// Tutorial chapter 3 — Grid layout demo.
//
// Shows row/column sizing with mixed fixed and fractional values,
// and cell spanning via Cell(x, y, w, h).
//
// Run with:  go run ./doc/tutorial/examples/03-layout-grid
package main

import (
	. "github.com/tekugo/zeichenwerk"
	"github.com/tekugo/zeichenwerk/themes"
)

func main() {
	NewBuilder(themes.TokyoNight()).
		// 3 rows × 3 columns. lines=true draws cell borders.
		Grid("layout", 3, 3, true).
		// Sizes: positive = fixed cells, negative = fraction.
		Columns(20, -1, -2). // 20 cells | 1 fraction | 2 fractions
		Rows(3, -1, 1).      // 3 rows   | grow       | 1 row
		// Cell(x, y, w, h) targets the next widget at column x, row y,
		// spanning w columns and h rows.
		Cell(0, 0, 3, 1).Static("banner", "Spans all 3 columns").
		Foreground("$cyan").Font("bold").Padding(1, 2).
		Cell(0, 1, 1, 1).Static("nav", "Navigation").Padding(1, 2).Background("$bg1").
		Cell(1, 1, 1, 1).Static("main", "Main content").Padding(1, 2).
		Cell(2, 1, 1, 1).Static("aside", "Aside (2× nav width)").Padding(1, 2).Background("$bg1").
		Cell(0, 2, 3, 1).Static("status", "Status bar").Padding(0, 2).Foreground("$gray").
		End().
		Run()
}
