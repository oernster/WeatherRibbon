# WeatherRibbon Architecture

A desktop application for Windows, macOS and Linux showing a ribbon of weather, one cell per chosen
city. It reads the system clock, its settings file, its forecast cache, the built-in city list, the
desktop (displays, tray, sign-in entry), MET Norway's forecasts and GitHub's latest release. The
domain and application are the same code everywhere; each platform's own half sits in files its
build tags or names select.

The ribbon itself (placing, dragging, the tab, the grip, opacity, the tray and menus, sign-in start,
the update check) is [ribbonkit](https://github.com/oernster/ribbonkit), a Go module and npm package
WeatherRibbon depends on at one tag, shared with TimeRibbon. Its own ARCHITECTURE.md describes how a
ribbon works; this document describes what WeatherRibbon puts in it and how the two are joined.

Its network requests are the forecasts and the sun times, which `internal/infrastructure/metno`
makes, plus the update check (FR-603), which is the kit's; WeatherRibbon hands it
`product.Repository`. No other file of WeatherRibbon's imports a network package, none starts
a program and its page asks no network. The donation page and a release's download go to the
desktop's browser.

FR, NFR and CON numbers are those of [REQUIREMENTS.md](REQUIREMENTS.md).

## Invariant

`UI -> Application -> Domain <- Infrastructure`

Dependencies point inward. Every rule below is a test; a guard not listed here does not exist. The
structural tests run on the kit's `structure` package, the code that holds the kit and TimeRibbon to
the same rules; the kit's own invariants are listed in its ARCHITECTURE.md.

| Invariant | Enforcing test | File |
|---|---|---|
| Domain imports nothing of WeatherRibbon or the kit outside a domain | `TestDomainHasNoOutwardImports` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Domain is pure: no IO, randomness or tz package; no wall clock read, no zone loaded (CON-5, FR-413) | `TestDomainIsPure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Application never imports infrastructure (WeatherRibbon's or the kit's) or Wails | `TestApplicationDoesNotImportInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Neither application nor infrastructure imports a UI package | `TestNothingBelowTheUIImportsIt` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Infrastructure never imports Wails | `TestWailsStaysOutOfInfrastructure` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Only `main.go` imports both application and infrastructure, the kit's included | `TestCompositionRootIsWhitelisted` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No Go file and no TypeScript or CSS file under `frontend/src` exceeds 400 lines (CON-2) | `TestNoFileExceedsLineLimit` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| No such file sits in the danger band of 381 to 400 lines (CON-2) | `TestNoFileInDangerBand` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| Every exported type carries a doc comment | `TestEveryExportedTypeIsDocumented` | [`boundary_test.go`](tests/structural/boundary_test.go) |
| `go.mod` and the front end's `package.json` name the same kit tag | `TestBothHalvesOfTheKitNameOneTag` | [`kittag_test.go`](tests/structural/kittag_test.go) |
| Only `internal/infrastructure/metno` imports a network package (NFR-S-1) | `TestOnlyTheForecastAndUpdateImportANetworkPackage` | [`network_test.go`](tests/structural/network_test.go) |
| No Go file of WeatherRibbon's starts a program or names a Windows library outside the kit's `SystemLibraries` (NFR-S-1) | `TestNothingOfWeatherRibbonsStartsAProcess` | [`network_test.go`](tests/structural/network_test.go) |
| The page uses no request API and names no web address, the SVG namespace aside (NFR-S-1) | `TestThePageMakesNoRequest` | [`network_test.go`](tests/structural/network_test.go) |
| Each platform's About credits exactly the third-party modules its build links (FR-610) | `TestEveryLinkedModuleIsCredited` | [`credits_test.go`](tests/structural/credits_test.go) |
| No platform credits a module twice | `TestAModuleIsCreditedOncePerPlatform` | [`credits_test.go`](tests/structural/credits_test.go) |
| WeatherRibbon's half of the wire is stated alike in `dto.go` and `frontend/src/wire.ts` | `TestTheWireIsStatedAlikeOnBothSides` | [`wire_test.go`](tests/structural/wire_test.go) |
| The page names each word `app.go` shares with it: opening Add city and the detail panel | `TestThePageNamesEveryWordAppShares` | [`wire_test.go`](tests/structural/wire_test.go) |
| Every method the page's `Bridge` calls is bound on `App`; none of the window's `Control` is | `TestEveryMethodThePageCallsIsBound`, `TestNothingOfTheControlIsBound` | [`page_api_test.go`](page_api_test.go) |
| `wails.json` names its executable as `internal/product` does | `TestEachWailsConfigNamesItsExecutableAsTheProductDoes` | [`names_test.go`](tests/structural/names_test.go) |
| The petrichor line pulses over four seconds between full and half opacity, standing still under reduced motion (FR-413) | `TestThePetrichorLinePulsesAndStandsStillWhenAsked` | [`petrichor_test.go`](tests/structural/petrichor_test.go) |
| The Licence panel is sized for the LICENSE's widest line | `TestTheLicencePanelIsSizedForTheLicencesWidestLine` | [`licence_test.go`](tests/structural/licence_test.go) |
| The settings file of the first release is read whole (NFR-C-1) | `TestA1Point0SettingsFileIsReadWhole` | [`contract_test.go`](internal/infrastructure/store/contract_test.go) |

The structural tests that read the kit's files (the Licence panel's width) read the kit Go builds
against, through `go list -m`; `page_api_test.go` reads the kit's `bridge.ts` as npm installed it,
since that is what the page compiles. `TestTheFixtureIsWhatThisVersionWrites` in the same file as the
contract test holds the fixture to what this version writes until 1.0.0 ships, when it is deleted and
the fixture frozen.

