package zeichenwerk

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v3"
)

// Editor is a multi-line text editor widget that provides comprehensive text editing
// capabilities with efficient gap buffer-based line storage. It supports all standard
// text editing operations including cursor movement, text insertion/deletion, line
// operations, and scrolling for large documents.
type Editor struct {
	Component
	content  []*GapBuffer // Editor line contents (one GapBuffer per line)
	line     int          // Current cursor line position (0-based)
	column   int          // Current cursor column position (0-based)
	longest  int          // Longest line visual width for horizontal scrolling
	offsetX  int          // Horizontal scroll offset
	offsetY  int          // Vertical scroll offset
	tab      int          // Tab width
	spaces   bool         // Insert spaces instead of tabs
	numbers  int          // Line numbers width (0 = none)
	indent   bool         // Auto-indent
	disabled bool         // Read-only flag
}

// NewEditor creates a new multi-line text editor widget with the specified ID.
func NewEditor(id string) *Editor {
	e := &Editor{
		Component: Component{id: id, hheight: 10}, // default preferred height
		content:   []*GapBuffer{NewGapBuffer(64)},
		line:      0,
		column:    0,
		longest:   0,
		offsetX:   0,
		offsetY:   0,
		tab:       4,
		spaces:    false,
		numbers:   0,
		indent:    true,
		disabled:  false,
	}
	e.SetFlag("focusable", true)
	OnKey(e, e.handleKey)
	return e
}

// Cursor returns the current cursor position relative to the content area.
// The position accounts for line numbers and scrolling.
func (e *Editor) Cursor() (int, int, string) {
	// Get content area dimensions
	_, _, cw, ch := e.Content()
	if cw <= 0 || ch <= 0 {
		return -1, -1, ""
	}

	// Calculate visual column (tabs expanded)
	visualCol := e.calculateVisualColumn()

	// Account for line numbers
	leftMargin := 0
	if e.numbers > 0 {
		leftMargin = e.numbers + 1 // line numbers + separator
	}

	cx := leftMargin + (visualCol - e.offsetX)
	cy := e.line - e.offsetY

	// Determine visible text area dimensions (similar to render)
	usableW := cw - leftMargin
	needV := len(e.content) > ch
	usableW -= b2i(needV)
	needH := e.longest > usableW
	usableH := ch - b2i(needH)

	if cx < 0 || cx >= usableW || cy < 0 || cy >= usableH {
		return -1, -1, ""
	}

	return cx, cy, "|"
}

// Refresh queues a redraw for the editor.
func (e *Editor) Refresh() {
	Redraw(e)
}

// SetContent replaces all editor content with the provided lines.
func (e *Editor) SetContent(lines []string) {
	e.content = make([]*GapBuffer, len(lines))
	for i, line := range lines {
		e.content[i] = NewGapBufferFromString(line, 32)
	}
	if len(e.content) == 0 {
		e.content = []*GapBuffer{NewGapBuffer(64)}
	}
	e.line = 0
	e.column = 0
	e.offsetX = 0
	e.offsetY = 0
	e.updateLongestLine()
	e.Dispatch(e, "change")
	e.Refresh()
}

// Load sets the editor content from a single string (lines separated by \n).
func (e *Editor) Load(text string) {
	lines := strings.Split(text, "\n")
	e.SetContent(lines)
}

// Lines returns the editor content as a slice of strings.
func (e *Editor) Lines() []string {
	lines := make([]string, len(e.content))
	for i, buf := range e.content {
		lines[i] = buf.String()
	}
	return lines
}

// Text returns the editor content as a single string with newlines.
func (e *Editor) Text() string {
	return strings.Join(e.Lines(), "\n")
}

// SetTabWidth configures tab width.
func (e *Editor) SetTabWidth(width int) {
	if width > 0 {
		e.tab = width
	}
	e.Refresh()
}

// UseSpaces configures whether to insert spaces instead of tabs.
func (e *Editor) UseSpaces(useSpaces bool) {
	e.spaces = useSpaces
}

// ShowLineNumbers enables or disables line numbers.
func (e *Editor) ShowLineNumbers(show bool) {
	if show {
		e.numbers = 3 // default width
	} else {
		e.numbers = 0
	}
	e.Refresh()
}

