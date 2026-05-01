// SQLite tutorial step 5 — polish.
//
// Adds:
//   - a status line under the result table that reports row counts and
//     errors instead of letting errors land inside the table,
//   - Ctrl-T to swap themes at runtime,
//   - the inspector toggle pre-wired (Ctrl-D opens a debug overlay).
//
// Run with:  go run ./doc/tutorial/examples/09-sqlite/step-5-wired
package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v3"
	_ "github.com/mattn/go-sqlite3"

	. "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
	"github.com/tekugo/zeichenwerk/v2/values"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

const dbPath = "./tutorial.db"

var (
	db        *sql.DB
	ui        *UI
	tableList *widgets.List
	editor    *widgets.Editor
	status    *widgets.Static

	themeRotation = []func() *core.Theme{
		themes.TokyoNight, themes.Nord, themes.GruvboxDark, themes.MidnightNeon,
	}
	themeIndex int
)

func main() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := bootstrap(db); err != nil {
		panic(err)
	}

	ui = createUI()
	wire()
	loadTables()
	ui.Run()
}

func createUI() *UI {
	return NewBuilder(themes.TokyoNight()).
		VFlex("root", core.Stretch, 0).
		With(header).
		With(content).
		With(footer).
		End().
		Build()
}

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
		Class("footer").Static("a1", " run  ").
		Class("shortcut").Static("k2", "[Ctrl-T]").
		Class("footer").Static("a2", " theme  ").
		Class("shortcut").Static("k3", "[Ctrl-D]").
		Class("footer").Static("a3", " inspector  ").
		Class("shortcut").Static("k4", "[Ctrl-Q]").
		Class("footer").Static("a4", " quit").
		Spacer().
		End().
		Class("")
}

// content now uses 3 rows on the right column: editor, result, status line.
func content(b *Builder) {
	b.Grid("body", 3, 2, true).Hint(0, -1).
		Columns(24, -1).Rows(-1, -1, 1).
		Cell(0, 0, 1, 3).
		List("tables").
		Cell(1, 0, 1, 1).
		Editor("sql").
		Cell(1, 1, 1, 1).
		Table("result", widgets.NewArrayTableProvider([]string{}, [][]string{}), true).
		Border("none").Border(":focused", "none").Border("grid:focused", "thin").
		Cell(1, 2, 1, 1).
		Static("status", "Ready").Foreground("$gray").Padding(0, 1).
		End()
}

func wire() {
	tableList = core.MustFind[*widgets.List](ui, "tables")
	editor = core.MustFind[*widgets.Editor](ui, "sql")
	status = core.MustFind[*widgets.Static](ui, "status")

	widgets.OnSelect(tableList, func(i int) bool {
		items := tableList.Items()
		if i < 0 || i >= len(items) {
			return false
		}
		editor.Load("SELECT * FROM " + items[i])
		return true
	})

	// Pressing Enter on a table row runs the query immediately.
	widgets.OnActivate(tableList, func(_ int) bool {
		runQuery()
		ui.Focus(editor)
		return true
	})

	// Global shortcuts on the root — Ctrl-R runs the query, Ctrl-T cycles
	// themes. Both have to be on `ui` rather than on the editor: events
	// bubble *up* from the focused widget through its parents, never
	// across to siblings, so an editor-only handler wouldn't fire when
	// focus was on the tables list. Ctrl-Q (quit) and Ctrl-D (inspector)
	// are already bound by the framework.
	widgets.OnKey(ui, func(ev *tcell.EventKey) bool {
		switch ev.Key() {
		case tcell.KeyCtrlR:
			runQuery()
			return true
		case tcell.KeyCtrlT:
			themeIndex = (themeIndex + 1) % len(themeRotation)
			ui.SetTheme(themeRotation[themeIndex]())
			return true
		}
		return false
	})
}

func runQuery() {
	q := strings.TrimSpace(editor.Text())
	// Guard: mattn/go-sqlite3 segfaults inside CGo when handed an empty
	// query, so refuse to call Query at all when there's nothing to run.
	if q == "" {
		status.Set("× empty query")
		return
	}

	rows, err := db.Query(q)
	if err != nil {
		status.Set(fmt.Sprintf("× %s", err.Error()))
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		status.Set(fmt.Sprintf("× %s", err.Error()))
		return
	}

	scratch := make([]any, len(cols))
	pointers := make([]any, len(cols))
	for i := range scratch {
		pointers[i] = &scratch[i]
	}

	var data [][]string
	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			continue
		}
		line := make([]string, len(cols))
		for i, v := range scratch {
			line[i] = fmt.Sprintf("%v", v)
		}
		data = append(data, line)
	}
	values.Update[widgets.TableProvider](ui, "result", widgets.NewArrayTableProvider(cols, data))
	status.Set(fmt.Sprintf("✓ %d row(s)", len(data)))
}

func loadTables() {
	rows, err := db.Query(
		"SELECT name FROM sqlite_schema WHERE type = 'table' ORDER BY name")
	if err != nil {
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tables = append(tables, name)
		}
	}
	values.Update(ui, "tables", tables)
}

func bootstrap(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS authors (
			id INTEGER PRIMARY KEY, name TEXT NOT NULL, country TEXT)`,
		`CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY, title TEXT NOT NULL,
			author_id INTEGER REFERENCES authors(id), year INTEGER)`,
		`CREATE TABLE IF NOT EXISTS loans (
			id INTEGER PRIMARY KEY, book_id INTEGER REFERENCES books(id),
			borrower TEXT, due_date TEXT)`,
		`INSERT OR IGNORE INTO authors (id, name, country) VALUES
			(1, 'Ursula K. Le Guin', 'US'), (2, 'Italo Calvino', 'IT'),
			(3, 'Stanisław Lem',     'PL'), (4, 'Ted Chiang',    'US')`,
		`INSERT OR IGNORE INTO books (id, title, author_id, year) VALUES
			(1, 'The Dispossessed',     1, 1974), (2, 'A Wizard of Earthsea', 1, 1968),
			(3, 'If on a winter''s night a traveler', 2, 1979),
			(4, 'Solaris',              3, 1961), (5, 'Stories of Your Life', 4, 2002)`,
		`INSERT OR IGNORE INTO loans (id, book_id, borrower, due_date) VALUES
			(1, 1, 'Alice', '2026-05-12'), (2, 4, 'Bob', '2026-05-15')`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
