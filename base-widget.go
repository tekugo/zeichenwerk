package tui

import (
	"fmt"
	"maps"
	"slices"

	"github.com/gdamore/tcell/v2"
)

// BaseWidget provides a default implementation of the Widget interface.
// It serves as a foundation for creating custom widgets by providing
// common functionality such as bounds management, parent-child relationships,
// event handling and style handling. Most concrete widget implementations
// should embed BaseWidget to inherit this basic functionality.
type BaseWidget struct {
	id                  string            // widget identification datum
	parent              Container         // reference to the parent container
	x, y, width, height int               // screen area of the widget (outer bounds)
	focusable           bool              // whether the widget can receive keyboard focus
	focussed            bool              // focus state for keyboard input
	hovered             bool              // hover state for mouse interaction
	styles              map[string]*Style // visual styling information
	handlers            map[string]func(Widget, string, ...any) bool
}

// Bounds returns the widget's outer boundaries as (x, y, width, height).
// This includes the full area occupied by the widget including margins, borders, and padding.
// The coordinates are always absolute screen coordinates.
func (w *BaseWidget) Bounds() (int, int, int, int) {
	return w.x, w.y, w.width, w.height
}

// Content returns the widget's inner content area as (x, y, width, height).
// This calculates the available space for actual content by subtracting
// margins, padding, and border space from the outer bounds. The returned
// coordinates represent the area where text or child widgets should be placed.
func (w *BaseWidget) Content() (int, int, int, int) {
	style := w.Style("")
	ix := w.x + style.Margin.Left + style.Padding.Left
	iy := w.y + style.Margin.Top + style.Padding.Top
	iw := w.width - style.Margin.Left - style.Margin.Right - style.Padding.Left - style.Padding.Right
	ih := w.height - style.Margin.Top - style.Margin.Bottom - style.Padding.Top - style.Padding.Bottom
	if style.Border != "" {
		ix++
		iy++
		iw -= 2
		ih -= 2
	}
	return ix, iy, iw, ih
}

// Cursor returns the current cursor position as (x, y) coordinates.
// The base implementation returns (-1, -1) indicating that no cursor
// is displayed by default. Widgets that support text input or cursor
// navigation should override this method to return the actual cursor position.
func (w *BaseWidget) Cursor() (int, int) {
	return -1, -1
}

// Emit triggers a custom event on the widget, calling any registered event handlers.
// This function is the core of the widget's event system, allowing widgets to notify
// listeners about state changes, user interactions, or other significant events.
//
// The function performs the following operations:
//   - Checks if any event handlers are registered for this widget
//   - Looks up the specific handler for the given event name
//   - Calls the handler function with the widget, event name, and any additional data
//   - Silently ignores the event if no handler is registered
//
// Event System Overview:
// The Emit function works in conjunction with the On() method to provide a flexible
// event-driven architecture. Widgets can emit events at any time, and external code
// can register handlers using On() to respond to these events.
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
//   - data: Optional additional data to pass to the event handler. The type and
//     meaning of this data depends on the specific event and widget type.
//
// Example usage:
//
//	// In a custom widget implementation
//	func (w *MyWidget) SetValue(value string) {
//		w.value = value
//		w.Emit("change", value) // Notify listeners of the change
//	}
//
//	// Emitting events with multiple data parameters
//	w.Emit("select", selectedIndex, selectedItem)
//
//	// Emitting simple notification events
//	w.Emit("activate")
//
// Event Handler Registration:
// Event handlers are registered using the On() method:
//
//	widget.On("change", func(w Widget, event string, data ...any) bool {
//		// Handle the change event
//		return true // Event was handled
//	})
//
// Note: If no handlers are registered or no handler exists for the specific event,
// the Emit call is silently ignored. This allows widgets to emit events freely
// without needing to check if anyone is listening.
//
// Attention: When using the BaseWidget Emit() method, only the BaseWidget is
// passed to the event handler function and it cannot be cast to the original
// widget type. If you want the original widget to be passed, you have to
// "copy" the Emit() method into it.
func (w *BaseWidget) Emit(event string, data ...any) {
	if w.handlers == nil {
		return
	}
	handler, found := w.handlers[event]
	if found {
		handler(w, event, data...)
	}
}

// Focusable returns whether the widget can receive keyboard focus.
// This determines if the widget can be included in focus navigation
// and receive keyboard input events. Non-focusable widgets (like labels)
// will be skipped during tab navigation.
//
// Returns:
//   - bool: true if the widget can receive focus, false otherwise
func (w *BaseWidget) Focusable() bool {
	return w.focusable
}

// Focussed returns the current focus state of the widget.
// A focused widget typically receives keyboard input and may display
// visual indicators such as highlighted borders or cursor.
//
// Returns:
//   - bool: true if the widget is currently focused, false otherwise
func (w *BaseWidget) Focussed() bool {
	return w.focussed
}

