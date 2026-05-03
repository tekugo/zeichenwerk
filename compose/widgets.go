package compose

import (
	"time"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// ---- Widget Construction --------------------------------------------------

// BarChart creates a BarChart widget for displaying multi-series stacked bar
// charts. Configure series, categories, and display options via
// [zeichenwerk.Find] after construction.
func BarChart(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewBarChart(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Breadcrumb creates a Breadcrumb path-indicator widget. Configure segments
// and display options via [zeichenwerk.Find] after construction.
func Breadcrumb(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewBreadcrumb(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Button adds a clickable button widget to the parent. text is the button
// label. Pass "dialog" as class to apply the primary/accent button style from
// the active theme.
func Button(id, class, text string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			button := widgets.NewButton(id, class, text)
			button.Apply(theme)
			container.Add(button)
			for _, option := range options {
				option(theme, button)
			}
		}
	}
}

// Canvas adds a drawable canvas widget to the parent. pages is the number of
// canvas layers; width and height set the canvas dimensions in cells.
func Canvas(id, class string, pages, width, height int, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewCanvas(id, class, pages, width, height)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewCheckbox(id, class, text, checked)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Clock adds an animated clock widget to the parent. interval controls how
// often the display refreshes; format is a Go time-layout string (e.g.
// "15:04:05"); prefix is an optional string prepended to the time (e.g. a
// Nerd Font icon). Start the animation imperatively after building:
//
//	clk := Find(ui, "my-clock").(*Clock)
//	clk.Start()
func Clock(id, class string, interval time.Duration, format, prefix string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewClock(id, class, interval, format, prefix)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// ColorPanel adds a single-colour editor panel (3-row swatch + RGB, HSL, and
// Hex inputs). The panel emits [EvtChange] on every value change with the
// panel pointer as payload.
func ColorPanel(id, class, title string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewColorPanel(id, class, title)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// ColorPicker adds a colour picker composite to the parent. In
// [widgets.ColorSingle] mode it shows one [ColorPanel]; in
// [widgets.ColorFgBg] mode it shows two [ColorPanel]s plus a [PreviewPanel]
// side by side and exposes the WCAG contrast ratio between fg and bg.
func ColorPicker(id, class string, mode widgets.ColorPickerMode, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewColorPicker(id, class, mode)
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
//	    On(z.EvtActivate, func(_ c.Widget, _ z.Event, data ...any) bool {
//	        fmt.Println("submitted:", data[0].(string))
//	        return true
//	    }),
//	)
func Combo(id, class string, items []string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewCombo(id, class, items)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// CRT adds an animated CRT power-on/off container to the parent. It acts as
// an invisible root wrapper during normal operation. Call [zeichenwerk.CRT.Start]
// to begin the power-on animation after layout, and [zeichenwerk.CRT.PowerOff]
// to play the shutdown animation (which calls the provided callback on completion).
func CRT(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewCRT(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Custom adds a widget whose rendering is handled entirely by fn to the
// parent. fn receives the widget and the renderer on every draw call.
func Custom(id, class string, fn func(core.Widget, *core.Renderer), options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewCustom(id, class, fn)
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
func Deck(id, class string, render widgets.ItemRender, itemHeight int, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewDeck(id, class, render, itemHeight)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewDigits(id, class, text)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewEditor(id, class)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewFilter(id, class)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewForm(id, class, title, data)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewFormGroup(id, class, title, horizontal, spacing)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
			if form, ok := widget.(*widgets.Form); ok {
				widgets.BuildFormGroup(form, w, "", theme)
			}
		}
	}
}

// Grow adds an animated expanding container to the parent. When horizontal is
// true it expands along the horizontal axis; otherwise vertically. The child
// widget is revealed progressively as the animation plays.
func Grow(id, class string, horizontal bool, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewGrow(id, class, horizontal)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewHeatmap(id, class, rows, cols)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewHRule(class, style)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Indicator adds a status Indicator widget to the parent. The level drives
// the glyph colour via the indicator:<level> style variants; the label
// always renders in the base "indicator" style.
func Indicator(id, class string, level core.Level, label string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewIndicator(id, class, level, label)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewInput(id, class, params...)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// List adds a scrollable list widget to the parent. items is the initial set
// of display strings.
func List(id, class string, items []string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			list := widgets.NewList(id, class, items)
			list.Apply(theme)
			container.Add(list)
			for _, option := range options {
				option(theme, list)
			}
		}
	}
}

// Marquee adds a horizontally scrolling text widget to the parent. Start the
// animation explicitly with Start and stop it with Stop.
func Marquee(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewMarquee(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// PreviewPanel adds a fg/bg preview swatch with a WCAG contrast ratio readout
// to the parent. The panel does not emit events and is normally driven by a
// surrounding [ColorPicker].
func PreviewPanel(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewPreviewPanel(id, class)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewProgress(id, class, horizontal)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewScanner(id, class, width, style)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewSelect(id, class, args...)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Shimmer adds an animated shimmer (loading placeholder) widget to the parent.
// Start the animation explicitly with Start and stop it with Stop.
func Shimmer(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewShimmer(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Shortcuts adds a keyboard-shortcut legend widget to the parent. pairs is a
// flat list of alternating key and label strings, e.g. "↑↓", "navigate", "q", "quit".
func Shortcuts(id, class string, pairs []string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewShortcuts(id, class, pairs...)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Spacer adds an unstyled component that acts as flexible empty space inside a
// Flex container. Combine with Hint to control how much space it consumes:
//
//	Spacer("", Hint(-1, 0))  // fills all remaining horizontal space
func Spacer(class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewComponent("spacer", class)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Sparkline adds a sparkline chart widget to the parent. Supply data and
// configure the scale mode imperatively via the returned *Sparkline.
func Sparkline(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewSparkline(id, class)
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
//	sp := Find(container, "my-spinner").(*Spinner)
//	sp.Start(80 * time.Millisecond)
func Spinner(id, class string, sequence string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewSpinner(id, class, sequence)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Static adds a read-only text label to the parent.
func Static(id, class, text string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			static := widgets.NewStatic(id, class, text)
			static.Apply(theme)
			container.Add(static)
			for _, option := range options {
				option(theme, static)
			}
		}
	}
}

// Styled adds a text widget that supports inline markup and word wrapping to
// the parent.
func Styled(id, class, text string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewStyled(id, class, text)
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
func Table(id, class string, provider widgets.TableProvider, cellNav bool, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewTable(id, class, provider, cellNav)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewTabs(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Terminal adds a Terminal widget to the parent container.
// Terminal is a leaf widget that renders arbitrary ANSI/VT terminal output.
func Terminal(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewTerminal(id, class)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewText(id, class, content, follow, max)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Tiles adds a tile-grid widget to the parent. render is called for every
// visible tile; tileWidth and tileHeight are the fixed cell dimensions of each
// tile. Populate the widget with [Items] or by calling SetItems imperatively.
func Tiles(id, class string, render widgets.ItemRender, tileWidth, tileHeight int, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewTiles(id, class, render, tileWidth, tileHeight)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewTree(id, class)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// TreeFS adds a file-system tree widget to the parent. root is the directory
// to display; dirsOnly restricts the listing to directories.
func TreeFS(id, class, root string, dirsOnly bool, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewTreeFS(id, class, root, dirsOnly)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewTypeahead(id, class, params...)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}

// Typewriter adds an animated typewriter text widget to the parent. Start the
// animation explicitly with Start and stop it with Stop.
func Typewriter(id, class string, options ...Option) Option {
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewTypewriter(id, class)
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
	return func(theme *core.Theme, widget core.Widget) {
		if container, ok := widget.(core.Container); ok {
			w := widgets.NewVRule(class, style)
			w.Apply(theme)
			container.Add(w)
			for _, option := range options {
				option(theme, w)
			}
		}
	}
}
