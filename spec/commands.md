# Commands

A floating fuzzy-search overlay that gives the user keyboard access to every
registered action in the application. The caller registers named commands with
optional keyboard shortcut hint strings; the palette is opened via
`ui.Commands().Open()`. Typing
narrows the list with a fuzzy match; Enter executes the focused command; Escape
dismisses without action. Commands can optionally be organised into named
groups, which appear as non-interactive section headers.

---

## Visual layout

**With groups, query "op":**

```
┌──────────────────────────────────────────┐
│ Commands                                 │
├──────────────────────────────────────────┤
│ [ op                                   ] │
├──────────────────────────────────────────┤
│ ── File ──────────────────────────────── │
│ › Open file                     Ctrl+O   │
│   Open recent…                  Ctrl+R   │
│ ── View ──────────────────────────────── │
│   Toggle theme                           │
└──────────────────────────────────────────┘
```

**No groups, empty query (all commands shown):**

```
┌──────────────────────────────────────────┐
│ Commands                                 │
├──────────────────────────────────────────┤
│ [                                      ] │
├──────────────────────────────────────────┤
│ › New file                      Ctrl+N   │
│   Open file                     Ctrl+O   │
│   Save file                     Ctrl+S   │
│   Split pane                    Ctrl+\   │
│   Toggle theme                           │
│   Quit                          Ctrl+Q   │
└──────────────────────────────────────────┘
```

The filter input sits at the top. Each command row shows the name left-aligned
and the shortcut right-aligned in a distinct style. Group headers span the full
width and are not selectable. The focused row is highlighted. When the filtered
list is longer than `maxItems`, the panel scrolls.

---

## Command struct

```go
type Command struct {
    Name     string // display label; used for fuzzy matching
    Shortcut string // hint string shown on the right ("Ctrl+O", ""); display only
    Group    string // optional section name; empty = ungrouped
    Action   func() // executed when the command is confirmed
}
```

`Action` is called after the popup is closed, so it may safely open another
popup or modify the widget tree.

---

## Commands struct

```go
type Commands struct {
    ui       *UI
    entries  []*Command // registration order preserved
    maxItems int        // max visible items before the list scrolls (default 10)
    width    int        // explicit popup width override; 0 = auto
    open     bool       // true while the popup is in the layer stack
}
```

`Commands` is not a `Widget` and is never added to the layout tree. It only
exists in memory and materialises a transient popup when `Open()` is called.

---

## Accessing Commands

`Commands` is a lazy singleton owned by the UI:

```go
func (ui *UI) Commands() *Commands
```

The first call allocates the instance; subsequent calls return the same pointer.
The singleton is appropriate because an application has one canonical command
registry, and it persists across open/close cycles so commands registered at
startup remain available.

---

## Registration

```go
func (c *Commands) Register(name, shortcut string, action func()) *Command
```

Appends a command at the end of the registration list. Returns the `*Command`
for chaining or later removal. `name` must be non-empty; `shortcut` and `action`
may be empty/nil.

```go
func (c *Commands) RegisterGroup(group, name, shortcut string, action func()) *Command
```

Same as `Register` but assigns a group label. Commands with the same group
string are gathered under a shared header when displayed in group order.

```go
func (c *Commands) Unregister(name string) bool
```

Removes the first command with the matching name (exact, case-sensitive).
Returns `true` if found and removed. No-op when the palette is open.

```go
func (c *Commands) All() []*Command
```

Returns a snapshot of the current registration slice.

---

## Display options

```go
func (c *Commands) SetMaxItems(n int)
```

Sets the maximum number of command rows visible before the list scrolls.
Clamped to minimum 3. Default 10.

```go
func (c *Commands) SetWidth(n int)
```

Forces the popup to a fixed column width, overriding auto-sizing. 0 = auto.

---

## Open and Close

```go
func (c *Commands) Open()
func (c *Commands) Close()
func (c *Commands) IsOpen() bool
```

`Open()` is a no-op when the palette is already open. It:

1. Computes popup dimensions (see *Sizing* below).
2. Constructs a transient popup — a fresh widget tree built with
   `c.ui.NewBuilder()` — containing a filter input and a command panel.
3. Seeds the command panel with all registered commands in registration order,
   grouped if any commands carry a group label.
4. Focuses the filter input immediately.
5. Wires all internal event handlers (see *Interaction* below).
6. Calls `c.ui.Popup(-1, -1, w, h, dialog)` to show it centred.
7. Sets `c.open = true`.

`Close()` calls `c.ui.Close()` and sets `c.open = false`. Calling it when
the palette is not open is a no-op.

---

## Sizing

**Width (auto):**

