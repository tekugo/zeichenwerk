package format

// Format handles one table syntax.
type Format interface {
	Name() string
	Extensions() []string
	Detect(data []byte) bool
	Parse(data []byte, opts ParseOpts) (*MutableTable, error)
	Serialize(t *MutableTable, opts SerialOpts) ([]byte, error)
}

// ParseOpts controls parsing behaviour.
type ParseOpts struct {
	Delimiter rune // CSV/TSV only; 0 = auto-detect
}

// SerialOpts controls serialisation behaviour.
type SerialOpts struct {
	Pretty    bool // pad columns to uniform width
	Delimiter rune // CSV/TSV only
}

var all = []Format{
	&MarkdownFormat{},
	&AsciiDocFormat{},
	&TypstFormat{},
	&HTMLFormat{},
	&CSVFormat{name: "tsv", ext: []string{"tsv"}, defaultDelim: '\t'},
	&CSVFormat{name: "csv", ext: []string{"csv"}, defaultDelim: ','},
}

// All returns all formats in detection order.
func All() []Format { return all }

// ByName returns the format with the given name, or nil.
func ByName(name string) Format {
	for _, f := range all {
		if f.Name() == name {
			return f
		}
	}
	return nil
}

// Detect returns the first format whose Detect heuristic matches data.
// CSV is the fallback; returns nil only for empty data.
func Detect(data []byte) Format {
	if len(data) == 0 {
		return nil
	}
	for _, f := range all {
		if f.Detect(data) {
			return f
		}
	}
	// fallback: CSV
	return ByName("csv")
}
