package main

import (
	"fmt"
	"regexp"

	"github.com/gdamore/tcell/v3"
	zw "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/cmd/tblr/format"
)

// selMode tracks the current selection type.
type selMode int

const (
	selCell selMode = iota
	selRange
	selRow
	selCol
	selAll
)

// uiState holds mutable TUI state outside the widget tree.
type uiState struct {
	table        *format.MutableTable
	activeFormat format.Format
	filePath     string
	dir          string

	// selection
	curRow, curCol       int
	anchorRow, anchorCol int
	mode                 selMode

	// sort
	sortCol    int
	sortAsc    bool
	sortActive bool

	// search
	searchActive  bool
	searchPattern string
	searchMatches [][2]int
	searchIdx     int

	// watch
	watchActive bool
	watcher     *ClipboardWatcher
}

// buildUI constructs the tblr widget tree.
func buildUI(theme *zw.Theme, tbl *format.MutableTable, dir string, filePath string, activeFormat format.Format) *zw.UI {
	state := &uiState{
		table:        tbl,
		activeFormat: activeFormat,
		filePath:     filePath,
		dir:          dir,
		sortCol:      -1,
	}

	b := zw.NewBuilder(theme)

	b.Flex("root", false, "stretch", 0).
		Table("tbl", tbl, true).Hint(0, -1).
		Static("status", statusText(state)).Hint(0, 1).
		Shortcuts("cmdbar", "e", "edit", "a", "add", "d", "del", "s", "sort", "/", "find", "f", "format", "w", "watch", "q", "quit").Hint(0, 1)

	ui := b.Build()

	tblWidget := zw.Find(ui, "tbl").(*zw.Table)
	statusBar := zw.Find(ui, "status").(*zw.Static)
	cmdBar := zw.Find(ui, "cmdbar").(*zw.Shortcuts)

	// cell styler for selection highlight and search matches
	tblWidget.SetCellStyler(func(row, col int, highlight bool) *zw.Style {
		if state.searchActive && state.searchPattern != "" {
			for _, m := range state.searchMatches {
				if m[0] == row && m[1] == col {
					return tblWidget.Style("highlight:focused")
				}
			}
		}
		switch state.mode {
		case selAll:
			return tblWidget.Style("highlight:focused")
		case selRow:
			if row == state.curRow {
				return tblWidget.Style("highlight:focused")
			}
		case selCol:
			if col == state.curCol {
				return tblWidget.Style("highlight:focused")
			}
		case selRange:
			r0, r1 := minmax(state.anchorRow, state.curRow)
			c0, c1 := minmax(state.anchorCol, state.curCol)
			if row >= r0 && row <= r1 && col >= c0 && col <= c1 {
				return tblWidget.Style("highlight:focused")
			}
		}
		return nil
	})

	refreshStatus := func() {
		statusBar.SetText(statusText(state))
	}

	// track cursor via EvtSelect
	tblWidget.On(zw.EvtSelect, func(_ zw.Widget, _ zw.Event, data ...any) bool {
		if len(data) >= 2 {
			if r, ok := data[0].(int); ok {
				state.curRow = r
			}
			if c, ok := data[1].(int); ok {
				state.curCol = c
			}
		}
		refreshStatus()
		return true
	})

	confirmQuit := func() {
		if state.table.Modified() {
			ui.Confirm("Quit", "Unsaved changes. Quit anyway?", func() {
				ui.Quit()
			}, nil)
		} else {
			ui.Quit()
		}
	}

	openCellEditor := func() {
		row, col := tblWidget.Selected()
		if row < 0 || col < 0 {
			return
		}
		current := tbl.Str(row, col)
		x, y, w, ok := tblWidget.CellBounds(row, col)
		if !ok {
			ui.Prompt("Edit Cell", fmt.Sprintf("[%d,%d]:", row+1, col+1), func(val string) {
				tbl.SetCell(row, col, val)
				tbl.RecalcWidths()
				tblWidget.Set(tbl)
				tblWidget.Refresh()
				refreshStatus()
			}, nil)
			return
		}

		b := ui.NewBuilder()
		b.Flex("cell-edit-wrap", true, "start", 0).
			Input("cell-edit-input", current).Hint(w, 1)
		wrapper := b.Container()
		inp := zw.Find(wrapper, "cell-edit-input").(*zw.Input)
		inp.End()

		zw.OnKey(inp, func(e *tcell.EventKey) bool {
			switch e.Key() {
			case tcell.KeyEnter:
				val := inp.Text()
				ui.Close()
				tbl.SetCell(row, col, val)
				tbl.RecalcWidths()
				tblWidget.Set(tbl)
				tblWidget.Refresh()
				refreshStatus()
				return true
			case tcell.KeyEscape:
				ui.Close()
				return true
			}
			return false
		})

		ui.Popup(x, y, w, 1, wrapper)
	}

	zw.OnKey(tblWidget, func(e *tcell.EventKey) bool {
		row, col := tblWidget.Selected()
		if row < 0 {
			row = 0
		}
		if col < 0 {
			col = 0
		}
		state.curRow = row
		state.curCol = col

		// modifier-key combos first
		mods := e.Modifiers()
		ctrl := mods&tcell.ModCtrl != 0
		alt := mods&tcell.ModAlt != 0
		shift := mods&tcell.ModShift != 0

		switch e.Key() {
		case tcell.KeyEnter:
			openCellEditor()
			return true

		case tcell.KeyF2:
			openCellEditor()
			return true

		case tcell.KeyEscape:
			state.mode = selCell
			if state.searchActive {
				state.searchActive = false
				state.searchMatches = nil
				cmdBar.SetPairs("e", "edit", "a", "add", "d", "del", "s", "sort", "/", "find", "f", "format", "w", "watch", "q", "quit")
			}
			tblWidget.Refresh()
			return true

		case tcell.KeyUp:
			if alt {
				tbl.MoveRow(row, max2(0, row-1))
				tbl.RecalcWidths()
				tblWidget.Set(tbl)
				tblWidget.SetSelected(max2(0, row-1), col)
				tblWidget.Refresh()
				refreshStatus()
				return true
			}
			if shift {
				if state.mode != selRange {
					state.anchorRow = row
					state.anchorCol = col
					state.mode = selRange
				}
				// let table handle movement; styler will reflect selection
			}
			return false

		case tcell.KeyDown:
			if alt {
				tbl.MoveRow(row, min2(tbl.Length()-1, row+1))
				tbl.RecalcWidths()
				tblWidget.Set(tbl)
				tblWidget.SetSelected(min2(tbl.Length()-1, row+1), col)
				tblWidget.Refresh()
				refreshStatus()
				return true
			}
			if shift {
				if state.mode != selRange {
					state.anchorRow = row
					state.anchorCol = col
					state.mode = selRange
				}
			}
			return false

		case tcell.KeyLeft:
			if alt {
				tbl.MoveCol(col, max2(0, col-1))
				tbl.RecalcWidths()
				tblWidget.Set(tbl)
				tblWidget.Refresh()
				return true
			}
			if shift {
				if state.mode != selRange {
					state.anchorRow = row
					state.anchorCol = col
					state.mode = selRange
				}
			}
			return false

		case tcell.KeyRight:
			if alt {
				tbl.MoveCol(col, min2(tbl.ColCount()-1, col+1))
				tbl.RecalcWidths()
				tblWidget.Set(tbl)
				tblWidget.Refresh()
				return true
			}
			if shift {
				if state.mode != selRange {
					state.anchorRow = row
					state.anchorCol = col
					state.mode = selRange
				}
			}
			return false

		case tcell.KeyCtrlZ:
			tbl.Undo()
			tbl.RecalcWidths()
			tblWidget.Set(tbl)
			tblWidget.Refresh()
			refreshStatus()
			return true

		case tcell.KeyCtrlY:
			tbl.Redo()
			tbl.RecalcWidths()
			tblWidget.Set(tbl)
			tblWidget.Refresh()
			refreshStatus()
			return true

		case tcell.KeyCtrlS:
			doSave(ui, state, tblWidget, refreshStatus)
			return true

		case tcell.KeyCtrlC:
			if state.activeFormat != nil {
				_ = WriteToClipboard(tbl, state.activeFormat, true)
			}
			return true

		case tcell.KeyCtrlV:
			t2, f2, err := ReadFromClipboard()
			if err == nil && t2 != nil {
				tbl.Load(t2.Headers(), t2.Data())
				tbl.LoadAlignments(t2.Alignments())
				tbl.RecalcWidths()
				if f2 != nil {
					state.activeFormat = f2
				}
				tblWidget.Set(tbl)
				tblWidget.Refresh()
				refreshStatus()
			}
			return true

		case tcell.KeyInsert:
			if ctrl {
				tbl.InsertRowAt(row)
				tbl.RecalcWidths()
				tblWidget.Set(tbl)
				tblWidget.Refresh()
				return true
			}
			return false
		}

		if e.Key() != tcell.KeyRune {
			return false
		}

		ch := e.Str()

		// Ctrl+Space → select column
		if ch == " " && ctrl {
			state.mode = selCol
			tblWidget.Refresh()
			return true
		}
		// Shift+Space → select row
		if ch == " " && shift {
			state.mode = selRow
			tblWidget.Refresh()
			return true
		}
		// Ctrl+A → select all
		if (ch == "a" || ch == "A") && ctrl {
			state.mode = selAll
			tblWidget.Refresh()
			return true
		}

		switch ch {
		case "e":
			openCellEditor()
			return true
		case "a":
			tbl.AppendRow()
			tbl.RecalcWidths()
			tblWidget.Set(tbl)
			tblWidget.SetSelected(tbl.Length()-1, col)
			tblWidget.Refresh()
			refreshStatus()
			return true
		case "A":
			ui.Prompt("Add Column", "Column header:", func(h string) {
				tbl.AppendCol(h)
				tbl.RecalcWidths()
				tblWidget.Set(tbl)
				tblWidget.Refresh()
				refreshStatus()
			}, nil)
			return true
		case "d":
			if state.mode == selCol {
				tbl.DeleteCol(col)
			} else if row >= 0 && row < tbl.Length() {
				tbl.DeleteRow(row)
			}
			tbl.RecalcWidths()
			tblWidget.Set(tbl)
			tblWidget.Refresh()
			refreshStatus()
			return true
		case "s":
			tbl.SortByCol(col, true)
			state.sortCol = col
			state.sortAsc = true
			state.sortActive = true
			tblWidget.Refresh()
			refreshStatus()
			return true
		case "S":
			tbl.SortByCol(col, false)
			state.sortCol = col
			state.sortAsc = false
			state.sortActive = true
			tblWidget.Refresh()
			refreshStatus()
			return true
		case "<":
			tbl.SetColAlignment(col, format.AlignLeft)
			tbl.RecalcWidths()
			tblWidget.Set(tbl)
			tblWidget.Refresh()
			refreshStatus()
			return true
		case ">":
			tbl.SetColAlignment(col, format.AlignRight)
			tbl.RecalcWidths()
			tblWidget.Set(tbl)
			tblWidget.Refresh()
			refreshStatus()
			return true
		case "^":
			tbl.SetColAlignment(col, format.AlignCenter)
			tbl.RecalcWidths()
			tblWidget.Set(tbl)
			tblWidget.Refresh()
			refreshStatus()
			return true
		case "/":
			openSearch(ui, state, tblWidget, cmdBar, statusBar, refreshStatus)
			return true
		case "n":
			if state.searchActive && len(state.searchMatches) > 0 {
				state.searchIdx = (state.searchIdx + 1) % len(state.searchMatches)
				m := state.searchMatches[state.searchIdx]
				tblWidget.SetSelected(m[0], m[1])
				tblWidget.Refresh()
			}
			return true
		case "N":
			if state.searchActive && len(state.searchMatches) > 0 {
				n := len(state.searchMatches)
				state.searchIdx = (state.searchIdx - 1 + n) % n
				m := state.searchMatches[state.searchIdx]
				tblWidget.SetSelected(m[0], m[1])
				tblWidget.Refresh()
			}
			return true
		case "f":
			openFormatPicker(ui, state, tblWidget, refreshStatus)
			return true
		case "w":
			toggleWatch(ui, state, tbl, tblWidget, cmdBar, refreshStatus)
			return true
		case "q":
			confirmQuit()
			return true
		}

		return false
	})

	return ui
}

