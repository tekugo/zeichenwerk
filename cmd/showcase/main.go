package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	. "github.com/tekugo/zeichenwerk"
)

type navItem struct{ icon, name, desc string }

func main() {
	createUI().Run()
}

// ── Shell ──────────────────────────────────────────────────────────────────────

func createUI() *UI {
	navItems := []any{
		navItem{"◈", "Dashboard", "System metrics & KPIs"},
		navItem{"◉", "User Admin", "Manage users & roles"},
		navItem{"≡", "Log Monitor", "Live log tail & stats"},
		navItem{"⬡", "Processes", "Running process list"},
		navItem{"◧", "Data Entry", "Forms & input controls"},
	}

	renderNav := func(r *Renderer, x, y, w, h, index int, data any, selected bool) {
		item := data.(navItem)
		bg := "$bg1"
		if selected {
			r.Set("$cyan", bg, "bold")
			r.Fill(x, y, w, 1, " ")
			r.Put(x, y, "┃")
			r.Text(x+2, y, item.icon+"  "+item.name, w-2)
			r.Set("$cyan", bg, "")
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

	ui := NewBuilder(MidnightNeonTheme()).
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
		End().
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
		End().
		Cell(1, 0, 1, 1).
		Switcher("content", false).
		With(dashboardScreen).
		With(userAdminScreen).
		With(logMonitorScreen).
		With(processScreen).
		With(dataEntryScreen).
		End().
		End().
		// ── Footer ────────────────────────────────────────────────────────────
		Flex("footer", true, "center", 0).Background("$bg1").Padding(0, 1).
		Static("footer-keys", " ↑↓ Navigate   Tab/Shift+Tab Focus   Enter/Space Activate   Esc Cancel").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("footer-brand", "Zeichenwerk v2.0 ◈").Foreground("$gray").
		End().
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
		End().
		HRule("thin").Padding(0, 0, 1, 0).
		// ── KPI cards ────────────────────────────────────────────────────────
		Flex("kpi-row", true, "stretch", 2).Padding(0, 0, 1, 0).
		// CPU card
		Flex("kpi-cpu", false, "start", 0).Border("", "round").Padding(1, 2).
		Static("kpi-cpu-lbl", "CPU Usage").Foreground("$gray").
		Digits("kpi-cpu-val", "64").Foreground("$yellow").
		Add(cpuProg).Hint(0, 1).
		Static("kpi-cpu-sub", "% · 8 cores · 3.2 GHz").Foreground("$gray").
		End().
		// Memory card
		Flex("kpi-mem", false, "start", 0).Border("", "round").Padding(1, 2).
		Static("kpi-mem-lbl", "Memory").Foreground("$gray").
		Digits("kpi-mem-val", "4.1").Foreground("$cyan").
		Add(memProg).Hint(0, 1).
		Static("kpi-mem-sub", "GB of 10 GB total").Foreground("$gray").
		End().
		// Disk card
		Flex("kpi-disk", false, "start", 0).Border("", "round").Padding(1, 2).
		Static("kpi-disk-lbl", "Disk").Foreground("$gray").
		Digits("kpi-disk-val", "780").Foreground("$orange").
		Add(diskProg).Hint(0, 1).
		Static("kpi-disk-sub", "GB of 1 TB · 78% full").Foreground("$gray").
		End().
		// Network card
		Flex("kpi-net", false, "start", 0).Border("", "round").Padding(1, 2).
		Static("kpi-net-lbl", "Network").Foreground("$gray").
		Digits("kpi-net-val", "2.4").Foreground("$green").
		Add(netProg).Hint(0, 1).
		Static("kpi-net-sub", "Mb/s ↑ · ↓8.1 Mb/s recv").Foreground("$gray").
		End().
		End().
		// ── Services + Alerts ────────────────────────────────────────────────
		Grid("dash-mid", 1, 2, false).Hint(0, 10).Columns(-2, -1).Padding(0, 0, 1, 0).Border("none").
		Cell(0, 0, 1, 1).
		Flex("svc-pane", false, "stretch", 0).Border("", "round").
		Static("svc-title", " Services").Font("bold").Foreground("$fg0").Background("$bg2").
		Table("svc-table", svcTable).Hint(0, -1).
		End().
		Cell(1, 0, 1, 1).
		Flex("alerts-pane", false, "stretch", 0).Border("", "round").
		Static("alerts-title", " Alerts & Notices").Font("bold").Foreground("$fg0").Background("$bg2").
		Static("alert1", "⚠  Disk usage on /var exceeds 80%").Foreground("$orange").Padding(0, 1).
		Static("alert2", "⚠  worker-03 memory pressure").Foreground("$orange").Padding(0, 1).
		Static("alert3", "✓  All SSL certificates valid").Foreground("$green").Padding(0, 1).
		Static("alert4", "✓  Last backup: 4h ago (success)").Foreground("$green").Padding(0, 1).
		Static("alert5", "✓  All database replicas in sync").Foreground("$green").Padding(0, 1).
		Static("alert6", "ℹ  Maintenance window: Sun 02:00 UTC").Foreground("$cyan").Padding(0, 1).
		End().
		End().
		// ── Activity log ─────────────────────────────────────────────────────
		Flex("activity-pane", false, "stretch", 0).Border("", "round").Hint(0, -1).
		Static("activity-title", " Recent Activity").Font("bold").Foreground("$fg0").Background("$bg2").
		Text("activity-log", activityLines, false, 200).Hint(0, -1).
		End().
		End()

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
		End().
		HRule("thin").Padding(0, 0, 1, 0).
		// Toolbar
		Flex("ua-toolbar", true, "center", 2).Padding(0, 0, 1, 0).
		Static("ua-search-lbl", "Search:").Foreground("$gray").
		Typeahead("ua-search", "", "name, email or role…").Hint(28, 1).
		Spacer().Hint(-1, 0).Class("dialog").
		Button("ua-btn-new", " + New User").
		Button("ua-btn-del", " ✕ Delete").
		Button("ua-btn-exp", " ↓ Export").
		End().
		// Split: table left, detail right
		Grid("ua-body", 1, 2, false).Hint(0, -1).Columns(-3, -2).Border("none").
		Cell(0, 0, 1, 1).
		Flex("ua-list-pane", false, "stretch", 0).Border("", "round").
		Static("ua-list-title", " Users").Font("bold").Background("$bg2").
		Table("ua-table", NewArrayTableProvider(userHeaders, userData)).Hint(0, -1).
		End().
		Cell(1, 0, 1, 1).
		Flex("ua-detail-pane", false, "stretch", 0).Border("", "round").
		Static("ua-detail-title", " Edit User").Font("bold").Background("$bg2").
		Form("ua-form", "", &selectedUser).
		Group("ua-group", "", "", false, 1).Padding(1).
		End().
		End().
		Flex("ua-detail-btns", true, "end", 2).Padding(1).
		Button("ua-save", " ✓ Save Changes").Class("dialog").
		Button("ua-reset", " ↺ Reset").
		Button("ua-deactivate", " ⊘ Deactivate").
		End().
		End().
		End().
		End()

	container := b.Find("user-admin").(Container)
	b.Find("ua-save").On(EvtClick, func(w Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "ua-detail-title").(*Static); ok {
			lbl.SetText(fmt.Sprintf(" Edit User  ✓ Saved %s", time.Now().Format("15:04:05")))
		}
		return true
	})
	b.Find("ua-reset").On(EvtClick, func(_ Widget, _ Event, _ ...any) bool {
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
		End().
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
		End().
		// Main split: log view + stats sidebar
		Grid("log-body", 1, 2, false).Hint(0, -1).Columns(-3, -1).Border("none").
		Cell(0, 0, 1, 1).
		Flex("log-view-pane", false, "stretch", 0).Border("", "round").
		Static("log-view-title", " Log Stream").Font("bold").Background("$bg2").
		Text("log-text", initial, true, 2000).Hint(0, -1).
		End().
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
		End().
		End().
		End()

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

	b.Find("log-pause").On(EvtClick, func(w Widget, _ Event, _ ...any) bool {
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
	b.Find("log-clear").On(EvtClick, func(_ Widget, _ Event, _ ...any) bool {
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
		End().
		HRule("thin").Padding(0, 0, 1, 0).
		// Resource summary bar
		Flex("proc-res-row", true, "stretch", 4).Padding(0, 0, 1, 0).
		Flex("proc-cpu-card", false, "start", 0).Border("", "round").Padding(0, 2).
		Static("proc-cpu-lbl", "Total CPU").Foreground("$gray").
		Static("proc-cpu-val", "31%").Font("bold").Foreground("$yellow").
		Add(cpuTotal).Hint(20, 1).
		End().
		Flex("proc-mem-card", false, "start", 0).Border("", "round").Padding(0, 2).
		Static("proc-mem-lbl", "Memory Used").Foreground("$gray").
		Static("proc-mem-val", "1.43 GB / 10 GB").Font("bold").Foreground("$cyan").
		Add(memTotal).Hint(20, 1).
		End().
		Spacer().Hint(-1, 0).
		Flex("proc-info-card", false, "start", 0).Border("", "round").Padding(0, 2).
		Static("proc-load-lbl", "Load Average").Foreground("$gray").
		Static("proc-load-val", "1.42  1.55  1.61").Font("bold").Foreground("$green").
		Static("proc-uptime-lbl", "Uptime").Foreground("$gray").
		Static("proc-uptime-val", "14d 06h 23m").Font("bold").Foreground("$fg1").
		End().
		End().
		// Toolbar
		Flex("proc-toolbar", true, "center", 2).Padding(0, 0, 1, 0).
		Static("proc-filter-lbl", "Filter:").Foreground("$gray").
		Input("proc-filter", "name or PID…").Hint(24, 1).
		Select("proc-sort", "cpu", "Sort: CPU%", "mem", "Sort: Memory", "pid", "Sort: PID", "name", "Sort: Name").
		Spacer().Hint(-1, 0).
		Button("proc-kill", " ✕ Kill").
		Button("proc-restart", " ↺ Restart").
		Button("proc-detail", " ⬡ Details").
		Button("proc-refresh", " ↻ Refresh").Class("dialog").
		End().
		// Process table
		Flex("proc-table-pane", false, "stretch", 0).Border("", "round").Hint(0, -1).
		Static("proc-table-title", " Processes").Font("bold").Background("$bg2").
		Table("proc-table", NewArrayTableProvider(procHeaders, procData)).Hint(0, -1).
		End().
		End()

	container := b.Find("process-mgr").(Container)

	b.Find("proc-kill").On(EvtClick, func(_ Widget, _ Event, _ ...any) bool {
		procTable := Find(container, "proc-table").(*Table)
		row := procTable.GetSelectedRow()
		if row >= 0 && row < len(procData) {
			pid := procData[row][0]
			name := strings.TrimSpace(procData[row][1])
			if lbl, ok := Find(container, "proc-summary").(*Static); ok {
				lbl.SetText(fmt.Sprintf("Sent SIGTERM to %s (PID %s)", name, pid))
			}
		}
		return true
	})
	b.Find("proc-refresh").On(EvtClick, func(_ Widget, _ Event, _ ...any) bool {
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
		End().
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
		End().
		End().
		End().
		End(). // de-cust-section
		Spacer().Hint(0, 1).
		Collapsible("de-ship-section", "  ② Shipping Details", true).
		Flex("de-ship-content", false, "stretch", 0).Padding(0, 2).
		Form("de-ship-form", "", &shipping).
		Group("de-ship-grp", "", "", false, 1).
		End().
		End().
		End().
		End(). // de-ship-section
		Spacer().Hint(0, 1).
		Collapsible("de-pay-section", "  ③ Payment & Terms", false).
		Flex("de-pay-content", false, "stretch", 0).Padding(0, 2).
		Form("de-pay-form", "", &payment).
		Group("de-pay-grp", "", "", false, 1).
		End().
		End().
		End().
		End(). // de-pay-section
		End(). // de-form-col
		// Right column: order items + summary
		Cell(1, 0, 1, 1).
		Flex("de-items-col", false, "stretch", 0).Padding(0, 0, 0, 2).
		Static("de-items-title", "Order Items").Font("bold").Foreground("$fg1").Padding(0, 0, 1, 0).
		Table("de-items-table", NewArrayTableProvider(orderHeaders, orderData)).Hint(0, 8).
		Spacer().Hint(0, 1).
		// Order summary box
		Flex("de-summary", false, "stretch", 0).Border("", "round").Padding(1, 2).
		Static("de-sum-title", "Order Summary").Font("bold").Foreground("$fg1").
		HRule("thin").
		Flex("de-sum-row1", true, "stretch", 0).
		Static("de-sum-subtotal-lbl", "Subtotal").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("de-sum-subtotal-val", "$2,688.00").
		End().
		Flex("de-sum-row2", true, "stretch", 0).
		Static("de-sum-tax-lbl", "Tax (0%)").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("de-sum-tax-val", "    $0.00").
		End().
		Flex("de-sum-row3", true, "stretch", 0).
		Static("de-sum-ship-lbl", "Shipping (Express)").Foreground("$gray").
		Spacer().Hint(-1, 0).
		Static("de-sum-ship-val", "   $35.00").
		End().
		HRule("thin").
		Flex("de-sum-total-row", true, "stretch", 0).
		Static("de-sum-total-lbl", "Total").Font("bold").Foreground("$fg0").
		Spacer().Hint(-1, 0).
		Static("de-sum-total-val", "$2,723.00").Font("bold").Foreground("$cyan").
		End().
		End().
		Spacer().Hint(0, -1).
		// Action buttons
		Flex("de-actions", true, "end", 2).
		Button("de-btn-draft", " ↓ Save Draft").
		Button("de-btn-cancel", " ✕ Cancel").
		Button("de-btn-submit", " ✓ Submit Order").Class("dialog").
		End().
		Static("de-status", "").Foreground("$green").Padding(1, 0, 0, 0).
		End().
		End().
		End()

	container := b.Find("data-entry").(Container)

	b.Find("de-btn-submit").On(EvtClick, func(_ Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "de-status").(*Static); ok {
			lbl.SetText(fmt.Sprintf("✓  Order REF #2026-0099 submitted at %s", time.Now().Format("15:04:05")))
		}
		if title, ok := Find(container, "de-ref").(*Static); ok {
			title.SetText(fmt.Sprintf("Submitted  ·  REF #2026-0099"))
		}
		return true
	})
	b.Find("de-btn-draft").On(EvtClick, func(_ Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "de-status").(*Static); ok {
			lbl.SetText(fmt.Sprintf("Draft saved at %s", time.Now().Format("15:04:05")))
		}
		return true
	})
	b.Find("de-btn-cancel").On(EvtClick, func(_ Widget, _ Event, _ ...any) bool {
		if lbl, ok := Find(container, "de-status").(*Static); ok {
			lbl.SetText("Changes discarded.")
		}
		return true
	})
}
