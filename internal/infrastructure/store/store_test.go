package store

// What every ribbon's settings file does alike (a file kept aside, never saved over when it could not
// be read, unknown keys kept, a byte order mark, ids told apart, ordering by position, atomic writes)
// is proved once, in the kit's settingsfile. These tests prove what WeatherRibbon's file holds.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/ribbonkit/infrastructure/settingsfile"
	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// productName is the name the store is handed; the composition root hands it the product's own.
const productName = "WeatherRibbon"

// write puts text in dir's settings file.
func write(t *testing.T, dir, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, settingsfile.FileName), []byte(text), settingsfile.FileMode); err != nil {
		t.Fatal(err)
	}
}

// read answers dir's settings file.
func read(t *testing.T, dir string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, settingsfile.FileName))
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

// full answers settings with every value away from its default: the settings of the 1.0.0 fixture.
func full() settings.Settings {
	s := settings.Settings{
		Choices: ribbon.Choices{
			Colour: ribbon.Sunset, Orientation: ribbon.Horizontal, Theme: ribbon.Dark, AlwaysOnTop: true,
			SkippedUpdate: "v1.2.0",
			Placement: &placement.Stored{
				Device: `\\.\DISPLAY2`, Work: placement.Rect{Left: 1920, Right: 4480, Bottom: 1392},
				DPI: 144, Offset: placement.Point{X: 180, Y: -4},
			},
			LastEdge: &placement.Against{Device: `\\.\DISPLAY2`, Edge: placement.Left},
			Opacity:  55, Scale: 150,
		},
		Units: units.Imperial, Format: localtime.TwelveHour, PetrichorCountdown: 7,
	}
	s = s.WithCityAdded(settings.City{ID: "Q2ZJ7L4MVXH3KSD5TW6RYBNPEA", GeoNamesID: 2643743, Label: "London"})
	return s.WithCityAdded(settings.City{ID: "ABCDEFGHIJKLMNOPQRSTUVWXYZ", GeoNamesID: 3143244, Label: "Mum"})
}

// FR-801: every value is written and read back as it was.
func TestSettingsRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := New(dir, productName).Save(full()); err != nil {
		t.Fatal(err)
	}
	loaded, err := New(dir, productName).Load()
	if err != nil || loaded.Notice != "" {
		t.Fatalf("load: %v %q", err, loaded.Notice)
	}
	assertSettings(t, loaded.Settings, full())
}

// assertSettings fails unless got holds every value want does.
func assertSettings(t *testing.T, got, want settings.Settings) {
	t.Helper()
	if got.Units != want.Units || got.Format != want.Format || got.PetrichorCountdown != want.PetrichorCountdown ||
		len(got.Cities) != len(want.Cities) || got.Cities[0] != want.Cities[0] || got.Cities[1] != want.Cities[1] {
		t.Errorf("got %+v", got)
	}
	if got.Colour != want.Colour || *got.Placement != *want.Placement || *got.LastEdge != *want.LastEdge ||
		got.Opacity != want.Opacity || got.Scale != want.Scale || got.Pinned != want.Pinned ||
		got.SkippedUpdate != want.SkippedUpdate || got.AlwaysOnTop != want.AlwaysOnTop {
		t.Errorf("the ribbon's choices read as %+v", got.Choices)
	}
}

// FR-801: the file is indented, in writing order, holding nothing derived and no pull out, which
// WeatherRibbon does not have.
func TestNoDerivedValueIsStored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := New(dir, productName).Save(full()); err != nil {
		t.Fatal(err)
	}
	text := read(t, dir)
	for _, derived := range []string{"forecast", "temperature", "symbol", "\"time\"", "\"zone\"", "latitude", settingsfile.KeyPullOutSide} {
		if strings.Contains(text, derived) {
			t.Errorf("the file holds %q:\n%s", derived, text)
		}
	}
	if !strings.HasPrefix(text, "{\n  \"version\": 1,\n  \"units\": \"imperial\",\n  \"format\": \"12h\",") {
		t.Errorf("not indented in writing order:\n%s", text)
	}
	if !strings.Contains(text, `"position": 1`) {
		t.Errorf("positions are not written:\n%s", text)
	}
}

