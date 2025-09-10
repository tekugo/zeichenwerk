package tui

// renderScrollbarV renders a vertical scrollbar indicating scroll position and content size.
// This method creates a visual scrollbar with a thumb that represents the current
// scroll position and the proportion of visible content relative to total content.
//
// Parameters:
//   - x, y: Top-left coordinates for the scrollbar
//   - height: Height of the scrollbar area
//   - offset: Current scroll offset (how far scrolled from top)
//   - total: Total number of items/lines in the content
//
// Scrollbar calculation:
//  1. Calculates thumb size based on the ratio of visible to total content
//  2. Determines thumb position based on current scroll offset
//  3. Ensures minimum thumb size for usability
//  4. Constrains thumb position within scrollbar bounds
//
// Visual representation:
//   - Track: Light shade characters (░) for the scrollable area
//   - Thumb: Solid block characters (█) indicating current position
//   - Proportional sizing: Thumb size reflects visible content ratio
//   - Position accuracy: Thumb position reflects scroll percentage
//
// The scrollbar provides immediate visual feedback about:
//   - Current scroll position within the content
//   - Amount of content above and below the current view
//   - Proportion of total content currently visible
//
// Edge cases handled:
//   - Zero or negative dimensions (no rendering)
//   - Content smaller than view (full-height thumb)
//   - Minimum thumb size for visibility and usability
func (r *Renderer) renderScrollbarV(x, y, height, offset, total int) {
	if height <= 0 || total <= 0 {
		return
	}

	// Calculate scrollbar thumb position and size
	thumb := height * height / total
	if thumb < 1 {
		thumb = 1
	}
	if thumb > height {
		thumb = height
	}

	pos := offset * (height - thumb) / (total - height)
	if pos < 0 {
		pos = 0
	}
	if pos > height-thumb {
		pos = height - thumb
	}

	// Render scrollbar track
	for i := 0; i < height; i++ {
		var ch rune
		if i >= pos && i < pos+thumb {
			ch = '█' // Solid block for thumb
		} else {
			ch = '░' // Light shade for track
		}
		r.screen.SetContent(x, y+i, ch, nil, r.style)
	}
}