// SetAutoIndent configures auto-indentation.
func (e *Editor) SetAutoIndent(auto bool) {
	e.indent = auto
}

// SetReadOnly configures read-only mode.
func (e *Editor) SetReadOnly(ro bool) {
	e.disabled = ro
	if ro {
		e.SetFlag("disabled", true)
	} else {
		e.SetFlag("disabled", false)
	}
}

// ---- Movement -------------------------------------------------------------

func (e *Editor) Left() {
	if e.column > 0 {
		e.column--
		e.adjustViewport()
		e.Refresh()
	} else if e.line > 0 {
		e.line--
		e.column = e.content[e.line].Length()
		e.adjustViewport()
		e.Refresh()
	}
}

func (e *Editor) Right() {
	lineLen := e.content[e.line].Length()
	if e.column < lineLen {
		e.column++
		e.adjustViewport()
		e.Refresh()
	} else if e.line < len(e.content)-1 {
		e.line++
		e.column = 0
		e.adjustViewport()
		e.Refresh()
	}
}

func (e *Editor) Up() {
	if e.line > 0 {
		e.line--
		lineLen := e.content[e.line].Length()
		if e.column > lineLen {
			e.column = lineLen
		}
		e.adjustViewport()
		e.Refresh()
	}
}

func (e *Editor) Down() {
	if e.line < len(e.content)-1 {
		e.line++
		lineLen := e.content[e.line].Length()
		if e.column > lineLen {
			e.column = lineLen
		}
		e.adjustViewport()
		e.Refresh()
	}
}

func (e *Editor) Home() {
	e.column = 0
	e.adjustViewport()
	e.Refresh()
}

func (e *Editor) End() {
	e.column = e.content[e.line].Length()
	e.adjustViewport()
	e.Refresh()
}

func (e *Editor) PageUp() {
	_, _, _, h := e.Content()
	target := max(e.line-h, 0)
	e.MoveTo(target, e.column)
	e.Refresh()
}

func (e *Editor) PageDown() {
	_, _, _, h := e.Content()
	target := min(e.line+h, len(e.content)-1)
	e.MoveTo(target, e.column)
	e.Refresh()
}

func (e *Editor) DocumentHome() {
	e.MoveTo(0, 0)
}

func (e *Editor) DocumentEnd() {
	lastLine := len(e.content) - 1
	lastCol := e.content[lastLine].Length()
	e.MoveTo(lastLine, lastCol)
}

// MoveTo moves the cursor to the specified line and column.
func (e *Editor) MoveTo(line, column int) {
	if line < 0 {
		line = 0
	} else if line >= len(e.content) {
		line = len(e.content) - 1
	}
	lineLen := e.content[line].Length()
	if column < 0 {
		column = 0
	} else if column > lineLen {
		column = lineLen
	}
	e.line = line
	e.column = column
	e.adjustViewport()
	e.Refresh()
}

// ---- Editing --------------------------------------------------------------

func (e *Editor) Insert(ch rune) {
	if e.disabled {
		return
	}

	// Handle tab character
	if ch == '\t' {
		if e.spaces {
			e.insertTabAsSpaces()
		} else {
			e.content[e.line].Move(e.column)
			e.content[e.line].Insert(ch)
			e.column++
		}
	} else {
		e.content[e.line].Move(e.column)
		e.content[e.line].Insert(ch)
		e.column++
	}

	e.updateLongestLine()
	e.adjustViewport()
	e.Dispatch(e, "change")
	e.Refresh()
}

func (e *Editor) Delete() {
	if e.disabled {
		return
	}

	if e.column > 0 {
		e.content[e.line].Move(e.column - 1)
		e.content[e.line].Delete()
		e.column--
	} else if e.line > 0 {
		// Join with previous line
		prev := e.line - 1
		e.column = e.content[prev].Length()

		// Append current line to previous
		currentText := e.content[e.line].String()
		for _, r := range currentText {
			e.content[prev].Move(e.content[prev].Length())
			e.content[prev].Insert(r)
		}

		// Remove current line
		e.content = append(e.content[:e.line], e.content[e.line+1:]...)
		e.line = prev
	}

	e.updateLongestLine()
	e.adjustViewport()
	e.Dispatch(e, "change")
	e.Refresh()
}

