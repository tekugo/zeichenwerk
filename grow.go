package zeichenwerk

type Grow struct {
	Animation
	child      Widget
	horizontal bool
	step       int
	end        int
	finished   bool
}

func NewGrow(horizontal bool) *Grow {
	grow := &Grow{
		Animation: Animation{
			stop: make(chan struct{}),
		},
		horizontal: horizontal,
		step:       1,
	}
	grow.Animation.fn = grow.tick
	return grow
}

func (g *Grow) Add(widget Widget) {
	if g.child != nil {
		g.child.SetParent(nil)
	}
	if widget != nil {
		widget.SetParent(g)
	}
	g.child = widget
}

func (g *Grow) Children() []Widget {
	if g.child != nil {
		return []Widget{g.child}
	} else {
		return []Widget{}
	}
}

func (g *Grow) Hint() (int, int) {
	w, h := g.child.Hint()
	style := g.child.Style()
	w += style.Horizontal()
	h += style.Vertical()
	return w, h
}

func (g *Grow) Layout() {
	if g.child != nil {
		cx, cy, cw, ch := g.Bounds()
		g.child.SetBounds(cx, cy, cw, ch)
		if g.horizontal {
			g.end = g.width
		} else {
			g.end = g.height
		}
	}
	Layout(g)
}

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
