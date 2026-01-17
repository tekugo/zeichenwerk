package next

// Screen represents an abstraction over terminal screen operations for rendering.
type Screen interface {
	// Clear the screen.
	//
	// This usually fills the screen with the default background color/style,
	// effectively erasing all previous content.
	Clear()

	// Sets the clipping region for subsequent put operations.
	//
	// The clipping region defines the bounds within which content can be rendered.
	// Any content drawn outside this region will be ignored.
	//
	// Parameters:
	//   - x, y: The top-left corner of the clipping region (Absolute Screen Coordinates).
	//   - width, height: The dimensions of the clipping region.
	//
	// Note: Calling Clip naturally restricts the drawing area. If a clipping
	// region was already set, the new method should probably be the intersection
	// of the previous and the new region!
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
	//   - string: The character at the position, or 0 if out of bounds.
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

	// Set the foreground, background and font for subsequent put operations.
	//
	// Parameters:
	//   - fg: The foreground color (e.g., "red", "#ff0000", or theme var).
	//   - bg: The background color.
	//   - font: The font attributes (e.g., "bold, italic").
	//
	// This state persists until Set is called again.
	Set(fg, bg, font string)
}
