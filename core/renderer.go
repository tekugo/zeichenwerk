package core

import "github.com/tekugo/zeichenwerk/renderer"

// Renderer is an extension of renderer.Renderer to add theme support for
// borders and colors. Due to dependencies, renderer.Renderer cannot use
// Border or Theme from this package.
type Renderer struct {
	renderer.Renderer
	theme *Theme
}

// NewRenderer creates a new themable renderer.
func NewRenderer(screen renderer.Screen, theme *Theme) *Renderer {
	return &Renderer{
		Renderer: renderer.Renderer{Screen: screen},
		theme:    theme,
	}
}

// ---- Additional Rendering Operations ----

// Border draws a complete border around a rectangular area using the specified BorderStyle.
// This method renders the four sides and corners of a border, creating a frame around
// the given coordinates using the border characters.
//
// Parameters:
//   - x, y: Top-left corner coordinates of the border area
//   - w, h: Width and height of the area to border (inner dimensions)
//   - box: Border containing the characters for each border element
func (r *Renderer) Border(x, y, w, h int, border string) {
	b := r.theme.Border(border)
	r.Line(x, y, 1, 0, w-2, b.TopLeft, b.Top, b.TopRight)
	r.Line(x, y+h-1, 1, 0, w-2, b.BottomLeft, b.Bottom, b.BottomRight)
	for i := range h - 2 {
		r.Put(x, y+i+1, b.Left)
		r.Put(x+w-1, y+i+1, b.Right)
	}
}

// Set sets the foreground colour, background colour, and font for subsequent
// drawing operations. Colour strings starting with "$" are resolved through
// the theme's colour registry.
func (r *Renderer) Set(foreground, background, font string) {
	r.Renderer.Set(r.theme.Color(foreground), r.theme.Color(background), font)
}

// Theme returns the renderer's theme.
func (r *Renderer) Theme() *Theme {
	return r.theme
}
