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

// Apply applies a theme style to the component.
func (d *Digits) Apply(theme *Theme) {
	theme.Apply(d, d.Selector("digits"))
}

// Render draws the digits widget using the renderer.
func (d *Digits) Render(r *Renderer) {
	// Let base component render background and borders
	d.Component.Render(r)

	// Get the content area for drawing
	x, y, _, h := d.Content()
	if h < 3 {
		return // Not enough vertical space
	}

	// Draw each character in the text
	cx := x
	for _, ch := range d.Text {
		if pattern, ok := digits[ch]; ok {
			for i, row := range pattern {
				r.Text(cx, y+i, row, 0)
			}
			cx += utf8.RuneCountInString(pattern[0])
		}
	}
}

// Set sets the digit text in a generic way.
func (d *Digits) Set(value any) bool {
	if text, ok := value.(string); ok {
		d.SetText(text)
		return true
	} else {
		return false
	}
}

// SetText updates the displayed text and triggers a refresh.
func (d *Digits) SetText(text string) {
	d.Text = text
	d.Refresh()
}
