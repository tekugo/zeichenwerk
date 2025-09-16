package zeichenwerk

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestNewEditor tests the creation of new editor widgets
func TestNewEditor(t *testing.T) {
	editor := NewEditor("test-editor")

	if editor == nil {
		t.Fatal("NewEditor returned nil")
	}

	if editor.ID() != "test-editor" {
		t.Errorf("NewEditor() ID = %q, want %q", editor.ID(), "test-editor")
	}

	if !editor.Focusable() {
		t.Error("NewEditor() should create focusable widget")
	}

	// Check initial state
	if editor.line != 0 {
		t.Errorf("NewEditor() line = %d, want 0", editor.line)
	}

	if editor.column != 0 {
		t.Errorf("NewEditor() column = %d, want 0", editor.column)
	}

	if len(editor.content) != 1 {
		t.Errorf("NewEditor() content length = %d, want 1", len(editor.content))
	}

	if editor.tab != 4 {
		t.Errorf("NewEditor() tabWidth = %d, want 4", editor.tab)
	}

	if editor.spaces {
		t.Error("NewEditor() useSpaces should be false by default")
	}

	if editor.numbers {
		t.Error("NewEditor() showLineNumbers should be false by default")
	}

	if !editor.indent {
		t.Error("NewEditor() autoIndent should be true by default")
	}

	if editor.disabled {
		t.Error("NewEditor() readOnly should be false by default")
	}
}

// TestSetContent tests setting editor content
func TestSetContent(t *testing.T) {
	editor := NewEditor("test-editor")

	lines := []string{
		"Line 1",
		"Line 2 with more content",
		"Line 3",
		"",
		"Line 5 after empty line",
	}

	editor.SetContent(lines)

	// Check content was set correctly
	if len(editor.content) != len(lines) {
		t.Errorf("SetContent() content length = %d, want %d", len(editor.content), len(lines))
	}

	// Verify each line
	for i, expectedLine := range lines {
		if i >= len(editor.content) {
			t.Errorf("SetContent() missing line %d", i)
			continue
		}

		actualLine := editor.content[i].String()
		if actualLine != expectedLine {
			t.Errorf("SetContent() line %d = %q, want %q", i, actualLine, expectedLine)
		}
	}

	// Check cursor was reset
	if editor.line != 0 || editor.column != 0 {
		t.Errorf("SetContent() cursor = (%d, %d), want (0, 0)", editor.line, editor.column)
	}

	// Check offsets were reset
	if editor.offsetX != 0 || editor.offsetY != 0 {
		t.Errorf("SetContent() offsets = (%d, %d), want (0, 0)", editor.offsetX, editor.offsetY)
	}
}

// TestSetContentEmpty tests setting empty content
func TestSetContentEmpty(t *testing.T) {
	editor := NewEditor("test-editor")

	editor.SetContent([]string{})

	// Should have at least one empty line
	if len(editor.content) != 1 {
		t.Errorf("SetContent([]) content length = %d, want 1", len(editor.content))
	}

	if editor.content[0].String() != "" {
		t.Errorf("SetContent([]) first line = %q, want empty string", editor.content[0].String())
	}
}

// TestLoadText tests loading text with newlines
func TestLoadText(t *testing.T) {
	editor := NewEditor("test-editor")

	text := "Line 1\nLine 2\nLine 3\n\nLine 5"
	editor.LoadText(text)

	expectedLines := []string{"Line 1", "Line 2", "Line 3", "", "Line 5"}

	if len(editor.content) != len(expectedLines) {
		t.Errorf("LoadText() content length = %d, want %d", len(editor.content), len(expectedLines))
	}

	for i, expectedLine := range expectedLines {
		if i >= len(editor.content) {
			continue
		}
		actualLine := editor.content[i].String()
		if actualLine != expectedLine {
			t.Errorf("LoadText() line %d = %q, want %q", i, actualLine, expectedLine)
		}
	}
}

