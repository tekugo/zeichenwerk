package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	. "github.com/tekugo/zeichenwerk"
)

type navItem struct{ icon, name, desc string }

func parseFlags() (*Theme, bool, bool, bool) {
	t := flag.String("t", "midnight", "Theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	dbg := flag.Bool("debug", false, "Start in debug mode")
	dmp := flag.Bool("dump", false, "Dump widget hierarchy to stdout and exit")
	dmpV := flag.Bool("dump-verbose", false, "Dump widget hierarchy with style details to stdout and exit")
	flag.Parse()
	var theme *Theme
	switch *t {
	case "tokyo":
		theme = TokyoNightTheme()
	case "nord":
		theme = NordTheme()
	case "gruvbox-dark":
		theme = GruvboxDarkTheme()
	case "gruvbox-light":
		theme = GruvboxLightTheme()
	case "lipstick":
		theme = LipstickTheme()
	default:
		theme = MidnightNeonTheme()
	}
	return theme, *dbg, *dmp, *dmpV
}

func main() {
	theme, dbg, dmp, dmpV := parseFlags()
	ui := createUI(theme)
	if dmp || dmpV {
		ui.SetBounds(0, 0, 120, 40)
		ui.Layout()
		ui.Dump(os.Stdout, DumpOptions{Style: dmpV})
		return
	}
	if dbg {
		ui.Debug()
	}
	ui.Run()
}

// ── Shell ──────────────────────────────────────────────────────────────────────

func createUI(theme *Theme) *UI {
	navItems := []any{
		navItem{"◈", "Dashboard", "System metrics & KPIs"},
		navItem{"◉", "User Admin", "Manage users & roles"},
		navItem{"≡", "Log Monitor", "Live log tail & stats"},
		navItem{"⬡", "Processes", "Running process list"},
		navItem{"◧", "Data Entry", "Forms & input controls"},
		navItem{"✎", "Code Editor", "Edit files & preview"},
	}

	renderNav := func(r *Renderer, x, y, w, h, index int, data any, selected, focused bool) {
		item := data.(navItem)
		bg := "$bg1"
		if selected {
			itemBg := bg
			if focused {
				itemBg = "$bg2"
			}
			r.Set("$cyan", itemBg, "bold")
			r.Fill(x, y, w, 1, " ")
			r.Put(x, y, "┃")
			r.Text(x+2, y, item.icon+"  "+item.name, w-2)
			r.Set("$cyan", itemBg, "")
			r.Fill(x, y+1, w, 1, " ")
			r.Put(x, y+1, "┃")
			r.Text(x+2, y+1, item.desc, w-2)
		} else {
			r.Set("$fg1", bg, "")
			r.Fill(x, y, w, 1, " ")
			r.Text(x+2, y, item.icon+"  "+item.name, w-2)
			r.Set("$gray", bg, "")
			r.Fill(x, y+1, w, 1, " ")
			r.Text(x+2, y+1, item.desc, w-2)
		}
		r.Set("$fg3", bg, "")
		r.Fill(x, y+2, w, 1, " ")
	}

	ui := NewBuilder(theme).
		Flex("root", false, "stretch", 0).
		// ── Header ────────────────────────────────────────────────────────────
		Flex("header", true, "center", 0).Background("$bg1").Padding(0, 1).
		Static("app-icon", "◈").Font("bold").Foreground("$cyan").Padding(0, 1, 0, 0).
		Static("app-name", "TUI Showcase").Font("bold").Foreground("$fg0").
		Spacer().Hint(2, 0).
		VRule("thin").
		Spacer().Hint(2, 0).
		Static("app-tagline", "Real-world terminal UI use cases").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("live-indicator", "● LIVE").Foreground("$green").Font("bold").Padding(0, 2, 0, 0).
		Static("header-sep", " | ").Foreground("$gray").
		Static("header-theme-lbl", "Theme ").Foreground("$gray").
		Select("theme-select", "neon", "Midnight Neon", "tokyo", "Tokyo Night", "gruvbox-dark", "Gruvbox Dark", "nrrd", "Nord").Padding(0, 1, 0, 0).
		End(). // Flex("header")
		// ── Body ──────────────────────────────────────────────────────────────
		Grid("body", 1, 2, false).Hint(0, -1).Columns(26, -1).
		Cell(0, 0, 1, 1).
		Flex("sidebar", false, "stretch", 0).Background("$bg1").
		Static("sidebar-brand", " ◈ SHOWCASE").Font("bold").Foreground("$magenta").Padding(1, 0).
		HRule("thin").
		Deck("nav", renderNav, 3).Hint(0, -1).
		HRule("thin").
		Static("sidebar-key1", "  ↑↓  Navigate").Foreground("$gray").Padding(0, 0, 0, 0).
		Static("sidebar-key2", "  ↵   Select").Foreground("$gray").
		Static("sidebar-key3", "  Tab Focus").Foreground("$gray").Padding(0, 0, 1, 0).
		End(). // Flex("sidebar")
		Cell(1, 0, 1, 1).
		Switcher("content", false).
		With(dashboardScreen).
		With(userAdminScreen).
		With(logMonitorScreen).
		With(processScreen).
		With(dataEntryScreen).
		With(codeEditorScreen).
		End(). // Switcher("content")
		End(). // Grid("body")
		// ── Footer ────────────────────────────────────────────────────────────
		Flex("footer", true, "center", 0).Background("$bg1").Padding(0, 1).
		Static("footer-keys", " ↑↓ Navigate   Tab/Shift+Tab Focus   Enter/Space Activate   Esc Cancel").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("footer-brand", "Zeichenwerk v2.0 ◈").Foreground("$gray").
		End(). // Flex("footer")
		Build()

	// Wire navigation
	switcher := Find(ui, "content").(*Switcher)
	nav := Find(ui, "nav").(*Deck)
	nav.SetItems(navItems)
	navSwitch := func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 1 {
			if sel, ok := data[0].(int); ok {
				switcher.Select(sel)
			}
		}
		return true
	}
	nav.On(EvtSelect, navSwitch)
	nav.On(EvtActivate, navSwitch)

	// Wire theme switcher
	themes := map[string]*Theme{
		"neon":         MidnightNeonTheme(),
		"tokyo":        TokyoNightTheme(),
		"gruvbox-dark": GruvboxDarkTheme(),
		"nrrd":         NordTheme(),
	}
	Find(ui, "theme-select").On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 1 {
			if key, ok := data[0].(string); ok {
				if theme, found := themes[key]; found {
					ui.SetTheme(theme)
				}
			}
		}
		return true
	})

	return ui
}

