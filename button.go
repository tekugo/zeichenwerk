package zeichenwerk

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

// Button represents a clickable button widget that responds to keyboard and
// mouse input. It displays text and can trigger actions when activated through
// various input methods.
//
// Features:
//   - Keyboard activation (Enter key and Space bar)
//   - Mouse click support with state update (pressed)
//   - Customizable action handlers
type Button struct {
	Component
	text string // The text displayed on the button
}

// NewButton creates a new button widget with the specified ID and text.
// The button is initialized in the default state with no action handler.
// An action handler should be registered for the "click" event with
// On("click", ...)
//
// Parameters:
//   - id: Unique identifier for the button widget
//   - text: The text to display on the button
//
// Returns:
//   - *Button: A new button widget instance
func NewButton(id, class, text string) *Button {
	button := &Button{
		Component: Component{id: id, class: class, hwidth: utf8.RuneCountInString(text), hheight: 1},
		text:      text,
	}
	button.SetFlag(FlagFocusable, true)
	OnKey(button, button.handleKey)
	OnMouse(button, button.handleMouse)
	return button
}

// ---- Widget Methods -------------------------------------------------------

// Activate programmatically triggers the button's action handler.
func (b *Button) Activate() {
	b.Dispatch(b, EvtActivate, 0)
}

// Apply applies a theme style to the component.
func (b *Button) Apply(theme *Theme) {
	theme.Apply(b, b.Selector("button"), "disabled", "focused", "hovered", "pressed")
}

func (b *Button) Refresh() {
	Redraw(b)
}

// Render implements the Widget interface for rendering the button.
func (b *Button) Render(r *Renderer) {
	b.Component.Render(r)
	x, y, w, _ := b.Content()
	r.Text(x, y, b.text, w)
}

// Set sets the button text. This is a generic method to allow
// using the Setter interface.
func (b *Button) Set(value any) bool {
	if text, ok := value.(string); ok {
		b.text = text
		return true
	} else {
		return false
	}
}

// ---- Internal methods -----------------------------------------------------

// handleKey processes keyboard input for the button widget.
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
func (b *Button) handleKey(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyEnter:
		b.Activate()
		return true
	case tcell.KeyRune:
		// Handle space bar activation (standard button behavior)
		if event.Str() == " " {
			b.Activate()
			return true
		}
	}

	return false
}

// handleMouse processes mouse input for the button widget.
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
func (b *Button) handleMouse(event *tcell.EventMouse) bool {
	x, y := event.Position()
	bx, by, bw, bh := b.Bounds()

	// Check if mouse is within button bounds
	if x >= bx && x < bx+bw && y >= by && y < by+bh {
		switch event.Buttons() {
		case tcell.Button1: // Left mouse button
			b.SetFlag(FlagPressed, true)
			return true
		case tcell.ButtonNone: // Mouse release
			if b.Flag(FlagPressed) {
				b.SetFlag(FlagPressed, false)
				b.Activate() // Trigger click on release
				return true
			}
		}
	} else if b.Flag(FlagPressed) {
		// Mouse moved outside button while pressed
		b.SetFlag(FlagPressed, false)
	}

	return false
}
