# Rating

A row of discrete symbol characters representing a score. Arrow keys and mouse
clicks change the integer value. When `FlagReadonly` is set the widget renders
as a static indicator with no interactivity — useful for priority columns, card
labels, or severity badges.

---

## Visual layout

**Standard mode** (value = 3 of 5):

```
★ ★ ★ ☆ ☆
```

**Half-star mode** (value = 7 of 10, i.e. 3.5 stars out of 5):

```
★ ★ ★ ⯨ ☆
```

Symbols are separated by `spacing` spaces (default 1). Each symbol occupies
exactly one character cell.

---

## Structure

```go
type Rating struct {
    Component
    value    int  // current value; range [0, count] or [0, count*2] in half-star mode
    count    int  // total number of symbol positions (default 5)
    halfStar bool // allow half-symbol increments
    spacing  int  // blank columns between adjacent symbols (default 1)
}
```

---

## Constructor

```go
func NewRating(id, class string) *Rating
```

- `count = 5`, `value = 0`, `spacing = 1`, `halfStar = false`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

---

## Methods

| Method | Description |
|--------|-------------|
| `SetValue(v int)` | Sets the current value (clamped to the valid range); dispatches `EvtChange(int)`; calls `Refresh()` |
| `Value() int` | Returns the current value |
| `SetCount(n int)` | Sets the number of symbol positions (minimum 1); clamps `value` to the new range; calls `Refresh()` |
| `SetHalfStar(v bool)` | Enables half-symbol mode; resets `value` to 0 when toggled; calls `Refresh()` |
| `SetSpacing(n int)` | Sets the number of blank columns between adjacent symbols (minimum 0); calls `Refresh()` |

### Value range

| Mode | Valid range | Interpretation |
|------|-------------|----------------|
| Standard | `[0, count]` | `n` = n filled symbols |
| Half-star | `[0, count*2]` | `n` = n half-symbols (e.g. `7` = 3 filled + 1 half + 1 empty in a 5-star widget) |

A value of `0` means no symbol is filled (unrated).

---

## Rendering

```go
func (r *Rating) Render(rnd *Renderer)
```

1. `r.Component.Render(rnd)` — draws background and border.
2. Determine the style for filled, half, and empty symbols based on focus state.
3. For each symbol position `i` in `0 … count-1`, draw one character at column
   `cx + i*(1+spacing)`, row `cy`:

**Standard mode:**

```
if i < value  → rating.filled character, "rating/filled" style
else          → rating.empty  character, "rating/empty"  style
```

**Half-star mode:**

```
fullStars = value / 2
hasHalf   = value % 2 == 1

if i < fullStars               → rating.filled character, "rating/filled" style
else if i == fullStars && hasHalf → rating.half character, "rating/half"   style
else                           → rating.empty  character, "rating/empty"   style
```

4. Apply `:focused` selector variants when the widget has focus.

---

## Hint

```go
func (r *Rating) Hint() (int, int)
```

- Width: manually set hint, or `count + spacing*(count-1)` (one cell per symbol,
  `spacing` gaps between them).
- Height: manually set hint, or `1`.

---

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `→` | Increase value by 1 |
| `←` | Decrease value by 1 |
| `Home` | Set value to `0` |
| `End` | Set value to maximum (`count` or `count*2`) |

In half-star mode `←` / `→` move in half-symbol increments (value ± 1). All
changes clamp to the valid range and dispatch `EvtChange(int)`. No action is
taken when `FlagReadonly` or `FlagDisabled` is set.

---

## Mouse interaction

**Standard mode — click on symbol at position `i` (0-based):**

```
value = i + 1
```

Clicking symbol `i` where `i+1 == value` (the rightmost filled symbol) and the
current value is already `i+1` sets `value = 0`, allowing the user to clear the
rating.

**Half-star mode — click at pixel column `mouseX` within the widget:**

