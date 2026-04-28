package main

import (
	"database/sql"
	"fmt"

	"github.com/gdamore/tcell/v3"
	_ "github.com/mattn/go-sqlite3"
	. "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
	"github.com/tekugo/zeichenwerk/values"
	"github.com/tekugo/zeichenwerk/widgets"
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

	ui.Log(ui, core.Debug, "Starting application")
	ui.Run()
}

func createUI() *UI {
	return NewBuilder(themes.TokyoNight()).
		VFlex("main", core.Stretch, 0).
		With(header).
		With(content).
		With(footer).
		Class("").
		Build()
}

func header(builder *Builder) {
	builder.Class("header").
		HFlex("header", core.Start, 0).Padding(0, 1).Hint(0, 1).
		Static("title", "DBU").Hint(30, 1).
		Static("", "SQLite Database Utility").Hint(-1, 1).
		Static("time", "12:00").Hint(5, 1).
		Class("").
		End()
}

func footer(builder *Builder) {
	builder.Class("footer").
		HFlex("footer", core.Start, 0).Padding(0, 1).Hint(0, 1).
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
		Table("result", widgets.NewArrayTableProvider([]string{}, [][]string{}), true).
		Border("none").
		Border("grid", "thin").
		Border(":focused", "none").
		Border("grid:focused", "none").
		End()

	sql := builder.Find("sql")
	widgets.OnKey(sql, func(event *tcell.EventKey) bool {
		ui.Log(sql, core.Debug, "Key handler for SQL")
		switch event.Key() {
		case tcell.KeyCtrlR:
			query()
			return true
		default:
			ui.Log(sql, core.Debug, "Unknown key", "key", event.Key())
			return false
		}
	})

	if editor, ok := builder.Find("sql").(*widgets.Editor); ok {
		editor.Load("SELECT * FROM sqlite_schema")
	}
}

func query() {
	widget := core.Find(ui, "sql")
	editor, ok := widget.(*widgets.Editor)
	if !ok {
		return
	}

	rows, err := db.Query(editor.Text())
	ui.Log(editor, core.Debug, "Executing query", "sql", editor.Text())
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
		data = append(data, line)
	}
	values.Update(ui, "result", widgets.NewArrayTableProvider(cols, data))
	core.Find(ui, "result").Refresh()
}

func loadTables() {
	ui.Log(ui, core.Debug, "Loading tables...")
	tables := []string{}
	var name string

	rows, err := db.Query("SELECT name FROM sqlite_schema WHERE type='table' ORDER BY name")
	if err != nil {
		ui.Log(ui, core.Error, "Error loading tables", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&name)
		tables = append(tables, name)
	}

	values.Update(ui, "tables", tables)
}
