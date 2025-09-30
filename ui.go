// Package ui.go contains the core UI type and application lifecycle management
// for the zeichenwerk terminal user interface framework.
//
// This file implements the main UI struct which serves as the root container
// and application orchestrator, managing:
//   - Terminal screen initialization and cleanup
//   - Event processing and propagation
//   - Widget hierarchy and layer management
//   - Focus and cursor management
//   - Rendering coordination and performance optimization
//   - Debug logging and visualization
//
// The UI type provides both the foundation for all TUI applications and
// the main entry point for running interactive terminal interfaces.

package zeichenwerk

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// UI represents the main TUI application that manages the screen, event handling,
// and widget hierarchy. It serves as the root container for all UI components
// and coordinates the rendering pipeline, focus management, and user input processing.
//
// The UI acts as the central orchestrator for the entire terminal user interface,
// providing a complete application framework with the following capabilities:
//
// # Core Responsibilities
//
//   - Screen initialization and management using tcell
//   - Event processing (keyboard, mouse, resize events)
//   - Focus navigation between widgets with Tab/Shift+Tab support
//   - Mouse interaction and hover state management
//   - Rendering coordination and dirty state management
//   - Debug logging and visualization
//   - Application lifecycle (startup, main loop, shutdown)
//   - Layer management for popups and modal dialogs
//
// # Event System
//
//   - Hierarchical event propagation from focused widget up the parent chain
//   - Global keyboard shortcuts (Tab navigation, Ctrl+C/Escape to quit)
//   - Mouse hover detection and click handling
//   - Automatic screen resize handling with layout recalculation
//
// # Rendering Pipeline
//
//   - Efficient dirty-flag based rendering to minimize screen updates
//   - Cursor positioning and styling based on focused widget
//   - Debug information overlay in debug mode
//   - Multi-layer rendering support for popups and overlays
//
// # Focus Management
//
//   - Automatic focus traversal through focusable widgets
//   - Focus wrapping (first to last, last to first)
//   - Programmatic focus control and widget finding
//
// # Architecture
//
// The UI maintains a widget hierarchy where containers can hold child widgets,
// enabling complex layouts and nested component structures. It provides both
// imperative APIs for direct control and declarative builder patterns for
// constructing interfaces.
//
// The UI uses a layer-based architecture where each layer represents a level
// of the interface (main UI + popups/modals). This enables proper event handling
// and rendering order for overlay elements.
type UI struct {
	BaseWidget

	// State management
	debug bool // Debug mode flag for showing debug information overlay and logging
	dirty bool // Flag indicating if a screen redraw is needed due to state changes

	// Event handling channels
	events chan tcell.Event // Buffered channel for incoming tcell events (keyboard, mouse, resize)
	quit   chan struct{}    // Channel for signaling graceful application shutdown

	// Rendering channels
	redraw  chan Widget   // Buffered channel for triggering individual widget redraws (performance optimization)
	refresh chan struct{} // Buffered channel for triggering full screen redraws

	// Widget state tracking
	focus Widget // Currently focused widget that receives keyboard input and cursor positioning
	hover Widget // Currently hovered widget for mouse interaction feedback and styling

	// Layer management
	layers []Container // Stack of widget layers (base layer + popups/modals) for proper z-order rendering

	// Debug infrastructure
	logger *Text // Debug log widget for runtime messages with auto-scrolling capability

	// Performance counters
	redraws  int // Counter for individual widget redraws (debugging and performance monitoring)
	refreshs int // Counter for full screen refreshes (debugging and performance monitoring)

	// Rendering system
	renderer Renderer     // Renderer instance responsible for drawing widgets to the terminal
	screen   tcell.Screen // The terminal screen interface for low-level cell manipulation and event polling
}

// ---- Constructor function -------------------------------------------------

