package application

import (
	"fmt"
	"slices"
	"time"

	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/ribbonkit/domain/placement"
	"github.com/oernster/weatherribbon/internal/domain/settings"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// daysPerWeek counts the weekdays an outlook can name.
const daysPerWeek = 7

// Layout is the size of one cell and the padding around the cells, all in DIP. It has one home, the
// composition root; the page draws cells at the sizes the snapshot hands it.
type Layout struct {
	Cell placement.Size
	// Prompt is the one cell an empty ribbon shows, holding the Add city control (FR-106).
	Prompt  placement.Size
	Padding int
}

// Samples is every text a cell can show whose width varies, for the page to measure in the font it
// really draws with (FR-103): every minute of a day in the time format, every weekday, every whole
// temperature in the units and every label held.
type Samples struct {
	Times        []string
	Weekdays     []string
	Temperatures []int
	Labels       []string
}

// Measured is the width in DIP a cell needs to show its widest text whole, as the page measured it,
// together with what it was measured under: the units and the time format set the texts, the labels
// are texts themselves (FR-103).
type Measured struct {
	Units     units.System
	Format    localtime.Format
	Labels    []string
	CellWidth int
}

// TextSamples answers the texts a cell can show under the current settings, for the page to
// measure (FR-103). Only the page can measure them: the font is whatever its web engine draws.
func (s *Service) TextSamples() Samples {
	current := s.Settings()
	weekdays := make([]string, 0, daysPerWeek)
	for day := range daysPerWeek {
		weekdays = append(weekdays, time.Weekday(day).String())
	}
	return Samples{
		Times:        localtime.TimeSamples(current.Format),
		Weekdays:     weekdays,
		Temperatures: units.Temperatures(current.Units),
		Labels:       labelsOf(current),
	}
}

// SetMeasured records the cell width the page measured the widest text to need; a width below zero
// is refused. Arranging the window afterwards gives the cells that width (FR-103).
func (s *Service) SetMeasured(measured Measured) error {
	if measured.CellWidth < 0 {
		return fmt.Errorf("%w: a cell of %d", placement.ErrNegativeLength, measured.CellWidth)
	}
	measured.Labels = slices.Clone(measured.Labels)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.measured = measured
	return nil
}

// layoutFor answers the layout cells are drawn at under current: the composition root's, with the
// cell widened to the width the page measured where that is wider and was measured under these very
// settings; a measurement for others waits for the page to measure again. The caller holds the
// mutex. Snapshot and arranging both read it, so the page's cells and the window's size cannot
// disagree (FR-103).
func (s *Service) layoutFor(current settings.Settings) Layout {
	layout, m := s.layout, s.measured
	if m.Units != current.Units || m.Format != current.Format || !slices.Equal(m.Labels, labelsOf(current)) {
		return layout
	}
	layout.Cell.Width = max(layout.Cell.Width, m.CellWidth)
	return layout
}

// labelsOf answers the label of every city held, in the order they were added.
func labelsOf(current settings.Settings) []string {
	labels := make([]string, 0, len(current.Cities))
	for _, city := range current.Cities {
		labels = append(labels, city.Label)
	}
	return labels
}
