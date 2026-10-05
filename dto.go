package main

// WeatherRibbon's half of the wire between Go and the page; ribbonkit's window states its own half
// (ribbonkit/ui/window/wire.go). Each type here is stated a second time in frontend/src/wire.ts; a
// structural test compares the two, since the type checker sees only the TypeScript and the
// marshaller sees only these. Every list goes out present, never null, so the page never meets one.

import (
	"fmt"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/forecast"
	"github.com/oernster/weatherribbon/internal/domain/place"
	"github.com/oernster/weatherribbon/internal/product"
)

// sizeDTO is a width and a height in DIP.
type sizeDTO struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// layoutDTO is the cell geometry the page draws with (FR-103, FR-106).
type layoutDTO struct {
	Cell    sizeDTO `json:"cell"`
	Prompt  sizeDTO `json:"prompt"`
	Padding int     `json:"padding"`
}

// textSamplesDTO is every text a cell can show whose width varies (FR-103).
type textSamplesDTO struct {
	Times        []string `json:"times"`
	Weekdays     []string `json:"weekdays"`
	Temperatures []int    `json:"temperatures"`
	Labels       []string `json:"labels"`
}

// measuredDTO is the cell width the page measured its widest text to need, with what it measured
// under (FR-103).
type measuredDTO struct {
	Units     string   `json:"units"`
	Format    string   `json:"format"`
	Labels    []string `json:"labels"`
	CellWidth int      `json:"cellWidth"`
}

// symbolDTO is a weather symbol: the icon drawn for it, else its words in its place (FR-412).
type symbolDTO struct {
	Icon  string `json:"icon"`
	Words string `json:"words"`
}

func symbolOf(s application.Symbol) symbolDTO { return symbolDTO{Icon: s.Icon, Words: s.Words} }

// dayDTO is one day of a cell (FR-404, FR-406); Date is written yyyy-mm-dd.
type dayDTO struct {
	Date    string    `json:"date"`
	Weekday string    `json:"weekday"`
	High    int       `json:"high"`
	Low     int       `json:"low"`
	Rain    float64   `json:"rain"`
	Symbol  symbolDTO `json:"symbol"`
	Known   bool      `json:"known"`
}

// cellDTO is one cell of the ribbon.
type cellDTO struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	Place       string    `json:"place"`
	Time        string    `json:"time"`
	ZoneMark    string    `json:"zoneMark"`
	Temperature int       `json:"temperature"`
	Symbol      symbolDTO `json:"symbol"`
	Today       dayDTO    `json:"today"`
	Outlook     []dayDTO  `json:"outlook"`
	Age         string    `json:"age"`
	Problem     string    `json:"problem"`
	Petrichor   bool      `json:"petrichor"`
}

// snapshotDTO is everything the ribbon draws.
type snapshotDTO struct {
	Cells         []cellDTO `json:"cells"`
	Units         string    `json:"units"`
	Format        string    `json:"format"`
	Colour        string    `json:"colour"`
	Orientation   string    `json:"orientation"`
	Theme         string    `json:"theme"`
	AlwaysOnTop   bool      `json:"alwaysOnTop"`
	Opacity       int       `json:"opacity"`
	MinOpacity    int       `json:"minOpacity"`
	Scale         float64   `json:"scale"`
	MinScale      int       `json:"minScale"`
	MaxScale      int       `json:"maxScale"`
	Layout        layoutDTO `json:"layout"`
	RefreshInMs   int64     `json:"refreshInMs"`
	Notices       []string  `json:"notices"`
	Scrolls       bool      `json:"scrolls"`
	DragThreshold sizeDTO   `json:"dragThreshold"`
	StartLabel    string    `json:"startLabel"`
	// Collapsed is true while the window is an unpinned ribbon's tab (FR-706).
	Collapsed bool `json:"collapsed"`
	// Choices are the menus' choices, which Settings offers as well (FR-701).
	Choices []choiceDTO `json:"choices"`
}

// choiceDTO is one of the menus' choices as Settings draws it (FR-701): either a group of Children
// or one item whose Action the page hands back to Choose.
type choiceDTO struct {
	Action    string      `json:"action"`
	Label     string      `json:"label"`
	Checkable bool        `json:"checkable"`
	Checked   bool        `json:"checked"`
	Children  []choiceDTO `json:"children"`
}

// placeDTO is one entry of the place search (FR-202): the place's GeoNames id, its name and its
// name in full.
type placeDTO struct {
	GeoNamesID  int    `json:"geoNamesId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// hourDTO is one hour of the detail panel (FR-410).
type hourDTO struct {
	Time        string    `json:"time"`
	Symbol      symbolDTO `json:"symbol"`
	Temperature int       `json:"temperature"`
	Rain        float64   `json:"rain"`
	WindSpeed   int       `json:"windSpeed"`
	WindFrom    float64   `json:"windFrom"`
}

// detailDTO is the detail panel for one city (FR-410, FR-411).
type detailDTO struct {
	ID      string    `json:"id"`
	Label   string    `json:"label"`
	Place   string    `json:"place"`
	Hours   []hourDTO `json:"hours"`
	Sunrise string    `json:"sunrise"`
	Sunset  string    `json:"sunset"`
	Problem string    `json:"problem"`
}

// choicesOf answers the wire form of menu items.
func choicesOf(items []menus.Item) []choiceDTO {
	out := make([]choiceDTO, 0, len(items))
	for _, item := range items {
		out = append(out, choiceDTO{
			Action: string(item.Action), Label: item.Label, Checkable: item.Checkable, Checked: item.Checked,
			Children: choicesOf(item.Children),
		})
	}
	return out
}

func sizeOf(size placement.Size) sizeDTO { return sizeDTO{Width: size.Width, Height: size.Height} }

// dateOf writes date as yyyy-mm-dd.
func dateOf(date forecast.Date) string {
	return fmt.Sprintf("%04d-%02d-%02d", date.Year, date.Month, date.Day)
}

func dayOf(day application.Day) dayDTO {
	return dayDTO{
		Date: dateOf(day.Date), Weekday: day.Weekday, High: day.High, Low: day.Low, Rain: day.Rain,
		Symbol: symbolOf(day.Symbol), Known: day.Known,
	}
}

func cellOf(c application.Cell) cellDTO {
	outlook := make([]dayDTO, 0, len(c.Outlook))
	for _, day := range c.Outlook {
		outlook = append(outlook, dayOf(day))
	}
	return cellDTO{
		ID: c.ID, Label: c.Label, Place: c.Place, Time: c.Time, ZoneMark: c.ZoneMark,
		Temperature: c.Temperature, Symbol: symbolOf(c.Symbol), Today: dayOf(c.Today), Outlook: outlook,
		Age: c.Age, Problem: c.Problem, Petrichor: c.Petrichor,
	}
}

// snapshotOf answers the wire form of a snapshot; how the window shows it is the facade's to add.
func snapshotOf(s application.Snapshot) snapshotDTO {
	cells := make([]cellDTO, 0, len(s.Cells))
	for _, c := range s.Cells {
		cells = append(cells, cellOf(c))
	}
	notices := append([]string{}, s.Notices...)
	choices := s.Choices
	return snapshotDTO{
		Cells: cells, Units: string(s.Units), Format: string(s.Format),
		Colour: string(choices.Colour), Orientation: string(choices.Orientation), Theme: string(choices.Theme),
		AlwaysOnTop: choices.AlwaysOnTop, Opacity: choices.Opacity, MinOpacity: ribbon.MinOpacity,
		Scale: s.Scale, MinScale: ribbon.MinScale, MaxScale: ribbon.MaxScale,
		Layout:      layoutDTO{Cell: sizeOf(s.Layout.Cell), Prompt: sizeOf(s.Layout.Prompt), Padding: s.Layout.Padding},
		RefreshInMs: s.NextRefresh.Sub(s.Now).Milliseconds(),
		Notices:     notices,
		StartLabel:  product.StartAtSignIn,
	}
}

func placesOf(places []place.Place) []placeDTO {
	out := make([]placeDTO, 0, len(places))
	for _, p := range places {
		out = append(out, placeDTO{GeoNamesID: p.GeoNamesID, Name: p.Name, Description: p.Description()})
	}
	return out
}

func detailOf(d application.Detail) detailDTO {
	hours := make([]hourDTO, 0, len(d.Hours))
	for _, h := range d.Hours {
		hours = append(hours, hourDTO{
			Time: h.Time, Symbol: symbolOf(h.Symbol), Temperature: h.Temperature, Rain: h.Rain, WindSpeed: h.WindSpeed,
			WindFrom: h.WindFrom,
		})
	}
	return detailDTO{
		ID: d.ID, Label: d.Label, Place: d.Place, Hours: hours, Sunrise: d.Sunrise, Sunset: d.Sunset,
		Problem: d.Problem,
	}
}
