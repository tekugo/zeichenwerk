# Drawer

A panel that slides in from a screen edge and overlays the existing UI as a
popup layer. The Drawer is a container: any widgets can be added to its content
area just like any other container. An optional title bar with a close indicator
appears at the top of the panel. The backdrop behind the drawer can optionally
be dimmed to visually separate it from the underlying content.

---

## Visual layout

**Left edge, 30 columns wide:**

```
╔═══════════════════════════╗────────────────────────────────┐
║ Navigation               ✕║ Search: [ filter...          ] │
╠═══════════════════════════╣                                 │
║ › Home                    ║  Name        Size   Modified   │
║   About                   ║  README.md   4 KB   2 days ago │
║   Settings                ║  src/        —      1 hour ago │
║   Help                    ║  tests/      —      3 days ago │
║                           ║                                 │
╚═══════════════════════════╝─────────────────────────────────┘
```

**Right edge, 28 columns wide:**

```
┌─────────────────────────────────╔══════════════════════════╗
│ Results                         ║ Details                 ✕║
│                                 ╠══════════════════════════╣
│ › README.md                     ║ Path:                    ║
│   src/                          ║ README.md                ║
│   tests/                        ║                          ║
│                                 ║ Size:  4 KB              ║
│                                 ║ Modified: 2 days ago     ║
└─────────────────────────────────╚══════════════════════════╝
```

**Bottom edge, 5 rows tall:**

```
┌──────────────────────────────────────────────────────────────┐
│ Table                                                        │
│ Name         Type    Modified                                │
│ README.md    file    2 days ago                              │
╔══════════════════════════════════════════════════════════════╗
║ Filters                                                     ✕║
╠══════════════════════════════════════════════════════════════╣
║ Status: Active  ▼    Type: All  ▼    From: ──────────────    ║
╚══════════════════════════════════════════════════════════════╝
```

**With dim backdrop (`dim = true`):**

The cells outside the drawer bounds are painted with the `"drawer/dim"` style
before the drawer itself is drawn, making the underlying content recede.

---

## Structure

```go
type DrawerEdge int

const (
    DrawerLeft   DrawerEdge = iota
    DrawerRight
    DrawerTop
    DrawerBottom
)

type Drawer struct {
    Component
    ui       *UI        // injected by ui.AddDrawer; required for Open/Close
    edge     DrawerEdge
    title    string
    size     int        // columns wide (Left/Right) or rows tall (Top/Bottom)
    dim      bool       // paint backdrop with "drawer/dim" style when open
    open     bool       // true while the drawer is in the popup layer stack
    children []Widget   // direct children laid out in the content area
    // Characters read from theme strings in Apply.
    chClose string      // close indicator, default "✕"
}
```

---

## Constructor

```go
func NewDrawer(id, class string, edge DrawerEdge, title string) *Drawer
```

Defaults:

- `size = 30` (columns or rows).
- `dim = false`.
- `chClose = "✕"`.
- Sets `FlagFocusable`.
- Registers a key handler: `Escape` calls `d.Close()`.
- Registers a mouse handler: click on the close indicator calls `d.Close()`.

---

## Methods

### Display

| Method | Description |
|--------|-------------|
| `SetTitle(s string)` | Updates the title bar text; calls `Refresh()` |
| `SetSize(n int)` | Sets the panel width (Left/Right) or height (Top/Bottom); minimum 4; calls `Refresh()` |
| `SetDim(v bool)` | Enables or disables backdrop dimming; calls `Refresh()` |
| `Edge() DrawerEdge` | Returns the configured edge |
| `IsOpen() bool` | Returns `true` while the drawer is in the popup layer stack |

### Lifecycle

| Method | Description |
|--------|-------------|
| `Open()` | Shows the drawer; no-op when already open |
| `Close()` | Hides the drawer; no-op when already closed |
| `Toggle()` | Calls `Open()` if closed, `Close()` if open |

`Open()` computes the position from the edge and the UI's current dimensions,
then calls `ui.Popup(x, y, w, h, d)`. It dispatches `EvtShow` after the popup
is registered and sets `open = true`.

`Close()` calls `ui.Close()`. It dispatches `EvtHide` when `EvtClose` is
received from the UI and sets `open = false`.

---

## Container interface

`Drawer` implements `Container`:

```go
func (d *Drawer) Add(w Widget)
func (d *Drawer) Children() []Widget
func (d *Drawer) Layout()
```

`Layout()` positions children in the content area: the rectangle inside the
border, with the title bar row excluded from the top.

```
content area:
    cx = d.x + leftBorder + leftPadding
    cy = d.y + topBorder + topPadding + 1   // +1 for title bar
    cw = d.width  - style.Horizontal()
    ch = d.height - style.Vertical() - 1    // -1 for title bar
```

Children are laid out top-to-bottom filling the content area. If only one child
is present it takes the full content area; if none are present the content area
is left empty. For more complex layouts the user is expected to add a single
`Flex` or `Grid` container as the sole child.

---

## Popup integration

`Open()` computes absolute screen coordinates from the edge and the stored `*UI`
reference, then delegates to `ui.Popup`:

| Edge | Call |
|------|------|
| `DrawerLeft` | `ui.Popup(0, 0, d.size, ui.height, d)` |
| `DrawerRight` | `ui.Popup(-2, 0, d.size, ui.height, d)` |
| `DrawerTop` | `ui.Popup(0, 0, ui.width, d.size, d)` |
| `DrawerBottom` | `ui.Popup(0, -2, ui.width, d.size, d)` |

