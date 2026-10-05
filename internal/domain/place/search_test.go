package place

import (
	"slices"
	"testing"
)

// sample is a few places shaped like the city list's, out of population order on purpose.
var sample = []Place{
	{GeoNamesID: 6058560, Name: "London", ASCIIName: "London", Region: "Ontario", Country: "Canada", Population: 422324},
	{GeoNamesID: 5128581, Name: "New York City", ASCIIName: "New York City", Region: "New York", Country: "United States", Population: 8804190},
	{GeoNamesID: 2643743, Name: "London", ASCIIName: "London", Region: "England", Country: "United Kingdom", Population: 8961989},
	{GeoNamesID: 2633352, Name: "York", ASCIIName: "York", Region: "England", Country: "United Kingdom", Population: 144202},
	{GeoNamesID: 3448439, Name: "São Paulo", ASCIIName: "Sao Paulo", Region: "São Paulo", Country: "Brazil", Population: 10021295},
	{GeoNamesID: 2636713, Name: "Stratford-upon-Avon", ASCIIName: "Stratford-upon-Avon", Region: "England", Country: "United Kingdom", Population: 27445},
	{GeoNamesID: 4409896, Name: "Springfield", ASCIIName: "Springfield", Region: "Missouri", Country: "United States", Population: 169176},
	{GeoNamesID: 4250542, Name: "Springfield", ASCIIName: "Springfield", Region: "Illinois", Country: "United States", Population: 114394},
	{GeoNamesID: 3166273, Name: "Oslo", ASCIIName: "Oslo", Country: "Norway", Population: 580000},
}

// ids answers the GeoNames ids of places, in order.
func ids(places []Place) []int {
	var found []int
	for _, each := range places {
		found = append(found, each.GeoNamesID)
	}
	return found
}

// FR-202: what is typed is found at the start of a word of the name, the ASCII name, the region or the
// country, whatever its case; never inside a word.
func TestPlaceSearchMatchesNameRegionOrCountry(t *testing.T) {
	t.Parallel()
	index := NewIndex(sample)
	for typed, want := range map[string][]int{
		"london":         {2643743, 6058560},
		"  LONDON ":      {2643743, 6058560},
		"ontario":        {6058560},
		"kingdom":        {2643743, 2633352, 2636713},
		"sao":            {3448439},
		"são":            {3448439},
		"upon":           {2636713},
		"new york":       {5128581},
		"ork":            nil,
		"don":            nil,
		"springf":        {4409896, 4250542},
		"zzz":            nil,
		"":               nil,
		"   ":            nil,
		"norway":         {3166273},
		"united kingdom": {2643743, 2633352, 2636713},
	} {
		if got := ids(index.Search(typed)); !slices.Equal(got, want) {
			t.Errorf("Search(%q) = %v; want %v", typed, got, want)
		}
	}
}

// FR-202: places whose name begins with the text come first, then the rest; within each, larger
// population first. York, small, leads New York City, found only by a later word.
func TestLargerPlacesComeFirst(t *testing.T) {
	t.Parallel()
	got := ids(NewIndex(sample).Search("york"))
	if want := []int{2633352, 5128581}; !slices.Equal(got, want) {
		t.Errorf("Search(york) = %v; want York before New York City", got)
	}
}

// A word starts at the start of the field or after anything that is neither a letter nor a digit.
func TestAWordStartsAfterANonLetter(t *testing.T) {
	t.Parallel()
	for field, want := range map[[2]string]bool{
		{"ab", "b"}:            false,
		{"a-b", "b"}:           true,
		{"a b", "b"}:           true,
		{"a1b", "b"}:           false,
		{"l'aquila", "aquila"}: true,
		{"bab", "b"}:           true,
	} {
		if got := atWordStart(field[0], field[1]); got != want {
			t.Errorf("atWordStart(%q, %q) = %v; want %v", field[0], field[1], got, want)
		}
	}
}

// NewIndex keeps its own copy of the places.
func TestAnIndexKeepsItsOwnPlaces(t *testing.T) {
	t.Parallel()
	places := []Place{{GeoNamesID: 1, Name: "Oslo", Country: "Norway"}}
	index := NewIndex(places)
	places[0].Name = "Changed"
	if got := index.Search("oslo"); len(got) != 1 || got[0].Name != "Oslo" {
		t.Errorf("the index followed the caller's slice: %+v", got)
	}
}
