package zeichenwerk

import (
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/gdamore/tcell/v2"
)

// RE to parse part:state style expressions
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
// This regex is used internally by the BaseWidget to parse style selectors
// and extract part and state information for theme application.
var stylePartRegExp, _ = regexp.Compile(`([0-9A-Za-z_\-]*):?([0-9A-Za-z_\-]*)`)

// BaseWidget provides a default implementation of the Widget interface.
// It serves as a foundation for creating custom widgets by providing
// common functionality such as bounds management, parent-child relationships,
// event handling and style handling. Most concrete widget implementations
// should embed BaseWidget to inherit this basic functionality.
//
// The Widget is not responsible for rendering, so all rendering is done in
// the renderer. For new widgets, you also have to extend the renderer to be
// able to render the widget.
//
// Also do not forget to add it to the builder for building and styling.
type BaseWidget struct {
	id                    string            // widget identification datum
	parent                Container         // reference to the parent container
	x, y, width, height   int               // screen area of the widget (outer bounds)
	widthHint, heightHint int               // preferred content width and height
	focusable             bool              // whether the widget can receive keyboard focus
	focused               bool              // focus state for keyboard input
	hovered               bool              // hover state for mouse interaction
	styles                map[string]*Style // visual styling information
	handlers              map[string]func(Widget, string, ...any) bool
}

// Bounds returns the widget's outer boundaries as (x, y, width, height).
// This includes the full area occupied by the widget including margins,
// borders, and padding. The coordinates are always absolute screen
// coordinates.
func (bw *BaseWidget) Bounds() (int, int, int, int) {
	return bw.x, bw.y, bw.width, bw.height
}

// Content returns the widget's inner content area as (x, y, width, height).
// This calculates the available space for actual content by subtracting
// margins, padding, and border space from the outer bounds. The returned
// coordinates represent the area where the actual content should be rendered.
func (bw *BaseWidget) Content() (int, int, int, int) {
	style := bw.Style()
	if style == nil {
		return bw.x, bw.y, bw.width, bw.height
	}
	if style.Margin() != nil && style.Padding() != nil {
		ix := bw.x + style.Margin().Left + style.Padding().Left
		iy := bw.y + style.Margin().Top + style.Padding().Top
		iw := bw.width - style.Horizontal()
		ih := bw.height - style.Vertical()
		border := style.Border()
		if border != "" && border != "none" {
			ix++
			iy++
		}
		return ix, iy, iw, ih
	} else {
		return bw.x, bw.y, bw.width, bw.height
	}
}

// Cursor returns the current cursor position as (x, y) coordinates.
// The base implementation returns (-1, -1) indicating that no cursor
// is displayed by default. Widgets that support text input or cursor
// navigation should override this method to return the actual cursor
// position.
//
// The returned position is relative to the content area of the widget,
// so 0,0 is the top-left corner of the content area.
func (bw *BaseWidget) Cursor() (int, int) {
	return -1, -1
}

// Emit triggers a custom event on the widget, calling any registered event
// handlers. This function is the core of the widget's event system, allowing
// widgets to notify listeners about state changes, user interactions, or
// other significant events.
//
// The function performs the following operations:
//   - Checks if any event handlers are registered for this widget
//   - Looks up the specific handler for the given event name
//   - Calls the handler function with the widget, event name, and any
//     optional data
//   - Silently ignores the event if no handler is registered
//
// Event System Overview:
// The Emit function works in conjunction with the On() method to provide a
// flexible event-driven architecture. Widgets can emit events at any time,
// and external code can register handlers using On() to respond to these events.
//
// Common widget events include:
//   - "change": Content or state has been modified
//   - "focus": Widget has gained keyboard focus
//   - "blur": Widget has lost keyboard focus
//   - "select": Item selection has changed (for lists, menus)
//   - "activate": Item has been activated (Enter key, double-click)
//   - "key": Raw keyboard event (automatically emitted by Handle())
//
// Parameters:
//   - event: The name of the event to emit (e.g., "change", "focus", "select")
//   - data: Optional additional data to pass to the event handler. The type
//     and meaning of this data depends on the specific event and widget type.
//
// Attention: When using the BaseWidget Emit() method, only the BaseWidget is
// passed to the event handler function and it cannot be cast to the original
// widget type. If you want the original widget to be passed, you have to
// "copy" the Emit() method into it.
func (bw *BaseWidget) Emit(event string, data ...any) bool {
	if bw.handlers == nil {
		return false
	}
	handler, found := bw.handlers[event]
	if found {
		return handler(bw, event, data...)
	}
	return false
}

