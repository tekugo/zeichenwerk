# Notification

A temporary overlay message that appears in the bottom-right corner and
auto-dismisses after a caller-specified duration. Added via a `UI` method;
no manual widget placement required.

## API

```go
func (ui *UI) Notify(message string, duration time.Duration, level ...NotificationLevel)
```

`level` is optional and defaults to `LevelInfo`. Multiple calls stack
notifications vertically from the bottom edge upward.

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

The level controls the style applied to the widget (`"notification:info"`,
`"notification:success"`, etc.).

## Structure

```go
type Notification struct {
    Animation              // provides the dismiss timer via Start/Stop
    message  string
    level    NotificationLevel
}
```

`Animation` is embedded for its goroutine-safe ticker. A single tick at
`duration` fires `dismiss()`.

## Lifecycle

`ui.Notify`:

1. Create a `Notification` with the given message and level.
2. Compute position: bottom-right, offset upward by `height * len(active)` to
   stack above any existing notifications.
3. Call `ui.Popup(x, y, w, h, notification)` to add it as a new layer.
   Width is `min(len(message) + padding, maxWidth)` with a minimum of 20.
   Height is 1 (single line) plus style padding.
4. Start the dismiss timer: `notification.Start(duration)`.

`dismiss()` (called from the `Animation.Tick`):

1. Remove this notification from `ui.layers`.
2. Shift any stacked notifications above it downward.
3. Call `ui.Refresh()`.

## Stacking

`UI` tracks active notifications:

```go
notifications []*Notification  // bottom-to-top order
```

New notifications are appended; dismissed ones are removed. Y position of
notification `i`:

```
y = ui.height - (i+1) * (notificationHeight + gap) - bottomMargin
```

`bottomMargin` defaults to 1; `gap` between notifications defaults to 0
(they touch).

## Rendering

```go
func (n *Notification) Render(r *Renderer)
```

1. `n.Component.Render(r)` — draws background and border in the level style.
2. Draw a level icon at the left edge:
   - Info: `ℹ`
   - Success: `✓`
   - Warning: `⚠`
   - Error: `✗`
3. Draw `n.message` after the icon, truncated to available width.

Icons are fetched from the theme via `theme.String("notification.*")` keys.

## Styling selectors

| Selector | Applied to |
|----------|-----------|
| `"notification"` | Base style (all levels) |
| `"notification:info"` | Info level |
| `"notification:success"` | Success level |
| `"notification:warning"` | Warning level |
| `"notification:error"` | Error level |

State selectors reuse the existing `:focused`/`:hovered` convention — the
notification is not focusable, so only level-based styling applies.

## Theme string keys

| Key | Default | Description |
|-----|---------|-------------|
| `notification.info` | `ℹ ` | Icon for info level |
| `notification.success` | `✓ ` | Icon for success level |
| `notification.warning` | `⚠ ` | Icon for warning level |
| `notification.error` | `✗ ` | Icon for error level |

## Events

`Notification` dispatches no events. Callers that need to react to dismissal
can wrap `ui.Notify` and check the `EvtHide` event if needed, but the
primary contract is fire-and-forget.

## Notes

- `Notification` is **not focusable**. It receives no keyboard input and does
  not interrupt the user's focus.
- Mouse clicks on a notification dismiss it immediately (calls `dismiss()`
  on `Button1` release within bounds).
- If `duration == 0`, the notification persists until clicked. This enables
  sticky error messages.
- The `Animation` goroutine is stopped in `dismiss()` before removing the
  layer, preventing a race between the timer firing and manual dismissal.

## Implementation plan

1. **`notification.go`** — new file
   - Define `NotificationLevel` type and constants.
   - Define `Notification` struct, embedding `Animation`.
   - Implement `NewNotification`, `dismiss`, `Apply`, `Hint`, `Render`,
     `handleMouse`.

2. **`ui.go`** — extend `UI`
   - Add `notifications []*Notification` field.
   - Implement `Notify(message string, duration time.Duration, level ...NotificationLevel)`.
   - Update `Close()` / layer removal to also clean up `notifications` slice
     when a notification layer is popped externally.

3. **Theme** — add `"notification"` family of style entries and
   `notification.*` string keys.

4. **Tests** — `notification_test.go`
   - `Notify` adds a layer and starts the timer.
   - Timer fires `dismiss`, which removes the layer.
   - Two notifications stack at correct Y positions.
   - Dismissing the bottom notification shifts the upper one down.
   - Click on notification dismisses it and stops the timer.
   - `duration == 0` leaves the notification until clicked.
