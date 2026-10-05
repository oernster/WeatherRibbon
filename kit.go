package main

// What WeatherRibbon hands ribbonkit's window: the page it serves, the terms Help shows and its
// application service as the window asks for it.

import (
	"embed"
	"errors"

	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/weatherribbon/internal/application"
)

// assets is the built page the window serves.
//
//go:embed all:frontend/dist
var assets embed.FS

// licenceText is the LICENSE file itself, so the Licence panel cannot disagree with what was built
// (FR-603).
//
//go:embed LICENSE
var licenceText string

// errNoPullOut is answered when the window is asked to open a pull out WeatherRibbon does not have.
var errNoPullOut = errors.New("WeatherRibbon has no pull out")

// kitService is WeatherRibbon's application service as the window asks for it: the ribbon's choices
// read out of WeatherRibbon's settings and no pull out, every other call the service's own.
type kitService struct {
	*application.Service
}

// Choices answers the ribbon's own choices out of WeatherRibbon's settings.
func (k kitService) Choices() ribbon.Choices { return k.Settings().Choices }

// PullOut answers false: no pull out stands beside WeatherRibbon's ribbon.
func (kitService) PullOut() bool { return false }

// SetPullOut refuses: the page offers no handle to ask for one with.
func (kitService) SetPullOut(bool) error { return errNoPullOut }
