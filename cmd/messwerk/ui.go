package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	. "github.com/tekugo/zeichenwerk"
)

const otlpConfig = `{
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "OTEL_METRICS_EXPORTER": "otlp",
    "OTEL_LOGS_EXPORTER": "otlp",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
    "OTEL_METRIC_EXPORT_INTERVAL": "10000",
    "OTEL_LOGS_EXPORT_INTERVAL": "5000"
  }
}`

// sparkCache holds per-session sparkline widgets in the UI layer so no
// Sparkline state leaks into the data model.
type sparkCache struct {
	total  *Sparkline
	input  *Sparkline
	output *Sparkline
}

func newSparkCache(id string) sparkCache {
	return sparkCache{
		total:  NewSparkline(id+"-total", ""),
		input:  NewSparkline(id+"-input", ""),
		output: NewSparkline(id+"-output", ""),
	}
}

func (sc sparkCache) update(buckets []Bucket) {
	sc.total.SetValues(sparkVals(buckets, func(v MetricValues) float64 {
		return float64(v.InputTokens + v.OutputTokens + v.CacheReadTokens + v.CacheCreationTokens)
	}))
	sc.input.SetValues(sparkVals(buckets, func(v MetricValues) float64 {
		return float64(v.InputTokens)
	}))
	sc.output.SetValues(sparkVals(buckets, func(v MetricValues) float64 {
		return float64(v.OutputTokens)
	}))
}

func sparkVals(buckets []Bucket, fn func(MetricValues) float64) []float64 {
	vs := make([]float64, len(buckets))
	for i, b := range buckets {
		vs[i] = fn(b.MetricValues)
	}
	return vs
}

