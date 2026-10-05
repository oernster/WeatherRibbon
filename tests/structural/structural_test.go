// Package structural enforces the architecture with tests rather than convention (CON-1, CON-2).
// The mechanics are ribbonkit's structure package, which holds the kit and TimeRibbon by the same
// code; the rules' data here is WeatherRibbon's own.
//
// Every assertion here has been proved to bite by planting a violation and watching it fail. An
// assertion never seen to fail is not yet a guard.
package structural

import (
	"os/exec"
	"path/filepath"
	"strings"
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

// pageExtensions are the page's source files the size rule governs.
var pageExtensions = []string{".ts", ".tsx", ".css"}

// pageFiles answers the page's own source files under frontend/src.
func pageFiles(t *testing.T) []string {
	t.Helper()
	return structure.FilesWith(t, pageExtensions, filepath.Join(structure.Root(t), "frontend", "src"))
}

// kitDir answers the folder Go builds ribbonkit from: the tagged module go.mod requires (else the
// working copy a local go.work names instead).
func kitDir(t *testing.T) string {
	t.Helper()
	command := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", kitModule)
	command.Dir = structure.Root(t)
	out, err := command.Output()
	dir := strings.TrimSpace(string(out))
	if err != nil || dir == "" {
		t.Fatalf("go list could not find %s (%v); run go mod download", kitModule, err)
	}
	return dir
}

// goFiles answers every Go file of WeatherRibbon.
func goFiles(t *testing.T) []string {
	t.Helper()
	return structure.GoFiles(t, structure.Root(t), skippedDirectories...)
}
