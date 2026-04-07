// Package compose provides a functional composition API for building
// zeichenwerk terminal UIs.
//
// Every widget is represented as an [Option] — a plain function value that
// adds a widget to its parent and applies child options recursively. Options
// can be nested, stored, and passed around freely. The active theme flows
// through the tree automatically; no global state is used.
//
// Entry points are [UI] (builds a complete UI and returns [*zeichenwerk.UI])
// and [Build] (builds a single widget subtree, useful for screen functions
// passed to [Include]).
//
// Styling, layout, and event options ([Bg], [Fg], [Font], [Border], [Padding],
// [Margin], [Hint], [On], etc.) are applied to the widget that receives them,
// not to their children, so they can appear at any position in the option list.
//
// When direct widget access is needed after construction — for example to wire
// events imperatively, populate a [Tree], or start animations — retrieve the
// widget with [zeichenwerk.Find] and call its methods directly.
package compose

import (
	"strings"

	"github.com/gdamore/tcell/v3"
	z "github.com/tekugo/zeichenwerk"
)

// ---- Composer State -------------------------------------------------------

// Option is a function that adds or configures a widget inside a parent widget.
// Every constructor in this package returns an Option, and styling / event
// helpers are Options as well, so the entire UI tree is expressed as nested
// Option calls.
//
// The theme parameter carries the active theme down the tree. The widget
// parameter is the parent widget that the Option should add its content to.
type Option func(*z.Theme, z.Widget)

// ---- Composer Functions ---------------------------------------------------

// UI creates a new [zeichenwerk.UI] with the given theme, applies all options
// to it, and returns it. Call Run on the result to start the event loop.
//
//	UI(TokyoNightTheme(),
//	    Flex("root", "", false, "stretch", 0,
//	        Static("title", "", "Hello"),
//	    ),
//	).Run()
func UI(theme *z.Theme, options ...Option) *z.UI {
	ui := z.NewUI(theme, nil)
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
//	func header(theme *z.Theme) z.Widget {
//	    return Build(theme, Static("title", "", "My App"))
//	}
func Include(fn func(*z.Theme) z.Widget) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			container.Add(fn(theme))
		}
	}
}

// Build applies options to a temporary container and returns the first child
// widget produced. It is the standard way to construct a widget subtree inside
// a screen function passed to [Include].
//
//	func header(theme *z.Theme) z.Widget {
//	    return Build(theme,
//	        Flex("header", "", true, "center", 0,
//	            Static("title", "", "My App", Font("bold")),
//	        ),
//	    )
//	}
func Build(theme *z.Theme, options ...Option) z.Widget {
	dummy := z.NewBox("__dummy__", "", "")
	for _, option := range options {
		option(theme, dummy)
	}
	children := dummy.Children()
	if len(children) == 0 {
		panic("compose.Build: no widget was added — check that at least one Option adds a child")
	}
	return children[0]
}

// ---- Container Construction -----------------------------------------------

// Box adds a titled box container to the parent. Child options are applied to
// the box, so widget and styling options nested inside Box affect the box
// itself or add children to it.
func Box(id, class, title string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			box := z.NewBox(id, class, title)
			box.Apply(theme)
			container.Add(box)
			for _, option := range options {
				option(theme, box)
			}
		}
	}
}

// Flex adds a flex container to the parent. When horizontal is true children
// are arranged in a row; otherwise in a column. alignment controls cross-axis
// alignment ("start", "center", "end", "stretch") and spacing sets the gap
// between children in cells.
func Flex(id, class string, horizontal bool, alignment string, spacing int, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			flex := z.NewFlex(id, class, horizontal, alignment, spacing)
			flex.Apply(theme)
			container.Add(flex)
			for _, option := range options {
				option(theme, flex)
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
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			grid := z.NewGrid(id, class, len(rows), len(columns), lines)
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

// List adds a scrollable list widget to the parent. items is the initial set
// of display strings.
func List(id, class string, items []string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			list := z.NewList(id, class, items)
			list.Apply(theme)
			container.Add(list)
			for _, option := range options {
				option(theme, list)
			}
		}
	}
}

