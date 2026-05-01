package widgets

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// Inline style flags for Span.Style.
const (
	Italic        = 1
	Bold          = 2
	Underline     = 4
	Strikethrough = 8
	Code          = 16
)

// Block is a parsed unit of styled text.
// Type is one of "p", "h1"–"h4", "ul", "ol", "code", "bq", "hr", or "table".
// Index holds the 1-based item number for "ol" blocks.
// Depth holds the nesting level (0 = top) for "ul" and "ol" blocks.
// Task and Done describe task-list items within "ul" blocks.
// Rows and Aligns hold the parsed cell data for "table" blocks.
type Block struct {
	Type    string
	Text    string
	Content []Span
	Index   int        // 1-based item number for "ol"
	Depth   int        // nesting depth for "ul" and "ol"
	Task    bool       // true if this is a task-list item (- [ ] / - [x])
	Done    bool       // true if the task is checked
	Rows    [][]string // table: Rows[0] = header cells, Rows[1:] = body rows
	Aligns  []string   // table: per-column alignment ("left", "center", "right")
}

// Span describes a contiguous run of text within a Block with a single inline
// style. Start and End are byte offsets into Block.Text.
type Span struct {
	Start  int
	End    int
	Length int // byte length (== rune count for ASCII)
	Style  int // bitmask of style flags
}

// Parse parses the inline markup in Block.Text and populates Block.Content.
// Supported markers: *italic*, **bold**, __underline__, ~~strikethrough~~, `code`.
func (b *Block) Parse() {
	b.Content = make([]Span, 0)
	n := len(b.Text)
	start := 0
	style := 0
	skip := false

	for i, r := range b.Text {
		if skip {
			skip = false
			continue
		}

		if style&Code != 0 {
			if r == '`' {
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style &^= Code
				start = i + 1
			}
			continue
		}

		switch r {
		case '*':
			if i+1 < n && b.Text[i+1] == '*' {
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Bold
				start = i + 2
				skip = true
			} else {
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Italic
				start = i + 1
			}
		case '_':
			if i+1 < n && b.Text[i+1] == '_' {
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Underline
				start = i + 2
				skip = true
			} else {
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Italic
				start = i + 1
			}
		case '~':
			if i+1 < n && b.Text[i+1] == '~' {
				if i > start {
					b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
				}
				style ^= Strikethrough
				start = i + 2
				skip = true
			}
		case '`':
			if i > start {
				b.Content = append(b.Content, Span{Start: start, End: i, Length: i - start, Style: style})
			}
			style |= Code
			start = i + 1
		}
	}

	if start < n {
		b.Content = append(b.Content, Span{Start: start, End: n, Length: n - start, Style: style})
	}
}

// ---- segment / line --------------------------------------------------------

// segment is a styled text fragment for one terminal column run.
// style holds inline style flags (Bold, Italic, etc.).
type segment struct {
	text  string
	style int
}

// line is one rendered terminal row, made up of segments.
type line []segment

// renderedLine is a laid-out terminal row together with its block kind
// ("h1"–"h4", "p", "ul", "ol", "pre", "code", "bq", "hr", "table", or "" for blank lines).
// The kind is used at render time to look up the appropriate theme style.
type renderedLine struct {
	kind string
	segs line
}

// segWidth returns the total rune width of a line.
func segWidth(l line) int {
	n := 0
	for _, s := range l {
		n += utf8.RuneCountInString(s.text)
	}
	return n
}

// styleFont converts a style bitmask to a space-separated font string accepted
// by Renderer.Set.
func styleFont(style int) string {
	var parts []string
	if style&Bold != 0 {
		parts = append(parts, "bold")
	}
	if style&Italic != 0 {
		parts = append(parts, "italic")
	}
	if style&Underline != 0 {
		parts = append(parts, "underline")
	}
	if style&Strikethrough != 0 {
		parts = append(parts, "strikethrough")
	}
	if style&Code != 0 {
		parts = append(parts, "code")
	}
	return strings.Join(parts, " ")
}

