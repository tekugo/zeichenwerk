package zeichenwerk

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
)

// Regular expression to parse part:state style expressions.
//
// stylePartRegExp is a compiled regular expression used to parse style part
// selectors. It matches patterns like "part:state" where both the part and
// state components are optional and consist of alphanumeric characters,
// underscores, and hyphens.
//
// Pattern breakdown:
//   - ([0-9A-Za-z_\-]*): Captures the part name (optional)
//   - :?: Matches an optional colon separator
//   - ([0-9A-Za-z_\-]*): Captures the state name (optional)
//
// This regex is used internally by the Component to parse style selectors
// and extract part and state information for theme application.
var stylePartRegExp = regexp.MustCompile(`([0-9A-Za-z_\-]*):?([0-9A-Za-z_\-]*)`)

// Component provides a default implementation of the Widget interface.
// It serves as a foundation for creating custom widgets by providing common
// functionality such as bounds management, parent-child relationships, event
// handling and style handling. Most concrete widget implementations should
// embed Component to inherit this basic functionality.
//
// Components can be used itself as spacers, if they have no styling.
//
// Also do not forget to add it to the builder for easy building and styling.
type Component struct {
	id                  string              // widget identification datum
	class               string              // style class for styling
	parent              Container           // reference to the parent container
	x, y, width, height int                 // screen area of the widget (outer bounds)
	hwidth, hheight     int                 // preferred content size for Hint() sizing; containers may use negative values for fractional sizing
	states              map[Flag]bool       // map of internal boolean states like visible
	styles              map[string]*Style   // visual styling information
	handlers            map[Event][]Handler // event handlers
}

// Creates a new component.
func NewComponent(id, class string) *Component {
	return &Component{id: id, class: class}
}

// Apply applies a theme style to the component.
func (c *Component) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("component"))
}

// Bounds returns the widget's outer boundaries as (x, y, width, height).
// This includes the full area occupied by the widget including margins,
// borders, and padding. The coordinates are always absolute screen
// coordinates.
func (c *Component) Bounds() (int, int, int, int) {
	return c.x, c.y, c.width, c.height
}

// Class returns the style class of the component.
func (c *Component) Class() string {
	return c.class
}

// Content returns the widget's inner content area as (x, y, width, height).
// This calculates the available space for actual content by subtracting
// margins, padding, and border space from the outer bounds. The returned
// coordinates represent the area where the actual content should be rendered.
func (c *Component) Content() (int, int, int, int) {
	style := c.Style()
	if style == nil {
		return c.x, c.y, c.width, c.height
	} else {
		return c.x + style.Left(),
			c.y + style.Top(),
			c.width - style.Horizontal(),
			c.height - style.Vertical()
	}
}

// Cursor returns the current cursor position as (x, y) coordinates and style.
// The base implementation returns (0, 0, "") indicating that no cursor is
// displayed by default. Widgets that support text input or cursor navigation
// should override this method to return the actual cursor position and style.
//
// The returned position is relative to the content area of the widget,
// so 0,0 is the top-left corner of the content area.
func (c *Component) Cursor() (int, int, string) {
	return 0, 0, ""
}

// Dispatch dispatches an event to the widget. The event is a string that
// represents the type of event, such as a key press or mouse click.
// It iterates over all registered handlers for the event and calls them.
// The source widget is passed as a parameter, so that this method does not
// need to be overwritten to pass the correct widget type.
//
// Parameters:
//   - source: Source widget
//   - event: The type of event to dispatch.
//   - data: Optional data associated with the event.
func (c *Component) Dispatch(source Widget, event Event, data ...any) bool {
	if c.handlers == nil {
		return false
	}
	handled := false
	handlers, ok := c.handlers[event]
	if ok {
		for _, handler := range handlers {
			handled = handler(source, event, data...)
			if handled {
				return true
			}
		}
	}
	return false
}

// Flag returns the value of a boolean state flag.
// If the state has not been set, it returns false by default.
//
// Parameters:
//   - state: The name of the state to query
//
// Returns:
//   - bool: The current value of the state
func (c *Component) Flag(flag Flag) bool {
	if c.states == nil {
		return false
	}
	return c.states[flag]
}

// Hint returns the widget's preferred content size of the widget.
// The preferred size is just an attribute of the widget and can be set using
// SetHint(w, h). The preferred size is the size of the content area and must
// not include margin, border or padding. Some containers support negative for
// with or height for fractional sizing.
//
// By default, the preferred size is 0, 0, but if it is not set to a
// negative value, containers might calculate their preferred size and
// return the calculated values.
//
// Returns (content width, content height)
func (c *Component) Hint() (int, int) {
	return c.hwidth, c.hheight
}

// ID returns the unique identifier string for this widget instance.
// This identifier can be used for styling, debugging, testing, or widget
// lookup within the widget hierarchy. The ID should be unique within the
// scope of the application to ensure proper widget identification.
//
// Returns:
//   - string: The widget's unique identifier
func (c *Component) ID() string {
	return c.id
}

