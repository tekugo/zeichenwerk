// Tutorial chapter 3 — Hint() value semantics.
//
// Three rows demonstrating positive, negative, and zero hints.
//
// Run with:  go run ./doc/tutorial/examples/03-layout-hints
package main

import (
	. "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
)

func main() {
	NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 1).Padding(1, 2).
		Static("h1", "Positive width = fixed cells").Font("bold").Foreground("$cyan").
		HFlex("row1", core.Stretch, 1).Hint(0, 1).
		Static("a", "fixed 10").Hint(10, 0).Background("$bg2").
		Static("b", "fixed 20").Hint(20, 0).Background("$bg1").
		Static("c", "auto").Background("$bg2").
		End().
		Static("h2", "Negative width = fractional weight").Font("bold").Foreground("$cyan").
		HFlex("row2", core.Stretch, 1).Hint(0, 1).
		Static("d", "weight -1").Hint(-1, 0).Background("$bg2").
		Static("e", "weight -2").Hint(-2, 0).Background("$bg1").
		Static("f", "weight -3").Hint(-3, 0).Background("$bg2").
		End().
		Static("h3", "Zero width = ask the widget").Font("bold").Foreground("$cyan").
		HFlex("row3", core.Start, 1).Hint(0, 1).
		Static("g", "short").Background("$bg2").
		Static("h", "a noticeably longer label").Background("$bg1").
		Static("i", "tiny").Background("$bg2").
		End().
		End().
		Run()
}
