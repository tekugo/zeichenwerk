package core

import "github.com/tekugo/zeichenwerk/v2/renderer"

// Renderer wraps the primitive drawing API from the renderer package and
// layers theme-aware operations on top of it: symbolic colour names are
// resolved against the theme's colour registry, and border strokes are
// looked up by name. The embedded renderer.Renderer remains accessible, so
// all low-level drawing primitives continue to work unchanged.
//
// The wrapper exists because renderer.Renderer cannot import this package
// (it would introduce an import cycle with Border and Theme), so theme
// resolution is done here at the boundary between the two layers.
type Renderer struct {
	Theme *Theme
	renderer.Renderer
}

// NewRenderer constructs a theme-aware renderer bound to the given screen
// and theme. The screen provides the raw cell-writing primitives while the
// theme supplies colour variables and named border glyph sets.
func NewRenderer(screen renderer.Screen, theme *Theme) *Renderer {
	return &Renderer{
		Renderer: renderer.Renderer{Screen: screen},
		Theme:    theme,
	}
}

// ---- Additional Rendering Operations ----

// Border draws a complete rectangular frame inside the given outer bounds
// using the border named in the theme. The four corners and both horizontal
// strokes are emitted via Line calls, and the vertical sides are written
// one cell per row between the corners. The content area is the inner
// rectangle at (x+1, y+1) with size (w-2, h-2); sizes smaller than 2×2
// produce a degenerate border and should be avoided by callers.
//
// Parameters:
//   - x, y:   Top-left corner of the border in screen coordinates.
//   - w, h:   Outer width and height of the framed rectangle.
//   - border: Name of the border style to look up in the theme.
func (r *Renderer) Border(x, y, w, h int, border string) {
	b := r.Theme.Border(border)
	r.Line(x, y, 1, 0, w-2, b.TopLeft, b.Top, b.TopRight)
	r.Line(x, y+h-1, 1, 0, w-2, b.BottomLeft, b.Bottom, b.BottomRight)
	for i := range h - 2 {
		r.Put(x, y+i+1, b.Left)
		r.Put(x+w-1, y+i+1, b.Right)
	}
}

// Set configures the foreground colour, background colour, and font used by
// subsequent drawing calls. Colour arguments starting with "$" are resolved
// through the theme's colour registry before being passed to the underlying
// renderer; literal colours (such as "#ffffff" or named terminal colours)
// are forwarded unchanged.
func (r *Renderer) Set(foreground, background, font string) {
	r.Renderer.Set(r.Theme.Color(foreground), r.Theme.Color(background), font)
}
