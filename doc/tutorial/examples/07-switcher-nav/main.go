// Tutorial chapter 7 — Switcher / navigation shell.
//
// A sidebar List swaps the visible content pane in a Switcher.
// Each pane is built by its own function for separation.
//
// Run with:  go run ./doc/tutorial/examples/07-switcher-nav
package main

import (
	. "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

func dashboard(b *Builder) {
	b.VFlex("dash", core.Stretch, 1).Padding(1, 2).
		Static("dash-title", "Dashboard").Font("bold").Foreground("$cyan").
		Static("dash-body", "Metrics, charts, KPIs go here.").
		End()
}

func settings(b *Builder) {
	b.VFlex("settings", core.Stretch, 1).Padding(1, 2).
		Static("set-title", "Settings").Font("bold").Foreground("$cyan").
		Checkbox("notifications", "Enable notifications", true).
		Checkbox("dark-mode", "Dark mode", true).
		Checkbox("telemetry", "Send telemetry", false).
		End()
}

func about(b *Builder) {
	b.VFlex("about", core.Stretch, 0).Padding(1, 2).
		Static("about-title", "About").Font("bold").Foreground("$cyan").
		Static("about-body", "A demo app built with zeichenwerk.").
		Static("about-version", "v0.0.1").Foreground("$gray").
		End()
}

func main() {
	ui := NewBuilder(themes.TokyoNight()).
		Grid("shell", 1, 2, false).Columns(20, -1).Rows(-1).
		// Sidebar: a List of section names.
		Cell(0, 0, 1, 1).
		List("nav", "Dashboard", "Settings", "About").
		Background("$bg1").
		// Content: a Switcher with three children, one per section.
		Cell(1, 0, 1, 1).
		Switcher("content", false).
		With(dashboard).
		With(settings).
		With(about).
		End().
		End().
		Build()

	switcher := core.MustFind[*widgets.Switcher](ui, "content")

	// Selecting a row in the nav list flips the switcher.
	widgets.OnSelect(core.MustFind[*widgets.List](ui, "nav"), func(i int) bool {
		switcher.Select(i)
		return true
	})

	ui.Run()
}
