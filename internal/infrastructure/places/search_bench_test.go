package places

import (
	"slices"
	"testing"
	"time"
)

const (
	// typedPrefixes is how many keystrokes NFR-P-2 is measured over.
	typedPrefixes = 200
	// keystrokesPerName is how many letters of each sampled name are typed, one search per letter.
	keystrokesPerName = 4
	// percentile is the share of keystrokes NFR-P-2 holds to its bound.
	percentile = 95
)

// keystrokes answers what typing the start of evenly spaced names in the list sends the search: the
// first letter, then the first two, up to keystrokesPerName, typedPrefixes searches in all.
func keystrokes(places *Places) []string {
	ids := make([]int, 0, len(places.byID))
	for id := range places.byID {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	names := typedPrefixes / keystrokesPerName
	step := len(ids) / names
	typed := make([]string, 0, typedPrefixes)
	for index := range names {
		name := []rune(places.byID[ids[index*step]].Name)
		for letters := 1; letters <= keystrokesPerName; letters++ {
			typed = append(typed, string(name[:min(letters, len(name))]))
		}
	}
	return typed
}

// NFR-P-2: a search over the whole city list for each of the typed keystrokes, reporting the 95th
// percentile of one keystroke's answer in milliseconds beside the mean.
func BenchmarkSearchTheWholeCityListAsTyped(b *testing.B) {
	places, err := New()
	if err != nil {
		b.Fatal(err)
	}
	typed := keystrokes(places)
	var took []time.Duration
	for b.Loop() {
		for _, each := range typed {
			start := time.Now()
			places.Search(each)
			took = append(took, time.Since(start))
		}
	}
	slices.Sort(took)
	at := len(took) * percentile / 100
	b.ReportMetric(float64(took[min(at, len(took)-1)])/float64(time.Millisecond), "p95-ms/keystroke")
}
