package main

// The facade the page calls. The ribbon's window is ribbonkit's, embedded here so its methods are page
// API alongside these; what is WeatherRibbon's own (its cities, their units and time format, the
// detail panel, the petrichor line and the measuring of their text) is here. Every method runs one use
// case, then has the window do what it needs afterwards. Methods that can be refused answer an error,
// which the page hears as a refusal.

import (
	"context"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/ui/window"
	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
	"github.com/oernster/weatherribbon/internal/product"
)

// openAtAddCity asks the page to open Settings on the place search (FR-106, FR-201).
const openAtAddCity = "add-city"

// detailPanel is the word the page names the hourly detail by, the window's own panel (FR-410).
const detailPanel = "detail"

// ribbonService is what WeatherRibbon's own half of the facade asks of the application layer:
// application.Service in production, a scripted stand-in in the facade's tests. The window asks for
// the rest through window.Service.
type ribbonService interface {
	Snapshot() application.Snapshot
	Settings() settings.Settings
	Search(typed string) []place.Place
	AddCity(geoNamesID int) (string, error)
	Relabel(id, typed string) error
	ChangePlace(id string, geoNamesID int) error
	RemoveCity(id string) error
	SetUnits(system units.System) error
	SetFormat(format localtime.Format) error
	DismissNotices()
	DismissPetrichor(id string)
	OpenDetail(ctx context.Context, id string) (application.Detail, error)
	TextSamples() application.Samples
	SetMeasured(measured application.Measured) error
	SettingsChoices() []menus.Item
}

// windowControl is what WeatherRibbon's own half asks of the window: the Control's calls, plus the
// two ribbon choices its menus carry. kitWindow is the real one; the facade's tests stand in for it.
type windowControl interface {
	Refitted(err error) error
	Redraw()
	Report(doing string, err error)
	ShowPanel(panel string)
	PageMeasured()
	Shown() window.Shown
	Offered(items []menus.Item) []menus.Item
	SetColour(colour string) error
	SetOrientation(orientation string) error
}

// fetcher is what the facade asks of the forecasts' own goroutine: to look again at what is due (as
// after a city is added or moved); to ask now for every city backing off (FR-309).
type fetcher interface {
	poke()
	askNow()
}

// kitWindow is the window's two halves together, as WeatherRibbon's own half reaches them.
type kitWindow struct {
	*window.Window
	*window.Control
}

// App is the facade Wails binds.
type App struct {
	*window.Window
	service ribbonService
	control windowControl
	fetch   fetcher
	// ctx bounds the requests the detail panel makes; it ends with the run.
	ctx context.Context
}

// newApp answers the facade over service, its window built from config with WeatherRibbon's product
// and menu actions, plus the Control the composition root runs it with.
func newApp(ctx context.Context, service ribbonService, fetch fetcher, config window.Config) (*App, *window.Control) {
	app := &App{service: service, fetch: fetch, ctx: ctx}
	config.Act = app.actOn
	config.Product = productOf()
	shown, control := window.New(config)
	app.Window, app.control = shown, kitWindow{shown, control}
	return app, control
}

// Snapshot answers what the ribbon shows now, how the window shows it included (FR-706).
func (a *App) Snapshot() snapshotDTO {
	seen := a.control.Shown()
	shown := snapshotOf(a.service.Snapshot())
	shown.Scrolls, shown.DragThreshold, shown.Collapsed = seen.Scrolls, sizeOf(seen.DragThreshold), seen.Collapsed
	shown.Choices = window.ChoicesOf(a.control.Offered(a.service.SettingsChoices()))
	return shown
}

// SearchPlaces answers the places matching typed, best first (FR-202).
func (a *App) SearchPlaces(typed string) []placeDTO { return placesOf(a.service.Search(typed)) }

// AddCity adds a city for the place with the GeoNames id and answers its id (FR-201); its forecast
// is asked for at once.
func (a *App) AddCity(geoNamesID int) (string, error) {
	id, err := a.service.AddCity(geoNamesID)
	a.fetch.poke()
	return id, a.control.Refitted(err)
}

// RenameCity stores what the user typed as a city's label (FR-204).
func (a *App) RenameCity(id, typed string) error {
	return a.control.Refitted(a.service.Relabel(id, typed))
}

// ChangePlace moves a city to another place (FR-207); its forecast is asked for at once.
func (a *App) ChangePlace(id string, geoNamesID int) error {
	err := a.service.ChangePlace(id, geoNamesID)
	a.fetch.poke()
	return a.control.Refitted(err)
}

// RemoveCity removes a city; the page has already asked (FR-205).
func (a *App) RemoveCity(id string) error {
	return a.control.Refitted(a.service.RemoveCity(id))
}

// SetUnits chooses metric or imperial (FR-703).
func (a *App) SetUnits(system string) error {
	return a.control.Refitted(a.service.SetUnits(units.System(system)))
}

// SetFormat chooses 12-hour or 24-hour time (FR-401).
func (a *App) SetFormat(format string) error {
	return a.control.Refitted(a.service.SetFormat(localtime.Format(format)))
}

// DismissNotices clears the notices the user has read, then fits the ribbon without their cells.
func (a *App) DismissNotices() {
	a.service.DismissNotices()
	_ = a.control.Refitted(nil)
}

// DismissPetrichor hides a city's petrichor line (FR-413).
func (a *App) DismissPetrichor(id string) { a.service.DismissPetrichor(id) }

// OpenDetail answers the detail panel for a city (FR-410, FR-411).
func (a *App) OpenDetail(id string) (detailDTO, error) {
	detail, err := a.service.OpenDetail(a.ctx, id)
	return detailOf(detail), err
}

// actOn carries out a menu action of WeatherRibbon's own (FR-109), then has the page redraw, since a
// choice made from a menu is one the page did not make. The window hands over every action it does
// not know; one that is not WeatherRibbon's either changes nothing.
func (a *App) actOn(action menus.Action) {
	switch action {
	case application.ActionAddCity:
		a.control.ShowPanel(openAtAddCity)
		return
	case application.ActionRefreshNow:
		a.fetch.askNow()
		return
	}
	if choose, doing, ok := a.choiceOf(action); ok {
		a.control.Report(doing, choose())
		a.control.Redraw()
	}
}

// choiceOf answers the choice action names with what making it is doing; false where action names
// none of WeatherRibbon's choices.
func (a *App) choiceOf(action menus.Action) (choose func() error, doing string, ok bool) {
	if system, ok := application.UnitsOf(action); ok {
		return func() error { return a.SetUnits(string(system)) }, "changing the units", true
	}
	if colour, ok := menus.ColourOf(action); ok {
		return func() error { return a.control.SetColour(string(colour)) }, "changing the colour", true
	}
	if orientation, ok := menus.OrientationOf(action); ok {
		return func() error { return a.control.SetOrientation(string(orientation)) }, "changing the orientation", true
	}
	return nil, "", false
}

// productOf answers what the window says about WeatherRibbon, every word from internal/product and
// the terms the LICENSE file itself (FR-603, FR-610).
func productOf() window.Product {
	credits := make([]window.Credit, 0, len(product.Credits))
	for _, credit := range product.Credits {
		credits = append(credits, window.Credit{Name: credit.Name, Licence: credit.Licence, Role: credit.Role})
	}
	return window.Product{
		App: product.App(), WindowClass: product.RibbonClass, Version: product.Version,
		Author: product.Author, Copyright: product.Copyright, Credits: credits,
		Licence: licenceText, DonateURL: product.DonateURL,
	}
}
