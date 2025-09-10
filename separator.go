package zeichenwerk

type Separator struct {
	BaseWidget
	Border string
}

func NewSeparator(id, border string) *Separator {
	return &Separator{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Border:     border,
	}
}
