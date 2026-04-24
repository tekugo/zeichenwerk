package core

import (
	"fmt"
	"iter"
)

// GapBuffer implements a gap buffer, an efficient data structure for storing
// and manipulating text when edits tend to cluster around a single moving
// cursor position. It is a classic building block of text editors.
//
// The buffer is a contiguous rune slice that contains a "gap" — a region of
// unused slots — positioned at the current cursor. Insertions consume cells
// at the left edge of the gap, forward deletions consume cells at the right
// edge, and cursor movement slides the gap by copying the runes that cross
// it. This makes insert and forward-delete O(1) at the cursor while cursor
// moves cost O(n) in the distance travelled.
//
// Invariants maintained by all operations:
//   - 0 <= start <= end <= len(buffer)
//   - the logical text is buffer[0:start] concatenated with buffer[end:]
//   - the gap (buffer[start:end]) holds unused slots whose contents are
//     undefined and must not be observed by callers
//   - the logical cursor position is always equal to start
//
// The zero value is not usable; construct instances with NewGapBuffer or
// NewGapBufferFromString. GapBuffer is not safe for concurrent use.
type GapBuffer struct {
	buffer     []rune // backing storage; gap cells hold undefined values
	start, end int    // gap occupies buffer[start:end] (end exclusive)
}

// NewGapBuffer creates an empty gap buffer with the specified initial
// backing capacity. The buffer will automatically grow (doubling each time)
// when the gap is exhausted, so the capacity is only an optimisation hint.
//
// Parameters:
//   - capacity: Initial size of the backing slice. Values below 2 are
//     silently raised to 2 to guarantee a minimal working gap.
//
// Returns:
//   - *GapBuffer: A new gap buffer whose cursor is positioned at offset 0
//     and whose entire backing slice is initially gap.
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

// NewGapBufferFromString creates a gap buffer pre-populated with text. The
// text is decoded into runes so multibyte characters occupy a single slot
// each, and the cursor (and therefore the gap) is placed immediately after
// the last rune — the typical starting state for an editor opening a file.
//
// Parameters:
//   - text:        Initial content of the buffer.
//   - gapCapacity: Desired initial gap size. Values below 1 are silently
//     raised to 1 so at least one edit is possible before a resize.
//
// Returns:
//   - *GapBuffer: A new gap buffer of length len([]rune(text)) with its
//     cursor positioned at the end of the text.
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

// ---- Public Methods -------------------------------------------------------

// Delete performs a forward delete: it removes the single rune immediately
// to the right of the cursor by expanding the gap. The cursor itself does
// not move. When the cursor is at the end of the buffer there is nothing to
// delete and the call is a no-op.
//
// Time complexity is O(1). To delete the rune to the left of the cursor
// (backspace), first Move the cursor one step left and then call Delete.
func (gb *GapBuffer) Delete() {
	if gb.end < len(gb.buffer) {
		gb.end++
	}
}

// Find returns the logical offsets of every non-overlapping occurrence of
// substr in the buffer, matched against the gap-free rune sequence (so
// matches can straddle the gap without being missed). The search uses the
// Knuth–Morris–Pratt algorithm for O(n + m) time complexity, where n is the
// buffer length and m is the pattern length.
//
// Parameters:
//   - substr: The substring to search for. An empty string yields no matches.
//
// Returns:
//   - []int: Rune-based starting positions of each match, or an empty (but
//     non-nil) slice when substr is empty, longer than the buffer, or does
//     not appear.
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

// Insert writes a rune at the current cursor position and advances the
// cursor past it. The logical buffer length increases by one. If the gap
// has been exhausted, the backing slice is doubled before the insertion so
// the call appears to always succeed.
//
// The amortised time complexity is O(1) — individual calls may be O(n) when
// they trigger a resize, but the total cost of any sequence of k insertions
// is O(k).
//
// Parameters:
//   - r: The rune to insert at the cursor position.
func (gb *GapBuffer) Insert(r rune) {
	if gb.start == gb.end {
		gb.resize()
	}
	gb.buffer[gb.start] = r
	gb.start++
}

// Length returns the number of runes currently stored in the buffer,
// excluding the gap. Because the two fragments (before the gap and after
// it) are tracked directly, this is an O(1) query.
//
// Returns:
//   - int: The logical length of the buffer in runes.
func (gb *GapBuffer) Length() int {
	return gb.start + (len(gb.buffer) - gb.end)
}

