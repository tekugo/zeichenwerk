package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// Test helper functions for Input widget testing

func TestNewInput(t *testing.T) {
	t.Run("creates input with default values", func(t *testing.T) {
		input := NewInput("test-input")
		
		if input == nil {
			t.Fatal("Expected input instance, got nil")
		}
		
		if input.ID() != "test-input" {
			t.Errorf("Expected ID 'test-input', got '%s'", input.ID())
		}
		
		if input.Text != "" {
			t.Errorf("Expected empty text, got '%s'", input.Text)
		}
		
		if input.Pos != 0 {
			t.Errorf("Expected cursor position 0, got %d", input.Pos)
		}
		
		if input.Offset != 0 {
			t.Errorf("Expected offset 0, got %d", input.Offset)
		}
		
		if input.Max != 0 {
			t.Errorf("Expected max length 0 (unlimited), got %d", input.Max)
		}
		
		if input.Placeholder != "" {
			t.Errorf("Expected empty placeholder, got '%s'", input.Placeholder)
		}
		
		if input.Masked {
			t.Error("Expected masking to be disabled")
		}
		
		if input.MaskChar != '*' {
			t.Errorf("Expected mask character '*', got '%c'", input.MaskChar)
		}
		
		if input.ReadOnly {
			t.Error("Expected read-only to be disabled")
		}
		
		if !input.Focusable() {
			t.Error("Expected input to be focusable")
		}
	})
}

func TestInput_SetText(t *testing.T) {
	t.Run("sets text content", func(t *testing.T) {
		input := NewInput("test")
		
		input.SetText("hello world")
		
		if input.Text != "hello world" {
			t.Errorf("Expected text 'hello world', got '%s'", input.Text)
		}
	})
	
	t.Run("respects read-only mode", func(t *testing.T) {
		input := NewInput("test")
		input.ReadOnly = true
		
		input.SetText("hello")
		
		if input.Text != "" {
			t.Errorf("Expected empty text in read-only mode, got '%s'", input.Text)
		}
	})
	
	t.Run("enforces maximum length", func(t *testing.T) {
		input := NewInput("test")
		input.Max = 5
		
		input.SetText("hello world")
		
		if len([]rune(input.Text)) > 5 {
			t.Errorf("Expected text length <= 5, got %d characters", len([]rune(input.Text)))
		}
	})
	
	t.Run("adjusts cursor position when text is shorter", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello world"
		input.Pos = 10
		
		input.SetText("hi")
		
		if input.Pos > len([]rune(input.Text)) {
			t.Errorf("Expected cursor position <= %d, got %d", len([]rune(input.Text)), input.Pos)
		}
	})
	
	t.Run("triggers change event", func(t *testing.T) {
		input := NewInput("test")
		eventTriggered := false
		
		input.On("change", func(w Widget, event string, data ...any) bool {
			eventTriggered = true
			if len(data) > 0 {
				if text, ok := data[0].(string); ok {
					if text != "test text" {
						t.Errorf("Expected event data 'test text', got '%s'", text)
					}
				}
			}
			return true
		})
		
		input.SetText("test text")
		
		if !eventTriggered {
			t.Error("Expected change event to be triggered")
		}
	})
}

func TestInput_SetMasked(t *testing.T) {
	t.Run("enables password masking", func(t *testing.T) {
		input := NewInput("test")
		
		input.SetMasked(true, '•')
		
		if !input.Masked {
			t.Error("Expected masking to be enabled")
		}
		
		if input.MaskChar != '•' {
			t.Errorf("Expected mask character '•', got '%c'", input.MaskChar)
		}
	})
	
	t.Run("disables password masking", func(t *testing.T) {
		input := NewInput("test")
		input.Masked = true
		
		input.SetMasked(false, '*')
		
		if input.Masked {
			t.Error("Expected masking to be disabled")
		}
	})
}

