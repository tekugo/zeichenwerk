package main

import (
	"fmt"
	"math"
	"math/rand"

	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/widgets"
)

var displayEntries = []Entry{
	{
		Category: "Display",
		Name:     "BarChart",
		Summary:  "Multi-series stacked bar chart with axes, grid, and legend.",
		DocFile:  "bar-chart.md",
		DemoFn:   barChartDemoFn,
		Builder: `builder.BarChart("chart").Hint(-1, 14)
chart := builder.Find("chart").(*BarChart)
chart.SetCategories([]string{"Q1", "Q2", "Q3", "Q4"})
chart.SetSeries([]BarSeries{
    {Label: "Revenue", Values: []float64{120, 145, 161, 188}},
    {Label: "Profit",  Values: []float64{ 38,  46,  52,  64}},
})
chart.SetShowValues(true)`,
		Compose: `chart := widgets.NewBarChart("chart", "")
chart.SetCategories([]string{"Q1", "Q2", "Q3", "Q4"})
chart.SetSeries([]widgets.BarSeries{ /* … */ })
compose.Include(func(_ *core.Theme) core.Widget { return chart })`,
	},
	{
		Category: "Display",
		Name:     "Breadcrumb",
		Summary:  "Path-style segment indicator with click + keyboard navigation.",
		DocFile:  "breadcrumb.md",
		DemoFn:   breadcrumbDemoFn,
		Builder: `builder.Breadcrumb("path").Hint(-1, 1)
bc := builder.Find("path").(*Breadcrumb)
bc.Set([]string{"Home", "Projects", "zeichenwerk", "cmd", "demo2"})`,
		Compose: `compose.Breadcrumb("path", "", compose.Hint(-1, 1))
// then: bc := core.Find(ui, "path").(*widgets.Breadcrumb); bc.Set(…)`,
	},
	{
		Category: "Display",
		Name:     "Canvas",
		Summary:  "Low-level pixel buffer with cursor and modal editing (NORMAL/INSERT).",
		DocFile:  "canvas.md",
		DemoFn:   canvasDemo,
		Builder: `c := NewCanvas("canvas", "", 1, 40, 12)
c.SetCell(0, 0, "★", NewStyle("").WithColors("yellow", "black"))
builder.Add(c)`,
		Compose: `compose.Canvas("canvas", "", 1, 40, 12)`,
	},
	{
		Category: "Display",
		Name:     "ColorPanel",
		Summary:  "Theme-colour preview palette — every named theme color as a swatch.",
		DocFile:  "color-panel.md",
		DemoFn:   colorPanelDemo,
		Builder: `builder.ColorPanel("palette", "Theme Colors")`,
		Compose: `compose.ColorPanel("palette", "", "Theme Colors")`,
	},
	{
		Category: "Display",
		Name:     "ColorPicker",
		Summary:  "Interactive RGB / HSL / Hex colour selector with optional contrast readout.",
		DocFile:  "color-picker.md",
		DemoFn:   colorPickerDemoFn,
		Builder: `builder.ColorPicker("cp-fgbg", ColorFgBg)
cp := builder.Find("cp-fgbg").(*ColorPicker)
cp.SetForeground("#ffffff")
cp.SetBackground("#1a1b26")`,
		Compose: `compose.ColorPicker("cp-fgbg", "", widgets.ColorFgBg)`,
	},
	{
		Category: "Display",
		Name:     "Deck",
		Summary:  "Fixed-height list of items with a custom renderer per cell.",
		DocFile:  "deck.md",
		DemoFn:   deckDemoFn,
		Builder: `render := func(r *Renderer, x, y, w, h, i int, d any, sel, foc bool) {
    item := d.(myItem)
    r.Set("$fg0", "$bg1", "")
    r.Text(x, y, item.Title, w)
    r.Set("$fg2", "$bg1", "")
    r.Text(x, y+1, item.Subtitle, w)
}
deck := NewDeck("deck", "", render, 3) // itemHeight = 3 rows
deck.Set(items)
builder.Add(deck)`,
		Compose: `// Build the Deck imperatively, then attach via Include:
compose.Include(func(_ *core.Theme) core.Widget { return deck })`,
	},
	{
		Category: "Display",
		Name:     "Digits",
		Summary:  "Large ASCII-art numerals built from box-drawing characters.",
		DocFile:  "digits.md",
		DemoFn:   digitsDemo,
		Builder: `builder.Digits("clock", "12:34")`,
		Compose: `compose.Digits("clock", "", "12:34")`,
	},
	{
		Category: "Display",
		Name:     "Heatmap",
		Summary:  "Coloured cell grid for matrix data; interpolates between zero/max colours.",
		DocFile:  "heatmap.md",
		DemoFn:   heatmapDemoFn,
		Builder: `builder.Heatmap("hm", 7, 24).Hint(-1, -1)
hm := builder.Find("hm").(*Heatmap)
hm.SetCellWidth(2)
hm.SetAll(values)`,
		Compose: `compose.Heatmap("hm", "", 7, 24)`,
	},
	{
		Category: "Display",
		Name:     "Indicator",
		Summary:  "Status glyph + label whose colour follows the indicator level.",
		DocFile:  "indicator.md",
		DemoFn:   indicatorDemoFn,
		Builder: `builder.Indicator("ok", Success, "Build succeeded")`,
		Compose: `compose.Indicator("ok", "", core.Success, "Build succeeded")`,
	},
	{
		Category: "Display",
		Name:     "Rule",
		Summary:  "Horizontal or vertical line separator. Style names map to theme borders.",
		DocFile:  "rule.md",
		DemoFn:   ruleDemo,
		Builder: `builder.HRule("thin")     // thin horizontal rule
builder.HRule("double")  // double horizontal rule
builder.VRule("thin")    // vertical separator (use inside HFlex)`,
		Compose: `compose.HRule("", "thin"),
compose.HRule("", "double"),
compose.VRule("", "thin")`,
	},
	{
		Category: "Display",
		Name:     "Shortcuts",
		Summary:  "Single-row keyboard hint bar built from alternating key/label pairs.",
		DocFile:  "shortcuts.md",
		DemoFn:   shortcutsDemo,
		Builder: `builder.Shortcuts("hints",
    "↑↓", "navigate",
    "Tab", "focus",
    "Enter", "activate",
    "q", "quit",
)`,
		Compose: `compose.Shortcuts("hints", "", []string{
    "↑↓", "navigate",
    "Tab", "focus",
    "Enter", "activate",
    "q", "quit",
})`,
	},
	{
		Category: "Display",
		Name:     "Sparkline",
		Summary:  "Inline trend chart: relative or absolute scale, optional threshold + gradient.",
		DocFile:  "sparkline.md",
		DemoFn:   sparklineDemoFn,
		Builder: `builder.Sparkline("sp").Hint(-1, 1)
sp := builder.Find("sp").(*Sparkline)
sp.SetValues(values)`,
		Compose: `compose.Sparkline("sp", "", compose.Hint(-1, 1))`,
	},
	{
		Category: "Display",
		Name:     "Static",
		Summary:  "Plain non-interactive text label.",
		DocFile:  "static.md",
		DemoFn:   staticDemo,
		Builder: `builder.Static("greeting", "Hello, world!")
builder.Static("centre", "Centred text").Padding(0, 4)`,
		Compose: `compose.Static("greeting", "", "Hello, world!"),
compose.Static("centre", "", "Centred text", compose.Padding(0, 4))`,
	},
	{
		Category: "Display",
		Name:     "Styled",
		Summary:  "Rich text with Markdown subset: headings, lists, code, blockquotes, tables.",
		DocFile:  "styled.md",
		DemoFn:   styledDemo,
		Builder: `builder.Styled("doc", "# Title\n\nSome **bold** and *italic* text.").Hint(0, -1)`,
		Compose: `compose.Styled("doc", "", "# Title\n\nSome **bold** and *italic* text.", compose.Hint(0, -1))`,
	},
	{
		Category: "Display",
		Name:     "Table",
		Summary:  "Tabular data display with optional cell navigation.",
		DocFile:  "table.md",
		DemoFn:   tableDemo,
		Builder: `provider := NewArrayTableProvider(headers, rows)
builder.Table("tbl", provider, false).Hint(0, -1)`,
		Compose: `compose.Table("tbl", "", provider, false, compose.Hint(0, -1))`,
	},
	{
		Category: "Display",
		Name:     "Tabs",
		Summary:  "Tab navigation strip; activate via click, Enter, or letter shortcut.",
		DocFile:  "tabs.md",
		DemoFn:   tabsDemo,
		Builder: `builder.Tabs("tabs", "First", "Second", "Third")`,
		Compose: `compose.Tabs("tabs", "", []string{"First", "Second", "Third"})`,
	},
	{
		Category: "Display",
		Name:     "Terminal",
		Summary:  "Embedded VT100/ANSI terminal emulator. Implements io.Writer.",
		DocFile:  "terminal.md",
		DemoFn:   terminalDemoFn,
		Builder: `term := NewTerminal("term", "")
term.SetHint(0, -1)
builder.Add(term)
// Pipe ANSI/VT100 output into it:
term.Write([]byte("\033[1;36mHello\033[0m\r\n"))`,
		Compose: `compose.Terminal("term", "", compose.Hint(0, -1))`,
	},
	{
		Category: "Display",
		Name:     "Text",
		Summary:  "Multi-line scrollable text. Optional follow-mode for live logs.",
		DocFile:  "text.md",
		DemoFn:   textDemo,
		Builder: `builder.Text("log", lines, true /* follow */, 1000 /* max lines */).Hint(0, -1)`,
		Compose: `compose.Text("log", "", lines, true, 1000, compose.Hint(0, -1))`,
	},
	{
		Category: "Display",
		Name:     "Tiles",
		Summary:  "Wrapping grid of fixed-size tiles rendered by a callback.",
		DocFile:  "tiles.md",
		DemoFn:   tilesDemoFn,
		Builder: `tiles := NewTiles("tiles", "", render, 14, 4) // tw=14, th=4
tiles.SetItems(items)
builder.Add(tiles)`,
		Compose: `compose.Include(func(_ *core.Theme) core.Widget { return tiles })`,
	},
}

