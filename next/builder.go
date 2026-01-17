package next

import (
	"fmt"
	"strings"
)

// Builder provides a fluent interface for constructing TUI components.
// It maintains a stack of containers and applies styling through themes.
// The builder pattern allows for method chaining to create complex UI
// layouts in an easy and descriptive way.
//
// The widget manipulation methods take variable arguments for optional
// selector and value. The selector is a CSS-like selector string (e.g.,
// ":focus", ":hover") and the value is the value to apply to the widget.
// If only one argument is provided, it is assumed to be the value and the
// selector is empty.
type Builder struct {
	theme      *Theme           // Current theme for styling widgets
	stack      Stack[Container] // Stack of container widgets for nesting
	current    Widget           // Currently active widget being configured
	class      string           // CSS-like class name for styling
	x, y, w, h int              // Grid cell coordinates and dimensions
}

// NewBuilder creates a new Builder instance with the specified theme.
// Returns a pointer to the newly created Builder.
func NewBuilder(theme *Theme) *Builder {
	return &Builder{theme: theme}
}

// ---- Internal Helper Methods -----------------------------------------------

// selector constructs a CSS-like selector string for styling widgets.
// It combines the widget type (t) with optional class and id modifiers.
// Format: "type.class#id" where class and id are optional.
// Example: "button.primary#submit" for a button with class "primary" and
// id "submit".
func (b *Builder) selector(t, id string) string {
	if b.class != "" {
		t = t + "." + b.class
	}
	if id != "" {
		t = t + "#" + id
	}
	return t
}

// ---- Builder Methods ------------------------------------------------------

// Build finalizes the UI construction and returns a complete UI instance.
// It creates a new UI with the current theme and root container from the
// stack.
func (b *Builder) Build() *UI {
	ui, _ := NewUI(b.theme, b.stack.Peek(), true)
	return ui
}

// Container returns the current top-level container from the builder's stack.
// Returns nil if no containers have been created yet.
func (b *Builder) Container() Container {
	return b.stack.Peek()
}

// Find tries to find the widget with the given id.
// Find searches the complete widget tree from the top of the stack, so also
// adjacent widgets will be found.
func (b *Builder) Find(id string) Widget {
	return Find(b.stack[0], id)
}

// Run builds and runs the UI in one go.
// Short-hand method for Build() and Run().
func (b *Builder) Run() {
	b.Build().Run()
}

// With applies a builder function to this builder instance, enabling
// composition and reusable UI building patterns.
//
// Parameters:
//   - fn: Builder function to apply
//
// Using With you can move building of parts of the UI into your own methods
// for better separation and still use the fluent API.
func (b *Builder) With(fn func(*Builder)) *Builder {
	fn(b)
	return b
}

// ---- Widget Construction --------------------------------------------------

// Add adds a widget to the current container and sets it as the current
// widget. If there's a container on the stack, the widget is added to it.
// The widget's parent is set to the container for proper hierarchy handling.
func (b *Builder) Add(widget Widget) *Builder {
	if len(b.stack) > 0 {
		top := b.stack.Peek()
		switch top := top.(type) {
		case *Box:
			top.Add(widget)
		case *Flex:
			top.Add(widget)
		case *Grid:
			top.Add(b.x, b.y, b.w, b.h, widget)
		}
		widget.SetParent(top)
	}
	b.Apply(widget)
	b.current = widget
	if container, ok := widget.(Container); ok {
		b.stack.Push(container)
	}
	return b
}

// End finalizes the current container and pops it from the stack.
// This should be called after adding all children to a container except
// the root container.
func (b *Builder) End() *Builder {
	if len(b.stack) > 1 {
		b.current = b.stack.Pop()
	}
	return b
}

