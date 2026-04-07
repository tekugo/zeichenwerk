package zeichenwerk

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gdamore/tcell/v3"
)

// UI represents the main TUI application that manages the screen, event
// handling, and widget hierarchy. It serves as the root container for all UI
// components and coordinates the rendering pipeline, focus management, and
// user input processing.
//
// The UI acts as the central orchestrator for the entire terminal user interface,
// providing a complete application framework with the following capabilities.
//
// # Core Responsibilities
//
//   - Screen initialization and management using tcell
//   - Event processing (keyboard, mouse, resize events)
//   - Focus navigation between widgets with Tab/Shift+Tab support
//   - Mouse interaction and hover state management
//   - Rendering coordination and dirty state management
//   - Debug logging
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
	Component

	// State management
	debug bool // Debug mode flag for showing debug information overlay and logging
	dirty bool // Flag indicating if a screen redraw is needed due to state changes

	// Event handling channels
	events   chan tcell.Event // Buffered channel for incoming tcell events (keyboard, mouse, resize)
	quit     chan struct{}    // Channel for signaling graceful application shutdown
	quitOnce sync.Once        // Guards the quit channel so Quit() is safe to call multiple times

	// Rendering channels
	redraw  chan Widget   // Buffered channel for triggering individual widget redraws (performance optimization)
	refresh chan struct{} // Buffered channel for triggering full screen redraws

	// Widget state tracking
	focus      Widget   // Currently focused widget that receives keyboard input and cursor positioning
	focusStack []Widget // Focus saved before each popup layer was opened; restored on Close
	hover      Widget   // Currently hovered widget for mouse interaction feedback and styling

	// Layer management
	layers []Container // Stack of widget layers (base layer + popups/modals) for proper z-order rendering

	// Logging infrastructure
	tableLog   *TableLog
	logger     *slog.Logger
	logHandler *UILogHandler

	// Performance counters
	redraws  int // Counter for individual widget redraws (debugging and performance monitoring)
	refreshs int // Counter for full screen refreshes (debugging and performance monitoring)

	// Rendering system
	renderer *Renderer    // Renderer instance responsible for drawing widgets to the terminal
	screen   tcell.Screen // The terminal screen interface for low-level cell manipulation and event polling
}

