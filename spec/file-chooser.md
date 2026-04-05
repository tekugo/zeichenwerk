# Specification: FileChooser

**File:** `file_chooser.go`  
**Package:** `zeichenwerk`  
**Depends on:** `Tree`, `TreeNode`, `Dialog`, `Input`, `Button`, `Checkbox`, `Flex`, `UI.Popup`

---

## Purpose

`FileChooser` is a modal overlay for selecting a file or directory from the
local filesystem. It is invoked through a `UI` method, similar to
`UI.Prompt` and `UI.Confirm`, and returns the chosen path through a callback.

---

## API

```go
// FileChooser shows a modal file/directory chooser dialog and returns the
// dialog widget. Attach event handlers to the returned widget before yielding
// control back to the event loop.
//
// title is shown in the dialog title bar.
//
// label is the confirm button text — e.g. "Open", "Save", or "Select".
//
// mode controls what can be selected:
//   - "dir"  — directories only (files are visible but not selectable)
//   - "file" — files only (directories are navigable but not selectable)
//   - "any"  — both files and directories are selectable
//
// initial is the starting path. If empty, os.Getwd() is used.
//
// showHidden controls whether dotfiles and dot-directories are initially
// visible. The user can toggle this at runtime via a checkbox.
func (ui *UI) FileChooser(title, label, mode, initial string, showHidden bool) Widget
```

### Events fired on the returned widget

| Event | Payload | When |
|-------|---------|------|
| `EvtAccept` | `string` — chosen absolute path | User confirms a valid selection |
| `EvtClose` | — | Dialog is closing for any reason (confirm or cancel) |

`EvtAccept` is dispatched before `ui.Close()` is called, so the handler runs
while the dialog is still visible. `EvtClose` is dispatched on every dismissal
— use it for cleanup or to detect cancellation (i.e. `EvtClose` without a
preceding `EvtAccept`).

### Example

```go
fc := ui.FileChooser("Choose Directory", "Select", "dir", cwd, false)
fc.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
    dir := data[0].(string)
    switchDirectory(dir)
    return true
})
fc.On(EvtClose, func(_ Widget, _ Event, _ ...any) bool {
    // cleanup if needed
    return true
})
```

---

## Layout

```
┌─ Choose Directory ──────────────────────────────────────────┐
│  /home/thomas/Projects/myapp________________________________│
│ ┌─────────────────────────────────────────────────────────┐ │
│ │  ▼ /                                                    │ │
│ │    ▶ etc                                                │ │
│ │    ▶ home                                               │ │
│ │    ▼ home                                               │ │
│ │      ▼ thomas                                           │ │
│ │        ▼ Projects                                       │ │
│ │        ► myapp                                          │ │
│ │          ▶ src                                          │ │
│ └─────────────────────────────────────────────────────────┘ │
│  [x] show hidden                         [Select]  [Cancel] │
└─────────────────────────────────────────────────────────────┘
```

### Regions

| Region | Widget | Notes |
|--------|--------|-------|
| Title bar | `Dialog` | Passed-in `title` string |
| Path input | `Input` | Editable absolute path; updated on tree navigation |
| Tree | `Tree` | Scrollable filesystem tree; lazy-loaded from `/` |
| Footer left | `Checkbox` | "show hidden"; initial state from `showHidden` param |
| Footer right | 2× `Button` | Confirm (`label`) + `Cancel`; right-aligned |

The dialog is sized to 60 × 20 characters and centered via `UI.Popup(-1, -1, 60, 20, …)`.

---

## Tree behaviour

### Population

- The tree always roots at `/`. The single top-level visible node is `/`.
- Each directory node is created with `NewLazyTreeNode`; on first expand,
  `os.ReadDir` populates its children.
- Files are added as leaf nodes (no loader).
- Entries are sorted: directories first, then files, both groups
  case-insensitively alphabetical.
- On open, the tree **pre-expands** the path components of `initial` one by
  one (loading each lazily in turn) and scrolls the final node into view.

### Hidden files

- Nodes whose name starts with `.` are hidden when `showHidden` is `false`.
- The "show hidden" `Checkbox` reflects the current state. Toggling it
  collapses and repopulates the entire tree with the new filter applied, then
  re-navigates to the path currently shown in the path input.

### Node selectability by mode

- **`"dir"`** — directory nodes are selectable; file nodes are dimmed and
  skipped by keyboard navigation (`FlagDisabled`).
- **`"file"`** — file nodes are selectable; directory nodes are navigable
  (expand/collapse) but cannot be confirmed.
- **`"any"`** — all nodes are selectable.

The confirm button is disabled whenever the currently focused node is not
selectable in the active mode.

---

## Path input

- Shows the absolute path of the currently focused tree node.
- The user may type a path directly. After each keystroke:
  - If the typed path resolves to a node already visible in the tree, that
    node is highlighted and scrolled into view.
  - If it resolves to a directory not yet loaded, the tree lazy-loads and
    expands path components one level at a time.
  - If the path does not exist or is not selectable in the current mode, the
    input is rendered with `"filechooser/input.error"` style and the confirm
    button is disabled.
- Pressing `↑` or `↓` while the input has focus moves focus to the tree.

---

## Keyboard bindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate tree; updates path input |
| `→` / `Space` | Expand focused directory node |
| `←` | Collapse focused node; if already collapsed, jump to parent |
| `Enter` | Confirm selection (same as clicking the confirm button) |
| `Tab` / `Shift+Tab` | Cycle focus: path input → tree → confirm button → cancel button → … |
| `Escape` | Cancel |
| `~` (tree focused) | Expand and jump to the user's home directory |
| `/` (tree focused) | Transfer focus to path input, set text to `/` |

---

## Confirm / cancel flow

**Confirm** (confirm button or `Enter` on a selectable node):
1. Resolve the path input to an absolute, cleaned path (`filepath.Clean`).
2. Dispatch `EvtAccept` with the path as payload.
3. Call `ui.Close()`.

**Cancel** (`Cancel` button or `Escape`):
1. Call `ui.Close()` — which dispatches `EvtClose` on the popup layer.

---

## Style keys

```
"filechooser/input"              — path input (default)
"filechooser/input.error"        — path input when path is invalid
"filechooser/node.dir"           — directory node label
"filechooser/node.file"          — file node label
"filechooser/node.file:disabled" — file node in "dir" mode (dimmed)
"filechooser/node.dir:disabled"  — directory node in "file" mode (reduced)
```

All six themes (`tokyo-night`, `midnight-neon`, `nord`, `gruvbox-dark`,
`gruvbox-light`, `lipstick`) must define these style keys.
