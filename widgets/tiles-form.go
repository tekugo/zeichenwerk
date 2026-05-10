package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// TilesForm is the WidgetForm for *Tiles. The ItemRender callback is
// a runtime dependency injected by the host application; the static
// editing surface only captures the tile dimensions and the
// scrollbar visibility.
type TilesForm struct {
	ComponentForm

	TileWidth  int  `group:"general" label:"Tile Width"`
	TileHeight int  `group:"general" label:"Tile Height"`
	Scrollbar  bool `group:"display" label:"Show Scrollbar"`
}

func (f *TilesForm) Name() string  { return "Tiles" }
func (f *TilesForm) Group() string { return "leaf" }
func (f *TilesForm) Help() string  { return "Scrollable 2D grid of fixed-size tiles" }

func (f *TilesForm) Load(w core.Widget) {
	t := w.(*Tiles)
	f.ComponentForm.Load(&t.Component)
	f.TileWidth = t.tileWidth
	f.TileHeight = t.tileHeight
	f.Scrollbar = t.scrollbar
}

func (f *TilesForm) Store(w core.Widget) {
	t := w.(*Tiles)
	f.ComponentForm.Store(&t.Component)
	if f.TileWidth >= 1 {
		t.tileWidth = f.TileWidth
	}
	if f.TileHeight >= 1 {
		t.tileHeight = f.TileHeight
	}
	t.scrollbar = f.Scrollbar
}

func (f *TilesForm) New() core.Widget {
	tw := f.TileWidth
	if tw < 1 {
		tw = 1
	}
	th := f.TileHeight
	if th < 1 {
		th = 1
	}
	t := NewTiles("", "", noopItemRender, tw, th)
	f.Store(t)
	return t
}

func (f *TilesForm) Validate(field string) error { return nil }

// Emit writes the Tiles constructor with a placeholder render
// identifier. The ItemRender is application-level state that
// codegen cannot synthesise; the user is expected to replace
// "tileRender" with the actual function.
func (f *TilesForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		tw := f.TileWidth
		if tw < 1 {
			tw = 1
		}
		th := f.TileHeight
		if th < 1 {
			th = 1
		}
		_, err := fmt.Fprintf(w, "Tiles(%q, tileRender /* TODO */, %d, %d).\n", f.ID, tw, th)
		return err
	})
}

// noopItemRender is a do-nothing ItemRender used by Tiles- and
// Deck-form New to keep the widget alive when constructed without a
// real render function.
func noopItemRender(_ *core.Renderer, _, _, _, _, _ int, _ any, _, _ bool) {}
