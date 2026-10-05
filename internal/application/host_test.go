package application

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/weatherribbon/internal/domain/place"
)

// horizontalRig answers a started rig whose ribbon runs horizontally and holds a city for each of
// places; the arithmetic below is worked for horizontal cells, whatever the default orientation is.
func horizontalRig(t *testing.T, places ...place.Place) *rig {
	t.Helper()
	r := newRig()
	r.store.loaded.Settings.Orientation = ribbon.Horizontal
	if err := r.service.Start(); err != nil {
		t.Fatal(err)
	}
	for _, each := range places {
		if _, err := r.service.AddCity(each.GeoNamesID); err != nil {
			t.Fatal(err)
		}
	}
	return r
}

// FR-106, FR-104: the arranger is handed the layout's cell, the prompt's when there are no cities,
// one cell per city and one more for each notice; the settings' choices are handed with it.
func TestTheServiceHandsTheArrangerItsContent(t *testing.T) {
	t.Parallel()
	for name, each := range map[string]struct {
		places []place.Place
		want   arranger.Content
	}{
		"empty":  {nil, arranger.Content{Cell: testLayout.Prompt, Cells: 1, Padding: testLayout.Padding}},
		"cities": {[]place.Place{london, tokyo}, arranger.Content{Cell: testLayout.Cell, Cells: 2, Padding: testLayout.Padding}},
	} {
		r := horizontalRig(t, each.places...)
		choices, got := host{r.service}.Ribbon()
		if got != each.want || choices != r.service.Settings().Choices {
			t.Errorf("%s: handed %+v with %+v, want %+v with the settings' own", name, got, choices, each.want)
		}
	}
}

// FR-702: a change the arranger makes is saved with the rest of the settings, so a save that fails
// raises the same notice.
func TestTheArrangersChangesAreSavedWithTheSettings(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t, london, tokyo)
	moved := func(c ribbon.Choices) ribbon.Choices {
		c.LastEdge = &placement.Against{Device: primaryMonitor.Device, Edge: placement.Top}
		return c
	}
	last := func() int { return len(r.store.saved) - 1 }
	if err := (host{r.service}).ChangeRibbon(moved); err != nil || r.store.saved[last()].LastEdge == nil || len(r.store.saved[last()].Cities) != 2 {
		t.Fatalf("saved %+v (%v), want the edge saved with the cities", r.store.saved[last()], err)
	}
	r.store.saveErr = errPlanted
	if err := (host{r.service}).ChangeRibbon(moved); !errors.Is(err, errPlanted) || len(r.service.Snapshot().Notices) != 1 {
		t.Errorf("a failed save answered %v with notices %v", err, r.service.Snapshot().Notices)
	}
}

// FR-104: one city at the top edge is 160 + 16 = 176 wide, centred at (1920 - 176) / 2 = 872. With a
// second city added it is 336 wide, centred again at (1920 - 336) / 2 = 792; that place is kept.
// A notice is one more cell: 3 x 160 + 16 = 496 wide.
func TestARibbonWhoseLengthChangesIsRecentredAndKept(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t, london)
	launched, err := r.service.Launch()
	if err != nil || launched.Size.Width != 176 || launched.At.X != 872 {
		t.Fatalf("launched %+v (%v)", launched, err)
	}
	if _, err := r.service.AddCity(tokyo.GeoNamesID); err != nil {
		t.Fatal(err)
	}
	grown, err := r.service.Rearrange(launched.At)
	if err != nil || grown.Size.Width != 336 || grown.At.X != 792 {
		t.Fatalf("grown %+v (%v)", grown, err)
	}
	if kept := r.service.Settings().Placement; kept == nil {
		t.Error("the re-centred place was not kept")
	}
	r.store.saveErr = errPlanted
	if err := r.service.SetColour(ribbon.Neon); !errors.Is(err, errPlanted) {
		t.Fatalf("the save did not fail: %v", err)
	}
	if noticed, _ := r.service.Rearrange(grown.At); noticed.Size.Width != 496 {
		t.Errorf("with a notice: %+v", noticed)
	}
}