// TestGetContent tests retrieving editor content
func TestGetContent(t *testing.T) {
	editor := NewEditor("test-editor")

	originalLines := []string{
		"First line",
		"Second line",
		"Third line",
	}

	editor.SetContent(originalLines)
	retrievedLines := editor.GetContent()

	if len(retrievedLines) != len(originalLines) {
		t.Errorf("GetContent() length = %d, want %d", len(retrievedLines), len(originalLines))
	}

	for i, expectedLine := range originalLines {
		if i >= len(retrievedLines) {
			continue
		}
		if retrievedLines[i] != expectedLine {
			t.Errorf("GetContent() line %d = %q, want %q", i, retrievedLines[i], expectedLine)
		}
	}
}

// TestGetText tests retrieving complete text
func TestGetText(t *testing.T) {
	editor := NewEditor("test-editor")

	lines := []string{"Line 1", "Line 2", "Line 3"}
	expectedText := "Line 1\nLine 2\nLine 3"

	editor.SetContent(lines)
	actualText := editor.GetText()

	if actualText != expectedText {
		t.Errorf("GetText() = %q, want %q", actualText, expectedText)
	}
}

// TestConfiguration tests editor configuration methods
func TestConfiguration(t *testing.T) {
	editor := NewEditor("test-editor")

	// Test SetTabWidth
	editor.SetTabWidth(8)
	if editor.tab != 8 {
		t.Errorf("SetTabWidth(8) tabWidth = %d, want 8", editor.tab)
	}

	// Test invalid tab width (should be ignored)
	editor.SetTabWidth(0)
	if editor.tab != 8 {
		t.Errorf("SetTabWidth(0) should not change tabWidth, got %d", editor.tab)
	}

	// Test UseSpaces
	editor.UseSpaces(true)
	if !editor.spaces {
		t.Error("UseSpaces(true) useSpaces should be true")
	}

	// Test ShowLineNumbers
	editor.ShowLineNumbers(true)
	if !editor.numbers {
		t.Error("ShowLineNumbers(true) showLineNumbers should be true")
	}

	// Test SetAutoIndent
	editor.SetAutoIndent(false)
	if editor.indent {
		t.Error("SetAutoIndent(false) autoIndent should be false")
	}

	// Test SetReadOnly
	editor.SetReadOnly(true)
	if !editor.disabled {
		t.Error("SetReadOnly(true) readOnly should be true")
	}
}

// TestMoveTo tests cursor positioning
func TestMoveTo(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{
		"Short",
		"Much longer line with content",
		"Medium line",
	})

	// Test normal positioning
	editor.MoveTo(1, 5)
	if editor.line != 1 || editor.column != 5 {
		t.Errorf("MoveTo(1, 5) cursor = (%d, %d), want (1, 5)", editor.line, editor.column)
	}

	// Test line constraint (negative)
	editor.MoveTo(-1, 5)
	if editor.line != 0 {
		t.Errorf("MoveTo(-1, 5) line = %d, want 0", editor.line)
	}

	// Test line constraint (beyond end)
	editor.MoveTo(10, 5)
	if editor.line != 2 {
		t.Errorf("MoveTo(10, 5) line = %d, want 2", editor.line)
	}

	// Test column constraint (negative)
	editor.MoveTo(1, -5)
	if editor.column != 0 {
		t.Errorf("MoveTo(1, -5) column = %d, want 0", editor.column)
	}

	// Test column constraint (beyond line end)
	editor.MoveTo(0, 100)
	lineLength := editor.content[0].Length()
	if editor.column != lineLength {
		t.Errorf("MoveTo(0, 100) column = %d, want %d", editor.column, lineLength)
	}
}

