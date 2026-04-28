package widgets

import (
	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// OnAccept registers an accept event handler for the given widget.
// The handler receives the accepted string value.
func OnAccept(widget Widget, handler func(string) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if value, ok := data[0].(string); ok {
			return handler(value)
		}
		return false
	})
}

// OnActivate registers an activate event handler for the given widget.
// The handler receives the index of the activated item.
func OnActivate(widget Widget, handler func(int) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if index, ok := data[0].(int); ok {
			return handler(index)
		}
		return false
	})
}

// OnChange registers a change event handler for the given widget.
// The handler receives the new value as a string.
func OnChange(widget Widget, handler func(string) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if value, ok := data[0].(string); ok {
			return handler(value)
		}
		return false
	})
}

// OnEnter registers an Enter event handler for the given widget.
func OnEnter(widget Widget, handler func(value string) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtEnter, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if value, ok := data[0].(string); ok {
			return handler(value)
		}
		return false
	})
}

// OnHide registers a hide event handler for the given widget.
func OnHide(widget Widget, handler func() bool) {
	if widget == nil {
		return
	}
	widget.On(EvtHide, func(_ Widget, _ Event, data ...any) bool {
		return handler()
	})
}

// OnKey registers a key event handler for the given widget.
func OnKey(widget Widget, handler func(*tcell.EventKey) bool) {
	if widget == nil {
		return
	}

	widget.On(EvtKey, func(_ Widget, _ Event, data ...any) bool {
		if len(data) != 1 {
			return false
		}
		if ev, ok := data[0].(*tcell.EventKey); ok {
			return handler(ev)
		} else {
			return false
		}
	})
}

// OnMouse registers a mouse event handler for the given widget.
func OnMouse(widget Widget, handler func(*tcell.EventMouse) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtMouse, func(_ Widget, _ Event, data ...any) bool {
		if len(data) != 1 {
			return false
		}
		if ev, ok := data[0].(*tcell.EventMouse); ok {
			return handler(ev)
		} else {
			return false
		}
	})
}

// OnSelect registers a select event handler for the given widget.
// The handler receives the index of the selected item.
func OnSelect(widget Widget, handler func(int) bool) {
	if widget == nil {
		return
	}
	widget.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		if len(data) < 1 {
			return false
		}
		if index, ok := data[0].(int); ok {
			return handler(index)
		}
		return false
	})
}

// OnShow registers a show event handler for the given widget.
func OnShow(widget Widget, handler func() bool) {
	if widget == nil {
		return
	}
	widget.On(EvtShow, func(_ Widget, _ Event, data ...any) bool {
		return handler()
	})
}
