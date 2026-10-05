package place

import (
	"cmp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// entry is one place with the text it is searched by, folded to lower case once rather than at every
// keystroke (NFR-P-2).
type entry struct {
	place  Place
	name   []string
	fields []string
}

// Index is the city list made ready to search (FR-202).
type Index struct {
	entries []entry
}

// NewIndex answers places made ready to search. The places are copied.
func NewIndex(places []Place) Index {
	entries := make([]entry, 0, len(places))
	for _, each := range places {
		name := []string{strings.ToLower(each.Name), strings.ToLower(each.ASCIIName)}
		fields := append(slices.Clone(name), strings.ToLower(each.Region), strings.ToLower(each.Country))
		entries = append(entries, entry{place: each, name: name, fields: fields})
	}
	return Index{entries: entries}
}

// Search answers the places matching typed (FR-202): typed, folded to lower case and trimmed, found
// at the start of a word of the place's name, its ASCII name, its region or its country. Places whose
// name begins with it come first, then the rest; within each, larger population first, then the
// order the list gives. Nothing is answered for nothing typed.
func (x Index) Search(typed string) []Place {
	wanted := strings.ToLower(strings.TrimSpace(typed))
	if wanted == "" {
		return nil
	}
	var leading, rest []Place
	for _, each := range x.entries {
		if !slices.ContainsFunc(each.fields, func(field string) bool { return atWordStart(field, wanted) }) {
			continue
		}
		if slices.ContainsFunc(each.name, func(name string) bool { return strings.HasPrefix(name, wanted) }) {
			leading = append(leading, each.place)
		} else {
			rest = append(rest, each.place)
		}
	}
	largestFirst := func(a, b Place) int { return cmp.Compare(b.Population, a.Population) }
	slices.SortStableFunc(leading, largestFirst)
	slices.SortStableFunc(rest, largestFirst)
	return append(leading, rest...)
}

// atWordStart answers whether wanted appears in field at its start or straight after a character
// that is neither a letter nor a digit: "york" in "new york", never "ork".
func atWordStart(field, wanted string) bool {
	for start := 0; start < len(field); {
		index := strings.Index(field[start:], wanted)
		if index < 0 {
			return false
		}
		at := start + index
		if previous, _ := utf8.DecodeLastRuneInString(field[:at]); at == 0 || !isWordRune(previous) {
			return true
		}
		start = at + 1
	}
	return false
}

// isWordRune answers whether r belongs inside a word.
func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }
