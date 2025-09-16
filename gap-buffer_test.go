package zeichenwerk

import (
	"strings"
	"testing"
)

// TestNewGapBuffer tests the creation of new gap buffers
func TestNewGapBuffer(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		wantCap  int
	}{
		{"normal capacity", 10, 10},
		{"minimum capacity", 2, 2},
		{"below minimum capacity", 1, 2},
		{"zero capacity", 0, 2},
		{"negative capacity", -5, 2},
		{"large capacity", 1000, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gb := NewGapBuffer(tt.capacity)
			if gb == nil {
				t.Fatal("NewGapBuffer returned nil")
			}
			if len(gb.buffer) != tt.wantCap {
				t.Errorf("NewGapBuffer() capacity = %d, want %d", len(gb.buffer), tt.wantCap)
			}
			if gb.start != 0 {
				t.Errorf("NewGapBuffer() start = %d, want 0", gb.start)
			}
			if gb.end != tt.wantCap {
				t.Errorf("NewGapBuffer() end = %d, want %d", gb.end, tt.wantCap)
			}
			if gb.Length() != 0 {
				t.Errorf("NewGapBuffer() Length() = %d, want 0", gb.Length())
			}
		})
	}
}

// TestNewGapBufferFromString tests the creation of gap buffers initialized with text
func TestNewGapBufferFromString(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		gapCapacity int
		wantGapCap  int
	}{
		{"normal text and gap", "hello", 5, 5},
		{"empty text", "", 10, 10},
		{"minimum gap capacity", "test", 1, 1},
		{"below minimum gap capacity", "test", 0, 1},
		{"negative gap capacity", "test", -5, 1},
		{"unicode text", "Hällö Wörld", 3, 3},
		{"large text", strings.Repeat("a", 1000), 100, 100},
		{"emoji text", "🚀🌟💫", 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gb := NewGapBufferFromString(tt.text, tt.gapCapacity)
			if gb == nil {
				t.Fatal("NewGapBufferFromString returned nil")
			}
			
			// Check total capacity
			expectedTotalCap := len([]rune(tt.text)) + tt.wantGapCap
			if len(gb.buffer) != expectedTotalCap {
				t.Errorf("NewGapBufferFromString() total capacity = %d, want %d", len(gb.buffer), expectedTotalCap)
			}
			
			// Check gap position (should be at end of text)
			expectedStart := len([]rune(tt.text))
			if gb.start != expectedStart {
				t.Errorf("NewGapBufferFromString() start = %d, want %d", gb.start, expectedStart)
			}
			
			// Check gap end
			if gb.end != expectedTotalCap {
				t.Errorf("NewGapBufferFromString() end = %d, want %d", gb.end, expectedTotalCap)
			}
			
			// Check length matches text length
			if gb.Length() != len([]rune(tt.text)) {
				t.Errorf("NewGapBufferFromString() Length() = %d, want %d", gb.Length(), len([]rune(tt.text)))
			}
			
			// Check content matches input text
			if gb.String() != tt.text {
				t.Errorf("NewGapBufferFromString() String() = %q, want %q", gb.String(), tt.text)
			}
		})
	}
}

// TestNewGapBufferFromStringOperations tests operations on buffers created from strings
func TestNewGapBufferFromStringOperations(t *testing.T) {
	// Test basic operations after creation from string
	gb := NewGapBufferFromString("hello", 5)
	
	// Test insertion at end (gap position)
	gb.Insert('!')
	if gb.String() != "hello!" {
		t.Errorf("After inserting at end, got %q, want %q", gb.String(), "hello!")
	}
	
	// Test insertion at beginning
	gb.Move(0)
	gb.Insert('H')
	if gb.String() != "Hhello!" {
		t.Errorf("After inserting at beginning, got %q, want %q", gb.String(), "Hhello!")
	}
	
	// Test deletion
	gb.Move(0)
	gb.Delete()
	if gb.String() != "hello!" {
		t.Errorf("After deleting first character, got %q, want %q", gb.String(), "hello!")
	}
	
	// Test find operation
	matches := gb.Find("ll")
	expected := []int{2}
	if len(matches) != len(expected) || (len(matches) > 0 && matches[0] != expected[0]) {
		t.Errorf("Find('ll') = %v, want %v", matches, expected)
	}
}