func (e *Editor) DeleteForward() {
	if e.disabled {
		return
	}

	lineLen := e.content[e.line].Length()
	if e.column < lineLen {
		e.content[e.line].Move(e.column)
		e.content[e.line].Delete()
	} else if e.line < len(e.content)-1 {
		// Join with next line
		nextText := e.content[e.line+1].String()
		for _, r := range nextText {
			e.content[e.line].Move(e.content[e.line].Length())
			e.content[e.line].Insert(r)
		}
		e.content = append(e.content[:e.line+1], e.content[e.line+2:]...)
	}

	e.updateLongestLine()
	e.adjustViewport()
	e.Dispatch(e, "change")
	e.Refresh()
}

func (e *Editor) Enter() {
	if e.disabled {
		return
	}

	// Split current line at cursor
	currentLine := e.content[e.line]
	// Get text after cursor
	rightText := ""
	for r := range currentLine.Runes(e.column) {
		rightText += string(r)
	}

	// Truncate current line at cursor
	currentLine.Move(e.column)
	for currentLine.Length() > e.column {
		currentLine.Delete()
	}

	// Create new line with rightText
	newBuf := NewGapBufferFromString(rightText, 32)

	// Auto-indent if enabled
	if e.indent {
		indent := e.getLineIndent(e.line)
		for i, ch := range []rune(indent) {
			newBuf.Move(i)
			newBuf.Insert(ch)
		}
	}

	// Insert new line after current
	e.content = append(e.content[:e.line+1], append([]*GapBuffer{newBuf}, e.content[e.line+1:]...)...)
	e.line++

	// Position cursor at start of new line
	if e.indent {
		indent := e.getLineIndent(e.line - 1)
		e.column = len([]rune(indent))
	} else {
		e.column = 0
	}

	e.updateLongestLine()
	e.adjustViewport()
	e.Dispatch(e, "change")
	e.Refresh()
}

// insertTabAsSpaces inserts spaces to reach the next tab stop.
func (e *Editor) insertTabAsSpaces() {
	spaces := e.tab - (e.column % e.tab)
	for range spaces {
		e.content[e.line].Insert(' ')
		e.column++
	}
	e.updateLongestLine()
	e.adjustViewport()
	e.Dispatch(e, "change")
	e.Refresh()
}

// getLineIndent returns leading whitespace of the given line.
func (e *Editor) getLineIndent(lineNum int) string {
	if lineNum < 0 || lineNum >= len(e.content) {
		return ""
	}
	line := e.content[lineNum].String()
	indent := ""
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			indent += string(ch)
		} else {
			break
		}
	}
	return indent
}

// ---- Viewport & Layout ----------------------------------------------------

// adjustViewport adjusts scroll offsets to keep the cursor visible.
func (e *Editor) adjustViewport() {
	_, _, cw, ch := e.Content()
	if cw <= 0 || ch <= 0 {
		return
	}

	leftMargin := 0
	if e.numbers > 0 {
		leftMargin = e.numbers + 1
	}
	usableW := cw - leftMargin
	usableH := ch

	// Determine if vertical scrollbar will be used (we need to reserve space)
	needV := len(e.content) > usableH
	usableW -= b2i(needV)

	// Determine horizontal scrollbar
	needH := e.longest > usableW
	usableH -= b2i(needH)

	// Horizontal scrolling
	visualCol := e.calculateVisualColumn()
	if visualCol < e.offsetX {
		e.offsetX = visualCol
	} else if visualCol >= e.offsetX+usableW {
		e.offsetX = visualCol - usableW + 1
	}
	if e.offsetX < 0 {
		e.offsetX = 0
	}
	// Limit offsetX to not scroll past end
	maxOffsetX := max(e.longest-usableW, 0)
	if e.offsetX > maxOffsetX {
		e.offsetX = maxOffsetX
	}

	// Vertical scrolling
	if e.line < e.offsetY {
		e.offsetY = e.line
	} else if e.line >= e.offsetY+usableH {
		e.offsetY = e.line - usableH + 1
	}
	if e.offsetY < 0 {
		e.offsetY = 0
	}
	// Limit offsetY
	maxOffsetY := max(len(e.content)-usableH, 0)
	if e.offsetY > maxOffsetY {
		e.offsetY = maxOffsetY
	}
}

