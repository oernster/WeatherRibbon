package application

import (
	"errors"
	"slices"
	"testing"

	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// FR-201: a chosen place is appended labelled with its name and saved; an unknown one is refused.
func TestAddingACityAppendsItWithItsName(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	id, err := r.service.AddCity(tokyo.GeoNamesID)
	cities := r.service.Settings().Cities
	if err != nil || id != "city-2" || len(cities) != 2 || cities[1].Label != "Tokyo" || cities[1].GeoNamesID != tokyo.GeoNamesID {
		t.Fatalf("added %q, %v: %+v", id, err, cities)
	}
	if _, err := r.service.AddCity(42); !errors.Is(err, ErrUnknownPlace) {
		t.Errorf("an unknown place answered %v", err)
	}
	if got := r.service.Search("tok"); len(got) != 1 || got[0].Name != "Tokyo" {
		t.Errorf("search %+v", got)
	}
}

// FR-204: a typed label is kept trimmed; an empty one is the place's name, else the city's own label
// where the list has lost the place.
func TestRelabellingKeepsTheTypedLabel(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID)
	if err := r.service.Relabel("city-1", " Mum "); err != nil || r.service.Settings().Cities[0].Label != "Mum" {
		t.Fatalf("relabel: %v %+v", err, r.service.Settings().Cities)
	}
	if err := r.service.Relabel("city-1", ""); err != nil || r.service.Settings().Cities[0].Label != "London" {
		t.Errorf("empty label: %v %+v", err, r.service.Settings().Cities)
	}
	r.service.current = settings.Defaults().WithCityAdded(settings.City{ID: "lost", GeoNamesID: 1, Label: "Atlantis"})
	if err := r.service.Relabel("lost", ""); err != nil || r.service.Settings().Cities[0].Label != "Atlantis" {
		t.Errorf("lost place: %v %+v", err, r.service.Settings().Cities)
	}
	if err := r.service.Relabel("nobody", "x"); !errors.Is(err, settings.ErrNoSuchCity) {
		t.Errorf("an unknown id answered %v", err)
	}
}

// FR-205: removing a city forgets its forecast; an unknown id changes nothing.
func TestRemovingACityForgetsItsForecast(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID, tokyo.GeoNamesID)
	r.service.weather["city-1"] = &weather{held: true}
	if err := r.service.RemoveCity("city-1"); err != nil {
		t.Fatal(err)
	}
	if _, found := r.service.weather["city-1"]; found || !slices.Equal(r.cache.forgotten, []string{"city-1"}) {
		t.Errorf("forecast not forgotten: forgotten %v", r.cache.forgotten)
	}
	if err := r.service.RemoveCity("city-1"); !errors.Is(err, settings.ErrNoSuchCity) {
		t.Errorf("removing twice answered %v", err)
	}
	if len(r.cache.forgotten) != 1 {
		t.Error("a refused removal forgot a forecast")
	}
}

// FR-207: changing place keeps a typed label and the position; the old forecast is forgotten.
func TestChangingPlaceForgetsTheOldForecast(t *testing.T) {
	t.Parallel()
	r := added(t, london.GeoNamesID, tokyo.GeoNamesID)
	_ = r.service.Relabel("city-1", "Mum")
	if err := r.service.ChangePlace("city-1", paris.GeoNamesID); err != nil {
		t.Fatal(err)
	}
	first := r.service.Settings().Cities[0]
	if first.ID != "city-1" || first.Label != "Mum" || first.GeoNamesID != paris.GeoNamesID {
		t.Errorf("changed city %+v", first)
	}
	if !slices.Equal(r.cache.forgotten, []string{"city-1"}) {
		t.Errorf("forgotten %v", r.cache.forgotten)
	}
	if err := r.service.ChangePlace("city-1", 42); !errors.Is(err, ErrUnknownPlace) {
		t.Errorf("an unknown place answered %v", err)
	}
	if err := r.service.ChangePlace("nobody", paris.GeoNamesID); !errors.Is(err, settings.ErrNoSuchCity) {
		t.Errorf("an unknown id answered %v", err)
	}
}
