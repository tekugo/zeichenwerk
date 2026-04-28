package format

import (
	"sort"
	"strconv"
	"unicode/utf8"

	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/widgets"
)

// Alignment is the per-column horizontal alignment.
type Alignment uint8

const (
	AlignLeft Alignment = iota // default
	AlignCenter
	AlignRight
)

// undoCmd records one reversible mutation.
type undoCmd struct {
	do   func()
	undo func()
}

const maxUndo = 100

// MutableTable is the in-memory table model used by tblr.
// Implements zeichenwerk.TableProvider.
type MutableTable struct {
	headers    []string
	alignments []Alignment
	columns    []widgets.TableColumn
	data       [][]string
	hasHeader  bool
	modified   bool
	delimiter  rune

	undoStack []undoCmd
	undoPos   int
}

// NewMutableTable creates an empty table.
func NewMutableTable() *MutableTable {
	return &MutableTable{
		hasHeader: true,
		delimiter: ',',
		undoStack: make([]undoCmd, 0, maxUndo),
	}
}

// --- TableProvider ---

func (t *MutableTable) Columns() []widgets.TableColumn { return t.columns }
func (t *MutableTable) Length() int                    { return len(t.data) }
func (t *MutableTable) Str(row, col int) string {
	if row < 0 || row >= len(t.data) {
		return ""
	}
	if col < 0 || col >= len(t.data[row]) {
		return ""
	}
	return t.data[row][col]
}

// --- Metadata ---

func (t *MutableTable) HasHeader() bool     { return t.hasHeader }
func (t *MutableTable) SetHasHeader(v bool) { t.hasHeader = v }
func (t *MutableTable) Modified() bool      { return t.modified }
func (t *MutableTable) ClearModified()      { t.modified = false }
func (t *MutableTable) Delimiter() rune     { return t.delimiter }
func (t *MutableTable) SetDelimiter(r rune) { t.delimiter = r }
func (t *MutableTable) ColCount() int       { return len(t.headers) }
func (t *MutableTable) Header(col int) string {
	if col < 0 || col >= len(t.headers) {
		return ""
	}
	return t.headers[col]
}

func (t *MutableTable) ColAlignment(col int) Alignment {
	if col < 0 || col >= len(t.alignments) {
		return AlignLeft
	}
	return t.alignments[col]
}

func (t *MutableTable) SetColAlignment(col int, a Alignment) {
	if col < 0 || col >= len(t.alignments) {
		return
	}
	old := t.alignments[col]
	t.push(undoCmd{
		do:   func() { t.alignments[col] = a; t.syncColumns(); t.modified = true },
		undo: func() { t.alignments[col] = old; t.syncColumns(); t.modified = true },
	})
	t.alignments[col] = a
	t.syncColumns()
	t.modified = true
}

// RecalcWidths recomputes column widths from cell content.
func (t *MutableTable) RecalcWidths() {
	for i := range t.columns {
		w := utf8.RuneCountInString(t.headers[i])
		for _, row := range t.data {
			if i < len(row) {
				if n := utf8.RuneCountInString(row[i]); n > w {
					w = n
				}
			}
		}
		if w < 1 {
			w = 1
		}
		t.columns[i].Width = w
	}
	t.syncColumns()
}

// syncColumns pushes header/alignment state into the stored columns slice.
func (t *MutableTable) syncColumns() {
	for i := range t.columns {
		t.columns[i].Header = t.headers[i]
		t.columns[i].Sortable = true
		switch t.alignments[i] {
		case AlignCenter:
			t.columns[i].Alignment = core.Center
		case AlignRight:
			t.columns[i].Alignment = core.Right
		default:
			t.columns[i].Alignment = core.Left
		}
	}
}

// Load replaces all content and resets the modified flag.
func (t *MutableTable) Load(headers []string, data [][]string) {
	t.headers = make([]string, len(headers))
	copy(t.headers, headers)
	t.alignments = make([]Alignment, len(headers))
	t.columns = make([]widgets.TableColumn, len(headers))
	t.data = make([][]string, len(data))
	for i, row := range data {
		r := make([]string, len(row))
		copy(r, row)
		t.data[i] = r
	}
	t.syncColumns()
	t.modified = false
	t.undoStack = t.undoStack[:0]
	t.undoPos = 0
}

// LoadAlignments sets per-column alignments after a Load call.
func (t *MutableTable) LoadAlignments(aligns []Alignment) {
	for i, a := range aligns {
		if i < len(t.alignments) {
			t.alignments[i] = a
		}
	}
	t.syncColumns()
}

// --- Cell mutation ---

func (t *MutableTable) SetCell(row, col int, value string) {
	if row < 0 || row >= len(t.data) || col < 0 {
		return
	}
	for len(t.data[row]) <= col {
		t.data[row] = append(t.data[row], "")
	}
	old := t.data[row][col]
	t.push(undoCmd{
		do:   func() { t.data[row][col] = value; t.modified = true },
		undo: func() { t.data[row][col] = old; t.modified = true },
	})
	t.data[row][col] = value
	t.modified = true
}

// --- Row operations ---