// calculateVisualColumn returns the visual column index of the cursor, expanding tabs.
func (e *Editor) calculateVisualColumn() int {
	if e.line >= len(e.content) {
		return 0
	}
	line := e.content[e.line].String()
	col := 0
	for i := 0; i < e.column && i < len(line); i++ {
		if line[i] == '\t' {
			col = ((col / e.tab) + 1) * e.tab
		} else {
			col++
		}
	}
	return col
}

// updateLongestLine recalculates the longest line's visual width.
func (e *Editor) updateLongestLine() {
	e.longest = 0
	for _, buf := range e.content {
		line := buf.String()
		width := visualWidth(line, e.tab)
		if width > e.longest {
			e.longest = width
		}
	}
}

// ---- Rendering ------------------------------------------------------------

func (e *Editor) Render(r *Renderer) {
	// Base rendering (background, border, etc.)
	e.Component.Render(r)

	x, y, w, h := e.Content()
	if w <= 0 || h <= 0 {
		return
	}

	// Margins and scrollbars
	leftMargin := 0
	if e.numbers > 0 {
		leftMargin = e.numbers + 1
	}
	usableW := w - leftMargin
	usableH := h

	// Determine scrollbars
	needV := len(e.content) > usableH
	needH := e.longest > usableW

	// Adjust usable dimensions for scrollbars
	if needV {
		usableW--
	}
	if needH {
		usableH--
	}

	// Render line numbers
	if e.numbers > 0 {
		e.renderLineNumbers(r, x, y, e.numbers, usableH)
		// Draw separator
		sepStyle := e.Style("separator")
		if sepStyle == nil {
			r.Set("white", "black", "")
		} else {
			r.Set(sepStyle.Foreground(), sepStyle.Background(), sepStyle.Font())
		}
		for i := 0; i < usableH; i++ {
			r.Put(x+e.numbers, y+i, "│")
		}
	}

	// Render text content
	textX := x + leftMargin
	textY := y
	for i := 0; i < usableH; i++ {
		lineIdx := e.offsetY + i
		if lineIdx >= len(e.content) {
			// Clear remaining lines
			r.Fill(textX, textY+i, usableW, 1, " ")
			continue
		}
		line := e.content[lineIdx].String()
		visible := e.getVisibleLineContent(line, e.offsetX, usableW, e.tab)
		// Choose style: current line highlighted if focused
		if e.Flag("focused") && lineIdx == e.line {
			style := e.Style("current-line")
			if style != nil {
				r.Set(style.Foreground(), style.Background(), style.Font())
			} else {
				r.Set(e.Style().Foreground(), e.Style().Background(), e.Style().Font())
			}
		} else {
			style := e.Style()
			r.Set(style.Foreground(), style.Background(), style.Font())
		}
		r.Text(textX, textY+i, visible, usableW)
	}

	// Render scrollbars
	if needV {
		// Vertical scrollbar at the far right of the widget (x + w - 1)
		r.ScrollbarV(x+w-1, y, usableH, e.offsetY, len(e.content))
	}
	if needH {
		// Horizontal scrollbar at the bottom of the widget (y + h - 1)
		r.ScrollbarH(x, y+h-1, w, e.offsetX, e.longest)
	}
}

