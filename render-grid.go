package zeichenwerk

// renderGrid renders a Grid widget with its cells and optional grid lines.
// This method handles the complete rendering of grid layouts, including
// internal grid lines and the rendering of all cell contents.
//
// Parameters:
//   - g: The Grid widget to render
//   - box: BorderStyle to use for drawing grid lines and separators
//
// Rendering process:
//  1. If grid lines are enabled, renders the internal grid structure
//  2. Renders each cell's content widget within its designated area
//
// The grid rendering respects the grid's configuration for:
//   - Line visibility (controlled by g.lines flag)
//   - Cell positioning and spanning
//   - Border style consistency with the overall widget theme
//
// Cell contents are rendered after grid lines to ensure proper layering,
// with cell content appearing above the grid structure.
func (r *Renderer) renderGrid(g *Grid, box BorderStyle) {
	if g.lines {
		r.renderGridLines(g, box)
	}
	for _, cell := range g.cells {
		r.render(cell.content)
	}
}

// renderGridLines draws the internal grid structure with proper line connections and junctions.
// This method creates a complex grid layout with T-junctions, crosses, and line segments
// that properly connect to form a cohesive grid appearance.
//
// Parameters:
//   - g: The Grid widget containing layout information
//   - box: BorderStyle providing Unicode characters for different line types
//
// Grid line rendering process:
//  1. Draws top and bottom T-connectors where vertical lines meet the border
//  2. Draws left and right T-connectors where horizontal lines meet the border
//  3. Renders internal grid lines with proper junction characters
//  4. Calculates connector types based on adjacent line presence
//
// Junction calculation:
//   - Uses bit flags to determine which directions have connecting lines
//   - Selects appropriate Unicode characters for each junction type
//   - Supports complex grid layouts with varying cell spans
//
// The method handles:
//   - Variable column widths and row heights
//   - Cell spanning that affects line connectivity
//   - Proper padding and margin calculations
//   - Border integration with internal grid structure
func (r *Renderer) renderGridLines(g *Grid, box BorderStyle) {
	style := g.Style("")
	_, _, iw, ih := g.Content()

	// draw top and bottom Ts
	cx := g.x + style.Margin.Left + style.Padding.Left
	by := g.y + style.Margin.Top + style.Padding.Top + style.Padding.Bottom + 1 + ih
	for i := range len(g.columns) - 1 {
		cx += g.widths[i] + 1
		if g.separators[0][i]&GridV > 0 {
			r.screen.SetContent(cx, g.y+style.Margin.Top, box.TopT, nil, r.style)
			r.repeat(cx, g.y+style.Margin.Top+1, 0, 1, style.Padding.Top, box.InnerV)
		}
		if g.separators[len(g.rows)-1][i]&GridV > 0 {
			r.screen.SetContent(cx, by, box.BottomT, nil, r.style)
			r.repeat(cx, by-1, 0, -1, style.Padding.Bottom, box.InnerV)
		}
	}

	// draw left and right Ts
	cy := g.y + style.Margin.Top + style.Padding.Top
	rx := g.x + style.Margin.Left + style.Padding.Left + style.Padding.Right + 1 + iw
	for i := range len(g.rows) - 1 {
		cy += g.heights[i] + 1
		if g.separators[i][0]&GridH > 0 {
			r.screen.SetContent(g.x+style.Margin.Left, cy, box.LeftT, nil, r.style)
			r.repeat(g.x+style.Margin.Left+1, cy, 1, 0, style.Padding.Left, box.InnerH)
		}
		if g.separators[i][len(g.columns)-1]&GridH > 0 {
			r.screen.SetContent(rx, cy, box.RightT, nil, r.style)
			r.repeat(rx-1, cy, -1, 0, style.Padding.Right, box.InnerH)
		}
	}

	// draw inner grid lines
	cy = g.y + style.Margin.Top + style.Padding.Top
	for row := range len(g.rows) {
		cx = g.x + style.Margin.Left + style.Padding.Left + g.widths[0] + 1
		cy += g.heights[row] + 1
		for c := range len(g.columns) {
			connector := 0
			if row < len(g.rows)-1 && g.separators[row][c]&GridH > 0 {
				r.repeat(cx-1, cy, -1, 0, g.widths[c], box.InnerH)
				connector |= 8
			}
			if c < len(g.columns)-1 && g.separators[row][c]&GridV > 0 {
				r.repeat(cx, cy-1, 0, -1, g.heights[row], box.InnerV)
				connector |= 1
			}
			if row < len(g.rows)-1 && c < len(g.columns)-1 {
				if g.separators[row+1][c]&GridV > 0 {
					connector |= 4
				}
				if g.separators[row][c+1]&GridH > 0 {
					connector |= 2
				}
				switch connector {
				case 5:
					r.screen.SetContent(cx, cy, box.InnerV, nil, r.style)
				case 7:
					r.screen.SetContent(cx, cy, box.InnerLeftT, nil, r.style)
				case 10:
					r.screen.SetContent(cx, cy, box.InnerH, nil, r.style)
				case 11:
					r.screen.SetContent(cx, cy, box.InnerBottomT, nil, r.style)
				case 13:
					r.screen.SetContent(cx, cy, box.InnerRightT, nil, r.style)
				case 14:
					r.screen.SetContent(cx, cy, box.InnerTopT, nil, r.style)
				case 15:
					r.screen.SetContent(cx, cy, box.InnerX, nil, r.style)
				}
			}
			if c < len(g.columns)-1 {
				cx += g.widths[c+1] + 1
			}
		}
	}
}
