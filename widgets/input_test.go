package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestInput_Defaults(t *testing.T) {
	inp := NewInput("i", "")
	if inp.Get() != "" {
		t.Errorf("Text() = %q; want empty", inp.Get())
	}
	if inp.Flag(FlagMasked) {
		t.Error("FlagMasked should be false by default")
	}
	if inp.Flag(FlagReadonly) {
		t.Error("FlagReadonly should be false by default")
	}
	if !inp.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
}

func TestInput_InitialText(t *testing.T) {
	inp := NewInput("i", "", "hello")
	if inp.Get() != "hello" {
		t.Errorf("Text() = %q; want %q", inp.Get(), "hello")
	}
}

func TestInput_InitialPlaceholder(t *testing.T) {
	inp := NewInput("i", "", "", "hint text")
	if inp.placeholder != "hint text" {
		t.Errorf("placeholder = %q; want %q", inp.placeholder, "hint text")
	}
}

// ── Set / Text ────────────────────────────────────────────────────────────────

func TestInput_Set_UpdatesText(t *testing.T) {
	inp := NewInput("i", "")
	inp.Set("world")
	if inp.Get() != "world" {
		t.Errorf("Text() = %q after Set; want %q", inp.Get(), "world")
	}
}

func TestInput_Set_Readonly_Ignored(t *testing.T) {
	inp := NewInput("i", "", "original")
	inp.SetFlag(FlagReadonly, true)
	inp.Set("changed")
	if inp.Get() != "original" {
		t.Errorf("Text() = %q after Set on readonly; want %q", inp.Get(), "original")
	}
}

func TestInput_Set_EmptyString(t *testing.T) {
	inp := NewInput("i", "", "hello")
	inp.Set("")
	if inp.Get() != "" {
		t.Errorf("Text() = %q after Set(\"\"); want empty", inp.Get())
	}
}

// ── Insert / Delete ───────────────────────────────────────────────────────────

func TestInput_Insert_AppendsAtEnd(t *testing.T) {
	inp := NewInput("i", "")
	inp.Insert("a")
	inp.Insert("b")
	inp.Insert("c")
	if inp.Get() != "abc" {
		t.Errorf("Text() = %q; want %q", inp.Get(), "abc")
	}
}

func TestInput_Insert_AtPosition(t *testing.T) {
	inp := NewInput("i", "", "ac")
	inp.Start()
	inp.Right() // cursor after 'a'
	inp.Insert("b")
	if inp.Get() != "abc" {
		t.Errorf("Text() = %q after insert at position 1; want %q", inp.Get(), "abc")
	}
}

func TestInput_Insert_Readonly_Ignored(t *testing.T) {
	inp := NewInput("i", "")
	inp.SetFlag(FlagReadonly, true)
	inp.Insert("x")
	if inp.Get() != "" {
		t.Errorf("Text() = %q after Insert on readonly; want empty", inp.Get())
	}
}

func TestInput_Insert_DispatchesEvtChange(t *testing.T) {
	inp := NewInput("i", "")
	fired := ""
	inp.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(string)
		return true
	})
	inp.Insert("x")
	if fired != "x" {
		t.Errorf("EvtChange data = %q; want %q", fired, "x")
	}
}

func TestInput_Delete_Backspace(t *testing.T) {
	inp := NewInput("i", "", "abc")
	inp.End()
	inp.Delete()
	if inp.Get() != "ab" {
		t.Errorf("Text() = %q after Delete at end; want %q", inp.Get(), "ab")
	}
}

func TestInput_Delete_AtStart_NoOp(t *testing.T) {
	inp := NewInput("i", "", "abc")
	inp.Start()
	inp.Delete() // no char before cursor
	if inp.Get() != "abc" {
		t.Errorf("Text() = %q after Delete at start; want unchanged %q", inp.Get(), "abc")
	}
}

func TestInput_Delete_DispatchesEvtChange(t *testing.T) {
	inp := NewInput("i", "", "a")
	inp.End()
	fired := ""
	inp.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		fired = data[0].(string)
		return true
	})
	inp.Delete()
	if fired != "" {
		t.Errorf("EvtChange data = %q; want empty string", fired)
	}
}

func TestInput_DeleteForward(t *testing.T) {
	inp := NewInput("i", "", "abc")
	inp.Start()
	inp.DeleteForward()
	if inp.Get() != "bc" {
		t.Errorf("Text() = %q after DeleteForward at start; want %q", inp.Get(), "bc")
	}
}

func TestInput_DeleteForward_AtEnd_NoOp(t *testing.T) {
	inp := NewInput("i", "", "abc")
	inp.End()
	inp.DeleteForward()
	if inp.Get() != "abc" {
		t.Errorf("Text() = %q after DeleteForward at end; want unchanged %q", inp.Get(), "abc")
	}
}

// ── Clear ─────────────────────────────────────────────────────────────────────

func TestInput_Clear(t *testing.T) {
	inp := NewInput("i", "", "hello world")
	inp.Clear()
	if inp.Get() != "" {
		t.Errorf("Text() = %q after Clear; want empty", inp.Get())
	}
}

func TestInput_Clear_DispatchesEvtChange(t *testing.T) {
	inp := NewInput("i", "", "hello")
	fired := false
	inp.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	inp.Clear()
	if !fired {
		t.Error("EvtChange should fire after Clear")
	}
}

func TestInput_Clear_Readonly_Ignored(t *testing.T) {
	inp := NewInput("i", "", "hello")
	inp.SetFlag(FlagReadonly, true)
	inp.Clear()
	if inp.Get() != "hello" {
		t.Errorf("Text() = %q after Clear on readonly; want %q", inp.Get(), "hello")
	}
}

// ── Cursor movement ───────────────────────────────────────────────────────────

func TestInput_Left_Right_MoveCursor(t *testing.T) {
	inp := NewInput("i", "", "abc")
	inp.SetBounds(0, 0, 20, 1)
	inp.End() // cursor at 3

	inp.Left()
	x, _, _ := inp.Cursor()
	if x != 2 {
		t.Errorf("Cursor x = %d after Left from end; want 2", x)
	}

	inp.Right()
	x, _, _ = inp.Cursor()
	if x != 3 {
		t.Errorf("Cursor x = %d after Right; want 3", x)
	}
}

func TestInput_Start_MovesToBeginning(t *testing.T) {
	inp := NewInput("i", "", "hello")
	inp.SetBounds(0, 0, 20, 1)
	inp.End()
	inp.Start()
	x, _, _ := inp.Cursor()
	if x != 0 {
		t.Errorf("Cursor x = %d after Start; want 0", x)
	}
}

func TestInput_End_MovesToEnd(t *testing.T) {
	inp := NewInput("i", "", "hello")
	inp.SetBounds(0, 0, 20, 1)
	inp.Start()
	inp.End()
	x, _, _ := inp.Cursor()
	if x != 5 {
		t.Errorf("Cursor x = %d after End; want 5 (len of \"hello\")", x)
	}
}

func TestInput_Left_AtStart_NoOp(t *testing.T) {
	inp := NewInput("i", "", "abc")
	inp.SetBounds(0, 0, 20, 1)
	inp.Start()
	inp.Left() // already at start
	x, _, _ := inp.Cursor()
	if x != 0 {
		t.Errorf("Cursor x = %d after Left at start; want 0", x)
	}
}

// ── SetMask ───────────────────────────────────────────────────────────────────

func TestInput_SetMask_EnablesMasked(t *testing.T) {
	inp := NewInput("i", "")
	inp.SetMask("*")
	if !inp.Flag(FlagMasked) {
		t.Error("FlagMasked should be set after SetMask")
	}
}

func TestInput_SetMask_Empty_DisablesMasked(t *testing.T) {
	inp := NewInput("i", "", "text", "", "*")
	inp.SetMask("")
	if inp.Flag(FlagMasked) {
		t.Error("FlagMasked should be cleared after SetMask(\"\")")
	}
}

func TestInput_Text_ReturnsPlaintext_EvenWhenMasked(t *testing.T) {
	inp := NewInput("i", "", "secret")
	inp.SetMask("*")
	if inp.Get() != "secret" {
		t.Errorf("Text() = %q with mask; want %q (unmasked)", inp.Get(), "secret")
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func TestInput_Keyboard_Backspace(t *testing.T) {
	inp := NewInput("i", "", "abc")
	inp.SetBounds(0, 0, 20, 1)
	inp.End()
	inp.handleKey(tcell.NewEventKey(tcell.KeyBackspace2, "", tcell.ModNone))
	if inp.Get() != "ab" {
		t.Errorf("Text() = %q after Backspace; want %q", inp.Get(), "ab")
	}
}

func TestInput_Keyboard_Delete(t *testing.T) {
	inp := NewInput("i", "", "abc")
	inp.SetBounds(0, 0, 20, 1)
	inp.Start()
	inp.handleKey(tcell.NewEventKey(tcell.KeyDelete, "", tcell.ModNone))
	if inp.Get() != "bc" {
		t.Errorf("Text() = %q after Delete key; want %q", inp.Get(), "bc")
	}
}

func TestInput_Keyboard_CtrlA_MoveToStart(t *testing.T) {
	inp := NewInput("i", "", "hello")
	inp.SetBounds(0, 0, 20, 1)
	inp.End()
	inp.handleKey(tcell.NewEventKey(tcell.KeyCtrlA, "", tcell.ModCtrl))
	x, _, _ := inp.Cursor()
	if x != 0 {
		t.Errorf("Cursor x = %d after Ctrl+A; want 0", x)
	}
}

func TestInput_Keyboard_CtrlE_MoveToEnd(t *testing.T) {
	inp := NewInput("i", "", "hello")
	inp.SetBounds(0, 0, 20, 1)
	inp.Start()
	inp.handleKey(tcell.NewEventKey(tcell.KeyCtrlE, "", tcell.ModCtrl))
	x, _, _ := inp.Cursor()
	if x != 5 {
		t.Errorf("Cursor x = %d after Ctrl+E; want 5", x)
	}
}

func TestInput_Keyboard_Rune_Inserts(t *testing.T) {
	inp := NewInput("i", "")
	inp.handleKey(tcell.NewEventKey(tcell.KeyRune, "x", tcell.ModNone))
	if inp.Get() != "x" {
		t.Errorf("Text() = %q after rune input; want %q", inp.Get(), "x")
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestInput_Render_ShowsText(t *testing.T) {
	inp := NewInput("i", "", "hi")
	cs2 := NewTestScreen()
	r2 := NewRenderer(cs2, NewTheme())
	inp.SetBounds(0, 0, 10, 1)
	inp.Render(r2)

	got := cs2.Get(0, 0) + cs2.Get(1, 0)
	if got != "hi" {
		t.Errorf("rendered cols 0-1 = %q; want %q", got, "hi")
	}
}

func TestInput_Render_ShowsPlaceholder_WhenEmpty(t *testing.T) {
	inp := NewInput("i", "", "", "type here")
	cs2 := NewTestScreen()
	r2 := NewRenderer(cs2, NewTheme())
	inp.SetBounds(0, 0, 20, 1)
	inp.Render(r2)

	// First char of placeholder should appear
	if cs2.Get(0, 0) != "t" {
		t.Errorf("col 0 = %q with placeholder; want %q", cs2.Get(0, 0), "t")
	}
}
