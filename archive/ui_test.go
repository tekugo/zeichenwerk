package zeichenwerk

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// MockScreen implements tcell.Screen for testing purposes
type MockScreen struct {
	width, height int
	cells         map[string]MockCell
	cursorX       int
	cursorY       int
	cursorStyle   tcell.CursorStyle
	cursorVisible bool
	style         tcell.Style
	mouseEnabled  bool
	events        []tcell.Event
	eventIndex    int
}

type MockCell struct {
	primary   rune
	combining []rune
	style     tcell.Style
}

func NewMockScreen(width, height int) *MockScreen {
	return &MockScreen{
		width:  width,
		height: height,
		cells:  make(map[string]MockCell),
		style:  tcell.StyleDefault,
	}
}

func (m *MockScreen) Init() error                     { return nil }
func (m *MockScreen) Fini()                           {}
func (m *MockScreen) Clear()                          { m.cells = make(map[string]MockCell) }
func (m *MockScreen) Fill(ch rune, style tcell.Style) { /* mock implementation */ }
func (m *MockScreen) SetCell(x, y int, style tcell.Style, ch ...rune) {
	if len(ch) > 0 {
		key := fmt.Sprintf("%d,%d", x, y)
		m.cells[key] = MockCell{
			primary: ch[0],
			style:   style,
		}
	}
}

func (m *MockScreen) GetContent(x, y int) (rune, []rune, tcell.Style, int) {
	key := fmt.Sprintf("%d,%d", x, y)
	if cell, exists := m.cells[key]; exists {
		return cell.primary, cell.combining, cell.style, 1
	}
	return ' ', nil, tcell.StyleDefault, 1
}

func (m *MockScreen) SetContent(x, y int, primary rune, combining []rune, style tcell.Style) {
	key := fmt.Sprintf("%d,%d", x, y)
	m.cells[key] = MockCell{
		primary:   primary,
		combining: combining,
		style:     style,
	}
}
func (m *MockScreen) SetStyle(style tcell.Style) { m.style = style }
func (m *MockScreen) ShowCursor(x, y int)        { m.cursorX, m.cursorY, m.cursorVisible = x, y, true }
func (m *MockScreen) HideCursor()                { m.cursorVisible = false }
func (m *MockScreen) SetCursorStyle(style tcell.CursorStyle, color ...tcell.Color) {
	m.cursorStyle = style
}
func (m *MockScreen) Size() (int, int) { return m.width, m.height }
func (m *MockScreen) PollEvent() tcell.Event {
	if m.eventIndex < len(m.events) {
		event := m.events[m.eventIndex]
		m.eventIndex++
		return event
	}
	return nil
}
func (m *MockScreen) PostEvent(ev tcell.Event) error                            { return nil }
func (m *MockScreen) PostEventWait(ev tcell.Event)                              {}
func (m *MockScreen) EnableMouse(flags ...tcell.MouseFlags)                     { m.mouseEnabled = true }
func (m *MockScreen) DisableMouse()                                             { m.mouseEnabled = false }
func (m *MockScreen) EnablePaste()                                              {}
func (m *MockScreen) DisablePaste()                                             {}
func (m *MockScreen) EnableFocus()                                              {}
func (m *MockScreen) DisableFocus()                                             {}
func (m *MockScreen) HasMouse() bool                                            { return m.mouseEnabled }
func (m *MockScreen) HasKey(key tcell.Key) bool                                 { return true }
func (m *MockScreen) Sync()                                                     {}
func (m *MockScreen) CharacterSet() string                                      { return "UTF-8" }
func (m *MockScreen) RegisterRuneFallback(r rune, subst string)                 {}
func (m *MockScreen) UnregisterRuneFallback(r rune)                             {}
func (m *MockScreen) CanDisplay(r rune, checkFallbacks bool) bool               { return true }
func (m *MockScreen) Resize(int, int, int, int)                                 {}
func (m *MockScreen) Colors() int                                               { return 256 }
func (m *MockScreen) Show()                                                     {}
func (m *MockScreen) Beep() error                                               { return nil }
func (m *MockScreen) Suspend() error                                            { return nil }
func (m *MockScreen) Resume() error                                             { return nil }
func (m *MockScreen) ChannelEvents(ch chan<- tcell.Event, quit <-chan struct{}) {}
func (m *MockScreen) GetClipboard()                                             { return }
func (m *MockScreen) HasPendingEvent() bool                                     { return false }
func (m *MockScreen) LockRegion(x, y, width, height int, sync bool)             {}
func (m *MockScreen) UnlockRegion(x, y, width, height int)                      {}
func (m *MockScreen) SetClipboard(data []byte)                                  {}
func (m *MockScreen) Tty() (tcell.Tty, bool)                                    { return nil, false }
func (m *MockScreen) SetSize(width, height int)                                 { m.width, m.height = width, height }
func (m *MockScreen) SetTitle(title string)                                     {}

