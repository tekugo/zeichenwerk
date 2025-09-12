package zeichenwerk

import (
	"regexp"
	"strings"
)

// CSS-like selector pattern: type/part.class#id/part:state
// This regular expression parses selectors in the format:
// - type: widget type (e.g., "button", "input")
// - part: optional widget part (e.g., "placeholder", "bar")
// - class: optional class name preceded by '.'
// - id: optional ID preceded by '#'
// - state: optional state preceded by ':' (e.g., "focus", "hover")
//
// It is on purpose, that the part is included 2 times, so that
// it can be either after the type or the id.
var (
	styleRegExp, _ = regexp.Compile(`([0-9A-Za-z_\-]*)/?([0-9A-Za-z_\-]*)\.?([0-9A-Za-z_\-]*)#?([0-9A-Za-z_\-]*)/?([0-9A-Za-z_\-]*):?([0-9A-Za-z_\-]*)`)
	defaultStyle   = NewStyle("green", "black").SetMargin(0).SetPadding(0)
)

// Theme provides a styling system for widgets using CSS-like selectors.
// It allows for hierarchical style resolution where more specific selectors
// override more general ones. The theme system supports widget types, classes,
// IDs, and parts/states for flexible and maintainable styling.
//
// Selector format: "type.class#id:part"
// Examples:
//   - "button" - styles all buttons
//   - "button.primary" - styles buttons with class "primary"
//   - "button#submit" - styles the button with ID "submit"
//   - "button:focus" - styles buttons in focus state
//   - "input.large:focus" - styles large input widgets when focused
type Theme interface {
	// Apply applies theme styles to a widget for multiple states/parts.
	// This is a convenience method that applies the base style and any
	// additional part-specific styles (e.g., focus, hover) to the widget.
	//
	// Parameters:
	//   - Widget: The widget to apply styles to
	//   - string: The base selector for the widget
	//   - ...string: Additional parts/states to apply (e.g., "focus", "hover")
	Apply(Widget, string, ...string)

	Border(string) BorderStyle

	Color(string) string

	Colors() map[string]string

	Flag(string) bool

	Get(string) Style

	Rune(string) rune

	// Set assigns a style to the specified selector in the theme.
	// This allows for dynamic theme customization and style registration.
	//
	// Parameters:
	//   - string: The CSS-like selector to associate with the style
	//   - *Style: The style configuration to assign
	Set(string, *Style)

	SetBorders(map[string]BorderStyle)
	SetColors(map[string]string)
	SetFlags(map[string]bool)
	SetRunes(map[string]rune)
	SetStyles(map[string]*Style)

	Styles() map[string]*Style
}

// MapTheme is a concrete implementation of the Theme interface that stores
// styles in a map using selectors as keys. It provides efficient style
// lookup and supports the full CSS-like selector hierarchy with fallback
// resolution for maintaining consistent styling across the application.
type MapTheme struct {
	borders map[string]BorderStyle // Border styles
	colors  map[string]string      // Color variables
	flags   map[string]bool        // Flags
	runes   map[string]rune        // Speziell rendering runes
	styles  map[string]*Style      // Map of selectors to their corresponding styles
}

// NewMapTheme creates a new MapTheme instance with an initialized style map.
// This is the recommended way to create a MapTheme for use in applications.
//
// Returns:
//   - *MapTheme: A new MapTheme instance ready for use
func NewMapTheme() *MapTheme {
	return &MapTheme{
		borders: make(map[string]BorderStyle),
		colors:  make(map[string]string),
		flags:   make(map[string]bool),
		runes:   make(map[string]rune),
		styles:  make(map[string]*Style),
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
func (m *MapTheme) Apply(widget Widget, selector string, states ...string) {
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

func (m *MapTheme) Border(border string) BorderStyle {
	return m.borders[border]
}

func (m *MapTheme) Cascade(style *Style, selector string) {
	other := m.styles[selector]
	style.Cascade(other)
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

func (m *MapTheme) Define(styles map[string]*Style) {
	m.styles = styles
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
	if parts[1] != "" && parts[2] != "" {
		m.Cascade(&result, parts[1]+"/"+parts[2])
	}
	if parts[3] != "" {
		m.Cascade(&result, "."+parts[3])
	}
	if parts[1] != "" && parts[3] != "" {
		m.Cascade(&result, parts[1]+"."+parts[3])
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
		m.styles[selector] = style
	}
}

func (m *MapTheme) SetBorders(borders map[string]BorderStyle) {
	m.borders = borders
}

func (m *MapTheme) SetColors(colors map[string]string) {
	m.colors = colors
}

func (m *MapTheme) SetFlags(flags map[string]bool) {
	m.flags = flags
}

func (m *MapTheme) SetRunes(runes map[string]rune) {
	m.runes = runes
}

func (m *MapTheme) SetStyles(styles map[string]*Style) {
	m.styles = styles
}

func (m *MapTheme) Styles() map[string]*Style {
	return m.styles
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
