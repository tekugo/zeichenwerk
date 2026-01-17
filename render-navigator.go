package zeichenwerk

import (
	"fmt"
	"strings"
)

// renderNavigator renders a Navigator widget with search input and navigation items
func (r *Renderer) renderNavigator(navigator *Navigator, x, y, w, h int) {
	if h < 2 {
		return
	}

	// Render search input at top if it exists
	if navigator.Input != nil {
		r.renderInput(navigator.Input, x, y, w, 1)
		y++
		h--
	}

	// Render navigation items
	r.renderNavigatorItems(navigator, x, y, w, h)
}

// renderNavigatorItems renders the list of navigation items from the root node
func (r *Renderer) renderNavigatorItems(navigator *Navigator, x, y, w, h int) {
	items := navigator.Items()
	start := navigator.Offset

	// Scrollbar logic
	tw := w
	if len(items) > h {
		tw = w - 1
	}

	for i := 0; i < h; i++ {
		idx := start + i
		if idx >= len(items) {
			break
		}

		item := items[idx]

		// Determine style
		if idx == navigator.Index {
			if navigator.Focused() {
				r.SetStyle(navigator.Style("highlight:focus"))
			} else {
				r.SetStyle(navigator.Style("highlight"))
			}
		} else {
			r.SetStyle(navigator.Style())
		}

		// Build label logic
		indent := strings.Repeat("  ", item.Level)
		label := fmt.Sprintf("%s%s %s", indent, item.Icon, item.Name)

		// Handle shortcut if present and space allows
		if item.Shortcut != "" {
			avail := tw - len([]rune(label)) - 1
			if avail > len([]rune(item.Shortcut)) {
				padding := strings.Repeat(" ", avail-len([]rune(item.Shortcut)))
				label += padding + item.Shortcut
			}
		}

		r.text(x, y+i, label, tw)

		r.SetStyle(navigator.Style())
	}

	// Render scrollbar
	if len(items) > h {
		r.renderScrollbarV(x+w-1, y, h, navigator.Offset, len(items))
	}
}
