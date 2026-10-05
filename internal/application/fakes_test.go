package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// errPlanted is the failure a fake answers when told to fail.
var errPlanted = errors.New("planted failure")

// The places the fake city list holds.
var (
	london = place.Place{GeoNamesID: 2643743, Name: "London", Region: "England", Country: "United Kingdom", Latitude: 51.50853, Longitude: -0.12574, Zone: "Europe/London"}
	tokyo  = place.Place{GeoNamesID: 1850147, Name: "Tokyo", Region: "Tokyo", Country: "Japan", Latitude: 35.6895, Longitude: 139.69171, Zone: "Asia/Tokyo"}
	paris  = place.Place{GeoNamesID: 2988507, Name: "Paris", Region: "Île-de-France", Country: "France", Latitude: 48.85341, Longitude: 2.3488, Zone: "Europe/Paris"}
	// newYork lies behind UTC; nowhere's zone is one the tz database does not know.
	newYork = place.Place{GeoNamesID: 5128581, Name: "New York City", Region: "New York", Country: "United States", Latitude: 40.71427, Longitude: -74.00597, Zone: "America/New_York"}
	nowhere = place.Place{GeoNamesID: 9, Name: "Nowhere", Country: "Nowhere", Zone: "Not/AZone"}
)

type fakeStore struct {
	loaded  Loaded
	loadErr error
	saveErr error
	saved   []settings.Settings
}

func (f *fakeStore) Load() (Loaded, error) { return f.loaded, f.loadErr }

func (f *fakeStore) Save(next settings.Settings) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, next)
	return nil
}

type fakePlaces struct{ known []place.Place }

func (f fakePlaces) Search(typed string) []place.Place { return place.NewIndex(f.known).Search(typed) }

func (f fakePlaces) Place(id int) (place.Place, bool) {
	for _, each := range f.known {
		if each.GeoNamesID == id {
			return each, true
		}
	}
	return place.Place{}, false
}

func (f fakePlaces) Resolve(zone string) (*time.Location, error) { return time.LoadLocation(zone) }

// fakeForecasts answers each request in turn from answers, recording what was asked.
type fakeForecasts struct {
	mutex   sync.Mutex
	asked   []Request
	answers []func(Request) (Answer, error)
}

func (f *fakeForecasts) Fetch(_ context.Context, request Request) (Answer, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.asked = append(f.asked, request)
	if len(f.answers) == 0 {
		return Answer{}, fmt.Errorf("no answer scripted for request %d", len(f.asked))
	}
	next := f.answers[0]
	f.answers = f.answers[1:]
	return next(request)
}

// fakeSun answers each sunrise request in turn from answers, recording the dates asked for.
type fakeSun struct {
	asked   []forecast.Date
	answers []func() (Sun, error)
}

func (f *fakeSun) Fetch(_ context.Context, _, _ float64, date forecast.Date, _ *time.Location) (Sun, error) {
	f.asked = append(f.asked, date)
	if len(f.answers) == 0 {
		return Sun{}, fmt.Errorf("no sun scripted for request %d", len(f.asked))
	}
	next := f.answers[0]
	f.answers = f.answers[1:]
	return next()
}

type fakeCache struct {
	entries   map[string]Cached
	loadErr   error
	saveErr   error
	forgotten []string
}

func (f *fakeCache) Load() (map[string]Cached, error) { return f.entries, f.loadErr }

func (f *fakeCache) Save(id string, cached Cached) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.entries[id] = cached
	return nil
}

func (f *fakeCache) Forget(id string) error {
	f.forgotten = append(f.forgotten, id)
	delete(f.entries, id)
	return nil
}

// fakeClock is a clock the test sets; the pacer moves it on by what it is asked to wait.
type fakeClock struct{ now time.Time }

func (f *fakeClock) Now() time.Time { return f.now }

type fakePacer struct {
	clock *fakeClock
	waits []time.Duration
}

func (f *fakePacer) Wait(ctx context.Context, d time.Duration) error {
	f.waits = append(f.waits, d)
	f.clock.now = f.clock.now.Add(d)
	return ctx.Err()
}

type fakeIDs struct{ next int }

func (f *fakeIDs) NewID() string {
	f.next++
	return fmt.Sprintf("city-%d", f.next)
}

// rig is a service over fakes, started with no stored settings at 2026-10-05T07:36Z.
type rig struct {
	service   *Service
	store     *fakeStore
	forecasts *fakeForecasts
	sun       *fakeSun
	cache     *fakeCache
	clock     *fakeClock
	pacer     *fakePacer
}

func newRig() *rig {
	clock := &fakeClock{now: time.Date(2026, time.October, 5, 7, 36, 4, 0, time.UTC)}
	r := &rig{
		store:     &fakeStore{loaded: Loaded{Settings: settings.Defaults()}},
		forecasts: &fakeForecasts{},
		sun:       &fakeSun{},
		cache:     &fakeCache{entries: map[string]Cached{}},
		clock:     clock,
		pacer:     &fakePacer{clock: clock},
	}
	r.service = New(Ports{
		Store: r.store, Places: fakePlaces{known: []place.Place{london, tokyo, paris, newYork, nowhere}}, Forecasts: r.forecasts, SunTimes: r.sun,
		Cache: r.cache, Clock: clock, Pacer: r.pacer, IDs: &fakeIDs{}, Draw: func(int) int { return 0 },
	})
	return r
}

// answer scripts one successful answer expiring at expires.
func answer(expires time.Time, lastModified string) func(Request) (Answer, error) {
	return func(Request) (Answer, error) {
		return Answer{Expires: expires, LastModified: lastModified}, nil
	}
}

// failure scripts one failed request answering err.
func failure(err error) func(Request) (Answer, error) {
	return func(Request) (Answer, error) { return Answer{}, err }
}
