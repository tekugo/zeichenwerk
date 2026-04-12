# Breadcrumb

A single-row path indicator that displays an ordered list of string segments
separated by a configurable separator. Each segment is individually focusable
and clickable. When the total rendered width exceeds the available space, leading
segments are collapsed to an overflow marker (`…`) from the left, always keeping
the focused segment visible. Useful for file browsers, settings hierarchies, and
drill-down navigation.

---

## Visual layout

**All segments fit (focused = "zeichenwerk"):**

```
Home › Projects › [zeichenwerk] › spec
```

**Overflowing — focus on rightmost segment:**

```
… › zeichenwerk › spec
```

**Overflowing — user navigated left to a collapsed segment:**

```
… › Projects › zeichenwerk › spec
```

The `[…]` brackets denote the focused segment styling; they are not drawn
literally. The overflow marker `…` appears with the separator appended
(`… › `) and is not focusable.

---

## Structure

```go
type Breadcrumb struct {
    Component
    segments  []string // ordered list of path segments
    selected  int      // index of the focused segment (-1 = none)
    first     int      // index of the first segment shown (0 = no overflow)
    // Characters read from theme strings in Apply.
    separator string   // separator between segments, default " › "
    overflow  string   // collapse marker, default "…"
}
```

---

## Constructor

```go
func NewBreadcrumb(id, class string) *Breadcrumb
```

Defaults:

- `selected = -1`, `first = 0`.
- `separator = " › "`, `overflow = "…"`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

---

## Methods

### Data

| Method | Description |
|--------|-------------|
| `SetSegments(segs []string)` | Replaces all segments; resets `first = 0`; clamps `selected`; calls `Refresh()` |
| `Push(seg string)` | Appends one segment; calls `Refresh()` |
| `Pop() string` | Removes and returns the last segment; clamps `selected`; calls `Refresh()` |
| `Truncate(index int)` | Removes all segments after `index` (inclusive truncation at `index+1`); clamps `selected`; calls `Refresh()` |
| `Segments() []string` | Returns the current segments slice |

### Navigation

| Method | Description |
|--------|-------------|
| `Select(index int)` | Focuses segment at `index`; clamps to valid range; ensures it is visible; dispatches `EvtSelect` |
| `Selected() int` | Returns the currently focused segment index (-1 if none) |

### Display

| Method | Description |
|--------|-------------|
| `SetSeparator(sep string)` | Overrides the separator string; calls `Refresh()` |
| `SetOverflow(marker string)` | Overrides the overflow marker; calls `Refresh()` |

---

## Hint

```go
func (c *Breadcrumb) Hint() (int, int)
```

- **Width**: natural width of all segments joined by the separator, plus style
  horizontal overhead. Returns 0 (fill parent) when no segments are set.
- **Height**: always 1 (plus style vertical overhead).

The natural width is:

```
sum(runeLen(seg) for seg in segments)
    + runeLen(sep) * (len(segments) - 1)
    + style.Horizontal()
```

where `runeLen` counts display columns (accounting for multi-byte runes).

---

## Overflow and visibility

The breadcrumb always fits within its rendered width by adjusting `first`.
The rule is applied at render time (not stored between renders) via
`computeFirstVis`:

```
computeFirstVis(availW int) int:
    start = max(c.first, 0)   // never move right on its own
    loop:
        width = renderWidth(start, availW)
        if width <= availW: return start
        start++
        if start >= len(segments)-1: return start  // always show at least one
```

`renderWidth(start, availW)`:

```
if start > 0: w = runeLen(overflow) + runeLen(separator)
else: w = 0
for i, seg in segments[start:]:
    w += runeLen(seg)
    if i < len(segments[start:])-1: w += runeLen(separator)
return w
```

`first` is updated (never increased past `selected`) in `Select` to guarantee
the newly focused segment is not hidden:

```go
if index < c.first {
    c.first = index
}
```

---

## Render

```go
func (c *Breadcrumb) Render(r *Renderer)
```

1. `c.Component.Render(r)` — draws background and border.
2. Obtain `(cx, cy, cw, _)` from `c.Content()`.
3. Compute `start = computeFirstVis(cw)`.
4. Draw the overflow prefix when `start > 0`:
   - `overflow` with the `"breadcrumb/separator"` style.
   - `separator` with the `"breadcrumb/separator"` style.
5. For each segment `i` from `start` to `len(segments)-1`:
   - Choose style: `"breadcrumb/segment:focused"` when `i == selected`,
     otherwise `"breadcrumb/segment"`.
   - Draw the segment text, truncated to remaining width.
   - If not the last segment: draw `separator` with `"breadcrumb/separator"` style.
6. Stop drawing if the cursor x position reaches `cx + cw`.