// TestNewGapBufferFromStringWithComplexText tests with complex text scenarios
func TestNewGapBufferFromStringWithComplexText(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"multiline text", "Line 1\nLine 2\nLine 3"},
		{"text with tabs", "Column1\tColumn2\tColumn3"},
		{"mixed unicode", "Hello 世界 🌍 Düsseldorf"},
		{"special characters", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"repeated patterns", "abababab"},
		{"palindrome", "racecar"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gb := NewGapBufferFromString(tt.text, 10)
			
			// Verify initial state
			if gb.String() != tt.text {
				t.Errorf("Initial string = %q, want %q", gb.String(), tt.text)
			}
			
			// Test cursor movement and operations
			textLen := len([]rune(tt.text))
			if textLen > 0 {
				// Move to middle and insert
				midPos := textLen / 2
				gb.Move(midPos)
				gb.Insert('X')
				
				// Verify insertion
				result := gb.String()
				runes := []rune(tt.text)
				expected := string(runes[:midPos]) + "X" + string(runes[midPos:])
				if result != expected {
					t.Errorf("After insertion at middle, got %q, want %q", result, expected)
				}
				
				// Remove the inserted character
				gb.Move(midPos)
				gb.Delete()
				if gb.String() != tt.text {
					t.Errorf("After removing insertion, got %q, want %q", gb.String(), tt.text)
				}
			}
		})
	}
}

// TestInsert tests insertion of runes
func TestInsert(t *testing.T) {
	gb := NewGapBuffer(10)

	// Test single insertion
	gb.Insert('a')
	if gb.Length() != 1 {
		t.Errorf("After inserting 'a', Length() = %d, want 1", gb.Length())
	}
	if gb.String() != "a" {
		t.Errorf("After inserting 'a', String() = %q, want %q", gb.String(), "a")
	}

	// Test multiple insertions
	gb.Insert('b')
	gb.Insert('c')
	if gb.Length() != 3 {
		t.Errorf("After inserting 'abc', Length() = %d, want 3", gb.Length())
	}
	if gb.String() != "abc" {
		t.Errorf("After inserting 'abc', String() = %q, want %q", gb.String(), "abc")
	}

	// Test Unicode characters
	gb.Insert('ä')
	gb.Insert('ü')
	gb.Insert('ö')
	if gb.Length() != 6 {
		t.Errorf("After inserting Unicode chars, Length() = %d, want 6", gb.Length())
	}
	if gb.String() != "abcäüö" {
		t.Errorf("After inserting Unicode chars, String() = %q, want %q", gb.String(), "abcäüö")
	}
}

// TestInsertWithResize tests that insertion works correctly when buffer needs to resize
func TestInsertWithResize(t *testing.T) {
	gb := NewGapBuffer(2) // Start with minimal capacity

	// Insert more characters than initial capacity
	text := "hello world"
	for _, r := range text {
		gb.Insert(r)
	}

	if gb.Length() != len(text) {
		t.Errorf("After inserting %q, Length() = %d, want %d", text, gb.Length(), len(text))
	}
	if gb.String() != text {
		t.Errorf("After inserting %q, String() = %q, want %q", text, gb.String(), text)
	}
	
	// Verify buffer has grown
	if len(gb.buffer) < len(text) {
		t.Errorf("Buffer capacity %d should be at least %d after resize", len(gb.buffer), len(text))
	}
}

// TestDelete tests deletion of characters
func TestDelete(t *testing.T) {
	gb := NewGapBuffer(10)
	
	// Insert some text
	text := "hello"
	for _, r := range text {
		gb.Insert(r)
	}

	// Move cursor to beginning and delete
	gb.Move(0)
	initialLength := gb.Length()
	gb.Delete()
	
	if gb.Length() != initialLength-1 {
		t.Errorf("After delete, Length() = %d, want %d", gb.Length(), initialLength-1)
	}
	if gb.String() != "ello" {
		t.Errorf("After delete, String() = %q, want %q", gb.String(), "ello")
	}

	// Delete from middle
	gb.Move(1)
	gb.Delete()
	if gb.String() != "elo" {
		t.Errorf("After delete from middle, String() = %q, want %q", gb.String(), "elo")
	}

	// Delete at end (should have no effect)
	gb.Move(gb.Length())
	lengthBefore := gb.Length()
	gb.Delete()
	if gb.Length() != lengthBefore {
		t.Errorf("Delete at end should not change length, got %d, want %d", gb.Length(), lengthBefore)
	}
}

