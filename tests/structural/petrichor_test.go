package structural

// FR-413: the petrichor line pulses gently, its opacity moving between 100 and 50 percent and back
// over four seconds; it stands still where the system asks for reduced motion. jsdom draws no
// animation and Vitest hands a test no stylesheet's text, so the rule is held here, in app.css as
// written.

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/oernster/ribbonkit/structure"
)

// petrichorRules are what app.css must say for the line to pulse and to stand still.
var petrichorRules = map[string]*regexp.Regexp{
	"a four-second pulse":        regexp.MustCompile(`\.petrichor \{[^}]*animation: petrichor 4s`),
	"half opacity at its middle": regexp.MustCompile(`@keyframes petrichor \{\s*50% \{\s*opacity: 0\.5;`),
	"still under reduced motion": regexp.MustCompile(`prefers-reduced-motion: reduce\)\s*\{\s*\.petrichor \{\s*animation: none;`),
}

func TestThePetrichorLinePulsesAndStandsStillWhenAsked(t *testing.T) {
	sheet := structure.Read(t, filepath.Join(structure.Root(t), "frontend", "src", "app.css"))
	for what, rule := range petrichorRules {
		if !rule.MatchString(sheet) {
			t.Errorf("app.css does not give the petrichor line %s", what)
		}
	}
}
