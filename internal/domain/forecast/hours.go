package forecast

import "time"

// DetailHours is how many hourly steps the detail panel shows (FR-410).
const DetailHours = 24

// Hour is one hourly step as the detail panel shows it (FR-410).
type Hour struct {
	// Time is the step's instant.
	Time time.Time
	// AirC is the air temperature in degrees Celsius.
	AirC float64
	// Symbol and RainMM are the step's 1-hour block's.
	Symbol string
	RainMM float64
	// WindMS and WindFrom are the wind speed in metres a second and the direction it blows from.
	WindMS, WindFrom float64
}

// HoursAt answers up to DetailHours hourly steps from the current one at now (FR-410): each a step
// carrying a 1-hour block. None where nothing is current (FR-403).
func (f Forecast) HoursAt(now time.Time) []Hour {
	if _, ok := f.CurrentAt(now); !ok {
		return nil
	}
	var hours []Hour
	for _, each := range f.steps[f.currentIndex(now):] {
		if len(hours) == DetailHours {
			break
		}
		if !each.Next1.Present() {
			continue
		}
		hours = append(hours, Hour{
			Time: each.Time, AirC: each.AirC, Symbol: each.Next1.Symbol, RainMM: each.Next1.RainMM,
			WindMS: each.WindMS, WindFrom: each.WindFrom,
		})
	}
	return hours
}
