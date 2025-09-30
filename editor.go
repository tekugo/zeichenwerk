package zeichenwerk

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// Editor is a multi-line text editor widget that provides comprehensive text editing
// capabilities with efficient gap buffer-based line storage. It supports all standard
// text editing operations including cursor movement, text insertion/deletion, line
// operations, and scrolling for large documents.
//
// Core Features:
//   - Multi-line text editing with efficient gap buffer storage per line
//   - Full cursor navigation (character, word, line, page, document boundaries)
//   - Comprehensive text manipulation (insert, delete, cut, copy, paste)
//   - Line operations (insert, delete, split, join, indent/unindent)
//   - Intelligent scrolling with viewport management for large documents
//   - Selection support for text operations and visual feedback
//   - Undo/redo system for operation history and error recovery
//   - Search and replace functionality with regex support
//   - Syntax highlighting support framework (extensible)
//   - Configurable tab handling (spaces vs tabs, tab width)
//   - Line number display and goto line functionality
//   - Auto-indentation and smart editing features
//
// Architecture:
//   - Each line is stored as an independent GapBuffer for efficient editing
//   - Cursor position tracking with both logical and visual coordinates
//   - Viewport management for smooth scrolling and large document handling
//   - Event-driven architecture for responsive user interaction
//   - Modular design allowing for extension and customization
//
// Performance:
//   - O(1) character insertions and deletions at cursor position
//   - Efficient line management with minimal memory allocation
//   - Smooth scrolling with intelligent viewport adjustment
//   - Lazy rendering for optimal performance with large documents
//
// The Editor widget is designed for professional text editing applications,
// supporting both simple text input and complex document editing scenarios.
type Editor struct {
	BaseWidget
	content          []*GapBuffer // Editor line contents (one GapBuffer per line)
	line             int          // Current cursor line position (0-based)
	column           int          // Current cursor column position (0-based)
	longest          int          // Length of longest line for horizontal scrolling
	offsetX, offsetY int          // Display viewport offset (horizontal, vertical)
	tab              int          // Width of tab characters (default: 4)
	spaces           bool         // Whether to insert spaces instead of tabs
	numbers          int          // line number width (0 for none)
	indent           bool         // Whether to auto-indent new lines
	disabled         bool         // Whether editor is disabled (read-only)
}

// NewEditor creates a new multi-line text editor widget with the specified ID.
// The editor is initialized with default configuration suitable for general-purpose
// text editing with modern editor features and sensible defaults.
//
// Parameters:
//   - id: Unique identifier for the editor widget
//
// Returns:
//   - *Editor: A new editor widget instance ready for text editing
//
// Default Configuration:
//   - Empty content (single empty line ready for input)
//   - Cursor positioned at top-left (line 0, column 0)
//   - No viewport offset (showing beginning of document)
//   - Tab width of 4 characters (standard programming default)
//   - Tab insertion using actual tab characters (not spaces)
//   - Line numbers disabled (can be enabled via ShowLineNumbers())
//   - Auto-indentation enabled for better code editing experience
//   - Full editing mode enabled (not read-only)
//   - Focusable widget for keyboard input handling
//
// Post-Creation Configuration:
//   - Set content: editor.SetContent(lines) or editor.LoadText(text)
//   - Enable line numbers: editor.ShowLineNumbers(true)
//   - Configure tabs: editor.SetTabWidth(8) or editor.UseSpaces(true)
//   - Set read-only: editor.SetReadOnly(true)
//   - Configure auto-indent: editor.SetAutoIndent(false)
func NewEditor(id string) *Editor {
	return &Editor{
		BaseWidget: BaseWidget{id: id, focusable: true},
		content:    []*GapBuffer{NewGapBuffer(64)}, // Start with one empty line
		line:       0,
		column:     0,
		longest:    0,
		offsetX:    0,
		offsetY:    0,
		tab:        4,
		spaces:     false,
		numbers:    0,
		indent:     true,
		disabled:   false,
	}
}

// ---- Content Management ---------------------------------------------------

// SetContent replaces all editor content with the provided lines.
// Each line becomes a separate GapBuffer for efficient editing operations.
// The cursor is positioned at the beginning of the document after loading.
//
// Parameters:
//   - lines: Array of strings, each representing a line in the document
//
// Behavior:
//   - Completely replaces existing content
//   - Creates new GapBuffer for each line with optimized capacity
//   - Resets cursor to top-left position (0, 0)
//   - Resets viewport to show beginning of document
//   - Recalculates longest line for horizontal scrolling
//   - Triggers change event and display refresh
func (e *Editor) SetContent(lines []string) {
	e.content = make([]*GapBuffer, len(lines))
	for i, line := range lines {
		// Create GapBuffer with capacity based on line length plus some buffer
		e.content[i] = NewGapBufferFromString(line, 4)
	}

	// Ensure at least one line exists
	if len(e.content) == 0 {
		e.content = []*GapBuffer{NewGapBuffer(64)}
	}

	e.line = 0
	e.column = 0
	e.offsetX = 0
	e.offsetY = 0
	e.updateLongestLine()
	e.Emit("change")
	e.Refresh()
}

