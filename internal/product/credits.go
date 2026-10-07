package product

import (
	"runtime"
	"slices"
)

// Author is who wrote WeatherRibbon, as About names them (FR-603).
const Author = "Oliver Ernster"

// Copyright is the copyright line About shows beneath the author (FR-603).
const Copyright = "© " + Author

// The platforms a build ships for, as Go names them.
const (
	Windows = "windows"
	Linux   = "linux"
	MacOS   = "darwin"
)

// Platforms is every platform WeatherRibbon ships for.
var Platforms = []string{Windows, MacOS, Linux}

// Credit is one component the application ships or shows: the Go module it comes from where it is
// one, the name it is credited under, its licence, what it does here and the platforms whose build
// ships it, none meaning every one.
type Credit struct {
	Module    string
	Name      string
	Licence   string
	Role      string
	Platforms []string
}

// Credits is every component this platform's build ships (FR-610).
var Credits = CreditsFor(runtime.GOOS)

// CreditsFor answers every component the build for goos ships.
func CreditsFor(goos string) []Credit {
	var out []Credit
	for _, credit := range allCredits {
		if len(credit.Platforms) == 0 || slices.Contains(credit.Platforms, goos) {
			out = append(out, credit)
		}
	}
	return out
}

// ccBy4 is the licence the forecasts and the city list are given under, with the link FR-610 asks
// for.
const ccBy4 = "CC BY 4.0, https://creativecommons.org/licenses/by/4.0/"

// The data WeatherRibbon shows, credited as their terms ask (FR-610, ASM-4).
var (
	metNorway = Credit{"", "MET Norway (the Norwegian Meteorological Institute)", ccBy4, "the forecasts with the sunrise and sunset times", nil}
	geoNames  = Credit{"", "GeoNames, cities15000", ccBy4, "the city list", nil}
	yrIcons   = Credit{"", "Yr weather icons, MET Norway", "MIT", "the weather symbols", nil}
)

// windowsOnly names the platform of a component only the Windows build ships.
var windowsOnly = []string{Windows}

// allCredits is every component any build ships, kept by hand so each names the right people in
// words a reader follows. TestEveryLinkedModuleIsCredited holds the modules here to the ones each
// platform's build links, in both directions.
var allCredits = []Credit{
	metNorway, geoNames, yrIcons,
	{"", "Go standard library", "BSD-3-Clause", "the language and its runtime", nil},
	{"", "tz database", "public domain, IANA", "the time zone rules, built in through Go's time/tzdata", nil},
	{"github.com/wailsapp/wails/v2", "Wails v2", "MIT", "the desktop shell", nil},
	{"", "React and React DOM", "MIT", "the user interface", nil},
	{"golang.org/x/sys", "golang.org/x/sys", "BSD-3-Clause", "the tray icon, the displays, the window and Start with Windows", windowsOnly},
	{"golang.org/x/sys", "golang.org/x/sys", "BSD-3-Clause", "sending the runtime's own error reports to the log", []string{MacOS}},
	{"golang.org/x/sys", "golang.org/x/sys", "BSD-3-Clause", "sending the runtime's own error reports to the log; also used by Wails", []string{Linux}},
	{"github.com/godbus/dbus/v5", "godbus", "BSD-2-Clause", "the tray icon and its menu; hearing a shutdown or restart coming", []string{Linux}},
	{"github.com/go-ole/go-ole", "go-ole", "MIT", "the setup program's Start Menu and Desktop shortcuts", windowsOnly},
	// Linked in by the modules above rather than named by this application; each role says which.
	{"github.com/wailsapp/go-webview2", "go-webview2", "MIT", "used by Wails", windowsOnly},
	{"github.com/wailsapp/mimetype", "mimetype", "MIT", "used by Wails", nil},
	{"git.sr.ht/~jackmordaunt/go-toast/v2", "go-toast", "MIT or Unlicense", "used by Wails", windowsOnly},
	{"github.com/bep/debounce", "debounce", "MIT", "used by Wails", windowsOnly},
	{"github.com/google/uuid", "google/uuid", "BSD-3-Clause", "used by Wails", windowsOnly},
	{"github.com/leaanthony/go-ansi-parser", "go-ansi-parser", "MIT", "used by Wails", nil},
	{"github.com/leaanthony/slicer", "slicer", "MIT", "used by Wails", nil},
	{"github.com/leaanthony/u", "leaanthony/u", "MIT", "used by Wails", nil},
	{"github.com/pkg/browser", "pkg/browser", "BSD-2-Clause", "used by Wails", nil},
	{"github.com/samber/lo", "lo", "MIT", "used by Wails", nil},
	{"github.com/tkrajina/go-reflector", "go-reflector", "Apache-2.0", "used by Wails", nil},
	{"golang.org/x/net", "golang.org/x/net", "BSD-3-Clause", "used by Wails and mimetype", nil},
	{"github.com/pkg/errors", "pkg/errors", "BSD-2-Clause", "used by Wails", nil},
	{"github.com/rivo/uniseg", "uniseg", "MIT", "used by go-ansi-parser", nil},
	{"golang.org/x/text", "golang.org/x/text", "BSD-3-Clause", "used by lo", nil},
}
