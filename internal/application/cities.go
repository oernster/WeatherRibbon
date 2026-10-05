package application

import (
	"errors"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// ErrUnknownPlace is answered when a city is set to a place the city list does not hold.
var ErrUnknownPlace = errors.New("no such place in the city list")

// ErrUnknownChoice is answered when a setting is given a value it does not offer.
var ErrUnknownChoice = errors.New("not one of the values offered")

// Search answers the places matching typed, best first (FR-202, FR-203).
func (s *Service) Search(typed string) []place.Place {
	return s.ports.Places.Search(typed)
}

// AddCity appends a city for the place with the GeoNames id, labelled with the place's name, then
// saves it (FR-201). It answers the new city's id.
func (s *Service) AddCity(geoNamesID int) (string, error) {
	chosen, found := s.ports.Places.Place(geoNamesID)
	if !found {
		return "", ErrUnknownPlace
	}
	id := s.ports.IDs.NewID()
	city := settings.City{ID: id, GeoNamesID: chosen.GeoNamesID, Label: chosen.Name}
	err := s.change(func(current settings.Settings) (settings.Settings, error) {
		return current.WithCityAdded(city), nil
	})
	return id, err
}

// Relabel stores what the user typed as the city's label: trimmed and capped, the place's name when
// nothing is left (FR-204).
func (s *Service) Relabel(id, typed string) error {
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		city, err := current.City(id)
		if err != nil {
			return current, err
		}
		chosen, found := s.ports.Places.Place(city.GeoNamesID)
		if !found {
			// A place the list no longer holds keeps the city's own label as its fallback (FR-803).
			chosen = place.Place{Name: city.Label}
		}
		city.Label = chosen.Label(typed)
		return current.WithCityReplaced(city)
	})
}

// RemoveCity removes the city and forgets its forecast (FR-205). The confirmation naming it is the
// page's.
func (s *Service) RemoveCity(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.changeLocked(func(current settings.Settings) (settings.Settings, error) {
		return current.WithoutCity(id)
	}); err != nil {
		return err
	}
	return s.forgetLocked(id)
}

// ChangePlace points the city at another place, keeping its position; its label follows only where
// it was the old place's name; the old forecast is forgotten (FR-207, FR-803).
func (s *Service) ChangePlace(id string, geoNamesID int) error {
	to, found := s.ports.Places.Place(geoNamesID)
	if !found {
		return ErrUnknownPlace
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.changeLocked(func(current settings.Settings) (settings.Settings, error) {
		city, err := current.City(id)
		if err != nil {
			return current, err
		}
		old, _ := s.ports.Places.Place(city.GeoNamesID)
		return current.WithCityReplaced(settings.Relocated(city, old, to))
	}); err != nil {
		return err
	}
	return s.forgetLocked(id)
}

// forgetLocked drops the city's forecast, held and saved. The caller holds the mutex.
func (s *Service) forgetLocked(id string) error {
	delete(s.weather, id)
	return s.ports.Cache.Forget(id)
}

// SetUnits chooses how measures are shown (FR-703).
func (s *Service) SetUnits(system units.System) error {
	if units.Normalise(system) != system {
		return ErrUnknownChoice
	}
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		current.Units = system
		return current, nil
	})
}

// SetFormat chooses 24-hour or 12-hour time (FR-401).
func (s *Service) SetFormat(format localtime.Format) error {
	if localtime.Normalise(format) != format {
		return ErrUnknownChoice
	}
	return s.change(func(current settings.Settings) (settings.Settings, error) {
		current.Format = format
		return current, nil
	})
}