// TestBasicMovement tests basic cursor movement operations
func TestBasicMovement(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{
		"Line 1",
		"Line 2",
		"Line 3",
	})

	// Test Right movement
	editor.MoveTo(0, 0)
	editor.Right()
	if editor.line != 0 || editor.column != 1 {
		t.Errorf("Right() cursor = (%d, %d), want (0, 1)", editor.line, editor.column)
	}

	// Test Right at line end (should move to next line)
	editor.MoveTo(0, 6) // End of "Line 1"
	editor.Right()
	if editor.line != 1 || editor.column != 0 {
		t.Errorf("Right() at line end cursor = (%d, %d), want (1, 0)", editor.line, editor.column)
	}

	// Test Left movement
	editor.MoveTo(1, 3)
	editor.Left()
	if editor.line != 1 || editor.column != 2 {
		t.Errorf("Left() cursor = (%d, %d), want (1, 2)", editor.line, editor.column)
	}

	// Test Left at line beginning (should move to previous line end)
	editor.MoveTo(1, 0)
	editor.Left()
	if editor.line != 0 || editor.column != 6 {
		t.Errorf("Left() at line beginning cursor = (%d, %d), want (0, 6)", editor.line, editor.column)
	}

	// Test Down movement
	editor.MoveTo(0, 3)
	editor.Down()
	if editor.line != 1 || editor.column != 3 {
		t.Errorf("Down() cursor = (%d, %d), want (1, 3)", editor.line, editor.column)
	}

	// Test Up movement
	editor.MoveTo(1, 3)
	editor.Up()
	if editor.line != 0 || editor.column != 3 {
		t.Errorf("Up() cursor = (%d, %d), want (0, 3)", editor.line, editor.column)
	}
}

// TestLineMovement tests line-based movement operations
func TestLineMovement(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{
		"Short",
		"Much longer line with more content",
		"Medium",
	})

	// Test Home
	editor.MoveTo(1, 10)
	editor.Home()
	if editor.line != 1 || editor.column != 0 {
		t.Errorf("Home() cursor = (%d, %d), want (1, 0)", editor.line, editor.column)
	}

	// Test End
	editor.MoveTo(1, 5)
	editor.End()
	expectedColumn := editor.content[1].Length()
	if editor.line != 1 || editor.column != expectedColumn {
		t.Errorf("End() cursor = (%d, %d), want (1, %d)", editor.line, editor.column, expectedColumn)
	}

	// Test DocumentHome
	editor.MoveTo(2, 5)
	editor.DocumentHome()
	if editor.line != 0 || editor.column != 0 {
		t.Errorf("DocumentHome() cursor = (%d, %d), want (0, 0)", editor.line, editor.column)
	}

	// Test DocumentEnd
	editor.MoveTo(0, 0)
	editor.DocumentEnd()
	lastLine := len(editor.content) - 1
	lastColumn := editor.content[lastLine].Length()
	if editor.line != lastLine || editor.column != lastColumn {
		t.Errorf("DocumentEnd() cursor = (%d, %d), want (%d, %d)", editor.line, editor.column, lastLine, lastColumn)
	}
}

// TestColumnConstraints tests column position constraints with different line lengths
func TestColumnConstraints(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{
		"Short",                               // 5 chars
		"Very long line with lots of content", // 33 chars
		"Med",                                 // 3 chars
	})

	// Move to long position in long line
	editor.MoveTo(1, 20)
	if editor.column != 20 {
		t.Errorf("MoveTo long line cursor column = %d, want 20", editor.column)
	}

	// Move up to shorter line (should constrain column)
	editor.Up()
	if editor.line != 0 || editor.column != 5 { // Should be constrained to line length
		t.Errorf("Up() to shorter line cursor = (%d, %d), want (0, 5)", editor.line, editor.column)
	}

	// Move down to longer line (should maintain column if possible)
	editor.Down()
	if editor.line != 1 || editor.column != 5 {
		t.Errorf("Down() to longer line cursor = (%d, %d), want (1, 5)", editor.line, editor.column)
	}

	// Move down to shorter line (should constrain again)
	editor.Down()
	if editor.line != 2 || editor.column != 3 { // Should be constrained to line length
		t.Errorf("Down() to shorter line cursor = (%d, %d), want (2, 3)", editor.line, editor.column)
	}
}

