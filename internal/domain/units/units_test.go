package units

import "testing"

// FR-703: 14.4 degrees Celsius shows as 14 metric and 58 imperial; 5 m/s as 18 km/h and 11 mph.
func TestEachUnitConvertsAndRounds(t *testing.T) {
	t.Parallel()
	if got := Temperature(14.4, Metric); got != 14 {
		t.Errorf("14.4 C metric is %d; want 14", got)
	}
	if got := Temperature(14.4, Imperial); got != 58 {
		t.Errorf("14.4 C imperial is %d; want 58", got)
	}
	if got := Temperature(-40, Imperial); got != -40 {
		t.Errorf("-40 C imperial is %d; want -40", got)
	}
	if got := WindSpeed(5, Metric); got != 18 {
		t.Errorf("5 m/s metric is %d; want 18", got)
	}
	if got := WindSpeed(5, Imperial); got != 11 {
		t.Errorf("5 m/s imperial is %d; want 11", got)
	}
	if got := Rain(25.4, Imperial); got != 1 {
		t.Errorf("25.4 mm imperial is %v; want 1 inch", got)
	}
	if got := Rain(3.2, Metric); got != 3.2 {
		t.Errorf("3.2 mm metric is %v; want 3.2", got)
	}
}

// FR-103: the samples hold every whole degree from -60 to 60 Celsius written in the units, -76 to
// 140 imperial, with no degree skipped between.
func TestSamplesHoldEveryTemperature(t *testing.T) {
	t.Parallel()
	for system, want := range map[System][2]int{Metric: {-60, 60}, Imperial: {-76, 140}} {
		got := Temperatures(system)
		if len(got) != want[1]-want[0]+1 || got[0] != want[0] || got[len(got)-1] != want[1] {
			t.Fatalf("%s: %d degrees from %d to %d; want %d to %d", system, len(got), got[0], got[len(got)-1], want[0], want[1])
		}
		for index := 1; index < len(got); index++ {
			if got[index] != got[index-1]+1 {
				t.Errorf("%s: %d follows %d", system, got[index], got[index-1])
			}
		}
	}
}

// OQ-1: an unknown system reads as Metric, the default.
func TestAnUnknownSystemIsMetric(t *testing.T) {
	t.Parallel()
	for _, given := range []System{"", "kelvin", Metric} {
		if got := Normalise(given); got != Metric {
			t.Errorf("Normalise(%q) = %q; want metric", given, got)
		}
	}
	if got := Temperature(14.4, "kelvin"); got != 14 {
		t.Errorf("an unknown system converts as metric: got %d", got)
	}
	if len(Systems) != 2 || Systems[0] != Metric {
		t.Errorf("Systems = %v; want metric first", Systems)
	}
}
