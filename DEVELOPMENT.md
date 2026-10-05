# Development

How to build WeatherRibbon from source and work on it. Commands run from the repository root:
PowerShell on Windows, the Terminal's shell elsewhere. Testing is in [TESTING.md](TESTING.md).

The delivery TimeRibbon has (`build.ps1` reading `VERSION` and running the gate first, the setup
program, the DMG and the Flatpak) is not yet ported; it is the next piece of work (REQUIREMENTS.md
section 5). Until then a build is `wails build` and carries the development version.

## What a Windows machine needs

| Tool | Version | What for | Where from |
|---|---|---|---|
| Go | 1.26.3, as `go.mod` declares | the application and its tools | [go.dev/dl](https://go.dev/dl/) |
| Node.js with npm | a current LTS | the front end, its lint, type check and tests | [nodejs.org](https://nodejs.org/) or `winget install OpenJS.NodeJS.LTS` |
| Wails CLI | v2.12.0, the module `go.mod` requires | the executable and running from source | `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0` |
| WebView2 runtime | any current | the window; Windows 11 ships it | Microsoft's WebView2 page |
| Python 3, Pillow | any current | only to regenerate the committed icons | [python.org](https://www.python.org/), `python -m pip install pillow` |

The gate fetches staticcheck through `go run` on its first run, so it needs the network once. No C
compiler is needed. If `wails` is not found, `%USERPROFILE%\go\bin` is not on the path:

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

Run the gate first and read its exit code ([TESTING.md](TESTING.md#running-it)), then build with cgo
off and the kit tag `go.mod` requires rather than a working copy a local `go.work` names:

```powershell
./test.ps1
```

```powershell
$env:CGO_ENABLED = '0'; $env:GOWORK = 'off'
```

```powershell
wails build
```

Wails builds the page (whose build runs `eslint` and `tsc --noEmit` first), then the executable,
`build/bin/WeatherRibbon.exe`. It reports `0.0.0-dev` as its version: `product.Version` is a var
holding that placeholder, which `build.ps1` will stamp from `VERSION` through `-ldflags -X` (CON-4).
The executable is not signed.

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

Go then builds WeatherRibbon against `../ribbonkit`. Once the kit is tagged, name the tag in `go.mod`
(`go get github.com/oernster/ribbonkit@<tag>`) and in `frontend/package.json`, run
`npm --prefix frontend install`, check that `package-lock.json` resolved the tag's commit, then delete
`go.work`. `TestBothHalvesOfTheKitNameOneTag` fails while the two name different tags.

## Generated files

All committed, so a clone builds without regenerating them.

- **Icons**, after changing an image in `assets/`: `python tools/genicons.py` writes
  `build/windows/icon.ico` (the executable, taskbar, tray and shortcuts), `build/appicon.png` and the
  page's artwork (`frontend/src/assets`: the donate mark, the Add city picture and the icon About
  shows).
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
| `test.ps1` | the gate |
| `frontend/src` | the React page |
| `tests/structural`, `tools/`, `assets/` | the architecture's tests; generators; master artwork |

## House rules worth knowing before a first change

- **The layer direction is enforced,** the kit's packages counting as layers too: the domain imports
  nothing outside a domain and reads no clock; the application imports neither infrastructure nor
  Wails; nothing below the UI imports it; only `main.go` wires application to infrastructure, handing
  the kit's window the desktop through the `shell.Desktop` port.
- **The ribbon's behaviour is the kit's.** A change to placing, dragging, the tab, the grip, the tray,
  sign-in or the update check is made in ribbonkit, tagged there, then named here.
- **Every exported method of `window.Window` is page API,** since `App` embeds it and Wails binds
  promoted methods too. What WeatherRibbon alone may call goes on `window.Control`, which is never
  embedded.
- **No file over 400 lines,** tests included; one between 381 and 400 goes down to 350 or fewer.
- **No magic numbers:** a literal needing a comment is a named constant or derived from data.
- **The product is named once,** in `internal/product/product.go`.
- **WeatherRibbon's half of the wire is written twice,** in `dto.go` and in `frontend/src/wire.ts`;
  change both sides.
- **Both halves of the kit name one tag,** in `go.mod` and `frontend/package.json`
  (`kittag_test.go`).
- **Only `metno` reaches the network,** besides the kit's update check: nothing else imports a network
  package, starts a program or asks a network from the page (`network_test.go`).
- **A page call Go can refuse takes a refusal handler** and answers null rather than rejecting.
- **Every new guard is proved by planting a violation** ([TESTING.md](TESTING.md#keeping-this-honest)).

See also [README.md](README.md), [TESTING.md](TESTING.md) and [ARCHITECTURE.md](ARCHITECTURE.md).