// NewUI creates and initializes a new TUI application with the specified theme and root container.
// It sets up the rendering system and prepares the application for the main event loop.
// The actual terminal screen initialization is deferred until Run() is called.
//
// Parameters:
//   - theme: The visual theme to use for styling widgets and colors
//   - root: The root container that will hold all UI widgets
//   - debug: Enable debug mode to show debug information overlay and logging
//
// Returns:
//   - *UI: The initialized UI application instance ready for Run()
//   - error: Any error that occurred during initialization (currently always nil)
//
// # Initialization Process
//
//  1. Creates the UI instance with communication channels and default settings
//  2. Establishes the parent-child relationship with the root container
//  3. Configures the renderer with the provided theme
//  4. Sets up the layer stack with the root container as the base layer
//  5. Connects to debug log widget if present (ID: "debug-log")
//
// # Debug Mode
//
// When debug is true, the UI will:
//   - Display a debug information bar at the bottom of the screen
//   - Show performance counters, focus/hover state, and layer information
//   - Log events and state changes to the debug log widget
//   - Reserve the bottom line of the screen for debug display
//
// # Example Usage
//
//	theme := zeichenwerk.TokyoNightTheme()
//	root := zeichenwerk.NewBuilder(theme).
//		Label("hello", "Hello World", 0).
//		Container()
//	ui, err := zeichenwerk.NewUI(theme, root, false)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := ui.Run(); err != nil {
//		log.Fatal(err)
//	}
func NewUI(theme Theme, root Container, debug bool) (*UI, error) {
	ui := &UI{
		BaseWidget: BaseWidget{id: "root", x: 0, y: 0, width: 0, height: 0},
		screen:     nil,
		renderer:   Renderer{theme: theme},
		layers:     []Container{root},
		debug:      debug,
		dirty:      true, // Initial draw needed
		quit:       make(chan struct{}),
		events:     make(chan tcell.Event, 10),
		redraw:     make(chan Widget, 10), // Initialize redraw channel with buffer
		refresh:    make(chan struct{}, 1),
	}

	root.SetParent(ui)

	// Connect debug log
	logger := ui.Find("debug-log", false)
	if logger != nil {
		if text, ok := logger.(*Text); ok {
			ui.logger = text
			ui.Log(ui, "debug", "==== Debug log started! ====")
			ui.Log(ui, "debug", "Screen size: %d:%d", ui.width, ui.height)
		}
	}

	return ui, nil
}

// ---- Widget methods -------------------------------------------------------
// Implementation of the Widget interface

// Handle processes tcell events and coordinates their handling throughout the
// application. This is the main event processing method that handles keyboard
// input, mouse events, and screen resize events.
//
// # Event Processing Flow
//
// The UI uses a hierarchical event handling system where events are first
// offered to the most specific widget (focused or hovered), then propagated
// up the parent chain until handled or reaching the root UI.
//
// # Event Types
//
// Keyboard Events (*tcell.EventKey):
//  1. Event is first propagated to the focused widget and its parent chain
//  2. If not handled, global application shortcuts are processed
//  3. Unhandled events are discarded
//
// Mouse Events (*tcell.EventMouse):
//   - Determines which widget is at the mouse position
//   - Updates hover state (clearing old, setting new)
//   - Propagates the event to the hovered widget
//   - Triggers screen refresh if hover state changed
//
// Resize Events (*tcell.EventResize):
//   - Updates UI dimensions from the screen
//   - Recalculates layout for all layers
//   - Synchronizes screen buffer and triggers refresh
//
// # Global Keyboard Shortcuts
//
//   - Tab: Navigate to next focusable widget
//   - Shift+Tab: Navigate to previous focusable widget
//   - Escape: Close current popup layer (if multiple layers exist)
//   - Ctrl+C, Ctrl+Q: Quit the application
//   - Ctrl+D: Open widget inspector (debug mode)
//   - 'q', 'Q': Quit the application
//
// # Parameters
//
//   - event: The tcell.Event to process (keyboard, mouse, resize, etc.)
//
// # Returns
//
//   - bool: Always returns true as the UI is the root event handler
func (ui *UI) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		// First try to handle the event with the focused widget
		// If the event is handled by the focused widget, we do not process the
		// event any further.
		if ui.propagate(ui.focus, event) {
			break
		}

		// Handle global app events, if the keyboard event was propagated
		ui.Log(ui, "debug", "Handling key event %v", event)
		switch event.Key() {
		case tcell.KeyTab:
			ui.SetFocus("next")
		case tcell.KeyBacktab:
			ui.SetFocus("previous")
		case tcell.KeyEscape:
			if len(ui.layers) > 1 {
				ui.Close()
			}
		case tcell.KeyCtrlC, tcell.KeyCtrlQ:
			close(ui.quit)
		case tcell.KeyCtrlD:
			ui.Log(ui, "debug", "Opening inspector")
			ui.Popup(-1, -1, 0, 0, NewInspector(ui).UI())
			ui.Refresh()
		case tcell.KeyRune:
			switch event.Rune() {
			case 'q', 'Q':
				close(ui.quit)
			}
		}

	case *tcell.EventMouse:
		at := ui.FindAt(event.Position())
		if at != ui.hover {
			if ui.hover != nil {
				ui.hover.SetHovered(false)
			}
			if at != nil {
				ui.hover = at
				at.SetHovered(true)
			}
			ui.Refresh()
		} else {
			switch event.Buttons() {
			case tcell.Button1:
				if at.Focusable() && at != ui.focus {
					ui.Focus(at)
				}
			}
		}
		ui.propagate(ui.hover, event)

	case *tcell.EventPaste:
		ui.propagate(ui.focus, event)

	case *tcell.EventResize:
		ui.width, ui.height = ui.screen.Size()
		ui.Layout()
		ui.screen.Sync()
		ui.Refresh() // Redraw after resize
	}

	return true
}

