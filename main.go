// Command WeatherRibbon shows a ribbon of cities, each with its local time, its weather now and the
// days ahead.
//
// This file is the composition root, the only file permitted to import both the application layer
// and concrete infrastructure (TestCompositionRootIsWhitelisted). The window reaches the desktop
// through the shell.Desktop port it is handed here.
package main

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/application/release"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/infrastructure/appdata"
	"github.com/oernster/ribbonkit/infrastructure/desktop"
	"github.com/oernster/ribbonkit/infrastructure/monitors"
	"github.com/oernster/ribbonkit/infrastructure/occupancy"
	"github.com/oernster/ribbonkit/infrastructure/platform"
	"github.com/oernster/ribbonkit/infrastructure/runlog"
	"github.com/oernster/ribbonkit/infrastructure/startup"
	"github.com/oernster/ribbonkit/infrastructure/system"
	"github.com/oernster/ribbonkit/infrastructure/update"
	"github.com/oernster/ribbonkit/ui/window"
	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/infrastructure/cache"
	"github.com/oernster/weatherribbon/internal/infrastructure/metno"
	"github.com/oernster/weatherribbon/internal/infrastructure/pacing"
	"github.com/oernster/weatherribbon/internal/infrastructure/places"
	"github.com/oernster/weatherribbon/internal/infrastructure/store"
	"github.com/oernster/weatherribbon/internal/product"
)

// layout is the size of one cell and of the empty ribbon's one cell, with the padding round the
// cells, in DIP: its one home. A cell is the least it is drawn at; it is widened to its widest text
// as the page measures it (FR-103) and the whole ribbon drawn at the chosen scale (FR-705). app.css
// sizes the cell's text to fit.
var layout = application.Layout{
	Cell:    placement.Size{Width: 196, Height: 212},
	Prompt:  placement.Size{Width: 196, Height: 212},
	Padding: 6,
}

// panels are the window's sizes in DIP while it shows a panel (CON-6): Settings wide enough for its
// choices to sit side by side; About, Licence, the update panel and a city's hourly detail narrower,
// for their text. Each fits its content's height once open.
var panels = window.PanelSizes{
	Settings: placement.Size{Width: 900, Height: 760},
	Other:    placement.Size{Width: 560, Height: 760},
	Own:      map[string]placement.Size{detailPanel: {Width: 560, Height: 760}},
}

func main() {
	if platform.GeneratingBindings {
		app, control := newApp(context.Background(), nil, nil, window.Config{Log: io.Discard, Panels: panels})
		if err := control.Run(app, assets, ""); err != nil {
			os.Exit(1)
		}
		return
	}
	log := keepLog()
	if err := run(log); err != nil {
		fmt.Fprintf(log, "%s: %v\n", product.Name, err)
		os.Exit(1)
	}
}

// keepLog opens the log and points standard error at it before anything can fail (NFR-O-1). A log
// that cannot be opened leaves standard error as it is; the run goes on.
func keepLog() io.Writer {
	dir, err := settingsDir()
	if err != nil {
		return os.Stderr
	}
	log, err := runlog.Open(dir, product.App(), time.Now())
	if err != nil {
		return os.Stderr
	}
	if err := runlog.Keep(log); err != nil {
		fmt.Fprintln(log, err)
	}
	return log
}

// settingsDir answers the settings folder appdata names; a folder of the same name in the temporary
// folder when the environment names none, so the ribbon still opens.
func settingsDir() (string, error) {
	dir, err := appdata.Dir(product.App(), os.LookupEnv)
	if err != nil {
		return filepath.Join(os.TempDir(), product.Name), err
	}
	return dir, nil
}

// openNeighbours takes WeatherRibbon's place in the folder every running ribbon shares, read from the
// environment through lookup (FR-506); where the environment names none it says so and is alone.
func openNeighbours(lookup func(string) (string, bool), log io.Writer) *occupancy.Folder {
	shared, err := occupancy.Dir(lookup)
	if err != nil {
		fmt.Fprintf(log, "%v: keeping off no other ribbon\n", err)
	}
	return occupancy.Open(shared, product.App(), log)
}

// run wires everything together and hands the facade to Wails. Only a failure to run the window at
// all ends it, as does a city list or icon set built into the binary that cannot be read; every
// other fault is carried to the ribbon or the log.
func run(log io.Writer) error {
	dir, err := settingsDir()
	if err != nil {
		fmt.Fprintf(log, "%v: keeping settings in %s\n", err, dir)
	}
	cityList, err := places.New()
	if err != nil {
		return err
	}
	icons, err := newIconSet(weatherIcons, iconPattern, log)
	if err != nil {
		return err
	}
	program, err := os.Executable()
	if err != nil {
		fmt.Fprintf(log, "finding this program's path: %v\n", err)
	}
	neighbours := openNeighbours(os.LookupEnv, log)
	defer neighbours.Close()
	service := application.New(application.Ports{
		Store:      store.New(dir, product.Name),
		Places:     cityList,
		Forecasts:  metno.New(product.UserAgent(), log),
		SunTimes:   metno.NewSunTimes(product.UserAgent(), log),
		Cache:      cache.New(dir, log),
		Clock:      system.Clock{},
		Pacer:      pacing.Pacer{},
		IDs:        system.IDs{},
		Draw:       rand.IntN,
		Icons:      icons,
		Monitors:   monitors.Monitors{},
		Neighbours: neighbours,
		Startup:    startup.New(product.App(), program),
		Releases:   update.New(product.Repository),
		Build:      release.Build{Version: product.Version, Platform: release.PlatformKeyFor(runtime.GOOS)},
	}, layout)
	if err := service.Start(); err != nil {
		fmt.Fprintf(log, "loading settings: %v\n", err)
	}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	var control *window.Control
	fetch := newRefresher(service.NextDue, service.Refresh, service.RefreshNow, func() {
		_ = control.Refitted(nil)
		control.Redraw()
	}, func(doing string, err error) { control.Report(doing, err) }, service.RefreshStopped)
	desk := desktop.New(product.App(), func() []menus.Item { return control.Offered(service.TrayMenu(control.Visible())) }, log)
	app, control := newApp(ctx, service, fetch, window.Config{Service: kitService{service}, Desktop: desk, Log: log, Panels: panels})
	platform.Prepare(desk, trayIcon, control.ExitWhen)
	if err := desk.Start(); err != nil {
		fmt.Fprintf(log, "starting the tray icon: %v; closing the ribbon will exit\n", err)
	} else {
		control.TrayStarted()
	}
	go fetch.run(ctx)
	return control.Run(app, assets, dir)
}
