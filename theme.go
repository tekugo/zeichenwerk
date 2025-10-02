package zeichenwerk

import (
	"maps"
	"regexp"
	"slices"
	"strings"
)

// styleRegExp is the compiled regular expression used to parse CSS-like selectors.
// It supports the format: type/part.class#id/part:state
//
// # Selector Components
//
// The regex captures the following groups:
//   - Group 1: Widget type (e.g., "button", "input", "list")
//   - Group 2: Widget part after type (e.g., "placeholder", "bar", "item")
//   - Group 3: CSS class name without '.' prefix (e.g., "primary", "large")
//   - Group 4: Widget ID without '#' prefix (e.g., "submit-button", "main-menu")
//   - Group 5: Widget part after ID (alternative placement for parts)
//   - Group 6: State without ':' prefix (e.g., "focus", "hover", "disabled")
//
// # Design Note
//
// The part component appears twice (groups 2 and 5) to allow flexible positioning:
//   - "button/text.primary" - part after type
//   - "button.primary#submit/text" - part after ID
//
// This dual placement provides compatibility with different styling conventions.
var (
	styleRegExp, _ = regexp.Compile(`([0-9A-Za-z_\-]*)/?([0-9A-Za-z_\-]*)\.?([0-9A-Za-z_\-]*)#?([0-9A-Za-z_\-]*)/?([0-9A-Za-z_\-]*):?([0-9A-Za-z_\-]*)`)
)

// Theme provides a comprehensive styling system for widgets using CSS-like selectors.
// It implements hierarchical style resolution where more specific selectors
// override more general ones, enabling flexible and maintainable theming.
//
// # Selector Format
//
// The theme system uses CSS-like selectors with the following components:
//   - type: Widget type (button, input, list, etc.)
//   - part: Widget sub-component (placeholder, bar, item, etc.)
//   - class: CSS class for categorization (.primary, .large, etc.)
//   - id: Unique widget identifier (#submit-button, #main-menu, etc.)
//   - state: Widget state (:focus, :hover, :disabled, etc.)
//
// Full selector format: "type/part.class#id/part:state"
//
// # Selector Examples
//
//   - "button" - Styles all button widgets
//   - "button.primary" - Styles buttons with "primary" class
//   - "button#submit" - Styles the button with ID "submit"
//   - "button:focus" - Styles buttons in focus state
//   - "input/placeholder" - Styles the placeholder part of input widgets
//   - "input.large:focus" - Styles large input widgets when focused
//   - "list/item.selected" - Styles selected items in lists
//
// # Specificity and Cascading
//
// Style resolution follows CSS-like specificity rules:
//  1. Base styles (type only)
//  2. Class styles (type + class)
//  3. State styles (type + state)
//  4. ID styles (highest specificity)
//  5. Combined selectors (class + state, ID + state, etc.)
//
// More specific styles override less specific ones, allowing for
// hierarchical theming with sensible defaults and targeted overrides.
type Theme interface {
	// Add adds a the given style to the theme.
	// The selector is taken from the style and the style is fixed, when
	// added to the theme.
	//
	// Parameters:
	//   - style: The style to add to the theme
	Add(*Style)

	// Apply applies theme styles to a widget for multiple states/parts.
	// This is a convenience method that applies the base style and any
	// additional part-specific styles (e.g., focus, hover) to the widget.
	//
	// The method resolves the base selector and each state combination,
	// applying them to the widget's style map for efficient runtime lookup.
	//
	// Parameters:
	//   - Widget: The widget to apply styles to
	//   - string: The base selector for the widget (e.g., "button.primary")
	//   - ...string: Additional states to apply (e.g., "focus", "hover")
	//
	// Example:
	//   theme.Apply(button, "button.primary", "focus", "hover")
	//   // Applies: "button.primary", "button.primary:focus", "button.primary:hover"
	Apply(Widget, string, ...string)

	// Border retrieves a border style by name from the theme's border registry.
	// Border styles define the visual appearance of widget borders including
	// line styles, corner characters, and drawing mode.
	//
	// Parameters:
	//   - string: The name of the border style to retrieve
	//
	// Returns:
	//   - BorderStyle: The border style configuration
	Border(string) BorderStyle

	// Color resolves a color name or variable to its actual color value.
	// Supports both direct color names and theme variables (prefixed with $).
	// Theme variables allow for consistent color schemes across the application.
	//
	// Parameters:
	//   - string: Color name or variable (e.g., "red", "$primary", "#FF0000")
	//
	// Returns:
	//   - string: The resolved color value
	Color(string) string

	// Colors returns the complete map of color variables defined in the theme.
	// This provides access to the theme's color palette for inspection or
	// dynamic color management.
	//
	// Returns:
	//   - map[string]string: Map of color variable names to their values
	Colors() map[string]string

	// Flag retrieves a boolean configuration flag from the theme.
	// Flags control various behavioral aspects of the theme system
	// and widget rendering.
	//
	// Parameters:
	//   - string: The name of the flag to retrieve
	//
	// Returns:
	//   - bool: The flag value (false if not set)
	Flag(string) bool

	// Get retrieves the resolved style for the specified selector.
	// This method implements the core style resolution algorithm,
	// applying cascading rules and specificity to produce the final style.
	//
	// Parameters:
	//   - string: The CSS-like selector to resolve
	//
	// Returns:
	//   - Style: The resolved style (never nil, returns default as fallback)
	Get(string) *Style

	// Rune retrieves a special Unicode character by name from the theme.
	// These characters are used for drawing borders, arrows, bullets,
	// and other decorative elements in the UI.
	//
	// Parameters:
	//   - string: The name of the Unicode character to retrieve
	//
	// Returns:
	//   - rune: The Unicode character
	Rune(string) rune

	// SetBorders updates the theme's border style registry with the provided map.
	// This replaces the existing border definitions with the new set.
	//
	// Parameters:
	//   - map[string]BorderStyle: Map of border names to their style definitions
	SetBorders(map[string]BorderStyle)

	// SetColors updates the theme's color variable registry with the provided map.
	// This replaces the existing color definitions with the new set.
	//
	// Parameters:
	//   - map[string]string: Map of color variable names to their values
	SetColors(map[string]string)

	// SetFlags updates the theme's flag registry with the provided map.
	// This replaces the existing flag definitions with the new set.
	//
	// Parameters:
	//   - map[string]bool: Map of flag names to their boolean values
	SetFlags(map[string]bool)

	// SetRunes updates the theme's Unicode character registry with the provided map.
	// This replaces the existing character definitions with the new set.
	//
	// Parameters:
	//   - map[string]rune: Map of character names to their Unicode values
	SetRunes(map[string]rune)

	// SetStyles updates the theme's style registry with the provided map.
	// This replaces the existing style definitions with the new set.
	//
	// Parameters:
	//   - map[string]*Style: Map of selectors to their style definitions
	SetStyles(...*Style)

	// Styles returns the complete map of selectors to styles defined in the theme.
	// This provides access to all style definitions for inspection or
	// dynamic style management.
	//
	// Returns:
	//   - map[string]*Style: Map of selectors to their style definitions
	Styles() []*Style
}