// Move repositions the cursor (and therefore the gap) to the given logical
// offset by shifting the runes that cross the gap. A position equal to
// Length() places the cursor after the last rune, which is the natural
// "append" position.
//
// The cost is O(d) where d is the distance between the current and target
// positions; moving to the current position is free.
//
// Parameters:
//   - pos: Target cursor offset in the range [0, Length()].
//
// Panics:
//   - If pos is outside the valid range [0, Length()].
func (gb *GapBuffer) Move(pos int) {
	if pos < 0 || pos > gb.Length() {
		panic(fmt.Sprintf("GapBuffer.Move: position %d out of range [0, %d]", pos, gb.Length()))
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

// Runes returns a range-compatible iterator that yields the buffer's runes
// in logical order, starting at the given offset and skipping over the
// gap. Typical usage:
//
//	for r := range gb.Runes(0) {
//	    // ...
//	}
//
// Parameters:
//   - start: First logical offset to emit. If start is negative or greater
//     than or equal to Length(), the iterator yields no values.
//
// Returns:
//   - iter.Seq[rune]: A rune sequence that terminates cleanly when the
//     caller breaks out of the range loop — no goroutine is spawned, so
//     early termination cannot leak resources.
//
// The buffer must not be mutated while iteration is in progress; the
// iterator reads directly from the underlying slice without a snapshot,
// so concurrent edits produce undefined behaviour.
func (gb *GapBuffer) Runes(start int) iter.Seq[rune] {
	return func(yield func(rune) bool) {
		if start < 0 || start >= gb.Length() {
			return
		}
		if start < gb.start {
			// Start in the left part of the buffer
			for i := start; i < gb.start; i++ {
				if !yield(gb.buffer[i]) {
					return
				}
			}
			// Continue with the right part of the buffer
			for i := gb.end; i < len(gb.buffer); i++ {
				if !yield(gb.buffer[i]) {
					return
				}
			}
		} else {
			// Start in the right part of the buffer
			offset := gb.end - gb.start
			for i := start + offset; i < len(gb.buffer); i++ {
				if !yield(gb.buffer[i]) {
					return
				}
			}
		}
	}
}

// String materialises the full logical content of the buffer as a Go
// string, concatenating the runes before the gap with those after it. The
// gap itself is not included. Time complexity is O(n) in the buffer
// length because all runes are copied.
//
// String also makes GapBuffer satisfy fmt.Stringer, so it works directly
// with fmt.Print and related helpers.
//
// Returns:
//   - string: The complete gap-free buffer content.
func (gb *GapBuffer) String() string {
	out := make([]rune, 0, gb.Length())
	out = append(out, gb.buffer[:gb.start]...)
	out = append(out, gb.buffer[gb.end:]...)
	return string(out)
}

// ---- Internal Methods -----------------------------------------------------

// computeLPSArray builds the "longest proper prefix which is also a suffix"
// table used by the Knuth–Morris–Pratt algorithm. Given a pattern P of
// length m, the returned slice lps has lps[i] equal to the length of the
// longest proper prefix of P[0..i] that is also a suffix of P[0..i].
// This table lets KMP skip comparisons after a mismatch without
// re-examining characters of the input. It runs in O(m) time.
//
// Parameters:
//   - pattern: The pattern to preprocess. An empty pattern yields an empty
//     table.
//
// Returns:
//   - []int: The LPS table, one entry per pattern rune.
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

// resize doubles the backing slice when the gap has been exhausted and
// copies the pre-gap and post-gap fragments into the new slice at their
// original logical offsets. The newly added capacity is absorbed into the
// gap, so the cursor position and logical content are both preserved.
// Because the capacity grows geometrically, amortised insert cost remains
// O(1) across arbitrarily long sequences of insertions.
func (gb *GapBuffer) resize() {
	newBuf := make([]rune, len(gb.buffer)*2)
	newEnd := len(newBuf) - (len(gb.buffer) - gb.end)

	copy(newBuf[:gb.start], gb.buffer[:gb.start])
	copy(newBuf[newEnd:], gb.buffer[gb.end:])

	gb.buffer = newBuf
	gb.end = newEnd
}
