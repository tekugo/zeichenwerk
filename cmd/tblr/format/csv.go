package format

import (
	"bytes"
	"encoding/csv"
	"strings"
)

// CSVFormat handles CSV and TSV files.
type CSVFormat struct {
	name         string
	ext          []string
	defaultDelim rune
}

func (f *CSVFormat) Name() string       { return f.name }
func (f *CSVFormat) Extensions() []string { return f.ext }

func (f *CSVFormat) Detect(data []byte) bool {
	line := firstNonEmpty(string(data))
	if line == "" {
		return false
	}
	switch f.defaultDelim {
	case '\t':
		return strings.Contains(line, "\t")
	default:
		return strings.Contains(line, ",")
	}
}

func (f *CSVFormat) Parse(data []byte, opts ParseOpts) (*MutableTable, error) {
	delim := opts.Delimiter
	if delim == 0 {
		delim = f.defaultDelim
		if delim == 0 {
			// auto-detect
			line := firstNonEmpty(string(data))
			if strings.Contains(line, "\t") {
				delim = '\t'
			} else {
				delim = ','
			}
		}
	}

	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = delim
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	t := NewMutableTable()
	t.SetDelimiter(delim)

	if len(records) == 0 {
		return t, nil
	}

	headers := records[0]
	var rows [][]string
	if len(records) > 1 {
		rows = records[1:]
	}
	t.Load(headers, rows)
	return t, nil
}

func (f *CSVFormat) Serialize(t *MutableTable, opts SerialOpts) ([]byte, error) {
	delim := opts.Delimiter
	if delim == 0 {
		delim = t.Delimiter()
		if delim == 0 {
			delim = f.defaultDelim
			if delim == 0 {
				delim = ','
			}
		}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	w.Comma = delim

	// header row
	headers := make([]string, t.ColCount())
	for i := range headers {
		headers[i] = t.Header(i)
	}
	if err := w.Write(headers); err != nil {
		return nil, err
	}

	// data rows
	for i := 0; i < t.Length(); i++ {
		row := make([]string, t.ColCount())
		for j := range row {
			row[j] = t.Str(i, j)
		}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

// firstNonEmpty returns the first non-empty line in s.
func firstNonEmpty(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}
