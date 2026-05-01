package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// Ensure Component implements Widget interface
var _ Widget = (*Component)(nil)

func TestComponent_Bounds(t *testing.T) {
	c := &Component{
		x:      10,
		y:      20,
		width:  100,
		height: 50,
	}
	x, y, w, h := c.Bounds()
	if x != 10 || y != 20 || w != 100 || h != 50 {
		t.Errorf("Bounds() = %d, %d, %d, %d; want 10, 20, 100, 50", x, y, w, h)
	}

	c.SetBounds(5, 5, 20, 20)
	x, y, w, h = c.Bounds()
	if x != 5 || y != 5 || w != 20 || h != 20 {
		t.Errorf("SetBounds() failed. Bounds() = %d, %d, %d, %d; want 5, 5, 20, 20", x, y, w, h)
	}
}

func TestComponent_Hint(t *testing.T) {
	c := &Component{
		hwidth:  200,
		hheight: 100,
	}
	w, h := c.Hint()
	if w != 200 || h != 100 {
		t.Errorf("Hint() = %d, %d; want 200, 100", w, h)
	}

	c.SetHint(300, 150)
	w, h = c.Hint()
	if w != 300 || h != 150 {
		t.Errorf("Hint() after SetHint = %d, %d; want 300, 150", w, h)
	}
}

func TestComponent_ID(t *testing.T) {
	c := &Component{id: "test-widget"}
	if c.ID() != "test-widget" {
		t.Errorf("ID() = %s; want test-widget", c.ID())
	}
}

func TestComponent_Flag(t *testing.T) {
	c := &Component{}

	if c.Flag(FlagHidden) {
		t.Errorf("Flag('hidden') should be false initially")
	}

	c.SetFlag(FlagHidden, true)
	if !c.Flag(FlagHidden) {
		t.Errorf("Flag('hidden') should be true after SetFlag")
	}

	c.SetFlag(FlagHidden, false)
	if c.Flag(FlagHidden) {
		t.Errorf("Flag('hidden') should be false after turning off")
	}
}

func TestComponent_Dispatch_On(t *testing.T) {
	c := &Component{}
	called := false
	var capturedEvent Event

	handler := func(_ Widget, event Event, data ...any) bool {
		called = true
		capturedEvent = event
		return true
	}

	c.On(EvtActivate, handler)
	c.Dispatch(c, EvtActivate)

	if !called {
		t.Error("Dispatch('activate') did not trigger the handler")
	}
	if capturedEvent != "activate" {
		t.Errorf("Handler captured event %s; want click", capturedEvent)
	}
}

func TestComponent_Content(t *testing.T) {
	c := &Component{x: 0, y: 0, width: 100, height: 100}
	cx, cy, cw, ch := c.Content()
	if cx != 0 || cy != 0 || cw != 100 || ch != 100 {
		t.Errorf("Content() no style = %d,%d %dx%d; want 0,0 100x100", cx, cy, cw, ch)
	}
}

// ── Parent / child wiring ─────────────────────────────────────────────────────

func TestComponent_SetParent_Parent(t *testing.T) {
	parent := NewSwitcher("parent", "") // Switcher implements Container
	child := NewComponent("child", "")
	child.SetParent(parent)
	if child.Parent() != parent {
		t.Error("Parent() should return the set parent")
	}
}

func TestComponent_SetParent_Nil_ClearsParent(t *testing.T) {
	parent := NewSwitcher("p", "")
	child := NewComponent("c", "")
	child.SetParent(parent)
	child.SetParent(nil)
	if child.Parent() != nil {
		t.Error("Parent() should be nil after SetParent(nil)")
	}
}

// ── State priority ────────────────────────────────────────────────────────────

func TestComponent_State_Disabled_Wins(t *testing.T) {
	c := NewComponent("c", "")
	c.SetFlag(FlagDisabled, true)
	c.SetFlag(FlagFocused, true)
	c.SetFlag(FlagHovered, true)
	if c.State() != string(FlagDisabled) {
		t.Errorf("State() = %q; want %q (disabled beats all)", c.State(), string(FlagDisabled))
	}
}

func TestComponent_State_Pressed_BeforesFocused(t *testing.T) {
	c := NewComponent("c", "")
	c.SetFlag(FlagPressed, true)
	c.SetFlag(FlagFocused, true)
	if c.State() != string(FlagPressed) {
		t.Errorf("State() = %q; want %q", c.State(), string(FlagPressed))
	}
}

