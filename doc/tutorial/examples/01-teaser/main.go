// Tutorial chapter 1 — A TUI in 20 lines.
//
// Run with:  go run ./doc/tutorial/examples/01-teaser
package main

import (
	. "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
	"github.com/tekugo/zeichenwerk/widgets"
)

func main() {
	tasks := []string{
		"Read the zeichenwerk teaser",
		"Build a tiny TUI",
		"Wire up an event handler",
		"Try a different theme",
	}

	ui := NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 0).Padding(1, 2).
		Static("title", "My First TUI").Font("bold").Foreground("$cyan").
		HRule("thin").
		List("tasks", tasks...).Hint(0, -1).
		HRule("thin").
		HFlex("actions", core.End, 2).
		Button("done", "Mark Done").
		Button("quit", "Quit").
		End().
		End().
		Build()

	list := core.MustFind[*widgets.List](ui, "tasks")

	widgets.OnActivate(core.MustFind[*widgets.Button](ui, "done"), func(_ int) bool {
		i := list.Selected()
		if i >= 0 {
			items := list.Items()
			items[i] = "✓ " + items[i]
			list.Set(items)
		}
		return true
	})

	widgets.OnActivate(core.MustFind[*widgets.Button](ui, "quit"), func(_ int) bool {
		ui.Quit()
		return true
	})

	ui.Run()
}
