package core

import (
	"iter"
	"time"
)

// TimeSeries is a fixed-size sliding window of numeric values indexed by
// evenly-spaced time slots. Every slot covers a half-open interval of
// length Interval(), and the full window spans [Start(), End()) — i.e.
// Size() consecutive slots starting at Start(). It is designed for feeding
// time-bucketed data into sparklines and similar visualisations where the
// most recent N buckets should remain in view and older ones can be
// discarded silently.
//
// Internally the values are kept in a circular buffer keyed by head index,
// so advancing the window (Shift) is O(1) per slot and allocation-free.
// When auto is true, Add, Set, and Touch shift the window forward on
// demand to keep freshly observed timestamps inside the window.
//
// TimeSeries is not safe for concurrent use; callers that share an
// instance between goroutines must provide their own synchronisation.
type TimeSeries[T Number] struct {
	start    time.Time
	interval time.Duration
	buf      []T  // circular backing array; len(buf) == slot count
	head     int  // index in buf corresponding to logical slot 0 (oldest)
	auto     bool // advance the window automatically when Add/Set/Touch see t >= End()
}

// NewTimeSeries creates a TimeSeries with the given number of slots, all
// pre-filled with the zero value of T. The supplied start time is
// truncated to a multiple of interval so that slot boundaries align with
// clock-wall units (the same semantics as time.Time.Truncate).
//
// Set auto to true for "live-tailing" data sources where the window
// should follow the most recent observations automatically; set it to
// false when the caller manages the window explicitly via Shift or Touch.
func NewTimeSeries[T Number](start time.Time, interval time.Duration, size int, auto bool) *TimeSeries[T] {
	return &TimeSeries[T]{
		start:    start.Truncate(interval),
		interval: interval,
		buf:      make([]T, size),
		auto:     auto,
	}
}

// Add increments (not replaces) the slot that covers t by value, which is
// the appropriate operation for counter-like metrics. Timestamps before
// Start() are dropped silently; timestamps at or beyond End() are dropped
// when auto is false, and trigger a forward shift that places t in the
// last slot when auto is true.
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

// All returns a range-compatible iterator over every slot in the window,
// oldest first. Each yielded pair consists of the slot's start time and
// its current value. Intended usage:
//
//	for t, v := range ts.All() {
//	    // ...
//	}
//
// The iterator takes a live view of the buffer; mutating the TimeSeries
// while iterating yields undefined results.
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

// At looks up the value stored in the slot that covers t. The ok flag
// distinguishes a legitimate zero from an out-of-range query: it is false
// (with a zero value) when t lies before Start() or at/after End().
func (ts *TimeSeries[T]) At(t time.Time) (T, bool) {
	i, ok := ts.slotIndex(t)
	if !ok {
		var zero T
		return zero, false
	}
	return ts.buf[ts.bufIndex(i)], true
}

// Clear resets every slot to the zero value of T. The window boundaries
// (Start, End, Interval) are unchanged.
func (ts *TimeSeries[T]) Clear() {
	clear(ts.buf)
}

// End returns the exclusive upper bound of the window: Start() plus
// Size() × Interval(). The last valid slot covers [End() - Interval, End()).
func (ts *TimeSeries[T]) End() time.Time {
	return ts.start.Add(time.Duration(len(ts.buf)) * ts.interval)
}

// Floats returns the slot values as []float64 in logical order (oldest
// first), converting from T as needed. The result is freshly allocated
// and can be handed directly to Sparkline.SetValues or any other consumer
// that expects float64 samples.
func (ts *TimeSeries[T]) Floats() []float64 {
	out := make([]float64, len(ts.buf))
	for i := range len(ts.buf) {
		out[i] = float64(ts.buf[ts.bufIndex(i)])
	}
	return out
}

// Get returns the value of the index-th most recent slot: index 0 is the
// newest, index Size()-1 is the oldest. This reversed indexing is the
// shape expected by the sparkline DataProvider interface, so
// TimeSeries[float64] can be passed to a Sparkline widget directly.
//
// No bounds checking is performed; indices outside [0, Size()) panic via
// the underlying slice access.
func (ts *TimeSeries[T]) Get(index int) T {
	return ts.buf[ts.bufIndex(len(ts.buf)-1-index)]
}

// Interval returns the width of a single slot.
func (ts *TimeSeries[T]) Interval() time.Duration { return ts.interval }

// Max returns the largest value in any slot. The window is assumed to be
// non-empty (Size() >= 1); zero-sized series will panic.
func (ts *TimeSeries[T]) Max() T {
	max := ts.buf[0]
	for i := 1; i < len(ts.buf); i++ {
		if v := ts.buf[i]; v > max {
			max = v
		}
	}
	return max
}

// Min returns the smallest value in any slot. The window is assumed to be
// non-empty (Size() >= 1); zero-sized series will panic.
func (ts *TimeSeries[T]) Min() T {
	min := ts.buf[0]
	for i := 1; i < len(ts.buf); i++ {
		if v := ts.buf[i]; v < min {
			min = v
		}
	}
	return min
}

// Set replaces the value of the slot that covers t, which is the
// appropriate operation for gauge-like metrics where each sample should
// overwrite the previous one. Timestamps before Start() are dropped;
// timestamps at or beyond End() are dropped when auto is false and
// trigger a forward shift that places t in the last slot when auto is true.
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

// Shift advances the window forward by d, rounded down to a whole number
// of intervals. Values smaller than one interval leave the window
// unchanged. Slots that fall off the old leading edge are discarded, and
// the corresponding number of new slots are appended at the trailing edge
// pre-filled with the zero value of T. When d covers the entire window or
// more, every slot is reset to zero in a single pass.
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

// Size returns the fixed number of slots in the window.
func (ts *TimeSeries[T]) Size() int { return len(ts.buf) }

// Start returns the start time of the oldest slot (inclusive).
func (ts *TimeSeries[T]) Start() time.Time { return ts.start }

// Touch advances the window forward by just enough intervals to include
// t, if t lies at or beyond End(). Timestamps already inside the window
// or before Start() are ignored. Use this when the auto flag is disabled
// but the caller occasionally needs to roll the window forward explicitly.
func (ts *TimeSeries[T]) Touch(t time.Time) {
	ts.maybeShift(t)
}

// Values returns a freshly allocated slice containing every slot value in
// logical order (oldest first). Mutating the returned slice does not
// affect the underlying buffer.
func (ts *TimeSeries[T]) Values() []T {
	out := make([]T, len(ts.buf))
	for i := range len(ts.buf) {
		out[i] = ts.buf[ts.bufIndex(i)]
	}
	return out
}

// ---- Internal Helpers -----------------------------------------------------

// slotIndex converts a timestamp to a logical slot index in [0, Size()).
// The timestamp is truncated to interval boundaries before indexing, so
// any t within a slot maps to the same index. Returns (0, false) when t
// falls outside the window.
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

// bufIndex converts a logical (oldest-first) slot index to the matching
// physical index in the circular buf by adding the head offset modulo
// the buffer size.
func (ts *TimeSeries[T]) bufIndex(slot int) int {
	return (ts.head + slot) % len(ts.buf)
}

// maybeShift rolls the window forward just enough for t to fall in the
// last slot, but only when t is already at or beyond End(). It is the
// shared helper behind the auto flag on Add/Set and behind Touch.
func (ts *TimeSeries[T]) maybeShift(t time.Time) {
	slot := t.Truncate(ts.interval)
	if !slot.Before(ts.End()) {
		last := ts.start.Add(time.Duration(len(ts.buf)-1) * ts.interval)
		ts.Shift(slot.Sub(last))
	}
}