// renderLineNumbers draws line numbers in the left margin.
func (e *Editor) renderLineNumbers(r *Renderer, x, y, width, height int) {
	// Style for line numbers
	styleNum := e.Style("line-numbers")
	if styleNum != nil {
		r.Set(styleNum.Foreground(), styleNum.Background(), styleNum.Font())
	} else {
		r.Set("gray", "", "")
	}
	for i := 0; i < height; i++ {
		lineIdx := e.offsetY + i
		if lineIdx >= len(e.content) {
			// Empty area: fill with spaces
			r.Fill(x, y+i, width, 1, " ")
			continue
		}
		// 1-based line number
		num := lineIdx + 1
		numStr := strconv.Itoa(num)
		// Right-align within width
		padding := width - len(numStr)
		if padding > 0 {
			r.Text(x, y+i, strings.Repeat(" ", padding)+numStr, width)
		} else {
			r.Text(x, y+i, numStr, width)
		}
		// Highlight current line number
		if lineIdx == e.line {
			curStyle := e.Style("current-line-number")
			if curStyle != nil {
				r.Set(curStyle.Foreground(), curStyle.Background(), curStyle.Font())
				r.Text(x, y+i, strings.Repeat(" ", padding)+numStr, width)
				// Restore line number style
				if styleNum != nil {
					r.Set(styleNum.Foreground(), styleNum.Background(), styleNum.Font())
				} else {
					r.Set("gray", "", "")
				}
			}
		}
	}
}

// getVisibleLineContent returns the portion of the line that is visible within the given width,
// taking into account horizontal scrolling and tab expansion.
func (e *Editor) getVisibleLineContent(line string, offsetX, maxWidth, tabWidth int) string {
	if line == "" || maxWidth <= 0 {
		return ""
	}
	// Expand tabs to spaces
	expanded := expandTabs(line, tabWidth)
	// Apply horizontal offset
	runes := []rune(expanded)
	start := offsetX
	if start >= len(runes) {
		return ""
	}
	end := min(start+maxWidth, len(runes))
	return string(runes[start:end])
}

// expandTabs converts tabs to spaces according to tabWidth.
func expandTabs(s string, tabWidth int) string {
	if !strings.Contains(s, "\t") {
		return s
	}
	var b strings.Builder
	col := 0
	for _, ch := range s {
		if ch == '\t' {
			spaces := tabWidth - (col % tabWidth)
			for i := 0; i < spaces; i++ {
				b.WriteRune(' ')
				col++
			}
		} else {
			b.WriteRune(ch)
			col++
		}
	}
	return b.String()
}

// visualWidth computes the visual column width of a string after tab expansion.
func visualWidth(s string, tabWidth int) int {
	width := 0
	col := 0
	for _, ch := range s {
		if ch == '\t' {
			width = ((col / tabWidth) + 1) * tabWidth
			col = width
		} else {
			width++
			col++
		}
	}
	return width
}

// b2i converts bool to int (1 for true, 0 for false).
func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ---- Event Handling -------------------------------------------------------

func (e *Editor) handleKey(_ Widget, evt *tcell.EventKey) bool {
	if e.disabled && evt.Key() != tcell.KeyLeft && evt.Key() != tcell.KeyRight &&
		evt.Key() != tcell.KeyUp && evt.Key() != tcell.KeyDown &&
		evt.Key() != tcell.KeyHome && evt.Key() != tcell.KeyEnd &&
		evt.Key() != tcell.KeyPgUp && evt.Key() != tcell.KeyPgDn {
		return false
	}

	switch evt.Key() {
	case tcell.KeyLeft:
		e.Left()
		return true
	case tcell.KeyRight:
		e.Right()
		return true
	case tcell.KeyUp:
		e.Up()
		return true
	case tcell.KeyDown:
		e.Down()
		return true
	case tcell.KeyHome:
		e.Home()
		return true
	case tcell.KeyEnd:
		e.End()
		return true
	case tcell.KeyPgUp:
		e.PageUp()
		return true
	case tcell.KeyPgDn:
		e.PageDown()
		return true
	case tcell.KeyCtrlA:
		e.DocumentHome()
		return true
	case tcell.KeyCtrlE:
		e.DocumentEnd()
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		e.Delete()
		return true
	case tcell.KeyDelete:
		e.DeleteForward()
		return true
	case tcell.KeyEnter:
		e.Enter()
		return true
	case tcell.KeyTab:
		e.Insert('\t')
		return true
	case tcell.KeyRune:
		chStr := evt.Str()
		if chStr != "" {
			e.Insert([]rune(chStr)[0])
		}
		return true
	default:
		return false
	}
}
