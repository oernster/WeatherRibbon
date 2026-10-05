package structural

// FR-610: About credits every component the application ships. The Go modules are the part a
// machine can check, so this asks the Go tool which modules each platform's build links, then
// holds that platform's credits to that list in both directions. A module
// linked but not credited fails; so does one credited that nothing links any more.

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/structure"
	"github.com/oernster/weatherribbon/internal/product"
)

// shippedBuild is how one platform's build compiles what ships: its tags, which change what the
// Wails module links; whether cgo is on, which decides which files count; its main packages.
type shippedBuild struct {
	tags     string
	cgo      string
	packages []string
}

// shippedBuilds are the builds per platform. The setup program joins the Windows list when it is
// ported.
var shippedBuilds = map[string]shippedBuild{
	product.Windows: {"desktop,production", "0", []string{"."}},
	product.Linux:   {"desktop,production,webkit2_41", "1", []string{"."}},
	product.MacOS:   {"desktop,production", "1", []string{"."}},
}

// ownModules are the author's own modules, which About does not credit as third-party components:
// WeatherRibbon and the ribbonkit it is built on.
var ownModules = []string{module, kitModule}

// linkedModules answers every module goos's shipped executables link, the author's own left out.
// Listing needs no C compiler, so every platform is listed from any machine.
func linkedModules(t *testing.T, goos string) []string {
	t.Helper()
	build := shippedBuilds[goos]
	args := append([]string{"list", "-tags", build.tags, "-deps", "-f", "{{with .Module}}{{.Path}}{{end}}"}, build.packages...)
	command := exec.Command("go", args...)
	command.Dir = structure.Root(t)
	command.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED="+build.cgo)
	out, err := command.Output()
	if err != nil {
		t.Fatalf("go list for %s: %v", goos, err)
	}
	var modules []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if linked := strings.TrimSpace(line); linked != "" && !slices.Contains(ownModules, linked) && !slices.Contains(modules, linked) {
			modules = append(modules, linked)
		}
	}
	if len(modules) == 0 {
		t.Fatal("go list named no module at all, the query is wrong")
	}
	return modules
}

func TestEveryLinkedModuleIsCredited(t *testing.T) {
	for _, goos := range product.Platforms {
		var credited []string
		for _, credit := range product.CreditsFor(goos) {
			if credit.Module != "" && !slices.Contains(credited, credit.Module) {
				credited = append(credited, credit.Module)
			}
		}
		linked := linkedModules(t, goos)
		for _, module := range linked {
			if !slices.Contains(credited, module) {
				t.Errorf("%s: %s is linked into what ships but About does not credit it", goos, module)
			}
		}
		for _, module := range credited {
			if !slices.Contains(linked, module) {
				t.Errorf("%s: About credits %s, which nothing that ships links any more", goos, module)
			}
		}
	}
}

// Each platform names one role per module, so About never lists the same module twice.
func TestAModuleIsCreditedOncePerPlatform(t *testing.T) {
	for _, goos := range product.Platforms {
		seen := map[string]bool{}
		for _, credit := range product.CreditsFor(goos) {
			if credit.Module != "" && seen[credit.Module] {
				t.Errorf("%s: %s is credited twice", goos, credit.Module)
			}
			seen[credit.Module] = true
		}
	}
}
