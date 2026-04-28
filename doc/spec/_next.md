# Widget backlog

Candidate widgets for future implementation, grouped by category.

---

## Data visualisation

### Line chart

The natural companion to `BarChart` and `Sparkline`. Multiple series as colored
lines over a shared time or category x-axis, with optional area fill below each
line. Sub-character vertical precision via Braille dots (same technique as
`Gauge`) gives smooth curves at any height. Supports a ring-buffer streaming
mode where new values scroll the x-axis left automatically — useful for live
dashboards. Shares `ScaleMode` and axis-layout logic with `BarChart`.

Selectors: `"line-chart"`, `"line-chart/s0"` … `"line-chart/s7"`,
`"line-chart/axis"`, `"line-chart/grid"`, `"line-chart/fill"`.

### Treemap

Nested rectangles sized proportionally to a value hierarchy (e.g. disk usage
by directory, module sizes, budget breakdown). Laid out with the squarified
algorithm. Each rectangle is drawn with a background color from a per-depth
palette and a truncated label. Selection navigates into sub-trees; `Esc` zooms
back out. A natural companion to `Tree` — the same `TreeNode` data feeds both.

Selectors: `"treemap"`, `"treemap/node"`, `"treemap/node:focused"`,
`"treemap/node:selected"`, with a `"treemap.depth"` theme variable for depth-
based colour interpolation (same pattern as `"heatmap.intensity"`).

### Donut / ring progress

A full-circle ring (not a semicircle like `Gauge`) drawn with Braille
characters. Can show a single value as an arc, or multiple concentric rings for
a hierarchy, or multiple arcs on the same ring for a stacked proportion. Works
at small sizes (as little as 5×5). Useful as a compact KPI widget or as a
loading indicator when animated.

Selectors: `"donut"`, `"donut/s0"` … `"donut/s7"`, `"donut/track"`,
`"donut/label"`.

### Flame graph

A horizontally stacked call-tree chart. Each row is a call stack depth; each
cell is a frame sized by its sample count. The root spans the full width; child
frames subdivide their parent. Clicking or navigating a frame zooms it to fill
the width. Color encodes the module or package. Very recognizable for profiling
output; pairs naturally with `Tree` data.

Selectors: `"flame-graph"`, `"flame-graph/frame"`, `"flame-graph/frame:focused"`,
`"flame-graph/frame:selected"`.

---

## Input

### Slider

A horizontal (or vertical) range input:

```
Min ─────────●──────── Max
             42
```

The thumb `●` moves with arrow keys (coarse step) and `Shift+Arrow` (fine step).
Mouse click on the track jumps to that position; drag moves the thumb. Dispatches
`EvtChange(float64)`. Works with `Value[float64]` binding. Implements `Setter`.

Selectors: `"slider"`, `"slider/track"`, `"slider/thumb"`,
`"slider/thumb:focused"`.

### Tag input

Inline labelled chips for entering a set of string values:

```
┌─────────────────────────────────┐
│ [Go ×] [Rust ×] [Python ×] ▌   │
└─────────────────────────────────┘
```

Each confirmed value is rendered as a styled chip with a remove button. Comma,
Tab, or Enter confirms the current word as a new tag. Backspace with an empty
cursor removes the last tag. Wraps to multiple lines when tags exceed the widget
width. Dispatches `EvtChange([]string)`. Implements `Filterable` so it can pair
with a `Filter` for suggestion.

Selectors: `"tag-input"`, `"tag-input/chip"`, `"tag-input/chip:hovered"`,
`"tag-input/remove"`.

### Rating

A row of discrete symbol characters representing a score:

```
★ ★ ★ ☆ ☆
```

Arrow keys and mouse clicks change the value. The symbol, filled form, and empty
form are theme strings. Dispatches `EvtChange(int)`. Also usable as a read-only
priority or severity indicator when `FlagReadonly` is set.

Theme strings: `"rating.filled"` (`★`), `"rating.empty"` (`☆`),
`"rating.half"` (`⯨`) for half-star mode.

### Color picker

An interactive color selector combining three zones:

```
  ┌──────────────┐  ┌──┐
  │  saturation  │  │  │  hue
  │   /value     │  │  │  strip
  │    gradient  │  │  │
  └──────────────┘  └──┘
  #1a1b26   R: 26  G: 27  B: 38
```

The main gradient panel (drawn with half-block characters and background colors)
selects saturation (x) and value (y) within the current hue. The hue strip on
the right selects hue. A hex input below allows direct entry. Arrow keys move
the crosshair in the active zone; Tab switches between zones. Dispatches
`EvtChange(Color)` on every movement, `EvtActivate` on Enter/click.

