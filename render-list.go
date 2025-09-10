package zeichenwerk

import (
	"fmt"
	"slices"
)

// renderList renders a List widget with items, selection highlighting, and optional scrollbar.
// This method handles the complete visual presentation of list widgets including
// item display, selection highlighting, line numbers, and scrollbar indicators.
//
// Parameters:
//   - list: The List widget to render
//   - x, y: Top-left coordinates of the list's content area
//   - w, h: Width and height of the list's content area
//
// Rendering features:
//  1. Displays visible items within the content area
//  2. Applies different styles for normal, highlighted, and disabled items
//  3. Shows optional line numbers with proper formatting
//  4. Renders scrollbar when content exceeds visible area
//  5. Handles item truncation for items wider than available space
//
// Visual elements:
//   - Item text with appropriate styling based on state
//   - Selection highlighting for the currently focused item
//   - Disabled item styling for non-selectable items
//   - Line numbers with consistent width formatting
//   - Vertical scrollbar indicating scroll position and content size
//
// The method automatically adjusts text width to accommodate scrollbars
// and line numbers, ensuring proper layout regardless of configuration.
func (r *Renderer) renderList(list *List, x, y, w, h int) {
	if h < 1 || w < 1 {
		return
	}

	items := list.Visible()

	// Calculate available width for text (reserve space for scrollbar if needed)
	tw := w
	if list.Scrollbar && len(list.Items) > h {
		tw = w - 1
	}

	// Calculate number width if showing numbers
	nw := 0
	if list.Numbers {
		nw = len(fmt.Sprintf("%d", len(list.Items)))
	}

	// Render each visible item
	for i, item := range items {
		if i >= h {
			break
		}

		current := list.Offset + i

		// Determine style for this item
		if slices.Contains(list.Disabled, i) {
			r.SetStyle(list.Style("disabled"))
		} else if current == list.Index {
			if list.focussed {
				r.SetStyle(list.Style("highlight"))
			} else {
				r.SetStyle(list.Style("highlight-blurred"))
			}
		}

		// Render line number if enabled
		if list.Numbers {
			r.text(x, y+i, fmt.Sprintf(" %*d \u2502 %s", nw, current+1, item), tw)
		} else {
			r.text(x, y+i, " "+item, tw)
		}

		// Reset style
		r.SetStyle(list.Style(""))
	}

	// Render scrollbar if needed
	if list.Scrollbar && len(list.Items) > h {
		scrollbarX := x + w - 1
		r.renderScrollbarV(scrollbarX, y, h, list.Offset, len(list.Items))
	}
}
