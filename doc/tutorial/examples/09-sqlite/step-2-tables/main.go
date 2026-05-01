// SQLite tutorial step 2 — populate the table list from a real database.
//
// On first run the program creates ./tutorial.db with a few demo tables
// (books, authors, loans) so the example is self-contained.
//
// Run with:  go run ./doc/tutorial/examples/09-sqlite/step-2-tables
package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	. "github.com/tekugo/zeichenwerk/v2"
	"github.com/tekugo/zeichenwerk/v2/core"
	"github.com/tekugo/zeichenwerk/v2/themes"
	"github.com/tekugo/zeichenwerk/v2/values"
	"github.com/tekugo/zeichenwerk/v2/widgets"
)

const dbPath = "./tutorial.db"

var (
	db *sql.DB
	ui *UI
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

// loadTables queries sqlite_schema and pushes the table names into the List.
//
// values.Update calls List.Set([]string) under the hood — so we don't need
// to MustFind the widget and assert its type ourselves.
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

// bootstrap creates demo tables once. Idempotent — IF NOT EXISTS keeps
// repeated runs fast.
func bootstrap(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS authors (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			country TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			author_id INTEGER REFERENCES authors(id),
			year INTEGER
		)`,
		`CREATE TABLE IF NOT EXISTS loans (
			id INTEGER PRIMARY KEY,
			book_id INTEGER REFERENCES books(id),
			borrower TEXT,
			due_date TEXT
		)`,
		`INSERT OR IGNORE INTO authors (id, name, country) VALUES
			(1, 'Ursula K. Le Guin', 'US'),
			(2, 'Italo Calvino',     'IT'),
			(3, 'Stanisław Lem',     'PL'),
			(4, 'Ted Chiang',        'US')`,
		`INSERT OR IGNORE INTO books (id, title, author_id, year) VALUES
			(1, 'The Dispossessed',          1, 1974),
			(2, 'A Wizard of Earthsea',      1, 1968),
			(3, 'If on a winter''s night a traveler', 2, 1979),
			(4, 'Solaris',                   3, 1961),
			(5, 'Stories of Your Life',      4, 2002)`,
		`INSERT OR IGNORE INTO loans (id, book_id, borrower, due_date) VALUES
			(1, 1, 'Alice', '2026-05-12'),
			(2, 4, 'Bob',   '2026-05-15')`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
