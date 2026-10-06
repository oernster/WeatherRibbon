# Decisions and trade-offs

The deliberate choices WeatherRibbon rests on, as the product makes them today: what was chosen,
what was given up and why. The detail and the tests behind each live in
[ARCHITECTURE.md](ARCHITECTURE.md) and [REQUIREMENTS.md](REQUIREMENTS.md);
[TECH_DEBT.md](TECH_DEBT.md) holds what is still open.

## The product as a whole

### Weather where your people are, at a glance

Each cell is a place with a label: its local time, the weather now, today and a short outlook.
WeatherRibbon is not a weather station, a radar viewer or a warning service.

- **Rather than:** contacts or friends' names; radar, maps, air quality, pollen, tides or marine
  forecasts; notifications of any kind.
- **Gains:** nothing personal is kept; the ribbon is read without being opened.
- **Costs:** those jobs need other tools.

### No weather warnings

The README says plainly that WeatherRibbon issues none and that the national weather service is the
place to rely on.

- **Rather than:** passing on alerts from a feed.
- **Gains:** no warning can arrive late or not at all while looking authoritative.
- **Costs:** someone wanting warnings must look elsewhere.

### Forecasts only, never past weather

- **Rather than:** observations or a history of what fell.
- **Gains:** one kind of data, from one source, read one way.
- **Costs:** the ribbon cannot say what the weather was this morning.

### No chance of rain

MET Norway gives a probability of rain for the Nordic countries alone, as measured across twelve
cities.

- **Rather than:** showing it where it exists.
- **Gains:** every cell shows the same things, wherever its city is.
- **Costs:** a Nordic city's probability is left unused.

### Go and Wails, on the shared ribbon

WeatherRibbon is TimeRibbon's stack: Go with Wails drawing a React page. Everything about being a
ribbon on a desktop comes from `ribbonkit`, the module TimeRibbon's desktop code was carved into
before WeatherRibbon was begun. WeatherRibbon holds no copy of it; its own work is the weather, the
cities and the composition.

- **Rather than:** a copy of TimeRibbon's desktop code; a different stack for a second product.
- **Gains:** a desktop fix lands once for both ribbons; two ribbons running together keep off each
  other through the folder they share.
- **Costs:** a change to the ribbon is made and tagged in the kit before WeatherRibbon can ship it;
  both halves of the kit (the Go module and the page's package) are named at one tag, which a test
  holds.

### Requirements before code

Every requirement was agreed before the first line of code, each naming its test; a later change
arrives as a numbered amendment. TimeRibbon's paid-for lessons (sizing from the page, showing the
ribbon only once sized, opacity on backgrounds alone, the pointer read from the desktop) were
written in as requirements rather than rediscovered.

- **Rather than:** building first and describing afterwards.
- **Gains:** a ruled-out idea stays ruled out; a requirement counts as met only once its test has
  failed without the code.
- **Costs:** the specification is work of its own and grows with every amendment.

### Free, with a donate button

- **Rather than:** a paid tier, a licence key or a feature held back.
- **Gains:** nothing stands between a user and the whole product.
- **Costs:** the work is paid for only by goodwill.

## The weather source

### MET Norway alone

Every forecast and every sunrise and sunset comes from MET Norway, under its terms.

- **Rather than:** a service needing a key or a paid tier; several sources blended.
- **Gains:** free use, commercial use included, under CC BY 4.0; each answer says when to ask again.
- **Costs:** WeatherRibbon is only as good as one forecaster; its terms are checked again before
  each release; its data and icons are credited in About.

### Ask only once a forecast has expired; ask only for what changed

A city is asked for again only a short grace after MET Norway's own expiry for its forecast has
passed; an answer that leaves the expiry behind waits that grace again. Each request carries the date
of the copy held, so an unchanged forecast comes back as "not modified" and is not sent again.
Refresh now asks at once only for what has expired or is backing off.

- **Rather than:** a fixed refresh timer; asking at the expiry itself, which, while MET Norway's new
  forecast was not yet ready, asked once a second until it was; Refresh now asking for everything.
