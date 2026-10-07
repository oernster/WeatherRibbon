#!/usr/bin/env bash
# Builds the WeatherRibbon macOS DMG for Apple Silicon. Run on an arm64 Mac from the repo root, in a
# Terminal at the Mac itself: signing and notarizing read the login keychain, which refuses a
# remote shell (measured 2026-09-28 over SSH: codesign failed with errSecInternalComponent).
#
#   bash builddmg.sh
#
# Ported from PigeonPost's builddmg.sh. PigeonPost builds with the wails command; WeatherRibbon builds
# with go build, as its Flatpak does, so the bundle is put together here. Flow: build the page, the
# executable and the icon, assemble WeatherRibbon.app, codesign it (hardened runtime), notarize and
# staple it, stage it with ditto, create-dmg, stamp the DMG file icon, sign the DMG, then notarize
# and staple the DMG.
#
# Notarization is mandatory. A Developer ID signature alone is not enough: since macOS 10.15
# Gatekeeper rejects signed but unnotarized apps with "Apple could not verify ... is free of
# malware". With APPLE_ID and APPLE_APP_PASSWORD both set it notarizes as that Apple ID (a password
# not shaped like an app-specific one stops the build before anything is built); with either unset
# it notarizes through the keychain profile.
#
# Environment overrides:
#   DEVELOPER_ID_APPLICATION   signing identity (defaults to Oliver's Developer ID; "-" signs ad hoc)
#   APPLE_ID, APPLE_APP_PASSWORD, APPLE_TEAM_ID   notarization credentials (else the profile)
#   APPLE_KEYCHAIN_PROFILE     keychain profile name (defaults to WeatherRibbon)
#   ALLOW_UNNOTARIZED=1        build without notarizing; local testing only, never released
#
# Output: WeatherRibbon.dmg in the repo root
set -euo pipefail

section() { printf '\n\033[1m== %s ==\033[0m\n' "$1"; }

section "Platform guard"
[ "$(uname -s)" = "Darwin" ] || { echo "error: this script must run on macOS" >&2; exit 1; }
[ "$(uname -m)" = "arm64" ] || { echo "error: this script targets Apple Silicon (arm64)" >&2; exit 1; }
command -v go > /dev/null 2>&1 || { echo "error: go is required (install from https://go.dev/dl/)" >&2; exit 1; }

# The names come from the one home each has in the code.
eval "$(go run ./tools/identity)"
MODULE="$(go list -m)"
VERSION="$(tr -d '[:space:]' < VERSION)"
APP_BUNDLE="build/bin/${APP_NAME}.app"
DIST_DIR="dist-dmg"
STAGE_DIR="${DIST_DIR}/stage"
WORK_DIR="${DIST_DIR}/work"
# DIST_DIR is scratch space; the final DMG lands in the repo root.
DMG_PATH="${APP_NAME}.dmg"
ICON_SOURCE="build/appicon.png"
ICON_FILE="iconfile"
# The point sizes an iconset holds, each at 1x and 2x, as iconutil expects them.
ICONSET_SIZES="16 32 128 256 512"

DEVELOPER_ID="${DEVELOPER_ID_APPLICATION:-Developer ID Application: Oliver Ernster (W7K465GKFJ)}"
APPLE_ID="${APPLE_ID:-}"
APPLE_APP_PASSWORD="${APPLE_APP_PASSWORD:-}"
APPLE_TEAM_ID="${APPLE_TEAM_ID:-W7K465GKFJ}"
# Escape hatch for local test builds. Distribution builds must never set this: an unnotarized DMG
# is rejected by Gatekeeper on every machine but the one that signed it; the failure is invisible
# at build time.
ALLOW_UNNOTARIZED="${ALLOW_UNNOTARIZED:-}"
# The notarization credential for this app, created once in a Terminal at the Mac with
#   xcrun notarytool store-credentials WeatherRibbon --apple-id <id> --team-id <team>
# which asks for the app-specific password. One profile per app means a leaked credential can be
# revoked for a single app. Stated explicitly rather than derived from APP_NAME: the profile is a
# fact registered in the keychain; deriving it would silently change which credential the build
# looks for if that name were ever edited. APPLE_KEYCHAIN_PROFILE overrides it.
NOTARY_PROFILE="${APPLE_KEYCHAIN_PROFILE:-WeatherRibbon}"
# The notary service accepts only an app-specific password from appleid.apple.com and rejects the
# Apple account password with HTTP 401. The shape is distinctive, so it is checked before the build
# rather than discovered after it.
APP_SPECIFIC_PASSWORD_RE='^[a-z]{4}-[a-z]{4}-[a-z]{4}-[a-z]{4}$'
# Notarization is the default and the keychain profile always resolves, so the only way to skip it
# is to ask for that explicitly.
NOTARIZING=1
[ "${ALLOW_UNNOTARIZED}" = "1" ] && NOTARIZING=0

