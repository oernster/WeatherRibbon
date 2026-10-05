package application

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/place"
)

// added answers a started rig holding the given places as cities, in order.
func added(t *testing.T, ids ...int) *rig {
	t.Helper()
	r := newRig()
	if err := r.service.Start(); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if _, err := r.service.AddCity(id); err != nil {
			t.Fatal(err)
		}
	}
	return r
}

// FR-303: a forecast cached at 07:36:04Z expiring 08:00:28Z is not asked for again before 08:00:28Z.
func TestNothingIsRequestedBeforeExpires(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	expires := time.Date(2026, time.October, 5, 8, 0, 28, 0, time.UTC)
	r.forecasts.answers = append(r.forecasts.answers, answer(expires, "Mon, 05 Oct 2026 07:30:00 GMT"))
	if err := r.service.Refresh(context.Background()); err != nil || len(r.forecasts.asked) != 1 {
		t.Fatalf("first refresh asked %d times, %v", len(r.forecasts.asked), err)
	}
	r.clock.now = expires.Add(-time.Second)
	_ = r.service.RefreshNow(context.Background())
	if len(r.forecasts.asked) != 1 {
		t.Fatalf("asked again before Expires: %d requests", len(r.forecasts.asked))
	}
	if got := r.service.NextDue(); !got.Equal(expires) {
		t.Errorf("next due %v; want %v", got, expires)
	}
	r.clock.now = expires
	r.forecasts.answers = append(r.forecasts.answers, answer(expires.Add(time.Hour), ""))
	_ = r.service.Refresh(context.Background())
	if len(r.forecasts.asked) != 2 {
		t.Errorf("not asked once Expires came: %d requests", len(r.forecasts.asked))
	}
}

// FR-302, FR-304: a request carries rounded coordinates; once something is cached it carries its
// Last-Modified too. A 304 keeps the forecast and takes the new Expires.
func TestANotModifiedAnswerKeepsTheForecast(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	first := r.clock.now.Add(24 * time.Minute)
	r.forecasts.answers = append(r.forecasts.answers, answer(first, "Mon, 05 Oct 2026 07:30:00 GMT"))
	_ = r.service.Refresh(context.Background())
	r.clock.now = first
	later := first.Add(30 * time.Minute)
	r.forecasts.answers = append(r.forecasts.answers, func(Request) (Answer, error) {
		return Answer{NotModified: true, Expires: later}, nil
	})
	_ = r.service.Refresh(context.Background())
	asked := r.forecasts.asked
	if asked[0].Latitude != 51.5085 || asked[0].Longitude != -0.1257 || asked[0].LastModified != "" {
		t.Errorf("first request %+v; want rounded coordinates and no Last-Modified", asked[0])
	}
	if asked[1].LastModified != "Mon, 05 Oct 2026 07:30:00 GMT" {
		t.Errorf("second request %+v; want the cached Last-Modified", asked[1])
	}
	kept := r.cache.entries["city-1"]
	if !kept.Expires.Equal(later) || kept.LastModified != "Mon, 05 Oct 2026 07:30:00 GMT" || !kept.Fetched.Equal(first) {
		t.Errorf("kept %+v; want the new Expires with the old Last-Modified, fetched at %v", kept, first)
	}
}

// FR-305: a failed request leaves the forecast already held in place, still shown, with no problem.
func TestAFailedRequestKeepsTheForecast(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	fetched := r.clock.now
	expires := fetched.Add(time.Hour)
	r.forecasts.answers = append(r.forecasts.answers, func(Request) (Answer, error) {
		return Answer{Forecast: hourlyAround(fetched, 14.4, 0), Expires: expires}, nil
	})
	_ = r.service.Refresh(context.Background())
	r.clock.now = expires
	r.forecasts.answers = append(r.forecasts.answers, failure(errPlanted))
	_ = r.service.Refresh(context.Background())
	if len(r.forecasts.asked) != 2 {
		t.Fatalf("asked %d times; want the success and the failure", len(r.forecasts.asked))
	}
	if kept := r.cache.entries["city-1"]; !kept.Fetched.Equal(fetched) || !kept.Expires.Equal(expires) {
		t.Errorf("kept %+v; want the forecast fetched at %v", kept, fetched)
	}
	if cell := r.service.Snapshot().Cells[0]; cell.Problem != "" || cell.Temperature != 14 {
		t.Errorf("cell %+v; want the held forecast shown", cell)
	}
}

