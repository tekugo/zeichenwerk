package main

import (
	z "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/compose"
)

func main() {
	UI(z.TokyoNightTheme(),
		Flex("main", "", false, "stretch", 0,
			Flex("header", "", true, "center", 1,
				Static("title", "", "My App"),
			),
			Grid("content", "", []int{-1}, []int{20, -1}, false, Hint(0, -1),
				Cell(0, 0, 1, 1, List("menu", "", []string{"Item 1", "Item 2", "Item 3"})),
				Cell(1, 0, 1, 1, Button("action", "", "Click Me")),
			),
		),
	).Run()
}
