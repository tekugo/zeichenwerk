# Designer

An interactive TUI application for building zeichenwerk widget hierarchies
visually and exporting them as Go source files. The designer maintains an
in-memory tree of `DesignerNode` values, renders a live preview of the layout,
and generates compilable Go code using either the Builder API or the Compose
API.

The existing `cmd/designer` (a VIM-style canvas editor) is replaced by this
application.

---

## Conceptual model

A *designer document* is a tree of nodes. Each node describes one widget:
its kind, identity, layout overrides, and kind-specific properties. The tree
mirrors the widget hierarchy that would be produced by a Builder chain.

```
Flex "root"  horizontal=false  align=stretch  spacing=0  hint=(-1,-1)
├─ Flex "header"  horizontal=true  align=center  spacing=2
│  ├─ Static "title"  text="My App"
│  └─ Button "quit"   text="Quit"
└─ Grid "body"  rows=["-1"]  cols=["28","-1"]
   ├─ List "nav"
   └─ Viewport "main"
      └─ Static "content"  text="Hello"
```

The designer does not execute the widgets. The live preview is produced by
a separate Builder pass that materialises the document tree into real widgets
and renders them onto a virtual screen.

---

## Visual layout

```
┌─ toolbar ─────────────────────────────────────────────────────────────────┐
│ [+Child] [+Sibling] [↑] [↓] [Indent] [Outdent] [✂ Cut] [⎘ Copy] [⏎ Paste]│
│ [Delete]                               [⬡ Preview]  [{ Generate}]  [Save] │
└───────────────────────────────────────────────────────────────────────────┘
┌─ Tree (32 cols) ─────────┐ ┌─ Properties (40 cols) ──────────────────────┐
│ ▼ Flex "root"            │ │ Kind      Flex                               │
│   ▼ Flex "header"        │ │ ID        [header                          ] │
│     Static "title"       │ │ Class     [                                ] │
│   ▼ Grid "body"          │ │ Hint W    [-1                              ] │
│   ▶ List "nav"           │ │ Hint H    [-1                              ] │
│   ▼ Viewport "main"      │ │ Padding   [0                               ] │
│       Static "content"   │ │ Margin    [                                ] │
│                          │ │ ──── Flex ────────────────────────────────  │
│                          │ │ Horizontal [✓]                               │
│                          │ │ Alignment  [stretch ▼]                       │
│                          │ │ Spacing    [2                               ] │
└──────────────────────────┘ └─────────────────────────────────────────────┘
┌─ Preview ─────────────────────────────────────────────────────────────────┐
│  My App                                                          [Quit]   │
│ ┌────────────────────────────────────────────────────────────────────────┐│
│ │ Hello                                                                  ││
│ └────────────────────────────────────────────────────────────────────────┘│
└───────────────────────────────────────────────────────────────────────────┘
```

The preview and code panels are shown in alternating mode toggled by `Tab` or
the toolbar buttons. The split is: tree + properties take the top 60 % of
terminal height; preview or code takes the bottom 40 %.

---

## Data model

### DesignerNode

```go
// DesignerNode is one node in the document tree.
type DesignerNode struct {
    Kind     string            // widget type, e.g. "Flex", "Static", "Button"
    ID       string            // builder id argument
    Class    string            // builder class argument (may be empty)
    HintW    int               // hint width override (0 = not set)
    HintH    int               // hint height override (0 = not set)
    Padding  *InsetSpec        // nil = not set (use theme default)
    Margin   *InsetSpec        // nil = not set
    Props    map[string]string // kind-specific properties (see §Widget palette)
    Children []*DesignerNode
}

// InsetSpec mirrors the inset format accepted by WithPadding / WithMargin.
type InsetSpec struct{ Top, Right, Bottom, Left int }
```

`Props` keys and values are always strings; the code generator parses them
back to the right Go types (bool, int, string).

### Document

```go
type Document struct {
    Root    *DesignerNode // single root node; always a container kind
    Theme   string        // theme key: "tokyo", "gruvbox-dark", etc.
    Package string        // package name for generated file (default "main")
    Import  string        // module path for the zeichenwerk import
}
```

Documents are serialised to/from JSON. A new document starts with a single
`Flex "root"` node.

---

## Widget palette

The following kinds are supported. Each entry lists the required and optional
`Props` keys.

### Containers (may have children)