// Spacer adds an unstyled component that acts as flexible empty space inside a
// Flex container. Combine with Hint to control how much space it consumes:
//
//	Spacer("", Hint(-1, 0))  // fills all remaining horizontal space
func Spacer(class string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewComponent("spacer", class)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Collapsible adds a collapsible section container to the parent. When
// expanded is true the section starts open. Child options are applied to the
// collapsible and can add widgets inside its body.
func Collapsible(id, class, title string, expanded bool, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewCollapsible(id, class, title, expanded)
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
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewDialog(id, class, title)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Form adds a data-bound form container to the parent. data must be a pointer
// to a struct; fields are rendered as labelled controls. Embed a [FormGroup]
// inside to control layout. Use struct tags to customise labels, widths, and
// control types.
func Form(id, class, title string, data any, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewForm(id, class, title, data)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// FormGroup adds a layout group inside a [Form]. When horizontal is true,
// controls are arranged in a row; otherwise in a column. spacing sets the gap
// between controls.
//
// When FormGroup is nested directly inside a [Form] option, it automatically
// generates form controls from the form's bound struct, mirroring the
// behaviour of the Builder's Group method.
func FormGroup(id, class, title string, horizontal bool, spacing int, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewFormGroup(id, class, title, horizontal, spacing)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
			if form, ok := widget.(*z.Form); ok {
				z.BuildFormGroup(form, w, "", theme)
			}
		}
	}
}

// Grow adds an animated expanding container to the parent. When horizontal is
// true it expands along the horizontal axis; otherwise vertically. The child
// widget is revealed progressively as the animation plays.
func Grow(id, class string, horizontal bool, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewGrow(id, class, horizontal)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Switcher adds a multi-pane container that shows one child at a time to the
// parent. Call Select on the retrieved [zeichenwerk.Switcher] to change the
// active pane. Pair with [Tabs] for a tabbed-panel layout.
func Switcher(id, class string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewSwitcher(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Viewport adds a scrollable single-child container to the parent. The child
// widget is scrolled with the keyboard or mouse; scrollbars are shown
// automatically when needed.
func Viewport(id, class, title string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewViewport(id, class, title)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// ---- Widget Construction --------------------------------------------------

// Button adds a clickable button widget to the parent. text is the button
// label. Pass "dialog" as class to apply the primary/accent button style from
// the active theme.
func Button(id, class, text string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			button := z.NewButton(id, class, text)
			button.Apply(theme)
			container.Add(button)
			for _, option := range options {
				option(theme, button)
			}
		}
	}
}

// Static adds a read-only text label to the parent.
func Static(id, class, text string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			static := z.NewStatic(id, class, text)
			static.Apply(theme)
			container.Add(static)
			for _, option := range options {
				option(theme, static)
			}
		}
	}
}

// Terminal adds a Terminal widget to the parent container.
// Terminal is a leaf widget that renders arbitrary ANSI/VT terminal output.
func Terminal(id, class string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewTerminal(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Checkbox adds a toggleable checkbox widget to the parent. checked sets the
// initial state.
func Checkbox(id, class, text string, checked bool, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewCheckbox(id, class, text, checked)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Input adds a single-line text input widget to the parent. params is passed
// directly to the underlying constructor; the first element is typically used
// as placeholder text.
func Input(id, class string, params []string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewInput(id, class, params...)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Typeahead adds a text input with autocomplete suggestions to the parent.
// params is passed directly to the underlying constructor; the first element
// is typically used as placeholder text.
func Typeahead(id, class string, params []string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewTypeahead(id, class, params...)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Filter adds a filter input widget that can be bound to a List or Tree.
func Filter(id, class string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewFilter(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Combo adds a traditional combo box to the parent: a free-text [Typeahead]
// input paired with a suggestion [List]. items is the initial suggestion set.
//
// The widget dispatches [EvtChange] (string) on every keystroke and
// [EvtActivate] (string) when Enter is pressed. Because the [EvtActivate]
// payload is a string rather than an int, use [On] instead of [OnActivate]
// to register a confirmation handler:
//
//	Combo("search", "", history,
//	    On(z.EvtActivate, func(_ z.Widget, _ z.Event, data ...any) bool {
//	        fmt.Println("submitted:", data[0].(string))
//	        return true
//	    }),
//	)
func Combo(id, class string, items []string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewCombo(id, class, items)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Select adds a dropdown selection widget to the parent. args is a flat list
// of alternating value/label pairs: []string{"key1", "Label 1", "key2", "Label 2", …}.
func Select(id, class string, args []string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewSelect(id, class, args...)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Tree adds a hierarchical tree widget to the parent. Populate the tree and
// expand nodes imperatively after construction via [zeichenwerk.Find]:
//
//	tree := z.Find(container, "my-tree").(*z.Tree)
//	root := z.NewTreeNode("root")
//	tree.Add(root)
//	tree.Expand(root)
func Tree(id, class string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewTree(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Tabs adds a tab-bar widget to the parent. Tab labels must be added
// imperatively after construction:
//
//	tabs := z.Find(container, "my-tabs").(*z.Tabs)
//	tabs.Add("Tab One")
//	tabs.Add("Tab Two")
//
// Pair with [Switcher] to build a full tabbed-panel layout.
func Tabs(id, class string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewTabs(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Text adds a scrollable multi-line text display to the parent. content is the
// initial slice of lines. When follow is true the view scrolls to the bottom
// automatically as new lines are added. max caps the total number of lines
// retained; older lines are discarded when the limit is reached.
func Text(id, class string, content []string, follow bool, max int, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewText(id, class, content, follow, max)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Styled adds a text widget that supports inline markup and word wrapping to
// the parent.
func Styled(id, class, text string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewStyled(id, class, text)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Digits adds a large-character numeric display widget to the parent. text is
// the initial value string.
func Digits(id, class, text string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewDigits(id, class, text)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Progress adds a progress bar widget to the parent. When horizontal is true
// the bar fills left-to-right; otherwise bottom-to-top. Use [Total] and
// [Value] to set the range and current value at construction time, or retrieve
// the widget imperatively to update it at runtime.
func Progress(id, class string, horizontal bool, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewProgress(id, class, horizontal)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Spinner adds an animated spinner widget to the parent. sequence is the
// animation frame string; use one of the entries in [zeichenwerk.Spinners]
// for a built-in sequence. Start and stop the animation imperatively:
//
//	sp := z.Find(container, "my-spinner").(*z.Spinner)
//	sp.Start(80 * time.Millisecond)
func Spinner(id, class string, sequence string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewSpinner(id, class, sequence)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// HRule adds a horizontal rule (divider line) to the parent. style selects
// the line style (e.g. "thin", "thick", "double").
func HRule(class, style string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewHRule(class, style)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// VRule adds a vertical rule (divider line) to the parent. style selects the
// line style (e.g. "thin", "thick", "double").
func VRule(class, style string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewVRule(class, style)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Editor adds a multi-line text editor widget to the parent. Set the initial
// content and enable line numbers with [Content] and [LineNumbers], or
// retrieve the widget imperatively for runtime updates.
func Editor(id, class string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewEditor(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Canvas adds a drawable canvas widget to the parent. pages is the number of
// canvas layers; width and height set the canvas dimensions in cells.
func Canvas(id, class string, pages, width, height int, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewCanvas(id, class, pages, width, height)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Deck adds a virtualised list widget to the parent that renders each item
// with a custom render function. render is called for every visible item;
// itemHeight is the fixed height of each item in cells. Populate the deck
// with [Items] or by calling SetItems imperatively.
func Deck(id, class string, render z.ItemRender, itemHeight int, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewDeck(id, class, render, itemHeight)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Table adds a scrollable data-grid widget to the parent. provider supplies
// the column headers and row data; use [zeichenwerk.NewArrayTableProvider]
// for a simple static data source.
func Table(id, class string, provider z.TableProvider, cellNav bool, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewTable(id, class, provider, cellNav)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Scanner adds an animated scanning-bar widget to the parent. width is the
// bar width in cells; style selects the visual style.
func Scanner(id, class string, width int, style string, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewScanner(id, class, width, style)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Heatmap creates a Heatmap widget with the given dimensions. Configure data,
// labels, and cell width via [zeichenwerk.Find] after construction.
func Heatmap(id, class string, rows, cols int, options ...Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			w := z.NewHeatmap(id, class, rows, cols)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// ---- Layout Options -------------------------------------------------------

// Cell wraps a single widget option for placement in a [Grid]. x and y are the
// zero-based column and row indices; width and height are the cell spans.
//
//	Grid("body", "", []int{0}, []int{20, -1}, false,
//	    Cell(0, 0, 1, 1, List("nav", "", []string{"Home"})),
//	    Cell(1, 0, 1, 1, Static("content", "", "…")),
//	)
func Cell(x, y, width, height int, option Option) Option {
	return func(theme *z.Theme, widget z.Widget) {
		if container, ok := widget.(z.Container); ok {
			dummy := z.NewBox("__cell__", "", "")
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
	return func(_ *z.Theme, widget z.Widget) {
		widget.SetHint(width, height)
	}
}

// ---- Styling --------------------------------------------------------------

// Bg sets the background colour of the widget. With one argument the value is
// applied to the default selector; with two arguments the first is the CSS-like
// selector and the second is the colour value (a theme variable such as "$bg1"
// or a colour name).
func Bg(params ...string) Option {
	return func(_ *z.Theme, widget z.Widget) {
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
	return func(_ *z.Theme, widget z.Widget) {
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
	return func(_ *z.Theme, widget z.Widget) {
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
	return func(_ *z.Theme, widget z.Widget) {
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
	return func(_ *z.Theme, widget z.Widget) {
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
	return func(_ *z.Theme, widget z.Widget) {
		style := widget.Style("")
		if style.Fixed() {
			widget.SetStyle("", style.WithPadding(a...))
		} else {
			style.WithPadding(a...)
		}
	}
}

// ---- Event Handling -------------------------------------------------------

// On registers a raw event handler on the widget. Use the typed helpers
// ([OnActivate], [OnChange], etc.) when available for a cleaner call site.
func On(event z.Event, handler z.Handler) Option {
	return func(_ *z.Theme, widget z.Widget) {
		widget.On(event, handler)
	}
}

// OnAccept registers a handler called when the user accepts a suggested or
// pending value (e.g. pressing Tab in a [Typeahead]). value is the accepted
// string.
func OnAccept(fn func(value string) bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnAccept(widget, fn)
	}
}

// OnActivate registers a handler called when the user activates an item
// (e.g. pressing Enter on a [List] row or a [Button]). index is the
// zero-based position of the activated item.
func OnActivate(fn func(index int) bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnActivate(widget, fn)
	}
}

// OnChange registers a handler called when the widget value changes
// (e.g. typing in an [Input] or toggling a [Checkbox]). value is the new
// string representation of the widget's value.
func OnChange(fn func(value string) bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnChange(widget, fn)
	}
}

// OnEnter registers a handler called when the user confirms input by pressing
// Enter (e.g. in an [Input] field). value is the current input string.
func OnEnter(fn func(value string) bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnEnter(widget, fn)
	}
}

// OnKey registers a handler called on every key event received by the widget.
func OnKey(fn func(*tcell.EventKey) bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnKey(widget, fn)
	}
}

// OnMouse registers a handler called on every mouse event received by the widget.
func OnMouse(fn func(*tcell.EventMouse) bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnMouse(widget, fn)
	}
}

// OnHide registers a handler called when the widget becomes hidden.
func OnHide(fn func() bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnHide(widget, fn)
	}
}

// OnSelect registers a handler called when the highlighted item changes
// (e.g. moving through a [List] or [Deck] before activation). index is the
// zero-based position of the newly highlighted item.
func OnSelect(fn func(index int) bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnSelect(widget, fn)
	}
}

// OnShow registers a handler called when the widget becomes visible.
func OnShow(fn func() bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		z.OnShow(widget, fn)
	}
}

// ---- Widget Options -------------------------------------------------------

// Total sets the maximum value of a [Progress] bar at construction time.
// It is a no-op when applied to any other widget type.
func Total(n int) Option {
	return func(_ *z.Theme, widget z.Widget) {
		if w, ok := widget.(*z.Progress); ok {
			w.SetTotal(n)
		}
	}
}

// Value sets the current value of a [Progress] bar at construction time.
// It is a no-op when applied to any other widget type.
func Value(n int) Option {
	return func(_ *z.Theme, widget z.Widget) {
		if w, ok := widget.(*z.Progress); ok {
			w.Set(n)
		}
	}
}

// Content sets the initial text content of an [Editor] widget at construction
// time. Each element of lines becomes one editor line. It is a no-op when
// applied to any other widget type.
func Content(lines []string) Option {
	return func(_ *z.Theme, widget z.Widget) {
		if w, ok := widget.(*z.Editor); ok {
			w.SetContent(lines)
		}
	}
}

// LineNumbers controls whether line numbers are shown in an [Editor] widget.
// It is a no-op when applied to any other widget type.
func LineNumbers(show bool) Option {
	return func(_ *z.Theme, widget z.Widget) {
		if w, ok := widget.(*z.Editor); ok {
			w.ShowLineNumbers(show)
		}
	}
}

// Items sets the initial data items of a [Deck] widget at construction time.
// It is a no-op when applied to any other widget type.
func Items(items []any) Option {
	return func(_ *z.Theme, widget z.Widget) {
		if w, ok := widget.(*z.Deck); ok {
			w.SetItems(items)
		}
	}
}
