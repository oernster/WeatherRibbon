// Package application holds WeatherRibbon's use cases: one named entry point per action the user can
// take, each executable from a test with no window open.
//
// It depends on the domain and on the ports declared here. Infrastructure implements the ports;
// the composition root wires them in.
package application

import (
	"context"
	"errors"
	"time"

	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/application/controls"
	"github.com/oernster/ribbonkit/application/release"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/petrichor"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// Loaded is what the store answers at launch.
type Loaded struct {
	// Settings is what was read; the defaults when nothing was there or the file was unreadable.
	Settings settings.Settings
	// Notice is a problem the user should read, such as a file kept aside; empty when none.
	Notice string
}

// Store keeps the settings between runs (FR-801, FR-802).
type Store interface {
	// Load reads the settings. Absence is not an error; an error is a fault reading them.
	Load() (Loaded, error)
	// Save writes the settings, replacing the previous copy atomically.
	Save(settings.Settings) error
}

// Places is the bundled city list (DATA-1).
type Places interface {
	// Search answers the places matching typed, best first (FR-202).
	Search(typed string) []place.Place
	// Place answers the place with the GeoNames id; false where the list holds none (FR-803).
	Place(geoNamesID int) (place.Place, bool)
	// Resolve answers the location for zone; an error when the tz database does not know it.
	Resolve(zone string) (*time.Location, error)
}

// ErrRefused is answered by Forecasts when MET Norway refuses the request outright (FR-307).
var ErrRefused = errors.New("MET Norway refused the forecast request")

// Request is one forecast request for one place.
type Request struct {
	// Latitude and Longitude are in degrees, already rounded as the terms allow (FR-302).
	Latitude, Longitude float64
	// LastModified is the cached forecast's Last-Modified header; empty when none is cached (FR-304).
	LastModified string
}

// Answer is what one successful request brought back.
type Answer struct {
	// NotModified says the cached forecast still stands (a 304); Forecast is then empty (FR-304).
	NotModified bool
	Forecast    forecast.Forecast
	// Expires is when the forecast may be asked for again (FR-303); LastModified is its header.
	Expires      time.Time
	LastModified string
}

// Forecasts asks MET Norway for forecasts (CON-9). Every failure is an error: no network, a timeout,
// a 5xx or 429, an answer that cannot be used (FR-305, FR-308); a 403 is ErrRefused (FR-307).
type Forecasts interface {
	Fetch(ctx context.Context, request Request) (Answer, error)
}

// Sun is a day's sunrise and sunset; each the zero time where the sun does not rise or set that day
// (FR-411).
type Sun struct {
	Rise, Set time.Time
}

// SunTimes asks MET Norway's Sunrise 3.0 for a place's sunrise and sunset on one local date in
// location (FR-411). Every failure is an error; a 403 is ErrRefused.
type SunTimes interface {
	Fetch(ctx context.Context, latitude, longitude float64, date forecast.Date, location *time.Location) (Sun, error)
}

// Cached is one city's forecast as kept between runs (FR-807).
type Cached struct {
	Forecast     forecast.Forecast
	Expires      time.Time
	LastModified string
	// Fetched is when the forecast was last fetched successfully, which says how stale it is (FR-305).
	Fetched time.Time
	// Spell is the city's dry spell (FR-413).
	Spell petrichor.Spell
}

// Cache keeps each city's forecast between runs (FR-807).
type Cache interface {
	// Load answers every readable entry by city id; an unreadable one is left out to be fetched again.
	Load() (map[string]Cached, error)
	// Save writes one city's entry atomically.
	Save(cityID string, cached Cached) error
	// Forget removes one city's entry (FR-205, FR-207).
	Forget(cityID string) error
}

// Clock answers the current instant. Only infrastructure reads the wall clock.
type Clock interface {
	Now() time.Time
}

// Pacer waits between forecast requests so that no two leave less than a second apart (NFR-S-2). It
// answers early, with the context's error, when the context ends.
type Pacer interface {
	Wait(ctx context.Context, d time.Duration) error
}

// IDs answers a new stable city id, unique for the life of the settings.
type IDs interface {
	NewID() string
}

// Icons is the weather icon set the page draws (FR-412).
type Icons interface {
	// Has answers whether the set holds an icon named name.
	Has(name string) bool
	// Missing hears a symbol code the set holds no icon for, once per code each run, to log it.
	Missing(code string)
}

// Ports gathers the collaborators a Service is built from.
type Ports struct {
	Store     Store
	Places    Places
	Forecasts Forecasts
	SunTimes  SunTimes
	Cache     Cache
	Clock     Clock
	Pacer     Pacer
	IDs       IDs
	// Draw is the countdown's source of chance (FR-413), injected so the domain holds none.
	Draw  petrichor.Draw
	Icons Icons
	// Monitors and Neighbours are what the kit's arranger places the ribbon among (FR-501 to FR-506);
	// a nil Neighbours is a ribbon alone.
	Monitors   arranger.Monitors
	Neighbours arranger.Neighbours
	// Startup is the start at sign-in value (FR-710).
	Startup  controls.Startup
	Releases release.Source
	// Build is not a collaborator but the facts about the running build the update check compares
	// against, given here so the composition root states them once (FR-603).
	Build release.Build
}
