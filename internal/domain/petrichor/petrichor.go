// Package petrichor decides the petrichor moment (FR-413): a rare, quiet line shown where rain
// arrives after a dry spell, inviting the curious to look the word up.
//
// Instants arrive as arguments and the countdown's randomness through a draw the caller injects, so
// this package reads no clock and holds no source of chance (OQ-14).
package petrichor

import (
	"time"

	"github.com/oernster/weatherribbon/internal/domain/forecast"
)

// Weather is what one snapshot found a city to be.
type Weather int

// The three findings (FR-413): rain forecast for the current period; none; nothing to judge by,
// because the forecast is stale or has run out.
const (
	Neither Weather = iota
	Dry
	Wet
)

// Judge answers what a snapshot finds a city to be from its current conditions (FR-413): wet where
// the current period forecasts rain above 0, dry where it forecasts none. Neither where the period
// gives no rain figure; neither too where usable is false, the forecast being stale (FR-305) or run
// out (FR-403).
func Judge(current forecast.Current, usable bool) Weather {
	switch {
	case !usable || !current.HasRain:
		return Neither
	case current.RainMM > 0:
		return Wet
	default:
		return Dry
	}
}

// SpellLength is how long a city must have been found dry before rain counts as an event (OQ-13).
const SpellLength = 72 * time.Hour

// Spell is one city's dry spell: the instant from which every snapshot has found it dry. The zero
// Spell is no spell.
type Spell struct {
	// Since is when the spell began; meaningful only while Dry holds.
	Since time.Time
	// Dry says whether a spell is under way.
	Dry bool
}

// Observe answers the spell after a snapshot finding weather at instant; with it, whether that
// snapshot is a petrichor event. A dry finding begins a spell where none is under way and leaves one under way
// alone; a wet finding ends the spell, an event where it had lasted SpellLength; a finding of neither
// changes nothing. Time while nothing was observed does not break a spell.
func (s Spell) Observe(weather Weather, instant time.Time) (Spell, bool) {
	switch weather {
	case Dry:
		if !s.Dry {
			return Spell{Since: instant, Dry: true}, false
		}
		return s, false
	case Wet:
		return Spell{}, s.Dry && instant.Sub(s.Since) >= SpellLength
	default:
		return s, false
	}
}

// The bounds a countdown is drawn between, both included (OQ-14).
const (
	FewestEvents = 10
	MostEvents   = 15
)

// Countdown is the petrichor events left before the next moment, one for all cities. Zero means none
// has been drawn yet.
type Countdown int

// Draw answers a uniformly random whole number from zero up to but not including n. It is injected so
// the domain holds no source of chance.
type Draw func(n int) int

// NewCountdown answers a countdown drawn uniformly from FewestEvents to MostEvents.
func NewCountdown(draw Draw) Countdown {
	return Countdown(FewestEvents + draw(MostEvents-FewestEvents+1))
}

// Valid answers whether c is undrawn or one a draw could have left: zero, else 1 to MostEvents.
func (c Countdown) Valid() bool { return c >= 0 && c <= MostEvents }

// Count answers the countdown after one petrichor event; with it, whether that event is a moment. An
// undrawn countdown is drawn first. The event that brings it to zero is a moment, after which a new
// countdown is drawn.
func (c Countdown) Count(draw Draw) (Countdown, bool) {
	if c <= 0 {
		c = NewCountdown(draw)
	}
	c--
	if c == 0 {
		return NewCountdown(draw), true
	}
	return c, false
}

// Shown says whether a moment showing for a city goes on after a snapshot finding weather (FR-413):
// it goes once the city is found dry; a wet finding keeps it and so does a finding of neither, which
// changes nothing. A click hides it whatever the weather; the caller handles that.
func Shown(weather Weather) bool { return weather != Dry }
