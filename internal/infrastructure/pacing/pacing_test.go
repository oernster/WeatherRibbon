package pacing

import (
	"context"
	"errors"
	"testing"
	"time"
)

// NFR-S-2: a wait lasts at least as long as it was asked to.
func TestAWaitLastsItsLength(t *testing.T) {
	t.Parallel()
	const asked = 20 * time.Millisecond
	started := time.Now()
	if err := (Pacer{}).Wait(context.Background(), asked); err != nil {
		t.Fatal(err)
	}
	if waited := time.Since(started); waited < asked {
		t.Errorf("waited %v, asked for %v", waited, asked)
	}
}

// A wait answers at once, with the context's error, once the context has ended.
func TestAnEndedContextEndsTheWait(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	if err := (Pacer{}).Wait(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("answered %v", err)
	}
	if waited := time.Since(started); waited > time.Second {
		t.Errorf("an ended context still waited %v", waited)
	}
}
