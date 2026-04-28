// Tutorial chapter 3 — Flex layout demo.
//
// Shows VFlex and HFlex working together to build a typical
// header / body / footer layout.
//
// Run with:  go run ./doc/tutorial/examples/03-layout-flex
package main

import (
	. "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
)

func main() {
	NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 0).
		// Header: full width, fixed height of 1 row.
		HFlex("header", core.Center, 0).Hint(0, 1).Background("$bg2").
		Static("title", "Flex Demo").Font("bold").Foreground("$cyan").
		End().
		// Body: takes all remaining vertical space.
		HFlex("body", core.Stretch, 0).Hint(0, -1).
		Static("sidebar", "Sidebar (24 wide)").Hint(24, 0).Background("$bg1").Padding(1, 2).
		Static("content", "Content (fills the rest)").Hint(-1, 0).Padding(1, 2).
		End().
		// Footer: full width, 1 row tall, content right-aligned.
		HFlex("footer", core.End, 2).Hint(0, 1).Background("$bg2").Padding(0, 1).
		Static("hint1", "[Tab] navigate").Foreground("$gray").
		Static("hint2", "[q] quit").Foreground("$gray").
		End().
		End().
		Run()
}