| Kind | Required Props | Optional Props |
|------|---------------|----------------|
| `Flex` | `horizontal` (bool), `alignment` (string), `spacing` (int) | — |
| `Grid` | `rows` (space-sep list), `cols` (space-sep list) | `lines` (bool) |
| `Box` | `title` (string) | — |
| `Dialog` | `title` (string) | — |
| `Collapsible` | `title` (string), `expanded` (bool) | — |
| `Tabs` | — | — |
| `Switcher` | — | — |
| `Viewport` | `title` (string) | — |

Container nodes expand/collapse in the tree panel. Attempting to add a child to
a non-container kind is rejected with an inline error message in the toolbar.

### Leaf widgets

| Kind | Required Props | Optional Props |
|------|---------------|----------------|
| `Static` | `text` (string) | — |
| `Button` | `text` (string) | — |
| `Input` | — | `placeholder` (string), `max` (int), `mask` (bool) |
| `Checkbox` | `text` (string), `checked` (bool) | — |
| `Select` | `options` (alternating val/label pairs, comma-sep) | — |
| `List` | — | — |
| `Deck` | `itemHeight` (int) | — |
| `Tiles` | `tileWidth` (int), `tileHeight` (int) | — |
| `Table` | — | — |
| `Tree` | — | — |
| `Text` | `follow` (bool), `max` (int) | — |
| `Progress` | `horizontal` (bool) | — |
| `Sparkline` | — | — |
| `Rule` | `horizontal` (bool) | — |
| `Breadcrumb` | — | — |
| `Marquee` | — | — |
| `Shimmer` | — | — |

For `List`, `Deck`, `Tiles`, `Table`, and `Tree`, the designer only generates
the constructor call. The ItemRender function and data-binding are left as
`// TODO` stubs in the generated code.

---

## UI components

### Toolbar

A horizontal `Flex` with `Button` widgets. Buttons are enabled/disabled based
on context (e.g. `Outdent` is disabled when the selected node is a direct child
of root; `Paste` is disabled when the clipboard is empty).

Keyboard shortcuts (available globally):

