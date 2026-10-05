package structural

// The Licence panel shows the LICENSE as written, unwrapped, with its type sized so the widest line
// fits (ribbonkit's help.css). That holds only while the width help.css states is the widest line of
// the LICENSE WeatherRibbon hands it.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/oernster/ribbonkit/structure"
)

// licenceColumns reads the width help.css sizes the licence's type for.
var licenceColumns = regexp.MustCompile(`--licence-columns:\s*(\d+);`)

func TestTheLicencePanelIsSizedForTheLicencesWidestLine(t *testing.T) {
	root := structure.Root(t)
	sheet, err := os.ReadFile(filepath.Join(kitDir(t), "web", "help.css"))
	if err != nil {
		t.Fatal(err)
	}
	match := licenceColumns.FindSubmatch(sheet)
	if match == nil {
		t.Fatal("help.css states no --licence-columns")
	}
	stated, _ := strconv.Atoi(string(match[1]))
	text, err := os.ReadFile(filepath.Join(root, "LICENSE"))
	if err != nil {
		t.Fatal(err)
	}
	widest := 0
	for _, line := range strings.Split(string(text), "\n") {
		widest = max(widest, utf8.RuneCountInString(strings.TrimRight(line, "\r")))
	}
	if stated != widest {
		t.Errorf("help.css sizes the licence for %d columns but its widest line is %d", stated, widest)
	}
}
