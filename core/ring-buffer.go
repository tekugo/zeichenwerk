package core

import "sync"

type RingBuffer[T any] struct {
	mu    sync.Mutex
	buf   []T
	size  int // buffer size
	start int // start - current writing position
	count int // buffer fill
}

// NewRingBuffer creates a ring buffer with the given capacity.
func NewRingBuffer[T any](size int) *RingBuffer[T] {
	return &RingBuffer[T]{buf: make([]T, size), size: size}
}

// Add adds a new entry to the ring buffer.
// The ring buffer overflows automatically, overwriting the earliest entry.
func (rb *RingBuffer[T]) Add(value T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.buf[rb.start] = value
	rb.start = (rb.start + 1) % rb.size

	if rb.count < rb.size {
		rb.count++
	}
}

// Clear resets all values to zero and empties the buffer.
func (rb *RingBuffer[T]) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	clear(rb.buf)
	rb.start = 0
	rb.count = 0
}

// Fill returns the current fill of the ring buffer up to size.
func (rb *RingBuffer[T]) Fill() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

// Get returns the n-th latest entry starting with 0.
// Get is not synchronized, so if there is adding in progress, it might
// not return the latest entry.
// Panics if index < 0 or index >= size.
func (rb *RingBuffer[T]) Get(index int) T {
	if index < 0 || index >= rb.size {
		panic("RingBuffer.Get: index out of range")
	}
	return rb.buf[(rb.size+rb.start-1-index)%rb.size]
}

// Length returns the size of the ring buffer.
func (rb *RingBuffer[T]) Size() int {
	return rb.size
}
