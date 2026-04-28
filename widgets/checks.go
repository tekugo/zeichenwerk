package widgets

import "github.com/tekugo/zeichenwerk/core"

// Compile-time interface type checks
var (
	_ core.Container = (*Box)(nil)

	_ core.Widget = (*Animation)(nil)
	_ core.Widget = (*Breadcrumb)(nil)
	_ core.Widget = (*BarChart)(nil)
	_ core.Widget = (*Button)(nil)
	_ core.Widget = (*Component)(nil)
)