- **Gains:** WeatherRibbon asks no more often than MET Norway says is useful, once per expiry.
- **Costs:** a new forecast arrives a few seconds later than it could; no press can fetch a forecast
  still within its expiry.

### One goroutine sends every request, a second apart

Forecasts are fetched on one goroutine of its own, one city at a time, never two for one city, at
least a second between any two requests. The sunrise and sunset requests wait their turn in the
same line.

- **Rather than:** a request per city as each falls due.
- **Gains:** the spacing and the one request per city hold by construction; a menu never waits on
  the network.
- **Costs:** many cities due together are fetched one after another; opening a city's detail can
  wait behind a refresh under way.

### Backing off after a failure

A failed request keeps the forecast held and tries again later, the wait doubling after each
further failure up to two hours; a success resets it.

- **Rather than:** retrying at once; giving up.
- **Gains:** an outage on either side costs MET Norway little.
- **Costs:** after a long outage a city may wait for its next try unless Refresh now asks.

### A refusal stops every request until the next launch

If MET Norway answers 403, nothing more is asked of it in this run and the ribbon says so.

- **Rather than:** retrying a refusal; backing off from it as from a failure.
- **Gains:** WeatherRibbon never keeps knocking on a door MET Norway has closed.
- **Costs:** one refusal silences every forecast until WeatherRibbon is started again.

### An answer it cannot use is a failure

An answer too large, not JSON or without a forecast in it is refused and treated as a failed
request; a size the server states is never trusted.

- **Rather than:** reading whatever arrives.
- **Gains:** a broken or hostile answer cannot fill memory or replace a good forecast.
- **Costs:** none recorded.

### A crashed refresher is a notice, not a loop

A fault that ends the refreshing is caught, logged and shown on the ribbon with its reason until
the next launch; the cells go on saying how old their forecasts grow.

- **Rather than:** restarting a refresh that would fail the same way at once; failing silently.
- **Gains:** a stopped ribbon says it has stopped and why.
- **Costs:** forecasts stay stopped until WeatherRibbon is started again.

### The forecasts survive a restart

Each city's last forecast is kept on disk in a file of its own, written whole, with its expiry and
its dry spell. An entry that cannot be read is removed and fetched again; a file the cache did not
write is left alone.

- **Rather than:** forecasts held only in memory.
- **Gains:** an offline start shows the last forecasts with their age; a restart asks for nothing
  not yet expired.
- **Costs:** a folder of files beside the settings; an unreadable entry is lost rather than mended.

### A stale forecast says its age; one that has run out shows nothing

A forecast still held after failed requests is shown, saying how old it is once that passes two
hours. A forecast that no longer covers now shows no value as current; the cell says the forecast
is unavailable instead.

- **Rather than:** hiding a stale forecast; showing the last value as though it were now.
- **Gains:** a cell is never wrong without saying so.
- **Costs:** an offline ribbon left long enough ends with every cell unavailable.

## Privacy and the network

### Coordinates and a User-Agent, nothing else

A forecast request sends the city's latitude and longitude rounded to four decimal places (the most
MET Norway's terms allow) with a User-Agent naming WeatherRibbon, its version and its repository.
No email address is sent. The update check, the kit's, sends nothing about the user.

- **Rather than:** exact coordinates; an email address in the User-Agent.
- **Gains:** MET Norway's terms are met with nothing personal sent.
- **Costs:** MET Norway learns which places are watched.

### Only the forecast client reaches the network, held by a test

Structural tests fail for any other Go code that imports a network package or starts a program;
they fail too for any request from the page. The donation page and downloads go to the browser.

- **Rather than:** a promise that the network is used sparingly.
- **Gains:** "nothing else touches the network" is a test result.
- **Costs:** any new outward feature has to change the test that forbids it.

## Places and time

### The city list is built in

The places offered are GeoNames' list of cities of more than 15,000 people, with their regions and
countries, carried inside the executable. A city is a place from that list with a label of the
user's own; coordinates are never typed.

- **Rather than:** an online geocoder; coordinates by hand.
- **Gains:** the search works offline and sends nothing about what is typed; every place offered
  has coordinates and a zone.
