// Package backoff says how long to wait before asking again for a city whose requests have failed
// (FR-306).
package backoff

import "time"

// The first wait after a failure and the longest any wait grows to (FR-306).
const (
	First   = 10 * time.Minute
	Longest = 2 * time.Hour
)

// Wait answers how long to wait after failures failures in a row: none after none, First after one,
// doubling after each further failure to at most Longest. A success resets the count to none.
func Wait(failures int) time.Duration {
	if failures <= 0 {
		return 0
	}
	wait := First
	for range failures - 1 {
		if wait >= Longest {
			break
		}
		wait *= 2
	}
	return min(wait, Longest)
}
