package zeichenwerk

// GapBuffer implements a gap buffer data structure, which is an efficient way
// to store and manipulate text data with frequent insertions and deletions at
// single cursor position. This data structure is commonly used in text
// editors.
//
// The gap buffer maintains a contiguous array with a "gap" (empty space) that
// moves to the cursor position. Insertions happen at the gap, and deletions
// expand the gap. This allows O(1) insertions and deletions at the cursor
// position, with O(n) cost only when moving the cursor to a different
// position.
type GapBuffer struct {
	buffer     []rune // buffer consists of printable runes
	start, end int    // gap start and end index (gap is from start to end-1)
}

// NewGapBuffer creates a new gap buffer with the specified initial capacity.
// The gap buffer will automatically resize when needed.
//
// Parameters:
//   - capacity: Initial capacity of the buffer. Must be at least 2.
//
// Returns:
//   - *GapBuffer: A new gap buffer instance with the gap positioned at the
//     beginning.
func NewGapBuffer(capacity int) *GapBuffer {
	if capacity < 2 {
		capacity = 2
	}
	return &GapBuffer{
		buffer: make([]rune, capacity),
		start:  0,
		end:    capacity,
	}
}

// NewGapBufferFromString creates a new gap buffer initialized with the given
// string and a specified gap capacity. The gap is positioned at the end of
// the initial text.
//
// Parameters:
//   - text: Initial text to populate the buffer with.
//   - gapCapacity: Size of the gap to maintain. Must be at least 1.
//
// Returns:
//   - *GapBuffer: A new gap buffer instance with the text loaded and gap at
//     the end.
func NewGapBufferFromString(text string, gapCapacity int) *GapBuffer {
	if gapCapacity < 1 {
		gapCapacity = 1
	}

	runes := []rune(text)
	textLen := len(runes)
	totalCapacity := textLen + gapCapacity

	gb := &GapBuffer{
		buffer: make([]rune, totalCapacity),
		start:  textLen,
		end:    totalCapacity,
	}

	// Copy the text to the beginning of the buffer
	copy(gb.buffer[:textLen], runes)

	return gb
}

// resize doubles the capacity of the gap buffer when the gap becomes empty.
// This is an internal method called automatically when more space is needed.
// The gap position is preserved during resizing.
func (gb *GapBuffer) resize() {
	newBuf := make([]rune, len(gb.buffer)*2)
	newEnd := len(newBuf) - (len(gb.buffer) - gb.end)

	copy(newBuf[:gb.start], gb.buffer[:gb.start])
	copy(newBuf[newEnd:], gb.buffer[gb.end:])

	gb.buffer = newBuf
	gb.end = newEnd
}

// Move repositions the gap (cursor) to the specified position in the buffer.
// This operation has O(n) time complexity in the worst case, where n is the
// distance between the current and target positions.
//
// Parameters:
//   - pos: Target position for the cursor (0-based index). Must be within
//     [0, Length()].
//
// Panics:
//   - If pos is outside the valid range [0, Length()].
func (gb *GapBuffer) Move(pos int) {
	if pos < 0 || pos > gb.Length() {
		panic("Cursor außerhalb des gültigen Bereichs")
	}

	if pos < gb.start {
		// Move gap left: copy characters from left of gap to right of gap
		n := gb.start - pos
		copy(gb.buffer[gb.end-n:gb.end], gb.buffer[pos:gb.start])
		gb.end -= n
		gb.start = pos
	} else if pos > gb.start {
		// Move gap right: copy characters from right of gap to left of gap
		n := pos - gb.start
		copy(gb.buffer[gb.start:gb.start+n], gb.buffer[gb.end:gb.end+n])
		gb.start += n
		gb.end += n
	}
	// If pos == gb.start, gap is already at the correct position
}

// Insert adds a rune at the current cursor position.
// This operation has O(1) time complexity. The buffer will automatically
// resize if needed.
//
// Parameters:
//   - r: The rune to insert at the current cursor position.
func (gb *GapBuffer) Insert(r rune) {
	if gb.start == gb.end {
		gb.resize()
	}
	gb.buffer[gb.start] = r
	gb.start++
}