// wrapSpans word-wraps the spans of a block into lines of at most width rune
// columns. Whitespace is used as the break opportunity; a single word longer
// than width is placed on its own line without truncation.
func wrapSpans(spans []Span, text string, width int) []line {
	if width <= 0 {
		return nil
	}

	type word struct {
		t     string
		style int
	}
	var words []word
	for _, span := range spans {
		if span.Start > len(text) || span.End > len(text) {
			continue
		}
		for _, w := range strings.Fields(text[span.Start:span.End]) {
			words = append(words, word{w, span.Style})
		}
	}

	if len(words) == 0 {
		return []line{{}}
	}

	var lines []line
	cur := line{}
	cx := 0

	isPunct := func(s string) bool {
		if s == "" {
			return false
		}
		r, _ := utf8.DecodeRuneInString(s)
		switch r {
		case ',', '.', ';', ':', '!', '?', ')', ']':
			return true
		}
		return false
	}

	for _, w := range words {
		ww := utf8.RuneCountInString(w.t)
		switch {
		case cx == 0:
			cur = append(cur, segment{w.t, w.style})
			cx = ww
		case isPunct(w.t):
			// Attach punctuation directly without a space.
			cur = append(cur, segment{w.t, w.style})
			cx += ww
		case cx+1+ww > width:
			lines = append(lines, cur)
			cur = line{segment{w.t, w.style}}
			cx = ww
		default:
			cur = append(cur, segment{" ", 0}) // style=0 so space never inherits inline decoration
			cur = append(cur, segment{w.t, w.style})
			cx += 1 + ww
		}
	}
	if len(cur) > 0 {
		lines = append(lines, cur)
	}
	return lines
}

// ---- Styled widget ---------------------------------------------------------

// Styled is a read-only widget that renders a subset of Markdown with word
// wrapping. Supported block types: paragraphs, # h1–h4 headings,
// - unordered lists (with task-list support), 1. ordered lists,
// ``` code blocks, > blockquotes, --- horizontal rules, and GFM-style
// pipe tables. Supported inline styles: *italic*, **bold**,
// __underline__, ~~strikethrough~~, `code`.
type Styled struct {
	Component
	text   string
	blocks []Block
	lines  []renderedLine
	scroll int
	lastW  int
}

// NewStyled creates a Styled widget and parses the markup in text.
func NewStyled(id, class, text string) *Styled {
	s := &Styled{Component: Component{id: id, class: class}}
	s.SetFlag(FlagFocusable, true)
	s.SetText(text)
	OnKey(s, s.handleKey)
	return s
}

// SetText replaces the widget's content and re-parses it.
func (s *Styled) SetText(text string) {
	s.text = text
	s.Parse()
	s.lastW = 0 // force re-layout on next render
	Redraw(s)
}

// Apply applies a theme's styles to the component, including per-element
// styles for h1–h4, p, ul, ol, pre, code, bq, hr, and table.
func (s *Styled) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("styled"))
	for _, part := range []string{"h1", "h2", "h3", "h4", "p", "ul", "ol", "pre", "code", "bq", "hr", "table"} {
		theme.Apply(s, "styled/"+part)
	}
	s.lastW = 0 // styles may affect rendering, force re-layout
}

// Refresh triggers a redraw of the widget.
func (s *Styled) Refresh() {
	Redraw(s)
}

// ScrollBy scrolls the content by delta rows (negative = up).
func (s *Styled) ScrollBy(delta int) {
	s.scroll += delta
	if s.scroll < 0 {
		s.scroll = 0
	}
	Redraw(s)
}

// isHR reports whether line is a Markdown horizontal rule (3+ dashes, stars,
// or underscores, optionally separated by spaces, with nothing else on the line).
func isHR(s string) bool {
	s = strings.ReplaceAll(s, " ", "")
	if len(s) < 3 {
		return false
	}
	ch := s[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != ch {
			return false
		}
	}
	return true
}

