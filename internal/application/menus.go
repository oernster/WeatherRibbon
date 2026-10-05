package application

import (
	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

// WeatherRibbon's own actions; every ribbon's are ribbonkit's menus package's, with their items and
// words. The words shown for WeatherRibbon's own live here alone.
const (
	ActionAddCity    menus.Action = "add-city"
	ActionRefreshNow menus.Action = "refresh-now"
	ActionMetric     menus.Action = "units-metric"
	ActionImperial   menus.Action = "units-imperial"
)

// Item words of WeatherRibbon's own, one home each.
const (
	labelAddCity    = "Add city"
	labelRefreshNow = "Refresh now"
	labelUnits      = "Units"
	labelMetric     = "Metric"
	labelImperial   = "Imperial"
)

// unitItem is a system's Units item: its action and its words.
type unitItem struct {
	action menus.Action
	label  string
}

// unitItems is each system's item; the order they are offered in is units.Systems'.
var unitItems = map[units.System]unitItem{
	units.Metric:   {ActionMetric, labelMetric},
	units.Imperial: {ActionImperial, labelImperial},
}

// UnitsOf answers the system a Units item chooses; false for any other action.
func UnitsOf(action menus.Action) (units.System, bool) {
	for system, each := range unitItems {
		if each.action == action {
			return system, true
		}
	}
	return "", false
}

// TrayMenu answers the tray menu for a ribbon that is or is not visible: the context menu's items
// with Show ribbon or Hide ribbon first in place of its own Hide ribbon (FR-601).
func (s *Service) TrayMenu(visible bool) []menus.Item {
	return append(append([]menus.Item{menus.Visibility(visible)}, s.ribbonItems()...), menus.ExitItem())
}

// ContextMenu answers the menu the ribbon offers when right-clicked (FR-109).
func (s *Service) ContextMenu() []menus.Item {
	return append(s.ribbonItems(), menus.HideItem(), menus.ExitItem())
}

// SettingsChoices answers every choice the menus offer, for Settings to offer as well (FR-701): the
// same items both menus hold, so their words and ticks have one home. What is left of the menus is
// commands, which choose nothing.
func (s *Service) SettingsChoices() []menus.Item {
	choices := s.Settings().Choices
	return []menus.Item{
		s.unitsItem(), menus.ColourItem(choices.Colour), menus.OrientationItem(choices.Orientation),
		menus.PositionItem(choices.Orientation), menus.AlwaysOnTopItem(choices.AlwaysOnTop),
		menus.PinItem(choices.Pinned),
	}
}

// CloseRequested answers what a request to close the ribbon does, such as Alt+F4: it hides the
// ribbon and the application keeps running; only Exit ends it (FR-602).
func (s *Service) CloseRequested() menus.Action {
	return menus.Hide
}

// ribbonItems is what both menus hold, in FR-109's order, up to Help.
func (s *Service) ribbonItems() []menus.Item {
	choices := s.Settings().Choices
	return []menus.Item{
		{Action: ActionAddCity, Label: labelAddCity}, menus.SettingsItem(), s.unitsItem(),
		menus.ColourItem(choices.Colour), menus.OrientationItem(choices.Orientation),
		menus.PositionItem(choices.Orientation), menus.AlwaysOnTopItem(choices.AlwaysOnTop),
		menus.PinItem(choices.Pinned), {Action: ActionRefreshNow, Label: labelRefreshNow}, menus.HelpItem(),
	}
}

// unitsItem is the Units submenu both menus hold, the current system ticked (FR-703).
func (s *Service) unitsItem() menus.Item {
	current := s.Settings().Units
	children := make([]menus.Item, 0, len(units.Systems))
	for _, system := range units.Systems {
		each := unitItems[system]
		children = append(children, menus.Item{
			Action: each.action, Label: each.label, Checkable: true, Checked: current == system,
		})
	}
	return menus.Item{Label: labelUnits, Children: children}
}
