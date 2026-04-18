package zeichenwerk

// ItemRender is the render function type for Deck slots. It is called once per
// visible slot with the renderer, slot bounds, item index, the data item,
// whether the slot is the currently highlighted item, and whether the Deck
// widget itself currently holds keyboard focus.
type ItemRender func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool)
