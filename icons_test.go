package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

// FR-412, ASM-5: the set built in holds all 83 icons, light sleet showers and thunder under its
// doubled s among them; a code it lacks is logged in words a reader can follow.
func TestTheBuiltInIconSetHoldsEveryIcon(t *testing.T) {
	var log strings.Builder
	icons, err := newIconSet(weatherIcons, iconPattern, &log)
	if err != nil {
		t.Fatal(err)
	}
	const yrIcons = 83
	if len(icons.names) != yrIcons || !icons.Has("clearsky_day") || !icons.Has("lightssleetshowersandthunder_day") || icons.Has("clearsky_day.svg") {
		t.Errorf("%d icons", len(icons.names))
	}
	icons.Missing("sandstorm")
	if got := log.String(); !strings.Contains(got, `"sandstorm"`) {
		t.Errorf("logged %q", got)
	}
}

// A pattern that cannot be read is answered, not hidden.
func TestAnUnreadableIconPatternIsAnswered(t *testing.T) {
	if _, err := newIconSet(fstest.MapFS{}, "[", &strings.Builder{}); err == nil {
		t.Error("a malformed pattern was accepted")
	}
}
