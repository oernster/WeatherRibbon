package structural

import (
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

func TestDomainHasNoOutwardImports(t *testing.T) {
	layout(t).CheckDomainHasNoOutwardImports(t, goFiles(t))
}

// No IO, randomness or tz package, no wall clock read and no zone loaded (CON-5, FR-413).
func TestDomainIsPure(t *testing.T) {
	layout(t).CheckDomainIsPure(t, goFiles(t))
}

func TestApplicationDoesNotImportInfrastructure(t *testing.T) {
	layout(t).CheckApplicationDoesNotImportInfrastructure(t, goFiles(t))
}

func TestNothingBelowTheUIImportsIt(t *testing.T) {
	layout(t).CheckNothingBelowTheUIImportsIt(t, goFiles(t))
}

func TestWailsStaysOutOfInfrastructure(t *testing.T) {
	layout(t).CheckWailsStaysOutOfInfrastructure(t, goFiles(t))
}

// Only main.go imports both application and infrastructure, the kit's included.
func TestCompositionRootIsWhitelisted(t *testing.T) {
	layout(t).CheckCompositionRootIsWhitelisted(t, goFiles(t))
}

// sourceFiles is every file the size rule governs. Until the page exists that is the Go alone;
// frontend/src joins it with the front end (CON-2).
func sourceFiles(t *testing.T) []string {
	t.Helper()
	return goFiles(t)
}

func TestNoFileExceedsLineLimit(t *testing.T) {
	structure.CheckNoFileExceedsLineLimit(t, sourceFiles(t))
}

func TestNoFileInDangerBand(t *testing.T) {
	structure.CheckNoFileInDangerBand(t, sourceFiles(t))
}

func TestEveryExportedTypeIsDocumented(t *testing.T) {
	structure.CheckEveryExportedTypeIsDocumented(t, goFiles(t))
}