// Handle processes the given tcell.Event and returns whether it was consumed.
// The base implementation always returns false, indicating that no events
// are handled by default. Concrete widget implementations should override
// this method to handle specific events like keyboard input or mouse clicks.
//
// Event handling follows a consumption model: if a widget handles an event,
// it should return true to prevent the event from being passed to other widgets.
// If the widget doesn't handle the event, it should return false to allow
// event propagation to continue.
//
// Parameters:
//   - event: The tcell.Event to process (keyboard, mouse, resize, etc.)
//
// Returns:
//   - bool: true if the event was handled and consumed, false otherwise
func (w *BaseWidget) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		w.Emit("key", event)
	}
	return false
}

// Hint returns the widget's preferred size based on the default style.
// The base implementation retrieves the width and height from the default
// style ("") if available. If no style is set or no dimensions are specified
// in the style, it returns (0, 0) as a fallback.
//
// Concrete widget implementations should override this method to provide
// more sophisticated size calculations based on content, text metrics,
// or other widget-specific requirements.
//
// Returns:
//   - int: Preferred width from default style, or 0 if not specified
//   - int: Preferred height from default style, or 0 if not specified
func (w *BaseWidget) Hint() (int, int) {
	style := w.Style("")
	if style != nil {
		w := style.Width
		h := style.Height
		if w < 0 {
			w = 10
		}
		if h < 0 {
			h = 1
		}
		return w, h
	}
	return 0, 0
}

// Hovered returns the current hover state of the widget.
// A hovered widget is one that the mouse cursor is currently positioned over.
// This state is typically used for visual feedback such as highlighting
// or showing tooltips when the mouse is over the widget.
//
// Returns:
//   - bool: true if the widget is currently hovered, false otherwise
func (w *BaseWidget) Hovered() bool {
	return w.hovered
}

// ID returns the unique identifier string for this widget instance.
// This identifier can be used for debugging, testing, or widget lookup
// within a widget hierarchy. The ID should be unique within the scope
// of the application to ensure proper widget identification.
//
// Returns:
//   - string: The widget's unique identifier
func (w *BaseWidget) ID() string {
	return w.id
}

// Info returns a human-readable description of the widget's current state.
// The base implementation returns basic information about the widget's
// type, ID, and bounds. This is primarily used for debugging and development.
func (w *BaseWidget) Info() string {
	cx, cy, cw, ch := w.Content()
	flags := make([]string, 0)
	parent := "<nil>"
	if w.focusable {
		flags = append(flags, "focusable")
	}
	if w.focussed {
		flags = append(flags, "focussed")
	}
	if w.hovered {
		flags = append(flags, "hovered")
	}
	if w.parent != nil {
		parent = w.parent.ID()
	}
	return fmt.Sprintf("id=%s, p=%s, bounds=(%d.%d %d/%d), content=(%d.%d %d/%d), styles=%d, handlers=%d, flags=%s",
		w.id, parent, w.x, w.y, w.width, w.height, cx, cy, cw, ch, len(w.styles), len(w.handlers), flags)
}

// Log logs a debug message to the application's debug log.
// The base implementation delegates to the parent widget's Log method.
// If there is no parent, the message is ignored. The method supports
// formatted messages with optional parameters.
//
// Parameters:
//   - msg: The debug message to log (can be a format string)
//   - params: Optional parameters for message formatting
func (w *BaseWidget) Log(msg string, params ...any) {
	if w.parent != nil {
		w.parent.Log(msg, params...)
	}
}

// On registers an event handler for widget-specific events.
// This event handler is called, whenever the widget emits an event.
//
// Parameters:
//   - handler: event handler function
func (w *BaseWidget) On(event string, handler func(Widget, string, ...any) bool) {
	if w.handlers == nil {
		w.handlers = make(map[string]func(Widget, string, ...any) bool)
	}
	w.handlers[event] = handler
}

// Parent returns the parent widget in the widget hierarchy.
// Returns nil if this widget has no parent (i.e., it's a root widget).
func (w *BaseWidget) Parent() Widget {
	return w.parent
}

// Position returns the widget's current position as (x, y) coordinates.
// These coordinates represent the top-left corner of the widget's outer bounds
// relative to its parent container or the screen.
//
// Returns:
//   - int: The x-coordinate of the widget's position
//   - int: The y-coordinate of the widget's position
func (w *BaseWidget) Position() (int, int) {
	return w.x, w.y
}

