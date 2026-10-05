package place

import (
	"strconv"
	"testing"
)

// FR-302: Kathmandu at 27.70169, 85.3206 is asked for as lat=27.7017&lon=85.3206.
func TestCoordinatesAreRoundedToFourPlaces(t *testing.T) {
	t.Parallel()
	for given, want := range map[float64]string{
		27.70169: "27.7017",
		85.3206:  "85.3206",
		-0.12784: "-0.1278",
		51.50735: "51.5074",
	} {
		if got := strconv.FormatFloat(RoundDegrees(given), 'f', -1, 64); got != want {
			t.Errorf("RoundDegrees(%v) writes as %s; want %s", given, got, want)
		}
	}
}
