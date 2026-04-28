# 3. Layout

Two containers cover 95 % of TUI layouts: **Flex** for linear stacks,
**Grid** for tables and dashboards. The third concept — **`Hint(w, h)`** —
controls how each child negotiates for space.

## Hint values

`.Hint(width, height)` sets a widget's preferred size. Each axis follows
the same three-case rule:

| Value | Meaning | Example |
|-------|---------|---------|
| `>0` | **Fixed** — exactly this many cells | `.Hint(20, 1)` — exactly 20 cells wide, 1 tall |
| `<0` | **Fractional** — share remaining space proportional to magnitude | `.Hint(-1, 0)` — gets 1× share; sibling with `-2` gets 2× |
| `0` | **Auto** — ask the widget for its preferred size | `.Hint(0, 0)` — Static's natural width = its text length |

Run [`examples/03-layout-hints`](examples/03-layout-hints/main.go) to see
all three side-by-side. The key file:

```go
HFlex("row1", core.Stretch, 1).Hint(0, 1).
    Static("a", "fixed 10").Hint(10, 0).
    Static("b", "fixed 20").Hint(20, 0).
    Static("c", "auto").                // no Hint → auto
End().

HFlex("row2", core.Stretch, 1).Hint(0, 1).
    Static("d", "weight -1").Hint(-1, 0).
    Static("e", "weight -2").Hint(-2, 0).
    Static("f", "weight -3").Hint(-3, 0).  // gets half of the row
End().
```

The mixed case is common and works the way you'd hope: fixed children take
their cells first, then fractional children divide the leftover space by
weight.

## Flex

```go
HFlex(id, alignment, spacing)   // children laid out left-to-right
VFlex(id, alignment, spacing)   // children laid out top-to-bottom
```

- **`alignment`** controls the *cross-axis* (vertical for HFlex, horizontal
  for VFlex). Values from `core`: `Start`, `End`, `Center`, `Stretch`,
  `Left`, `Right`, `Default`.
- **`spacing`** is the gap (in cells) between consecutive children.

A typical app shell:

```go
VFlex("root", core.Stretch, 0).
    HFlex("header", core.Center, 0).Hint(0, 1).Background("$bg2").
        Static("title", "App Name").Font("bold").
    End().
    HFlex("body", core.Stretch, 0).Hint(0, -1).         // -1 height → fill
        Static("sidebar", "…").Hint(24, 0).             // fixed 24 wide
        Static("content", "…").Hint(-1, 0).             // fill rest
    End().
    HFlex("footer", core.End, 2).Hint(0, 1).
        Static("hint1", "[q] quit").
    End().
End()
```

Full version: [`examples/03-layout-flex`](examples/03-layout-flex/main.go).

### Common Flex mistake: forgetting `End()`

`End()` pops the current container off the builder stack. Skip it once and
the next widget you add lands as a *child* of the previous container
instead of a sibling. The compiler can't catch this — the symptom is a
wonky layout. When something nests where you didn't expect, count your
`End()`s first.

## Grid

```go
Grid(id, rows, columns, lines).
    Columns(c1, c2, …).      // size of each column
    Rows(r1, r2, …).         // size of each row
    Cell(x, y, w, h).Static(…).   // place next widget
    Cell(…).Button(…).
End()
```

- **`rows`, `columns`** are the number of tracks. They must match what you
  pass to `Rows()` and `Columns()`.
- **`lines = true`** draws thin separators between cells.
- **`Cell(x, y, w, h)`** — column index `x`, row index `y`, spanning `w`
  columns and `h` rows. Call it before *each* widget you place.

Track sizes follow the same Hint rules: positive = fixed, negative =
fractional, zero = auto. Mix freely:

```go
Columns(20, -1, -2)   // 20 fixed | 1× share | 2× share
Rows(3, -1, 1)        // 3 fixed  | fill     | 1 fixed
```

Full example with row/column spanning:
[`examples/03-layout-grid`](examples/03-layout-grid/main.go).

### Grid pitfall: no fractional track

If every column is fixed and the terminal is wider, the leftover space sits
unused on the right. Same for rows. The fix is to give at least one track
on each axis a negative weight (commonly `-1`).

## Margin, border, padding

The CSS-like model. Order from outside to inside is:

```
┌── margin ─────────────────────────────────┐
│  ┌── border ───────────────────────────┐  │
│  │  ┌── padding ─────────────────────┐ │  │
│  │  │                                 │ │  │
│  │  │       content                   │ │  │
│  │  │                                 │ │  │
│  │  └─────────────────────────────────┘ │  │
│  └─────────────────────────────────────┘  │
└───────────────────────────────────────────┘
```

`Hint(w, h)` describes the **content** box — margin, border, and padding
are *added on top* by the framework. So `.Hint(20, 1).Padding(0, 2)` claims
24 cells of horizontal space (20 + 2×2 padding).

Both `.Padding(...)` and `.Margin(...)` accept 1-4 ints, CSS-style:

```go
.Padding(1)           // 1 on all four sides
.Padding(1, 2)        // top/bottom = 1, left/right = 2
.Padding(1, 2, 3)     // top = 1, left/right = 2, bottom = 3
.Padding(1, 2, 3, 4)  // top, right, bottom, left
```

`.Border(style)` accepts the names registered by your theme: `"none"`,
`"thin"`, `"thick"`, `"round"`, `"double"`, `"lines"`. The selector form
`.Border(":focus", "double")` applies only when the widget is focused.

## Spacers and rules

Two cosmetic helpers worth knowing early:

- **`Spacer()`** — an invisible flexible widget. In an HFlex, drop one
  between two groups to push them to opposite sides.
- **`HRule(style)` / `VRule(style)`** — a 1-cell separator line.

```go
HFlex("toolbar", core.Start, 1).
    Button("save", "Save").
    Button("open", "Open").
    Spacer().
    Button("quit", "Quit").
End()
```

## Try it

```bash
go run ./doc/tutorial/examples/03-layout-flex
go run ./doc/tutorial/examples/03-layout-grid
go run ./doc/tutorial/examples/03-layout-hints
```

Resize your terminal while each one runs — fractional tracks redistribute,
fixed ones don't.

→ [Next: Styling & themes](04-styling-themes.md)
