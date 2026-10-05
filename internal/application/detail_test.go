package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// londonSun is London's sunrise and sunset on 5 October 2026 as measured (section 2.3): 07:07 and
// 18:29 BST.
func londonSun() (Sun, error) {
	return Sun{
		Rise: time.Date(2026, time.October, 5, 6, 7, 0, 0, time.UTC),
		Set:  time.Date(2026, time.October, 5, 17, 29, 0, 0, time.UTC),
	}, nil
}

// withForecast answers a started rig with London holding a fresh forecast.
func withForecast(t *testing.T) *rig {
	t.Helper()
	r := added(t, london.GeoNamesID)
	hold(r, "city-1", hourlyAround(r.clock.now, 14.4, 0.2), r.clock.now)
	return r
}

// FR-410: the panel shows 24 hours from the current one in the city's local time and the chosen units.
func TestTheDetailShowsTheHours(t *testing.T) {
	t.Parallel()
	r := withForecast(t)
	r.sun.answers = append(r.sun.answers, londonSun)
	detail, err := r.service.OpenDetail(context.Background(), "city-1")
	if err != nil || detail.Problem != "" || len(detail.Hours) != forecast.DetailHours {
		t.Fatalf("detail %+v, %v", detail, err)
	}
	first := detail.Hours[0]
	if first.Time != "08:00" || first.Temperature != 14 || first.Rain != 0.2 || first.Symbol != "rain" {
		t.Errorf("first hour %+v; want the 07:00Z step at 08:00 BST", first)
	}
	if detail.Label != "London" || detail.Place != "London, England, United Kingdom" {
		t.Errorf("detail %+v", detail)
	}
}

// FR-411: sunrise and sunset are asked once per local date, then held; a new date asks again.
func TestSunTimesAreAskedOncePerDay(t *testing.T) {
	t.Parallel()
	r := withForecast(t)
	r.sun.answers = append(r.sun.answers, londonSun, londonSun)
	first, _ := r.service.OpenDetail(context.Background(), "city-1")
	second, _ := r.service.OpenDetail(context.Background(), "city-1")
	if first.Sunrise != "07:07" || first.Sunset != "18:29" || second.Sunrise != "07:07" || len(r.sun.asked) != 1 {
		t.Fatalf("first %q %q, second %q, asked %d; want 07:07 and 18:29 asked once", first.Sunrise, first.Sunset, second.Sunrise, len(r.sun.asked))
	}
	r.clock.now = r.clock.now.Add(24 * time.Hour)
	hold(r, "city-1", hourlyAround(r.clock.now, 10, 0), r.clock.now)
	_, _ = r.service.OpenDetail(context.Background(), "city-1")
	if len(r.sun.asked) != 2 || r.sun.asked[1] != (forecast.Date{Year: 2026, Month: time.October, Day: 6}) {
		t.Errorf("asked %v; want the next date asked for", r.sun.asked)
	}
}

// FR-411: a failed request leaves the times out; the next opening asks again. A day without a sunset
// shows none.
func TestAFailedSunRequestIsAskedAgain(t *testing.T) {
	t.Parallel()
	r := withForecast(t)
	r.sun.answers = append(r.sun.answers, func() (Sun, error) { return Sun{}, errPlanted },
		func() (Sun, error) { return Sun{Rise: time.Date(2026, time.October, 5, 6, 7, 0, 0, time.UTC)}, nil })
	failed, err := r.service.OpenDetail(context.Background(), "city-1")
	if err != nil || failed.Sunrise != "" || len(failed.Hours) == 0 {
		t.Fatalf("after a failure %+v, %v; want the hours without times", failed, err)
	}
	again, _ := r.service.OpenDetail(context.Background(), "city-1")
	if len(r.sun.asked) != 2 || again.Sunrise != "07:07" || again.Sunset != "" {
		t.Errorf("asked %d, sunrise %q, sunset %q", len(r.sun.asked), again.Sunrise, again.Sunset)
	}
}

// FR-307: a refusal of the sunrise request stops every request in the run.
func TestARefusedSunRequestStopsEveryRequest(t *testing.T) {
	t.Parallel()
	r := withForecast(t)
	r.sun.answers = append(r.sun.answers, func() (Sun, error) { return Sun{}, ErrRefused })
	_, _ = r.service.OpenDetail(context.Background(), "city-1")
	_, _ = r.service.OpenDetail(context.Background(), "city-1")
	if len(r.sun.asked) != 1 || !r.service.refused {
		t.Errorf("asked %d, refused %v; want one request and the run refused", len(r.sun.asked), r.service.refused)
	}
}

// FR-410: a city with no forecast, no place or an unknown id is answered without asking anything.
func TestADetailThatCannotShowTheHoursSaysWhy(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	if detail, err := r.service.OpenDetail(context.Background(), "city-1"); err != nil || detail.Problem != forecastUnavailable {
		t.Errorf("no forecast: %+v, %v", detail, err)
	}
	r.service.current = r.service.current.WithCityAdded(settings.City{ID: "lost", GeoNamesID: 1, Label: "Atlantis"})
	if detail, _ := r.service.OpenDetail(context.Background(), "lost"); detail.Problem != placeNotFound || detail.Place != "" {
		t.Errorf("lost place: %+v", detail)
	}
	if _, err := r.service.OpenDetail(context.Background(), "nobody"); !errors.Is(err, settings.ErrNoSuchCity) {
		t.Errorf("unknown id answered %v", err)
	}
	if len(r.sun.asked) != 0 {
		t.Errorf("asked %d sunrise requests; want none", len(r.sun.asked))
	}
}

// A city removed while its sunrise request is out keeps nothing; an ended context asks nothing.
func TestARemovedCityKeepsNoSunTimes(t *testing.T) {
	t.Parallel()
	r := withForecast(t)
	r.sun.answers = append(r.sun.answers, func() (Sun, error) {
		delete(r.service.weather, "city-1")
		return londonSun()
	})
	if detail, err := r.service.OpenDetail(context.Background(), "city-1"); err != nil || detail.Sunrise != "07:07" {
		t.Errorf("detail %+v, %v", detail, err)
	}
	if _, found := r.service.weather["city-1"]; found {
		t.Error("times were kept for a removed city")
	}
	ended, cancel := context.WithCancel(context.Background())
	cancel()
	hold(r, "city-1", hourlyAround(r.clock.now, 10, 0), r.clock.now)
	if detail, _ := r.service.OpenDetail(ended, "city-1"); detail.Sunrise != "" || len(r.sun.asked) != 1 {
		t.Errorf("an ended context asked: %d requests", len(r.sun.asked))
	}
}
