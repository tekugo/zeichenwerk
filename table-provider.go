package zeichenwerk

// TableColumn defines the structure and properties of a table column.
// Each column has a header, width, and optional sorting/filtering capabilities.
type TableColumn struct {
	Header     string // Display text for the column header
	Width      int    // Column width in characters (auto-calculated for ArrayTableProvider)
	Sortable   bool   // Whether this column supports sorting (not yet implemented)
	Filterable bool   // Whether this column supports filtering (not yet implemented)
}

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

// ArrayTableProvider is a concrete implementation of TableProvider that
// stores table data in memory as a 2D slice of strings. This is the most
// common provider for static or small datasets.
//
// The provider automatically calculates optimal column widths based on
// the longest content in each column, considering both header text and data.
type ArrayTableProvider struct {
	columns []TableColumn // Column definitions with calculated widths
	data    [][]string    // 2D array of string data (rows x columns)
}

// NewArrayTableProvider creates a new array-based table provider with
// automatic column width calculation. The width of each column is determined
// by the longest content (header or data) in that column.
//
// Parameters:
//   - headers: Slice of column header strings
//   - data: 2D slice of string data [rows][columns]
//
// Returns:
//   - *ArrayTableProvider: Configured provider ready for use with Table widget
//
// Example:
//   headers := []string{"Name", "Age", "City"}
//   data := [][]string{
//       {"John Doe", "25", "New York"},
//       {"Jane Smith", "30", "Los Angeles"},
//   }
//   provider := NewArrayTableProvider(headers, data)
//
// Note: The function assumes all data rows have the same number of columns
// as the headers slice. Mismatched row lengths may cause runtime panics.
func NewArrayTableProvider(headers []string, data [][]string) *ArrayTableProvider {
	table := ArrayTableProvider{
		columns: make([]TableColumn, len(headers)),
		data:    data,
	}

	// Determine column width by longest value in all rows
	for i, header := range headers {
		column := TableColumn{Header: header, Width: len([]rune(header))}
		for j := range len(data) {
			if len(data[j][i]) > column.Width {
				column.Width = len([]rune(data[j][i]))
			}
		}
		table.columns[i] = column
	}
	return &table
}

// Columns returns the column definitions for this table.
// Each column includes the header text and calculated width.
//
// Returns:
//   - []TableColumn: Slice of column definitions
func (a *ArrayTableProvider) Columns() []TableColumn {
	return a.columns
}

// Length returns the number of data rows in the table.
// This count excludes the header row.
//
// Returns:
//   - int: Number of data rows
func (a *ArrayTableProvider) Length() int {
	return len(a.data)
}

// Str returns the string content of the cell at the specified position.
// Row and column indices are zero-based, with row 0 being the first data row.
//
// Parameters:
//   - row: Zero-based row index
//   - column: Zero-based column index
//
// Returns:
//   - string: Cell content as string
//
// Note: This method does not perform bounds checking. Accessing invalid
// indices will cause a runtime panic.
func (a *ArrayTableProvider) Str(row, column int) string {
	return a.data[row][column]
}
