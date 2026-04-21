package core

type Root interface {
	Bounds() (int, int, int, int)
	Close()
	Focus(widget Widget)
	Layout()
	Popup(x, y, w, h int, container Container)
	Redraw(widget Widget)
	Refresh()
	Theme() *Theme
}
