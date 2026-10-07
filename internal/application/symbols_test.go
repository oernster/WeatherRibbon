package application

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// iconFolder is the Yr icon set the page ships, read here only for its file names.
const iconFolder = "../../frontend/src/assets/weather"

// FR-412, NFR-U-2: a code with no icon shows its words in its place and is told missing once, however
// often it is met; a code with an icon shows the icon, named by its words; no code is no symbol.
func TestAnUnknownSymbolShowsItsWords(t *testing.T) {
	t.Parallel()
	r := newRig()
	for range 2 {
		if got := r.service.symbolOf("lightrainshowers_day"); got != (Symbol{Words: "Light rain showers"}) {
			t.Errorf("a code with no icon showed %+v", got)
		}
	}
	if !slices.Equal(r.icons.missing, []string{"lightrainshowers_day"}) {
		t.Errorf("told missing %v, want the code once", r.icons.missing)
	}
	if got := r.service.symbolOf("rain"); got != (Symbol{Icon: "rain", Words: "Rain"}) {
		t.Errorf("a known code showed %+v", got)
	}
	if got := r.service.symbolOf(""); got != (Symbol{}) {
		t.Errorf("no code showed %+v", got)
	}
}

// FR-412: a code the legend lacks is written as itself, its variant mark a space, its first letter
// capital.
func TestASymbolOutsideTheLegendIsWrittenAsItself(t *testing.T) {
	t.Parallel()
	r := newRig()
	if got := r.service.symbolOf("sandstorm_day"); got != (Symbol{Words: "Sandstorm day"}) {
		t.Errorf("a code outside the legend showed %+v", got)
	}
}

// FR-412, NFR-U-2: light sleet showers and thunder and light snow showers and thunder each find the
// icon the set files under the doubled s whichever spelling MET Norway sends; each is named in the
// single s.
func TestBothSpellingsOfTheDoubledSSymbolsHaveAnIcon(t *testing.T) {
	t.Parallel()
	r := newRig()
	cases := []struct {
		codes []string
		want  Symbol
	}{
		{
			[]string{"lightsleetshowersandthunder_day", "lightssleetshowersandthunder_day"},
			Symbol{Icon: "lightssleetshowersandthunder_day", Words: "Light sleet showers and thunder"},
		},
		{
			[]string{"lightsnowshowersandthunder_night", "lightssnowshowersandthunder_night"},
			Symbol{Icon: "lightssnowshowersandthunder_night", Words: "Light snow showers and thunder"},
		},
	}
	for _, c := range cases {
		for _, code := range c.codes {
			if got := r.service.symbolOf(code); got != c.want {
				t.Errorf("%s showed %+v", code, got)
			}
		}
	}
	if len(r.icons.missing) != 0 {
		t.Errorf("told missing %v", r.icons.missing)
	}
}

// NFR-U-2: every icon the page ships is named in English from symbolWords, never written as its code.
func TestEveryShippedIconIsNamedInEnglish(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(iconFolder)
	if err != nil {
		t.Fatal(err)
	}
	named := make([]string, 0, len(symbolWords))
	for _, words := range symbolWords {
		named = append(named, words)
	}
	r := newRig()
	for _, entry := range entries {
		code, isIcon := strings.CutSuffix(entry.Name(), ".svg")
		if !isIcon {
			continue
		}
		if got := r.service.symbolOf(code); !slices.Contains(named, got.Words) {
			t.Errorf("%s is named %q, which symbolWords does not hold", code, got.Words)
		}
	}
}
