package widgets

import (
	"slices"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// ==== AI ===================================================================

// Deck is a scrollable list widget where every item occupies a fixed number of
// rows. Rendering is delegated to a caller-supplied ItemRender function so
// each slot can display rich, multi-line content without per-item widget
// allocations.
type Deck struct {
	Component
	render     ItemRender // Render function for each slot
	items      []any      // Data items, one per slot
	disabled   []int      // Indices of non-selectable items
	itemHeight int        // Fixed row count per item slot (>= 1)
	index      int        // Currently highlighted item index (-1 if empty)
	offset     int        // Index of first visible item
	scrollbar  bool       // Whether to draw a vertical scrollbar
}

// NewDeck creates a new Deck widget.
//
// Parameters:
//   - id: Unique identifier for the widget
//   - class: CSS-like class name for styling
//   - render: Function called to draw each item slot
//   - itemHeight: Fixed row count per slot; panics if < 1
func NewDeck(id, class string, render ItemRender, itemHeight int) *Deck {
	if itemHeight < 1 {
		panic("deck: itemHeight must be >= 1")
	}
	d := &Deck{
		Component:  Component{id: id, class: class},
		render:     render,
		items:      nil,
		disabled:   nil,
		itemHeight: itemHeight,
		index:      -1,
		offset:     0,
		scrollbar:  true,
	}
	d.SetFlag(FlagFocusable, true)
	OnKey(d, d.handleKey)
	OnMouse(d, d.handleMouse)
	return d
}

// ---- Widget Methods -------------------------------------------------------

// Apply applies the deck's theme styles.
func (d *Deck) Apply(theme *Theme) {
	theme.Apply(d, d.Selector("deck"), "disabled", "focused", "hovered")
}

// Hint returns the preferred size. If a hint override has been set via
// SetHint (e.g. Hint(0, -1) for flexible height), that value is returned.
// Otherwise the natural height is len(items)*itemHeight.
func (d *Deck) Hint() (int, int) {
	if d.hwidth != 0 || d.hheight != 0 {
		return d.hwidth, d.hheight
	}
	return 0, len(d.items) * d.itemHeight
}

// ---- Data -----------------------------------------------------------------

// Get returns the current items slice.
func (d *Deck) Get() []any {
	return d.items
}

// Set replaces all items, resets index to 0 (or -1 if empty) and offset
// to 0, then redraws.
func (d *Deck) Set(items []any) {
	d.items = items
	if len(items) == 0 {
		d.index = -1
		d.offset = 0
	} else if d.index < 0 || d.index >= len(items) {
		d.index = 0
		d.offset = 0
	}
	Redraw(d)
}

// SetDisabled replaces the list of non-selectable item indices.
func (d *Deck) SetDisabled(indices []int) {
	d.disabled = indices
}

// ---- Selection ------------------------------------------------------------

// Select highlights the item at index, adjusts the scroll offset, and
// dispatches EvtSelect.
func (d *Deck) Select(index int) {
	if index < 0 || index >= len(d.items) {
		return
	}
	d.index = index
	d.adjust()
	d.Dispatch(d, EvtSelect, d.index)
	Redraw(d)
}

// Selected returns the currently highlighted item index (-1 if none).
func (d *Deck) Selected() int {
	return d.index
}

// ---- Navigation -----------------------------------------------------------

// First highlights the first enabled item.
func (d *Deck) First() {
	d.index = -1
	for i := range d.items {
		if !slices.Contains(d.disabled, i) {
			d.index = i
			d.adjust()
			d.Dispatch(d, EvtSelect, d.index)
			break
		}
	}
	Redraw(d)
}

// Last highlights the last enabled item.
func (d *Deck) Last() {
	d.index = -1
	for i := len(d.items) - 1; i >= 0; i-- {
		if !slices.Contains(d.disabled, i) {
			d.index = i
			d.adjust()
			d.Dispatch(d, EvtSelect, d.index)
			break
		}
	}
	Redraw(d)
}

// Move advances the highlight by count steps (positive = down, negative = up),
// skipping disabled items and clamping at the list boundaries.
func (d *Deck) Move(count int) {
	if len(d.items) == 0 || count == 0 {
		return
	}
	steps := count
	if steps < 0 {
		steps = -steps
	}
	newIndex := d.index
	for i := 0; i < steps; i++ {
		if count > 0 {
			newIndex = d.skip(newIndex+1, 1)
		} else {
			newIndex = d.skip(newIndex-1, -1)
		}
	}
	if newIndex != d.index {
		d.index = newIndex
		d.adjust()
		d.Dispatch(d, EvtSelect, d.index)
		Redraw(d)
	}
}

// PageDown moves the highlight down by the number of fully visible slots.
func (d *Deck) PageDown() {
	_, _, _, ch := d.Content()
	slots := ch / d.itemHeight
	if slots < 1 {
		slots = 1
	}
	d.Move(slots)
}

// PageUp moves the highlight up by the number of fully visible slots.
func (d *Deck) PageUp() {
	_, _, _, ch := d.Content()
	slots := ch / d.itemHeight
	if slots < 1 {
		slots = 1
	}
	d.Move(-slots)
}

// ---- Internal helpers -----------------------------------------------------

// skip advances from index in direction (±1), skipping disabled items. When
// it overshoots the bounds it wraps to the nearest enabled item at the other
// boundary.
func (d *Deck) skip(index, direction int) int {
	next := index
	for next >= 0 && next < len(d.items) && slices.Contains(d.disabled, next) {
		next += direction
	}
	if next < 0 {
		next = -1
		for i := range d.items {
			if !slices.Contains(d.disabled, i) {
				next = i
				break
			}
		}
	} else if next >= len(d.items) {
		next = -1
		for i := len(d.items) - 1; i >= 0; i-- {
			if !slices.Contains(d.disabled, i) {
				next = i
				break
			}
		}
	}
	return next
}

// adjust ensures the highlighted item is visible within the current viewport.
func (d *Deck) adjust() {
	_, _, _, ch := d.Content()
	if ch <= 0 || d.itemHeight <= 0 {
		return
	}
	slots := ch / d.itemHeight

	if d.index < d.offset {
		d.offset = d.index
	} else if d.index >= d.offset+slots {
		d.offset = d.index - slots + 1
	}
	if d.offset < 0 {
		d.offset = 0
	}
	maxOffset := max(len(d.items)-slots, 0)
	if d.offset > maxOffset {
		d.offset = maxOffset
	}
}

// ---- Event Handlers -------------------------------------------------------

func (d *Deck) handleKey(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyUp:
		d.Move(-1)
		return true
	case tcell.KeyDown:
		d.Move(1)
		return true
	case tcell.KeyHome:
		d.First()
		return true
	case tcell.KeyEnd:
		d.Last()
		return true
	case tcell.KeyPgUp:
		d.PageUp()
		return true
	case tcell.KeyPgDn:
		d.PageDown()
		return true
	case tcell.KeyEnter:
		if d.index >= 0 {
			d.Dispatch(d, EvtActivate, d.index)
		}
		return true
	}
	return false
}

