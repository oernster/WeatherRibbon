package application

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// noticesOf answers the service's standing notices.
func noticesOf(s *Service) []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.notices()
}

// A fault reading the settings keeps the defaults and raises a notice; a fault reading the saved
// forecasts raises its own; saved forecasts are held from the start (FR-802, FR-807).
func TestStartKeepsGoingWhatEverCannotBeRead(t *testing.T) {
	t.Parallel()
	r := newRig()
	r.store.loadErr = errPlanted
	r.cache.loadErr = errPlanted
	if err := r.service.Start(); !errors.Is(err, errPlanted) {
		t.Fatalf("Start answered %v", err)
	}
	notices := noticesOf(r.service)
	if len(notices) != 2 || !strings.HasPrefix(notices[0], loadFailedPrefix) || !strings.HasPrefix(notices[1], cacheFailedPrefix) {
		t.Errorf("notices %v", notices)
	}
	r.service.DismissNotices()
	if got := noticesOf(r.service); len(got) != 0 {
		t.Errorf("dismissed notices remain: %v", got)
	}

	held := newRig()
	held.store.loaded = Loaded{Settings: settings.Settings{Units: "kelvin"}, Notice: "kept aside"}
	held.cache.entries["city-1"] = Cached{LastModified: "then"}
	if err := held.service.Start(); err != nil {
		t.Fatal(err)
	}
	if got := held.service.Settings(); got.Units != units.Metric {
		t.Errorf("stored settings were not normalised: %+v", got)
	}
	if state := held.service.weather["city-1"]; state == nil || !state.held || state.cached.LastModified != "then" {
		t.Errorf("saved forecast not held: %+v", state)
	}
	if got := noticesOf(held.service); !slices.Equal(got, []string{"kept aside"}) {
		t.Errorf("notices %v; want the store's own", got)
	}
}

// FR-702, FR-802: a failed save keeps the change in effect and raises a notice until a save succeeds.
func TestWriteFailureIsReportedAndCleared(t *testing.T) {
	t.Parallel()
	r := added(t)
	r.store.saveErr = errPlanted
	if err := r.service.SetUnits(units.Imperial); !errors.Is(err, errPlanted) {
		t.Fatalf("SetUnits answered %v", err)
	}
	if r.service.Settings().Units != units.Imperial {
		t.Error("the change was not kept in effect")
	}
	if got := noticesOf(r.service); len(got) != 1 || !strings.HasPrefix(got[0], saveFailedPrefix) {
		t.Errorf("notices %v", got)
	}
	r.store.saveErr = nil
	if err := r.service.SetUnits(units.Metric); err != nil || len(noticesOf(r.service)) != 0 {
		t.Errorf("a later save did not clear the notice: %v %v", err, noticesOf(r.service))
	}
}

// FR-703, FR-401: a value a setting does not offer is refused and changes nothing.
func TestAValueASettingDoesNotOfferIsRefused(t *testing.T) {
	t.Parallel()
	r := added(t)
	if err := r.service.SetUnits("kelvin"); !errors.Is(err, ErrUnknownChoice) {
		t.Errorf("SetUnits(kelvin) answered %v", err)
	}
	if err := r.service.SetFormat("13h"); !errors.Is(err, ErrUnknownChoice) {
		t.Errorf("SetFormat(13h) answered %v", err)
	}
	if err := r.service.SetFormat("12h"); err != nil || r.service.Settings().Format != "12h" {
		t.Errorf("SetFormat(12h): %v", err)
	}
	if len(r.store.saved) != 1 {
		t.Errorf("saved %d times; want the one accepted change alone", len(r.store.saved))
	}
}

// A failed forecast save raises a notice; an answer for a city removed meanwhile is passed over.
func TestAForecastThatCannotBeKeptIsSaid(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	r.cache.saveErr = errPlanted
	r.forecasts.answers = append(r.forecasts.answers, answer(r.clock.now.Add(time.Hour), ""))
	_ = r.service.Refresh(context.Background())
	if got := noticesOf(r.service); len(got) != 1 || !strings.HasPrefix(got[0], cacheSaveFailedPrefix) {
		t.Errorf("notices %v", got)
	}
	r.service.keep("gone", Answer{}, nil)
	if _, found := r.service.weather["gone"]; found {
		t.Error("an answer for a removed city was kept")
	}
}

// With no cities nothing falls due; a city whose place is gone is never asked for (FR-803).
func TestNothingIsAskedForWithoutAPlace(t *testing.T) {
	t.Parallel()
	r := added(t)
	if !r.service.NextDue().IsZero() {
		t.Error("with no cities something fell due")
	}
	gone := settings.Defaults().WithCityAdded(settings.City{ID: "x", GeoNamesID: 1, Label: "Atlantis"}).
		WithCityAdded(settings.City{ID: "y", GeoNamesID: london.GeoNamesID, Unreadable: "bad", Original: "{}"})
	r.service.current = gone
	if claimed := r.service.claim(true); len(claimed) != 0 {
		t.Errorf("claimed %+v; want neither a lost place nor an unreadable city", claimed)
	}
}
