package zeichenwerk

import "github.com/gdamore/tcell/v2"

// Custom represents a widget with user-defined rendering and event handling
// behavior. It provides a flexible foundation for creating specialized
// widgets that don't fit the standard widget patterns, allowing complete
// control over visual appearance and interaction logic.
//
// Features:
//   - Custom rendering function for complete visual control
//   - Custom event handling function for specialized interactions
//   - Full access to the underlying screen drawing primitives
//   - Integration with the standard widget hierarchy and layout system
//   - Configurable focus behavior (focusable or non-focusable)
//
// Use cases:
//   - Creating specialized visual elements (charts, graphs, diagrams)
//   - Implementing unique interaction patterns not covered by standard widgets
//   - Prototyping new widget concepts before full implementation
//   - Integrating external drawing libraries or custom graphics
//   - Creating application-specific UI components
//
// The Custom widget bridges the gap between the high-level widget system and
// low-level terminal drawing capabilities, enabling developers to create
// sophisticated custom UI elements while benefiting from the framework's
// layout management and event propagation systems.
type Custom struct {
	BaseWidget

	// handler is the user-defined function that processes events for this widget.
	// It should return true if the event was handled, false otherwise.
	// If nil, the widget will not handle any events.
	handler func(tcell.Event) bool

	// renderer is the user-defined function that draws the widget's visual content.
	// It receives the widget instance and screen interface for drawing operations.
	// If nil, the widget will not render anything.
	renderer func(Widget, Screen)
}

// NewCustom creates a new custom widget with user-defined rendering behavior.
// The widget can be configured as focusable or non-focusable depending on whether
// it needs to receive keyboard input. The renderer function is called during
// the rendering phase to draw the widget's visual content.
//
// Parameters:
//   - id: Unique identifier for the custom widget
//   - focusable: Whether the widget can receive keyboard focus
//   - renderer: Function that handles the visual rendering of the widget
//
// Returns:
//   - *Custom: A new custom widget instance
//
// The renderer function receives:
//   - Widget: Reference to this custom widget instance
//   - Screen: Screen interface for drawing operations
//
// Example usage:
//
//	// Create a simple horizontal line widget
//	custom := NewCustom("line", false, func(widget Widget, screen Screen) {
//		width, height := widget.Size()
//		x, y := widget.Position()
//		style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
//
//		// Draw a horizontal line across the widget
//		for i := 0; i < width; i++ {
//			screen.SetContent(x+i, y, '─', nil, style)
//		}
//	})
//
//	// Add event handling
//	custom.SetHandler(func(event tcell.Event) bool {
//		if keyEv, ok := event.(*tcell.EventKey); ok {
//			if keyEv.Key() == tcell.KeyEnter {
//				// Handle enter key press
//				return true
//			}
//		}
//		return false
//	})
func NewCustom(id string, focusable bool, renderer func(Widget, Screen)) *Custom {
	return &Custom{
		BaseWidget: BaseWidget{id: id, focusable: focusable},
		renderer:   renderer,
	}
}

// Handle processes events for the custom widget using the user-defined event handler.
// If no custom handler is set, it returns false to indicate the event was not handled.
//
// This method is called by the framework when events need to be processed by the widget.
// The custom handler function should return true if it handled the event, or false to
// allow the event to propagate to other widgets or the default handling mechanism.
//
// Parameters:
//   - event: The tcell.Event to be processed
//
// Returns:
//   - bool: true if the event was handled, false otherwise
func (c *Custom) Handle(event tcell.Event) bool {
	if c.handler != nil {
		return c.handler(event)
	}
	return false
}

// SetHandler sets a custom event handler function for the widget.
// The handler function will be called whenever events need to be processed by this widget.
// If handler is nil, the widget will not handle any events (Handle will return false).
//
// The handler function should:
//   - Return true if it successfully handled the event
//   - Return false if the event should be passed to other handlers or default processing
//
// Parameters:
//   - handler: Function that processes tcell.Event and returns bool indicating if handled
//
// Example usage:
//
//	custom.SetHandler(func(event tcell.Event) bool {
//		switch ev := event.(type) {
//		case *tcell.EventKey:
//			if ev.Key() == tcell.KeyEnter {
//				// Handle enter key
//				return true
//			}
//		}
//		return false
//	})
func (c *Custom) SetHandler(handler func(tcell.Event) bool) {
	c.handler = handler
}

// Render draws the custom widget using the user-defined renderer function.
// This method is called by the framework during the rendering phase.
// If no renderer is set, the widget will not draw anything.
//
// Parameters:
//   - screen: Screen interface for drawing operations
func (c *Custom) Render(screen Screen) {
	if c.renderer != nil {
		c.renderer(c, screen)
	}
}
