package zeichenwerk

// Hidden represents an invisible widget that occupies space but renders nothing.
// It is primarily used as a spacer or placeholder in layouts where you need
// to reserve space without displaying any visible content.
//
// Features:
//   - Invisible rendering (draws nothing to the screen)
//   - Non-focusable (cannot receive keyboard input)
//   - Space allocation (respects size hints for layout purposes)
//   - Layout participation (acts as a normal widget for containers)
//
// Common use cases:
//   - Creating flexible spacers in layouts
//   - Reserving space for future content
//   - Implementing gaps between visible widgets
//   - Placeholder widgets during dynamic UI construction
//
// The Hidden widget behaves like any other widget in terms of layout
// calculations and hierarchy management, but produces no visual output
// during rendering operations.
type Hidden struct {
	BaseWidget
}

// NewHidden creates a new hidden widget with the specified identifier.
// The widget is initialized as non-focusable and invisible, making it
// suitable for use as a spacer or placeholder in layouts.
//
// Parameters:
//   - id: Unique identifier for the hidden widget
//
// Returns:
//   - *Hidden: A new hidden widget instance
//
// Example usage:
//
//	spacer := NewHidden("spacer-1")
//	spacer.SetHint(20, 1)  // Reserve 20x1 character space
//
//	// Use in layout to create spacing
//	builder.Flex("row", "horizontal", "start", 0).
//		Button("btn1", "Button 1").
//		Add(spacer).  // Creates gap
//		Button("btn2", "Button 2")
func NewHidden(id string) *Hidden {
	return &Hidden{
		BaseWidget: BaseWidget{id: id, focusable: false},
	}
}