// ── Demo functions ────────────────────────────────────────────────────────────

func barChartDemoFn(b *Builder) {
	categories := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	series := []BarSeries{
		{Label: "Revenue", Values: []float64{42, 55, 61, 49, 70, 88, 95, 83, 74, 66, 52, 78}},
		{Label: "Costs", Values: []float64{30, 32, 35, 33, 38, 41, 44, 40, 37, 35, 30, 36}},
		{Label: "Profit", Values: []float64{12, 23, 26, 16, 32, 47, 51, 43, 37, 31, 22, 42}},
	}
	b.VFlex("bc-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Stacked bar chart with y-axis, grid, and legend.").
		Padding(0, 0, 1, 0).
		Static("v-label", "Vertical:").
		BarChart("bc-vert").Hint(-1, 14).
		Static("h-label", "Horizontal:").Padding(1, 0, 0, 0).
		BarChart("bc-horiz").Hint(-1, 10).
		End()

	v := b.Find("bc-vert").(*BarChart)
	v.SetCategories(categories)
	v.SetSeries(series)
	v.SetShowValues(true)

	h := b.Find("bc-horiz").(*BarChart)
	h.SetCategories(categories)
	h.SetSeries(series)
	h.SetHorizontal(true)
}

func breadcrumbDemoFn(b *Builder) {
	path := []string{"Home", "Projects", "zeichenwerk", "cmd", "demo2"}
	b.VFlex("bc-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Click a segment or press Enter on it to truncate the path. ←→ navigate.").
		Padding(0, 0, 1, 0).
		Breadcrumb("bc").Hint(-1, 1).
		HFlex("ctl", Center, 2).Padding(1, 0, 0, 0).
		Button("push", "Push").
		Button("pop", "Pop").
		Button("reset", "Reset").
		End().
		Static("status", "").Padding(1, 0, 0, 0).
		End()

	bc := b.Find("bc").(*Breadcrumb)
	bc.Set(path)
	st := b.Find("status").(*Static)
	bc.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		if i, ok := data[0].(int); ok {
			segs := bc.Segments()
			if i < len(segs) {
				st.Set(fmt.Sprintf("Selected: [%d] %s", i, segs[i]))
			}
		}
		return true
	})
	bc.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if i, ok := data[0].(int); ok {
			bc.Truncate(i)
			st.Set(fmt.Sprintf("Truncated to %d", i))
		}
		return true
	})
	b.Find("push").(*Button).On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		bc.Push(fmt.Sprintf("dir%d", len(bc.Segments())))
		return true
	})
	b.Find("pop").(*Button).On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		bc.Pop()
		return true
	})
	b.Find("reset").(*Button).On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		bc.Set(path)
		return true
	})
}