// Apply applies the styles of a theme to a widget based on its type and ID.
// The theme styles are applied automatically when the builder is used,
// but if widgets are created outside of the builder, they can be styled
// using this method.
func (b *Builder) Apply(widget Widget) {
	switch widget := widget.(type) {
	case *Box:
		b.theme.Apply(widget, b.selector("box", widget.ID()))
		b.theme.Apply(widget, b.selector("box/shadow", widget.ID()))
	case *Flex:
		b.theme.Apply(widget, b.selector("flex", widget.ID()))
	case *Grid:
		b.theme.Apply(widget, b.selector("grid", widget.ID()))
	case *List:
		b.theme.Apply(widget, b.selector("list", widget.ID()), "disabled", "focus")
		b.theme.Apply(widget, b.selector("list/highlight", widget.ID()), "focus")
	case *Static:
		b.theme.Apply(widget, b.selector("static", widget.ID()))
	default:
		panic(fmt.Errorf("no style for widget type %T", widget))
	}
}

// ---- Widget Creation ------------------------------------------------------

// Box creates a new box widget with the specified id and display title.
// The box is automatically styled with theme styles for the border and
// the title.
func (b *Builder) Box(id, title string) *Builder {
	box := NewBox(id, title)
	b.Add(box)
	return b
}

// Flex creates a new flex container widget for arranging child widgets.
//
// Parameters:
//   - id: unique identifier for the flex container
//   - horizontal: whether the flex container is horizontal
//   - alignment: how children are aligned ("start", "center", "end", "stretch")
//   - spacing: cells between child widgets (columns or rows)
func (b *Builder) Flex(id string, horizontal bool, alignment string, spacing int) *Builder {
	flex := NewFlex(id, horizontal, alignment, spacing)
	b.Add(flex)
	return b
}

// Grid creates a new grid container widget for arranging widgets in a table
// layout.
//
// Parameters:
//   - id: unique identifier for the grid container
//   - rows: number of rows in the grid
//   - columns: number of columns in the grid
//   - lines: whether to show grid lines
//
// Use Cell() to specify where subsequent widgets should be placed.
// Initially all rows and columns are initialized to use fractional sizes
// at one fraction each (i.e. -1).
func (b *Builder) Grid(id string, rows, columns int, lines bool) *Builder {
	grid := NewGrid(id, rows, columns, lines)
	b.Add(grid)
	return b
}

// List creates a new list widget for displaying selectable items.
//
// Parameters:
//   - id: unique identifier for the list widget
//   - values: slice of strings to display as list items
func (b *Builder) List(id string, values []string) *Builder {
	list := NewList(id, values)
	b.Add(list)
	return b
}

// Static creates a new static widget with the specified id and text.
// The static widget is styled with theme styles for the text.
func (b *Builder) Static(id, text string) *Builder {
	static := NewStatic(id, text)
	b.Add(static)
	return b
}

// ---- Widget Manipulation --------------------------------------------------

// Background sets the background color for the current widget.
// A selector for the part/state can be specified.
func (b *Builder) Background(params ...string) *Builder {
	selector := ""
	if len(params) == 2 {
		selector = params[0]
	}
	style := b.current.Style(selector)
	if style.Fixed() {
		b.current.SetStyle(selector, style.WithBackground(params[len(params)-1]))
	} else {
		style.WithBackground(params[len(params)-1])
	}
	return b
}

// Border sets the border style for the current widget.
// The border parameter should be a valid border style string.
func (b *Builder) Border(params ...string) *Builder {
	selector := ""
	value := "none"
	if len(params) == 1 {
		value = params[0]
	} else if len(params) >= 2 {
		selector = params[0]
		value = strings.Join(params[1:], " ")
	}
	style := b.current.Style(selector)
	if style.Fixed() {
		b.current.SetStyle(selector, style.WithBorder(value))
	} else {
		style.WithBorder(value)
	}
	return b
}

// Bounds sets the absolute position and size of the current widget.
//
// Parameters:
//   - x, y: position coordinates relative to the parent container
//   - w, h: width and height in characters/cells
//
// The position is automatically offset by the parent container's content
// area. Be careful, because during UI creation, the position of must
// widgets is not set yet, because the layout was not calculated yet.
func (b *Builder) Bounds(x, y, w, h int) *Builder {
	var ox, oy int // x and y offset for content area

	if len(b.stack) > 0 {
		ox, oy, _, _ = b.stack.Peek().Content()
	}

	b.current.SetBounds(x+ox, y+oy, w, h)
	return b
}

