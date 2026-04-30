package widgets

import (
	"math"
	"strings"
	"testing"
)

func TestPreviewPanel_DefaultIsBlackOnWhite(t *testing.T) {
	pp := NewPreviewPanel("pp", "")
	if pp.Foreground() != (RGB{0, 0, 0}) {
		t.Errorf("default fg = %+v; want black", pp.Foreground())
	}
	if pp.Background() != (RGB{255, 255, 255}) {
		t.Errorf("default bg = %+v; want white", pp.Background())
	}
}

func TestPreviewPanel_Contrast_BlackOnWhite(t *testing.T) {
	pp := NewPreviewPanel("pp", "")
	if got := pp.Contrast(); math.Abs(got-21) > 0.01 {
		t.Errorf("contrast = %.4f; want ~21.0", got)
	}
}

func TestPreviewPanel_SetColors_UpdatesContrast(t *testing.T) {
	pp := NewPreviewPanel("pp", "")
	pp.SetColors(RGB{255, 128, 64}, RGB{16, 16, 32})
	r := pp.Contrast()
	if r < 1 || r > 21 {
		t.Errorf("contrast out of bounds: %.4f", r)
	}
	if !strings.HasPrefix(pp.contrastLabel.Text, "Contrast ") {
		t.Errorf("contrast label = %q; want prefix %q", pp.contrastLabel.Text, "Contrast ")
	}
}

func TestPreviewPanel_SetForegroundOnly(t *testing.T) {
	pp := NewPreviewPanel("pp", "")
	pp.SetForeground(RGB{10, 20, 30})
	if pp.Foreground() != (RGB{10, 20, 30}) {
		t.Errorf("Foreground = %+v; want {10 20 30}", pp.Foreground())
	}
	if pp.Background() != (RGB{255, 255, 255}) {
		t.Errorf("Background should not change; got %+v", pp.Background())
	}
}

func TestPreviewPanel_HintIsContentSize(t *testing.T) {
	pp := NewPreviewPanel("pp", "")
	w, h := pp.Hint()
	if w != 13 || h != 7 {
		t.Errorf("Hint = (%d,%d); want (13,7) content size", w, h)
	}
}

func TestPreviewPanel_ContrastLabelFormat(t *testing.T) {
	pp := NewPreviewPanel("pp", "")
	pp.SetColors(RGB{0, 0, 0}, RGB{0, 0, 0})
	got := pp.contrastLabel.Text
	if got != "Contrast  1.0" {
		t.Errorf("same-colour contrast label = %q; want %q", got, "Contrast  1.0")
	}
}