// statusText builds the status bar string.
func statusText(s *uiState) string {
	fmtName := "csv"
	if s.activeFormat != nil {
		fmtName = s.activeFormat.Name()
	}
	nrows := s.table.Length()
	ncols := s.table.ColCount()
	cursor := fmt.Sprintf("[%d,%d]", s.curRow+1, s.curCol+1)

	alignSym := "←"
	alignName := "left"
	if ncols > 0 && s.curCol < ncols {
		switch s.table.ColAlignment(s.curCol) {
		case format.AlignCenter:
			alignSym = "↔"
			alignName = "center"
		case format.AlignRight:
			alignSym = "→"
			alignName = "right"
		}
	}

	result := fmtName + " · " +
		fmt.Sprintf("%d rows", nrows) + " · " +
		fmt.Sprintf("%d cols", ncols) + " · " +
		cursor + " · " +
		alignSym + " " + alignName

	if s.sortActive && s.sortCol >= 0 && s.sortCol < ncols {
		dir := "↑"
		if !s.sortAsc {
			dir = "↓"
		}
		result += " · sorted: " + s.table.Header(s.sortCol) + " " + dir
	}

	if s.watchActive {
		result += " ●"
	}

	return result
}

// doSave serialises and writes the table to disk.
func doSave(ui *zw.UI, state *uiState, tblWidget *zw.Table, refresh func()) {
	if state.filePath == "" {
		ui.Prompt("Save As", "File path:", func(path string) {
			if path == "" {
				return
			}
			state.filePath = path
			doSaveToPath(ui, state, refresh)
		}, nil)
		return
	}
	doSaveToPath(ui, state, refresh)
}