- **Costs:** a few megabytes in the executable; a smaller town is reached only through a nearby city
  and a label; a newer list comes only with a new release.

### Search matches the start of a word

What is typed matches the start of a word of a place's name, region or country. Places whose name
begins with it come first; within each group, larger places first. A place whose region GeoNames
does not name shows its name and country alone.

- **Rather than:** matching anywhere in a word; alphabetical order.
- **Gains:** `london` lists London, England before London, Ontario; the eight Springfields each show
  their state.
- **Costs:** a fragment inside a word finds nothing.

### A place gone from the list is kept, never replaced

A city whose place a newer list has dropped keeps its place on the ribbon and reads that its place
was not found, with Change place and Remove offered; one city entry that cannot be read is kept and
written back as found while the others work.

- **Rather than:** substituting the nearest match; dropping the city.
- **Gains:** a city is never silently swapped for another; one bad entry costs only itself.
- **Costs:** such a city shows no weather until changed or removed.

### The same place may be added twice

- **Rather than:** refusing a duplicate.
- **Gains:** one city can stand for two people under two labels.
- **Costs:** two cells show the same weather.

### Cities run east from Greenwich, as TimeRibbon's clocks do

London first, then places ahead of UTC by how far ahead, then those behind it; ties keep the order
they were added. Each cell's local time, its zone mark and the 12-hour or 24-hour choice are the
kit's, 24-hour by default, with no seconds.

- **Rather than:** an order by hand; an order of its own.
- **Gains:** the two ribbons read round the world the same way.
- **Costs:** cities cannot be arranged by hand.

## Today and the outlook

### Today means now to local midnight

Today's high, low and rain cover the forecast from the current hour to the end of the city's own
date. The README says so.

- **Rather than:** the whole calendar day, past hours included.
- **Gains:** the figures are forecasts still to come.
- **Costs:** late in the day they describe only the hours left.

### Three days ahead

Each cell shows the three days after today, each with its weekday, a symbol, its high and its low.

- **Rather than:** a longer outlook; none.
- **Gains:** a cell stays small enough to sit beside the others.
- **Costs:** a week ahead needs another tool.

### A period belongs to the day of its midpoint

A forecast period counts for the local date its midpoint falls on in the city's zone; a day's rain
adds the hourly periods where there are any and the six-hour ones after, never both for one hour.
A day's symbol is that of the six-hour period nearest local noon.

- **Rather than:** a period counting for the day it starts; adding every period that touches a day.
- **Gains:** a period split by midnight is counted once, in the right day, whatever the zone's
  offset; rain is never counted twice.
- **Costs:** a day's symbol says nothing of a stormy evening after a fine noon.

### One choice of units for every city

Metric is the default: Celsius, millimetres and kilometres an hour. Imperial is Fahrenheit, inches
and miles an hour. The choice is made once; temperatures are shown to whole degrees.

- **Rather than:** units per city.
- **Gains:** every cell reads alike.
- **Costs:** a Fahrenheit city beside a Celsius one is not possible.

### Every number says what it measures

A temperature carries its scale, a high and a low their letters and rain its measure; the detail
panel names wind's measure in its heading. The outlook puts each day's high above its low.

- **Rather than:** bare numbers, which the first build drew.
- **Gains:** no figure can be misread as another measure or another scale.
- **Costs:** more text in each cell, which the stacked high and low keep narrow.

### No dates in the cells

The outlook names each day by its weekday alone; no cell draws a date.

- **Rather than:** a weekday and date for each day.
- **Gains:** narrower cells.
- **Costs:** nothing in a cell says which date a day is.

### Hourly detail on a click

A click on a cell opens the next 24 hours for that city: the temperature, symbol, rain and wind each
hour, with its sunrise and sunset. The sun times are asked once per city per local date; a failure
is asked again at the next opening.

- **Rather than:** hours drawn in the cell; sun times fetched for every city as forecasts are.
- **Gains:** the cell stays small; the sun times cost a request only when looked at.
- **Costs:** a second kind of request to MET Norway; the detail replaces the ribbon while open.

### The icon set says which symbols have icons