// Info returns a human-readable description of the widget's current state.
// The returned string includes: widget ID, parent ID (or "<nil>" if no parent),
// outer bounds (x,y,w,h), and content area bounds (cx,cy,cw,ch). It also reports
// counts of styles, handlers, and current state flags. This is primarily used for
// debugging and development purposes. The format is:
//
//	id=<id>, p=<parent-id>, bounds=(<x>.<y> <w>/<h>), content=(<cx>.<cy> <cw>/<ch>), styles=<count>, handlers=<count>, states=<map>
func (c *Component) Info() string {
	cx, cy, cw, ch := c.Content()
	parent := "<nil>"
	if c.parent != nil {
		parent = c.parent.ID()
	}
	return fmt.Sprintf("id=%s, p=%s, bounds=(%d.%d %d/%d), content=(%d.%d %d/%d), styles=%d, handlers=%d, states=%v",
		c.id, parent, c.x, c.y, c.width, c.height, cx, cy, cw, ch, len(c.styles), len(c.handlers), c.states)
}

// Log logs a debug message to the application's debug log.
// The base implementation delegates to the parent widget's Log method.
// If there is no parent, the message is ignored. The method supports
// formatted messages with optional parameters.
//
// Parameters:
//   - source: Source widget
//   - level: Log level
//   - msg: The debug message to log (can be a format string)
//   - params: Optional parameters for message formatting
func (c *Component) Log(source Widget, level Level, msg string, params ...any) {
	if c.parent != nil {
		c.parent.Log(source, level, msg, params...)
	}
}

// On registers an event handler for widget-specific events.
// This event handler is called, whenever the widget emits the specific event.
// Multiple handlers can be registered for the same event and they are called
// in *reverse* order of registration. So new handlers can consume or override
// events for pre-registered handlers.
//
// Parameters:
//   - event: The event to listen for
//   - handler: event handler function
func (c *Component) On(event Event, handler Handler) {
	if c.handlers == nil {
		c.handlers = make(map[Event][]Handler)
	}
	c.handlers[event] = append([]Handler{handler}, c.handlers[event]...)
}

// Parent returns the parent container of this widget.
// Returns nil if this widget has no parent (i.e., it's a root widget,
// or is not in the widget hierarchy yet).
func (c *Component) Parent() Container {
	return c.parent
}

// Refresh triggers a redraw of the widget and its visual representation.
// The base implementation delegates the refresh request to the parent widget,
// which eventually propagates to the root UI for screen updates. This method
// should be called whenever the widget's visual state changes and needs to
// be reflected on screen.
//
// It should be overridden by concrete widgets to optimize screen refreshes,
// as the base implementation redraws the whole UI.
func (c *Component) Refresh() {
	if c.parent != nil {
		c.parent.Refresh()
	}
}

// Render renders the widget to the screen using the Renderer.
//
// Component implements the basic rendering of the common widget parts like
// margin, border and padding. It also considers the state of the widget for
// rendering and style resolution.
//
// If the widget is hidden, rendering is skipped. The widget's state is
// determined by State() with priority: disabled > pressed > focused > hovered.
// The style for the state (e.g., ":focused") is looked up via Style() with
// fallback order: exact selector, then part, then state, then default.
//
// Parameters:
//   - r: The renderer to use for rendering the widget
func (c *Component) Render(r *Renderer) {
	// Check if the widget is visible
	if c.Flag(FlagHidden) {
		return
	}

	// Determine the style to use based on the widget state
	state := c.State()
	if state != "" {
		state = ":" + state
	}
	style := c.Style(state)
	r.Set(style.Foreground(), style.Background(), style.Font())

	// Clear the area covered by the widget
	if style.Background() != "" {
		margin := style.Margin()
		// Margin() never returns nil; if unset, returns NoInsets (all zeros)
		r.Fill(c.x+margin.Left, c.y+margin.Top, c.width-margin.Left-margin.Right, c.height-margin.Top-margin.Bottom, " ")
	}

	// Draw the border if specified
	border := style.Border()
	if border != "" && border != "none" {
		parts := strings.Split(border, " ")
		if len(parts) > 1 {
			border = parts[0]
			fg := parts[1]
			bg := style.Background()
			if len(parts) > 2 {
				bg = parts[2]
			}
			r.Set(fg, bg, "")
		} else {
			r.Set(style.Foreground(), style.Background(), "")
		}
		margin := style.Margin()
		r.Border(c.x+margin.Left, c.y+margin.Top, c.width-margin.Left-margin.Right, c.height-margin.Top-margin.Bottom, border)
		r.Set(style.Foreground(), style.Background(), style.Font())
	}
}

// Selector constructs a CSS-like selector string for styling widgets.
// It combines the widget type (t) with optional class and id modifiers.
// Format: "type.class#id" where class and id are optional.
// Example: "button.primary#submit" for a button with class "primary" and
// id "submit".
func (c *Component) Selector(t string) string {
	if c.class != "" {
		t = t + "." + c.class
	}
	if c.id != "" {
		t = t + "#" + c.id
	}
	return t
}