func doSaveToPath(ui *zw.UI, state *uiState, refresh func()) {
	if state.activeFormat == nil {
		state.activeFormat = format.ByName("csv")
	}
	data, err := state.activeFormat.Serialize(state.table, format.SerialOpts{Pretty: true})
	if err != nil {
		ui.Confirm("Save Error", err.Error(), nil, nil)
		return
	}
	if err := writeFile(state.filePath, data); err != nil {
		ui.Confirm("Save Error", err.Error(), nil, nil)
		return
	}
	state.table.ClearModified()
	refresh()
}

// openSearch shows a search prompt and highlights matches.
func openSearch(ui *zw.UI, state *uiState, tblWidget *zw.Table, cmdBar *zw.Shortcuts, statusBar *zw.Static, refresh func()) {
	state.searchActive = true
	cmdBar.SetPairs("Enter", "next", "Esc", "cancel")

	ui.Prompt("Search", "Pattern (regex):", func(pattern string) {
		state.searchPattern = pattern
		state.searchMatches = nil
		state.searchIdx = 0
		cmdBar.SetPairs("e", "edit", "a", "add", "d", "del", "s", "sort", "/", "find", "f", "format", "w", "watch", "q", "quit")

		if pattern == "" {
			state.searchActive = false
			tblWidget.Refresh()
			return
		}

		re, err := regexp.Compile(pattern)
		if err != nil {
			state.searchActive = false
			tblWidget.Refresh()
			return
		}

		for row := 0; row < state.table.Length(); row++ {
			for col := 0; col < state.table.ColCount(); col++ {
				if re.MatchString(state.table.Str(row, col)) {
					state.searchMatches = append(state.searchMatches, [2]int{row, col})
				}
			}
		}

		if len(state.searchMatches) > 0 {
			m := state.searchMatches[0]
			tblWidget.SetSelected(m[0], m[1])
		}
		tblWidget.Refresh()
		refresh()
	}, func() {
		state.searchActive = false
		state.searchMatches = nil
		cmdBar.SetPairs("e", "edit", "a", "add", "d", "del", "s", "sort", "/", "find", "f", "format", "w", "watch", "q", "quit")
		tblWidget.Refresh()
	})
}

