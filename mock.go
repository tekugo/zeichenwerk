package zeichenwerk

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

// NewMockScreen creates a tcell.Screen backed by a vt.MockTerm, suitable for
// headless testing. The returned screen has already been initialised.
func NewMockScreen(opts ...vt.MockOpt) (tcell.Screen, error) {
	mt := vt.NewMockTerm(opts...)
	scr, err := tcell.NewTerminfoScreenFromTty(mt)
	if err != nil {
		return nil, err
	}
	if err = scr.Init(); err != nil {
		return nil, err
	}
	return scr, nil
}
