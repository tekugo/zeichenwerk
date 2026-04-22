package compose

import (
	"strings"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// ---- Layout Options -------------------------------------------------------

// Cell wraps a single widget option for placement in a [Grid]. x and y are the
// zero-based column and row indices; width and height are the cell spans.
//
//	Grid("body", "", []int{0}, []int{20, -1}, false,
//	    Cell(0, 0, 1, 1, List("nav", "", []string{"Home"})),
//	    Cell(1, 0, 1, 1, Static("content", "", "…")),
//	)
func Cell(x, y, width, height int, option Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			dummy := widgets.NewBox("__cell__", "", "")
			option(theme, dummy)
			children := dummy.Children()
			if len(children) > 0 {
				container.Add(children[0], x, y, width, height)
			}
		}
	}
}

// Hint sets the size hint of the widget. A value of -1 means "fill remaining
// space"; 0 means "auto-size"; positive values are fixed sizes in cells.
// Hint is typically applied directly to container options or to a [Spacer]:
//
//	Spacer("", Hint(-1, 0))  // fills remaining horizontal space
//	Flex("col", "", false, "stretch", 0, Hint(0, -1), …)
func Hint(width, height int) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widget.SetHint(width, height)
	}
}

// Flag sets or clears a state flag on the widget. Use the Flag constants from
// the zeichenwerk package (e.g. [zeichenwerk.FlagRight], [zeichenwerk.FlagDisabled]):
//
//	Digits("cost", "", "0.00", Fg("$yellow"), Flag(FlagRight, true))
func Flag(flag core.Flag, value ...bool) Option {
	v := len(value) == 0 || value[0]
	return func(_ *core.Theme, widget core.Widget) {
		widget.SetFlag(flag, v)
	}
}

// ---- Styling --------------------------------------------------------------

// Bg sets the background colour of the widget. With one argument the value is
// applied to the default selector; with two arguments the first is the CSS-like
// selector and the second is the colour value (a theme variable such as "$bg1"
// or a colour name).
func Bg(params ...string) Option {
	return func(_ *core.Theme, widget core.Widget) {
		var selector, value string
		switch len(params) {
		case 0:
			return
		case 1:
			selector = ""
			value = params[0]
		default:
			selector = params[0]
			value = params[1]
		}
		style := widget.Style(selector)
		if style.Fixed() {
			widget.SetStyle(selector, style.WithBackground(value))
		} else {
			style.WithBackground(value)
		}
	}
}

// Border sets the border style of the widget. With one argument the value is
// applied to the default selector; with two or more arguments the first is the
// selector and the remaining parts are joined as the border value (e.g. "round",
// "none", "thin").
func Border(params ...string) Option {
	return func(_ *core.Theme, widget core.Widget) {
		var selector, value string
		switch len(params) {
		case 0:
			return
		case 1:
			selector = ""
			value = params[0]
		case 2:
			selector = params[0]
			value = params[1]
		default:
			selector = params[0]
			value = strings.Join(params[1:], " ")
		}
		style := widget.Style(selector)
		if style.Fixed() {
			widget.SetStyle(selector, style.WithBorder(value))
		} else {
			style.WithBorder(value)
		}
	}
}

// Font sets the font / text attribute of the widget. With one argument the
// value is applied to the default selector; with two or more arguments the
// first is the selector and the remaining parts are joined as the font value
// (e.g. "bold", "italic", "bold italic").
func Font(params ...string) Option {
	return func(_ *core.Theme, widget core.Widget) {
		var selector, value string
		switch len(params) {
		case 0:
			return
		case 1:
			selector = ""
			value = params[0]
		case 2:
			selector = params[0]
			value = params[1]
		default:
			selector = params[0]
			value = strings.Join(params[1:], " ")
		}
		style := widget.Style(selector)
		if style.Fixed() {
			widget.SetStyle(selector, style.WithFont(value))
		} else {
			style.WithFont(value)
		}
	}
}

// Fg sets the foreground (text) colour of the widget. With one argument the
// value is applied to the default selector; with two arguments the first is
// the selector and the second is the colour value.
func Fg(params ...string) Option {
	return func(_ *core.Theme, widget core.Widget) {
		var selector, value string
		switch len(params) {
		case 0:
			return
		case 1:
			selector = ""
			value = params[0]
		default:
			selector = params[0]
			value = params[1]
		}
		style := widget.Style(selector)
		if style.Fixed() {
			widget.SetStyle(selector, style.WithForeground(value))
		} else {
			style.WithForeground(value)
		}
	}
}

// Margin sets the outer margin of the widget. Arguments follow the same
// shorthand as CSS: one value sets all sides, two values set
// vertical/horizontal, and four values set top/right/bottom/left.
func Margin(a ...int) Option {
	return func(_ *core.Theme, widget core.Widget) {
		style := widget.Style("")
		if style.Fixed() {
			widget.SetStyle("", style.WithMargin(a...))
		} else {
			style.WithMargin(a...)
		}
	}
}

// Padding sets the inner padding of the widget. Arguments follow the same
// shorthand as CSS: one value sets all sides, two values set
// vertical/horizontal, and four values set top/right/bottom/left.
func Padding(a ...int) Option {
	return func(_ *core.Theme, widget core.Widget) {
		style := widget.Style("")
		if style.Fixed() {
			widget.SetStyle("", style.WithPadding(a...))
		} else {
			style.WithPadding(a...)
		}
	}
}