`x = -2` and `y = -2` use the existing `Popup` negative-offset semantics to
place the panel flush with the right or bottom edge:

```
x = -2  →  ui.width  - w + (-2) + 2 = ui.width  - w
y = -2  →  ui.height - h + (-2) + 2 = ui.height - h
```

`ui.AddDrawer(d *Drawer)` stores `ui` in `d.ui` and is the only way the drawer
acquires its `*UI` reference. It does not otherwise modify the widget tree; the
drawer is not part of any layer until `Open()` is called.

---

## Backdrop dimming

When `dim = true`, `Render()` paints all screen cells outside the drawer's own
bounds with a single space using the `"drawer/dim"` style before drawing the
panel. This darkens the underlying content without erasing it structurally —
the cells are only overwritten for the duration of this render pass.

The dimmed region is:

| Edge | Dimmed area |
|------|-------------|
| `DrawerLeft` | `(d.size, 0) … (screenW, screenH)` |
| `DrawerRight` | `(0, 0) … (screenW - d.size, screenH)` |
| `DrawerTop` | `(0, d.size) … (screenW, screenH)` |
| `DrawerBottom` | `(0, 0) … (screenW, screenH - d.size)` |

---

## Render

```
Render order:
  1. If open and dim: paint dimmed backdrop.
  2. Draw background and border via Component.Render().
  3. Draw title bar:
       left of title bar: "drawer/title" style, full panel width minus borders
       title text: left-aligned, truncated to fit
       close indicator (chClose): right-aligned in the title bar
  4. Render children in the content area.
```

The title bar occupies row `cy - 1` (the row immediately above the content
area, inside the border). A horizontal separator rule using the border's
horizontal character is drawn on the row between the title bar and the content
area when the border style provides one; otherwise the separator is omitted and
the title bar sits directly above the first child.

---

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `Escape` | Calls `d.Close()` |

Tab and Shift+Tab cycle focus among the drawer's children in the normal way,
because the drawer is a popup layer and the UI's standard focus traversal
applies within the topmost layer.

---

## Mouse interaction

A click on the close indicator (last column of the title bar row, inside the
border) calls `d.Close()`.

Clicks outside the drawer bounds while `dim = true` also call `d.Close()`,
providing a "click outside to dismiss" behaviour. When `dim = false` clicks
outside the drawer reach the underlying layer normally.

---

## Events

| Event | When |
|-------|------|
| `EvtShow` | Dispatched on the drawer itself after `Open()` pushes it onto the layer stack |
| `EvtHide` | Dispatched on the drawer itself when `EvtClose` is received from the UI (i.e. when `ui.Close()` removes the layer) |

Both events are already defined in `events.go`.

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"drawer"` | Background and border of the panel |
| `"drawer/title"` | Title bar row (background, foreground, font) |
| `"drawer/dim"` | Backdrop cells painted outside the panel when `dim = true` |

Example theme entries (Tokyo Night):

```go
NewStyle("drawer").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(0, 1),
NewStyle("drawer/title").WithColors("$bg1", "$blue").WithBorder("none").WithMargin(0).WithPadding(0, 1),
NewStyle("drawer/dim").WithColors("$bg0", "$bg0"),
```

---

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `"drawer.close"` | `"✕"` | Close indicator drawn right-aligned in the title bar |

---

## Builder usage

```go
ui.NewBuilder().
    Drawer("nav", DrawerLeft, "Navigation", 32).
    List("nav-list", "Home", "About", "Settings").
    End()

drawer := builder.Find("nav").(*Drawer)
ui.AddDrawer(drawer)
drawer.Open()
```

---

## Implementation plan

1. **`drawer.go`** — new file
   - Define `DrawerEdge` type and `DrawerLeft`, `DrawerRight`, `DrawerTop`,
     `DrawerBottom` constants.
   - Define `Drawer` struct embedding `Component`.
   - Implement `NewDrawer(id, class string, edge DrawerEdge, title string) *Drawer`
     with defaults and key/mouse handler registration.
   - Implement display setters: `SetTitle`, `SetSize`, `SetDim`.
   - Implement `Edge() DrawerEdge`, `IsOpen() bool`.
   - Implement `Open()`, `Close()`, `Toggle()`.
   - Implement `Container` interface: `Add`, `Children`, `Layout`.
   - Implement `Apply(t *Theme)`: register `"drawer"`, `"drawer/title"`,
     `"drawer/dim"` selectors; read `"drawer.close"` string.
   - Implement `Hint() (int, int)`: returns `(size, 0)` for Left/Right and
     `(0, size)` for Top/Bottom (explicit `Popup` call overrides these anyway).
   - Implement `Render(r *Renderer)`: dim backdrop, border, title bar with close
     indicator, then render children.

2. **`ui.go`** — add `AddDrawer` method
   ```go
   func (ui *UI) AddDrawer(d *Drawer) {
       d.ui = ui
   }
   ```

3. **`builder.go`** — add `Drawer` method
   ```go
   func (b *Builder) Drawer(id string, edge DrawerEdge, title string, size int) *Builder
   ```
   Creates a `Drawer`, calls `Apply(b.theme)`, sets size, pushes a child-scope
   context so subsequent builder calls add children to the drawer.

4. **Theme** — add `"drawer"`, `"drawer/title"`, `"drawer/dim"` styles and the
   `"drawer.close"` string key to all built-in themes.

5. **`cmd/demo/main.go`** — add a `"Drawer"` entry to the navigation list with
   `drawerDemo`, demonstrating left, right, and bottom variants, and toggling
   `dim` on or off via a checkbox.