## Layers

- **Domain** (`internal/domain`), pure Go: time arrives as an argument and a zone already resolved.
  - `forecast`: a forecast as steps and their 1, 6 and 12 hour periods; the current step (FR-402,
    FR-403), today from now to local midnight (FR-404), the outlook of three days (FR-406), each day's
    high, low, rain and symbol (FR-405, FR-407 to FR-409) and the next 24 hours for the detail
    (FR-410).
  - `units`: Metric and Imperial with each conversion and rounding (FR-703); every whole degree from
    minus 60 to 60 Celsius written in a system, for the page to measure (FR-103).
  - `place`: a place from the city list, its description, its label capped by the kit's
    `ribbon.Label`, its local time through the kit's `localtime`, coordinates rounded to four places
    (FR-302) and the search's ordering (FR-202).
  - `settings`: the user's choices as one value, every operation answering a new one: the kit's
    `ribbon.Choices` embedded, then the units, the time format, the petrichor countdown and the cities.
  - `backoff`: the wait after failures, from 10 minutes doubling to 2 hours (FR-306).
  - `petrichor`: wet and dry, a city's dry spell, events and the countdown, the draw injected
    (FR-413).
- **Application** (`internal/application`): one `Service` over its ports in `ports.go` (`Store`,
  `Places`, `Forecasts`, `SunTimes`, `Cache`, `Clock`, `Pacer`, `IDs`, `Draw`, `Icons`, the kit's
  `arranger.Monitors` and `arranger.Neighbours`, the kit's `controls.Startup` and `release.Source`).
  It edits cities and searches places (`cities.go`), refreshes forecasts (`refresh.go`), builds the
  snapshot (`snapshot.go`), keeps the petrichor moment (`moments.go`), opens the detail
  (`detail.go`), names each symbol's icon (`symbols.go`), takes the page's measurements
  (`measure.go`) and answers the menus (`menus.go`). It embeds the kit's `arranger.Arranger` and
  `controls.Controls`, so arranging the ribbon and the ribbon's own choices are the service's own
  method set. The arranger asks its `Host` for the ribbon's choices and content read together; the
  service answers through `host.go` (a cell per notice and per city, the prompt when there are none)
  and saves the arranger's changes through its one save path, so they raise the same notice.
- **Infrastructure** (`internal/infrastructure`): `metno` (the forecast and sunrise client),
  `cache` (each city's last forecast and dry spell on disk), `store` (the settings file), `places`
  (the built-in city list) and `pacing` (the wait between requests). Everything else it reaches
  (displays, sign-in, the log, the data folder, the desktop, the shared ribbon folder, the update
  check) is the kit's.
- **UI**: the React front end in `frontend/src` over the kit's page half; the Wails facade in
  package `main`, which embeds the kit's window and maps the service's answers into `dto.go`.
- **Outside the layers**: `internal/product` holds the name, app id, repository, setup program's name,
  window class, donation address, version, User-Agent, author, copyright line, sign-in label and
  credits. The domain and application never read it.
- **Tools**, never shipped: `gencities` (the city list from GeoNames) and `genicons.py` (every
  committed icon).

## Composition root

`main.go` points standard error at the run log before anything can fail, builds the adapters (the
kit's among them), injects them into the service, prepares the platform, starts the tray, starts the
refresher and hands the facade to Wails. What it does for the platform is the kit's `platform`
package: on Linux and macOS `platform.Prepare` hands the desktop the icon and ends the run on SIGTERM
or SIGINT; on Linux importing it sends GTK through X11 and turns off the DMABUF renderer.
WeatherRibbon says only where its tray icon lies (`trayicon_unix.go`; none on Windows, whose tray
reads the executable's). The cell size (`layout`) and panel sizes (`panels`) live there. No service
is held in a global.

The facade Wails binds is `App` in `app.go`, in two halves. The window is the kit's `window.Window`,
embedded, so every exported method of the window is page API. What is WeatherRibbon's own stays in
`app.go` (the cities, units and time format, the detail, the petrichor line, the snapshot) and
`measure.go`, over the `ribbonService` interface, with the refresher behind `fetcher`. WeatherRibbon reaches its window through the kit's
`window.Control`, a named field that is never embedded and so never bound; menu actions the kit does
not know reach WeatherRibbon through the `Act` hook it hands the window. `kit.go` embeds the page and
the LICENSE and adapts the service to the window's port, answering that there is no pull out;
`dto.go` is WeatherRibbon's half of the wire. `icons.go` embeds the Yr icon set's file names as the
`Icons` port: the files are the one home of which symbols have an icon (FR-412). The kit's
`platform.GeneratingBindings` keeps the binding-generation run from writing the log or showing a tray
icon.

```
             +-----------------------------------+
   Wails/UI  | app.go (App) embeds window.Window |
             +--------------+--------------------+
                            | calls
             +--------------v---------------+
             |  application: Service, ports |
             +------+----------------^------+
          depends on|                | implements
             +------v-----+          |
             |   domain   |          |
             +------------+          |
                     +---------------+----------------+
                     |        infrastructure          |
                     +--------------------------------+
```

The page half joins the same way: `frontend/src/api.ts` extends the kit's bridge with WeatherRibbon's
own calls over the kit's guarded call; `App.tsx` is the kit's shell (`useShell`) around the cells,
with the detail as a panel of WeatherRibbon's own; `Ribbon.tsx` draws the cells (`Cell.tsx`,
`Symbol.tsx`) in the kit's `Band`. The front end extends the kit's `tsconfig.json`, uses its eslint
rules and runs its suites under its test set-up.

## The window and the ribbon

How the window opens hidden, places itself, collapses to its tab, scales under the grip, fades with
opacity and shares itself with panels is the kit's (its ARCHITECTURE.md, "One window" and "Place and
drag"). What WeatherRibbon decides:

- **Panels.** Settings opens 900 DIP wide, the others and a city's detail 560 (`panels` in
  `main.go`); each then fits its content's height.
- **Opacity (FR-704).** `app.css` mixes `--window-opacity` into the surface and the tab's accent, so
  only backgrounds fade and the weather stays solid.
- **Size (FR-103, FR-105, FR-106).** The ribbon is sized from its notices and cities (the Add city
  prompt when there are none): cells plus padding along the orientation up to the work area, beyond
  which they scroll; one cell across. The page measures the widest time, weekday, temperature and
  label its font draws (`measure.ts`, `SetMeasured`); the service widens the cell to that width, so
  the cells drawn and the window sized cannot disagree.
- **Refits.** Every change that can alter the cells refits the ribbon where it stands; where the
  length changed, the ribbon is centred along it with its position across kept and that place stored
  (FR-104).
- **No pull out.** WeatherRibbon has none: the window's port answers false and refuses to open one.

## Forecasts

The refresher (`refresher.go`) keeps the forecasts fresh on a goroutine of its own: it waits until
the next city falls due, a wake (a city added, moved or Refresh now) or the end of the run, refreshes,
then has the page redraw. Every forecast request leaves from that one goroutine, so a menu never
waits on the network. A panic there is recovered on that goroutine and logged; the loop then stops.
The cells say how old their forecasts grow from then on (FR-305).

The service decides what is asked (`refresh.go`): a city only once its cached forecast has passed its
`Expires` (FR-303), with `If-Modified-Since` from its `Last-Modified` (FR-304); after a failure not
before its back-off ends (FR-306), unless Refresh now asks (FR-309); never two requests for one city
at once (FR-310); at least a second between any two requests, through the `Pacer` (NFR-S-2). A 403
stops every request for the rest of the run (FR-307).

`metno` sends the User-Agent `product.UserAgent()` names (FR-301), waits at most 10 seconds for an
answer and refuses one over 1 MiB, one that is not JSON or one without a `timeseries` (FR-308),
logging one line per request. Sunrise 3.0 is asked at most once per city per local date; a failure
is asked again on the next opening of the detail (FR-411).

Each successful forecast is written to `forecasts` in the settings folder, one file per city named
after its id hex encoded, with its `Expires`, its `Last-Modified` and the city's dry spell (FR-807,
FR-413). At launch the cache is read, so an offline start shows the last forecasts with their age; an
entry that cannot be read is discarded and fetched again.

## Places and time

The city list, `internal/infrastructure/places/cities.tsv`, is GeoNames' `cities15000` with its
regions and countries, written by `tools/gencities` and embedded (DATA-1): 34,153 places, its head
line recording the download. Zones resolve through the kit's `zones.Resolver`: `time.LoadLocation`
with `time/tzdata` built in (CON-5), each zone loaded once. Each snapshot carries the time to the next
minute and the page takes the next one then.

## The settings file

`settings.json` in the settings folder ([Data locations](#data-locations)) is indented JSON with a
format version (FR-801); derived values are never stored. `store` says what the file holds (its keys
in writing order and its cities, `store/codec.go`); what every ribbon's file does alike is the kit's
`settingsfile`, proved by its own tests: atomic writing, a file that is not JSON kept aside, a file
that could not be read never saved over, a byte order mark passed over and unknown keys written back
(FR-802). A city entry that cannot be read is kept as found and shown as unreadable while the rest
load (FR-804); a city whose place a newer list has dropped keeps its place and reads `Place not
found` (FR-803).

**The file is a contract from the first release (NFR-C-1).** No key is renamed, dropped or given
another meaning. `TestA1Point0SettingsFileIsReadWhole` reads the fixture
`store/testdata/settings-1.0.0.json`, every key set away from its default; once 1.0.0 ships the fixture
is never regenerated.

## Colour

WeatherRibbon states no colour of its own. Every colour on the page is one of the kit's tokens, for
the ten schemes in light and dark (FR-707); the weather symbols are the Yr icons as drawn.

## Menus

Both menus are the kit's native popups. Their items have one home, `internal/application/menus.go`,
built from the kit's `menus` package (the `Item` type and the actions every ribbon offers) plus
WeatherRibbon's own: Add city, the Units submenu and Refresh now, in FR-109's order. Settings offers
every menu choice from the same items (`Service.SettingsChoices`). The tray menu (in `main.go`) and
the Settings choices (in the snapshot) pass through the kit's `Control.Offered` as the right-click
menu does, so a Position item that would leave the ribbon where it stands is greyed in all three
(FR-505). The window carries out every ribbon's actions itself and hands any other to WeatherRibbon's
`actOn`.

## Help, About and Licence

The three panels are the kit's. About shows WeatherRibbon's picture, name and version, author,
copyright line and a credit for MET Norway's data, GeoNames, the Yr icons and every third-party
component this platform's build ships (FR-610), from one table in `internal/product/credits.go`; the
kit is WeatherRibbon's author's own and is not credited. Licence shows the embedded `LICENSE` exactly
as written.

## Data locations

| What | Where |
|---|---|
| Settings | `settings.json`: `%APPDATA%\WeatherRibbon` on Windows, `~/Library/Application Support/WeatherRibbon` on macOS, `$XDG_CONFIG_HOME/WeatherRibbon` else `~/.config/WeatherRibbon` on Linux; kept-aside copies beside it |
| Forecast cache | `forecasts` beside the settings, one JSON file per city |
| Run log | `WeatherRibbon.log` beside the settings |
| Web view data on Windows | `%APPDATA%\WeatherRibbon\WebView2`, so forgetting the settings removes it |
| Ribbons on this desktop | `uk.codecrafter.WeatherRibbon.json` and `.lock` in `%LOCALAPPDATA%\ribbonkit`, `~/Library/Application Support/ribbonkit` or `$XDG_RUNTIME_DIR/ribbonkit` (FR-506) |
| City list, time zone rules, weather icons | built in |
| Start at sign-in | the kit's entry under the name `WeatherRibbon` (FR-710) |

## Errors

Errors are wrapped with `%w` at each boundary, so `errors.Is` finds sentinels such as
`settings.ErrNoSuchCity` beneath.

- **Before the window** only a failure to run the window, to read the built-in city list or to read
  the built-in icon set ends the run. Standard error goes to the log first (the kit's `runlog.Keep`),
  so even a runtime panic is kept. A missing settings folder falls back to the temporary folder;
  unreadable settings, a failed tray icon and a missing executable path are logged and the ribbon
  still opens.
- **On the ribbon:** a kept-aside file and a failed save as notices; a city that cannot be read or
  whose place is gone in words; a forecast that is stale, unavailable or refused in its cell.
- **Beneath the control pressed:** every page call Go can refuse takes a refusal handler and answers
  null rather than rejecting, so a call without one does not compile.

## Quality enforcement

- The structural tests above run with the suite; the kit's run in the kit's own gate.
- `test.ps1` checks formatting, vet and staticcheck, runs the Go suite and the front end's lint, type
  check and tests, holds the domain and application to 100% and every other gated package to its
  measured floor ([TESTING.md](TESTING.md)).

## Design decisions

| Decision | Why | Rejected alternative |
|---|---|---|
| The ribbon's desktop half in ribbonkit, at one tag | TimeRibbon needs the same ribbon; a fix lands once and reaches both | A copy of TimeRibbon's desktop code |
| Forecasts from MET Norway alone (CON-9) | Free under CC BY 4.0, commercial use included, with `Expires` saying when to ask again | A service needing a key or a paid tier |
| The city list built in (DATA-1) | The search works offline and sends nothing about what is typed | An online geocoder |
| One goroutine sends every forecast request | Spacing and one request per city hold by construction; a menu never waits on the network | A request per city as it falls due |
| Which symbols have an icon is read from the embedded files | The icon set is the one home of that answer, so a code and its icon cannot disagree | A list of codes kept beside the icons |
| The page measures the cells | Only the page knows its font and the ratio it is drawn at | Widths written into Go |

See also [TESTING.md](TESTING.md) and [DEVELOPMENT.md](DEVELOPMENT.md).
