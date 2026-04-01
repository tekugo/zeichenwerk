package zeichenwerk

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

// ---- helpers ---------------------------------------------------------------

func newEditor(lines ...string) *Editor {
	e := NewEditor("e", "")
	e.SetContent(lines)
	return e
}

// ---- selectionBounds -------------------------------------------------------

func TestEditor_SelectionBounds_NoSelection(t *testing.T) {
	e := newEditor("hello world")
	_, _, _, _, ok := e.selectionBounds()
	if ok {
		t.Error("selectionBounds should return ok=false when not selecting")
	}
}

func TestEditor_SelectionBounds_MarkBeforeCursor(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 2
	e.line = 0
	e.column = 7
	sl, sc, el, ec, ok := e.selectionBounds()
	if !ok {
		t.Fatal("selectionBounds should return ok=true")
	}
	if sl != 0 || sc != 2 || el != 0 || ec != 7 {
		t.Errorf("got (%d,%d)-(%d,%d); want (0,2)-(0,7)", sl, sc, el, ec)
	}
}

func TestEditor_SelectionBounds_MarkAfterCursor(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 7
	e.line = 0
	e.column = 2
	sl, sc, el, ec, ok := e.selectionBounds()
	if !ok {
		t.Fatal("selectionBounds should return ok=true")
	}
	if sl != 0 || sc != 2 || el != 0 || ec != 7 {
		t.Errorf("got (%d,%d)-(%d,%d); want (0,2)-(0,7)", sl, sc, el, ec)
	}
}

func TestEditor_SelectionBounds_MarkEqualsCursor(t *testing.T) {
	e := newEditor("hello")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 3
	e.line = 0
	e.column = 3
	_, _, _, _, ok := e.selectionBounds()
	if ok {
		t.Error("selectionBounds should return ok=false when mark equals cursor")
	}
}

func TestEditor_SelectionBounds_MultiLine(t *testing.T) {
	e := newEditor("abc", "def", "ghi")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 1
	e.line = 2
	e.column = 2
	sl, sc, el, ec, ok := e.selectionBounds()
	if !ok {
		t.Fatal("selectionBounds should return ok=true")
	}
	if sl != 0 || sc != 1 || el != 2 || ec != 2 {
		t.Errorf("got (%d,%d)-(%d,%d); want (0,1)-(2,2)", sl, sc, el, ec)
	}
}

// ---- HasSelection / ClearSelection -----------------------------------------

func TestEditor_HasSelection_FalseWhenEmpty(t *testing.T) {
	e := newEditor("hello")
	if e.HasSelection() {
		t.Error("HasSelection should be false initially")
	}
}

func TestEditor_ClearSelection(t *testing.T) {
	e := newEditor("hello")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 0
	e.line = 0
	e.column = 3
	if !e.HasSelection() {
		t.Fatal("expected selection to be active")
	}
	e.ClearSelection()
	if e.HasSelection() {
		t.Error("ClearSelection should clear the selection")
	}
}

// ---- SelectAll -------------------------------------------------------------

func TestEditor_SelectAll_SingleLine(t *testing.T) {
	e := newEditor("hello world")
	e.SelectAll()
	if !e.HasSelection() {
		t.Fatal("SelectAll should produce an active selection")
	}
	if e.markLine != 0 || e.markColumn != 0 {
		t.Errorf("mark = (%d,%d); want (0,0)", e.markLine, e.markColumn)
	}
	if e.line != 0 || e.column != 11 {
		t.Errorf("cursor = (%d,%d); want (0,11)", e.line, e.column)
	}
}

func TestEditor_SelectAll_MultiLine(t *testing.T) {
	e := newEditor("abc", "defg")
	e.SelectAll()
	if e.line != 1 || e.column != 4 {
		t.Errorf("cursor = (%d,%d); want (1,4)", e.line, e.column)
	}
}

// ---- ShiftLeft / plain Left collapses selection ----------------------------

func TestEditor_ShiftLeft_ExtendSelection(t *testing.T) {
	e := newEditor("hello")
	e.line = 0
	e.column = 3
	e.ShiftLeft()
	if !e.HasSelection() {
		t.Fatal("ShiftLeft should create a selection")
	}
	if e.markLine != 0 || e.markColumn != 3 {
		t.Errorf("mark = (%d,%d); want (0,3)", e.markLine, e.markColumn)
	}
	if e.column != 2 {
		t.Errorf("column = %d; want 2 after ShiftLeft", e.column)
	}
}

