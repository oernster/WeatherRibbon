package places

import (
	"strings"
	"testing"
)

// listedPlaces is how many places GeoNames' cities15000 held when the list was built (DATA-1).
const listedPlaces = 34153

// newPlaces answers the places over the embedded list.
func newPlaces(t *testing.T) *Places {
	t.Helper()
	places, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return places
}

// Every place of the embedded list is read, each with a name, a country and a zone the built-in tz
// database resolves (DATA-1, FR-401).
func TestEveryPlaceIsReadAndItsZoneResolves(t *testing.T) {
	t.Parallel()
	places := newPlaces(t)
	if len(places.byID) != listedPlaces {
		t.Fatalf("read %d places, want %d", len(places.byID), listedPlaces)
	}
	for id, each := range places.byID {
		if each.Name == "" || each.Country == "" {
			t.Errorf("%d has no name or country: %+v", id, each)
		}
		if _, err := places.Resolve(each.Zone); err != nil {
			t.Errorf("%d: %v", id, err)
		}
	}
}

// FR-202's acceptance: London in England before London in Ontario; eight US Springfields, each with
// its state.
func TestTheSearchExamplesHold(t *testing.T) {
	t.Parallel()
	places := newPlaces(t)
	found := places.Search("london")
	england, ontario := -1, -1
	for index, each := range found {
		switch each.Description() {
		case "London, England, United Kingdom":
			england = index
		case "London, Ontario, Canada":
			ontario = index
		}
	}
	if england < 0 || ontario < 0 || england > ontario {
		t.Errorf("London in England at %d, in Ontario at %d", england, ontario)
	}
	springfields := 0
	for _, each := range places.Search("springf") {
		if each.Name == "Springfield" && each.Country == "United States" && each.Region != "" {
			springfields++
		}
	}
	if springfields != 8 {
		t.Errorf("%d US Springfields with a state, want 8", springfields)
	}
}

// A stored GeoNames id is found; one the list does not hold is not (FR-803).
func TestAPlaceIsFoundByItsGeoNamesID(t *testing.T) {
	t.Parallel()
	places := newPlaces(t)
	if london, ok := places.Place(2643743); !ok || london.Description() != "London, England, United Kingdom" {
		t.Errorf("answered %+v, %v", london, ok)
	}
	if _, ok := places.Place(0); ok {
		t.Error("an id the list does not hold was found")
	}
}

// A zone the tz database does not know is refused.
func TestAnUnknownZoneIsRefused(t *testing.T) {
	t.Parallel()
	if _, err := newPlaces(t).Resolve("Not/AZone"); err == nil {
		t.Error("an unknown zone resolved")
	}
}

// A list line that is not a place is refused, naming its line.
func TestAMalformedLineIsRefused(t *testing.T) {
	t.Parallel()
	good := "1\tx\tx\tr\tc\t1.5\t2.5\t100\tEurope/London"
	if places, err := fromText("# head\r\n" + good + "\r\n"); err != nil || len(places.byID) != 1 {
		t.Errorf("a good list answered %v", err)
	}
	for name, line := range map[string]string{
		"too few columns":    "1\tx",
		"an id not a number": strings.Replace(good, "1\t", "one\t", 1),
		"no coordinates":     strings.Replace(good, "1.5", "north", 1),
		"no population":      strings.Replace(good, "100", "many", 1),
	} {
		if _, err := fromText("# head\n" + line); err == nil || !strings.Contains(err.Error(), "cities.tsv line 2") {
			t.Errorf("%s: answered %v", name, err)
		}
	}
}