// Focusable returns whether the widget can receive keyboard focus.
// This determines if the widget can be included in focus navigation
// and receive keyboard input events. Non-focusable widgets (like labels)
// will be skipped during tab navigation.
//
// Returns:
//   - bool: true if the widget can receive focus, false otherwise
func (bw *BaseWidget) Focusable() bool {
	return bw.focusable
}

// Focused returns the current focus state of the widget.
// A focused widget typically receives keyboard input and may display
// visual indicators such as highlighted borders or cursor.
//
// Returns:
//   - bool: true if the widget is currently focused, false otherwise
func (bw *BaseWidget) Focused() bool {
	return bw.focused
}

// Handle processes the given tcell.Event and returns whether it was consumed.
// The base implementation always returns false, indicating that no events
// are handled by default. Concrete widget implementations should override
// this method to handle specific events like keyboard input or mouse clicks.
//
// Event handling follows a consumption model: if a widget handles an event,
// it should return true to prevent the event from being passed to the parent.
// If the widget doesn't handle the event, it should return false to allow
// event propagation to continue.
//
// Parameters:
//   - event: The tcell.Event to process (keyboard, mouse, resize, etc.)
//
// Returns:
//   - bool: true if the event was handled and consumed, false otherwise
func (bw *BaseWidget) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		return bw.Emit("key", event)
	}
	return false
}

// Hint returns the widget's preferred content size of the widget.
// THe preferred size is just an attribute of the widget and can be set
// using SetHint(w, h). The preferred size is the size of the content area
// and must not include margin, border or padding. Some containers support
// negative for with or height for fractional sizing.
//
// By default, the preferred size is 0, 0, but if it is not set to a
// negative value, containers might calculate their preferred size and
// return the calculated values.
//
// Returns:
//   - int: preferred width, might be negative for fractional sizes
//   - int: preferred height, might be negative for fractional sizes
func (bw *BaseWidget) Hint() (int, int) {
	return bw.widthHint, bw.heightHint
}

// Hovered returns the current hover state of the widget.
// A hovered widget is one that the mouse cursor is currently positioned over.
// This state is typically used for visual feedback such as highlighting
// or showing tooltips when the mouse is over the widget.
//
// Returns:
//   - bool: true if the widget is currently hovered, false otherwise
func (bw *BaseWidget) Hovered() bool {
	return bw.hovered
}

// ID returns the unique identifier string for this widget instance.
// This identifier can be used for styling, debugging, testing, or widget
// lookup within the widget hierarchy. The ID should be unique within the
// scope of the application to ensure proper widget identification.
//
// Returns:
//   - string: The widget's unique identifier
func (bw *BaseWidget) ID() string {
	return bw.id
}

