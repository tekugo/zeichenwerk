package next

import (
	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"
)

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
