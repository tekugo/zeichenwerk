# 9. Building a SQLite query tool

A real application brings the concepts from Parts II and III together —
layout, events, focus, table providers, status reporting. This chapter
builds **DBU**, a small SQLite query tool, in five incremental steps:

```
┌─ DBU ───────────── SQLite Tutorial ─────────────────────────────┐
│ authors  │ SELECT * FROM books;                                  │
│ books    │                                                       │
│ loans    ├───────────────────────────────────────────────────────┤
│          │ id │ title                  │ author_id │ year       │
│          │  1 │ The Dispossessed       │     1     │ 1974       │
│          │  2 │ A Wizard of Earthsea   │     1     │ 1968       │
│          │  …                                                    │
├──────────┴───────────────────────────────────────────────────────┤
│ ✓ 5 row(s)                                                        │
├──────────────────────────────────────────────────────────────────┤
│ [Ctrl-R] run  [Ctrl-T] theme  [Ctrl-D] inspector  [Ctrl-Q] quit │
└──────────────────────────────────────────────────────────────────┘
```

Five steps, each a runnable `main.go`:

| Step | What it adds | Source |
|------|--------------|--------|
| 1 | Layout skeleton — header, sidebar, editor, result, footer | [`step-1-skeleton`](examples/09-sqlite/step-1-skeleton/main.go) |
| 2 | Real database; populate the sidebar from `sqlite_schema` | [`step-2-tables`](examples/09-sqlite/step-2-tables/main.go) |
| 3 | Editor wiring — selecting a table prefills `SELECT * FROM …` | [`step-3-editor`](examples/09-sqlite/step-3-editor/main.go) |
| 4 | Run the query, push rows into the result table | [`step-4-results`](examples/09-sqlite/step-4-results/main.go) |
| 5 | Polish — status bar, theme switching, inspector | [`step-5-wired`](examples/09-sqlite/step-5-wired/main.go) |

Each program creates `./tutorial.db` with three demo tables (authors,
books, loans) on first launch. Run them from the project root:

```bash
go run ./doc/tutorial/examples/09-sqlite/step-5-wired
```

The DB lives in your current directory; delete it any time, the next run
re-creates it.

## Step 1 — skeleton

We start with no database — just the layout shape, so we can confirm the
geometry before we add behaviour.

```go
func content(b *Builder) {
    b.Grid("body", 2, 2, true).Hint(0, -1).
        Columns(24, -1).Rows(-1, -1).
        Cell(0, 0, 1, 2).
            List("tables", "(no database yet)").
        Cell(1, 0, 1, 1).
            Editor("sql").
        Cell(1, 1, 1, 1).
            Table("result", widgets.NewArrayTableProvider([]string{}, [][]string{}), true).
    End()
}
```

What's interesting:

- The grid is **2 × 2 with the sidebar spanning both rows** — that's
  what `Cell(0, 0, 1, 2)` does (column 0, row 0, span 1 column × 2 rows).
- The sidebar gets a fixed 24-cell width; everything else flexes.
- The result table uses an empty `ArrayTableProvider` as a placeholder.
  Step 4 will swap in a populated provider.

Header and footer go in their own `func(b *Builder)` helpers and are
wired in via `.With(header)`, `.With(content)`, `.With(footer)` — the
composition trick from Chapter 7.

[`step-1-skeleton/main.go`](examples/09-sqlite/step-1-skeleton/main.go)

## Step 2 — load tables

Now we open SQLite, create demo tables on first run, and populate the
sidebar from `sqlite_schema`:

```go
func loadTables() {
    rows, _ := db.Query(
        "SELECT name FROM sqlite_schema WHERE type = 'table' ORDER BY name")
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var name string
        rows.Scan(&name)
        tables = append(tables, name)
    }
    values.Update(ui, "tables", tables)
}
```

Two pieces worth highlighting:

- **`values.Update(ui, "tables", tables)`** is the data-binding helper
  from chapter 6. It looks up the widget with that ID, checks if it
  implements `values.Setter[[]string]`, and calls `Set([]string)` on it.
  No type assertion, no `MustFind` — just push data in by ID.