// Cell specifies the grid cell coordinates and span for the next widget in a
// grid container.
//
// Parameters:
//   - x, y: starting grid cell coordinates (0-based)
//   - w, h: number of cells to span horizontally and vertically
//
// This method must be called before adding a widget to a grid container.
// The coordinates are used when the next widget is added to the grid.
func (b *Builder) Cell(x, y, w, h int) *Builder {
	b.x = x
	b.y = y
	b.w = w
	b.h = h
	return b
}

// Class sets a CSS-like class name that will be applied to subsequently
// created widgets. The class name is used in selector generation for styling
// purposes. For example, setting class to "primary" will generate selectors
// like "button.primary".
func (b *Builder) Class(class string) *Builder {
	b.class = class
	return b
}

// Columns sets the column sizes for the current grid container.
func (b *Builder) Columns(columns ...int) *Builder {
	b.current.(*Grid).Columns(columns...)
	return b
}

// Rows sets the row sizes for the current grid container.
func (b *Builder) Rows(rows ...int) *Builder {
	b.current.(*Grid).Rows(rows...)
	return b
}

// Font set the font options for the current widget.
// The font can be bold, italic, strikethrough or underline or any combination
// concatenated by commas.
func (b *Builder) Font(params ...string) *Builder {
	selector := ""
	if len(params) == 2 {
		selector = params[0]
	}
	style := b.current.Style(selector)
	if style.Fixed() {
		b.current.SetStyle(selector, style.WithFont(params[len(params)-1]))
	} else {
		style.WithFont(params[len(params)-1])
	}
	return b
}

// Foreground sets the foreground (text) color for the current widget.
// The color parameter should be a valid color name or hex code.
func (b *Builder) Foreground(params ...string) *Builder {
	selector := ""
	if len(params) == 2 {
		selector = params[0]
	}
	style := b.current.Style(selector)
	if style.Fixed() {
		b.current.SetStyle(selector, style.WithForeground(params[len(params)-1]))
	} else {
		style.WithForeground(params[len(params)-1])
	}
	return b
}

// Hint sets the preferred size hint for the current widget.
// Parameters:
//   - width: preferred width in characters
//   - height: preferred height in lines
//
// Size hints help the layout system determine optimal widget sizing.
func (b *Builder) Hint(width, height int) *Builder {
	b.current.SetHint(width, height)
	return b
}

// Margin sets the margin spacing around the current widget.
// Accepts 1-4 integer values following CSS margin conventions:
//   - 1 value: all sides
//   - 2 values: vertical, horizontal
//   - 3 values: top, horizontal, bottom
//   - 4 values: top, right, bottom, left
func (b *Builder) Margin(a ...int) *Builder {
	style := b.current.Style()
	if style.Fixed() {
		b.current.SetStyle("", style.WithMargin(a...))
	} else {
		style.WithMargin(a...)
	}
	return b
}

// Padding sets the internal padding for the current widget.
// Accepts 1-4 integer values following CSS padding conventions:
//   - 1 value: all sides
//   - 2 values: vertical, horizontal
//   - 3 values: top, horizontal, bottom
//   - 4 values: top, right, bottom, left
func (b *Builder) Padding(a ...int) *Builder {
	style := b.current.Style()
	if style.Fixed() {
		b.current.SetStyle("", style.WithPadding(a...))
	} else {
		style.WithPadding(a...)
	}
	return b
}

// Position sets the absolute position of the current widget.
//
// Parameters:
//   - x: horizontal position in characters
//   - y: vertical position in lines
//
// This is typically used for widgets that are not in containers.
func (b *Builder) Position(x, y int) *Builder {
	_, _, w, h := b.current.Bounds()
	b.current.SetBounds(x, y, w, h)
	return b
}

// Size sets the absolute size of the current widget.
//
// Parameters:
//   - width: width in characters
//   - height: height in lines
//
// This overrides any size hints or automatic sizing.
func (b *Builder) Size(width, height int) *Builder {
	x, y, _, _ := b.current.Bounds()
	b.current.SetBounds(x, y, width, height)
	return b
}
