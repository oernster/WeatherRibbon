package structural

// Every text in a cell meets 4.5:1 against the cell in both themes, on every scheme (NFR-U-1).
// WeatherRibbon states no palette of its own: the page loads ribbonkit's, whose own suite holds the
// ribbon's text tokens to the floor on its backgrounds (TestTheKitsTextMeetsTheContrastFloor). That
// holds for WeatherRibbon only while its stylesheets write no colour of their own, draw text only in
// those tokens and lay it only on those backgrounds, which is what this checks.

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

var (
	// colourLiteral is a colour written out rather than taken from a token.
	colourLiteral = regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b|\b(?:rgba?|hsla?|hwb|lab|lch|oklab|oklch)\(`)
	// textColour and backgroundColour find what text is drawn in and what it is laid on.
	textColour       = regexp.MustCompile(`(?m)^\s*color:\s*([^;]+);`)
	backgroundColour = regexp.MustCompile(`(?m)^\s*background(?:-color)?:\s*([^;]+);`)
	// tokenValue reads the one token a value names, as var(--name).
	tokenValue = regexp.MustCompile(`^var\(--([\w-]+)\)$`)
)

// fadedSurface is app.css's --surface drawn at the ribbon's opacity, which is --surface itself at the
// 100 percent NFR-U-1 is stated at.
const fadedSurface = "surface-shown"

// seeThrough are backgrounds that paint nothing, so what is drawn there lies on what is beneath: the
// page behind #root, the petrichor button on its cell.
var seeThrough = []string{"transparent", "none"}

// cssFiles answers WeatherRibbon's own stylesheets.
func cssFiles(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, path := range pageFiles(t) {
		if strings.HasSuffix(path, ".css") {
			out = append(out, path)
		}
	}
	return out
}

func TestWeatherRibbonDrawsTextOnlyInTheKitsCheckedColours(t *testing.T) {
	texts, backgrounds := structure.TextTokens(), append(structure.Backgrounds(), fadedSurface)
	files := cssFiles(t)
	if len(files) == 0 {
		t.Fatal("no stylesheet found; the walk is wrong")
	}
	for _, path := range files {
		text := structure.Read(t, path)
		name := structure.Relative(structure.Root(t), path)
		for _, literal := range colourLiteral.FindAllString(text, -1) {
			t.Errorf("%s writes the colour %s; take it from the kit's palette", name, literal)
		}
		check := func(form *regexp.Regexp, allowed []string, what string) {
			for _, match := range form.FindAllStringSubmatch(text, -1) {
				value := strings.TrimSpace(match[1])
				if form == backgroundColour && slices.Contains(seeThrough, value) {
					continue
				}
				named := tokenValue.FindStringSubmatch(value)
				if named == nil || !slices.Contains(allowed, named[1]) {
					t.Errorf("%s draws %s in %s; the kit checks only %v", name, what, value, allowed)
				}
			}
		}
		check(textColour, texts, "text")
		check(backgroundColour, backgrounds, "a background")
	}
}
