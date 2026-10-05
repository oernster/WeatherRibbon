// Package units turns MET Norway's measures (degrees Celsius, millimetres, metres a second) into
// the system the user chose (FR-703).
package units

import "math"

// System is how measures are shown.
type System string

// The two systems. The string values are what the settings file holds.
const (
	Metric   System = "metric"
	Imperial System = "imperial"
)

// Systems lists the systems in the order they are offered.
var Systems = []System{Metric, Imperial}

// Conversion factors, each a definition rather than a measurement.
const (
	fahrenheitPerCelsius = 9.0 / 5.0
	fahrenheitAtFreezing = 32
	millimetresPerInch   = 25.4
	metresPerKilometre   = 1000
	metresPerMile        = 1609.344
	secondsPerHour       = 3600
)

// The coldest and hottest air a cell is sized to show, in degrees Celsius (FR-103).
const (
	coldestC = -60
	hottestC = 60
)

// Temperatures answers every whole temperature a cell can show in system, coldest first: every
// degree from the coldest air written in system to the hottest, so imperial holds each Fahrenheit
// degree between rather than only those a whole Celsius degree lands on (FR-103).
func Temperatures(system System) []int {
	coldest, hottest := Temperature(coldestC, system), Temperature(hottestC, system)
	all := make([]int, 0, hottest-coldest+1)
	for degree := coldest; degree <= hottest; degree++ {
		all = append(all, degree)
	}
	return all
}

// Normalise answers system where it is one of Systems; Metric otherwise, so a file holding an
// unknown word still draws (OQ-1).
func Normalise(system System) System {
	if system == Imperial {
		return Imperial
	}
	return Metric
}

// Temperature answers celsius in system, rounded to whole degrees: 14.4 is 14 metric and 58 imperial.
func Temperature(celsius float64, system System) int {
	if Normalise(system) == Imperial {
		return int(math.Round(celsius*fahrenheitPerCelsius + fahrenheitAtFreezing))
	}
	return int(math.Round(celsius))
}

// WindSpeed answers metresPerSecond in kilometres or miles an hour, rounded to whole units: 5 m/s is
// 18 km/h and 11 mph.
func WindSpeed(metresPerSecond float64, system System) int {
	perHour := metresPerSecond * secondsPerHour
	if Normalise(system) == Imperial {
		return int(math.Round(perHour / metresPerMile))
	}
	return int(math.Round(perHour / metresPerKilometre))
}

// Rain answers millimetres in millimetres or inches, unrounded; how many places are shown is the
// page's choice.
func Rain(millimetres float64, system System) float64 {
	if Normalise(system) == Imperial {
		return millimetres / millimetresPerInch
	}
	return millimetres
}