// parseLevel converts a Level to slog.Level.
func parseLevel(l Level) slog.Level {
	switch l {
	case Debug:
		return slog.LevelDebug
	case Info:
		return slog.LevelInfo
	case Warning:
		return slog.LevelWarn
	case Error, Fatal:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
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
func NewUI(theme *Theme, root Container) *UI {
	ui := &UI{
		Component: Component{id: "__ui__", x: 0, y: 0, width: 0, height: 0},
		screen:    nil,
		renderer:  &Renderer{theme: theme},
		layers:    []Container{},
		debug:     false,
		tableLog:  NewTableLog(2000),
		dirty:     true, // Initial draw needed
		quit:      make(chan struct{}),
		events:    make(chan tcell.Event, 10),
		redraw:    make(chan Widget, 10), // Initialize redraw channel with buffer
		refresh:   make(chan struct{}, 1),
	}

	if root != nil {
		ui.Add(root)
	}

	return ui
}

// Debug sets the UI into debug mode
func (ui *UI) Debug() *UI {
	ui.debug = true

	level := slog.LevelDebug
	ui.logHandler = &UILogHandler{
		tableLog: ui.tableLog,
		level:    level,
		console:  false,
	}
	ui.logger = slog.New(ui.logHandler)
	ui.logger.Debug("==== Debug log started! =====", "source", "ui", "widgetType", "UI")

	// Connect debug log widget and initialize structured logging
	loggerWidget := Find(ui, "debug-log")
	if loggerWidget != nil {
		if text, ok := loggerWidget.(*Text); ok {
			ui.logHandler.text = text
		}
	}

	return ui
}

// Creates a new builder with the current UI theme.
func (ui *UI) NewBuilder() *Builder {
	return NewBuilder(ui.renderer.theme)
}

// ---- Event Handling -------------------------------------------------------

// Handle processes tcell events and coordinates their handling throughout the
// application. This is the main event processing method that handles keyboard
// input, mouse events, and screen resize events.
// # Parameters
//
//   - event: The tcell.Event to process (keyboard, mouse, resize, etc.)
//
// # Returns
//   - bool: Always returns true as the UI is the root event handler
func (ui *UI) Handle(event tcell.Event) bool {
	switch event := event.(type) {
	case *tcell.EventKey:
		// First try to handle the event with the focused widget.
		// If the event is handled by the focused widget or one of its parents
		// we do not process the event any further.
		if ui.dispatch(ui.focus, EvtKey, event) {
			break
		}

		// Handle global app events, if the keyboard event was propagated
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyRight, tcell.KeyDown:
			ui.SetFocus("next")
		case tcell.KeyBacktab, tcell.KeyLeft, tcell.KeyUp:
			ui.SetFocus("previous")
		case tcell.KeyEscape:
			if len(ui.layers) > 1 {
				ui.Close()
			}
		case tcell.KeyCtrlC, tcell.KeyCtrlQ:
			ui.Quit()
		case tcell.KeyCtrlD:
			ui.Log(ui, Debug, "Opening inspector")
			animation := NewGrow("inspector-grow", "", false)
			animation.Add(NewInspector(ui).UI())
			hw, hh := animation.Hint()
			animation.Start(10 * time.Millisecond)
			ui.Log(ui, Debug, "Inspection hint", "hw", hw, "hh", hh)
			ui.Popup(-1, -1, 0, 0, animation)
			ui.Refresh()
		case tcell.KeyRune:
			switch event.Str() {
			case "q", "Q":
				ui.Quit()
			}
		}

	case *tcell.EventMouse:
		// We only search the highest layer for hovering
		mx, my := event.Position()
		at := FindAt(ui.layers[len(ui.layers)-1], mx, my)
		if at != ui.hover {
			if ui.hover != nil {
				ui.hover.SetFlag(FlagHovered, false)
				ui.Redraw(ui.hover)
			}
			if at != nil {
				ui.hover = at
				at.SetFlag(FlagHovered, true)
				ui.Redraw(at)
			}
		} else {
			switch event.Buttons() {
			case tcell.Button1:
				if at.Flag(FlagFocusable) && at != ui.focus {
					ui.Focus(at)
				}
			}
		}
		ui.dispatch(ui.hover, EvtHover, event)
		ui.dispatch(at, EvtMouse, event)

	case *tcell.EventPaste:
		ui.dispatch(ui.focus, EvtPaste, event)

	case *tcell.EventResize:
		w, h := ui.screen.Size()
		if ui.width != w || ui.height != h {
			ui.width, ui.height = w, h
			ui.Layout()
			ui.Log(ui, Debug, "Screen size: %d:%d", ui.width, ui.height)
			ui.Refresh()
			ui.screen.Sync()
		}
	}

	return true
}

// dispatch sends an event up the widget hierarchy starting from the target widget.
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
func (ui *UI) dispatch(target Widget, event Event, data ...any) bool {
	current := target
	handled := false
	for current != nil && !handled && current != ui {
		handled = current.Dispatch(current, event, data...)
		current = current.Parent()
	}
	return handled
}

// ---- Container Methods ----------------------------------------------------

// Add adds a new container layer to the UI.
// This should only be done for the base layer, other layers should be added
// via Popup().
func (ui *UI) Add(widget Widget, _ ...any) error {
	if container, ok := widget.(Container); ok {
		container.SetParent(ui)
		ui.layers = append(ui.layers, container)
		return nil
	} else {
		return ErrNoContainer
	}
}

// Children returns the child widgets of the App.
// Since UI acts as the root container, it returns a slice containing
// only the root container widget.
func (ui *UI) Children() []Widget {
	result := make([]Widget, 0, len(ui.layers))
	for _, layer := range ui.layers {
		result = append(result, layer)
	}
	return result
}

// Layout recalculates and applies the layout for all widget layers in the UI.
// This method is called automatically when the screen is resized or when
// the UI structure changes. It ensures that all widgets are properly
// positioned and sized according to their layout constraints.
func (ui *UI) Layout() error {
	// Set the bounds of the root widget to the screen bounds.
	// In debug mode, the bottom line is reserved for debug information
	if ui.debug {
		ui.layers[0].SetBounds(0, 0, ui.width, ui.height-1)
	} else {
		ui.layers[0].SetBounds(0, 0, ui.width, ui.height)
	}

	// Lay out the root widget.
	return ui.layers[0].Layout()
}

// ---- Drawing Methods -------------------------------------------------------

// Draw renders the entire application to the screen.
// This method handles cursor positioning, debug information display, and
// coordinates the rendering of all widgets through the renderer.
func (ui *UI) Draw() {
	if !ui.dirty {
		return
	}

	ui.refreshs++
	for i := range len(ui.layers) {
		ui.layers[i].Render(ui.renderer)
	}

	ui.ShowCursor()
	ui.ShowDebug()
	ui.renderer.Flush()

	ui.dirty = false
}

// Redraw renders just a single widget, if its state changed. No new layout is
// performed, no other widgets are affected.
//
// Parameters:
//   - widget: Widget to redraw
func (ui *UI) DrawWidget(widget Widget) {
	// Refresh, if there is more than one layer
	if len(ui.layers) > 1 {
		ui.dirty = true
		ui.Draw()
		return
	}

	ui.redraws++
	widget.Render(ui.renderer)
	ui.ShowCursor()
	ui.ShowDebug()
	ui.renderer.Flush()
}

// Redraw queues the specified widget for individual redraw optimization.
// This method provides a performance optimization by redrawing only the
// changed widget instead of the entire screen.
func (ui *UI) Redraw(widget Widget) {
	select {
	case ui.redraw <- widget:
	default:
		ui.Refresh()
	}
}

// Refresh triggers a complete screen redraw for all visible widgets.
// This method sets the dirty flag and signals the main event loop to perform
// a full rendering pass on the next iteration.
func (ui *UI) Refresh() {
	ui.dirty = true
	select {
	case ui.refresh <- struct{}{}:
	default: // Channel is full, redraw already pending
	}
}

// ShowDebug renders the debug information bar at the bottom of the screen.
// This method displays real-time debugging information when debug mode is enabled,
// providing insights into the application's current state and performance metrics.
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
		ui.renderer.Set("black", "green", "")
		ui.renderer.Text(
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
func (ui *UI) Focus(widget Widget) {
	if ui.focus != nil && ui.focus != widget {
		ui.focus.SetFlag(FlagFocused, false)
		ui.focus.Dispatch(ui.focus, EvtBlur)
	}
	if widget != nil {
		widget.SetFlag(FlagFocused, true)
	}
	ui.focus = widget
	if widget != nil {
		widget.Dispatch(widget, EvtFocus)
	}
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
func (ui *UI) SetFocus(which string) {
	var first, previous, next, last Widget
	found := false
	Traverse(ui.layers[len(ui.layers)-1], func(widget Widget) bool {
		if widget.Flag(FlagHidden) {
			return false
		}
		if !widget.Flag(FlagFocusable) || widget.Flag(FlagSkip) {
			return true
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
		return true
	})
	ui.Log(ui, Debug, "SetFocus", "which", which, "layer", ui.layers[len(ui.layers)-1].ID(), "first", ID(first), "previous", ID(previous), "next", ID(next), "last", ID(last))
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
// Supported cursor styles:
//   - "|", "bar", "steady-bar": Steady vertical bar cursor
//   - "*|", "*bar", "blinking-bar": Blinking vertical bar cursor
//   - "#", "block", "steady-block": Steady block cursor
//   - "*#", "blinking-block", "*block": Blinking block cursor
//   - "_", "underline", "steady-underline": Steady underline cursor
//   - "*_", "blinking-underline", "*underline": Blinking underline cursor
func (ui *UI) ShowCursor() {
	// Show cursor
	if ui.focus != nil {
		x, y, _, _ := ui.focus.Content()
		cx, cy, cursor := ui.focus.Cursor()
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

// ---- Loggging ------------------------------------------------------------

// Log adds a structured log entry using the application's slog logger.
// The log is routed to both the debug Text widget (human-readable) and the
// TableLog (for tabular display). The level is given as a string (e.g., "debug",
// "info", "warn", "error"). Additional key-value pairs can be provided via params.
//
// This method maintains backward compatibility with existing component logging.
func (ui *UI) Log(source Widget, level Level, message string, params ...any) {
	if ui.logger == nil {
		// Logger not initialized; ignore.
		return
	}
	slogLevel := parseLevel(level)
	args := []any{
		slog.String("source", source.ID()),
		slog.String("widgetType", WidgetType(source)),
	}
	// Append params as attributes (key, value pairs)
	for i := 0; i < len(params); i += 2 {
		if i+1 < len(params) {
			key := fmt.Sprintf("%v", params[i])
			val := params[i+1]
			args = append(args, slog.Any(key, val))
		}
	}
	ui.logger.Log(context.Background(), slogLevel, message, args...)
}

// SetLogLevel changes the minimum log level at runtime.
func (ui *UI) SetLogLevel(level slog.Level) {
	if ui.logHandler != nil {
		ui.logHandler.level = level
	}
}

// Logs returns the TableLog containing structured log entries.
func (ui *UI) Logs() *TableLog {
	return ui.tableLog
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
func (ui *UI) Popup(x, y, w, h int, popup Container) {
	// Set parent first for logging to work immediately
	popup.SetParent(ui)

	// Auto sizing
	if w == 0 || h == 0 {
		pw, ph := popup.Hint()
		style := popup.Style()
		pw += style.Horizontal()
		ph += style.Vertical()
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

	popup.SetBounds(x, y, w, h)
	popup.Layout()

	ui.focusStack = append(ui.focusStack, ui.focus)
	ui.layers = append(ui.layers, popup)
	ui.SetFocus("first")
	ui.Refresh()
}

// Close removes the topmost layer from the UI layer stack.
// This method is typically used to close popup dialogs, modal windows,
// or other overlay widgets that were added as additional layers.
// The base layer (main UI) cannot be closed using this method.
// EvtClose is dispatched to the removed layer before it is discarded.
func (ui *UI) Close() {
	if len(ui.layers) > 1 {
		top := ui.layers[len(ui.layers)-1]
		ui.layers = ui.layers[:len(ui.layers)-1]
		top.Dispatch(top, EvtClose)
		var prev Widget
		if len(ui.focusStack) > 0 {
			prev = ui.focusStack[len(ui.focusStack)-1]
			ui.focusStack = ui.focusStack[:len(ui.focusStack)-1]
		}
		// Restore focus without dispatching EvtFocus — the widget that opened
		// the popup (e.g. Combo) listens on EvtFocus to open it; re-dispatching
		// here would cause it to reopen immediately.
		if ui.focus != nil && ui.focus != prev {
			ui.focus.SetFlag(FlagFocused, false)
			ui.focus.Dispatch(ui.focus, EvtBlur)
		}
		if prev != nil {
			prev.SetFlag(FlagFocused, true)
		}
		ui.focus = prev
	}
	ui.Refresh()
}

// Confirm shows a centered modal dialog with a message and OK/Cancel buttons.
// onConfirm is called (then the dialog is closed) when the user activates OK;
// onCancel when they activate Cancel or press Escape. Either callback may be nil.
func (ui *UI) Confirm(title, message string, onConfirm, onCancel func()) {
	if title == "" {
		title = "Confirm"
	}
	b := ui.NewBuilder()
	dialog := b.
		Dialog("confirm-dialog", title).
		Class("dialog").
		Flex("confirm-body", false, "stretch", 1).
		Static("confirm-msg", message).
		Flex("confirm-buttons", true, "end", 2).
		Button("confirm-ok", "OK").
		Button("confirm-cancel", "Cancel").
		End().
		End().
		Class("").
		Container()

	Find(dialog, "confirm-ok").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		ui.Close()
		if onConfirm != nil {
			onConfirm()
		}
		return true
	})
	Find(dialog, "confirm-cancel").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		ui.Close()
		if onCancel != nil {
			onCancel()
		}
		return true
	})

	ui.Popup(-1, -1, 0, 0, dialog)
}

// Prompt shows a centered modal dialog with a message, a text input, and
// OK/Cancel buttons. onAccept is called with the input value when the user
// confirms; onCancel when they cancel or press Escape. Either callback may be nil.
func (ui *UI) Prompt(title, message string, onAccept func(string), onCancel func()) {
	if title == "" {
		title = "Prompt"
	}
	b := ui.NewBuilder()
	dialog := b.
		Dialog("prompt-dialog", title).
		Class("dialog").
		Flex("prompt-body", false, "stretch", 1).
		Static("prompt-msg", message).
		Input("prompt-input").Hint(0, 1).
		Flex("prompt-buttons", true, "end", 2).
		Button("prompt-ok", "OK").
		Button("prompt-cancel", "Cancel").
		End().
		End().
		Class("").
		Container()

	input := Find(dialog, "prompt-input").(*Input)
	accept := func() {
		text := input.Text()
		ui.Close()
		if onAccept != nil {
			onAccept(text)
		}
	}
	OnKey(input, func(e *tcell.EventKey) bool {
		if e.Key() == tcell.KeyEnter {
			accept()
			return true
		}
		return false
	})
	Find(dialog, "prompt-ok").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		accept()
		return true
	})
	Find(dialog, "prompt-cancel").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		ui.Close()
		if onCancel != nil {
			onCancel()
		}
		return true
	})

	ui.Popup(-1, -1, 0, 0, dialog)
}

// Quit signals the application to exit cleanly. Safe to call multiple times.
// This is the programmatic equivalent of pressing Ctrl+C or Ctrl+Q.
func (ui *UI) Quit() {
	ui.quitOnce.Do(func() { close(ui.quit) })
}

// ---- Run Loop -------------------------------------------------------------

// Run starts the main application event loop and blocks until the application
// exits. This is the primary method that drives the TUI application, handling
// all events, rendering, and application lifecycle management.
//
// Run spawns a background goroutine (EventLoop) to poll for tcell events
// and forward them to the main event loop channel. This prevents the main
// loop from blocking on event polling.
func (ui *UI) Run() error {
	var err error

	// Initialize screen
	ui.screen, err = tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrScreenInit, err)
	}

	if err := ui.screen.Init(); err != nil {
		return fmt.Errorf("%w: %w", ErrScreenInit, err)
	}

	defer func() {
		ui.screen.Fini()
		if r := recover(); r != nil {
			fmt.Printf("Panic: %v\n", r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	// Enable mouse events
	ui.screen.EnableMouse()

	style := tcell.StyleDefault
	ui.screen.SetStyle(style)
	ui.screen.Clear()
	ui.renderer.screen = NewTcellScreen(ui.screen)
	ui.renderer.Set("white", "black", "")

	// Take screen size for the root
	ui.width, ui.height = ui.screen.Size()
	ui.Layout()

	// Set initial focus
	ui.SetFocus("first")

	// Event handling loop
	go ui.EventLoop()

	for {
		select {
		case <-ui.quit:
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
func (ui *UI) EventLoop() {
	for {
		ev := <-ui.screen.EventQ()
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
func (ui *UI) SetTheme(theme *Theme) {
	ui.renderer.theme = theme

	// Re-apply theme styles to all widgets
	Traverse(ui, func(widget Widget) bool {
		widget.Apply(theme)
		return true
	})

	ui.Refresh()
}

// Theme returns the currently active theme.
// This method provides access to the current theme for inspection or
// for storing the current theme before switching to a different one.
//
// Returns:
//   - Theme: The currently active theme
func (ui *UI) Theme() *Theme {
	return ui.renderer.theme
}

// Dump writes a human- and LLM-readable text tree of the full UI state to w.
// It prints a header line with the screen dimensions and layer count, followed
// by the widget hierarchy of each layer. Use this to give an AI agent a
// snapshot of the current interface without running the full event loop.
func (ui *UI) Dump(w io.Writer, opts ...DumpOptions) {
	fmt.Fprintf(w, "[UI] %dx%d layers=%d\n", ui.width, ui.height, len(ui.layers))
	for i, layer := range ui.layers {
		if i > 0 {
			fmt.Fprintf(w, "  ── layer %d ──\n", i)
		}
		Dump(w, layer, opts...)
	}
}
