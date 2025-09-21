// Package theme.go implements the theming and styling system for zeichenwerk.
//
// This file provides a CSS-like styling framework that allows widgets to be
// styled using hierarchical selectors. The theming system supports:
//   - Widget type-based styling (e.g., "button", "input")
//   - Class-based styling (e.g., ".primary", ".large")
//   - ID-based styling (e.g., "#submit-button")
//   - State-based styling (e.g., ":focus", ":hover")
//   - Part-based styling (e.g., "/placeholder", "/bar")
//   - Complex combinations of the above
//
// The system uses a cascading resolution algorithm where more specific
// selectors override more general ones, similar to CSS precedence rules.

package zeichenwerk

import (
	"regexp"
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
//
// defaultStyle provides the base styling foundation for all widgets when no
// specific theme styles are defined. It uses a green-on-black color scheme
// with no margins or padding.
var (
	styleRegExp, _ = regexp.Compile(`([0-9A-Za-z_\-]*)/?([0-9A-Za-z_\-]*)\.?([0-9A-Za-z_\-]*)#?([0-9A-Za-z_\-]*)/?([0-9A-Za-z_\-]*):?([0-9A-Za-z_\-]*)`)
	defaultStyle   = NewStyle("green", "black").SetMargin(0).SetPadding(0)
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
	Get(string) Style

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

	// Set assigns a style to the specified selector in the theme.
	// This allows for dynamic theme customization and style registration
	// at runtime, enabling flexible theming systems.
	//
	// Parameters:
	//   - string: The CSS-like selector to associate with the style
	//   - *Style: The style configuration to assign
	Set(string, *Style)

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
	SetStyles(map[string]*Style)

	// Styles returns the complete map of selectors to styles defined in the theme.
	// This provides access to all style definitions for inspection or
	// dynamic style management.
	//
	// Returns:
	//   - map[string]*Style: Map of selectors to their style definitions
	Styles() map[string]*Style
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

// Apply resolves and assigns theme styles to a widget for multiple states.
// This is the primary method for applying theme styling to widgets, handling
// both base styles and state-specific variations in a single operation.
//
// # Style Application Process
//
// The method follows these steps:
//  1. Parses the selector to determine if it contains a widget part
//  2. Resolves the base style using the theme's cascading algorithm
//  3. Applies the base style to the widget's style registry
//  4. For each additional state, resolves "selector:state" and applies it
//  5. Stores all styles in the widget for efficient runtime lookup
//
// # Part Handling
//
// If the selector contains a part (e.g., "input/placeholder"), the styles
// are applied to that specific part of the widget. Otherwise, they are
// applied to the widget's base styling.
//
// # State Combinations
//
// Each state parameter creates a combined selector:
//   - Base: "button.primary" → applied to widget's default style
//   - State: "focus" → creates "button.primary:focus" selector
//   - Result: Widget has both default and focus-specific styling
//
// # Performance Benefits
//
// Batch application reduces:
//   - Individual style resolution calls
//   - Widget style map updates
//   - Memory allocations for style objects
//
// # Example Usage
//
//	theme.Apply(button, "button.primary", "focus", "hover", "disabled")
//	// Results in widget having styles for:
//	// - "button.primary" (base state)
//	// - "button.primary:focus" (focused state)
//	// - "button.primary:hover" (hovered state)
//	// - "button.primary:disabled" (disabled state)
//
// Parameters:
//   - widget: The widget to apply styles to
//   - selector: The base CSS-like selector for the widget
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

// Border retrieves a border style by name from the theme's border registry.
// Returns the zero value BorderStyle if the border name is not found.
func (m *MapTheme) Border(border string) BorderStyle {
	return m.borders[border]
}

// Cascade applies styles from the specified selector to the target style object.
// This is an internal method used by the Get() method to implement the cascading
// algorithm. If the selector exists in the theme, its style properties are
// merged into the target style, with the selector's properties taking precedence.
//
// Parameters:
//   - style: The target style to cascade properties into
//   - selector: The selector whose style should be cascaded
func (m *MapTheme) Cascade(style *Style, selector string) {
	other := m.styles[selector]
	style.Cascade(other)
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

// Define replaces the entire style registry with the provided map.
// This method is used for bulk theme loading and complete theme replacement.
//
// # Warning
//
// This method completely replaces the existing styles map, discarding
// any previously defined styles. Use Set() for individual style updates
// or SetStyles() for safer bulk operations.
//
// Parameters:
//   - styles: The new style registry to replace the current one
func (m *MapTheme) Define(styles map[string]*Style) {
	m.styles = styles
}

// Flag retrieves a boolean configuration flag from the theme's flag registry.
// Returns false if the flag name is not found.
func (m *MapTheme) Flag(flag string) bool {
	return m.flags[flag]
}

// Get retrieves the resolved style for the given selector using hierarchical cascading.
// This method implements the core style resolution algorithm that mimics CSS
// specificity rules, building the final style through progressive cascading
// from general to specific selectors.
//
// # Resolution Algorithm
//
// The method builds the final style through cascading in this order:
//  1. Default style (base foundation)
//  2. Empty selector ("") - global base styles
//  3. Type-only selectors (e.g., "button")
//  4. Type + part selectors (e.g., "button/text")
//  5. Class-only selectors (e.g., ".primary")
//  6. Type + class selectors (e.g., "button.primary")
//  7. Type + part + class selectors (e.g., "button/text.primary")
//  8. State-only selectors (e.g., ":focus")
//  9. Type + state selectors (e.g., "button:focus")
// 10. Type + class + state selectors (e.g., "button.primary:focus")
// 11. Type + part + state selectors (e.g., "button/text:focus")
// 12. Type + part + class + state selectors (e.g., "button/text.primary:focus")
// 13. ID-only selectors (e.g., "#submit") - highest specificity
// 14. ID + state selectors (e.g., "#submit:focus")
// 15. ID + part selectors (e.g., "#submit/text")
// 16. ID + part + state selectors (e.g., "#submit/text:focus")
//
// # Specificity Rules
//
// Higher specificity selectors override lower specificity ones:
//   - ID selectors have highest precedence
//   - Type + class + state combinations come next
//   - Type + class combinations follow
//   - Type-only selectors have lowest precedence
//
// # Cascading Process
//
// Each matching selector cascades its properties onto the result style,
// with more specific selectors overriding properties set by less specific ones.
// This ensures that the final style contains the most appropriate value
// for each style property.
//
// # Fallback Guarantee
//
// The method always returns a valid Style object, using the default style
// as the foundation. This prevents nil pointer exceptions and ensures
// consistent rendering even for undefined selectors.
//
// Parameters:
//   - selector: The CSS-like selector string to resolve (e.g., "button.primary:focus")
//
// Returns:
//   - Style: The fully resolved style with all applicable properties cascaded
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

// Set assigns a style to the specified selector in the theme registry.
// This method validates the selector format and stores the style for later
// retrieval during the cascading resolution process.
//
// # Validation
//
// The method validates the selector using the internal regex parser.
// Invalid selectors (those that don't match the expected CSS-like format)
// are silently ignored to prevent corrupting the theme registry.
//
// # Storage
//
// Valid selectors are stored exactly as provided, maintaining the original
// string format for efficient lookup during style resolution. The style
// object is stored by reference, allowing for shared style instances.
//
// # Usage Patterns
//
// Common usage includes:
//   - Defining base widget styles: Set("button", style)
//   - Creating themed variants: Set("button.primary", style)
//   - Specifying state styles: Set("button:focus", style)
//   - Complex combinations: Set("input/placeholder.large:focus", style)
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
func (m *MapTheme) SetStyles(styles map[string]*Style) {
	m.styles = styles
}

// Styles returns a direct reference to the theme's style registry.
// This allows for inspection and iteration over all defined styles.
func (m *MapTheme) Styles() map[string]*Style {
	return m.styles
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
