package application

import (
	"fmt"
	"slices"
	"time"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// Words a cell is shown with when it cannot show the weather. They say what is wrong rather than
// relying on colour (NFR-U-2).
const (
	unreadablePrefix    = "This city could not be read: "
	placeNotFound       = "Place not found"
	unknownZonePrefix   = "Unknown time zone: "
	forecastUnavailable = "Forecast unavailable"
)

// staleAfter is how old a forecast's last successful fetch may be before the cell says its age
// (FR-305).
const staleAfter = 2 * time.Hour

// hoursBeforeDays is the age from which a stale forecast's age is told in days rather than hours.
const hoursBeforeDays = 48 * time.Hour

// hoursPerDay converts an age in hours to days.
const hoursPerDay = 24

// Day is one day of a cell, in the chosen units (FR-404, FR-406).
type Day struct {
	Date    forecast.Date
	Weekday string
	High    int
	Low     int
	Rain    float64
	Symbol  string
	// Known is false for a day the forecast does not reach; nothing else is then meaningful.
	Known bool
}

// Cell is what one cell of the ribbon shows.
type Cell struct {
	ID    string
	Label string
	// Place names the city's place in full, as the search does, for Settings (FR-202).
	Place    string
	Time     string
	ZoneMark string
	// Temperature and Symbol are the current conditions (FR-402); meaningful only where Problem is
	// empty.
	Temperature int
	Symbol      string
	Today       Day
	Outlook     []Day
	// Age is "Updated <age> ago" while the forecast is stale (FR-305); empty while it is fresh.
	Age string
	// Problem is why the cell cannot show the weather; empty when it can (FR-305, FR-403, FR-803,
	// FR-804). A cell with a problem shows its label and the problem, never another place's weather.
	Problem string
	// Petrichor is true while the cell shows the petrichor line (FR-413).
	Petrichor bool
}

// Snapshot is everything the ribbon draws at one instant.
type Snapshot struct {
	Cells  []Cell
	Units  units.System
	Format localtime.Format
	// Now is the instant the snapshot was taken at; NextRefresh the minute boundary to take the next
	// at (FR-401).
	Now         time.Time
	NextRefresh time.Time
	// Notices are problems for the user to read, oldest kind first.
	Notices []string
}

// placedCell is a cell with its zone's offset at the snapshot's instant; placed is false for a cell
// with no zone to order by, which goes after every cell that has one.
type placedCell struct {
	cell          Cell
	offsetSeconds int
	placed        bool
}

// Snapshot answers what the ribbon shows now, one cell per city ordered east from Greenwich; cities
// keeping the same offset keep the order they were added in (FR-108). Each snapshot also observes
// every city for the petrichor moment (FR-413).
func (s *Service) Snapshot() Snapshot {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	now := s.ports.Clock.Now()
	current := s.current.Normalised()
	placed := make([]placedCell, 0, len(current.Cities))
	for _, city := range current.Cities {
		placed = append(placed, s.cellFor(city, now, current))
	}
	slices.SortStableFunc(placed, func(a, b placedCell) int {
		if a.placed != b.placed {
			if a.placed {
				return -1
			}
			return 1
		}
		return localtime.EastFromGreenwich(a.offsetSeconds, b.offsetSeconds)
	})
	cells := make([]Cell, 0, len(placed))
	for _, each := range placed {
		cells = append(cells, each.cell)
	}
	return Snapshot{
		Cells: cells, Units: current.Units, Format: current.Format,
		Now: now, NextRefresh: localtime.NextRefresh(now), Notices: s.notices(),
	}
}

// cellFor answers city's cell at now. The caller holds the mutex.
func (s *Service) cellFor(city settings.City, now time.Time, current settings.Settings) placedCell {
	cell := Cell{ID: city.ID, Label: city.Label}
	if city.Unreadable != "" {
		cell.Problem = unreadablePrefix + city.Unreadable
		return placedCell{cell: cell}
	}
	chosen, found := s.ports.Places.Place(city.GeoNamesID)
	if !found {
		cell.Problem = placeNotFound
		return placedCell{cell: cell}
	}
	cell.Place = chosen.Description()
	location, err := s.ports.Places.Resolve(chosen.Zone)
	if err != nil {
		cell.Problem = unknownZonePrefix + chosen.Zone
		return placedCell{cell: cell}
	}
	clock := place.ClockAt(now, location, current.Format)
	cell.Time, cell.ZoneMark = clock.Time, clock.ZoneMark
	s.fillWeather(&cell, s.weatherOf(city.ID), now, location, current.Units)
	return placedCell{cell: cell, offsetSeconds: clock.OffsetSeconds, placed: true}
}

// fillWeather writes the city's forecast into cell, then observes the city for the petrichor moment.
// The caller holds the mutex.
func (s *Service) fillWeather(cell *Cell, state *weather, now time.Time, location *time.Location, system units.System) {
	conditions, current := state.cached.Forecast.CurrentAt(now)
	usable := state.held && current
	stale := state.held && now.Sub(state.cached.Fetched) > staleAfter
	s.observe(cell.ID, state, conditions, usable && !stale, now)
	cell.Petrichor = state.moment
	if !usable {
		cell.Problem = forecastUnavailable
		return
	}
	if stale {
		cell.Age = age(now.Sub(state.cached.Fetched))
	}
	cell.Temperature = units.Temperature(conditions.AirC, system)
	cell.Symbol = conditions.Symbol
	today, _ := state.cached.Forecast.TodayAt(now, location)
	cell.Today = dayIn(today, system)
	for _, each := range state.cached.Forecast.OutlookAt(now, location) {
		cell.Outlook = append(cell.Outlook, dayIn(each, system))
	}
}

// dayIn answers day in system.
func dayIn(day forecast.Day, system units.System) Day {
	return Day{
		Date: day.Date, Weekday: day.Date.Weekday().String(),
		High: units.Temperature(day.HighC, system), Low: units.Temperature(day.LowC, system),
		Rain: units.Rain(day.RainMM, system), Symbol: day.Symbol, Known: day.Known,
	}
}

// age answers how a stale forecast's age is told (FR-305): "Updated 3 hours ago", "Updated 2 days
// ago" from two days on.
func age(old time.Duration) string {
	hours := int(old / time.Hour)
	if old < hoursBeforeDays {
		return fmt.Sprintf("Updated %d hours ago", hours)
	}
	return fmt.Sprintf("Updated %d days ago", hours/hoursPerDay)
}
