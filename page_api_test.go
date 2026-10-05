package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/oernster/ribbonkit/ui/window"
)

// The page states the facade in two halves: WeatherRibbon's Bridge in api.ts, which extends the
// window's WindowBridge in ribbonkit's bridge.ts. bridgeMethod finds each method named in one.
var (
	bridgeInterface = regexp.MustCompile(`(?s)interface Bridge extends WindowBridge \{(.*?)\n\}`)
	windowInterface = regexp.MustCompile(`(?s)interface WindowBridge \{(.*?)\n\}`)
	bridgeMethod    = regexp.MustCompile(`(?m)^\s+(\w+)\(`)
)

// pageHalves names each file stating a half of the facade with the pattern that finds it: the
// window's half is read from the kit the page is built from, as npm installed it.
var pageHalves = map[string]*regexp.Regexp{
	filepath.Join("frontend", "src", "api.ts"):                                              bridgeInterface,
	filepath.Join("frontend", "node_modules", "@oernster", "ribbonkit", "web", "bridge.ts"): windowInterface,
}

// pageCalls answers every method the page calls on the facade, as its two halves state them.
func pageCalls(t *testing.T) []string {
	t.Helper()
	var names []string
	for file, half := range pageHalves {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		found := half.FindSubmatch(raw)
		if found == nil {
			t.Fatalf("%s states no %s", file, half)
		}
		for _, method := range bridgeMethod.FindAllSubmatch(found[1], -1) {
			names = append(names, string(method[1]))
		}
	}
	return names
}

// Wails binds every exported method of the App it is handed, the window's promoted through the
// embedded Window included; the page reaches each by name (window.go.main.App). Every method the page
// calls must be in that set, whichever half of the facade holds it.
func TestEveryMethodThePageCallsIsBound(t *testing.T) {
	app, _ := newApp(context.Background(), nil, nil, window.Config{Log: io.Discard})
	bound := reflect.TypeOf(app)
	calls := pageCalls(t)
	if len(calls) == 0 {
		t.Fatal("api.ts's Bridge names no method")
	}
	for _, name := range calls {
		if _, ok := bound.MethodByName(name); !ok {
			t.Errorf("the page calls %s, which the bound App does not have", name)
		}
	}
}

// The Control is how WeatherRibbon reaches its window; none of it may be bound, so the page can never
// run the window's life, fit it or end it.
func TestNothingOfTheControlIsBound(t *testing.T) {
	bound := reflect.TypeOf(&App{})
	control := reflect.TypeOf(&window.Control{})
	for index := range control.NumMethod() {
		name := control.Method(index).Name
		if _, ok := bound.MethodByName(name); ok {
			t.Errorf("the Control's %s is bound to the page", name)
		}
	}
}
