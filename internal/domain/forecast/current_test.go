package forecast

import (
	"testing"
	"time"
	_ "time/tzdata"
)

// at answers the instant written in RFC 3339, failing the test when it is not one.
func at(t *testing.T, text string) time.Time {
	t.Helper()
	instant, err := time.Parse(time.RFC3339, text)
	if err != nil {
		t.Fatalf("%q is not an instant: %v", text, err)
	}
	return instant
}

// hourly answers a 1-hour block of symbol giving rain.
func hourly(symbol string, rain float64) Period {
	return Period{Hours: 1, Symbol: symbol, RainMM: rain, HasRain: true}
}

// sixHourly answers a 6-hour block of symbol giving rain and extremes.
func sixHourly(symbol string, rain, high, low float64) Period {
	return Period{Hours: 6, Symbol: symbol, RainMM: rain, HasRain: true, MaxC: high, MinC: low, HasExtremes: true}
}

// FR-402: at 07:36Z with steps at 07:00Z and 08:00Z, the 07:00Z step is current.
func TestTheCurrentStepIsTheLatestNotAfterNow(t *testing.T) {
	t.Parallel()
	f := New([]Step{
		{Time: at(t, "2026-10-05T08:00:00Z"), AirC: 15, Next1: hourly("cloudy", 0)},
		{Time: at(t, "2026-10-05T07:00:00Z"), AirC: 14.4, Next1: hourly("partlycloudy_day", 0.3)},
	})
	got, ok := f.CurrentAt(at(t, "2026-10-05T07:36:00Z"))
	if !ok || got.AirC != 14.4 || got.Symbol != "partlycloudy_day" || got.RainMM != 0.3 || !got.HasRain {
		t.Fatalf("CurrentAt = %+v, %v; want the 07:00Z step", got, ok)
	}
	exact, ok := f.CurrentAt(at(t, "2026-10-05T08:00:00Z"))
	if !ok || exact.AirC != 15 {
		t.Fatalf("a step at now itself is current: got %+v, %v", exact, ok)
	}
}

// FR-402: the symbol is the 1-hour block's, else the 6-hour block's, else none.
func TestTheCurrentSymbolPrefersTheShortestBlock(t *testing.T) {
	t.Parallel()
	now := at(t, "2026-10-05T12:00:00Z")
	for name, step := range map[string]struct {
		step Step
		want string
	}{
		"both":        {Step{Time: now, Next1: hourly("rain", 1), Next6: sixHourly("cloudy", 2, 9, 4)}, "rain"},
		"six alone":   {Step{Time: now, Next6: sixHourly("fog", 0, 9, 4)}, "fog"},
		"twelve":      {Step{Time: now, Next12: Period{Hours: 12, Symbol: "clearsky_day"}}, ""},
		"no blocks":   {Step{Time: now}, ""},
		"one, no six": {Step{Time: now, Next1: hourly("snow", 1)}, "snow"},
	} {
		got, ok := New([]Step{step.step}).CurrentAt(now)
		if !ok || got.Symbol != step.want {
			t.Errorf("%s: symbol %q, %v; want %q", name, got.Symbol, ok, step.want)
		}
	}
}

// FR-403: before the first step, past the last block or with no steps, nothing is current.
func TestAForecastThatHasRunOutShowsNothingAsCurrent(t *testing.T) {
	t.Parallel()
	f := New([]Step{
		{Time: at(t, "2026-10-05T07:00:00Z"), Next1: hourly("cloudy", 0)},
		{Time: at(t, "2026-10-05T08:00:00Z"), Next1: hourly("cloudy", 0), Next6: sixHourly("cloudy", 0, 9, 4)},
	})
	for name, now := range map[string]string{
		"before the first step":    "2026-10-05T06:59:00Z",
		"past the last 6-hour end": "2026-10-05T14:01:00Z",
	} {
		if got, ok := f.CurrentAt(at(t, now)); ok {
			t.Errorf("%s: CurrentAt = %+v; want nothing current", name, got)
		}
	}
	if _, ok := f.CurrentAt(at(t, "2026-10-05T14:00:00Z")); !ok {
		t.Error("the last block's own end is still covered")
	}
	if _, ok := New(nil).CurrentAt(at(t, "2026-10-05T07:00:00Z")); ok {
		t.Error("an empty forecast has nothing current")
	}
	bare := New([]Step{{Time: at(t, "2026-10-05T07:00:00Z")}})
	if _, ok := bare.CurrentAt(at(t, "2026-10-05T07:00:01Z")); ok {
		t.Error("a last step with no blocks covers its own instant alone")
	}
}

// New orders the steps and keeps its own copy; Steps answers a copy.
func TestAForecastKeepsItsOwnOrderedSteps(t *testing.T) {
	t.Parallel()
	given := []Step{{Time: at(t, "2026-10-05T09:00:00Z")}, {Time: at(t, "2026-10-05T08:00:00Z")}}
	f := New(given)
	given[0].AirC = 99
	steps := f.Steps()
	if !steps[0].Time.Equal(at(t, "2026-10-05T08:00:00Z")) || steps[1].AirC == 99 {
		t.Fatalf("steps %+v; want time order, untouched by the caller", steps)
	}
	steps[0].AirC = 42
	if f.Steps()[0].AirC == 42 {
		t.Fatal("Steps handed out the forecast's own slice")
	}
}

// Dates move across months and years and know their weekday.
func TestDatesMoveByWholeDays(t *testing.T) {
	t.Parallel()
	last := Date{Year: 2026, Month: time.December, Day: 31}
	if got := last.AddDays(1); got != (Date{Year: 2027, Month: time.January, Day: 1}) {
		t.Errorf("31 December plus a day is %+v", got)
	}
	if got := last.AddDays(-31); got != (Date{Year: 2026, Month: time.November, Day: 30}) {
		t.Errorf("31 December less 31 days is %+v", got)
	}
	if got := (Date{Year: 2026, Month: time.October, Day: 5}).Weekday(); got != time.Monday {
		t.Errorf("5 October 2026 is a %v; want Monday", got)
	}
}
