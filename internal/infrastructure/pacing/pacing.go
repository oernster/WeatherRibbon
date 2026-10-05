// Package pacing waits between forecast requests on the real clock, so that no two leave less than a
// second apart (NFR-S-2). The application says how long; this package only waits.
package pacing

import (
	"context"
	"time"
)

// Pacer waits on a timer.
type Pacer struct{}

// Wait waits for d, answering early with the context's error when the context ends first.
func (Pacer) Wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
