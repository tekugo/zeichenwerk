package zeichenwerk

import (
	"strings"
)

// MapTheme is a concrete implementation of the Theme interface that stores
// styles in a map using selectors as keys. It provides efficient style
// lookup and supports the full CSS-like selector hierarchy with fallback
// resolution for maintaining consistent styling across the application.
type MapTheme struct {
	borders map[string]BorderStyle // Border styles
	colors  map[string]string      // Color variables
	flags   map[string]bool        // Flags
	runes   map[string]rune        // Speziell rendering runes
	styles  map[string]Style       // Map of selectors to their corresponding styles
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

func (m *MapTheme) Border(border string) BorderStyle {
	return BorderStyles[border]
}

func (m *MapTheme) Cascade(style *Style, selector string) {
	other := m.styles[selector]
	style.Cascade(&other)
}

func (m *MapTheme) Color(color string) string {
	if strings.HasPrefix(color, "$") {
		if name, ok := m.colors[color]; ok {
			return name
		}
	}
	return color
}

func (m *MapTheme) Colors() map[string]string {
	return m.colors
}

func (m *MapTheme) Flag(flag string) bool {
	return m.flags[flag]
}

// Get retrieves the style for the given selector using hierarchical resolution.
// This method implements the Theme interface's style resolution algorithm,
// searching for the most specific match and falling back to more general
// selectors if needed.
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
	if parts[3] != "" {
		m.Cascade(&result, "."+parts[3])
	}
	if parts[1] != "" && parts[2] != "" {
		m.Cascade(&result, parts[1]+"/"+parts[2])
	}
	if parts[1] != "" && parts[2] != "" && parts[3] != "" {
		m.Cascade(&result, parts[1]+"/"+parts[2]+"."+parts[3])
	}
	if parts[6] != "" {
		m.Cascade(&result, ":"+parts[6])
	}
	if parts[1] != "" && parts[6] != "" {
		m.Cascade(&result, parts[1]+":"+parts[6])
	}
	if parts[1] != "" && parts[3] != "" && parts[6] != "" {
		m.Cascade(&result, parts[1]+"."+parts[3]+":"+parts[6])
	}
	if parts[1] != "" && parts[2] != "" && parts[6] != "" {
		m.Cascade(&result, parts[1]+"/"+parts[2]+":"+parts[6])
	}
	if parts[1] != "" && parts[2] != "" && parts[3] != "" && parts[6] != "" {
		m.Cascade(&result, parts[1]+"/"+parts[2]+"."+parts[3]+":"+parts[6])
	}
	if parts[4] != "" {
		m.Cascade(&result, "#"+parts[4])
	}
	if parts[4] != "" && parts[6] != "" {
		m.Cascade(&result, "#"+parts[4]+":"+parts[6])
	}
	if parts[4] != "" && parts[5] != "" {
		m.Cascade(&result, "#"+parts[4]+"/"+parts[5])
	}
	if parts[4] != "" && parts[5] != "" && parts[6] != "" {
		m.Cascade(&result, "#"+parts[4]+"/"+parts[5]+":"+parts[6])
	}

	return result
}

func (m *MapTheme) Rune(name string) rune {
	return m.runes[name]
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
//   - states: Additional states to apply (e.g., "focus", "hover", "disabled")
func (m MapTheme) SetStyles(widget Widget, selector string, states ...string) {
	parts := split(selector)
	part := ""
	if parts[2] != "" {
		part = parts[2]
	} else if parts[5] != "" {
		part = parts[5]
	}

	if part != "" {
		style := m.Get(selector)
		widget.SetStyle(part, &style)
		for _, state := range states {
			style := m.Get(selector + ":" + state)
			widget.SetStyle(part+":"+state, &style)
		}
	} else {
		style := m.Get(selector)
		widget.SetStyle("", &style)
		for _, state := range states {
			style := m.Get(selector + ":" + state)
			widget.SetStyle(":"+state, &style)
		}
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
	result := styleRegExp.FindStringSubmatch(selector)
	return result
}
