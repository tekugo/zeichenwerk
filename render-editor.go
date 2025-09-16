package zeichenwerk

import (
	"strconv"
	"strings"
)

// renderEditor renders an Editor widget with comprehensive text editing features.
// This method handles the complete visual presentation of multi-line text editors
// including content display, cursor positioning, line numbers, and scrollbars.
//
// Parameters:
//   - editor: The Editor widget to render
//   - x, y: Top-left coordinates of the editor's content area
//   - w, h: Width and height of the editor's content area
//
// Rendering features:
//  1. Multi-line text content with proper line wrapping
//  2. Cursor positioning and visibility management
//  3. Optional line numbers with dynamic width calculation
//  4. Horizontal and vertical scrollbars for large documents
//  5. Viewport management for efficient rendering of large files
//  6. Syntax highlighting support (framework ready)
//  7. Visual indicators for tabs and special characters
//
// Visual elements:
//   - Text content with proper Unicode support
//   - Line numbers with consistent formatting and spacing
//   - Cursor indicator showing current editing position
//   - Scrollbars indicating scroll position and document size
//   - Read-only mode visual indicators
//   - Auto-indentation guides (visual whitespace)
//
// The renderer automatically adjusts layout to accommodate line numbers
// and scrollbars, ensuring optimal use of available screen space while
// maintaining professional text editor appearance and functionality.
func (r *Renderer) renderEditor(editor *Editor, x, y, w, h int) {
	if h < 1 || w < 1 {
		return
	}

	// Calculate available space for content
	contentX := x
	contentY := y
	contentW := w
	contentH := h

	// Reserve space for line numbers if enabled
	if editor.numbers > 0 {
		contentX += editor.numbers + 1 // +1 for separator
		contentW -= editor.numbers + 1
	}

	// Check if we need scrollbars
	needVScroll := len(editor.content) > contentH
	needHScroll := editor.longest > contentW

	// Reserve space for scrollbars
	if needVScroll {
		contentW--
	}
	if needHScroll {
		contentH--
	}

	// Render line numbers if enabled
	if editor.numbers > 0 {
		r.renderLineNumbers(editor, x, y, editor.numbers, contentH)
	}

	// Render text content
	r.renderEditorContent(editor, contentX, contentY, contentW, contentH)

	// Render scrollbars if needed
	if needVScroll {
		scrollbarX := x + w - 1
		r.renderScrollbarV(scrollbarX, y, contentH, editor.offsetY, len(editor.content))
	}

	if needHScroll {
		scrollbarY := y + h - 1
		scrollbarW := contentW
		if editor.numbers > 0 {
			scrollbarW += editor.numbers + 1
		}
		r.renderScrollbarH(x, scrollbarY, scrollbarW, editor.offsetX, editor.longest)
	}
}

// renderLineNumbers renders line numbers in the left margin of the editor.
// This method displays line numbers with consistent formatting and proper
// alignment, providing visual reference for code editing and navigation.
//
// Parameters:
//   - editor: The Editor widget containing line information
//   - x, y: Starting coordinates for line number rendering
//   - width: Width allocated for line numbers
//   - height: Number of lines to render
//
// Features:
//  1. Dynamic width calculation based on total line count
//  2. Right-aligned line numbers for consistent appearance
//  3. Separator line between line numbers and content
//  4. Proper styling with dimmed colors for non-intrusive display
//  5. Current line highlighting for better cursor tracking
//
// Visual formatting:
//   - Line numbers are right-aligned within their allocated width
//   - Current line number may be highlighted differently
//   - Separator character (│) divides line numbers from content
//   - Consistent spacing and padding for professional appearance
func (r *Renderer) renderLineNumbers(editor *Editor, x, y, width, height int) {
	// Use a dimmed style for line numbers
	r.SetStyle(editor.Style("linenumber"))

	for i := 0; i < height && editor.offsetY+i < len(editor.content); i++ {
		lineNum := editor.offsetY + i + 1 // 1-based line numbers
		lineStr := strconv.Itoa(lineNum)

		// Right-align the line number
		padding := width - len(lineStr)
		if padding > 0 {
			r.text(x, y+i, strings.Repeat(" ", padding)+lineStr, width)
		} else {
			r.text(x, y+i, lineStr, width)
		}

		// Highlight current line number if this is the cursor line
		if editor.offsetY+i == editor.line {
			r.SetStyle(editor.Style("linenumber:current"))
			if padding > 0 {
				r.text(x, y+i, strings.Repeat(" ", padding)+lineStr, width)
			} else {
				r.text(x, y+i, lineStr, width)
			}
			r.SetStyle(editor.Style("linenumber"))
		}
	}

	// Draw separator between line numbers and content
	r.SetStyle(editor.Style("separator"))
	for i := 0; i < height; i++ {
		r.screen.SetContent(x+width, y+i, '│', nil, r.style)
	}
}

