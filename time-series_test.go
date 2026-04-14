package zeichenwerk

import (
	"testing"
	"time"
)

// base is a convenient fixed reference time (already aligned to minutes).
var base = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

const minute = time.Minute

func TestNewTimeSeries(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 5, false)

	if ts.Size() != 5 {
		t.Errorf("Size() = %d, want 5", ts.Size())
	}
	if ts.Interval() != minute {
		t.Errorf("Interval() = %v, want %v", ts.Interval(), minute)
	}
	if !ts.Start().Equal(base) {
		t.Errorf("Start() = %v, want %v", ts.Start(), base)
	}
	wantEnd := base.Add(5 * minute)
	if !ts.End().Equal(wantEnd) {
		t.Errorf("End() = %v, want %v", ts.End(), wantEnd)
	}
	for i, v := range ts.Values() {
		if v != 0 {
			t.Errorf("Values()[%d] = %d, want 0", i, v)
		}
	}
}

func TestNewTimeSeriesStartTruncation(t *testing.T) {
	// Start time with sub-minute offset should be truncated.
	unaligned := base.Add(30 * time.Second)
	ts := NewTimeSeries[int](unaligned, minute, 3, false)
	if !ts.Start().Equal(base) {
		t.Errorf("Start() = %v, want truncated %v", ts.Start(), base)
	}
}

func TestTimeSeriesAdd(t *testing.T) {
	ts := NewTimeSeries[int64](base, minute, 5, false)

	// Add to the first slot.
	ts.Add(base, 10)
	ts.Add(base.Add(30*time.Second), 5) // same slot
	if got := ts.Values()[0]; got != 15 {
		t.Errorf("slot 0 after two adds = %d, want 15", got)
	}

	// Add to the last slot.
	ts.Add(base.Add(4*minute), 99)
	if got := ts.Values()[4]; got != 99 {
		t.Errorf("slot 4 = %d, want 99", got)
	}

	// Other slots untouched.
	for i, v := range ts.Values() {
		if i == 0 || i == 4 {
			continue
		}
		if v != 0 {
			t.Errorf("slot %d = %d, want 0", i, v)
		}
	}
}

func TestTimeSeriesSet(t *testing.T) {
	ts := NewTimeSeries[float64](base, minute, 3, false)

	ts.Set(base, 1.0)
	ts.Set(base, 2.0) // second Set replaces, not accumulates
	if got := ts.Values()[0]; got != 2.0 {
		t.Errorf("slot 0 after two sets = %f, want 2.0", got)
	}
}

func TestTimeSeriesOutOfRange(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 3, false)
	ts.Add(base, 1)

	// Before start — ignored.
	ts.Add(base.Add(-minute), 99)
	// After end — ignored (autoShift off).
	ts.Add(base.Add(10*minute), 99)

	vals := ts.Values()
	if vals[0] != 1 {
		t.Errorf("slot 0 = %d, want 1", vals[0])
	}
	for i := 1; i < 3; i++ {
		if vals[i] != 0 {
			t.Errorf("slot %d = %d, want 0 (out-of-range write should be ignored)", i, vals[i])
		}
	}
}

func TestTimeSeriesShift(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 4, false)
	ts.Add(base, 1)
	ts.Add(base.Add(minute), 2)
	ts.Add(base.Add(2*minute), 3)
	ts.Add(base.Add(3*minute), 4)

	ts.Shift(minute)

	if !ts.Start().Equal(base.Add(minute)) {
		t.Errorf("Start() after shift = %v, want %v", ts.Start(), base.Add(minute))
	}
	want := []int{2, 3, 4, 0}
	for i, v := range ts.Values() {
		if v != want[i] {
			t.Errorf("slot %d = %d, want %d", i, v, want[i])
		}
	}
}

func TestTimeSeriesShiftFull(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 3, false)
	ts.Add(base, 1)
	ts.Add(base.Add(minute), 2)
	ts.Add(base.Add(2*minute), 3)

	ts.Shift(10 * minute) // far beyond size

	if !ts.Start().Equal(base.Add(10*minute)) {
		t.Errorf("Start() = %v, want %v", ts.Start(), base.Add(10*minute))
	}
	for i, v := range ts.Values() {
		if v != 0 {
			t.Errorf("slot %d = %d, want 0 after full shift", i, v)
		}
	}
}

func TestTimeSeriesShiftPartial(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 5, false)
	for i := range 5 {
		ts.Add(base.Add(time.Duration(i)*minute), i+1) // 1,2,3,4,5
	}

	ts.Shift(2 * minute)

	want := []int{3, 4, 5, 0, 0}
	for i, v := range ts.Values() {
		if v != want[i] {
			t.Errorf("slot %d = %d, want %d", i, v, want[i])
		}
	}
	if !ts.Start().Equal(base.Add(2 * minute)) {
		t.Errorf("Start() = %v, want %v", ts.Start(), base.Add(2*minute))
	}
}

