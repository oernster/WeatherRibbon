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
