package compose

import (
	"github.com/gdamore/tcell/v3"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

// ---- Event Handling -------------------------------------------------------

// On registers a raw event handler on the widget. Use the typed helpers
// ([OnActivate], [OnChange], etc.) when available for a cleaner call site.
func On(event core.Event, handler core.Handler) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widget.On(event, handler)
	}
}

// OnAccept registers a handler called when the user accepts a suggested or
// pending value (e.g. pressing Tab in a [Typeahead]). value is the accepted
// string.
func OnAccept(fn func(value string) bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnAccept(widget, fn)
	}
}

// OnActivate registers a handler called when the user activates an item
// (e.g. pressing Enter on a [List] row or a [Button]). index is the
// zero-based position of the activated item.
func OnActivate(fn func(index int) bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnActivate(widget, fn)
	}
}

// OnChange registers a handler called when the widget value changes
// (e.g. typing in an [Input] or toggling a [Checkbox]). value is the new
// string representation of the widget's value.
func OnChange(fn func(value string) bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnChange(widget, fn)
	}
}

// OnEnter registers a handler called when the user confirms input by pressing
// Enter (e.g. in an [Input] field). value is the current input string.
func OnEnter(fn func(value string) bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnEnter(widget, fn)
	}
}

// OnKey registers a handler called on every key event received by the widget.
func OnKey(fn func(*tcell.EventKey) bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnKey(widget, fn)
	}
}

// OnMouse registers a handler called on every mouse event received by the widget.
func OnMouse(fn func(*tcell.EventMouse) bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnMouse(widget, fn)
	}
}

// OnHide registers a handler called when the widget becomes hidden.
func OnHide(fn func() bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnHide(widget, fn)
	}
}

// OnSelect registers a handler called when the highlighted item changes
// (e.g. moving through a [List] or [Deck] before activation). index is the
// zero-based position of the newly highlighted item.
func OnSelect(fn func(index int) bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnSelect(widget, fn)
	}
}

// OnShow registers a handler called when the widget becomes visible.
func OnShow(fn func() bool) Option {
	return func(_ *core.Theme, widget core.Widget) {
		widgets.OnShow(widget, fn)
	}
}