// openFormatPicker prompts for a format name.
func openFormatPicker(ui *zw.UI, state *uiState, tblWidget *zw.Table, refresh func()) {
	names := ""
	for i, f := range format.All() {
		if i > 0 {
			names += ", "
		}
		names += f.Name()
	}
	ui.Prompt("Format", "Format ("+names+"):", func(name string) {
		if f := format.ByName(name); f != nil {
			state.activeFormat = f
		}
		refresh()
	}, nil)
}

// toggleWatch starts or stops the clipboard watcher.
func toggleWatch(ui *zw.UI, state *uiState, tbl *format.MutableTable, tblWidget *zw.Table, cmdBar *zw.Shortcuts, refresh func()) {
	if state.watchActive {
		if state.watcher != nil {
			state.watcher.Stop()
			state.watcher = nil
		}
		state.watchActive = false
		cmdBar.SetPairs("e", "edit", "a", "add", "d", "del", "s", "sort", "/", "find", "f", "format", "w", "watch", "q", "quit")
		refresh()
		return
	}

	state.watchActive = true
	cmdBar.SetPairs("e", "edit", "a", "add", "d", "del", "s", "sort", "/", "find", "f", "format", "w", "stop", "q", "quit")

	state.watcher = NewClipboardWatcher(func(t2 *format.MutableTable, f format.Format) {
		tbl.Load(t2.Headers(), t2.Data())
		tbl.LoadAlignments(t2.Alignments())
		tbl.RecalcWidths()
		if f != nil {
			state.activeFormat = f
		}
		tblWidget.Set(tbl)
		tblWidget.Refresh()
		refresh()
	})
	state.watcher.Start()
}

func minmax(a, b int) (int, int) {
	if a <= b {
		return a, b
	}
	return b, a
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}
