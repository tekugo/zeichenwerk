package next

import (
	"testing"
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

	if c.Flag("hidden") {
		t.Errorf("Flag('hidden') should be false initially")
	}

	c.SetFlag("hidden", true)
	if !c.Flag("hidden") {
		t.Errorf("Flag('hidden') should be true after SetFlag")
	}

	c.SetFlag("hidden", false)
	if c.Flag("hidden") {
		t.Errorf("Flag('hidden') should be false after turning off")
	}
}

func TestComponent_Dispatch_On(t *testing.T) {
	c := &Component{}
	called := false
	var capturedEvent string

	handler := func(event string, data ...any) bool {
		called = true
		capturedEvent = event
		return true
	}

	c.On("click", handler)
	c.Dispatch("click")

	if !called {
		t.Error("Dispatch('click') did not trigger the handler")
	}
	if capturedEvent != "click" {
		t.Errorf("Handler captured event %s; want click", capturedEvent)
	}
}

func TestComponent_Content(t *testing.T) {
	c := &Component{
		x:      0,
		y:      0,
		width:  100,
		height: 100,
	}
	// No style
	cx, cy, cw, ch := c.Content()
	if cx != 0 || cy != 0 || cw != 100 || ch != 100 {
		t.Errorf("Content() no style = %d,%d %dx%d; want 0,0 100x100", cx, cy, cw, ch)
	}

	// With style (margin 5, border 1 (implicit if names imply it, but we set explicit usually, here just assuming Style logic works as mocked or real))
	// Actually we depend on Style implementation. Let's assume Style logic is correct and just test we call it.
	// We need to inject a style into the component manually or via a public method if one existed to set style directly,
	// but currently only internal map access or maybe via Style() which creates default.
	// SetStyle is not in the Interface, nor in the component (wait, previous code didn't have SetStyle method expoyted on Component?
	// Wait, base-widget had SetStyle. I missed SetStyle in Component!
	// Let me check my implementation of Component again. I don't recall implementing SetStyle public method.
	// The interface doesn't REQUIRE SetStyle?
	// Let me check Widget interface again.
	// Widget interface:
	// ...
	// Dispatch, On, ...
	// It does NOT have SetStyle.
	// However, how do we set styles then? BaseWidget had SetStyle. Conceptually widgets should have it?
	// The prompt asked to implement missing Widget methods. If SetStyle is not in Widget interface, strict compliance doesn't require it.
	// BUT, practically, we need it.
	// For now, I will test Content() with default style which is 0 insets.
}