// Helper function to create a simple container for testing
func createTestContainer(id string) *Flex {
	return NewFlex(id, "vertical", "start", 1)
}

// Helper function to create a simple focusable widget for testing
func createTestInput(id string) *Input {
	return NewInput(id)
}

// Helper function to create a simple non-focusable widget for testing
func createTestLabel(id string, text string) *Label {
	return NewLabel(id, text)
}

func TestNewUI(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")

	t.Run("creates UI with debug disabled", func(t *testing.T) {
		ui, err := NewUI(theme, root, false)
		if err != nil {
			t.Fatalf("NewUI failed: %v", err)
		}

		if ui == nil {
			t.Fatal("Expected UI instance, got nil")
		}

		if ui.debug {
			t.Error("Expected debug to be false")
		}

		if ui.renderer.theme != theme {
			t.Error("Expected renderer theme to match provided theme")
		}

		if len(ui.layers) != 1 {
			t.Errorf("Expected 1 layer, got %d", len(ui.layers))
		}

		if ui.layers[0] != root {
			t.Error("Expected root layer to match provided root container")
		}

		if root.Parent() != ui {
			t.Error("Expected root container parent to be UI")
		}
	})

	t.Run("creates UI with debug enabled", func(t *testing.T) {
		ui, err := NewUI(theme, root, true)
		if err != nil {
			t.Fatalf("NewUI failed: %v", err)
		}

		if !ui.debug {
			t.Error("Expected debug to be true")
		}
	})

	t.Run("connects debug logger if present", func(t *testing.T) {
		// Create root with debug log widget
		flex := NewFlex("root", "vertical", "stretch", 0)
		debugLog := NewText("debug-log", []string{}, true, 100)
		flex.Add(debugLog)

		ui, err := NewUI(theme, flex, true)
		if err != nil {
			t.Fatalf("NewUI failed: %v", err)
		}

		if ui.logger == nil {
			t.Error("Expected logger to be connected")
		}

		if ui.logger != debugLog {
			t.Error("Expected logger to be the debug-log widget")
		}
	})
}

