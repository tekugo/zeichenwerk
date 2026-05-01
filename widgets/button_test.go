package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestButton_Defaults(t *testing.T) {
	b := NewButton("b", "", "OK")
	if !b.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
}

// ── Set / Summary ─────────────────────────────────────────────────────────────

func TestButton_Set_UpdatesLabel(t *testing.T) {
	b := NewButton("b", "", "Old")
	b.Set("New")
	if b.Summary() != "New" {
		t.Errorf("Summary() = %q after Set; want %q", b.Summary(), "New")
	}
}

// ── Hint ──────────────────────────────────────────────────────────────────────

func TestButton_Hint_WidthEqualsRuneCount(t *testing.T) {
	b := NewButton("b", "", "Save")
	w, h := b.Hint()
	if w != 4 {
		t.Errorf("Hint width = %d; want 4 (rune count of \"Save\")", w)
	}
	if h != 1 {
		t.Errorf("Hint height = %d; want 1", h)
	}
}

func TestButton_Hint_MultiByte(t *testing.T) {
	b := NewButton("b", "", "über") // 4 runes
	w, _ := b.Hint()
	if w != 4 {
		t.Errorf("Hint width = %d; want 4 (rune count)", w)
	}
}

// ── Activate ──────────────────────────────────────────────────────────────────

func TestButton_Activate_DispatchesEvtActivate(t *testing.T) {
	b := NewButton("b", "", "OK")
	fired := false
	b.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	b.Activate()
	if !fired {
		t.Error("EvtActivate should fire after Activate()")
	}
}

func TestButton_Activate_PayloadIsZero(t *testing.T) {
	b := NewButton("b", "", "OK")
	var got any
	b.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		got = data[0]
		return true
	})
	b.Activate()
	if got != 0 {
		t.Errorf("EvtActivate payload = %v; want 0", got)
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func TestButton_Keyboard_Enter_Activates(t *testing.T) {
	b := NewButton("b", "", "OK")
	fired := false
	b.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	b.handleKey(BuildKey(tcell.KeyEnter))
	if !fired {
		t.Error("EvtActivate should fire on Enter")
	}
}

func TestButton_Keyboard_Space_Activates(t *testing.T) {
	b := NewButton("b", "", "OK")
	fired := false
	b.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	b.handleKey(tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone))
	if !fired {
		t.Error("EvtActivate should fire on Space")
	}
}

func TestButton_Keyboard_OtherKey_NotHandled(t *testing.T) {
	b := NewButton("b", "", "OK")
	handled := b.handleKey(BuildKey(tcell.KeyLeft))
	if handled {
		t.Error("Left key should not be handled by button")
	}
}

// ── Mouse ─────────────────────────────────────────────────────────────────────

func TestButton_Mouse_PressSetsFlagPressed(t *testing.T) {
	b := NewButton("b", "", "OK")
	b.SetBounds(0, 0, 4, 1)
	b.handleMouse(tcell.NewEventMouse(1, 0, tcell.Button1, tcell.ModNone))
	if !b.Flag(FlagPressed) {
		t.Error("FlagPressed should be set after Button1 press")
	}
}

func TestButton_Mouse_ReleaseAfterPress_Activates(t *testing.T) {
	b := NewButton("b", "", "OK")
	b.SetBounds(0, 0, 4, 1)
	fired := false
	b.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	b.handleMouse(tcell.NewEventMouse(1, 0, tcell.Button1, tcell.ModNone))    // press
	b.handleMouse(tcell.NewEventMouse(1, 0, tcell.ButtonNone, tcell.ModNone)) // release
	if !fired {
		t.Error("EvtActivate should fire on mouse release within bounds")
	}
}

func TestButton_Mouse_PressOutsideBounds_Ignored(t *testing.T) {
	b := NewButton("b", "", "OK")
	b.SetBounds(0, 0, 4, 1)
	handled := b.handleMouse(tcell.NewEventMouse(10, 10, tcell.Button1, tcell.ModNone))
	if handled {
		t.Error("click outside bounds should not be handled")
	}
}

func TestButton_Mouse_Disabled_StillActivates(t *testing.T) {
	// Button does not check FlagDisabled in handleMouse — documents current behaviour
	b := NewButton("b", "", "OK")
	b.SetBounds(0, 0, 4, 1)
	b.SetFlag(FlagDisabled, true)
	fired := false
	b.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	b.handleMouse(tcell.NewEventMouse(1, 0, tcell.Button1, tcell.ModNone))
	b.handleMouse(tcell.NewEventMouse(1, 0, tcell.ButtonNone, tcell.ModNone))
	if !fired {
		t.Error("button currently fires even when disabled (documents implementation)")
	}
}
