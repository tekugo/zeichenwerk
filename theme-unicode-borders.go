package zeichenwerk

// BorderStyle defines the Unicode characters used to draw borders and grid lines for widgets.
// It provides a comprehensive set of border elements including outer borders, corners,
// connectors, and inner grid elements for creating visually consistent widget boundaries
// and complex grid layouts.
//
// The BorderStyle supports two main use cases:
//   - Simple widget borders: Uses outer border elements (Top, Right, Bottom, Left, corners)
//   - Complex grid layouts: Uses inner elements for drawing internal grid lines and connectors
//
// Border elements are organized into categories:
//   - Outer borders: Form the perimeter of widgets
//   - Corners: Connect perpendicular border segments
//   - T-connectors: Join borders at T-junctions on the perimeter
//   - Inner elements: Create internal grid structures within widgets
//
// All rune values should be Unicode box-drawing characters for proper terminal display.
// The style system automatically selects appropriate characters based on the context
// and neighboring elements to create seamless border connections.
type BorderStyle struct {
	// Outer border elements - form the perimeter of widgets
	Top    rune // Horizontal line for the top border
	Right  rune // Vertical line for the right border
	Bottom rune // Horizontal line for the bottom border
	Left   rune // Vertical line for the left border

	// Corner elements - connect perpendicular border segments
	TopLeft     rune // Corner connecting top and left borders
	TopRight    rune // Corner connecting top and right borders
	BottomRight rune // Corner connecting bottom and right borders
	BottomLeft  rune // Corner connecting bottom and left borders

	// Outer T-connectors - join borders at T-junctions on the perimeter
	TopT    rune // T-connector extending downward from the top border
	RightT  rune // T-connector extending leftward from the right border
	BottomT rune // T-connector extending upward from the bottom border
	LeftT   rune // T-connector extending rightward from the left border

	// Inner grid elements - create internal structures within widgets
	InnerH rune // Horizontal line for internal grid divisions
	InnerV rune // Vertical line for internal grid divisions
	InnerX rune // Cross connector for internal grid intersections

	// Inner T-connectors - join internal grid lines at T-junctions
	InnerTopT    rune // T-connector extending downward from internal horizontal lines
	InnerRightT  rune // T-connector extending leftward from internal vertical lines
	InnerBottomT rune // T-connector extending upward from internal horizontal lines
	InnerLeftT   rune // T-connector extending rightward from internal vertical lines
}

