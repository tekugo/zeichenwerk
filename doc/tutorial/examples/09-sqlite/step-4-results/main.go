// SQLite tutorial step 4 — execute SQL and show the results.
//
// Ctrl-R in the editor runs the query and pushes the rows into the result
// table by swapping its TableProvider.
//
// Run with:  go run ./doc/tutorial/examples/09-sqlite/step-4-results
package main

import (
	"database/sql"
	"fmt"
	"strings"

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
		Border("none").Border(":focused", "none").Border("grid:focused", "thin").
		End()
}

func wire() {
	tableList = core.MustFind[*widgets.List](ui, "tables")
	editor = core.MustFind[*widgets.Editor](ui, "sql")

	widgets.OnSelect(tableList, func(i int) bool {
		items := tableList.Items()
		if i < 0 || i >= len(items) {
			return false
		}
		editor.Load("SELECT * FROM " + items[i])
		return true
	})

	// Ctrl-R is a global shortcut: registered on ui (the root) so it
	// fires no matter which widget currently has focus. A handler on the
	// editor would only run when the editor itself was focused — events
	// bubble up through *parents*, not across siblings, so a Ctrl-R
	// pressed while the tables list was focused would never reach an
	// editor-only handler.
	widgets.OnKey(ui, func(ev *tcell.EventKey) bool {
		ui.Log(ui, core.Debug, "On Key", "key", ev.Key())
		if ev.Key() == tcell.KeyCtrlR {
			runQuery()
			return true
		}
		return false
	})
}

// runQuery executes the editor's text against the database and pushes the
// rows into the result Table by replacing its TableProvider.
func runQuery() {
	q := strings.TrimSpace(editor.Text())
	ui.Log(ui, core.Info, "Run query", "sql", q)
	// Guard: mattn/go-sqlite3 segfaults inside CGo when handed an empty
	// query, so refuse to call Query at all when there's nothing to run.
	if q == "" {
		values.Update(ui, "result",
			widgets.NewArrayTableProvider([]string{"info"}, [][]string{{"(empty query)"}}))
		core.Find(ui, "result").Refresh()
		return
	}

	rows, err := db.Query(q)
	if err != nil {
		// Show the error as a single-row "result". Production code would
		// surface this in a status bar (see step 5).
		values.Update(ui, "result",
			widgets.NewArrayTableProvider([]string{"error"}, [][]string{{err.Error()}}))
		core.Find(ui, "result").Refresh()
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
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
	ui.Log(ui, core.Debug, "SQL Result", "columns", len(cols), "rows", len(data))
	values.Update[widgets.TableProvider](ui, "result", widgets.NewArrayTableProvider(cols, data))
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
