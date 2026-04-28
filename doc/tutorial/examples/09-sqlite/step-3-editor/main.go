// SQLite tutorial step 3 — wire the editor and the tables list.
//
// Selecting a table in the sidebar prefills the editor with
// `SELECT * FROM <table>`. Ctrl-R doesn't run a query yet — that's step 4.
//
// Run with:  go run ./doc/tutorial/examples/09-sqlite/step-3-editor
package main

import (
	"database/sql"

	"github.com/gdamore/tcell/v3"
	_ "github.com/mattn/go-sqlite3"

	. "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
	"github.com/tekugo/zeichenwerk/values"
	"github.com/tekugo/zeichenwerk/widgets"
)

const dbPath = "./tutorial.db"

var (
	db        *sql.DB
	ui        *UI
	tableList *widgets.List
	editor    *widgets.Editor
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
	b.Grid("body", 2, 2, true).Hint(0, -1).
		Columns(24, -1).Rows(-1, -1).
		Cell(0, 0, 1, 2).
		List("tables").
		Cell(1, 0, 1, 1).
		Editor("sql").
		Cell(1, 1, 1, 1).
		Table("result", widgets.NewArrayTableProvider([]string{}, [][]string{}), true).
		End()
}

// wire fishes the list and editor out of the tree once and registers
// handlers. Doing this in a separate function keeps createUI() declarative
// and the wiring imperative — the pattern from chapter 2.
func wire() {
	tableList = core.MustFind[*widgets.List](ui, "tables")
	editor = core.MustFind[*widgets.Editor](ui, "sql")

	// Selecting a row in the sidebar prefills the editor with
	// "SELECT * FROM <table>". EvtSelect fires on every arrow-key move,
	// so the editor follows the highlight live.
	widgets.OnSelect(tableList, func(i int) bool {
		items := tableList.Items()
		if i < 0 || i >= len(items) {
			return false
		}
		editor.Load("SELECT * FROM " + items[i])
		return true
	})

	// Ctrl-R will run the query in the next step. For now, just log it
	// so we can confirm the key handler is wired.
	widgets.OnKey(editor, func(ev *tcell.EventKey) bool {
		if ev.Key() == tcell.KeyCtrlR {
			ui.Log(editor, core.Info, "Ctrl-R pressed — query stub")
			return true // consume; next step will execute
		}
		return false
	})
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
