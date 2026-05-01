// SQLite tutorial step 1 — skeleton layout.
//
// Just the chrome: header, 3-pane content area, footer. No database yet.
// The shape we're targeting:
//
//   ┌─ DBU ───────────────── SQLite Tutorial ─────────────────────┐
//   │ tables   │ SQL editor                                        │
//   │  …       │                                                   │
//   │          ├───────────────────────────────────────────────────┤
//   │          │ result table                                      │
//   └──────────┴───────────────────────────────────────────────────┘
//   [q] quit  [Ctrl-D] inspector
//
// Run with:  go run ./doc/tutorial/examples/09-sqlite/step-1-skeleton
package main

import (
	. "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

func header(b *Builder) {
	b.Class("header").
		HFlex("hdr", core.Start, 0).Hint(0, 1).Padding(0, 1).
		Static("title", "DBU").Hint(10, 1).Font("bold").
		Static("subtitle", "SQLite Tutorial").Hint(-1, 1).
		End().
		Class("")
}

func footer(b *Builder) {
	b.Class("footer").
		HFlex("ftr", core.Start, 0).Hint(0, 1).Padding(0, 1).
		Class("shortcut").Static("k1", "[Ctrl-R]").
		Class("footer").Static("a1", " run query  ").
		Class("shortcut").Static("k2", "[Ctrl-D]").
		Class("footer").Static("a2", " inspector  ").
		Class("shortcut").Static("k3", "[Ctrl-Q]").
		Class("footer").Static("a3", " quit").
		Spacer().
		End().
		Class("")
}

func content(b *Builder) {
	// 2 columns × 2 rows. Left pane spans both rows.
	b.Grid("body", 2, 2, true).Hint(0, -1).
		Columns(24, -1).Rows(-1, -1).
		Cell(0, 0, 1, 2). // sidebar — both rows
		List("tables", "(no database yet)").
		Cell(1, 0, 1, 1).
		Editor("sql").
		Cell(1, 1, 1, 1).
		Table("result", widgets.NewArrayTableProvider([]string{}, [][]string{}), true).
		End()
}

func main() {
	NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 0).
		With(header).
		With(content).
		With(footer).
		End().
		Run()
}
