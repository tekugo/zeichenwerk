package main

// Edit is one cell-level change captured for undo / redo.
type Edit struct {
	X, Y   int
	Before Cell
	After  Cell
}

// Batch is a list of Edits performed by a single user action.
type Batch []Edit

// History is a bounded linear undo stack.
type History struct {
	max     int
	batches []Batch
	cursor  int     // points one past the most recent applied batch
	current Batch   // open batch being recorded
	open    bool    // true between Begin and Commit
}

// NewHistory creates a history with the given capacity. Older batches are
// dropped when the size exceeds max.
func NewHistory(max int) *History {
	if max < 1 {
		max = 1
	}
	return &History{max: max}
}

// Begin opens a new batch. Calling Begin while a batch is already open is
// a no-op.
func (h *History) Begin() {
	if h.open {
		return
	}
	h.current = h.current[:0]
	h.open = true
}

// Record adds a single cell delta to the open batch. Begin must have been
// called first; otherwise the call is dropped.
func (h *History) Record(x, y int, before, after Cell) {
	if !h.open {
		return
	}
	if before == after {
		return
	}
	h.current = append(h.current, Edit{X: x, Y: y, Before: before, After: after})
}

// Commit closes the open batch and pushes it onto the stack. An empty
// batch is discarded so undo doesn't no-op.
func (h *History) Commit() {
	if !h.open {
		return
	}
	h.open = false
	if len(h.current) == 0 {
		return
	}
	// Discard any redo branch.
	h.batches = h.batches[:h.cursor]
	batch := make(Batch, len(h.current))
	copy(batch, h.current)
	h.batches = append(h.batches, batch)
	h.cursor++
	if len(h.batches) > h.max {
		drop := len(h.batches) - h.max
		h.batches = h.batches[drop:]
		h.cursor -= drop
		if h.cursor < 0 {
			h.cursor = 0
		}
	}
}

// Undo reverses the most recent batch on the document.
func (h *History) Undo(d *Document, e *Editor) {
	if h.cursor == 0 {
		return
	}
	h.cursor--
	batch := h.batches[h.cursor]
	for i := len(batch) - 1; i >= 0; i-- {
		ed := batch[i]
		d.Cells[ed.Y][ed.X] = ed.Before
	}
	d.Dirty = true
	e.Refresh()
	e.app.refreshStatus()
}

// Redo re-applies the next batch.
func (h *History) Redo(d *Document, e *Editor) {
	if h.cursor >= len(h.batches) {
		return
	}
	batch := h.batches[h.cursor]
	h.cursor++
	for _, ed := range batch {
		d.Cells[ed.Y][ed.X] = ed.After
	}
	d.Dirty = true
	e.Refresh()
	e.app.refreshStatus()
}
