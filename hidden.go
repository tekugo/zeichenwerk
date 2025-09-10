package tui

type Hidden struct {
	BaseWidget
}

func NewHidden(id string) *Hidden {
	return &Hidden{
		BaseWidget: BaseWidget{id: id, focusable: false},
	}
}