// ── Screen 1: System Dashboard ─────────────────────────────────────────────────

func dashboardScreen(b *Builder) {
	// Pre-configure progress bars
	cpuProg := NewProgress("dash-cpu-prog", "", true)
	cpuProg.SetTotal(100)
	cpuProg.SetValue(64)

	memProg := NewProgress("dash-mem-prog", "", true)
	memProg.SetTotal(100)
	memProg.SetValue(41)

	diskProg := NewProgress("dash-disk-prog", "", true)
	diskProg.SetTotal(100)
	diskProg.SetValue(78)

	netProg := NewProgress("dash-net-prog", "", true)
	netProg.SetTotal(100)
	netProg.SetValue(23)

	// Service status table
	svcHeaders := []string{"Service", "Status", "CPU", "Memory", "Uptime"}
	svcData := [][]string{
		{"nginx", "● running", "0.3%", " 24 MB", "14d 06h"},
		{"postgresql", "● running", "2.1%", "512 MB", "14d 06h"},
		{"redis", "● running", "0.1%", " 64 MB", "14d 05h"},
		{"celery", "○ stopped", "  —  ", "    —  ", "     —  "},
		{"prometheus", "● running", "1.8%", "128 MB", " 2d 11h"},
		{"grafana", "● running", "0.9%", " 96 MB", " 2d 11h"},
	}
	svcTable := NewArrayTableProvider(svcHeaders, svcData)

	// Activity log seed entries
	activityLines := []string{
		"[12:34:01] INFO    User admin logged in from 192.168.1.10",
		"[12:33:55] WARN    High memory usage on worker-03 (87%)",
		"[12:33:12] INFO    Backup job completed: 4.2 GB archived",
		"[12:32:47] INFO    SSL certificate renewed for api.example.com",
		"[12:31:30] ERROR   Connection timeout to external API (retry 3/3)",
		"[12:31:28] WARN    Rate limit reached for endpoint /api/v2/search",
		"[12:30:55] INFO    Deployment finished: app-server v2.4.1",
		"[12:30:10] INFO    Scheduled task 'db-cleanup' started",
		"[12:29:44] INFO    New user registered: jane.doe@example.com",
		"[12:29:01] DEBUG   Cache invalidated: product catalog (1,248 keys)",
	}

	b.Flex("dashboard", false, "stretch", 0).Padding(1, 2).
		// Title row
		Flex("dash-hdr", true, "center", 2).Padding(0, 0, 1, 0).
		Static("dash-title", "System Dashboard").Font("bold").Foreground("$cyan").
		Spacer().Hint(-1, 0).
		Static("dash-date", time.Now().Format("Mon, 02 Jan 2006")).Foreground("$gray").
		Static("dash-sep", "  ·  ").Foreground("$gray").
		Static("dash-time", time.Now().Format("15:04 MST")).Foreground("$fg1").Font("bold").
		End(). // Flex("dash-hdr")
		HRule("thin").Padding(0, 0, 1, 0).
		// ── KPI cards ────────────────────────────────────────────────────────
		Flex("kpi-row", true, "stretch", 2).Padding(0, 0, 1, 0).
		// CPU card
		Flex("kpi-cpu", false, "start", 0).Border("", "round").Padding(1, 2).
		Static("kpi-cpu-lbl", "CPU Usage").Foreground("$gray").
		Digits("kpi-cpu-val", "64").Foreground("$yellow").
		Add(cpuProg).Hint(0, 1).
		Static("kpi-cpu-sub", "% · 8 cores · 3.2 GHz").Foreground("$gray").
		End(). // Flex("kpi-cpu")
		// Memory card
		Flex("kpi-mem", false, "start", 0).Border("", "round").Padding(1, 2).
		Static("kpi-mem-lbl", "Memory").Foreground("$gray").
		Digits("kpi-mem-val", "4.1").Foreground("$cyan").
		Add(memProg).Hint(0, 1).
		Static("kpi-mem-sub", "GB of 10 GB total").Foreground("$gray").
		End(). // Flex("kpi-mem")
		// Disk card
		Flex("kpi-disk", false, "start", 0).Border("", "round").Padding(1, 2).
		Static("kpi-disk-lbl", "Disk").Foreground("$gray").
		Digits("kpi-disk-val", "780").Foreground("$orange").
		Add(diskProg).Hint(0, 1).
		Static("kpi-disk-sub", "GB of 1 TB · 78% full").Foreground("$gray").
		End(). // Flex("kpi-disk")
		// Network card
		Flex("kpi-net", false, "start", 0).Border("", "round").Padding(1, 2).
		Static("kpi-net-lbl", "Network").Foreground("$gray").
		Digits("kpi-net-val", "2.4").Foreground("$green").
		Add(netProg).Hint(0, 1).
		Static("kpi-net-sub", "Mb/s ↑ · ↓8.1 Mb/s recv").Foreground("$gray").
		End(). // Flex("kpi-net")
		End(). // Flex("kpi-row")
		// ── Services + Alerts ────────────────────────────────────────────────
		Grid("dash-mid", 1, 2, false).Hint(0, 10).Columns(-2, -1).Padding(0, 0, 1, 0).Border("none").
		Cell(0, 0, 1, 1).
		Flex("svc-pane", false, "stretch", 0).Border("", "round").
		Static("svc-title", " Services").Font("bold").Foreground("$fg0").Background("$bg2").
		Table("svc-table", svcTable, false).Hint(0, -1).
		End(). // Flex("svc-pane")
		Cell(1, 0, 1, 1).
		Flex("alerts-pane", false, "stretch", 0).Border("", "round").
		Static("alerts-title", " Alerts & Notices").Font("bold").Foreground("$fg0").Background("$bg2").
		Static("alert1", "⚠  Disk usage on /var exceeds 80%").Foreground("$orange").Padding(0, 1).
		Static("alert2", "⚠  worker-03 memory pressure").Foreground("$orange").Padding(0, 1).
		Static("alert3", "✓  All SSL certificates valid").Foreground("$green").Padding(0, 1).
		Static("alert4", "✓  Last backup: 4h ago (success)").Foreground("$green").Padding(0, 1).
		Static("alert5", "✓  All database replicas in sync").Foreground("$green").Padding(0, 1).
		Static("alert6", "ℹ  Maintenance window: Sun 02:00 UTC").Foreground("$cyan").Padding(0, 1).
		End(). // Flex("alerts-pane")
		End(). // Grid("dash-mid")
		// ── Activity log ─────────────────────────────────────────────────────
		Flex("activity-pane", false, "stretch", 0).Border("", "round").Hint(0, -1).
		Static("activity-title", " Recent Activity").Font("bold").Foreground("$fg0").Background("$bg2").
		Text("activity-log", activityLines, false, 200).Hint(0, -1).
		End(). // Flex("activity-pane")
		End()  // Flex("dashboard")

	// Animate spinners when visible
	container := b.Find("dashboard").(Container)
	container.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		for _, sp := range FindAll[*Spinner](container) {
			sp.Start(120 * time.Millisecond)
		}
		return true
	})
	container.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		for _, sp := range FindAll[*Spinner](container) {
			sp.Stop()
		}
		return true
	})
}

