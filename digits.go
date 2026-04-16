package zeichenwerk

import (
	"unicode/utf8"
)

// digits contains the ASCII art patterns for rendering large-format characters.
// Each character is represented as a slice of strings, with each string representing
// one horizontal line of the character's visual representation.
//
// The character map includes digits 0-9, letters A-F, and common symbols.
var digits = map[rune][]string{
	'0': {
		"╭──╮",
		"│  │",
		"╰──╯",
	},
	'1': {
		" ╶╮ ",
		"  │ ",
		" ╶┴╴",
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
	'G': {
		"╭──╮",
		"│ ─┐",
		"╰──╯",
	},
	'H': {
		"╷  ╷",
		"├──┤",
		"╵  ╵",
	},
	'I': {
		"╶┬╴",
		" │ ",
		"╶┴╴",
	},
	'J': {
		"  ╷",
		"  │",
		"╰─╯",
	},
	'K': {
		"╷ ╷ ",
		"├─┴╮",
		"╵  ╵",
	},
	'L': {
		"╷  ",
		"│  ",
		"└──",
	},
	'M': {
		"┌─┬─┐",
		"│ │ │",
		"╵   ╵",
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
		"   ",
		" : ",
		"   ",
	},
	' ': {
		"   ",
		"   ",
		"   ",
	},
}

// Digits displays text using large ASCII art-style characters.
// It renders characters using Unicode box-drawing glyphs for visual impact.
type Digits struct {
	Component
	Text string // Text to render as large digits/characters
}

// ---- Constructor ----------------------------------------------------------

// NewDigits creates a digits widget with the given ID and initial text.
func NewDigits(id, class, text string) *Digits {
	d := &Digits{
		Component: Component{id: id, class: class},
		Text:      text,
	}
	// Compute preferred size: sum of character widths, height = 3
	width := 0
	for _, ch := range text {
		if pattern, ok := digits[ch]; ok {
			width += utf8.RuneCountInString(pattern[0])
		}
	}
	d.hwidth = width
	d.hheight = 3
	return d
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies a theme style to the component.
func (d *Digits) Apply(theme *Theme) {
	theme.Apply(d, d.Selector("digits"))
}

// Render draws the digits widget using the renderer.
func (d *Digits) Render(r *Renderer) {
	// Let base component render background and borders
	d.Component.Render(r)

	// Get the content area for drawing
	x, y, w, h := d.Content()
	if h < 3 {
		return // Not enough vertical space
	}

	// Calculate the starting position depending on alignment
	var cx int
	if d.Flag(FlagRight) {
		cx = x + w
		for _, ch := range d.Text {
			if pattern, ok := digits[ch]; ok {
				cx -= utf8.RuneCountInString(pattern[0])
			}
		}
		cx = max(x, cx)
	} else {
		cx = x
	}

	// Draw each character in the text
	for _, ch := range d.Text {
		if pattern, ok := digits[ch]; ok {
			for i, row := range pattern {
				r.Text(cx, y+i, row, 0)
			}
			cx += utf8.RuneCountInString(pattern[0])
		}
	}
}

// ---- Setter ---------------------------------------------------------------

// Set updates the displayed text and triggers a refresh.
func (d *Digits) Set(value string) {
	d.Text = value
	d.Refresh()
}
