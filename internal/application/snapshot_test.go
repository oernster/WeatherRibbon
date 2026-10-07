package application

import (
	"slices"
	"testing"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// hourlyAround answers a forecast with hourly steps from an hour before now for four days, each at
// airC with rain mm in its hour and 6-hour blocks.
func hourlyAround(now time.Time, airC, rain float64) forecast.Forecast {
	var steps []forecast.Step
	start := now.Truncate(time.Hour).Add(-time.Hour)
	for hour := range 4 * 24 {
		steps = append(steps, forecast.Step{
			Time: start.Add(time.Duration(hour) * time.Hour), AirC: airC,
			Next1: forecast.Period{Hours: 1, Symbol: "rain", RainMM: rain, HasRain: true},
			Next6: forecast.Period{Hours: 6, Symbol: "cloudy", RainMM: rain, HasRain: true, MaxC: airC, MinC: airC, HasExtremes: true},
		})
	}
	return forecast.New(steps)
}

// hold gives the city a forecast fetched at fetched.
func hold(r *rig, id string, f forecast.Forecast, fetched time.Time) {
	r.service.mutex.Lock()
	defer r.service.mutex.Unlock()
	state := r.service.weatherOf(id)
	state.cached.Forecast, state.cached.Fetched, state.cached.Expires, state.held = f, fetched, fetched.Add(time.Hour), true
}

// labels answers the cells' labels in order.
func labels(cells []Cell) []string {
	var out []string
	for _, each := range cells {
		out = append(out, each.Label)
	}
	return out
}

// FR-108: given New York, Tokyo and London added in that order, the ribbon shows London, Tokyo, New
// York; a cell with no zone to order by goes last.
func TestTheRibbonRunsEastFromGreenwich(t *testing.T) {
	t.Parallel()
	r := added(t, newYork.GeoNamesID, nowhere.GeoNamesID, tokyo.GeoNamesID, london.GeoNamesID)
	got := labels(r.service.Snapshot().Cells)
	if want := []string{"London", "Tokyo", "New York City", "Nowhere"}; !slices.Equal(got, want) {
		t.Errorf("order %v; want %v", got, want)
	}
}

// FR-401, FR-402, FR-404, FR-406, FR-703: a city with a fresh forecast shows its local time, its
// current conditions, today and three days ahead in the chosen units.
func TestACellShowsTheWeatherInTheChosenUnits(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	hold(r, "city-1", hourlyAround(r.clock.now, 14.4, 0.5), r.clock.now)
	cell := r.service.Snapshot().Cells[0]
	if cell.Time != "08:36" || cell.ZoneMark != "BST" || cell.Place != "London, England, United Kingdom" {
		t.Errorf("time %q %q, place %q", cell.Time, cell.ZoneMark, cell.Place)
	}
	if cell.Problem != "" || cell.Age != "" || cell.Temperature != 14 || cell.Symbol != (Symbol{Icon: "rain", Words: "Rain"}) {
		t.Errorf("current %+v", cell)
	}
	if !cell.Today.Known || cell.Today.High != 14 || len(cell.Outlook) != 3 || cell.Outlook[0].Weekday != "Tuesday" {
		t.Errorf("today %+v, outlook %+v", cell.Today, cell.Outlook)
	}
	_ = r.service.SetUnits(units.Imperial)
	if imperial := r.service.Snapshot().Cells[0]; imperial.Temperature != 58 {
		t.Errorf("imperial temperature %d; want 58", imperial.Temperature)
	}
}

// FR-305: past two hours since the last fetch the cell says its age, in days from two days on.
func TestAStaleForecastSaysItsAge(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	hold(r, "city-1", hourlyAround(r.clock.now, 10, 0), r.clock.now.Add(-3*time.Hour))
	if got := r.service.Snapshot().Cells[0].Age; got != "Updated 3 hours ago" {
		t.Errorf("age %q", got)
	}
	hold(r, "city-1", hourlyAround(r.clock.now, 10, 0), r.clock.now.Add(-50*time.Hour))
	if got := r.service.Snapshot().Cells[0].Age; got != "Updated 2 days ago" {
		t.Errorf("age %q", got)
	}
	hold(r, "city-1", hourlyAround(r.clock.now, 10, 0), r.clock.now.Add(-2*time.Hour))
	if got := r.service.Snapshot().Cells[0].Age; got != "" {
		t.Errorf("at exactly two hours the age is not yet said: %q", got)
	}
}

// FR-305, FR-403, FR-803, FR-804: a cell that cannot show the weather says why and shows nothing else.
func TestACellThatCannotShowTheWeatherSaysWhy(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID, nowhere.GeoNamesID)
	r.service.current = r.service.current.
		WithCityAdded(settings.City{ID: "lost", GeoNamesID: 1, Label: "Atlantis"}).
		WithCityAdded(settings.City{ID: "bad", Label: "?", Unreadable: "geonamesId is not a number", Original: "{}"})
	hold(r, "city-1", hourlyAround(r.clock.now.Add(-200*time.Hour), 10, 0), r.clock.now)
	problems := map[string]string{}
	for _, each := range r.service.Snapshot().Cells {
		problems[each.Label] = each.Problem
	}
	want := map[string]string{
		"London":   forecastUnavailable,
		"Nowhere":  unknownZonePrefix + "Not/AZone",
		"Atlantis": placeNotFound,
		"?":        unreadablePrefix + "geonamesId is not a number",
	}
	for label, problem := range want {
		if problems[label] != problem {
			t.Errorf("%s: problem %q; want %q", label, problems[label], problem)
		}
	}
}

// The snapshot carries the next minute boundary and the standing notices.
func TestTheSnapshotCarriesTheNextMinuteAndTheNotices(t *testing.T) {
	t.Parallel()
	r := added(t)
	r.service.loadNotice = "kept aside"
	snapshot := r.service.Snapshot()
	if !snapshot.NextRefresh.Equal(time.Date(2026, time.October, 5, 7, 37, 0, 0, time.UTC)) {
		t.Errorf("next refresh %v", snapshot.NextRefresh)
	}
	if !slices.Equal(snapshot.Notices, []string{"kept aside"}) || snapshot.Units != units.Metric {
		t.Errorf("snapshot %+v", snapshot)
	}
}
