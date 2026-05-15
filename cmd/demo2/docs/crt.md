# CRT

Single-child container that wraps the UI in a Matrix-style power-on / power-off animation. During the boot animation a horizontal band of green character rain expands from the centre line until the full height is visible; during shutdown it contracts back. In between, CRT is an invisible pass-through wrapper.

**Constructor:** `NewCRT(id, class string) *CRT`

## Methods

- `Add(widget Widget) error` — set the wrapped child (single-child container)
- `Children() []Widget` — returns the wrapped child
- `Start(interval time.Duration)` — begin the power-on animation
- `PowerOff(interval time.Duration, onDone func())` — run the power-off animation, then invoke `onDone` (typically `ui.Quit`)
- `Layout() error` — positions child to fill the CRT bounds; records animation end position
- `Render(r *Renderer)` — clips the child to the current animation band and overlays the scanline / phosphor effect

## Notes

Wrap the root container with CRT, then call `Start` after `ui.Layout()` has run so screen dimensions are known. The animation areas are filled with a Matrix-style green character rain that grows brighter toward the scan edge, flanked by a flashing phosphor scanline. Rows of the child content fade from green monochrome to true colour after they are revealed (power-on) or before they are swallowed (power-off).

```go
crt := zw.NewCRT("crt", "")
crt.Add(myRootContainer)
ui := zw.NewUI(theme, crt)
crt.Start(30 * time.Millisecond)
// later, wired to a quit shortcut:
//   crt.PowerOff(30*time.Millisecond, ui.Quit)
ui.Run()
```