// propagate sends an event up the widget hierarchy starting from the target widget.
// This implements the event bubbling pattern where events are first handled by
// the most specific widget (target), then by its parent, and so on up the chain
// until either a widget handles the event or the root UI is reached.
//
// Parameters:
//   - target: The widget to start event propagation from (typically focused or hovered widget)
//   - event: The tcell.Event to propagate (keyboard, mouse, etc.)
//
// Returns:
//   - bool: true if any widget in the chain handled the event, false otherwise
//
// Propagation process:
//  1. Start with the target widget
//  2. Call the widget's Handle method with the event
//  3. If handled, stop propagation and return true
//  4. If not handled, move to the widget's parent
//  5. Repeat until handled or root UI is reached
//
// This allows for hierarchical event handling where child widgets can handle
// specific events while parent containers can provide fallback behavior.
func (ui *UI) propagate(target Widget, event tcell.Event) bool {
	current := target
	handled := false
	for current != nil && !handled && current != ui {
		handled = current.Handle(event)
		current = current.Parent()
	}
	return handled
}

// Redraw queues the specified widget for individual redraw optimization.
// This method provides a performance optimization by redrawing only the
// changed widget instead of the entire screen.
//
// # Behavior
//
// The method attempts to enqueue the widget in the redraw channel without blocking.
// If the redraw channel is full (indicating heavy redraw activity), it falls back
// to triggering a complete screen refresh instead.
//
// # Usage Guidelines
//
// Use Redraw when:
//   - Only a single widget's content or state has changed
//   - The widget's position and size remain unchanged
//   - Other widgets on screen are not affected by the change
//   - Performance optimization is desired for frequent updates
//
// Use Refresh instead when:
//   - Multiple widgets have changed
//   - Layout has been modified
//   - Widget positions or sizes have changed
//   - Focus or hover states have changed
//
// # Performance Benefits
//
// Individual widget redraws are significantly faster than full screen refreshes
// because they:
//   - Skip layout calculations
//   - Only render the specific widget
//   - Avoid traversing the entire widget hierarchy
//   - Minimize terminal output operations
func (ui *UI) Redraw(widget Widget) {
	select {
	case ui.redraw <- widget:
	default:
		ui.Log(ui, "debug", "Redraw queue full")
		ui.Refresh()
	}
}

// Refresh triggers a complete screen redraw for all visible widgets.
// This method sets the dirty flag and signals the main event loop to perform
// a full rendering pass on the next iteration.
//
// # Behavior
//
// The method performs two operations:
//  1. Sets the dirty flag to true, marking the screen as needing an update
//  2. Attempts to send a signal through the refresh channel without blocking
//
// If the refresh channel is full (indicating a refresh is already pending),
// the signal is skipped to avoid blocking and prevent redundant refreshes.
//
// # When to Use
//
// Call Refresh when:
//   - Multiple widgets have changed and need updating
//   - Layout modifications have occurred (position, size changes)
//   - Focus or hover states have changed
//   - Theme changes have been applied
//   - Layer stack modifications (popups opened/closed)
//   - Initial rendering or major state changes
//
// # Performance Considerations
//
// Full screen refreshes are more expensive than individual widget redraws
// because they:
//   - Traverse the entire widget hierarchy
//   - Recalculate cursor positioning
//   - Render debug information (if enabled)
//   - Update all visible layers
//
// For single widget updates, prefer using Redraw() for better performance.
func (ui *UI) Refresh() {
	ui.dirty = true
	select {
	case ui.refresh <- struct{}{}:
	default: // Channel is full, redraw already pending
	}
}

// ---- Container Methods ----------------------------------------------------

// Children returns the child widgets of the App.
// Since UI acts as the root container, it returns a slice containing
// only the root container widget.
func (ui *UI) Children(_ bool) []Widget {
	result := make([]Widget, 0, len(ui.layers))
	for _, layer := range ui.layers {
		result = append(result, layer)
	}
	return result
}

// Find searches for a widget with the specified ID in the widget hierarchy.
// It first checks if the root container matches the ID, then delegates
// the search to the root container's Find method to search recursively
// through all child widgets.
//
// Parameters:
//   - id: The unique identifier of the widget to find
//   - visible: Only look for visible children
//
// Returns:
//   - Widget: The widget with the matching ID, or nil if not found
func (ui *UI) Find(id string, visible bool) Widget {
	var widget Widget
	for _, layer := range ui.layers {
		if layer.ID() == id {
			return layer
		} else {
			widget = layer.Find(id, visible)
			if widget != nil {
				return widget
			}
		}
	}
	return nil
}

