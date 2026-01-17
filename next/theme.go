package next

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
type Theme struct {
	borders map[string]*Border // Registry of border styles indexed by name
	colors  map[string]string  // Registry of color variables (e.g., "$primary" -> "#007ACC")
	flags   map[string]bool    // Registry of boolean configuration flags
	strings map[string]string  // Registry of special Unicode characters for UI elements
	styles  map[string]*Style  // Registry of style definitions indexed by CSS-like selectors
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
func NewTheme() *Theme {
	return &Theme{
		borders: make(map[string]*Border),
		colors:  make(map[string]string),
		flags:   make(map[string]bool),
		strings: make(map[string]string),
		styles:  make(map[string]*Style),
	}
}

// Add adds a style to the theme.
// The style is automatically assigned a parent based on the style selector.
// THrough this mechanism, inheritance/cascading is implemented.
func (t *Theme) Add(style *Style) {
	parts := split(style.selector)
	key, priority := selector(0, parts)

	var parent *Style
	for priority > 0 {
		key, priority = selector(priority, parts)
		parent = t.styles[key]
		if parent != nil {
			style.WithParent(parent)
			break
		}
	}

	t.styles[style.selector] = style
	style.Fix()
}

// Apply applies styles to a widget based on its type and id.
//
// Parameters:
//   - widget: The widget to apply the styles to
//   - selector: Style selector
//   - states: Additional states to assign styles to
func (t *Theme) Apply(widget Widget, selector string, states ...string) {
	parts := split(selector)
	part := ""
	if parts[2] != "" {
		part = parts[2]
	} else if parts[5] != "" {
		part = parts[5]
	}

	if part != "" {
		style := t.Get(selector)
		widget.SetStyle(part, style)
		for _, state := range states {
			style := t.Get(selector + ":" + state)
			widget.SetStyle(part+":"+state, style)
		}
	} else {
		style := t.Get(selector)
		widget.SetStyle("", style)
		for _, state := range states {
			style := t.Get(selector + ":" + state)
			widget.SetStyle(":"+state, style)
		}
	}
}

// Border retrieves a border style by name from the theme's border registry.
// Returns the zero value BorderStyle if the border name is not found.
func (t *Theme) Border(border string) *Border {
	return t.borders[border]
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
func (t *Theme) Color(color string) string {
	if strings.HasPrefix(color, "$") {
		if name, ok := t.colors[color]; ok {
			return name
		}
	}
	return color
}

// Colors returns a direct reference to the theme's color variable registry.
// This allows for inspection and iteration over all defined color variables.
func (t *Theme) Colors() map[string]string {
	return t.colors
}

// Flag retrieves a boolean configuration flag from the theme's flag registry.
// Returns false if the flag name is not found.
func (t *Theme) Flag(flag string) bool {
	return t.flags[flag]
}

// Get returns a style which maps the given selector the best.
//
// Parameters:
//   - s: Style selector
func (t *Theme) Get(s string) *Style {
	parts := split(s)
	key := s
	prio := 0

	for prio >= 0 {
		key, prio = selector(prio, parts)
		style := t.styles[key]
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
func (t *Theme) String(name string) string {
	return t.strings[name]
}

// SetBorders replaces the theme's border style registry with the provided map.
// This method is used for bulk border style configuration and theme initialization.
//
// Parameters:
//   - borders: Map of border names to their BorderStyle definitions
func (t *Theme) SetBorders(borders map[string]*Border) {
	t.borders = borders
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
func (t *Theme) SetColors(colors map[string]string) {
	t.colors = colors
}

// SetFlags replaces the theme's flag registry with the provided map.
// This method is used for bulk flag configuration and theme initialization.
//
// Parameters:
//   - flags: Map of flag names to their boolean values
func (t *Theme) SetFlags(flags map[string]bool) {
	t.flags = flags
}

// SetRunes replaces the theme's Unicode character registry with the provided map.
// This method is used for bulk rune configuration and theme initialization.
//
// Parameters:
//   - runes: Map of character names to their Unicode values
func (t *Theme) SetStrings(strings map[string]string) {
	t.strings = strings
}

// SetStyles replaces the theme's style registry with the provided map.
// This method is used for bulk style configuration and theme initialization.
// Unlike Define(), this method is the official way to replace all styles.
//
// Parameters:
//   - styles: Map of CSS-like selectors to their Style definitions
func (t *Theme) SetStyles(styles ...*Style) {
	for _, style := range styles {
		t.Add(style)
	}
}

// Styles returns a direct reference to the theme's style registry.
// This allows for inspection and iteration over all defined styles.
func (t *Theme) Styles() []*Style {
	return slices.Collect(maps.Values(t.styles))
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
