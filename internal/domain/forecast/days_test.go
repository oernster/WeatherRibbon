package forecast

import (
	"testing"
	"time"
)

// zone answers the named zone, failing the test when it cannot be loaded.
func zone(t *testing.T, name string) *time.Location {
	t.Helper()
	location, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("loading %s: %v", name, err)
	}
	return location
}

// FR-404: at 22:30 London time with steps to midnight, today's high and low come from the 22:00 and
// 23:00 steps alone.
func TestTodayCoversNowToMidnightLocal(t *testing.T) {
	t.Parallel()
	london := zone(t, "Europe/London")
	f := New([]Step{
		{Time: at(t, "2026-10-05T20:00:00Z"), AirC: 5, Next1: hourly("rain", 3)},
		{Time: at(t, "2026-10-05T21:00:00Z"), AirC: 9, Next1: hourly("rain", 0.5)},
		{Time: at(t, "2026-10-05T22:00:00Z"), AirC: 7, Next1: hourly("rain", 0.25)},
		{Time: at(t, "2026-10-05T23:00:00Z"), AirC: 1, Next1: hourly("cloudy", 0)},
	})
	today, ok := f.TodayAt(at(t, "2026-10-05T21:30:00Z"), london)
	if !ok || !today.Known || today.HighC != 9 || today.LowC != 7 || today.RainMM != 0.75 {
		t.Fatalf("today = %+v, %v; want high 9, low 7, rain 0.75 from 22:00 and 23:00 alone", today, ok)
	}
	if today.Date != (Date{Year: 2026, Month: time.October, Day: 5}) {
		t.Errorf("today is %+v", today.Date)
	}
	if _, ok := f.TodayAt(at(t, "2026-10-05T19:00:00Z"), london); ok {
		t.Error("a forecast not yet begun describes no part of today")
	}
}

// FR-405: a block counts for the local date of its midpoint.
func TestAPeriodBelongsToTheDayOfItsMidpoint(t *testing.T) {
	t.Parallel()
	block := sixHourly("rain", 2, 20, 10)
	tokyo := time.FixedZone("Tokyo", 9*3600)
	start := at(t, "2026-10-05T18:00:00Z")
	if got := DateOf(midpoint(start, block), tokyo); got != (Date{Year: 2026, Month: time.October, Day: 6}) {
		t.Errorf("Tokyo's 18:00Z to 00:00Z block counts for %+v; want 6 October", got)
	}
	kathmandu := time.FixedZone("Kathmandu", 5*3600+45*60)
	morning := at(t, "2026-10-05T06:00:00Z")
	if got := DateOf(midpoint(morning, block), kathmandu); got != DateOf(morning, kathmandu) {
		t.Errorf("Kathmandu's 06:00Z to 12:00Z block counts for %+v; want its own date", got)
	}
}

// FR-406: on a Monday the outlook reads Tuesday, Wednesday and Thursday; a day the forecast does not
// reach is not known.
func TestTheOutlookCoversTheFollowingDays(t *testing.T) {
	t.Parallel()
	var steps []Step
	for day := 5; day <= 7; day++ {
		steps = append(steps, Step{Time: time.Date(2026, time.October, day, 12, 0, 0, 0, time.UTC), AirC: float64(day)})
	}
	outlook := New(steps).OutlookAt(at(t, "2026-10-05T09:00:00Z"), time.UTC)
	want := []time.Weekday{time.Tuesday, time.Wednesday, time.Thursday}
	if len(outlook) != OutlookDays {
		t.Fatalf("outlook has %d days; want %d", len(outlook), OutlookDays)
	}
	for index, each := range outlook {
		if each.Date.Weekday() != want[index] {
			t.Errorf("day %d is a %v; want %v", index, each.Date.Weekday(), want[index])
		}
	}
	if !outlook[0].Known || !outlook[1].Known || outlook[2].Known {
		t.Errorf("known days %v %v %v; want the forecast's two and not the third",
			outlook[0].Known, outlook[1].Known, outlook[2].Known)
	}
}

// FR-407: a day's high and low take every step's temperature and every block's extremes.
func TestADaysHighAndLowTakeEveryValue(t *testing.T) {
	t.Parallel()
	f := New([]Step{
		{Time: at(t, "2026-10-06T06:00:00Z"), AirC: 8, Next6: sixHourly("cloudy", 0, 13, 6)},
		{Time: at(t, "2026-10-06T12:00:00Z"), AirC: 12, Next6: sixHourly("cloudy", 0, 11, 9)},
	})
	day := f.OutlookAt(at(t, "2026-10-05T12:00:00Z"), time.UTC)[0]
	if day.HighC != 13 || day.LowC != 6 {
		t.Fatalf("high %v, low %v; want 13 from a block's maximum and 6 from its minimum", day.HighC, day.LowC)
	}
}

