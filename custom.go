package zeichenwerk

import "github.com/gdamore/tcell/v2"

type Custom struct {
	BaseWidget
	handler  func(tcell.Event) bool
	renderer func(Screen)
}

func NewCustom(id string, focusable bool, renderer func(Screen)) *Custom {
	return &Custom{
		BaseWidget: BaseWidget{id: id, focusable: focusable},
		renderer:   renderer,
	}
}
