package zeichenwerk

// Grow is an animated container that reveals its child widget progressively —
// either horizontally (left to right) or vertically (top to bottom) — by
// expanding a clip region one cell per tick. Once the full size is reached the
// animation stops and the child is rendered without clipping.
type Grow struct {
	Animation
	child      Widget
	horizontal bool
	step       int
	end        int
	finished   bool
}

// NewGrow creates a Grow container. When horizontal is true the child expands
// left to right; when false it expands top to bottom. Call Start to begin the
// animation.
func NewGrow(id, class string, horizontal bool) *Grow {
	grow := &Grow{
		Animation: Animation{
			Component: Component{id: id, class: class},
			stop:      make(chan struct{}),
		},
		horizontal: horizontal,
		step:       1,
	}
	grow.Animation.fn = grow.tick
	return grow
}

// Add sets the single child widget, replacing any previous child.
func (g *Grow) Add(widget Widget, params ...any) error {
	if g.child != nil {
		g.child.SetParent(nil)
	}
	if widget != nil {
		widget.SetParent(g)
	}
	g.child = widget
	return nil
}

// Apply applies a theme's styles to the component.
func (g *Grow) Apply(theme *Theme) {
	theme.Apply(g, g.Selector("grow"))
}

// Children returns the child widget slice (empty if no child has been set).
func (g *Grow) Children() []Widget {
	if g.child != nil {
		return []Widget{g.child}
	} else {
		return []Widget{}
	}
}

// Hint returns the preferred size of the child widget including its style overhead.
func (g *Grow) Hint() (int, int) {
	if g.hwidth != 0 || g.hheight != 0 {
		return g.hwidth, g.hheight
	}
	w, h := g.child.Hint()
	style := g.child.Style()
	w += style.Horizontal()
	h += style.Vertical()
	return w, h
}

// Layout positions the child to fill the Grow widget's full bounds and records
// the animation end position.
func (g *Grow) Layout() error {
	if g.child != nil {
		cx, cy, cw, ch := g.Bounds()
		g.child.SetBounds(cx, cy, cw, ch)
		if g.horizontal {
			g.end = g.width
		} else {
			g.end = g.height
		}
	}
	return Layout(g)
}

// Render draws the child clipped to the current animation step. Once the
// animation finishes the child is rendered without clipping.
func (g *Grow) Render(r *Renderer) {
	if !g.finished {
		if g.ticker == nil {
			return
		}
		if g.horizontal {
			r.Clip(g.x, g.y, g.step, g.height)
		} else {
			r.Clip(g.x, g.y, g.width, g.step)
		}
		r.Translate(-g.x, -g.y)
		g.child.Render(r)
		r.Clip(0, 0, 0, 0)
		r.Translate(0, 0)
	} else {
		g.child.Render(r)
	}
}

func (g *Grow) tick() {
	g.step++
	if g.step > g.end {
		g.Stop()
		g.finished = true
	} else {
		Redraw(g)
	}
}