// FR-309: Refresh now asks for the city whose forecast has expired and the one backing off, never
// the one still within its Expires.
func TestRefreshNowAsksOnlyTheExpired(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID, tokyo.GeoNamesID, paris.GeoNamesID)
	start := r.clock.now
	parisAsked, tokyoAsked := place.RoundDegrees(paris.Latitude), place.RoundDegrees(tokyo.Latitude)
	byCity := func(request Request) (Answer, error) {
		switch request.Latitude {
		case parisAsked:
			return Answer{}, errPlanted
		case tokyoAsked:
			return Answer{Expires: start.Add(5 * time.Minute)}, nil
		}
		return Answer{Expires: start.Add(time.Hour)}, nil
	}
	for range 5 {
		r.forecasts.answers = append(r.forecasts.answers, byCity)
	}
	_ = r.service.Refresh(context.Background())
	r.clock.now = start.Add(5 * time.Minute)
	_ = r.service.RefreshNow(context.Background())
	var again []float64
	for _, each := range r.forecasts.asked[3:] {
		again = append(again, each.Latitude)
	}
	slices.Sort(again)
	if want := []float64{tokyoAsked, parisAsked}; !slices.Equal(again, want) {
		t.Errorf("Refresh now asked for latitudes %v; want Tokyo and Paris alone", again)
	}
}

// FR-306: failures back off from 10 minutes; a success resets it.
func TestFailuresBackOff(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	start := r.clock.now
	r.forecasts.answers = append(r.forecasts.answers, failure(errPlanted))
	_ = r.service.Refresh(context.Background())
	if got := r.service.NextDue(); !got.Equal(start.Add(10 * time.Minute)) {
		t.Fatalf("after one failure next due %v; want 10 minutes on", got)
	}
	r.clock.now = start.Add(5 * time.Minute)
	_ = r.service.Refresh(context.Background())
	if len(r.forecasts.asked) != 1 {
		t.Fatal("asked again while backing off")
	}
	r.forecasts.answers = append(r.forecasts.answers, answer(r.clock.now.Add(time.Hour), ""))
	_ = r.service.RefreshNow(context.Background())
	if len(r.forecasts.asked) != 2 {
		t.Fatal("Refresh now did not ask a city backing off (FR-309)")
	}
	if state := r.service.weather["city-1"]; state.failures != 0 || !state.retryAt.IsZero() {
		t.Errorf("a success did not reset the back-off: %+v", state)
	}
}

// FR-307: a refusal stops every request in this run and stands as a notice.
func TestARefusalStopsEveryRequest(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID, tokyo.GeoNamesID)
	r.forecasts.answers = append(r.forecasts.answers, failure(ErrRefused))
	_ = r.service.Refresh(context.Background())
	_ = r.service.RefreshNow(context.Background())
	if len(r.forecasts.asked) != 1 || r.service.weather["city-2"].asking {
		t.Fatalf("asked %d times; want the refused request alone, Tokyo released", len(r.forecasts.asked))
	}
	r.service.DismissNotices()
	r.service.mutex.Lock()
	notices := r.service.notices()
	r.service.mutex.Unlock()
	if !slices.Contains(notices, refusedNotice) {
		t.Errorf("notices %v; want the refusal to stand", notices)
	}
	if !r.service.NextDue().IsZero() {
		t.Error("after a refusal nothing falls due")
	}
}

// NFR-S-2: requests leave at least a second apart, across refreshes too.
func TestRequestsAreSpacedAtLeastASecondApart(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID, tokyo.GeoNamesID, paris.GeoNamesID)
	for range 3 {
		r.forecasts.answers = append(r.forecasts.answers, answer(r.clock.now.Add(time.Hour), ""))
	}
	_ = r.service.Refresh(context.Background())
	if want := []time.Duration{time.Second, time.Second}; !slices.Equal(r.pacer.waits, want) {
		t.Errorf("waits %v; want %v", r.pacer.waits, want)
	}
}

// FR-310: a city already being asked for is not claimed again.
func TestASecondRefreshWaitsForTheFirst(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	if claimed := r.service.claim(false); len(claimed) != 1 {
		t.Fatalf("claimed %d", len(claimed))
	}
	if again := r.service.claim(true); len(again) != 0 {
		t.Errorf("claimed a city already being asked for: %+v", again)
	}
	r.service.release("city-1")
	r.service.release("gone")
	if again := r.service.claim(false); len(again) != 1 {
		t.Error("a released city is not claimed again")
	}
}

// A context that ends releases the cities not yet asked for, answering its error.
func TestAnEndedRefreshReleasesItsCities(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID, tokyo.GeoNamesID)
	r.forecasts.answers = append(r.forecasts.answers, answer(r.clock.now.Add(time.Hour), ""))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := r.service.Refresh(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("refresh answered %v; want the context's error", err)
	}
	if len(r.forecasts.asked) != 0 || r.service.weather["city-1"].asking || r.service.weather["city-2"].asking {
		t.Errorf("asked %d; cities left claimed", len(r.forecasts.asked))
	}
}
