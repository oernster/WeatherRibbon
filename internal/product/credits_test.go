package product

import (
	"runtime"
	"slices"
	"strings"
	"testing"
)

// modulesOf answers the modules credits name.
func modulesOf(credits []Credit) []string {
	var out []string
	for _, credit := range credits {
		out = append(out, credit.Module)
	}
	return out
}

// FR-610: each platform's About names what that platform's build ships; a component with no
// platforms named ships on every one.
func TestEachPlatformCreditsWhatItShips(t *testing.T) {
	t.Parallel()
	windows, linux, mac := modulesOf(CreditsFor(Windows)), modulesOf(CreditsFor(Linux)), modulesOf(CreditsFor(MacOS))
	if !slices.Contains(windows, "github.com/wailsapp/go-webview2") || slices.Contains(linux, "github.com/wailsapp/go-webview2") {
		t.Error("the Windows web view is not credited on Windows alone")
	}
	if !slices.Contains(linux, "github.com/godbus/dbus/v5") || slices.Contains(mac, "github.com/godbus/dbus/v5") {
		t.Error("the Linux tray's bus is not credited on Linux alone")
	}
	for _, list := range [][]string{windows, linux, mac} {
		if !slices.Contains(list, "github.com/wailsapp/wails/v2") {
			t.Error("Wails is not credited on every platform")
		}
	}
	if len(Credits) != len(CreditsFor(runtime.GOOS)) {
		t.Error("Credits is not this platform's list")
	}
}

// FR-610: every platform's About credits MET Norway's data and GeoNames under CC BY 4.0 with a link
// to the licence; the Yr weather icons under MIT.
func TestEveryDataSourceIsCredited(t *testing.T) {
	t.Parallel()
	for _, goos := range Platforms {
		credits := CreditsFor(goos)
		for _, want := range []struct{ name, licence string }{
			{"MET Norway", "CC BY 4.0"}, {"GeoNames", "CC BY 4.0"}, {"Yr weather icons", "MIT"},
		} {
			index := slices.IndexFunc(credits, func(c Credit) bool { return strings.HasPrefix(c.Name, want.name) })
			if index < 0 || !strings.HasPrefix(credits[index].Licence, want.licence) {
				t.Errorf("%s: %s is not credited under %s", goos, want.name, want.licence)
				continue
			}
			if want.licence == "CC BY 4.0" && !strings.Contains(credits[index].Licence, "https://creativecommons.org/licenses/by/4.0/") {
				t.Errorf("%s: %s carries no link to its licence", goos, want.name)
			}
		}
	}
}
