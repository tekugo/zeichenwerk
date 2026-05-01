// Tutorial chapter 5 — Events, focus, and redraw vs. relayout.
//
// A small form: an Input echoes its value into a Static, a List shows
// what was clicked, a Button quits.
//
// Run with:  go run ./doc/tutorial/examples/05-events
package main

import (
	"fmt"

	. "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

func main() {
	ui := NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 1).Padding(1, 2).
		Static("title", "Events demo (Tab to move focus)").
		Font("bold").Foreground("$cyan").
		HRule("thin").
		HFlex("row", core.Stretch, 1).Hint(0, 1).
		Static("label", "Type something:").Hint(20, 1).
		Input("name").Hint(-1, 1).
		End().
		Static("echo", "(your text shows up here)").
		Foreground("$gray").Hint(0, 1).
		HRule("thin").
		Static("listLabel", "Pick a colour:").
		List("colours", "Red", "Green", "Blue", "Cyan", "Magenta", "Yellow").
		Hint(0, -1).
		Static("status", "(events appear here)").
		Foreground("$gray").Hint(0, 1).
		HRule("thin").
		HFlex("actions", core.End, 0).Hint(0, 1).
		Button("quit", "Quit").
		End().
		End().
		Build()

	echo := core.MustFind[*widgets.Static](ui, "echo")
	status := core.MustFind[*widgets.Static](ui, "status")

	// EvtChange fires on every keystroke into the Input.
	widgets.OnChange(core.MustFind[*widgets.Input](ui, "name"), func(value string) bool {
		echo.Set("→ " + value) // Set() also queues a refresh internally
		return true
	})

	colours := core.MustFind[*widgets.List](ui, "colours")

	// EvtSelect fires when the highlighted row moves (arrow keys, click).
	widgets.OnSelect(colours, func(index int) bool {
		status.Set(fmt.Sprintf("selected: %s", colours.Items()[index]))
		return true
	})

	// EvtActivate fires on Enter or double-click.
	widgets.OnActivate(colours, func(index int) bool {
		status.Set(fmt.Sprintf("activated: %s !", colours.Items()[index]))
		return true
	})

	widgets.OnActivate(core.MustFind[*widgets.Button](ui, "quit"), func(_ int) bool {
		ui.Quit()
		return true
	})

	ui.Run()
}
