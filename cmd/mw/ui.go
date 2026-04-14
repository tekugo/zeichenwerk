package main

import (
	"fmt"
	"strings"
	"time"

	z "github.com/tekugo/zeichenwerk"
	c "github.com/tekugo/zeichenwerk/compose"
)

// Sparkline window: 30 buckets × 2 minutes = last 60 minutes.
const (
	sparkBuckets  = 30
	sparkInterval = 2 * time.Minute
)

// sparkCache holds per-session sparkline widgets and their backing TimeSeries.
type sparkCache struct {
	input  *z.Sparkline
	output *z.Sparkline
	tsIn   *z.TimeSeries[int64]
	tsOut  *z.TimeSeries[int64]
}

func newSparkCache(id string) sparkCache {
	now := time.Now()
	start := now.Add(-time.Duration(sparkBuckets-1) * sparkInterval)
	return sparkCache{
		input:  z.NewSparkline(id+"-in", ""),
		output: z.NewSparkline(id+"-out", ""),
		tsIn:   z.NewTimeSeries[int64](start, sparkInterval, sparkBuckets, true),
		tsOut:  z.NewTimeSeries[int64](start, sparkInterval, sparkBuckets, true),
	}
}

func (sc sparkCache) update(metrics []Metrics) {
	// Advance window to now, clear previous counts, then replay all data points.
	now := time.Now()
	sc.tsIn.Touch(now)
	sc.tsOut.Touch(now)
	sc.tsIn.Clear()
	sc.tsOut.Clear()
	for i := range metrics {
		m := &metrics[i]
		sc.tsIn.Add(m.Time, m.Input)
		sc.tsOut.Add(m.Time, m.Output)
	}
	sc.input.SetValues(sc.tsIn.Floats())
	sc.output.SetValues(sc.tsOut.Floats())
}

// kpiCard returns a rounded-border KPI box with a label, Digits widget, and sub-label.
// align controls the horizontal alignment of content ("start" or "end").
func kpiCard(digID, label, initial, fg, sub, align string) c.Option {
	return c.Flex("", "", false, align, 0,
		c.Border("round"),
		c.Padding(0, 2),
		c.Static("", "", label, c.Fg("$gray")),
		c.Digits(digID, "", initial, c.Fg(fg)),
		c.Static("", "", sub, c.Fg("$gray")),
	)
}

