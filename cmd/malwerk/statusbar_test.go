package main

import (
	"testing"

	"github.com/tekugo/zeichenwerk/themes"
)

func TestStatusBar_ApplyInstallsLocalStyle(t *testing.T) {
	sb := NewStatusBar("statusbar")
	theme := themes.TokyoNight()
	sb.Apply(theme)

	style := sb.Style()
	if got := style.Foreground(); got != "$fg0" {
		t.Errorf("fg = %q; want %q", got, "$fg0")
	}
	if got := style.Background(); got != "$bg2" {
		t.Errorf("bg = %q; want %q", got, "$bg2")
	}
}

func TestStatusBar_HintHonoursVisibility(t *testing.T) {
	sb := NewStatusBar("statusbar")
	if w, h := sb.Hint(); w != 0 || h != 1 {
		t.Errorf("visible Hint = (%d, %d); want (0, 1)", w, h)
	}
	sb.visible = false
	if w, h := sb.Hint(); w != 0 || h != 0 {
		t.Errorf("hidden Hint = (%d, %d); want (0, 0)", w, h)
	}
}
