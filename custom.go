package zeichenwerk

import "github.com/gdamore/tcell/v2"

// Custom represents a widget with user-defined rendering and event handling behavior.
// It provides a flexible foundation for creating specialized widgets that don't fit
// the standard widget patterns, allowing complete control over visual appearance
// and interaction logic.
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
	handler  func(tcell.Event) bool // Custom event handler function
	renderer func(Widget, Screen)   // Custom rendering function
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
//	custom := NewCustom("chart", false, func(widget Widget, screen Screen) {
//		width, height := widget.Size()
//		// Draw custom content using screen.SetContent()
//		for x := 0; x < width; x++ {
//			screen.SetContent(x, 0, '-', nil, style)
//		}
//	})
func NewCustom(id string, focusable bool, renderer func(Widget, Screen)) *Custom {
	return &Custom{
		BaseWidget: BaseWidget{id: id, focusable: focusable},
		renderer:   renderer,
	}
}

func (c *Custom) Handle(event tcell.Event) bool {
	if c.handler != nil {
		return c.handler(event)
	} else {
		return false
	}
}
