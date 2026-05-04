# Clock

Live wall-clock display. Re-renders on a fixed interval, formatting `time.Now()` with a Go time layout.

**Constructor:** `NewClock(id, class string, interval time.Duration, params ...string) *Clock`

`params[0]` is the Go time-layout (default `"15:04"`); `params[1]` is a prefix prepended to the time string (default `""`).

## Methods

- `Start()` — begin ticking using the interval given at construction
- `Stop()` — stop the animation goroutine (inherited from `Animation`)
- `Tick()` — called once per frame; triggers a redraw

## Notes

Embeds `Animation`. Width hint is computed once from the format layout (the Go reference time formatted with the user's layout) so the widget reserves a stable width that doesn't oscillate as digits change.

```go
c := zw.NewClock("now", "", time.Second, "15:04:05", " ")
c.Start()
// … later, on shutdown:
c.Stop()
```
