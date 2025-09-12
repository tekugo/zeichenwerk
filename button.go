package zeichenwerk

import (
	"github.com/gdamore/tcell/v2"
)

// Button represents a clickable button widget that responds to keyboard and mouse input.
// It displays text and can trigger actions when activated through various input methods.
//
// Features:
//   - Keyboard activation (Enter key and Space bar)
//   - Mouse click support with visual feedback
//   - Focus and hover state management
//   - Customizable action handlers
//   - State-based styling support (normal, focus, hover, pressed, disabled)
//
// The button supports multiple activation methods and provides visual feedback
// for different interaction states to enhance user experience.
type Button struct {
	BaseWidget
	Text  string // The text displayed on the button
	state string // Current button state (pressed, disabled, etc.)
}

// NewButton creates a new button widget with the specified ID and text.
// The button is initialized in the default state with no action handler.
// You should set the Action field to define what happens when the button is clicked.
//
// Parameters:
//   - id: Unique identifier for the button widget
//   - text: The text to display on the button
//
// Returns:
//   - *Button: A new button widget instance
//
// Example usage:
//
//	button := NewButton("ok-btn", "OK")
//	button.Action = func(b *Button) { fmt.Println("Button clicked!") }
func NewButton(id, text string) *Button {
	return &Button{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Text:       text,
		state:      "",
	}
}

// Click programmatically triggers the button's action handler.
// This method executes the Action callback if one is set. If no Action
// is defined, calling this method has no effect.
//
// This method can be called programmatically to simulate a button click
// or can be used internally when the button is activated through user input.
// The method also emits a "click" event for any registered event handlers.
func (b *Button) Click() {
	b.Emit("click")
}

// Cursor returns the cursor position for the button widget.
// Buttons don't typically display a cursor, so this always returns (-1, -1)
// to indicate that no cursor should be shown.
//
// Returns:
//   - int: x-coordinate (always -1)
//   - int: y-coordinate (always -1)
func (b *Button) Cursor() (int, int) {
	return -1, -1
}

func (b *Button) Emit(event string, data ...any) {
	if b.handlers == nil {
		return
	}
	handler, found := b.handlers[event]
	if found {
		handler(b, event, data...)
	}
}

// Handle processes keyboard and mouse events for the button widget.
// The button responds to various input methods and provides appropriate
// visual feedback and action triggering.
//
// Supported events:
//   - Keyboard: Enter key and Space bar trigger the button action
//   - Mouse: Left click with press/release cycle and bounds checking
//   - State management: Handles pressed, hover, and disabled states
//
// Parameters:
//   - event: The tcell.Event to process (keyboard or mouse)
//
// Returns:
//   - bool: true if the event was handled, false otherwise
func (b *Button) Handle(event tcell.Event) bool {
	if b.state == "disabled" {
		return false
	}

	switch event := event.(type) {
	case *tcell.EventKey:
		return b.handleKeyEvent(event)
	case *tcell.EventMouse:
		return b.handleMouseEvent(event)
	}

	return false
}

// handleKeyEvent processes keyboard input for the button widget.
// This method handles the standard button activation keys and provides
// immediate feedback by triggering the button action.
//
// Supported keys:
//   - Enter: Activates the button immediately
//   - Space: Activates the button immediately (standard button behavior)
//
// Parameters:
//   - event: The keyboard event to process
//
// Returns:
//   - bool: true if the key was handled, false otherwise
func (b *Button) handleKeyEvent(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyEnter:
		b.Click()
		return true
	case tcell.KeyRune:
		// Handle space bar activation (standard button behavior)
		if event.Rune() == ' ' {
			b.Click()
			return true
		}
	}

	return false
}

// handleMouseEvent processes mouse input for the button widget.
// This method implements standard button mouse interaction patterns
// including press/release cycles and bounds checking for proper UX.
//
// Mouse interaction behavior:
//   - Left button press: Sets button to "pressed" state
//   - Left button release: Triggers action if released within bounds
//   - Mouse movement: Cancels press state if moved outside bounds
//   - Bounds checking: Only responds to clicks within button area
//
// Parameters:
//   - event: The mouse event to process
//
// Returns:
//   - bool: true if the mouse event was handled, false otherwise
func (b *Button) handleMouseEvent(event *tcell.EventMouse) bool {
	x, y := event.Position()
	bx, by, bw, bh := b.Bounds()

	// Check if mouse is within button bounds
	if x >= bx && x < bx+bw && y >= by && y < by+bh {
		switch event.Buttons() {
		case tcell.Button1: // Left mouse button
			if b.state != "pressed" {
				b.state = "pressed"
			}
			return true
		case tcell.ButtonNone: // Mouse release
			if b.state == "pressed" {
				b.state = ""
				b.Click() // Trigger click on release
				return true
			}
		}
	} else if b.state == "pressed" {
		// Mouse moved outside button while pressed
		b.state = ""
	}

	return false
}

// Info returns a human-readable description of the button's current state.
// This includes the button's position, dimensions, text content, and current state.
// This method is primarily used for debugging and development purposes.
//
// Returns:
//   - string: Formatted string with button information
func (b *Button) Info() string {
	return "button [" + b.BaseWidget.Info() + "]"
}

// SetEnabled sets the enabled/disabled state of the button.
// When disabled, the button will not respond to user input and will
// display using the "disabled" style. When enabled, the button returns
// to normal interactive behavior.
//
// Parameters:
//   - enabled: true to enable the button, false to disable it
func (b *Button) SetEnabled(enabled bool) {
	if enabled {
		if b.state == "disabled" {
			b.state = ""
		}
	} else {
		b.state = "disabled"
	}
}

// IsEnabled returns whether the button is currently enabled.
// A disabled button will not respond to user input and displays
// using the "disabled" visual style.
//
// Returns:
//   - bool: true if the button is enabled, false if disabled
func (b *Button) IsEnabled() bool {
	return b.state != "disabled"
}

// State returns the current state of the button for styling purposes.
// The state determines which visual style should be applied to the button
// and follows a priority order for multiple simultaneous states.
//
// State priority (highest to lowest):
//  1. "focus" - when the button has keyboard focus
//  2. "hover" - when the mouse is over the button
//  3. Internal state - "pressed", "disabled", or "" (default)
//
// Available states:
//   - "": default/normal state
//   - "focus": button has keyboard focus
//   - "hover": mouse cursor is over the button
//   - "pressed": button is currently being pressed
//   - "disabled": button is disabled and non-interactive
//
// Returns:
//   - string: The current button state identifier for styling
func (b *Button) State() string {
	if b.focused {
		return "focus"
	} else if b.hovered {
		return "hover"
	} else {
		return b.state
	}
}
