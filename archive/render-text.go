package zeichenwerk

// renderText renders a Text widget's content with scrolling support.
// This method displays the visible portion of multi-line text content,
// handling both horizontal and vertical scrolling offsets.
//
// Parameters:
//   - text: The Text widget containing content and scroll offsets
//   - x, y: Top-left coordinates of the text rendering area
//   - w, h: Width and height of the text rendering area
//
// Rendering process:
//  1. Calculates visible content range based on vertical scroll offset
//  2. Iterates through visible lines within the height constraint
//  3. Applies horizontal scroll offset to each line
//  4. Renders each line using the text renderer with width limiting
//
// Scrolling behavior:
//   - Vertical scrolling: Shows lines starting from offsetY
//   - Horizontal scrolling: Shows characters starting from offsetX
//   - Automatic clipping: Content beyond w×h area is not rendered
//   - Line-by-line rendering: Each line is rendered independently
//
// The method handles:
//   - Content that exceeds the widget's display area
//   - Empty content (no lines to render)
//   - Lines shorter than the horizontal offset
//   - Proper text truncation and width management
//
// This enables Text widgets to display large documents, logs, or any
// multi-line content with smooth scrolling in both directions.
func (r *Renderer) renderText(text *Text, x, y, w, h int) {
	// Check, if we need to render a vertical scroll bar
	iw := w
	if len(text.content) > h {
		iw--
	}

	// Check, if we need to render a horizontal scroll bar
	ih := h
	if text.longest > iw {
		ih--
	}

	if iw < w {
		r.renderScrollbarV(x+w-1, y, ih, text.offsetY, len(text.content))
	}

	if ih < h {
		r.renderScrollbarH(x, y+h-1, iw, text.offsetX, text.longest)
	}

	// Render visible text content
	for i := range len(text.content) - text.offsetY {
		if i >= ih {
			break
		}
		// TODO - use runes, not bytes
		line := text.content[text.offsetY+i]
		if text.offsetX < len(line) {
			r.text(x, y+i, line[text.offsetX:], iw)
		}
	}
}
