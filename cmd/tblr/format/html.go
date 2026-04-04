package format

import (
	"bytes"
	"regexp"
	"strings"
)

// HTMLFormat handles <table> HTML tables.
type HTMLFormat struct{}

func (f *HTMLFormat) Name() string         { return "html" }
func (f *HTMLFormat) Extensions() []string { return []string{"html", "htm"} }

func (f *HTMLFormat) Detect(data []byte) bool {
	return strings.Contains(strings.ToLower(string(data)), "<table")
}

var (
	reTag      = regexp.MustCompile(`(?i)<(/?)(\w+)[^>]*>`)
	reEntities = strings.NewReplacer(
		"&amp;", "&", "&lt;", "<", "&gt;", ">",
		"&nbsp;", " ", "&quot;", `"`, "&#39;", "'",
	)
)

func (f *HTMLFormat) Parse(data []byte, _ ParseOpts) (*MutableTable, error) {
	s := string(data)
	t := NewMutableTable()

	// extract <table>...</table>
	lower := strings.ToLower(s)
	ts := strings.Index(lower, "<table")
	if ts == -1 {
		return t, nil
	}
	te := strings.Index(lower[ts:], "</table>")
	if te == -1 {
		te = len(s) - ts
	} else {
		te += 8
	}
	tableHTML := s[ts : ts+te]

	rows := extractRows(tableHTML)
	if len(rows) == 0 {
		return t, nil
	}

	// first row (th or td) is the header
	headers := rows[0]
	var data2 [][]string
	if len(rows) > 1 {
		data2 = rows[1:]
	}
	t.Load(headers, data2)
	return t, nil
}

func (f *HTMLFormat) Serialize(t *MutableTable, opts SerialOpts) ([]byte, error) {
	var buf bytes.Buffer
	cols := t.Columns()
	ncols := len(cols)

	buf.WriteString("<table>\n")

	// thead
	buf.WriteString("  <thead>\n    <tr>\n")
	for i := 0; i < ncols; i++ {
		h := htmlEscape(cols[i].Header)
		buf.WriteString("      <th>" + h + "</th>\n")
	}
	buf.WriteString("    </tr>\n  </thead>\n")

	// tbody
	buf.WriteString("  <tbody>\n")
	for row := 0; row < t.Length(); row++ {
		buf.WriteString("    <tr>\n")
		for col := 0; col < ncols; col++ {
			cell := htmlEscape(t.Str(row, col))
			buf.WriteString("      <td>" + cell + "</td>\n")
		}
		buf.WriteString("    </tr>\n")
	}
	buf.WriteString("  </tbody>\n</table>\n")
	return buf.Bytes(), nil
}

// extractRows returns all rows as slices of cell strings.
func extractRows(html string) [][]string {
	var rows [][]string
	lower := strings.ToLower(html)
	i := 0
	for i < len(html) {
		trStart := strings.Index(lower[i:], "<tr")
		if trStart == -1 {
			break
		}
		trStart += i
		// skip past >
		gtIdx := strings.Index(html[trStart:], ">")
		if gtIdx == -1 {
			break
		}
		cellStart := trStart + gtIdx + 1

		trEnd := strings.Index(lower[cellStart:], "</tr>")
		if trEnd == -1 {
			trEnd = len(html) - cellStart
		}
		rowHTML := html[cellStart : cellStart+trEnd]

		cells := extractCells(rowHTML)
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
		i = cellStart + trEnd + 5
	}
	return rows
}

// extractCells pulls text content from <td> and <th> elements.
func extractCells(html string) []string {
	var cells []string
	lower := strings.ToLower(html)
	i := 0
	for i < len(html) {
		// find next td or th
		tdIdx := strings.Index(lower[i:], "<td")
		thIdx := strings.Index(lower[i:], "<th")
		start := -1
		tag := ""
		if tdIdx != -1 && (thIdx == -1 || tdIdx < thIdx) {
			start = i + tdIdx
			tag = "td"
		} else if thIdx != -1 {
			start = i + thIdx
			tag = "th"
		}
		if start == -1 {
			break
		}
		// skip to end of opening tag
		gtIdx := strings.Index(html[start:], ">")
		if gtIdx == -1 {
			break
		}
		contentStart := start + gtIdx + 1

		endTag := "</" + tag + ">"
		endIdx := strings.Index(lower[contentStart:], endTag)
		var content string
		if endIdx == -1 {
			content = html[contentStart:]
			i = len(html)
		} else {
			content = html[contentStart : contentStart+endIdx]
			i = contentStart + endIdx + len(endTag)
		}
		// strip inner tags
		content = reTag.ReplaceAllString(content, "")
		content = reEntities.Replace(content)
		content = strings.TrimSpace(content)
		cells = append(cells, content)
	}
	return cells
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
