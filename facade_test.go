package main

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/ribbonkit/ui/window"
	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/units"
	"github.com/oernster/weatherribbon/internal/product"
)

// Every change is followed by fitting the ribbon whether or not it saved, since a change whose save
// failed raises a notice, one more cell to fit (FR-702); the service's error is answered.
func TestEveryChangeFitsTheRibbonAndAnswersTheServicesError(t *testing.T) {
	changes := map[string]func(app *App) error{
		"AddCity":        func(app *App) error { _, err := app.AddCity(2643743); return err },
		"Relabel":        func(app *App) error { return app.RenameCity("city-1", "Home") },
		"ChangePlace":    func(app *App) error { return app.ChangePlace("city-1", 1850147) },
		"RemoveCity":     func(app *App) error { return app.RemoveCity("city-1") },
		"SetUnits":       func(app *App) error { return app.SetUnits("imperial") },
		"SetFormat":      func(app *App) error { return app.SetFormat("12h") },
		"SetMeasured":    func(app *App) error { return app.SetMeasured(measuredDTO{CellWidth: 200}) },
		"DismissNotices": func(app *App) error { app.DismissNotices(); return nil },
	}
	for name, change := range changes {
		for _, failure := range []error{nil, errPlanted} {
			app, service, control, _ := newTestApp(t)
			service.changeErr = failure
			if err := change(app); name != "DismissNotices" && !errors.Is(err, failure) {
				t.Errorf("%s answered %v, want the service's %v", name, err, failure)
			}
			if !slices.Contains(service.calls, name) {
				t.Errorf("%s never reached the service", name)
			}
			if control.count("Refitted") != 1 {
				t.Errorf("%s (service answered %v) fitted the ribbon %d times, want once", name, failure, control.count("Refitted"))
			}
		}
	}
}

// FR-201, FR-207: a city added or moved has its forecast asked for at once; nothing else does.
func TestANewPlaceIsAskedForAtOnce(t *testing.T) {
	app, _, _, fetch := newTestApp(t)
	_, _ = app.AddCity(2643743)
	_ = app.ChangePlace("city-1", 1850147)
	_ = app.RenameCity("city-1", "Home")
	_ = app.RemoveCity("city-1")
	if fetch.pokes != 2 {
		t.Errorf("the refresher was woken %d times, want twice", fetch.pokes)
	}
}

// The snapshot reaches the page whole: every list present rather than null, since the page counts
// them; the ribbon's choices and the window's reading of it carried.
func TestTheSnapshotCarriesEveryCellAndTheWindowsReading(t *testing.T) {
	app, service, control, _ := newTestApp(t)
	service.snapshot = application.Snapshot{
		Cells: []application.Cell{{
			ID: "city-1", Label: "London", Time: "08:36", Symbol: application.Symbol{Icon: "rain"},
			Today:   application.Day{Date: forecast.Date{Year: 2026, Month: 10, Day: 5}, Symbol: application.Symbol{Words: "fog"}},
			Outlook: []application.Day{{Date: forecast.Date{Year: 2026, Month: 10, Day: 6}, Weekday: "Tuesday", High: 15, Known: true}},
		}},
		Units:   units.Imperial,
		Choices: ribbon.Choices{Colour: ribbon.Ocean, Opacity: 40},
		Scale:   125,
	}
	service.choices = []menus.Item{{Action: menus.Pin, Label: "Pin ribbon", Checkable: true}}
	control.shown = window.Shown{Collapsed: true, Scrolls: true, DragThreshold: placement.Size{Width: 4, Height: 4}}
	got := app.Snapshot()
	if len(got.Cells) != 1 || got.Cells[0].Symbol != (symbolDTO{Icon: "rain"}) || got.Cells[0].Today.Date != "2026-10-05" {
		t.Errorf("cells %+v, want the one cell the service answered", got.Cells)
	}
	if got.Notices == nil || got.Cells[0].Today.Symbol.Words != "fog" {
		t.Error("the notices went out as null or the day's words were lost")
	}
	if outlook := got.Cells[0].Outlook; len(outlook) != 1 || outlook[0].Date != "2026-10-06" || outlook[0].Weekday != "Tuesday" || outlook[0].High != 15 || !outlook[0].Known {
		t.Errorf("outlook %+v, want the service's one day", outlook)
	}
	if got.Units != "imperial" || got.Colour != "ocean" || got.Opacity != 40 || got.Scale != 125 || got.MinScale != ribbon.MinScale {
		t.Errorf("units %q, colour %q, opacity %d, scale %v of at least %d", got.Units, got.Colour, got.Opacity, got.Scale, got.MinScale)
	}
	if !got.Collapsed || !got.Scrolls || got.DragThreshold != (sizeDTO{Width: 4, Height: 4}) || got.StartLabel != product.StartAtSignIn {
		t.Errorf("collapsed %v, scrolls %v, threshold %v, start label %q", got.Collapsed, got.Scrolls, got.DragThreshold, got.StartLabel)
	}
	if len(got.Choices) != 1 || got.Choices[0].Action != string(menus.Pin) || got.Choices[0].Children == nil || !got.Choices[0].Disabled {
		t.Errorf("choices %+v, want the service's as the window offers them, never a null list of children", got.Choices)
	}
}

