package core

// Alignment specifies how content is positioned within a widget or container
// along a given axis. It is used by layout containers, table columns, and any
// widget that needs to align child content horizontally or vertically.
type Alignment int

// Alignment values used to position content within widgets and containers.
//
// Start/End are axis-agnostic (leading/trailing edge) while Left/Right are
// explicitly horizontal. Stretch expands the content to fill the available
// space instead of positioning it at a single point.
const (
	Default Alignment = 0 // widget- or container-specific default alignment
	Start   Alignment = 1 // align to the leading edge (top or left)
	Left    Alignment = 2 // align to the left edge
	Center  Alignment = 3 // center within the available space
	Right   Alignment = 4 // align to the right edge
	End     Alignment = 5 // align to the trailing edge (bottom or right)
	Stretch Alignment = 6 // expand content to fill the available space
)

// String returns the lowercase name of the alignment value (for example
// "center" or "stretch"). Unknown or zero values return "default", which makes
// the method safe to call on any Alignment without panicking.
func (a Alignment) String() string {
	switch a {
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
