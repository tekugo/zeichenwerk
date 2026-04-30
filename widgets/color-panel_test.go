package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

func TestColorPanel_DefaultColor(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	if cp.Color() != "#000000" {
		t.Errorf("default Color() = %q; want #000000", cp.Color())
	}
	if cp.RGB() != (RGB{0, 0, 0}) {
		t.Errorf("default RGB() = %+v; want zero", cp.RGB())
	}
}

func TestColorPanel_SetColor_HexSixDigits(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.SetColor("#ff8040")
	if got := cp.RGB(); got != (RGB{255, 128, 64}) {
		t.Errorf("after SetColor(#ff8040), RGB = %+v; want {255 128 64}", got)
	}
	if cp.Color() != "#ff8040" {
		t.Errorf("Color() = %q; want #ff8040", cp.Color())
	}
}

func TestColorPanel_SetColor_HexThreeDigits(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.SetColor("#abc")
	if got := cp.RGB(); got != (RGB{0xaa, 0xbb, 0xcc}) {
		t.Errorf("after SetColor(#abc), RGB = %+v; want {170 187 204}", got)
	}
}

func TestColorPanel_SetColor_InvalidIsNoOp(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.SetRGB(RGB{10, 20, 30})
	cp.SetColor("not-a-color")
	if got := cp.RGB(); got != (RGB{10, 20, 30}) {
		t.Errorf("invalid SetColor should be no-op; RGB = %+v", got)
	}
}

func TestColorPanel_RefreshUpdatesAllInputs(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.SetRGB(RGB{255, 0, 0})
	if got := cp.inR.Get(); got != "255" {
		t.Errorf("inR = %q; want 255", got)
	}
	if got := cp.inG.Get(); got != "0" {
		t.Errorf("inG = %q; want 0", got)
	}
	if got := cp.inB.Get(); got != "0" {
		t.Errorf("inB = %q; want 0", got)
	}
	if got := cp.inH.Get(); got != "0" {
		t.Errorf("inH for red = %q; want 0", got)
	}
	if got := cp.inS.Get(); got != "100" {
		t.Errorf("inS for red = %q; want 100", got)
	}
	if got := cp.inL.Get(); got != "50" {
		t.Errorf("inL for red = %q; want 50", got)
	}
	if got := cp.inHex.Get(); got != "#ff0000" {
		t.Errorf("inHex = %q; want #ff0000", got)
	}
}

func TestColorPanel_ApplyRGB_FromInputs(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.inR.Set("255")
	cp.inG.Set("128")
	cp.inB.Set("64")
	cp.applyRGB()
	if got := cp.RGB(); got != (RGB{255, 128, 64}) {
		t.Errorf("applyRGB → RGB = %+v; want {255 128 64}", got)
	}
	if got := cp.inHex.Get(); got != "#ff8040" {
		t.Errorf("inHex after applyRGB = %q; want #ff8040", got)
	}
}

func TestColorPanel_ApplyHex_FromInput(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.inHex.Set("#ff8040")
	cp.applyHex()
	if got := cp.RGB(); got != (RGB{255, 128, 64}) {
		t.Errorf("applyHex → RGB = %+v; want {255 128 64}", got)
	}
}

func TestColorPanel_ApplyHSL_FromInputs(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.inH.Set("0")
	cp.inS.Set("100")
	cp.inL.Set("50")
	cp.applyHSL()
	if got := cp.RGB(); got != (RGB{255, 0, 0}) {
		t.Errorf("applyHSL(0,100,50) → RGB = %+v; want {255 0 0}", got)
	}
}

func TestColorPanel_DispatchesEvtChangeOnce(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")

	// Drain the EvtChange that initial setup may have queued by listening only
	// after construction completes.
	count := 0
	var payload any
	cp.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		count++
		if len(data) > 0 {
			payload = data[0]
		}
		return false
	})

	cp.SetColor("#112233")
	if count != 1 {
		t.Errorf("EvtChange fired %d times; want 1", count)
	}
	if payload != cp {
		t.Errorf("EvtChange payload = %v; want the panel itself", payload)
	}
}

