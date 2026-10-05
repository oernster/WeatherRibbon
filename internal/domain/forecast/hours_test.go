package forecast

import (
	"testing"
	"time"
)

// FR-410: the detail shows the next 24 hourly steps from the current one, each with its 1-hour
// block's symbol and rain and the step's wind; a 6-hourly step is passed over.
func TestTheDetailShowsTheNextTwentyFourHours(t *testing.T) {
	t.Parallel()
	start := at(t, "2026-10-05T07:00:00Z")
	var steps []Step
	for hour := range 30 {
		steps = append(steps, Step{
			Time: start.Add(time.Duration(hour) * time.Hour), AirC: float64(hour),
			WindMS: 5, WindFrom: 225, Next1: hourly("rain", 0.1),
		})
	}
	steps = append(steps, Step{Time: start.Add(36 * time.Hour), Next6: sixHourly("cloudy", 0, 9, 4)})
	hours := New(steps).HoursAt(at(t, "2026-10-05T07:36:00Z"))
	if len(hours) != DetailHours {
		t.Fatalf("%d hours; want %d", len(hours), DetailHours)
	}
	first, last := hours[0], hours[DetailHours-1]
	if !first.Time.Equal(start) || first.Symbol != "rain" || first.RainMM != 0.1 || first.WindMS != 5 || first.WindFrom != 225 {
		t.Errorf("first hour %+v; want the current 07:00 step", first)
	}
	if last.AirC != DetailHours-1 {
		t.Errorf("last hour %+v; want the 24th step", last)
	}
	short := New(steps[28:]).HoursAt(steps[28].Time)
	if len(short) != 2 {
		t.Errorf("%d hours near the end; want the two hourly steps left, the 6-hourly one passed over", len(short))
	}
	if New(steps).HoursAt(at(t, "2026-10-05T06:00:00Z")) != nil {
		t.Error("before the forecast begins there are no hours")
	}
}
