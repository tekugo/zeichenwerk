package zeichenwerk

import (
	"sync"
	"testing"
)

func TestNewRingBuffer(t *testing.T) {
	rb := NewRingBuffer[int](5)
	if rb == nil {
		t.Fatal("NewRingBuffer returned nil")
	}
	if rb.size != 5 {
		t.Errorf("size = %d, want 5", rb.size)
	}
	if len(rb.buf) != 5 {
		t.Errorf("buffer len = %d, want 5", len(rb.buf))
	}
	if rb.Length() != 0 {
		t.Errorf("Length() = %d, want 0", rb.Length())
	}
}

func TestRingBufferAdd(t *testing.T) {
	rb := NewRingBuffer[int](3)

	rb.Add(10)
	if rb.Length() != 1 {
		t.Errorf("after 1 add: Length() = %d, want 1", rb.Length())
	}

	rb.Add(20)
	rb.Add(30)
	if rb.Length() != 3 {
		t.Errorf("after 3 adds: Length() = %d, want 3", rb.Length())
	}

	// Overflow: count must not exceed size
	rb.Add(40)
	if rb.Length() != 3 {
		t.Errorf("after overflow: Length() = %d, want 3", rb.Length())
	}
}

func TestRingBufferGet(t *testing.T) {
	rb := NewRingBuffer[string](3)
	rb.Add("A")
	rb.Add("B")
	rb.Add("C")

	if got := rb.Get(0); got != "C" {
		t.Errorf("Get(0) = %q, want %q", got, "C")
	}
	if got := rb.Get(1); got != "B" {
		t.Errorf("Get(1) = %q, want %q", got, "B")
	}
	if got := rb.Get(2); got != "A" {
		t.Errorf("Get(2) = %q, want %q", got, "A")
	}
}

func TestRingBufferGetAfterOverflow(t *testing.T) {
	rb := NewRingBuffer[int](3)
	rb.Add(1)
	rb.Add(2)
	rb.Add(3)
	rb.Add(4) // overwrites 1

	if got := rb.Get(0); got != 4 {
		t.Errorf("Get(0) = %d, want 4", got)
	}
	if got := rb.Get(1); got != 3 {
		t.Errorf("Get(1) = %d, want 3", got)
	}
	if got := rb.Get(2); got != 2 {
		t.Errorf("Get(2) = %d, want 2", got)
	}
}

func TestRingBufferGetSingleElement(t *testing.T) {
	rb := NewRingBuffer[int](4)
	rb.Add(99)

	if got := rb.Get(0); got != 99 {
		t.Errorf("Get(0) = %d, want 99", got)
	}
}

func TestRingBufferGetPanicsOutOfRange(t *testing.T) {
	rb := NewRingBuffer[int](3)
	rb.Add(1)

	cases := []int{-1, 3, 100}
	for _, idx := range cases {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Get(%d) should have panicked", idx)
				}
			}()
			rb.Get(idx)
		}()
	}
}

func TestRingBufferLength(t *testing.T) {
	rb := NewRingBuffer[int](4)

	for i := 1; i <= 4; i++ {
		rb.Add(i)
		if rb.Length() != i {
			t.Errorf("after %d adds: Length() = %d, want %d", i, rb.Length(), i)
		}
	}

	// Adding beyond capacity keeps length at size
	rb.Add(99)
	if rb.Length() != 4 {
		t.Errorf("after overflow: Length() = %d, want 4", rb.Length())
	}
}

func TestRingBufferConcurrentAdd(t *testing.T) {
	const size = 64
	const goroutines = 8
	const addsPerGoroutine = 1000

	rb := NewRingBuffer[int](size)

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for i := 0; i < addsPerGoroutine; i++ {
				rb.Add(base + i)
			}
		}(g * addsPerGoroutine)
	}
	wg.Wait()

	if rb.Length() != size {
		t.Errorf("after concurrent adds: Length() = %d, want %d", rb.Length(), size)
	}
}

func TestRingBufferSize1(t *testing.T) {
	rb := NewRingBuffer[string](1)
	rb.Add("first")
	if got := rb.Get(0); got != "first" {
		t.Errorf("Get(0) = %q, want %q", got, "first")
	}

	rb.Add("second")
	if got := rb.Get(0); got != "second" {
		t.Errorf("after overflow Get(0) = %q, want %q", got, "second")
	}
	if rb.Length() != 1 {
		t.Errorf("Length() = %d, want 1", rb.Length())
	}
}

func TestRingBufferOrdering(t *testing.T) {
	// Verify ordering is preserved across multiple wrap-arounds
	rb := NewRingBuffer[int](3)
	for i := 1; i <= 9; i++ {
		rb.Add(i)
	}
	// Last three added: 7, 8, 9
	if got := rb.Get(0); got != 9 {
		t.Errorf("Get(0) = %d, want 9", got)
	}
	if got := rb.Get(1); got != 8 {
		t.Errorf("Get(1) = %d, want 8", got)
	}
	if got := rb.Get(2); got != 7 {
		t.Errorf("Get(2) = %d, want 7", got)
	}
}

func BenchmarkRingBufferAdd(b *testing.B) {
	rb := NewRingBuffer[int](256)
	for i := 0; i < b.N; i++ {
		rb.Add(i)
	}
}

func BenchmarkRingBufferGet(b *testing.B) {
	rb := NewRingBuffer[int](256)
	for i := 0; i < 256; i++ {
		rb.Add(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rb.Get(i % 256)
	}
}