func TestTimeSeriesShiftSubInterval(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 3, false)
	ts.Add(base, 42)

	// Shift by less than one interval — should be a no-op.
	ts.Shift(30 * time.Second)

	if !ts.Start().Equal(base) {
		t.Errorf("Start() changed after sub-interval shift")
	}
	if ts.Values()[0] != 42 {
		t.Errorf("slot 0 changed after sub-interval shift")
	}
}

func TestTimeSeriesAutoShift(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 3, true)
	ts.Add(base, 1)            // slot 0
	ts.Add(base.Add(minute), 2) // slot 1

	// Add a value 2 minutes beyond End() — should auto-shift and land in last slot.
	future := base.Add(5 * minute)
	ts.Add(future, 99)

	wantStart := future.Add(-2 * minute) // last slot (index 2) holds future
	if !ts.Start().Equal(wantStart) {
		t.Errorf("Start() after auto-shift = %v, want %v", ts.Start(), wantStart)
	}
	vals := ts.Values()
	if vals[2] != 99 {
		t.Errorf("last slot = %d, want 99", vals[2])
	}
}

func TestTimeSeriesAutoShiftSet(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 3, true)

	future := base.Add(10 * minute)
	ts.Set(future, 7)

	wantStart := future.Add(-2 * minute)
	if !ts.Start().Equal(wantStart) {
		t.Errorf("Start() after auto-shift Set = %v, want %v", ts.Start(), wantStart)
	}
	if got := ts.Values()[2]; got != 7 {
		t.Errorf("last slot = %d, want 7", got)
	}
}

func TestTimeSeriesAutoShiftWithinWindow(t *testing.T) {
	// When t is within the window, autoShift should not move it.
	ts := NewTimeSeries[int](base, minute, 5, true)
	ts.Add(base.Add(2*minute), 42) // within window, no shift expected

	if !ts.Start().Equal(base) {
		t.Errorf("Start() moved unexpectedly: %v", ts.Start())
	}
	if ts.Values()[2] != 42 {
		t.Errorf("slot 2 = %d, want 42", ts.Values()[2])
	}
}

func TestTimeSeriesValues(t *testing.T) {
	ts := NewTimeSeries[int](base, minute, 3, false)
	ts.Add(base, 1)
	ts.Add(base.Add(minute), 2)
	ts.Add(base.Add(2*minute), 3)

	vals := ts.Values()

	// Mutating the returned slice should not affect the TimeSeries.
	vals[0] = 999
	if ts.Values()[0] != 1 {
		t.Error("Values() returned a reference into internal state (not a copy)")
	}

	// Correct order: oldest first.
	if vals[0] != 999 || ts.Values()[1] != 2 || ts.Values()[2] != 3 {
		t.Errorf("Values() order wrong: %v", ts.Values())
	}
}

func TestTimeSeriesFloats(t *testing.T) {
	ts := NewTimeSeries[int64](base, minute, 3, false)
	ts.Add(base, 10)
	ts.Add(base.Add(minute), 20)
	ts.Add(base.Add(2*minute), 30)

	fv := ts.Floats()
	want := []float64{10, 20, 30}
	for i, v := range fv {
		if v != want[i] {
			t.Errorf("Floats()[%d] = %f, want %f", i, v, want[i])
		}
	}
}

func TestTimeSeriesEndToEnd(t *testing.T) {
	// Simulate a sparkline feed: collect values, then scroll the window forward.
	ts := NewTimeSeries[float64](base, minute, 5, false)

	for i := range 5 {
		ts.Add(base.Add(time.Duration(i)*minute), float64(i+1))
	}

	// Scroll forward by 2 minutes (two new zero buckets appear at end).
	ts.Shift(2 * minute)

	fv := ts.Floats()
	want := []float64{3, 4, 5, 0, 0}
	for i, v := range fv {
		if v != want[i] {
			t.Errorf("Floats()[%d] = %f, want %f", i, v, want[i])
		}
	}
}

func TestTimeSeriesMultipleShifts(t *testing.T) {
	// Shifting in multiple small steps should equal one large shift.
	ts1 := NewTimeSeries[int](base, minute, 4, false)
	ts2 := NewTimeSeries[int](base, minute, 4, false)
	for i := range 4 {
		ts1.Add(base.Add(time.Duration(i)*minute), i+1)
		ts2.Add(base.Add(time.Duration(i)*minute), i+1)
	}

	ts1.Shift(3 * minute)
	ts2.Shift(minute)
	ts2.Shift(minute)
	ts2.Shift(minute)

	v1, v2 := ts1.Values(), ts2.Values()
	for i := range v1 {
		if v1[i] != v2[i] {
			t.Errorf("slot %d: single-shift=%d, step-shift=%d", i, v1[i], v2[i])
		}
	}
	if !ts1.Start().Equal(ts2.Start()) {
		t.Errorf("Start() diverged: %v vs %v", ts1.Start(), ts2.Start())
	}
}