// ── Screen 2: User Administration ──────────────────────────────────────────────

func userAdminScreen(b *Builder) {
	userHeaders := []string{"ID", "Name", "Email", "Role", "Status", "Last Login"}
	userData := [][]string{
		{"001", "Alice Johnson", "alice@example.com", "Admin", "● Active", "2026-03-31 11:42"},
		{"002", "Bob Martinez", "bob@example.com", "Editor", "● Active", "2026-03-31 09:15"},
		{"003", "Carol White", "carol@example.com", "Viewer", "● Active", "2026-03-30 17:03"},
		{"004", "David Kim", "david@example.com", "Editor", "○ Inactive", "2026-03-28 08:55"},
		{"005", "Eva Müller", "eva@example.com", "Admin", "● Active", "2026-03-31 10:30"},
		{"006", "Frank Chen", "frank@example.com", "Viewer", "● Active", "2026-03-29 14:22"},
		{"007", "Grace Okafor", "grace@example.com", "Editor", "⊘ Pending", "          —     "},
		{"008", "Henry Dubois", "henry@example.com", "Viewer", "○ Inactive", "2026-03-15 09:00"},
		{"009", "Iris Nakamura", "iris@example.com", "Editor", "● Active", "2026-03-31 07:48"},
		{"010", "Jack O'Brien", "jack@example.com", "Viewer", "● Active", "2026-03-30 22:11"},
		{"011", "Karen Singh", "karen@example.com", "Admin", "● Active", "2026-03-31 11:01"},
		{"012", "Leon Petrov", "leon@example.com", "Editor", "⊘ Pending", "          —     "},
	}

	selectedUser := struct {
		ID         string `readonly:"true"`
		Name       string `width:"30"`
		Email      string `label:"E-Mail" width:"30"`
		Role       string `control:"select" options:"admin,Admin,editor,Editor,viewer,Viewer"`
		Department string `width:"30"`
		Active     bool   `label:"Active"`
		Admin      bool   `label:"Administrator"`
	}{
		ID: "001", Name: "Alice Johnson", Email: "alice@example.com",
		Role: "admin", Department: "Engineering", Active: true, Admin: true,
	}

	b.Flex("user-admin", false, "stretch", 0).Padding(1, 2).
		// Title
		Flex("ua-hdr", true, "center", 2).Padding(0, 0, 1, 0).
		Static("ua-title", "User Administration").Font("bold").Foreground("$cyan").
		Spacer().Hint(-1, 0).
		Static("ua-count", "12 users  ·  9 active  ·  3 pending/inactive").Foreground("$gray").
		End(). // Flex("ua-hdr")
		HRule("thin").Padding(0, 0, 1, 0).
		// Toolbar
		Flex("ua-toolbar", true, "center", 2).Padding(0, 0, 1, 0).
		Static("ua-search-lbl", "Search:").Foreground("$gray").
		Typeahead("ua-search", "", "name, email or role…").Hint(28, 1).
		Spacer().Hint(-1, 0).
		Button("ua-btn-new", " + New User").
		Button("ua-btn-del", " ✕ Delete").
		Button("ua-btn-exp", " ↓ Export").
		End(). // Flex("ua-toolbar")
		// Split: table left, detail right
		Grid("ua-body", 1, 2, false).Hint(0, -1).Columns(-3, -2).Border("none").
		Cell(0, 0, 1, 1).
		Flex("ua-list-pane", false, "stretch", 0).Border("", "round").
		Static("ua-list-title", " Users").Font("bold").Background("$bg2").
		Table("ua-table", NewArrayTableProvider(userHeaders, userData), true).Hint(0, -1).
		End(). // Flex("ua-list-pane")
		Cell(1, 0, 1, 1).
		Flex("ua-detail-pane", false, "stretch", 0).Border("", "round").
		Static("ua-detail-title", " Edit User").Font("bold").Background("$bg2").
		Form("ua-form", "", &selectedUser).
		Group("ua-group", "", "", false, 1).Padding(1).
		End(). // Group("ua-group")
		End(). // Form("ua-form")
		Flex("ua-detail-btns", true, "end", 2).Padding(1).
		Button("ua-save", " ✓ Save Changes").
		Button("ua-reset", " ↺ Reset").
		Button("ua-deactivate", " ⊘ Deactivate").
		End(). // Flex("ua-detail-btns")
		End(). // Flex("ua-detail-pane")
		End(). // Grid("ua-body")
		End()  // Flex("user-admin")

	container := b.Find("user-admin").(Container)
	b.Find("ua-save").On(EvtActivate, func(w Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "ua-detail-title").(*Static); ok {
			lbl.SetText(fmt.Sprintf(" Edit User  ✓ Saved %s", time.Now().Format("15:04:05")))
		}
		return true
	})
	b.Find("ua-reset").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "ua-detail-title").(*Static); ok {
			lbl.SetText(" Edit User")
		}
		return true
	})
}

