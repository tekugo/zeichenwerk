package zeichenwerk

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Builder provides a fluent interface for constructing TUI components.
// It maintains a stack of containers and applies styling through themes.
// The builder pattern allows for method chaining to create complex UI
// layouts in an easy and descriptive way.
type Builder struct {
	theme      Theme            // Current theme for styling widgets
	stack      Stack[Container] // Stack of container widgets for nesting
	current    Widget           // Currently active widget being configured
	tabs       *Tabs            // Last tabs widget to add new tabs
	class      string           // CSS-like class name for styling
	x, y, w, h int              // Grid cell coordinates and dimensions
}

// NewBuilder creates a new Builder instance with the specified theme.
// Returns a pointer to the newly created Builder.
func NewBuilder(theme Theme) *Builder {
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
//
// Returns the constructed UI ready for rendering and interaction.
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
	return Find(b.stack[0], id, false)
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
		case *Dialog:
			top.Add(widget)
		case *Flex:
			top.Add(widget)
		case *Form:
			top.Add(widget)
		case *Grid:
			top.Add(b.x, b.y, b.w, b.h, widget)
		case *Scroller:
			top.Add(widget)
		case *Switcher:
			top.Add(widget.ID(), widget)
		case *ThemeSwitch:
			top.Add(widget)
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

// Apply applies the styles of a theme to a widget based on its type and ID.
// The theme styles are applied automatically when the builder is used,
// but if widgets are created outside of the builder, they can be styled
// using this method.
func (b *Builder) Apply(widget Widget) {
	switch widget := widget.(type) {
	case *Box:
		b.theme.Apply(widget, b.selector("box", widget.ID()), "title")
	case *Button:
		b.theme.Apply(widget, b.selector("button", widget.ID()), "disabled", "focus", "hover", "pressed")
	case *Checkbox:
		b.theme.Apply(widget, b.selector("checkbox", widget.ID()), "disabled", "focus", "hover")
	case *Custom:
		b.theme.Apply(widget, b.selector("custom", widget.ID()))
	case *Dialog:
		b.theme.Apply(widget, b.selector("dialog", widget.ID()), "focus")
		b.theme.Apply(widget, b.selector("dialog/title", widget.ID()), "focus")
	case *Digits:
		b.theme.Apply(widget, b.selector("digits", widget.ID()))
	case *Editor:
		b.theme.Apply(widget, b.selector("editor", widget.ID()))
		b.theme.Apply(widget, b.selector("editor/current-line", widget.ID()))
		b.theme.Apply(widget, b.selector("editor/current-line-number", widget.ID()))
		b.theme.Apply(widget, b.selector("editor/line-numbers", widget.ID()))
		b.theme.Apply(widget, b.selector("editor/separator", widget.ID()))
	case *Flex:
		b.theme.Apply(widget, b.selector("flex", widget.ID()))
		b.theme.Apply(widget, b.selector("flex/shadow", widget.ID()))
	case *Form:
		b.theme.Apply(widget, b.selector("form", widget.ID()))
	case *FormGroup:
		b.theme.Apply(widget, b.selector("form-group", widget.ID()))
	case *Grid:
		b.theme.Apply(widget, b.selector("grid", widget.ID()))
	case *Hidden:
		b.theme.Apply(widget, b.selector("hidden", widget.ID()))
	case *Input:
		b.theme.Apply(widget, b.selector("input", widget.ID()), "focus")
	case *Label:
		b.theme.Apply(widget, b.selector("label", widget.ID()))
	case *List:
		b.theme.Apply(widget, b.selector("list", widget.ID()), "disabled", "focus")
		b.theme.Apply(widget, b.selector("list/highlight", widget.ID()), "focus")
	case *ProgressBar:
		b.theme.Apply(widget, b.selector("progress-bar", widget.ID()))
		b.theme.Apply(widget, b.selector("progress-bar/bar", widget.ID()))
	case *Scroller:
		b.theme.Apply(widget, b.selector("scroller", widget.ID()), "focus")
	case *Separator:
		b.theme.Apply(widget, b.selector("separator", widget.ID()))
	case *Spinner:
		b.theme.Apply(widget, b.selector("spinner", widget.ID()))
	case *Switcher:
		b.theme.Apply(widget, b.selector("switcher", widget.ID()))
	case *Table:
		b.theme.Apply(widget, b.selector("table", widget.ID()), "focus")
		b.theme.Apply(widget, b.selector("table/grid", widget.ID()), "focus")
		b.theme.Apply(widget, b.selector("table/header", widget.ID()), "focus")
		b.theme.Apply(widget, b.selector("table/highlight", widget.ID()), "focus")
	case *Tabs:
		b.theme.Apply(widget, b.selector("tabs", widget.ID()), "focus")
		b.theme.Apply(widget, b.selector("tabs/line", widget.ID()), "focus")
		b.theme.Apply(widget, b.selector("tabs/highlight", widget.ID()), "focus")
		b.theme.Apply(widget, b.selector("tabs/highlight-line", widget.ID()), "focus")
	case *Text:
		b.theme.Apply(widget, b.selector("text", widget.ID()))
	case *ThemeSwitch:
		b.theme.Apply(widget, b.selector("theme-switch", widget.ID()))
	default:
		panic(fmt.Errorf("no style for widget type %T", widget))
	}
}

// Box creates a new box widget with the specified id and display title.
// The box is automatically styled with theme styles for the border and
// the title.
func (b *Builder) Box(id, title string) *Builder {
	box := NewBox(id, title)
	b.Add(box)
	return b
}

// Button creates a new button widget with the specified id and display text.
// The button is automatically styled with theme styles for various states
// (disabled, focus, hover, pressed). The button's size hint is set to the
// text length.
func (b *Builder) Button(id string, text string) *Builder {
	button := NewButton(id, text)
	b.Add(button)
	button.SetHint(len(text), 1)
	return b
}

// Checkbox creates a new checkbox widget with the specified id, label text,
// and initial state.
// The checkbox is automatically styled with theme styles for various states
// (disabled, focus, hover). The checkbox's size hint is set based on the
// label length.
func (b *Builder) Checkbox(id string, text string, checked bool) *Builder {
	checkbox := NewCheckbox(id, text, checked)
	b.Add(checkbox)
	// Size hint: 4 characters for "[x] " plus text length
	checkbox.SetHint(4+len([]rune(text)), 1)
	return b
}

// Dialog creates a new dialog container with the given id and title.
// A dialog is not normally created as port of the initial UI creation, but
// afterwards. Use UI.Builder() to get a builder for the dialog at runtime.
func (b *Builder) Dialog(id, title string) *Builder {
	dialog := NewDialog(id, title)
	b.Add(dialog)
	return b
}

// Digits creates a big digit display label.
func (b *Builder) Digits(id, text string) *Builder {
	digits := NewDigits(id, text)
	b.Add(digits)
	digits.SetHint(len(text)*4, 3)
	return b
}

// Editor creates a multi-line text editor widget with the specified id.
// The editor will be empty and should be configured separately.
func (b *Builder) Editor(id string) *Builder {
	editor := NewEditor(id)
	b.Add(editor)
	return b
}

// Flex creates a new flex container widget for arranging child widgets.
//
// Parameters:
//   - id: unique identifier for the flex container
//   - orientation: layout direction ("horizontal" or "vertical")
//   - alignment: how children are aligned ("start", "center", "end", "stretch")
//   - spacing: cells between child widgets (columns or rows)
func (b *Builder) Flex(id string, orientation, alignment string, spacing int) *Builder {
	flex := NewFlex(id, orientation, alignment, spacing)
	b.Add(flex)
	return b
}

// Form creates a new form widget that manages data binding between Go structs
// and form controls. The form automatically generates appropriate controls
// based on struct field types and tags.
//
// Parameters:
//   - id: Unique identifier for the form widget
//   - title: Display title for the form container
//   - value: Pointer to struct containing form data - must be a pointer for updates to work
//
// The struct should use tags to control form generation:
//   - `label:"Display Name"` - Custom label (default: field name)
//   - `width:"20"` - Control width in characters
//   - `control:"input|checkbox|password"` - Control type (auto-detected if omitted)
//   - `group:"groupname"` - Group fields together
//   - `line:"1"` - Line number within group
//   - `readOnly:"true"` - Make field read-only
//
// Example:
//
//	type User struct {
//	  Name  string `label:"Full Name" width:"30"`
//	  Email string `label:"Email Address" width:"40"`
//	  Admin bool   `label:"Administrator"`
//	}
//
//	user := &User{}
//	builder.Form("user-form", "User Registration", user)
func (b *Builder) Form(id, title string, value any) *Builder {
	form := NewForm(id, title, value)
	b.Add(form)
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

// Group creates a form group that organizes related form fields with
// consistent labeling and layout. This method automatically generates form
// controls based on struct fields that match the specified group.
//
// Parameters:
//   - id: Unique identifier for the form group
//   - title: Display title for the group (shown in border if styled)
//   - name: Group name to match against struct field `group` tags (empty string matches untagged fields)
//   - placement: Label placement ("horizontal" or "vertical")
//   - spacing: Vertical spacing between field lines (typically 1)
//
// The method searches for the nearest parent Form widget and automatically
// generates controls for struct fields that have a matching `group` tag value.
//
// Struct tag example:
//
//	type User struct {
//	  Name  string `group:"basic" label:"Full Name" width:"30"`
//	  Email string `group:"basic" label:"Email" width:"40"`
//	  Phone string `group:"contact" label:"Phone" width:"20"`
//	}
//
// Example:
//
//	builder.Form("user-form", "User", &user).
//	  Group("basic-group", "Basic Information", "basic", "horizontal", 1).
//	  End().
//	  Group("contact-group", "Contact Details", "contact", "vertical", 1).
//	  End()
func (b *Builder) Group(id, title, name, placement string, spacing int) *Builder {
	group := NewFormGroup(id, title, placement)
	group.Spacing = 1
	b.Add(group)

	// Find the nearest form
	var current Widget = b.stack.Peek()
	for current != nil {
		if form, ok := current.(*Form); ok {
			b.buildGroup(form, group, name, form.Data)
			break
		}
		current = current.Parent()
	}

	return b
}

// buildGroup processes struct fields and creates form controls for fields
// matching the specified group. This is an internal helper method called
// by Group() to generate form controls automatically.
//
// Parameters:
//   - form: The parent form widget for data binding
//   - group: The form group to add controls to
//   - name: Group name to filter fields (empty matches untagged fields)
//   - data: Pointer to struct containing the data
//
// Supported struct tags:
//   - `group:"name"` - Assigns field to named group
//   - `label:"text"` - Custom label (default: field name, "-" to hide)
//   - `control:"type"` - Control type (input, checkbox, password)
//   - `width:"20"` - Control width in characters (default: 10)
//   - `line:"1"` - Specific line number (default: auto-increment)
//   - `readOnly:"true"` - Make control read-only
func (b *Builder) buildGroup(form *Form, group *FormGroup, name string, data any) {
	line := 0
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		panic("expecting pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := range v.NumField() {
		sf := t.Field(i)
		fv := v.Field(i)
		g := sf.Tag.Get("group")
		if name != "" && name != g {
			continue
		}
		label := sf.Tag.Get("label")
		if label == "-" {
			continue
		} else if label == "" {
			label = sf.Name
		}
		control := sf.Tag.Get("control")
		width, err := strconv.Atoi(sf.Tag.Get("width"))
		if err != nil {
			width = 10
		}
		l, err := strconv.Atoi(sf.Tag.Get("line"))
		if err == nil {
			line = l
		}

		widget := b.buildFormControl(control, sf.Name, fv)
		widget.SetHint(width, 1)
		widget.SetParent(b.stack.Peek())
		widget.On("change", form.Update(fv))
		group.Add(line, label, widget)

		line++
	}
}

// buildFormControl creates an appropriate form control widget based on the
// field type and control tag. This is an internal helper method that handles
// the automatic selection and configuration of controls.
//
// Parameters:
//   - control: Explicit control type from struct tag ("input", "checkbox", "password")
//   - id: Widget ID (typically the struct field name)
//   - v: reflect.Value of the struct field
//
// Returns:
//   - Widget: Configured form control widget
//
// Control type selection:
//   - If control tag is specified, uses that type
//   - If no control tag: bool fields become checkboxes, others become inputs
//   - Supported types: "input", "checkbox", "password"
func (b *Builder) buildFormControl(control, id string, v reflect.Value) Widget {
	if control == "" {
		switch v.Kind() {
		case reflect.Bool:
			control = "checkbox"
		default:
			control = "input"
		}
	}

	switch control {
	case "checkbox":
		checkbox := NewCheckbox(id, id, v.Bool())
		b.Apply(checkbox)
		checkbox.Checked = v.Bool()
		return checkbox
	default:
		input := NewInput(id)
		b.Apply(input)
		input.Text = v.String()
		return input
	}
}

// Input creates a new text input widget for user text entry.
//
// Parameters:
//   - id: unique identifier for the input widget
//   - label: label text (currently unused in implementation)
//   - width: preferred width in characters
//
// The input is styled with focus states and sized according to the width
// parameter.
func (b *Builder) Input(id string, label string, width int) *Builder {
	input := NewInput(id)
	b.Add(input)
	input.SetHint(width, 1)
	return b
}

// Label creates a new text label widget for displaying static text.
//
// Parameters:
//   - id: unique identifier for the label widget
//   - text: text content to display
//   - width: preferred width in characters (0 = auto-size to text length)
//
// If width is 0, the label is sized to fit the text content exactly.
func (b *Builder) Label(id string, text string, width int) *Builder {
	label := NewLabel(id, text)
	b.Add(label)
	if width == 0 {
		label.SetHint(len([]rune(text)), 1)
	} else {
		label.SetHint(width, 1)
	}
	return b
}

// List creates a new list widget for displaying selectable items.
//
// Parameters:
//   - id: unique identifier for the list widget
//   - values: slice of strings to display as list items
//
// The list is styled with states for disabled, focus, and highlight.
// Users can navigate and select items using keyboard input.
func (b *Builder) List(id string, values []string) *Builder {
	list := NewList(id, values)
	b.Add(list)
	return b
}

// ProgressBar creates a new progress bar widget for showing completion status.
//
// Parameters:
//   - id: unique identifier for the progress bar widget
//   - value: current progress value
//   - min: minimum value of the range
//   - max: maximum value of the range
//
// The progress bar visually represents the value as a percentage of the
// min-max range.
func (b *Builder) ProgressBar(id string, value, min, max int) *Builder {
	bar := NewProgressBar(id)
	b.Add(bar)
	bar.SetRange(min, max)
	bar.SetValue(value)
	bar.SetHint(20, 1)
	return b
}

// Scroller creates a new scroll pane for displaying the child in a viewport.
//
// Parameters:
//   - id: unique identifier for the scroller widget
//   - title: optional scroll pane title rendered in its border
//
// The Scroller allows to display children larger than the available content
// area and if it is larger, scroll bars are shown.
func (b *Builder) Scroller(id, title string) *Builder {
	scroller := NewScroller(id, title)
	b.Add(scroller)
	return b
}

// Separator creates a new separator for displaying a horizontal or vertical line.
//
// Parameters:
//   - id: unique identifier for the separator widget
//   - border: border style
//   - width: separator width, should be 1 for vertical separators
//   - height: separator height, should be 1 for horizontal separators
//
// The separator is just a visual element with no interaction possibility.
func (b *Builder) Separator(id, border string, width, height int) *Builder {
	separator := NewSeparator(id, border)
	b.Add(separator)
	separator.SetHint(width, height)
	return b
}

// Spacer creates a hidden widget, which just takes up space in the container.
// It takes no parameters and no ID, because it does nothing and should do
// nothing.
func (b *Builder) Spacer() *Builder {
	spacer := NewHidden("")
	b.Add(spacer)
	spacer.SetHint(-1, -1)
	return b
}

// Spinner creates a new spinner widget.
func (b *Builder) Spinner(id string, runes []rune) *Builder {
	spinner := NewSpinner(id, runes)
	b.Add(spinner)
	return b
}

// Switcher creates a content switcher container.
// The switcher can be automatically connected to the last tabs component for
// tab switching. If so, every pane should be accompanied by a Tab() call
// with the tab title.
func (b *Builder) Switcher(id string, connect bool) *Builder {
	switcher := NewSwitcher(id)
	b.Add(switcher)
	if connect && b.tabs != nil {
		b.tabs.On("activate", func(widget Widget, _ string, params ...any) bool {
			if len(params) > 0 {
				if selected, ok := params[0].(int); ok {
					switcher.Select(selected)
				}
			}
			return true
		})
	}
	return b
}

// Table creates a table widget with the passed data provider.
func (b *Builder) Table(id string, provider TableProvider) *Builder {
	table := NewTable(id, provider)
	b.Add(table)
	return b
}

// Tab adds a new tab for a switcher, if a Tabs was added before.
func (b *Builder) Tab(name string) *Builder {
	if b.tabs != nil {
		b.tabs.Add(name)
	}
	return b
}

// Tabs creates a tabl selection widget.
//
// The last Tabs widget is stored by the builder and new tabs can be added
// by subsequent tab calls, which can be placed inside the switcher.
func (b *Builder) Tabs(id string, tabs ...string) *Builder {
	t := NewTabs(id)
	for _, tab := range tabs {
		t.Add(tab)
	}
	b.Add(t)
	b.tabs = t
	return b
}

// Text creates a new text widget for displaying multi-line text content.
//
// Parameters:
//   - id: unique identifier for the text widget
//   - content: slice of strings, each representing a line of text
//   - follow: whether to automatically scroll to show new content (like tail -f)
//   - max: maximum number of lines to retain (0 = unlimited)
//
// The text widget supports scrolling and can be used for logs or large text display.
func (b *Builder) Text(id string, content []string, follow bool, max int) *Builder {
	text := NewText(id, content, follow, max)
	b.Add(text)
	return b
}

// ThemeSwitch creates a new theme switcher widget, for temporarily changing
// the theme.
//
// Parameters:
//   - id: unique identifier for the theme switch
//   - theme: new theme
func (b *Builder) ThemeSwitch(id string, theme Theme) *Builder {
	ts := NewThemeSwitch(id, theme)
	b.Add(ts)
	return b
}

// ---- Widget Manipulation --------------------------------------------------

// Background sets the background color for the current widget.
// A selector for the part/state can be specified.
//
// Parameters:
//   - selector: optional CSS-like selector string (e.g., ":focus", ":hover")
//   - color: color name or hex code for the background
//
// This allows different background colors for different widget states.
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

// End finalizes the current container and pops it from the stack.
// This should be called after adding all children to a container.
// The current widget is refreshed and the container becomes the new current
// widget. If only one container remains on the stack, it stays as the root
// container.
func (b *Builder) End() *Builder {
	if len(b.stack) > 1 {
		b.current = b.stack.Pop()
	}
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
	b.current.SetPosition(x, y)
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
	b.current.SetSize(width, height)
	return b
}
