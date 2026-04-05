package zeichenwerk

import (
	"strconv"
	"sync"
	"testing"
)

// captureSetter records calls to Set for use in Bind tests.
type captureSetter[T any] struct {
	mu     sync.Mutex
	values []T
}

func (c *captureSetter[T]) Set(value T) {
	c.mu.Lock()
	c.values = append(c.values, value)
	c.mu.Unlock()
}

func (c *captureSetter[T]) last() (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.values) == 0 {
		var zero T
		return zero, false
	}
	return c.values[len(c.values)-1], true
}

// TestNewValue verifies that NewValue stores the initial value correctly.
func TestNewValue(t *testing.T) {
	v := NewValue("hello")
	if got := v.Get(); got != "hello" {
		t.Errorf("Get() = %q, want %q", got, "hello")
	}
}

// TestValueGet verifies Get for several types.
func TestValueGet(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		v := NewValue("abc")
		if got := v.Get(); got != "abc" {
			t.Errorf("got %q, want %q", got, "abc")
		}
	})
	t.Run("int", func(t *testing.T) {
		v := NewValue(42)
		if got := v.Get(); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})
	t.Run("bool", func(t *testing.T) {
		v := NewValue(true)
		if got := v.Get(); !got {
			t.Errorf("got false, want true")
		}
	})
}

// TestValueSet verifies that Set updates the stored value.
func TestValueSet(t *testing.T) {
	v := NewValue(0)
	v.Set(99)
	if got := v.Get(); got != 99 {
		t.Errorf("Get() after Set = %d, want 99", got)
	}
}

// TestValueSetNotifiesSubscribers verifies that all subscribers are called with
// the new value.
func TestValueSetNotifiesSubscribers(t *testing.T) {
	v := NewValue(0)

	var calls []int
	v.Subscribe(func(n int) { calls = append(calls, n) }) // called once immediately
	v.Subscribe(func(n int) { calls = append(calls, n) }) // called once immediately

	// At this point calls = [0, 0] from the two immediate invocations.
	calls = nil

	v.Set(7)

	if len(calls) != 2 {
		t.Fatalf("expected 2 subscriber calls, got %d", len(calls))
	}
	for i, c := range calls {
		if c != 7 {
			t.Errorf("calls[%d] = %d, want 7", i, c)
		}
	}
}

// TestValueSubscribeImmediateCall verifies that Subscribe calls the callback
// with the current value right away.
func TestValueSubscribeImmediateCall(t *testing.T) {
	v := NewValue("initial")

	var received string
	v.Subscribe(func(s string) { received = s })

	if received != "initial" {
		t.Errorf("immediate call got %q, want %q", received, "initial")
	}
}

// TestValueSubscribeChaining verifies that Subscribe returns the receiver for
// method chaining.
func TestValueSubscribeChaining(t *testing.T) {
	v := NewValue(0)
	returned := v.Subscribe(func(int) {})
	if returned != v {
		t.Error("Subscribe did not return the receiver")
	}
}

// TestValueSubscribeFutureSets verifies that a subscriber is called on every
// subsequent Set.
func TestValueSubscribeFutureSets(t *testing.T) {
	v := NewValue(0)

	var got []int
	v.Subscribe(func(n int) { got = append(got, n) })
	got = nil // discard the immediate call

	v.Set(1)
	v.Set(2)
	v.Set(3)

	want := []int{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

// TestValueBindImmediateCall verifies that Bind calls Set on the setter with
// the current value immediately.
func TestValueBindImmediateCall(t *testing.T) {
	v := NewValue("hello")
	s := &captureSetter[string]{}
	v.Bind(s)

	if val, ok := s.last(); !ok || val != "hello" {
		t.Errorf("Bind immediate call: got %q ok=%v, want %q", val, ok, "hello")
	}
}

// TestValueBindForwardsSets verifies that the setter is called each time the
// value changes after Bind.
func TestValueBindForwardsSets(t *testing.T) {
	v := NewValue(0)
	s := &captureSetter[int]{}
	v.Bind(s)
	s.values = nil // discard the immediate call

	v.Set(10)
	v.Set(20)

	s.mu.Lock()
	got := append([]int(nil), s.values...)
	s.mu.Unlock()

	want := []int{10, 20}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

// TestValueBindChaining verifies that Bind returns the receiver for method
// chaining.
func TestValueBindChaining(t *testing.T) {
	v := NewValue(0)
	s := &captureSetter[int]{}
	returned := v.Bind(s)
	if returned != v {
		t.Error("Bind did not return the receiver")
	}
}

// TestValueBindMultipleSetters verifies that multiple Bind calls each receive
// all updates.
func TestValueBindMultipleSetters(t *testing.T) {
	v := NewValue("")
	a := &captureSetter[string]{}
	b := &captureSetter[string]{}
	v.Bind(a)
	v.Bind(b)
	a.values = nil
	b.values = nil

	v.Set("x")

	if val, ok := a.last(); !ok || val != "x" {
		t.Errorf("setter a: got %q ok=%v, want %q", val, ok, "x")
	}
	if val, ok := b.last(); !ok || val != "x" {
		t.Errorf("setter b: got %q ok=%v, want %q", val, ok, "x")
	}
}

// TestValueConcurrentSetGet exercises concurrent reads and writes to verify
// the mutex protects internal state.
func TestValueConcurrentSetGet(t *testing.T) {
	v := NewValue(0)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)
		n := i
		go func() { defer wg.Done(); v.Set(n) }()
		go func() { defer wg.Done(); _ = v.Get() }()
	}

	wg.Wait()
}

// TestValueConcurrentSubscribe exercises concurrent Subscribe and Set calls.
func TestValueConcurrentSubscribe(t *testing.T) {
	v := NewValue(0)
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(2)
		n := i
		go func() { defer wg.Done(); v.Subscribe(func(int) {}) }()
		go func() { defer wg.Done(); v.Set(n) }()
	}

	wg.Wait()
}

// TestValueObserveCheckbox verifies that a Value[bool] updates when a Checkbox
// is toggled.
func TestValueObserveCheckbox(t *testing.T) {
	check := NewCheckbox("cb", "", "agree", false)
	v := NewValue(false)
	v.Observe(check)

	check.Toggle() // true
	if got := v.Get(); !got {
		t.Errorf("after Toggle: got %v, want true", got)
	}

	check.Toggle() // false
	if got := v.Get(); got {
		t.Errorf("after second Toggle: got %v, want false", got)
	}
}

// TestValueObserveInput verifies that a Value[string] tracks each character
// inserted into an Input.
func TestValueObserveInput(t *testing.T) {
	input := NewInput("in", "")
	v := NewValue("")
	v.Observe(input)

	for _, ch := range "hi" {
		input.Insert(string(ch))
	}

	if got := v.Get(); got != "hi" {
		t.Errorf("got %q, want %q", got, "hi")
	}
}

// TestValueDerivedFromInput verifies that a Derived Value[int] stays in sync
// with an Input via an intermediate Value[string].
func TestValueDerivedFromInput(t *testing.T) {
	input := NewInput("in", "")
	str := NewValue("")
	str.Observe(input)

	num := Derived(str, func(s string) int {
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return n
	})

	for _, ch := range "42" {
		input.Insert(string(ch))
	}

	if got := num.Get(); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}
