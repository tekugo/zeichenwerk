package tui

import "strings"

// renderProgressBar renders a ProgressBar widget using the configured visual style.
// This method delegates to the horizontal progress bar renderer, as the current
// implementation focuses on horizontal progress indicators.
//
// Parameters:
//   - widget: The ProgressBar widget to render
//   - x, y: Top-left coordinates of the progress bar's area
//   - w, h: Width and height of the progress bar's area
//
// The method serves as the main entry point for progress bar rendering,
// allowing for future extension to support different orientations or styles
// while maintaining a consistent interface.
func (r *Renderer) renderProgressBar(widget *ProgressBar, x, y, w, h int) {
	r.renderHorizontalProgressBar(widget, x, y, w, h)
}

// renderHorizontalProgressBar renders a horizontal progress bar with multiple visual styles.
// This method supports various rendering modes including Unicode blocks, custom fonts,
// and ASCII-based progress indicators.
//
// Parameters:
//   - widget: The ProgressBar widget containing value, range, and style information
//   - x, y: Top-left coordinates for rendering
//   - w, h: Width and height of the rendering area
//
// Supported rendering styles:
//   - "fira-code": Uses Fira Code font's progress bar glyphs for smooth appearance
//   - "bar": Uses Unicode heavy horizontal line characters (━)
//   - "unicode": Uses Unicode block characters (█ for filled, ░ for empty)
//   - default: Uses ASCII characters (# for filled, . for empty)
//
// Progress calculation:
//  1. Calculates filled portion based on current value relative to min/max range
//  2. Determines remaining empty portion
//  3. Renders appropriate characters for filled and empty sections
//  4. Applies different styles for filled vs empty portions when supported
//
// The method automatically handles edge cases like zero progress, full progress,
// and ensures the visual representation accurately reflects the current value.
func (r *Renderer) renderHorizontalProgressBar(widget *ProgressBar, x, y, w, h int) {
	hint := widget.Style("").Render
	size := w * widget.Value / (widget.Max - widget.Min)
	rest := w - size

	switch hint {
	case "fira-code":
		var text string
		if size > 0 {
			text = "\uee03"
		} else {
			text = "\uee00"
		}
		if size > 1 {
			text += strings.Repeat("\uee04", size-1)
		}
		if size > 0 && rest > 1 {
			text += strings.Repeat("\uee01", rest)
		} else if rest > 1 {
			text += strings.Repeat("\uee01", rest-1)
		}
		if rest > 0 {
			text += "\uee02"
		} else {
			text += "\uee05"
		}
		r.text(x, y, text, w)

	case "bar":
		r.SetStyle(widget.Style("bar"))
		r.text(x, y, strings.Repeat("\u2501", size), size)
		r.SetStyle(widget.Style(""))
		r.text(x+size, y, strings.Repeat("\u2501", rest), rest)

	case "unicode":
		r.SetStyle(widget.Style("bar"))
		r.text(x, y, strings.Repeat("\u2588", size), size)
		r.SetStyle(widget.Style(""))
		r.text(x+size, y, strings.Repeat("\u2591", rest), rest)

	default:
		r.text(x, y, strings.Repeat("#", size)+strings.Repeat(".", rest), w)
	}
}