// TestTextInsertion tests character insertion
func TestTextInsertion(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{"Hello world"})

	// Test basic insertion
	editor.MoveTo(0, 5) // After "Hello"
	editor.Insert(' ')
	editor.Insert('t')
	editor.Insert('e')
	editor.Insert('s')
	editor.Insert('t')

	expected := "Hello test world"
	actual := editor.content[0].String()
	if actual != expected {
		t.Errorf("After insertion got %q, want %q", actual, expected)
	}

	// Check cursor position
	if editor.column != 10 { // After "Hello test"
		t.Errorf("After insertion cursor column = %d, want 10", editor.column)
	}
}

// TestTextDeletion tests character deletion (backspace)
func TestTextDeletion(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{"Hello world"})

	// Test basic deletion
	editor.MoveTo(0, 6) // After space after "Hello"
	editor.Delete()     // Should delete the space

	expected := "Helloworld"
	actual := editor.content[0].String()
	if actual != expected {
		t.Errorf("After deletion got %q, want %q", actual, expected)
	}

	// Check cursor position
	if editor.column != 5 { // After "Hello"
		t.Errorf("After deletion cursor column = %d, want 5", editor.column)
	}
}

// TestDeleteForward tests forward deletion
func TestDeleteForward(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{"Hello world"})

	// Test forward deletion
	editor.MoveTo(0, 5)    // At space after "Hello"
	editor.DeleteForward() // Should delete the space

	expected := "Helloworld"
	actual := editor.content[0].String()
	if actual != expected {
		t.Errorf("After forward deletion got %q, want %q", actual, expected)
	}

	// Check cursor position (should not move)
	if editor.column != 5 {
		t.Errorf("After forward deletion cursor column = %d, want 5", editor.column)
	}
}

// TestLineOperations tests line splitting and joining
func TestLineOperations(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{"Hello world"})

	// Test line splitting with Enter
	editor.MoveTo(0, 5) // After "Hello"
	editor.Enter()

	if len(editor.content) != 2 {
		t.Errorf("After Enter content length = %d, want 2", len(editor.content))
	}

	if editor.content[0].String() != "Hello" {
		t.Errorf("First line after Enter = %q, want %q", editor.content[0].String(), "Hello")
	}

	if editor.content[1].String() != " world" {
		t.Errorf("Second line after Enter = %q, want %q", editor.content[1].String(), " world")
	}

	// Check cursor position (should be at beginning of new line)
	if editor.line != 1 || editor.column != 0 {
		t.Errorf("After Enter cursor = (%d, %d), want (1, 0)", editor.line, editor.column)
	}

	// Test line joining with backspace at line beginning
	editor.MoveTo(1, 0) // Beginning of second line
	editor.Delete()     // Should join lines

	if len(editor.content) != 1 {
		t.Errorf("After line join content length = %d, want 1", len(editor.content))
	}

	if editor.content[0].String() != "Hello world" {
		t.Errorf("After line join got %q, want %q", editor.content[0].String(), "Hello world")
	}

	// Check cursor position (should be at join point)
	if editor.line != 0 || editor.column != 5 {
		t.Errorf("After line join cursor = (%d, %d), want (0, 5)", editor.line, editor.column)
	}
}

// TestAutoIndent tests automatic indentation
func TestAutoIndent(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{"    indented line"})
	editor.SetAutoIndent(true)

	// Test Enter with auto-indent
	editor.MoveTo(0, 16) // End of line
	editor.Enter()

	if len(editor.content) != 2 {
		t.Errorf("After Enter content length = %d, want 2", len(editor.content))
	}

	// Second line should inherit indentation
	secondLine := editor.content[1].String()
	if !strings.HasPrefix(secondLine, "    ") {
		t.Errorf("Second line %q should start with 4 spaces", secondLine)
	}

	// Cursor should be positioned after indentation
	if editor.column != 4 {
		t.Errorf("After auto-indent cursor column = %d, want 4", editor.column)
	}
}

