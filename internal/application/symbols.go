package application

import "strings"

// Symbol is a weather symbol as a cell shows it (FR-412): the icon drawn for it, named by the code's
// words (NFR-U-2); where there is no icon, the words in its place.
type Symbol struct {
	// Icon names the icon's file, without its extension; empty when there is none.
	Icon string
	// Words name the symbol as an everyday forecast does, such as "Light showers", whichever
	// spelling MET Norway sent (symbolWords); a code the legend lacks is written as itself, its first
	// letter capital.
	Words string
}

// variantMark joins a symbol to its variant in a code, as in "clearsky_day"; a code the legend lacks
// is written with a space where it stood.
const variantMark = "_"

// iconSpelling maps a symbol's correct spelling to the one the icon set files it under (FR-412): the
// legend spells light sleet showers and thunder and light snow showers and thunder with a doubled s.
// MET Norway may send either; the icon is found under the second and the words name the first.
var iconSpelling = map[string]string{
	"lightsleetshowersandthunder": "lightssleetshowersandthunder",
	"lightsnowshowersandthunder":  "lightssnowshowersandthunder",
}

// symbolOf answers how code is shown: its icon where the set holds one, always with its words; the
// words alone where there is none, the code then told to the icons port the first time it is met in
// this run. An empty code is no symbol.
func (s *Service) symbolOf(code string) Symbol {
	if code == "" {
		return Symbol{}
	}
	name, variant, hasVariant := strings.Cut(code, variantMark)
	correct, icon := name, name
	for spelt, filed := range iconSpelling {
		if name == spelt || name == filed {
			correct, icon = spelt, filed
		}
	}
	said, known := symbolWords[correct]
	if !known {
		said = strings.ReplaceAll(code, variantMark, " ")
		said = strings.ToUpper(said[:1]) + said[1:]
	}
	if hasVariant {
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