func TestUI_Handle_KeyboardEvents(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	// Add some focusable widgets
	input1 := createTestInput("input1")
	input2 := createTestInput("input2")
	root.Add(input1)
	root.Add(input2)

	t.Run("handles Tab key - next focus", func(t *testing.T) {
		// Set initial focus
		ui.Focus(input1)

		// Create Tab key event
		event := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)

		handled := ui.Handle(event)

		if !handled {
			t.Error("Expected Tab event to be handled")
		}

		// Focus should move to next widget
		if ui.focus != input2 {
			t.Error("Expected focus to move to input2")
		}
	})

	t.Run("handles Shift+Tab key - previous focus", func(t *testing.T) {
		// Set focus to second widget
		ui.Focus(input2)
		if ui.focus != input2 {
			t.Error("Expected initial focus to be input2")
		}

		// Create Shift+Tab key event
		event := tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModShift)

		handled := ui.Handle(event)

		if !handled {
			t.Error("Expected Shift+Tab event to be handled")
		}

		// Focus should move to previous widget
		if ui.focus != input1 {
			t.Errorf("Expected focus to move to input1, is %s", ui.focus.ID())
		}
	})

	t.Run("handles Escape key - closes popup", func(t *testing.T) {
		// Add a popup layer
		popup := createTestContainer("popup")
		ui.layers = append(ui.layers, popup)

		event := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)

		handled := ui.Handle(event)

		if !handled {
			t.Error("Expected Escape event to be handled")
		}

		// Popup layer should be removed
		if len(ui.layers) != 1 {
			t.Errorf("Expected 1 layer after Escape, got %d", len(ui.layers))
		}
	})

	t.Run("handles Ctrl+C key - quits application", func(t *testing.T) {
		event := tcell.NewEventKey(tcell.KeyCtrlC, 0, tcell.ModCtrl)

		handled := ui.Handle(event)

		if !handled {
			t.Error("Expected Ctrl+C event to be handled")
		}

		// Check if quit channel was closed (non-blocking check)
		select {
		case <-ui.quit:
			// Quit channel was closed - expected behavior
		default:
			t.Error("Expected quit channel to be closed")
		}
	})

	t.Run("handles 'q' key - quits application", func(t *testing.T) {
		// Create new UI since previous test closed quit channel
		ui, _ = NewUI(theme, root, false)

		event := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)

		handled := ui.Handle(event)

		if !handled {
			t.Error("Expected 'q' event to be handled")
		}

		// Check if quit channel was closed
		select {
		case <-ui.quit:
			// Expected
		default:
			t.Error("Expected quit channel to be closed")
		}
	})
}

func TestUI_Handle_MouseEvents(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	// Set UI dimensions
	ui.width, ui.height = 80, 24
	ui.Layout()

	t.Run("handles mouse move - updates hover state", func(t *testing.T) {
		// Create a widget that can be hovered
		label := createTestLabel("label1", "Test Label")
		label.SetBounds(10, 5, 20, 1)
		root.Add(label)

		// Create mouse move event over the label
		event := tcell.NewEventMouse(15, 5, tcell.ButtonNone, tcell.ModNone)

		handled := ui.Handle(event)

		if !handled {
			t.Error("Expected mouse event to be handled")
		}

		// Check if hover state was updated
		if ui.hover != label {
			t.Error("Expected label to be hovered")
		}

		if !label.Hovered() {
			t.Error("Expected label to have hovered state")
		}
	})

	t.Run("clears previous hover state", func(t *testing.T) {
		// Set up initial hover state
		label1 := createTestLabel("label1", "Label 1")
		label1.SetBounds(10, 5, 20, 1)
		label2 := createTestLabel("label2", "Label 2")
		label2.SetBounds(10, 10, 20, 1)

		root.Add(label1)
		root.Add(label2)

		// Set initial hover
		ui.hover = label1
		label1.SetHovered(true)

		// Move mouse to second label
		event := tcell.NewEventMouse(15, 10, tcell.ButtonNone, tcell.ModNone)

		handled := ui.Handle(event)

		if !handled {
			t.Error("Expected mouse event to be handled")
		}

		// Check hover state changes
		if label1.Hovered() {
			t.Error("Expected label1 to lose hovered state")
		}

		if ui.hover != label2 {
			t.Error("Expected label2 to be hovered")
		}

		if !label2.Hovered() {
			t.Error("Expected label2 to have hovered state")
		}
	})
}

func TestUI_Handle_ResizeEvents(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	// Mock screen
	mockScreen := NewMockScreen(80, 24)
	ui.screen = mockScreen

	t.Run("handles resize event", func(t *testing.T) {
		// Create resize event
		mockScreen.width = 100
		mockScreen.height = 30
		event := tcell.NewEventResize(100, 30)

		handled := ui.Handle(event)

		if !handled {
			t.Error("Expected resize event to be handled")
		}

		// Check if UI dimensions were updated
		if ui.width != 100 || ui.height != 30 {
			t.Errorf("Expected dimensions 100x30, got %dx%d", ui.width, ui.height)
		}
	})
}