// TestMove tests cursor movement
func TestMove(t *testing.T) {
	gb := NewGapBuffer(10)
	
	// Insert some text
	text := "abcdef"
	for _, r := range text {
		gb.Insert(r)
	}

	// Test moving to various positions
	positions := []int{0, 1, 3, 5, 6}
	for _, pos := range positions {
		gb.Move(pos)
		
		// Insert a marker to verify position
		marker := 'X'
		gb.Insert(marker)
		
		// Check that marker is at expected position
		result := gb.String()
		if pos >= len(result) || result[pos] != byte(marker) {
			t.Errorf("After moving to %d and inserting %c, result = %q, marker not at expected position", pos, marker, result)
		}
		
		// Remove the marker for next test
		gb.Move(pos)
		gb.Delete()
	}
}

// TestMovePanic tests that Move panics on invalid positions
func TestMovePanic(t *testing.T) {
	gb := NewGapBuffer(10)
	gb.Insert('a')
	gb.Insert('b')

	tests := []struct {
		name string
		pos  int
	}{
		{"negative position", -1},
		{"position beyond length", gb.Length() + 1},
		{"large negative position", -100},
		{"large positive position", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Move(%d) should have panicked", tt.pos)
				}
			}()
			gb.Move(tt.pos)
		})
	}
}

// TestLength tests the Length method
func TestLength(t *testing.T) {
	gb := NewGapBuffer(10)

	// Empty buffer
	if gb.Length() != 0 {
		t.Errorf("Empty buffer Length() = %d, want 0", gb.Length())
	}

	// Add characters and test length
	for i := 1; i <= 5; i++ {
		gb.Insert('a')
		if gb.Length() != i {
			t.Errorf("After inserting %d characters, Length() = %d, want %d", i, gb.Length(), i)
		}
	}

	// Delete characters and test length
	gb.Move(0)
	for i := 4; i >= 0; i-- {
		gb.Delete()
		if gb.Length() != i {
			t.Errorf("After deleting, Length() = %d, want %d", gb.Length(), i)
		}
	}
}

// TestString tests the String method
func TestString(t *testing.T) {
	gb := NewGapBuffer(10)

	// Empty buffer
	if gb.String() != "" {
		t.Errorf("Empty buffer String() = %q, want empty string", gb.String())
	}

	// Test with various strings
	testStrings := []string{
		"a",
		"hello",
		"hello world",
		"äöü",
		"🚀🌟💫",
		"mixed äöü 🚀 text",
	}

	for _, str := range testStrings {
		gb = NewGapBuffer(10)
		for _, r := range str {
			gb.Insert(r)
		}
		if gb.String() != str {
			t.Errorf("After inserting %q, String() = %q, want %q", str, gb.String(), str)
		}
	}
}

// TestRunes tests the Runes iterator
func TestRunes(t *testing.T) {
	gb := NewGapBuffer(10)
	text := "hello"
	
	for _, r := range text {
		gb.Insert(r)
	}

	// Test iteration from beginning
	var result []rune
	for r := range gb.Runes(0) {
		result = append(result, r)
	}
	
	if string(result) != text {
		t.Errorf("Runes(0) iteration result = %q, want %q", string(result), text)
	}

	// Test iteration from middle
	result = nil
	for r := range gb.Runes(2) {
		result = append(result, r)
	}
	
	expected := text[2:]
	if string(result) != expected {
		t.Errorf("Runes(2) iteration result = %q, want %q", string(result), expected)
	}

	// Test iteration from invalid position
	result = nil
	for r := range gb.Runes(-1) {
		result = append(result, r)
	}
	
	if len(result) != 0 {
		t.Errorf("Runes(-1) should return empty iteration, got %d runes", len(result))
	}

	result = nil
	for r := range gb.Runes(gb.Length()) {
		result = append(result, r)
	}
	
	if len(result) != 0 {
		t.Errorf("Runes(Length()) should return empty iteration, got %d runes", len(result))
	}
}

