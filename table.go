package zeichenwerk

import "github.com/gdamore/tcell/v2"

type Table struct {
	BaseWidget
	provider         TableProvider
	row, column      int // highlight position
	offsetX, offsetY int
}

func NewTable(id string, provider TableProvider) *Table {
	return &Table{
		BaseWidget: BaseWidget{id: id, focusable: true},
		provider:   provider,
	}
}

func (t *Table) Hint() (int, int) {
	w := 0
	h := t.provider.Length()
	for i, column := range t.provider.Columns() {
		if i > 1 {
			w++
		}
		w += column.Width
	}
	return w, h
}

func (t *Table) Handle(evt tcell.Event) bool {
	event, ok := evt.(*tcell.EventKey)
	if !ok {
		return false
	}

	switch event.Key() {
	case tcell.KeyDown:
		if t.row < t.provider.Length()-1 {
			t.row++
			t.adjust()
		}
		return true
	case tcell.KeyUp:
		if t.row > 0 {
			t.row--
			t.adjust()
		}
		return true
	default:
		return false
	}
}

func (t *Table) adjust() {
	// Get actual content height
	_, h := t.Size()

	// We do not need to adjust anything, if all rows fit
	if t.provider.Length() < h-2 {
		return
	}
	if t.row < t.offsetY {
		t.offsetY = t.row
	}
	if t.row > t.offsetY+h-3 {
		t.offsetY = t.row - h + 3
	}
	t.Refresh()
}