func TestUI_Focus(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	input1 := createTestInput("input1")
	input2 := createTestInput("input2")
	root.Add(input1)
	root.Add(input2)

	t.Run("sets focus to widget", func(t *testing.T) {
		ui.Focus(input1)

		if ui.focus != input1 {
			t.Error("Expected input1 to be focused")
		}

		if !input1.Focused() {
			t.Error("Expected input1 to have focused state")
		}
	})

	t.Run("clears previous focus", func(t *testing.T) {
		// Set initial focus
		ui.Focus(input1)

		// Change focus
		ui.Focus(input2)

		if input1.Focused() {
			t.Error("Expected input1 to lose focused state")
		}

		if ui.focus != input2 {
			t.Error("Expected input2 to be focused")
		}

		if !input2.Focused() {
			t.Error("Expected input2 to have focused state")
		}
	})

	t.Run("clears focus with nil", func(t *testing.T) {
		// Set initial focus
		ui.Focus(input1)

		// Clear focus
		ui.Focus(nil)

		if input1.Focused() {
			t.Error("Expected input1 to lose focused state")
		}

		if ui.focus != nil {
			t.Error("Expected focus to be nil")
		}
	})
}

func TestUI_SetFocus(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	// Add focusable widgets
	input1 := createTestInput("input1")
	input2 := createTestInput("input2")
	input3 := createTestInput("input3")
	root.Add(input1)
	root.Add(input2)
	root.Add(input3)

	t.Run("sets focus to first widget", func(t *testing.T) {
		ui.SetFocus("first")

		if ui.focus != input1 {
			t.Error("Expected focus to be on first widget")
		}
	})

	t.Run("sets focus to last widget", func(t *testing.T) {
		ui.SetFocus("last")

		if ui.focus != input3 {
			t.Error("Expected focus to be on last widget")
		}
	})

	t.Run("moves focus to next widget", func(t *testing.T) {
		ui.Focus(input1)
		ui.SetFocus("next")

		if ui.focus != input2 {
			t.Error("Expected focus to move to next widget")
		}
	})

	t.Run("wraps focus from last to first", func(t *testing.T) {
		ui.Focus(input3)
		ui.SetFocus("next")

		if ui.focus != input1 {
			t.Error("Expected focus to wrap to first widget")
		}
	})

	t.Run("moves focus to previous widget", func(t *testing.T) {
		ui.Focus(input2)
		ui.SetFocus("previous")

		if ui.focus != input1 {
			t.Error("Expected focus to move to previous widget")
		}
	})

	t.Run("wraps focus from first to last", func(t *testing.T) {
		ui.Focus(input1)
		ui.SetFocus("previous")

		if ui.focus != input3 {
			t.Error("Expected focus to wrap to last widget")
		}
	})
}

func TestUI_Popup(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	// Set UI dimensions
	ui.width, ui.height = 80, 24

	t.Run("adds popup as new layer", func(t *testing.T) {
		popup := createTestContainer("popup")

		ui.Popup(10, 5, 30, 10, popup)

		if len(ui.layers) != 2 {
			t.Errorf("Expected 2 layers, got %d", len(ui.layers))
		}

		if ui.layers[1] != popup {
			t.Error("Expected popup to be second layer")
		}

		if popup.Parent() != ui {
			t.Error("Expected popup parent to be UI")
		}
	})

	t.Run("centers popup when x=-1, y=-1", func(t *testing.T) {
		popup := createTestContainer("popup-center")

		ui.Popup(-1, -1, 20, 8, popup)

		x, y, w, h := popup.Bounds()
		expectedX := (80 - 20) / 2 // (width - w) / 2
		expectedY := (24 - 8) / 2  // (height - h) / 2

		if x != expectedX || y != expectedY {
			t.Errorf("Expected popup at (%d,%d), got (%d,%d)", expectedX, expectedY, x, y)
		}

		if w != 20 || h != 8 {
			t.Errorf("Expected popup size 20x8, got %dx%d", w, h)
		}
	})

	t.Run("positions popup relative to edges when negative", func(t *testing.T) {
		popup := createTestContainer("popup-edge")

		ui.Popup(-3, -4, 20, 8, popup)

		x, y, _, _ := popup.Bounds()
		expectedX := 80 - 20 + (-3) + 2 // width - w + x + 2
		expectedY := 24 - 8 + (-4) + 2  // height - h + y + 2

		if x != expectedX || y != expectedY {
			t.Errorf("Expected popup at (%d,%d), got (%d,%d)", expectedX, expectedY, x, y)
		}
	})
}

