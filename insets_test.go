package zeichenwerk

import (
	"testing"
)

// TestNewInsets tests the creation of new Insets instances using various value configurations
func TestNewInsets(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   Insets
	}{
		{
			name:   "no values - all zeros",
			values: []int{},
			want:   Insets{Top: 0, Right: 0, Bottom: 0, Left: 0},
		},
		{
			name:   "single value - uniform insets",
			values: []int{5},
			want:   Insets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		},
		{
			name:   "two values - vertical/horizontal",
			values: []int{10, 20},
			want:   Insets{Top: 10, Right: 20, Bottom: 10, Left: 20},
		},
		{
			name:   "three values - top, horizontal, bottom",
			values: []int{1, 2, 3},
			want:   Insets{Top: 1, Right: 2, Bottom: 3, Left: 2},
		},
		{
			name:   "four values - clockwise from top",
			values: []int{1, 2, 3, 4},
			want:   Insets{Top: 1, Right: 2, Bottom: 3, Left: 4},
		},
		{
			name:   "more than four values - uses first four",
			values: []int{1, 2, 3, 4, 5, 6},
			want:   Insets{Top: 1, Right: 2, Bottom: 3, Left: 4},
		},
		{
			name:   "negative values",
			values: []int{-1, -2, -3, -4},
			want:   Insets{Top: -1, Right: -2, Bottom: -3, Left: -4},
		},
		{
			name:   "mixed positive and negative values",
			values: []int{10, -5, 8, -3},
			want:   Insets{Top: 10, Right: -5, Bottom: 8, Left: -3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewInsets(tt.values...)
			if got != tt.want {
				t.Errorf("NewInsets(%v) = %+v, want %+v", tt.values, got, tt.want)
			}
		})
	}
}

// TestInsetsSet tests the Set method for configuring insets
func TestInsetsSet(t *testing.T) {
	tests := []struct {
		name   string
		values []int
		want   Insets
	}{
		{
			name:   "no values - all zeros",
			values: []int{},
			want:   Insets{Top: 0, Right: 0, Bottom: 0, Left: 0},
		},
		{
			name:   "single value - uniform insets",
			values: []int{7},
			want:   Insets{Top: 7, Right: 7, Bottom: 7, Left: 7},
		},
		{
			name:   "two values - vertical/horizontal",
			values: []int{15, 25},
			want:   Insets{Top: 15, Right: 25, Bottom: 15, Left: 25},
		},
		{
			name:   "three values - top, horizontal, bottom",
			values: []int{5, 10, 15},
			want:   Insets{Top: 5, Right: 10, Bottom: 15, Left: 10},
		},
		{
			name:   "four values - clockwise from top",
			values: []int{2, 4, 6, 8},
			want:   Insets{Top: 2, Right: 4, Bottom: 6, Left: 8},
		},
		{
			name:   "more than four values - uses first four",
			values: []int{10, 20, 30, 40, 50, 60, 70},
			want:   Insets{Top: 10, Right: 20, Bottom: 30, Left: 40},
		},
		{
			name:   "zero values",
			values: []int{0, 0, 0, 0},
			want:   Insets{Top: 0, Right: 0, Bottom: 0, Left: 0},
		},
		{
			name:   "large values",
			values: []int{1000, 2000, 3000, 4000},
			want:   Insets{Top: 1000, Right: 2000, Bottom: 3000, Left: 4000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insets := Insets{}
			insets.Set(tt.values...)
			if insets != tt.want {
				t.Errorf("Set(%v) = %+v, want %+v", tt.values, insets, tt.want)
			}
		})
	}
}

// TestInsetsSetOverwrite tests that Set properly overwrites existing values
func TestInsetsSetOverwrite(t *testing.T) {
	insets := Insets{Top: 10, Right: 20, Bottom: 30, Left: 40}
	
	// Set with new values should overwrite existing ones
	insets.Set(1, 2, 3, 4)
	want := Insets{Top: 1, Right: 2, Bottom: 3, Left: 4}
	
	if insets != want {
		t.Errorf("Set() after existing values = %+v, want %+v", insets, want)
	}
	
	// Set with single value should overwrite all
	insets.Set(99)
	want = Insets{Top: 99, Right: 99, Bottom: 99, Left: 99}
	
	if insets != want {
		t.Errorf("Set() single value overwrite = %+v, want %+v", insets, want)
	}
}

