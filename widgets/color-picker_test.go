package widgets

import (
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
)

func TestColorPicker_SingleMode_NoBackgroundOrPreview(t *testing.T) {
	cp := NewColorPicker("cp", "", ColorSingle)
	if cp.ForegroundPanel() == nil {
		t.Errorf("foreground panel must be present in single mode")
	}
	if cp.BackgroundPanel() != nil {
		t.Errorf("background panel must be nil in single mode; got %v", cp.BackgroundPanel())
	}
	if cp.Preview() != nil {
		t.Errorf("preview must be nil in single mode; got %v", cp.Preview())
	}
	if cp.Background() != "" {
		t.Errorf("Background() in single mode = %q; want empty string", cp.Background())
	}
	if cp.Contrast() != 1.0 {
		t.Errorf("Contrast() in single mode = %v; want 1.0", cp.Contrast())
	}
}

func TestColorPicker_FgBgMode_HasAllPanels(t *testing.T) {
	cp := NewColorPicker("cp", "", ColorFgBg)
	if cp.ForegroundPanel() == nil || cp.BackgroundPanel() == nil || cp.Preview() == nil {
		t.Errorf("all panels must be present in fg/bg mode")
	}
	if cp.Foreground() != "#000000" {
		t.Errorf("default fg = %q; want #000000", cp.Foreground())
	}
	if cp.Background() != "#ffffff" {
		t.Errorf("default bg = %q; want #ffffff", cp.Background())
	}
}

func TestColorPicker_HintMatchesMode(t *testing.T) {
	single := NewColorPicker("a", "", ColorSingle)
	if w, h := single.Hint(); w != 24 || h != 9 {
		t.Errorf("ColorSingle Hint = (%d,%d); want (24,9)", w, h)
	}
	fgbg := NewColorPicker("b", "", ColorFgBg)
	if w, h := fgbg.Hint(); w != 65 || h != 9 {
		t.Errorf("ColorFgBg Hint = (%d,%d); want (65,9)", w, h)
	}
}

func TestColorPicker_PanelChange_RemitsChange(t *testing.T) {
	cp := NewColorPicker("cp", "", ColorFgBg)

	count := 0
	var payload any
	cp.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		count++
		if len(data) > 0 {
			payload = data[0]
		}
		return false
	})

	cp.SetForeground("#ff8040")
	if count != 1 {
		t.Errorf("SetForeground fired %d EvtChange events; want 1", count)
	}
	if payload != cp {
		t.Errorf("EvtChange payload = %v; want the picker itself", payload)
	}

	cp.SetBackground("#101020")
	if count != 2 {
		t.Errorf("SetBackground brought total to %d; want 2", count)
	}
}

func TestColorPicker_PreviewSyncsWithPanels(t *testing.T) {
	cp := NewColorPicker("cp", "", ColorFgBg)
	cp.SetForeground("#ff0000")
	cp.SetBackground("#000000")

	if cp.Preview().Foreground() != (RGB{255, 0, 0}) {
		t.Errorf("preview fg = %+v; want red", cp.Preview().Foreground())
	}
	if cp.Preview().Background() != (RGB{0, 0, 0}) {
		t.Errorf("preview bg = %+v; want black", cp.Preview().Background())
	}
}

func TestColorPicker_BackgroundSetterIgnoredInSingleMode(t *testing.T) {
	cp := NewColorPicker("cp", "", ColorSingle)
	// Should not panic, should not emit, should leave Background empty.
	cp.SetBackground("#abcdef")
	if cp.Background() != "" {
		t.Errorf("Background() after SetBackground in single mode = %q; want empty", cp.Background())
	}
}