// FR-202: a place goes out with its id, its name and its name in full.
func TestSearchPlacesPassesThrough(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	service.places = []place.Place{{GeoNamesID: 2643743, Name: "London", Region: "England", Country: "United Kingdom"}}
	got := app.SearchPlaces("lon")
	if len(got) != 1 || got[0] != (placeDTO{GeoNamesID: 2643743, Name: "London", Description: service.places[0].Description()}) {
		t.Errorf("SearchPlaces answered %+v", got)
	}
}

// FR-410, FR-413: the detail panel and the petrichor line reach the service; the detail goes out with
// its hours present and the service's error answered.
func TestTheDetailAndThePetrichorLineReachTheService(t *testing.T) {
	app, service, _, _ := newTestApp(t)
	service.detail = application.Detail{ID: "city-1", Sunrise: "07:07"}
	got, err := app.OpenDetail("city-1")
	if err != nil || got.ID != "city-1" || got.Sunrise != "07:07" || got.Hours == nil {
		t.Errorf("detail %+v (%v)", got, err)
	}
	service.detail.Hours = []application.HourView{{Time: "08:00", Symbol: application.Symbol{Icon: "rain"}, Temperature: 14, Rain: 0.2, WindSpeed: 18, WindFrom: 225}}
	got, _ = app.OpenDetail("city-1")
	want := hourDTO{Time: "08:00", Symbol: symbolDTO{Icon: "rain"}, Temperature: 14, Rain: 0.2, WindSpeed: 18, WindFrom: 225}
	if len(got.Hours) != 1 || got.Hours[0] != want {
		t.Errorf("hours %+v, want %+v", got.Hours, want)
	}
	service.changeErr = errPlanted
	if _, err := app.OpenDetail("city-1"); !errors.Is(err, errPlanted) {
		t.Errorf("a refused detail answered %v", err)
	}
	app.DismissPetrichor("city-1")
	if !slices.Contains(service.calls, "DismissPetrichor") {
		t.Error("the petrichor line was not dismissed")
	}
}

// FR-109: Add city opens Settings on the place search; Refresh now asks the refresher; each choice
// is made, the page told to redraw.
func TestWeatherRibbonsMenuActions(t *testing.T) {
	app, service, control, fetch := newTestApp(t)
	app.actOn(application.ActionAddCity)
	if !slices.Equal(control.panels, []string{openAtAddCity}) {
		t.Errorf("Add city opened %v, want the place search", control.panels)
	}
	app.actOn(application.ActionRefreshNow)
	if fetch.asked != 1 {
		t.Errorf("Refresh now asked the refresher %d times", fetch.asked)
	}
	app.actOn(application.ActionImperial)
	app.actOn(menus.Action("colour-neon"))
	app.actOn(menus.OrientHorizontal)
	if service.units != units.Imperial || !slices.Equal(control.colours, []string{"neon"}) || !slices.Equal(control.turned, []string{"horizontal"}) {
		t.Errorf("units %q, colours %v, orientations %v; want each choice made", service.units, control.colours, control.turned)
	}
	if control.count("Redraw") != 3 {
		t.Errorf("the page was told to redraw %d times, want once for each of the three choices", control.count("Redraw"))
	}
}

func TestAMenuChoiceThatFailedIsReported(t *testing.T) {
	app, service, control, _ := newTestApp(t)
	service.changeErr, control.choiceErr = errPlanted, errPlanted
	for _, action := range []menus.Action{application.ActionMetric, "colour-ocean", menus.OrientVertical} {
		app.actOn(action)
	}
	want := []string{"changing the units", "changing the colour", "changing the orientation"}
	if !slices.Equal(control.reported, want) {
		t.Errorf("reported %v, want %v", control.reported, want)
	}
}

// An action that is neither the kit's nor WeatherRibbon's changes nothing.
func TestAnUnknownActionChangesNothing(t *testing.T) {
	app, service, control, fetch := newTestApp(t)
	app.actOn("no-such-action")
	if len(service.calls) != 0 || len(control.calls) != 0 || len(control.panels) != 0 || fetch.asked != 0 {
		t.Errorf("service %v, window %v, panels %v; want nothing", service.calls, control.calls, control.panels)
	}
}

// The window is handed WeatherRibbon's choices out of its settings and no pull out: on a first run,
// the first-run ones; asked for a pull out, it refuses.
func TestTheWindowReadsTheRibbonsChoicesAndHasNoPullOut(t *testing.T) {
	kit := kitService{application.New(application.Ports{}, application.Layout{})}
	if kit.Choices() != ribbon.Defaults() || kit.PullOut() {
		t.Errorf("choices %+v, pull out %v; want the first run's and none", kit.Choices(), kit.PullOut())
	}
	if err := kit.SetPullOut(true); !errors.Is(err, errNoPullOut) {
		t.Errorf("a pull out was not refused: %v", err)
	}
}

// What Help shows comes from internal/product and the LICENSE file (FR-603, FR-610).
func TestTheWindowIsHandedWeatherRibbonsProduct(t *testing.T) {
	got := productOf()
	if got.App != product.App() || got.WindowClass != product.RibbonClass || got.Version != product.Version {
		t.Errorf("names %+v, %q, %q; want internal/product's", got.App, got.WindowClass, got.Version)
	}
	if len(got.Credits) != len(product.Credits) || got.DonateURL != product.DonateURL {
		t.Errorf("credits %d, donation %q; want internal/product's", len(got.Credits), got.DonateURL)
	}
	if got.Licence != licenceText || licenceText == "" {
		t.Error("the licence is not the embedded LICENSE")
	}
}
