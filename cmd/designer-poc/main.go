// Command designer-poc demonstrates the designer package against a
// small canned target tree. The target lives inside a Preview so
// the framework's focus / hit-testing / event walks skip the
// subtree entirely while Layout and Render still drive it normally
// — the designer mutates the live widgets and the user sees those
// edits behind the popup.
//
// Ctrl+Space opens the designer popup; Ctrl+D opens the runtime
// inspector. ESC closes the active popup. Ctrl+Q quits.
package main

import (
	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/designer"
	"github.com/tekugo/zeichenwerk/inspector"
	"github.com/tekugo/zeichenwerk/themes"
	. "github.com/tekugo/zeichenwerk/widgets"
)

func main() {
	theme := themes.TokyoNight()

	ui := NewBuilder(theme).
		VFlex("ui-root", Stretch, 0).
		Preview("preview").Hint(0, -1).
		VFlex("target-root", Stretch, 1).
		HFlex("header", Center, 2).
		Static("title", "Designer PoC").Padding(0, 1).
		Input("search", "", "", "type to filter…").
		End(). // closes header
		Grid("g1", 2, 2, false).
		Cell(0, 0, 1, 1).Static("s1", "Hello").
		Cell(1, 0, 1, 1).Class("highlight").Static("s2", "World").
		End(). // closes Grid
		End(). // closes target-root
		End(). // closes Preview
		Static("main-status", " Ctrl+Space → designer    Ctrl+D → inspector    Ctrl+Q → quit ").Hint(0, 1).
		End(). // closes ui-root
		Build()

	preview := MustFind[*Preview](ui, "preview")
	target := preview.Target().(Container)

	designer.Open(ui, target, theme, designer.DefaultProject())
	inspector.Open(ui)
	ui.Run()
}
