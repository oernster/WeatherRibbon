#!/usr/bin/env bash
# Builds the WeatherRibbon Flatpak for Linux. Run from the repo root on a Linux machine:
#
#   bash build_flatpak.sh
#
# Ported from PigeonPost's build_flatpak.sh, the house's Go and Wails Flatpak. Flow: install the
# flatpak tooling if missing, add flathub, install the GNOME runtime with the golang and node SDK
# extensions, write the desktop file, metainfo and manifest, build inside the sandbox (the page with
# npm, the icons, then a CGO Go build against the runtime's webkit2gtk-4.1), install it for the
# current user and export a bundle.
#
# The GNOME runtime is required because Wails v2 renders through webkit2gtk-4.1, which the
# freedesktop runtime does not carry; the Go build therefore uses -tags webkit2_41.
#
# The sandbox's permissions are only what WeatherRibbon uses: X11 and no Wayland, since the ribbon
# must place its own window; the tray host's bus name; the bus name of Wails' single-instance lock;
# the session's autostart folder for Start at sign-in; the runtime folder's ribbonkit, which every
# ribbon running shares (FR-506); the network, for the forecasts and sun times from api.met.no and
# the update check to GitHub (NFR-S-1). No other files.
#
# Outputs: weatherribbon.flatpak (installable anywhere) and a user install of the app.
set -euo pipefail

eval "$(go run ./tools/identity)"
MODULE="$(go list -m)"
VERSION="$(tr -d '[:space:]' < VERSION)"
RELEASE_DATE="$(date +%F)"
APP_SUMMARY="A ribbon of weather for the cities you choose"
HOMEPAGE="https://github.com/oernster/WeatherRibbon"
RUNTIME="org.gnome.Platform"
SDK="org.gnome.Sdk"
RUNTIME_VERSION="50"
SDK_EXT_VERSION="25.08"   # the freedesktop base of GNOME 50; extensions pair with it
GOLANG_EXT="org.freedesktop.Sdk.Extension.golang"
NODE_EXT="org.freedesktop.Sdk.Extension.node22"
BUILD_DIR=".flatpak-build"
REPO_DIR=".flatpak-repo"
BUNDLE="${BIN_NAME}.flatpak"
MANIFEST="${APP_ID}.yml"
PACKAGING_DIR="packaging"
ICON_DIR="build/linux/icons"

section() { printf '\n\033[1m== %s ==\033[0m\n' "$1"; }

install_if_missing() {
    local tool="$1"
    command -v "$tool" > /dev/null 2>&1 && return 0
    section "Installing missing tool: $tool"
    if command -v apt-get > /dev/null 2>&1; then sudo apt-get install -y "$tool"
    elif command -v dnf > /dev/null 2>&1; then sudo dnf install -y "$tool"
    elif command -v pacman > /dev/null 2>&1; then sudo pacman -S --noconfirm "$tool"
    elif command -v zypper > /dev/null 2>&1; then sudo zypper install -y "$tool"
    else echo "error: install $tool with your package manager and re-run" >&2; exit 1
    fi
}

section "Tooling"
install_if_missing flatpak
install_if_missing flatpak-builder

section "Flathub remote and runtimes"
flatpak remote-add --if-not-exists --user flathub https://dl.flathub.org/repo/flathub.flatpakrepo
flatpak install --user --noninteractive flathub \
    "${RUNTIME}//${RUNTIME_VERSION}" \
    "${SDK}//${RUNTIME_VERSION}" \
    "${GOLANG_EXT}//${SDK_EXT_VERSION}" \
    "${NODE_EXT}//${SDK_EXT_VERSION}"

section "Writing packaging files"
mkdir -p "${PACKAGING_DIR}"

# StartupWMClass is the window class GTK gives the ribbon, the executable's name (measured on
# Ubuntu: WM_CLASS "weatherribbon").
cat > "${PACKAGING_DIR}/${APP_ID}.desktop" << DESKTOP
[Desktop Entry]
Name=${APP_NAME}
Comment=${APP_SUMMARY}
Exec=${BIN_NAME}
Icon=${APP_ID}
Terminal=false
Type=Application
Categories=Utility;
StartupWMClass=${BIN_NAME}
DESKTOP

