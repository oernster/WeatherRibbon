package structural

// Wails reads each executable's name from its wails.json and build.ps1 reads it back from there, so
// the two files must spell the name themselves. This test holds them to internal/product, so a
// rename that misses either file fails here rather than shipping an executable under the old name.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/oernster/ribbonkit/structure"

	"github.com/oernster/weatherribbon/internal/product"
)

// wailsNames names each wails.json with the name internal/product gives its executable.
var wailsNames = map[string]string{
	"wails.json": product.Name,
}

// wailsIdentity is the part of a wails.json that names the executable.
type wailsIdentity struct {
	Name           string `json:"name"`
	OutputFilename string `json:"outputfilename"`
}

func TestEachWailsConfigNamesItsExecutableAsTheProductDoes(t *testing.T) {
	for file, want := range wailsNames {
		raw, err := os.ReadFile(filepath.Join(structure.Root(t), file))
		if err != nil {
			t.Fatalf("reading %s: %v", filepath.ToSlash(file), err)
		}
		var got wailsIdentity
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatalf("reading %s: %v", filepath.ToSlash(file), err)
		}
		if got.Name != want || got.OutputFilename != want {
			t.Errorf("%s names %q with output %q; internal/product says %q", filepath.ToSlash(file), got.Name, got.OutputFilename, want)
		}
	}
}