```
symbolWidth = 1 + spacing
i           = (mouseX - cx) / symbolWidth         // which symbol (0-based)
leftHalf    = (mouseX - cx) % symbolWidth == 0    // left cell of the symbol

if leftHalf → candidate = i*2 + 1   // half step
else        → candidate = i*2 + 2   // full step

value = candidate if candidate != value else 0    // toggle-off if same
```

Each click dispatches `EvtChange(int)`. Clicks outside the symbol columns
(i.e. in the spacing gaps) are ignored.

No mouse interaction occurs when `FlagReadonly` or `FlagDisabled` is set.

---

## Events

| Event | Data | Description |
|-------|------|-------------|
| `EvtChange` | `int` | Value changed by keyboard or mouse |

---

## Read-only mode

When `FlagReadonly` is set:

- `FlagFocusable` is not set; the widget is excluded from Tab focus traversal.
- Key and mouse handlers are disabled.
- The widget renders identically to the interactive version using the same
  symbol characters and styles, making it a drop-in display component.

---

## Theme strings

| Key | Default | Description |
|-----|---------|-------------|
| `rating.filled` | `★` | Symbol for a filled position |
| `rating.empty`  | `☆` | Symbol for an empty position |
| `rating.half`   | `⯨` | Symbol for a half-filled position (half-star mode) |

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"rating"` | Background and border |
| `"rating/filled"` | Filled symbol characters |
| `"rating/filled:focused"` | Filled symbols when the widget has focus |
| `"rating/half"` | Half symbol character (half-star mode) |
| `"rating/half:focused"` | Half symbol when the widget has focus |
| `"rating/empty"` | Empty symbol characters |
| `"rating/empty:focused"` | Empty symbols when the widget has focus |

Example theme entries (Tokyo Night):

```go
NewStyle("rating").WithBorder("none"),
NewStyle("rating/filled").WithForeground("$yellow"),
NewStyle("rating/filled:focused").WithForeground("$orange"),
NewStyle("rating/half").WithForeground("$yellow"),
NewStyle("rating/half:focused").WithForeground("$orange"),
NewStyle("rating/empty").WithForeground("$bg3"),
NewStyle("rating/empty:focused").WithForeground("$bg4"),
```

---

## Implementation plan

1. **`rating.go`** — new file
   - Define `Rating` struct and `NewRating`.
   - Implement setters: `SetValue`, `SetCount`, `SetHalfStar`, `SetSpacing`.
   - Implement helpers: `maxValue()` (`count` or `count*2`), `symbolAt(i int) (rune, string)`
     (returns the character and selector for position `i` given the current value and mode).
   - Implement `Hint`, `Render`, `handleKey`, `handleMouse`.

2. **`builder.go`** — add `Rating` method
   ```go
   func (b *Builder) Rating(id string) *Builder
   ```

3. **Theme** — add `"rating"` family and `rating.*` string keys to all built-in
   themes.

4. **Tests** — `rating_test.go`
   - `SetValue` clamps at `0` and at `count` (standard) / `count*2` (half-star).
   - Standard mode, value 0: all symbols render as empty.
   - Standard mode, value = count: all symbols render as filled.
   - Standard mode, value 3 of 5: first 3 filled, last 2 empty.
   - Half-star mode, value 7 of 10: 3 filled, 1 half, 1 empty.
   - Half-star mode, value 6 of 10: 3 filled, 0 half, 2 empty.
   - `→` at max does not exceed the valid range; no `EvtChange` dispatched.
   - `←` at 0 does not go below 0; no `EvtChange` dispatched.
   - `Home` sets value to 0; `End` sets value to `maxValue()`.
   - Mouse click on symbol `i` sets value to `i+1` (standard mode).
   - Mouse click on the same symbol again sets value to 0 (toggle-off).
   - Half-star click on left cell sets a half value; right cell sets a full value.
   - `FlagReadonly`: key and mouse handlers produce no state change.
   - `SetHalfStar` toggle resets value to 0.
   - `Hint` width equals `count + spacing*(count-1)` for the default spacing.
   - `SetCount` with a smaller count clamps an out-of-range value.
