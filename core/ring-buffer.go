package core

import "sync"

// RingBuffer is a fixed-capacity, overwrite-on-full circular buffer for
// values of arbitrary type T. It is intended as a rolling history store
// (log messages, recent inputs, sparkline samples) where the newest N
// entries are always kept and older entries are silently discarded.
//
// The buffer is safe for concurrent use: Add and Clear acquire an internal
// mutex. Get, in contrast, is intentionally unlocked so that read-mostly
// consumers (such as a renderer drawing from a live log) do not block
// producers; see Get for the consistency caveat.
type RingBuffer[T any] struct {
	mu    sync.Mutex
	buf   []T
	size  int // total capacity; len(buf) is kept equal to this
	start int // index at which the next Add will write (== 1 + newest index)
	count int // number of valid entries, 0..size
}

// NewRingBuffer creates a ring buffer with the given fixed capacity. The
// buffer starts empty; up to size entries can be stored before older ones
// begin to be overwritten.
func NewRingBuffer[T any](size int) *RingBuffer[T] {
	return &RingBuffer[T]{buf: make([]T, size), size: size}
}

// Add inserts a new value at the current write position and advances that
// position by one, wrapping at capacity. When the buffer is already full
// the oldest entry is overwritten silently. The fill count is updated up
// to the capacity.
func (rb *RingBuffer[T]) Add(value T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buf[rb.start] = value
	rb.start = (rb.start + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	}
}

// Clear zeroes every slot of the underlying storage and resets the write
// position and fill count so the buffer appears empty. It is safe for
// concurrent callers but, like any mutating operation, races with
// unsynchronised Get calls.
func (rb *RingBuffer[T]) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	clear(rb.buf)
	rb.start = 0
	rb.count = 0
}

// Fill returns the number of valid entries currently held, in the range
// [0, Size()]. Once the buffer has been filled once, subsequent Adds do
// not grow this count further — they overwrite existing entries instead.
func (rb *RingBuffer[T]) Fill() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

// Get returns the index-th most recent entry: index 0 is the newest value
// previously passed to Add, index 1 the one before that, and so on up to
// Size()-1. If the buffer is not yet full, indices beyond Fill()-1 return
// the zero value of T (the unused slot) rather than an error.
//
// Get is intentionally lock-free so that read-heavy consumers do not
// contend with producers. As a consequence, a concurrent Add may have
// partially updated the slot being read; callers that require a consistent
// snapshot should synchronise externally.
//
// Panics if index is outside [0, Size()).
func (rb *RingBuffer[T]) Get(index int) T {
	if index < 0 || index >= rb.size {
		panic("RingBuffer.Get: index out of range")
	}
	return rb.buf[(rb.size+rb.start-1-index)%rb.size]
}

// Size returns the fixed capacity of the ring buffer.
func (rb *RingBuffer[T]) Size() int {
	return rb.size
}
