package widgets

import (
	"sync"
	"time"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// Animation provides a base implementation for timed animations.
// It manages the ticker and stop channel, and calls the tick callback
// on each animation frame. Components that embed Animation should
// implement the Tick() method to define their specific animation behavior.
type Animation struct {
	Component
	mu     sync.Mutex
	ticker *time.Ticker
	fn     func()
	stop   chan struct{}
}

// ---- Widget Methods -------------------------------------------------------

// Refresh triggers a redraw of the widget. This method should be called
// by the Tick implementation to update the display.
func (a *Animation) Refresh() {
	Redraw(a)
}

// ---- Animation Control ----------------------------------------------------

// Running returns whether the animation is currently active.
func (a *Animation) Running() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ticker != nil
}

// Start begins the animation with the specified time interval.
// The animation runs in a separate goroutine. If already running,
// this method does nothing.
func (a *Animation) Start(interval time.Duration) {
	a.mu.Lock()
	if a.ticker != nil {
		a.mu.Unlock()
		a.Log(a, Error, "Animation already running")
		return
	}
	a.ticker = time.NewTicker(interval)
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			if a.ticker != nil {
				a.ticker.Stop()
				a.ticker = nil
			}
			a.mu.Unlock()
		}()

		for {
			select {
			case <-a.stop:
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
}

// Tick is called on each animation frame. Subtypes must implement this method
// to define their specific animation behavior.
func (a *Animation) Tick() {
	if a.fn != nil {
		a.fn()
	} else {
		a.Log(a, Warning, "No animation method set")
	}
}
