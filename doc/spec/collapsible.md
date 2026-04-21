# Collapsible

A single-child container with a clickable header that expands and collapses the
body. When collapsed, only the header row is visible and the widget's height hint
shrinks accordingly, causing the parent layout to reclaim the freed space.

## Structure

```go
type Collapsible struct {
    Component
    title    string  // Header label
    child    Widget  // Body content (nil until Add is called)
    expanded bool    // Whether the body is currently visible
}
```

## Constructor

```go
func NewCollapsible(id, class, title string, expanded bool) *Collapsible
```

- Sets `FlagFocusable` — the header itself is keyboard-navigable.
- Registers key and mouse handlers.
- `expanded` sets the initial state; child defaults to `nil`.

## Container interface

`Collapsible` implements `Container`.

| Method | Behaviour |
|--------|-----------|
| `Add(widget, ...)` | Replaces any existing child; sets `child.SetParent(c)`; sets `FlagHidden` on the child when collapsed |
| `Children() []Widget` | Returns `[child]` if child is set, otherwise `[]` |
| `Layout()` | Positions child within the body area when expanded; calls `Layout(c)` to recurse into child containers |

The child is never removed on collapse — it retains its state. `FlagHidden` is
the only toggle, consistent with how `Switcher` handles invisible panes.

## Expanding and collapsing

```go
func (c *Collapsible) Expand()
func (c *Collapsible) Collapse()
func (c *Collapsible) Toggle()
func (c *Collapsible) Expanded() bool
```

All three mutating methods follow the same sequence:

1. Set `c.expanded`.
2. If child is set: `child.SetFlag(FlagHidden, !c.expanded)`.
3. Call `Relayout(c)` — a new package-level helper (see Implementation Plan)
   that walks up the parent chain to the root container and re-runs layout
   top-down, then calls `Redraw` on the root. This ensures the parent
   reclaims or allocates space for the body.
4. Dispatch `EvtChange` with `c.expanded` as data.

## Hint

```go
func (c *Collapsible) Hint() (int, int)
```

The hint drives the parent layout's space allocation:

- **Collapsed**: `(childW, headerH)` — header row only (1 character tall before
  style overhead).
- **Expanded**: `(childW, headerH + childH)` — header plus child.

Where `childW` and `childH` come from `child.Hint()` (or 0 if no child), and
`headerH = 1`. Style overhead (border, padding, margin) is added by the parent's
layout engine from the widget's own style, consistent with how `Box.Hint()` works.

## Layout

```go
func (c *Collapsible) Layout() error
```

1. Compute content area `(cx, cy, cw, ch)` from `c.Content()`.
2. Header occupies `(cx, cy, cw, 1)` — this is only used for mouse hit-testing
   and focus area; it is not a sub-widget.
3. If expanded and child is set: `child.SetBounds(cx, cy+1, cw, ch-1)`.
4. Call `Layout(c)` to recurse into child containers.

When collapsed, child bounds are not updated; the child retains its last known
bounds but is hidden.

## Keyboard interaction

`Collapsible` is focusable so the header row can receive keyboard input without
requiring the user to click.

| Key | Behaviour |
|-----|-----------|
| `Enter` / `Space` | `Toggle()` |
| `→` | `Expand()` if collapsed, no-op if expanded |
| `←` | `Collapse()` if expanded, no-op if collapsed |

## Mouse interaction

A mouse click anywhere on the header row (y == `c.y + style.Margin().Top`) calls
`Toggle()`. Clicks in the body area are routed to the child widget as normal.

## Render

```go
func (c *Collapsible) Render(r *Renderer)
```

1. `c.Component.Render(r)` — draws background and border.
2. Resolve header style (`"collapsible/header"` + focused/hovered state).
3. Draw indicator (`▼` or `▶`) and title in the header row.
4. If expanded and child is set: `child.Render(r)`.

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"change"` | `bool` | Expanded state changed; `true` = now expanded |

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"collapsible"` | Outer widget (background, border) |
| `"collapsible/header"` | Header row text and background |
| `"collapsible/header:focused"` | Header row when widget is focused |
| `"collapsible/header:hovered"` | Header row on mouse hover |

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `collapsible.expanded` | `▼ ` | Indicator for expanded state |
| `collapsible.collapsed` | `▶ ` | Indicator for collapsed state |

## Builder usage

```go
builder.
    Collapsible("section", "Details", false).
        List("items", "One", "Two", "Three").
    End()
```

`Collapsible` is pushed onto the builder's container stack (like `Box` and
`Flex`), so `End()` closes it and subsequent widgets are added to the parent.

## Implementation Plan

1. **`helper.go`** — add `Relayout(widget Widget)` package-level helper
   - Walks up `widget.Parent()` to find the root container.
   - Calls `root.Layout()` to recompute bounds top-down.
   - Calls `Redraw(root)` to queue a repaint.
   - This helper will also be useful for any future widget that changes its
     own size at runtime.

2. **`collapsible.go`** — new file
   - Define `Collapsible` struct and `NewCollapsible`.
   - Implement `Add`, `Children`, `Layout`, `Hint`, `Apply`, `Render`.
   - Implement `Expand`, `Collapse`, `Toggle`, `Expanded`.
   - Implement `handleKey` and `handleMouse`.

3. **`builder.go`** — add `Collapsible` method
   ```go
   func (b *Builder) Collapsible(id, title string, expanded bool) *Builder
   ```

4. **Theme** — add `"collapsible/header"` style entry and `collapsible.*`
   string keys to built-in themes.

5. **Tests** — `collapsible_test.go`
   - `Hint()` returns header-only height when collapsed, full height when
     expanded.
   - `Toggle()` sets `FlagHidden` on the child correctly.
   - `EvtChange` carries the new expanded state.
   - `←` / `→` keys collapse and expand respectively.
   - `Layout()` positions child at `y+1` within the content area.
   - Collapsing and re-expanding restores the child to the same state.