func (t *MutableTable) InsertRowAt(at int) {
	if at < 0 {
		at = 0
	}
	if at > len(t.data) {
		at = len(t.data)
	}
	row := make([]string, len(t.headers))
	t.push(undoCmd{
		do:   func() { t.insertRow(at, row); t.modified = true },
		undo: func() { t.deleteRow(at); t.modified = true },
	})
	t.insertRow(at, row)
	t.modified = true
}

func (t *MutableTable) AppendRow() { t.InsertRowAt(len(t.data)) }

func (t *MutableTable) DeleteRow(at int) {
	if at < 0 || at >= len(t.data) {
		return
	}
	saved := make([]string, len(t.data[at]))
	copy(saved, t.data[at])
	t.push(undoCmd{
		do:   func() { t.deleteRow(at); t.modified = true },
		undo: func() { t.insertRow(at, saved); t.modified = true },
	})
	t.deleteRow(at)
	t.modified = true
}

func (t *MutableTable) MoveRow(from, to int) {
	if from == to || from < 0 || from >= len(t.data) || to < 0 || to >= len(t.data) {
		return
	}
	t.push(undoCmd{
		do:   func() { t.doMoveRow(from, to); t.modified = true },
		undo: func() { t.doMoveRow(to, from); t.modified = true },
	})
	t.doMoveRow(from, to)
	t.modified = true
}

func (t *MutableTable) insertRow(at int, row []string) {
	t.data = append(t.data, nil)
	copy(t.data[at+1:], t.data[at:])
	t.data[at] = row
}

func (t *MutableTable) deleteRow(at int) {
	t.data = append(t.data[:at], t.data[at+1:]...)
}

func (t *MutableTable) doMoveRow(from, to int) {
	row := t.data[from]
	newData := make([][]string, 0, len(t.data))
	for i, r := range t.data {
		if i == from {
			continue
		}
		newData = append(newData, r)
	}
	final := make([][]string, len(t.data))
	copy(final[:to], newData[:to])
	final[to] = row
	copy(final[to+1:], newData[to:])
	t.data = final
}

// --- Column operations ---

func (t *MutableTable) InsertColAt(at int) {
	if at < 0 {
		at = 0
	}
	if at > len(t.headers) {
		at = len(t.headers)
	}
	t.push(undoCmd{
		do:   func() { t.insertCol(at, ""); t.modified = true },
		undo: func() { t.deleteCol(at); t.modified = true },
	})
	t.insertCol(at, "")
	t.modified = true
}

func (t *MutableTable) AppendCol(header string) {
	at := len(t.headers)
	t.push(undoCmd{
		do:   func() { t.insertCol(at, header); t.modified = true },
		undo: func() { t.deleteCol(at); t.modified = true },
	})
	t.insertCol(at, header)
	t.modified = true
}

func (t *MutableTable) DeleteCol(at int) {
	if at < 0 || at >= len(t.headers) {
		return
	}
	savedHeader := t.headers[at]
	savedAlign := t.alignments[at]
	savedCells := make([]string, len(t.data))
	for i, row := range t.data {
		if at < len(row) {
			savedCells[i] = row[at]
		}
	}
	t.push(undoCmd{
		do: func() { t.deleteCol(at); t.modified = true },
		undo: func() {
			t.insertCol(at, savedHeader)
			t.alignments[at] = savedAlign
			for i := range t.data {
				if at < len(t.data[i]) {
					t.data[i][at] = savedCells[i]
				}
			}
			t.syncColumns()
			t.modified = true
		},
	})
	t.deleteCol(at)
	t.modified = true
}

func (t *MutableTable) MoveCol(from, to int) {
	if from == to || from < 0 || from >= len(t.headers) || to < 0 || to >= len(t.headers) {
		return
	}
	t.push(undoCmd{
		do:   func() { t.doMoveCol(from, to); t.modified = true },
		undo: func() { t.doMoveCol(to, from); t.modified = true },
	})
	t.doMoveCol(from, to)
	t.modified = true
}

func (t *MutableTable) RenameCol(col int, name string) {
	if col < 0 || col >= len(t.headers) {
		return
	}
	old := t.headers[col]
	t.push(undoCmd{
		do:   func() { t.headers[col] = name; t.syncColumns(); t.modified = true },
		undo: func() { t.headers[col] = old; t.syncColumns(); t.modified = true },
	})
	t.headers[col] = name
	t.syncColumns()
	t.modified = true
}

func (t *MutableTable) insertCol(at int, header string) {
	t.headers = append(t.headers, "")
	copy(t.headers[at+1:], t.headers[at:])
	t.headers[at] = header

	t.alignments = append(t.alignments, AlignLeft)
	copy(t.alignments[at+1:], t.alignments[at:])
	t.alignments[at] = AlignLeft

	w := utf8.RuneCountInString(header)
	if w < 1 {
		w = 1
	}
	t.columns = append(t.columns, widgets.TableColumn{})
	copy(t.columns[at+1:], t.columns[at:])
	t.columns[at] = widgets.TableColumn{Header: header, Width: w, Sortable: true}

	for i, row := range t.data {
		nr := make([]string, len(row)+1)
		copy(nr[:at], row[:at])
		copy(nr[at+1:], row[at:])
		t.data[i] = nr
	}
}

