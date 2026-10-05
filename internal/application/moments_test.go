package application

import (
	"testing"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/petrichor"
)

// FR-413: London dry from now, wet 73 hours on with the countdown at 1: its cell shows the line, a new
// countdown is drawn and kept; a dry finding later takes the line away.
func TestRainAfterADrySpellShowsTheMomentWhereItRained(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID, tokyo.GeoNamesID)
	r.service.current.PetrichorCountdown = 1
	start := r.clock.now
	hold(r, "city-1", hourlyAround(start, 10, 0), start)
	_ = r.service.Snapshot()
	if spell := r.cache.entries["city-1"].Spell; !spell.Dry || !spell.Since.Equal(start) {
		t.Fatalf("the spell was not begun and kept: %+v", spell)
	}
	r.clock.now = start.Add(73 * time.Hour)
	hold(r, "city-1", hourlyAround(r.clock.now, 10, 0.3), r.clock.now)
	cells := r.service.Snapshot().Cells
	if !cells[0].Petrichor || cells[1].Petrichor {
		t.Fatalf("petrichor %v %v; want London alone", cells[0].Petrichor, cells[1].Petrichor)
	}
	if got := r.service.Settings().PetrichorCountdown; got != petrichor.FewestEvents {
		t.Errorf("countdown %d; want a fresh one drawn and kept", got)
	}
	if again := r.service.Snapshot().Cells[0]; !again.Petrichor {
		t.Error("the line went while it still rained")
	}
	hold(r, "city-1", hourlyAround(r.clock.now, 10, 0), r.clock.now)
	if dry := r.service.Snapshot().Cells[0]; dry.Petrichor {
		t.Error("a dry finding did not take the line away")
	}
}

// FR-413: an event that does not end the countdown shows nothing; a click hides a line.
func TestAnEventBeforeTheCountdownEndsShowsNothing(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	r.service.current.PetrichorCountdown = 5
	start := r.clock.now
	hold(r, "city-1", hourlyAround(start, 10, 0), start)
	_ = r.service.Snapshot()
	r.clock.now = start.Add(80 * time.Hour)
	hold(r, "city-1", hourlyAround(r.clock.now, 10, 1), r.clock.now)
	if cell := r.service.Snapshot().Cells[0]; cell.Petrichor {
		t.Error("an event short of the countdown showed the line")
	}
	if got := r.service.Settings().PetrichorCountdown; got != 4 {
		t.Errorf("countdown %d; want 4", got)
	}
	r.service.weather["city-1"].moment = true
	r.service.DismissPetrichor("city-1")
	r.service.DismissPetrichor("nobody")
	if r.service.weather["city-1"].moment {
		t.Error("a click did not hide the line")
	}
}