// FindAt searches for the widget at the specified screen coordinates.
// This method is used for mouse interaction to determine which widget
// is located at a given position on the screen.
//
// Parameters:
//   - x: The x-coordinate on the screen
//   - y: The y-coordinate on the screen
//
// Returns:
//   - Widget: The widget at the specified position, or nil if no widget is found
func (ui *UI) FindAt(x, y int) Widget {
	var widget Widget
	for current := len(ui.layers) - 1; current >= 0; current-- {
		widget = ui.layers[current].FindAt(x, y)
		if widget != nil {
			return widget
		}
	}
	return widget
}

// Layout recalculates and applies the layout for all widget layers in the UI.
// This method is called automatically when the screen is resized or when
// the UI structure changes. It ensures that all widgets are properly
// positioned and sized according to their layout constraints.
//
// Layout process:
//  1. Calculate available screen space (reserving bottom line for debug if enabled)
//  2. Set bounds for the base layer (root container) to fill available space
//  3. Trigger layout calculation for the root container and all its children
//  4. Additional layers (popups) maintain their existing bounds and layouts
//
// Debug mode considerations:
//   - In debug mode, the bottom line is reserved for debug information display
//   - This reduces the available height for the main UI by one line
//   - Popup layers are not affected by debug mode space reservation
//
// This method should be called whenever:
//   - The terminal window is resized
//   - The UI structure changes (widgets added/removed)
//   - Debug mode is toggled
//   - Manual layout refresh is needed
func (ui *UI) Layout() {
	// Set the bounds of the root widget to the screen bounds.
	// In debug mode, the bottom line is reserved for debug information
	if ui.debug {
		ui.layers[0].SetBounds(0, 0, ui.width, ui.height-1)
	} else {
		ui.layers[0].SetBounds(0, 0, ui.width, ui.height)
	}

	// Lay out the root widget.
	ui.layers[0].Layout()
}

// ---- Drawing Methods -------------------------------------------------------

// Draw renders the entire application to the screen.
// This method handles cursor positioning, debug information display,
// and coordinates the rendering of all widgets through the renderer.
//
// The draw process includes:
//   - Positioning the cursor based on the focused widget
//   - Rendering debug information if debug mode is enabled
//   - Triggering the main widget rendering pipeline
//   - Displaying debug logs in debug mode
//   - Showing the final rendered frame to the user
//
// The method only performs actual rendering if the dirty flag is set,
// providing efficient updates by avoiding unnecessary redraws.
func (ui *UI) Draw() {
	if !ui.dirty {
		return
	}

	ui.refreshs++
	for _, layer := range ui.layers {
		ui.renderer.render(layer)
	}
	ui.ShowCursor()
	ui.ShowDebug()
	ui.screen.Show()
	ui.dirty = false
}

// Redraw renders just a single widget, if its state changed. No new layout is
// performed, no other widgets are affected.
//
// Parameters:
//   - widget: Widget to redraw
func (ui *UI) DrawWidget(widget Widget) {
	ui.redraws++
	ui.renderer.render(widget)
	ui.ShowCursor()
	ui.ShowDebug()
	ui.screen.Show()
}

// ShowDebug renders the debug information bar at the bottom of the screen.
// This method displays real-time debugging information when debug mode is enabled,
// providing insights into the application's current state and performance metrics.
//
// Debug information displayed:
//   - Frame counter: Number of frames rendered (performance indicator)
//   - Screen dimensions: Current terminal width and height
//   - Layer count: Number of active UI layers (main + popups)
//   - Focused widget: ID of the currently focused widget (or "<nil>")
//   - Hovered widget: ID of the currently hovered widget (or "<nil>")
//
// Visual formatting:
//   - Uses green background with black text for high visibility
//   - Positioned at the bottom line of the screen (height-1)
//   - Spans the full width of the terminal
//   - Uses Unicode box-drawing characters (│) as separators
//
// The debug bar is only rendered when debug mode is enabled and provides
// valuable information for:
//   - Performance monitoring (frame counter)
//   - Layout debugging (screen dimensions, layer count)
//   - Interaction debugging (focus and hover state)
//   - Widget hierarchy troubleshooting
//
// This method is called automatically during the rendering process.
func (ui *UI) ShowDebug() {
	if ui.debug {
		focus := "<nil>"
		hover := "<nil>"
		if ui.focus != nil {
			focus = ui.focus.ID()
		}
		if ui.hover != nil {
			hover = ui.hover.ID()
		}
		ui.renderer.SetStyle(NewStyle("black", "green"))
		ui.renderer.text(
			0,
			ui.height-1,
			fmt.Sprintf(" DEBUG \u2502 Refresh %6d \u2502 Redraw %6d \u2502 Screen %3d:%3d \u2502 Layers %2d \u2502 Focus %-20s \u2502 Hover %-20s", ui.refreshs, ui.redraws, ui.width, ui.height, len(ui.layers), focus, hover),
			ui.width)
	}
}

