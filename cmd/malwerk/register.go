package main

// Register holds a single rectangular yank.
type Register struct {
	W, H  int
	Cells [][]Cell
}

// Yank copies the inclusive rectangle (x1, y1)-(x2, y2) into the
// register.
func (r *Register) Yank(d *Document, x1, y1, x2, y2 int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	w := x2 - x1 + 1
	h := y2 - y1 + 1
	r.W, r.H = w, h
	r.Cells = make([][]Cell, h)
	for dy := range h {
		row := make([]Cell, w)
		for dx := range w {
			row[dx] = d.At(x1+dx, y1+dy)
		}
		r.Cells[dy] = row
	}
}

// Put pastes the register at (x, y) — the register's top-left aligns
// with the given cell. Out-of-bounds cells are clipped. Records edits
// to history.
func (r *Register) Put(d *Document, e *Editor, x, y int) {
	if r.W == 0 || r.H == 0 {
		return
	}
	hist := e.app.history
	hist.Begin()
	for dy := range r.H {
		for dx := range r.W {
			tx, ty := x+dx, y+dy
			if tx < 0 || ty < 0 || tx >= d.Width || ty >= d.Height {
				continue
			}
			before := d.Cells[ty][tx]
			after := r.Cells[dy][dx]
			if before == after {
				continue
			}
			hist.Record(tx, ty, before, after)
			d.Cells[ty][tx] = after
		}
	}
	hist.Commit()
	d.Dirty = true
}
