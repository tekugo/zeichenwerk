package widgets

import (
	. "github.com/tekugo/zeichenwerk/core"
)

const (
	// EvtAccept is dispatched when the user accepts a suggested or pending
	// value (e.g. Tab in a Typeahead widget).
	EvtAccept Event = "accept"
	// EvtActivate is dispatched when the user activates an item, typically
	// by pressing Enter or double-clicking (e.g. List, Tree, Button).
	EvtActivate Event = "activate"
	// EvtChange is dispatched when a widget's value or state changes
	// (e.g. Checkbox toggled, Tree node expanded/collapsed).
	EvtChange Event = "change"
	// EvtBlur is dispatched when a widget loses keyboard focus.
	EvtBlur Event = "blur"
	// EvtClick is dispatched on a single mouse button-1 click.
	EvtClick Event = "click"
	// EvtClose is dispatched to a popup layer just before it is removed by
	// UI.Close, giving widgets inside the dialog a chance to clean up state.
	EvtClose Event = "close"
	// EvtEnter is dispatched if the Enter key is pressed.
	EvtEnter Event = "enter"
	// EvtFocus is dispatched when a widget gains keyboard focus.
	EvtFocus Event = "focus"
	// EvtHide is dispatched when a widget becomes hidden.
	EvtHide Event = "hide"
	// EvtHover is dispatched while the mouse cursor is over a widget.
	EvtHover Event = "hover"
	// EvtKey is dispatched for unhandled key events that bubble up.
	EvtKey Event = "key"
	// EvtMode is dispatched when a widget changes its editing mode
	// (e.g. Canvas switching between normal and insert mode).
	EvtMode Event = "mode"
	// EvtMouse is dispatched for raw mouse events.
	EvtMouse Event = "mouse"
	// EvtMove is dispatched when the highlighted/selected position changes
	// due to mouse movement.
	EvtMove Event = "move"
	// EvtPaste is dispatched when text is pasted into a widget.
	EvtPaste Event = "paste"
	// EvtSelect is dispatched when the highlighted item changes
	// (e.g. List, Tree, Deck — before activation).
	EvtSelect Event = "select"
	// EvtShow is dispatched when a widget becomes visible.
	EvtShow Event = "show"
)