// ---- Focus and Cursor Handling --------------------------------------------

// Focus sets the keyboard focus to the specified widget.
// This method handles the focus transition by properly updating the focus
// state of both the previously focused widget and the new target widget.
//
// Parameters:
//   - widget: The widget to receive focus, or nil to clear focus
//
// Focus transition process:
//  1. If a widget is currently focused and different from the target, clear its focus state
//  2. If the target widget is not nil, set its focus state to true
//  3. Update the UI's internal focus reference
//  4. Trigger a screen refresh to update visual focus indicators
//
// Visual effects:
//   - Previously focused widget loses focus styling (borders, highlights)
//   - Newly focused widget gains focus styling according to its theme
//   - Cursor positioning is updated based on the new focused widget
//
// The method is safe to call with nil to clear focus entirely, and handles
// the case where the same widget is already focused without unnecessary updates.
func (ui *UI) Focus(widget Widget) {
	if ui.focus != nil && ui.focus != widget {
		ui.focus.SetFocused(false)
	}
	if widget != nil {
		widget.SetFocused(true)
	}
	ui.focus = widget
	ui.Refresh()
}

// SetFocus navigates focus between widgets using directional or positional commands.
// This method implements keyboard navigation patterns commonly used in terminal
// applications, providing consistent focus traversal behavior.
//
// Parameters:
//   - which: Direction or position command for focus navigation
//
// Supported commands:
//   - "first": Focus the first focusable widget in the current layer
//   - "last": Focus the last focusable widget in the current layer
//   - "next": Focus the next focusable widget (wraps to first if at end)
//   - "previous": Focus the previous focusable widget (wraps to last if at beginning)
//   - Any other value: Defaults to "first"
//
// Navigation behavior:
//   - Only considers widgets that return true for Focusable()
//   - Operates on the topmost layer (current popup or main UI)
//   - Implements wrapping: next from last goes to first, previous from first goes to last
//   - Traverses widgets in document order (depth-first tree traversal)
//
// Use cases:
//   - Tab key navigation (next)
//   - Shift+Tab navigation (previous)
//   - Home key (first)
//   - End key (last)
//   - Initial focus setup when opening dialogs
func (ui *UI) SetFocus(which string) {
	var first, previous, next, last Widget
	found := false
	Traverse(ui.layers[len(ui.layers)-1], true, func(widget Widget) {
		if !widget.Focusable() {
			return
		}
		if first == nil {
			first = widget
		}
		if widget == ui.focus {
			found = true
		} else if !found {
			previous = widget
		} else if next == nil && found {
			next = widget
		}
		last = widget
	})
	switch which {
	case "last":
		ui.Focus(last)
	case "next":
		if next == nil {
			ui.Focus(first)
		} else {
			ui.Focus(next)
		}
	case "previous":
		if previous == nil {
			ui.Focus(last)
		} else {
			ui.Focus(previous)
		}
	default:
		ui.Focus(first)
	}
}

// ShowCursor positions and displays the cursor based on the currently focused widget.
// The cursor appearance and position are determined by the focused widget's
// cursor style configuration and current cursor position.
//
// Cursor positioning:
//   - Uses the focused widget's content area as the base coordinate system
//   - Adds the widget's internal cursor offset to determine final screen position
//   - Automatically hides cursor if no widget is focused or cursor is disabled
//
// Supported cursor styles:
//   - "|", "bar", "steady-bar": Steady vertical bar cursor
//   - "*|", "*bar", "blinking-bar": Blinking vertical bar cursor
//   - "#", "block", "steady-block": Steady block cursor
//   - "*#", "blinking-block", "*block": Blinking block cursor
//   - "_", "underline", "steady-underline": Steady underline cursor
//   - "*_", "blinking-underline", "*underline": Blinking underline cursor
//
// Cursor visibility rules:
//   - Cursor is shown only if a widget is focused
//   - Widget must have a non-empty cursor style configured
//   - Widget must provide valid cursor coordinates (>= 0)
//   - Cursor is hidden if any of the above conditions are not met
//
// This method is called automatically during the rendering process
// and should not typically be called directly by application code.
func (ui *UI) ShowCursor() {
	// Show cursor
	if ui.focus != nil {
		x, y, _, _ := ui.focus.Content()
		cx, cy := ui.focus.Cursor()
		cursor := ui.focus.Style("").Cursor
		if cursor != "" && cx >= 0 && cy >= 0 {
			cs := tcell.CursorStyleDefault
			switch cursor {
			case "|", "bar", "steady-bar":
				cs = tcell.CursorStyleSteadyBar
			case "*|", "*bar", "blinking-bar":
				cs = tcell.CursorStyleBlinkingBar
			case "*#", "blinking-block", "*block":
				cs = tcell.CursorStyleBlinkingBlock
			case "*_", "blinking-underline", "*underline":
				cs = tcell.CursorStyleBlinkingUnderline
			case "#", "block", "steady-block":
				cs = tcell.CursorStyleSteadyBlock
			case "_", "underline", "steady-underline":
				cs = tcell.CursorStyleSteadyUnderline
			}
			ui.screen.SetCursorStyle(cs)
			ui.screen.ShowCursor(x+cx, y+cy)
			ui.screen.Show()
		} else {
			ui.screen.HideCursor()
		}
	}
}

