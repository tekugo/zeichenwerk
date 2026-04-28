# Notification

A structured message widget with an icon, title, and body text. Suitable for
toast-style alerts, informational prompts, or confirmation messages. Notifications
auto-dismiss after a caller-specified duration, or persist until the user closes
them. Every notification is appended to a package-level log.

Notifications are typically displayed as popup layers stacked in a corner, each
wrapped in `Grow(horizontal=true)` for an animated reveal. They can also be
placed statically in any layout.

---

## Visual structure

```
╭──────────────────────────────────────────────────╮
│ ●  Title text                               [×]  │
│    Body text that describes the event in         │
│    more detail, word-wrapped to width.           │
╰──────────────────────────────────────────────────╯
```

- **Icon column** (3 chars wide) — level icon drawn in the accent colour.
- **Title** — first line of the text area, rendered bold.
- **Body** — subsequent lines, word-wrapped to the available width.
- **Close button** `[×]` — rendered top-right when `closeable` is true.

The border and icon accent colour change per level:

| Level   | Accent    | Icon |
|---------|-----------|------|
| info    | `$blue`   | `ℹ`  |
| success | `$green`  | `✓`  |
| warning | `$yellow` | `⚠`  |
| error   | `$red`    | `✗`  |

---

## Notification levels

```go
type NotificationLevel int

const (
    LevelInfo    NotificationLevel = iota
    LevelSuccess
    LevelWarning
    LevelError
)
```

---

## Structure

```go
type Notification struct {
    Animation                  // dismiss timer via Start/Stop
    level     NotificationLevel
    icon      string           // overrides theme default when non-empty
    title     string
    text      string
    closeable bool             // whether to show the close button
}
```

`Animation` is embedded for its goroutine-safe ticker. A single tick at the
configured duration fires `dismiss()`. When `duration == 0` the timer is never
started.

---

## Constructor

```go
func NewNotification(id, class string) *Notification
```

Returns a notification at `LevelInfo` with no title, no text, `closeable = true`,
and no auto-dismiss timer. Use the setters below to configure it before display.

### Setters

```go
func (n *Notification) SetLevel(level NotificationLevel) *Notification
func (n *Notification) SetIcon(icon string) *Notification   // overrides theme icon
func (n *Notification) SetTitle(title string) *Notification
func (n *Notification) SetText(text string) *Notification
func (n *Notification) SetCloseable(v bool) *Notification
```

All setters return the receiver for chaining.

---

## Showing a notification

### Via UI helper

```go
func (ui *UI) Notify(level NotificationLevel, title, text string, duration time.Duration) *Notification
```

1. Creates a `Notification` with the given parameters.
2. Wraps it in `Grow(horizontal=true)` and calls `grow.Start()`.
3. Computes the stacking position (see *Stacking* below) and calls
   `ui.Popup(x, y, w, h, grow)`.
4. If `duration > 0`, calls `notification.Start(duration)`.
5. Appends an entry to the notification log.
6. Returns the `*Notification` for caller use (e.g. to dismiss early).

### Standalone placement

A notification can be added to any layout directly. In this case the caller is
responsible for positioning, wrapping in `Grow` if desired, and starting the
timer.

```go
n := NewNotification("warn", "").
    SetLevel(LevelWarning).
    SetTitle("Unsaved changes").
    SetText("Your work has not been saved.").
    SetCloseable(true)

n.Start(10 * time.Second)
flex.Add(n)
```

---

## Stacking

`UI` tracks active popup notifications:

```go
notifications []*Notification  // bottom-to-top order
```

Y position of notification at index `i` (bottom-right placement):

```
x = ui.width  - notificationWidth  - rightMargin
y = ui.height - (i+1) * (notificationHeight + gap) - bottomMargin
```

`notificationWidth` defaults to 48 characters; `rightMargin` and
`bottomMargin` default to 1; `gap` between notifications defaults to 0.

When a notification is dismissed, any above it shift downward by one slot.

---

## Auto-dismiss and close button

When `duration > 0`, the embedded `Animation` fires a single tick after the
duration and calls `dismiss()`.

`dismiss()`:
1. Stops the animation goroutine (prevents races with manual close).
2. Removes this notification from `ui.notifications` (if present).
3. Shifts stacked notifications above it downward.
4. Removes the popup layer.
5. Calls `ui.Refresh()`.
6. Dispatches `EvtHide`.

The close button `[×]` is drawn at `(content right - 3, content top)`. A
`Button1` click anywhere within it calls `dismiss()`. When `closeable = false`
the button is omitted and that column is used for body text.

If `duration == 0` and `closeable = false` the notification is permanent — the
caller must call `n.Dismiss()` programmatically.

---

## Notification log

Every notification displayed via `ui.Notify` or `n.Log()` is appended to a
package-level log. The log is never automatically trimmed.

```go
type NotificationEntry struct {
    Time  time.Time
    Level NotificationLevel
    Icon  string
    Title string
    Text  string
}

func NotificationLog() []NotificationEntry  // returns a copy of the log
func ClearNotificationLog()                 // empties the log
```

