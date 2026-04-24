package renderer

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

// NewMockScreen constructs a tcell.Screen wired to a vt.MockTerm so
// rendering tests can run without a real terminal. The screen is
// returned already initialised (Init has been called); callers only
// need to wrap it in a TcellScreen and exercise their widgets.
//
// The opts are forwarded to vt.NewMockTerm and can be used to preset
// the mock terminal size, capabilities, or other behaviour.
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