// FR-802 as TimeRibbon's FR-703: no file is the defaults with no notice.
func TestAbsentFileMeansDefaults(t *testing.T) {
	t.Parallel()
	loaded, err := New(t.TempDir(), productName).Load()
	if err != nil || loaded.Notice != "" || loaded.Settings.Units != units.Metric || len(loaded.Settings.Cities) != 0 {
		t.Errorf("got %+v (%v)", loaded, err)
	}
}

// FR-802 as TimeRibbon's FR-704: cities that are not a list mean nothing in the file can be trusted,
// so it is kept aside.
func TestCitiesThatAreNotAListKeepTheFileAside(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"cities": {"id": "a"}}`)
	loaded, err := New(dir, productName).Load()
	if err != nil || loaded.Notice != settingsfile.KeptAsideNotice(settingsfile.UnreadableName) || len(loaded.Settings.Cities) != 0 {
		t.Errorf("got %+v (%v)", loaded, err)
	}
}

// oneBadCity is FR-804's acceptance file: three cities, the second holding a GeoNames id that is not a
// number.
const oneBadCity = `{"cities": [
	{"id": "a", "geoNamesId": 2643743, "label": "London", "position": 0},
	{"id": "b", "geoNamesId": "Oslo", "label": "Gran", "position": 1},
	{"id": "c", "geoNamesId": 3143244, "label": "Oslo", "position": 2},
	{"geoNamesId": 1850147},
	{"id": "e", "label": "no place"}
]}`

// FR-804: the first and third load; the second, like any entry with no id or no place, is unreadable
// with its reason and keeps its place.
func TestOneBadCityLeavesTheOthersWorking(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, oneBadCity)
	loaded, err := New(dir, productName).Load()
	cities := loaded.Settings.Cities
	if err != nil || len(cities) != 5 || cities[0] != (settings.City{ID: "a", GeoNamesID: 2643743, Label: "London"}) ||
		cities[2] != (settings.City{ID: "c", GeoNamesID: 3143244, Label: "Oslo"}) {
		t.Fatalf("cities %+v (%v)", cities, err)
	}
	for index, reason := range map[int]string{1: "fields are not what a city holds", 3: "has no id", 4: "names no place"} {
		if !strings.Contains(cities[index].Unreadable, reason) || !strings.HasPrefix(cities[index].ID, settingsfile.UnreadableIDPrefix) {
			t.Errorf("entry %d: %+v", index, cities[index])
		}
	}
}

// FR-804: a save writes the unreadable city back exactly as it was found.
func TestAnUnreadableCityIsWrittenBackAsFound(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, oneBadCity)
	store := New(dir, productName)
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(loaded.Settings); err != nil {
		t.Fatal(err)
	}
	var file struct{ Cities []json.RawMessage }
	if err := json.Unmarshal([]byte(read(t, dir)), &file); err != nil || len(file.Cities) != 5 {
		t.Fatalf("the file reads as %+v (%v)", file, err)
	}
	var written, found bytes.Buffer
	bad := `{"id": "b", "geoNamesId": "Oslo", "label": "Gran", "position": 1}`
	if json.Compact(&written, file.Cities[1]) != nil || json.Compact(&found, []byte(bad)) != nil || written.String() != found.String() {
		t.Errorf("the bad entry was written back as %s", file.Cities[1])
	}
}

// A bad value leaves its default and the rest load.
func TestABadValueLeavesItsDefaultAndTheRestLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, dir, `{"units": 7, "theme": "dark", "petrichorCountdown": "soon", "cities": null}`)
	loaded, err := New(dir, productName).Load()
	got := loaded.Settings
	if err != nil || got.Units != units.Metric || got.Theme != ribbon.Dark || got.PetrichorCountdown != 0 {
		t.Errorf("got %+v (%v)", got, err)
	}
}