Which symbols have an icon is read from the built-in Yr icon files themselves. A symbol without one
shows its words and is logged once. One code MET Norway may spell two ways finds its icon either
way and is said in its correct spelling.

- **Rather than:** a list of codes kept beside the icons.
- **Gains:** a code and its icon cannot disagree; an unknown symbol still says what it is.
- **Costs:** a new symbol shows as words until its icon is added.

## The petrichor moment

### Rain after a dry spell, rarely

When rain arrives in a city that has been found dry for at least three days, that is an event. One
countdown for every city, drawn between ten and fifteen events, decides which event shows a quiet
line in that city's cell inviting the word to be looked up. Dryness is judged from the forecast's
current period; a stale forecast, one that has run out or a period giving no rain figure is
neither wet nor dry. The draw is handed to the domain, never taken there.

- **Rather than:** showing it at every rain; a chance at each event, which could go unseen for good.
- **Gains:** the moment stays rare yet is bounded; a forecast that cannot be judged never counts.
- **Costs:** it is judged from forecasts, not from rain that fell; it may be weeks between moments.

### A gentle line that goes with the rain

The line pulses gently between full and half opacity, standing still where the system asks for
reduced motion. It goes when the city is found dry or when it is clicked, which opens no detail.

- **Rather than:** a notification; a line that stays until dismissed.
- **Gains:** it belongs to the weather it describes and never interrupts.
- **Costs:** it is not restored after a restart.

### The spell and the countdown survive a restart

Each city's dry spell is kept with its cached forecast; the countdown is kept in the settings file.
Time while WeatherRibbon is closed does not break a spell; only a wet finding does.

- **Rather than:** starting afresh at every launch.
- **Gains:** a ribbon used only by day still sees its dry spells end.
- **Costs:** rain that came and went while it was closed is never counted.

## The settings file

### One readable file, a contract from the first release

Every choice and city lives in one indented JSON file; derived values are never stored. No key the
first release writes is renamed, dropped or given another meaning; a test reads a frozen file of
the first release. Writing whole, keeping a damaged file aside and writing unknown keys back are the
kit's, which every ribbon's file shares.

- **Rather than:** a database; reshaping the file as the product grows.
- **Gains:** a person can read and repair it; an upgrade never loses anybody's cities.
- **Costs:** a key named badly once is named so for good.

## The ribbon on the desktop

### The ribbon's behaviour is the kit's

Placing, dragging, snapping, the thin tab of an unpinned ribbon, the corner grip, opacity on the
backgrounds alone, the tray and both menus, one copy at a time, start at sign-in, the update check,
Linux through X11 and the Dock icon kept on macOS all come from ribbonkit, as TimeRibbon's
DECISIONS-TRADEOFFS.md describes them. That behaviour is verified by the kit's own tests, never by a
second copy here.

- **Rather than:** writing it again; testing it twice.
- **Gains:** WeatherRibbon behaves as TimeRibbon does and inherits every fix.
- **Costs:** WeatherRibbon cannot vary what the kit decides without a change in the kit.

### Never on another ribbon

Beside TimeRibbon (or any other ribbon on the kit) WeatherRibbon takes the nearest free place along
its edge, else the opposite edge; a Position choice that would leave it where it stands is greyed.

- **Rather than:** two ribbons stacking on one edge's centre.
- **Gains:** the two products share a desktop without being arranged by hand.
- **Costs:** the folder they share is one more thing the Flatpak is granted.

### The page measures the cells

The page measures the widest text a cell can show in the font it draws with: every minute of the
day, every weekday, every whole degree of the span of weather the cell is sized for written in the
chosen units; the longest label. Go sizes the cells and the window from that width.

- **Rather than:** widths written into Go; a span of degrees fixed in whichever unit is chosen,
  which in Fahrenheit stopped short of a warm day.
- **Gains:** the cells drawn and the window sized cannot disagree at any font or scaling.
- **Costs:** the page measures again on every change of units, format or label.

### Size by the grip alone

- **Rather than:** TimeRibbon's Large and Small sizes beside the grip.
- **Gains:** one way to size the cells.
- **Costs:** none recorded.

