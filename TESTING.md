# Testing

What is tested, what is not and why the line falls where it does. Every figure was measured by the
commands in [Running it](#running-it) when it was written; every shortfall is named with its reason.
The ribbon itself is [ribbonkit](https://github.com/oernster/ribbonkit), tested by its own gate and
its own TESTING.md; this document covers what WeatherRibbon adds.

## Before the first run on Windows

Some anti-virus programs quarantine a freshly built test binary in Go's scratch folder, which stops
the suite. If a run fails with access denied or a missing file on a test binary, allow the folder
`go env GOTMPDIR` names (the system temporary folder where it prints nothing). Install the front end's
packages once, which fetches the kit's page half from GitHub at the tag `package.json` names. Build
the page once too, since the application embeds `frontend/dist`, which git does not hold:

```powershell
npm --prefix frontend install
```

```powershell
npm --prefix frontend run build
```

Text files check out with LF endings (`.gitattributes`), since gofmt refuses CRLF.

## The standard

**A floor is a measurement, never an aspiration.** Each floor in `test.ps1` is its package's measured
figure with the fraction dropped, so it fails once cover is lost.

**A gap is named or it is closed.** An unexplained shortfall cannot be told from an oversight.

## What the numbers are

### Go

| Package | Coverage | Floor |
|---|---|---|
| `internal/domain/backoff`, `forecast`, `petrichor`, `place`, `settings`, `units` | 100% | 100% |
| `internal/application` | 100% | 100% |
| `internal/infrastructure/cache`, `metno`, `pacing`, `places`, `store` | 100% | 100% |
| `internal/product` | 100% | 100% |
| `tools/gencities` | 90.9% | 90% |
| the root package (the Wails facade) | 68.4% | 68% |
| `installer` | 0% | not gated |
| `tools/versioninfo`, `payload`, `identity`, `linuxicons` | no tests | not gated |

`installer` is the setup program's composition root; its one test reads the pictures it carries,
which runs no statement. The four tools are mains that hand `internal/product` to the kit's
`delivery` package, where their work and its tests live.

Every figure is the Windows build's, which `test.ps1` measures. That build compiles 177 Go test
functions, counted from the test files `go list` selects. Twenty-two are the structural tests in
`tests/structural`, which read the source and are the same on every platform; two more in
`page_api_test.go` hold the page's calls to what is bound. [ARCHITECTURE.md](ARCHITECTURE.md) lists
each against its rule. `TestA1Point0SettingsFileIsReadWhole` in `store` holds the settings file's
promise (NFR-C-1); `TestThePetrichorLinePulsesAndStandsStillWhenAsked` holds FR-413's pulse in Go
because jsdom draws no animation and Vitest hands a test no stylesheet's text. The macOS and Linux
builds compile 175. Two are Windows only: the setup program's test, since setup is built for Windows
alone; `cache/held_windows_test.go`, a cache entry another program holds open being left out and
fetched again, since only Windows refuses to read a file held open
([On macOS and Linux](#on-macos-and-linux)).

### The front end

30 tests in 8 files under Vitest with jsdom, run from `frontend`: the cells in the kit's band, each
city's time, weather now, today and outlook, a symbol shown as words, a problem said in place of the
weather, a stale forecast's age, the empty ribbon and opening the detail (`ribbon.test.tsx`); the
detail's hours, its sun times and closing it (`detail.test.tsx`); Settings, its search, removing a
city, the menus' choices with a greyed Position choice, the time format and the donate button
(`settings.test.tsx`); measuring a cell's widest text (`measure.test.ts`, FR-103); every number written with its unit,
a high and a low with their letters (`units.test.ts`, FR-703); the petrichor line
(`petrichor.test.tsx`, FR-413); every shipped weather icon named by its code in words
(`a11y.test.tsx`, NFR-U-2); every timer the page schedules, the kit's included, against a reasoned
allow-list through the kit's `describePageTimers` (`timers.test.ts`, NFR-P-3). No coverage provider is installed, so no figure is claimed.

## How each layer is tested

| Layer | Kind of test | Touches |
|---|---|---|
| `internal/domain` | pure unit over fixed instants, zones loaded with the tz database embedded and an injected draw | nothing |
| `internal/application` | unit over hand-written fakes of its ports, the clock and the pacer among them | nothing |
| `internal/infrastructure` | integration over temporary folders; the forecast client over a stand-in HTTP client answering recorded MET Norway answers from `testdata` | the filesystem |
| the root package | unit over a scripted service and a stand-in window; the refresher over recording stand-ins at a fixed instant; WeatherRibbon's place among the ribbons over a temporary folder | reads `frontend/src/api.ts` and the kit's `bridge.ts` as npm installed it; the filesystem |
| `tests/structural` | source and AST scans, a `go list` per platform, the kit's files read through `go list -m` | reads files |
| `tools/gencities` | unit over small GeoNames-shaped files in a temporary folder | the filesystem |
| the front end | component tests under jsdom over `fakeBridge.ts`, which records every call and builds on the kit's stand-in (`@oernster/ribbonkit/testing`) | nothing |

No Go test uses a mocking library. **No test writes to the user's own settings, cache, sign-in
entry or Apps list** and **no test reaches the network**: MET Norway is answered by the stand-in client; the update
check is the kit's and tested there.

## What is not tested and why

- **The root package (68.4%).** WeatherRibbon's own half of the facade is tested over a scripted
  service and a stand-in window: every change fits the ribbon and answers the service's error, a new
  place is asked for at once, the snapshot carries every cell and the window's reading, the search,
  the detail and the petrichor line reach the service, WeatherRibbon's menu actions (a failed one
  reported, an unknown one changing nothing), the product handed to the window, the
  measurements, the refresher's waiting, Refresh now and its recovery from a panic, the built-in icon
  set, WeatherRibbon's place in the folder every ribbon shares, the adapter
  reading the ribbon's choices out of the settings and having no pull out; every method the page's `Bridge` calls is bound,
  with nothing of the `Control`. Not reached: the composition root (`main`, `keepLog`, `settingsDir`,
  `run`).
- **`installer` (0%).** Only the composition root; the setup window and its policy are tested in the
  kit.
- **`tools/gencities` (90.9%):** its `main` handing `run` its real arguments. Its reading, joining and
  writing are tested. The other tools' work is the kit's and tested there.

What the window, the tray, focus, paint, the install and a real MET Norway answer do on a real
desktop is checked by hand in a real build of each platform.

## On macOS and Linux

The root package compiles its own halves there and links the kit's cgo half. Check it on a machine
of that platform, set up as [DEVELOPMENT.md](DEVELOPMENT.md) says, with the page built. The kit's own
macOS and Linux checks are in its TESTING.md.

| What | macOS | Linux |
|---|---|---|
| Tags | `desktop,production` | `desktop,production,webkit2_41` |
| Go test functions | 175 | 175 |

With the platform's tags in `TAGS`, run each and read its exit code:

```bash
test -z "$(gofmt -l . | grep -v node_modules)"
```

```bash
go vet -tags "$TAGS" $(go list ./... | grep -v node_modules)
```

```bash
go run honnef.co/go/tools/cmd/staticcheck@v0.8.1 -tags "$TAGS" $(go list ./... | grep -v node_modules)
```

```bash
go test -count=1 -tags "$TAGS" $(go list ./... | grep -v node_modules)
```

staticcheck is the version `test.ps1` pins. To run from source with a scratch settings folder (`HOME`
on macOS, `XDG_CONFIG_HOME` on Linux):

```bash
go run -tags "$TAGS" .
```

## Running it

The whole gate:

```powershell
./test.ps1
```

It checks formatting, vet and staticcheck, runs every Go test, runs the front end's `lint`,
`typecheck` and `test`, holds the domain and application to 100% and every other gated package to its
floor. A front end without its packages stops the gate. Read the exit code: `0` means every check
passed and every floor held.

Another floor for the domain and application, for a deliberate check:

```powershell
./test.ps1 -Floor 95
```

The front end alone, from `frontend`:

```powershell
npx vitest run
```

One package's coverage in detail; the profile's total can differ slightly from the `-cover` figure
the floors hold:

```powershell
go test -coverprofile=cover.out ./internal/infrastructure/cache
```

```powershell
go tool cover -func=cover.out
```

## Keeping this honest

**Prove a new guard bites:** plant the violation, read the exit code, restore the file in a
`finally`; confirm first that the clean tree passes. **Re-measure before quoting:** a figure copied
forward describes a repository that no longer exists. **Read the exit code, never the last line.**

See also [DEVELOPMENT.md](DEVELOPMENT.md) and [ARCHITECTURE.md](ARCHITECTURE.md).
