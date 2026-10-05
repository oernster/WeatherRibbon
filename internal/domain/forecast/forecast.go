// Package forecast reads one MET Norway Locationforecast answer, already decoded into steps, into
// what a cell shows: the current conditions, today and the outlook (FR-402 to FR-409).
//
// Every instant arrives as an argument and every zone already resolved, so this package reads no
// clock and loads no zone (CON-5).
package forecast

import (
	"slices"
	"time"
)

// The spans of a step's next_1_hours, next_6_hours and next_12_hours blocks, in hours.
const (
	Next1Hours  = 1
	Next6Hours  = 6
	Next12Hours = 12
)

// Period is one of a step's period blocks: the span from its step's instant for Hours hours, the
// condition MET Norway names for it plus its rain and its extremes where the block gives them.
type Period struct {
	// Hours is the span in hours: Next1Hours, Next6Hours or Next12Hours. Zero means the step carries
	// no such block.
	Hours int
	// Symbol is MET Norway's symbol code, such as "partlycloudy_day".
	Symbol string
	// RainMM is precipitation_amount in millimetres; HasRain says whether the block gave one.
	RainMM  float64
	HasRain bool
	// MaxC and MinC are air_temperature_max and air_temperature_min in degrees Celsius;
	// HasExtremes says whether the block gave them.
	MaxC, MinC  float64
	HasExtremes bool
}

// Present answers whether the step carries this block.
func (p Period) Present() bool { return p.Hours > 0 }

// length answers the span the block covers.
func (p Period) length() time.Duration { return time.Duration(p.Hours) * time.Hour }

// Step is one entry of the forecast's timeseries.
type Step struct {
	// Time is the step's instant.
	Time time.Time
	// AirC is the air temperature at Time in degrees Celsius.
	AirC float64
	// WindMS is the wind speed at Time in metres a second; WindFrom the direction it blows from, in
	// degrees clockwise from north (FR-410).
	WindMS, WindFrom float64
	// Next1, Next6 and Next12 are its next_1_hours, next_6_hours and next_12_hours blocks.
	Next1, Next6, Next12 Period
}

// periods answers the blocks the step carries, shortest first.
func (s Step) periods() []Period {
	var present []Period
	for _, each := range []Period{s.Next1, s.Next6, s.Next12} {
		if each.Present() {
			present = append(present, each)
		}
	}
	return present
}

// Forecast is one answer for one city: its steps in time order.
type Forecast struct {
	steps []Step
}

// New answers the forecast holding steps, put in time order whatever order they arrived in. The
// steps are copied, so the caller's slice can change without changing the forecast.
func New(steps []Step) Forecast {
	ordered := slices.Clone(steps)
	slices.SortStableFunc(ordered, func(a, b Step) int { return a.Time.Compare(b.Time) })
	return Forecast{steps: ordered}
}

// Steps answers a copy of the forecast's steps in time order.
func (f Forecast) Steps() []Step { return slices.Clone(f.steps) }

// end answers the instant the forecast stops covering: the end of its last step's longest block,
// else that step's own instant. Asked only once a current step has been found, so there is always a
// last step.
func (f Forecast) end() time.Time {
	last := f.steps[len(f.steps)-1]
	present := last.periods()
	if len(present) == 0 {
		return last.Time
	}
	return last.Time.Add(present[len(present)-1].length())
}

// Date is a calendar date with no time of day and no zone: a city's local date.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// DateOf answers the date instant falls on in location.
func DateOf(instant time.Time, location *time.Location) Date {
	local := instant.In(location)
	return Date{Year: local.Year(), Month: local.Month(), Day: local.Day()}
}

// AddDays answers the date days after d; before it when days is negative.
func (d Date) AddDays(days int) Date {
	moved := time.Date(d.Year, d.Month, d.Day+days, 0, 0, 0, 0, time.UTC)
	return Date{Year: moved.Year(), Month: moved.Month(), Day: moved.Day()}
}

// Weekday answers the day of the week d falls on.
func (d Date) Weekday() time.Weekday {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).Weekday()
}

// noon answers local noon of d in location: what a day's symbol is chosen nearest to (FR-409).
func (d Date) noon(location *time.Location) time.Time {
	const noonHour = 12
	return time.Date(d.Year, d.Month, d.Day, noonHour, 0, 0, 0, location)
}

// midpoint answers the instant half way through a block starting at start (FR-405).
func midpoint(start time.Time, block Period) time.Time {
	return start.Add(block.length() / 2)
}
