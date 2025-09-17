package zeichenwerk

import (
	"reflect"
	"testing"
)

// TestStackPush tests the Push operation
func TestStackPush(t *testing.T) {
	t.Run("Push to empty stack", func(t *testing.T) {
		var stack Stack[int]
		
		length := stack.Push(42)
		
		if length != 1 {
			t.Errorf("Push() returned length = %d, want 1", length)
		}
		
		if stack.Len() != 1 {
			t.Errorf("Stack length = %d, want 1", stack.Len())
		}
		
		if stack.Peek() != 42 {
			t.Errorf("Stack top = %d, want 42", stack.Peek())
		}
	})
	
	t.Run("Push multiple elements", func(t *testing.T) {
		var stack Stack[string]
		
		length1 := stack.Push("first")
		length2 := stack.Push("second")
		length3 := stack.Push("third")
		
		if length1 != 1 || length2 != 2 || length3 != 3 {
			t.Errorf("Push() lengths = %d, %d, %d, want 1, 2, 3", length1, length2, length3)
		}
		
		if stack.Len() != 3 {
			t.Errorf("Stack length = %d, want 3", stack.Len())
		}
		
		if stack.Peek() != "third" {
			t.Errorf("Stack top = %q, want %q", stack.Peek(), "third")
		}
	})
	
	t.Run("Push large number of elements", func(t *testing.T) {
		var stack Stack[int]
		
		for i := 0; i < 1000; i++ {
			length := stack.Push(i)
			if length != i+1 {
				t.Errorf("Push(%d) returned length = %d, want %d", i, length, i+1)
			}
		}
		
		if stack.Len() != 1000 {
			t.Errorf("Stack length = %d, want 1000", stack.Len())
		}
		
		if stack.Peek() != 999 {
			t.Errorf("Stack top = %d, want 999", stack.Peek())
		}
	})
}

// TestStackPop tests the Pop operation
func TestStackPop(t *testing.T) {
	t.Run("Pop from stack with multiple elements", func(t *testing.T) {
		stack := Stack[string]{"first", "second", "third"}
		
		value1 := stack.Pop()
		if value1 != "third" {
			t.Errorf("Pop() = %q, want %q", value1, "third")
		}
		
		if stack.Len() != 2 {
			t.Errorf("Stack length after pop = %d, want 2", stack.Len())
		}
		
		value2 := stack.Pop()
		if value2 != "second" {
			t.Errorf("Pop() = %q, want %q", value2, "second")
		}
		
		value3 := stack.Pop()
		if value3 != "first" {
			t.Errorf("Pop() = %q, want %q", value3, "first")
		}
		
		if !stack.IsEmpty() {
			t.Error("Stack should be empty after popping all elements")
		}
	})
	
	t.Run("Pop from single element stack", func(t *testing.T) {
		stack := Stack[int]{42}
		
		value := stack.Pop()
		if value != 42 {
			t.Errorf("Pop() = %d, want 42", value)
		}
		
		if !stack.IsEmpty() {
			t.Error("Stack should be empty after popping last element")
		}
	})
}

// TestStackPopPanic tests that Pop panics on empty stack
func TestStackPopPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Pop() on empty stack should panic")
		}
	}()
	
	var stack Stack[int]
	stack.Pop() // Should panic
}

// TestStackPeek tests the Peek operation
func TestStackPeek(t *testing.T) {
	t.Run("Peek at single element", func(t *testing.T) {
		stack := Stack[float64]{3.14}
		
		value := stack.Peek()
		if value != 3.14 {
			t.Errorf("Peek() = %f, want 3.14", value)
		}
		
		// Verify stack is unchanged
		if stack.Len() != 1 {
			t.Errorf("Stack length after peek = %d, want 1", stack.Len())
		}
	})
	
	t.Run("Peek at multiple elements", func(t *testing.T) {
		stack := Stack[string]{"bottom", "middle", "top"}
		
		value1 := stack.Peek()
		if value1 != "top" {
			t.Errorf("Peek() = %q, want %q", value1, "top")
		}
		
		value2 := stack.Peek()
		if value2 != "top" {
			t.Errorf("Second Peek() = %q, want %q", value2, "top")
		}
		
		// Verify stack is unchanged
		if stack.Len() != 3 {
			t.Errorf("Stack length after peek = %d, want 3", stack.Len())
		}
	})
}

// TestStackPeekPanic tests that Peek panics on empty stack
func TestStackPeekPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Peek() on empty stack should panic")
		}
	}()
	
	var stack Stack[string]
	stack.Peek() // Should panic
}