func (t *MutableTable) deleteCol(at int) {
	t.headers = append(t.headers[:at], t.headers[at+1:]...)
	t.alignments = append(t.alignments[:at], t.alignments[at+1:]...)
	t.columns = append(t.columns[:at], t.columns[at+1:]...)
	for i, row := range t.data {
		if at < len(row) {
			t.data[i] = append(row[:at], row[at+1:]...)
		}
	}
}

func (t *MutableTable) doMoveCol(from, to int) {
	h := t.headers[from]
	t.headers = append(t.headers[:from], t.headers[from+1:]...)
	t.headers = append(t.headers, "")
	copy(t.headers[to+1:], t.headers[to:])
	t.headers[to] = h

	a := t.alignments[from]
	t.alignments = append(t.alignments[:from], t.alignments[from+1:]...)
	t.alignments = append(t.alignments, 0)
	copy(t.alignments[to+1:], t.alignments[to:])
	t.alignments[to] = a

	c := t.columns[from]
	t.columns = append(t.columns[:from], t.columns[from+1:]...)
	t.columns = append(t.columns, widgets.TableColumn{})
	copy(t.columns[to+1:], t.columns[to:])
	t.columns[to] = c

	for i, row := range t.data {
		if from >= len(row) {
			continue
		}
		val := row[from]
		newRow := make([]string, len(row))
		copy(newRow[:from], row[:from])
		copy(newRow[from:], row[from+1:])
		newRow[len(newRow)-1] = ""
		// re-insert at to
		nr := make([]string, len(row))
		copy(nr[:to], newRow[:to])
		nr[to] = val
		copy(nr[to+1:], newRow[to:])
		t.data[i] = nr
	}
}

// --- Sort ---

func (t *MutableTable) SortByCol(col int, asc bool) {
	if col < 0 || col >= len(t.headers) || len(t.data) == 0 {
		return
	}
	snapshot := make([][]string, len(t.data))
	for i, row := range t.data {
		r := make([]string, len(row))
		copy(r, row)
		snapshot[i] = r
	}

	numeric := true
	for _, row := range t.data {
		if col >= len(row) || row[col] == "" {
			continue
		}
		if _, err := strconv.ParseFloat(row[col], 64); err != nil {
			numeric = false
			break
		}
	}

	t.push(undoCmd{
		do: func() { t.doSort(col, asc, numeric); t.modified = true },
		undo: func() {
			for i, row := range snapshot {
				r := make([]string, len(row))
				copy(r, row)
				t.data[i] = r
			}
			t.modified = true
		},
	})
	t.doSort(col, asc, numeric)
	t.modified = true
}

func (t *MutableTable) doSort(col int, asc, numeric bool) {
	sort.SliceStable(t.data, func(i, j int) bool {
		var a, b string
		if col < len(t.data[i]) {
			a = t.data[i][col]
		}
		if col < len(t.data[j]) {
			b = t.data[j][col]
		}
		var less bool
		if numeric {
			fa, ea := strconv.ParseFloat(a, 64)
			fb, eb := strconv.ParseFloat(b, 64)
			switch {
			case ea != nil && eb != nil:
				less = a < b
			case ea != nil:
				less = false
			case eb != nil:
				less = true
			default:
				less = fa < fb
			}
		} else {
			less = a < b
		}
		if asc {
			return less
		}
		return !less
	})
}

// --- Undo / Redo ---

func (t *MutableTable) push(cmd undoCmd) {
	t.undoStack = t.undoStack[:t.undoPos]
	if len(t.undoStack) >= maxUndo {
		t.undoStack = t.undoStack[1:]
		t.undoPos--
	}
	t.undoStack = append(t.undoStack, cmd)
	t.undoPos++
}

// Undo undoes the last command; returns false if nothing to undo.
func (t *MutableTable) Undo() bool {
	if t.undoPos == 0 {
		return false
	}
	t.undoPos--
	t.undoStack[t.undoPos].undo()
	return true
}

// Redo re-applies the last undone command; returns false if nothing to redo.
func (t *MutableTable) Redo() bool {
	if t.undoPos >= len(t.undoStack) {
		return false
	}
	t.undoStack[t.undoPos].do()
	t.undoPos++
	return true
}

// Headers returns a copy of the header slice.
func (t *MutableTable) Headers() []string {
	h := make([]string, len(t.headers))
	copy(h, t.headers)
	return h
}

// Data returns a deep copy of the data rows.
func (t *MutableTable) Data() [][]string {
	d := make([][]string, len(t.data))
	for i, row := range t.data {
		r := make([]string, len(row))
		copy(r, row)
		d[i] = r
	}
	return d
}

// Alignments returns a copy of the alignments slice.
func (t *MutableTable) Alignments() []Alignment {
	a := make([]Alignment, len(t.alignments))
	copy(a, t.alignments)
	return a
}

// SortState records sort status for display.
type SortState struct {
	Col    int
	Asc    bool
	Active bool
}
