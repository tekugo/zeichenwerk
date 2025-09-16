package zeichenwerk

type TableColumn struct {
	Header     string
	Width      int
	Sortable   bool
	Filterable bool
}

type TableProvider interface {
	Columns() []TableColumn
	Length() int
	Str(int, int) string
}

type ArrayTableProvider struct {
	columns []TableColumn
	data    [][]string
}

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

func (a *ArrayTableProvider) Columns() []TableColumn {
	return a.columns
}

func (a *ArrayTableProvider) Length() int {
	return len(a.data)
}

func (a *ArrayTableProvider) Str(row, column int) string {
	return a.data[row][column]
}