// ── Screen 3: Log Monitor ──────────────────────────────────────────────────────

// logLevels and sample sources for generating entries
var (
	logLevels  = []string{"DEBUG", "INFO ", "INFO ", "INFO ", "WARN ", "WARN ", "ERROR"}
	logSources = []string{
		"api-server", "worker-01", "worker-02", "db-pool", "cache",
		"scheduler", "mailer", "auth", "storage", "metrics",
	}
)

var logMessages = []string{
	"Request processed in 42ms GET /api/v2/users",
	"Request processed in 128ms POST /api/v2/orders",
	"Cache hit ratio: 94.3%",
	"Connection pool: 12/50 in use",
	"Scheduled job 'sync-inventory' started",
	"Scheduled job 'sync-inventory' completed (812 records)",
	"High latency detected: p99=820ms (threshold=500ms)",
	"Retry 1/3 for external payment gateway",
	"User session expired: uid=2048",
	"Rate limit applied to client 203.0.113.42",
	"Database query slow: 1.24s on table orders",
	"Health check OK: all 5 downstream services reachable",
	"Certificate expiry check: api.example.com valid 45d",
	"Memory usage threshold crossed: 87% on worker-03",
	"Deployment hook triggered: app v2.4.1 → staging",
	"Email delivery failed: bounce for user@domain.invalid",
	"New OAuth token issued for app-client-07",
	"Backup archive: 4.18 GB written to s3://backups/daily",
}

func generateLogLine() string {
	level := logLevels[rand.Intn(len(logLevels))]
	src := logSources[rand.Intn(len(logSources))]
	msg := logMessages[rand.Intn(len(logMessages))]
	return fmt.Sprintf("[%s] %-5s %-12s  %s", time.Now().Format("15:04:05"), level, src, msg)
}

func logMonitorScreen(b *Builder) {
	// Seed initial log lines
	initial := make([]string, 30)
	base := time.Now().Add(-30 * time.Minute)
	for i := range initial {
		level := logLevels[rand.Intn(len(logLevels))]
		src := logSources[rand.Intn(len(logSources))]
		msg := logMessages[rand.Intn(len(logMessages))]
		ts := base.Add(time.Duration(i) * time.Minute).Format("15:04:05")
		initial[i] = fmt.Sprintf("[%s] %-5s %-12s  %s", ts, level, src, msg)
	}

	b.Flex("log-monitor", false, "stretch", 0).Padding(1, 2).
		// Title + stats row
		Flex("log-hdr", true, "center", 2).Padding(0, 0, 1, 0).
		Static("log-title", "Log Monitor").Font("bold").Foreground("$cyan").
		Spacer().Hint(-1, 0).
		Spinner("log-spinner", Spinners["braille"]).
		Static("log-live-lbl", " streaming").Foreground("$green").
		End(). // Flex("log-hdr")
		HRule("thin").Padding(0, 0, 1, 0).
		// Toolbar
		Flex("log-toolbar", true, "center", 2).Padding(0, 0, 1, 0).
		Static("log-filter-lbl", "Filter:").Foreground("$gray").
		Input("log-filter", "keyword…").Hint(24, 1).
		Spacer().Hint(2, 0).
		Static("log-src-lbl", "Source:").Foreground("$gray").
		Select("log-src", "all", "All sources",
			"api-server", "api-server",
			"worker-01", "worker-01",
			"worker-02", "worker-02",
			"db-pool", "db-pool",
			"cache", "cache",
		).
		Spacer().Hint(2, 0).
		Static("log-level-lbl", "Level:").Foreground("$gray").
		Select("log-level", "all", "All levels",
			"debug", "DEBUG",
			"info", "INFO",
			"warn", "WARN",
			"error", "ERROR",
		).
		Spacer().Hint(-1, 0).
		Button("log-clear", " ✕ Clear").
		Button("log-pause", " ⏸ Pause").
		Button("log-save", " ↓ Save").
		End(). // Flex("log-toolbar")
		// Main split: log view + stats sidebar
		Grid("log-body", 1, 2, false).Hint(0, -1).Columns(-3, -1).Border("none").
		Cell(0, 0, 1, 1).
		Flex("log-view-pane", false, "stretch", 0).Border("", "round").
		Static("log-view-title", " Log Stream").Font("bold").Background("$bg2").
		Text("log-text", initial, true, 2000).Hint(0, -1).
		End(). // Flex("log-view-pane")
		Cell(1, 0, 1, 1).
		Flex("log-stats-pane", false, "stretch", 0).Border("", "round").
		Static("log-stats-title", " Statistics").Font("bold").Background("$bg2").
		Spacer().Hint(0, 1).
		Static("log-stat-hdr", "Last 30 min").Font("bold").Foreground("$fg1").Padding(0, 1).
		HRule("thin").
		Static("log-stat-total", "  Total      1,842").Foreground("$fg0").Padding(0, 1).
		Static("log-stat-info", "  INFO       1,209").Foreground("$cyan").Padding(0, 1).
		Static("log-stat-warn", "  WARN         487").Foreground("$yellow").Padding(0, 1).
		Static("log-stat-error", "  ERROR         93").Foreground("$red").Padding(0, 1).
		Static("log-stat-debug", "  DEBUG          53").Foreground("$gray").Padding(0, 1).
		HRule("thin").
		Static("log-rate-hdr", "Rate (msg/min)").Font("bold").Foreground("$fg1").Padding(0, 1).
		Static("log-rate-val", "  ≈ 61 / min").Foreground("$green").Padding(0, 1).
		HRule("thin").
		Static("log-top-hdr", "Top Sources").Font("bold").Foreground("$fg1").Padding(0, 1).
		Static("log-top1", "  api-server  38%").Foreground("$fg0").Padding(0, 1).
		Static("log-top2", "  worker-01   22%").Foreground("$fg0").Padding(0, 1).
		Static("log-top3", "  db-pool     18%").Foreground("$fg0").Padding(0, 1).
		Static("log-top4", "  scheduler    9%").Foreground("$fg0").Padding(0, 1).
		Static("log-top5", "  others      13%").Foreground("$gray").Padding(0, 1).
		End(). // Flex("log-stats-pane")
		End(). // Grid("log-body")
		End()  // Flex("log-monitor")

	container := b.Find("log-monitor").(Container)
	var ticker *time.Ticker
	var paused bool

	container.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		for _, sp := range FindAll[*Spinner](container) {
			sp.Start(80 * time.Millisecond)
		}
		ticker = time.NewTicker(1200 * time.Millisecond)
		go func() {
			for range ticker.C {
				if !paused {
					if logText, ok := Find(container, "log-text").(*Text); ok {
						logText.Add(generateLogLine())
					}
				}
			}
		}()
		return true
	})
	container.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		for _, sp := range FindAll[*Spinner](container) {
			sp.Stop()
		}
		if ticker != nil {
			ticker.Stop()
		}
		return true
	})

	b.Find("log-pause").On(EvtActivate, func(w Widget, _ Event, _ ...any) bool {
		paused = !paused
		if btn, ok := w.(*Button); ok {
			if paused {
				btn.Set(" ▶ Resume")
			} else {
				btn.Set(" ⏸ Pause")
			}
			btn.Refresh()
		}
		return true
	})
	b.Find("log-clear").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if logText, ok := Find(container, "log-text").(*Text); ok {
			logText.Clear()
		}
		return true
	})
}

