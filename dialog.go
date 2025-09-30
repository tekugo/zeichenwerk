package zeichenwerk

import "github.com/gdamore/tcell/v2"

type Dialog struct {
	BaseWidget
	Title string
	child Widget
	runes map[rune]string
}

func NewDialog(id, title string) *Dialog {
	return &Dialog{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Title:      title,
		runes:      make(map[rune]string),
	}
}

func (d *Dialog) Action(key rune, emit string) {
	d.runes[key] = emit
}

func (d *Dialog) Add(widget Widget) {
	d.child = widget
}

func (d *Dialog) Children(_ bool) []Widget {
	if d.child == nil {
		return []Widget{}
	}
	return []Widget{d.child}
}

func (d *Dialog) Emit(event string, data ...any) bool {
	if d.handlers == nil {
		return false
	}
	handler, found := d.handlers[event]
	if found {
		return handler(d, event, data...)
	}
	return false
}

func (d *Dialog) Find(id string, visible bool) Widget {
	return Find(d, id, visible)
}

func (d *Dialog) FindAt(x, y int) Widget {
	return FindAt(d, x, y)
}

func (d *Dialog) Handle(evt tcell.Event) bool {
	switch event := evt.(type) {
	case *tcell.EventKey:
		switch event.Key() {
		case tcell.KeyEsc:
			FindUI(d).Close()
			d.Emit("close", event)
		case tcell.KeyRune:
			if action, ok := d.runes[event.Rune()]; ok {
				FindUI(d).Close()
				d.Emit(action, event)
			}
		default:
			return false
		}
	default:
		return false
	}

	return true
}

func (d *Dialog) Hint() (int, int) {
	if d.child != nil {
		w, h := d.child.Hint()
		w += d.Style("").Horizontal()
		h += d.Style("").Vertical()
		return w, h
	} else {
		return 0, 0
	}
}

func (d *Dialog) Info() string {
	return "dialog [" + d.BaseWidget.Info() + "]"
}

func (d *Dialog) Layout() {
	if d.child != nil {
		cx, cy, cw, ch := d.Content()
		d.child.SetBounds(cx, cy, cw, ch)
	}
	Layout(d)
}
