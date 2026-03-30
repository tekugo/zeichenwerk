package zeichenwerk

// Event represents a named event that widgets can dispatch and subscribe to.
type Event string

const (
	EvtAccept   Event = "accept"
	EvtActivate Event = "activate"
	EvtChange   Event = "change"
	EvtClick    Event = "click"
	EvtEnter    Event = "enter"
	EvtHide     Event = "hide"
	EvtHover    Event = "hover"
	EvtKey      Event = "key"
	EvtMode     Event = "mode"
	EvtMouse    Event = "mouse"
	EvtMove     Event = "move"
	EvtPaste    Event = "paste"
	EvtSelect   Event = "select"
	EvtShow     Event = "show"
)