// ── Screen 4: Process Manager ──────────────────────────────────────────────────

func processScreen(b *Builder) {
	procHeaders := []string{"PID", "Name", "User", "CPU%", "MEM", "Status", "Threads", "Started"}
	procData := [][]string{
		{"1", "systemd", "root", "0.0", "  8 MB", "● running", " 1", "14d ago"},
		{"412", "nginx", "www-data", "0.3", " 24 MB", "● running", " 4", "14d ago"},
		{"1204", "postgresql", "postgres", "2.1", "512 MB", "● running", "12", "14d ago"},
		{"1831", "redis-server", "redis", "0.1", " 64 MB", "● running", " 4", "14d ago"},
		{"2048", "app-server", "app", "8.4", "256 MB", "● running", "16", " 2d ago"},
		{"2051", "app-server", "app", "7.9", "248 MB", "● running", "16", " 2d ago"},
		{"2055", "app-server", "app", "9.2", "261 MB", "● running", "16", " 2d ago"},
		{"3100", "celery", "app", "0.0", "  0 MB", "○ stopped", " 0", "  —    "},
		{"3210", "prometheus", "monitor", "1.8", "128 MB", "● running", " 8", " 2d ago"},
		{"3215", "grafana", "monitor", "0.9", " 96 MB", "● running", " 6", " 2d ago"},
		{"4001", "node_exporter", "monitor", "0.2", " 16 MB", "● running", " 4", " 2d ago"},
		{"5500", "sshd", "root", "0.0", " 12 MB", "● running", " 1", "14d ago"},
		{"8912", "bash", "admin", "0.0", "  4 MB", "● running", " 1", " 1h ago"},
		{"14022", "top", "admin", "0.4", "  2 MB", "● running", " 1", " 0h ago"},
	}

	// Summary progress bars
	cpuTotal := NewProgress("proc-cpu-total", "", true)
	cpuTotal.SetTotal(100)
	cpuTotal.SetValue(31) // sum of process CPUs

	memTotal := NewProgress("proc-mem-total", "", true)
	memTotal.SetTotal(100)
	memTotal.SetValue(41)

	b.Flex("process-mgr", false, "stretch", 0).Padding(1, 2).
		// Title
		Flex("proc-hdr", true, "center", 2).Padding(0, 0, 1, 0).
		Static("proc-title", "Process Manager").Font("bold").Foreground("$cyan").
		Spacer().Hint(-1, 0).
		Static("proc-summary", "14 processes  ·  13 running  ·  1 stopped").Foreground("$gray").
		End(). // Flex("proc-hdr")
		HRule("thin").Padding(0, 0, 1, 0).
		// Resource summary bar
		Flex("proc-res-row", true, "stretch", 4).Padding(0, 0, 1, 0).
		Flex("proc-cpu-card", false, "start", 0).Border("", "round").Padding(0, 2).
		Static("proc-cpu-lbl", "Total CPU").Foreground("$gray").
		Static("proc-cpu-val", "31%").Font("bold").Foreground("$yellow").
		Add(cpuTotal).Hint(20, 1).
		End(). // Flex("proc-cpu-card")
		Flex("proc-mem-card", false, "start", 0).Border("", "round").Padding(0, 2).
		Static("proc-mem-lbl", "Memory Used").Foreground("$gray").
		Static("proc-mem-val", "1.43 GB / 10 GB").Font("bold").Foreground("$cyan").
		Add(memTotal).Hint(20, 1).
		End(). // Flex("proc-mem-card")
		Spacer().Hint(-1, 0).
		Flex("proc-info-card", false, "start", 0).Border("", "round").Padding(0, 2).
		Static("proc-load-lbl", "Load Average").Foreground("$gray").
		Static("proc-load-val", "1.42  1.55  1.61").Font("bold").Foreground("$green").
		Static("proc-uptime-lbl", "Uptime").Foreground("$gray").
		Static("proc-uptime-val", "14d 06h 23m").Font("bold").Foreground("$fg1").
		End(). // Flex("proc-info-card")
		End(). // Flex("proc-res-row")
		// Toolbar
		Flex("proc-toolbar", true, "center", 2).Padding(0, 0, 1, 0).
		Static("proc-filter-lbl", "Filter:").Foreground("$gray").
		Input("proc-filter", "name or PID…").Hint(24, 1).
		Select("proc-sort", "cpu", "Sort: CPU%", "mem", "Sort: Memory", "pid", "Sort: PID", "name", "Sort: Name").
		Spacer().Hint(-1, 0).
		Button("proc-kill", " ✕ Kill").
		Button("proc-restart", " ↺ Restart").
		Button("proc-detail", " ⬡ Details").
		Button("proc-refresh", " ↻ Refresh").
		End(). // Flex("proc-toolbar")
		// Process table
		Flex("proc-table-pane", false, "stretch", 0).Border("", "round").Hint(0, -1).
		Static("proc-table-title", " Processes").Font("bold").Background("$bg2").
		Table("proc-table", NewArrayTableProvider(procHeaders, procData), false).Hint(0, -1).
		End(). // Flex("proc-table-pane")
		End()  // Flex("process-mgr")

	container := b.Find("process-mgr").(Container)

	b.Find("proc-kill").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		procTable := Find(container, "proc-table").(*Table)
		row, _ := procTable.Selected()
		if row >= 0 && row < len(procData) {
			pid := procData[row][0]
			name := strings.TrimSpace(procData[row][1])
			if lbl, ok := Find(container, "proc-summary").(*Static); ok {
				lbl.SetText(fmt.Sprintf("Sent SIGTERM to %s (PID %s)", name, pid))
			}
		}
		return true
	})
	b.Find("proc-refresh").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "proc-summary").(*Static); ok {
			lbl.SetText(fmt.Sprintf("14 processes  ·  refreshed %s", time.Now().Format("15:04:05")))
		}
		return true
	})
}

