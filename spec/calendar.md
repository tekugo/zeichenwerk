# Calendar

A month-grid widget that displays a single month as a 7-column day grid.
The user navigates days, weeks, and months with the keyboard and confirms a
date with Enter. Can be used standalone or embedded in a dialog.

## Visual layout

```
◀  March 2026  ▶
Mo Tu We Th Fr Sa Su
                    1
 2  3  4  5  6  7  8
 9 10 11 12 13 14 15
16 17 18 19 20 21 22
23 24 25 26 27 28 29
30 31
```

Fixed dimensions:
- Width: `7 * cellWidth + 6 * gap` where `cellWidth = 2`, `gap = 1` → **20 chars**
- Height: 1 header + 1 weekday row + up to 6 day rows → **8 rows** (always 6 day rows, some cells empty)

The header row shows the month name and year, with navigation arrows at the
left and right edges. The weekday row shows abbreviated day names in the
configured locale. Day cells are right-aligned 2-character numbers.

## Structure

```go
type Calendar struct {
    Component
    selected  time.Time     // Currently highlighted date
    viewing   time.Time     // Month/year being displayed
    weekStart time.Weekday  // First day of week (Monday or Sunday, default Monday)
    minDate   time.Time     // Optional lower bound (zero = no bound)
    maxDate   time.Time     // Optional upper bound (zero = no bound)
}
```

`selected` and `viewing` may differ: the user can navigate to a different month
without losing the previously selected date.

## Constructor

```go
func NewCalendar(id, class string) *Calendar
```

- Initialises `selected` and `viewing` to today (`time.Now()` truncated to
  the calendar day in the local timezone).
- `weekStart = time.Monday`.
- Sets `FlagFocusable`.
- Registers key and mouse handlers.

## Methods

| Method | Description |
|--------|-------------|
| `SetDate(t time.Time)` | Sets `selected` and `viewing` to the date's month; calls `Refresh()` |
| `Date() time.Time` | Returns the currently selected date |
| `SetWeekStart(d time.Weekday)` | Sets the first column day (Monday or Sunday) |
| `SetMinDate(t time.Time)` | Lower bound; dates before it are disabled |
| `SetMaxDate(t time.Time)` | Upper bound; dates after it are disabled |
| `PrevMonth()` | Moves `viewing` back one month |
| `NextMonth()` | Moves `viewing` forward one month |
| `PrevYear()` | Moves `viewing` back one year |
| `NextYear()` | Moves `viewing` forward one year |

## Keyboard interaction

| Key | Behaviour |
|-----|-----------|
| `←` / `→` | Move `selected` one day back/forward |
| `↑` / `↓` | Move `selected` one week back/forward (7 days) |
| `PgUp` | `PrevMonth()` — keep the same day-of-month where possible |
| `PgDn` | `NextMonth()` — same |
| `Ctrl+PgUp` | `PrevYear()` |
| `Ctrl+PgDn` | `NextYear()` |
| `Home` | Move `selected` to the first day of the current viewing month |
| `End` | Move `selected` to the last day of the current viewing month |
| `Enter` | Dispatch `EvtActivate` with `selected` as `time.Time` |
| `Esc` | Dispatch `EvtChange` with zero `time.Time` (cancellation signal) |

When `←`/`→`/`↑`/`↓` moves `selected` outside the current `viewing` month,
`viewing` follows automatically (scrolls to the new month). This keeps the
selected date always visible.

When moving to a disabled date (outside `minDate`/`maxDate`), the move is
silently ignored.

## Mouse interaction

- Click on a day cell: set `selected` to that date, dispatch `EvtSelect`.
  If `viewing` differs from `selected`'s month, update `viewing`.
- Double-click on a day cell: dispatch `EvtActivate`.
- Click on `◀`: `PrevMonth()`.
- Click on `▶`: `NextMonth()`.

## Events

| Event | Data | Description |
|-------|------|-------------|
| `"select"` | `time.Time` | Selected date changed by navigation |
| `"activate"` | `time.Time` | Date confirmed with Enter or double-click |
| `"change"` | `time.Time` | Viewing month changed (zero value = cancelled via Esc) |

## Rendering

### Grid calculation

