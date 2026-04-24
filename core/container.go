package core

// Container represents a widget that can contain and manage child widgets.
// It extends the Widget interface with methods for adding, enumerating, and
// laying out children, enabling complex UI layouts and hierarchical widget
// trees such as forms, panels, grids, and flex boxes.
//
// Implementations are responsible for their own child-storage strategy (for
// example a flat slice for flex containers or a positional map for grids) and
// decide how the optional params supplied to Add are interpreted.
type Container interface {
	Widget

	// Add inserts a widget into the container as a new child.
	//
	// The params argument carries container-specific placement or layout hints
	// (for example a grid cell specification or a flex weight). Implementations
	// should document which params they accept and return an error if the
	// arguments are missing, of the wrong type, or otherwise invalid.
	//
	// When widget is nil, implementations MUST return ErrChildIsNil rather
	// than panicking or silently ignoring the call — accepting a nil child
	// would corrupt layout traversal and event dispatch.
	//
	// Single-child containers (for example Box or Collapsible) MUST replace
	// their existing child when Add is called a second time rather than
	// returning an error; this matches the mental model of "the container
	// holds at most one child" and lets callers re-parent a widget without
	// having to explicitly remove it first.
	Add(widget Widget, params ...any) error

	// Children returns a slice of all direct child widgets in their current
	// layout order. Both visible and hidden children are included so that
	// traversal, focus handling, and debugging tools see the full tree.
	//
	// Implementations MUST return an empty (but non-nil) slice when the
	// container has no children, so callers can iterate or count without a
	// nil check. The returned slice is owned by the caller and may be
	// iterated freely, but callers must not rely on mutations to it
	// affecting the container.
	Children() []Widget

	// Layout arranges the direct children within the container's content area
	// according to the container's layout strategy. It computes and assigns
	// each child's position and size, and MUST recurse into nested
	// containers so that the entire subtree is laid out in one pass —
	// callers rely on Layout to leave no unlaid descendants behind. The
	// Layout(c) helper in this package is a convenient way to perform that
	// recursion, but inline `child.Layout()` is equally acceptable.
	//
	// Layout MUST confine its effects to calling SetBounds on its children
	// (and recursing into child containers). It MUST NOT draw anything,
	// emit events, or mutate widget state beyond what SetBounds changes —
	// the renderer assumes that by the time Render runs, all geometry is
	// already resolved and no further side effects remain.
	//
	// Layout is expected to be called after the container's own bounds have
	// been set (for example by a parent layout pass or the top-level UI).
	// It returns an error if the layout cannot be satisfied — for instance
	// when required constraints conflict or a child reports an invalid size.
	Layout() error
}
