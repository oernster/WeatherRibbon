package structural

// ribbonkit is one release in two halves: the Go module in go.mod and the npm package in the front
// end's package.json. The window and the page half it serves change together, so both must name the
// same tag; a page built from one release against a window from another breaks only at run time.

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// kitPackage is the npm half's name; kitSpecPrefix is how package.json names its tag on GitHub.
const (
	kitPackage    = "@oernster/ribbonkit"
	kitSpecPrefix = "github:oernster/ribbonkit#"
)

// goModRequire finds the version go.mod requires the kit at.
var goModRequire = regexp.MustCompile(`(?m)^\s*(?:require\s+)?` + regexp.QuoteMeta(kitModule) + `\s+(v\S+)`)

func TestBothHalvesOfTheKitNameOneTag(t *testing.T) {
	root := structure.Root(t)
	required := goModRequire.FindStringSubmatch(structure.Read(t, filepath.Join(root, "go.mod")))
	if required == nil {
		t.Fatalf("go.mod requires no version of %s", kitModule)
	}
	var stated struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(structure.Read(t, filepath.Join(root, "frontend", "package.json"))), &stated); err != nil {
		t.Fatal(err)
	}
	spec := stated.Dependencies[kitPackage]
	tag, ok := strings.CutPrefix(spec, kitSpecPrefix)
	if !ok {
		t.Fatalf("package.json names %s as %q, not a tag of %s", kitPackage, spec, kitSpecPrefix)
	}
	if tag != required[1] {
		t.Errorf("go.mod requires the kit at %s but package.json names %s", required[1], tag)
	}
}