// Refresh triggers a redraw of the widget and its visual representation.
// The base implementation delegates the refresh request to the parent widget,
// which eventually propagates to the root UI for screen updates. This method
// should be called whenever the widget's visual state changes and needs to
// be reflected on screen.
//
// Concrete widget implementations may override this method to perform
// widget-specific refresh operations before delegating to the parent.
func (w *BaseWidget) Refresh() {
	if w.parent != nil {
		w.parent.Refresh()
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
func (w *BaseWidget) SetBounds(x, y, width, height int) {
	w.x, w.y, w.width, w.height = x, y, width, height
}

// SetFocusable sets whether the widget can receive keyboard focus.
// This controls the widget's participation in focus navigation and
// keyboard input handling. Setting this to false will exclude the
// widget from tab navigation and prevent it from receiving focus.
//
// Parameters:
//   - focusable: true to allow the widget to receive focus, false to prevent it
func (w *BaseWidget) SetFocusable(focusable bool) {
	w.focusable = focusable
}

// SetFocussed sets the focus state of the widget.
// When a widget gains focus, it typically becomes the target for keyboard input
// and may update its visual appearance to indicate the focused state.
//
// Parameters:
//   - focussed: true to focus the widget, false to unfocus it
func (w *BaseWidget) SetFocussed(focussed bool) {
	if focussed {
		w.Emit("focus")
	} else {
		w.Emit("blur")
	}
	w.focussed = focussed
}

// SetHint sets the sizing hint of the widget.
// The sizing hint is part of the default styling of the widget and with
// this method, it can be set dynamically according e.g. to label width.
//
// Parameters:
//   - width: preferred widget content width
//   - height: preferred widget content height
func (w *BaseWidget) SetHint(width, height int) {
	style := w.Style("")
	if style != nil {
		style.Width = width
		style.Height = height
	}
}

// SetHovered sets the hover state of the widget.
// This method is typically called by the application's mouse event handling
// system when the mouse cursor enters or leaves the widget's bounds.
// Widgets may use this state to provide visual feedback to users.
//
// Parameters:
//   - hovered: true when the mouse is over the widget, false when it leaves
func (w *BaseWidget) SetHovered(hovered bool) {
	if hovered {
		w.Emit("hover")
	}
	w.hovered = hovered
}

// SetParent establishes a parent-child relationship by setting the parent widget.
// Pass nil to remove the widget from its current parent. This is typically
// called during widget hierarchy construction or when moving widgets between containers.
func (w *BaseWidget) SetParent(parent Container) {
	w.parent = parent
}

// SetPosition sets the widget's position to the specified coordinates.
// This method updates the widget's location within its parent container
// or on the screen, affecting where the widget will be rendered.
//
// Parameters:
//   - x: The new x-coordinate for the widget's position
//   - y: The new y-coordinate for the widget's position
func (w *BaseWidget) SetPosition(x, y int) {
	w.x = x
	w.y = y
}

// SetSize sets the content size of the widget, automatically calculating outer bounds.
// This method takes the desired content dimensions and adds margins, padding, and
// border space to determine the widget's total outer bounds. The content size
// represents the actual usable area for text or child widgets.
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
func (w *BaseWidget) SetSize(width, height int) {
	style := w.Style("")
	if style != nil {
		w.width = width + style.Margin.Left + style.Margin.Right + style.Padding.Left + style.Padding.Right
		w.height = height + style.Margin.Top + style.Margin.Bottom + style.Padding.Top + style.Padding.Bottom
		if style.Border != "" {
			w.width += 2
			w.height += 2
		}
	} else {
		w.width = width
		w.height = height
	}
}

// SetStyle applies the given style configuration to the widget for a specific selector.
// The style controls visual appearance including colors, borders, margins, and padding.
// Selectors allow different styles for different widget states (e.g., "", "focus", "hover").
//
// Parameters:
//   - selector: The style selector (e.g., "", "focus", "hover", "disabled")
//   - style: The style configuration to apply, or nil to remove the style for this selector
func (w *BaseWidget) SetStyle(selector string, style *Style) {
	if w.styles == nil {
		w.styles = make(map[string]*Style)
	}
	if style == nil {
		delete(w.styles, selector)
	} else {
		w.styles[selector] = style
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
func (w *BaseWidget) Size() (int, int) {
	style := w.Style("")
	if style != nil {
		width := w.width - style.Margin.Left - style.Margin.Right - style.Padding.Left - style.Padding.Right
		height := w.height - style.Margin.Top - style.Margin.Bottom - style.Padding.Top - style.Padding.Bottom
		if style.Border != "" {
			width -= 2
			height -= 2
		}
		return width, height
	} else {
		return w.width, w.height
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
func (w *BaseWidget) State() string {
	if w.focussed {
		return "focus"
	} else {
		return ""
	}
}

// Style returns the style configuration for the specified selector.
// If the requested selector is not found, it falls back to the default style ("").
// Returns nil if no styles have been set for the widget.
//
// The style system supports CSS-like selectors for different widget states:
//   - "": default/base style
//   - "focus": style when widget has keyboard focus
//   - "hover": style when mouse is over the widget
//   - "disabled": style when widget is disabled
//
// Parameters:
//   - selector: The style selector to retrieve (e.g., "", "focus", "hover")
//
// Returns:
//   - *Style: The style configuration for the selector, or nil if not found
func (w *BaseWidget) Style(selector string) *Style {
	if w.styles == nil {
		return nil
	}
	style, ok := w.styles[selector]
	if ok {
		return style
	} else {
		return w.styles[""]
	}
}

// Styles returns a list of all style selectors currently defined for this widget.
// This is useful for debugging, introspection, or iterating over all available
// widget styles. The returned slice contains selector strings like "", "focus", "hover", etc.
//
// Returns:
//   - []string: A slice of all style selector names defined for this widget
func (w *BaseWidget) Styles() []string {
	return slices.Collect(maps.Keys(w.styles))
}