| Key | Action |
|-----|--------|
| `a` | Add child |
| `A` | Add sibling after |
| `d` / `Delete` | Delete selected node |
| `J` / `K` | Move node down / up among siblings |
| `>` / `<` | Indent (make last sibling's child) / Outdent |
| `x` | Cut |
| `c` | Copy subtree |
| `v` | Paste as last child of selected |
| `Tab` | Toggle preview / code panel |
| `Ctrl+S` | Save document to file |
| `Ctrl+G` | Generate and copy code to clipboard |
| `Ctrl+N` | New document |
| `Ctrl+O` | Open document |

### Tree panel

A `Tree` widget. Each `TreeNode` carries its `*DesignerNode` as opaque data.
Selecting a node fires `EvtSelect`, which populates the properties panel.
Expanding/collapsing shows/hides children.

Label format: `Kind "id"` for identified nodes, `Kind` for nodes with empty ID.
Nodes with no children and a container kind are shown with a hollow triangle
`▷` (expandable but empty).

### Properties panel

A `FormGroup` with fields that change based on the selected node's `Kind`.
The top section (ID, Class, Hint W, Hint H, Padding, Margin) is always shown.
Below a divider, kind-specific fields appear.

Editing any field:
1. Validates the input (type check, range check where applicable).
2. Updates the `*DesignerNode` immediately.
3. Re-generates the code panel content.
4. Triggers a preview rebuild (debounced 200 ms).

Validation errors are shown inline below the offending field using a red
`Static` widget.

### Preview panel

A `Viewport` containing a fixed-size virtual render area. The preview is
rebuilt by:

1. Walking the document tree and calling the corresponding `NewBuilder`
   methods, passing placeholder data/render-functions for widgets that require
   them (List gets `[]any{"Item 1", "Item 2", "Item 3"}`, Deck gets a no-op
   ItemRender, etc.).
2. Calling `builder.Build()` to get a `*UI`.
3. Setting bounds to the preview panel's content dimensions.
4. Running `ui.Layout()`.
5. Calling `ui.Render()` onto a `CellBuffer` screen.
6. Blitting the cell buffer into the preview viewport.

If the document tree has errors (e.g. a non-container node with children),
the preview shows an error message instead.

### Code panel

A `Text` widget (read-only) showing the generated Go source. Updated whenever
the document changes. The full generated file is shown, syntax-highlighted
using `$cyan` for keywords, `$green` for string literals, `$yellow` for
identifiers where possible (simple tokeniser — not a full Go parser).

---

## Code generation

### Builder API output

Walking the document tree emits a `func buildUI(theme *Theme) *UI` function:

```go
func buildUI(theme *zeichenwerk.Theme) *zeichenwerk.UI {
    return zeichenwerk.NewBuilder(theme).
        Flex("root", false, "stretch", 0).Hint(-1, -1).
            Flex("header", true, "center", 2).
                Static("title", "My App").
                Button("quit", "Quit").
            End().
            Grid("body", 1, 2, false).Hint(-1, -1).
                Columns(28, -1).
                List("nav").Hint(0, -1).
                Viewport("main", "").Hint(0, -1).
                    Static("content", "Hello").
                End().
            End().
        End().
        Build()
}
```

Rules:
- Each container calls `.End()` after its children.
- Hint overrides are emitted only when HintW or HintH is non-zero.
- Padding/Margin overrides are emitted only when set.
- For `Grid`, `Columns(…)` and `Rows(…)` are emitted immediately after the
  `Grid(…)` call with values from the `cols` / `rows` props.
- `Tabs` children are emitted with `.Tab("label")` before each child.
- `Switcher` children are emitted as separate `.With(func(b){…})` blocks.
- Widgets needing ItemRender (Deck, Tiles) receive a stub:
  ```go
  Deck("items", /* TODO: add ItemRender */ nil, 3).
  ```

### Compose API output (optional toggle)

An alternative generator emits the `compose` package style:

```go
func buildUI(theme *zeichenwerk.Theme) *zeichenwerk.UI {
    return compose.UI(theme,
        compose.Flex("root", "", false, "stretch", 0,
            compose.Flex("header", "", true, "center", 2,
                compose.Static("title", "", "My App"),
                compose.Button("quit", "", "Quit"),
            ),
            // …
        ),
    )
}
```

### File header

The generated file starts with:

```go
package main

import (
    z "github.com/tekugo/zeichenwerk"
    // "github.com/tekugo/zeichenwerk/compose"  // uncomment for Compose API
)
```

The import alias `z` is used only when the dot-import is not preferred. A
`--dot-import` flag switches to `. "github.com/tekugo/zeichenwerk"`.

---

## Add-child flow

Pressing `a` (or the `[+Child]` toolbar button) when a node is selected:

1. Opens a modal `Dialog` with a `List` of all valid widget kinds (containers
   at top, leaves below a divider).
2. The user selects a kind and presses `Enter`.
3. A new `DesignerNode` is created with default Props for that kind, an
   auto-generated ID (`kind + sequential number`, e.g. `"flex2"`), and appended
   as the last child of the selected node.
4. The dialog closes, the tree expands the parent node, and the new node is
   selected.
5. Focus moves to the ID field in the properties panel for immediate editing.

If the selected node is a leaf kind, the new node is added as a sibling after
it instead, and the toolbar shows a brief warning: `"Non-container — added as
sibling"`.

---

## Clipboard

Cut/copy stores a deep copy of the selected subtree in an in-process
`*DesignerNode` variable. Paste inserts it as the last child of the currently
selected node (deep copy again, with IDs suffixed `_copy` to avoid duplicates).
There is no cross-process clipboard integration.

---

## Persistence

Documents are saved as JSON. The default filename is `ui.designer.json` in the
current working directory. The file path is shown in the toolbar title.

```json
{
  "root": {
    "kind": "Flex",
    "id": "root",
    "hintW": -1,
    "hintH": -1,
    "props": { "horizontal": "false", "alignment": "stretch", "spacing": "0" },
    "children": [ … ]
  },
  "theme": "tokyo",
  "package": "main",
  "import": "github.com/tekugo/zeichenwerk"
}
```

On save failure, the toolbar status area shows the error in red for 3 seconds.

---

## Application structure (`cmd/designer`)

```
cmd/designer/
    main.go          — flag parsing, theme selection, NewDesignerApp
    app.go           — DesignerApp (top-level widget wiring, global key handler)
    model.go         — DesignerNode, InsetSpec, Document, deep-copy, JSON marshalling
    palette.go       — widgetKinds list, defaultProps(), isContainer()
    tree-panel.go    — tree panel widget, sync between DesignerNode tree and TreeNode tree
    props-panel.go   — properties panel, per-kind field generation, validation
    preview.go       — preview rebuilder, virtual-screen blitting
    codegen.go       — Builder and Compose code generators
    id-counter.go    — auto-ID generation, uniqueness enforcement
```

### DesignerApp

`DesignerApp` is the top-level composite widget (implements `Container`). Its
inner layout:

```go
NewBuilder(theme).
    Flex("app", false, "stretch", 0).
        Flex("toolbar", true, "center", 1).Hint(0, 1).
            // toolbar buttons …
        End().
        Flex("workspace", true, "stretch", 0).Hint(0, -1).
            Tree("node-tree", "").Hint(32, 0).
            PropsPanel("props", "").Hint(40, 0).
            Switcher("right-panel", false).Hint(-1, 0).
                Viewport("preview-pane", "").
                End().
                // code pane …
            End().
        End().
    End()
```

`PropsPanel` is an internal composite widget (not exported from the library)
that rebuilds its field layout whenever the selected node kind changes.

---

## Events

`DesignerApp` dispatches the following events on itself:

| Event | Payload | Description |
|-------|---------|-------------|
| `EvtChange` | `*Document` | Any node was added, deleted, or had a property modified |
| `EvtSelect` | `*DesignerNode` | A different node was selected in the tree |

---

## Implementation plan

1. **`cmd/designer/model.go`**
   - `DesignerNode`, `InsetSpec`, `Document`.
   - `Clone(node) *DesignerNode` (deep copy for clipboard/undo).
   - JSON marshal/unmarshal round-trip.
   - `Validate(doc) []error` — checks for duplicate IDs, invalid prop values.

2. **`cmd/designer/palette.go`**
   - `widgetKinds []KindSpec` — ordered list of all supported kinds.
   - `defaultProps(kind string) map[string]string`.
   - `isContainer(kind string) bool`.
   - `propFields(kind string) []PropField` — field descriptors for the
     properties panel (name, label, type, validation).

3. **`cmd/designer/id-counter.go`**
   - `IDCounter` — tracks used IDs, generates unique ones.
   - `Allocate(prefix string) string`.
   - `ScanDocument(doc *Document)` — seed from existing document.

4. **`cmd/designer/codegen.go`**
   - `GenerateBuilder(doc *Document) string`.
   - `GenerateCompose(doc *Document) string`.
   - Both functions return a complete Go source file as a string.

5. **`cmd/designer/preview.go`**
   - `PreviewBuilder` — builds a real widget tree from a `*Document`,
     renders onto a `CellBuffer`, returns the populated buffer.
   - Uses a goroutine + debounce channel to avoid rebuilding on every
     keystroke.

6. **`cmd/designer/tree-panel.go`**
   - `syncTree(root *DesignerNode, t *Tree)` — rebuilds the `Tree` widget's
     `TreeNode` hierarchy from the document tree.
   - Preserves expanded state across rebuilds by matching node IDs.

7. **`cmd/designer/props-panel.go`**
   - `PropsPanel` widget (internal).
   - `Load(node *DesignerNode)` — populates fields.
   - `fieldChanged(name, value string)` — validates and writes back to node,
     fires parent `EvtChange`.

8. **`cmd/designer/app.go`**
   - `DesignerApp` composite widget.
   - `NewDesignerApp(theme *Theme, doc *Document) *DesignerApp`.
   - Global key handler, toolbar wiring, clipboard, file I/O.
   - `AddChild()`, `AddSibling()`, `DeleteSelected()`, `MoveUp()`,
     `MoveDown()`, `Indent()`, `Outdent()`, `Cut()`, `Copy()`, `Paste()`.

9. **`cmd/designer/main.go`**
   - Flags: `--file` (path), `--theme`, `--dot-import`, `--compose`.
   - Loads existing document or starts with empty Flex root.
   - Wires `Ctrl+S` → save, `Ctrl+G` → generate + print to stdout.

---

## Non-goals

- Drag-and-drop node rearrangement (use `J`/`K`/Indent/Outdent instead).
- Undo/redo (left as a future enhancement; `Clone` in `model.go` is the
  building block).
- Rendering pixel-accurate previews for widgets that need live data (Deck,
  Tiles, Table, Tree) — these show placeholder content.
- Round-tripping existing Go source back into the designer.
