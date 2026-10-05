package application

import (
	"slices"
	"testing"
)

// FR-412, NFR-U-2: a code with no icon shows its words in its place and is told missing once, however
// often it is met; a code with an icon shows the icon, named by its words; no code is no symbol.
func TestAnUnknownSymbolShowsItsWords(t *testing.T) {
	t.Parallel()
	r := newRig()
	for range 2 {
		if got := r.service.symbolOf("lightrainshowers_day"); got != (Symbol{Words: "lightrainshowers day"}) {
			t.Errorf("an unknown code showed %+v", got)
		}
	}
	if !slices.Equal(r.icons.missing, []string{"lightrainshowers_day"}) {
		t.Errorf("told missing %v, want the code once", r.icons.missing)
	}
	if got := r.service.symbolOf("rain"); got != (Symbol{Icon: "rain", Words: "rain"}) {
		t.Errorf("a known code showed %+v", got)
	}
	if got := r.service.symbolOf(""); got != (Symbol{}) {
		t.Errorf("no code showed %+v", got)
	}
}

// FR-412, NFR-U-2: light sleet showers and thunder finds the icon the set files under the doubled s
// whichever spelling MET Norway sends; it is named in the single s.
func TestBothSpellingsOfLightSleetThunderHaveAnIcon(t *testing.T) {
	t.Parallel()
	r := newRig()
	want := Symbol{Icon: "lightssleetshowersandthunder_day", Words: "lightsleetshowersandthunder day"}
	for _, code := range []string{"lightsleetshowersandthunder_day", "lightssleetshowersandthunder_day"} {
		if got := r.service.symbolOf(code); got != want {
			t.Errorf("%s showed %+v", code, got)
		}
	}
	if len(r.icons.missing) != 0 {
		t.Errorf("told missing %v", r.icons.missing)
	}
}