func TestComponent_State_Empty_WhenNoFlags(t *testing.T) {
	c := NewComponent("c", "")
	if c.State() != "" {
		t.Errorf("State() = %q; want empty when no flags set", c.State())
	}
}

// ── Style fallback chain ──────────────────────────────────────────────────────

func TestComponent_Style_ExactMatch(t *testing.T) {
	c := NewComponent("c", "")
	s := NewStyle("").WithColors("red", "blue")
	c.SetStyle("header:focused", s)
	got := c.Style("header:focused")
	if got != s {
		t.Error("Style() should return exact match for registered selector")
	}
}

func TestComponent_Style_FallsBackToPart(t *testing.T) {
	c := NewComponent("c", "")
	s := NewStyle("").WithColors("red", "blue")
	c.SetStyle("header", s)
	got := c.Style("header:focused")
	if got != s {
		t.Error("Style(\"header:focused\") should fall back to \"header\" when exact not found")
	}
}

func TestComponent_Style_FallsBackToState(t *testing.T) {
	c := NewComponent("c", "")
	s := NewStyle("").WithColors("red", "blue")
	c.SetStyle(":focused", s)
	got := c.Style("header:focused")
	if got != s {
		t.Error("Style(\"header:focused\") should fall back to \":focused\" when part not found")
	}
}

func TestComponent_Style_FallsBackToDefault(t *testing.T) {
	c := NewComponent("c", "")
	got := c.Style("unknown:state")
	if got == nil {
		t.Error("Style() should never return nil — falls back to default")
	}
}

// ── SetStyle / Styles ─────────────────────────────────────────────────────────

func TestComponent_SetStyle_Styles(t *testing.T) {
	c := NewComponent("c", "")
	c.SetStyle("", NewStyle("").WithColors("white", "black"))
	c.SetStyle(":focused", NewStyle("").WithColors("black", "white"))
	styles := c.Styles()
	if len(styles) != 2 {
		t.Errorf("Styles() len = %d; want 2", len(styles))
	}
}

func TestComponent_SetStyle_Nil_Removes(t *testing.T) {
	c := NewComponent("c", "")
	c.SetStyle("test", NewStyle(""))
	c.SetStyle("test", nil)
	styles := c.Styles()
	for _, s := range styles {
		if s == "test" {
			t.Error("nil SetStyle should remove the selector")
		}
	}
}

// ── Selector ─────────────────────────────────────────────────────────────────

func TestComponent_Selector_TypeOnly(t *testing.T) {
	c := NewComponent("", "")
	if c.Selector("button") != "button" {
		t.Errorf("Selector() = %q; want %q", c.Selector("button"), "button")
	}
}

func TestComponent_Selector_WithClass(t *testing.T) {
	c := NewComponent("", "primary")
	if c.Selector("button") != "button.primary" {
		t.Errorf("Selector() = %q; want %q", c.Selector("button"), "button.primary")
	}
}

func TestComponent_Selector_WithID(t *testing.T) {
	c := NewComponent("submit", "")
	if c.Selector("button") != "button#submit" {
		t.Errorf("Selector() = %q; want %q", c.Selector("button"), "button#submit")
	}
}

func TestComponent_Selector_WithClassAndID(t *testing.T) {
	c := NewComponent("submit", "primary")
	got := c.Selector("button")
	want := "button.primary#submit"
	if got != want {
		t.Errorf("Selector() = %q; want %q", got, want)
	}
}

// ── FlagHidden prevents Render ────────────────────────────────────────────────

func TestComponent_FlagHidden_SkipsBackgroundFill(t *testing.T) {
	c := NewComponent("c", "")
	c.SetBounds(0, 0, 5, 1)
	c.SetStyle("", NewStyle("").WithColors("white", "black")) // background set → would Fill with " "

	cs := NewTestScreen()
	r := NewRenderer(cs, NewTheme())

	// Without hidden: background fill should write spaces
	c.Render(r)
	if cs.Get(0, 0) != " " {
		t.Errorf("visible component with background should fill with space; got %q", cs.Get(0, 0))
	}

	// With hidden: nothing written
	cs2 := NewTestScreen()
	r2 := NewRenderer(cs2, NewTheme())
	c.SetFlag(FlagHidden, true)
	c.Render(r2)
	if cs2.Get(0, 0) != "" {
		t.Errorf("hidden component should not fill background; got %q at (0,0)", cs2.Get(0, 0))
	}
}