In compact mode (`SetCompact(true)`) only the hue strip and hex field are shown —
useful when embedded inside a `Form`.

Selectors: `"color-picker"`, `"color-picker/gradient"`,
`"color-picker/hue"`, `"color-picker/hex"`, `"color-picker/hex:focused"`.

---

## Layout and navigation

### Status bar

A single-row container with named slots that each have an independent style and
update independently without reflowing neighbours:

```
 NORMAL │ main.go │ ln 42, col 8 │ UTF-8 │ Go │ 14:23 
```

Three alignment zones — left, centre, right — each holding an ordered list of
slots. Each slot is a `(key, text)` pair addressed by key; `Set(key, text)` 
updates one slot and redraws. Unlike a `Flex`, slots do not participate in 
fractional sizing; they shrink from the right when space is insufficient.

Selectors: `"status-bar"`, `"status-bar/slot"`, plus per-key selectors
`"status-bar/slot.mode"` etc. for slot-specific styling.

### Accordion

A vertical stack of `Collapsible` panels where expanding one automatically
collapses the previously open panel. The grouping contract (at most one open)
lives in the `Accordion` container, not in the individual panels. Supports
`ExpandAll()` / `CollapseAll()` escape hatches for non-exclusive mode.
Dispatches `EvtChange(int)` with the newly expanded panel index.

Constructed by adding `Collapsible` children in the Builder:

```go
builder.Accordion("settings", "").
    Collapsible("network", "Network", true).
        // … children
    End().
    Collapsible("display", "Display", false).
        // … children
    End().
End()
```

### Drawer

A panel that slides in from a screen edge (`Grow` animation) and overlays the
existing UI as a popup layer. Has a configurable edge (`Left`, `Right`,
`Bottom`), a title, and a `Toggle()` method. The backdrop outside the drawer
can optionally dim the rest of the UI. Dispatches `EvtShow` / `EvtHide`.

```go
drawer := NewDrawer("nav", "", DrawerLeft, "Navigation")
ui.AddDrawer(drawer)
drawer.Toggle()
```

---

## Text and display

### Diff viewer

Side-by-side or unified diff with standard coloring (`+` green, `-` red,
`@@` cyan) and line numbers. Parses unified diff format (`[]string` of lines).
Wraps a `Viewport` internally for scrolling. Unchanged context lines are dimmed.
In side-by-side mode, changed lines are shown aligned on left and right panels
with intra-line highlighting of the changed words.

Selectors: `"diff"`, `"diff/added"`, `"diff/removed"`, `"diff/hunk"`,
`"diff/context"`, `"diff/line-number"`, `"diff/changed-word"`.

### Clock

A live display built on the existing `Digits` widget. Shows the current time
(or a caller-supplied `time.Time`) in large box-drawing characters, updating
every second via `Animation`. Format is configurable (`HH:MM`, `HH:MM:SS`,
`HH:MM:SS.ms` for a stopwatch). Can also count up or down from a given time
(stopwatch / countdown mode). Dispatches `EvtActivate` when a countdown
reaches zero.

Selectors: `"clock"`, `"clock/digits"`, `"clock/separator"`.

---

## Animation and effects

### Glitter

Like `Shimmer`, but individual characters twinkle independently at random
intervals rather than following a sweeping band. Each character cycles through
a small palette of accent colors with a random phase offset, producing a
sparkling effect. The density (fraction of characters twinkling at any moment)
and palette are configurable.

Suitable for celebratory moments, splash screens, or ambient decoration.
Can be applied to any static text — it is a drop-in replacement for `Static`
with animation.

```go
gl := NewGlitter("title", "").
    SetText("  zeichenwerk  ").
    SetDensity(0.3).
    SetPalette("$blue", "$cyan", "$purple", "$fg0")
gl.Start(80 * time.Millisecond)
```

Selectors: `"glitter"`. Per-character color is controlled by the palette slice,
not by sub-part selectors.

---

## Compound widgets

### Kanban column

A `Deck`-backed scrollable column of multi-line cards. Each card has a title,
optional body text, and an optional color label strip on the left edge. Cards
can be reordered within the column via `Ctrl+Up` / `Ctrl+Down`. A horizontal
`Flex` of columns forms a full board. Dispatches `EvtChange` when card order
changes, `EvtActivate` when a card is opened.

Selectors: `"kanban-column"`, `"kanban-card"`, `"kanban-card:focused"`,
`"kanban-card/label"` (colored left strip), `"kanban-card/title"`,
`"kanban-card/body"`.