// ── Screen 5: Data Entry — Order Form ──────────────────────────────────────────

func dataEntryScreen(b *Builder) {
	type CustomerInfo struct {
		Company string `width:"36"`
		Name    string `label:"Contact Name" width:"36"`
		Email   string `label:"E-Mail" width:"36"`
		Phone   string `width:"24"`
		Country string `control:"select" options:",de,Germany,us,United States,fr,France,gb,United Kingdom,jp,Japan,other,Other"`
	}

	type ShippingInfo struct {
		Address    string `label:"Street Address" width:"36"`
		City       string `width:"24"`
		PostalCode string `label:"Postal Code" width:"12"`
		Priority   string `label:"Shipping Method" control:"select" options:"std,Standard (5-7d),exp,Express (2d),next,Next Day"`
		Insurance  bool   `label:"Shipment Insurance"`
	}

	type PaymentInfo struct {
		Method    string `label:"Payment Method" control:"select" options:"invoice,Invoice,cc,Credit Card,wire,Wire Transfer,crypto,Crypto"`
		PO        string `label:"PO Number" width:"24"`
		VATNumber string `label:"VAT / Tax ID" width:"24"`
		Notes     string `label:"Order Notes" width:"40"`
		Agreed    bool   `label:"I agree to the terms & conditions"`
	}

	customer := CustomerInfo{
		Company: "Acme Corporation",
		Name:    "Jane Smith",
		Email:   "jane.smith@acme.example",
		Phone:   "+1-555-010-2048",
		Country: "us",
	}
	shipping := ShippingInfo{
		Address:    "1600 Amphitheatre Parkway",
		City:       "Mountain View",
		PostalCode: "94043",
		Priority:   "exp",
		Insurance:  true,
	}
	payment := PaymentInfo{
		Method:    "invoice",
		PO:        "PO-2026-0042",
		VATNumber: "US-123456789",
	}

	orderHeaders := []string{"#", "SKU", "Product", "Qty", "Unit Price", "Total"}
	orderData := [][]string{
		{"1", "ZW-2001", "Zeichenwerk Pro License", "5", "  $249.00", " $1,245.00"},
		{"2", "SW-4412", "Support & Maintenance 1yr", "5", "   $49.00", "   $245.00"},
		{"3", "TR-0801", "Training Package (remote)", "2", "  $599.00", " $1,198.00"},
	}

	b.Flex("data-entry", false, "stretch", 0).Padding(1, 2).
		// Title
		Flex("de-hdr", true, "center", 2).Padding(0, 0, 1, 0).
		Static("de-title", "New Order Entry").Font("bold").Foreground("$cyan").
		Spacer().Hint(-1, 0).
		Static("de-ref", "Draft  ·  REF #2026-0099").Foreground("$gray").
		End(). // Flex("de-hdr")
		HRule("thin").Padding(0, 0, 1, 0).
		// Two-column layout: form sections left, order items right
		Grid("de-body", 1, 2, false).Hint(0, -1).Columns(-1, -1).
		// Left column: collapsible form sections
		Cell(0, 0, 1, 1).
		Flex("de-form-col", false, "stretch", 0).
		Collapsible("de-cust-section", "  ① Customer Information", true).
		Flex("de-cust-content", false, "stretch", 0).Padding(0, 2).
		Form("de-cust-form", "", &customer).
		Group("de-cust-grp", "", "", false, 1).
		End(). // Group("de-cust-grp")
		End(). // Form("de-cust-form")
		End(). // Flex("de-cust-content")
		End(). // Collapsible("de-cust-section")
		Spacer().Hint(0, 1).
		Collapsible("de-ship-section", "  ② Shipping Details", true).
		Flex("de-ship-content", false, "stretch", 0).Padding(0, 2).
		Form("de-ship-form", "", &shipping).
		Group("de-ship-grp", "", "", false, 1).
		End(). // Group("de-ship-grp")
		End(). // Form("de-ship-form")
		End(). // Flex("de-ship-content")
		End(). // Collapsible("de-ship-section")
		Spacer().Hint(0, 1).
		Collapsible("de-pay-section", "  ③ Payment & Terms", false).
		Flex("de-pay-content", false, "stretch", 0).Padding(0, 2).
		Form("de-pay-form", "", &payment).
		Group("de-pay-grp", "", "", false, 1).
		End(). // Group("de-pay-grp")
		End(). // Form("de-pay-form")
		End(). // Flex("de-pay-content")
		End(). // Collapsible("de-pay-section")
		End(). // Flex("de-form-col")
		// Right column: order items + summary
		Cell(1, 0, 1, 1).
		Flex("de-items-col", false, "stretch", 0).Padding(0, 0, 0, 2).
		Static("de-items-title", "Order Items").Font("bold").Foreground("$fg1").Padding(0, 0, 1, 0).
		Table("de-items-table", NewArrayTableProvider(orderHeaders, orderData), false).Hint(0, 8).
		Spacer().Hint(0, 1).
		// Order summary box
		Flex("de-summary", false, "stretch", 0).Border("", "round").Padding(1, 2).
		Static("de-sum-title", "Order Summary").Font("bold").Foreground("$fg1").
		HRule("thin").
		Flex("de-sum-row1", true, "stretch", 0).
		Static("de-sum-subtotal-lbl", "Subtotal").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("de-sum-subtotal-val", "$2,688.00").
		End(). // Flex("de-sum-row1")
		Flex("de-sum-row2", true, "stretch", 0).
		Static("de-sum-tax-lbl", "Tax (0%)").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("de-sum-tax-val", "    $0.00").
		End(). // Flex("de-sum-row2")
		Flex("de-sum-row3", true, "stretch", 0).
		Static("de-sum-ship-lbl", "Shipping (Express)").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("de-sum-ship-val", "   $35.00").
		End(). // Flex("de-sum-row3")
		HRule("thin").
		Flex("de-sum-total-row", true, "stretch", 0).
		Static("de-sum-total-lbl", "Total").Font("bold").Foreground("$fg0").
		Spacer().Hint(-1, 0).
		Static("de-sum-total-val", "$2,723.00").Font("bold").Foreground("$cyan").
		End(). // Flex("de-sum-total-row")
		End(). // Flex("de-summary")
		Spacer().Hint(0, -1).
		// Action buttons
		Flex("de-actions", true, "end", 2).
		Button("de-btn-draft", " ↓ Save Draft").
		Button("de-btn-cancel", " ✕ Cancel").
		Button("de-btn-submit", " ✓ Submit Order").
		End(). // Flex("de-actions")
		Static("de-status", "").Foreground("$green").Padding(1, 0, 0, 0).
		End(). // Flex("de-items-col")
		End(). // Grid("de-body")
		End()  // Flex("data-entry")

	container := b.Find("data-entry").(Container)

	b.Find("de-btn-submit").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "de-status").(*Static); ok {
			lbl.SetText(fmt.Sprintf("✓  Order REF #2026-0099 submitted at %s", time.Now().Format("15:04:05")))
		}
		if title, ok := Find(container, "de-ref").(*Static); ok {
			title.SetText(fmt.Sprintf("Submitted  ·  REF #2026-0099"))
		}
		return true
	})
	b.Find("de-btn-draft").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "de-status").(*Static); ok {
			lbl.SetText(fmt.Sprintf("Draft saved at %s", time.Now().Format("15:04:05")))
		}
		return true
	})
	b.Find("de-btn-cancel").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "de-status").(*Static); ok {
			lbl.SetText("Changes discarded.")
		}
		return true
	})
}