// buildUI constructs the messwerk TUI and wires the store's onChange callback.
func buildUI(theme *Theme, store *Store) *UI {
	b := NewBuilder(theme)
	configLines := strings.Split(otlpConfig, "\n")

	// UI-layer sparkline cache — declared before the builder so makeItemRender
	// can close over it. The map is mutated in refresh(); map reference
	// semantics mean the render closure always sees current entries.
	sparks := map[string]sparkCache{}

	b.Flex("main", false, "stretch", 0).
		Tabs("tabs", "Start", "Monitor", "Log", "Sessions", "About").
		Switcher("view", false).Hint(0, -1).
		With(func(b *Builder) {
			// ── Start pane ────────────────────────────────────────────────────
			b.Flex("start-pane", false, "stretch", 0).Hint(0, -1).Padding(3, 6).
				Static("info-heading", "Waiting for telemetry data").
				Static("info-desc", "Add this to ~/.claude/settings.json to enable OTLP export to messwerk:").
				Text("info-config", configLines, false, 0).Hint(0, len(configLines)).
				Button("copy-btn", "Copy config to clipboard").
			End()
		}).
		With(func(b *Builder) {
			// ── Monitor pane ──────────────────────────────────────────────────
			b.Grid("nav-grid", 1, 2, false).Hint(0, -1).
				Rows(-1).
				Columns(42, -1).
				// Left column: session list
				Cell(0, 0, 1, 1).Deck("nav", makeItemRender(theme, sparks), 4).
				// Right column: session dashboard
				Cell(1, 0, 1, 1).Flex("detail", false, "stretch", 0).
				// Header bar
				Flex("detail-hdr", true, "center", 2).Background("$bg1").Padding(0, 1).
				Static("detail-icon", "◈").Font("bold").Foreground("$cyan").Padding(0, 1, 0, 0).
				Static("detail-name", "Select a session").Font("bold").Foreground("$fg0").
				Spacer().Hint(-1, 0).
				Static("detail-status", "").Foreground("$gray").
				End().
				HRule("thin").
				// KPI cards — big-number current-window metrics
				Flex("kpi-row", true, "start", 1).Padding(0, 0, 1, 0).
				Flex("kpi-in", false, "start", 0).Border("", "round").Padding(0, 2).
				Static("kpi-in-lbl", "Input Tokens").Foreground("$gray").
				Digits("kpi-in-val", "0000").Foreground("$cyan").
				Static("kpi-in-sub", "per window").Foreground("$gray").
				End().
				Flex("kpi-out", false, "start", 0).Border("", "round").Padding(0, 2).
				Static("kpi-out-lbl", "Output Tokens").Foreground("$gray").
				Digits("kpi-out-val", "0000").Foreground("$green").
				Static("kpi-out-sub", "per window").Foreground("$gray").
				End().
				Flex("kpi-cost", false, "start", 0).Border("", "round").Padding(0, 2).
				Static("kpi-cost-lbl", "Cost").Foreground("$gray").
				Digits("kpi-cost-val", "0.00").Foreground("$yellow").
				Static("kpi-cost-sub", "USD / window").Foreground("$gray").
				End().
				Spacer().Hint(-1, 0).
				End().
				// Sparkline pane
				Flex("spark-pane", false, "stretch", 0).Border("", "round").Hint(0, 8).
				Static("spark-title", " Token Activity  ·  2 min buckets").Font("bold").Foreground("$fg0").Background("$bg2").
				Sparkline("detail-spark").Hint(0, 5).
				End().
				// Session details pane
				Flex("info-pane", false, "stretch", 0).Border("", "round").Hint(0, -1).
				Static("info-title", " Session Details").Font("bold").Foreground("$fg0").Background("$bg2").
				Text("detail-info", nil, false, 100).Padding(0, 1).Hint(0, -1).
				End().
				End(). // Flex("detail")
				End()  // Grid("nav-grid")
		}).
		With(func(b *Builder) {
			// ── Log pane ──────────────────────────────────────────────────────
			b.Flex("log-pane", false, "stretch", 0).Hint(0, -1).
				Table("otlp-log", store.Log, false).Hint(0, -1).
			End()
		}).
		With(func(b *Builder) {
			// ── Sessions pane — aggregate view + session list ─────────────────
			b.Flex("sessions-pane", false, "stretch", 0).Hint(0, -1).Padding(1, 2).
				Flex("total-hdr", true, "center", 2).Padding(0, 0, 1, 0).
				Static("total-title", "All Sessions").Font("bold").Foreground("$cyan").
				Spacer().Hint(-1, 0).
				Static("total-subtitle", "Aggregated totals").Foreground("$gray").
				End().
				HRule("thin").Padding(0, 0, 1, 0).
				// Aggregate KPI row
				Flex("total-kpi-row", true, "start", 1).Padding(0, 0, 1, 0).
				Flex("tot-in", false, "start", 0).Border("", "round").Padding(0, 2).
				Static("tot-in-lbl", "Input Tokens").Foreground("$gray").
				Static("val-tot-input", "—").Font("bold").Foreground("$cyan").
				End().
				Flex("tot-out", false, "start", 0).Border("", "round").Padding(0, 2).
				Static("tot-out-lbl", "Output Tokens").Foreground("$gray").
				Static("val-tot-output", "—").Font("bold").Foreground("$green").
				End().
				Flex("tot-cst", false, "start", 0).Border("", "round").Padding(0, 2).
				Static("tot-cost-lbl", "Total Cost").Foreground("$gray").
				Static("val-tot-cost", "—").Font("bold").Foreground("$yellow").
				End().
				Flex("tot-act", false, "start", 0).Border("", "round").Padding(0, 2).
				Static("tot-act-lbl", "Active Time").Foreground("$gray").
				Static("val-tot-active", "—").Font("bold").Foreground("$magenta").
				End().
				Spacer().Hint(-1, 0).
				End().
				// Session list
				Flex("sess-list-pane", false, "stretch", 0).Border("", "round").Hint(0, -1).
				Static("sess-list-title", " Sessions").Font("bold").Foreground("$fg0").Background("$bg2").
				Text("session-list", nil, false, 500).Padding(0, 1).Hint(0, -1).
				End().
			End()
		}).
		With(func(b *Builder) {
			// ── About pane ────────────────────────────────────────────────────
			b.Flex("about-pane", false, "stretch", 0).Hint(0, -1).Padding(3, 6).
				Static("about-placeholder", "About — coming soon").
			End()
		}).
		End().
		Shortcuts("footer", "↑↓", "navigate", "←→", "switch tabs", "q", "quit").
		End()

	ui := b.Build()

	tabs    := Find(ui, "tabs").(*Tabs)
	viewSw  := Find(ui, "view").(*Switcher)
	deck    := Find(ui, "nav").(*Deck)
	copyBtn := Find(ui, "copy-btn").(*Button)

	// Monitor detail widgets
	detailName   := Find(ui, "detail-name").(*Static)
	detailStatus := Find(ui, "detail-status").(*Static)
	digInput     := Find(ui, "kpi-in-val").(*Digits)
	digOutput    := Find(ui, "kpi-out-val").(*Digits)
	digCost      := Find(ui, "kpi-cost-val").(*Digits)
	detailSpark  := Find(ui, "detail-spark").(*Sparkline)
	detailInfo   := Find(ui, "detail-info").(*Text)
	detailSpark.SetMode(Absolute)
	detailSpark.SetMin(0)

	// Sessions tab widgets
	valTotInput  := Find(ui, "val-tot-input").(*Static)
	valTotOutput := Find(ui, "val-tot-output").(*Static)
	valTotCost   := Find(ui, "val-tot-cost").(*Static)
	valTotActive := Find(ui, "val-tot-active").(*Static)
	sessionList  := Find(ui, "session-list").(*Text)

	var currentTab int

	switchToTab := func(idx int) {
		currentTab = idx
		viewSw.Select(idx)
	}
	tabs.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			if idx, ok := data[0].(int); ok {
				switchToTab(idx)
			}
		}
		return true
	})
	tabs.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if len(data) > 0 {
			if idx, ok := data[0].(int); ok {
				switchToTab(idx)
			}
		}
		return true
	})

	copyBtn.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if err := copyToClipboard(otlpConfig); err == nil {
			copyBtn.Set("✓ Copied!")
			Redraw(copyBtn)
			go func() {
				time.Sleep(2 * time.Second)
				copyBtn.Set("Copy config to clipboard")
				Redraw(copyBtn)
			}()
		}
		return true
	})

	var currentItems []*SessionItem

	updateDetail := func(item *SessionItem) {
		// Header
		name := item.Name
		if name == "" {
			name = truncate(item.ID, 36)
		}
		detailName.Set(name)

		_, dot := statusDot(item.Status, theme)
		statusStr := dot + " "
		switch item.Status {
		case StatusActive:
			statusStr += "Active"
		case StatusIdle:
			statusStr += "Idle"
		default:
			statusStr += "Ended"
		}
		if !item.LastSeen.IsZero() {
			statusStr += "  ·  " + item.LastSeen.Format("15:04:05")
		}
		detailStatus.Set(statusStr)

		// KPI big numbers — last metric window
		last := item.Last
		digInput.Set(formatDigitsTokens(last.InputTokens))
		digOutput.Set(formatDigitsTokens(last.OutputTokens))
		digCost.Set(formatDigitsCost(last.CostUSD))

		// Sparkline — total tokens from 2-min buckets
		if sc := sparks[item.ID]; sc.total != nil {
			vals := sc.total.Values()
			maxV := 1.0
			for _, v := range vals {
				if v > maxV {
					maxV = v
				}
			}
			detailSpark.SetMax(maxV)
			detailSpark.SetValues(vals)
		}

		// Session details text
		t := item.Total
		detailInfo.Clear()
		if item.ID != "" {
			detailInfo.Add(fmt.Sprintf("%-12s %s", "Session", truncate(item.ID, 40)))
		}
		if item.Name != "" && item.Name != item.ID {
			detailInfo.Add(fmt.Sprintf("%-12s %s", "Name", item.Name))
		}
		if item.OrgID != "" {
			detailInfo.Add(fmt.Sprintf("%-12s %s", "Org", truncate(item.OrgID, 40)))
		}
		if item.UserEmail != "" {
			detailInfo.Add(fmt.Sprintf("%-12s %s", "User", item.UserEmail))
		}
		if item.TerminalType != "" {
			detailInfo.Add(fmt.Sprintf("%-12s %s", "Terminal", item.TerminalType))
		}
		if !item.FirstSeen.IsZero() {
			detailInfo.Add(fmt.Sprintf("%-12s %s", "Started", item.FirstSeen.Format("2006-01-02  15:04:05")))
		}
		if !item.LastSeen.IsZero() {
			detailInfo.Add(fmt.Sprintf("%-12s %s", "Last seen", item.LastSeen.Format("2006-01-02  15:04:05")))
		}
		detailInfo.Add("")
		detailInfo.Add("── Totals ─────────────────────────────────────────")
		detailInfo.Add(fmt.Sprintf("%-20s %s tokens", "Input", formatTokens(t.InputTokens)))
		detailInfo.Add(fmt.Sprintf("%-20s %s tokens", "Output", formatTokens(t.OutputTokens)))
		detailInfo.Add(fmt.Sprintf("%-20s %s tokens", "Cache read", formatTokens(t.CacheReadTokens)))
		detailInfo.Add(fmt.Sprintf("%-20s %s tokens", "Cache creation", formatTokens(t.CacheCreationTokens)))
		detailInfo.Add(fmt.Sprintf("%-20s $%.6f", "Cost", t.CostUSD))
		if t.ActiveTimeUser > 0 || t.ActiveTimeCLI > 0 {
			detailInfo.Add(fmt.Sprintf("%-20s %.1fs user  ·  %.1fs CLI", "Active time", t.ActiveTimeUser, t.ActiveTimeCLI))
		}
		if t.LinesAdded > 0 || t.LinesRemoved > 0 {
			detailInfo.Add(fmt.Sprintf("%-20s +%d  −%d", "Lines changed", t.LinesAdded, t.LinesRemoved))
		}
		if t.EditDecisionsAccepted > 0 || t.EditDecisionsRejected > 0 {
			detailInfo.Add(fmt.Sprintf("%-20s %d accepted  ·  %d rejected", "Edit decisions", t.EditDecisionsAccepted, t.EditDecisionsRejected))
		}
	}

	updateTotalView := func(tv TotalView) {
		t := tv.Total
		valTotInput.Set(formatTokens(t.InputTokens))
		valTotOutput.Set(formatTokens(t.OutputTokens))
		valTotCost.Set(fmt.Sprintf("$%.4f", t.CostUSD))
		active := t.ActiveTimeUser + t.ActiveTimeCLI
		valTotActive.Set(fmt.Sprintf("%.0fs", active))

		sessionList.Clear()
		sessionList.Add(fmt.Sprintf("  %-36s  %-19s  %-19s  %s",
			"Session", "Started", "Last seen", "Status"))
		sessionList.Add("  " + strings.Repeat("─", 85))
		for _, s := range tv.Sessions {
			_, dot := statusDot(s.Status, theme)
			name := s.Name
			if name == "" || name == s.ID {
				name = truncate(s.ID, 36)
			} else {
				name = truncate(name+" ("+truncate(s.ID, 8)+"…)", 36)
			}
			start, end := "—", "—"
			if !s.FirstSeen.IsZero() {
				start = s.FirstSeen.Format("2006-01-02 15:04")
			}
			if !s.LastSeen.IsZero() {
				end = s.LastSeen.Format("2006-01-02 15:04")
			}
			sessionList.Add(fmt.Sprintf("%s %-36s  %-19s  %-19s", dot, name, start, end))
		}
	}

	refresh := func() {
		items, tv := store.Items()
		currentItems = items

		// Keep sparkline cache in sync with current sessions.
		for _, item := range items {
			sc, ok := sparks[item.ID]
			if !ok {
				sc = newSparkCache(item.ID)
				sparks[item.ID] = sc
			}
			sc.update(item.Buckets)
		}

		// Auto-switch from Start to Monitor on first session.
		if currentTab == 0 && len(items) > 0 {
			tabs.Select(1)
		}

		any := make([]any, len(items))
		for i, it := range items {
			any[i] = it
		}
		prevIdx := deck.Selected()
		deck.SetItems(any)
		if prevIdx > 0 && prevIdx < len(items) {
			deck.Select(prevIdx)
		} else if len(items) > 0 {
			updateDetail(items[deck.Selected()])
		}

		updateTotalView(tv)
		ui.Refresh()
	}

	deck.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		idx, ok := data[0].(int)
		if !ok || idx < 0 || idx >= len(currentItems) {
			return false
		}
		updateDetail(currentItems[idx])
		return true
	})

	store.SetOnChange(refresh)
	refresh()
	return ui
}