```
nameW     = max display-column width across all command names
shortcutW = max display-column width across all shortcut strings
                (0 if all shortcuts are empty)
minW      = nameW + (shortcutW > 0 ? shortcutW + 3 : 0) + 4  // 4 = padding + border
popupW    = clamp(max(minW, 44), 44, screenWidth - 4)
```

**Height:**

```
visibleRows = min(len(filteredCommands) + groupHeaderCount, maxItems)
popupH      = 1            // title bar
            + 1            // separator
            + 1            // filter input row
            + 1            // separator
            + visibleRows  // command rows
            + 1            // bottom border
```

Height is recomputed and the popup is not resized after opening — the panel
scrolls instead when the filtered set shrinks or grows.

---

## Fuzzy matching

The command panel re-filters on every keystroke in the filter input. The
matching algorithm is a character-order fuzzy match (not substring):

**Match test:** query characters must all appear in the command name in order
(case-insensitive), but need not be adjacent.

```
isMatch("op fi", "Open file") → true   // o…p…fi present in order
isMatch("opfi",  "Open file") → true
isMatch("oz",    "Open file") → false  // 'z' not in name
```

**Score:** computed over the matched character positions `[p₀, p₁, …]`:

| Bonus | Condition |
|-------|-----------|
| +5 | Matched character is at a word boundary (position 0, or preceded by ` `, `/`, `_`, `-`) |
| +3 | Matched character is immediately adjacent to the previous match (`pᵢ == pᵢ₋₁ + 1`) |
| +1 | Exact case match (matched without lowercasing) |

Commands are presented sorted by score descending; ties are broken by
registration order. Commands with a score of 0 (no bonuses, matches only by
character presence) are still shown.

When the query is empty, all commands are shown in registration order with no
scoring applied.

**Group ordering:** when groups are in use, the sort is stable within each
group. The group order follows first-appearance order in the registration list.
Group headers are shown even if only one item in the group matches.

---

## Popup structure

The transient popup tree is built with `ui.NewBuilder()`:

```
Dialog("commands-dialog", "commands", "Commands")
  └─ Flex("commands-flex", vertical, "stretch", 0)
       ├─ Filter("commands-input")          // always focused on open
       └─ commandsPanel("commands-panel")   // unexported widget
```

`commandsPanel` is an unexported widget defined inside `commands.go`. It holds
the filtered+scored command slice and implements:

- `SetItems(items []rankedCommand)` — replaces the visible list and resets
  scroll position and selection to the first non-header item.
- Up/Down/Home/End navigation skipping group header rows.
- Per-row rendering: name left-aligned in `"commands/item"` style, shortcut
  right-aligned in `"commands/shortcut"` style, focused row in
  `"commands/item:focused"`.
- Group header rows rendered in `"commands/group"` style, spanning full width.
- Vertical scrolling when `len(items) > maxItems`.
- `EvtActivate` dispatched with `*Command` data on Enter or double-click.
- `FlagFocusable` is **not** set — keyboard focus stays in the filter input at
  all times; Up/Down key events bubble from the input to the panel via the
  normal propagation chain.

```go
type rankedCommand struct {
    cmd   *Command
    score int
    isHeader bool   // true for group header rows
}
```

---

## Interaction

### Keyboard

| Key | Behaviour |
|-----|-----------|
| Any printable character | Appended to filter; panel re-filters |
| `Backspace` | Deletes last filter character; panel re-filters |
| `↑` / `↓` | Moves selection in the command panel (skipping group headers) |
| `Home` / `End` | Jumps to first / last selectable row |
| `Enter` | Executes the focused command's `Action`; closes the popup |
| `Escape` | Closes the popup without executing any command |
| `Tab` | Accepts ghost-text suggestion in the filter input (inherited from `Filter`) |

Escape is handled at the UI level (global `KeyEscape → ui.Close()` when a
popup layer is open), so no special wiring is required.

The filter input receives focus on `Open()`. Up/Down keys are handled in the
filter's key handler (before the default Tab navigation) and forwarded to the
panel, so the user never needs to move focus out of the input field.

### Mouse

