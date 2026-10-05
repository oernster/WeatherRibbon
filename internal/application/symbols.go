package application

import "strings"

// Symbol is a weather symbol as a cell shows it (FR-412): the icon drawn for it; where there is none,
// the code's words in its place.
type Symbol struct {
	// Icon names the icon's file, without its extension; empty when there is none.
	Icon string
	// Words are the code written as words, such as "lightrainshowers day"; empty when Icon is not.
	Words string
}

// variantMark joins a symbol to its variant in a code, as in "clearsky_day"; a code's words put a
// space where it stood.
const variantMark = "_"

// iconSpelling maps a symbol MET Norway may spell one way to the spelling the icon set files it
// under (FR-412): the legend spells light sleet showers and thunder with a doubled s.
var iconSpelling = map[string]string{"lightsleetshowersandthunder": "lightssleetshowersandthunder"}

// symbolOf answers how code is shown: its icon where the set holds one, else its words, the code
// then told to the icons port the first time it is met in this run. An empty code is no symbol.
func (s *Service) symbolOf(code string) Symbol {
	if code == "" {
		return Symbol{}
	}
	name, variant, hasVariant := strings.Cut(code, variantMark)
	if spelled, ok := iconSpelling[name]; ok {
		name = spelled
	}
	if hasVariant {
		name += variantMark + variant
	}
	if s.ports.Icons.Has(name) {
		return Symbol{Icon: name}
	}
	if _, seen := s.missing.LoadOrStore(code, true); !seen {
		s.ports.Icons.Missing(code)
	}
	return Symbol{Words: strings.ReplaceAll(code, variantMark, " ")}
}
