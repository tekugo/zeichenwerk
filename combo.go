package zeichenwerk

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v3"
)

// Combo is a traditional combo box: a single-line display of the current value
// that opens a popup when focused. The popup contains a free-text [Typeahead]
// input and a filtered suggestion [List]. The user can type anything or pick an
// item from the list; either way the confirmed value is what gets submitted.
//
// Typical use: search fields with a history or a set of common candidates.
//
// Events:
//   - [EvtChange]   – string: current input text while the popup is open
//   - [EvtActivate] – string: confirmed value when Enter is pressed
type Combo struct {
	Component
	value string
	items []string
}

// NewCombo creates a new Combo widget with the given suggestion items.
func NewCombo(id, class string, items []string) *Combo {
	c := &Combo{
		Component: Component{id: id, class: class},
		items:     items,
	}
	c.SetFlag(FlagFocusable, true)
	c.On(EvtFocus, func(_ Widget, _ Event, _ ...any) bool {
		c.popup()
		return false
	})
	OnKey(c, func(evt *tcell.EventKey) bool {
		if evt.Key() == tcell.KeyEnter {
			c.popup()
			return true
		}
		return false
	})
	return c
}

// ---- Widget Methods ------------------------------------------------------

// Apply applies theme styles to the Combo.
func (c *Combo) Apply(theme *Theme) {
	theme.Apply(c, c.Selector("combo"), "disabled", "focused", "hovered")
}

// Hint returns the preferred size: width = widest item + 2, height = 1.
func (c *Combo) Hint() (int, int) {
	if c.hwidth != 0 || c.hheight != 0 {
		return c.hwidth, c.hheight
	}
	maxW := 0
	for _, item := range c.items {
		if w := utf8.RuneCountInString(item); w > maxW {
			maxW = w
		}
	}
	return maxW + 2, 1
}

// Render draws the current value with a ▼ indicator at the right edge.
func (c *Combo) Render(r *Renderer) {
	if c.Flag(FlagHidden) {
		return
	}
	c.Component.Render(r)
	cx, cy, cw, _ := c.Content()
	if cw < 1 {
		return
	}
	r.Text(cx, cy, c.value, cw-1)
	r.Text(cx+cw-1, cy, "▼", 1)
}

// ---- Combo Methods --------------------------------------------------------

// SetItems replaces the suggestion list.
func (c *Combo) SetItems(items []string) {
	c.items = items
}

// ---- Getter and Setter ----------------------------------------------------

// Get returns the last confirmed value.
func (c *Combo) Get() string {
	return c.value
}

// Set sets the combo box value.
func (c *Combo) Set(value string) {
	c.value = value
}

// ---- Internal Methods -----------------------------------------------------

// popup opens a floating input+list panel below the Combo.
func (c *Combo) popup() {
	ui := FindUI(c)
	if ui == nil {
		return
	}

	listH := len(c.items)
	if listH > 8 {
		listH = 8
	}
	popupH := listH + 1 // +1 for the input row
	if popupH < 3 {
		popupH = 3
	}

	b := ui.NewBuilder().Class("popup")
	popupContainer := b.
		Box("combo-popup", "").
		Flex("combo-popup-body", "stretch", 0).Flag(FlagVertical).
		Typeahead("combo-popup-input", c.value).Hint(0, 1).
		List("combo-popup-list", c.items...).Hint(0, -1).
		End().
		End().
		Class("").
		Container()

	input := Find(popupContainer, "combo-popup-input").(*Typeahead)
	list := Find(popupContainer, "combo-popup-list").(*List)

	// Wire ghost-text from list prefix matching.
	input.SetSuggest(list.Suggest)

	// Filter list and propagate EvtChange on the Combo as the user types.
	OnChange(input, func(value string) bool {
		list.Filter(value)
		c.Dispatch(c, EvtChange, value)
		return false
	})

	// Navigation and confirmation keys.
	OnKey(input, func(evt *tcell.EventKey) bool {
		switch evt.Key() {
		case tcell.KeyDown:
			list.Move(+1)
			comboPopupCopy(input, list)
			return true
		case tcell.KeyUp:
			list.Move(-1)
			comboPopupCopy(input, list)
			return true
		case tcell.KeyPgDn:
			list.PageDown()
			comboPopupCopy(input, list)
			return true
		case tcell.KeyPgUp:
			list.PageUp()
			comboPopupCopy(input, list)
			return true
		case tcell.KeyEnter:
			c.value = input.Get()
			c.Dispatch(c, EvtActivate, c.value)
			ui.Close()
			return true
		case tcell.KeyEsc:
			ui.Close()
			return true
		}
		return false
	})

	_, _, _, uiHeight := ui.Bounds()
	py := c.y + c.height
	if py+popupH > uiHeight {
		py = c.y - popupH
	}
	ui.Popup(c.x, py, c.width, popupH, popupContainer)
}

// comboPopupCopy copies the highlighted list item into the input field without
// triggering EvtChange (which would reset the list filter).
func comboPopupCopy(input *Typeahead, list *List) {
	items := list.Items()
	if len(items) == 0 {
		return
	}
	idx := list.Selected()
	if idx < 0 || idx >= len(items) {
		return
	}
	input.Set(items[idx])
	input.End()
}
