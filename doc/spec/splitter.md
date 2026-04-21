# Splitter

A thin, draggable handle widget placed between two siblings in a `Flex`
container. Dragging it adjusts the size hints of its two neighbours, causing
the parent to re-layout and redistribute space. Keyboard nudging (arrow keys)
is also supported. No changes to the `Container` API are required — only
`SetHint` calls on the adjacent children.

## Placement

`Splitter` is inserted as an explicit child between two panes in a `Flex`:

```go
NewBuilder(theme).
    Flex("root", true, "stretch", 0).   // horizontal flex
        Editor("left").
        Splitter("split", "").
        List("right", items...).
    End()
```

In a **horizontal** flex (children side by side) the splitter occupies a single
column and is dragged left/right. In a **vertical** flex (children stacked) it
occupies a single row and is dragged up/down.

## Structure

```go
type Splitter struct {
    Component
    minSize   int  // Minimum size enforced on each neighbour (default 1)
    dragging  bool // Drag in progress
    dragOrigin int // Mouse coordinate at drag start (x or y depending on orientation)
    dragHintA  int // Left/top neighbour's hint at drag start
    dragHintB  int // Right/bottom neighbour's hint at drag start
}
```

## Constructor

```go
func NewSplitter(id, class string) *Splitter
```

- `minSize = 1`.
- Sets `FlagFocusable` so the handle can receive keyboard input.
- Registers key and mouse handlers.

## Orientation

Inferred from the splitter's own bounds after layout — no stored flag needed:

- `width == 1` → the parent Flex is **horizontal**; drag axis is X.
- `height == 1` → the parent Flex is **vertical**; drag axis is Y.

This is always correct because the parent Flex gives the splitter exactly
`(1, fullHeight)` or `(fullWidth, 1)` during layout.

## Hint

Returns `(1, 1)`. The Flex stretches it to full cross-axis size via the
`"stretch"` alignment applied to all children. The 1-unit main-axis hint
reserves exactly one column/row for the handle.

## Finding neighbours

```go
func (s *Splitter) neighbours() (before Widget, after Widget)
```

Walks `s.Parent().Children()`, finds the index of `s`, and returns
`children[i-1]` and `children[i+1]`. Returns `nil, nil` if the parent is not
a `Flex` or if there are no neighbours on one side.

## Mouse interaction

### Drag start — `Button1` press on the splitter

1. Record `dragOrigin` = mouse X (horizontal) or mouse Y (vertical).
2. Record `dragHintA`, `dragHintB` from `before.Hint()` and `after.Hint()`.
   If either hint is negative (fractional), convert it to the widget's current
   rendered width/height (from `Bounds()`) so drag arithmetic is always in
   fixed units.
3. Set `dragging = true`.
4. Call `ui.Capture(s)` — see *Mouse capture* below.

### Drag move — `ButtonNone` (no button) while `dragging`

1. Compute `delta` = current drag-axis coordinate − `dragOrigin`.
2. `newA = clamp(dragHintA + delta, minSize, dragHintA + dragHintB - minSize)`.
3. `newB = dragHintA + dragHintB - newA`.
4. Call `before.SetHint(newA, beforeHintH)` and `after.SetHint(newB, afterHintH)`,
   preserving the cross-axis hint component.
5. Call `Relayout(s)` to propagate the new hints through the parent Flex.

### Drag end — mouse button released while `dragging`

1. Set `dragging = false`.
2. Call `ui.Release(s)`.
3. Dispatch `EvtChange` with the new split position (fixed-axis coordinate of
   the splitter after layout) as `int` data.

## Mouse capture

Dragging requires receiving `EvtMouse` even when the cursor moves beyond the
splitter's 1-column/1-row area into a neighbouring pane. This requires a
UI-level capture mechanism:

**New fields on `UI`:**

```go
capture Widget  // receives all EvtMouse events when non-nil
```

**New methods on `UI`:**

```go
func (ui *UI) Capture(w Widget)   // sets ui.capture = w
func (ui *UI) Release(w Widget)   // clears ui.capture if w matches
```

**Change to mouse dispatch in `ui.go`:**

