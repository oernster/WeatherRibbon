package application

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// measuredFor answers a measurement of width taken under the rig's current settings.
func measuredFor(r *rig, width int) Measured {
	current := r.service.Settings()
	return Measured{Units: current.Units, Format: current.Format, Labels: labelsOf(current), CellWidth: width}
}

// FR-103: the samples are every time in the format, every weekday, every temperature in the units
// and every label held.
func TestTheTextSamplesFollowTheSettings(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t, london, tokyo)
	if err := r.service.SetUnits(units.Imperial); err != nil {
		t.Fatal(err)
	}
	if err := r.service.SetFormat(localtime.TwelveHour); err != nil {
		t.Fatal(err)
	}
	got := r.service.TextSamples()
	if !slices.Contains(got.Times, "11:59 PM") || !slices.Equal(got.Times, localtime.TimeSamples(localtime.TwelveHour)) {
		t.Errorf("%d times starting %q", len(got.Times), got.Times[0])
	}
	if len(got.Weekdays) != 7 || got.Weekdays[0] != "Sunday" || got.Weekdays[6] != "Saturday" {
		t.Errorf("weekdays %v", got.Weekdays)
	}
	if !slices.Equal(got.Temperatures, units.Temperatures(units.Imperial)) {
		t.Errorf("temperatures from %d to %d", got.Temperatures[0], got.Temperatures[len(got.Temperatures)-1])
	}
	if !slices.Equal(got.Labels, []string{london.Name, tokyo.Name}) {
		t.Errorf("labels %v", got.Labels)
	}
}

// FR-103: a measured width wider than the layout's widens every cell, in the snapshot the page draws
// from and in the window alike: two cells of 168 are 2 x 168 + 16 = 352 wide. A narrower one leaves
// the layout's own.
func TestAMeasuredWidthWidensTheCells(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t, london, tokyo)
	wider := testLayout.Cell.Width + testLayout.Padding
	if err := r.service.SetMeasured(measuredFor(r, wider)); err != nil {
		t.Fatal(err)
	}
	if got := r.service.Snapshot().Layout.Cell.Width; got != wider {
		t.Errorf("snapshot cell %d, want %d", got, wider)
	}
	if got, err := r.service.Launch(); err != nil || got.Size.Width != 352 {
		t.Errorf("window %+v (%v)", got.Size, err)
	}
	if err := r.service.SetMeasured(measuredFor(r, testLayout.Cell.Width-1)); err != nil {
		t.Fatal(err)
	}
	if got := r.service.Snapshot().Layout.Cell.Width; got != testLayout.Cell.Width {
		t.Errorf("a narrower measurement gave %d", got)
	}
}

// FR-103: a measurement taken under other units, another format or other labels widens nothing until
// the page measures again; one handed in is not changed by its caller afterwards; a width below zero
// is refused.
func TestAMeasurementCountsOnlyForWhatItWasTakenUnder(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t, london, tokyo)
	wider := testLayout.Cell.Width + testLayout.Padding
	for name, other := range map[string]func(*Measured){
		"units":  func(m *Measured) { m.Units = units.Imperial },
		"format": func(m *Measured) { m.Format = localtime.TwelveHour },
		"labels": func(m *Measured) { m.Labels = []string{london.Name} },
	} {
		measured := measuredFor(r, wider)
		other(&measured)
		if err := r.service.SetMeasured(measured); err != nil {
			t.Fatal(err)
		}
		if got := r.service.Snapshot().Layout.Cell.Width; got != testLayout.Cell.Width {
			t.Errorf("%s: widened the cell to %d", name, got)
		}
	}
	measured := measuredFor(r, wider)
	if err := r.service.SetMeasured(measured); err != nil {
		t.Fatal(err)
	}
	measured.Labels[0] = "Elsewhere"
	if got := r.service.Snapshot().Layout.Cell.Width; got != wider {
		t.Errorf("the caller's change reached the measurement: %d", got)
	}
	if err := r.service.SetMeasured(measuredFor(r, -1)); !errors.Is(err, placement.ErrNegativeLength) {
		t.Errorf("a negative width: %v", err)
	}
}
