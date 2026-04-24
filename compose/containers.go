package compose

import (
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// ---- Container Construction -----------------------------------------------

// Box adds a titled box container to the parent. Child options are applied to
// the box, so widget and styling options nested inside Box affect the box
// itself or add children to it.
func Box(id, class, title string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			box := widgets.NewBox(id, class, title)
			box.Apply(theme)
			container.Add(box)
			for _, option := range options {
				option(theme, box)
			}
		}
	}
}

// Card adds a titled card container to the parent. The title is rendered
// inline with the top border line. Child options are applied to the card in
// order: the first option that adds a widget becomes the content; the second
// becomes the footer. Style options (Border, Padding, Fg, etc.) apply to the
// card itself regardless of position in the option list.
//
// Style selectors: "card", "card/title".
func Card(id, class, title string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			card := widgets.NewCard(id, class, title)
			card.Apply(theme)
			container.Add(card)
			for _, option := range options {
				option(theme, card)
			}
		}
	}
}

// Collapsible adds a collapsible section container to the parent. When
// expanded is true the section starts open. Child options are applied to the
// collapsible and can add widgets inside its body.
func Collapsible(id, class, title string, expanded bool, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewCollapsible(id, class, title, expanded)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Dialog adds a dialog overlay container to the parent. Use [zeichenwerk.UI.Popup]
// to display it on screen after building it with [Build] and adding content
// to it imperatively.
func Dialog(id, class, title string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewDialog(id, class, title)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Grid adds a grid container to the parent. rows and columns are slices of
// size weights: positive values are fixed cell sizes, negative values share
// the remaining space proportionally, and zero means auto-size. lines controls
// whether separator lines are drawn between cells.
//
//	Grid("body", "", []int{0}, []int{20, -1}, false,
//	    Cell(0, 0, 1, 1, List("nav", "", []string{"Home", "About"})),
//	    Cell(1, 0, 1, 1, Static("content", "", "…")),
//	)
func Grid(id, class string, rows, columns []int, lines bool, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			grid := widgets.NewGrid(id, class, len(rows), len(columns), lines)
			grid.Apply(theme)
			grid.Rows(rows...)
			grid.Columns(columns...)
			container.Add(grid)
			for _, option := range options {
				option(theme, grid)
			}
		}
	}
}

// HFlex adds a flex container to the parent. Flex is by default horizontal,
// so children are arranged in a row. Use VFlex for vertical orientation
// (column). Although HFlex and VFlex both create a Flex widget with a
// differing flag, there is not Flex() in the composition API.
// alignment controls cross-axis alignment ("start", "center", "end",
// "stretch") and spacing sets the gap between children in cells.
func HFlex(id, class string, alignment core.Alignment, spacing int, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			flex := widgets.NewFlex(id, class, alignment, spacing)
			flex.Apply(theme)
			container.Add(flex)
			for _, option := range options {
				option(theme, flex)
			}
		}
	}
}

// Switcher adds a multi-pane container that shows one child at a time to the
// parent. Call Select on the retrieved [zeichenwerk.Switcher] to change the
// active pane. Pair with [Tabs] for a tabbed-panel layout.
func Switcher(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewSwitcher(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// VFlex adds a flex container to the parent. This flex is vertically,
// oriented, so children are arranged in a column. lthough HFlex and
// VFlex both create a Flex widget with a differing flag, there is no
// Flex() in the composition API.
// alignment controls cross-axis alignment ("start", "center", "end",
// "stretch") and spacing sets the gap between children in cells.
func VFlex(id, class string, alignment core.Alignment, spacing int, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			flex := widgets.NewFlex(id, class, alignment, spacing)
			flex.SetFlag(core.FlagVertical, true)
			flex.Apply(theme)
			container.Add(flex)
			for _, option := range options {
				option(theme, flex)
			}
		}
	}
}

// Viewport adds a scrollable single-child container to the parent. The child
// widget is scrolled with the keyboard or mouse; scrollbars are shown
// automatically when needed.
func Viewport(id, class, title string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewViewport(id, class, title)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}