// MapTheme is a concrete implementation of the Theme interface that stores
// styles and theme resources in maps using string keys for efficient lookup.
// It provides the complete CSS-like selector hierarchy with cascading resolution
// for maintaining consistent styling across the application.
//
// # Architecture
//
// MapTheme organizes theme resources into separate registries:
//   - styles: Core styling rules indexed by CSS-like selectors
//   - colors: Named color variables for consistent color schemes
//   - borders: Border style definitions for drawing widget frames
//   - runes: Special Unicode characters for decorative elements
//   - flags: Boolean configuration options for theme behavior
//
// # Performance Characteristics
//
// The map-based storage provides:
//   - O(1) direct style lookup for exact selector matches
//   - Efficient memory usage through shared style instances
//   - Fast theme switching by replacing map contents
//   - Minimal overhead for style resolution cascading
//
// # Thread Safety
//
// MapTheme is not thread-safe. Applications that modify themes from
// multiple goroutines must provide external synchronization.
type MapTheme struct {
	borders map[string]BorderStyle // Registry of border styles indexed by name
	colors  map[string]string      // Registry of color variables (e.g., "$primary" -> "#007ACC")
	flags   map[string]bool        // Registry of boolean configuration flags
	runes   map[string]rune        // Registry of special Unicode characters for UI elements
	styles  map[string]*Style      // Registry of style definitions indexed by CSS-like selectors
}

// NewMapTheme creates a new MapTheme instance with empty initialized registries.
// This is the recommended constructor for creating themes that will be populated
// with styles, colors, and other resources.
//
// # Initialization
//
// The constructor initializes all internal maps to empty states:
//   - styles: Empty style registry ready for selector-based style definitions
//   - colors: Empty color variable registry for theme color schemes
//   - borders: Empty border style registry for widget frame styles
//   - runes: Empty Unicode character registry for decorative elements
//   - flags: Empty flag registry for behavioral configuration
//
// # Usage Pattern
//
// Typically used in combination with theme builder methods or bulk assignment:
//
//	theme := NewMapTheme()
//	theme.SetColors(map[string]string{
//		"$primary":   "#007ACC",
//		"$secondary": "#FF6B35",
//	})
//	theme.Set("button.primary", NewStyle("$primary", "white"))
//
// Returns:
//   - *MapTheme: A new MapTheme instance with empty registries ready for configuration
func NewMapTheme() *MapTheme {
	return &MapTheme{
		borders: make(map[string]BorderStyle),
		colors:  make(map[string]string),
		flags:   make(map[string]bool),
		runes:   make(map[string]rune),
		styles:  make(map[string]*Style),
	}
}

