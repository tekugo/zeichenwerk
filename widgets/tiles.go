package widgets

import (
	"slices"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/core"
)

// ==== AI ===================================================================

// Tiles is a scrollable 2D grid widget where every item occupies a fixed
// tileWidth × tileHeight cell. Items fill left-to-right, wrapping to the next
// row when the viewport width is exhausted. Rendering is delegated to an
// ItemRender function — one function, repositioned per visible cell.
//
// The number of columns is computed at render time from the content width:
//
//	cols = max(1, contentWidth / tileWidth)
//
// Moving left or right in reading order wraps between rows: going right past
// the last column advances to the first column of the next row, and vice versa.
// Moving up or down keeps the current column position.
type Tiles struct {
	Component
	render      ItemRender // Render function for each tile slot
	items       []any      // Data items
	disabled    []int      // Non-selectable item indices
	tileWidth   int        // Fixed tile width in columns (>= 1)
	tileHeight  int        // Fixed tile height in rows (>= 1)
	index       int        // Highlighted item index (-1 if empty)
	offsetRow   int        // First visible row index (0-based)
	scrollbar   bool       // Whether to draw a vertical scrollbar
	defaultCols int        // Hint column count when content width unknown
}

// NewTiles creates a new Tiles widget.
//
// Parameters:
//   - id, class: widget identity and styling class
//   - render: called once per visible tile
//   - tileWidth, tileHeight: fixed cell dimensions; panics if either < 1
func NewTiles(id, class string, render ItemRender, tileWidth, tileHeight int) *Tiles {
	if tileWidth < 1 || tileHeight < 1 {
		panic("tiles: tileWidth and tileHeight must be >= 1")
	}
	t := &Tiles{
		Component:   Component{id: id, class: class},
		render:      render,
		tileWidth:   tileWidth,
		tileHeight:  tileHeight,
		index:       -1,
		scrollbar:   true,
		defaultCols: 4,
	}
	t.SetFlag(FlagFocusable, true)
	OnKey(t, t.handleKey)
	OnMouse(t, t.handleMouse)
	return t
}

// ---- Widget Methods --------------------------------------------------------

// Apply applies the tiles' theme styles.
func (t *Tiles) Apply(theme *Theme) {
	theme.Apply(t, t.Selector("tiles"), "disabled", "focused")
}

// Hint returns the preferred size.
// Width: tileWidth × defaultCols (+ border overhead).
// Height: rows × tileHeight (+ border overhead), based on the tile count and
// defaultCols — this collapses to the actual row count once layout is known.
func (t *Tiles) Hint() (int, int) {
	if t.hwidth != 0 || t.hheight != 0 {
		return t.hwidth, t.hheight
	}
	s := t.Style()
	dc := t.defaultCols
	if dc < 1 {
		dc = 1
	}
	rows := (len(t.items) + dc - 1) / dc
	if rows < 1 {
		rows = 1
	}
	return t.tileWidth*dc + s.Horizontal(), rows*t.tileHeight + s.Vertical()
}

// ---- Derived values --------------------------------------------------------

// cols returns the number of tile columns for the current content width.
// It is only valid after layout (Content() returns non-zero width).
func (t *Tiles) cols() int {
	_, _, cw, _ := t.Content()
	if cw < t.tileWidth {
		return 1
	}
	return cw / t.tileWidth
}

// rows returns the total number of tile rows for the current item count.
func (t *Tiles) rows() int {
	c := t.cols()
	if len(t.items) == 0 {
		return 0
	}
	return (len(t.items) + c - 1) / c
}

// row returns the grid row of the item at index.
func (t *Tiles) row(index int) int {
	return index / t.cols()
}

// col returns the grid column of the item at index.
func (t *Tiles) col(index int) int {
	return index % t.cols()
}

// ---- Data ------------------------------------------------------------------

// SetItems replaces all items, resets index to 0 (or -1 if empty) and
// offsetRow to 0, then redraws.
func (t *Tiles) SetItems(items []any) {
	t.items = items
	t.offsetRow = 0
	if len(items) == 0 {
		t.index = -1
	} else {
		t.index = 0
	}
	Redraw(t)
}

// Items returns the current items slice.
func (t *Tiles) Items() []any { return t.items }

