package zeichenwerk

import "testing"

// ── NewArrayTableProvider ─────────────────────────────────────────────────────

func TestArrayTableProvider_Columns_Headers(t *testing.T) {
	p := NewArrayTableProvider([]string{"Name", "Age"}, nil)
	cols := p.Columns()
	if len(cols) != 2 {
		t.Fatalf("len(Columns()) = %d; want 2", len(cols))
	}
	if cols[0].Header != "Name" {
		t.Errorf("cols[0].Header = %q; want %q", cols[0].Header, "Name")
	}
	if cols[1].Header != "Age" {
		t.Errorf("cols[1].Header = %q; want %q", cols[1].Header, "Age")
	}
}

func TestArrayTableProvider_Width_FromHeader(t *testing.T) {
	p := NewArrayTableProvider([]string{"Name", "Age"}, [][]string{
		{"Al", "5"},
	})
	cols := p.Columns()
	if cols[0].Width != 4 { // "Name" = 4
		t.Errorf("cols[0].Width = %d; want 4", cols[0].Width)
	}
	if cols[1].Width != 3 { // "Age" = 3
		t.Errorf("cols[1].Width = %d; want 3", cols[1].Width)
	}
}

func TestArrayTableProvider_Width_FromData(t *testing.T) {
	p := NewArrayTableProvider([]string{"N", "C"}, [][]string{
		{"Alice", "Berlin"},
		{"Bob", "New York City"},
	})
	cols := p.Columns()
	if cols[0].Width != 5 { // "Alice" = 5
		t.Errorf("cols[0].Width = %d; want 5 (longest data)", cols[0].Width)
	}
	if cols[1].Width != 13 { // "New York City" = 13
		t.Errorf("cols[1].Width = %d; want 13 (longest data)", cols[1].Width)
	}
}

func TestArrayTableProvider_Width_HeaderWinsOverData(t *testing.T) {
	p := NewArrayTableProvider([]string{"Location"}, [][]string{
		{"NY"},
	})
	cols := p.Columns()
	if cols[0].Width != 8 { // "Location" = 8, beats "NY" = 2
		t.Errorf("cols[0].Width = %d; want 8 (header longer than data)", cols[0].Width)
	}
}

func TestArrayTableProvider_Width_MultibyteRunes(t *testing.T) {
	// "über" = 4 runes, "日本語" = 3 runes
	p := NewArrayTableProvider([]string{"über"}, [][]string{
		{"日本語"},
	})
	cols := p.Columns()
	if cols[0].Width != 4 { // header "über" has 4 runes
		t.Errorf("cols[0].Width = %d; want 4 (rune count of header)", cols[0].Width)
	}
}

func TestArrayTableProvider_Length_Empty(t *testing.T) {
	p := NewArrayTableProvider([]string{"X"}, nil)
	if p.Length() != 0 {
		t.Errorf("Length() = %d; want 0 for nil data", p.Length())
	}
}

func TestArrayTableProvider_Length(t *testing.T) {
	p := NewArrayTableProvider([]string{"X"}, [][]string{
		{"a"}, {"b"}, {"c"},
	})
	if p.Length() != 3 {
		t.Errorf("Length() = %d; want 3", p.Length())
	}
}

func TestArrayTableProvider_Str(t *testing.T) {
	p := NewArrayTableProvider([]string{"Name", "Age"}, [][]string{
		{"Alice", "30"},
		{"Bob", "25"},
	})
	cases := []struct{ row, col int; want string }{
		{0, 0, "Alice"},
		{0, 1, "30"},
		{1, 0, "Bob"},
		{1, 1, "25"},
	}
	for _, tc := range cases {
		got := p.Str(tc.row, tc.col)
		if got != tc.want {
			t.Errorf("Str(%d, %d) = %q; want %q", tc.row, tc.col, got, tc.want)
		}
	}
}

func TestArrayTableProvider_Columns_EmptyHeaders(t *testing.T) {
	p := NewArrayTableProvider(nil, nil)
	if len(p.Columns()) != 0 {
		t.Errorf("len(Columns()) = %d; want 0 for nil headers", len(p.Columns()))
	}
}

func TestArrayTableProvider_ImplementsInterface(t *testing.T) {
	var _ TableProvider = (*ArrayTableProvider)(nil)
}
