// designer is a full-screen TUI canvas editor with VIM-style modal editing.
package main

import (
	"fmt"

	. "github.com/tekugo/zeichenwerk"
)

type Designer struct {
	Component
	children []Widget
	canvas   *Canvas
	command  *Input
	position *Static
}

func NewDesigner() *Designer {
	designer := &Designer{
		Component: *NewComponent("designer", ""),
		children:  make([]Widget, 3),
	}

	designer.canvas = NewCanvas("canvas", "", 1, 80, 25)
	designer.position = NewStatic("position", "", " position ")
	designer.command = NewInput("command", "")
	designer.command.SetFlag(FlagHidden, true)

	designer.canvas.On(EvtMove, func(widget Widget, event Event, data ...any) bool {
		x, y, _ := designer.canvas.Cursor()
		designer.position.Set(fmt.Sprintf(" %d:%d %s ", x, y, designer.canvas.Mode()))
		designer.position.SetHint(len(designer.position.Text), 1)
		return true
	})

	designer.children = []Widget{designer.canvas, designer.position, designer.command}
	for _, child := range designer.children {
		child.SetParent(designer)
	}
	return designer
}

func (d *Designer) Children() []Widget {
	return d.children
}

func (d *Designer) Add(_ Widget, _ ...any) error {
	return nil
}

func (d *Designer) Layout() error {
	x, y, w, h := d.Content()
	lw, _ := d.children[1].Hint()
	d.children[0].SetBounds(x, y, w, h)
	d.children[1].SetBounds(x+w-lw-1, y+h-1, lw, 1)
	d.children[2].SetBounds(x, y+h-1, w, 1)
	return nil
}

func (d *Designer) Render(r *Renderer) {
	d.Component.Render(r)
	for _, child := range d.children {
		child.Render(r)
	}
}

// main function
func main() {
	ui := NewUI(TokyoNightTheme(), NewDesigner())
	ui.Run()
}
