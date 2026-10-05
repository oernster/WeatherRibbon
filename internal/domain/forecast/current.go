package forecast

import "time"

// Current is what a cell shows as now (FR-402).
type Current struct {
	// AirC is the current step's air temperature in degrees Celsius.
	AirC float64
	// Symbol is the symbol code of the current step's next_1_hours block, else its next_6_hours
	// block; empty when the step carries neither.
	Symbol string
	// RainMM is the same block's rain in millimetres; HasRain says whether it gave one (FR-413).
	RainMM  float64
	HasRain bool
}

// currentIndex answers the index of the latest step at or before now; -1 when there is none.
func (f Forecast) currentIndex(now time.Time) int {
	found := -1
	for index, each := range f.steps {
		if each.Time.After(now) {
			break
		}
		found = index
	}
	return found
}

// CurrentAt answers the conditions at now: the latest step at or before now (FR-402). It answers
// false when no step lies at or before now or when now is past the end of the forecast's last block,
// so an old value is never shown as current (FR-403).
func (f Forecast) CurrentAt(now time.Time) (Current, bool) {
	index := f.currentIndex(now)
	if index < 0 || now.After(f.end()) {
		return Current{}, false
	}
	step := f.steps[index]
	current := Current{AirC: step.AirC}
	for _, block := range []Period{step.Next1, step.Next6} {
		if block.Present() {
			current.Symbol, current.RainMM, current.HasRain = block.Symbol, block.RainMM, block.HasRain
			break
		}
	}
	return current, true
}