```go
case *tcell.EventMouse:
    at := FindAt(ui.layers[len(ui.layers)-1], mx, my)
    // hover update as before ...
    target := at
    if ui.capture != nil {
        target = ui.capture
    }
    ui.dispatch(target, EvtMouse, event)
```

Hover tracking continues to follow the physical mouse position regardless of
capture, so other widgets still show hover state during a drag.

## Keyboard interaction

When the splitter has focus, arrow keys nudge the split by one unit at a time.

| Key | Behaviour |
|-----|-----------|
| `←` / `↑` | Move splitter one unit toward `before` |
| `→` / `↓` | Move splitter one unit toward `after` |

Which pair of keys is active depends on orientation: `←`/`→` in a horizontal
flex, `↑`/`↓` in a vertical flex. The other pair is ignored (returns `false`).

Each nudge: read current hints of neighbours, apply ±1 delta with `minSize`
clamping, call `SetHint` on both, call `Relayout(s)`, dispatch `EvtChange`.

## Rendering

```go
func (s *Splitter) Render(r *Renderer)
```

1. `s.Component.Render(r)` — fills the 1×N or N×1 background.
2. Determine orientation from bounds.
3. Draw the handle character along the full length using `r.Repeat`:
   - Vertical bar (horizontal flex): `│`
   - Horizontal bar (vertical flex): `─`
4. When `FlagFocused` or `FlagHovered`: use `"splitter:focused"` /
   `"splitter:hovered"` style (typically a brighter colour).
5. Draw a centre grip marker at the midpoint:
   - Vertical: `╪` (or `┿`)
   - Horizontal: `╫` (or `┿`)
   This gives the user a visible drag target, especially in tall/wide splits.

Handle characters are fetched from the theme via `theme.String("splitter.*")`
keys, consistent with `Tree` and `Select`.

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `int` | Drag or nudge completed; data is the new absolute split position |

`EvtChange` fires on mouse-up and on each keyboard nudge. It does not fire
continuously during a drag — callers that need live feedback can observe
`EvtChange` on the neighbouring panes' `Bounds()` after `Relayout`.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"splitter"` | Handle background and foreground |
| `"splitter:focused"` | Handle when keyboard-focused |
| `"splitter:hovered"` | Handle on mouse hover |

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `splitter.vertical` | `│` | Character for vertical bar (horizontal flex) |
| `splitter.horizontal` | `─` | Character for horizontal bar (vertical flex) |
| `splitter.grip.vertical` | `╪` | Centre grip for vertical bar |
| `splitter.grip.horizontal` | `╫` | Centre grip for horizontal bar |

## Implementation plan

1. **`ui.go`** — add mouse capture
   - Add `capture Widget` field to `UI`.
   - Implement `Capture(w Widget)` and `Release(w Widget)`.
   - Route `EvtMouse` to `capture` when set, while keeping hover tracking
     unconditional.

2. **`helper.go`** — `Relayout`
   - If not already added (required by `Collapsible`): implement
     `Relayout(widget Widget)` walking up to the root and calling
     `root.Layout()` followed by `Redraw(root)`.

3. **`splitter.go`** — new file
   - Define `Splitter` struct and `NewSplitter`.
   - Implement `neighbours`, `Hint`, `Apply`, `Render`.
   - Implement `handleKey` and `handleMouse` with drag start/move/end logic.

4. **`flex.go`** — expose `Horizontal()`
   - Add `func (f *Flex) Horizontal() bool` so external code can check
     orientation without relying on bounds inference. Used by the drag
     logic to know which hint dimension (width vs height) to adjust on
     neighbours.

5. **`builder.go`** — add `Splitter` method
   ```go
   func (b *Builder) Splitter(id string) *Builder
   ```

6. **Theme** — add `"splitter"` style entry and `splitter.*` string keys to
   built-in themes.

7. **Tests** — `splitter_test.go`
   - `neighbours()` returns correct before/after widgets in a Flex.
   - Drag delta correctly clamps to `minSize` on both sides.
   - Fractional hints are converted to fixed before drag starts.
   - `EvtChange` fires on mouse-up with the correct position.
   - Keyboard nudge adjusts hints by exactly 1 and dispatches `EvtChange`.
   - Orientation is inferred correctly from a 1×N vs N×1 bounds.
