// Tutorial chapter 7 — Form auto-generation from a struct.
//
// The Form binds to a *struct and auto-generates Input / Checkbox / Select
// controls for its fields based on `label`, `control`, and `options` tags.
//
// Run with:  go run ./doc/tutorial/examples/07-form
package main

import (
	"fmt"

	. "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

// Profile is the data the form is bound to.
// Tags drive how each field is rendered:
//
//	label    — visible label (default: field name)
//	control  — "input" (default), "password", "checkbox", or "select"
//	options  — comma-separated values for "select"
//	width    — input width in cells
//	group    — group tag, used by Group("…", …, "<group>", …)
type Profile struct {
	Name      string `label:"Full name"        width:"30"`
	Email     string `label:"Email address"    width:"30"`
	Password  string `label:"Password"         control:"password" width:"20"`
	Theme     string `label:"Preferred theme"  control:"select"   options:"tokyo,nord,gruvbox,lipstick"`
	Newsletter bool  `label:"Subscribe to newsletter"`
}

func main() {
	profile := &Profile{
		Name:      "Ada Lovelace",
		Theme:     "tokyo",
		Newsletter: true,
	}

	ui := NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 1).Padding(1, 2).
		Static("title", "Profile").Font("bold").Foreground("$cyan").
		HRule("thin").
		Form("profile", "", profile).
		Group("fields", "", "", false, 1). // empty group name → all fields
		End(). // Group
		End(). // Form — required! Form is a single-child container, so any
		//       sibling added after the Group would *replace* it.
		HFlex("actions", core.End, 2).Hint(0, 1).
		Button("save", "Save").
		Button("quit", "Quit").
		End().
		End().
		Build()

	out := core.MustFind[*widgets.Static](ui, "title")

	widgets.OnActivate(core.MustFind[*widgets.Button](ui, "save"), func(_ int) bool {
		// Form.Update handlers have already written through to `profile`.
		out.Set(fmt.Sprintf("Saved: %+v", profile))
		return true
	})

	widgets.OnActivate(core.MustFind[*widgets.Button](ui, "quit"), func(_ int) bool {
		ui.Quit()
		return true
	})

	ui.Run()
}