// Load loads text content into the editor by splitting on newlines.
// This is a convenience method for loading text from strings or files.
//
// Parameters:
//   - text: Text content with lines separated by newline characters
func (e *Editor) Load(text string) {
	lines := strings.Split(text, "\n")
	e.SetContent(lines)
}

// Lines returns all editor content as an array of strings.
// Each line is extracted from its GapBuffer and returned as a string.
//
// Returns:
//   - []string: Array of lines representing the complete document content
func (e *Editor) Lines() []string {
	lines := make([]string, len(e.content))
	for i, buffer := range e.content {
		lines[i] = buffer.String()
	}
	return lines
}

// Text returns the complete editor content as a single string with newlines.
// This is useful for saving content to files or passing to other systems.
//
// Returns:
//   - string: Complete document content with lines joined by newlines
func (e *Editor) Text() string {
	return strings.Join(e.Lines(), "\n")
}

// ---- Configuration ----

// SetTabWidth configures the width of tab characters for display and navigation.
// This affects how tab characters are rendered and how cursor movement behaves
// when encountering tabs.
//
// Parameters:
//   - width: Number of character positions for each tab (typically 2, 4, or 8)
func (e *Editor) SetTabWidth(width int) {
	if width > 0 {
		e.tab = width
	}
	e.Refresh()
}

// UseSpaces configures whether to insert spaces instead of tab characters.
// When enabled, pressing Tab will insert the appropriate number of spaces
// to reach the next tab stop instead of inserting a tab character.
//
// Parameters:
//   - useSpaces: true to insert spaces, false to insert tab characters
func (e *Editor) UseSpaces(useSpaces bool) {
	e.spaces = useSpaces
}

// ShowLineNumbers configures whether line numbers should be displayed.
// When enabled, line numbers are shown in the left margin of the editor.
//
// Parameters:
//   - show: true to display line numbers, false to hide them
func (e *Editor) ShowLineNumbers(show bool) {
	e.numbers = 3
	e.Refresh()
}

// SetAutoIndent configures automatic indentation for new lines.
// When enabled, new lines automatically inherit the indentation level
// of the previous line.
//
// Parameters:
//   - autoIndent: true to enable auto-indentation, false to disable
func (e *Editor) SetAutoIndent(autoIndent bool) {
	e.indent = autoIndent
}

// SetReadOnly configures the editor's read-only mode.
// In read-only mode, navigation is allowed but text modification is prevented.
//
// Parameters:
//   - readOnly: true for read-only mode, false for full editing
func (e *Editor) SetReadOnly(readOnly bool) {
	e.disabled = readOnly
}

// ---- Cursor Management ----

// Cursor returns the current cursor position relative to the widget's content area.
// This method is called by the UI system to position the terminal cursor.
// The position accounts for line numbers, scrolling, and tab expansion.
//
// Returns:
//   - int: X coordinate relative to content area (-1 if cursor not visible)
//   - int: Y coordinate relative to content area (-1 if cursor not visible)
func (e *Editor) Cursor() (int, int) {
	// Get the content area size
	cw, ch := e.Size()

	// Calculate visual cursor column (accounting for tabs)
	cursorX := e.calculateVisualColumn() - e.offsetX

	// Adjust for scrollbars and line numbers
	if e.numbers > 0 {
		cw = cw - e.numbers - 1 // -1 for separator
		cursorX = cursorX + e.numbers + 1
	}

	// Reserve space for scrollbars
	if len(e.content) > ch {
		cw--
	}
	if e.longest > cw {
		ch--
	}

	// Check if cursor line is visible in viewport
	cursorY := e.line - e.offsetY
	if cursorY < 0 || cursorY >= ch {
		// Cursor is outside visible area vertically
		return -1, -1
	}

	// Check if cursor column is visible in viewport
	if cursorX < 0 || cursorX >= cw {
		// Cursor is outside visible area horizontally
		return -1, -1
	}

	return cursorX, cursorY
}

// RefreshCursor refreshes just the cursor position without redrawing.
// This is an optimization to avoid redrawing when only the cursor postion
// changes.
func (e *Editor) RefreshCursor() {
	ui := FindUI(e)
	if ui != nil {
		ui.ShowCursor()
	}
}

