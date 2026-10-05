package backoff

import (
	"testing"
	"time"
)

// FR-306: 10 minutes after a failure, doubling after each further one to at most 2 hours; none
// after none.
func TestFailuresBackOffToTwoHours(t *testing.T) {
	t.Parallel()
	for failures, want := range map[int]time.Duration{
		-1:  0,
		0:   0,
		1:   10 * time.Minute,
		2:   20 * time.Minute,
		3:   40 * time.Minute,
		4:   80 * time.Minute,
		5:   2 * time.Hour,
		500: 2 * time.Hour,
	} {
		if got := Wait(failures); got != want {
			t.Errorf("Wait(%d) = %v; want %v", failures, got, want)
		}
	}
}
