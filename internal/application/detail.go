package application

import (
	"context"
	"errors"
	"time"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// HourView is one hour of the detail panel, in the chosen units and format (FR-410).
type HourView struct {
	// Time is the hour's local time, such as "09:00" or "9:00 AM".
	Time        string
	Symbol      Symbol
	Temperature int
	Rain        float64
	WindSpeed   int
	// WindFrom is the direction the wind blows from, in degrees clockwise from north.
	WindFrom float64
}

// Detail is what the detail panel shows for one city (FR-410, FR-411).
type Detail struct {
	ID    string
	Label string
	Place string
	Hours []HourView
	// Sunrise and Sunset are local times; empty where the sun does not rise or set that day or the
	// times could not be had (FR-411).
	Sunrise, Sunset string
	// Problem is why the panel cannot show the hours; empty when it can (FR-410).
	Problem string
}

// OpenDetail answers the detail panel for the city id at now (FR-410). Its sunrise and sunset are
// asked of MET Norway at most once per local date, paced like the forecasts; a request that fails
// leaves them out and the next opening asks again (FR-411). After a refusal nothing is asked
// (FR-307).
func (s *Service) OpenDetail(ctx context.Context, id string) (Detail, error) {
	detail, ask, err := s.detailOf(id)
	if err != nil || ask == nil {
		return detail, err
	}
	sun, sunErr := s.askSun(ctx, ask)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if sunErr != nil {
		if errors.Is(sunErr, ErrRefused) {
			s.refused = true
		}
		return detail, nil
	}
	if state, found := s.weather[id]; found {
		state.sun, state.sunDate, state.sunHeld = sun, ask.date, true
	}
	detail.Sunrise, detail.Sunset = sunText(sun, ask.location, s.current.Format)
	return detail, nil
}

// sunAsking is a sunrise request about to leave.
type sunAsking struct {
	place    place.Place
	date     forecast.Date
	location *time.Location
}

// detailOf answers the panel as far as the held forecast goes, with the sunrise request to make when
// the day's times are not yet held. The caller holds no lock.
func (s *Service) detailOf(id string) (Detail, *sunAsking, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	city, err := s.current.City(id)
	if err != nil {
		return Detail{}, nil, err
	}
	detail := Detail{ID: id, Label: city.Label}
	where := s.locate(city)
	if where.found {
		detail.Place = where.place.Description()
	}
	if where.problem != "" {
		detail.Problem = where.problem
		return detail, nil, nil
	}
	now := s.ports.Clock.Now()
	state := s.weatherOf(id)
	current := s.current.Normalised()
	detail.Hours = s.hourViews(state.cached.Forecast.HoursAt(now), where.location, current)
	if !state.held || len(detail.Hours) == 0 {
		detail.Problem = forecastUnavailable
		return detail, nil, nil
	}
	today := forecast.DateOf(now, where.location)
	if state.sunHeld && state.sunDate == today {
		detail.Sunrise, detail.Sunset = sunText(state.sun, where.location, current.Format)
		return detail, nil, nil
	}
	if s.refused {
		return detail, nil, nil
	}
	return detail, &sunAsking{place: where.place, date: today, location: where.location}, nil
}

// askSun asks for the day's sunrise and sunset, one request among the forecasts' (NFR-S-2).
func (s *Service) askSun(ctx context.Context, ask *sunAsking) (Sun, error) {
	s.refreshing.Lock()
	defer s.refreshing.Unlock()
	if err := s.pace(ctx); err != nil {
		return Sun{}, err
	}
	sun, err := s.ports.SunTimes.Fetch(ctx, place.RoundDegrees(ask.place.Latitude), place.RoundDegrees(ask.place.Longitude), ask.date, ask.location)
	s.lastRequest = s.ports.Clock.Now()
	return sun, err
}

// hourViews answers hours as the panel shows them in location under current's units and format.
func (s *Service) hourViews(hours []forecast.Hour, location *time.Location, current settings.Settings) []HourView {
	views := make([]HourView, 0, len(hours))
	for _, each := range hours {
		views = append(views, HourView{
			Time: localtime.Text(each.Time.In(location), current.Format), Symbol: s.symbolOf(each.Symbol),
			Temperature: units.Temperature(each.AirC, current.Units), Rain: units.Rain(each.RainMM, current.Units),
			WindSpeed: units.WindSpeed(each.WindMS, current.Units), WindFrom: each.WindFrom,
		})
	}
	return views
}

// sunText answers sun's times as local times in format; empty for a time the day does not have.
func sunText(sun Sun, location *time.Location, format localtime.Format) (rise, set string) {
	if !sun.Rise.IsZero() {
		rise = localtime.Text(sun.Rise.In(location), format)
	}
	if !sun.Set.IsZero() {
		set = localtime.Text(sun.Set.In(location), format)
	}
	return rise, set
}
