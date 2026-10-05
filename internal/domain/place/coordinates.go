// Package place holds what WeatherRibbon knows about a city's place apart from its weather.
package place

import "math"

// coordinateScale is ten to the power of the decimal places MET Norway's terms allow in a request's
// latitude and longitude: four (FR-302).
const coordinateScale = 10000

// RoundDegrees answers degrees rounded to the four decimal places a request may carry (FR-302):
// Kathmandu's 27.70169 is asked for as 27.7017.
func RoundDegrees(degrees float64) float64 {
	return math.Round(degrees*coordinateScale) / coordinateScale
}
