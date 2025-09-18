package zeichenwerk

type Digits struct {
	BaseWidget
	Text string
}

func NewDigits(id string, text string) *Digits {
	return &Digits{
		BaseWidget: BaseWidget{id: id, focusable: false},
		Text:       text,
	}
}