func TestInput_CursorMovement(t *testing.T) {
	input := NewInput("test")
	input.Text = "hello"
	input.SetBounds(0, 0, 10, 1) // Set content area for cursor calculations
	
	t.Run("Left moves cursor left", func(t *testing.T) {
		input.Pos = 3
		
		input.Left()
		
		if input.Pos != 2 {
			t.Errorf("Expected cursor position 2, got %d", input.Pos)
		}
	})
	
	t.Run("Left at beginning has no effect", func(t *testing.T) {
		input.Pos = 0
		
		input.Left()
		
		if input.Pos != 0 {
			t.Errorf("Expected cursor position 0, got %d", input.Pos)
		}
	})
	
	t.Run("Right moves cursor right", func(t *testing.T) {
		input.Pos = 2
		
		input.Right()
		
		if input.Pos != 3 {
			t.Errorf("Expected cursor position 3, got %d", input.Pos)
		}
	})
	
	t.Run("Right at end has no effect", func(t *testing.T) {
		input.Pos = len(input.Text)
		
		input.Right()
		
		if input.Pos != len(input.Text) {
			t.Errorf("Expected cursor position %d, got %d", len(input.Text), input.Pos)
		}
	})
	
	t.Run("Start moves to beginning", func(t *testing.T) {
		input.Pos = 3
		
		input.Start()
		
		if input.Pos != 0 {
			t.Errorf("Expected cursor position 0, got %d", input.Pos)
		}
	})
	
	t.Run("End moves to end", func(t *testing.T) {
		input.Pos = 2
		
		input.End()
		
		if input.Pos != len(input.Text) {
			t.Errorf("Expected cursor position %d, got %d", len(input.Text), input.Pos)
		}
	})
}

func TestInput_Insert(t *testing.T) {
	t.Run("inserts character at cursor position", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "helo"
		input.Pos = 3
		
		input.Insert('l')
		
		if input.Text != "hello" {
			t.Errorf("Expected text 'hello', got '%s'", input.Text)
		}
		
		if input.Pos != 4 {
			t.Errorf("Expected cursor position 4, got %d", input.Pos)
		}
	})
	
	t.Run("inserts at beginning", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "ello"
		input.Pos = 0
		
		input.Insert('h')
		
		if input.Text != "hello" {
			t.Errorf("Expected text 'hello', got '%s'", input.Text)
		}
		
		if input.Pos != 1 {
			t.Errorf("Expected cursor position 1, got %d", input.Pos)
		}
	})
	
	t.Run("inserts at end", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hell"
		input.Pos = 4
		
		input.Insert('o')
		
		if input.Text != "hello" {
			t.Errorf("Expected text 'hello', got '%s'", input.Text)
		}
		
		if input.Pos != 5 {
			t.Errorf("Expected cursor position 5, got %d", input.Pos)
		}
	})
	
	t.Run("respects read-only mode", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.ReadOnly = true
		originalText := input.Text
		
		input.Insert('x')
		
		if input.Text != originalText {
			t.Errorf("Expected text unchanged in read-only mode, got '%s'", input.Text)
		}
	})
	
	t.Run("respects maximum length", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Max = 5
		input.Pos = 5
		
		input.Insert('x')
		
		if input.Text != "hello" {
			t.Errorf("Expected text unchanged when at max length, got '%s'", input.Text)
		}
	})
	
	t.Run("handles Unicode characters", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hell"
		input.Pos = 4
		
		input.Insert('ø')
		
		if input.Text != "hellø" {
			t.Errorf("Expected text 'hellø', got '%s'", input.Text)
		}
	})
	
	t.Run("triggers change event", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		eventTriggered := false
		
		input.On("change", func(w Widget, event string, data ...any) bool {
			eventTriggered = true
			return true
		})
		
		input.Insert('!')
		
		if !eventTriggered {
			t.Error("Expected change event to be triggered")
		}
	})
}

func TestInput_Delete(t *testing.T) {
	t.Run("deletes character before cursor", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 3
		
		input.Delete()
		
		if input.Text != "helo" {
			t.Errorf("Expected text 'helo', got '%s'", input.Text)
		}
		
		if input.Pos != 2 {
			t.Errorf("Expected cursor position 2, got %d", input.Pos)
		}
	})
	
	t.Run("no effect at beginning", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 0
		
		input.Delete()
		
		if input.Text != "hello" {
			t.Errorf("Expected text unchanged, got '%s'", input.Text)
		}
		
		if input.Pos != 0 {
			t.Errorf("Expected cursor position 0, got %d", input.Pos)
		}
	})
	
	t.Run("respects read-only mode", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 3
		input.ReadOnly = true
		
		input.Delete()
		
		if input.Text != "hello" {
			t.Errorf("Expected text unchanged in read-only mode, got '%s'", input.Text)
		}
	})
}

