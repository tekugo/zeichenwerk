package zeichenwerk

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

// MapTheme is a concrete implementation of the Theme interface that stores
// styles in a map using selectors as keys. It provides efficient style
// lookup and supports the full CSS-like selector hierarchy with fallback
// resolution for maintaining consistent styling across the application.
type MapTheme struct {
	colors map[string]string // Color variables
	styles map[string]Style  // Map of selectors to their corresponding styles
}

// NewMapTheme creates a new MapTheme instance with an initialized style map.
// This is the recommended way to create a MapTheme for use in applications.
//
// Returns:
//   - *MapTheme: A new MapTheme instance ready for use
func NewMapTheme() *MapTheme {
	return &MapTheme{
		styles: make(map[string]Style),
	}
}

// Get retrieves the style for the given selector using hierarchical resolution.
// This method implements the Theme interface's style resolution algorithm,
// searching for the most specific match and falling back to more general
// selectors if needed.
//
// The resolution follows CSS-like specificity rules:
//  1. Exact selector match (highest priority)
//  2. type.class:part combination
//  3. type:part combination
//  4. type only
//  5. Default style (fallback)
//
// Parameters:
//   - selector: The CSS-like selector string to resolve
//
// Returns:
//   - Style: The resolved style, never nil (returns default style as fallback)
func (m *MapTheme) Get(selector string) Style {
	// Copy the default style first
	result := *defaultStyle

	parts := split(selector)
	m.Cascade(&result, "")

	if parts[1] != "" {
		m.Cascade(&result, parts[1])
	}
	if parts[2] != "" {
		m.Cascade(&result, "."+parts[2])
	}
	if parts[4] != "" {
		m.Cascade(&result, ":"+parts[4])
	}
	if parts[1] != "" && parts[2] != "" {
		m.Cascade(&result, parts[1]+"."+parts[2])
	}
	if parts[1] != "" && parts[4] != "" {
		m.Cascade(&result, parts[1]+":"+parts[4])
	}
	if parts[1] != "" && parts[2] != "" && parts[4] != "" {
		m.Cascade(&result, parts[1]+"."+parts[2]+":"+parts[4])
	}
	if parts[3] != "" {
		m.Cascade(&result, "#"+parts[3])
	}
	if parts[3] != "" && parts[4] != "" {
		m.Cascade(&result, "#"+parts[3]+":"+parts[4])
	}
	return result
}

func (m *MapTheme) Cascade(style *Style, selector string) {
	other := m.styles[selector]
	style.Cascade(&other)
}

func (m *MapTheme) Color(color string) tcell.Color {
	if strings.HasPrefix(color, "$") {
		if name, ok := m.colors[color]; ok {
			color = name
		}
	}
	result, err := ParseColor(color)
	if err == nil {
		return result
	} else {
		return tcell.ColorDefault
	}
}

// Set assigns a style to the specified selector in the theme.
// The method validates the selector format and stores the style
// in the internal map for later retrieval during style resolution.
//
// Parameters:
//   - selector: The CSS-like selector string (e.g., "button.primary:focus")
//   - style: The style configuration to associate with the selector
func (m *MapTheme) Set(selector string, style *Style) {
	parts := split(selector)
	if parts != nil {
		m.styles[selector] = *style
	}
}

// SetStyles applies theme styles to a widget for the base state and additional parts.
// This is a convenience method that automatically resolves and applies multiple
// style states to a widget in a single call, reducing boilerplate code.
//
// The method:
//  1. Applies the base style (resolved from the base selector) to the widget
//  2. For each additional part, resolves "selector:part" and applies it
//
// Example usage:
//
//	theme.SetStyles(button, "button.primary", "focus", "hover")
//	// This applies:
//	// - "button.primary" style as the base style
//	// - "button.primary:focus" style for the focus state
//	// - "button.primary:hover" style for the hover state
//
// Parameters:
//   - widget: The widget to apply styles to
//   - selector: The base selector for the widget
//   - styles: Additional parts/states to apply (e.g., "focus", "hover", "disabled")
func (m MapTheme) SetStyles(widget Widget, selector string, styles ...string) {
	style := m.Get(selector)
	widget.SetStyle("", &style)

	for _, part := range styles {
		style := m.Get(selector + ":" + part)
		widget.SetStyle(part, &style)
	}
}

// split parses a CSS-like selector string into its component parts using
// the predefined regular expression. This is an internal utility function
// that breaks down selectors for hierarchical style resolution.
//
// The function returns a slice with the following structure:
//
//	[0]: Full match
//	[1]: Type (widget type)
//	[2]: Class (without the '.' prefix)
//	[3]: ID (without the '#' prefix)
//	[4]: Part/state (without the ':' prefix)
//
// Parameters:
//   - selector: The CSS-like selector string to parse
//
// Returns:
//   - []string: Parsed components of the selector, or nil if invalid
func split(selector string) []string {
	result := re.FindStringSubmatch(selector)
	return result
}