// buildUI constructs the mw TUI and wires the store's onChange callback.
func buildUI(theme *z.Theme, store *Store) *z.UI {
	mp  := &metricsProvider{}
	lp  := &logProvider{}
	mtp := &modelProvider{}

	// UI-layer sparkline cache — closed over by makeItemRender.
	sparks := map[string]sparkCache{}

	ui := c.UI(theme,
		c.Flex("root", "", false, "stretch", 0,
			// Title bar
			c.Static("title", "", "mw  ·  OTLP monitor",
				c.Font("bold"),
				c.Bg("$bg1"),
				c.Padding(0, 1),
			),
			// Content grid: left column fixed 42, right column fills
			c.Grid("content", "", []int{-1}, []int{42, -1}, false,
				c.Hint(0, -1),
				c.Cell(0, 0, 1, 1,
					c.Deck("sessions", "", makeItemRender(theme, sparks), 4),
				),
				c.Cell(1, 0, 1, 1,
					c.Switcher("view", "",
						c.Hint(0, -1),

						// Pane 0 — waiting for data
						c.Flex("pane-start", "", false, "center", 0,
							c.Hint(0, -1),
							c.Padding(3, 6),
							c.Static("", "", "Waiting for OTLP data…"),
						),

						// Pane 1 — all-sessions dashboard
						c.Flex("pane-dash", "", false, "stretch", 0,
							c.Hint(0, -1),
							c.Padding(1, 2),
							c.Static("", "", "All Sessions",
								c.Font("bold"),
								c.Fg("$cyan"),
							),
							c.HRule("", "thin"),
							// Aggregate KPI row
							c.Flex("dash-kpi-row", "", true, "start", 1,
								c.Padding(1, 0, 1, 0),
								kpiCard("dig-tot-in", "Input Tokens", "0000K", "$cyan", "total", "start"),
								kpiCard("dig-tot-out", "Output Tokens", "0000K", "$green", "total", "start"),
								kpiCard("dig-tot-cost", "Cost", "000.00", "$yellow", "USD", "end"),
								c.Spacer("", c.Hint(-1, 0)),
							),
							// Session list
							c.Flex("sess-list-box", "", false, "stretch", 0,
								c.Border("round"),
								c.Hint(0, -1),
								c.Static("", "", " Sessions",
									c.Font("bold"),
									c.Fg("$fg0"),
									c.Bg("$bg2"),
								),
								c.Text("dash-text", "", nil, false, 500,
									c.Hint(0, -1),
									c.Padding(0, 1),
								),
							),
						),

						// Pane 2 — session detail
						c.Flex("pane-session", "", false, "stretch", 0,
							c.Hint(0, -1),
							c.Tabs("session-tabs", ""),
							c.Switcher("session-view", "",
								c.Hint(0, -1),

								// Tab 0: Overview
								c.Flex("overview-pane", "", false, "stretch", 0,
									c.Hint(0, -1),
									// KPI row
									c.Flex("session-kpi-row", "", true, "start", 1,
										c.Padding(1, 2, 0, 2),
										kpiCard("dig-in", "Input Tokens", "0000K", "$cyan", "total", "start"),
										kpiCard("dig-out", "Output Tokens", "0000K", "$green", "total", "start"),
										kpiCard("dig-cache", "Cache", "0", "$magenta", "tokens", "start"),
										kpiCard("dig-cost", "Cost", "000.00", "$yellow", "USD", "end"),
										c.Spacer("", c.Hint(-1, 0)),
									),
									// Session info in a bordered box
									c.Flex("session-info-box", "", false, "stretch", 0,
										c.Border("round"),
										c.Hint(0, -1),
										c.Text("overview-text", "", nil, false, 500,
											c.Padding(0, 1),
										),
										c.HRule("", "thin"),
										c.Table("model-table", "", mtp, false,
											c.Hint(0, -1),
										),
									),
								),

								// Tab 1: Metrics
								c.Table("metrics-table", "", mp, false,
									c.Hint(0, -1),
								),

								// Tab 2: Logs
								c.Table("logs-table", "", lp, false,
									c.Hint(0, -1),
								),
							),
						),
					),
				),
			),
			// Footer shortcuts
			c.Shortcuts("footer", "",
				[]string{"↑↓", "navigate", "Tab", "switch tabs", "q", "quit"},
			),
		),
	)

	// ---- Post-build wiring --------------------------------------------------

	deck         := z.Find(ui, "sessions").(*z.Deck)
	view         := z.Find(ui, "view").(*z.Switcher)
	sessionTabs  := z.Find(ui, "session-tabs").(*z.Tabs)
	sessionView  := z.Find(ui, "session-view").(*z.Switcher)
	dashText     := z.Find(ui, "dash-text").(*z.Text)
	digTotIn     := z.Find(ui, "dig-tot-in").(*z.Digits)
	digTotOut    := z.Find(ui, "dig-tot-out").(*z.Digits)
	digTotCost   := z.Find(ui, "dig-tot-cost").(*z.Digits)
	overviewText := z.Find(ui, "overview-text").(*z.Text)
	digIn        := z.Find(ui, "dig-in").(*z.Digits)
	digOut       := z.Find(ui, "dig-out").(*z.Digits)
	digCache     := z.Find(ui, "dig-cache").(*z.Digits)
	digCost      := z.Find(ui, "dig-cost").(*z.Digits)
	metricsTable := z.Find(ui, "metrics-table").(*z.Table)
	logsTable    := z.Find(ui, "logs-table").(*z.Table)
	modelTable   := z.Find(ui, "model-table").(*z.Table)

	sessionTabs.Add("Overview")
	sessionTabs.Add("Metrics")
	sessionTabs.Add("Logs")

	switchSessionTab := func(idx int) { sessionView.Select(idx) }
	sessionTabs.On(z.EvtChange, func(_ z.Widget, _ z.Event, data ...any) bool {
		if len(data) > 0 {
			if idx, ok := data[0].(int); ok {
				switchSessionTab(idx)
			}
		}
		return true
	})
	sessionTabs.On(z.EvtActivate, func(_ z.Widget, _ z.Event, data ...any) bool {
		if len(data) > 0 {
			if idx, ok := data[0].(int); ok {
				switchSessionTab(idx)
			}
		}
		return true
	})

	// updateDash populates the all-sessions dashboard.
	updateDash := func(sessions []*Session) {
		var totIn, totOut int64
		var totCost float64
		for _, s := range sessions {
			for _, t := range s.Totals {
				totIn += t.Input
				totOut += t.Output
				totCost += t.Cost
			}
		}
		digTotIn.Set(formatDigitsTokens(totIn))
		digTotOut.Set(formatDigitsTokens(totOut))
		digTotCost.Set(formatDigitsCost(totCost))

		dashText.Clear()
		if len(sessions) == 0 {
			dashText.Add("  no sessions yet")
			return
		}
		dashText.Add(fmt.Sprintf("  %-36s  %-16s  %s", "Session", "Started", "Models"))
		dashText.Add("  " + strings.Repeat("─", 70))
		for _, s := range sessions {
			name := s.ServiceName
			if name == "" {
				name = truncate(s.ID, 36)
			} else {
				name = truncate(name, 36)
			}
			start := "—"
			if !s.Start.IsZero() {
				start = s.Start.Format("01-02 15:04:05")
			}
			models := make([]string, 0, len(s.Totals))
			for m := range s.Totals {
				models = append(models, truncate(m, 30))
			}
			dashText.Add(fmt.Sprintf("  %-36s  %-16s  %s", name, start, strings.Join(models, ", ")))
		}
	}

	// updateOverview populates the session overview pane.
	updateOverview := func(s *Session) {
		// Aggregate totals across all models.
		var totIn, totOut, totCache int64
		var totCost float64
		for _, t := range s.Totals {
			totIn += t.Input
			totOut += t.Output
			totCache += t.CacheRead + t.CacheCreation
			totCost += t.Cost
		}
		digIn.Set(formatDigitsTokens(totIn))
		digOut.Set(formatDigitsTokens(totOut))
		digCache.Set(formatDigitsTokens(totCache))
		digCost.Set(formatDigitsCost(totCost))

		overviewText.Clear()
		if s.ID != "" {
			overviewText.Add(fmt.Sprintf("  %-16s %s", "ID", truncate(s.ID, 52)))
		}
		if s.ServiceName != "" {
			overviewText.Add(fmt.Sprintf("  %-16s %s", "Service", s.ServiceName))
		}
		if s.ServiceVersion != "" {
			overviewText.Add(fmt.Sprintf("  %-16s %s", "Version", s.ServiceVersion))
		}
		if s.UserEmail != "" {
			overviewText.Add(fmt.Sprintf("  %-16s %s", "User", s.UserEmail))
		}
		if s.OrgID != "" {
			overviewText.Add(fmt.Sprintf("  %-16s %s", "Org", truncate(s.OrgID, 52)))
		}
		if s.TerminalType != "" {
			overviewText.Add(fmt.Sprintf("  %-16s %s", "Terminal", s.TerminalType))
		}
		if s.HostArch != "" {
			overviewText.Add(fmt.Sprintf("  %-16s %s / %s %s",
				"Platform", s.HostArch, s.OSType, s.OSVersion))
		}
		if !s.Start.IsZero() {
			overviewText.Add(fmt.Sprintf("  %-16s %s",
				"Started", s.Start.Format("2006-01-02  15:04:05")))
		}
	}

	// currentItems holds the last-known deck item list for EvtSelect lookup.
	var currentItems []any // [nil, *Session, *Session, …]
	var currentPane = -1  // tracks which outer pane is active to avoid spurious re-selects
	var inRefresh bool    // suppresses EvtSelect during deck.SetItems / deck.Select

	selectPane := func(n int) {
		if currentPane != n {
			currentPane = n
			view.Select(n)
		}
	}

	// updatePane updates the right pane content for the currently selected item.
	// selectPane guards against re-selecting the outer pane when it hasn't changed,
	// preserving the inner session-view tab selection across data refreshes.
	updatePane := func(sel int) {
		if len(currentItems) == 0 || sel < 0 || sel >= len(currentItems) {
			selectPane(0)
			return
		}
		item := currentItems[sel]
		if item == nil {
			selectPane(1)
			updateDash(store.Items())
		} else {
			s := item.(*Session)
			selectPane(2)
			updateOverview(s)
			mp.set(s)
			lp.set(s)
			mtp.set(s)
			metricsTable.Refresh()
			logsTable.Refresh()
			modelTable.Refresh()
		}
	}

	deck.On(z.EvtSelect, func(_ z.Widget, _ z.Event, data ...any) bool {
		if inRefresh {
			return true // ignore events fired by SetItems / Select during refresh
		}
		if len(data) == 0 {
			return false
		}
		idx, ok := data[0].(int)
		if !ok {
			return false
		}
		updatePane(idx)
		ui.Refresh()
		return true
	})

	refresh := func() {
		sessions := store.Items()

		// Keep sparkline cache in sync with current sessions.
		for _, s := range sessions {
			sc, ok := sparks[s.ID]
			if !ok {
				sc = newSparkCache(s.ID)
				sparks[s.ID] = sc
			}
			sc.update(s.Metrics)
		}

		items := make([]any, 1+len(sessions))
		items[0] = nil // "All Sessions" sentinel
		for i, s := range sessions {
			items[i+1] = s
		}

		prev := deck.Selected()

		// Block EvtSelect during SetItems/Select so intermediate events don't
		// corrupt currentPane and trigger a spurious view.Select(2).
		inRefresh = true
		deck.SetItems(items)
		currentItems = items
		if prev > 0 && prev < len(items) {
			deck.Select(prev)
		}
		inRefresh = false

		updatePane(deck.Selected())
		ui.Refresh()
	}

	store.SetOnChange(refresh)
	refresh()
	return ui
}