func TestEditor_ShiftLeft_KeepsMark(t *testing.T) {
	e := newEditor("hello")
	e.line = 0
	e.column = 4
	e.ShiftLeft() // sets mark at 4, cursor at 3
	e.ShiftLeft() // cursor moves to 2, mark stays at 4
	if e.markColumn != 4 {
		t.Errorf("mark column = %d; want 4 (unchanged)", e.markColumn)
	}
	if e.column != 2 {
		t.Errorf("column = %d; want 2", e.column)
	}
}

func TestEditor_PlainLeft_CollapsesToStart(t *testing.T) {
	e := newEditor("hello")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 1
	e.line = 0
	e.column = 4
	e.Left()
	if e.HasSelection() {
		t.Error("Left should clear selection")
	}
	if e.line != 0 || e.column != 1 {
		t.Errorf("cursor = (%d,%d); want (0,1) (selection start)", e.line, e.column)
	}
}

func TestEditor_PlainRight_CollapsesToEnd(t *testing.T) {
	e := newEditor("hello")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 1
	e.line = 0
	e.column = 4
	e.Right()
	if e.HasSelection() {
		t.Error("Right should clear selection")
	}
	if e.line != 0 || e.column != 4 {
		t.Errorf("cursor = (%d,%d); want (0,4) (selection end)", e.line, e.column)
	}
}

// ---- ShiftRight ------------------------------------------------------------

func TestEditor_ShiftRight_ExtendSelection(t *testing.T) {
	e := newEditor("hello")
	e.line = 0
	e.column = 2
	e.ShiftRight()
	if !e.HasSelection() {
		t.Fatal("ShiftRight should create a selection")
	}
	if e.markColumn != 2 {
		t.Errorf("mark column = %d; want 2", e.markColumn)
	}
	if e.column != 3 {
		t.Errorf("column = %d; want 3", e.column)
	}
}

// ---- ShiftUp / ShiftDown ---------------------------------------------------

func TestEditor_ShiftUp_ExtendSelection(t *testing.T) {
	e := newEditor("abc", "def")
	e.line = 1
	e.column = 2
	e.ShiftUp()
	if !e.HasSelection() {
		t.Fatal("ShiftUp should create a selection")
	}
	if e.line != 0 {
		t.Errorf("line = %d; want 0", e.line)
	}
}

func TestEditor_ShiftDown_ExtendSelection(t *testing.T) {
	e := newEditor("abc", "def")
	e.line = 0
	e.column = 1
	e.ShiftDown()
	if !e.HasSelection() {
		t.Fatal("ShiftDown should create a selection")
	}
	if e.line != 1 {
		t.Errorf("line = %d; want 1", e.line)
	}
}

// ---- ShiftHome / ShiftEnd --------------------------------------------------

func TestEditor_ShiftHome(t *testing.T) {
	e := newEditor("hello world")
	e.line = 0
	e.column = 5
	e.ShiftHome()
	if !e.HasSelection() {
		t.Fatal("ShiftHome should create a selection")
	}
	if e.column != 0 {
		t.Errorf("column = %d; want 0", e.column)
	}
	if e.markColumn != 5 {
		t.Errorf("mark column = %d; want 5", e.markColumn)
	}
}

func TestEditor_ShiftEnd(t *testing.T) {
	e := newEditor("hello")
	e.line = 0
	e.column = 2
	e.ShiftEnd()
	if !e.HasSelection() {
		t.Fatal("ShiftEnd should create a selection")
	}
	if e.column != 5 {
		t.Errorf("column = %d; want 5", e.column)
	}
	if e.markColumn != 2 {
		t.Errorf("mark column = %d; want 2", e.markColumn)
	}
}

// ---- PlainUp/Down/Home/End clear selection ---------------------------------

func TestEditor_Up_ClearsSelection(t *testing.T) {
	e := newEditor("abc", "def")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 0
	e.line = 1
	e.column = 2
	e.Up()
	if e.HasSelection() {
		t.Error("Up should clear selection")
	}
}

func TestEditor_Home_ClearsSelection(t *testing.T) {
	e := newEditor("hello")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 0
	e.line = 0
	e.column = 3
	e.Home()
	if e.HasSelection() {
		t.Error("Home should clear selection")
	}
}

// ---- SelectionText ---------------------------------------------------------

func TestEditor_SelectionText_SingleLine(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 6
	e.line = 0
	e.column = 11
	got := e.SelectionText()
	if got != "world" {
		t.Errorf("SelectionText = %q; want %q", got, "world")
	}
}

func TestEditor_SelectionText_MultiLine(t *testing.T) {
	e := newEditor("abc", "def", "ghi")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 1
	e.line = 2
	e.column = 2
	got := e.SelectionText()
	want := "bc\ndef\ngh"
	if got != want {
		t.Errorf("SelectionText = %q; want %q", got, want)
	}
}

func TestEditor_SelectionText_Empty(t *testing.T) {
	e := newEditor("hello")
	if e.SelectionText() != "" {
		t.Error("SelectionText should be empty when no selection")
	}
}

// ---- DeleteSelection -------------------------------------------------------

func TestEditor_DeleteSelection_SameLine(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 5
	e.line = 0
	e.column = 11
	e.DeleteSelection()
	if e.HasSelection() {
		t.Error("selection should be cleared after DeleteSelection")
	}
	if e.line != 0 || e.column != 5 {
		t.Errorf("cursor = (%d,%d); want (0,5)", e.line, e.column)
	}
	got := e.content[0].String()
	if got != "hello" {
		t.Errorf("line content = %q; want %q", got, "hello")
	}
}

func TestEditor_DeleteSelection_MultiLine(t *testing.T) {
	e := newEditor("hello", "world", "end")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 3 // "hel|lo"
	e.line = 2
	e.column = 1 // "e|nd"
	e.DeleteSelection()
	if len(e.content) != 1 {
		t.Errorf("content lines = %d; want 1", len(e.content))
	}
	got := e.content[0].String()
	if got != "helnd" {
		t.Errorf("line content = %q; want %q", got, "helnd")
	}
	if e.line != 0 || e.column != 3 {
		t.Errorf("cursor = (%d,%d); want (0,3)", e.line, e.column)
	}
}

func TestEditor_DeleteSelection_NoOp(t *testing.T) {
	e := newEditor("hello")
	e.DeleteSelection() // should not panic
	if e.content[0].String() != "hello" {
		t.Error("DeleteSelection without selection should not modify content")
	}
}

// ---- Insert replaces selection ---------------------------------------------

func TestEditor_Insert_ReplacesSelection(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 6
	e.line = 0
	e.column = 11
	e.Insert('X')
	if e.HasSelection() {
		t.Error("Insert should clear selection")
	}
	got := e.content[0].String()
	if got != "hello X" {
		t.Errorf("content = %q; want %q", got, "hello X")
	}
}

// ---- Delete / DeleteForward replace selection ------------------------------

func TestEditor_Delete_ReplacesSelection(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 0
	e.line = 0
	e.column = 6
	e.Delete()
	if e.HasSelection() {
		t.Error("Delete should clear selection")
	}
	got := e.content[0].String()
	if got != "world" {
		t.Errorf("content = %q; want %q", got, "world")
	}
}

func TestEditor_DeleteForward_ReplacesSelection(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 5
	e.line = 0
	e.column = 11
	e.DeleteForward()
	if e.HasSelection() {
		t.Error("DeleteForward should clear selection")
	}
	got := e.content[0].String()
	if got != "hello" {
		t.Errorf("content = %q; want %q", got, "hello")
	}
}

// ---- Copy / Paste round-trip -----------------------------------------------

func TestEditor_Copy_Paste_RoundTrip(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 6
	e.line = 0
	e.column = 11
	e.Copy()
	if editorClipboard != "world" {
		t.Errorf("clipboard = %q; want %q", editorClipboard, "world")
	}
	// Clear selection and move cursor, then paste
	e.ClearSelection()
	e.line = 0
	e.column = 0
	e.Paste()
	got := e.content[0].String()
	if got != "worldhello world" {
		t.Errorf("after paste content = %q; want %q", got, "worldhello world")
	}
}

func TestEditor_Cut(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 0
	e.line = 0
	e.column = 6
	e.Cut()
	if e.HasSelection() {
		t.Error("Cut should clear selection")
	}
	got := e.content[0].String()
	if got != "world" {
		t.Errorf("after cut content = %q; want %q", got, "world")
	}
	if editorClipboard != "hello " {
		t.Errorf("clipboard = %q; want %q", editorClipboard, "hello ")
	}
}

// ---- Paste multi-line ------------------------------------------------------

func TestEditor_Paste_MultiLine(t *testing.T) {
	editorClipboard = "foo\nbar"
	e := newEditor("hello")
	e.line = 0
	e.column = 5 // end of "hello"
	e.Paste()
	if len(e.content) != 2 {
		t.Fatalf("line count = %d; want 2", len(e.content))
	}
	if e.content[0].String() != "hellofoo" {
		t.Errorf("line 0 = %q; want %q", e.content[0].String(), "hellofoo")
	}
	if e.content[1].String() != "bar" {
		t.Errorf("line 1 = %q; want %q", e.content[1].String(), "bar")
	}
}

// ---- charToVisualCol -------------------------------------------------------

func TestEditor_CharToVisualCol_NoTabs(t *testing.T) {
	col := charToVisualCol("hello", 3, 4)
	if col != 3 {
		t.Errorf("charToVisualCol = %d; want 3", col)
	}
}

func TestEditor_CharToVisualCol_WithTab(t *testing.T) {
	// "\thello" with tabWidth=4: tab expands to 4 spaces, so char 1 = visual 4
	col := charToVisualCol("\thello", 1, 4)
	if col != 4 {
		t.Errorf("charToVisualCol = %d; want 4", col)
	}
}

func TestEditor_CharToVisualCol_MultiTab(t *testing.T) {
	// "ab\tc" with tabWidth=4: 'a'=0, 'b'=1, tab→4, 'c'=4 visual
	// charCol=2 (at the tab) → visual col 2 (a+b)
	col := charToVisualCol("ab\tc", 2, 4)
	if col != 2 {
		t.Errorf("charToVisualCol at tab = %d; want 2", col)
	}
	// charCol=3 (at 'c') → visual col 4 (after tab expansion)
	col = charToVisualCol("ab\tc", 3, 4)
	if col != 4 {
		t.Errorf("charToVisualCol after tab = %d; want 4", col)
	}
}

// ---- Key handler -----------------------------------------------------------

func TestEditor_KeyShiftLeft_ExtendSelection(t *testing.T) {
	e := newEditor("hello")
	e.line = 0
	e.column = 3
	ev := tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModShift)
	e.handleKey(ev)
	if !e.HasSelection() {
		t.Error("Shift+Left should create a selection")
	}
	if e.column != 2 {
		t.Errorf("column = %d; want 2", e.column)
	}
}

func TestEditor_KeyCtrlA_SelectAll(t *testing.T) {
	e := newEditor("hello", "world")
	ev := tcell.NewEventKey(tcell.KeyCtrlA, "", tcell.ModNone)
	e.handleKey(ev)
	if !e.HasSelection() {
		t.Error("Ctrl+A should select all")
	}
	if e.line != 1 {
		t.Errorf("cursor line = %d; want 1", e.line)
	}
}

func TestEditor_KeyCtrlC_Copy(t *testing.T) {
	e := newEditor("hello world")
	e.selecting = true
	e.markLine = 0
	e.markColumn = 0
	e.line = 0
	e.column = 5
	ev := tcell.NewEventKey(tcell.KeyCtrlC, "", tcell.ModNone)
	e.handleKey(ev)
	if editorClipboard != "hello" {
		t.Errorf("clipboard = %q; want %q", editorClipboard, "hello")
	}
}

func TestEditor_KeyCtrlHome_DocumentHome(t *testing.T) {
	e := newEditor("abc", "def")
	e.line = 1
	e.column = 3
	ev := tcell.NewEventKey(tcell.KeyHome, "", tcell.ModCtrl)
	e.handleKey(ev)
	if e.line != 0 || e.column != 0 {
		t.Errorf("cursor = (%d,%d); want (0,0)", e.line, e.column)
	}
}

func TestEditor_KeyCtrlEnd_DocumentEnd(t *testing.T) {
	e := newEditor("abc", "def")
	e.line = 0
	e.column = 0
	ev := tcell.NewEventKey(tcell.KeyEnd, "", tcell.ModCtrl)
	e.handleKey(ev)
	if e.line != 1 || e.column != 3 {
		t.Errorf("cursor = (%d,%d); want (1,3)", e.line, e.column)
	}
}