# notarytool_submit uploads a file to Apple and waits for the verdict. notarytool exits non-zero on
# an Invalid verdict; set -e then fails the build rather than leaving an artifact that looks
# distributable. Stapling is a separate step because the file submitted and the file carrying the
# ticket differ for a bundle: a zip goes up, the .app gets stapled. The password never reaches the
# echoed command: with a keychain profile it is not on the command line at all; the fallback branch
# prints a masked form.
notarytool_submit() {
    if [ -n "${APPLE_ID}" ] && [ -n "${APPLE_APP_PASSWORD}" ]; then
        echo "\$ xcrun notarytool submit $1 --apple-id ${APPLE_ID} --password ******** --team-id ${APPLE_TEAM_ID} --wait"
        xcrun notarytool submit "$1" \
            --apple-id "${APPLE_ID}" \
            --password "${APPLE_APP_PASSWORD}" \
            --team-id "${APPLE_TEAM_ID}" \
            --wait
        return
    fi
    echo "\$ xcrun notarytool submit $1 --keychain-profile ${NOTARY_PROFILE} --wait"
    xcrun notarytool submit "$1" \
        --keychain-profile "${NOTARY_PROFILE}" \
        --wait
}

# require ensures a tool is on PATH, running the given install command if it is not.
require() {
    local tool="$1" install="$2"
    if ! command -v "$tool" > /dev/null 2>&1; then
        section "Installing missing tool: $tool"
        eval "$install"
    fi
    command -v "$tool" > /dev/null 2>&1 || { echo "error: $tool is required (install: $install)" >&2; exit 1; }
}

# stamp_dmg_icon gives the .dmg file itself a custom Finder icon. create-dmg only sets the mounted
# volume icon, so without this the .dmg shows the generic disk-image icon in Finder. Cosmetic: warn
# and skip if the icns or the classic resource tools are unavailable rather than failing the build.
stamp_dmg_icon() {
    local dmg="$1" icns="$2" tool
    [ -f "$icns" ] || { echo "warning: ${icns} missing; DMG file icon not set" >&2; return 0; }
    for tool in sips DeRez Rez SetFile; do
        command -v "$tool" > /dev/null 2>&1 || { echo "warning: ${tool} missing; DMG file icon not set" >&2; return 0; }
    done
    local work
    work="$(mktemp -d)"
    cp "$icns" "${work}/icon.icns"
    sips -i "${work}/icon.icns" > /dev/null
    DeRez -only icns "${work}/icon.icns" > "${work}/icon.rsrc"
    Rez -append "${work}/icon.rsrc" -o "$dmg"
    SetFile -a C "$dmg"
    rm -rf "$work"
}

section "Notarization credentials"
# Checked before any build work so a missing password costs a second rather than a full build.
if [ "${NOTARIZING}" -eq 0 ]; then
    echo "warning: ALLOW_UNNOTARIZED=1; local test build only, do not release the result" >&2
elif [ -n "${APPLE_ID}" ] && [ -n "${APPLE_APP_PASSWORD}" ]; then
    if ! [[ "${APPLE_APP_PASSWORD}" =~ ${APP_SPECIFIC_PASSWORD_RE} ]]; then
        cat >&2 << EOF
error: APPLE_APP_PASSWORD is not an app-specific password.
Expected four lowercase groups of four, like abcd-efgh-ijkl-mnop.
An Apple account password is rejected by the notary service with
'HTTP status code: 401. Invalid credentials'.
Generate one at https://appleid.apple.com (Sign-In and Security, App-Specific
Passwords). Alternatively leave both variables unset and store the credential
in the keychain as profile ${NOTARY_PROFILE}.
EOF
        exit 1
    fi
    echo "Notarizing as ${APPLE_ID} (team ${APPLE_TEAM_ID})"
else
    echo "Notarizing with keychain profile ${NOTARY_PROFILE}"
fi

section "Tooling"
require npm "brew install node"
require create-dmg "brew install create-dmg"

rm -rf "${DIST_DIR}" "${APP_BUNDLE}"
rm -f "${DMG_PATH}"
mkdir -p "${WORK_DIR}" "${STAGE_DIR}"

