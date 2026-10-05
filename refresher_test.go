package main

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"
)

// refresherRig is a refresher over recording stand-ins at a fixed instant; its timer fires at once
// and records how long it was asked to wait.
type refresherRig struct {
	r        *refresher
	due      time.Time
	waits    []time.Duration
	asked    []string
	reported []string
	redraws  int
	halts    []error
	err      error
}

var refresherNow = time.Date(2026, time.October, 5, 7, 36, 0, 0, time.UTC)

func newRefresherRig() *refresherRig {
	g := &refresherRig{}
	ask := func(name string) func(context.Context) error {
		return func(context.Context) error {
			g.asked = append(g.asked, name)
			return g.err
		}
	}
	g.r = newRefresher(func() time.Time { return g.due }, ask("Refresh"), ask("RefreshNow"),
		func() { g.redraws++ }, func(doing string, _ error) { g.reported = append(g.reported, doing) },
		func(err error) { g.halts = append(g.halts, err) })
	g.r.now = func() time.Time { return refresherNow }
	g.r.after = func(wait time.Duration) <-chan time.Time {
		g.waits = append(g.waits, wait)
		fired := make(chan time.Time, 1)
		fired <- refresherNow.Add(wait)
		return fired
	}
	return g
}

// FR-303: the refresher waits until the next city falls due, refreshes, then has the page redraw; a
// city already due is refreshed without waiting.
func TestTheRefresherWaitsUntilACityIsDue(t *testing.T) {
	g := newRefresherRig()
	g.due = refresherNow.Add(5 * time.Minute)
	if !g.r.once(context.Background()) {
		t.Fatal("the loop ended")
	}
	g.due = refresherNow.Add(-time.Minute)
	g.r.once(context.Background())
	if !slices.Equal(g.waits, []time.Duration{5 * time.Minute, 0}) || !slices.Equal(g.asked, []string{"Refresh", "Refresh"}) || g.redraws != 2 {
		t.Errorf("waited %v, asked %v, redrew %d", g.waits, g.asked, g.redraws)
	}
}

// With nothing ever due the refresher waits only for a wake; a wake with nothing asked for refreshes
// nothing and looks again.
func TestAWakeWithNothingDueRefreshesNothing(t *testing.T) {
	g := newRefresherRig()
	g.r.poke()
	g.r.poke()
	if !g.r.once(context.Background()) || len(g.waits) != 0 || len(g.asked) != 0 || g.redraws != 0 {
		t.Errorf("waited %v, asked %v, redrew %d", g.waits, g.asked, g.redraws)
	}
}

// FR-309: Refresh now asks for every city backing off, on the refresher's goroutine, once.
func TestRefreshNowAsksOnTheRefreshersGoroutine(t *testing.T) {
	g := newRefresherRig()
	g.r.askNow()
	g.r.once(context.Background())
	g.due = refresherNow
	g.r.once(context.Background())
	if !slices.Equal(g.asked, []string{"RefreshNow", "Refresh"}) {
		t.Errorf("asked %v, want Refresh now once then the ordinary refresh", g.asked)
	}
}

// A refresh that fails is reported; one ended by the run's end is not; once the run has ended the
// loop stops.
func TestAFailedRefreshIsReportedAndTheEndStopsTheLoop(t *testing.T) {
	g := newRefresherRig()
	g.due, g.err = refresherNow, errPlanted
	g.r.once(context.Background())
	if !slices.Equal(g.reported, []string{refreshing}) {
		t.Errorf("reported %v", g.reported)
	}
	ended, end := context.WithCancel(context.Background())
	end()
	g.due = time.Time{}
	if g.r.once(ended) {
		t.Error("the loop went on after the run ended")
	}
	if len(g.reported) != 1 {
		t.Errorf("the run's end was reported: %v", g.reported)
	}
}

// A refresh that panics is caught on the refresher's goroutine, reported and raised as the notice
// that refreshing has stopped, with a redraw to show it; the loop then stops rather than spin on a
// refresh that fails the same way at once.
func TestAPanickingRefreshIsReportedAndStopsTheLoop(t *testing.T) {
	g := newRefresherRig()
	g.due = refresherNow
	g.r.refresh = func(context.Context) error { panic("planted") }
	done := make(chan struct{})
	go func() {
		defer close(done)
		g.r.run(context.Background())
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("the loop did not stop after the panic")
	}
	if !slices.Equal(g.reported, []string{refreshing}) {
		t.Errorf("reported %v", g.reported)
	}
	if len(g.halts) != 1 || !strings.Contains(g.halts[0].Error(), "planted") || g.redraws != 1 {
		t.Errorf("halted %v with %d redraws; want the panic said once and drawn", g.halts, g.redraws)
	}
}
