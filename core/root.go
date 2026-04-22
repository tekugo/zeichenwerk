package core

type Root interface {
	Container
	Close()
	Focus(widget Widget)
	Popup(x, y, w, h int, container Container)
	Redraw(widget Widget)
	Theme() *Theme
}
