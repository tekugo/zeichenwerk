package zeichenwerk

import (
	"regexp"

	"github.com/gdamore/tcell/v2"
)

// CSS-like selector pattern: <type>.class#id:part
// This regular expression parses selectors in the format:
// - type: widget type (e.g., "button", "input")
// - class: optional class name preceded by '.'
// - id: optional ID preceded by '#'
// - part: optional part/state preceded by ':' (e.g., "focus", "hover")
var (
	re, _        = regexp.Compile(`([0-9A-Za-z_\-]*)\.?([0-9A-Za-z_\-]*)#?([0-9A-Za-z_\-]*):?([0-9A-Za-z_\-]*)`)
	defaultStyle = NewStyle("green", "black").SetMargin(0).SetPadding(0)
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
	Color(string) tcell.Color

	Get(string) Style

	// Set assigns a style to the specified selector in the theme.
	// This allows for dynamic theme customization and style registration.
	//
	// Parameters:
	//   - string: The CSS-like selector to associate with the style
	//   - *Style: The style configuration to assign
	Set(string, *Style)

	// SetStyles applies theme styles to a widget for multiple states/parts.
	// This is a convenience method that applies the base style and any
	// additional part-specific styles (e.g., focus, hover) to the widget.
	//
	// Parameters:
	//   - Widget: The widget to apply styles to
	//   - string: The base selector for the widget
	//   - ...string: Additional parts/states to apply (e.g., "focus", "hover")
	SetStyles(Widget, string, ...string)
}