// Parse re-parses s.text into blocks. Called automatically by SetText.
func (s *Styled) Parse() {
	s.blocks = nil
	var cur *Block
	// olIndex tracks per-depth ordered list counters (depth → next index).
	olIndex := map[int]int{}

	flush := func() {
		if cur != nil {
			switch cur.Type {
			case "table":
				parseTableBlock(cur)
			case "code":
				// verbatim — no inline parsing
			default:
				if cur.Content == nil {
					cur.Parse()
				}
			}
			s.blocks = append(s.blocks, *cur)
			cur = nil
		}
	}

	for _, raw := range strings.Split(s.text, "\n") {
		trimmed := strings.TrimSpace(raw)

		// Code block fence
		if strings.HasPrefix(trimmed, "```") {
			if cur != nil && cur.Type == "code" {
				flush()
			} else {
				flush()
				cur = &Block{Type: "code"}
			}
			olIndex = map[int]int{}
			continue
		}
		if cur != nil && cur.Type == "code" {
			if cur.Text != "" {
				cur.Text += "\n"
			}
			cur.Text += raw
			continue
		}

		// Headings — check #### before ### before ## before #
		if strings.HasPrefix(raw, "#### ") {
			flush()
			b := Block{Type: "h4", Text: strings.TrimPrefix(raw, "#### ")}
			b.Parse()
			s.blocks = append(s.blocks, b)
			olIndex = map[int]int{}
			continue
		}
		if strings.HasPrefix(raw, "### ") {
			flush()
			b := Block{Type: "h3", Text: strings.TrimPrefix(raw, "### ")}
			b.Parse()
			s.blocks = append(s.blocks, b)
			olIndex = map[int]int{}
			continue
		}
		if strings.HasPrefix(raw, "## ") {
			flush()
			b := Block{Type: "h2", Text: strings.TrimPrefix(raw, "## ")}
			b.Parse()
			s.blocks = append(s.blocks, b)
			olIndex = map[int]int{}
			continue
		}
		if strings.HasPrefix(raw, "# ") {
			flush()
			b := Block{Type: "h1", Text: strings.TrimPrefix(raw, "# ")}
			b.Parse()
			s.blocks = append(s.blocks, b)
			olIndex = map[int]int{}
			continue
		}

		// Horizontal rule — must be checked before paragraph accumulation
		if isHR(trimmed) {
			flush()
			s.blocks = append(s.blocks, Block{Type: "hr"})
			olIndex = map[int]int{}
			continue
		}

		// Blockquote — accumulate consecutive "> " lines into one block
		if strings.HasPrefix(trimmed, "> ") {
			text := strings.TrimPrefix(trimmed, "> ")
			if cur != nil && cur.Type == "bq" {
				cur.Text += " " + text
			} else {
				flush()
				cur = &Block{Type: "bq", Text: text}
			}
			continue
		}

		// Table row — accumulate consecutive pipe-delimited rows into one block
		if strings.HasPrefix(trimmed, "|") {
			if cur != nil && cur.Type == "table" {
				cur.Text += "\n" + trimmed
			} else {
				flush()
				cur = &Block{Type: "table", Text: trimmed}
				olIndex = map[int]int{}
			}
			continue
		}

		// List items — measure indent depth (2 spaces per level)
		leading := len(raw) - len(strings.TrimLeft(raw, " \t"))
		depth := leading / 2

		// Unordered list (supports task list markers - [ ] and - [x])
		if strings.HasPrefix(trimmed, "- ") {
			flush()
			text := strings.TrimPrefix(trimmed, "- ")
			b := Block{Type: "ul", Depth: depth}
			if strings.HasPrefix(text, "[ ] ") {
				b.Task, b.Done, b.Text = true, false, strings.TrimPrefix(text, "[ ] ")
			} else if strings.HasPrefix(text, "[x] ") || strings.HasPrefix(text, "[X] ") {
				b.Task, b.Done, b.Text = true, true, text[4:]
			} else {
				b.Text = text
			}
			b.Parse()
			s.blocks = append(s.blocks, b)
			olIndex = map[int]int{}
			continue
		}

		// Ordered list: "N. text" (depth-aware counter)
		if text, n, ok := parseOLPrefix(trimmed); ok {
			flush()
			if n == 1 {
				olIndex[depth] = 0
			}
			olIndex[depth]++
			b := Block{Type: "ol", Text: text, Index: olIndex[depth], Depth: depth}
			b.Parse()
			s.blocks = append(s.blocks, b)
			continue
		}

		// Blank line — flush current paragraph
		if trimmed == "" {
			flush()
			olIndex = map[int]int{}
			continue
		}

		// Paragraph — accumulate continuation lines
		if cur == nil {
			cur = &Block{Type: "p", Text: trimmed}
		} else if cur.Type == "p" {
			cur.Text += " " + trimmed
		} else {
			flush()
			cur = &Block{Type: "p", Text: trimmed}
		}
	}
	flush()
}