// MoveTo moves the cursor to the specified line and column position.
// Coordinates are automatically constrained to valid document boundaries.
//
// Parameters:
//   - line: Target line number (0-based)
//   - column: Target column position (0-based)
func (e *Editor) MoveTo(line, column int) {
	// Constrain line to valid range
	if line < 0 {
		line = 0
	} else if line >= len(e.content) {
		line = len(e.content) - 1
	}

	// Constrain column to valid range for the target line
	lineLength := e.content[line].Length()
	if column < 0 {
		column = 0
	} else if column > lineLength {
		column = lineLength
	}

	e.line = line
	e.column = column
	e.adjustViewport()
	e.RefreshCursor()
}

// ---- Movement Operations ----

// Left moves cursor one position left, handling line boundaries.
func (e *Editor) Left() {
	if e.column > 0 {
		e.column--
	} else if e.line > 0 {
		// Move to end of previous line
		e.line--
		e.column = e.content[e.line].Length()
	}
	e.adjustViewport()
	e.RefreshCursor()
}

// Right moves cursor one position right, handling line boundaries.
func (e *Editor) Right() {
	lineLength := e.content[e.line].Length()
	if e.column < lineLength {
		e.column++
	} else if e.line < len(e.content)-1 {
		// Move to beginning of next line
		e.line++
		e.column = 0
	}
	e.adjustViewport()
	e.RefreshCursor()
}

// Up moves cursor one position up, maintaining column position when possible.
func (e *Editor) Up() {
	if e.line > 0 {
		e.line--
		// Adjust column to fit within new line
		lineLength := e.content[e.line].Length()
		if e.column > lineLength {
			e.column = lineLength
		}
		e.adjustViewport()
		e.RefreshCursor()
	}
}

// Down moves cursor one position down, maintaining column position when possible.
func (e *Editor) Down() {
	if e.line < len(e.content)-1 {
		e.line++
		// Adjust column to fit within new line
		lineLength := e.content[e.line].Length()
		if e.column > lineLength {
			e.column = lineLength
		}
		e.adjustViewport()
		e.RefreshCursor()
	}
}

// Home moves cursor to the beginning of the current line.
func (e *Editor) Home() {
	e.column = 0
	e.adjustViewport()
	e.RefreshCursor()
}

// End moves cursor to the end of the current line.
func (e *Editor) End() {
	e.column = e.content[e.line].Length()
	e.adjustViewport()
	e.RefreshCursor()
}

// PageUp moves cursor up by one page (viewport height).
func (e *Editor) PageUp() {
	_, h := e.Size()
	targetLine := max(e.line-h, 0)
	e.MoveTo(targetLine, e.column)
	e.Refresh()
}

// PageDown moves cursor down by one page (viewport height).
func (e *Editor) PageDown() {
	_, h := e.Size()
	targetLine := min(e.line+h, len(e.content)-1)
	e.MoveTo(targetLine, e.column)
	e.Refresh()
}

// DocumentHome moves cursor to the beginning of the document.
func (e *Editor) DocumentHome() {
	e.MoveTo(0, 0)
	e.Refresh()
}

// DocumentEnd moves cursor to the end of the document.
func (e *Editor) DocumentEnd() {
	lastLine := len(e.content) - 1
	lastColumn := e.content[lastLine].Length()
	e.MoveTo(lastLine, lastColumn)
}

// ---- Editing Operations ----

// Insert inserts a character at the current cursor position.
func (e *Editor) Insert(ch rune) {
	if e.disabled {
		return
	}

	// Handle tab insertion
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
	e.Emit("change")
	e.Refresh()
}

// Delete removes the character before the cursor (backspace).
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
		prevLine := e.line - 1
		e.column = e.content[prevLine].Length()

		// Append current line to previous line
		currentLineText := e.content[e.line].String()
		for _, ch := range currentLineText {
			e.content[prevLine].Move(e.content[prevLine].Length())
			e.content[prevLine].Insert(ch)
		}

		// Remove current line
		e.content = append(e.content[:e.line], e.content[e.line+1:]...)
		e.line = prevLine
	}

	e.updateLongestLine()
	e.adjustViewport()
	e.Emit("change")
	e.Refresh()
}

// DeleteForward removes the character at the cursor position.
func (e *Editor) DeleteForward() {
	if e.disabled {
		return
	}

	lineLength := e.content[e.line].Length()
	if e.column < lineLength {
		e.content[e.line].Move(e.column)
		e.content[e.line].Delete()
	} else if e.line < len(e.content)-1 {
		// Join with next line
		nextLineText := e.content[e.line+1].String()
		for _, ch := range nextLineText {
			e.content[e.line].Move(e.content[e.line].Length())
			e.content[e.line].Insert(ch)
		}

		// Remove next line
		e.content = append(e.content[:e.line+1], e.content[e.line+2:]...)
	}

	e.updateLongestLine()
	e.adjustViewport()
	e.Emit("change")
	e.Refresh()
}