func canvasDemo(b *Builder) {
	c := NewCanvas("canvas", "", 1, 40, 12)
	normal := NewStyle("").WithColors("white", "black").WithCursor("block")
	insert := NewStyle("").WithColors("cyan", "black").WithCursor("bar")
	c.SetStyle("", normal)
	c.SetStyle(":insert", insert)
	c.Fill("", normal)
	border := NewStyle("").WithColors("yellow", "black")
	for x := 1; x < 39; x++ {
		c.SetCell(x, 0, "─", border)
		c.SetCell(x, 11, "─", border)
	}
	for y := 1; y < 11; y++ {
		c.SetCell(0, y, "│", border)
		c.SetCell(39, y, "│", border)
	}
	c.SetCell(0, 0, "┌", border)
	c.SetCell(39, 0, "┐", border)
	c.SetCell(0, 11, "└", border)
	c.SetCell(39, 11, "┘", border)
	title := NewStyle("").WithColors("green", "black")
	c.SetCell(2, 2, "C", title)
	c.SetCell(3, 2, "a", title)
	c.SetCell(4, 2, "n", title)
	c.SetCell(5, 2, "v", title)
	c.SetCell(6, 2, "a", title)
	c.SetCell(7, 2, "s", title)

	b.VFlex("canvas-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Canvas is a 2D buffer of (rune, style) cells. Press 'i' to enter INSERT mode, ESC to leave.").
		Padding(0, 0, 1, 0).
		Add(c).
		End()
}

