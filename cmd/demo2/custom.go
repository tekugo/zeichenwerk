package main

import (
	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/widgets"
)

var customEntries = []Entry{
	{
		Category: "Custom",
		Name:     "Custom",
		Summary:  "Widget that delegates rendering to a user-supplied function.",
		DocFile:  "custom.md",
		DemoFn:   customDemo,
		Builder: `c := NewCustom("stars", "", func(w Widget, r *Renderer) {
    _, _, w_, h_ := w.Content()
    for y := 0; y < h_; y += 2 {
        for x := 0; x < w_; x += 4 {
            r.Put(x, y, "✦")
        }
    }
})
c.SetStyle("", NewStyle().WithColors("$cyan", "$bg0"))
builder.Add(c)`,
		Compose: `compose.Custom("stars", "", func(w core.Widget, r *core.Renderer) {
    _, _, w_, h_ := w.Content()
    for y := 0; y < h_; y += 2 {
        for x := 0; x < w_; x += 4 {
            r.Put(x, y, "✦")
        }
    }
})`,
	},
}

func customDemo(b *Builder) {
	c := NewCustom("stars", "", func(w Widget, r *Renderer) {
		_, _, w_, h_ := w.Content()
		for y := 0; y < h_; y++ {
			for x := 0; x < w_; x++ {
				if (x*7+y*13)%23 == 0 {
					r.Put(x, y, "✦")
				} else if (x*5+y*11)%37 == 0 {
					r.Put(x, y, "·")
				}
			}
		}
	})
	c.SetStyle("", NewStyle("").WithColors("$cyan", "$bg0").WithMargin(0).WithPadding(0))

	b.VFlex("custom-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Custom is the simplest extension point: pass a render function to NewCustom and you have a new widget.").
		Padding(0, 0, 1, 0).
		Add(c).Hint(0, -1).
		End()
}