section "Building the page"
# Go embeds frontend/dist, so the page is built before the executable.
npm --prefix frontend install --no-audit --no-fund
npm --prefix frontend run build

# minos answers the oldest macOS an executable says it runs on.
minos() { vtool -show-build "$1" | awk '$1 == "minos" { print $2; exit }'; }

section "Finding the oldest macOS Go supports"
# Left alone, cgo compiles for the macOS the build runs on, so a build on the newest macOS refuses
# to open on any older one (measured 2026-09-28: minos 26.0 on macOS 26). The oldest macOS worth
# claiming is the one the Go runtime itself needs, which a pure Go program built by this toolchain
# records; it is read from one rather than written here, so it follows the toolchain.
FLOOR_DIR="${WORK_DIR}/floor"
mkdir -p "${FLOOR_DIR}"
printf 'module floor\n' > "${FLOOR_DIR}/go.mod"
printf 'package main\n\nfunc main() {}\n' > "${FLOOR_DIR}/main.go"
(cd "${FLOOR_DIR}" && CGO_ENABLED=0 go build -o floor .)
TARGET_MACOS="$(minos "${FLOOR_DIR}/floor")"
[ -n "${TARGET_MACOS}" ] || { echo "error: could not read Go's oldest macOS" >&2; exit 1; }
echo "Go supports macOS ${TARGET_MACOS} and later"

section "Building ${APP_NAME} ${VERSION} (darwin/arm64)"
# The target reaches the C and Objective-C compiler through the cgo flags rather than the
# MACOSX_DEPLOYMENT_TARGET variable: Go's build cache keys on the flags but not on that variable,
# so with it alone objects compiled earlier for a newer macOS were reused (measured 2026-09-28: the
# linker warned that objects built for 26.0 were linked for 12.0).
VERSION_FLAG="-mmacosx-version-min=${TARGET_MACOS}"
export CGO_CFLAGS="${CGO_CFLAGS:--O2 -g} ${VERSION_FLAG}"
export CGO_LDFLAGS="${CGO_LDFLAGS:-} ${VERSION_FLAG}"
EXECUTABLE="${WORK_DIR}/${APP_NAME}"
BUILD_LOG="${WORK_DIR}/build.log"
# What ships is built from the ribbonkit tag go.mod requires, never a local go.work's working copy.
GOWORK=off go build -tags desktop,production -ldflags "-s -w -X ${MODULE}/internal/product.Version=${VERSION}" -o "${EXECUTABLE}" . 2>&1 | tee "${BUILD_LOG}"
# Linking code built for a newer macOS succeeds with only a warning, leaving an executable that
# claims an older macOS than its code was built for. That is refused rather than shipped.
if grep -q 'was built for newer' "${BUILD_LOG}"; then
    echo "error: code built for a newer macOS than ${TARGET_MACOS} was linked in; see ${BUILD_LOG}" >&2
    exit 1
fi
# The oldest macOS the bundle claims is the one the executable was built for, read from it rather
# than stated here, so the two cannot disagree.
MIN_MACOS="$(minos "${EXECUTABLE}")"
if [ "${MIN_MACOS}" != "${TARGET_MACOS}" ]; then
    echo "error: ${EXECUTABLE} needs macOS ${MIN_MACOS}, not ${TARGET_MACOS} as asked" >&2
    exit 1
fi
echo "Minimum macOS: ${MIN_MACOS}"

section "Making the icon"
ICONSET="${WORK_DIR}/${ICON_FILE}.iconset"
mkdir -p "${ICONSET}"
for size in ${ICONSET_SIZES}; do
    sips -z "${size}" "${size}" "${ICON_SOURCE}" --out "${ICONSET}/icon_${size}x${size}.png" > /dev/null
    double=$((size * 2))
    sips -z "${double}" "${double}" "${ICON_SOURCE}" --out "${ICONSET}/icon_${size}x${size}@2x.png" > /dev/null
done
iconutil -c icns "${ICONSET}" -o "${WORK_DIR}/${ICON_FILE}.icns"

