package compose

import (
	"github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

// ---- Composer State -------------------------------------------------------

// Option is a function that adds or configures a widget inside a parent widget.
// Every constructor in this package returns an Option, and styling / event
// helpers are Options as well, so the entire UI tree is expressed as nested
// Option calls.
//
// The theme parameter carries the active theme down the tree. The widget
// parameter is the parent widget that the Option should add its content to.
type Option func(*core.Theme, core.Widget)

// ---- Composer Functions ---------------------------------------------------

// UI creates a new [zeichenwerk.UI] with the given theme, applies all options
// to it, and returns it. Call Run on the result to start the event loop.
//
//	UI(TokyoNightTheme(),
//	    Flex("root", "", false, "stretch", 0,
//	        Static("title", "", "Hello"),
//	    ),
//	).Run()
func UI(theme *core.Theme, options ...Option) *zeichenwerk.UI {
	ui := zeichenwerk.NewUI(theme, nil)
	for _, option := range options {
		option(theme, ui)
	}
	return ui
}

// Include wraps a screen function so it can be used as an Option inside a
// container. fn receives the current theme and must return the root widget of
// the sub-tree it builds, typically via [Build].
//
//	UI(theme,
//	    Flex("root", "", false, "stretch", 0,
//	        Include(header),
//	        Include(footer),
//	    ),
//	).Run()
//
//	func header(theme *c.Theme) c.Widget {
//	    return Build(theme, Static("title", "", "My App"))
//	}
func Include(fn func(*core.Theme) core.Widget) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			container.Add(fn(theme))
		}
	}
}

// Build applies options to a temporary container and returns the first child
// widget produced. It is the standard way to construct a widget subtree inside
// a screen function passed to [Include].
//
//	func header(theme *c.Theme) c.Widget {
//	    return Build(theme,
//	        Flex("header", "", true, "center", 0,
//	            Static("title", "", "My App", Font("bold")),
//	        ),
//	    )
//	}
func Build(theme *core.Theme, options ...Option) core.Widget {
	dummy := widgets.NewBox("__dummy__", "", "")
	for _, option := range options {
		option(theme, dummy)
	}
	children := dummy.Children()
	if len(children) == 0 {
		panic("compose.Build: no widget was added — check that at least one Option adds a child")
	}
	return children[0]
}