func colorPanelDemo(b *Builder) {
	b.VFlex("cp-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "ColorPanel shows every named colour from the active theme. Switch themes in the header to see it update.").
		Padding(0, 0, 1, 0).
		ColorPanel("palette", "Theme Colors").Hint(0, -1).
		End()
}

func colorPickerDemoFn(b *Builder) {
	b.VFlex("cp-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Edit any of R/G/B, H/S/L, or Hex — all representations stay in sync.").
		Padding(0, 0, 1, 0).
		Static("single", "Single colour:").
		ColorPicker("single-cp", ColorSingle).Padding(1, 0).
		Static("fgbg", "Foreground / background with contrast ratio:").Padding(1, 0, 0, 0).
		ColorPicker("fgbg-cp", ColorFgBg).Padding(1, 0).
		End()
	if cp, ok := b.Find("single-cp").(*ColorPicker); ok {
		cp.SetForeground("#ff8040")
	}
	if cp, ok := b.Find("fgbg-cp").(*ColorPicker); ok {
		cp.SetForeground("#ffffff")
		cp.SetBackground("#1a1b26")
	}
}

type colorItem struct {
	name string
	hex  string
}

func deckDemoFn(b *Builder) {
	items := []any{
		colorItem{"Cyan", "#7dcfff"},
		colorItem{"Blue", "#7aa2f7"},
		colorItem{"Purple", "#bb9af7"},
		colorItem{"Green", "#9ece6a"},
		colorItem{"Yellow", "#e0af68"},
		colorItem{"Orange", "#ff9e64"},
		colorItem{"Red", "#f7768e"},
		colorItem{"Magenta", "#ad8ee6"},
	}
	render := func(r *Renderer, x, y, w, h, _ int, data any, selected, focused bool) {
		item := data.(colorItem)
		bg := "$bg1"
		fg := "$fg0"
		font := ""
		if selected {
			font = "bold"
			if focused {
				fg = "$cyan"
			}
		}
		r.Set(fg, bg, font)
		r.Fill(x, y, w, h, " ")
		r.Text(x+2, y, item.name, w-2)
		r.Set("$fg2", bg, "")
		r.Text(x+2, y+1, item.hex, w-2)
		// Swatch on the right.
		r.Set("", item.hex, "")
		r.Fill(x+w-10, y, 8, 2, " ")
	}
	deck := NewDeck("deck", "", render, 3)
	deck.Set(items)

	b.VFlex("deck-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Deck renders a list of items with a custom callback. itemHeight controls each cell.").
		Padding(0, 0, 1, 0).
		Add(deck).Hint(0, -1).
		End()
}

func digitsDemo(b *Builder) {
	b.VFlex("digits-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Digits renders large ASCII-art numerals. Useful for clocks, counters, KPIs.").
		Padding(0, 0, 1, 0).
		HFlex("row", Center, 2).
		Digits("d1", "12:34").
		End().
		Spacer().Hint(0, 1).
		HFlex("row2", Center, 2).
		Digits("d2", "99.5").
		End().
		End()
}

func heatmapDemoFn(b *Builder) {
	const rows, cols = 24, 7
	b.VFlex("hm-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "24×7 activity heatmap. Cells interpolate between heatmap/zero and heatmap/max colours.").
		Padding(0, 0, 1, 0).
		Heatmap("hm", rows, cols).Hint(-1, -1).
		End()
	hm := b.Find("hm").(*Heatmap)
	hm.SetCellWidth(2)
	hm.SetColLabels([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"})
	rowLabels := make([]string, rows)
	for i := range rowLabels {
		rowLabels[i] = fmt.Sprintf("%2dh", i)
	}
	hm.SetRowLabels(rowLabels)
	data := make([][]float64, rows)
	for r := range data {
		data[r] = make([]float64, cols)
		base := 0.1
		if r >= 8 && r < 18 {
			base = 0.45
		}
		for c := range data[r] {
			b := base
			if c >= 5 {
				b *= 0.4
			}
			data[r][c] = b + rand.Float64()*0.5
		}
	}
	hm.SetAll(data)
}

func indicatorDemoFn(b *Builder) {
	b.VFlex("ind-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Indicators render a level-coloured glyph followed by a label. Ideal for status rows.").
		Padding(0, 0, 1, 0).
		Indicator("ind1", Debug, "Debug — verbose diagnostics").
		Indicator("ind2", Info, "Info — connection established").
		Indicator("ind3", Success, "Success — build completed").
		Indicator("ind4", Warning, "Warning — disk usage at 85%").
		Indicator("ind5", Error, "Error — request timed out").
		Indicator("ind6", Fatal, "Fatal — process aborted").
		End()
}

func ruleDemo(b *Builder) {
	b.VFlex("rule-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Horizontal and vertical rules. Style names refer to theme borders.").
		Padding(0, 0, 1, 0).
		Static("h1", "Thin horizontal:").
		HRule("thin").
		Spacer().Hint(0, 1).
		Static("h2", "Double horizontal:").
		HRule("double").
		Spacer().Hint(0, 1).
		Static("h3", "Vertical (between two columns):").
		HFlex("v-row", Stretch, 0).Hint(0, 5).
		Static("left", "Left side").
		VRule("thin").
		Static("right", "Right side").
		End().
		End()
}

func shortcutsDemo(b *Builder) {
	b.VFlex("sc-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Shortcuts shows alternating (key, label) pairs in a single row. Common at the bottom of an app.").
		Padding(0, 0, 1, 0).
		Shortcuts("hints",
			"↑↓", "navigate",
			"Tab", "focus",
			"Enter", "activate",
			"Esc", "cancel",
			"q", "quit").
		End()
}

func sparklineDemoFn(b *Builder) {
	makeData := func(fn func(int) float64) []float64 {
		v := make([]float64, 60)
		for i := range v {
			v[i] = fn(i)
		}
		return v
	}
	sin := func(i int) float64 { return math.Sin(float64(i)*0.3) + rand.Float64()*0.3 - 0.15 }
	abs := func(i int) float64 { return (math.Sin(float64(i)*0.2) + 1.0) / 2.0 }

	b.VFlex("sp-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Sparklines summarise time series in a single row. h>1 packs more vertical resolution.").
		Padding(0, 0, 1, 0).
		Static("l1", "Relative scale (h=1):").
		Sparkline("sp1").Hint(-1, 1).
		Static("l2", "Absolute scale, h=2:").Padding(1, 0, 0, 0).
		Sparkline("sp2").Hint(-1, 2).
		Static("l3", "Threshold + gradient, h=3:").Padding(1, 0, 0, 0).
		Sparkline("sp3").Hint(-1, 3).
		Static("l4", "Multi-row absolute, h=4:").Padding(1, 0, 0, 0).
		Sparkline("sp4").Hint(-1, 4).
		End()

	feed := func(sp *Sparkline, vs []float64) {
		rb := NewRingBuffer[float64](120)
		for _, v := range vs {
			rb.Add(v)
		}
		sp.SetProvider(rb)
	}
	feed(b.Find("sp1").(*Sparkline), makeData(sin))
	sp2 := b.Find("sp2").(*Sparkline)
	sp2.SetAbsolute(true)
	sp2.SetMin(-1)
	sp2.SetMax(1)
	feed(sp2, makeData(sin))
	sp3 := b.Find("sp3").(*Sparkline)
	sp3.SetAbsolute(true)
	sp3.SetMin(0)
	sp3.SetMax(1)
	sp3.SetThreshold(0.65)
	sp3.SetGradient(true)
	feed(sp3, makeData(abs))
	sp4 := b.Find("sp4").(*Sparkline)
	sp4.SetAbsolute(true)
	sp4.SetMin(-1)
	sp4.SetMax(1)
	feed(sp4, makeData(sin))
}

func staticDemo(b *Builder) {
	b.VFlex("static-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Static is a non-interactive label. Use it everywhere you need plain text.").
		Padding(0, 0, 1, 0).
		Static("plain", "Plain text").
		Spacer().Hint(0, 1).
		Static("centred", "Centred — uses HFlex Center alignment.").Padding(0, 4).
		Spacer().Hint(0, 1).
		HFlex("row", Stretch, 2).
		Static("a", "Left").
		Static("b", "Centre").Hint(-1, 0).
		Static("c", "Right").
		End().
		End()
}

const styledDemoText = `# Styled Widget

The **Styled** widget renders a Markdown subset with word wrapping.

## Inline styles

Plain text alongside *italic*, **bold**, __underlined__, ~~strikethrough~~ and ` + "`code`" + `.

## Lists

- First item
- Second item
- Third item

1. Step one
2. Step two
3. Step three

## Code block

` + "```" + `
func hello() {
    fmt.Println("Hello")
}
` + "```" + `

## Blockquote

> A quoted paragraph rendered with a left border and muted colours.

---

Use **↑ ↓** to scroll · **PgUp PgDn** for page · **Home End** for top/bottom.`

func styledDemo(b *Builder) {
	b.VFlex("styled-demo", Stretch, 0).Hint(0, -1).
		Styled("doc", styledDemoText).Hint(0, -1).
		End()
}

func tableDemo(b *Builder) {
	headers := []string{"Service", "Status", "CPU", "Memory", "Uptime"}
	rows := [][]string{
		{"nginx", "● running", "0.3%", " 24 MB", "14d 06h"},
		{"postgresql", "● running", "2.1%", "512 MB", "14d 06h"},
		{"redis", "● running", "0.1%", " 64 MB", "14d 05h"},
		{"celery", "○ stopped", "  —  ", "    —  ", "     —  "},
		{"prometheus", "● running", "1.8%", "128 MB", " 2d 11h"},
		{"grafana", "● running", "0.9%", " 96 MB", " 2d 11h"},
		{"elasticsearch", "● running", "3.4%", "1.2 GB", "11d 03h"},
		{"kibana", "● running", "0.7%", "256 MB", "11d 03h"},
	}
	b.VFlex("tbl-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Table renders any TableProvider. Use NewArrayTableProvider for in-memory data.").
		Padding(0, 0, 1, 0).
		Table("tbl", NewArrayTableProvider(headers, rows), false).Hint(0, -1).
		End()
}

func tabsDemo(b *Builder) {
	b.VFlex("tabs-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Tabs is a navigation strip. Pair with Switcher and connect=true to drive panes.").
		Padding(0, 0, 1, 0).
		Tabs("tabs", "First", "Second", "Third", "Fourth").
		Spacer().Hint(0, 1).
		Static("hint", "←/→ navigate · Enter activate · 1234 jump by index.").
		End()
}

func terminalDemoFn(b *Builder) {
	term := NewTerminal("term", "")
	term.SetHint(0, -1)
	b.VFlex("term-demo", Stretch, 0).Hint(0, -1).Padding(0, 1).
		Static("desc", "Terminal renders ANSI/VT100 sequences. Pipe any io.Writer-compatible source into it.").
		Padding(0, 0, 1, 0).
		Add(term).Hint(0, -1).
		End()
	pane := b.Find("term-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		term.Clear()
		w := func(s string) { _, _ = term.Write([]byte(s)) }
		w("\033[1;36mTerminal — VT100/ANSI demo\033[0m\r\n")
		w("\033[2m─────────────────────────────────────────\033[0m\r\n\r\n")
		w("\033[1mBold\033[0m  \033[2mDim\033[0m  \033[3mItalic\033[0m  \033[4mUnderline\033[0m\r\n\r\n")
		w("Standard colors: ")
		for i := 0; i < 8; i++ {
			w(fmt.Sprintf("\033[%dm  \033[0m", 40+i))
		}
		w("\r\n")
		w("Bright colors:   ")
		for i := 0; i < 8; i++ {
			w(fmt.Sprintf("\033[%dm  \033[0m", 100+i))
		}
		w("\r\n\r\n")
		w("256-colour cube:\r\n")
		for row := 0; row < 4; row++ {
			w("  ")
			for col := 0; col < 36; col++ {
				idx := 16 + row*36 + col
				if idx > 231 {
					break
				}
				w(fmt.Sprintf("\033[48;5;%dm  \033[0m", idx))
			}
			w("\r\n")
		}
		w("\r\nTrue colour gradient: ")
		for i := 0; i < 48; i++ {
			r := 255 * i / 47
			b := 255 - r
			w(fmt.Sprintf("\033[48;2;%d;128;%dm \033[0m", r, b))
		}
		w("\r\n\r\nBox-drawing: ┌──────────┐\r\n             │  \033[1;32mHello!\033[0m  │\r\n             └──────────┘\r\n")
		return true
	})
}

func textDemo(b *Builder) {
	lines := []string{}
	for i := 1; i <= 60; i++ {
		lines = append(lines, fmt.Sprintf("Line %2d — Text widget renders a wrapped, scrollable, multi-line string buffer.", i))
	}
	b.VFlex("text-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Text is a multi-line scrollable buffer. follow=true keeps the bottom in view (live logs).").
		Padding(0, 0, 1, 0).
		Text("body", lines, false, 1000).Hint(0, -1).
		End()
}

type tileItem struct {
	name  string
	icon  string
	color string
}

func tilesDemoFn(b *Builder) {
	cards := []tileItem{
		{"Dashboard", "◈", "$blue"},
		{"Analytics", "▦", "$green"},
		{"Reports", "▤", "$yellow"},
		{"Settings", "⚙", "$fg2"},
		{"Users", "◉", "$blue"},
		{"Billing", "◎", "$orange"},
		{"Security", "◆", "$red"},
		{"Integrations", "⬡", "$cyan"},
		{"Logs", "≡", "$fg2"},
		{"API Keys", "◆", "$purple"},
		{"Webhooks", "◈", "$blue"},
		{"Audit Trail", "▤", "$yellow"},
	}
	items := make([]any, len(cards))
	for i, c := range cards {
		items[i] = c
	}
	render := func(r *Renderer, x, y, w, h, _ int, data any, selected, focused bool) {
		c := data.(tileItem)
		bg := "$bg2"
		fg := "$fg1"
		if selected && focused {
			bg = "$blue"
			fg = "$bg0"
		} else if selected {
			bg = "$bg3"
			fg = "$fg0"
		}
		r.Set(fg, bg, "")
		r.Fill(x, y, w, h, " ")
		iconX := x + max(0, (w-1)/2)
		r.Set(c.color, bg, "bold")
		r.Put(iconX, y+1, c.icon)
		nameRunes := []rune(c.name)
		nameX := x + max(0, (w-len(nameRunes))/2)
		r.Set(fg, bg, "")
		r.Text(nameX, y+2, c.name, w-(nameX-x))
	}

	b.VFlex("tiles-demo", Stretch, 0).Padding(0, 2).
		Static("desc", "Tiles is a wrapping grid of fixed-size cells rendered by a callback.").
		Padding(0, 0, 0, 0).
		Tiles("tiles", render, 14, 4).Hint(-1, -1).
		End()
	g := b.Find("tiles").(*Tiles)
	g.SetItems(items)
}

