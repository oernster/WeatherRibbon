package application

import (
	"github.com/oernster/ribbonkit/application/arranger"
	"github.com/oernster/ribbonkit/domain/ribbon"
	"github.com/oernster/weatherribbon/internal/domain/settings"
)

// host is the service as the arranger's host: the cities are the ribbon's content and the ribbon's
// choices are saved with the rest of the settings, so a save that fails raises the same notice
// (FR-702). Not the service itself, so its two methods stay off the service's surface.
type host struct{ s *Service }

// Ribbon answers the ribbon's choices with its content, read together under one lock: a cell for
// each notice, then for each city, the prompt standing in for them when there are none (FR-106);
// each cell the layout's, widened to the measured text where that applies (FR-103).
func (h host) Ribbon() (ribbon.Choices, arranger.Content) {
	s := h.s
	s.mutex.Lock()
	defer s.mutex.Unlock()
	current := s.current.Normalised()
	layout := s.layoutFor(current)
	cell := layout.Cell
	if len(current.Cities) == 0 {
		cell = layout.Prompt
	}
	return current.Choices, arranger.Content{
		Cell:    cell,
		Cells:   len(s.notices()) + max(len(current.Cities), 1),
		Padding: layout.Padding,
	}
}

// ChangeRibbon applies edit to the ribbon's choices through the one save path, change.
func (h host) ChangeRibbon(edit func(ribbon.Choices) ribbon.Choices) error {
	return h.s.change(func(current settings.Settings) (settings.Settings, error) {
		current.Choices = edit(current.Choices)
		return current, nil
	})
}
