//go:build windows

package main

import (
	"io/fs"
	"testing"

	"github.com/oernster/ribbonkit/installer"
)

// The setup page shows WeatherRibbon's pictures; one the build does not carry shows broken.
func TestWeatherRibbonCarriesEveryPictureTheSetupPageShows(t *testing.T) {
	shown, err := fs.Sub(pictures, picturesRoot)
	if err != nil {
		t.Fatal(err)
	}
	if missing := installer.Missing(shown); len(missing) != 0 {
		t.Errorf("the setup program carries no %v; tools/genicons.py makes them", missing)
	}
}