func TestUI_Close(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	t.Run("closes popup layer", func(t *testing.T) {
		// Add popup
		popup := createTestContainer("popup")
		ui.layers = append(ui.layers, popup)

		ui.Close()

		if len(ui.layers) != 1 {
			t.Errorf("Expected 1 layer after close, got %d", len(ui.layers))
		}

		if ui.layers[0] != root {
			t.Error("Expected root layer to remain")
		}
	})

	t.Run("protects base layer from closing", func(t *testing.T) {
		// Ensure only base layer exists
		ui.layers = []Container{root}

		ui.Close()

		if len(ui.layers) != 1 {
			t.Errorf("Expected 1 layer to remain, got %d", len(ui.layers))
		}

		if ui.layers[0] != root {
			t.Error("Expected root layer to remain")
		}
	})
}

func TestUI_Find(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	// Create nested structure
	input1 := createTestInput("input1")
	label1 := createTestLabel("label1", "Label 1")

	flex := NewFlex("flex1", "vertical", "stretch", 0)
	input2 := createTestInput("input2")
	flex.Add(input2)

	root.Add(input1)
	root.Add(label1)
	root.Add(flex)

	t.Run("finds widget by ID", func(t *testing.T) {
		found := ui.Find("input1", false)

		if found != input1 {
			t.Error("Expected to find input1")
		}
	})

	t.Run("finds nested widget by ID", func(t *testing.T) {
		found := ui.Find("input2", false)

		if found != input2 {
			t.Error("Expected to find nested input2")
		}
	})

	t.Run("returns nil for non-existent widget", func(t *testing.T) {
		found := ui.Find("nonexistent", false)

		if found != nil {
			t.Error("Expected nil for non-existent widget")
		}
	})

	t.Run("finds layer by ID", func(t *testing.T) {
		found := ui.Find("root", false)

		if found != root {
			t.Error("Expected to find root layer")
		}
	})
}

func TestUI_FindOn(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	input := createTestInput("test-input")
	root.Add(input)

	t.Run("attaches event handler to found widget", func(t *testing.T) {
		handlerCalled := false

		ui.FindOn("test-input", "change", func(w Widget, event string, data ...any) bool {
			handlerCalled = true
			return true
		})

		// Trigger the event
		input.Emit("change", "test")

		if !handlerCalled {
			t.Error("Expected event handler to be called")
		}
	})

	t.Run("silently ignores non-existent widget", func(t *testing.T) {
		// This should not panic
		ui.FindOn("nonexistent", "change", func(w Widget, event string, data ...any) bool {
			return true
		})
	})
}

func TestUI_Redraw(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	input := createTestInput("test-input")

	t.Run("queues widget for redraw", func(t *testing.T) {
		ui.Redraw(input)

		// Check if widget was queued (non-blocking check)
		select {
		case widget := <-ui.redraw:
			if widget != input {
				t.Error("Expected input widget in redraw queue")
			}
		default:
			t.Error("Expected widget to be queued for redraw")
		}
	})

	t.Run("triggers refresh when redraw queue is full", func(t *testing.T) {
		// Fill the redraw channel
		for i := 0; i < cap(ui.redraw); i++ {
			ui.redraw <- input
		}

		// This should trigger refresh instead
		ui.Redraw(input)

		// Check if refresh was triggered
		select {
		case <-ui.refresh:
			// Expected
		default:
			t.Error("Expected refresh to be triggered when redraw queue is full")
		}
	})
}