// TestInsetsInfo tests the Info method for string representation
func TestInsetsInfo(t *testing.T) {
	tests := []struct {
		name   string
		insets Insets
		want   string
	}{
		{
			name:   "all zeros",
			insets: Insets{Top: 0, Right: 0, Bottom: 0, Left: 0},
			want:   "(0 0 0 0)",
		},
		{
			name:   "uniform values",
			insets: Insets{Top: 5, Right: 5, Bottom: 5, Left: 5},
			want:   "(5 5 5 5)",
		},
		{
			name:   "different values",
			insets: Insets{Top: 1, Right: 2, Bottom: 3, Left: 4},
			want:   "(1 2 3 4)",
		},
		{
			name:   "negative values",
			insets: Insets{Top: -1, Right: -2, Bottom: -3, Left: -4},
			want:   "(-1 -2 -3 -4)",
		},
		{
			name:   "mixed positive and negative",
			insets: Insets{Top: 10, Right: -5, Bottom: 8, Left: -3},
			want:   "(10 -5 8 -3)",
		},
		{
			name:   "large values",
			insets: Insets{Top: 1000, Right: 2000, Bottom: 3000, Left: 4000},
			want:   "(1000 2000 3000 4000)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.insets.Info()
			if got != tt.want {
				t.Errorf("Info() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestInsetsHorizontal tests the Horizontal method
func TestInsetsHorizontal(t *testing.T) {
	tests := []struct {
		name   string
		insets Insets
		want   int
	}{
		{
			name:   "all zeros",
			insets: Insets{Top: 0, Right: 0, Bottom: 0, Left: 0},
			want:   0,
		},
		{
			name:   "positive values",
			insets: Insets{Top: 10, Right: 5, Bottom: 15, Left: 8},
			want:   13, // 5 + 8
		},
		{
			name:   "negative values",
			insets: Insets{Top: 10, Right: -3, Bottom: 15, Left: -7},
			want:   -10, // -3 + (-7)
		},
		{
			name:   "mixed positive and negative",
			insets: Insets{Top: 10, Right: 15, Bottom: 20, Left: -5},
			want:   10, // 15 + (-5)
		},
		{
			name:   "large values",
			insets: Insets{Top: 100, Right: 1000, Bottom: 200, Left: 2000},
			want:   3000, // 1000 + 2000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.insets.Horizontal()
			if got != tt.want {
				t.Errorf("Horizontal() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestInsetsVertical tests the Vertical method
func TestInsetsVertical(t *testing.T) {
	tests := []struct {
		name   string
		insets Insets
		want   int
	}{
		{
			name:   "all zeros",
			insets: Insets{Top: 0, Right: 0, Bottom: 0, Left: 0},
			want:   0,
		},
		{
			name:   "positive values",
			insets: Insets{Top: 10, Right: 5, Bottom: 15, Left: 8},
			want:   25, // 10 + 15
		},
		{
			name:   "negative values",
			insets: Insets{Top: -3, Right: 10, Bottom: -7, Left: 15},
			want:   -10, // -3 + (-7)
		},
		{
			name:   "mixed positive and negative",
			insets: Insets{Top: 20, Right: 10, Bottom: -5, Left: 15},
			want:   15, // 20 + (-5)
		},
		{
			name:   "large values",
			insets: Insets{Top: 1000, Right: 100, Bottom: 2000, Left: 200},
			want:   3000, // 1000 + 2000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.insets.Vertical()
			if got != tt.want {
				t.Errorf("Vertical() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestInsetsTotal tests the Total method
func TestInsetsTotal(t *testing.T) {
	tests := []struct {
		name        string
		insets      Insets
		wantHoriz   int
		wantVert    int
	}{
		{
			name:      "all zeros",
			insets:    Insets{Top: 0, Right: 0, Bottom: 0, Left: 0},
			wantHoriz: 0,
			wantVert:  0,
		},
		{
			name:      "positive values",
			insets:    Insets{Top: 10, Right: 5, Bottom: 15, Left: 8},
			wantHoriz: 13, // 5 + 8
			wantVert:  25, // 10 + 15
		},
		{
			name:      "negative values",
			insets:    Insets{Top: -3, Right: -5, Bottom: -7, Left: -2},
			wantHoriz: -7, // -5 + (-2)
			wantVert:  -10, // -3 + (-7)
		},
		{
			name:      "mixed positive and negative",
			insets:    Insets{Top: 20, Right: 15, Bottom: -5, Left: -3},
			wantHoriz: 12, // 15 + (-3)
			wantVert:  15, // 20 + (-5)
		},
		{
			name:      "large values",
			insets:    Insets{Top: 1000, Right: 500, Bottom: 2000, Left: 1500},
			wantHoriz: 2000, // 500 + 1500
			wantVert:  3000, // 1000 + 2000
		},
		{
			name:      "asymmetric values",
			insets:    Insets{Top: 1, Right: 100, Bottom: 2, Left: 200},
			wantHoriz: 300, // 100 + 200
			wantVert:  3,   // 1 + 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHoriz, gotVert := tt.insets.Total()
			if gotHoriz != tt.wantHoriz {
				t.Errorf("Total() horizontal = %d, want %d", gotHoriz, tt.wantHoriz)
			}
			if gotVert != tt.wantVert {
				t.Errorf("Total() vertical = %d, want %d", gotVert, tt.wantVert)
			}
		})
	}
}

// TestInsetsConsistency tests that Total() returns the same values as Horizontal() and Vertical()
func TestInsetsConsistency(t *testing.T) {
	testCases := []Insets{
		{Top: 0, Right: 0, Bottom: 0, Left: 0},
		{Top: 5, Right: 5, Bottom: 5, Left: 5},
		{Top: 1, Right: 2, Bottom: 3, Left: 4},
		{Top: -1, Right: -2, Bottom: -3, Left: -4},
		{Top: 10, Right: -5, Bottom: 8, Left: -3},
		{Top: 1000, Right: 2000, Bottom: 3000, Left: 4000},
	}

	for i, insets := range testCases {
		t.Run(insets.Info(), func(t *testing.T) {
			horizontal := insets.Horizontal()
			vertical := insets.Vertical()
			totalHoriz, totalVert := insets.Total()

			if horizontal != totalHoriz {
				t.Errorf("Case %d: Horizontal() = %d, but Total() horizontal = %d", i, horizontal, totalHoriz)
			}
			if vertical != totalVert {
				t.Errorf("Case %d: Vertical() = %d, but Total() vertical = %d", i, vertical, totalVert)
			}
		})
	}
}

// TestInsetsEdgeCases tests edge cases and boundary conditions
func TestInsetsEdgeCases(t *testing.T) {
	t.Run("zero initialization", func(t *testing.T) {
		var insets Insets // zero value
		if insets.Top != 0 || insets.Right != 0 || insets.Bottom != 0 || insets.Left != 0 {
			t.Errorf("Zero value Insets = %+v, want all fields zero", insets)
		}
		if insets.Horizontal() != 0 {
			t.Errorf("Zero value Horizontal() = %d, want 0", insets.Horizontal())
		}
		if insets.Vertical() != 0 {
			t.Errorf("Zero value Vertical() = %d, want 0", insets.Vertical())
		}
	})

	t.Run("large positive values", func(t *testing.T) {
		const largeInt = 1000000
		insets := NewInsets(largeInt, 0, largeInt, 0)
		
		horizontal := insets.Horizontal()
		vertical := insets.Vertical()
		
		if horizontal != 0 {
			t.Errorf("Large int horizontal = %d, want 0", horizontal)
		}
		if vertical != 2*largeInt {
			t.Errorf("Large int vertical = %d, want %d", vertical, 2*largeInt)
		}
	})

	t.Run("large negative values", func(t *testing.T) {
		const largeNegInt = -1000000
		insets := NewInsets(largeNegInt, 0, largeNegInt, 0)
		
		horizontal := insets.Horizontal()
		vertical := insets.Vertical()
		
		if horizontal != 0 {
			t.Errorf("Large negative int horizontal = %d, want 0", horizontal)
		}
		if vertical != 2*largeNegInt {
			t.Errorf("Large negative int vertical = %d, want %d", vertical, 2*largeNegInt)
		}
	})

	t.Run("set empty slice multiple times", func(t *testing.T) {
		insets := NewInsets(1, 2, 3, 4)
		insets.Set() // Should reset to all zeros
		want := Insets{Top: 0, Right: 0, Bottom: 0, Left: 0}
		
		if insets != want {
			t.Errorf("Set() empty = %+v, want %+v", insets, want)
		}
		
		// Do it again to ensure it's consistent
		insets.Set()
		if insets != want {
			t.Errorf("Set() empty second time = %+v, want %+v", insets, want)
		}
	})
}