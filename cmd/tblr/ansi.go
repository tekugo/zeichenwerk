package main

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tekugo/zeichenwerk/cmd/tblr/format"
)

// ANSIOpts controls ANSI rendering.
type ANSIOpts struct {
	Border    string // thin, double, rounded, thick, none
	Theme     string // auto, dark, light, 16
	Zebra     bool
	Width     int
	NoNumeric bool
}

// DefaultANSIOpts returns sensible defaults.
func DefaultANSIOpts() ANSIOpts {
	return ANSIOpts{Border: "thin", Theme: "auto", Zebra: true, Width: 0}
}

// borderChars holds box-drawing characters for a border style.
type borderChars struct {
	tl, tr, bl, br string // corners
	h, v           string // horizontal, vertical
	tT, bT, lT, rT string // T-junctions
	cross          string
	hr             string // header separator (same as h)
}

var borders = map[string]borderChars{
	"thin": {
		tl: "┌", tr: "┐", bl: "└", br: "┘",
		h: "─", v: "│",
		tT: "┬", bT: "┴", lT: "├", rT: "┤",
		cross: "┼", hr: "─",
	},
	"double": {
		tl: "╔", tr: "╗", bl: "╚", br: "╝",
		h: "═", v: "║",
		tT: "╦", bT: "╩", lT: "╠", rT: "╣",
		cross: "╬", hr: "═",
	},
	"rounded": {
		tl: "╭", tr: "╮", bl: "╰", br: "╯",
		h: "─", v: "│",
		tT: "┬", bT: "┴", lT: "├", rT: "┤",
		cross: "┼", hr: "─",
	},
	"thick": {
		tl: "┏", tr: "┓", bl: "┗", br: "┛",
		h: "━", v: "┃",
		tT: "┳", bT: "┻", lT: "┣", rT: "┫",
		cross: "╋", hr: "━",
	},
	"none": {
		tl: " ", tr: " ", bl: " ", br: " ",
		h: " ", v: " ",
		tT: " ", bT: " ", lT: " ", rT: " ",
		cross: " ", hr: " ",
	},
}

// ansiTheme holds ANSI colour codes.
type ansiTheme struct {
	reset   string
	bold    string
	headerFg string
	headerBg string
	borderFg string
	zebra    string // bg for odd rows
	normal   string // bg for even rows
	numFg    string // numeric cell foreground
}

func resolveTheme(name string) ansiTheme {
	switch name {
	case "light":
		return ansiTheme{
			reset: "\033[0m", bold: "\033[1m",
			headerFg: "\033[30m", headerBg: "\033[47m",
			borderFg: "\033[90m",
			zebra:    "\033[48;5;254m", normal: "",
			numFg:    "\033[34m",
		}
	case "16":
		return ansiTheme{
			reset: "\033[0m", bold: "\033[1m",
			headerFg: "\033[97m", headerBg: "\033[44m",
			borderFg: "\033[90m",
			zebra:    "\033[40m", normal: "",
			numFg:    "\033[36m",
		}
	default: // auto / dark
		return ansiTheme{
			reset: "\033[0m", bold: "\033[1m",
			headerFg: "\033[97m", headerBg: "\033[48;5;237m",
			borderFg: "\033[38;5;240m",
			zebra:    "\033[48;5;236m", normal: "",
			numFg:    "\033[38;5;81m",
		}
	}
}

// RenderANSI writes the table to w with ANSI escape codes.
func RenderANSI(w io.Writer, t *format.MutableTable, opts ANSIOpts) error {
	bc, ok := borders[opts.Border]
	if !ok {
		bc = borders["thin"]
	}
	th := resolveTheme(opts.Theme)

	cols := t.Columns()
	ncols := len(cols)
	if ncols == 0 {
		return nil
	}

	// compute widths respecting opts.Width
	widths := make([]int, ncols)
	totalW := 0
	for i, c := range cols {
		widths[i] = c.Width
		if widths[i] < utf8.RuneCountInString(c.Header) {
			widths[i] = utf8.RuneCountInString(c.Header)
		}
		totalW += widths[i]
	}
	totalW += ncols + 1 // borders between and on sides

	if opts.Width > 0 && totalW > opts.Width {
		// proportionally shrink, min 3 per column
		excess := totalW - opts.Width
		for excess > 0 && ncols > 0 {
			for i := range widths {
				if widths[i] > 3 {
					widths[i]--
					excess--
					if excess == 0 {
						break
					}
				}
			}
		}
	}

	// top border
	fmt.Fprint(w, th.borderFg+bc.tl)
	for i, ww := range widths {
		fmt.Fprint(w, strings.Repeat(bc.h, ww+2))
		if i < ncols-1 {
			fmt.Fprint(w, bc.tT)
		}
	}
	fmt.Fprint(w, bc.tr+th.reset+"\n")

	// header row
	fmt.Fprint(w, th.borderFg+bc.v+th.reset)
	for i, c := range cols {
		cell := truncate(c.Header, widths[i])
		cell = padCell(cell, widths[i], format.AlignLeft)
		fmt.Fprint(w, " "+th.headerBg+th.headerFg+th.bold+cell+th.reset+" ")
		fmt.Fprint(w, th.borderFg+bc.v+th.reset)
	}
	fmt.Fprint(w, "\n")

	// header separator
	fmt.Fprint(w, th.borderFg+bc.lT)
	for i, ww := range widths {
		fmt.Fprint(w, strings.Repeat(bc.hr, ww+2))
		if i < ncols-1 {
			fmt.Fprint(w, bc.cross)
		}
	}
	fmt.Fprint(w, bc.rT+th.reset+"\n")

	// data rows
	for row := 0; row < t.Length(); row++ {
		rowBg := ""
		if opts.Zebra && row%2 == 1 {
			rowBg = th.zebra
		} else {
			rowBg = th.normal
		}

		fmt.Fprint(w, th.borderFg+bc.v+th.reset)
		for col, c := range cols {
			cell := t.Str(row, col)
			cell = truncate(cell, widths[col])
			cell = padCell(cell, widths[col], format.Alignment(c.Alignment))

			fg := ""
			if isNumeric(t.Str(row, col)) {
				fg = th.numFg
			}

			fmt.Fprint(w, " "+rowBg+fg+cell+th.reset+" ")
			fmt.Fprint(w, th.borderFg+bc.v+th.reset)
		}
		fmt.Fprint(w, "\n")
	}

	// bottom border
	fmt.Fprint(w, th.borderFg+bc.bl)
	for i, ww := range widths {
		fmt.Fprint(w, strings.Repeat(bc.h, ww+2))
		if i < ncols-1 {
			fmt.Fprint(w, bc.bT)
		}
	}
	fmt.Fprint(w, bc.br+th.reset+"\n")

	return nil
}

func truncate(s string, maxW int) string {
	runes := []rune(s)
	if len(runes) <= maxW {
		return s
	}
	if maxW <= 1 {
		return "…"
	}
	return string(runes[:maxW-1]) + "…"
}

func padCell(s string, width int, align format.Alignment) string {
	runes := []rune(s)
	n := len(runes)
	if n >= width {
		return s
	}
	pad := width - n
	switch align {
	case format.AlignRight:
		return strings.Repeat(" ", pad) + s
	case format.AlignCenter:
		l := pad / 2
		return strings.Repeat(" ", l) + s + strings.Repeat(" ", pad-l)
	default:
		return s + strings.Repeat(" ", pad)
	}
}

func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