// Log adds a debug message to the application's log buffer.
// The log maintains only the most recent 100 messages to prevent
// unlimited memory growth. Messages are displayed in debug mode
// and can be useful for troubleshooting widget behavior and events.
//
// Parameters:
//   - source: Source widget
//   - level: Log level
//   - message: The debug message to add to the log
func (ui *UI) Log(source Widget, level, message string, params ...any) {
	if ui.logger != nil {
		ui.logger.Add(time.Now().Format("15:04:05.000") + fmt.Sprintf(" %-6s %-16s %-10s ", level, source.ID(), WidgetType(source)) + fmt.Sprintf(message, params...))
	}
}

// ---- Popup and Layer Handling ---------------------------------------------

// Popup displays a container widget as an overlay on top of the current UI.
// This method adds the popup as a new layer in the layer stack, allowing for
// modal dialogs, context menus, and other overlay interfaces.
//
// Parameters:
//   - x: Horizontal position (-1 for center, negative for right-aligned offset)
//   - y: Vertical position (-1 for center, negative for bottom-aligned offset)
//   - w: Width of the popup in characters
//   - h: Height of the popup in characters
//   - popup: The container widget to display as a popup
//
// Positioning behavior:
//   - x = -1: Center horizontally on screen
//   - x < -1: Position relative to right edge (e.g., -3 = 1 chars from right)
//   - x >= 0: Absolute position from left edge
//   - y = -1: Center vertically on screen
//   - y < -1: Position relative to bottom edge (e.g., -3 = 1 chars from bottom)
//   - y >= 0: Absolute position from top edge
//
// Layer management:
//   - Adds popup as new topmost layer
//   - Sets UI as parent of the popup container
//   - Applies specified bounds and triggers layout
//   - Automatically focuses first focusable widget in popup
//   - Triggers screen refresh to display the popup
//
// Example usage:
//
//	// Centered dialog
//	ui.Popup(-1, -1, 40, 10, dialog)
//
//	// Bottom-right positioned popup
//	ui.Popup(-2, -3, 30, 8, contextMenu)
func (ui *UI) Popup(x, y, w, h int, popup Container) {
	// Auto sizing
	if w == 0 || h == 0 {
		pw, ph := popup.Hint()
		if w == 0 {
			w = pw
		}
		if h == 0 {
			h = ph
		}
	}

	// if x is -1, center the popup horizontally
	if x == -1 {
		x = (ui.width - w) / 2
	} else if x < 0 {
		x = ui.width - w + x + 2
	}

	// if y is -1, center the popup vertically
	if y == -1 {
		y = (ui.height - h) / 2
	} else if y < 0 {
		y = ui.height - h + y + 2
	}

	popup.SetParent(ui)
	popup.SetBounds(x, y, w, h)
	popup.Layout()

	ui.layers = append(ui.layers, popup)
	ui.SetFocus("first")
	ui.Refresh()
}

// Close removes the topmost layer from the UI layer stack.
// This method is typically used to close popup dialogs, modal windows,
// or other overlay widgets that were added as additional layers.
// The base layer (main UI) cannot be closed using this method.
//
// Behavior:
//   - Only closes layers if more than one layer exists (protects base layer)
//   - Removes the topmost layer from the stack
//   - Automatically sets focus to the first focusable widget in the remaining top layer
//   - Logs the close operation for debugging purposes
//   - Triggers a screen refresh to update the display
//
// Use cases:
//   - Closing popup dialogs (OK/Cancel buttons)
//   - Dismissing modal windows (Escape key handler)
//   - Removing temporary overlays or context menus
//   - Implementing navigation back functionality
//
// The method is safe to call even when only the base layer exists,
// as it includes protection against closing the main UI layer.
func (ui *UI) Close() {
	if len(ui.layers) > 1 {
		ui.layers = ui.layers[:len(ui.layers)-1]
		ui.SetFocus("first")
	}
	ui.Refresh()
}

