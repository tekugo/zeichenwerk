package zeichenwerk

// Stack is a generic LIFO (Last In, First Out) data structure implemented as a slice.
// It provides standard stack operations including push, pop, and peek functionality.
// The stack is type-safe and can hold any type specified by the generic parameter T.
//
// Features:
//   - Generic type support for type safety
//   - Standard LIFO stack operations
//   - Efficient slice-based implementation
//   - No capacity limits (grows dynamically)
//   - Zero-allocation peek operations
//
// The stack is commonly used within the TUI package for managing widget hierarchies,
// container nesting, and maintaining state during layout operations.
//
// Example usage:
//
//	// Create a stack of integers
//	var intStack Stack[int]
//	intStack.Push(1)
//	intStack.Push(2)
//	top := intStack.Peek() // Returns 2
//	value := intStack.Pop() // Returns 2
//
//	// Create a stack of widgets
//	var widgetStack Stack[Widget]
//	widgetStack.Push(someWidget)
type Stack[T any] []T

// Peek returns the top element of the stack without removing it.
// This operation does not modify the stack and has O(1) time complexity.
//
// Returns:
//   - T: The top element of the stack
//
// Panics:
//   - If the stack is empty (no bounds checking is performed)
//
// Example usage:
//
//	stack := Stack[string]{"a", "b", "c"}
//	top := stack.Peek() // Returns "c", stack remains ["a", "b", "c"]
func (s Stack[T]) Peek() T {
	return s[len(s)-1]
}

// Push adds a new element to the top of the stack.
// The stack grows dynamically to accommodate new elements.
// This operation has amortized O(1) time complexity.
//
// Parameters:
//   - value: The element to add to the top of the stack
//
// Returns:
//   - int: The new length of the stack after the push operation
//
// Example usage:
//
//	var stack Stack[int]
//	length := stack.Push(42)  // Returns 1, stack is now [42]
//	length = stack.Push(100)  // Returns 2, stack is now [42, 100]
func (s *Stack[T]) Push(value T) int {
	*s = append(*s, value)
	return len(*s)
}

// Pop removes and returns the top element from the stack.
// This operation modifies the stack by removing the most recently added element.
// The operation has O(1) time complexity.
//
// Returns:
//   - T: The element that was removed from the top of the stack
//
// Panics:
//   - If the stack is empty (no bounds checking is performed)
//
// Example usage:
//
//	stack := Stack[string]{"a", "b", "c"}
//	value := stack.Pop() // Returns "c", stack becomes ["a", "b"]
//	value = stack.Pop()  // Returns "b", stack becomes ["a"]
func (s *Stack[T]) Pop() T {
	l := len(*s)
	v := (*s)[l-1]
	*s = (*s)[:l-1]
	return v
}

// IsEmpty returns whether the stack contains no elements.
// This is a convenience method for checking if the stack is empty
// before performing operations that require elements.
//
// Returns:
//   - bool: true if the stack is empty, false otherwise
//
// Example usage:
//
//	var stack Stack[int]
//	if stack.IsEmpty() {
//	    fmt.Println("Stack is empty")
//	}
//	stack.Push(42)
//	if !stack.IsEmpty() {
//	    value := stack.Pop() // Safe to pop
//	}
func (s Stack[T]) IsEmpty() bool {
	return len(s) == 0
}

// Len returns the number of elements currently in the stack.
// This operation has O(1) time complexity.
//
// Returns:
//   - int: The number of elements in the stack
//
// Example usage:
//
//	stack := Stack[string]{"a", "b", "c"}
//	count := stack.Len() // Returns 3
func (s Stack[T]) Len() int {
	return len(s)
}

// Clear removes all elements from the stack, making it empty.
// This operation resets the stack to its initial state.
//
// Example usage:
//
//	stack := Stack[int]{1, 2, 3}
//	stack.Clear()
//	// stack is now empty
func (s *Stack[T]) Clear() {
	*s = (*s)[:0]
}
