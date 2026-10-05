package forecast

import "time"

// OutlookDays is how many days after today a cell shows (FR-406, OQ-3).
const OutlookDays = 3

// Day is what a cell shows for one local date (FR-404, FR-406 to FR-409).
type Day struct {
	// Date is the city's local date this day covers.
	Date Date
	// HighC and LowC are the day's highest and lowest temperatures in degrees Celsius (FR-407).
	HighC, LowC float64
	// RainMM is the day's rain in millimetres (FR-408).
	RainMM float64
	// Symbol is the symbol code of the 6-hour block whose midpoint is nearest local noon; empty when
	// no 6-hour block falls in the day (FR-409).
	Symbol string
	// Known says whether the forecast reaches the day at all; a day it does not reach shows nothing.
	Known bool
}

// TodayAt answers today in location at now: its high, low and rain over the forecast from the
// current step to the end of the local date (FR-404, OQ-5). It answers false where CurrentAt does,
// since a forecast that has run out describes no part of today (FR-403).
func (f Forecast) TodayAt(now time.Time, location *time.Location) (Day, bool) {
	if _, ok := f.CurrentAt(now); !ok {
		return Day{}, false
	}
	remaining := f.steps[f.currentIndex(now):]
	return summarise(remaining, DateOf(now, location), location), true
}

// OutlookAt answers the OutlookDays days after today in location at now, each summarised over the
// whole forecast (FR-406). A day the forecast does not reach is answered with Known false.
func (f Forecast) OutlookAt(now time.Time, location *time.Location) []Day {
	today := DateOf(now, location)
	days := make([]Day, 0, OutlookDays)
	for ahead := 1; ahead <= OutlookDays; ahead++ {
		days = append(days, summarise(f.steps, today.AddDays(ahead), location))
	}
	return days
}

// summarise answers date as steps describe it. A step counts for the date its instant falls on; a
// block for the date its midpoint falls on (FR-405).
func summarise(steps []Step, date Date, location *time.Location) Day {
	day := Day{Date: date}
	for _, each := range steps {
		if DateOf(each.Time, location) == date {
			day.include(each.AirC, each.AirC)
		}
	}
	for _, each := range blockSpans(steps) {
		if each.block.HasExtremes && DateOf(midpoint(each.start, each.block), location) == date {
			day.include(each.block.MaxC, each.block.MinC)
		}
	}
	for _, each := range rainSpans(steps) {
		if DateOf(midpoint(each.start, each.block), location) == date {
			day.RainMM += each.block.RainMM
		}
	}
	day.Symbol = symbolNearestNoon(steps, date, location)
	return day
}

// include widens the day's high and low to take in high and low (FR-407).
func (d *Day) include(high, low float64) {
	if !d.Known {
		d.HighC, d.LowC, d.Known = high, low, true
		return
	}
	d.HighC = max(d.HighC, high)
	d.LowC = min(d.LowC, low)
}

// symbolNearestNoon answers the symbol of the 6-hour block falling in date whose midpoint is nearest
// local noon; the earlier of two equally near. Empty when no 6-hour block falls in the date (FR-409).
func symbolNearestNoon(steps []Step, date Date, location *time.Location) string {
	noon := date.noon(location)
	symbol := ""
	nearest := time.Duration(-1)
	for _, each := range blockSpans(steps) {
		// A day's symbol is chosen from the 6-hour blocks alone (FR-409).
		if each.block.Hours != Next6Hours {
			continue
		}
		middle := midpoint(each.start, each.block)
		if DateOf(middle, location) != date {
			continue
		}
		distance := middle.Sub(noon).Abs()
		if nearest < 0 || distance < nearest {
			symbol, nearest = each.block.Symbol, distance
		}
	}
	return symbol
}