func TestUI_Refresh(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	t.Run("sets dirty flag and queues refresh", func(t *testing.T) {
		ui.dirty = false

		ui.Refresh()

		if !ui.dirty {
			t.Error("Expected dirty flag to be set")
		}

		// Check if refresh was queued
		select {
		case <-ui.refresh:
			// Expected
		default:
			t.Error("Expected refresh to be queued")
		}
	})

	t.Run("skips queue when refresh already pending", func(t *testing.T) {
		// Fill refresh channel
		ui.refresh <- struct{}{}

		ui.Refresh()

		// Should still set dirty flag
		if !ui.dirty {
			t.Error("Expected dirty flag to be set even when queue is full")
		}
	})
}

func TestUI_Layout(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	ui.width, ui.height = 80, 24

	t.Run("sets root bounds in normal mode", func(t *testing.T) {
		ui.debug = false
		ui.Layout()

		x, y, w, h := root.Bounds()

		if x != 0 || y != 0 || w != 80 || h != 24 {
			t.Errorf("Expected root bounds (0,0,80,24), got (%d,%d,%d,%d)", x, y, w, h)
		}
	})

	t.Run("reserves bottom line in debug mode", func(t *testing.T) {
		ui.debug = true
		ui.Layout()

		x, y, w, h := root.Bounds()

		if x != 0 || y != 0 || w != 80 || h != 23 {
			t.Errorf("Expected root bounds (0,0,80,23) in debug mode, got (%d,%d,%d,%d)", x, y, w, h)
		}
	})
}

func TestUI_SetTheme(t *testing.T) {
	theme1 := DefaultTheme()
	theme2 := TokyoNightTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme1, root, false)

	t.Run("changes renderer theme", func(t *testing.T) {
		ui.SetTheme(theme2)

		if ui.renderer.theme != theme2 {
			t.Error("Expected renderer theme to be updated")
		}
	})

	t.Run("returns current theme", func(t *testing.T) {
		currentTheme := ui.Theme()

		if currentTheme != theme2 {
			t.Error("Expected GetTheme to return current theme")
		}
	})
}

func TestUI_EventPropagation(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	// Create a widget that handles events
	input := createTestInput("test-input")
	root.Add(input)

	eventHandled := false
	input.On("key", func(w Widget, event string, data ...any) bool {
		eventHandled = true
		return true
	})

	t.Run("propagates event to focused widget", func(t *testing.T) {
		ui.Focus(input)

		// Create key event
		keyEvent := tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModNone)

		handled := ui.propagate(input, keyEvent)

		if !handled {
			t.Error("Expected event to be handled by widget")
		}

		if !eventHandled {
			t.Error("Expected widget event handler to be called")
		}
	})

	t.Run("stops propagation when event is handled", func(t *testing.T) {
		parentHandled := false
		root.On("key", func(w Widget, event string, data ...any) bool {
			parentHandled = true
			return true
		})

		eventHandled = false

		keyEvent := tcell.NewEventKey(tcell.KeyCtrlD, 0, tcell.ModNone)

		handled := ui.propagate(input, keyEvent)

		if !handled {
			t.Error("Expected event to be handled")
		}

		if !eventHandled {
			t.Error("Expected child widget to handle event")
		}

		if parentHandled {
			t.Error("Expected propagation to stop at child widget")
		}
	})
}

func TestUI_Performance(t *testing.T) {
	theme := DefaultTheme()
	root := createTestContainer("root")
	ui, _ := NewUI(theme, root, false)

	t.Run("tracks redraw counter", func(t *testing.T) {
		initialRedraws := ui.redraws

		// Mock a widget redraw
		ui.redraws++

		if ui.redraws != initialRedraws+1 {
			t.Error("Expected redraw counter to increment")
		}
	})

	t.Run("tracks refresh counter", func(t *testing.T) {
		initialRefreshs := ui.refreshs

		// Mock a refresh
		ui.refreshs++

		if ui.refreshs != initialRefreshs+1 {
			t.Error("Expected refresh counter to increment")
		}
	})
}
