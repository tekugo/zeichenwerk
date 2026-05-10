package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// ScannerForm is the WidgetForm for *Scanner. Width is clamped to >=
// 1 in Store; the Style options match the keys registered in
// scannerConfigs (blocks, diamonds, circles).
type ScannerForm struct {
	ComponentForm

	Width  int    `group:"general" label:"Width"`
	Glyphs string `group:"general" label:"Style" control:"select" options:"blocks,diamonds,circles"`
}

func (f *ScannerForm) Name() string  { return "Scanner" }
func (f *ScannerForm) Group() string { return "leaf" }
func (f *ScannerForm) Help() string  { return "Back-and-forth scanning animation with a fading trail" }

func (f *ScannerForm) Load(w core.Widget) {
	s := w.(*Scanner)
	f.ComponentForm.Load(&s.Component)
	f.Width = s.width
	for name, cfg := range scannerConfigs {
		if cfg.active.ch == s.config.active.ch {
			f.Glyphs = name
			break
		}
	}
	if f.Glyphs == "" {
		f.Glyphs = "blocks"
	}
}

func (f *ScannerForm) Store(w core.Widget) {
	s := w.(*Scanner)
	f.ComponentForm.Store(&s.Component)
	if f.Width >= 1 {
		s.width = f.Width
	}
	if cfg, ok := scannerConfigs[f.Glyphs]; ok {
		s.config = cfg
	}
}

func (f *ScannerForm) New() core.Widget {
	width := f.Width
	if width < 1 {
		width = 1
	}
	s := NewScanner("", "", width, f.Glyphs)
	f.Store(s)
	return s
}

func (f *ScannerForm) Validate(field string) error { return nil }

func (f *ScannerForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		width := f.Width
		if width < 1 {
			width = 1
		}
		style := f.Glyphs
		if style == "" {
			style = "blocks"
		}
		_, err := fmt.Fprintf(w, "Scanner(%q, %d, %q).\n", f.ID, width, style)
		return err
	})
}