```go
func (c *Calendar) firstCell() time.Time
```

Returns the date to display in the top-left cell: the `weekStart` day on or
before the first day of the `viewing` month. May be a day from the previous
month.

The 6×7 grid covers 42 cells. Cells before the month's first day and after
its last day are rendered empty (blank, using the `"calendar/overflow"` style).

### Row structure

```
Row 0: header  — "◀  March 2026  ▶"
Row 1: days    — "Mo Tu We Th Fr Sa Su"
Rows 2–7: day cells
```

**Header** (row 0):

- Left arrow `◀` at column 0 (width 1).
- Month + year string centered in the remaining width.
- Right arrow `▶` at column 19 (rightmost).
- Arrows use `"calendar/arrow"` style; `"calendar/arrow:hovered"` when the
  mouse is over them.

**Weekday row** (row 1):

- 7 two-character abbreviations separated by single spaces.
- Uses `"calendar/header"` style.
- Abbreviations come from `theme.String("calendar.weekdays")` — a
  space-separated list of 7 two-character strings in `weekStart` order
  (e.g., `"Mo Tu We Th Fr Sa Su"`).

**Day cells** (rows 2–7):

- Each cell: right-aligned 2-character number, 1-space gap to the next cell.
- Empty cells (overflow from adjacent months): blank spaces.
- Style selection per cell:

| Condition | Selector |
|-----------|----------|
| Overflow cell (outside viewing month) | `"calendar/overflow"` |
| Today | `"calendar/today"` |
| Selected date | `"calendar/selected"` |
| Selected + focused | `"calendar/selected:focused"` |
| Disabled (out of bounds) | `"calendar/day:disabled"` |
| Normal | `"calendar/day"` |

Multiple conditions can apply; priority order is: selected > today > disabled
> overflow > normal.

## Hint

Returns `(20, 8)` — always fixed. The calendar does not resize.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"calendar"` | Outer background and border |
| `"calendar/header"` | Weekday abbreviation row |
| `"calendar/arrow"` | Navigation arrows |
| `"calendar/arrow:hovered"` | Arrow on hover |
| `"calendar/day"` | Normal day cell |
| `"calendar/day:disabled"` | Disabled day cell |
| `"calendar/today"` | Cell displaying today's date |
| `"calendar/selected"` | Currently selected cell |
| `"calendar/selected:focused"` | Selected cell when widget is focused |
| `"calendar/overflow"` | Cells outside the viewing month |

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `calendar.prev` | `◀` | Previous-month arrow |
| `calendar.next` | `▶` | Next-month arrow |
| `calendar.weekdays` | `Mo Tu We Th Fr Sa Su` | Abbreviated day names, space-separated, starting from Monday |
| `calendar.weekdays.sun` | `Su Mo Tu We Th Fr Sa` | Sunday-first variant (used when `weekStart == Sunday`) |

## Implementation plan

1. **`calendar.go`** — new file
   - Define `Calendar` struct and `NewCalendar`.
   - Implement `firstCell`, `SetDate`, `Date`, `SetWeekStart`,
     `SetMinDate`, `SetMaxDate`, `PrevMonth`, `NextMonth`, `PrevYear`,
     `NextYear`.
   - Implement `handleKey`, `handleMouse` (arrow-click hit testing,
     day-cell hit testing).
   - Implement `Apply`, `Hint`, `Render` (header, weekday row, day grid).

2. **`builder.go`** — add `Calendar` method
   ```go
   func (b *Builder) Calendar(id string) *Builder
   ```

3. **Theme** — add `"calendar/*"` style entries and `calendar.*` string
   keys to built-in themes.

4. **Tests** — `calendar_test.go`
   - `firstCell` returns the correct start cell for months beginning on
     each day of the week, for both Monday-first and Sunday-first configs.
   - `←`/`→` across a month boundary updates `viewing`.
   - `PgDn` on January 31 lands on February 28/29 (last valid day).
   - `PgDn` on March 30 in a leap year lands on April 30.
   - Dates outside `minDate`/`maxDate` are not selectable.
   - `EvtActivate` carries the correct `time.Time` value.
   - Grid renders exactly 42 cells across 6 rows × 7 columns.