`Notification.Log()` can be called manually when a notification is placed
statically (without `ui.Notify`) to ensure it appears in the log.

---

## Hint

```go
func (n *Notification) Hint() (int, int)
```

- Width: manually set hint, or 0 (fills parent — natural when inside Grow or a
  full-width Flex).
- Height: `2 + wrappedLines(n.text, availableWidth)`, where `availableWidth =
  hintWidth - iconWidth - closeWidth - style.Horizontal()`. Falls back to a
  minimum of 2 (title row + one text row) when hint width is not yet known.

---

## Rendering

```go
func (n *Notification) Render(r *Renderer)
```

1. `n.Component.Render(r)` — draws background and border in the level style.
2. Compute content area `(cx, cy, cw, ch)`.
3. Reserve `closeW = 4` on the right when `closeable`; available text width
   `tw = cw - 3 - closeW`.
4. Resolve the icon string: `n.icon` if non-empty, else `theme.String("notification.<level>")`.
5. Draw icon at `(cx, cy)` using the `"notification/icon"` style.
6. Draw title at `(cx+3, cy)` using the `"notification/title"` style,
   truncated to `tw`.
7. Word-wrap `n.text` to `tw`; draw each line at `(cx+3, cy+1+i)`, stopping
   at `cy+ch-1`.
8. If `closeable`: draw `[×]` at `(cx+cw-3, cy)` using the
   `"notification/close"` style.

---

## Events

| Event     | Data | Description |
|-----------|------|-------------|
| `EvtHide` | —    | Notification dismissed (timer or close button) |

---

## Keyboard interaction

The notification is focusable when `closeable = true`.

| Key     | Behaviour          |
|---------|--------------------|
| `Esc`   | `dismiss()`        |
| `Enter` | `dismiss()`        |

---

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"notification"` | Outer border and background (all levels) |
| `"notification:info"` | Border and accent colour for info level |
| `"notification:success"` | Border and accent colour for success level |
| `"notification:warning"` | Border and accent colour for warning level |
| `"notification:error"` | Border and accent colour for error level |
| `"notification/icon"` | Icon cell (typically bold accent foreground) |
| `"notification/title"` | Title line (typically bold) |
| `"notification/close"` | The `[×]` button |
| `"notification/close:hovered"` | `[×]` on mouse hover |

Example theme entries (Tokyo Night):

```go
NewStyle("notification").WithBorder("round").WithPadding(0, 1),
NewStyle("notification:info").    WithColors("$fg0", "$bg1").WithForeground("$blue"),
NewStyle("notification:success"). WithColors("$fg0", "$bg1").WithForeground("$green"),
NewStyle("notification:warning"). WithColors("$fg0", "$bg1").WithForeground("$yellow"),
NewStyle("notification:error").   WithColors("$fg0", "$bg1").WithForeground("$red"),
NewStyle("notification/icon").    WithFont("bold"),
NewStyle("notification/title").   WithFont("bold"),
NewStyle("notification/close").   WithColors("$fg1", "$bg1"),
NewStyle("notification/close:hovered"). WithColors("$fg0", "$red"),
```

The border colour inherits from the level foreground, giving each level a
distinctly coloured frame.

---

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `notification.info`    | `ℹ` | Icon for info level |
| `notification.success` | `✓` | Icon for success level |
| `notification.warning` | `⚠` | Icon for warning level |
| `notification.error`   | `✗` | Icon for error level |

---

## Builder / Compose usage

```go
// Builder
builder.Notification("n1", LevelSuccess, "Saved", "File written to disk.", 3*time.Second)

// Compose
compose.Notification("n1", "", LevelWarning, "Low disk space", "Less than 1 GB remaining.", 0)
```

Both wrap the notification in `Grow(horizontal=true)` automatically.

---

## Implementation plan

1. **`notification.go`** — new file
   - Define `NotificationLevel`, `NotificationEntry`.
   - Define package-level `notificationLog []NotificationEntry`,
     `NotificationLog()`, `ClearNotificationLog()`.
   - Define `Notification` struct embedding `Animation`.
   - Implement `NewNotification`, setters, `Log()`, `Dismiss()`.
   - Implement `Apply`, `Hint`, `Render`, `handleKey`, `handleMouse`.

2. **`ui.go`** — extend `UI`
   - Add `notifications []*Notification` field.
   - Implement `Notify(level, title, text string, duration)`.
   - Clean up `notifications` slice when a layer is popped externally.

3. **Builder / Compose** — add `Notification` method to both.

4. **Theme** — add `"notification"` style family and `notification.*` string
   keys to all built-in themes.

5. **Tests** — `notification_test.go`
   - `Notify` adds a popup layer and logs an entry.
   - Timer fires `dismiss`, removes the layer and dispatches `EvtHide`.
   - Two notifications stack at correct Y positions.
   - Dismissing the lower notification shifts the upper one.
   - `closeable = false` omits the `[×]` column.
   - `duration == 0` leaves the notification until `Dismiss()` is called.
   - `SetIcon` overrides the theme icon in rendering.
   - `Log()` appends even when called on a standalone notification.