func TestInput_DeleteForward(t *testing.T) {
	t.Run("deletes character at cursor", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 2
		
		input.DeleteForward()
		
		if input.Text != "helo" {
			t.Errorf("Expected text 'helo', got '%s'", input.Text)
		}
		
		if input.Pos != 2 {
			t.Errorf("Expected cursor position unchanged at 2, got %d", input.Pos)
		}
	})
	
	t.Run("no effect at end", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 5
		
		input.DeleteForward()
		
		if input.Text != "hello" {
			t.Errorf("Expected text unchanged, got '%s'", input.Text)
		}
	})
	
	t.Run("respects read-only mode", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 2
		input.ReadOnly = true
		
		input.DeleteForward()
		
		if input.Text != "hello" {
			t.Errorf("Expected text unchanged in read-only mode, got '%s'", input.Text)
		}
	})
}

func TestInput_Clear(t *testing.T) {
	t.Run("clears all text", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello world"
		input.Pos = 5
		input.Offset = 2
		
		input.Clear()
		
		if input.Text != "" {
			t.Errorf("Expected empty text, got '%s'", input.Text)
		}
		
		if input.Pos != 0 {
			t.Errorf("Expected cursor position 0, got %d", input.Pos)
		}
		
		if input.Offset != 0 {
			t.Errorf("Expected offset 0, got %d", input.Offset)
		}
	})
	
	t.Run("respects read-only mode", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.ReadOnly = true
		
		input.Clear()
		
		if input.Text != "hello" {
			t.Errorf("Expected text unchanged in read-only mode, got '%s'", input.Text)
		}
	})
	
	t.Run("triggers change event", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		eventTriggered := false
		
		input.On("change", func(w Widget, event string, data ...any) bool {
			eventTriggered = true
			return true
		})
		
		input.Clear()
		
		if !eventTriggered {
			t.Error("Expected change event to be triggered")
		}
	})
}

func TestInput_Visible(t *testing.T) {
	t.Run("returns full text when it fits", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.SetBounds(0, 0, 10, 1) // Wide enough content area
		
		visible := input.Visible()
		
		if visible != "hello" {
			t.Errorf("Expected visible text 'hello', got '%s'", visible)
		}
	})
	
	t.Run("returns scrolled portion for long text", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello world"
		input.SetBounds(0, 0, 5, 1) // Narrow content area
		input.Offset = 6 // Scroll to show "world"
		
		visible := input.Visible()
		
		if visible != "world" {
			t.Errorf("Expected visible text 'world', got '%s'", visible)
		}
	})
	
	t.Run("returns masked text when masking enabled", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "secret"
		input.SetMasked(true, '*')
		input.SetBounds(0, 0, 10, 1)
		
		visible := input.Visible()
		
		if visible != "******" {
			t.Errorf("Expected masked text '******', got '%s'", visible)
		}
	})
	
	t.Run("returns masked scrolled portion", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "password123"
		input.SetMasked(true, '•')
		input.SetBounds(0, 0, 5, 1)
		input.Offset = 4
		
		visible := input.Visible()
		
		if visible != "•••••" {
			t.Errorf("Expected masked visible text '•••••', got '%s'", visible)
		}
	})
	
	t.Run("handles empty content area", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.SetBounds(0, 0, 0, 1) // Zero width
		
		visible := input.Visible()
		
		if visible != "" {
			t.Errorf("Expected empty visible text, got '%s'", visible)
		}
	})
}

func TestInput_Cursor(t *testing.T) {
	t.Run("returns cursor position relative to visible area", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello world"
		input.SetBounds(0, 0, 5, 1)
		input.Pos = 7
		input.Offset = 6
		
		x, y := input.Cursor()
		
		if x != 1 || y != 0 {
			t.Errorf("Expected cursor at (1,0), got (%d,%d)", x, y)
		}
	})
	
	t.Run("clamps cursor to content bounds", func(t *testing.T) {
		input := NewInput("test")
		input.SetBounds(0, 0, 5, 1)
		input.Pos = 0
		input.Offset = 5 // Cursor would be at negative position
		
		x, y := input.Cursor()
		
		if x < 0 {
			t.Errorf("Expected cursor x >= 0, got %d", x)
		}
		
		if y != 0 {
			t.Errorf("Expected cursor y = 0, got %d", y)
		}
	})
}

