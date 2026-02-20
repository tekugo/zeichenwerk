package zeichenwerk

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
	thumb := min(max(height*height/total, 1), height)
	pos := min(max(offset*(height-thumb)/(total-height), 0), height-thumb)

	// Render scrollbar track
	for i := range height {
		var ch rune
		if i >= pos && i < pos+thumb {
			ch = '█' // Solid block for thumb
		} else {
			ch = '░' // Light shade for track
		}
		r.screen.SetContent(x, y+i, ch, nil, r.style)
	}
}

// renderScrollbarH renders a horizontal scrollbar indicating scroll position and content width.
// This method creates a visual scrollbar with a thumb that represents the current
// horizontal scroll position and the proportion of visible content relative to total content width.
//
// Parameters:
//   - x, y: Top-left coordinates for the scrollbar
//   - width: Width of the scrollbar area
//   - offset: Current horizontal scroll offset (how far scrolled from left)
//   - total: Total width of the content (in characters)
//
// Scrollbar calculation:
//  1. Calculates thumb size based on the ratio of visible to total content width
//  2. Determines thumb position based on current horizontal scroll offset
//  3. Ensures minimum thumb size for usability
//  4. Constrains thumb position within scrollbar bounds
//
// Visual representation:
//   - Track: Light shade characters (░) for the scrollable area
//   - Thumb: Solid block characters (█) indicating current position
//   - Proportional sizing: Thumb size reflects visible content ratio
//   - Position accuracy: Thumb position reflects scroll percentage
//
// The horizontal scrollbar provides immediate visual feedback about:
//   - Current horizontal scroll position within the content
//   - Amount of content to the left and right of the current view
//   - Proportion of total content width currently visible
//
// Edge cases handled:
//   - Zero or negative dimensions (no rendering)
//   - Content narrower than view (full-width thumb)
//   - Minimum thumb size for visibility and usability
//   - Division by zero when total width equals view width
func (r *Renderer) renderScrollbarH(x, y, width, offset, total int) {
	if width <= 0 || total <= 0 {
		return
	}

	// Calculate scrollbar thumb position and size
	thumb := min(max(width*width/total, 1), width)

	// Calculate thumb position, handling edge case where total <= width
	var pos int
	if total > width {
		pos = min(max(offset*(width-thumb)/(total-width), 0), width-thumb)
	} else {
		pos = 0 // Content fits within view, thumb starts at beginning
	}

	// Render horizontal scrollbar track
	for i := range width {
		var ch rune
		if i >= pos && i < pos+thumb {
			ch = '█' // Solid block for thumb
		} else {
			ch = '░' // Light shade for track
		}
		r.screen.SetContent(x+i, y, ch, nil, r.style)
	}
}
