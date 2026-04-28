package widgets

import (
	"testing"
	"time"
)

// ── Running ───────────────────────────────────────────────────────────────────

func TestAnimation_Running_FalseInitially(t *testing.T) {
	a := &Animation{stop: make(chan struct{}, 1)}
	if a.Running() {
		t.Error("Running() should be false before Start()")
	}
}

// ── Start ─────────────────────────────────────────────────────────────────────

func TestAnimation_Start_SetsRunning(t *testing.T) {
	a := &Animation{stop: make(chan struct{}, 1)}
	a.Start(100 * time.Millisecond)
	defer a.Stop()
	if !a.Running() {
		t.Error("Running() should be true after Start()")
	}
}

func TestAnimation_Start_Twice_IsNoOp(t *testing.T) {
	a := &Animation{stop: make(chan struct{}, 1)}
	a.Start(100 * time.Millisecond)
	defer a.Stop()
	// Second Start should silently do nothing (no panic, still running)
	a.Start(100 * time.Millisecond)
	if !a.Running() {
		t.Error("Running() should still be true after double Start()")
	}
}

// ── Stop ─────────────────────────────────────────────────────────────────────

func TestAnimation_Stop_StopsRunning(t *testing.T) {
	a := &Animation{stop: make(chan struct{}, 1)}
	a.Start(10 * time.Millisecond)
	a.Stop()
	// Allow the goroutine time to clean up
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !a.Running() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	if a.Running() {
		t.Error("Running() should be false after Stop()")
	}
}

func TestAnimation_Stop_BeforeStart_NoPanic(t *testing.T) {
	a := &Animation{stop: make(chan struct{}, 1)}
	a.Stop() // should not panic
}

// ── Tick / fn ────────────────────────────────────────────────────────────────

func TestAnimation_Tick_CallsFn(t *testing.T) {
	called := false
	a := &Animation{
		stop: make(chan struct{}, 1),
		fn:   func() { called = true },
	}
	a.Tick()
	if !called {
		t.Error("Tick() should call fn when set")
	}
}

func TestAnimation_Tick_NoFn_NoPanic(t *testing.T) {
	a := &Animation{stop: make(chan struct{}, 1)}
	a.Tick() // fn is nil — should not panic
}