// copyToClipboard writes text to the system clipboard.
func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	default:
		cmd = exec.Command("xclip", "-selection", "clipboard")
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// makeItemRender returns the Deck ItemRender for session cards.
// Sparklines are looked up from the UI-layer cache, not the model.
//
//	Row 0: status dot + session name
//	Row 1: total input, output counts + cost (right-aligned)
//	Row 2: "Input " label + input-token sparkline
//	Row 3: "Output" label + output-token sparkline
func makeItemRender(theme *Theme, sparks map[string]sparkCache) ItemRender {
	return func(r *Renderer, x, y, w, h, _ int, data any, selected, focused bool) {
		item := data.(*SessionItem)

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
		dotFg, dot := statusDot(item.Status, theme)
		r.Set(dotFg, bg, "")
		r.Put(contentX, y, dot)

		nameFg, nameFont := theme.Color("$fg1"), ""
		if selected {
			nameFg, nameFont = theme.Color("$fg0"), "bold"
		}
		name := item.Name
		if name == "" {
			name = truncate(item.ID, w-(contentX+2-x))
		}
		r.Set(nameFg, bg, nameFont)
		r.Text(contentX+2, y, name, w-(contentX+2-x))

		if h > 1 {
			t := item.Total
			line1 := fmt.Sprintf("in %s  out %s  $%.2f",
				formatTokens(t.InputTokens), formatTokens(t.OutputTokens), t.CostUSD)
			startX := x + w - len(line1) - 1
			if startX < contentX {
				startX = contentX
			}
			r.Set(theme.Color("$fg2"), bg, "")
			r.Text(startX, y+1, line1, w-(startX-x))
		}

		renderSparkRow := func(row int, label string, sp *Sparkline) {
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
			spFg := theme.Color(spStyle.Foreground())
			sp.SetStyle("", NewStyle().WithForeground(spStyle.Foreground()).WithBackground(bg))
			sp.SetBounds(spX, y+row, spW, 1)
			sp.Render(r)
			if emptyCount := spW - len(sp.Values()); emptyCount > 0 {
				r.Set(spFg, bg, "")
				for i := 0; i < emptyCount; i++ {
					r.Put(spX+i, y+row, "▁")
				}
			}
		}

		sc := sparks[item.ID]
		renderSparkRow(2, "Input ", sc.input)
		renderSparkRow(3, "Output", sc.output)
	}
}

// statusDot returns the (fg colour, glyph) for the given status.
func statusDot(st SessionStatus, theme *Theme) (string, string) {
	switch st {
	case StatusActive:
		return theme.Color("$green"), "●"
	case StatusIdle:
		return theme.Color("$yellow"), "●"
	default:
		return theme.Color("$fg2"), "○"
	}
}

// truncate shortens s to at most n runes, appending … if needed.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
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
// Digits supports 0-9, '.', ':', ',', ' ' only — no k/M suffixes.
// Values ≥ 10 000 are shown as "X.X" (thousands); the sub-label carries the unit.
func formatDigitsTokens(n int64) string {
	if n < 10_000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%.1f", float64(n)/1_000.0)
}

// formatDigitsCost formats a USD cost for the Digits widget (2 decimal places).
func formatDigitsCost(usd float64) string {
	if usd < 10 {
		return fmt.Sprintf("%.2f", usd)
	}
	return fmt.Sprintf("%.1f", usd)
}