// TestStackIsEmpty tests the IsEmpty method
func TestStackIsEmpty(t *testing.T) {
	t.Run("Empty stack", func(t *testing.T) {
		var stack Stack[int]
		
		if !stack.IsEmpty() {
			t.Error("New stack should be empty")
		}
	})
	
	t.Run("Non-empty stack", func(t *testing.T) {
		stack := Stack[string]{"element"}
		
		if stack.IsEmpty() {
			t.Error("Stack with elements should not be empty")
		}
	})
	
	t.Run("Stack after operations", func(t *testing.T) {
		var stack Stack[int]
		
		// Initially empty
		if !stack.IsEmpty() {
			t.Error("Initial stack should be empty")
		}
		
		// After push
		stack.Push(1)
		if stack.IsEmpty() {
			t.Error("Stack should not be empty after push")
		}
		
		// After pop
		stack.Pop()
		if !stack.IsEmpty() {
			t.Error("Stack should be empty after popping last element")
		}
	})
}

// TestStackLen tests the Len method
func TestStackLen(t *testing.T) {
	t.Run("Empty stack length", func(t *testing.T) {
		var stack Stack[bool]
		
		if stack.Len() != 0 {
			t.Errorf("Empty stack length = %d, want 0", stack.Len())
		}
	})
	
	t.Run("Stack length during operations", func(t *testing.T) {
		var stack Stack[int]
		
		// Start empty
		if stack.Len() != 0 {
			t.Errorf("Initial length = %d, want 0", stack.Len())
		}
		
		// Push elements
		for i := 1; i <= 5; i++ {
			stack.Push(i)
			if stack.Len() != i {
				t.Errorf("Length after %d pushes = %d, want %d", i, stack.Len(), i)
			}
		}
		
		// Pop elements
		for i := 5; i > 0; i-- {
			stack.Pop()
			if stack.Len() != i-1 {
				t.Errorf("Length after pop = %d, want %d", stack.Len(), i-1)
			}
		}
	})
}

// TestStackClear tests the Clear method
func TestStackClear(t *testing.T) {
	t.Run("Clear empty stack", func(t *testing.T) {
		var stack Stack[int]
		
		stack.Clear()
		
		if !stack.IsEmpty() {
			t.Error("Stack should be empty after clear")
		}
		
		if stack.Len() != 0 {
			t.Errorf("Stack length after clear = %d, want 0", stack.Len())
		}
	})
	
	t.Run("Clear stack with elements", func(t *testing.T) {
		stack := Stack[string]{"a", "b", "c", "d", "e"}
		
		stack.Clear()
		
		if !stack.IsEmpty() {
			t.Error("Stack should be empty after clear")
		}
		
		if stack.Len() != 0 {
			t.Errorf("Stack length after clear = %d, want 0", stack.Len())
		}
	})
	
	t.Run("Use stack after clear", func(t *testing.T) {
		stack := Stack[int]{1, 2, 3}
		
		stack.Clear()
		
		// Should be able to use normally after clear
		length := stack.Push(42)
		if length != 1 {
			t.Errorf("Push after clear returned length = %d, want 1", length)
		}
		
		if stack.Peek() != 42 {
			t.Errorf("Peek after clear = %d, want 42", stack.Peek())
		}
	})
}

// TestStackGenericTypes tests the stack with different types
func TestStackGenericTypes(t *testing.T) {
	t.Run("Stack of interfaces", func(t *testing.T) {
		var stack Stack[interface{}]
		
		stack.Push("string")
		stack.Push(42)
		stack.Push(3.14)
		stack.Push(true)
		
		if stack.Len() != 4 {
			t.Errorf("Stack length = %d, want 4", stack.Len())
		}
		
		// Check order (LIFO)
		if val := stack.Pop(); val != true {
			t.Errorf("Pop() = %v, want true", val)
		}
		if val := stack.Pop(); val != 3.14 {
			t.Errorf("Pop() = %v, want 3.14", val)
		}
		if val := stack.Pop(); val != 42 {
			t.Errorf("Pop() = %v, want 42", val)
		}
		if val := stack.Pop(); val != "string" {
			t.Errorf("Pop() = %v, want \"string\"", val)
		}
	})
	
	t.Run("Stack of structs", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		
		var stack Stack[Person]
		
		alice := Person{Name: "Alice", Age: 30}
		bob := Person{Name: "Bob", Age: 25}
		
		stack.Push(alice)
		stack.Push(bob)
		
		if stack.Len() != 2 {
			t.Errorf("Stack length = %d, want 2", stack.Len())
		}
		
		topPerson := stack.Peek()
		if topPerson.Name != "Bob" || topPerson.Age != 25 {
			t.Errorf("Peek() = %+v, want %+v", topPerson, bob)
		}
		
		poppedPerson := stack.Pop()
		if !reflect.DeepEqual(poppedPerson, bob) {
			t.Errorf("Pop() = %+v, want %+v", poppedPerson, bob)
		}
	})
	
	t.Run("Stack of pointers", func(t *testing.T) {
		var stack Stack[*int]
		
		val1, val2 := 10, 20
		stack.Push(&val1)
		stack.Push(&val2)
		
		if stack.Len() != 2 {
			t.Errorf("Stack length = %d, want 2", stack.Len())
		}
		
		peeked := stack.Peek()
		if *peeked != 20 {
			t.Errorf("*Peek() = %d, want 20", *peeked)
		}
		
		popped := stack.Pop()
		if *popped != 20 {
			t.Errorf("*Pop() = %d, want 20", *popped)
		}
	})
}