func AddUnicodeBorders(theme Theme) {
	theme.SetBorders(map[string]BorderStyle{ // "thin" - Standard single-line borders (┌─┐│└─┘)
		// Most commonly used style, provides clear boundaries without visual heaviness.
		// Ideal for general-purpose widgets, forms, and content containers.
		"thin": {
			Top:          rune(0x2500), // ─ (horizontal line)
			Right:        rune(0x2502), // │ (vertical line)
			Bottom:       rune(0x2500), // ─ (horizontal line)
			Left:         rune(0x2502), // │ (vertical line)
			TopLeft:      rune(0x250c), // ┌ (top-left corner)
			TopRight:     rune(0x2510), // ┐ (top-right corner)
			BottomRight:  rune(0x2518), // ┘ (bottom-right corner)
			BottomLeft:   rune(0x2514), // └ (bottom-left corner)
			TopT:         rune(0x252c), // ┬ (top T-junction)
			RightT:       rune(0x2524), // ┤ (right T-junction)
			BottomT:      rune(0x2534), // ┴ (bottom T-junction)
			LeftT:        rune(0x251c), // ├ (left T-junction)
			InnerH:       rune(0x2500), // ─ (inner horizontal)
			InnerV:       rune(0x2502), // │ (inner vertical)
			InnerX:       rune(0x253c), // ┼ (inner cross)
			InnerTopT:    rune(0x252c), // ┬ (inner top T)
			InnerRightT:  rune(0x2524), // ┤ (inner right T)
			InnerBottomT: rune(0x2534), // ┴ (inner bottom T)
			InnerLeftT:   rune(0x251c), // ├ (inner left T)
		},

		// "double" - Double-line borders (╔═╗║╚═╝)
		// Creates strong visual emphasis and clear hierarchy separation.
		// Best for important dialogs, primary containers, or emphasized sections.
		"double": {
			Top:          rune(0x2550), // ═ (double horizontal)
			Right:        rune(0x2551), // ║ (double vertical)
			Bottom:       rune(0x2550), // ═ (double horizontal)
			Left:         rune(0x2551), // ║ (double vertical)
			TopLeft:      rune(0x2554), // ╔ (double top-left)
			TopRight:     rune(0x2557), // ╗ (double top-right)
			BottomRight:  rune(0x255d), // ╝ (double bottom-right)
			BottomLeft:   rune(0x255a), // ╚ (double bottom-left)
			TopT:         rune(0x2566), // ╦ (double top T)
			RightT:       rune(0x2563), // ╣ (double right T)
			BottomT:      rune(0x2569), // ╩ (double bottom T)
			LeftT:        rune(0x2560), // ╠ (double left T)
			InnerH:       rune(0x2550), // ═ (double inner horizontal)
			InnerV:       rune(0x2551), // ║ (double inner vertical)
			InnerX:       rune(0x256c), // ╬ (double inner cross)
			InnerTopT:    rune(0x2566), // ╦ (double inner top T)
			InnerRightT:  rune(0x2563), // ╣ (double inner right T)
			InnerBottomT: rune(0x2569), // ╩ (double inner bottom T)
			InnerLeftT:   rune(0x2560), // ╠ (double inner left T)
		},

		// "round" - Rounded corners with thin lines (╭─╮│╰─╯)
		// Provides a modern, friendly appearance with softer visual impact.
		// Ideal for user-friendly interfaces, welcome screens, or casual applications.
		"round": {
			Top:          rune(0x2500), // ─ (horizontal line)
			Right:        rune(0x2502), // │ (vertical line)
			Bottom:       rune(0x2500), // ─ (horizontal line)
			Left:         rune(0x2502), // │ (vertical line)
			TopLeft:      rune(0x256d), // ╭ (rounded top-left)
			TopRight:     rune(0x256e), // ╮ (rounded top-right)
			BottomRight:  rune(0x256f), // ╯ (rounded bottom-right)
			BottomLeft:   rune(0x2570), // ╰ (rounded bottom-left)
			TopT:         rune(0x252c), // ┬ (top T-junction)
			RightT:       rune(0x2524), // ┤ (right T-junction)
			BottomT:      rune(0x2534), // ┴ (bottom T-junction)
			LeftT:        rune(0x251c), // ├ (left T-junction)
			InnerH:       rune(0x2500), // ─ (inner horizontal)
			InnerV:       rune(0x2502), // │ (inner vertical)
			InnerX:       rune(0x253c), // ┼ (inner cross)
			InnerTopT:    rune(0x252c), // ┬ (inner top T)
			InnerRightT:  rune(0x2524), // ┤ (inner right T)
			InnerBottomT: rune(0x2534), // ┴ (inner bottom T)
			InnerLeftT:   rune(0x251c), // ├ (inner left T)
		},

		// "thick" - Bold single-line borders (┏━┓┃┗━┛)
		// Creates strong visual weight and clear boundaries.
		// Suitable for alerts, warnings, or primary action areas requiring attention.
		"thick": {
			Top:          rune(0x2501), // ━ (thick horizontal)
			Right:        rune(0x2503), // ┃ (thick vertical)
			Bottom:       rune(0x2501), // ━ (thick horizontal)
			Left:         rune(0x2503), // ┃ (thick vertical)
			TopLeft:      rune(0x250f), // ┏ (thick top-left)
			TopRight:     rune(0x2513), // ┓ (thick top-right)
			BottomRight:  rune(0x251b), // ┛ (thick bottom-right)
			BottomLeft:   rune(0x2517), // ┗ (thick bottom-left)
			TopT:         rune(0x2533), // ┳ (thick top T)
			RightT:       rune(0x252b), // ┫ (thick right T)
			BottomT:      rune(0x253b), // ┻ (thick bottom T)
			LeftT:        rune(0x2523), // ┣ (thick left T)
			InnerH:       rune(0x2501), // ━ (thick inner horizontal)
			InnerV:       rune(0x2503), // ┃ (thick inner vertical)
			InnerX:       rune(0x254b), // ╋ (thick inner cross)
			InnerTopT:    rune(0x2533), // ┳ (thick inner top T)
			InnerRightT:  rune(0x252b), // ┫ (thick inner right T)
			InnerBottomT: rune(0x253b), // ┻ (thick inner bottom T)
			InnerLeftT:   rune(0x2523), // ┣ (thick inner left T)
		},

		// "thick-thin" - Mixed weight borders (thick outer, thin inner)
		// Combines strong outer boundaries with subtle inner grid lines.
		// Perfect for complex layouts like tables or dashboards with hierarchical data.
		"thick-thin": {
			Top:          rune(0x2501), // ━ (thick horizontal)
			Right:        rune(0x2503), // ┃ (thick vertical)
			Bottom:       rune(0x2501), // ━ (thick horizontal)
			Left:         rune(0x2503), // ┃ (thick vertical)
			TopLeft:      rune(0x250f), // ┏ (thick top-left)
			TopRight:     rune(0x2513), // ┓ (thick top-right)
			BottomRight:  rune(0x251b), // ┛ (thick bottom-right)
			BottomLeft:   rune(0x2517), // ┗ (thick bottom-left)
			TopT:         rune(0x252f), // ┰ (thick-thin top T)
			RightT:       rune(0x2528), // ┨ (thick-thin right T)
			BottomT:      rune(0x2537), // ┸ (thick-thin bottom T)
			LeftT:        rune(0x2520), // ┠ (thick-thin left T)
			InnerH:       rune(0x2500), // ─ (thin inner horizontal)
			InnerV:       rune(0x2502), // │ (thin inner vertical)
			InnerX:       rune(0x253c), // ┼ (thin inner cross)
			InnerTopT:    rune(0x252c), // ┬ (thin inner top T)
			InnerRightT:  rune(0x2524), // ┤ (thin inner right T)
			InnerBottomT: rune(0x2534), // ┴ (thin inner bottom T)
			InnerLeftT:   rune(0x251c), // ├ (thin inner left T)
		},

		// "thick-slashed" - Thick borders with dashed inner lines
		// Provides strong outer definition with subtle, non-intrusive inner divisions.
		// Useful for data tables where inner grid should be present but not dominant.
		"thick-slashed": {
			Top:          rune(0x2501), // ━ (thick horizontal)
			Right:        rune(0x2503), // ┃ (thick vertical)
			Bottom:       rune(0x2501), // ━ (thick horizontal)
			Left:         rune(0x2503), // ┃ (thick vertical)
			TopLeft:      rune(0x250f), // ┏ (thick top-left)
			TopRight:     rune(0x2513), // ┓ (thick top-right)
			BottomRight:  rune(0x251b), // ┛ (thick bottom-right)
			BottomLeft:   rune(0x2517), // ┗ (thick bottom-left)
			TopT:         rune(0x252f), // ┰ (thick-thin top T)
			RightT:       rune(0x2528), // ┨ (thick-thin right T)
			BottomT:      rune(0x2537), // ┸ (thick-thin bottom T)
			LeftT:        rune(0x2520), // ┠ (thick-thin left T)
			InnerH:       rune(0x2508), // ┈ (dashed horizontal)
			InnerV:       rune(0x250a), // ┊ (dashed vertical)
			InnerX:       rune(0x253c), // ┼ (standard cross)
			InnerTopT:    rune(0x252c), // ┬ (standard top T)
			InnerRightT:  rune(0x2524), // ┤ (standard right T)
			InnerBottomT: rune(0x2534), // ┴ (standard bottom T)
			InnerLeftT:   rune(0x251c), // ├ (standard left T)
		},

		// "lines" - Minimalist horizontal lines only
		// Provides subtle content separation without visual clutter.
		// Ideal for clean, minimal interfaces or content that needs gentle organization.
		"lines": {
			Top:          rune(0x2594), // ▔ (upper block)
			Right:        ' ',          //   (space - no right border)
			Bottom:       rune(0x2581), // ▁ (lower block)
			Left:         ' ',          //   (space - no left border)
			TopLeft:      rune(0x2594), // ▔ (upper block)
			TopRight:     rune(0x2594), // ▔ (upper block)
			BottomRight:  rune(0x2581), // ▁ (lower block)
			BottomLeft:   rune(0x2581), // ▁ (lower block)
			TopT:         rune(0x2594), // ▔ (upper block)
			RightT:       ' ',          //   (space)
			BottomT:      rune(0x2581), // ▁ (lower block)
			LeftT:        ' ',          //   (space)
			InnerH:       ' ',          //   (space - no inner grid)
			InnerV:       ' ',          //   (space - no inner grid)
			InnerX:       ' ',          //   (space - no inner grid)
			InnerTopT:    ' ',          //   (space - no inner grid)
			InnerRightT:  ' ',          //   (space - no inner grid)
			InnerBottomT: ' ',          //   (space - no inner grid)
			InnerLeftT:   ' ',          //   (space - no inner grid)
		},
	})
}