// FindOn locates a widget by ID and attaches an event handler to it.
// This is a convenience method that combines widget lookup with event handler
// registration, providing a simple way to set up event handling for specific widgets.
//
// Parameters:
//   - id: The unique identifier of the widget to find
//   - event: The event name to listen for (e.g., "click", "change", "select")
//   - handler: The event handler function to attach
//
// Handler function signature:
//   - func(Widget, string, ...any) bool
//   - First parameter: The widget that triggered the event
//   - Second parameter: The event name
//   - Remaining parameters: Event-specific data
//   - Return value: true if event was handled, false to continue propagation
//
// Behavior:
//   - Searches the entire widget hierarchy for the specified ID
//   - If widget is found, attaches the handler for the specified event
//   - If widget is not found, the operation is silently ignored
//   - Multiple handlers can be attached to the same widget/event combination
//
// Common event types:
//   - "click": Button clicks, list item activation
//   - "change": Input field text changes, selection changes
//   - "select": List item selection, focus changes
//   - "key": Raw keyboard events not handled by widget
//
// Example usage:
//
//	ui.FindOn("ok-button", "click", func(w Widget, event string, data ...any) bool {
//		fmt.Println("OK button clicked")
//		return true
//	})
func (ui *UI) FindOn(id, event string, handler func(Widget, string, ...any) bool) {
	widget := ui.Find(id, false)
	if widget != nil {
		widget.On(event, handler)
	}
}

// ---- Builder method -------------------------------------------------------

// Builder returns a builder to construct UIs or parts like popups using the
// currently set theme.
func (ui *UI) Builder() *Builder {
	return NewBuilder(ui.Theme())
}

// ---- Dialog methds --------------------------------------------------------

// Confirm displays a popup dialog with a confirmation message.
func (ui *UI) Confirm(title, message, ok, cancel string, fn func()) {
	dialog := ui.Builder().Dialog("std-confirm", title).Class("dialog").
		Flex("std-confirm-flex", "vertical", "stretch", 1).Padding(1, 2).
		Label("std-confirm-label", message, 0).
		Flex("std-confirm-buttons", "horizontal", "start", 2).
		Spacer().
		Button("std-confirm-ok", ok).
		Button("std-confirm-cancel", cancel).
		End().
		End().
		Container().(*Dialog)

	With(dialog, "std-confirm-ok", func(btn *Button) {
		btn.On("click", func(_ Widget, _ string, _ ...any) bool {
			ui.Close()
			fn()
			return true
		})
	})

	With(dialog, "std-confirm-cancel", func(btn *Button) {
		btn.On("click", func(_ Widget, _ string, _ ...any) bool {
			ui.Close()
			return true
		})
	})

	dialog.Layout()
	w, h := dialog.Hint()
	ui.Log(dialog, "debug", "Dialog Hint %d.%d", w, h)
	ui.Popup(-1, -1, 0, 0, dialog)
}

// Info shows an information message.
func (ui *UI) Message(title, message, ok string) {
	dialog := ui.Builder().Dialog("std-info", title).Class("dialog").
		Flex("std-info-flex", "vertical", "stretch", 1).Padding(1, 2).
		Label("std-info-label", message, 0).
		Flex("std-info-buttons", "horizontal", "start", 2).
		Spacer().
		Button("std-info.ok", ok).
		End().
		End().
		Container().(*Dialog)

	With(dialog, "std-info-ok", func(btn *Button) {
		btn.On("click", func(_ Widget, _ string, _ ...any) bool {
			ui.Close()
			return true
		})
	})
}

// ---- Run Loop -------------------------------------------------------------

// Run starts the main application event loop and blocks until the application
// exits. This is the primary method that drives the TUI application, handling
// all events, rendering, and application lifecycle management.
//
// # Initialization Phase
//
// Before entering the event loop, Run performs the following initialization:
//
//  1. Creates and initializes a new tcell screen for terminal interaction
//  2. Enables mouse event support for hover detection and click interactions
//  3. Sets up the default screen style (black background, white foreground)
//  4. Configures the renderer with the screen interface
//  5. Determines initial screen dimensions
//  6. Sets initial focus to the first focusable widget
//  7. Performs initial layout calculation and screen rendering
//  8. Starts the background event polling goroutine
//
// # Event Loop
//
// The main event loop handles the following in order of priority:
//  1. Quit signals - highest priority for graceful shutdown
//  2. Individual widget redraw requests - performance optimization
//  3. Full screen refresh requests - complete re-rendering
//  4. Input events - keyboard, mouse, and resize events
//
// # Shutdown Process
//
// When a quit signal is received, the method:
//  1. Calls screen.Fini() to restore terminal state
//  2. Returns nil to indicate successful shutdown
//
// # Error Conditions
//
// The method returns an error if:
//   - Screen creation fails (terminal not available, permissions, etc.)
//   - Screen initialization fails (terminal capabilities, size detection)
//
// # Concurrency
//
// Run spawns a background goroutine (EventLoop) to poll for tcell events
// and forward them to the main event loop channel. This prevents the main
// loop from blocking on event polling.
func (ui *UI) Run() error {
	var err error

	// Initialize screen
	ui.screen, err = tcell.NewScreen()
	if err != nil {
		return err
	}

	if err := ui.screen.Init(); err != nil {
		return err
	}

	// Enable mouse events
	ui.screen.EnableMouse()

	style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	ui.screen.SetStyle(style)
	ui.screen.Clear()
	ui.renderer.screen = ui.screen
	ui.renderer.style = style

	// Take screen size for the root
	ui.width, ui.height = ui.screen.Size()

	// Set initial focus
	ui.SetFocus("first")

	// Initial draw
	ui.Layout()
	ui.Draw()

	go ui.EventLoop()

	// Event handling loop
	for {
		select {
		case <-ui.quit:
			ui.screen.Fini()
			return nil
		case widget := <-ui.redraw:
			ui.DrawWidget(widget)
		case <-ui.refresh:
			ui.Draw()
		case event := <-ui.events:
			ui.Handle(event)
		}
	}
}