// Info returns a human-readable description of the widget's current state.
// The base implementation returns basic information about the widget's
// type, ID, and bounds. This is primarily used for debugging and development.
func (bw *BaseWidget) Info() string {
	cx, cy, cw, ch := bw.Content()
	flags := make([]string, 0)
	parent := "<nil>"
	if bw.focusable {
		flags = append(flags, "focusable")
	}
	if bw.focused {
		flags = append(flags, "focused")
	}
	if bw.hovered {
		flags = append(flags, "hovered")
	}
	if bw.parent != nil {
		parent = bw.parent.ID()
	}
	return fmt.Sprintf("id=%s, p=%s, bounds=(%d.%d %d/%d), content=(%d.%d %d/%d), styles=%d, handlers=%d, flags=%s",
		bw.id, parent, bw.x, bw.y, bw.width, bw.height, cx, cy, cw, ch, len(bw.styles), len(bw.handlers), flags)
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
func (bw *BaseWidget) Log(source Widget, level string, msg string, params ...any) {
	if bw.parent != nil {
		bw.parent.Log(source, level, msg, params...)
	}
}

// On registers an event handler for widget-specific events.
// This event handler is called, whenever the widget emits an event.
// Currently, only one handler per event is supported. If a second handler is
// registered, it will overwrite the previous one.
//
// Parameters:
//   - handler: event handler function
func (bw *BaseWidget) On(event string, handler func(Widget, string, ...any) bool) {
	if bw.handlers == nil {
		bw.handlers = make(map[string]func(Widget, string, ...any) bool)
	}
	bw.handlers[event] = handler
}

// Parent returns the parent widget in the widget hierarchy.
// Returns nil if this widget has no parent (i.e., it's a root widget,
// or is not in the widget hierarchy yet).
func (bw *BaseWidget) Parent() Widget {
	return bw.parent
}

// Position returns the widget's current position as (x, y) coordinates.
// These coordinates represent the top-left corner of the widget's outer bounds
// as absolute screen coordinates.
//
// Returns:
//   - int: The x-coordinate of the widget's position
//   - int: The y-coordinate of the widget's position
func (bw *BaseWidget) Position() (int, int) {
	return bw.x, bw.y
}

// Refresh triggers a redraw of the widget and its visual representation.
// The base implementation delegates the refresh request to the parent widget,
// which eventually propagates to the root UI for screen updates. This method
// should be called whenever the widget's visual state changes and needs to
// be reflected on screen.
//
// It should be overridden by concrete widgets to optimize screen refreshes,
// as the base implementation redraws the whole UI.
func (bw *BaseWidget) Refresh() {
	if bw.parent != nil {
		bw.parent.Refresh()
	}
}

// SetBounds sets the widget's position and size as (x, y, width, height).
// This defines the outer boundaries of the widget including margins, borders,
// and padding. The coordinates are relative to the parent container or screen.
//
// Parameters:
//   - x: The x-coordinate of the widget's position
//   - y: The y-coordinate of the widget's position
//   - width: The total width of the widget
//   - height: The total height of the widget
func (bw *BaseWidget) SetBounds(x, y, width, height int) {
	bw.x, bw.y, bw.width, bw.height = x, y, width, height
}

// SetFocusable sets whether the widget can receive keyboard focus.
// This controls the widget's participation in focus navigation and
// keyboard input handling. Setting this to false will exclude the
// widget from tab navigation and prevent it from receiving focus.
//
// Parameters:
//   - focusable: true to allow the widget to receive focus, false to prevent it
func (bw *BaseWidget) SetFocusable(focusable bool) {
	bw.focusable = focusable
}

// SetFocused sets the focus state of the widget.
// When a widget gains focus, it typically becomes the target for keyboard input
// and may update its visual appearance to indicate the focused state. Changing
// the focus makes the widget to emit "blur" and "focus" events.
//
// Parameters:
//   - focused: true to focus the widget, false to unfocus it
func (bw *BaseWidget) SetFocused(focused bool) {
	if focused {
		bw.Emit("focus")
	} else {
		bw.Emit("blur")
	}
	bw.focused = focused
}

// SetHint sets the sizing hint/preferred size of the widget.
// Some containers support negative values for fractional sizes.
//
// Parameters:
//   - width: preferred widget content width
//   - height: preferred widget content height
func (bw *BaseWidget) SetHint(width, height int) {
	bw.widthHint = width
	bw.heightHint = height
}

// SetHovered sets the hover state of the widget.
// This method is typically called by the application's mouse event handling
// system when the mouse cursor enters or leaves the widget's bounds.
// Widgets may use this state to provide visual feedback to users.
//
// Parameters:
//   - hovered: true when the mouse is over the widget, false when it leaves
func (bw *BaseWidget) SetHovered(hovered bool) {
	if hovered {
		bw.Emit("hover")
	}
	bw.hovered = hovered
}

// SetParent establishes a parent-child relationship by setting the parent
// widget. Pass nil to remove the widget from its current parent. This is
// typically called during widget hierarchy construction or when moving
// widgets between containers.
func (bw *BaseWidget) SetParent(parent Container) {
	bw.parent = parent
}

// SetPosition sets the widget's position to the specified coordinates.
// This method updates the widget's location on the screen, affecting where
// the widget will be rendered. The widget coordinates are normally absolute
// screen coordinates (exceptions are e.g. Scroller).
//
// Parameters:
//   - x: The new x-coordinate for the widget's position
//   - y: The new y-coordinate for the widget's position
func (bw *BaseWidget) SetPosition(x, y int) {
	bw.x = x
	bw.y = y
}

// SetSize sets the content size of the widget, automatically calculating
// outer bounds. This method takes the desired content dimensions and adds
// margins, padding, and border space to determine the widget's total outer
// bounds. The content size represents the actual usable area for content or
// child widgets.
//
// The calculation includes:
//   - Content size (the specified width and height)
//   - Padding (inner spacing around content)
//   - Margins (outer spacing around the widget)
//   - Border space (if a border is present, adds 2 to each dimension)
//
// Parameters:
//   - width: The desired content width in characters/cells
//   - height: The desired content height in characters/cells
func (bw *BaseWidget) SetSize(width, height int) {
	style := bw.Style()
	if style != nil {
		bw.width = width + style.Horizontal()
		bw.height = height + style.Vertical()
	} else {
		bw.width = width
		bw.height = height
	}
}

// SetStyle applies the given style configuration to the widget for a
// specific selector. The style controls visual appearance including colors,
// borders, margins, and padding. Selectors allow different styles for
// different widget states (e.g., "", "focus", "hover").
//
// Parameters:
//   - selector: The style selector (e.g., "", "focus", "hover", "disabled")
//   - style: The style to apply, or nil to remove the style for the selector
func (bw *BaseWidget) SetStyle(selector string, style *Style) {
	if bw.styles == nil {
		bw.styles = make(map[string]*Style)
	}
	if style == nil {
		delete(bw.styles, selector)
	} else {
		bw.styles[selector] = style
	}
}

// Size returns the widget's current content size as (width, height).
// This method calculates the available content area by subtracting margins,
// padding, and border space from the widget's outer bounds. The returned
// size represents the actual space available for content rendering.
//
// The calculation considers:
//   - Margins (outer spacing)
//   - Padding (inner spacing)
//   - Borders (if present, reduces size by 2 in each dimension)
//
// Returns:
//   - int: The content width in characters/cells
//   - int: The content height in characters/cells
func (bw *BaseWidget) Size() (int, int) {
	style := bw.Style()
	if style != nil {
		return bw.width - style.Horizontal(), bw.height - style.Vertical()
	} else {
		return bw.width, bw.height
	}
}

// State returns the current state of the widget for styling purposes.
// The base implementation returns "focus" when the widget is focused,
// and an empty string for the default state. The hover state is typically
// handled by concrete widget implementations that may override this method
// to return "hover" or combined states like "focus:hover".
//
// Common states:
//   - "" (empty): default state
//   - "focus": widget has keyboard focus
//   - "hover": mouse is over the widget (handled by concrete implementations)
//
// Returns:
//   - string: The current widget state identifier for styling
func (bw *BaseWidget) State() string {
	if bw.focused {
		return "focus"
	} else {
		return ""
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
//
// Parameters:
//   - params: The style selector to retrieve (e.g., "part", ":state", "part:state")
//
// Returns:
//   - *Style: The style configuration for the selector, or nil if not found
func (bw *BaseWidget) Style(params ...string) *Style {
	// If no parameter is specified, we get the default style ""
	selector := ""
	if len(params) > 0 {
		selector = params[0]
	}

	// If no style is set, we create an empty default one
	if bw.styles == nil {
		bw.styles = make(map[string]*Style)
		bw.styles[""] = NewStyle("")
	}

	style, ok := bw.styles[selector]
	if ok {
		return style
	} else {
		parts := stylePartRegExp.FindStringSubmatch(selector)
		if style, ok = bw.styles[":"+parts[2]]; ok {
			return style
		} else if style, ok = bw.styles["/"+parts[1]]; ok {
			return style
		} else {
			return bw.styles[""]
		}
	}
}

// Styles returns a list of all style selectors currently defined for this widget.
// This is useful for debugging, introspection, or iterating over all available
// widget styles. The returned slice contains selector strings like "", "part",
// ":focus", "part:focus", etc.
//
// Returns:
//   - []string: A slice of all style selector names defined for this widget
func (bw *BaseWidget) Styles() []string {
	return slices.Collect(maps.Keys(bw.styles))
}