// makeItemRender returns the Deck ItemRender for session cards.
//
//	Row 0: status dot + session name
//	Row 1: token/cost summary
//	Row 2: "Input " label + input-token sparkline
//	Row 3: "Output" label + output-token sparkline
func makeItemRender(theme *z.Theme, sparks map[string]sparkCache) z.ItemRender {
	return func(r *z.Renderer, x, y, w, h, _ int, data any, selected, focused bool) {
		bg := theme.Color("$bg1")
		if selected {
			bg = theme.Color("$bg3")
		}
		r.Set("", bg, "")
		r.Fill(x, y, w, h, " ")

		indicatorFg := theme.Color("$fg2")
		indicator := " "
		if selected {
			indicator = "▍"
			if focused {
				indicatorFg = theme.Color("$blue")
			}
		}
		r.Set(indicatorFg, bg, "")
		for row := 0; row < h; row++ {
			r.Put(x, y+row, indicator)
		}

		contentX := x + 2

		// Special "All Sessions" item (nil data).
		if data == nil {
			r.Set(theme.Color("$cyan"), bg, "bold")
			r.Text(contentX, y, "◈ All Sessions", w-(contentX-x))
			if h > 1 {
				r.Set(theme.Color("$fg2"), bg, "")
				r.Text(contentX, y+1, "select for totals", w-(contentX-x))
			}
			return
		}

		s := data.(*Session)

		// Row 0: status dot + name
		dotFg, dot := sessionStatusDot(s, theme)
		r.Set(dotFg, bg, "")
		r.Put(contentX, y, dot)

		nameFg, nameFont := theme.Color("$fg1"), ""
		if selected {
			nameFg, nameFont = theme.Color("$fg0"), "bold"
		}
		name := s.ServiceName
		if name == "" {
			name = truncate(s.ID, w-(contentX+2-x))
		}
		r.Set(nameFg, bg, nameFont)
		r.Text(contentX+2, y, name, w-(contentX+2-x))

		// Row 1: aggregate token/cost summary
		if h > 1 {
			var totIn, totOut int64
			var totCost float64
			for _, t := range s.Totals {
				totIn += t.Input
				totOut += t.Output
				totCost += t.Cost
			}
			line1 := fmt.Sprintf("in %s  out %s  $%.4f",
				formatTokens(totIn), formatTokens(totOut), totCost)
			startX := x + w - len([]rune(line1)) - 1
			if startX < contentX {
				startX = contentX
			}
			r.Set(theme.Color("$fg2"), bg, "")
			r.Text(startX, y+1, line1, w-(startX-x))
		}

		// Rows 2–3: sparklines
		renderSparkRow := func(row int, label string, sp *z.Sparkline) {
			if sp == nil || row >= h {
				return
			}
			r.Set(theme.Color("$fg2"), bg, "")
			r.Put(contentX, y+row, label)
			spX := contentX + len([]rune(label))
			spW := w - (spX - x)
			if spW <= 0 {
				return
			}
			sp.Apply(theme)
			spStyle := sp.Style()
			sp.SetStyle("", z.NewStyle().
				WithForeground(spStyle.Foreground()).
				WithBackground(bg))
			sp.SetBounds(spX, y+row, spW, 1)
			sp.Render(r)
			spFg := theme.Color(spStyle.Foreground())
			if emptyCount := spW - len(sp.Values()); emptyCount > 0 {
				r.Set(spFg, bg, "")
				for i := 0; i < emptyCount; i++ {
					r.Put(spX+i, y+row, "▁")
				}
			}
		}

		sc := sparks[s.ID]
		renderSparkRow(2, "Input ", sc.input)
		renderSparkRow(3, "Output", sc.output)
	}
}