func TestColorPanel_NoEventOnSameColor(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.SetRGB(RGB{10, 20, 30})

	count := 0
	cp.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		count++
		return false
	})
	cp.SetRGB(RGB{10, 20, 30})
	if count != 0 {
		t.Errorf("setting same colour fired EvtChange %d times; want 0", count)
	}
}

func TestColorPanel_InvalidRGBChannelDoesNotCrash(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.SetRGB(RGB{10, 20, 30})
	cp.inR.Set("not a number")
	cp.applyRGB() // should be a no-op for the colour
	if got := cp.RGB(); got != (RGB{10, 20, 30}) {
		t.Errorf("invalid R should leave colour untouched; got %+v", got)
	}
}

func TestColorPanel_HSLEditing_DoesNotLoop(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")

	count := 0
	cp.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		count++
		return false
	})

	// Simulate the user typing "100" into S; applyHSL should fire EvtChange
	// exactly once even though it programmatically rewrites R/G/B/Hex.
	cp.inH.Set("0")
	cp.inS.Set("100")
	cp.inL.Set("50")
	cp.applyHSL()

	if count != 1 {
		t.Errorf("applyHSL fired EvtChange %d times; want 1 (no feedback loop)", count)
	}
}

func TestColorPanel_ApplyHex_PartialDigitsIsNoOp(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.SetRGB(RGB{10, 20, 30})

	for _, partial := range []string{"#", "#f", "#ff", "#ffff", "#fffff", "#fffffff"} {
		cp.inHex.Set(partial)
		cp.applyHex()
		if got := cp.RGB(); got != (RGB{10, 20, 30}) {
			t.Errorf("partial Hex %q should not change colour; got %+v", partial, got)
		}
	}
}

func TestColorPanel_ApplyHex_ThreeDigitsTriggersUpdate(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.inHex.Set("#abc")
	cp.applyHex()
	if got := cp.RGB(); got != (RGB{0xaa, 0xbb, 0xcc}) {
		t.Errorf("#abc → RGB = %+v; want {170 187 204}", got)
	}
}

func TestColorPanel_ApplyHex_NoLeadingHash(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	cp.inHex.Set("ff8040")
	cp.applyHex()
	if got := cp.RGB(); got != (RGB{255, 128, 64}) {
		t.Errorf("hex without # → RGB = %+v; want {255 128 64}", got)
	}
}

func TestColorPanel_RefreshSkipsFocusedInput(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	// User has focused the Hex input and is typing "#fff" (a valid 3-digit
	// hex). The other fields should sync but the Hex field must NOT be
	// rewritten to its 6-digit canonical form.
	cp.inHex.SetFlag(FlagFocused, true)
	cp.inHex.Set("#fff")
	cp.applyHex()

	if got := cp.RGB(); got != (RGB{255, 255, 255}) {
		t.Errorf("after #fff → RGB = %+v; want {255 255 255}", got)
	}
	if got := cp.inHex.Get(); got != "#fff" {
		t.Errorf("focused Hex should remain %q; got %q", "#fff", got)
	}
	if got := cp.inR.Get(); got != "255" {
		t.Errorf("R should be updated to 255; got %q", got)
	}
}

func TestColorPanel_RefreshSkipsFocusedRInput(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	// Focus R and type a value. The R input must be left alone after
	// applyRGB so the cursor doesn't jump while typing.
	cp.inR.SetFlag(FlagFocused, true)
	cp.inR.Set("128")
	cp.applyRGB()

	if got := cp.RGB(); got != (RGB{128, 0, 0}) {
		t.Errorf("after R=128 → RGB = %+v; want {128 0 0}", got)
	}
	if got := cp.inR.Get(); got != "128" {
		t.Errorf("focused R should remain %q; got %q", "128", got)
	}
	if got := cp.inHex.Get(); got != "#800000" {
		t.Errorf("Hex should reflect new colour; got %q want #800000", got)
	}
}

func TestColorPanel_HintIsContentSize(t *testing.T) {
	cp := NewColorPanel("p", "", "Color")
	w, h := cp.Hint()
	if w != 22 || h != 7 {
		t.Errorf("Hint = (%d,%d); want (22,7) content size", w, h)
	}
}
