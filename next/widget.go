package next

// Widget represents a UI component that can be rendered and interact with
// user input. All UI elements in the TUI framework must implement this
// interface to participate in the rendering pipeline, event handling system,
// and layout management.
//
// All widgets share common functionality through BaseWidget, which provides
// default implementations for most interface methods.
type Widget interface {
	// Bounds returns the widget's position and size as absolute screen coordinates.
	// Returns (x, y, width, height).
	Bounds() (int, int, int, int)

	// Content returns the widget's content area as absolute screen coordinates.
	// The content area is the inner area of the widget where the widget's content
	// is rendered, excluding margin, border and padding.
	// Returns (x, y, width, height).
	Content() (int, int, int, int)

	// Cursor returns the widget's cursor position as relative coordinates to the
	// upper left corner of the widget's content area. If the cursor should not be
	// displayed, returns (0, 0, "").
	// Returns (x, y, cursor-style).
	Cursor() (int, int, string)

	// Dispatch dispatches an event to the widget. The event is a struct that
	// represents a user input event, such as a key press or mouse click.
	//
	// Parameters:
	//   - event: The type of event to dispatch.
	//   - data: Optional data associated with the event.
	Dispatch(event string, data ...any) bool

	// Flag returns a widget's state flag.
	Flag(string) bool

	// Hint returns the widgets preferred content size for optimal display.
	// Container may use this as guidance for the content layout. Values can be:
	//   - Fixed size (positive value)
	//   - Fractional size (negative value)
	//   - Automatic/available space (zero)
	Hint() (int, int)

	// ID returns the widget's unique identifier. The ID is used for widget lookup,
	// but can also be used for debugging, testing and logging. The ID should be
	// globally unique.
	ID() string

	// Info returns a human-readable description of the widget and its current state.
	Info() string

	// Log los a debug message to the application's debug log.
	//
	// Parameters:
	//   - widget: The widget that is logging the message.
	//   - level: The severity level of the message.
	//   - message: The message to log.
	//   - data: Optional data associated with the message.
	Log(widget Widget, level string, message string, data ...any)

	// On registers an event handler for widget-specific actions.
	//
	// Parameters:
	//   - event: The type of event to handle.
	//   - handler: The handler function to register.
	On(string, Handler)

	// Parent returns the parent container of this widget. Returns nil if the widget
	// has no parent.
	Parent() Container

	// Refresh triggers a redraw of the widget.
	// This should be called when the widget's visual state has changed
	// and needs to be updated on the screen.
	Refresh()

	// Render renders the widget to the screen using the Renderer.
	Render(r *Renderer)

	// SetBounds sets the widget's position and size as absolute screen coordinates.
	SetBounds(x, y, width, height int)

	// SetFlag sets a widget's state flag.
	//
	// Parameters:
	//   - state: The state flag to set.
	//   - value: The value to set the state flag to.
	SetFlag(string, bool)

	// SetHint sets the widget's preferred content size for optimal display.
	//
	// Parameters:
	//   - width: The preferred width of the widget.
	//   - height: The preferred height of the widget.
	SetHint(width, height int)

	// SetParent sets the parent container of this widget.
	//
	// Parameters:
	//   - parent: The parent container of this widget.
	SetParent(parent Container)

	// SetStyle applies the given style to the widget for a specific selector.
	// Controls visual appearance such as colors, borders, and text formatting.
	//
	// Parameters:
	//   - string: Style selector (e.g., "", ":focus", "bar", "bar:hover")
	//   - *Style: The style configuration to apply, or nil to remove
	SetStyle(string, *Style)

	// State returns the widget state for rendering.
	State() string

	// Style returns the widget's style for the given selector.
	// The style may consist of a part and the state prefixed by a colon. The
	// rendering method decides, what style to use, but the widget should return
	// sensible defaults for styles not set. So if there is no part:state
	// combination, first use the part, then the state.
	//
	// Parameters:
	//   - selector: The selector to get the style for or nothing for default.
	Style(...string) *Style
}
