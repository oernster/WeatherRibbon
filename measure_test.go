package main

import (
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// FR-103: the page is handed the samples whole, every list present; its measurement reaches the
// service in the service's own terms; once it is applied, the window hears the page has been
// measured (FR-102).
func TestTheMeasurementRoundTrip(t *testing.T) {
	app, service, control, _ := newTestApp(t)
	service.samples = application.Samples{Times: []string{"23:59"}, Temperatures: []int{-60, 60}}
	samples := app.TextSamples()
	if !slices.Equal(samples.Times, []string{"23:59"}) || !slices.Equal(samples.Temperatures, []int{-60, 60}) ||
		samples.Weekdays == nil || samples.Labels == nil {
		t.Errorf("samples %+v", samples)
	}
	if err := app.SetMeasured(measuredDTO{Units: "imperial", Format: "12h", Labels: []string{"London"}, CellWidth: 201}); err != nil {
		t.Fatal(err)
	}
	got := service.measured
	if got.Units != units.Imperial || got.Format != localtime.TwelveHour || !slices.Equal(got.Labels, []string{"London"}) || got.CellWidth != 201 {
		t.Errorf("service was handed %+v", got)
	}
	if !slices.Equal(control.calls, []string{"Refitted", "PageMeasured"}) {
		t.Errorf("the window was asked %v, want the ribbon fitted then the page counted measured", control.calls)
	}
}

// A measurement the service refused does not count as applied, so the launched ribbon waits on.
func TestARefusedMeasurementIsNotCountedApplied(t *testing.T) {
	app, service, control, _ := newTestApp(t)
	service.changeErr = errPlanted
	_ = app.SetMeasured(measuredDTO{CellWidth: 201})
	if slices.Contains(control.calls, "PageMeasured") {
		t.Error("a refused measurement counted the page measured")
	}
}
