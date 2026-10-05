package settings

import (
	"errors"
	"testing"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/weatherribbon/internal/domain/petrichor"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// FR-107, FR-401, FR-703: a first run is vertical, metric and 24-hour, with no cities and no
// countdown drawn.
func TestDefaultsAreVerticalMetricAndTwentyFourHour(t *testing.T) {
	t.Parallel()
	got := Defaults()
	if got.Orientation != ribbon.Vertical || got.Units != units.Metric || got.Format != localtime.TwentyFourHour ||
		len(got.Cities) != 0 || got.PetrichorCountdown != 0 {
		t.Errorf("defaults %+v", got)
	}
}

// Unknown words and an impossible countdown are replaced by their defaults; known ones are kept.
func TestUnknownChoicesAreNormalisedToDefaults(t *testing.T) {
	t.Parallel()
	odd := Settings{Units: "kelvin", Format: "13h", PetrichorCountdown: petrichor.MostEvents + 1}
	got := odd.Normalised()
	if got.Units != units.Metric || got.Format != localtime.TwentyFourHour || got.PetrichorCountdown != 0 {
		t.Errorf("normalised %+v", got)
	}
	known := Settings{Units: units.Imperial, Format: localtime.TwelveHour, PetrichorCountdown: 7}
	if kept := known.Normalised(); kept.Units != units.Imperial || kept.Format != localtime.TwelveHour || kept.PetrichorCountdown != 7 {
		t.Errorf("known choices were changed: %+v", kept)
	}
}

// FR-201, FR-205: cities are appended, found, replaced in place and removed with the gap closed; an
// unknown id is refused and changes nothing.
func TestCitiesAreAddedReplacedAndRemoved(t *testing.T) {
	t.Parallel()
	base := Defaults().
		WithCityAdded(City{ID: "a", GeoNamesID: 2643743, Label: "London"}).
		WithCityAdded(City{ID: "b", GeoNamesID: 1850147, Label: "Tokyo"}).
		WithCityAdded(City{ID: "c", GeoNamesID: 5128581, Label: "New York"})
	replaced, err := base.WithCityReplaced(City{ID: "b", GeoNamesID: 1850147, Label: "Hiro"})
	if err != nil || replaced.Cities[1].Label != "Hiro" || base.Cities[1].Label != "Tokyo" {
		t.Fatalf("replace: %v %+v; the original must be untouched", err, replaced.Cities)
	}
	removed, err := replaced.WithoutCity("b")
	if err != nil || len(removed.Cities) != 2 || removed.Cities[1].ID != "c" || len(replaced.Cities) != 3 {
		t.Fatalf("remove: %v %+v", err, removed.Cities)
	}
	if found, err := removed.City("c"); err != nil || found.Label != "New York" {
		t.Errorf("City(c) = %+v, %v", found, err)
	}
	if _, err := removed.City("b"); !errors.Is(err, ErrNoSuchCity) {
		t.Errorf("a removed city is not found: %v", err)
	}
	if _, err := removed.WithoutCity("x"); !errors.Is(err, ErrNoSuchCity) {
		t.Errorf("removing an unknown id: %v", err)
	}
	if _, err := removed.WithCityReplaced(City{ID: "x"}); !errors.Is(err, ErrNoSuchCity) {
		t.Errorf("replacing an unknown id: %v", err)
	}
}

// FR-206: the same place may be added twice under different labels.
func TestTheSamePlaceMayBeAddedTwice(t *testing.T) {
	t.Parallel()
	got := Defaults().
		WithCityAdded(City{ID: "a", GeoNamesID: 2643743, Label: "London"}).
		WithCityAdded(City{ID: "b", GeoNamesID: 2643743, Label: "Mum"})
	if len(got.Cities) != 2 {
		t.Errorf("cities %+v; want both kept", got.Cities)
	}
}

// FR-207: a new place replaces the label only where it was the old place's name.
func TestChangingPlaceKeepsACustomLabel(t *testing.T) {
	t.Parallel()
	london := place.Place{GeoNamesID: 2643743, Name: "London"}
	paris := place.Place{GeoNamesID: 2988507, Name: "Paris"}
	named := Relocated(City{ID: "a", GeoNamesID: london.GeoNamesID, Label: "London"}, london, paris)
	if named.Label != "Paris" || named.GeoNamesID != paris.GeoNamesID {
		t.Errorf("a label that was the place's name follows it: %+v", named)
	}
	custom := Relocated(City{ID: "a", GeoNamesID: london.GeoNamesID, Label: "Mum"}, london, paris)
	if custom.Label != "Mum" || custom.GeoNamesID != paris.GeoNamesID {
		t.Errorf("a typed label is kept: %+v", custom)
	}
	gone := Relocated(City{ID: "a", GeoNamesID: 999, Label: ""}, place.Place{}, paris)
	if gone.Label != "Paris" {
		t.Errorf("an empty label takes the new place's name: %+v", gone)
	}
}
