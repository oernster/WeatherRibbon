// Package structural enforces the architecture with tests rather than convention (CON-1, CON-2).
// The mechanics are ribbonkit's structure package, which holds the kit and TimeRibbon by the same
// code; the rules' data here is WeatherRibbon's own.
//
// Every assertion here has been proved to bite by planting a violation and watching it fail. An
// assertion never seen to fail is not yet a guard.
package structural

import (
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// module is WeatherRibbon's module path; kitModule is ribbonkit's, laid out in the same layers.
const (
	module    = "github.com/oernster/weatherribbon"
	kitModule = "github.com/oernster/ribbonkit"
)

// skippedDirectories are walked past besides git's and npm's: the page's source is walked on its
// own; build holds generated output.
var skippedDirectories = []string{"frontend", "build"}

// layout answers WeatherRibbon's layout: its layers under internal with the kit's laid out alike;
// main.go is the one file that wires the application to the infrastructure.
func layout(t *testing.T) structure.Layout {
	t.Helper()
	return structure.Layout{
		Root:            structure.Root(t),
		Module:          module,
		Layered:         []string{"internal"},
		Dependencies:    []string{kitModule},
		CompositionRoot: []string{"main.go"},
	}
}

// goFiles answers every Go file of WeatherRibbon.
func goFiles(t *testing.T) []string {
	t.Helper()
	return structure.GoFiles(t, structure.Root(t), skippedDirectories...)
}
