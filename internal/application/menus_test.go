package application

import (
	"slices"
	"testing"

	"github.com/oernster/ribbonkit/application/menus"
	"github.com/oernster/weatherribbon/internal/domain/units"
)

func itemLabels(items []menus.Item) []string {
	var out []string
	for _, item := range items {
		out = append(out, item.Label)
	}
	return out
}

// find answers the item of menu labelled label, failing the test when there is none.
func find(t *testing.T, menu []menus.Item, label string) menus.Item {
	t.Helper()
	index := slices.IndexFunc(menu, func(item menus.Item) bool { return item.Label == label })
	if index < 0 {
		t.Fatalf("no %s in %v", label, itemLabels(menu))
	}
	return menu[index]
}

// FR-109.
func TestContextMenuOffersTheRibbonsActions(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t)
	menu := r.service.ContextMenu()
	if got := itemLabels(menu); !slices.Equal(got, []string{
		"Add city", "Settings", "Units", "Colour", "Orientation", "Position", "Always on top", "Pin ribbon",
		"Refresh now", "Help", "Hide ribbon", "Exit",
	}) {
		t.Errorf("got %v", got)
	}
	if menu[0].Action != ActionAddCity || menu[8].Action != ActionRefreshNow || menu[len(menu)-1].Action != menus.Exit {
		t.Errorf("acts as %q, %q and %q", menu[0].Action, menu[8].Action, menu[len(menu)-1].Action)
	}
}

// FR-601: the tray holds the context menu's items with Show ribbon or Hide ribbon first in place of
// its own Hide ribbon.
func TestTrayMenuNamesTheOppositeOfTheVisibility(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t)
	shown := r.service.TrayMenu(true)
	if !slices.Equal(itemLabels(shown), []string{
		"Hide ribbon", "Add city", "Settings", "Units", "Colour", "Orientation", "Position", "Always on top",
		"Pin ribbon", "Refresh now", "Help", "Exit",
	}) || shown[0].Action != menus.Hide {
		t.Errorf("visible: %v", itemLabels(shown))
	}
	if hidden := r.service.TrayMenu(false); hidden[0].Label != "Show ribbon" || hidden[0].Action != menus.Show {
		t.Errorf("hidden: %+v", hidden[0])
	}
}

// FR-703: Units is a submenu in both menus offering every system with the current one ticked; each
// item names its system and nothing else does.
func TestBothMenusOfferUnitsWithTheCurrentTicked(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t)
	if err := r.service.SetUnits(units.Imperial); err != nil {
		t.Fatal(err)
	}
	for name, menu := range map[string][]menus.Item{"tray": r.service.TrayMenu(true), "context": r.service.ContextMenu()} {
		item := find(t, menu, labelUnits)
		if item.Action != "" || !slices.Equal(itemLabels(item.Children), []string{"Metric", "Imperial"}) {
			t.Fatalf("%s: %+v", name, item)
		}
		for index, child := range item.Children {
			got, ok := UnitsOf(child.Action)
			if !ok || got != units.Systems[index] || !child.Checkable || child.Checked != (got == units.Imperial) {
				t.Errorf("%s: %+v answered %s, %v", name, child, got, ok)
			}
		}
	}
	if _, ok := UnitsOf(ActionRefreshNow); ok {
		t.Error("Refresh now was taken for a system")
	}
}

// FR-701: Settings offers every choice the menus offer and no command.
func TestSettingsOffersEveryMenuChoice(t *testing.T) {
	t.Parallel()
	r := horizontalRig(t)
	if got := itemLabels(r.service.SettingsChoices()); !slices.Equal(got, []string{
		"Units", "Colour", "Orientation", "Position", "Always on top", "Pin ribbon",
	}) {
		t.Errorf("got %v", got)
	}
}

// FR-602.
func TestCloseRequestHidesRatherThanQuits(t *testing.T) {
	t.Parallel()
	if got := horizontalRig(t).service.CloseRequested(); got != menus.Hide {
		t.Errorf("got %s", got)
	}
}