// SetDisabled replaces the list of non-selectable item indices.
func (t *Tiles) SetDisabled(indices []int) { t.disabled = indices }

// ---- Selection -------------------------------------------------------------

// Selected returns the currently highlighted item index (-1 if none).
func (t *Tiles) Selected() int { return t.index }

// Select highlights the item at index, adjusts the scroll offset, and
// dispatches EvtSelect.
func (t *Tiles) Select(index int) {
	if index < 0 || index >= len(t.items) {
		return
	}
	t.index = index
	t.adjust()
	t.Dispatch(t, EvtSelect, t.index)
	Redraw(t)
}

// ---- Navigation ------------------------------------------------------------

// First highlights the first enabled item.
func (t *Tiles) First() {
	for i := range t.items {
		if !slices.Contains(t.disabled, i) {
			t.index = i
			t.adjust()
			t.Dispatch(t, EvtSelect, t.index)
			Redraw(t)
			return
		}
	}
}

// Last highlights the last enabled item.
func (t *Tiles) Last() {
	for i := len(t.items) - 1; i >= 0; i-- {
		if !slices.Contains(t.disabled, i) {
			t.index = i
			t.adjust()
			t.Dispatch(t, EvtSelect, t.index)
			Redraw(t)
			return
		}
	}
}

// Move moves the highlight by dr rows and dc columns.
//
// Column navigation (dc != 0) uses reading order: moving right past the last
// column of a row advances to the first column of the next row, and vice versa.
// Row navigation (dr != 0) keeps the current column, clamping if the target
// cell would be past the end of items.
//
// Disabled items are skipped. The highlight is clamped to valid bounds.
func (t *Tiles) Move(dr, dc int) {
	if len(t.items) == 0 || (dr == 0 && dc == 0) {
		return
	}
	if t.index < 0 {
		t.First()
		return
	}

	c := t.cols()
	newIndex := t.index

	if dc != 0 {
		// Reading-order (flat) traversal.
		dir := 1
		if dc < 0 {
			dir = -1
		}
		steps := dc
		if steps < 0 {
			steps = -steps
		}
		for i := 0; i < steps; i++ {
			newIndex = t.skipFlat(newIndex+dir, dir)
		}
	} else {
		// Column-preserving row navigation.
		curCol := t.index % c
		dir := 1
		if dr < 0 {
			dir = -1
		}
		steps := dr
		if steps < 0 {
			steps = -steps
		}
		for i := 0; i < steps; i++ {
			targetRow := newIndex/c + dir
			targetCol := newIndex % c
			candidate := targetRow*c + targetCol
			if candidate < 0 {
				// Clamp at first row — keep column
				candidate = curCol
			} else if candidate >= len(t.items) {
				// Clamp at last item
				candidate = len(t.items) - 1
			}
			newIndex = t.skipRow(candidate, dir)
		}
	}

	if newIndex != t.index && newIndex >= 0 {
		t.index = newIndex
		t.adjust()
		t.Dispatch(t, EvtSelect, t.index)
		Redraw(t)
	}
}

// PageDown moves the highlight down by the number of fully visible rows.
func (t *Tiles) PageDown() {
	_, _, _, ch := t.Content()
	rows := ch / t.tileHeight
	if rows < 1 {
		rows = 1
	}
	t.Move(rows, 0)
}

// PageUp moves the highlight up by the number of fully visible rows.
func (t *Tiles) PageUp() {
	_, _, _, ch := t.Content()
	rows := ch / t.tileHeight
	if rows < 1 {
		rows = 1
	}
	t.Move(-rows, 0)
}

// ---- Internal helpers ------------------------------------------------------

// skipFlat advances from index in direction (±1) in reading order, skipping
// disabled items. Clamps at list boundaries.
func (t *Tiles) skipFlat(index, direction int) int {
	n := len(t.items)
	next := index
	for next >= 0 && next < n && slices.Contains(t.disabled, next) {
		next += direction
	}
	if next < 0 {
		next = 0
	}
	if next >= n {
		next = n - 1
	}
	return next
}

