package zeichenwerk

import (
	"time"
)

// Animation provides a base implementation for timed animations.
// It manages the ticker and stop channel, and calls the tick callback
// on each animation frame. Components that embed Animation should
// implement the Tick() method to define their specific animation behavior.
type Animation struct {
	Component
	ticker *time.Ticker
	fn     func()
	stop   chan struct{}
}

// Start begins the animation with the specified time interval.
// The animation runs in a separate goroutine. If already running,
// this method does nothing.
func (a *Animation) Start(interval time.Duration) {
	go func() {
		if a.ticker != nil {
			a.Log(a, "error", "Animation already running")
			return // already running
		}

		a.ticker = time.NewTicker(interval)
		defer a.ticker.Stop()

		for {
			select {
			case <-a.stop:
				a.ticker = nil
				return
			case <-a.ticker.C:
				a.Tick()
			}
		}
	}()
}

// Stop gracefully halts the animation.
func (a *Animation) Stop() {
	select {
	case a.stop <- struct{}{}:
	default:
	}
	a.ticker = nil
}

// Running returns whether the animation is currently active.
func (a *Animation) Running() bool {
	return a.ticker != nil
}

// Refresh triggers a redraw of the widget. This method should be called
// by the Tick implementation to update the display.
func (a *Animation) Refresh() {
	Redraw(a)
}

// Tick is called on each animation frame. Subtypes must implement this method
// to define their specific animation behavior.
func (a *Animation) Tick() {
	if a.fn != nil {
		a.fn()
	} else {
		a.Log(a, "warn", "No animation method set")
	}
}