// parseOLPrefix matches "N. text" at the start of line and returns the text
// and the numeric prefix. ok is false when the line is not an ordered list item.
func parseOLPrefix(line string) (text string, n int, ok bool) {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == 0 || i >= len(line)-1 || line[i] != '.' || line[i+1] != ' ' {
		return "", 0, false
	}
	num, err := strconv.Atoi(line[:i])
	if err != nil {
		return "", 0, false
	}
	return line[i+2:], num, true
}

// parseTableBlock parses the raw accumulated pipe-row text in b.Text into
// b.Rows (header + body) and b.Aligns (per-column alignment).
// The expected format is standard GFM:
//
//	| Col A | Col B |
//	|-------|:-----:|
//	| val   | val   |
func parseTableBlock(b *Block) {
	rawLines := strings.Split(b.Text, "\n")
	if len(rawLines) < 2 {
		return
	}

	splitCells := func(s string) []string {
		s = strings.TrimSpace(s)
		s = strings.Trim(s, "|")
		parts := strings.Split(s, "|")
		cells := make([]string, len(parts))
		for i, p := range parts {
			cells[i] = strings.TrimSpace(p)
		}
		return cells
	}

	header := splitCells(rawLines[0])
	cols := len(header)
	if cols == 0 {
		return
	}

	// Parse separator row for column alignment.
	b.Aligns = make([]string, cols)
	for i, cell := range splitCells(rawLines[1]) {
		if i >= cols {
			break
		}
		cell = strings.TrimSpace(cell)
		left := strings.HasPrefix(cell, ":")
		right := strings.HasSuffix(cell, ":")
		switch {
		case left && right:
			b.Aligns[i] = "center"
		case right:
			b.Aligns[i] = "right"
		default:
			b.Aligns[i] = "left"
		}
	}

	b.Rows = make([][]string, 0, len(rawLines)-1)
	b.Rows = append(b.Rows, header)

	for _, raw := range rawLines[2:] {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		row := splitCells(raw)
		for len(row) < cols {
			row = append(row, "")
		}
		b.Rows = append(b.Rows, row[:cols])
	}
}

// ---- Layout ----------------------------------------------------------------

func (s *Styled) rl(kind string, segs line) renderedLine {
	return renderedLine{kind: kind, segs: segs}
}

func (s *Styled) layout(w int) {
	s.lines = nil
	s.lastW = w

	for i, b := range s.blocks {
		if i > 0 {
			prev := s.blocks[i-1]
			sameList := (b.Type == "ul" && prev.Type == "ul") ||
				(b.Type == "ol" && prev.Type == "ol") ||
				(b.Type == "bq" && prev.Type == "bq")
			if !sameList {
				s.lines = append(s.lines, renderedLine{})
			}
		}
		switch b.Type {
		case "h1":
			s.layoutH1(b, w)
		case "h2":
			s.layoutH2(b, w)
		case "h3":
			s.layoutH3(b, w)
		case "h4":
			s.layoutH4(b, w)
		case "p":
			s.layoutP(b, w)
		case "ul":
			s.layoutUL(b, w)
		case "ol":
			s.layoutOL(b, w)
		case "code":
			s.layoutCode(b, w)
		case "bq":
			s.layoutBQ(b, w)
		case "hr":
			s.layoutHR(b, w)
		case "table":
			s.layoutTable(b, w)
		}
	}
}