// skipRow advances from index by rows (in direction ±1), keeping the column,
// skipping disabled items. Clamps at list boundaries.
func (t *Tiles) skipRow(index, direction int) int {
	c := t.cols()
	n := len(t.items)
	next := index
	if next < 0 {
		next = 0
	}
	if next >= n {
		next = n - 1
	}
	for slices.Contains(t.disabled, next) {
		candidate := next/c*c + direction*c + next%c
		if candidate < 0 || candidate >= n {
			break
		}
		next = candidate
	}
	return next
}

// adjust ensures the highlighted item's row is within the visible viewport.
func (t *Tiles) adjust() {
	_, _, _, ch := t.Content()
	if ch <= 0 || t.tileHeight <= 0 || t.index < 0 {
		return
	}
	visibleRows := ch / t.tileHeight
	if visibleRows < 1 {
		visibleRows = 1
	}
	itemRow := t.row(t.index)
	if itemRow < t.offsetRow {
		t.offsetRow = itemRow
	} else if itemRow >= t.offsetRow+visibleRows {
		t.offsetRow = itemRow - visibleRows + 1
	}
	if t.offsetRow < 0 {
		t.offsetRow = 0
	}
	maxOffset := max(t.rows()-visibleRows, 0)
	if t.offsetRow > maxOffset {
		t.offsetRow = maxOffset
	}
}

// ---- Event Handlers --------------------------------------------------------

func (t *Tiles) handleKey(event *tcell.EventKey) bool {
	switch event.Key() {
	case tcell.KeyLeft:
		t.Move(0, -1)
		return true
	case tcell.KeyRight:
		t.Move(0, +1)
		return true
	case tcell.KeyUp:
		t.Move(-1, 0)
		return true
	case tcell.KeyDown:
		t.Move(+1, 0)
		return true
	case tcell.KeyPgUp:
		t.PageUp()
		return true
	case tcell.KeyPgDn:
		t.PageDown()
		return true
	case tcell.KeyHome:
		t.First()
		return true
	case tcell.KeyEnd:
		t.Last()
		return true
	case tcell.KeyEnter:
		if t.index >= 0 {
			t.Dispatch(t, EvtActivate, t.index)
		}
		return true
	}
	return false
}

func (t *Tiles) handleMouse(event *tcell.EventMouse) bool {
	if event.Buttons() != tcell.Button1 {
		return false
	}
	mx, my := event.Position()
	cx, cy, cw, ch := t.Content()
	if mx < cx || mx >= cx+cw || my < cy || my >= cy+ch {
		return false
	}

	c := t.cols()
	clickedCol := (mx - cx) / t.tileWidth
	clickedRow := (my-cy)/t.tileHeight + t.offsetRow
	clickedIndex := clickedRow*c + clickedCol

	if clickedIndex < 0 || clickedIndex >= len(t.items) {
		return false
	}
	if slices.Contains(t.disabled, clickedIndex) {
		return false
	}

	if clickedIndex == t.index {
		t.Dispatch(t, EvtActivate, clickedIndex)
	} else {
		t.Select(clickedIndex)
	}
	return true
}

// ---- Rendering -------------------------------------------------------------

func (t *Tiles) Render(r *Renderer) {
	if t.Flag(FlagHidden) {
		return
	}
	t.Component.Render(r)

	cx, cy, cw, ch := t.Content()
	if cw <= 0 || ch <= 0 || t.render == nil {
		return
	}

	// Reserve rightmost column for scrollbar when content exceeds viewport.
	tw := cw
	if t.scrollbar && t.rows()*t.tileHeight > ch {
		tw = cw - 1
	}

	c := max(1, tw/t.tileWidth)
	visibleRows := ch / t.tileHeight
	focused := t.Flag(FlagFocused)

	for row := t.offsetRow; row < t.offsetRow+visibleRows; row++ {
		for col := 0; col < c; col++ {
			itemIndex := row*c + col
			if itemIndex >= len(t.items) {
				break
			}
			slotX := cx + col*t.tileWidth
			slotY := cy + (row-t.offsetRow)*t.tileHeight
			t.render(r, slotX, slotY, t.tileWidth, t.tileHeight, itemIndex, t.items[itemIndex], itemIndex == t.index, focused)
		}
	}

	if t.scrollbar && t.rows()*t.tileHeight > ch {
		r.ScrollbarV(cx+tw, cy, ch, t.offsetRow*t.tileHeight, t.rows()*t.tileHeight)
	}
}
