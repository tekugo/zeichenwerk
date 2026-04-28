package core

// Stack is a generic LIFO (last-in, first-out) container backed by a
// slice. It is deliberately declared as a type alias for []T rather than a
// struct so the zero value is immediately usable (`var s Stack[T]`), no
// explicit capacity is needed, and the underlying slice can be inspected
// or ranged over when that is convenient.
//
// The stack is not safe for concurrent use; callers that share a stack
// between goroutines must provide their own synchronisation. Peek and Pop
// assume the stack is non-empty; empty checks are the caller's
// responsibility via Empty or Len.
//
// Typical uses inside the framework include tracking container nesting
// during tree construction (Builder) and maintaining iteration state in
// tree traversal helpers.
//
// Example:
//
//	var s Stack[int]
//	s.Push(1)
//	s.Push(2)
//	top := s.Peek() // 2, stack unchanged
//	v := s.Pop()    // 2, stack is now [1]
type Stack[T any] []T

// Clear removes every element from the stack by truncating it to length
// zero. The underlying array is retained so subsequent pushes can reuse
// the existing capacity without reallocating. Held element values are not
// overwritten; if the stack contains pointers and retention matters for
// garbage collection, discard the stack entirely instead of clearing it.
func (s *Stack[T]) Clear() {
	*s = (*s)[:0]
}

// Empty reports whether the stack has no elements. It is equivalent to
// `s.Len() == 0` and is intended as a readable guard before a Peek or Pop.
func (s Stack[T]) Empty() bool {
	return len(s) == 0
}

// Len returns the number of elements currently on the stack in O(1).
func (s Stack[T]) Len() int {
	return len(s)
}

// Peek returns the top element without removing it. The stack is not
// modified and the operation runs in O(1).
//
// Panics if the stack is empty — callers should check Empty or Len first.
func (s Stack[T]) Peek() T {
	return s[len(s)-1]
}

// Pop removes and returns the top element. The operation runs in O(1) and
// retains the underlying array capacity so future pushes can reuse it.
//
// Panics if the stack is empty — callers should check Empty or Len first.
func (s *Stack[T]) Pop() T {
	l := len(*s)
	v := (*s)[l-1]
	*s = (*s)[:l-1]
	return v
}

// Push appends a value to the top of the stack and returns the new length.
// The backing slice grows automatically; amortised time complexity is O(1)
// while individual calls may be O(n) when a reallocation is required.
func (s *Stack[T]) Push(value T) int {
	*s = append(*s, value)
	return len(*s)
}
