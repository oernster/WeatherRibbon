package main

import (
	"context"
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/ui/window/windowtest"
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

// recordingFetcher stands in for the refresher, counting what the facade asked of it.
type recordingFetcher struct{ pokes, asked int }

func (f *recordingFetcher) poke() { f.pokes++ }

func (f *recordingFetcher) askNow() { f.asked++ }

// newTestApp answers WeatherRibbon's half of the facade over a scripted service, its window and the
// refresher the recording stand-ins answered with it.
func newTestApp(t *testing.T) (*App, *scriptedService, *windowtest.Control, *recordingFetcher) {
	t.Helper()
	service, control, fetch := &scriptedService{}, &windowtest.Control{}, &recordingFetcher{}
	return &App{service: service, control: control, fetch: fetch, ctx: context.Background()}, service, control, fetch
}
