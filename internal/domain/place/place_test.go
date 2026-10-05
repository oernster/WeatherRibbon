package place

import (
	"strings"
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/ribbon"
)

// FR-202: a place is described by its name, region and country; by name and country where GeoNames
// names no region.
func TestAPlaceIsDescribedWithItsRegionWhereThereIsOne(t *testing.T) {
	t.Parallel()
	london := Place{Name: "London", Region: "England", Country: "United Kingdom"}
	if got := london.Description(); got != "London, England, United Kingdom" {
		t.Errorf("description %q", got)
	}
	unnamed := Place{Name: "Vaduz", Country: "Liechtenstein"}
	if got := unnamed.Description(); got != "Vaduz, Liechtenstein" {
		t.Errorf("description without a region %q", got)
	}
}

// FR-204: a label empty after trimming stores the place's name; a typed one is capped by the kit.
func TestEmptyLabelFallsBackToThePlaceName(t *testing.T) {
	t.Parallel()
	london := Place{Name: "London"}
	for _, typed := range []string{"", "  "} {
		if got := london.Label(typed); got != "London" {
			t.Errorf("Label(%q) = %q; want the place's name", typed, got)
		}
	}
	if got := london.Label(" Mum "); got != "Mum" {
		t.Errorf("a typed label is kept trimmed: got %q", got)
	}
	if got := []rune(london.Label(strings.Repeat("x", ribbon.MaxLabelLength+1))); len(got) != ribbon.MaxLabelLength {
		t.Errorf("a long label holds %d characters; want %d", len(got), ribbon.MaxLabelLength)
	}
}

// FR-401: at 2026-10-05T07:36:00Z London shows 08:36 BST.
func TestLocalTimeAndZoneMark(t *testing.T) {
	t.Parallel()
	london, err := time.LoadLocation("Europe/London")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, time.October, 5, 7, 36, 0, 0, time.UTC)
	got := ClockAt(at, london, localtime.TwentyFourHour)
	if got.Time != "08:36" || got.ZoneMark != "BST" || got.OffsetSeconds != 3600 {
		t.Errorf("London at 07:36Z = %+v; want 08:36 BST, an hour ahead", got)
	}
	if twelve := ClockAt(at, london, localtime.TwelveHour); twelve.Time != "8:36 AM" {
		t.Errorf("in 12-hour time %q; want 8:36 AM", twelve.Time)
	}
}
