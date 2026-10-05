//go:build windows

// Command installer is WeatherRibbon's setup program (REQUIREMENTS.md section 5, as
// TimeRibbon's FR-801 to FR-811).
//
// It is a second Wails application in the module, carrying the built application as an embedded
// payload. It installs, updates, goes back a version, repairs, reinstalls and uninstalls, all per
// user with no administrator rights. The window, its page, its wiring and the install policy are
// ribbonkit's (ribbonkit/installer over ribbonkit/infrastructure/setup), which name no product; this
// is the composition root that names it and carries what is WeatherRibbon's own.
package main

import (
	"embed"
	"os"

	"github.com/oernster/ribbonkit/infrastructure/setup"
	"github.com/oernster/ribbonkit/installer"
	"github.com/oernster/weatherribbon/internal/product"
)

// pictures are the page's pictures, which tools/genicons.py makes from WeatherRibbon's artwork: the
// header mark and the theme switch's sun and moon.
//
//go:embed all:frontend/dist
var pictures embed.FS

// picturesRoot is the folder pictures holds them under.
const picturesRoot = "frontend/dist"

// payload is the built application as a zip archive, embedded as a string so Go keeps it in the
// read-only image rather than charging it to the process. build.ps1 writes the real one before it
// builds and puts the empty placeholder back after.
//
//go:embed payload.zip
var payload string

// App is what Wails binds, so the page reaches the kit's setup facade as main.App.
type App struct{ *installer.Setup }

func main() {
	os.Exit(installer.Main(installer.Program{
		Product:      setup.Product{App: product.App(), Publisher: product.Author},
		SetupID:      product.SetupName,
		RibbonClass:  product.RibbonClass,
		Version:      product.Version,
		Payload:      payload,
		Pictures:     pictures,
		PicturesRoot: picturesRoot,
		Bind:         func(facade *installer.Setup) any { return &App{facade} },
	}))
}
