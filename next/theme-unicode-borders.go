package next

func AddUnicodeBorders(theme *Theme) {
	theme.SetBorders(map[string]*Border{
		// "thin" - Standard single-line borders (┌─┐│└─┘)
		// Most commonly used style, provides clear boundaries without visual heaviness.
		// Ideal for general-purpose widgets, forms, and content containers.
		"thin": &Border{
			Top:          "\u2500", // ─ (horizontal line)
			Right:        "\u2502", // │ (vertical line)
			Bottom:       "\u2500", // ─ (horizontal line)
			Left:         "\u2502", // │ (vertical line)
			TopLeft:      "\u250c", // ┌ (top-left corner)
			TopRight:     "\u2510", // ┐ (top-right corner)
			BottomRight:  "\u2518", // ┘ (bottom-right corner)
			BottomLeft:   "\u2514", // └ (bottom-left corner)
			TopT:         "\u252c", // ┬ (top T-junction)
			RightT:       "\u2524", // ┤ (right T-junction)
			BottomT:      "\u2534", // ┴ (bottom T-junction)
			LeftT:        "\u251c", // ├ (left T-junction)
			InnerH:       "\u2500", // ─ (inner horizontal)
			InnerV:       "\u2502", // │ (inner vertical)
			InnerX:       "\u253c", // ┼ (inner cross)
			InnerTopT:    "\u252c", // ┬ (inner top T)
			InnerRightT:  "\u2524", // ┤ (inner right T)
			InnerBottomT: "\u2534", // ┴ (inner bottom T)
			InnerLeftT:   "\u251c", // ├ (inner left T)
		},

		// "double" - Double-line borders (╔═╗║╚═╝)
		// Creates strong visual emphasis and clear hierarchy separation.
		// Best for important dialogs, primary containers, or emphasized sections.
		"double": &Border{
			Top:          "\u2550", // ═ (double horizontal)
			Right:        "\u2551", // ║ (double vertical)
			Bottom:       "\u2550", // ═ (double horizontal)
			Left:         "\u2551", // ║ (double vertical)
			TopLeft:      "\u2554", // ╔ (double top-left)
			TopRight:     "\u2557", // ╗ (double top-right)
			BottomRight:  "\u255d", // ╝ (double bottom-right)
			BottomLeft:   "\u255a", // ╚ (double bottom-left)
			TopT:         "\u2566", // ╦ (double top T)
			RightT:       "\u2563", // ╣ (double right T)
			BottomT:      "\u2569", // ╩ (double bottom T)
			LeftT:        "\u2560", // ╠ (double left T)
			InnerH:       "\u2550", // ═ (double inner horizontal)
			InnerV:       "\u2551", // ║ (double inner vertical)
			InnerX:       "\u256c", // ╬ (double inner cross)
			InnerTopT:    "\u2566", // ╦ (double inner top T)
			InnerRightT:  "\u2563", // ╣ (double inner right T)
			InnerBottomT: "\u2569", // ╩ (double inner bottom T)
			InnerLeftT:   "\u2560", // ╠ (double inner left T)
		},

		// "double-thin" - Double outer borders with thin inner grid lines (╔═╗║╚═╝ with ─│┼)
		// Combines strong visual emphasis of double borders with subtle inner organization.
		// Perfect for important containers that need internal structure, like data tables
		// in dialogs, primary dashboard widgets, or featured content sections with subdivisions.
		"double-thin": &Border{
			Top:          "\u2550", // ═ (double horizontal)
			Right:        "\u2551", // ║ (double vertical)
			Bottom:       "\u2550", // ═ (double horizontal)
			Left:         "\u2551", // ║ (double vertical)
			TopLeft:      "\u2554", // ╔ (double top-left)
			TopRight:     "\u2557", // ╗ (double top-right)
			BottomRight:  "\u255d", // ╝ (double bottom-right)
			BottomLeft:   "\u255a", // ╚ (double bottom-left)
			TopT:         "\u2564", // ╤ (double-thin top T)
			RightT:       "\u2562", // ╡ (double-thin right T)
			BottomT:      "\u2567", // ╧ (double-thin bottom T)
			LeftT:        "\u255f", // ╞ (double-thin left T)
			InnerH:       "\u2500", // ─ (thin inner horizontal)
			InnerV:       "\u2502", // │ (thin inner vertical)
			InnerX:       "\u253c", // ┼ (thin inner cross)
			InnerTopT:    "\u252c", // ┬ (thin inner top T)
			InnerRightT:  "\u2524", // ┤ (thin inner right T)
			InnerBottomT: "\u2534", // ┴ (thin inner bottom T)
			InnerLeftT:   "\u251c", // ├ (thin inner left T)
		},

		// "round" - Rounded corners with thin lines (╭─╮│╰─╯)
		// Provides a modern, friendly appearance with softer visual impact.
		// Ideal for user-friendly interfaces, welcome screens, or casual applications.
		"round": &Border{
			Top:          "\u2500", // ─ (horizontal line)
			Right:        "\u2502", // │ (vertical line)
			Bottom:       "\u2500", // ─ (horizontal line)
			Left:         "\u2502", // │ (vertical line)
			TopLeft:      "\u256d", // ╭ (rounded top-left)
			TopRight:     "\u256e", // ╮ (rounded top-right)
			BottomRight:  "\u256f", // ╯ (rounded bottom-right)
			BottomLeft:   "\u2570", // ╰ (rounded bottom-left)
			TopT:         "\u252c", // ┬ (top T-junction)
			RightT:       "\u2524", // ┤ (right T-junction)
			BottomT:      "\u2534", // ┴ (bottom T-junction)
			LeftT:        "\u251c", // ├ (left T-junction)
			InnerH:       "\u2500", // ─ (inner horizontal)
			InnerV:       "\u2502", // │ (inner vertical)
			InnerX:       "\u253c", // ┼ (inner cross)
			InnerTopT:    "\u252c", // ┬ (inner top T)
			InnerRightT:  "\u2524", // ┤ (inner right T)
			InnerBottomT: "\u2534", // ┴ (inner bottom T)
			InnerLeftT:   "\u251c", // ├ (inner left T)
		},

		// "thick" - Bold single-line borders (┏━┓┃┗━┛)
		// Creates strong visual weight and clear boundaries.
		// Suitable for alerts, warnings, or primary action areas requiring attention.
		"thick": &Border{
			Top:          "\u2501", // ━ (thick horizontal)
			Right:        "\u2503", // ┃ (thick vertical)
			Bottom:       "\u2501", // ━ (thick horizontal)
			Left:         "\u2503", // ┃ (thick vertical)
			TopLeft:      "\u250f", // ┏ (thick top-left)
			TopRight:     "\u2513", // ┓ (thick top-right)
			BottomRight:  "\u251b", // ┛ (thick bottom-right)
			BottomLeft:   "\u2517", // ┗ (thick bottom-left)
			TopT:         "\u2533", // ┳ (thick top T)
			RightT:       "\u252b", // ┫ (thick right T)
			BottomT:      "\u253b", // ┻ (thick bottom T)
			LeftT:        "\u2523", // ┣ (thick left T)
			InnerH:       "\u2501", // ━ (thick inner horizontal)
			InnerV:       "\u2503", // ┃ (thick inner vertical)
			InnerX:       "\u254b", // ╋ (thick inner cross)
			InnerTopT:    "\u2533", // ┳ (thick inner top T)
			InnerRightT:  "\u252b", // ┫ (thick inner right T)
			InnerBottomT: "\u253b", // ┻ (thick inner bottom T)
			InnerLeftT:   "\u2523", // ┣ (thick inner left T)
		},

		// "thick-thin" - Mixed weight borders (thick outer, thin inner)
		// Combines strong outer boundaries with subtle inner grid lines.
		// Perfect for complex layouts like tables or dashboards with hierarchical data.
		"thick-thin": &Border{
			Top:          "\u2501", // ━ (thick horizontal)
			Right:        "\u2503", // ┃ (thick vertical)
			Bottom:       "\u2501", // ━ (thick horizontal)
			Left:         "\u2503", // ┃ (thick vertical)
			TopLeft:      "\u250f", // ┏ (thick top-left)
			TopRight:     "\u2513", // ┓ (thick top-right)
			BottomRight:  "\u251b", // ┛ (thick bottom-right)
			BottomLeft:   "\u2517", // ┗ (thick bottom-left)
			TopT:         "\u252f", // ┰ (thick-thin top T)
			RightT:       "\u2528", // ┨ (thick-thin right T)
			BottomT:      "\u2537", // ┸ (thick-thin bottom T)
			LeftT:        "\u2520", // ┠ (thick-thin left T)
			InnerH:       "\u2500", // ─ (thin inner horizontal)
			InnerV:       "\u2502", // │ (thin inner vertical)
			InnerX:       "\u253c", // ┼ (thin inner cross)
			InnerTopT:    "\u252c", // ┬ (thin inner top T)
			InnerRightT:  "\u2524", // ┤ (thin inner right T)
			InnerBottomT: "\u2534", // ┴ (thin inner bottom T)
			InnerLeftT:   "\u251c", // ├ (thin inner left T)
		},

		// "thick-slashed" - Thick borders with dashed inner lines
		// Provides strong outer definition with subtle, non-intrusive inner divisions.
		// Useful for data tables where inner grid should be present but not dominant.
		"thick-slashed": &Border{
			Top:          "\u2501", // ━ (thick horizontal)
			Right:        "\u2503", // ┃ (thick vertical)
			Bottom:       "\u2501", // ━ (thick horizontal)
			Left:         "\u2503", // ┃ (thick vertical)
			TopLeft:      "\u250f", // ┏ (thick top-left)
			TopRight:     "\u2513", // ┓ (thick top-right)
			BottomRight:  "\u251b", // ┛ (thick bottom-right)
			BottomLeft:   "\u2517", // ┗ (thick bottom-left)
			TopT:         "\u252f", // ┰ (thick-thin top T)
			RightT:       "\u2528", // ┨ (thick-thin right T)
			BottomT:      "\u2537", // ┸ (thick-thin bottom T)
			LeftT:        "\u2520", // ┠ (thick-thin left T)
			InnerH:       "\u2508", // ┈ (dashed horizontal)
			InnerV:       "\u250a", // ┊ (dashed vertical)
			InnerX:       "\u253c", // ┼ (standard cross)
			InnerTopT:    "\u252c", // ┬ (standard top T)
			InnerRightT:  "\u2524", // ┤ (standard right T)
			InnerBottomT: "\u2534", // ┴ (standard bottom T)
			InnerLeftT:   "\u251c", // ├ (standard left T)
		},

		// "lines" - Minimalist horizontal lines only
		// Provides subtle content separation without visual clutter.
		// Ideal for clean, minimal interfaces or content that needs gentle organization.
		"lines": &Border{
			Top:          "\u2594", // ▔ (upper block)
			Right:        " ",      //   (space - no right border)
			Bottom:       "\u2581", // ▁ (lower block)
			Left:         " ",      //   (space - no left border)
			TopLeft:      "\u2594", // ▔ (upper block)
			TopRight:     "\u2594", // ▔ (upper block)
			BottomRight:  "\u2581", // ▁ (lower block)
			BottomLeft:   "\u2581", // ▁ (lower block)
			TopT:         "\u2594", // ▔ (upper block)
			RightT:       " ",      //   (space)
			BottomT:      "\u2581", // ▁ (lower block)
			LeftT:        " ",      //   (space)
			InnerH:       " ",      //   (space - no inner grid)
			InnerV:       " ",      //   (space - no inner grid)
			InnerX:       " ",      //   (space - no inner grid)
			InnerTopT:    " ",      //   (space - no inner grid)
			InnerRightT:  " ",      //   (space - no inner grid)
			InnerBottomT: " ",      //   (space - no inner grid)
			InnerLeftT:   " ",      //   (space - no inner grid)
		},

		// "lines2" - Minimalist horizontal lines only
		// Provides subtle content separation without visual clutter.
		// Ideal for clean, minimal interfaces or content that needs gentle organization.
		"lines2": &Border{
			Top:          "\u2581",
			Right:        " ",
			Bottom:       "\u2594",
			Left:         " ",
			TopLeft:      "\u2581",
			TopRight:     "\u2581",
			BottomRight:  "\u2594",
			BottomLeft:   "\u2594",
			TopT:         "\u2581",
			RightT:       " ",
			BottomT:      "\u2594",
			LeftT:        " ",
			InnerH:       " ",
			InnerV:       " ",
			InnerX:       " ",
			InnerTopT:    " ",
			InnerRightT:  " ",
			InnerBottomT: " ",
			InnerLeftT:   " ",
		},
	})
}
