package application

import "strings"

// Symbol is a weather symbol as a cell shows it (FR-412): the icon drawn for it, named by the code's
// words (NFR-U-2); where there is no icon, the words in its place.
type Symbol struct {
	// Icon names the icon's file, without its extension; empty when there is none.
	Icon string
	// Words are the code written as words, such as "lightrainshowers day", in the spelling that
	// reads correctly whichever one MET Norway sent.
	Words string
}

// variantMark joins a symbol to its variant in a code, as in "clearsky_day"; a code's words put a
// space where it stood.
const variantMark = "_"

// iconSpelling maps a symbol's correct spelling to the one the icon set files it under (FR-412): the
// legend spells light sleet showers and thunder with a doubled s. MET Norway may send either; the
// icon is found under the second and the words are written in the first.
var iconSpelling = map[string]string{"lightsleetshowersandthunder": "lightssleetshowersandthunder"}

// symbolOf answers how code is shown: its icon where the set holds one, always with its words; the
// words alone where there is none, the code then told to the icons port the first time it is met in
// this run. An empty code is no symbol.
func (s *Service) symbolOf(code string) Symbol {
	if code == "" {
		return Symbol{}
	}
	name, variant, hasVariant := strings.Cut(code, variantMark)
	said, icon := name, name
	for correct, filed := range iconSpelling {
		if name == correct || name == filed {
			said, icon = correct, filed
		}
	}
	if hasVariant {
		said += " " + variant
		icon += variantMark + variant
	}
	if s.ports.Icons.Has(icon) {
		return Symbol{Icon: icon, Words: said}
	}
	if _, seen := s.missing.LoadOrStore(code, true); !seen {
		s.ports.Icons.Missing(code)
	}
	return Symbol{Words: said}
}
