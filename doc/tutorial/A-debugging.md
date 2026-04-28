# Appendix A — Debugging

Three tools are built into every zeichenwerk app and worth knowing the
moment you hit a layout you don't understand.

## `.Debug()` — turn on the debug bar

Call `.Debug()` on the `*UI` before `Run()`:

```go
ui := NewBuilder(themes.TokyoNight()).
    /* … */ .
    Build()

ui.Debug()
ui.Run()
```

This adds a one-line debug bar at the bottom showing:

- Current focused widget (ID and type)
- Hovered widget under the mouse
- Mouse coordinates
- Last event key
- Redraw / refresh counters (handy for spotting layout thrash)

The bar updates live; nothing else changes about your UI.

## `Ctrl-D` — the inspector

With `.Debug()` enabled, `Ctrl-D` opens an inspector popup over the live
UI. It's a tree view of every widget in your app:

```
╭─ Inspector ──────────────────────────────╮
│ ▼ ui                                      │
│   ▼ root (VFlex)         24×80   bg=$bg0  │
│     ▶ header (HFlex)      1×80            │
│     ▼ body (Grid)        20×80            │
│       ▶ tables (List)    20×24            │
│       ▼ sql (Editor)     10×56  focused   │
│         …                                 │
╰───────────────────────────────────────────╯
```

You can:

- Expand / collapse nodes to focus on one subtree.
- See bounds for every widget — perfect for spotting "why is this 0
  cells wide?"
- See the resolved style (foreground, background, border, padding) for
  any selected widget.
- See the active flags (focused, hovered, focusable, hidden, …).

Press `Esc` to close the inspector and return to the app.

## Logging

Every widget can call `Log(self, level, message, kvs...)`:

```go
editor.Log(editor, core.Debug, "running query", "sql", q)
```

Logs land in a rotating buffer (`*widgets.TableLog`) the UI keeps
internally. To see them:

- They're rendered live in the debug bar's "last log" entry when
  `.Debug()` is enabled.
- `ui.Logs()` returns the `*TableLog`. You can drop it into your UI
  yourself with `Builder.Add(ui.Logs())` to get a permanent log pane.
- `ui.SetLogLevel(slog.LevelDebug)` changes the threshold at runtime.

Levels in `core`: `Debug`, `Info`, `Warning`, `Error`, `Fatal`.

## Dump

Sometimes you want the widget tree as text — for diffing, for a bug
report, for git. `*UI` has a `Dump` method:

```go
ui.SetBounds(0, 0, 120, 40)
ui.Layout()
ui.Dump(os.Stdout, DumpOptions{Style: true})
```

You'll get an indented tree with bounds, types, and (with `Style: true`)
the resolved styles for every node. The showcase command-line uses
exactly this approach for its `--dump` / `--dump-verbose` flags — see
[`cmd/showcase/main.go`](../../cmd/showcase/main.go).

## Common pitfalls and what they look like

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| Widget doesn't appear | Forgotten `End()` — it landed inside a previous container | Count `End()`s; open inspector to confirm where it is. |
| Whole panel won't grow with the terminal | No fractional track on that axis | Give one column / row `-1`. |
| `MustFind` panics | Wrong ID or wrong type | Search source: `grep '"name"'`. Or untyped `core.Find` and check for `nil`. |
| Tab skips a widget you want focusable | Custom widget didn't set `FlagFocusable`, or `FlagSkip` is on | `widget.SetFlag(core.FlagFocusable, true)` in your constructor; clear `FlagSkip`. |
| `OnChange` handler never fires | Event data is the wrong type (e.g. Checkbox sends `bool`, not `string`) | Use raw `widget.On(EvtChange, …)` and assert `data[0].(bool)`. |
| Theme variable shows up as literal `$cyan` | Renderer used directly without theme wrapping | Always render through `core.Renderer` (you do automatically inside `Render`). |

## When to ask the framework

Most "weird layout" questions are answered by the inspector in five
seconds. Make `.Debug() + Ctrl-D` your first move; the rest follows.
