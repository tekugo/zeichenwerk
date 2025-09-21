// Package render-digits.go implements specialized rendering for the Digits widget.
//
// This file provides ASCII art-style rendering for displaying large, decorative
// digits and letters using Unicode box-drawing characters. The Digits widget is
// typically used for displaying prominent numeric information like counters,
// clocks, scores, or status indicators where visual impact is important.
//
// # Character Set
//
// The implementation supports:
//   - Digits 0-9: Complete numeric character set
//   - Letters A-F: Hexadecimal digit support
//   - Special characters: space, comma, period, colon, hash
//   - Each character is rendered as a 3-line, variable-width design
//
// # Design Philosophy
//
// The character designs prioritize:
//   - Visual clarity and readability at a distance
//   - Consistent visual weight across all characters
//   - Proper proportions for digital display aesthetics
//   - Unicode box-drawing characters for clean terminal rendering

package zeichenwerk

// digits contains the ASCII art patterns for rendering large-format characters.
// Each character is represented as a slice of strings, with each string representing
// one horizontal line of the character's visual representation.
//
// # Character Design Specifications
//
// All characters follow consistent design rules:
//   - Height: Exactly 3 lines for uniform vertical alignment
//   - Width: Variable width optimized for each character's visual needs
//   - Style: Unicode box-drawing characters for clean, professional appearance
//   - Spacing: Consistent internal spacing for optical balance
//
// # Supported Character Set
//
// The character map includes:
//   - Digits: 0, 1, 2, 3, 4, 5, 6, 7, 8, 9
//   - Hex Letters: A, B, C, D, E, F (for hexadecimal display)
//   - Punctuation: space, comma (,), period (.), colon (:)
//   - Special: hash (#) for placeholder or decoration
//
// # Unicode Box-Drawing Characters
//
// The designs use Unicode box-drawing characters including:
//   - Corners: ╭ ╮ ╰ ╯ ┌ ┐ └ ┘
//   - Lines: ─ │ ╶ ╴ ╷ ╵
//   - Intersections: ├ ┤ ┬ ┴ ┼
//   - Special: ● ○ for dots and decoration
//
// # Visual Examples
//
// Sample character rendering:
//   '0': ╭──╮  '1':  ╶╮   '2': ╭──╮
//        │  │        │         ╭──╯
//        ╰──╯       ╶┴╴        └──╴
var digits = map[rune][]string{
	'0': {
		"╭──╮",
		"│  │",
		"╰──╯",
	},
	'1': {
		" ╶╮ ",
		"  │ ",
		" ╶┴╴ ",
	},
	'2': {
		"╭──╮",
		"╭──╯",
		"└──╴",
	},
	'3': {
		"╶──╮",
		" ──┤",
		"╶──╯",
	},
	'4': {
		"╷  ╷",
		"╰──┤",
		"   ╵",
	},
	'5': {
		"┌──╴",
		"╰──╮",
		"╰──┘",
	},
	'6': {
		"╭──╮",
		"├──╮",
		"╰──╯",
	},
	'7': {
		"╶──┐",
		"   │",
		"   ╵",
	},
	'8': {
		"╭──╮",
		"├──┤",
		"╰──╯",
	},
	'9': {
		"╭──╮",
		"╰──┤",
		"╰──╯",
	},
	'A': {
		"╭──╮",
		"├──┤",
		"╵  ╵",
	},
	'B': {
		"┌─╮ ",
		"├─┴╮",
		"└──╯",
	},
	'C': {
		"╭──╮",
		"│   ",
		"╰──╯",
	},
	'D': {
		"┌──╮",
		"│  │",
		"└──╯",
	},
	'E': {
		"┌──╴",
		"├── ",
		"└──╴",
	},
	'F': {
		"┌──╴",
		"├── ",
		"╵   ",
	},
	'#': {
		"    ",
		" ┼┼ ",
		" ┼┼ ",
	},
	',': {
		"   ",
		"   ",
		" , ",
	},
	'.': {
		"   ",
		"   ",
		" ● ",
	},
	':': {
		" ○ ",
		"   ",
		" ○ ",
	},
	' ': {
		"   ",
		"   ",
		"   ",
	},
}

// renderDigits renders the Digits widget content using large ASCII art-style characters.
// This method converts each character in the widget's text into a multi-line visual
// representation using Unicode box-drawing characters for enhanced visual impact.
//
// # Rendering Process
//
// The method processes each character in the widget text:
//  1. Looks up the character in the digits map
//  2. Skips unsupported characters silently
//  3. Renders each line of the character pattern at the appropriate position
//  4. Advances the horizontal position by the character's width
//  5. Continues until all characters are processed
//
// # Character Positioning
//
// Characters are positioned using these rules:
//   - Horizontal: Characters are placed side-by-side with no spacing
//   - Vertical: All characters align to the same baseline (y coordinate)
//   - Width: Each character's width is determined by its pattern's first line
//   - Height: All characters use exactly 3 lines (consistent across all patterns)
//
// # Unsupported Character Handling
//
// When encountering characters not in the digits map:
//   - The character is silently skipped (no error or placeholder)
//   - No horizontal space is allocated for the missing character
//   - Rendering continues with the next character
//   - This allows graceful degradation for mixed content
//
// # Visual Layout
//
// The rendering creates a continuous string of large characters:
//   - No automatic spacing between characters
//   - Characters flow left-to-right in reading order
//   - Vertical alignment maintained across all characters
//   - Total width determined by sum of individual character widths
//
// # Performance Considerations
//
// The method is optimized for:
//   - Direct character lookup using map access (O(1))
//   - Minimal string operations during rendering
//   - Efficient horizontal positioning calculation
//   - No unnecessary memory allocations for character data
//
// # Use Cases
//
// This rendering method is ideal for:
//   - Digital clocks and timers
//   - Numeric counters and scoreboards
//   - Status displays and indicators
//   - Hexadecimal value displays
//   - Decorative text elements
//
// # Example Output
//
// For widget.Text = "123":
//   ╶╮ ╭──╮╶──╮
//    │ ╭──╯ ──┤
//   ╶┴╴└──╴╶──╯
//
// Parameters:
//   - widget: The Digits widget containing the text to render
//   - x, y: Top-left coordinates of the content area
//   - w, h: Available width and height (h should be at least 3 for proper display)
func (r *Renderer) renderDigits(widget *Digits, x, y, w, h int) {
	cx := x
	for _, ch := range widget.Text {
		m, found := digits[ch]
		if !found {
			continue
		}
		for i, row := range m {
			r.text(cx, y+i, row, 0)
		}
		cx += len([]rune(m[0]))
	}
}