// Add adds a style to the theme.
// The style is automatically assigned a parent based on the style selector.
// THrough this mechanism, inheritance/cascading is implemented.
func (m *MapTheme) Add(style *Style) {
	parts := split(style.selector)
	key, priority := selector(0, parts)

	var parent *Style
	for priority > 0 {
		key, priority = selector(priority, parts)
		parent = m.styles[key]
		if parent != nil {
			style.WithParent(parent)
			break
		}
	}

	m.styles[style.selector] = style
	style.Fix()
}

// Apply applies styles to a widget based on its type and id.
//
// Parameters:
//   - widget: The widget to apply the styles to
//   - selector: Style selector
//   - states: Additional states to assign styles to
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
		widget.SetStyle(part, style)
		for _, state := range states {
			style := m.Get(selector + ":" + state)
			widget.SetStyle(part+":"+state, style)
		}
	} else {
		style := m.Get(selector)
		widget.SetStyle("", style)
		for _, state := range states {
			style := m.Get(selector + ":" + state)
			widget.SetStyle(":"+state, style)
		}
	}
}

// Border retrieves a border style by name from the theme's border registry.
// Returns the zero value BorderStyle if the border name is not found.
func (m *MapTheme) Border(border string) BorderStyle {
	return m.borders[border]
}

// Color resolves a color name or variable to its actual color value.
// Theme variables (prefixed with $) are looked up in the color registry,
// while direct color values are returned unchanged.
//
// # Variable Resolution
//
// Color variables starting with "$" are resolved from the colors map:
//   - "$primary" → looks up m.colors["$primary"]
//   - If found, returns the mapped value
//   - If not found, returns the original variable name
//
// # Direct Colors
//
// Non-variable colors are returned as-is:
//   - "red", "#FF0000", "rgb(255,0,0)" → returned unchanged
//
// Parameters:
//   - color: Color name or variable to resolve
//
// Returns:
//   - string: The resolved color value
func (m *MapTheme) Color(color string) string {
	if strings.HasPrefix(color, "$") {
		if name, ok := m.colors[color]; ok {
			return name
		}
	}
	return color
}

// Colors returns a direct reference to the theme's color variable registry.
// This allows for inspection and iteration over all defined color variables.
func (m *MapTheme) Colors() map[string]string {
	return m.colors
}

// Flag retrieves a boolean configuration flag from the theme's flag registry.
// Returns false if the flag name is not found.
func (m *MapTheme) Flag(flag string) bool {
	return m.flags[flag]
}

// Get returns a style which maps the given selector the best.
//
// Parameters:
//   - s: Style selector
func (m *MapTheme) Get(s string) *Style {
	parts := split(s)
	key := s
	prio := 0

	for prio >= 0 {
		key, prio = selector(prio, parts)
		style := m.styles[key]
		if style != nil {
			return style
		}
	}

	return &DefaultStyle
}

// Rune retrieves a special Unicode character by name from the theme's rune registry.
// Returns the zero value (rune 0) if the rune name is not found.
//
// # Common Rune Names
//
// Typical rune names include:
//   - "arrow-up", "arrow-down", "arrow-left", "arrow-right"
//   - "bullet", "checkbox", "radio"
//   - "spinner1", "spinner2", etc.
//   - Border drawing characters for custom borders
func (m *MapTheme) Rune(name string) rune {
	return m.runes[name]
}

// SetBorders replaces the theme's border style registry with the provided map.
// This method is used for bulk border style configuration and theme initialization.
//
// Parameters:
//   - borders: Map of border names to their BorderStyle definitions
func (m *MapTheme) SetBorders(borders map[string]BorderStyle) {
	m.borders = borders
}

// SetColors replaces the theme's color variable registry with the provided map.
// This method is used for bulk color configuration and theme initialization.
//
// # Color Variable Format
//
// Color variables should use the "$" prefix convention:
//   - "$primary": "#007ACC"
//   - "$secondary": "#FF6B35"
//   - "$background": "#1E1E1E"
//
// Parameters:
//   - colors: Map of color variable names to their color values
func (m *MapTheme) SetColors(colors map[string]string) {
	m.colors = colors
}

// SetFlags replaces the theme's flag registry with the provided map.
// This method is used for bulk flag configuration and theme initialization.
//
// Parameters:
//   - flags: Map of flag names to their boolean values
func (m *MapTheme) SetFlags(flags map[string]bool) {
	m.flags = flags
}