// layoutH1 renders a heading with a tight surrounding box border.
//
//	╔══════════════════╗
//	║ Heading text     ║
//	╚══════════════════╝
func (s *Styled) layoutH1(b Block, w int) {
	wrapped := wrapSpans(b.Content, b.Text, w-4)

	maxW := 0
	for _, wl := range wrapped {
		if cw := segWidth(wl); cw > maxW {
			maxW = cw
		}
	}
	border := maxW + 4
	top := "╔" + strings.Repeat("═", border-2) + "╗"
	bot := "╚" + strings.Repeat("═", border-2) + "╝"

	s.lines = append(s.lines, s.rl("h1", line{segment{top, 0}}))
	for _, wl := range wrapped {
		pad := maxW - segWidth(wl)
		l := line{segment{"║ ", 0}}
		l = append(l, wl...)
		if pad > 0 {
			l = append(l, segment{strings.Repeat(" ", pad), 0})
		}
		l = append(l, segment{" ║", 0})
		s.lines = append(s.lines, s.rl("h1", l))
	}
	s.lines = append(s.lines, s.rl("h1", line{segment{bot, 0}}))
}

// layoutH2 renders a heading with a bottom border rule.
//
//	Heading text
//	────────────
func (s *Styled) layoutH2(b Block, w int) {
	for _, wl := range wrapSpans(b.Content, b.Text, w) {
		s.lines = append(s.lines, s.rl("h2", wl))
	}
	s.lines = append(s.lines, s.rl("h2", line{segment{strings.Repeat("─", w), 0}}))
}

// layoutH3 renders a heading using the h3 theme style (typically bold+underline).
func (s *Styled) layoutH3(b Block, w int) {
	for _, wl := range wrapSpans(b.Content, b.Text, w) {
		s.lines = append(s.lines, s.rl("h3", wl))
	}
}

// layoutH4 renders a heading using the h4 theme style (typically bold).
func (s *Styled) layoutH4(b Block, w int) {
	for _, wl := range wrapSpans(b.Content, b.Text, w) {
		s.lines = append(s.lines, s.rl("h4", wl))
	}
}

func (s *Styled) layoutP(b Block, w int) {
	for _, wl := range wrapSpans(b.Content, b.Text, w) {
		s.lines = append(s.lines, s.rl("p", wl))
	}
}

// ulBullet returns the bullet string for a given nesting depth.
func ulBullet(depth int) string {
	switch depth % 3 {
	case 1:
		return "◦ "
	case 2:
		return "▸ "
	default:
		return "• "
	}
}

// layoutUL renders an unordered list item with a bullet and continuation indent.
// Depth is reflected as extra leading spaces (2 per level).
// Task-list items use ☐/☑ instead of a bullet.
func (s *Styled) layoutUL(b Block, w int) {
	margin := strings.Repeat("  ", b.Depth)
	var bullet string
	if b.Task {
		if b.Done {
			bullet = "☑ "
		} else {
			bullet = "☐ "
		}
	} else {
		bullet = ulBullet(b.Depth)
	}
	prefix := margin + bullet
	indent := utf8.RuneCountInString(prefix)
	cont := margin + strings.Repeat(" ", utf8.RuneCountInString(bullet))
	for i, wl := range wrapSpans(b.Content, b.Text, w-indent) {
		var l line
		if i == 0 {
			l = append(l, segment{prefix, 0})
		} else {
			l = append(l, segment{cont, 0})
		}
		l = append(l, wl...)
		s.lines = append(s.lines, s.rl("ul", l))
	}
}

