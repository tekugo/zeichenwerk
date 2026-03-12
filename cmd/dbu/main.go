package main

import (
	"database/sql"
	"fmt"

	"github.com/gdamore/tcell/v3"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/tekugo/zeichenwerk"
)

var (
	db *sql.DB
	ui *UI
)

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./cmd/dbu/test.db")
	if err != nil {
		panic(err)
	}

	ui = createUI()
	loadTables()

	ui.Run()
}

func createUI() *UI {
	return NewBuilder(TokyoNightTheme()).
		Flex("main", false, "stretch", 0).
		With(header).
		With(content).
		With(footer).
		Class("").
		Build()
}

func header(builder *Builder) {
	builder.Class("header").
		Flex("header", true, "start", 0).Padding(0, 1).Hint(0, 1).
		Static("title", "DBU").Hint(30, 1).
		Static("", "SQLite Database Utility").Hint(-1, 1).
		Static("time", "12:00").Hint(5, 1).
		Class("").
		End()
}

func footer(builder *Builder) {
	builder.Class("footer").
		Flex("footer", true, "start", 0).Padding(0, 1).Hint(0, 1).
		Class("shortcut").Static("1", "Esc").
		Class("footer").Static("2", "Close \u2502").
		Class("shortcut").Static("3", "Ctrl-D").
		Class("footer").Static("4", "Inspector \u2502").
		Class("shortcut").Static("5", "Ctrl-Q").
		Class("footer").Static("6", "Quit Application \u2502").
		Class("").
		Spacer().
		End()
}

func content(builder *Builder) {
	builder.Grid("grid", 2, 2, true).Columns(30, -1).Rows(5, -1).Hint(0, -1).
		Cell(0, 0, 1, 2).
		List("tables").
		Cell(1, 0, 1, 1).
		Editor("sql").
		Cell(1, 1, 1, 1).
		Flex("main", false, "stretch", 0).
		Tabs("tabs").
		Switcher("switcher", true).Hint(-1, -1).
		Tab("Query Result").Table("result", NewArrayTableProvider([]string{}, [][]string{})).
		Tab("Debug Log").Text("debug-log", []string{}, true, 1000).
		End().
		End().
		End()

	OnKey(builder.Find("sql"), func(widget Widget, event *tcell.EventKey) bool {
		switch event.Key() {
		case tcell.KeyCtrlR:
			query()
			return true
		default:
			return false
		}
	})

	if editor, ok := builder.Find("sql").(*Editor); ok {
		editor.Load("SELECT * FROM sqlite_schema")
	}
}

func query() {
	widget := Find(ui, "sql")
	editor, ok := widget.(*Editor)
	if !ok {
		return
	}

	rows, err := db.Query(editor.Text())
	ui.Log(editor, "debug", "Executing query %s", editor.Text())
	if err != nil {
		panic(err)
	}

	fill(rows)
}

func fill(rows *sql.Rows) {
	data := [][]string{}

	cols, _ := rows.Columns()
	row := make([]any, len(cols))
	rowPtrs := make([]any, len(cols))
	for i := range cols {
		rowPtrs[i] = &row[i]
	}
	for rows.Next() {
		rows.Scan(rowPtrs...)
		line := make([]string, len(cols))
		for i := range len(cols) {
			line[i] = fmt.Sprintf("%v", row[i])
		}
		ui.Log(ui, "debug", "%v", line)
		data = append(data, line)
	}
	ui.Log(ui, "debug", "Result returned %d rows", len(data))
	Update(ui, "result", NewArrayTableProvider(cols, data))
	Find(ui, "result").Refresh()
}

func loadTables() {
	ui.Log(ui, "debug", "Loading tables...")
	tables := []string{}
	var name string

	rows, err := db.Query("SELECT name FROM sqlite_schema WHERE type='table' ORDER BY name")
	if err != nil {
		ui.Log(ui, "error", "Error loading tables: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&name)
		ui.Log(ui, "debug", "Table %s", name)
		tables = append(tables, name)
	}

	Update(ui, "tables", tables)
}