func (d *Deck) handleMouse(event *tcell.EventMouse) bool {
	if event.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := event.Position()
	cx, cy, cw, ch := d.Content()
	if mx < cx || mx >= cx+cw || my < cy || my >= cy+ch {
		return false
	}

	clickedItem := (my-cy)/d.itemHeight + d.offset
	if clickedItem < 0 || clickedItem >= len(d.items) {
		return false
	}
	if slices.Contains(d.disabled, clickedItem) {
		return false
	}

	if clickedItem == d.index {
		d.Dispatch(d, EvtActivate, clickedItem)
	} else {
		d.Select(clickedItem)
	}
	return true
}

// ---- Rendering ------------------------------------------------------------

func (d *Deck) Render(r *Renderer) {
	d.Component.Render(r)

	cx, cy, cw, ch := d.Content()
	if cw <= 0 || ch <= 0 || d.render == nil {
		return
	}

	// Reserve rightmost column for scrollbar when needed.
	tw := cw
	if d.scrollbar && len(d.items)*d.itemHeight > ch {
		tw = cw - 1
	}

	slots := ch / d.itemHeight

	for s := 0; s < slots; s++ {
		itemIndex := d.offset + s
		if itemIndex >= len(d.items) {
			break
		}
		slotY := cy + s*d.itemHeight
		d.render(r, cx, slotY, tw, d.itemHeight, itemIndex, d.items[itemIndex], itemIndex == d.index, d.Flag(FlagFocused))
	}

	if d.scrollbar && len(d.items)*d.itemHeight > ch {
		r.ScrollbarV(cx+tw, cy, ch, d.offset*d.itemHeight, len(d.items)*d.itemHeight)
	}
}
