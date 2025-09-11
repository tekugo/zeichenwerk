package zeichenwerk

type Tabs struct {
	BaseWidget
	Tabs  []string
	Index int
}

func NewTabs(id string) *Tabs {
	return &Tabs{
		BaseWidget: BaseWidget{id: id, focusable: true},
		Tabs:       make([]string, 0),
		Index:      0,
	}
}

func (t *Tabs) Add(title string) {
	t.Tabs = append(t.Tabs, title)
}

func (t *Tabs) Hint() (int, int) {
	width := 0
	for _, tab := range t.Tabs {
		width += len([]rune(tab))
	}
	width += len(t.Tabs)*2 + 2
	return width, 2
}