cat > "${PACKAGING_DIR}/${APP_ID}.metainfo.xml" << METAINFO
<?xml version="1.0" encoding="UTF-8"?>
<component type="desktop-application">
  <id>${APP_ID}</id>
  <name>${APP_NAME}</name>
  <summary>${APP_SUMMARY}</summary>
  <metadata_license>CC0-1.0</metadata_license>
  <project_license>GPL-3.0-only</project_license>
  <description>
    <p>
      A slim ribbon of weather that stands against an edge of the screen, one cell for each city
      chosen: its local time, the weather now, today and three days ahead, from MET Norway. It
      asks MET Norway for each city's forecast by its coordinates and GitHub for the latest
      release; it issues no weather warnings.
    </p>
  </description>
  <launchable type="desktop-id">${APP_ID}.desktop</launchable>
  <url type="homepage">${HOMEPAGE}</url>
  <content_rating type="oars-1.1"/>
  <releases>
    <release version="${VERSION}" date="${RELEASE_DATE}"/>
  </releases>
</component>
METAINFO

section "Writing manifest"
# This heredoc is unquoted, so each dollar meant for the build's own shell is escaped.
cat > "${MANIFEST}" << MANIFEST_EOF
app-id: ${APP_ID}
runtime: ${RUNTIME}
runtime-version: '${RUNTIME_VERSION}'
sdk: ${SDK}
sdk-extensions:
  - ${GOLANG_EXT}
  - ${NODE_EXT}
command: ${BIN_NAME}
finish-args:
  - --share=ipc
  - --socket=x11
  - --device=dri
  - --talk-name=org.kde.StatusNotifierWatcher
  - --own-name=${SINGLE_INSTANCE_NAME}
  - --filesystem=xdg-config/autostart:create
  # The folder every ribbon on ribbonkit shares, so two products never land on each other (FR-506).
  - --filesystem=xdg-run/ribbonkit:create
  # The forecasts and sun times from api.met.no and the update check to GitHub (NFR-S-1).
  - --share=network
build-options:
  append-path: /usr/lib/sdk/golang/bin:/usr/lib/sdk/node22/bin
  build-args:
    - --share=network
  env:
    GOPATH: /run/build/${BIN_NAME}/gopath
    GOCACHE: /run/build/${BIN_NAME}/gocache
    GOFLAGS: -buildvcs=false
    # What ships is built from the ribbonkit tag go.mod requires, never a local go.work's copy.
    GOWORK: "off"
    npm_config_cache: /run/build/${BIN_NAME}/npm-cache
modules:
  - name: ${BIN_NAME}
    buildsystem: simple
    build-commands:
      - cd frontend && npm install --no-audit --no-fund && npm run build
      - go run ./tools/linuxicons -in build/appicon.png -out ${ICON_DIR} -name ${BIN_NAME}
      - go build -tags desktop,production,webkit2_41 -ldflags "-s -w -X ${MODULE}/internal/product.Version=${VERSION}" -o ${BIN_NAME} .
      - install -Dm755 ${BIN_NAME} /app/bin/${BIN_NAME}
      - chmod -R u+w gopath gocache 2>/dev/null || true
      - install -Dm644 packaging/${APP_ID}.desktop /app/share/applications/${APP_ID}.desktop
      - install -Dm644 packaging/${APP_ID}.metainfo.xml /app/share/metainfo/${APP_ID}.metainfo.xml
      - for icon in ${ICON_DIR}/${BIN_NAME}_*.png; do size=\$(basename "\$icon" .png); size=\${size##*_}; install -Dm644 "\$icon" /app/share/icons/hicolor/\${size}x\${size}/apps/${APP_ID}.png; done
    sources:
      - type: dir
        path: .
        skip:
          - .git
          - ${BUILD_DIR}
          - ${REPO_DIR}
          - .flatpak-builder
          - ${BUNDLE}
          - frontend/node_modules
MANIFEST_EOF

# A copy left running holds the single-instance lock, so the new build's first launch would only
# toggle the old ribbon and exit (FR-602). flatpak kill fails when nothing runs, hence the check.
section "Stopping a running ${APP_NAME}"
if flatpak ps --columns=application | grep -x "${APP_ID}" > /dev/null; then
    flatpak kill "${APP_ID}"
    echo "  Stopped."
else
    echo "  Not running, skipping."
fi

section "Building ${APP_NAME} ${VERSION} with flatpak-builder"
flatpak-builder --user --install --force-clean \
    --install-deps-from=flathub \
    --repo="${REPO_DIR}" \
    "${BUILD_DIR}" "${MANIFEST}"

section "Exporting bundle"
flatpak build-bundle \
    --runtime-repo=https://dl.flathub.org/repo/flathub.flatpakrepo \
    "${REPO_DIR}" "${BUNDLE}" "${APP_ID}"

section "Done"
echo "Installed for the current user: flatpak run ${APP_ID}"
echo "Distributable bundle: ${BUNDLE}"
