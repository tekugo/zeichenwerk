package widgets

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ── Constructor ───────────────────────────────────────────────────────────────

func TestCheckbox_Defaults(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	if cb.Flag(FlagChecked) {
		t.Error("FlagChecked should be false by default")
	}
	if !cb.Flag(FlagFocusable) {
		t.Error("FlagFocusable should be set")
	}
}

func TestCheckbox_InitiallyChecked(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", true)
	if !cb.Flag(FlagChecked) {
		t.Error("FlagChecked should be true when checked=true")
	}
}

// ── Toggle ────────────────────────────────────────────────────────────────────

func TestCheckbox_Toggle_FlipsState(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.Toggle()
	if !cb.Flag(FlagChecked) {
		t.Error("FlagChecked should be true after Toggle on unchecked")
	}
	cb.Toggle()
	if cb.Flag(FlagChecked) {
		t.Error("FlagChecked should be false after second Toggle")
	}
}

func TestCheckbox_Toggle_DispatchesEvtChange(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	var got bool
	cb.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		got = data[0].(bool)
		return true
	})
	cb.Toggle()
	if !got {
		t.Error("EvtChange should fire with true after toggling from unchecked")
	}
}

func TestCheckbox_Toggle_Readonly_Ignored(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.SetFlag(FlagReadonly, true)
	cb.Toggle()
	if cb.Flag(FlagChecked) {
		t.Error("Toggle on readonly checkbox should not change state")
	}
}

// ── Set ───────────────────────────────────────────────────────────────────────

func TestCheckbox_Set_True(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.Set(true)
	if !cb.Flag(FlagChecked) {
		t.Error("FlagChecked should be true after Set(true)")
	}
}

func TestCheckbox_Set_False(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", true)
	cb.Set(false)
	if cb.Flag(FlagChecked) {
		t.Error("FlagChecked should be false after Set(false)")
	}
}

func TestCheckbox_Set_DoesNotDispatchEvent(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	fired := false
	cb.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		fired = true
		return true
	})
	cb.Set(true)
	if fired {
		t.Error("Set() should not dispatch EvtChange (use Toggle for that)")
	}
}

// ── Keyboard ──────────────────────────────────────────────────────────────────

func TestCheckbox_Keyboard_Space_Toggles(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.handleKey(tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone))
	if !cb.Flag(FlagChecked) {
		t.Error("Space should toggle checkbox")
	}
}

func TestCheckbox_Keyboard_Enter_Toggles(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.handleKey(BuildKey(tcell.KeyEnter))
	if !cb.Flag(FlagChecked) {
		t.Error("Enter should toggle checkbox")
	}
}

func TestCheckbox_Keyboard_Readonly_Ignored(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.SetFlag(FlagReadonly, true)
	handled := cb.handleKey(tcell.NewEventKey(tcell.KeyRune, " ", tcell.ModNone))
	if handled {
		t.Error("Space on readonly checkbox should return false")
	}
	if cb.Flag(FlagChecked) {
		t.Error("FlagChecked should not change when readonly")
	}
}

// ── Mouse ─────────────────────────────────────────────────────────────────────

func TestCheckbox_Mouse_Click_Toggles(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.SetBounds(0, 0, 10, 1)
	cb.handleMouse(tcell.NewEventMouse(1, 0, tcell.Button1, tcell.ModNone))
	if !cb.Flag(FlagChecked) {
		t.Error("Button1 click should toggle checkbox")
	}
}

func TestCheckbox_Mouse_OutOfBounds_Ignored(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.SetBounds(0, 0, 10, 1)
	handled := cb.handleMouse(tcell.NewEventMouse(50, 50, tcell.Button1, tcell.ModNone))
	if handled {
		t.Error("click outside bounds should not be handled")
	}
	if cb.Flag(FlagChecked) {
		t.Error("state should not change on click outside bounds")
	}
}

func TestCheckbox_Mouse_Readonly_Ignored(t *testing.T) {
	cb := NewCheckbox("cb", "", "Enable", false)
	cb.SetBounds(0, 0, 10, 1)
	cb.SetFlag(FlagReadonly, true)
	handled := cb.handleMouse(tcell.NewEventMouse(1, 0, tcell.Button1, tcell.ModNone))
	if handled {
		t.Error("click on readonly checkbox should return false")
	}
}

// ── Render ────────────────────────────────────────────────────────────────────

func TestCheckbox_Render_ShowsUnchecked(t *testing.T) {
	cb := NewCheckbox("cb", "", "x", false)
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	cb.SetBounds(0, 0, 10, 1)
	cb.Render(r)
	got := cs.Get(0, 0) + cs.Get(1, 0) + cs.Get(2, 0)
	if got != "[ ]" {
		t.Errorf("rendered indicator = %q; want %q", got, "[ ]")
	}
}

func TestCheckbox_Render_ShowsChecked(t *testing.T) {
	cb := NewCheckbox("cb", "", "x", true)
	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())
	cb.SetBounds(0, 0, 10, 1)
	cb.Render(r)
	got := cs.Get(0, 0) + cs.Get(1, 0) + cs.Get(2, 0)
	if got != "[x]" {
		t.Errorf("rendered indicator = %q; want %q", got, "[x]")
	}
}