// ── Screen 6: Code Editor ──────────────────────────────────────────────────────

func codeEditorScreen(b *Builder) {
	// ── File contents (short illustrative snippets) ────────────────────────────
	mainGoContent := []string{
		"package main",
		"",
		"import (",
		`    . "github.com/tekugo/zeichenwerk"`,
		")",
		"",
		"func main() {",
		"    createUI().Run()",
		"}",
		"",
		"func createUI() *UI {",
		"    return NewBuilder(MidnightNeonTheme()).",
		`        Flex("root", false, "stretch", 0).`,
		`        Static("title", "Hello, World!").`,
		`            Font("bold").Foreground("$cyan").`,
		`        End(). // Flex("root")`,
		"        Build()",
		"}",
	}

	tableGoContent := []string{
		"// Table is a scrollable data grid widget.",
		"// It renders rows from a TableProvider with",
		"// optional column separators and highlights.",
		"type Table struct {",
		"    Component",
		"    provider   TableProvider",
		"    row        int",
		"    column     int",
		"    offsetX    int",
		"    offsetY    int",
		"    grid       *Border",
		"    inner      bool",
		"    outer      bool",
		"}",
		"",
		"func (t *Table) Render(r *Renderer) {",
		"    t.Component.Render(r)",
		"    x, y, w, h := t.Content()",
		"    t.renderTableHeader(r, x, y, w, h)",
		"    t.renderTableContent(r, x, y+2, w, h-2)",
		"}",
	}

	readmeContent := "# Zeichenwerk\n\n" +
		"**Zeichenwerk** is a lightweight TUI framework for Go.\n\n" +
		"## Features\n\n" +
		"- *Declarative* builder API\n" +
		"- CSS-like theming with `$variables`\n" +
		"- **Responsive** layout engine (Flex, Grid)\n" +
		"- Rich widget set: `Table`, `Tree`, `Editor`, `Deck`\n\n" +
		"## Quick Start\n\n" +
		"Create a `NewBuilder`, compose screens with fluent calls,\n" +
		"then call `Run()` to start the event loop.\n\n" +
		"## Keyboard Shortcuts\n\n" +
		"- `Tab` / `Shift+Tab`  move focus between widgets\n" +
		"- `Ctrl+D`  open the live widget inspector\n" +
		"- `Ctrl+Q`  quit\n\n" +
		"*Tip: Press Ctrl+D to inspect the widget tree of this screen.*"

	// ── File tree ──────────────────────────────────────────────────────────────
	root := NewTreeNode("zeichenwerk")
	widgets := NewTreeNode("widgets")
	widgets.Add(NewTreeNode("main.go", 0)) // data = tab index
	widgets.Add(NewTreeNode("table.go", 1))
	widgets.Add(NewTreeNode("flex.go"))
	widgets.Add(NewTreeNode("editor.go"))
	cmd := NewTreeNode("cmd")
	cmd.Add(NewTreeNode("showcase"))
	cmd.Add(NewTreeNode("demo"))
	root.Add(widgets)
	root.Add(cmd)
	root.Add(NewTreeNode("README.md", 2))
	root.Add(NewTreeNode("go.mod"))

	// ── Layout ─────────────────────────────────────────────────────────────────
	b.Flex("code-editor", false, "stretch", 0).Padding(1, 2).
		// Header
		Flex("ce-hdr", true, "center", 2).Padding(0, 0, 1, 0).
		Static("ce-title", "Code Editor").Font("bold").Foreground("$cyan").
		Spacer().Hint(-1, 0).
		Static("ce-status", "main.go — Ln 1, Col 1").Foreground("$gray").
		Button("ce-btn-new", " + New").
		End(). // Flex("ce-hdr")
		HRule("thin").Padding(0, 0, 1, 0).
		// Body: file tree + tabbed editor
		Grid("ce-body", 1, 2, false).Hint(0, -1).Columns(26, -1).Border("none").
		Cell(0, 0, 1, 1).
		Flex("ce-tree-pane", false, "stretch", 0).Border("", "round").
		Static("ce-tree-title", " Project").Font("bold").Background("$bg2").
		Tree("ce-tree").Hint(0, -1).
		End(). // Flex("ce-tree-pane")
		Cell(1, 0, 1, 1).
		Flex("ce-edit-col", false, "stretch", 0).Hint(0, -1).
		Tabs("ce-tabs").
		Switcher("ce-switcher", true).Hint(0, -1). // auto-wired to Tabs via EvtActivate
		Tab("main.go").
		Editor("ce-editor-main").Hint(0, -1).
		Tab("table.go").
		Editor("ce-editor-table").Hint(0, -1).
		Tab("README.md").
		Viewport("ce-viewport", "").
		Styled("ce-preview", readmeContent).
		End(). // Viewport("ce-viewport")
		End(). // Switcher("ce-switcher")
		End(). // Tabs("ce-tabs")
		End(). // Flex("ce-edit-col")
		End(). // Grid("ce-body")
		End()  // Flex("code-editor")

	// ── Post-build wiring ──────────────────────────────────────────────────────
	container := b.Find("code-editor").(Container)

	// Set editor contents and enable line numbers
	mainEditor := Find(container, "ce-editor-main").(*Editor)
	mainEditor.SetContent(mainGoContent)
	mainEditor.ShowLineNumbers(true)

	tableEditor := Find(container, "ce-editor-table").(*Editor)
	tableEditor.SetContent(tableGoContent)
	tableEditor.ShowLineNumbers(true)

	// Populate the file tree and expand top-level directories
	tree := Find(container, "ce-tree").(*Tree)
	tree.Add(root)
	tree.Expand(root)
	tree.Expand(widgets)

	// Wire tree selection → switch active tab + pane
	tabs := Find(container, "ce-tabs").(*Tabs)
	switcher := Find(container, "ce-switcher").(*Switcher)
	tree.On(EvtSelect, func(_ Widget, _ Event, _ ...any) bool {
		if node := tree.Selected(); node != nil {
			if idx, ok := node.Data().(int); ok {
				tabs.Select(idx)
				switcher.Select(idx)
			}
		}
		return false
	})

	// Wire editor changes → update status bar with cursor position
	statusLbl := Find(container, "ce-status").(*Static)
	mainEditor.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		cx, cy, _ := mainEditor.Cursor()
		statusLbl.SetText(fmt.Sprintf("main.go — Ln %d, Col %d", cy+1, cx+1))
		return false
	})

	// Wire "New" button → popup a dialog
	b.Find("ce-btn-new").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		ui := FindUI(container)
		dlg := NewDialog("ce-dlg", "dialog", " New File")
		body := ui.NewBuilder().
			Flex("ce-dlg-body", false, "stretch", 0).
			Static("ce-dlg-lbl", "Filename:").Foreground("$fg1").
			Input("ce-dlg-input", "untitled.go").Hint(28, 1).
			Flex("ce-dlg-btns", true, "end", 2).Padding(1, 0, 0, 0).
			Button("ce-dlg-ok", " ✓ Create").Class("dialog").
			Button("ce-dlg-cancel", " ✕ Cancel").
			End(). // Flex("ce-dlg-btns")
			Container()
		dlg.Add(body)
		Find(body, "ce-dlg-cancel").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
			ui.Close()
			return true
		})
		ui.Popup(-1, -1, 0, 0, dlg)
		return true
	})
}