// SetRunes replaces the theme's Unicode character registry with the provided map.
// This method is used for bulk rune configuration and theme initialization.
//
// Parameters:
//   - runes: Map of character names to their Unicode values
func (m *MapTheme) SetRunes(runes map[string]rune) {
	m.runes = runes
}

// SetStyles replaces the theme's style registry with the provided map.
// This method is used for bulk style configuration and theme initialization.
// Unlike Define(), this method is the official way to replace all styles.
//
// Parameters:
//   - styles: Map of CSS-like selectors to their Style definitions
func (m *MapTheme) SetStyles(styles ...*Style) {
	for _, style := range styles {
		m.Add(style)
	}
}

// Styles returns a direct reference to the theme's style registry.
// This allows for inspection and iteration over all defined styles.
func (m *MapTheme) Styles() []*Style {
	return slices.Collect(maps.Values(m.styles))
}

// split parses a CSS-like selector string into its component parts using
// the predefined regular expression. This is an internal utility function
// that enables the hierarchical style resolution system.
//
// # Parsing Logic
//
// The function uses the global styleRegExp to extract selector components
// according to the pattern: type/part.class#id/part:state
//
// # Return Structure
//
// The function returns a slice with the following indices:
//   - [0]: Full match (entire selector string)
//   - [1]: Widget type (e.g., "button", "input", "list")
//   - [2]: Widget part after type (e.g., "text", "placeholder", "item")
//   - [3]: CSS class without '.' prefix (e.g., "primary", "large")
//   - [4]: Widget ID without '#' prefix (e.g., "submit-button", "main-menu")
//   - [5]: Widget part after ID (alternative part placement)
//   - [6]: State without ':' prefix (e.g., "focus", "hover", "disabled")
//
// # Error Handling
//
// Invalid selectors that don't match the regex pattern return nil,
// which callers should check to avoid processing malformed selectors.
//
// # Usage Examples
//
//   - "button" → ["button", "button", "", "", "", "", ""]
//   - "button.primary" → ["button.primary", "button", "", "primary", "", "", ""]
//   - "input:focus" → ["input:focus", "input", "", "", "", "", "focus"]
//   - "list/item.selected:hover" → ["list/item.selected:hover", "list", "item", "selected", "", "", "hover"]
//
// Parameters:
//   - selector: The CSS-like selector string to parse
//
// Returns:
//   - []string: Parsed components of the selector, or nil if invalid format
func split(selector string) []string {
	result := styleRegExp.FindStringSubmatch(selector)
	return result
}

// selector returns a stripped down selector string based on the priority.
//
// The priority of the selectors implicitly defined the cascading of styles.
func selector(priority int, parts []string) (string, int) {
	if priority < 1 && parts[4] != "" && parts[5] != "" && parts[6] != "" {
		return "#" + parts[4] + "/" + parts[5] + ":" + parts[6], 1
	}
	if priority < 2 && parts[4] != "" && parts[5] != "" {
		return "#" + parts[4] + "/" + parts[5], 2
	}
	if priority < 3 && parts[4] != "" && parts[6] != "" {
		return "#" + parts[4] + ":" + parts[6], 3
	}
	if priority < 4 && parts[4] != "" {
		return "#" + parts[4], 4
	}
	if priority < 5 && parts[1] != "" && parts[2] != "" && parts[3] != "" && parts[6] != "" {
		return parts[1] + "/" + parts[2] + "." + parts[3] + ":" + parts[6], 5
	}
	if priority < 6 && parts[1] != "" && parts[2] != "" && parts[6] != "" {
		return parts[1] + "/" + parts[2] + ":" + parts[6], 6
	}
	if priority < 7 && parts[1] != "" && parts[3] != "" && parts[6] != "" {
		return parts[1] + "." + parts[3] + ":" + parts[6], 7
	}
	if priority < 8 && parts[1] != "" && parts[6] != "" {
		return parts[1] + ":" + parts[6], 8
	}
	if priority < 9 && parts[6] != "" {
		return ":" + parts[6], 9
	}
	if priority < 10 && parts[1] != "" && parts[2] != "" && parts[3] != "" {
		return parts[1] + "/" + parts[2] + "." + parts[3], 10
	}
	if priority < 11 && parts[1] != "" && parts[3] != "" {
		return parts[1] + "." + parts[3], 11
	}
	if priority < 12 && parts[3] != "" {
		return "." + parts[3], 12
	}
	if priority < 13 && parts[1] != "" && parts[2] != "" {
		return parts[1] + "/" + parts[2], 13
	}
	if priority < 14 && parts[1] != "" {
		return parts[1], 14
	}
	if priority < 15 {
		return "", 15
	}

	return "", -1
}
