package cache

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/petrichor"
)

// at answers the instant hours after 2026-10-05T06:00Z.
func at(hours int) time.Time {
	return time.Date(2026, time.October, 5, 6, 0, 0, 0, time.UTC).Add(time.Duration(hours) * time.Hour)
}

// sample answers an entry using every kind of value: a block with rain and extremes, one with
// neither, a step with no blocks, a step out of order and a dry spell.
func sample() application.Cached {
	steps := []forecast.Step{
		{Time: at(1), AirC: 11.5, WindMS: 2, WindFrom: 270},
		{
			Time: at(0), AirC: 10.2, WindMS: 3.4, WindFrom: 225.5,
			Next1:  forecast.Period{Hours: forecast.Next1Hours, Symbol: "rain", RainMM: 0.3, HasRain: true},
			Next6:  forecast.Period{Hours: forecast.Next6Hours, Symbol: "cloudy", RainMM: 0, HasRain: true, MaxC: 14, MinC: 9.5, HasExtremes: true},
			Next12: forecast.Period{Hours: forecast.Next12Hours, Symbol: "fair_day"},
		},
	}
	return application.Cached{
		Forecast: forecast.New(steps), Expires: at(1), LastModified: "Mon, 05 Oct 2026 07:30:04 GMT",
		Fetched: at(0), Spell: petrichor.Spell{Since: at(-80), Dry: true},
	}
}

// assertSame fails unless got holds what want does, step by step.
func assertSame(t *testing.T, got, want application.Cached) {
	t.Helper()
	gotSteps, wantSteps := got.Forecast.Steps(), want.Forecast.Steps()
	if len(gotSteps)+len(wantSteps) > 0 && !reflect.DeepEqual(gotSteps, wantSteps) {
		t.Errorf("steps\n got %+v\nwant %+v", got.Forecast.Steps(), want.Forecast.Steps())
	}
	got.Forecast, want.Forecast = forecast.Forecast{}, forecast.Forecast{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("entry\n got %+v\nwant %+v", got, want)
	}
}

// Every value written is read back as it was, whether a figure was given or not.
func TestAnEntryReadsBackAsWritten(t *testing.T) {
	t.Parallel()
	for name, cached := range map[string]application.Cached{
		"dry spell": sample(),
		"no spell":  {Forecast: forecast.New(nil), Expires: at(1), Fetched: at(0)},
	} {
		body, err := encode(cached)
		if err != nil {
			t.Fatal(err)
		}
		read, err := decode(body)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		assertSame(t, read, cached)
	}
}

// unwritable answers a forecast holding a figure JSON cannot hold.
func unwritable() forecast.Forecast {
	return forecast.New([]forecast.Step{{Time: at(0), AirC: math.NaN()}})
}

// A figure JSON cannot hold refuses the write rather than writing something else.
func TestAFigureJSONCannotHoldIsRefused(t *testing.T) {
	t.Parallel()
	cached := sample()
	cached.Forecast = unwritable()
	if _, err := encode(cached); err == nil || !strings.Contains(err.Error(), "encoding the forecast") {
		t.Errorf("answered %v; want the write refused", err)
	}
}

// Anything this version did not write is refused with the reason.
func TestWhatThisVersionDidNotWriteIsRefused(t *testing.T) {
	t.Parallel()
	const times = `"expires":"2026-10-05T07:00:00Z","fetched":"2026-10-05T06:00:00Z"`
	for name, tc := range map[string]struct {
		body string
		want error
	}{
		"not JSON":              {body: `{"version":`},
		"another version":       {body: `{"version":2,` + times + `}`, want: errVersion},
		"no expires":            {body: `{"version":1,"fetched":"2026-10-05T06:00:00Z"}`, want: errMissing},
		"no fetched":            {body: `{"version":1,"expires":"2026-10-05T07:00:00Z"}`, want: errMissing},
		"a spell with no start": {body: `{"version":1,` + times + `,"spell":{"dry":true}}`, want: errMissing},
		"a step with no time":   {body: `{"version":1,` + times + `,"steps":[{"airC":1}]}`, want: errMissing},
		"half the extremes": {
			body: `{"version":1,` + times + `,"steps":[{"time":"2026-10-05T06:00:00Z","next6":{"symbol":"fog","maxC":3}}]}`,
			want: errExtremes,
		},
	} {
		_, err := decode([]byte(tc.body))
		if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) {
			t.Errorf("%s: answered %v; want %v", name, err, tc.want)
		}
	}
}