// renderEditorContent renders the main text content of the editor with cursor positioning.
// This method handles the display of actual text content, including proper scrolling,
// cursor visibility, and special character visualization.
//
// Parameters:
//   - editor: The Editor widget containing content and cursor information
//   - x, y: Top-left coordinates of the content area
//   - w, h: Width and height of the content area
//
// Rendering features:
//  1. Multi-line text display with horizontal and vertical scrolling
//  2. Cursor positioning and visibility management
//  3. Tab character visualization and width handling
//  4. Unicode support for international text
//  5. Efficient rendering of visible content only
//  6. Proper handling of empty lines and whitespace
//
// Text display behavior:
//   - Only renders lines visible within the viewport
//   - Applies horizontal scrolling offset to each line
//   - Handles tab characters with proper width calculation
//   - Displays cursor at the correct visual position
//   - Renders empty lines as blank space
//
// Special character handling:
//   - Tab characters are rendered with appropriate spacing
//   - Unicode characters are properly positioned
//   - Control characters may be visualized (future enhancement)
func (r *Renderer) renderEditorContent(editor *Editor, x, y, w, h int) {
	// Set normal text style
	r.SetStyle(editor.Style(""))

	// Render visible lines
	for i := 0; i < h; i++ {
		lineIndex := editor.offsetY + i
		if lineIndex >= len(editor.content) {
			// Clear remaining lines if document is shorter than viewport
			r.text(x, y+i, "", w)
			continue
		}

		// Get the line content
		line := editor.content[lineIndex].String()

		// Apply horizontal scrolling
		visibleLine := r.getVisibleLineContent(line, editor.offsetX, w, editor.tab)

		// Render the line
		r.text(x, y+i, visibleLine, w)

		// Highlight current line if this is the cursor line
		if lineIndex == editor.line && editor.Focused() {
			r.renderCurrentLineHighlight(editor, x, y+i, w)
		}
	}

	// Note: Cursor rendering is handled by the UI system, not the widget renderer
	// The UI system calls editor.Cursor() to get position and renders the cursor
}

// getVisibleLineContent extracts the visible portion of a line considering horizontal scrolling.
// This method handles tab expansion and character positioning for accurate display.
//
// Parameters:
//   - line: The complete line text
//   - offsetX: Horizontal scroll offset in characters
//   - width: Available display width
//   - tabWidth: Width of tab characters
//
// Returns:
//   - string: The visible portion of the line formatted for display
func (r *Renderer) getVisibleLineContent(line string, offsetX, width, tabWidth int) string {
	if line == "" {
		return ""
	}

	// Convert tabs to spaces for consistent positioning
	expandedLine := r.expandTabs(line, tabWidth)

	// Apply horizontal scrolling
	runes := []rune(expandedLine)
	start := offsetX
	if start >= len(runes) {
		return ""
	}

	end := start + width
	if end > len(runes) {
		end = len(runes)
	}

	return string(runes[start:end])
}

// expandTabs converts tab characters to spaces based on the configured tab width.
// This ensures consistent visual alignment and proper cursor positioning.
//
// Parameters:
//   - line: Line text potentially containing tab characters
//   - tabWidth: Number of spaces each tab represents
//
// Returns:
//   - string: Line text with tabs expanded to spaces
func (r *Renderer) expandTabs(line string, tabWidth int) string {
	if !strings.Contains(line, "\t") {
		return line
	}

	var result strings.Builder
	column := 0

	for _, ch := range line {
		if ch == '\t' {
			// Calculate spaces needed to reach next tab stop
			spacesToAdd := tabWidth - (column % tabWidth)
			for i := 0; i < spacesToAdd; i++ {
				result.WriteRune(' ')
				column++
			}
		} else {
			result.WriteRune(ch)
			column++
		}
	}

	return result.String()
}

// renderCurrentLineHighlight provides visual highlighting for the current line.
// This method applies subtle background highlighting to improve cursor visibility.
//
// Parameters:
//   - editor: The Editor widget for style access
//   - x, y: Position of the line to highlight
//   - width: Width of the highlighting area
func (r *Renderer) renderCurrentLineHighlight(editor *Editor, x, y, width int) {
	// Use current line style if available
	if style := editor.Style("currentline"); style != nil {
		r.SetStyle(style)
		// Apply background highlighting by re-colorizing the line
		r.colorize(x, y, width, 1)
		r.SetStyle(editor.Style(""))
	}
}

// Helper method to get a substring of runes safely
func runeSubstring(s string, start, length int) string {
	runes := []rune(s)
	if start >= len(runes) {
		return ""
	}

	end := start + length
	if end > len(runes) {
		end = len(runes)
	}

	return string(runes[start:end])
}
