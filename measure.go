package main

import (
	"github.com/oernster/ribbonkit/domain/localtime"
	"github.com/oernster/weatherribbon/internal/application"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// TextSamples answers every text a cell can show whose width varies, for the page to measure (FR-103).
func (a *App) TextSamples() textSamplesDTO {
	samples := a.service.TextSamples()
	return textSamplesDTO{
		Times: append([]string{}, samples.Times...), Weekdays: append([]string{}, samples.Weekdays...),
		Temperatures: append([]int{}, samples.Temperatures...), Labels: append([]string{}, samples.Labels...),
	}
}

// SetMeasured takes the cell width the page measured its widest text to need in the font it really
// draws with, then fits the ribbon to it (FR-103). Once applied, the launched ribbon may be shown
// (FR-102).
func (a *App) SetMeasured(measured measuredDTO) error {
	err := a.control.Refitted(a.service.SetMeasured(application.Measured{
		Units:     units.System(measured.Units),
		Format:    localtime.Format(measured.Format),
		Labels:    measured.Labels,
		CellWidth: measured.CellWidth,
	}))
	if err == nil {
		a.control.PageMeasured()
	}
	return err
}
