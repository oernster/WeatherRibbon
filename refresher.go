package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// refreshing is what the refresher says it was doing when it reports a failure.
const refreshing = "refreshing the forecasts"

// refresher keeps the forecasts fresh on a goroutine of its own: it waits until the next city falls
// due (FR-303), a wake or the end of the run, refreshes, then has the page redraw. Every forecast
// request leaves from this one goroutine, Refresh now's included (FR-309), so a menu never waits on
// the network.
type refresher struct {
	// due answers when the next city falls due; the zero time where none ever will (application's
	// NextDue). refresh and refreshNow ask for what is due (application's Refresh and RefreshNow).
	due        func() time.Time
	refresh    func(ctx context.Context) error
	refreshNow func(ctx context.Context) error
	// redraw has the page take a fresh snapshot; report writes a failure to the log; halted raises
	// the notice saying refreshing has stopped (application's RefreshStopped).
	redraw func()
	report func(doing string, err error)
	halted func(err error)
	// now and after are the clock and the timer, fields so a test can stand in for them.
	now   func() time.Time
	after func(wait time.Duration) <-chan time.Time

	wake chan struct{}
	// forced says Refresh now was asked for since the last refresh.
	forced atomic.Bool
}

// newRefresher answers a refresher over the given calls on the real clock.
func newRefresher(due func() time.Time, refresh, refreshNow func(context.Context) error, redraw func(), report func(string, error), halted func(error)) *refresher {
	return &refresher{
		due: due, refresh: refresh, refreshNow: refreshNow, redraw: redraw, report: report, halted: halted,
		now: time.Now, after: time.After, wake: make(chan struct{}, 1),
	}
}

// poke has the refresher look again at what is due, as after a city is added or moved. It never
// waits: a wake already waiting covers this one.
func (r *refresher) poke() {
	select {
	case r.wake <- struct{}{}:
	default:
	}
}

// askNow has the next refresh ask for every city still backing off as well, on the refresher's
// goroutine (FR-309).
func (r *refresher) askNow() {
	r.forced.Store(true)
	r.poke()
}

// run refreshes until ctx ends or a refresh panics. A panic is caught here (on the goroutine that
// raised it), logged and said on the ribbon as a notice, which a redraw shows at once. The loop stops
// rather than retry a refresh that fails the same way at once; the cells then say how old their
// forecasts grow (FR-305).
func (r *refresher) run(ctx context.Context) {
	defer func() {
		if failure := recover(); failure != nil {
			err := fmt.Errorf("stopped after a panic: %v", failure)
			r.report(refreshing, err)
			r.halted(err)
			r.redraw()
		}
	}()
	for r.once(ctx) {
	}
}

// once waits for the next thing to do and does it; false once ctx has ended.
func (r *refresher) once(ctx context.Context) bool {
	var due <-chan time.Time
	if at := r.due(); !at.IsZero() || r.forced.Load() {
		due = r.after(max(at.Sub(r.now()), 0))
	}
	select {
	case <-ctx.Done():
		return false
	case <-r.wake:
		if !r.forced.Load() {
			return true
		}
	case <-due:
	}
	// The refresh about to run looks at every city, so a wake already waiting is spent by it; one that
	// arrives while it runs still has the loop look again.
	select {
	case <-r.wake:
	default:
	}
	ask := r.refresh
	if r.forced.Swap(false) {
		ask = r.refreshNow
	}
	if err := ask(ctx); err != nil && ctx.Err() == nil {
		r.report(refreshing, err)
	}
	r.redraw()
	return ctx.Err() == nil
}
