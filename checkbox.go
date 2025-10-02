package zeichenwerk

import (
	"github.com/gdamore/tcell/v2"
)

// Checkbox represents a boolean input widget that can be toggled between
// checked and unchecked states. It displays a checkbox indicator with
// optional label text and responds to keyboard and mouse input.
//
// Features:
//   - Boolean checked/unchecked state
//   - Keyboard activation (Space bar and Enter key)
//   - Mouse click support
//   - Focus and hover state management
//   - Customizable label text
//   - State-based styling support (normal, focus, hover, disabled)
//
// The checkbox provides a simple way to capture boolean user input in forms
// and settings interfaces.
type Checkbox struct {
	BaseWidget
	Text    string // The label text displayed next to the checkbox
	Checked bool   // Current checked state of the checkbox
	state   string // Current checkbox state (disabled, etc.)
}

// NewCheckbox creates a new checkbox widget with the specified ID, label
// text, and initial checked state. The checkbox is initialized as focusable
// and responds to user input.
//
// Parameters:
//   - id: Unique identifier for the checkbox widget
//   - text: The label text to display next to the checkbox
//   - checked: Initial checked state (true for checked, false for unchecked)
//
// Returns:
//   - *Checkbox: A new checkbox widget instance
//
// Example usage:
//
//	checkbox := NewCheckbox("remember-me", "Remember me", false)
//	checkbox.On("change", func(w Widget, event string, data ...any) bool {
//		fmt.Printf("Checkbox is now: %v\n", data[0].(bool))
//		return true
//	})
func NewCheckbox(id, text string, checked bool) *Checkbox {
	return &Checkbox{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Text:       text,
		Checked:    checked,
		state:      "",
	}
}

// Refresh redraws the widget.
func (c *Checkbox) Refresh() {
	Redraw(c)
}

// Toggle switches the checkbox state between checked and unchecked.
// This method triggers the "change" event with the new state as data.
// The checkbox will be refreshed to reflect the visual state change.
func (c *Checkbox) Toggle() {
	c.Checked = !c.Checked
	c.Emit("change", c.Checked)
	c.Refresh()
}

// Cursor returns the cursor position for the checkbox widget.
// Checkboxes don't typically display a cursor, so this always returns
// (-1, -1) to indicate that no cursor should be shown.
//
// Returns:
//   - int: x-coordinate (always -1)
//   - int: y-coordinate (always -1)
func (c *Checkbox) Cursor() (int, int) {
	return -1, -1
}

// Emit triggers a custom event on the widget, calling any registered event
// handlers. This function is the core of the widget's event system, allowing
// widgets to notify listeners about state changes, user interactions, or
// other significant events.
//
// The function performs the following operations:
//   - Checks if any event handlers are registered for this widget
//   - Looks up the specific handler for the given event name
//   - Calls the handler function with the widget, event name, and any
//     optional data
//   - Silently ignores the event if no handler is registered
//
// Event System Overview:
// The Emit function works in conjunction with the On() method to provide a
// flexible event-driven architecture. Widgets can emit events at any time,
// and external code can register handlers using On() to respond to these events.
//
// Common widget events include:
//   - "change": Content or state has been modified
//   - "focus": Widget has gained keyboard focus
//   - "blur": Widget has lost keyboard focus
//   - "select": Item selection has changed (for lists, menus)
//   - "activate": Item has been activated (Enter key, double-click)
//   - "key": Raw keyboard event (automatically emitted by Handle())
//
// Parameters:
//   - event: The name of the event to emit (e.g., "change", "focus", "select")
//   - data: Optional additional data to pass to the event handler. The type
//     and meaning of this data depends on the specific event and widget type.
func (c *Checkbox) Emit(event string, data ...any) bool {
	if c.handlers == nil {
		return false
	}
	handler, found := c.handlers[event]
	if found {
		return handler(c, event, data...)
	} else {
		return false
	}
}

