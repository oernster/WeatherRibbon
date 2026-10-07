package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/oernster/ribbonkit/domain/placement"
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

type fakeMonitors struct{ monitors []placement.Monitor }

func (f fakeMonitors) Monitors() ([]placement.Monitor, error) { return f.monitors, nil }

type fakeStartup struct{ enabled bool }

func (f *fakeStartup) Enabled() (bool, error) { return f.enabled, nil }

func (f *fakeStartup) Enable() error {
	f.enabled = true
	return nil
}

func (f *fakeStartup) Disable() error {
	f.enabled = false
	return nil
}

// fakeIcons is an icon set holding the names in has, recording each code told missing.
type fakeIcons struct {
	has     map[string]bool
	mutex   sync.Mutex
	missing []string
}

func (f *fakeIcons) Has(name string) bool { return f.has[name] }

func (f *fakeIcons) Missing(code string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.missing = append(f.missing, code)
}

// testIcons are the icons the rig's set holds: the symbols the tests' forecasts carry plus the
// doubled-s spelling of light sleet showers and thunder (FR-412).
var testIcons = []string{"rain", "cloudy", "lightssleetshowersandthunder_day", "lightssnowshowersandthunder_night"}

// testLayout is the cell geometry the tests arrange with, in DIP.
var testLayout = Layout{
	Cell:    placement.Size{Width: 160, Height: 120},
	Prompt:  placement.Size{Width: 200, Height: 140},
	Padding: 8,
}

// primaryMonitor is one 1920 by 1032 work area at 100 percent.
var primaryMonitor = placement.Monitor{
	Device: `\\.\DISPLAY1`, Work: placement.Rect{Right: 1920, Bottom: 1032}, DPI: placement.BaseDPI, Primary: true,
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
	icons     *fakeIcons
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
		icons:     &fakeIcons{has: map[string]bool{}},
	}
	for _, name := range testIcons {
		r.icons.has[name] = true
	}
	r.service = New(Ports{
		Store: r.store, Places: fakePlaces{known: []place.Place{london, tokyo, paris, newYork, nowhere}}, Forecasts: r.forecasts, SunTimes: r.sun,
		Cache: r.cache, Clock: clock, Pacer: r.pacer, IDs: &fakeIDs{}, Draw: func(int) int { return 0 },
		Monitors: fakeMonitors{monitors: []placement.Monitor{primaryMonitor}}, Startup: &fakeStartup{}, Icons: r.icons,
	}, testLayout)
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
