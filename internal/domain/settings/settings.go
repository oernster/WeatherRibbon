// Package settings holds the user's choices as one value and the operations that change them.
//
// Every operation answers a new value and leaves the receiver as it was, so a change that is refused
// changes nothing. Derived values (forecasts, local times, offsets) are never held here (FR-801).
package settings

import (
	"errors"
	"slices"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/weatherribbon/internal/domain/petrichor"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// ErrNoSuchCity is answered when an operation names a city id that is not configured.
var ErrNoSuchCity = errors.New("no city has that id")

// City is one configured city as stored. Its position is its index in Settings.Cities.
type City struct {
	// ID is stable for the city's lifetime and never reused.
	ID string
	// GeoNamesID names the city's place in the city list; it may name one a newer list has dropped
	// (FR-803).
	GeoNamesID int
	// Label is the name shown.
	Label string
}

// Settings is every choice the user has made: the ribbon's own, which ribbonkit holds, then the
// weather's. The ribbon's are embedded, so they read as fields of Settings.
type Settings struct {
	ribbon.Choices
	Units  units.System
	Format localtime.Format
	// PetrichorCountdown is the petrichor events left before the next moment; zero until the first is
	// drawn (FR-413).
	PetrichorCountdown petrichor.Countdown
	// Cities is the configured cities in the order they were added (FR-108).
	Cities []City
}

// Defaults answers the settings of a first run: the ribbon's own (vertical, FR-107), then metric
// (FR-703), 24-hour (FR-401), no countdown drawn, no cities.
func Defaults() Settings {
	return Settings{
		Choices: ribbon.Defaults(),
		Units:   units.Metric,
		Format:  localtime.TwentyFourHour,
	}
}

// Normalised answers the settings with any choice that is not one of the known values replaced by
// its default, so a hand-edited file holding a word it should not cannot leave a choice unset. A
// countdown outside the bounds a draw gives is forgotten, so the next event draws afresh.
func (s Settings) Normalised() Settings {
	defaults := Defaults()
	s.Choices = s.Choices.Normalised()
	s.Units = units.Normalise(s.Units)
	if !slices.Contains(localtime.Formats, s.Format) {
		s.Format = defaults.Format
	}
	if !s.PetrichorCountdown.Valid() {
		s.PetrichorCountdown = defaults.PetrichorCountdown
	}
	s.Cities = slices.Clone(s.Cities)
	return s
}

// WithCityAdded answers the settings with city appended at the end (FR-201).
func (s Settings) WithCityAdded(city City) Settings {
	s.Cities = append(slices.Clone(s.Cities), city)
	return s
}

// WithoutCity answers the settings with the city id removed and the gap closed (FR-205).
func (s Settings) WithoutCity(id string) (Settings, error) {
	index := s.indexOf(id)
	if index < 0 {
		return s, ErrNoSuchCity
	}
	s.Cities = slices.Delete(slices.Clone(s.Cities), index, index+1)
	return s, nil
}

// WithCityReplaced answers the settings with the city of city's id replaced by city, keeping its
// position (FR-204, FR-207).
func (s Settings) WithCityReplaced(city City) (Settings, error) {
	index := s.indexOf(city.ID)
	if index < 0 {
		return s, ErrNoSuchCity
	}
	s.Cities = slices.Clone(s.Cities)
	s.Cities[index] = city
	return s, nil
}

// City answers the configured city with id.
func (s Settings) City(id string) (City, error) {
	index := s.indexOf(id)
	if index < 0 {
		return City{}, ErrNoSuchCity
	}
	return s.Cities[index], nil
}

func (s Settings) indexOf(id string) int {
	return slices.IndexFunc(s.Cities, func(city City) bool { return city.ID == id })
}

// Relocated answers city moved from the place it named, old, to another, to (FR-207). Its label
// follows the new place only where it was the old place's name; a label the user typed is kept.
func Relocated(city City, old, to place.Place) City {
	if city.Label == old.Name || city.Label == "" {
		city.Label = to.Name
	}
	city.GeoNamesID = to.GeoNamesID
	return city
}