// layoutOL renders an ordered list item with its number and continuation indent.
// Depth is reflected as extra leading spaces (2 per level).
func (s *Styled) layoutOL(b Block, w int) {
	margin := strings.Repeat("  ", b.Depth)
	num := fmt.Sprintf("%d. ", b.Index)
	prefix := margin + num
	indent := utf8.RuneCountInString(prefix)
	cont := margin + strings.Repeat(" ", utf8.RuneCountInString(num))
	for i, wl := range wrapSpans(b.Content, b.Text, w-indent) {
		var l line
		if i == 0 {
			l = append(l, segment{prefix, 0})
		} else {
			l = append(l, segment{cont, 0})
		}
		l = append(l, wl...)
		s.lines = append(s.lines, s.rl("ol", l))
	}
}

// layoutBQ renders a blockquote with a leading "│ " border.
func (s *Styled) layoutBQ(b Block, w int) {
	const prefix = "│ "
	const indent = 2
	for _, wl := range wrapSpans(b.Content, b.Text, w-indent) {
		l := line{segment{prefix, 0}}
		l = append(l, wl...)
		s.lines = append(s.lines, s.rl("bq", l))
	}
}

// layoutHR renders a horizontal rule. It is one character shorter than the
// full content width to leave a visible gap before the scrollbar.
func (s *Styled) layoutHR(_ Block, w int) {
	s.lines = append(s.lines, s.rl("hr", line{segment{strings.Repeat("─", w-1), 0}}))
}

// layoutCode renders a code block verbatim, one source line per terminal row.
func (s *Styled) layoutCode(b Block, w int) {
	for _, codeLine := range strings.Split(b.Text, "\n") {
		runes := []rune(codeLine)
		if len(runes) > w {
			runes = runes[:w]
		}
		s.lines = append(s.lines, s.rl("pre", line{segment{string(runes), 0}}))
	}
}

// layoutTable renders a GFM pipe table: a header row, a single separator
// line, and zero or more body rows. All lines use the "table" theme style.
// Column widths are derived from the widest cell in each column; cells are
// padded to that width and aligned per the separator markers.
//
//	│ Header A │ Header B │
//	├──────────┼──────────┤
//	│ cell     │    cell  │
func (s *Styled) layoutTable(b Block, w int) {
	if len(b.Rows) == 0 {
		return
	}
	cols := len(b.Rows[0])
	if cols == 0 {
		return
	}

	// Compute per-column content widths.
	colW := make([]int, cols)
	for _, row := range b.Rows {
		for c, cell := range row {
			if c >= cols {
				break
			}
			if cw := utf8.RuneCountInString(cell); cw > colW[c] {
				colW[c] = cw
			}
		}
	}

	// clip trims a rendered row to the available width.
	clip := func(s string) string {
		r := []rune(s)
		if len(r) > w {
			return string(r[:w])
		}
		return s
	}

	// makeRow formats one data row into a renderedLine.
	makeRow := func(row []string) renderedLine {
		var sb strings.Builder
		for c := 0; c < cols; c++ {
			cell := ""
			if c < len(row) {
				cell = row[c]
			}
			cw := colW[c]
			if rlen := utf8.RuneCountInString(cell); rlen > cw {
				cell = string([]rune(cell)[:cw])
			}
			pad := cw - utf8.RuneCountInString(cell)
			align := "left"
			if c < len(b.Aligns) {
				align = b.Aligns[c]
			}
			if c == 0 {
				sb.WriteString("│ ")
			} else {
				sb.WriteString(" │ ")
			}
			switch align {
			case "right":
				sb.WriteString(strings.Repeat(" ", pad))
				sb.WriteString(cell)
			case "center":
				lpad := pad / 2
				sb.WriteString(strings.Repeat(" ", lpad))
				sb.WriteString(cell)
				sb.WriteString(strings.Repeat(" ", pad-lpad))
			default: // left
				sb.WriteString(cell)
				sb.WriteString(strings.Repeat(" ", pad))
			}
		}
		sb.WriteString(" │")
		return s.rl("table", line{segment{clip(sb.String()), 0}})
	}

	// makeBorder builds a horizontal border line using the given left, mid, right,
	// and fill characters (e.g. ┌, ┬, ┐, ─ for the top border).
	makeBorder := func(left, mid, right, fill string) renderedLine {
		var sb strings.Builder
		for c := 0; c < cols; c++ {
			if c == 0 {
				sb.WriteString(left)
			} else {
				sb.WriteString(mid)
			}
			sb.WriteString(strings.Repeat(fill, colW[c]+2))
		}
		sb.WriteString(right)
		return s.rl("table", line{segment{clip(sb.String()), 0}})
	}

	// Top border.
	s.lines = append(s.lines, makeBorder("┌", "┬", "┐", "─"))

	// Header row.
	s.lines = append(s.lines, makeRow(b.Rows[0]))

	// Header/body separator.
	s.lines = append(s.lines, makeBorder("├", "┼", "┤", "─"))

	// Body rows.
	for _, row := range b.Rows[1:] {
		s.lines = append(s.lines, makeRow(row))
	}

	// Bottom border.
	s.lines = append(s.lines, makeBorder("└", "┴", "┘", "─"))
}

