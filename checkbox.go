package zeichenwerk

import (
	"fmt"

	"github.com/gdamore/tcell/v3"
)

// Checkbox represents a boolean input widget that can be toggled between
// checked and unchecked states. It displays a checkbox indicator with
// optional label text and responds to keyboard and mouse input.
type Checkbox struct {
	Component
	text string // The label text displayed next to the checkbox
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
func NewCheckbox(id, class, text string, checked bool) *Checkbox {
	checkbox := &Checkbox{
		Component: Component{id: id, class: class},
		text:      text,
	}
	checkbox.SetHint(len(text)+4, 1)
	checkbox.SetFlag(FlagFocusable, true)
	checkbox.SetFlag(FlagChecked, checked)
	checkbox.SetFlag(FlagReadonly, false)
	OnKey(checkbox, checkbox.handleKey)
	OnMouse(checkbox, checkbox.handleMouse)
	return checkbox
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies a theme style to the component.
func (c *Checkbox) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("checkbox"), "checked", "disabled", "focused", "hovered")
}

// Refresh redraws the widget.
func (c *Checkbox) Refresh() {
	Redraw(c)
}

// ---- Setter ---------------------------------------------------------------

// Set set's the checkbox value in a generic way.
func (c *Checkbox) Set(value bool) {
	c.SetFlag(FlagChecked, value)
	c.Refresh()
}

// Label returns the checkbox's label text.
func (c *Checkbox) Label() string { return c.text }

// ---- Summarizer -----------------------------------------------------------

// Summary returns label and checked state for Dump output.
func (c *Checkbox) Summary() string {
	return fmt.Sprintf("%q checked=%v", c.text, c.Flag(FlagChecked))
}

// ---- Checkbox Methods -----------------------------------------------------

// Toggle switches the checkbox state between checked and unchecked.
// This method triggers the "change" event with the new state as data.
// The checkbox will be refreshed to reflect the visual state change.
func (c *Checkbox) Toggle() {
	if c.Flag(FlagReadonly) {
		return
	}
	c.SetFlag(FlagChecked, !c.Flag(FlagChecked))
	c.Dispatch(c, EvtChange, c.Flag(FlagChecked))
	c.Refresh()
}

// ---- Internal Event Handling ----------------------------------------------

// handleKey processes keyboard input for the checkbox widget.
// This method handles the standard checkbox activation keys.
//
// Supported keys:
//   - Space: Toggles the checkbox state (standard checkbox behavior)
//   - Enter: Toggles the checkbox state
func (c *Checkbox) handleKey(event *tcell.EventKey) bool {
	if c.Flag(FlagReadonly) {
		return false
	}
	switch event.Key() {
	case tcell.KeyEnter:
		c.Toggle()
		return true
	case tcell.KeyRune:
		// Handle space bar activation (standard checkbox behavior)
		if event.Str() == " " {
			c.Toggle()
			return true
		}
	}
	return false
}

// handleMouse processes mouse input for the checkbox widget.
// This method implements standard checkbox mouse interaction patterns.
//
// Mouse interaction behavior:
//   - Left button click: Toggles checkbox state
//   - Bounds checking: Only responds to clicks within checkbox area
func (c *Checkbox) handleMouse(event *tcell.EventMouse) bool {
	if c.Flag(FlagReadonly) {
		return false
	}
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

// ---- Rendering ------------------------------------------------------------

func (c *Checkbox) Render(r *Renderer) {
	x, y, w, h := c.Content()
	if h < 1 || w < 1 {
		return
	}

	// Render the general component
	c.Component.Render(r)

	// Determine checkbox indicator based on state
	var indicator string
	if c.Flag(FlagChecked) {
		indicator = "[x]"
	} else {
		indicator = "[ ]"
	}

	// Render checkbox indicator (takes 3 characters)
	if w >= 3 {
		r.Text(x, y, indicator, 3)
	}

	// Render label text after the checkbox and a space
	if w > 4 && c.text != "" {
		labelX := x + len(indicator) // Position after "[x] "
		labelWidth := w - len(indicator)
		r.Text(labelX, y, c.text, labelWidth)
	}
}