### No pull out

TimeRibbon's sun map is not carried over; WeatherRibbon answers the kit that it has no pull out.

- **Rather than:** a map or radar sliding out beside the ribbon.
- **Gains:** a single lane of cells; no imagery to carry or fetch.
- **Costs:** none recorded.

### No colour of its own

Every colour on the page is one of the kit's tokens for its ten schemes; WeatherRibbon's stylesheets
are held by a test to the kit's checked text and background colours, so the kit's contrast test
covers every cell. The weather symbols are the Yr icons as drawn.

- **Rather than:** colours written for the weather cells.
- **Gains:** one palette across both ribbons; an unreadable cell fails the kit's suite.
- **Costs:** a colour the cells need is added in the kit first; how each icon reads on a dark cell
  is checked by hand.

## Building and installing

### Three platforms from the first release

A setup program on Windows, a DMG on macOS and a Flatpak on Linux, each built by the kit's tools and
TimeRibbon's scripts with the names changed.

- **Rather than:** Windows first and the others later.
- **Gains:** a friend on any desktop can run it.
- **Costs:** three builds to check by hand before each release.

### Per user, with the kit's setup program on Windows

Everything is written under the user's own folders; nothing asks for administrator rights. The
setup program is the kit's, carrying WeatherRibbon's name and pictures. The settings and the kept
forecasts are removed only when asked.

- **Rather than:** a machine-wide install; a generic installer.
- **Gains:** nothing needs an administrator; uninstalling keeps a user's cities unless told
  otherwise.
- **Costs:** each account installs separately.

### macOS: Apple Silicon only, signed and notarised

- **Rather than:** a universal build; shipping unsigned.
- **Gains:** Gatekeeper lets it open.
- **Costs:** Intel Macs are not served; signing needs a Terminal at the Mac itself.

### A Flatpak with a narrow sandbox

The sandbox holds X11, the GPU, the tray, the single-instance name, the autostart folder, the folder
the ribbons share and the network.

- **Rather than:** a native package; wider access.
- **Gains:** the application holds only what it uses.
- **Costs:** Linux users need Flatpak; the network is granted to the whole application for MET
  Norway and the update check.

## Engineering

### Layers with one place where they meet

Domain, application, infrastructure and interface, each depending only inward, the kit's packages
counting as layers too; only the composition root sees both sides. The domain reads no clock and
holds no chance: time, zones and the petrichor draw arrive as arguments. Structural tests hold every
boundary.

- **Rather than:** convention; a dependency injection framework.
- **Gains:** forecast, units, petrichor and refresh rules are tested with no disk, network, clock or
  screen.
- **Costs:** more packages and explicit wiring.

### Complete coverage where it means something

The domain and application are held to 100 percent; every other gated package to the coverage it
measured.

- **Rather than:** one figure over everything; aspirational floors.
- **Gains:** a shortfall in the pure layers is a decision nobody made; every other one is named.
- **Costs:** the facade sits lower and the window relies on checks by hand.

### Tests with real parts; never the network

No mocking library; hand-written stand-ins against real interfaces. MET Norway is answered from
recorded answers by a stand-in client; no test reaches the network or the user's own settings. Every
guard is proved by planting a violation. The Windows build runs the whole gate first with no switch
to skip it.

- **Rather than:** mocks; tests against the live service; an optional test step.
- **Gains:** a passing test means the real behaviour holds; the tests never spend MET Norway's
  goodwill.
- **Costs:** stand-ins and recorded answers are kept by hand.

### The wire written twice and compared

WeatherRibbon's shapes crossing between Go and the page are stated in both languages and compared by
test; every method the page calls is checked to be bound and nothing of the window's own control is.

- **Rather than:** Wails' generated bindings.
- **Gains:** a contract both sides are checked against.
- **Costs:** every wire change is made twice.

### Small files; every name and the version have one home

No source file passes 400 lines. The product's name, repository, User-Agent, donation address and
credits are stated once; the version lives in one file.

- **Rather than:** letting files grow; copies where needed.
- **Gains:** files split at real seams; a rename is made once.
- **Costs:** many small files.
