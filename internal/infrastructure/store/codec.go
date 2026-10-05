package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/oernster/ribbonkit/infrastructure/settingsfile"
	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// formatVersion is written into every file so a later version can tell what it is reading.
const formatVersion = 1

// WeatherRibbon's own top-level keys; the ribbon's are the kit's.
const (
	keyVersion   = "version"
	keyUnits     = "units"
	keyFormat    = "format"
	keyCountdown = "petrichorCountdown"
	keyCities    = "cities"
)

// knownKeys lists the keys this version reads, in writing order (FR-801). WeatherRibbon has no pull
// out, so the kit's pullOutSide is not among them: one written by hand is kept as found.
var knownKeys = []string{
	keyVersion, keyUnits, keyFormat, settingsfile.KeyTheme, settingsfile.KeyColour,
	settingsfile.KeyOrientation, settingsfile.KeyAlwaysOnTop, settingsfile.KeyPlacement,
	settingsfile.KeyPinned, settingsfile.KeyLastEdge, settingsfile.KeyOpacity, settingsfile.KeyScale,
	settingsfile.KeySkippedUpdate, keyCountdown, keyCities,
}

// codec is WeatherRibbon's settings file.
var codec = settingsfile.Codec[settings.Settings]{
	Keys:     knownKeys,
	Defaults: settings.Defaults,
	Decode:   decode,
	Values:   values,
}

// Why a city entry cannot be read, as the user is shown it (FR-804).
var (
	errNoID    = errors.New("it has no id")
	errNoPlace = errors.New("it names no place")
)

// storedCity is one city as the file holds it. Pointers tell a missing field from an empty one.
type storedCity struct {
	ID         *string `json:"id"`
	GeoNamesID *int    `json:"geoNamesId"`
	Label      *string `json:"label"`
	Position   *int    `json:"position"`
}

// cities reads each city entry on its own, so one bad entry is kept as found and the rest load
// (FR-804).
var cities = settingsfile.Entries[settings.City]{
	Decode: decodeCity,
	Encode: func(city settings.City, position int) (json.RawMessage, error) {
		return json.Marshal(storedCity{ID: &city.ID, GeoNamesID: &city.GeoNamesID, Label: &city.Label, Position: &position})
	},
	Unreadable: func(id, reason, original string) settings.City {
		return settings.City{ID: id, Unreadable: reason, Original: original}
	},
	ID:       func(city settings.City) (string, bool) { return city.ID, city.Unreadable == "" },
	WithID:   func(city settings.City, id string) settings.City { city.ID = id; return city },
	Original: func(city settings.City) string { return city.Original },
}

// decode reads a settings file's object. It answers false when the cities are not a list, since then
// nothing can be trusted. Any other bad value leaves its default.
func decode(object settingsfile.Object) (settings.Settings, bool) {
	entries, ok := object.List(keyCities)
	if !ok {
		return settings.Settings{}, false
	}
	decoded := settings.Defaults()
	settingsfile.ReadChoices(object, &decoded.Choices)
	settingsfile.Read(object, keyUnits, &decoded.Units)
	settingsfile.Read(object, keyFormat, &decoded.Format)
	settingsfile.Read(object, keyCountdown, &decoded.PetrichorCountdown)
	decoded.Cities = cities.Read(entries)
	return decoded, true
}

// decodeCity reads one entry with its stored position; the reason where it cannot be read.
func decodeCity(raw json.RawMessage) (settings.City, *int, error) {
	var stored storedCity
	if err := json.Unmarshal(raw, &stored); err != nil {
		return settings.City{}, nil, fmt.Errorf("its fields are not what a city holds (%v)", err)
	}
	switch {
	case stored.ID == nil || *stored.ID == "":
		return settings.City{}, nil, errNoID
	case stored.GeoNamesID == nil:
		return settings.City{}, nil, errNoPlace
	}
	city := settings.City{ID: *stored.ID, GeoNamesID: *stored.GeoNamesID}
	if stored.Label != nil {
		city.Label = *stored.Label
	}
	return city, stored.Position, nil
}

// values answers the value of every key, with any failure writing the cities, which refuses the save.
// Derived values are never written (FR-801).
func values(current settings.Settings) (map[string]any, error) {
	written, err := cities.Write(current.Cities)
	all := map[string]any{
		keyVersion: formatVersion, keyUnits: current.Units, keyFormat: current.Format,
		keyCountdown: current.PetrichorCountdown, keyCities: written,
	}
	maps.Copy(all, settingsfile.ChoiceValues(current.Choices))
	return all, err
}