// TestRunesWithGapMovement tests Runes iterator with different gap positions
func TestRunesWithGapMovement(t *testing.T) {
	gb := NewGapBuffer(10)
	text := "abcdef"
	
	for _, r := range text {
		gb.Insert(r)
	}

	// Move gap to different positions and test iteration
	positions := []int{0, 1, 3, len(text)}
	for _, pos := range positions {
		gb.Move(pos)
		
		var result []rune
		for r := range gb.Runes(0) {
			result = append(result, r)
		}
		
		if string(result) != text {
			t.Errorf("After moving gap to %d, Runes(0) = %q, want %q", pos, string(result), text)
		}
	}
}

// TestFind tests the Find method
func TestFind(t *testing.T) {
	gb := NewGapBuffer(20)
	text := "hello world hello"
	
	for _, r := range text {
		gb.Insert(r)
	}

	tests := []struct {
		name     string
		pattern  string
		expected []int
	}{
		{"simple match", "hello", []int{0, 12}},
		{"single character", "o", []int{4, 7, 16}},
		{"no match", "xyz", []int{}},
		{"empty pattern", "", []int{}},
		{"pattern longer than text", "this is way too long", []int{}},
		{"exact match", text, []int{0}},
		{"space character", " ", []int{5, 11}},
		{"overlapping pattern", "ll", []int{2, 14}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gb.Find(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("Find(%q) returned %d matches, want %d", tt.pattern, len(result), len(tt.expected))
				return
			}
			for i, pos := range result {
				if pos != tt.expected[i] {
					t.Errorf("Find(%q) match %d at position %d, want %d", tt.pattern, i, pos, tt.expected[i])
				}
			}
		})
	}
}

// TestFindWithGapMovement tests Find with different gap positions
func TestFindWithGapMovement(t *testing.T) {
	gb := NewGapBuffer(20)
	text := "abcdefghijk"
	
	for _, r := range text {
		gb.Insert(r)
	}

	pattern := "def"
	expected := []int{3}

	// Test find with gap at different positions
	positions := []int{0, 1, 5, 8, len(text)}
	for _, pos := range positions {
		gb.Move(pos)
		result := gb.Find(pattern)
		
		if len(result) != len(expected) {
			t.Errorf("With gap at %d, Find(%q) returned %d matches, want %d", pos, pattern, len(result), len(expected))
			continue
		}
		for i, matchPos := range result {
			if matchPos != expected[i] {
				t.Errorf("With gap at %d, Find(%q) match %d at position %d, want %d", pos, pattern, i, matchPos, expected[i])
			}
		}
	}
}

// TestFindUnicode tests Find with Unicode characters
func TestFindUnicode(t *testing.T) {
	gb := NewGapBuffer(20)
	text := "Hällö Wörld Hällö"
	
	for _, r := range text {
		gb.Insert(r)
	}

	tests := []struct {
		name     string
		pattern  string
		expected []int
	}{
		{"unicode pattern", "Hällö", []int{0, 12}},
		{"umlaut", "ö", []int{4, 7, 16}},
		{"mixed", "ö W", []int{4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gb.Find(tt.pattern)
			if len(result) != len(tt.expected) {
				t.Errorf("Find(%q) returned %d matches, want %d", tt.pattern, len(result), len(tt.expected))
				return
			}
			for i, pos := range result {
				if pos != tt.expected[i] {
					t.Errorf("Find(%q) match %d at position %d, want %d", tt.pattern, i, pos, tt.expected[i])
				}
			}
		})
	}
}

// TestComputeLPSArray tests the LPS array computation for KMP algorithm
func TestComputeLPSArray(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []int
	}{
		{"simple pattern", "ABABCABAB", []int{0, 0, 1, 2, 0, 1, 2, 3, 4}},
		{"no prefix-suffix", "ABCDEF", []int{0, 0, 0, 0, 0, 0}},
		{"all same", "AAAA", []int{0, 1, 2, 3}},
		{"empty pattern", "", []int{}},
		{"single character", "A", []int{0}},
		{"repeating pattern", "ABAB", []int{0, 0, 1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := []rune(tt.pattern)
			result := computeLPSArray(pattern)
			
			if len(result) != len(tt.expected) {
				t.Errorf("computeLPSArray(%q) length = %d, want %d", tt.pattern, len(result), len(tt.expected))
				return
			}
			
			for i, val := range result {
				if val != tt.expected[i] {
					t.Errorf("computeLPSArray(%q)[%d] = %d, want %d", tt.pattern, i, val, tt.expected[i])
				}
			}
		})
	}
}

