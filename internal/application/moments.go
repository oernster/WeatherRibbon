package application

import (
	"time"

	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/petrichor"
	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// observe judges one city at a snapshot for the petrichor moment (FR-413): its dry spell moves on,
// kept with its forecast when it changes; an event counts down, the countdown kept in the settings;
// the event ending the countdown sets the line on the city's cell. A dry finding takes the line away.
// The caller holds the mutex.
func (s *Service) observe(id string, state *weather, conditions forecast.Current, usable bool, now time.Time) {
	finding := petrichor.Judge(conditions, usable)
	spell, event := state.cached.Spell.Observe(finding, now)
	if spell != state.cached.Spell {
		// Only a wet or dry finding moves a spell; either needs a forecast held.
		state.cached.Spell = spell
		s.saveCachedLocked(id, state)
	}
	if !petrichor.Shown(finding) {
		state.moment = false
	}
	if !event {
		return
	}
	var moment bool
	// A failed save keeps the countdown in effect and raises its own notice (FR-802).
	_ = s.changeLocked(func(current settings.Settings) (settings.Settings, error) {
		current.PetrichorCountdown, moment = current.PetrichorCountdown.Count(s.ports.Draw)
		return current, nil
	})
	if moment {
		state.moment = true
	}
}

// DismissPetrichor hides the petrichor line on the city's cell, as a click on it does (FR-413).
func (s *Service) DismissPetrichor(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if state, found := s.weather[id]; found {
		state.moment = false
	}
}
