package zeichenwerk

import (
	"github.com/gdamore/tcell/v2"
)

// Widget represents a UI component that can be rendered and interact with user input.
// All UI elements in the TUI framework must implement this interface to participate
// in the rendering pipeline and event handling system.
type Widget interface {
	// Bounds returns the widget's position and size as (x, y, width, height).
	// These coordinates define the widget's outer boundaries including any borders or padding.
	Bounds() (int, int, int, int)

	// Content returns the widget's content area as (x, y, width, height).
	// This represents the inner area available for actual content, excluding borders and padding.
	Content() (int, int, int, int)

	// Cursor returns the current cursor position as (x, y) coordinates.
	// Returns (-1, -1) if the widget doesn't support or currently show a cursor.
	Cursor() (int, int)

	// Focusable returns the focusability of the widget.
	Focusable() bool

	// Focused returns the focus state of the widget.
	// A focused widget typically receives keyboard input and may have
	// visual indicators such as highlighted borders or cursor display.
	//
	// Returns:
	//   - bool: true if the widget is currently focused, false otherwise
	Focused() bool

	// Handle processes the given tcell.Event and returns true if the event was consumed.
	// Events include keyboard input, mouse events, and resize events.
	// If the widget handles the event, it should return true to prevent further propagation.
	Handle(tcell.Event) bool

	// Hint returns the widget's preferred size for optimal display.
	// This represents the ideal dimensions the widget would like to have,
	// taking into account its content, styling, and layout requirements.
	// Layout containers should use this as guidance when allocating space.
	//
	// Returns:
	//   - int: Preferred width in characters/cells
	//   - int: Preferred height in characters/cells
	Hint() (int, int)

	// Hovered returns the hover state of the widget.
	// A hovered widget is one that the mouse cursor is currently positioned over.
	// This state is typically used for visual feedback such as highlighting
	// or showing tooltips when the mouse is over the widget.
	//
	// Returns:
	//   - bool: true if the widget is currently hovered, false otherwise
	Hovered() bool

	// ID returns a unique identifier for this widget instance.
	// This can be used for debugging, testing, or widget lookup purposes.
	ID() string

	// Info returns a human-readable description of the widget's current state.
	// This is primarily used for debugging and development purposes.
	Info() string

	// Log logs a debug message to the application's debug log.
	// Messages are typically displayed in debug mode and can be useful
	// for troubleshooting widget behavior and state changes. The method
	// supports formatted messages with optional parameters.
	//
	// Parameters:
	//   - string: The debug message to log (can be a format string)
	//   - ...any: Optional parameters for message formatting
	Log(string, ...any)

	// On registers an action handler, which is called, whenever a
	// widget-specific action occurs. The handler is called for every
	// action.
	On(string, func(Widget, string, ...any) bool)

	// Parent returns the parent container in the widget hierarchy.
	// Returns nil if this is a root widget or has no parent.
	Parent() Widget

	// Position returns the widget's current position as (x, y) coordinates.
	// These coordinates represent the top-left corner of the widget's outer bounds
	// relative to its parent container or the screen.
	//
	// Returns:
	//   - int: The x-coordinate of the widget's position
	//   - int: The y-coordinate of the widget's position
	Position() (int, int)

	// Refresh triggers a redraw of the widget.
	// This should be called when the widget's visual state has changed
	// and needs to be updated on the screen.
	Refresh()

	// SetBounds sets the widget's position and size as (x, y, width, height).
	// This defines the widget's outer boundaries and may trigger a layout update.
	SetBounds(int, int, int, int)

	// SetFocused sets the widget's focus state.
	// When a widget gains focus, it typically becomes the target for keyboard input
	// and may update its visual appearance to indicate the focused state.
	//
	// Parameters:
	//   - bool: true to focus the widget, false to unfocus it
	SetFocused(bool)

	// SetHint sets the sizing hint of the widget.
	// The sizing hint is part of the style but is not context-sensitive, when
	// set via the style. With this method, set sizing hint can be set
	// dynamically
	//
	// Parameters:
	//   - int: Preferred content width of the widget
	//   - int: Preferred content height of the widget
	SetHint(int, int)

	// SetHovered sets the hover state of the widget.
	// This method is typically called by the application's mouse event handling
	// system when the mouse cursor enters or leaves the widget's bounds.
	// Widgets may use this state to provide visual feedback to users.
	//
	// Parameters:
	//   - bool: true when the mouse is over the widget, false when it leaves
	SetHovered(bool)

	// SetParent sets the parent widget in the widget hierarchy.
	// This establishes the parent-child relationship that enables event propagation,
	// layout management, and widget tree traversal.
	//
	// Parameters:
	//   - Container: The parent container widget, or nil to remove from current parent
	SetParent(Container)

	// SetPosition sets the widget's position to the specified coordinates.
	// This method updates the widget's location within its parent container
	// or on the screen, affecting where the widget will be rendered.
	//
	// Parameters:
	//   - int: The new x-coordinate for the widget's position
	//   - int: The new y-coordinate for the widget's position
	SetPosition(int, int)

	// SetSize sets the content size of the widget, taking into account margin,
	// padding and border. This method calculates the total widget bounds needed
	// to accommodate the specified content size plus any styling elements.
	//
	// Parameters:
	//   - int: The desired content width
	//   - int: The desired content height
	SetSize(int, int)

	// SetStyle applies the given style to the widget. The style can depend
	// on the state or be for different parts of the widget. The style controls
	// visual appearance such as colors, borders, and text formatting.
	//
	// Parameters:
	//   - string: The style selector (e.g., "", "focus", "hover") for state-specific styling
	//   - *Style: The style configuration to apply, or nil to remove the style
	SetStyle(string, *Style)

	// Size returns the widget's content size as (width, height).
	// This represents the available content area by subtracting margins,
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
	Size() (int, int)

	// State returns the current state of the widget for rendering purposes.
	// Common states include "" (default), "focus", "hover", "disabled", etc.
	// The state is used to determine which style configuration to apply.
	//
	// Returns:
	//   - string: The current widget state identifier
	State() string

	// Style returns the current style configuration applied to the widget for the given selector.
	// This method is used during rendering to determine the visual appearance of the widget.
	//
	// Parameters:
	//   - string: The style selector to retrieve (e.g., "", "focus", "hover")
	//
	// Returns:
	//   - *Style: The style configuration for the selector, or nil if no style is set
	Style(string) *Style

	// Styles returns a list of all defined part/state styles selector names.
	// This is mainly for debugging or introspection purposes.
	//
	// Returns:
	//   - []string: All selector names
	Styles() []string
}