// TestStackLIFOBehavior tests Last-In-First-Out behavior
func TestStackLIFOBehavior(t *testing.T) {
	var stack Stack[int]
	
	// Push numbers 1-10
	for i := 1; i <= 10; i++ {
		stack.Push(i)
	}
	
	// Pop and verify they come out in reverse order
	for i := 10; i >= 1; i-- {
		value := stack.Pop()
		if value != i {
			t.Errorf("Pop() = %d, want %d (LIFO violation)", value, i)
		}
	}
	
	if !stack.IsEmpty() {
		t.Error("Stack should be empty after popping all elements")
	}
}

// TestStackOperationSequences tests complex operation sequences
func TestStackOperationSequences(t *testing.T) {
	t.Run("Mixed push/pop operations", func(t *testing.T) {
		var stack Stack[string]
		
		// Push, peek, push, pop, push, pop, pop
		stack.Push("first")
		
		if stack.Peek() != "first" {
			t.Errorf("Peek() = %q, want %q", stack.Peek(), "first")
		}
		
		stack.Push("second")
		
		val1 := stack.Pop()
		if val1 != "second" {
			t.Errorf("Pop() = %q, want %q", val1, "second")
		}
		
		stack.Push("third")
		
		val2 := stack.Pop()
		if val2 != "third" {
			t.Errorf("Pop() = %q, want %q", val2, "third")
		}
		
		val3 := stack.Pop()
		if val3 != "first" {
			t.Errorf("Pop() = %q, want %q", val3, "first")
		}
		
		if !stack.IsEmpty() {
			t.Error("Stack should be empty at end")
		}
	})
	
	t.Run("Push, clear, push sequence", func(t *testing.T) {
		var stack Stack[int]
		
		// Fill stack
		for i := 0; i < 5; i++ {
			stack.Push(i)
		}
		
		if stack.Len() != 5 {
			t.Errorf("Stack length before clear = %d, want 5", stack.Len())
		}
		
		// Clear and verify
		stack.Clear()
		
		if !stack.IsEmpty() {
			t.Error("Stack should be empty after clear")
		}
		
		// Use after clear
		stack.Push(100)
		
		if stack.Len() != 1 {
			t.Errorf("Stack length after push post-clear = %d, want 1", stack.Len())
		}
		
		if stack.Peek() != 100 {
			t.Errorf("Peek() after clear and push = %d, want 100", stack.Peek())
		}
	})
}

// TestStackConcurrentSafety tests basic concurrent access patterns
// Note: The stack is not designed to be thread-safe, but we can test
// basic scenarios to document expected behavior
func TestStackNonConcurrentSafety(t *testing.T) {
	// This test documents that the stack is NOT thread-safe
	// It's included for completeness but doesn't test concurrent access
	// as that would be inherently unsafe
	
	var stack Stack[int]
	
	// Sequential operations work fine
	stack.Push(1)
	stack.Push(2)
	stack.Push(3)
	
	if stack.Len() != 3 {
		t.Errorf("Sequential operations failed: length = %d, want 3", stack.Len())
	}
	
	// Document that concurrent access would require external synchronization
	t.Log("Note: Stack[T] is not thread-safe and requires external synchronization for concurrent access")
}

// BenchmarkStackPush benchmarks the Push operation
func BenchmarkStackPush(b *testing.B) {
	var stack Stack[int]
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.Push(i)
	}
}

// BenchmarkStackPop benchmarks the Pop operation
func BenchmarkStackPop(b *testing.B) {
	var stack Stack[int]
	
	// Pre-fill stack
	for i := 0; i < b.N; i++ {
		stack.Push(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.Pop()
	}
}

// BenchmarkStackPeek benchmarks the Peek operation
func BenchmarkStackPeek(b *testing.B) {
	stack := Stack[int]{42}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stack.Peek()
	}
}

// BenchmarkStackPushPop benchmarks mixed Push/Pop operations
func BenchmarkStackPushPop(b *testing.B) {
	var stack Stack[int]
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.Push(i)
		if i%2 == 1 && !stack.IsEmpty() {
			stack.Pop()
		}
	}
}