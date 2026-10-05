package main

import (
	"context"
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/ui/window"
	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// errPlanted is the failure a stand-in answers when a test asks it to fail.
var errPlanted = errors.New("planted failure")

// scriptedService stands in for application.Service in WeatherRibbon's own half of the facade. It
// answers what a test sets and records the calls the facade made.
type scriptedService struct {
	settings settings.Settings
	snapshot application.Snapshot
	places   []place.Place
	choices  []menus.Item
	detail   application.Detail
	samples  application.Samples
	// measured is the last measurement SetMeasured was handed; units the last system chosen.
	measured application.Measured
	units    units.System
	// changeErr answers every change.
	changeErr error
	calls     []string
}

func (s *scriptedService) record(call string) { s.calls = append(s.calls, call) }

func (s *scriptedService) change(call string) error {
	s.record(call)
	return s.changeErr
}

func (s *scriptedService) Snapshot() application.Snapshot { return s.snapshot }

func (s *scriptedService) Settings() settings.Settings { return s.settings }

func (s *scriptedService) Search(string) []place.Place { return s.places }

func (s *scriptedService) AddCity(int) (string, error) {
	s.record("AddCity")
	return "city-1", s.changeErr
}

func (s *scriptedService) Relabel(string, string) error { return s.change("Relabel") }

func (s *scriptedService) ChangePlace(string, int) error { return s.change("ChangePlace") }

func (s *scriptedService) RemoveCity(string) error { return s.change("RemoveCity") }

func (s *scriptedService) SetUnits(system units.System) error {
	s.units = system
	return s.change("SetUnits")
}

func (s *scriptedService) SetFormat(localtime.Format) error { return s.change("SetFormat") }

func (s *scriptedService) DismissNotices() { s.record("DismissNotices") }

func (s *scriptedService) DismissPetrichor(string) { s.record("DismissPetrichor") }

func (s *scriptedService) OpenDetail(context.Context, string) (application.Detail, error) {
	s.record("OpenDetail")
	return s.detail, s.changeErr
}

func (s *scriptedService) TextSamples() application.Samples { return s.samples }

func (s *scriptedService) SetMeasured(measured application.Measured) error {
	s.measured = measured
	return s.change("SetMeasured")
}

func (s *scriptedService) SettingsChoices() []menus.Item { return s.choices }

// recordingControl stands in for the window, recording what WeatherRibbon's half asked of it in order.
type recordingControl struct {
	calls []string
	// reported is each failure reported, as its doing; panels each panel shown.
	reported []string
	panels   []string
	colours  []string
	turned   []string
	shown    window.Shown
	// choiceErr answers SetColour and SetOrientation.
	choiceErr error
}

func (c *recordingControl) record(call string) { c.calls = append(c.calls, call) }

func (c *recordingControl) Refitted(err error) error {
	c.record("Refitted")
	return err
}

func (c *recordingControl) Redraw() { c.record("Redraw") }

func (c *recordingControl) Report(doing string, err error) {
	if err != nil {
		c.reported = append(c.reported, doing)
	}
}

func (c *recordingControl) ShowPanel(panel string) { c.panels = append(c.panels, panel) }

func (c *recordingControl) PageMeasured() { c.record("PageMeasured") }

func (c *recordingControl) Shown() window.Shown { return c.shown }

// Offered greys every item, so a test reads which menu passed through it.
func (c *recordingControl) Offered(items []menus.Item) []menus.Item {
	greyed := make([]menus.Item, len(items))
	for index, item := range items {
		item.Disabled = true
		greyed[index] = item
	}
	return greyed
}

func (c *recordingControl) SetColour(colour string) error {
	c.colours = append(c.colours, colour)
	return c.choiceErr
}

func (c *recordingControl) SetOrientation(orientation string) error {
	c.turned = append(c.turned, orientation)
	return c.choiceErr
}

// count answers how often the facade asked the window for call.
func (c *recordingControl) count(call string) int {
	n := 0
	for _, each := range c.calls {
		if each == call {
			n++
		}
	}
	return n
}

// recordingFetcher stands in for the refresher, counting what the facade asked of it.
type recordingFetcher struct{ pokes, asked int }

func (f *recordingFetcher) poke() { f.pokes++ }

func (f *recordingFetcher) askNow() { f.asked++ }

// newTestApp answers WeatherRibbon's half of the facade over a scripted service, its window and the
// refresher the recording stand-ins answered with it.
func newTestApp(t *testing.T) (*App, *scriptedService, *recordingControl, *recordingFetcher) {
	t.Helper()
	service, control, fetch := &scriptedService{}, &recordingControl{}, &recordingFetcher{}
	return &App{service: service, control: control, fetch: fetch, ctx: context.Background()}, service, control, fetch
}
