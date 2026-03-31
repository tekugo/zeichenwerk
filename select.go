package zeichenwerk

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

// option represents a single option in the select widget.
type option struct {
	value string
	text  string
}

// Select is a widget that allows selecting from a list of options.
type Select struct {
	Component
	index   int
	options []option
}

// NewSelect creates a new Select widget with the given ID and options.
// The args parameter should contain alternating value and text strings
// (value1, text1, value2, text2, ...).
func NewSelect(id, class string, args ...string) *Select {
	s := &Select{
		Component: Component{id: id, class: class},
		options:   make([]option, 0, len(args)/2),
	}
	s.SetFlag(FlagFocusable, true)
	for i := 0; i+1 < len(args); i += 2 {
		s.options = append(s.options, option{value: args[i], text: args[i+1]})
	}
	OnKey(s, s.handleKey)
	return s
}

// Apply applies a theme's styles to the component.
func (s *Select) Apply(theme *Theme) {
	theme.Apply(s, s.Selector("select"), "disabled", "focused", "hovered")
}

// Hint returns the preferred size of the Select widget as width and height in characters.
func (s *Select) Hint() (int, int) {
	if s.hwidth != 0 && s.hheight != 0 {
		return s.hwidth, s.hheight
	}

	mw := 0
	for _, option := range s.options {
		mw = max(mw, utf8.RuneCountInString(option.text))
	}

	// TODO: Get real dropdown width? Renderer not available here
	return mw + 2, 1
}

// Render draws the Select widget using the given renderer.
func (s *Select) Render(r *Renderer) {
	if s.Flag(FlagHidden) {
		return
	}

	// Render the component style
	s.Component.Render(r)

	// Render the content
	dropdown := r.theme.String("select.dropdown")
	dw := utf8.RuneCountInString(dropdown)
	cx, cy, cw, _ := s.Content()
	r.Text(cx, cy, s.options[s.index].text, 0)
	r.Text(cx+cw-dw, cy, dropdown, dw)
}

// Select selects the specified value.
func (s *Select) Select(value string) {
	s.index = 0
	for i, option := range s.options {
		if option.value == value {
			s.index = i
			return
		}
	}
}

// Text returns the display text of the currently selected option.
func (s *Select) Text() string {
	return s.options[s.index].text
}

// Value returns the value of the currently selected option.
func (s *Select) Value() string {
	return s.options[s.index].value
}

// handleKey processes key events for the Select widget.
func (s *Select) handleKey(_ Widget, evt *tcell.EventKey) bool {
	switch evt.Key() {
	case tcell.KeyEnter:
		s.popup()
		return true
	}
	return false
}

// popup shows the dropdown list of options.
func (s *Select) popup() {
	s.Log(s, Debug, "Show list popup")
	ui := FindUI(s)
	if ui == nil {
		return
	}
	items := make([]string, len(s.options))
	for i, option := range s.options {
		items[i] = option.text
	}
	popup := ui.NewBuilder().
		Class("popup").
		Box("select-popup", "").
		List("select-list", items...).
		End().
		Class("").
		Container()
	list, ok := Find(popup, "select-list").(*List)
	if !ok {
		s.Log(s, Error, "Cannot create popup")
		return
	}
	list.Select(s.index)
	list.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		s.index = list.Selected()
		s.Dispatch(s, EvtChange, s.Value())
		ui.Close()
		ui.Focus(s)
		return true
	})
	OnKey(list, func(_ Widget, evt *tcell.EventKey) bool {
		switch evt.Key() {
		case tcell.KeyEsc:
			ui.Close()
			ui.Focus(s)
			return true
		}
		return false
	})
	const popupHeight = 10
	_, _, _, uiHeight := ui.Bounds()
	py := s.y + s.height
	if py+popupHeight > uiHeight {
		py = s.y - popupHeight
	}
	ui.Popup(s.x, py, s.width, popupHeight, popup)
}
