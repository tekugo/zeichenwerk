package main

import (
	"database/sql"
	"fmt"

	"github.com/gdamore/tcell/v2"
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
		Flex("main", "vertical", "stretch", 0).
		With(header).
		With(content).
		With(footer).
		Class("").
		Build()
}

func header(builder *Builder) {
	builder.Class("header").
		Flex("header", "horizontal", "start", 0).Padding(0, 1).Hint(0, 1).
		Label("title", "DBU", 30).Hint(30, 1).
		Label("", "SQLite Database Utility", 0).Hint(-1, 1).
		Label("time", "12:00", 0).Hint(5, 1).
		Class("").
		End()
}

func footer(builder *Builder) {
	builder.Class("footer").
		Flex("footer", "horizontal", "start", 0).Padding(0, 1).Hint(0, 1).
		Class("shortcut").Label("1", "Esc", 0).
		Class("footer").Label("2", "Close \u2502", 0).
		Class("shortcut").Label("3", "Ctrl-D", 0).
		Class("footer").Label("4", "Inspector \u2502", 0).
		Class("shortcut").Label("5", "Ctrl-Q", 0).
		Class("footer").Label("6", "Quit Application \u2502", 0).
		Class("").
		Spacer().
		End()
}

func content(builder *Builder) {
	builder.Grid("grid", 2, 2, true).Hint(0, -1).
		Cell(0, 0, 1, 2).
		List("tables", []string{}).
		Cell(1, 0, 1, 1).
		Editor("sql").
		Cell(1, 1, 1, 1).
		Flex("main", "vertical", "stretch", 0).
		Tabs("tabs", "Result", "Debug").
		Switcher("switcher").Hint(-1, -1).
		Table("result", NewArrayTableProvider([]string{}, [][]string{})).
		Text("debug-log", []string{}, true, 1000).
		End().
		End().
		End()

	With[*Grid](builder.Container(), "grid", func(grid *Grid) {
		grid.Columns(30, -1)
		grid.Rows(5, -1)

		HandleKeyEvent(grid, "sql", func(widget Widget, event *tcell.EventKey) bool {
			switch event.Key() {
			case tcell.KeyCtrlR:
				query()
				return true
			default:
				return false
			}
		})
	})

	current := builder.Container()
	With[*Tabs](current, "tabs", func(tabs *Tabs) {
		tabs.On("activate", func(widget Widget, event string, params ...any) bool {
			if t, ok := widget.(*Tabs); ok {
				switch t.Selected {
				case 0:
					Update(current, "switcher", "result")
				case 1:
					Update(current, "switcher", "debug-log")
				}
			}
			return true
		})
	})

	if editor, ok := builder.Container().Find("sql", false).(*Editor); ok {
		editor.Load("SELECT * FROM sqlite_schema")
	}
}

func query() {
	widget := ui.Find("sql", false)
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
	ui.Find("result", false).Refresh()
}

func loadTables() {
	tables := []string{}
	var name string

	rows, err := db.Query("SELECT name FROM sq1lite_schema WHERE type='table'")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&name)
		tables = append(tables, name)
	}

	Update(ui, "tables", tables)
}