- The DB is opened once globally (`var db *sql.DB`). For a real app
  you'd plumb it through, but for a focused example, package-level
  state keeps the diff readable.

`bootstrap(db)` runs idempotent `CREATE TABLE IF NOT EXISTS` and
`INSERT OR IGNORE` statements, so repeated runs are fast and re-runs
don't pile up duplicates.

[`step-2-tables/main.go`](examples/09-sqlite/step-2-tables/main.go)

## Step 3 — wire the editor

We split UI assembly from event wiring — the **build declaratively, wire
imperatively** pattern from Chapter 2:

```go
func wire() {
    tableList = core.MustFind[*widgets.List](ui, "tables")
    editor    = core.MustFind[*widgets.Editor](ui, "sql")

    widgets.OnSelect(tableList, func(i int) bool {
        editor.Load("SELECT * FROM " + tableList.Items()[i])
        return true
    })

    widgets.OnKey(editor, func(ev *tcell.EventKey) bool {
        if ev.Key() == tcell.KeyCtrlR {
            ui.Log(editor, core.Info, "Ctrl-R pressed — query stub")
            return true
        }
        return false
    })
}
```

`EvtSelect` fires every time the highlight in the list moves, so the
editor follows along live. `editor.Load(text)` resets the editor's
contents — `Editor` exposes `Load`, `Text`, and `Lines` for content
access.

The Ctrl-R handler is a stub here; step 4 wires it up.

> **Heads-up for step 4.** This handler is on `editor`, so it only fires
> when the editor is focused. Pressing Ctrl-R from the tables list won't
> reach it — events bubble *up* from the focused widget through its
> parents, never across to siblings. Step 4 promotes Ctrl-R to a global
> shortcut by attaching it to `ui` instead.

[`step-3-editor/main.go`](examples/09-sqlite/step-3-editor/main.go)

## Step 4 — run the query

Two changes: write `runQuery`, and promote Ctrl-R from an editor-local
handler to a **global** shortcut on `ui` so it fires regardless of which
pane has focus:

```go
// Note the receiver: ui, not editor.
widgets.OnKey(ui, func(ev *tcell.EventKey) bool {
    if ev.Key() == tcell.KeyCtrlR {
        runQuery()
        return true
    }
    return false
})
```

A handler on the editor would only run while the editor was focused.
Events propagate *up* from the focused widget through its parents — never
sideways to siblings — so a Ctrl-R pressed while the tables list was
focused would never reach an editor-only handler. Putting it on `ui`
lets every keystroke that no other widget consumes pass through this
handler.

`runQuery` itself is plain `database/sql` glue with two TUI-side
gotchas you need to handle:

1. **`Table.Set` does not redraw on its own.** Unlike `List.Set` (which
   calls `Refresh()` internally), the `Table` widget only updates its
   provider and recalculates widths — you have to trigger the redraw
   yourself with `core.Find(ui, "result").Refresh()`. Without it,
   `runQuery` will succeed silently and you'll see no rows.
2. **`mattn/go-sqlite3` SIGSEGVs on an empty query string.** It's a CGo
   nil-pointer in `sqlite3_clear_bindings` — not a Go error, an honest
   segfault. Guard with `strings.TrimSpace(...)` before calling
   `db.Query`.

Putting both together:

```go
func runQuery() {
    q := strings.TrimSpace(editor.Text())
    if q == "" {
        values.Update(ui, "result",
            widgets.NewArrayTableProvider([]string{"info"}, [][]string{{"(empty query)"}}))
        core.Find(ui, "result").Refresh()
        return
    }

    rows, err := db.Query(q)
    if err != nil {
        values.Update(ui, "result",
            widgets.NewArrayTableProvider([]string{"error"}, [][]string{{err.Error()}}))
        core.Find(ui, "result").Refresh()
        return
    }
    defer rows.Close()

    cols, _ := rows.Columns()

    scratch  := make([]any, len(cols))
    pointers := make([]any, len(cols))
    for i := range scratch {
        pointers[i] = &scratch[i]
    }

    var data [][]string
    for rows.Next() {
        rows.Scan(pointers...)
        line := make([]string, len(cols))
        for i, v := range scratch {
            line[i] = fmt.Sprintf("%v", v)
        }
        data = append(data, line)
    }
    values.Update(ui, "result", widgets.NewArrayTableProvider(cols, data))
    core.Find(ui, "result").Refresh()
}
```

