package format

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// TypstFormat handles Typst #table(...) syntax.
type TypstFormat struct{}

func (f *TypstFormat) Name() string         { return "typst" }
func (f *TypstFormat) Extensions() []string { return []string{"typ"} }

func (f *TypstFormat) Detect(data []byte) bool {
	return strings.Contains(string(data), "#table(")
}

func (f *TypstFormat) Parse(data []byte, _ ParseOpts) (*MutableTable, error) {
	s := string(data)
	t := NewMutableTable()

	// find #table( ... )
	start := strings.Index(s, "#table(")
	if start == -1 {
		return t, nil
	}
	inner := extractParens(s[start+7:])

	// parse columns: N argument
	ncols := 0
	colsIdx := strings.Index(inner, "columns:")
	if colsIdx != -1 {
		rest := inner[colsIdx+8:]
		rest = strings.TrimSpace(rest)
		end := strings.IndexAny(rest, ",)")
		if end != -1 {
			rest = rest[:end]
		}
		rest = strings.TrimSpace(rest)
		n, err := strconv.Atoi(rest)
		if err == nil {
			ncols = n
		}
	}

	// extract cell contents from [...]
	var cells []string
	i := 0
	for i < len(inner) {
		bi := strings.Index(inner[i:], "[")
		if bi == -1 {
			break
		}
		bi += i
		cell, skip := extractBracket(inner[bi:])
		cells = append(cells, cell)
		i = bi + skip
	}

	if len(cells) == 0 {
		return t, nil
	}
	if ncols == 0 {
		ncols = len(cells) // single row
	}

	headers := make([]string, ncols)
	for i := 0; i < ncols && i < len(cells); i++ {
		headers[i] = cells[i]
	}

	var rows [][]string
	for offset := ncols; offset < len(cells); offset += ncols {
		row := make([]string, ncols)
		for j := 0; j < ncols && offset+j < len(cells); j++ {
			row[j] = cells[offset+j]
		}
		rows = append(rows, row)
	}

	t.Load(headers, rows)
	return t, nil
}

func (f *TypstFormat) Serialize(t *MutableTable, opts SerialOpts) ([]byte, error) {
	cols := t.Columns()
	ncols := len(cols)
	if ncols == 0 {
		return []byte("#table(columns: 0)\n"), nil
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("#table(\n  columns: %d,\n", ncols))

	// header cells
	buf.WriteString("  // header\n")
	for i, c := range cols {
		h := c.Header
		if opts.Pretty {
			h = padCell(h, c.Width, AlignLeft)
		}
		if i < ncols-1 {
			buf.WriteString(fmt.Sprintf("  [%s],\n", h))
		} else {
			buf.WriteString(fmt.Sprintf("  [%s],\n", h))
		}
	}

	// data rows
	for row := 0; row < t.Length(); row++ {
		for col := 0; col < ncols; col++ {
			cell := t.Str(row, col)
			if opts.Pretty {
				cell = padCell(cell, cols[col].Width, Alignment(cols[col].Alignment))
			}
			buf.WriteString(fmt.Sprintf("  [%s],\n", cell))
		}
	}

	buf.WriteString(")\n")
	return buf.Bytes(), nil
}

// extractParens returns everything inside the outermost parentheses starting at s.
func extractParens(s string) string {
	depth := 1
	for i, ch := range s {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return s[:i]
			}
		}
	}
	return s
}

// extractBracket returns the content of [...] and the number of bytes consumed (incl. brackets).
func extractBracket(s string) (string, int) {
	if len(s) == 0 || s[0] != '[' {
		return "", 0
	}
	depth := 1
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return s[1:i], i + 1
			}
		}
	}
	return s[1:], len(s)
}
