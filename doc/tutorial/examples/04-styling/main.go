// Tutorial chapter 4 — Styling, themes, classes, and selector states.
//
// Run with:  go run ./doc/tutorial/examples/04-styling
package main

import (
	. "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
)

func main() {
	NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 1).Padding(1, 2).
		// Inline overrides on a single widget.
		Static("title", "Styling demo").
		Font("bold").Foreground("$cyan").Padding(0, 1).
		// Class assigns a CSS-like style group to the next widget(s).
		// The default Tokyo Night theme styles classes like "header"
		// and "footer" — you can also define your own (see below).
		Class("header").
		Static("subtitle", " A small showcase of styling tools ").
		Class(""). // reset back to default class
		HRule("thin").
		// Borders: theme-defined names.
		Box("box1", "thin border").Border("thin").Padding(0, 1).Hint(0, 3).
		Static("inner1", "a box with a thin border").
		End().
		Box("box2", "round border").Border("round").Padding(0, 1).Hint(0, 3).
		Static("inner2", "round corners — same theme, different border style").
		End().
		// State selectors: ":focus" applies only when focused.
		// Try Tab to move focus between the buttons and watch the
		// background colour change.
		HFlex("buttons", core.Start, 2).Hint(0, 3).Padding(1, 0).
		Button("a", "Button A").
		Background("$bg2").Background(":focus", "$blue").
		Foreground(":focus", "$bg0").
		Button("b", "Button B").
		Background("$bg2").Background(":focus", "$blue").
		Foreground(":focus", "$bg0").
		End().
		HRule("thin").
		// Theme variables vs literal colours: literals lock you to one
		// palette but work without a theme entry.
		Static("v1", "Theme variable: $magenta").Foreground("$magenta").
		Static("v2", "Literal colour: #ff6347 (tomato)").Foreground("#ff6347").
		End().
		Run()
}