// FR-408: hourly steps to 00:00Z then 6-hourly; a day spanning the change counts each hour once.
func TestRainIsNeverCountedTwice(t *testing.T) {
	t.Parallel()
	var steps []Step
	for hour := 18; hour <= 23; hour++ {
		steps = append(steps, Step{
			Time:  time.Date(2026, time.October, 7, hour, 0, 0, 0, time.UTC),
			Next1: hourly("rain", 1), Next6: sixHourly("rain", 6, 10, 5),
		})
	}
	// The last hourly step, at 00:00Z, carries both blocks; the 6-hourly steps after it the 6-hour
	// block alone, as MET Norway's answers do.
	steps = append(steps, Step{
		Time:  time.Date(2026, time.October, 8, 0, 0, 0, 0, time.UTC),
		Next1: hourly("rain", 1), Next6: sixHourly("rain", 6, 10, 5),
	})
	for hour := 6; hour <= 18; hour += 6 {
		steps = append(steps, Step{
			Time:  time.Date(2026, time.October, 8, hour, 0, 0, 0, time.UTC),
			Next6: sixHourly("rain", 6, 10, 5),
		})
	}
	outlook := New(steps).OutlookAt(at(t, "2026-10-06T12:00:00Z"), time.UTC)
	if outlook[0].RainMM != 6 || outlook[1].RainMM != 24 {
		t.Fatalf("rain %v on the 7th and %v on the 8th; want 6 from six hours and 24 from four blocks",
			outlook[0].RainMM, outlook[1].RainMM)
	}
}

// FR-408: a block starting inside one already taken is passed over; a step giving no rain is skipped.
func TestAnOverlappingBlockIsPassedOver(t *testing.T) {
	t.Parallel()
	spans := rainSpans([]Step{
		{Time: at(t, "2026-10-08T00:00:00Z"), Next6: sixHourly("rain", 6, 9, 4)},
		{Time: at(t, "2026-10-08T01:00:00Z"), Next6: sixHourly("rain", 6, 9, 4)},
		{Time: at(t, "2026-10-08T06:00:00Z"), Next12: Period{Hours: 12, Symbol: "cloudy"}},
	})
	if len(spans) != 1 || !spans[0].start.Equal(at(t, "2026-10-08T00:00:00Z")) {
		t.Fatalf("spans %+v; want the 00:00Z block alone", spans)
	}
}

// FR-409: the 6-hour block whose midpoint is nearest local noon gives the day's symbol; of two equally
// near, the earlier; none when no 6-hour block falls in the day.
func TestADaysSymbolIsTheOneNearestNoon(t *testing.T) {
	t.Parallel()
	day := Date{Year: 2026, Month: time.October, Day: 6}
	nearest := []Step{
		{Time: at(t, "2026-10-06T06:00:00Z"), Next6: sixHourly("fog", 0, 9, 4)},
		{Time: at(t, "2026-10-06T09:00:00Z"), Next1: hourly("rain", 1), Next6: sixHourly("cloudy", 0, 9, 4)},
		{Time: at(t, "2026-10-06T12:00:00Z"), Next6: sixHourly("clearsky_day", 0, 9, 4)},
		{Time: at(t, "2026-10-06T21:00:00Z"), Next6: sixHourly("clearsky_night", 0, 9, 4)},
	}
	if got := symbolNearestNoon(nearest, day, time.UTC); got != "cloudy" {
		t.Errorf("symbol %q; want the block with its midpoint at noon", got)
	}
	tied := []Step{
		{Time: at(t, "2026-10-06T08:00:00Z"), Next6: sixHourly("fog", 0, 9, 4)},
		{Time: at(t, "2026-10-06T10:00:00Z"), Next6: sixHourly("rain", 0, 9, 4)},
	}
	if got := symbolNearestNoon(tied, day, time.UTC); got != "fog" {
		t.Errorf("symbol %q; want the earlier of two equally near", got)
	}
	if got := symbolNearestNoon([]Step{{Time: at(t, "2026-10-06T09:00:00Z"), Next1: hourly("rain", 1)}}, day, time.UTC); got != "" {
		t.Errorf("symbol %q; want none without a 6-hour block", got)
	}
}
