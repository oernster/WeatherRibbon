# <img width="128" height="128" alt="app-icon" src="https://github.com/user-attachments/assets/be9428f9-c11c-4641-a401-5afc64d0476a" /> WeatherRibbon

A small ribbon of weather for the cities you choose, horizontal or vertical, for Windows, macOS and
Linux.

> **Commercial licences available.** WeatherRibbon is free and open source under the GNU General Public License, version 3 (GPL-3.0). If those terms do not suit what you are building, such as a closed-source product, a commercial licence can be bought from me separately. It covers my own code; third-party libraries keep their own licences. See [commercial licensing](https://ernster.dev/commercial-licensing.html).

What is the weather where your friends and family are, today and over the next few days?
WeatherRibbon shows one cell per city you choose, each with its local time, the weather now, today's
high, low and rain and the three days after, in a small frameless ribbon you can put anywhere on any
monitor.

## Who it is for

- Anyone with people in other cities who wants their weather and their time in view.
- Anyone on Windows 10 or 11, macOS 12 or later on Apple Silicon or a Linux desktop that runs
  Flatpaks.

## Who it is not for

- **Anyone needing weather warnings.** WeatherRibbon issues none; rely on your national weather
  service for those.
- Anyone wanting radar, maps, air quality, pollen, tides or marine forecasts.
- Anyone wanting the chance of rain: MET Norway gives it for the Nordic countries alone, so cells
  would disagree in what they show.
- Anyone wanting past weather or observations: it shows forecasts.
- Anyone wanting to track people: a cell is a place, never a person.
- Anyone on an Intel Mac.

## What it does

- **The weather now** in each city: the temperature and a weather symbol, from MET Norway's
  forecast for the hour.
- **Today and three days ahead.** Today's high, low and rain cover the hours from now to midnight in
  that city, so late in the day they describe the hours left, not the whole day. Each of the next
  three days shows its weekday, a symbol, its high and its low.
- **The hours in detail.** A click on a cell opens the next 24 hours for that city: the temperature,
  symbol, rain and wind each hour, with the city's sunrise and sunset.
- **Each city's own time,** with its zone's abbreviation (BST) or its offset where the zone has none.
- **Runs east from Greenwich,** as TimeRibbon does: London, Berlin, Tokyo, then New York.
- **Finds cities offline** among the 34,153 places of GeoNames' list of cities over 15,000 people,
  built in, so a search never touches the network. Typing matches the start of a word of the name,
  region or country; `london` lists London, England before London, Ontario. A label can be anything up
  to 32 characters.
- **Metric or Imperial.** Metric is degrees Celsius, millimetres and kilometres an hour; Imperial is
  degrees Fahrenheit, inches and miles an hour.
- **Asks no more often than it should.** A forecast is asked for again only once MET Norway says it
  has expired, ten seconds after, so MET Norway has its new forecast ready. The request carries the date of the copy already held, so an unchanged forecast is not
  sent twice. Requests leave at least a second apart. Refresh now asks at once for every city whose
  forecast has expired or that is waiting to try again after a failure.
- **Keeps working offline.** The last forecasts are kept on disk; with no network the cells show them,
  saying how old they are once that passes two hours. A failed request is tried again after 10
  minutes, waiting twice as long after each further failure, up to two hours.
- **Stays out of the way.** No title bar, border or taskbar button; its icon sits in the notification
  area, menu bar or tray. The icon's menu and the ribbon's right-click menu offer Add city, Settings,
  Units, Colour, Orientation, Position, Always on top, Pin ribbon, Refresh now, Help, Hide and Exit.
  Hiding the ribbon leaves WeatherRibbon running.
- **Unpinned, it waits as a thin tab** against the edge it stands on, opening when the pointer rests
  on it.
- **Goes where you put it** on any monitor and opens there next time; Position centres it on an edge,
  greying an edge that would leave it where it stands. Beside TimeRibbon it never lands on it.
- **Fits its cells, then scrolls,** each cell as wide as the widest time, weekday, temperature and
  label its font can draw.
- **Choices that apply at once:** ten colour schemes, light, dark or the system's theme, 12-hour or
  24-hour time. The Opacity slider fades the ribbon's background from 20 to 100 percent while the
  weather stays solid. The grip in the ribbon's corner resizes the cells from 75 to 200 percent.
- **Keeps your cities in one readable file,** `settings.json`, written whole or not at all. A damaged
  file is kept aside and never overwritten; one city that cannot be read leaves the others working.
- **Starts when you sign in,** when asked.
- **Tells you when a new release is out.** Help's Check for updates asks at any time.
- Now and then, after a long dry spell, a city may have a question for you.

## What it does not do

- **It issues no weather warnings** and is no substitute for your national weather service.
- **Its network requests are these and no others:** the forecasts and the sunrise and sunset times
  from MET Norway (`api.met.no`), sending each city's latitude and longitude rounded to four decimal
  places and a User-Agent naming WeatherRibbon, its version and this repository; and the update check
  to GitHub, sending nothing about you. If MET Norway refuses a request, WeatherRibbon asks it nothing
  more until it is started again. The donation page and downloads are handed to your browser.
- **It never installs an update itself.**
- **It never shows a forecast as current once the forecast has run out;** the cell says so instead.
- **On Linux it draws through X11,** XWayland on a Wayland desktop, as TimeRibbon does.

## Built with

| Part | Choice |
|---|---|
| Backend | Go |
| Desktop shell | Wails v2: WebView2 on Windows, WKWebView on macOS, WebKitGTK on Linux |
| Front end | React and TypeScript, built with Vite |
| Forecasts | [MET Norway](https://api.met.no/)'s Locationforecast 2.0 and Sunrise 3.0, under CC BY 4.0 |
| City list | [GeoNames](https://www.geonames.org/) `cities15000`, under CC BY 4.0, built in |
| Weather symbols | the Yr weather icons from MET Norway, under MIT, built in |
| Time zone rules | the tz database built in through Go's `time/tzdata` |
| The ribbon itself | [ribbonkit](https://github.com/oernster/ribbonkit), the placing, dragging, tab, tray and scaling every ribbon shares, at one tag |

## Getting it

Each file below is attached to the [latest release](https://github.com/oernster/WeatherRibbon/releases/latest);
[DEVELOPMENT.md](DEVELOPMENT.md) builds them from source.

### Windows

Run `WeatherRibbonSetup.exe`. It installs for your account alone under
`%LOCALAPPDATA%\Programs\WeatherRibbon`, never asks for administrator rights and offers a Start Menu
entry, a Desktop shortcut and Start with Windows. Remove it from the Apps list; your cities and the
kept forecasts stay unless you tick **Also forget my settings**.

### macOS

Open `WeatherRibbon.dmg` and drag WeatherRibbon to Applications. To remove it, turn off Open at
Login, then move it to the Bin.

### Linux

```bash
flatpak install --user weatherribbon.flatpak
```

```bash
flatpak run uk.codecrafter.WeatherRibbon
```

It runs on the GNOME 50 runtime from Flathub. Exit it before installing a newer release, since a
copy left running keeps the old one. To remove it, turn off Start when I sign in, Exit, then:

```bash
flatpak uninstall --user uk.codecrafter.WeatherRibbon
```

### Your settings

| Platform | Folder holding `settings.json`, the log and the `forecasts` folder |
|---|---|
| Windows | `%APPDATA%\WeatherRibbon` |
| macOS | `~/Library/Application Support/WeatherRibbon` |
| Linux (the Flatpak) | `~/.var/app/uk.codecrafter.WeatherRibbon/config/WeatherRibbon` |
| Linux (from source) | `$XDG_CONFIG_HOME/WeatherRibbon`, else `~/.config/WeatherRibbon` |

## Testing

```powershell
./test.ps1
```

The gate checks formatting, vet and staticcheck, runs every Go test plus the front end's lint, type
check and tests, then holds the domain and application to 100 percent coverage and other packages to
measured floors. [TESTING.md](TESTING.md) has the figures.

## Building

| Platform | Command | Output |
|---|---|---|
| Windows | `./build.ps1` | `build/bin/WeatherRibbon.exe` and `dist-installer/WeatherRibbonSetup.exe` |
| macOS | `bash builddmg.sh` | `WeatherRibbon.dmg` |
| Linux | `bash build_flatpak.sh` | `weatherribbon.flatpak` and an install for your account |

`build.ps1` runs the gate first. [DEVELOPMENT.md](DEVELOPMENT.md) sets up each machine;
[ARCHITECTURE.md](ARCHITECTURE.md) explains the layering; [DECISIONS-TRADEOFFS.md](DECISIONS-TRADEOFFS.md)
gives the decisions with what each costs; [TECH_DEBT.md](TECH_DEBT.md) lists what is still open, what
is deliberately left and what only looks like debt.

## Supporting WeatherRibbon

WeatherRibbon is free and stays free: no paid tier, no licence key, no feature held back behind a
donation. If it earns its place on your screen, a donation is welcome. The same button sits at the
foot of Settings.

<a href="https://www.paypal.com/ncp/payment/88LQG589TJEM6"><img src="assets/donate.png" alt="Donate to WeatherRibbon" width="120"></a>

## Licence

GNU General Public License, version 3: see [LICENSE](LICENSE), also shown under Help, then Licence.
The forecasts and the city list are CC BY 4.0 and the weather symbols MIT, credited under Help, then
About.

Commercial licences are available: see [commercial licensing](https://ernster.dev/commercial-licensing.html).
