package application

import (
	"fmt"
	"sync"
	"time"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/controls"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// Notices the service raises for the user to read.
const (
	saveFailedPrefix  = "Settings could not be saved: "
	loadFailedPrefix  = "Settings could not be read: "
	cacheFailedPrefix = "Saved forecasts could not be read: "
)

// weather is what the service knows of one city's forecast while it runs.
type weather struct {
	// cached is the last forecast fetched; held says whether there is one.
	cached Cached
	held   bool
	// failures counts failed requests in a row; retryAt is the earliest the next may leave (FR-306).
	failures int
	retryAt  time.Time
	// asking is true while a request for the city is outstanding (FR-310).
	asking bool
	// moment is true while the city's cell shows the petrichor line (FR-413).
	moment bool
	// sun is the day's sunrise and sunset, held for sunDate once asked for successfully (FR-411).
	sun     Sun
	sunDate forecast.Date
	sunHeld bool
}

// Service runs every use case over the current settings. It is safe to call from several
// goroutines: Wails, the tray and the refresh loop each call in on their own.
//
// The ribbon is arranged by the kit's Arranger and its choices set by the kit's Controls, both
// embedded so their use cases are the service's own; the service is their host, answering the
// cities as the ribbon's content and saving the choices with the rest (see host).
type Service struct {
	*arranger.Arranger
	*controls.Controls

	ports  Ports
	layout Layout

	// refreshing lets one refresh run at a time, so requests from two never interleave closer than a
	// second (NFR-S-2); lastRequest is when the last request left. Both are held by refreshing alone.
	refreshing  sync.Mutex
	lastRequest time.Time

	mutex       sync.Mutex
	current     settings.Settings
	weather     map[string]*weather
	refused     bool
	stopped     string
	loadNotice  string
	cacheNotice string
	saveNotice  string
	// measured is the cell width the page last measured its widest text to need, with what it was
	// measured under; the zero value, before it says, widens nothing (FR-103).
	measured Measured
	// missing holds each symbol code met with no icon, so each is told once a run (FR-412).
	missing sync.Map
}

// New answers a service over ports drawing cells at layout, with the first-run settings; Start
// loads the stored ones.
func New(ports Ports, layout Layout) *Service {
	s := &Service{ports: ports, layout: layout, current: settings.Defaults(), weather: map[string]*weather{}}
	s.Arranger = arranger.New(host{s}, ports.Monitors, ports.Neighbours)
	s.Controls = controls.New(host{s}, ports.Startup, ports.Releases, ports.Build)
	return s
}

// Start loads the stored settings, then the saved forecasts. A fault reading the settings is
// answered and the defaults kept, so the ribbon still opens; a fault reading the forecasts leaves
// them to be fetched again (FR-807). Each fault raises a notice.
func (s *Service) Start() error {
	loaded, err := s.ports.Store.Load()
	saved, cacheErr := s.ports.Cache.Load()
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if cacheErr != nil {
		s.cacheNotice = cacheFailedPrefix + cacheErr.Error()
	}
	for id, each := range saved {
		s.weather[id] = &weather{cached: each, held: true}
	}
	if err != nil {
		s.loadNotice = fmt.Sprintf("%s%v", loadFailedPrefix, err)
		return err
	}
	s.current = loaded.Settings.Normalised()
	s.loadNotice = loaded.Notice
	return nil
}

// Settings answers a copy of the current settings.
func (s *Service) Settings() settings.Settings {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.current.Normalised()
}

// DismissNotices clears the notices the user has read. A save failure that still holds is raised
// again by the next save that fails.
func (s *Service) DismissNotices() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadNotice, s.cacheNotice, s.saveNotice = "", "", ""
}

// change applies edit to the current settings and saves the result. An edit that answers an error
// changes nothing. A save that fails keeps the change in effect and raises a notice until a later
// save succeeds (FR-702, FR-802); it is answered too, so a caller can tell. The caller holds no lock.
func (s *Service) change(edit func(settings.Settings) (settings.Settings, error)) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.changeLocked(edit)
}

// changeLocked is change for a caller already holding the mutex.
func (s *Service) changeLocked(edit func(settings.Settings) (settings.Settings, error)) error {
	next, err := edit(s.current)
	if err != nil {
		return err
	}
	s.current = next
	if err := s.ports.Store.Save(next); err != nil {
		s.saveNotice = saveFailedPrefix + err.Error()
		return err
	}
	s.saveNotice = ""
	return nil
}

// notices answers the notices standing now, oldest kind first. The caller holds the mutex.
func (s *Service) notices() []string {
	var out []string
	for _, notice := range []string{s.loadNotice, s.cacheNotice, s.saveNotice} {
		if notice != "" {
			out = append(out, notice)
		}
	}
	if s.refused {
		// Not dismissed: it stands until the next launch (FR-307).
		out = append(out, refusedNotice)
	}
	if s.stopped != "" {
		// Not dismissed either: nothing refreshes again in this run.
		out = append(out, s.stopped)
	}
	return out
}

// weatherOf answers the city's forecast state, made empty when there is none yet. The caller holds
// the mutex.
func (s *Service) weatherOf(id string) *weather {
	state, found := s.weather[id]
	if !found {
		state = &weather{}
		s.weather[id] = state
	}
	return state
}
