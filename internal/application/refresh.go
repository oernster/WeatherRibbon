package application

import (
	"context"
	"errors"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/backoff"
	"github.com/oernster/weatherribbon/internal/domain/place"
)

// requestSpacing is the least time between any two forecast requests (NFR-S-2).
const requestSpacing = time.Second

// cacheSaveFailedPrefix begins the notice raised while forecasts cannot be kept on disk (FR-807).
const cacheSaveFailedPrefix = "Forecasts could not be saved: "

// refusedNotice is shown from MET Norway's refusal until the next launch (FR-307).
const refusedNotice = "MET Norway refused the forecast request"

// stoppedPrefix begins the notice shown once refreshing has stopped after a fault, until the next
// launch: the cells go on showing what is held as it grows older; the reason is said.
const stoppedPrefix = "Forecasts stopped after a fault and resume at the next launch: "

// RefreshStopped raises the notice saying refreshing has stopped, with why, so a fault that ended the
// refreshing is read on the ribbon and not only in the log.
func (s *Service) RefreshStopped(reason error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stopped = stoppedPrefix + reason.Error()
}

// asking is one request about to leave: the city and what to ask for.
type asking struct {
	id      string
	request Request
}

// Refresh asks for the forecast of every city that is due (FR-303, FR-306) and keeps each answer
// (FR-304, FR-305, FR-807). Requests leave one at a time at least a second apart (NFR-S-2); a city
// already being asked for is not asked for again (FR-310). After a refusal nothing more is asked in
// this run (FR-307). It answers the context's error when the context ends first.
func (s *Service) Refresh(ctx context.Context) error {
	return s.refresh(ctx, false)
}

// RefreshNow is Refresh asking too for every city still backing off; a city within its Expires is
// still left alone (FR-309).
func (s *Service) RefreshNow(ctx context.Context) error {
	return s.refresh(ctx, true)
}

func (s *Service) refresh(ctx context.Context, now bool) error {
	s.refreshing.Lock()
	defer s.refreshing.Unlock()
	for _, each := range s.claim(now) {
		if s.wasRefused() || s.pace(ctx) != nil {
			s.release(each.id)
			continue
		}
		answer, err := s.ports.Forecasts.Fetch(ctx, each.request)
		s.lastRequest = s.ports.Clock.Now()
		s.keep(each.id, answer, err)
	}
	return ctx.Err()
}

// pace waits until a second has passed since the last request left.
func (s *Service) pace(ctx context.Context) error {
	if s.lastRequest.IsZero() {
		return ctx.Err()
	}
	wait := s.lastRequest.Add(requestSpacing).Sub(s.ports.Clock.Now())
	if wait <= 0 {
		return ctx.Err()
	}
	return s.ports.Pacer.Wait(ctx, wait)
}

// claim marks every due city as being asked for and answers their requests. With backingOff, a city
// waiting out a failure is due too. A city whose place the list no longer holds is never asked for
// (FR-803).
func (s *Service) claim(backingOff bool) []asking {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.refused {
		return nil
	}
	now := s.ports.Clock.Now()
	var claimed []asking
	for _, city := range s.current.Cities {
		chosen, found := s.ports.Places.Place(city.GeoNamesID)
		state := s.weatherOf(city.ID)
		if !found || city.Unreadable != "" || !due(state, now, backingOff) {
			continue
		}
		state.asking = true
		claimed = append(claimed, asking{id: city.ID, request: requestFor(chosen, state)})
	}
	return claimed
}

// due answers whether a city may be asked for at now: not already being asked for, past any
// back-off unless backingOff, holding nothing or holding a forecast past its Expires.
func due(state *weather, now time.Time, backingOff bool) bool {
	if state.asking || (!backingOff && now.Before(state.retryAt)) {
		return false
	}
	return !state.held || !now.Before(state.cached.Expires)
}

// requestFor answers the request for chosen, conditional on the forecast held (FR-302, FR-304).
func requestFor(chosen place.Place, state *weather) Request {
	request := Request{Latitude: place.RoundDegrees(chosen.Latitude), Longitude: place.RoundDegrees(chosen.Longitude)}
	if state.held {
		request.LastModified = state.cached.LastModified
	}
	return request
}

// wasRefused answers whether MET Norway has refused a request in this run (FR-307).
func (s *Service) wasRefused() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.refused
}

// release marks the city as no longer being asked for.
func (s *Service) release(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if state, found := s.weather[id]; found {
		state.asking = false
	}
}

// keep records the outcome of one request. A city removed meanwhile is passed over.
func (s *Service) keep(id string, answer Answer, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	state, found := s.weather[id]
	if !found {
		return
	}
	state.asking = false
	now := s.ports.Clock.Now()
	switch {
	case errors.Is(err, ErrRefused):
		s.refused = true
		return
	case err != nil:
		state.failures++
		state.retryAt = now.Add(backoff.Wait(state.failures))
		return
	}
	state.failures, state.retryAt = 0, time.Time{}
	if !answer.NotModified {
		state.cached.Forecast = answer.Forecast
	}
	if answer.LastModified != "" {
		state.cached.LastModified = answer.LastModified
	}
	state.cached.Expires, state.cached.Fetched, state.held = answer.Expires, now, true
	s.saveCachedLocked(id, state)
}

// saveCachedLocked writes the city's forecast to disk, raising a notice while that fails. The caller
// holds the mutex.
func (s *Service) saveCachedLocked(id string, state *weather) {
	if err := s.ports.Cache.Save(id, state.cached); err != nil {
		s.cacheNotice = cacheSaveFailedPrefix + err.Error()
	}
}

// NextDue answers when the next city falls due: now where one is due already, the zero time where
// none ever will in this run (none configured; MET Norway refused).
func (s *Service) NextDue() time.Time {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.refused {
		return time.Time{}
	}
	now := s.ports.Clock.Now()
	var next time.Time
	for _, city := range s.current.Cities {
		state := s.weatherOf(city.ID)
		at := now
		if state.held && state.cached.Expires.After(at) {
			at = state.cached.Expires
		}
		if state.retryAt.After(at) {
			at = state.retryAt
		}
		if next.IsZero() || at.Before(next) {
			next = at
		}
	}
	return next
}
