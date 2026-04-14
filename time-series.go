package zeichenwerk

import (
	"iter"
	"time"
)

// Number is the type constraint for TimeSeries values.
type Number interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}

// TimeSeries is a fixed-size sliding window of numeric values indexed by
// evenly-spaced time slots. It is suitable for feeding sparklines with
// time-bucketed data.
//
// Internally it uses a ring buffer so that Shift runs in O(1) without allocation.
type TimeSeries[T Number] struct {
	start    time.Time
	interval time.Duration
	buf      []T  // ring buffer backing array, length == size
	head     int  // buf index of the oldest slot (logical index 0)
	auto     bool // advance the window automatically in Add/Set when t >= End()
}

// NewTimeSeries creates a TimeSeries with size zero-value slots.
// start is truncated to interval so that slot boundaries align with clock-wall
// multiples (the same time.Truncate semantics used for time bucketing).
// When auto is true, Add and Set automatically advance the window when a
// timestamp falls at or beyond End(), keeping the most recent data in view.
func NewTimeSeries[T Number](start time.Time, interval time.Duration, size int, auto bool) *TimeSeries[T] {
	return &TimeSeries[T]{
		start:    start.Truncate(interval),
		interval: interval,
		buf:      make([]T, size),
		auto:     auto,
	}
}

// All returns an iterator over all slots, oldest-first, yielding each slot's
// start time and value. Use as: for t, v := range ts.All() { ... }
func (ts *TimeSeries[T]) All() iter.Seq2[time.Time, T] {
	return func(yield func(time.Time, T) bool) {
		for i := range len(ts.buf) {
			t := ts.start.Add(time.Duration(i) * ts.interval)
			if !yield(t, ts.buf[ts.bufIndex(i)]) {
				return
			}
		}
	}
}

// Get returns the value of the slot covering t.
// Returns the zero value and false if t is out of range.
func (ts *TimeSeries[T]) Get(t time.Time) (T, bool) {
	i, ok := ts.slotIndex(t)
	if !ok {
		var zero T
		return zero, false
	}
	return ts.buf[ts.bufIndex(i)], true
}

// Add accumulates value into the slot that covers t.
// When auto is enabled and t >= End(), the window is shifted forward first
// so that t lands in the last slot.
// Out-of-range timestamps are silently ignored.
func (ts *TimeSeries[T]) Add(t time.Time, value T) {
	if ts.auto {
		ts.maybeShift(t)
	}
	i, ok := ts.slotIndex(t)
	if !ok {
		return
	}
	ts.buf[ts.bufIndex(i)] += value
}

// Clear resets all slot values to zero without changing the window position.
func (ts *TimeSeries[T]) Clear() {
	clear(ts.buf)
}

// End returns the exclusive end time of the window (Start + Size×Interval).
func (ts *TimeSeries[T]) End() time.Time {
	return ts.start.Add(time.Duration(len(ts.buf)) * ts.interval)
}

// Floats returns the slot values as []float64, oldest-first.
// Convenient for passing directly to Sparkline.SetValues.
func (ts *TimeSeries[T]) Floats() []float64 {
	out := make([]float64, len(ts.buf))
	for i := range len(ts.buf) {
		out[i] = float64(ts.buf[ts.bufIndex(i)])
	}
	return out
}

// Interval returns the slot width.
func (ts *TimeSeries[T]) Interval() time.Duration { return ts.interval }

// Set replaces the value of the slot that covers t.
// When auto is enabled and t >= End(), the window is shifted forward first
// so that t lands in the last slot.
// Out-of-range timestamps are silently ignored.
func (ts *TimeSeries[T]) Set(t time.Time, value T) {
	if ts.auto {
		ts.maybeShift(t)
	}
	i, ok := ts.slotIndex(t)
	if !ok {
		return
	}
	ts.buf[ts.bufIndex(i)] = value
}

// Shift advances the window by d, rounded down to a multiple of Interval.
// Slots that fall off the front are discarded; new slots at the back are zeroed.
func (ts *TimeSeries[T]) Shift(d time.Duration) {
	n := int(d / ts.interval)
	if n <= 0 {
		return
	}
	size := len(ts.buf)
	if n >= size {
		// All slots are recycled — zero the entire buffer.
		for i := range ts.buf {
			ts.buf[i] = 0
		}
		ts.head = (ts.head + n) % size
	} else {
		for range n {
			ts.buf[ts.head] = 0
			ts.head = (ts.head + 1) % size
		}
	}
	ts.start = ts.start.Add(time.Duration(n) * ts.interval)
}

// Size returns the fixed number of slots.
func (ts *TimeSeries[T]) Size() int { return len(ts.buf) }

// Start returns the start time of the oldest slot.
func (ts *TimeSeries[T]) Start() time.Time { return ts.start }

// Touch shifts the window forward until t falls within [Start, End), if needed.
// If t is already inside the window or before Start, it is a no-op.
func (ts *TimeSeries[T]) Touch(t time.Time) {
	ts.maybeShift(t)
}

// Values returns the slot values as a new slice, oldest-first.
func (ts *TimeSeries[T]) Values() []T {
	out := make([]T, len(ts.buf))
	for i := range len(ts.buf) {
		out[i] = ts.buf[ts.bufIndex(i)]
	}
	return out
}

// ---- Internal Helpers -----------------------------------------------------

// slotIndex maps t to a logical slot index in [0, size). Returns false if out of range.
func (ts *TimeSeries[T]) slotIndex(t time.Time) (int, bool) {
	slot := t.Truncate(ts.interval)
	if slot.Before(ts.start) {
		return 0, false
	}
	i := int(slot.Sub(ts.start) / ts.interval)
	if i >= len(ts.buf) {
		return 0, false
	}
	return i, true
}

// bufIndex converts a logical slot index to a physical buf index.
func (ts *TimeSeries[T]) bufIndex(slot int) int {
	return (ts.head + slot) % len(ts.buf)
}

// maybeShift advances the window if t falls at or beyond End().
// It shifts the minimum amount needed to place t in the last slot.
func (ts *TimeSeries[T]) maybeShift(t time.Time) {
	slot := t.Truncate(ts.interval)
	if !slot.Before(ts.End()) {
		last := ts.start.Add(time.Duration(len(ts.buf)-1) * ts.interval)
		ts.Shift(slot.Sub(last))
	}
}