// TestComplexOperations tests complex sequences of operations
func TestComplexOperations(t *testing.T) {
	gb := NewGapBuffer(5)

	// Build text by inserting at different positions
	gb.Insert('h')
	gb.Insert('e')
	gb.Insert('l')
	gb.Insert('l')
	gb.Insert('o')
	
	if gb.String() != "hello" {
		t.Errorf("After building 'hello', got %q", gb.String())
	}

	// Insert in the middle
	gb.Move(2)
	gb.Insert('X')
	if gb.String() != "heXllo" {
		t.Errorf("After inserting X at position 2, got %q", gb.String())
	}

	// Delete from middle
	gb.Move(2)
	gb.Delete()
	if gb.String() != "hello" {
		t.Errorf("After deleting X, got %q", gb.String())
	}

	// Insert at beginning
	gb.Move(0)
	gb.Insert('H')
	if gb.String() != "Hhello" {
		t.Errorf("After inserting H at beginning, got %q", gb.String())
	}

	// Delete first character
	gb.Move(0)
	gb.Delete()
	if gb.String() != "hello" {
		t.Errorf("After deleting first character, got %q", gb.String())
	}
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	// Test with minimal buffer that needs frequent resizing
	gb := NewGapBuffer(2)
	
	// Insert many characters to force multiple resizes
	text := "This is a longer text that will force multiple buffer resizes"
	for _, r := range text {
		gb.Insert(r)
	}
	
	if gb.String() != text {
		t.Errorf("After multiple resizes, got %q, want %q", gb.String(), text)
	}

	// Test operations on empty buffer
	emptyGB := NewGapBuffer(10)
	
	// Delete on empty buffer should not crash
	emptyGB.Delete()
	if emptyGB.Length() != 0 {
		t.Errorf("Delete on empty buffer changed length to %d", emptyGB.Length())
	}

	// Find on empty buffer
	matches := emptyGB.Find("test")
	if len(matches) != 0 {
		t.Errorf("Find on empty buffer returned %d matches, want 0", len(matches))
	}

	// Iteration on empty buffer
	var runes []rune
	for r := range emptyGB.Runes(0) {
		runes = append(runes, r)
	}
	if len(runes) != 0 {
		t.Errorf("Iteration on empty buffer returned %d runes, want 0", len(runes))
	}
}

// BenchmarkInsert benchmarks insertion operations
func BenchmarkInsert(b *testing.B) {
	gb := NewGapBuffer(1000)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		gb.Insert('a')
	}
}

// BenchmarkMove benchmarks cursor movement
func BenchmarkMove(b *testing.B) {
	gb := NewGapBuffer(1000)
	
	// Fill buffer with some data
	for i := 0; i < 500; i++ {
		gb.Insert('a')
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pos := i % gb.Length()
		gb.Move(pos)
	}
}

// BenchmarkFind benchmarks the Find operation
func BenchmarkFind(b *testing.B) {
	gb := NewGapBuffer(1000)
	
	// Create text with multiple occurrences of pattern
	text := strings.Repeat("hello world ", 100)
	for _, r := range text {
		gb.Insert(r)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gb.Find("hello")
	}
}

// BenchmarkString benchmarks String conversion
func BenchmarkString(b *testing.B) {
	gb := NewGapBuffer(1000)
	
	// Fill buffer with data
	for i := 0; i < 500; i++ {
		gb.Insert('a')
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gb.String()
	}
}

// BenchmarkNewGapBufferFromString benchmarks creation from string
func BenchmarkNewGapBufferFromString(b *testing.B) {
	text := strings.Repeat("hello world ", 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewGapBufferFromString(text, 100)
	}
}

// BenchmarkNewGapBufferFromStringLarge benchmarks creation from large string
func BenchmarkNewGapBufferFromStringLarge(b *testing.B) {
	text := strings.Repeat("This is a longer text with more content for benchmarking purposes. ", 1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewGapBufferFromString(text, 1000)
	}
}