The pattern with `scratch` / `pointers` is the standard `database/sql`
trick for scanning rows into `[]any` when columns are dynamic.

`widgets.NewArrayTableProvider(cols, data)` builds a provider from a
header slice and a `[][]string`. For non-array data (a streaming cursor,
a paged API), implement the `TableProvider` interface yourself — three
methods: `Columns`, `Length`, `Str(row, col)`.

[`step-4-results/main.go`](examples/09-sqlite/step-4-results/main.go)

## Step 5 — polish

Errors landing inside the result table is ugly. Let's add a one-line
status under the result and report there. While we're here, we'll add
runtime theme switching (Ctrl-T) and Enter-to-run on the table list:

```go
// Grid grew a row.
b.Grid("body", 3, 2, true).Hint(0, -1).
    Columns(24, -1).Rows(-1, -1, 1).
    Cell(0, 0, 1, 3).             // sidebar spans all 3 rows now
        List("tables").
    Cell(1, 0, 1, 1).
        Editor("sql").
    Cell(1, 1, 1, 1).
        Table("result", …, true).
    Cell(1, 2, 1, 1).              // status line
        Static("status", "Ready").Foreground("$gray").Padding(0, 1).
End()
```

```go
// Pressing Enter on a row runs the query and jumps focus to the editor.
widgets.OnActivate(tableList, func(_ int) bool {
    runQuery()
    ui.Focus(editor)
    return true
})

// One global key handler on `ui` covers Ctrl-R (run) and Ctrl-T (cycle
// themes). Both have to be on the root, not on the editor — events
// bubble *up* from the focused widget through its parents, never across
// to siblings.
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
```

Wiring `OnKey` on `ui` catches keys that no focused widget consumed —
great for global shortcuts. The framework already binds `Ctrl-Q` /
`Ctrl-C` (quit), `Tab` / `Shift-Tab` (focus traversal), and `Ctrl-D`
(inspector overlay) — see [Debugging](A-debugging.md). You don't need
to wire those.

[`step-5-wired/main.go`](examples/09-sqlite/step-5-wired/main.go)

## What you didn't have to write

The 200 lines you did write are entirely about **your** problem —
opening a database, building rows, deciding what to display. Compare to
what the framework handled silently:

- A multi-pane layout that resizes with the terminal.
- A scrollable, focusable, mouse-aware list.
- A multi-line text editor with proper cursor, selection, copy/paste.
- A scrollable table with column widths, headers, and cell navigation.
- Tab/Shift-Tab focus traversal across all interactive panes.
- Theme-aware styling for every widget, swappable at runtime.
- A built-in inspector you can pop open with Ctrl-D.

## Where to take it from here

Half a dozen extensions are obvious next steps; pick whichever you'd
actually want to use:

- **Query history** — wrap the editor in a `Switcher` and keep the last
  N queries; flip with Alt-↑ / Alt-↓.
- **Schema inspector** — replace the sidebar `List` with a `Tree`
  (`widgets.Tree`) so each table expands to show its columns.
- **CSV export** — add a Ctrl-E shortcut that writes the current result
  set to a file. The data is already in `[][]string` after `runQuery`.
- **Multi-statement** — split the editor on `;` and run statements in
  sequence. `db.Exec` for non-`SELECT`, `db.Query` for `SELECT`.
- **Dialogs** — confirm DESTRUCTIVE statements (DROP, DELETE) with
  `ui.Confirm(title, msg, onYes, onNo)`.

You now have everything you need to build any of these. The next-step
references in this tutorial:

- The [reference docs](../reference/overview.md) for every widget's API.
- The [showcase app](../../cmd/showcase) for exercises in styling.
- The [design principles](../principles.md) for the small list of
  invariants the framework expects.

→ [Appendix: Debugging](A-debugging.md)
→ [Appendix: Cheatsheet](B-cheatsheet.md)
