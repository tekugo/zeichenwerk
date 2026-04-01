package compose

import (
	"strings"

	zw "github.com/tekugo/zeichenwerk"
)

// ---- Composer State -------------------------------------------------------

var currentTheme *zw.Theme

type Option func(zw.Widget)

// ---- Composer Functions ---------------------------------------------------

func UI(theme *zw.Theme, options ...Option) *zw.UI {
	ui, _ := zw.NewUI(theme, nil)
	currentTheme = theme
	for _, option := range options {
		option(ui)
	}
	// Apply theme after all widgets have been added
	return ui
}

func Include(fn func() zw.Widget) Option {
	return func(widget zw.Widget) {
		if container, ok := widget.(zw.Container); ok {
			container.Add(fn())
		}
	}
}

func Build(options ...Option) zw.Widget {
	dummy := zw.NewBox("__dummy__", "", "")
	for _, option := range options {
		option(dummy)
	}
	return dummy.Children()[0]
}

// ---- Container Construction -----------------------------------------------

func Box(id, class, title string, options ...Option) Option {
	return func(widget zw.Widget) {
		if container, ok := widget.(zw.Container); ok {
			box := zw.NewBox(id, class, title)
			box.Apply(currentTheme)
			container.Add(box)
			for _, option := range options {
				option(box)
			}
		}
	}
}

func Flex(id, class string, horizontal bool, alignment string, spacing int, options ...Option) Option {
	return func(widget zw.Widget) {
		if container, ok := widget.(zw.Container); ok {
			flex := zw.NewFlex(id, "", horizontal, alignment, spacing)
			flex.Apply(currentTheme)
			container.Add(flex)
			for _, option := range options {
				option(flex)
			}
		}
	}
}

func Grid(id, class string, rows, columns []int, lines bool, options ...Option) Option {
	return func(widget zw.Widget) {
		if container, ok := widget.(zw.Container); ok {
			grid := zw.NewGrid(id, class, len(rows), len(columns), lines)
			grid.Apply(currentTheme)
			grid.Rows(rows...)
			grid.Columns(columns...)
			container.Add(grid)
			for _, option := range options {
				option(grid)
			}
		}
	}
}

func List(id, class string, items []string, options ...Option) Option {
	return func(widget zw.Widget) {
		if container, ok := widget.(zw.Container); ok {
			list := zw.NewList(id, class, items)
			list.Apply(currentTheme)
			container.Add(list)
			for _, option := range options {
				option(list)
			}
		}
	}
}

// ---- Widget Construction --------------------------------------------------

func Button(id, class, text string, options ...Option) Option {
	return func(widget zw.Widget) {
		if container, ok := widget.(zw.Container); ok {
			button := zw.NewButton(id, class, text)
			button.Apply(currentTheme)
			container.Add(button)
			for _, option := range options {
				option(button)
			}
		}
	}
}

func Static(id, class, text string, options ...Option) Option {
	return func(widget zw.Widget) {
		if container, ok := widget.(zw.Container); ok {
			static := zw.NewStatic(id, "", text)
			static.Apply(currentTheme)
			container.Add(static)
			for _, option := range options {
				option(static)
			}
		}
	}
}

// ---- Layout Options -------------------------------------------------------

func Cell(x, y, width, height int, option Option) Option {
	return func(widget zw.Widget) {
		if container, ok := widget.(zw.Container); ok {
			dummy := zw.NewBox("__cell__", "", "")
			option(dummy)
			children := dummy.Children()
			if len(children) > 0 {
				container.Add(children[0], x, y, width, height)
			}
		}
	}
}

func Hint(width, height int) Option {
	return func(widget zw.Widget) {
		widget.SetHint(width, height)
	}
}

// ---- Styling --------------------------------------------------------------

func Bg(params ...string) Option {
	return func(widget zw.Widget) {
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

func Border(params ...string) Option {
	return func(widget zw.Widget) {
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

func Font(params ...string) Option {
	return func(widget zw.Widget) {
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

func Fg(params ...string) Option {
	return func(widget zw.Widget) {
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

func Margin(a ...int) Option {
	return func(widget zw.Widget) {
		style := widget.Style("")
		if style.Fixed() {
			widget.SetStyle("", style.WithMargin(a...))
		} else {
			style.WithMargin(a...)
		}
	}
}

func Padding(a ...int) Option {
	return func(widget zw.Widget) {
		style := widget.Style("")
		if style.Fixed() {
			widget.SetStyle("", style.WithPadding(a...))
		} else {
			style.WithPadding(a...)
		}
	}
}