func TestInput_ShouldShowPlaceholder(t *testing.T) {
	t.Run("shows placeholder when text is empty", func(t *testing.T) {
		input := NewInput("test")
		input.Text = ""
		input.Placeholder = "Enter text..."
		
		if !input.ShouldShowPlaceholder() {
			t.Error("Expected placeholder to be shown for empty text")
		}
	})
	
	t.Run("hides placeholder when text is not empty", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Placeholder = "Enter text..."
		
		if input.ShouldShowPlaceholder() {
			t.Error("Expected placeholder to be hidden when text is present")
		}
	})
	
	t.Run("hides placeholder when no placeholder text", func(t *testing.T) {
		input := NewInput("test")
		input.Text = ""
		input.Placeholder = ""
		
		if input.ShouldShowPlaceholder() {
			t.Error("Expected placeholder to be hidden when no placeholder text")
		}
	})
}

func TestInput_Handle_KeyboardEvents(t *testing.T) {
	t.Run("handles arrow key navigation", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 2
		
		// Test Left arrow
		leftEvent := tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
		handled := input.Handle(leftEvent)
		
		if !handled {
			t.Error("Expected Left arrow event to be handled")
		}
		
		if input.Pos != 1 {
			t.Errorf("Expected cursor position 1 after Left, got %d", input.Pos)
		}
		
		// Test Right arrow
		rightEvent := tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
		handled = input.Handle(rightEvent)
		
		if !handled {
			t.Error("Expected Right arrow event to be handled")
		}
		
		if input.Pos != 2 {
			t.Errorf("Expected cursor position 2 after Right, got %d", input.Pos)
		}
	})
	
	t.Run("handles Home and End keys", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello world"
		input.Pos = 5
		
		// Test Home key
		homeEvent := tcell.NewEventKey(tcell.KeyHome, 0, tcell.ModNone)
		handled := input.Handle(homeEvent)
		
		if !handled {
			t.Error("Expected Home key event to be handled")
		}
		
		if input.Pos != 0 {
			t.Errorf("Expected cursor position 0 after Home, got %d", input.Pos)
		}
		
		// Test End key
		endEvent := tcell.NewEventKey(tcell.KeyEnd, 0, tcell.ModNone)
		handled = input.Handle(endEvent)
		
		if !handled {
			t.Error("Expected End key event to be handled")
		}
		
		if input.Pos != len(input.Text) {
			t.Errorf("Expected cursor position %d after End, got %d", len(input.Text), input.Pos)
		}
	})
	
	t.Run("handles Ctrl+A and Ctrl+E shortcuts", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 3
		
		// Test Ctrl+A (beginning)
		ctrlAEvent := tcell.NewEventKey(tcell.KeyCtrlA, 0, tcell.ModCtrl)
		handled := input.Handle(ctrlAEvent)
		
		if !handled {
			t.Error("Expected Ctrl+A event to be handled")
		}
		
		if input.Pos != 0 {
			t.Errorf("Expected cursor position 0 after Ctrl+A, got %d", input.Pos)
		}
		
		// Test Ctrl+E (end)
		ctrlEEvent := tcell.NewEventKey(tcell.KeyCtrlE, 0, tcell.ModCtrl)
		handled = input.Handle(ctrlEEvent)
		
		if !handled {
			t.Error("Expected Ctrl+E event to be handled")
		}
		
		if input.Pos != len(input.Text) {
			t.Errorf("Expected cursor position %d after Ctrl+E, got %d", len(input.Text), input.Pos)
		}
	})
	
	t.Run("handles Backspace and Delete keys", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.Pos = 3
		
		// Test Backspace
		backspaceEvent := tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone)
		handled := input.Handle(backspaceEvent)
		
		if !handled {
			t.Error("Expected Backspace event to be handled")
		}
		
		if input.Text != "helo" {
			t.Errorf("Expected text 'helo' after Backspace, got '%s'", input.Text)
		}
		
		if input.Pos != 2 {
			t.Errorf("Expected cursor position 2 after Backspace, got %d", input.Pos)
		}
		
		// Test Delete key
		deleteEvent := tcell.NewEventKey(tcell.KeyDelete, 0, tcell.ModNone)
		handled = input.Handle(deleteEvent)
		
		if !handled {
			t.Error("Expected Delete event to be handled")
		}
		
		if input.Text != "heo" {
			t.Errorf("Expected text 'heo' after Delete, got '%s'", input.Text)
		}
	})
	
	t.Run("handles character input", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hllo"
		input.Pos = 1
		
		// Test character insertion
		charEvent := tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone)
		handled := input.Handle(charEvent)
		
		if !handled {
			t.Error("Expected character input event to be handled")
		}
		
		if input.Text != "hello" {
			t.Errorf("Expected text 'hello' after character input, got '%s'", input.Text)
		}
		
		if input.Pos != 2 {
			t.Errorf("Expected cursor position 2 after character input, got %d", input.Pos)
		}
	})
	
	t.Run("handles Ctrl+K and Ctrl+U shortcuts", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello world"
		input.Pos = 5
		
		// Test Ctrl+K (kill to end)
		ctrlKEvent := tcell.NewEventKey(tcell.KeyCtrlK, 0, tcell.ModCtrl)
		handled := input.Handle(ctrlKEvent)
		
		if !handled {
			t.Error("Expected Ctrl+K event to be handled")
		}
		
		if input.Text != "hello" {
			t.Errorf("Expected text 'hello' after Ctrl+K, got '%s'", input.Text)
		}
		
		// Reset for Ctrl+U test
		input.Text = "hello world"
		input.Pos = 6
		
		// Test Ctrl+U (kill to beginning)
		ctrlUEvent := tcell.NewEventKey(tcell.KeyCtrlU, 0, tcell.ModCtrl)
		handled = input.Handle(ctrlUEvent)
		
		if !handled {
			t.Error("Expected Ctrl+U event to be handled")
		}
		
		if input.Text != "world" {
			t.Errorf("Expected text 'world' after Ctrl+U, got '%s'", input.Text)
		}
		
		if input.Pos != 0 {
			t.Errorf("Expected cursor position 0 after Ctrl+U, got %d", input.Pos)
		}
	})
	
	t.Run("handles Enter key", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		eventTriggered := false
		
		input.On("enter", func(w Widget, event string, data ...any) bool {
			eventTriggered = true
			if len(data) > 0 {
				if text, ok := data[0].(string); ok {
					if text != "hello" {
						t.Errorf("Expected enter event data 'hello', got '%s'", text)
					}
				}
			}
			return true
		})
		
		enterEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		handled := input.Handle(enterEvent)
		
		if !handled {
			t.Error("Expected Enter event to be handled")
		}
		
		if !eventTriggered {
			t.Error("Expected enter event to be triggered")
		}
	})
	
	t.Run("respects read-only mode for editing", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		input.ReadOnly = true
		input.Pos = 2
		
		// Character input should be ignored
		charEvent := tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone)
		handled := input.Handle(charEvent)
		
		if handled {
			t.Error("Expected character input to be ignored in read-only mode")
		}
		
		if input.Text != "hello" {
			t.Errorf("Expected text unchanged in read-only mode, got '%s'", input.Text)
		}
		
		// Navigation should still work
		leftEvent := tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
		handled = input.Handle(leftEvent)
		
		if !handled {
			t.Error("Expected navigation to work in read-only mode")
		}
		
		if input.Pos != 1 {
			t.Errorf("Expected cursor position 1 after navigation in read-only mode, got %d", input.Pos)
		}
	})
	
	t.Run("ignores non-printable characters", func(t *testing.T) {
		input := NewInput("test")
		input.Text = "hello"
		originalText := input.Text
		
		// Test control character
		ctrlCharEvent := tcell.NewEventKey(tcell.KeyRune, '\x01', tcell.ModNone)
		handled := input.Handle(ctrlCharEvent)
		
		if handled {
			t.Error("Expected non-printable character to be ignored")
		}
		
		if input.Text != originalText {
			t.Errorf("Expected text unchanged after non-printable character, got '%s'", input.Text)
		}
	})
	
	t.Run("ignores non-keyboard events", func(t *testing.T) {
		input := NewInput("test")
		
		// Test mouse event
		mouseEvent := tcell.NewEventMouse(10, 5, tcell.ButtonNone, tcell.ModNone)
		handled := input.Handle(mouseEvent)
		
		if handled {
			t.Error("Expected mouse event to be ignored")
		}
	})
}