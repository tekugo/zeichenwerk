package core

// Widget represents a UI component that can be rendered and interact with
// user input. All UI elements in the TUI framework must implement this
// interface to participate in the rendering pipeline, event handling system,
// and layout management.
//
// Widgets form a tree rooted at a top-level container: each widget has an
// optional parent, an identifier, a set of styles, a bounds rectangle
// assigned during layout, and a size hint used by parent containers when
// computing that layout. Widgets receive input through Dispatch, notify
// observers via handlers registered with On, and draw themselves in Render.
//
// Most widgets share a common implementation through Component, which
// provides sensible default behavior for the majority of these methods so
// that concrete widgets only need to override what is specific to them.
type Widget interface {
	// Apply resolves the widget's styles from the given theme and assigns
	// them to the widget. It is typically invoked once when the widget is
	// attached to the UI and again whenever the theme is swapped at runtime.
	//
	// Convention: implementations are expected to delegate to Theme.Apply,
	// passing a selector derived from the widget's type (and optionally
	// parts, class, or state). Concrete widgets in this framework do this
	// through a Selector(part string) string helper on their common base
	// (see widgets.Component), which is not part of this interface because
	// it is a convention of that implementation rather than a minimal
	// requirement of every Widget. Custom widgets should still follow the
	// same pattern so that theme cascading works uniformly.
	//
	// Parameters:
	//   - theme: The theme to look up selectors in.
	Apply(theme *Theme)

	// Bounds returns the widget's outer rectangle in absolute screen
	// coordinates — the area reserved for the widget including its margin,
	// border, and padding. The values are valid only after a layout pass has
	// assigned them via SetBounds.
	//
	// Returns (x, y, width, height).
	Bounds() (int, int, int, int)

	// Content returns the widget's content rectangle in absolute screen
	// coordinates. The content area is the inner region in which the widget's
	// actual content is drawn, with margin, border, and padding already
	// subtracted from the outer bounds.
	//
	// Returns (x, y, width, height).
	Content() (int, int, int, int)

	// Cursor returns the widget's desired text cursor position, expressed in
	// coordinates relative to the upper-left corner of the content area,
	// together with the cursor style to use. When the widget does not want a
	// cursor to be shown (for example because it is not focused or does not
	// take text input), it returns an empty style string; in that case the
	// coordinates should be ignored by the caller.
	//
	// Returns (x, y, cursor-style).
	Cursor() (int, int, string)

	// Dispatch delivers an event to the widget and returns whether the event
	// was consumed. A return value of true indicates that the widget handled
	// the event and it should not propagate further (for example to parent
	// containers or global shortcut handlers); false means the event is still
	// available for other consumers.
	//
	// Parameters:
	//   - source: The widget that originated the event. This is often the
	//     same as the receiver, but for bubbled events it identifies the
	//     originating child.
	//   - event:  The event being dispatched (for example a key press,
	//     mouse event, or custom application event).
	//   - data:   Optional event-specific payload.
	Dispatch(source Widget, event Event, data ...any) bool

	// Flag returns the current value of a widget state flag (for example
	// "focused", "disabled", or "hovered"). Unknown flags return false.
	//
	// Parameters:
	//   - flag: The state flag to query.
	Flag(flag Flag) bool

	// Hint returns the widget's preferred content size. Containers consult
	// this value during layout; the exact interpretation depends on the
	// container, but the encoding of each axis is:
	//   - positive value: a fixed size in cells
	//   - negative value: a fractional/weighted size (the magnitude is the
	//     weight relative to siblings with fractional hints)
	//   - zero:           no preference — take whatever space is available
	//
	// Returns (content width, content height).
	Hint() (int, int)

	// ID returns the widget's identifier. IDs are used for widget lookup in
	// the tree and are also handy for debugging, testing, and logging. They
	// should be unique within the UI; uniqueness is the caller's
	// responsibility and is not enforced by the framework.
	ID() string

	// Info returns a human-readable, one-line description of the widget and
	// its current state. It is intended for debugging, logging, and tooling
	// rather than end-user display.
	Info() string

	// Log writes a debug message to the application's log on behalf of the
	// widget. Implementations typically forward to the UI's central logger so
	// that all widget output shares a single destination.
	//
	// Parameters:
	//   - widget:  The widget that is emitting the log entry. This is usually
	//              the receiver, but may differ when logging is delegated.
	//   - level:   The severity level of the message.
	//   - message: The log message.
	//   - data:    Optional structured payload associated with the message.
	Log(widget Widget, level Level, message string, data ...any)

	// On registers a handler to be invoked when the given event occurs on
	// the widget. Multiple handlers may be registered for the same event;
	// they are invoked in reverse registration order (LIFO), so the most
	// recently added handler sees the event first and can consume it (by
	// returning true) to prevent earlier handlers from running. See the
	// Handler type for the full consumption contract.
	//
	// Parameters:
	//   - event:   The event to listen for (for example "click" or "focus").
	//   - handler: The callback invoked when the event fires.
	On(event Event, handler Handler)

	// Parent returns the container that owns this widget in the widget tree,
	// or nil if the widget is a root or has not yet been attached to a
	// container.
	Parent() Container

	// Refresh requests that the widget be redrawn on the next rendering
	// cycle. It should be called whenever the widget's visual state has
	// changed in a way that is not already tracked by the framework (for
	// example after updating a model value that influences the display).
	//
	// The default implementation (in Component) bubbles up to the parent
	// and ultimately triggers a full-screen refresh — always correct but
	// heavier than strictly necessary. Widgets that can guarantee no
	// surrounding state depends on their own may override Refresh to call
	// Redraw(self) instead, restricting the repaint to their own bounds.
	// Choosing between the two is a trade-off between correctness and
	// performance and is the widget author's responsibility.
	Refresh()

	// Render draws the widget using the provided Renderer. Implementations
	// should render only within the widget's content area and are expected
	// to honor the currently active style and state.
	//
	// Render MUST be a pure drawing pass: it must not modify widget state,
	// assign bounds, dispatch events, or trigger layout. By the time Render
	// is invoked, all geometry is already resolved, and the framework
	// relies on this to call Render repeatedly (for partial redraws) or to
	// reorder draws without side effects.
	//
	// Parameters:
	//   - r: The renderer used to emit drawing primitives.
	Render(r *Renderer)

	// SetBounds assigns the widget's outer rectangle in absolute screen
	// coordinates. It is normally called by the parent container during a
	// layout pass; calling it directly bypasses layout and should be done
	// with care.
	//
	// Parameters:
	//   - x, y:          Top-left corner in screen coordinates.
	//   - width, height: Size in cells.
	SetBounds(x, y, width, height int)

	// SetFlag updates the value of a widget state flag. Flags typically
	// influence rendering (for example by switching to a ":focus" style) and
	// may trigger a refresh depending on the implementation.
	//
	// Parameters:
	//   - flag:  The state flag to set.
	//   - value: The new value for the flag.
	SetFlag(Flag, bool)

	// SetHint updates the widget's preferred content size. See Hint for the
	// encoding of fixed, fractional, and automatic values. Changing the hint
	// usually requires a re-layout of the parent container to take effect.
	//
	// Parameters:
	//   - width:  The preferred content width.
	//   - height: The preferred content height.
	SetHint(width, height int)

	// SetParent records the container that owns this widget. It is normally
	// called by the container's Add method and should not be invoked
	// directly by application code.
	//
	// Parameters:
	//   - parent: The owning container, or nil to detach the widget.
	SetParent(parent Container)

	// SetStyle installs the style used for the given selector. The selector
	// identifies a part and/or state (for example "" for the default,
	// ":focus" for the focused state, "bar" for a named part, or
	// "bar:hover" for a part in a specific state). Passing a nil style
	// removes any previously registered style for the selector.
	//
	// Parameters:
	//   - selector: Style selector (for example "", ":focus", "bar",
	//               "bar:hover").
	//   - style:    The style configuration to install, or nil to remove.
	SetStyle(selector string, style *Style)

	// State returns a short string describing the widget's current visual
	// state (for example "focus" or "hover"). The state is used during
	// rendering to pick the appropriate style variant.
	State() string

	// Style returns the style associated with the given selector, falling
	// back through sensible defaults when an exact match is not present.
	//
	// A selector may combine a part and a state separated by a colon
	// ("part:state"). Lookup proceeds in decreasing specificity: the
	// part:state combination is tried first, then the bare part, then the
	// bare state, and finally the default style. Implementations are
	// expected to always return a non-nil style so that callers — typically
	// the renderer — can rely on the result without additional nil checks.
	//
	// Parameters:
	//   - selector: The selector to resolve, or no arguments to request the
	//               default style.
	Style(...string) *Style
}