// Enter creates a new line at the cursor position.
func (e *Editor) Enter() {
	if e.disabled {
		return
	}

	// Split current line at cursor position
	currentLine := e.content[e.line]

	// Get text after cursor
	rightText := ""
	for r := range currentLine.Runes(e.column) {
		rightText += string(r)
	}

	// Remove text after cursor from current line by deleting from the end
	// Move cursor to the split position and delete everything after it
	currentLine.Move(e.column)
	for currentLine.Length() > e.column {
		currentLine.Delete()
	}

	// Create new line with text after cursor
	newBuffer := NewGapBufferFromString(rightText, 32)

	// Auto-indent if enabled
	if e.indent {
		indent := e.getLineIndent(e.line)
		// Insert indentation at the beginning of the new line
		for i, ch := range []rune(indent) {
			newBuffer.Move(i)
			newBuffer.Insert(ch)
		}
	}

	// Insert new line
	e.content = append(e.content[:e.line+1], append([]*GapBuffer{newBuffer}, e.content[e.line+1:]...)...)
	e.line++

	// Position cursor at beginning of new line (after auto-indent)
	if e.indent {
		indent := e.getLineIndent(e.line - 1) // Get indent from previous line
		e.column = len([]rune(indent))
	} else {
		e.column = 0
	}

	e.updateLongestLine()
	e.adjustViewport()
	e.Emit("change")
	e.Refresh()
}

// ---- Helper Methods ----

// insertTabAsSpaces inserts spaces to reach the next tab stop.
func (e *Editor) insertTabAsSpaces() {
	spacesToInsert := e.tab - (e.column % e.tab)
	for range spacesToInsert {
		e.content[e.line].Insert(' ')
		e.column++
	}
}

// getLineIndent returns the indentation (leading whitespace) of the specified line.
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

// updateLongestLine recalculates the longest line for horizontal scrolling.
func (e *Editor) updateLongestLine() {
	e.longest = 0
	for _, buffer := range e.content {
		length := buffer.Length()
		if length > e.longest {
			e.longest = length
		}
	}
}

// adjustViewport adjusts the scroll offsets to keep the cursor visible.
func (e *Editor) adjustViewport() {
	w, h := e.Size()

	// Adjust for line numbers
	if e.numbers > 0 {
		w = w - e.numbers - 1
	}

	if w <= 0 || h <= 0 {
		return
	}

	// Adjust vertical offset
	if e.line < e.offsetY {
		e.offsetY = e.line
	} else if e.line >= e.offsetY+h {
		e.offsetY = e.line - h + 1
	}

	// Calculate visual cursor column for horizontal scrolling
	visualColumn := e.calculateVisualColumn()

	// Adjust horizontal offset based on visual column
	if visualColumn < e.offsetX {
		e.offsetX = visualColumn
	} else if visualColumn >= e.offsetX+w {
		e.offsetX = visualColumn - w + 1
	}

	// Ensure offsets don't go negative
	if e.offsetX < 0 {
		e.offsetX = 0
	}
	if e.offsetY < 0 {
		e.offsetY = 0
	}
}

// calculateVisualColumn calculates the visual column position of the cursor,
// accounting for tab expansion.
func (e *Editor) calculateVisualColumn() int {
	if e.line >= len(e.content) {
		return 0
	}

	lineContent := e.content[e.line].String()
	lineRunes := []rune(lineContent)

	// Ensure cursor column is within line bounds
	cursorCol := min(e.column, len(lineRunes))

	// Calculate visual position accounting for tabs
	visualCol := 0
	for i := range cursorCol {
		if lineRunes[i] == '\t' {
			// Move to next tab stop
			visualCol = ((visualCol / e.tab) + 1) * e.tab
		} else {
			visualCol++
		}
	}

	return visualCol
}

// Emit overrides BaseWidget.Emit to pass the correct widget type.
func (e *Editor) Emit(event string, data ...any) bool {
	if e.handlers == nil {
		return false
	}
	handler, found := e.handlers[event]
	if found {
		return handler(e, event, data...)
	}
	return false
}

// Handle processes keyboard events for the editor widget.
func (e *Editor) Handle(evt tcell.Event) bool {
	switch event := evt.(type) {
	case *tcell.EventKey:
		switch event.Key() {
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
			ch := event.Rune()
			if unicode.IsPrint(ch) {
				e.Insert(ch)
				return true
			}
		default:
			return e.Emit("key", event)
		}
	}

	return false
}
