# Development

How to build WeatherRibbon from nothing to what ships: the setup program on Windows, the DMG on
macOS and the Flatpak on Linux, each built on a machine of its platform. Commands run from the
repository root: PowerShell on Windows, the Terminal's shell elsewhere. Testing is in
[TESTING.md](TESTING.md).

## What a Windows machine needs

| Tool | Version | What for | Where from |
|---|---|---|---|
| Go | 1.26.3, as `go.mod` declares | the application, setup and tools | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm | a current LTS | the front end, its lint, type check and tests | [nodejs.org](https://nodejs.org/) or `winget install OpenJS.NodeJS.LTS` |
| Wails CLI | v2.12.0, the module `go.mod` requires | both executables and running from source | `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0` |
| WebView2 runtime | any current | the window; Windows 11 ships it | Microsoft's WebView2 page |
| Python 3, Pillow | any current | stamping the site; Pillow only to regenerate the committed icons | [python.org](https://www.python.org/), `python -m pip install pillow` |

The gate fetches staticcheck through `go run` on its first run, so it needs the network once. No C
compiler is needed: `build.ps1` pins cgo off. If `wails` is not found, `%USERPROFILE%\go\bin` is not
on the path:

```powershell
$env:PATH = "$env:USERPROFILE\go\bin;$env:PATH"
```

## Getting the source

```powershell
git clone https://github.com/oernster/WeatherRibbon.git
```

```powershell
cd WeatherRibbon
```

```powershell
go mod download
```

```powershell
npm --prefix frontend install
```

That fetches the kit's page half from GitHub at the tag `package.json` names, so it needs git on the
path; `go mod download` fetched its Go half at the tag `go.mod` names.

The application embeds the built page, which git does not hold, so build it once on a fresh clone;
without it the gate stops at `pattern all:frontend/dist: no matching files found`:

```powershell
npm --prefix frontend run build
```

## Building on Windows

```powershell
./build.ps1
```

It stops at the first failure:

1. Reads `VERSION`, refusing anything but `major.minor.patch`; stamps the site's version tokens and
   asset hashes through `python stamp_version.py`; reads the module path and executable names.
2. Pins `CGO_ENABLED=0`, so the tests exercise what ships; pins `GOWORK=off`, so what ships is the
   kit tag `go.mod` requires rather than a working copy a local `go.work` names.
3. Runs `test.ps1` ([TESTING.md](TESTING.md#running-it)), with no switch to skip it.
4. Refuses to go on without the committed `build/windows/icon.ico` and `build/appicon.png`.
5. Writes the version resource through `go run ./tools/versioninfo`, then runs `wails build` with the
   version in `-ldflags`; the page's build runs `eslint` and `tsc --noEmit` first.
6. Packs the application and `LICENSE` into `installer/payload.zip` (`go run ./tools/payload`).
7. Builds the setup program the same way and copies it to `dist-installer`.
8. Writes the empty placeholder back over `installer/payload.zip` whatever happened, so the real
   payload never reaches a commit.

| Output | What it is |
|---|---|
| `build/bin/WeatherRibbon.exe` | the application |
| `dist-installer/WeatherRibbonSetup.exe` | the setup program, application included |

`./build.ps1 -SkipInstaller` stops after the application. If a build is interrupted between steps 6
and 8, run it again rather than committing `installer/payload.zip`.

The version reaches both executables through
`-X github.com/oernster/weatherribbon/internal/product.Version=<version>` (CON-4). `-X` writes only to
a `var`, so `Version` is a var holding a development placeholder; an executable reporting it was not
built by `build.ps1`.

## Running it while working

```powershell
wails dev
```

It serves the page from Vite with hot reload and rebuilds Go on change. Each run appends to
`WeatherRibbon.log` in the settings folder (`%APPDATA%\WeatherRibbon`; the others are in
[ARCHITECTURE.md](ARCHITECTURE.md#data-locations)), the runtime's own panic report and a line per
forecast request among it. One copy runs per user: a second launch toggles the first ribbon and
exits. On macOS and Linux build the page and use `go run` with the build's tags
([TESTING.md](TESTING.md#on-macos-and-linux)).

Install what you built with `./dist-installer/WeatherRibbonSetup.exe`. Neither executable is signed.

Every run asks MET Norway for real forecasts, under its terms: at most one request a second, each
city again only once its forecast has expired. A refusal (HTTP 403) stops every request until the
next launch, so a run that is refused should be read in the log, not restarted in a loop.

## Changing the kit beside WeatherRibbon

The ribbon's desktop half is [ribbonkit](https://github.com/oernster/ribbonkit). To change it and see
the change here before it is tagged, clone it beside this repository and give WeatherRibbon a
`go.work` (gitignored), from the repository root:

```powershell
go work init . ../ribbonkit
```

Go then builds WeatherRibbon against `../ribbonkit`; `build.ps1` and the macOS and Linux scripts set
`GOWORK=off`, so a release is always built from the tag. Once the kit is tagged, name the tag in
`go.mod` (`go get github.com/oernster/ribbonkit@<tag>`) and in `frontend/package.json`, run
`npm --prefix frontend install`, check that `package-lock.json` resolved the tag's commit, then delete
`go.work`. `TestBothHalvesOfTheKitNameOneTag` fails while the two name different tags.

## Building on macOS

An Apple Silicon Mac with:

| Tool | What for | Where from |
|---|---|---|
| Xcode | cgo's compiler, `codesign`, `notarytool`, `stapler`, `vtool` | the App Store |
| Go, as `go.mod` declares | the application | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm | the front end | `brew install node`; the script installs it where missing |
| create-dmg | the DMG | `brew install create-dmg`; the script installs it where missing |
| A Developer ID Application certificate | signing | the Apple Developer account, in the login keychain |

Store the notary credential once, in a Terminal at the Mac, with an app-specific password from
appleid.apple.com. The script reads the profile `WeatherRibbon` (`APPLE_KEYCHAIN_PROFILE` names
another; `APPLE_ID` with `APPLE_APP_PASSWORD` notarises as that Apple ID instead):

```bash
xcrun notarytool store-credentials WeatherRibbon --apple-id <Apple ID> --team-id W7K465GKFJ
```

Build in a Terminal at the Mac itself, since the keychain refuses a remote shell:

```bash
bash builddmg.sh
```

It stops at the first failure:

1. Refuses anything but an Apple Silicon Mac; reads names from `tools/identity` and `VERSION`; refuses
   an `APPLE_APP_PASSWORD` not shaped like an app-specific one.
2. Builds the page.
3. Reads the oldest macOS the Go toolchain supports from an empty Go program and hands it to cgo.
4. Builds with `go build -tags desktop,production`, refusing a link of code built for a newer macOS.
5. Makes `iconfile.icns` with `sips` and `iconutil`.
6. Assembles `build/bin/WeatherRibbon.app` with its `Info.plist`, which declares `LSUIElement` so
   macOS treats the application as an agent and the Dock never records it as a recent app.
7. Signs it with the hardened runtime, notarises and staples it.
8. Makes, signs, notarises and staples the DMG, then runs `stapler validate` and `spctl --assess`.

The output is `WeatherRibbon.dmg`. For a local trial, `DEVELOPER_ID_APPLICATION=-` signs ad hoc and
`ALLOW_UNNOTARIZED=1` skips notarising; such a DMG opens only on the Mac that built it. The script is
TimeRibbon's with the names changed.

## Building on Linux

For running and testing from source, with the packages as Ubuntu names them:

```bash
sudo apt-get install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev nodejs npm flatpak flatpak-builder
```

Go from the distribution or [go.dev/dl](https://go.dev/dl/). The Flatpak compiles inside the GNOME
SDK, which the script installs from Flathub with its golang and node22 extensions; it still needs Go
locally to read the names through `tools/identity`.

```bash
bash build_flatpak.sh
```

It writes the desktop entry, metainfo and manifest (gitignored), stops a copy left running, builds the
page, icons and executable (`-tags desktop,production,webkit2_41`) in the sandbox, installs the
result for the current user and exports `weatherribbon.flatpak`. The sandbox is granted the network,
which MET Norway's forecasts need; it is granted logind on the system bus too, so the ribbon leaves
as soon as a shutdown or restart is announced. `cleanup_flatpak.sh` wipes it entirely: the app, all its data in
`~/.var/app/uk.codecrafter.WeatherRibbon` (settings with the cities, the forecast cache, the log),
its sign-in entry and the build outputs, so the next build's first launch is a true first run. Both
scripts are TimeRibbon's with the names changed.

## Generated files

All committed, so a clone builds without regenerating them.

- **Icons**, after changing an image in `assets/`: `python tools/genicons.py` (the work is ribbonkit's `tools/genicons.py`, read where Go builds the kit from) writes
  `build/windows/icon.ico` (both executables, taskbar, tray and shortcuts), `build/appicon.png`, the
  setup page's mark and toggles (`installer/frontend/dist`) and the page's artwork
  (`frontend/src/assets`: the donate mark, the Add city picture and the icon About shows).
- **The weather symbols** in `frontend/src/assets/weather` are MET Norway's Yr icons, the SVGs of
  `metno/weathericons` as published, with their `LICENSE`. They are not generated; which symbols have
  an icon is read from these files (FR-412).
- **The city list**, after a new GeoNames export, from a folder holding `cities15000.zip`,
  `admin1CodesASCII.txt` and `countryInfo.txt` from
  [download.geonames.org/export/dump](https://download.geonames.org/export/dump/). Its head line
  records the date given:

```powershell
go run ./tools/gencities -dir C:\path\to\downloads -downloaded 2026-10-05
```

## Versioning and releases

`VERSION` holds the one version string. Each build script passes it through the same `-X` flag; the
version resources, the DMG's `Info.plist`, the Flatpak's metainfo and the site's tokens are written
from it. The setup program compares it with the Apps list's record to choose its screen; the update
check compares it with GitHub's latest release tag, so a development placeholder is never offered one.

1. Set `VERSION`.
2. Run `./build.ps1` and read its exit code; commit the site's stamped version with `VERSION`.
3. Run `bash builddmg.sh` on the Mac and `bash build_flatpak.sh` on Linux, with the checks in
   [TESTING.md](TESTING.md#on-macos-and-linux).
4. Check by hand what no test reaches (window, tray, focus, paint, the install) in all three builds.
5. Tag the commit `v` plus the version and publish a release with the three attached. The update check
   reads only the latest published release and offers the first asset ending `.exe`, `.dmg` or
   `.flatpak`, so setup must be the only `.exe`.

## Where things live

| Path | What it holds |
|---|---|
| `main.go` | the composition root; the cell size and the panel sizes |
| `app.go`, `measure.go` | the facade Wails binds: the kit's window embedded, plus WeatherRibbon's own calls (the cities, units and time format, the detail, the petrichor line, the measured cell width) and its menu actions |
| `kit.go`, `dto.go` | the page and the LICENSE embedded, the service adapted to the window's port; WeatherRibbon's half of the wire |
| `refresher.go` | the one goroutine every forecast request leaves from |
| `icons.go` | the Yr icon set's file names as the `Icons` port |
| `trayicon_*.go` | where the tray icon lies on macOS and Linux; none on Windows |
| `*_test.go` in the root | WeatherRibbon's half tested over a scripted service and a stand-in window (`fakes_test.go`); the page's calls checked against what is bound (`page_api_test.go`) |
| `internal/domain`, `internal/application` | WeatherRibbon's pure rules (forecasts, units, places, settings, back-off, petrichor); its use cases over their ports |
| `internal/infrastructure` | WeatherRibbon's own adapters: MET Norway, the forecast cache, the settings store, the city list and the request pacing |
| `internal/product` | names, repository, User-Agent, version, donation address, author, sign-in label, credits |
| `build.ps1`, `test.ps1`, `VERSION`, `stamp_version.py` | the Windows build; the gate; the version; the site stamp |
| `builddmg.sh`, `build_flatpak.sh`, `cleanup_flatpak.sh` | the macOS DMG; the Flatpak and its removal |
| `frontend/src`, `installer/` | the React page; the setup program's composition root, which carries the payload and the page's pictures (`installer/frontend/dist`, made by `tools/genicons.py`) |
| `tools/` | thin mains over the kit's delivery (`identity`, `linuxicons`, `payload`, `versioninfo`); WeatherRibbon's own generators (`gencities`, `genicons.py`) |
| `tests/structural`, `assets/`, `docs/` | the architecture's tests; master artwork; the site |

## House rules worth knowing before a first change

- **The layer direction is enforced,** the kit's packages counting as layers too: the domain imports
  nothing outside a domain and reads no clock; the application imports neither infrastructure nor
  Wails; nothing below the UI imports it; only `main.go` wires application to infrastructure, handing
  the kit's window the desktop through the `shell.Desktop` port.
- **The ribbon's behaviour is the kit's.** A change to placing, dragging, the tab, the grip, the tray,
  sign-in, the update check or setup is made in ribbonkit, tagged there, then named here.
- **Every exported method of `window.Window` is page API,** since `App` embeds it and Wails binds
  promoted methods too. What WeatherRibbon alone may call goes on `window.Control`, which is never
  embedded.
- **No file over 400 lines,** tests included; one between 381 and 400 goes down to 350 or fewer.
  Build scripts are exempt.
- **No magic numbers:** a literal needing a comment is a named constant or derived from data.
- **The product is named once,** in `internal/product/product.go`; the setup page is handed it.
- **WeatherRibbon's half of the wire is written twice,** in `dto.go` and in `frontend/src/wire.ts`;
  change both sides.
- **Both halves of the kit name one tag,** in `go.mod` and `frontend/package.json`
  (`kittag_test.go`).
- **Only `metno` reaches the network,** besides the kit's update check: nothing else imports a network
  package, starts a program or asks a network from the page (`network_test.go`).
- **A page call Go can refuse takes a refusal handler** and answers null rather than rejecting.
- **Every new guard is proved by planting a violation** ([TESTING.md](TESTING.md#keeping-this-honest)).

See also [README.md](README.md), [TESTING.md](TESTING.md) and [ARCHITECTURE.md](ARCHITECTURE.md).
