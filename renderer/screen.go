package renderer

// Screen is the minimum interface that a rendering back-end must provide.
// It models the terminal as a grid of styled cells and exposes just enough
// primitives for the higher-level Renderer to build on: writing and
// reading single cells, changing the active style, scoping output with a
// clipping region and a coordinate translation, clearing, and flushing.
//
// A Screen is owned by a single goroutine; there is no internal locking.
// The Put, Get, Set, Clip and Translate operations are all stateful, so
// calling code must establish the intended clip/translate/style context
// before a batch of draw calls.
//
// TcellScreen is the production implementation (driven by gdamore/tcell);
// tests usually use core.TestScreen, which records every Put for later
// inspection.
type Screen interface {
	// Clear the screen.
	//
	// This usually fills the screen with the default background color/style,
	// effectively erasing all previous content.
	Clear()

	// Clip installs the clipping region for subsequent Put operations. The
	// supplied (x, y) also becomes the new coordinate origin: after Clip,
	// a Put at (0, 0) targets the top-left cell of the clip rectangle in
	// absolute screen coordinates. Content drawn outside the region is
	// silently discarded by the implementation.
	//
	// The current implementation replaces any previously installed clip
	// rather than intersecting with it, so callers that want to restrict
	// a nested clip must intersect explicitly before calling Clip.
	//
	// Parameters:
	//   - x, y:          Top-left corner in absolute screen coordinates,
	//     also the new local origin for Put/Get.
	//   - width, height: Dimensions of the clip rectangle. A value of 0
	//     means "unlimited on this axis".
	Clip(x, y, width, height int)

	// Flush the screen to applying all changes visible.
	//
	// This synchronizes the internal buffer with the actual terminal output.
	// Should be called after a batch of drawing operations.
	Flush()

	// Get the rune from the specified position.
	//
	// Parameters:
	//   - x, y: The position of the rune to get (relative to the clipping region).
	//
	// Returns:
	//   - string: The character at the position, or empty string if out of bounds.
	//
	// Note: This method only returns the primary rune. Styling information
	// or combining characters are not returned.
	Get(x, y int) string

	// Put a rune at the specified position.
	//
	// Parameters:
	//   - x, y: The position of the rune to put (relative to the clipping region).
	//   - ch: The rune to display.
	//
	// Note: This method relies on the current style set via Set().
	// Combining characters: This interface currently supports simple runes.
	// Complex combining character sequences might essentially need a string
	// or []rune interface.
	Put(x, y int, ch string)

	// Set installs the foreground, background and font for subsequent Put
	// operations. The new style remains in effect until Set is called
	// again, so callers typically Set once at the start of a draw region
	// and reuse it across many Put calls.
	//
	// Colour strings are literal values understood by the back-end (named
	// colours such as "red" or hex triplets such as "#ff0000"). The
	// renderer package does not know about theme variables; those are
	// resolved at a higher layer (core.Renderer) before reaching Screen.
	//
	// Parameters:
	//   - fg:   Foreground colour, or an empty string to keep the default.
	//   - bg:   Background colour, or an empty string to keep the default.
	//   - font: Space-separated font attributes (for example
	//     "bold italic"). Recognised tokens include blink, bold, italic,
	//     normal, strikethrough and underline.
	Set(fg, bg, font string)

	// Translate shifts the coordinate system used by Put and Get by an
	// additional offset applied on top of the current clip origin. It is
	// used to implement scrolling views: after Translate(-tx, -ty), a
	// child widget that draws at its own origin (0, 0) will appear at
	// screen offset (clip_x - tx, clip_y - ty), effectively scrolling its
	// content into view.
	//
	// Pass (0, 0) to remove any translation.
	//
	// Parameters:
	//   - x, y: Offsets added to coordinates passed to Put and Get.
	Translate(x, y int)

	// SetUnderline sets the underline style and colour used by subsequent
	// Put calls. The underline is written in addition to the rest of the
	// current Set style; calling Set again resets the underline along
	// with the other attributes.
	//
	// Parameters:
	//   - style: 0 = none, 1 = single, 2 = double, 3 = curly, 4 = dotted,
	//     5 = dashed.
	//   - color: Empty string keeps the terminal default colour;
	//     otherwise a named or hex colour.
	SetUnderline(style int, color string)
}
