package core

type Alignment int

// Column alignment values for TableColumn.Alignment.
const (
	Default Alignment = 0 // widget/container specific default
	Start   Alignment = 1
	Left    Alignment = 2
	Center  Alignment = 3
	Right   Alignment = 4
	End     Alignment = 5
	Stretch Alignment = 6
)

func AlignmentString(align Alignment) string {
	switch align {
	case Start:
		return "start"
	case Left:
		return "left"
	case Center:
		return "center"
	case Right:
		return "right"
	case End:
		return "end"
	case Stretch:
		return "stretch"
	default:
		return "default"
	}
}
