package format

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

// MarkdownFormat handles GFM pipe tables.
type MarkdownFormat struct{}

func (f *MarkdownFormat) Name() string         { return "markdown" }
func (f *MarkdownFormat) Extensions() []string { return []string{"md", "markdown"} }

func (f *MarkdownFormat) Detect(data []byte) bool {
	s := string(data)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "|---|") || strings.Contains(line, "|:--") ||
			strings.Contains(line, "|---") {
			return true
		}
	}
	return false
}

func (f *MarkdownFormat) Parse(data []byte, _ ParseOpts) (*MutableTable, error) {
	lines := splitLines(string(data))

	// find first pipe table block
	var tableLines []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "|") || isPipeSep(trimmed) {
			tableLines = append(tableLines, trimmed)
		} else if len(tableLines) > 0 {
			break
		}
	}

	t := NewMutableTable()
	if len(tableLines) < 2 {
		return t, nil
	}

	headers := parsePipeRow(tableLines[0])
	aligns := parseSepRow(tableLines[1])

	var rows [][]string
	for _, l := range tableLines[2:] {
		if isPipeSep(l) {
			continue
		}
		rows = append(rows, parsePipeRow(l))
	}

	t.Load(headers, rows)
	// pad alignments to column count
	for len(aligns) < len(headers) {
		aligns = append(aligns, AlignLeft)
	}
	t.LoadAlignments(aligns)
	return t, nil
}

func (f *MarkdownFormat) Serialize(t *MutableTable, opts SerialOpts) ([]byte, error) {
	cols := t.Columns()
	if len(cols) == 0 {
		return nil, nil
	}

	pretty := opts.Pretty
	// pretty is default for markdown
	if !pretty {
		pretty = true
	}

	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = c.Width
		if widths[i] < 3 {
			widths[i] = 3 // min for separator row
		}
	}

	var buf bytes.Buffer

	// header row
	buf.WriteString("|")
	for i, c := range cols {
		cell := c.Header
		if pretty {
			cell = padCell(cell, widths[i], AlignLeft)
		}
		buf.WriteString(" " + cell + " |")
	}
	buf.WriteString("\n")

	// separator row
	buf.WriteString("|")
	for i, c := range cols {
		w := widths[i]
		align := Alignment(c.Alignment)
		sep := sepCell(w, align)
		buf.WriteString(" " + sep + " |")
		_ = i
	}
	buf.WriteString("\n")

	// data rows
	for row := 0; row < t.Length(); row++ {
		buf.WriteString("|")
		for i, c := range cols {
			cell := t.Str(row, i)
			if pretty {
				cell = padCell(cell, widths[i], Alignment(c.Alignment))
			}
			buf.WriteString(" " + cell + " |")
		}
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// parsePipeRow splits a pipe-delimited row into trimmed cells.
func parsePipeRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

// parseSepRow extracts alignment from a separator row like |:---|:---:|---:|
func parseSepRow(line string) []Alignment {
	cells := parsePipeRow(line)
	aligns := make([]Alignment, len(cells))
	for i, c := range cells {
		c = strings.TrimSpace(c)
		left := strings.HasPrefix(c, ":")
		right := strings.HasSuffix(c, ":")
		switch {
		case left && right:
			aligns[i] = AlignCenter
		case right:
			aligns[i] = AlignRight
		default:
			aligns[i] = AlignLeft
		}
	}
	return aligns
}

// isPipeSep returns true for separator rows like |---|:---:|---:|
func isPipeSep(line string) bool {
	line = strings.TrimSpace(line)
	if !strings.Contains(line, "|") {
		return false
	}
	inner := strings.Trim(line, "|")
	for _, cell := range strings.Split(inner, "|") {
		cell = strings.TrimSpace(cell)
		clean := strings.Trim(cell, ":-")
		if clean != "" {
			return false
		}
	}
	return true
}

// padCell pads s to width runes according to alignment.
func padCell(s string, width int, align Alignment) string {
	runes := []rune(s)
	n := len(runes)
	if n >= width {
		return s
	}
	pad := width - n
	switch align {
	case AlignRight:
		return strings.Repeat(" ", pad) + s
	case AlignCenter:
		l := pad / 2
		return strings.Repeat(" ", l) + s + strings.Repeat(" ", pad-l)
	default:
		return s + strings.Repeat(" ", pad)
	}
}

// sepCell builds a markdown separator cell of the given width and alignment.
func sepCell(width int, align Alignment) string {
	if width < 3 {
		width = 3
	}
	inner := strings.Repeat("-", width)
	switch align {
	case AlignCenter:
		// ensure at least 3 dashes between colons
		if width < 3 {
			inner = "---"
		} else {
			inner = strings.Repeat("-", width)
		}
		return ":" + inner + ":"
	case AlignRight:
		return strings.Repeat("-", width) + ":"
	default:
		return strings.Repeat("-", width)
	}
}

func splitLines(s string) []string {
	return strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
}

// runeWidth returns the rune count of s (for alignment padding).
func runeWidth(s string) int { return utf8.RuneCountInString(s) }
