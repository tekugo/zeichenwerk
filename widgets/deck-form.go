package widgets

import (
	"fmt"
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// DeckForm is the WidgetForm for *Deck. The ItemRender callback is a
// runtime dependency; only the per-slot height and the scrollbar
// flag are part of the static editing surface.
type DeckForm struct {
	ComponentForm

	ItemHeight int  `group:"general" label:"Item Height"`
	Scrollbar  bool `group:"display" label:"Show Scrollbar"`
}

func (f *DeckForm) Name() string  { return "Deck" }
func (f *DeckForm) Group() string { return "leaf" }
func (f *DeckForm) Help() string  { return "Scrollable list of fixed-height rich items" }

func (f *DeckForm) Load(w core.Widget) {
	d := w.(*Deck)
	f.ComponentForm.Load(&d.Component)
	f.ItemHeight = d.itemHeight
	f.Scrollbar = d.scrollbar
}

func (f *DeckForm) Store(w core.Widget) {
	d := w.(*Deck)
	f.ComponentForm.Store(&d.Component)
	if f.ItemHeight >= 1 {
		d.itemHeight = f.ItemHeight
	}
	d.scrollbar = f.Scrollbar
}

func (f *DeckForm) New() core.Widget {
	h := f.ItemHeight
	if h < 1 {
		h = 1
	}
	d := NewDeck("", "", noopItemRender, h)
	f.Store(d)
	return d
}

func (f *DeckForm) Validate(field string) error { return nil }

func (f *DeckForm) Emit(w io.Writer, mode string) error {
	return f.EmitFrame(w, mode, func() error {
		h := f.ItemHeight
		if h < 1 {
			h = 1
		}
		_, err := fmt.Fprintf(w, "Deck(%q, deckRender /* TODO */, %d).\n", f.ID, h)
		return err
	})
}