The render pass is strictly left-to-right; no wrapping occurs (the widget is
always one row tall).

---

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `←` | `Select(selected - 1)`; wraps to last when at first |
| `→` | `Select(selected + 1)`; wraps to first when at last |
| `Home` | `Select(0)` |
| `End` | `Select(len(segments) - 1)` |
| `Enter` | Dispatch `EvtActivate` with `selected` as data |

When the widget gains focus (`EvtFocus`) and `selected == -1`, focus is set to
the last segment automatically (the most common case for path navigation is
starting at the leaf).

---

## Mouse interaction

A click anywhere in the content row hits one of the rendered segments. The hit
is resolved by replaying the render geometry without actually drawing:

```
for i, seg in visibleSegments:
    segEnd = segStart + runeLen(seg)
    if mx >= segStart and mx < segEnd:
        Select(i)
        if i == previously selected: Dispatch(EvtActivate, i)
        break
    segStart = segEnd + runeLen(separator)
```

The overflow marker is not clickable. Clicks on it or on separator text between
segments are ignored.

A second click on the already-selected segment dispatches `EvtActivate`.

---

## Events

| Event | Data | Description |
|-------|------|-------------|
| `EvtSelect` | `int` | Focused segment index changed |
| `EvtActivate` | `int` | Enter pressed, or segment clicked twice |

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"breadcrumb"` | Widget background and border |
| `"breadcrumb/segment"` | Unfocused segment text |
| `"breadcrumb/segment:focused"` | Focused segment text |
| `"breadcrumb/separator"` | Separator text and overflow marker `…` |

Example theme entries (Tokyo Night):

```go
NewStyle("breadcrumb").WithColors("$fg0", "$bg0"),
NewStyle("breadcrumb/segment").WithColors("$fg1", "$bg0"),
NewStyle("breadcrumb/segment:focused").WithColors("$bg0", "$blue").WithFont("bold"),
NewStyle("breadcrumb/separator").WithColors("$fg2", "$bg0"),
```

---

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `"breadcrumb.separator"` | `" › "` | Text drawn between segments and after the overflow marker |
| `"breadcrumb.overflow"` | `"…"` | Marker replacing hidden leading segments |

---

## Builder usage

```go
builder.Breadcrumb("path")

bc := builder.Find("path").(*Breadcrumb)
bc.SetSegments([]string{"Home", "Projects", "zeichenwerk", "spec"})
bc.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
    index := data[0].(int)
    bc.Truncate(index)
    return true
})
```

Activating a segment (clicking it or pressing Enter on it) typically truncates
the path to that point, making the breadcrumb behave like a drill-up control.

---

## Implementation plan

1. **`breadcrumb.go`** — new file
   - Define `Breadcrumb` struct and `NewBreadcrumb`.
   - Implement data methods: `SetSegments`, `Push`, `Pop`, `Truncate`,
     `Segments`.
   - Implement navigation: `Select`, `Selected`.
   - Implement display setters: `SetSeparator`, `SetOverflow`.
   - Implement `computeFirstVis(availW int) int` as a private method.
   - Implement `Apply(t *Theme)`: register the four style selectors; read
     `"breadcrumb.separator"` and `"breadcrumb.overflow"` string keys.
   - Implement `Hint() (int, int)`.
   - Implement `Render(r *Renderer)`.
   - Implement `handleKey` and `handleMouse` (including double-click detection
     for `EvtActivate`).
   - On `EvtFocus`: if `selected == -1` and segments are non-empty, call
     `Select(len(segments) - 1)`.

2. **`builder.go`** — add `Breadcrumb` method
   ```go
   func (b *Builder) Breadcrumb(id string) *Builder
   ```

3. **Theme** — add `"breadcrumb"`, `"breadcrumb/segment"`,
   `"breadcrumb/segment:focused"`, and `"breadcrumb/separator"` styles, plus
   `"breadcrumb.separator"` and `"breadcrumb.overflow"` string keys to all
   built-in themes.

4. **`cmd/demo/main.go`** — add a `"Breadcrumb"` entry with `breadcrumbDemo`,
   showing a file-system-style path with `Push`/`Pop` controls and an
   `EvtActivate` handler that truncates the path on segment click.

5. **Tests** — `breadcrumb_test.go`
   - `Hint()` returns correct natural width for a multi-segment path.
   - `computeFirstVis` collapses the minimum number of segments needed to fit.
   - `computeFirstVis` always leaves at least one segment visible.
   - `Select` on a collapsed segment decreases `first` to reveal it.
   - `Push` / `Pop` / `Truncate` clamp `selected` correctly.
   - `EvtActivate` fires on Enter and on a repeated click.
   - `EvtFocus` auto-selects the last segment when `selected == -1`.
   - Separator and overflow marker characters are drawn in the correct style.