// sessionStatusDot returns the (fg colour, glyph) for a session.
func sessionStatusDot(s *Session, theme *z.Theme) (string, string) {
	if s.End.IsZero() {
		return theme.Color("$green"), "●"
	}
	if time.Since(s.End) < 2*time.Minute {
		return theme.Color("$yellow"), "●"
	}
	return theme.Color("$fg2"), "○"
}

// truncate shortens s to at most n runes, appending … if needed.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 0 {
		return ""
	}
	return string(r[:n-1]) + "…"
}

// formatTokens formats a token count compactly: 123, 12.4k, 1.2M.
func formatTokens(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// formatDigitsTokens formats a token count for the Digits widget.
// Up to 9 999 the raw integer is shown (max 4 digits).
// Above that a K / M / B suffix is used so the numeric part stays ≤ 4 digits.
// The result is always left-padded to 5 chars to match the initial "0000K" frame.
func formatDigitsTokens(n int64) string {
	const width = 5 // matches initial "0000K"
	var s string
	switch {
	case n < 10_000:
		s = fmt.Sprintf("%d", n)
	case n < 10_000_000:
		s = fmt.Sprintf("%dK", n/1_000)
	case n < 10_000_000_000:
		s = fmt.Sprintf("%dM", n/1_000_000)
	default:
		s = fmt.Sprintf("%dB", n/1_000_000_000)
	}
	for len(s) < width {
		s = " " + s
	}
	return s
}

// formatDigitsCost formats a USD cost for the Digits widget.
// The result is left-padded with spaces to match the initial frame width ("000.00" = 7 chars),
// so the number always renders flush-right inside the card.
func formatDigitsCost(usd float64) string {
	const width = 6 // matches initial "000.00"
	s := fmt.Sprintf("%.2f", usd)
	for len(s) < width {
		s = " " + s
	}
	return s
}
