package widgets

import "github.com/tekugo/zeichenwerk/core"

// TableColumn defines the structure and properties of a table column.
type TableColumn struct {
	Header     string         // Display text for the column header
	Width      int            // Column width in characters (auto-calculated for ArrayTableProvider)
	Alignment  core.Alignment // AlignLeft, AlignCenter, or AlignRight
	Sortable   bool           // Whether this column supports sorting (not yet implemented)
	Filterable bool           // Whether this column supports filtering (not yet implemented)
}