// EventLoop continuously polls for tcell events and forwards them to the main event loop.
// This method runs in a separate goroutine to prevent blocking the main event loop
// during event polling operations.
//
// # Operation
//
// The method runs an infinite loop that:
//  1. Calls screen.PollEvent() to wait for the next terminal event
//  2. Forwards non-nil events to the main event loop via the events channel
//  3. Continues polling until the application terminates
//
// # Concurrency Design
//
// This separation of event polling from event processing allows the main event loop
// to handle multiple types of operations (rendering, widget updates, shutdown) without
// being blocked by the synchronous nature of tcell's event polling.
//
// The goroutine automatically terminates when the screen is finalized during
// application shutdown, as PollEvent() will return nil or panic.
func (ui *UI) EventLoop() {
	for {
		ev := ui.screen.PollEvent()
		if ev != nil {
			ui.events <- ev
		}
	}
}

// ---- Theme Management -----------------------------------------------------

// SetTheme changes the active theme and re-applies styling to all widgets.
// This method enables runtime theme switching by updating the renderer's theme
// and triggering a complete re-styling of the entire widget hierarchy.
//
// Parameters:
//   - theme: The new theme to apply to the application
//
// Process:
//  1. Updates the renderer's theme reference
//  2. Recursively re-applies theme styles to all widgets in all layers
//  3. Triggers a complete screen refresh to show the new styling
//
// The method traverses the entire widget hierarchy and re-applies the appropriate
// theme styles based on each widget's type, class, and ID. This ensures that
// all visual elements immediately reflect the new theme.
//
// Example usage:
//
//	// Switch to Nord theme
//	ui.SetTheme(NordTheme())
//
//	// Switch to Tokyo Night theme
//	ui.SetTheme(TokyoNightTheme())
//
// SetTheme applies a new theme to the UI and all its widgets.
// This method updates the renderer's theme and re-applies theme styles
// to all widgets in the widget hierarchy, ensuring consistent visual
// appearance throughout the application.
//
// The method performs the following operations:
//   - Updates the renderer's theme reference
//   - Creates a new builder with the new theme
//   - Traverses all widgets and applies the new theme styles
//   - Triggers a UI refresh to update the display
//
// Parameters:
//   - theme: The new theme to apply to the UI and all widgets
//
// Example usage:
//
//	// Switch to dark theme
//	ui.SetTheme(NewDarkTheme())
//
//	// Apply a custom theme
//	customTheme := NewCustomTheme()
//	ui.SetTheme(customTheme)
//
// Note: This operation affects all widgets in the UI hierarchy and
// triggers a complete visual refresh of the interface.
func (ui *UI) SetTheme(theme Theme) {
	ui.renderer.theme = theme

	// Re-apply theme styles to all widgets
	builder := NewBuilder(theme)
	Traverse(ui, false, func(widget Widget) {
		builder.Apply(widget)
	})

	ui.Refresh()
}

// Theme returns the currently active theme.
// This method provides access to the current theme for inspection or
// for storing the current theme before switching to a different one.
//
// Returns:
//   - Theme: The currently active theme
func (ui *UI) Theme() Theme {
	return ui.renderer.theme
}

// ---- Helper functions -----------------------------------------------------

// print recursively logs information about a container and its child widgets.
// This is a debug utility function that traverses the widget hierarchy and
// outputs detailed information about each widget including its type, ID, and state.
// The level parameter controls indentation for nested containers.
func print(level int, container Container) {
	for i, widget := range container.Children(false) {
		container.Log(widget, "debug", "%s %3d %T %s %s\n", strings.Repeat(" ", level), i, widget, widget.ID(), widget.Info())
		c, ok := widget.(Container)
		if ok {
			print(level+1, c)
		}
	}
}