// Delete removes the character immediately after the cursor position.
// This operation has O(1) time complexity. If the cursor is at the end
// of the buffer, this operation has no effect.
func (gb *GapBuffer) Delete() {
	if gb.end < len(gb.buffer) {
		gb.end++
	}
}

// Length returns the number of characters currently stored in the buffer,
// excluding the gap. This operation has O(1) time complexity.
//
// Returns:
//   - int: The number of characters in the buffer.
func (gb *GapBuffer) Length() int {
	return gb.start + (len(gb.buffer) - gb.end)
}

// String returns the entire buffer content as a string, excluding the gap.
// This operation has O(n) time complexity where n is the buffer length.
//
// Returns:
//   - string: The complete buffer content as a string.
func (gb *GapBuffer) String() string {
	out := make([]rune, 0, gb.Length())
	out = append(out, gb.buffer[:gb.start]...)
	out = append(out, gb.buffer[gb.end:]...)
	return string(out)
}

// Runes returns a channel that iterates over all runes in the buffer starting
// from the specified position. The iteration skips the gap and provides a
// sequential view of the buffer content.
//
// Parameters:
//   - start: Starting position for iteration (0-based index).
//
// Returns:
//   - <-chan rune: A channel that yields runes from the buffer sequentially.
//     The channel is closed when iteration is complete.
//
// Note: If start is outside the valid range, the channel will be closed
// immediately.
func (gb *GapBuffer) Runes(start int) <-chan rune {
	ch := make(chan rune)
	go func() {
		defer close(ch)
		if start < 0 || start >= gb.Length() {
			return
		}
		if start < gb.start {
			// Start in the left part of the buffer
			for i := start; i < gb.start; i++ {
				ch <- gb.buffer[i]
			}
			// Continue with the right part of the buffer
			for i := gb.end; i < len(gb.buffer); i++ {
				ch <- gb.buffer[i]
			}
		} else {
			// Start in the right part of the buffer
			offset := gb.end - gb.start
			for i := start + offset; i < len(gb.buffer); i++ {
				ch <- gb.buffer[i]
			}
		}
	}()
	return ch
}

// computeLPSArray computes the Longest Proper Prefix which is also Suffix
// (LPS) array for the KMP (Knuth-Morris-Pratt) string matching algorithm.
// This is used internally by the Find method.
//
// Parameters:
//   - pattern: The pattern to compute LPS array for.
//
// Returns:
//   - []int: LPS array where lps[i] contains the length of the longest proper
//     prefix of pattern[0..i] which is also a suffix of pattern[0..i].
func computeLPSArray(pattern []rune) []int {
	lps := make([]int, len(pattern))
	length := 0
	i := 1
	for i < len(pattern) {
		if pattern[i] == pattern[length] {
			length++
			lps[i] = length
			i++
		} else {
			if length != 0 {
				length = lps[length-1]
			} else {
				lps[i] = 0
				i++
			}
		}
	}
	return lps
}

// Find searches for all occurrences of a substring in the buffer using the
// Knuth-Morris-Pratt (KMP) algorithm, which has O(n + m) time complexity
// where n is the buffer length and m is the pattern length.
//
// Parameters:
//   - substr: The substring to search for in the buffer.
//
// Returns:
//   - []int: A slice containing the starting positions of all matches.
//     Returns an empty slice if no matches are found or if substr is empty.
func (gb *GapBuffer) Find(substr string) []int {
	result := []int{}
	if substr == "" {
		return result
	}
	needle := []rune(substr)
	needleLen := len(needle)
	bufLen := gb.Length()
	if needleLen > bufLen {
		return result
	}

	// Helper function to get character at logical position i, accounting for the gap
	charAt := func(i int) rune {
		if i < gb.start {
			return gb.buffer[i]
		}
		return gb.buffer[i+(gb.end-gb.start)]
	}

	lps := computeLPSArray(needle)
	i, j := 0, 0
	for i < bufLen {
		if charAt(i) == needle[j] {
			i++
			j++
		}
		if j == needleLen {
			result = append(result, i-j)
			j = lps[j-1]
		} else if i < bufLen && charAt(i) != needle[j] {
			if j != 0 {
				j = lps[j-1]
			} else {
				i++
			}
		}
	}
	return result
}