// TestTabHandling tests tab insertion behavior
func TestTabHandling(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{""})

	// Test tab character insertion (default)
	editor.MoveTo(0, 0)
	editor.Insert('\t')

	if editor.content[0].String() != "\t" {
		t.Errorf("Tab insertion got %q, want tab character", editor.content[0].String())
	}

	// Test spaces instead of tabs
	editor.SetContent([]string{""})
	editor.UseSpaces(true)
	editor.SetTabWidth(4)
	editor.MoveTo(0, 0)
	editor.Insert('\t')

	expected := "    " // 4 spaces
	actual := editor.content[0].String()
	if actual != expected {
		t.Errorf("Tab as spaces got %q, want %q", actual, expected)
	}

	if editor.column != 4 {
		t.Errorf("After tab as spaces cursor column = %d, want 4", editor.column)
	}
}

// TestReadOnlyMode tests read-only functionality
func TestReadOnlyMode(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{"Read only content"})
	editor.SetReadOnly(true)

	originalContent := editor.content[0].String()

	// Test that insertion is blocked
	editor.Insert('X')
	if editor.content[0].String() != originalContent {
		t.Error("Insert should be blocked in read-only mode")
	}

	// Test that deletion is blocked
	editor.MoveTo(0, 5)
	editor.Delete()
	if editor.content[0].String() != originalContent {
		t.Error("Delete should be blocked in read-only mode")
	}

	// Test that Enter is blocked
	lineCount := len(editor.content)
	editor.Enter()
	if len(editor.content) != lineCount {
		t.Error("Enter should be blocked in read-only mode")
	}

	// Test that movement still works
	editor.MoveTo(0, 10)
	if editor.column != 10 {
		t.Error("Movement should still work in read-only mode")
	}
}

// TestKeyboardEvents tests keyboard event handling
func TestKeyboardEvents(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{"Test line"})

	tests := []struct {
		name        string
		key         tcell.Key
		rune        rune
		startLine   int
		startColumn int
		expectLine  int
		expectCol   int
	}{
		{"Left Arrow", tcell.KeyLeft, 0, 0, 5, 0, 4},
		{"Right Arrow", tcell.KeyRight, 0, 0, 4, 0, 5},
		{"Up Arrow", tcell.KeyUp, 0, 1, 3, 0, 3},
		{"Down Arrow", tcell.KeyDown, 0, 0, 3, 0, 3},
		{"Home", tcell.KeyHome, 0, 0, 5, 0, 0},
		{"End", tcell.KeyEnd, 0, 0, 0, 0, 9},
		{"Backspace", tcell.KeyBackspace, 0, 0, 5, 0, 4},
		{"Delete", tcell.KeyDelete, 0, 0, 5, 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset content for each test
			editor.SetContent([]string{"Test line"})
			editor.MoveTo(tt.startLine, tt.startColumn)

			// Create and handle the event
			var event tcell.Event
			if tt.key == tcell.KeyRune {
				event = tcell.NewEventKey(tt.key, tt.rune, tcell.ModNone)
			} else {
				event = tcell.NewEventKey(tt.key, 0, tcell.ModNone)
			}

			handled := editor.Handle(event)
			if !handled {
				t.Errorf("Event %s should be handled", tt.name)
			}

			if editor.line != tt.expectLine || editor.column != tt.expectCol {
				t.Errorf("After %s cursor = (%d, %d), want (%d, %d)",
					tt.name, editor.line, editor.column, tt.expectLine, tt.expectCol)
			}
		})
	}
}

// TestRuneInsertion tests printable character insertion via keyboard
func TestRuneInsertion(t *testing.T) {
	editor := NewEditor("test-editor")
	editor.SetContent([]string{""})

	// Test inserting printable characters
	chars := []rune{'H', 'e', 'l', 'l', 'o', '!', '世', '界'}

	for _, ch := range chars {
		event := tcell.NewEventKey(tcell.KeyRune, ch, tcell.ModNone)
		handled := editor.Handle(event)
		if !handled {
			t.Errorf("Rune event for %c should be handled", ch)
		}
	}

	expected := "Hello!世界"
	actual := editor.content[0].String()
	if actual != expected {
		t.Errorf("After rune insertion got %q, want %q", actual, expected)
	}
}