// ---- Render ----------------------------------------------------------------

func (s *Styled) handleKey(ev *tcell.EventKey) bool {
	_, _, _, h := s.Content()
	switch ev.Key() {
	case tcell.KeyUp:
		s.ScrollBy(-1)
	case tcell.KeyDown:
		s.ScrollBy(1)
	case tcell.KeyPgUp:
		s.ScrollBy(-max(1, h-1))
	case tcell.KeyPgDn:
		s.ScrollBy(max(1, h-1))
	case tcell.KeyHome:
		s.ScrollBy(-len(s.lines))
	case tcell.KeyEnd:
		s.ScrollBy(len(s.lines))
	default:
		return false
	}
	return true
}

// Render draws the pre-laid-out lines, respecting the current scroll offset.
func (s *Styled) Render(r *Renderer) {
	// Content() already accounts for the style's padding/border/margin.
	x, y, w, h := s.Content()
	if w != s.lastW {
		s.layout(w)
	}

	// Clamp scroll.
	maxScroll := len(s.lines) - h
	if maxScroll < 0 {
		maxScroll = 0
	}
	if s.scroll > maxScroll {
		s.scroll = maxScroll
	}

	// Clear content area plus all padding using the widget's base style,
	// so every padding strip shares the same background as the content.
	base := s.Style()
	r.Set(base.Foreground(), base.Background(), "")
	pad := base.Padding()
	r.Fill(x-pad.Left, y-pad.Top, w+pad.Left+pad.Right, h+pad.Top+pad.Bottom, " ")

	for i, rl := range s.lines[s.scroll:] {
		if i >= h {
			break
		}
		st := s.Style(rl.kind)
		fg, bg := st.Foreground(), st.Background()
		if bg == "" {
			bg = base.Background()
		}
		if fg == "" {
			fg = base.Foreground()
		}
		cx := x
		for _, seg := range rl.segs {
			font := st.Font()
			if inline := styleFont(seg.style); inline != "" {
				if font != "" {
					font += " " + inline
				} else {
					font = inline
				}
			}
			r.Set(fg, bg, font)
			r.Text(cx, y+i, seg.text, 0)
			cx += utf8.RuneCountInString(seg.text)
		}
		// Fill the remainder of the line so the block's background extends
		// to the full content width (relevant for styled/pre, styled/code, etc.).
		if rem := x + w - cx; rem > 0 {
			r.Set("", bg, "")
			r.Fill(cx, y+i, rem, 1, " ")
		}
	}

	// Scrollbar — placed at the far right of the raw widget bounds, outside
	// the padding, so it is never indented by the content padding.
	if len(s.lines) > h {
		wx, _, ww, _ := s.Bounds()
		r.Set("", "", "")
		r.ScrollbarV(wx+ww-1, y, h, s.scroll, len(s.lines))
	}
}
