package zeichenwerk

// Separator represents a visual divider widget used to create visual separation
// between UI elements. It renders as a horizontal or vertical line using the
// specified border style, providing clear visual grouping in layouts.
//
// Features:
//   - Horizontal and vertical line rendering
//   - Customizable border styles (thin, thick, double, etc.)
//   - Non-interactive (non-focusable)
//   - Flexible sizing based on container dimensions
//   - Theme-aware styling
//
// Common use cases:
//   - Dividing sections in forms or dialogs
//   - Creating visual groups in lists or menus
//   - Separating header/footer from content areas
//   - Adding structure to complex layouts
//
// The separator automatically adapts its orientation based on its size hints:
//   - Width > Height: Renders as horizontal line
//   - Height > Width: Renders as vertical line
//   - Equal dimensions: Renders as horizontal line (default)
//
// Border styles depend on the current theme's border definitions.
type Separator struct {
	BaseWidget
	Border string // Border style identifier for the separator line
}

// NewSeparator creates a new separator widget with the specified border style.
// The separator is initialized as non-focusable and will render using the
// provided border style according to the current theme.
//
// Parameters:
//   - id: Unique identifier for the separator widget
//   - border: Border style identifier (e.g., "thin", "thick", "double")
//
// Returns:
//   - *Separator: A new separator widget instance
//
// Example usage:
//
//	// Horizontal separator
//	hSep := NewSeparator("h-sep", "thin")
//	hSep.SetHint(0, 1)  // Full width, 1 line height
//
//	// Vertical separator
//	vSep := NewSeparator("v-sep", "thick")
//	vSep.SetHint(1, 0)  // 1 character width, full height
//
//	// Use in layouts
//	builder.Flex("content", "vertical", "stretch", 0).
//		Label("header", "Header Content").
//		Add(hSep).  // Divider line
//		Label("body", "Body Content")
func NewSeparator(id, border string) *Separator {
	return &Separator{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Border:     border,
	}
}
