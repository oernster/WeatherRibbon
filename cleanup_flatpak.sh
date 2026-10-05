#!/usr/bin/env bash
# cleanup_flatpak.sh: uninstalls the WeatherRibbon Flatpak and removes its build artefacts. Run from the
# repo root. Ported from PigeonPost's.
#
# It wipes the Flatpak entirely: the app, all of its data in ~/.var/app/<id> (settings with the
# cities, the forecast cache, the log, the web view's cache) and the Start at sign-in entry it may
# have written, which would otherwise go on trying to start an app that is gone. The next build's first launch is then a
# true first run, opening flush against the right edge, centred (FR-502; Oliver, 2026-10-04). It
# deliberately does not touch what the other build paths produce, so they stay independent.
set -euo pipefail

eval "$(go run ./tools/identity)"
AUTOSTART="${XDG_CONFIG_HOME:-${HOME}/.config}/autostart/${APP_ID}.desktop"
APP_DATA="${HOME}/.var/app/${APP_ID}"

bold=$(tput bold 2>/dev/null || true)
reset=$(tput sgr0 2>/dev/null || true)
section() { echo; echo "${bold}=== $* ===${reset}"; }

# Uninstalling leaves a running copy running, holding the single-instance lock, so the next
# install's first launch would only toggle it and exit (FR-602). flatpak kill fails when nothing
# runs, hence the check.
section "Stopping a running ${APP_NAME}"
if flatpak ps --columns=application | grep -x "${APP_ID}" > /dev/null; then
    flatpak kill "${APP_ID}"
    echo "  Stopped."
else
    echo "  Not running, skipping."
fi

section "Uninstalling ${APP_ID}"
if flatpak list --user --app --columns=application | grep -qx "${APP_ID}"; then
    flatpak uninstall --user -y --delete-data "${APP_ID}"
    echo "  Uninstalled."
else
    echo "  Not installed, skipping."
fi

# --delete-data covers an installed app; this also catches data an earlier uninstall left behind.
section "Removing ${APP_NAME}'s data and settings"
if [[ -d "${APP_DATA}" ]]; then
    rm -rf "${APP_DATA}"
    echo "  Removed ${APP_DATA}"
else
    echo "  None, skipping."
fi

section "Removing the Start at sign-in entry"
if [[ -f "${AUTOSTART}" ]]; then
    rm -f "${AUTOSTART}"
    echo "  Removed ${AUTOSTART}"
else
    echo "  None, skipping."
fi

section "Removing flatpak build artefacts"
rm -f "${BIN_NAME}.flatpak"
# Go writes its module cache read-only. A build that fails leaves its build directory behind with
# that cache in it, before the manifest's own chmod has run, so plain rm cannot remove it.
[[ -d .flatpak-builder ]] && chmod -R u+w .flatpak-builder
rm -rf .flatpak-build .flatpak-repo .flatpak-builder build/linux
rm -f "${APP_ID}.yml"
rm -rf packaging/
echo "  Done."

echo
echo "${bold}Purge complete.${reset}"
