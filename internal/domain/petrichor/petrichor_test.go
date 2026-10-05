package petrichor

import (
	"testing"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/forecast"
)

// at answers an instant hours after 2026-10-01T09:00Z, where FR-413's worked example begins.
func at(hours int) time.Time {
	return time.Date(2026, time.October, 1, 9, 0, 0, 0, time.UTC).Add(time.Duration(hours) * time.Hour)
}

// observeAll answers the spell after each finding in turn, with whether the last was an event.
func observeAll(findings []Weather, hours []int) (Spell, bool) {
	var spell Spell
	event := false
	for index, each := range findings {
		spell, event = spell.Observe(each, at(hours[index]))
	}
	return spell, event
}

// FR-413: London dry from 2026-10-01T09:00Z, wet at 2026-10-04T10:00Z, 73 hours later: an event.
func TestRainAfterThreeDryDaysIsAPetrichorEvent(t *testing.T) {
	t.Parallel()
	spell, event := observeAll([]Weather{Dry, Dry, Dry, Wet}, []int{0, 24, 72, 73})
	if !event || spell.Dry {
		t.Fatalf("event %v, spell %+v; want an event and the spell ended", event, spell)
	}
	if _, exact := observeAll([]Weather{Dry, Wet}, []int{0, 72}); !exact {
		t.Error("rain exactly 72 hours into a spell is an event")
	}
}

// FR-413: rain 48 hours in is no event; the spell starts afresh at the next dry snapshot.
func TestRainSoonerIsNoEvent(t *testing.T) {
	t.Parallel()
	spell, event := observeAll([]Weather{Dry, Wet}, []int{0, 48})
	if event || spell.Dry {
		t.Fatalf("event %v, spell %+v; want no event and no spell", event, spell)
	}
	again, _ := spell.Observe(Dry, at(49))
	if !again.Dry || !again.Since.Equal(at(49)) {
		t.Errorf("the next dry snapshot begins a spell at 49 hours: got %+v", again)
	}
	if _, event := (Spell{}).Observe(Wet, at(100)); event {
		t.Error("rain with no spell under way is no event")
	}
}

// FR-413: a stale forecast is neither wet nor dry and changes nothing; a gap in snapshots does not
// break a spell.
func TestAStaleForecastIsNeitherWetNorDry(t *testing.T) {
	t.Parallel()
	spell, event := observeAll([]Weather{Dry, Neither, Neither, Wet}, []int{0, 10, 60, 80})
	if !event || spell.Dry {
		t.Fatalf("event %v; want the spell from hour 0 to hold through findings of neither", event)
	}
	if kept, _ := (Spell{}).Observe(Neither, at(5)); kept.Dry {
		t.Error("a finding of neither begins no spell")
	}
}

// FR-413, OQ-14: a countdown is drawn uniformly from 10 to 15, both included.
func TestTheCountdownIsDrawnFromTenToFifteen(t *testing.T) {
	t.Parallel()
	asked := -1
	lowest := NewCountdown(func(n int) int { asked = n; return 0 })
	highest := NewCountdown(func(n int) int { return n - 1 })
	if asked != MostEvents-FewestEvents+1 || lowest != FewestEvents || highest != MostEvents {
		t.Fatalf("drew from %d, giving %d to %d; want 6 choices giving 10 to 15", asked, lowest, highest)
	}
}

// FR-413: each event takes one; the one bringing the countdown to zero is a moment and a new one is
// drawn; an undrawn countdown is drawn first.
func TestTheEventThatEndsTheCountdownIsAMoment(t *testing.T) {
	t.Parallel()
	lowest := func(int) int { return 0 }
	left, moment := Countdown(2).Count(lowest)
	if moment || left != 1 {
		t.Fatalf("from 2: %d left, moment %v; want 1 left and no moment", left, moment)
	}
	left, moment = left.Count(lowest)
	if !moment || left != FewestEvents {
		t.Fatalf("from 1: %d left, moment %v; want a moment and a fresh countdown of 10", left, moment)
	}
	if first, moment := Countdown(0).Count(lowest); moment || first != FewestEvents-1 {
		t.Errorf("an undrawn countdown counting one event leaves %d, moment %v; want 9", first, moment)
	}
}

// A countdown is valid when undrawn or within the draw's bounds.
func TestACountdownOutsideItsBoundsIsInvalid(t *testing.T) {
	t.Parallel()
	for countdown, want := range map[Countdown]bool{-1: false, 0: true, 1: true, MostEvents: true, MostEvents + 1: false} {
		if got := countdown.Valid(); got != want {
			t.Errorf("Countdown(%d).Valid() = %v; want %v", countdown, got, want)
		}
	}
}

// FR-413: rain above 0 is wet, none is dry; a stale forecast or a period giving no rain is neither.
func TestAStaleForecastJudgesNeither(t *testing.T) {
	t.Parallel()
	for name, each := range map[string]struct {
		current forecast.Current
		usable  bool
		want    Weather
	}{
		"rain":          {forecast.Current{RainMM: 0.3, HasRain: true}, true, Wet},
		"none":          {forecast.Current{RainMM: 0, HasRain: true}, true, Dry},
		"stale":         {forecast.Current{RainMM: 0.3, HasRain: true}, false, Neither},
		"no rain given": {forecast.Current{}, true, Neither},
	} {
		if got := Judge(each.current, each.usable); got != each.want {
			t.Errorf("%s: Judge = %v; want %v", name, got, each.want)
		}
	}
}

// FR-413: the line goes once the city is found dry; wet or neither keeps it.
func TestTheMomentEndsWhenTheRainDoes(t *testing.T) {
	t.Parallel()
	for weather, want := range map[Weather]bool{Wet: true, Neither: true, Dry: false} {
		if got := Shown(weather); got != want {
			t.Errorf("Shown(%v) = %v; want %v", weather, got, want)
		}
	}
}