A click on a command row selects it (updates the panel's focused index). A
second click on the already-selected row — or a double-click — executes the
command and closes the popup. Clicks on group header rows are ignored.

---

## Events

`Commands` dispatches no events itself. All user feedback is delivered through
the `Action` callbacks.

---

## Lifecycle and Action safety

`Action` is called **after** `ui.Close()` removes the popup from the layer
stack. This ordering guarantees that:

- `Action` may open a new popup without interfering with the closing animation
  of the commands palette.
- `Action` may call `ui.Commands().Open()` to re-open the palette (e.g. for
  sub-command flows) without a reentrancy hazard.

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"commands"` | The Dialog border, background, and title bar |
| `"commands/input"` | The Filter input widget |
| `"commands/item"` | Unfocused command row (name and background) |
| `"commands/item:focused"` | Focused command row |
| `"commands/shortcut"` | Shortcut hint text on unfocused rows |
| `"commands/shortcut:focused"` | Shortcut hint text on the focused row |
| `"commands/group"` | Group header rows |

The `"commands"` selector is applied to the `Dialog` container; all other
selectors are applied inside `commandsPanel`.

Example theme entries (Tokyo Night):

```go
NewStyle("commands").WithColors("$fg0", "$bg2").WithBorder("round").WithPadding(0, 0),
NewStyle("commands/input").WithColors("$fg0", "$bg3").WithCursor("*bar"),
NewStyle("commands/item").WithColors("$fg1", "$bg2"),
NewStyle("commands/item:focused").WithColors("$bg0", "$blue").WithFont("bold"),
NewStyle("commands/shortcut").WithColors("$fg2", "$bg2"),
NewStyle("commands/shortcut:focused").WithColors("$bg1", "$blue"),
NewStyle("commands/group").WithColors("$fg2", "$bg2").WithFont("bold"),
```

---

## Usage example

```go
cmds := ui.Commands()

cmds.RegisterGroup("File", "New file",   "Ctrl+N", func() { newFile(ui) })
cmds.RegisterGroup("File", "Open file",  "Ctrl+O", func() { openFile(ui) })
cmds.RegisterGroup("File", "Save file",  "Ctrl+S", func() { saveFile(ui) })
cmds.RegisterGroup("View", "Toggle theme", "",     func() { toggleTheme(ui) })
cmds.RegisterGroup("View", "Split pane", "Ctrl+\\", func() { splitPane(ui) })

// Bind Ctrl+K to open the palette
OnKey(root, func(e *tcell.EventKey) bool {
    if e.Key() == tcell.KeyCtrlK {
        ui.Commands().Open()
        return true
    }
    return false
})
```

Commands can also be registered at any time after startup — for example,
context-sensitive commands added when a specific widget gains focus and removed
when it loses it:

```go
var editCmd *Command

editor.On(EvtFocus, func(_ Widget, _ Event, _ ...any) bool {
    editCmd = ui.Commands().Register("Format document", "Ctrl+Shift+F", formatDoc)
    return false
})
editor.On(EvtBlur, func(_ Widget, _ Event, _ ...any) bool {
    if editCmd != nil {
        ui.Commands().Unregister(editCmd.Name)
        editCmd = nil
    }
    return false
})
```

---

## Implementation plan

1. **`commands.go`** — new file
   - Define `Command` struct.
   - Define `Commands` struct.
   - Implement `newCommands(ui *UI) *Commands` (unexported constructor).
   - Implement `Register`, `RegisterGroup`, `Unregister`, `All`.
   - Implement `SetMaxItems`, `SetWidth`.
   - Implement `Open`, `Close`, `IsOpen`.
   - Implement `fuzzyMatch(query, name string) (matched bool, score int)`.
   - Implement `filterCommands(query string) []rankedCommand`.
   - Define unexported `commandsPanel` struct and its widget methods:
     `SetItems`, `Hint`, `Render`, `handleKey`, `handleMouse`.

2. **`ui.go`** — add `Commands` accessor
   - Add `commands *Commands` field to `UI`.
   - Implement `func (ui *UI) Commands() *Commands`.

3. **Theme** — add `"commands"`, `"commands/input"`, `"commands/item"`,
   `"commands/item:focused"`, `"commands/shortcut"`, `"commands/shortcut:focused"`,
   and `"commands/group"` style entries to all built-in themes.

4. **`cmd/demo/main.go`** — add a `"Commands"` entry with `commandsDemo` that
   registers a representative set of commands (including grouped ones) and
   documents how to bind Ctrl+K, demonstrating the palette on a live UI with
   commands that modify visible state.

5. **Tests** — `commands_test.go`
   - `fuzzyMatch` returns `false` when query characters are not all present
     in the name.
   - `fuzzyMatch` awards word-boundary bonus at position 0 and after ` `/`-`/`_`.
   - `fuzzyMatch` awards adjacency bonus for consecutive matched positions.
   - `filterCommands` returns all commands when query is empty.
   - `filterCommands` sorts by score descending, stable within equal scores.
   - `filterCommands` preserves group order across score-sorted results.
   - `Register` appends; `Unregister` removes the first matching name.
   - `Open` while already open is a no-op (no second popup layer pushed).
   - `Action` is called after the popup is removed from the layer stack.
   - Group headers are included in `rankedCommand` slice but skipped by
     Up/Down navigation.