section "Assembling ${APP_BUNDLE}"
mkdir -p "${APP_BUNDLE}/Contents/MacOS" "${APP_BUNDLE}/Contents/Resources"
cp "${EXECUTABLE}" "${APP_BUNDLE}/Contents/MacOS/${APP_NAME}"
cp "${WORK_DIR}/${ICON_FILE}.icns" "${APP_BUNDLE}/Contents/Resources/${ICON_FILE}.icns"
# LSUIElement makes the app check in with macOS as an agent, so the Dock never records it as a
# recent app. Without it the app checked in as a regular one before any code ran and left a recents
# tile behind at every launch (measured 2026-10-07 from the Mac's log). Wails still makes the app
# regular as it launches; ribbonkit switches it back, so both are needed.
cat > "${APP_BUNDLE}/Contents/Info.plist" << PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleName</key>
	<string>${APP_NAME}</string>
	<key>CFBundleDisplayName</key>
	<string>${APP_NAME}</string>
	<key>CFBundleExecutable</key>
	<string>${APP_NAME}</string>
	<key>CFBundleIdentifier</key>
	<string>${APP_ID}</string>
	<key>CFBundleShortVersionString</key>
	<string>${VERSION}</string>
	<key>CFBundleVersion</key>
	<string>${VERSION}</string>
	<key>CFBundleIconFile</key>
	<string>${ICON_FILE}</string>
	<key>LSMinimumSystemVersion</key>
	<string>${MIN_MACOS}</string>
	<key>LSUIElement</key>
	<true/>
	<key>NSHighResolutionCapable</key>
	<true/>
	<key>NSHumanReadableCopyright</key>
	<string>${COPYRIGHT}</string>
</dict>
</plist>
PLIST
plutil -lint "${APP_BUNDLE}/Contents/Info.plist"

section "Codesigning the app bundle"
codesign --force --deep --options runtime --sign "${DEVELOPER_ID}" "${APP_BUNDLE}"
codesign --verify --deep --strict "${APP_BUNDLE}"

# Notarize the .app before it goes into the DMG. Stapling only the DMG leaves the copied-out .app
# with no local ticket, so Gatekeeper falls back to an online check and the app fails to launch for
# anyone offline or behind a restrictive network. notarytool takes archives only, so ditto zips the
# bundle first; the ticket goes on the bundle, since a zip cannot carry one.
if [ "${NOTARIZING}" -eq 1 ]; then
    section "Notarizing the app bundle (this waits on Apple)"
    APP_ZIP="$(mktemp -d)/${APP_NAME}.zip"
    ditto -c -k --keepParent "${APP_BUNDLE}" "${APP_ZIP}"
    notarytool_submit "${APP_ZIP}"
    xcrun stapler staple "${APP_BUNDLE}"
    rm -rf "$(dirname "${APP_ZIP}")"
fi

section "Creating the DMG"
# ditto preserves the symlinks and metadata the embedded signature depends on.
ditto "${APP_BUNDLE}" "${STAGE_DIR}/${APP_NAME}.app"
VOLICON="${APP_BUNDLE}/Contents/Resources/${ICON_FILE}.icns"
CREATE_DMG_ARGS=(
    --volname "${APP_NAME}"
    --window-size 540 380
    --icon-size 128
    --icon "${APP_NAME}.app" 140 190
    --app-drop-link 400 190
    --volicon "${VOLICON}"
)
set +e
create-dmg "${CREATE_DMG_ARGS[@]}" "${DMG_PATH}" "${STAGE_DIR}"
STATUS=$?
set -e
# create-dmg exits 2 when it cannot set a custom window background (headless); still a good DMG.
if [ "${STATUS}" -ne 0 ] && [ "${STATUS}" -ne 2 ]; then
    echo "error: create-dmg failed with exit ${STATUS}" >&2
    exit "${STATUS}"
fi

# The icon stamp writes a resource fork into the DMG, so it runs before signing and notarization.
# Doing it afterwards would modify a file Gatekeeper has already been told the hash of.
section "Stamping the DMG file icon"
stamp_dmg_icon "${DMG_PATH}" "${VOLICON}"
rm -rf "${DIST_DIR}"

section "Signing the DMG"
codesign --force --sign "${DEVELOPER_ID}" "${DMG_PATH}"
codesign --verify "${DMG_PATH}"

if [ "${NOTARIZING}" -eq 1 ]; then
    section "Notarizing the DMG (this waits on Apple)"
    notarytool_submit "${DMG_PATH}"
    xcrun stapler staple "${DMG_PATH}"
    # stapler validate proves a ticket is attached; spctl replays the check Gatekeeper runs on the
    # end user's machine. Together they catch the silent case where signing succeeded but
    # notarization never happened.
    xcrun stapler validate "${DMG_PATH}"
    spctl --assess --type install -vv "${DMG_PATH}"
else
    section "Notarization skipped: unnotarized DMG, do not publish this build"
fi

section "Done"
echo "${DMG_PATH}"
