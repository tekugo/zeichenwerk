package widgets

// TableProvider defines the interface that data sources must implement
// to supply data to a Table widget. This abstraction allows tables to
// work with various data sources (arrays, databases, APIs, etc.).
type TableProvider interface {
	// Columns returns the column definitions for the table.
	// This includes headers, widths, and column properties.
	Columns() []TableColumn

	// Length returns the total number of data rows available.
	// This excludes the header row.
	Length() int

	// Str returns the string representation of the cell at the
	// specified row and column indices. Row indices start at 0
	// for the first data row (header is handled separately).
	Str(row, column int) string
}