// Handle processes keyboard and mouse events for the checkbox widget.
// The checkbox responds to space bar, enter key, and mouse clicks for
// toggling.
//
// Supported events:
//   - Keyboard: Space bar and Enter key toggle the checkbox state
//   - Mouse: Left click toggles the checkbox state
//   - State management: Handles disabled state
//
// Parameters:
//   - event: The tcell.Event to process (keyboard or mouse)
//
// Returns:
//   - bool: true if the event was handled, false otherwise
func (c *Checkbox) Handle(event tcell.Event) bool {
	if c.state == "disabled" {
		return false
	}

	switch event := event.(type) {
	case *tcell.EventKey:
		return c.handleKeyEvent(event)
	case *tcell.EventMouse:
		return c.handleMouseEvent(event)
	}

	return false
}

// handleKeyEvent processes keyboard input for the checkbox widget.
// This method handles the standard checkbox activation keys.
//
// Supported keys:
//   - Space: Toggles the checkbox state (standard checkbox behavior)
//   - Enter: Toggles the checkbox state
//
// Parameters:
//   - event: The keyboard event to process
//
// Returns:
//   - bool: true if the key was handled, false otherwise
func (c *Checkbox) handleKeyEvent(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyEnter:
		c.Toggle()
		return true
	case tcell.KeyRune:
		// Handle space bar activation (standard checkbox behavior)
		if event.Rune() == ' ' {
			c.Toggle()
			return true
		} else {
			c.Emit("key", event)
		}
	default:
		return c.Emit("key", event)
	}

	return false
}

// handleMouseEvent processes mouse input for the checkbox widget.
// This method implements standard checkbox mouse interaction patterns.
//
// Mouse interaction behavior:
//   - Left button click: Toggles checkbox state
//   - Bounds checking: Only responds to clicks within checkbox area
//
// Parameters:
//   - event: The mouse event to process
//
// Returns:
//   - bool: true if the mouse event was handled, false otherwise
func (c *Checkbox) handleMouseEvent(event *tcell.EventMouse) bool {
	x, y := event.Position()
	bx, by, bw, bh := c.Bounds()

	// Check if mouse is within checkbox bounds
	if x >= bx && x < bx+bw && y >= by && y < by+bh {
		switch event.Buttons() {
		case tcell.Button1: // Left mouse button
			c.Toggle()
			return true
		}
	}

	return false
}

// Info returns a human-readable description of the checkbox's current state.
// This includes the checkbox's position, dimensions, text content, and checked state.
// This method is primarily used for debugging and development purposes.
//
// Returns:
//   - string: Formatted string with checkbox information
func (c *Checkbox) Info() string {
	return "checkbox [" + c.BaseWidget.Info() + "]"
}

// SetEnabled sets the enabled/disabled state of the checkbox.
// When disabled, the checkbox will not respond to user input and will
// display using the "disabled" style. When enabled, the checkbox returns
// to normal interactive behavior.
//
// Parameters:
//   - enabled: true to enable the checkbox, false to disable it
func (c *Checkbox) SetEnabled(enabled bool) {
	if enabled {
		if c.state == "disabled" {
			c.state = ""
		}
	} else {
		c.state = "disabled"
	}
}

// IsEnabled returns whether the checkbox is currently enabled.
// A disabled checkbox will not respond to user input and displays
// using the "disabled" visual style.
//
// Returns:
//   - bool: true if the checkbox is enabled, false if disabled
func (c *Checkbox) IsEnabled() bool {
	return c.state != "disabled"
}

// State returns the current state of the checkbox for styling purposes.
// The state determines which visual style should be applied to the checkbox
// and follows a priority order for multiple simultaneous states.
//
// State priority (highest to lowest):
//  1. "focus" - when the checkbox has keyboard focus
//  2. "hover" - when the mouse is over the checkbox
//  3. Internal state - "disabled" or "" (default)
//
// Available states:
//   - "": default/normal state
//   - "focus": checkbox has keyboard focus
//   - "hover": mouse cursor is over the checkbox
//   - "disabled": checkbox is disabled and non-interactive
//
// Returns:
//   - string: The current checkbox state identifier for styling
func (c *Checkbox) State() string {
	if c.focused {
		return "focus"
	} else if c.hovered {
		return "hover"
	} else {
		return c.state
	}
}
