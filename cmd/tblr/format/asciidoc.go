package format

import (
	"bytes"
	"strings"
)

// AsciiDocFormat handles AsciiDoc |=== block tables.
type AsciiDocFormat struct{}

func (f *AsciiDocFormat) Name() string         { return "asciidoc" }
func (f *AsciiDocFormat) Extensions() []string { return []string{"adoc", "asciidoc"} }

func (f *AsciiDocFormat) Detect(data []byte) bool {
	return strings.Contains(string(data), "|===")
}

func (f *AsciiDocFormat) Parse(data []byte, _ ParseOpts) (*MutableTable, error) {
	lines := splitLines(string(data))
	t := NewMutableTable()

	// find |=== block
	start := -1
	end := -1
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed == "|===" {
			if start == -1 {
				start = i
			} else {
				end = i
				break
			}
		}
	}
	if start == -1 {
		return t, nil
	}
	if end == -1 {
		end = len(lines)
	}

	block := lines[start+1 : end]

	// parse cols attribute for alignment
	var aligns []Alignment
	for i := start - 1; i >= 0 && i >= start-5; i-- {
		l := strings.TrimSpace(lines[i])
		if strings.HasPrefix(l, "[") || strings.HasPrefix(l, "cols=") || strings.Contains(l, "cols=") {
			aligns = parseCols(l)
			break
		}
	}

	// collect cell tokens (each cell starts with |)
	var tokens []string
	for _, l := range block {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		// cells can be on one line: | a | b | c
		// or each cell on its own line: | cell content
		if strings.HasPrefix(l, "|") {
			// split on | — but first char is |, so trim it
			parts := strings.Split(l[1:], "|")
			for _, p := range parts {
				tokens = append(tokens, strings.TrimSpace(p))
			}
		}
	}

	if len(tokens) == 0 {
		return t, nil
	}

	// first row is header
	ncols := guessColCount(tokens, aligns)
	if ncols == 0 {
		return t, nil
	}

	headers := make([]string, ncols)
	for i := 0; i < ncols && i < len(tokens); i++ {
		headers[i] = tokens[i]
	}

	var rows [][]string
	for offset := ncols; offset < len(tokens); offset += ncols {
		row := make([]string, ncols)
		for j := 0; j < ncols && offset+j < len(tokens); j++ {
			row[j] = tokens[offset+j]
		}
		rows = append(rows, row)
	}

	t.Load(headers, rows)
	if len(aligns) > 0 {
		t.LoadAlignments(aligns)
	}
	return t, nil
}

func (f *AsciiDocFormat) Serialize(t *MutableTable, opts SerialOpts) ([]byte, error) {
	var buf bytes.Buffer

	cols := t.Columns()
	ncols := len(cols)

	// cols attribute
	buf.WriteString("[cols=\"")
	for i, c := range cols {
		if i > 0 {
			buf.WriteString(",")
		}
		switch Alignment(c.Alignment) {
		case AlignCenter:
			buf.WriteString("^")
		case AlignRight:
			buf.WriteString(">")
		default:
			buf.WriteString("<")
		}
	}
	buf.WriteString("\"]\n")

	buf.WriteString("|===\n")

	// header row
	for i := 0; i < ncols; i++ {
		h := cols[i].Header
		if opts.Pretty {
			h = padCell(h, cols[i].Width, AlignLeft)
		}
		buf.WriteString("| " + h + " ")
	}
	buf.WriteString("\n\n")

	// data rows
	for row := 0; row < t.Length(); row++ {
		for col := 0; col < ncols; col++ {
			cell := t.Str(row, col)
			if opts.Pretty {
				cell = padCell(cell, cols[col].Width, Alignment(cols[col].Alignment))
			}
			buf.WriteString("| " + cell + " ")
		}
		buf.WriteString("\n")
	}

	buf.WriteString("|===\n")
	return buf.Bytes(), nil
}

// parseCols parses a cols attribute like [cols="<,^,>"] or cols="<,^,>"
func parseCols(line string) []Alignment {
	// extract between quotes
	start := strings.Index(line, "\"")
	end := strings.LastIndex(line, "\"")
	if start == -1 || end <= start {
		return nil
	}
	spec := line[start+1 : end]
	parts := strings.Split(spec, ",")
	aligns := make([]Alignment, len(parts))
	for i, p := range parts {
		p = strings.TrimSpace(p)
		// may have width prefix like "1<" or just "<"
		ch := lastRune(p)
		switch ch {
		case '^':
			aligns[i] = AlignCenter
		case '>':
			aligns[i] = AlignRight
		default:
			aligns[i] = AlignLeft
		}
	}
	return aligns
}

func lastRune(s string) rune {
	if s == "" {
		return 0
	}
	runes := []rune(s)
	return runes[len(runes)-1]
}

func guessColCount(tokens []string, aligns []Alignment) int {
	if len(aligns) > 0 {
		return len(aligns)
	}
	// heuristic: assume first row is the header; look for the "empty line" separator
	// or just use all tokens as 1 row if unknown
	if len(tokens) > 0 {
		// try small factor
		for n := 1; n <= len(tokens); n++ {
			if len(tokens)%n == 0 {
				return n
			}
		}
	}
	return len(tokens)
}
