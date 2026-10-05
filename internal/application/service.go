package application

import (
	"fmt"
	"sync"
	"time"

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
}

// Service runs every use case over the current settings. It is safe to call from several
// goroutines: Wails, the tray and the refresh loop each call in on their own.
type Service struct {
	ports Ports

	// refreshing lets one refresh run at a time, so requests from two never interleave closer than a
	// second (NFR-S-2); lastRequest is when the last request left. Both are held by refreshing alone.
	refreshing  sync.Mutex
	lastRequest time.Time

	mutex       sync.Mutex
	current     settings.Settings
	weather     map[string]*weather
	refused     bool
	loadNotice  string
	cacheNotice string
	saveNotice  string
}

// New answers a service over ports with the first-run settings; Start loads the stored ones.
func New(ports Ports) *Service {
	return &Service{ports: ports, current: settings.Defaults(), weather: map[string]*weather{}}
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
