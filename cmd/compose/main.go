// Compose API example
package main

import (
	zw "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/compose"
)

func main() {
	UI(zw.TokyoNightTheme(),
		Flex("main", "", false, "stretch", 0,
			Include(header),
			Include(content),
			Include(footer),
		),
	).Debug().Run()
}

func header() zw.Widget {
	return Build(
		Static("header", "", "Header"),
	)
}

func footer() zw.Widget {
	return Build(
		Static("footer", "", "Footer"),
	)
}

func content() zw.Widget {
	return Build(
		Grid("main", "", []int{-1, 10}, []int{30, -1}, true,
			Hint(-1, -1),
			Cell(0, 0, 1, 2, List("list", "", []string{"One", "Two", "Three"})),
			Cell(1, 0, 1, 1, Static("s1", "", "Panel 1")),
			Cell(1, 1, 1, 1, Static("s2", "", "Panel 2")),
		),
	)
}