// SetBounds sets the widget's position and size as (x, y, width, height).
// This defines the outer boundaries of the widget including margins, borders,
// and padding. The coordinates are absolute screen coordinates.
//
// Parameters:
//   - x: The x-coordinate of the widget's position
//   - y: The y-coordinate of the widget's position
//   - width: The total width of the widget
//   - height: The total height of the widget
func (c *Component) SetBounds(x, y, width, height int) {
	c.x, c.y, c.width, c.height = x, y, width, height
}

// SetClass sets the style class of the component.
func (c *Component) SetClass(class string) {
	c.class = class
}

// SetFlag sets a boolean state flag for the widget.
// This can be used to track states like "hover", "focus", "checked", etc.
//
// Parameters:
//   - state: The name of the state
//   - value: The boolean value to set
func (c *Component) SetFlag(flag Flag, value bool) {
	if c.states == nil {
		c.states = make(map[Flag]bool)
	}
	c.states[flag] = value
}

// SetHint sets the sizing hint/preferred size of the widget.
// Some containers support negative values for fractional sizes.
//
// Parameters:
//   - width: preferred widget content width
//   - height: preferred widget content height
func (c *Component) SetHint(width, height int) {
	c.hwidth = width
	c.hheight = height
}

// SetParent establishes a parent-child relationship by setting the parent
// container. Pass nil to remove the widget from its current parent. This is
// typically called during widget hierarchy construction or when moving
// widgets between containers.
func (c *Component) SetParent(parent Container) {
	c.parent = parent
}

// State returns the current widget state based on the set flags.
func (c *Component) State() string {
	switch {
	case c.Flag(FlagDisabled):
		return string(FlagDisabled)
	case c.Flag(FlagPressed):
		return string(FlagPressed)
	case c.Flag(FlagFocused):
		return string(FlagFocused)
	case c.Flag(FlagHovered):
		return string(FlagHovered)
	}
	return ""
}

// SetStyle applies the given style configuration to the widget for a
// specific selector. The style controls visual appearance including colors,
// borders, margins, and padding. Selectors allow different styles for
// different widget states (e.g., "", ":state", "part:state").
//
// Parameters:
//   - selector: The style selector
//   - style: The style to apply, or nil to remove the style for the selector
func (c *Component) SetStyle(selector string, style *Style) {
	if c.styles == nil {
		c.styles = make(map[string]*Style)
	}
	if style == nil {
		delete(c.styles, selector)
	} else {
		c.styles[selector] = style
	}
}

// Style returns the style for the specified selector.
// If the requested selector is not found, it falls back to the default style.
//
// The style system supports CSS-like selectors for different widget states:
//   - "": default/base style
//   - ":focus": style when widget has keyboard focus
//   - ":hover": style when mouse is over the widget
//   - ":disabled": style when widget is disabled
//   - "part": style for a specific part of the widget
//   - "part:state": style for a specific part in a specific state
//
// The fallback order for a selector "part:state" is:
//   - exact match "part:state"
//   - part match "part"
//   - state match ":state"
//   - default ""
//
// Parameters:
//   - selector: The style selector to retrieve (e.g., "part", ":state", "part:state")
//
// Returns:
//   - *Style: The style configuration for the selector. Never returns nil; falls back to a default.
func (c *Component) Style(selector ...string) *Style {
	// If no selector is specified, we get the default style "" depending on the
	// current State()
	var actual string
	if len(selector) > 0 {
		actual = selector[0]
	} else {
		actual = c.State()
		if actual != "" {
			actual = ":" + actual
		}
	}

	// Ensure styles map exists
	if c.styles == nil {
		c.styles = make(map[string]*Style)
	}

	// Try exact match
	if style, ok := c.styles[actual]; ok {
		return style
	}

	// Fallback: if selector is "part:state", try "part" then ":state"
	parts := stylePartRegExp.FindStringSubmatch(actual)
	if len(parts) >= 3 {
		// Try part only (without state)
		if parts[1] != "" {
			if style, ok := c.styles[parts[1]]; ok {
				return style
			}
		}
		// Try state only (with colon)
		if parts[2] != "" {
			if style, ok := c.styles[":"+parts[2]]; ok {
				return style
			}
		}
	}

	// Ensure default style exists and return it
	if c.styles[""] == nil {
		c.styles[""] = NewStyle("")
	}
	return c.styles[""]
}

// Styles returns a list of all style selectors currently defined for this widget.
// This is useful for debugging, introspection, or iterating over all available
// widget styles. The returned slice contains selector strings like "", "part",
// ":focus", "part:focus", etc.
//
// Returns:
//   - []string: A slice of all style selector names defined for this widget
func (c *Component) Styles() []string {
	if c.styles == nil {
		return []string{}
	}
	return slices.Collect(maps.Keys(c.styles))
}
