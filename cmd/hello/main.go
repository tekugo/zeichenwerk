package main

import (
	. "github.com/tekugo/zeichenwerk/compose"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
)

func main() {
	UI(themes.TokyoNight(),
		VFlex("main", "", core.Stretch, 0,
			HFlex("header", "", core.Center, 1,
				Static("title", "", "My App"),
			),
			Grid("content", "", []int{-1}, []int{20, -1}, false, Hint(0, -1),
				Cell(0, 0, 1, 1, List("menu", "", []string{"Item 1", "Item 2", "Item 3"})),
				Cell(1, 0, 1, 1, Button("action", "", "Click Me")),
			),
		),
	).Run()
}
