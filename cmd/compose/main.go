// Compose API showcase — mirrors cmd/showcase/main.go using the compose package.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"
	z "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/compose"
	"github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
	"github.com/tekugo/zeichenwerk/widgets"
)

type navItem struct{ icon, name, desc string }

func parseFlags() (*core.Theme, bool, bool) {
	t := flag.String("t", "midnight", "Theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	dmp := flag.Bool("dump", false, "Dump widget hierarchy to stdout and exit")
	dmpV := flag.Bool("dump-verbose", false, "Dump widget hierarchy with style details to stdout and exit")
	flag.Parse()
	var theme *core.Theme
	switch *t {
	case "tokyo":
		theme = themes.TokyoNight()
	case "nord":
		theme = themes.Nord()
	case "gruvbox-dark":
		theme = themes.GruvboxDark()
	case "gruvbox-light":
		theme = themes.GruvboxLight()
	case "lipstick":
		theme = themes.Lipstick()
	default:
		theme = themes.MidnightNeon()
	}
	return theme, *dmp, *dmpV
}

func main() {
	theme, dmp, dmpV := parseFlags()
	ui := createUI(theme)
	if dmp || dmpV {
		ui.SetBounds(0, 0, 120, 40)
		ui.Layout()
		ui.Dump(os.Stdout, widgets.DumpOptions{Style: dmpV})
		return
	}

	crt := core.Find(ui, "crt").(*widgets.CRT)

	// Intercept quit shortcuts so the power-off animation plays before exit.
	widgets.OnKey(crt, func(e *tcell.EventKey) bool {
		switch e.Key() {
		case tcell.KeyCtrlC, tcell.KeyCtrlQ:
			crt.PowerOff(20*time.Millisecond, ui.Quit)
			return true
		case tcell.KeyRune:
			if s := e.Str(); s == "q" || s == "Q" {
				crt.PowerOff(20*time.Millisecond, ui.Quit)
				return true
			}
		}
		return false
	})

	crt.Start(20 * time.Millisecond)
	ui.Run()
}

// ── Shell ──────────────────────────────────────────────────────────────────────

func createUI(theme *core.Theme) *z.UI {
	navItems := []any{
		navItem{"◈", "Dashboard", "System metrics & KPIs"},
		navItem{"◉", "User Admin", "Manage users & roles"},
		navItem{"≡", "Log Monitor", "Live log tail & stats"},
		navItem{"⬡", "Processes", "Running process list"},
		navItem{"◧", "Data Entry", "Forms & input controls"},
		navItem{"✎", "Code Editor", "Edit files & preview"},
	}

	renderNav := func(r *core.Renderer, x, y, w, h, index int, data any, selected, focused bool) {
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

	ui := UI(theme,
		CRT("crt", "",
			VFlex("root", "", "stretch", 0,
				// ── Header ──────────────────────────────────────────────────────────
				HFlex("header", "", "center", 0,
					Bg("$bg1"), Padding(0, 1),
					Static("app-icon", "", "◈", Font("bold"), Fg("$cyan"), Padding(0, 1, 0, 0)),
					Static("app-name", "", "TUI Showcase", Font("bold"), Fg("$fg0")),
					Spacer("", Hint(2, 0)),
					VRule("", "thin"),
					Spacer("", Hint(2, 0)),
					Static("app-tagline", "", "Real-world terminal UI use cases", Fg("$gray")),
					Spacer("", Hint(-1, 0)),
					Static("live-indicator", "", "● LIVE", Fg("$green"), Font("bold"), Padding(0, 2, 0, 0)),
					Static("header-sep", "", " | ", Fg("$gray")),
					Static("header-theme-lbl", "", "Theme ", Fg("$gray")),
					Select("theme-select", "", []string{
						"neon", "Midnight Neon",
						"tokyo", "Tokyo Night",
						"gruvbox-dark", "Gruvbox Dark",
						"nrrd", "Nord",
					}, Padding(0, 1, 0, 0)),
				),
				// ── Body ────────────────────────────────────────────────────────────
				Grid("body", "", []int{-1}, []int{26, -1}, false,
					Hint(0, -1),
					Cell(0, 0, 1, 1,
						VFlex("sidebar", "", "stretch", 0,
							Bg("$bg1"),
							Static("sidebar-brand", "", " ◈ SHOWCASE", Font("bold"), Fg("$magenta"), Padding(1, 0)),
							HRule("", "thin"),
							Deck("nav", "", renderNav, 3, Hint(0, -1)),
							HRule("", "thin"),
							Static("sidebar-key1", "", "  ↑↓  Navigate", Fg("$gray"), Padding(0, 0, 0, 0)),
							Static("sidebar-key2", "", "  ↵   Select", Fg("$gray")),
							Static("sidebar-key3", "", "  Tab Focus", Fg("$gray"), Padding(0, 0, 1, 0)),
						),
					),
					Cell(1, 0, 1, 1,
						Switcher("content", "",
							Include(dashboardScreen),
							Include(userAdminScreen),
							Include(logMonitorScreen),
							Include(processScreen),
							Include(dataEntryScreen),
							Include(codeEditorScreen),
						),
					),
				),
				// ── Footer ──────────────────────────────────────────────────────────
				HFlex("footer", "", "center", 0,
					Bg("$bg1"), Padding(0, 1),
					Static("footer-keys", "", " ↑↓ Navigate   Tab/Shift+Tab Focus   Enter/Space Activate   Esc Cancel", Fg("$gray")),
					Spacer("", Hint(-1, 0)),
					Static("footer-brand", "", "Zeichenwerk v2.0 ◈", Fg("$gray")),
				),
			),
		),
	)

	// Wire navigation
	switcher := core.Find(ui, "content").(*widgets.Switcher)
	nav := core.Find(ui, "nav").(*widgets.Deck)
	nav.Set(navItems)
	navSwitch := func(_ core.Widget, _ core.Event, data ...any) bool {
		if len(data) == 1 {
			if sel, ok := data[0].(int); ok {
				switcher.Select(sel)
			}
		}
		return true
	}
	nav.On(widgets.EvtSelect, navSwitch)
	nav.On(widgets.EvtActivate, navSwitch)

	// Wire theme switcher
	themes := map[string]*core.Theme{
		"neon":         themes.MidnightNeon(),
		"tokyo":        themes.TokyoNight(),
		"gruvbox-dark": themes.GruvboxDark(),
		"nrrd":         themes.Nord(),
	}
	core.Find(ui, "theme-select").On(widgets.EvtChange, func(_ core.Widget, _ core.Event, data ...any) bool {
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

func dashboardScreen(theme *core.Theme) core.Widget {
	cpuProg := widgets.NewProgress("dash-cpu-prog", "", true)
	cpuProg.SetTotal(100)
	cpuProg.Set(64)

	memProg := widgets.NewProgress("dash-mem-prog", "", true)
	memProg.SetTotal(100)
	memProg.Set(41)

	diskProg := widgets.NewProgress("dash-disk-prog", "", true)
	diskProg.SetTotal(100)
	diskProg.Set(78)

	netProg := widgets.NewProgress("dash-net-prog", "", true)
	netProg.SetTotal(100)
	netProg.Set(23)

	svcHeaders := []string{"Service", "Status", "CPU", "Memory", "Uptime"}
	svcData := [][]string{
		{"nginx", "● running", "0.3%", " 24 MB", "14d 06h"},
		{"postgresql", "● running", "2.1%", "512 MB", "14d 06h"},
		{"redis", "● running", "0.1%", " 64 MB", "14d 05h"},
		{"celery", "○ stopped", "  —  ", "    —  ", "     —  "},
		{"prometheus", "● running", "1.8%", "128 MB", " 2d 11h"},
		{"grafana", "● running", "0.9%", " 96 MB", " 2d 11h"},
	}
	svcTable := widgets.NewArrayTableProvider(svcHeaders, svcData)

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

	w := Build(theme,
		VFlex("dashboard", "", "stretch", 0,
			Padding(1, 2),
			// Title row
			HFlex("dash-hdr", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("dash-title", "", "System Dashboard", Font("bold"), Fg("$cyan")),
				Spacer("", Hint(-1, 0)),
				Static("dash-date", "", time.Now().Format("Mon, 02 Jan 2006"), Fg("$gray")),
				Static("dash-sep", "", "  ·  ", Fg("$gray")),
				Static("dash-time", "", time.Now().Format("15:04 MST"), Fg("$fg1"), Font("bold")),
			),
			HRule("", "thin", Padding(0, 0, 1, 0)),
			// KPI cards
			HFlex("kpi-row", "", "stretch", 2,
				Padding(0, 0, 1, 0),
				VFlex("kpi-cpu", "", "start", 0,
					Border("", "round"), Padding(1, 2),
					Static("kpi-cpu-lbl", "", "CPU Usage", Fg("$gray")),
					Digits("kpi-cpu-val", "", "64", Fg("$yellow")),
					Progress("dash-cpu-prog", "", true, Value(64), Total(100)),
					Static("kpi-cpu-sub", "", "% · 8 cores · 3.2 GHz", Fg("$gray")),
				),
				VFlex("kpi-mem", "", "start", 0,
					Border("", "round"), Padding(1, 2),
					Static("kpi-mem-lbl", "", "Memory", Fg("$gray")),
					Digits("kpi-mem-val", "", "4.1", Fg("$cyan")),
					Include(func(_ *core.Theme) core.Widget { return memProg }),
					Static("kpi-mem-sub", "", "GB of 10 GB total", Fg("$gray")),
				),
				VFlex("kpi-disk", "", "start", 0,
					Border("", "round"), Padding(1, 2),
					Static("kpi-disk-lbl", "", "Disk", Fg("$gray")),
					Digits("kpi-disk-val", "", "780", Fg("$orange")),
					Include(func(_ *core.Theme) core.Widget { return diskProg }),
					Static("kpi-disk-sub", "", "GB of 1 TB · 78% full", Fg("$gray")),
				),
				VFlex("kpi-net", "", "start", 0,
					Border("", "round"), Padding(1, 2),
					Static("kpi-net-lbl", "", "Network", Fg("$gray")),
					Digits("kpi-net-val", "", "2.4", Fg("$green")),
					Include(func(_ *core.Theme) core.Widget { return netProg }),
					Static("kpi-net-sub", "", "Mb/s ↑ · ↓8.1 Mb/s recv", Fg("$gray")),
				),
			),
			// Services + Alerts
			Grid("dash-mid", "", []int{0}, []int{-2, -1}, false,
				Hint(0, 10), Padding(0, 0, 1, 0), Border("none"),
				Cell(0, 0, 1, 1,
					VFlex("svc-pane", "", "stretch", 0,
						Border("", "round"),
						Static("svc-title", "", " Services", Font("bold"), Fg("$fg0"), Bg("$bg2")),
						Table("svc-table", "", svcTable, false, Hint(0, -1)),
					),
				),
				Cell(1, 0, 1, 1,
					VFlex("alerts-pane", "", "stretch", 0,
						Border("", "round"),
						Static("alerts-title", "", " Alerts & Notices", Font("bold"), Fg("$fg0"), Bg("$bg2")),
						Static("alert1", "", "⚠  Disk usage on /var exceeds 80%", Fg("$orange"), Padding(0, 1)),
						Static("alert2", "", "⚠  worker-03 memory pressure", Fg("$orange"), Padding(0, 1)),
						Static("alert3", "", "✓  All SSL certificates valid", Fg("$green"), Padding(0, 1)),
						Static("alert4", "", "✓  Last backup: 4h ago (success)", Fg("$green"), Padding(0, 1)),
						Static("alert5", "", "✓  All database replicas in sync", Fg("$green"), Padding(0, 1)),
						Static("alert6", "", "ℹ  Maintenance window: Sun 02:00 UTC", Fg("$cyan"), Padding(0, 1)),
					),
				),
			),
			// Activity log
			VFlex("activity-pane", "", "stretch", 0,
				Border("", "round"), Hint(0, -1),
				Static("activity-title", "", " Recent Activity", Font("bold"), Fg("$fg0"), Bg("$bg2")),
				Text("activity-log", "", activityLines, false, 200, Hint(0, -1)),
			),
		),
	)

	container := w.(core.Container)
	container.On(widgets.EvtShow, func(_ core.Widget, _ core.Event, _ ...any) bool {
		for _, sp := range core.FindAll[*widgets.Spinner](container) {
			sp.Start(120 * time.Millisecond)
		}
		return true
	})
	container.On(widgets.EvtHide, func(_ core.Widget, _ core.Event, _ ...any) bool {
		for _, sp := range core.FindAll[*widgets.Spinner](container) {
			sp.Stop()
		}
		return true
	})

	return w
}

// ── Screen 2: User Administration ──────────────────────────────────────────────

func userAdminScreen(theme *core.Theme) core.Widget {
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

	w := Build(theme,
		VFlex("user-admin", "", "stretch", 0,
			Padding(1, 2),
			HFlex("ua-hdr", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("ua-title", "", "User Administration", Font("bold"), Fg("$cyan")),
				Spacer("", Hint(-1, 0)),
				Static("ua-count", "", "12 users  ·  9 active  ·  3 pending/inactive", Fg("$gray")),
			),
			HRule("", "thin", Padding(0, 0, 1, 0)),
			HFlex("ua-toolbar", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("ua-search-lbl", "", "Search:", Fg("$gray")),
				Typeahead("ua-search", "", []string{"name, email or role…"}, Hint(28, 1)),
				Spacer("dialog", Hint(-1, 0)),
				Button("ua-btn-new", "", " + New User"),
				Button("ua-btn-del", "", " ✕ Delete"),
				Button("ua-btn-exp", "", " ↓ Export"),
			),
			Grid("ua-body", "", []int{-1}, []int{-3, -2}, false,
				Hint(0, -1), Border("none"),
				Cell(0, 0, 1, 1,
					VFlex("ua-list-pane", "", "stretch", 0,
						Border("", "round"),
						Static("ua-list-title", "", " Users", Font("bold"), Bg("$bg2")),
						Table("ua-table", "", widgets.NewArrayTableProvider(userHeaders, userData), true, Hint(0, -1)),
					),
				),
				Cell(1, 0, 1, 1,
					VFlex("ua-detail-pane", "", "stretch", 0,
						Border("", "round"),
						Static("ua-detail-title", "", " Edit User", Font("bold"), Bg("$bg2")),
						Form("ua-form", "", "", &selectedUser,
							FormGroup("ua-group", "", "", false, 1,
								Padding(1),
							),
						),
						HFlex("ua-detail-btns", "", "end", 2,
							Padding(1),
							Button("ua-save", "dialog", " ✓ Save Changes"),
							Button("ua-reset", "", " ↺ Reset"),
							Button("ua-deactivate", "", " ⊘ Deactivate"),
						),
					),
				),
			),
		),
	)

	container := w.(core.Container)
	core.Find(container, "ua-save").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		if lbl, ok := core.Find(container, "ua-detail-title").(*widgets.Static); ok {
			lbl.Set(fmt.Sprintf(" Edit User  ✓ Saved %s", time.Now().Format("15:04:05")))
		}
		return true
	})
	core.Find(container, "ua-reset").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		if lbl, ok := core.Find(container, "ua-detail-title").(*widgets.Static); ok {
			lbl.Set(" Edit User")
		}
		return true
	})

	return w
}

// ── Screen 3: Log Monitor ──────────────────────────────────────────────────────

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

func logMonitorScreen(theme *core.Theme) core.Widget {
	initial := make([]string, 30)
	base := time.Now().Add(-30 * time.Minute)
	for i := range initial {
		level := logLevels[rand.Intn(len(logLevels))]
		src := logSources[rand.Intn(len(logSources))]
		msg := logMessages[rand.Intn(len(logMessages))]
		ts := base.Add(time.Duration(i) * time.Minute).Format("15:04:05")
		initial[i] = fmt.Sprintf("[%s] %-5s %-12s  %s", ts, level, src, msg)
	}

	w := Build(theme,
		VFlex("log-monitor", "", "stretch", 0,
			Padding(1, 2),
			HFlex("log-hdr", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("log-title", "", "Log Monitor", Font("bold"), Fg("$cyan")),
				Spacer("", Hint(-1, 0)),
				Spinner("log-spinner", "", widgets.Spinners["braille"]),
				Static("log-live-lbl", "", " streaming", Fg("$green")),
			),
			HRule("", "thin", Padding(0, 0, 1, 0)),
			HFlex("log-toolbar", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("log-filter-lbl", "", "Filter:", Fg("$gray")),
				Input("log-filter", "", []string{"keyword…"}, Hint(24, 1)),
				Spacer("", Hint(2, 0)),
				Static("log-src-lbl", "", "Source:", Fg("$gray")),
				Select("log-src", "", []string{
					"all", "All sources",
					"api-server", "api-server",
					"worker-01", "worker-01",
					"worker-02", "worker-02",
					"db-pool", "db-pool",
					"cache", "cache",
				}),
				Spacer("", Hint(2, 0)),
				Static("log-level-lbl", "", "Level:", Fg("$gray")),
				Select("log-level", "", []string{
					"all", "All levels",
					"debug", "DEBUG",
					"info", "INFO",
					"warn", "WARN",
					"error", "ERROR",
				}),
				Spacer("", Hint(-1, 0)),
				Button("log-clear", "", " ✕ Clear"),
				Button("log-pause", "", " ⏸ Pause"),
				Button("log-save", "", " ↓ Save"),
			),
			Grid("log-body", "", []int{-1}, []int{-3, -1}, false,
				Hint(0, -1), Border("none"),
				Cell(0, 0, 1, 1,
					VFlex("log-view-pane", "", "stretch", 0,
						Border("", "round"),
						Static("log-view-title", "", " Log Stream", Font("bold"), Bg("$bg2")),
						Text("log-text", "", initial, true, 2000, Hint(0, -1)),
					),
				),
				Cell(1, 0, 1, 1,
					VFlex("log-stats-pane", "", "stretch", 0,
						Border("", "round"),
						Static("log-stats-title", "", " Statistics", Font("bold"), Bg("$bg2")),
						Spacer("", Hint(0, 1)),
						Static("log-stat-hdr", "", "Last 30 min", Font("bold"), Fg("$fg1"), Padding(0, 1)),
						HRule("", "thin"),
						Static("log-stat-total", "", "  Total      1,842", Fg("$fg0"), Padding(0, 1)),
						Static("log-stat-info", "", "  INFO       1,209", Fg("$cyan"), Padding(0, 1)),
						Static("log-stat-warn", "", "  WARN         487", Fg("$yellow"), Padding(0, 1)),
						Static("log-stat-error", "", "  ERROR         93", Fg("$red"), Padding(0, 1)),
						Static("log-stat-debug", "", "  DEBUG          53", Fg("$gray"), Padding(0, 1)),
						HRule("", "thin"),
						Static("log-rate-hdr", "", "Rate (msg/min)", Font("bold"), Fg("$fg1"), Padding(0, 1)),
						Static("log-rate-val", "", "  ≈ 61 / min", Fg("$green"), Padding(0, 1)),
						HRule("", "thin"),
						Static("log-top-hdr", "", "Top Sources", Font("bold"), Fg("$fg1"), Padding(0, 1)),
						Static("log-top1", "", "  api-server  38%", Fg("$fg0"), Padding(0, 1)),
						Static("log-top2", "", "  worker-01   22%", Fg("$fg0"), Padding(0, 1)),
						Static("log-top3", "", "  db-pool     18%", Fg("$fg0"), Padding(0, 1)),
						Static("log-top4", "", "  scheduler    9%", Fg("$fg0"), Padding(0, 1)),
						Static("log-top5", "", "  others      13%", Fg("$gray"), Padding(0, 1)),
					),
				),
			),
		),
	)

	container := w.(core.Container)
	var ticker *time.Ticker
	var paused bool

	container.On(widgets.EvtShow, func(_ core.Widget, _ core.Event, _ ...any) bool {
		for _, sp := range core.FindAll[*widgets.Spinner](container) {
			sp.Start(80 * time.Millisecond)
		}
		ticker = time.NewTicker(1200 * time.Millisecond)
		go func() {
			for range ticker.C {
				if !paused {
					if logText, ok := core.Find(container, "log-text").(*widgets.Text); ok {
						logText.Add(generateLogLine())
					}
				}
			}
		}()
		return true
	})
	container.On(widgets.EvtHide, func(_ core.Widget, _ core.Event, _ ...any) bool {
		for _, sp := range core.FindAll[*widgets.Spinner](container) {
			sp.Stop()
		}
		if ticker != nil {
			ticker.Stop()
		}
		return true
	})

	core.Find(container, "log-pause").On(widgets.EvtActivate, func(w core.Widget, _ core.Event, _ ...any) bool {
		paused = !paused
		if btn, ok := w.(*widgets.Button); ok {
			if paused {
				btn.Set(" ▶ Resume")
			} else {
				btn.Set(" ⏸ Pause")
			}
			btn.Refresh()
		}
		return true
	})
	core.Find(container, "log-clear").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		if logText, ok := core.Find(container, "log-text").(*widgets.Text); ok {
			logText.Clear()
		}
		return true
	})

	return w
}

// ── Screen 4: Process Manager ──────────────────────────────────────────────────

func processScreen(theme *core.Theme) core.Widget {
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

	w := Build(theme,
		VFlex("process-mgr", "", "stretch", 0,
			Padding(1, 2),
			HFlex("proc-hdr", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("proc-title", "", "Process Manager", Font("bold"), Fg("$cyan")),
				Spacer("", Hint(-1, 0)),
				Static("proc-summary", "", "14 processes  ·  13 running  ·  1 stopped", Fg("$gray")),
			),
			HRule("", "thin", Padding(0, 0, 1, 0)),
			HFlex("proc-res-row", "", "stretch", 4,
				Padding(0, 0, 1, 0),
				VFlex("proc-cpu-card", "", "stretch", 0,
					Border("", "round"), Padding(0, 2),
					Static("proc-cpu-lbl", "", "Total CPU", Fg("$gray")),
					Static("proc-cpu-val", "", "31%", Font("bold"), Fg("$yellow")),
					Progress("proc-cpu-total", "", true, Total(100), Value(31)),
					Hint(20, 1),
				),
				VFlex("proc-mem-card", "", "stretch", 0,
					Border("", "round"), Padding(0, 2),
					Static("proc-mem-lbl", "", "Memory Used", Fg("$gray")),
					Static("proc-mem-val", "", "1.43 GB / 10 GB", Font("bold"), Fg("$cyan")),
					Progress("proc-mem-total", "", true, Total(100), Value(41)),
					Hint(20, 1),
				),
				Spacer("", Hint(-1, 0)),
				VFlex("proc-info-card", "", "stretch", 0,
					Border("", "round"), Padding(0, 2),
					Static("proc-load-lbl", "", "Load Average", Fg("$gray")),
					Static("proc-load-val", "", "1.42  1.55  1.61", Font("bold"), Fg("$green")),
					Static("proc-uptime-lbl", "", "Uptime", Fg("$gray")),
					Static("proc-uptime-val", "", "14d 06h 23m", Font("bold"), Fg("$fg1")),
				),
			),
			HFlex("proc-toolbar", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("proc-filter-lbl", "", "Filter:", Fg("$gray")),
				Input("proc-filter", "", []string{"name or PID…"}, Hint(24, 1)),
				Select("proc-sort", "", []string{
					"cpu", "Sort: CPU%",
					"mem", "Sort: Memory",
					"pid", "Sort: PID",
					"name", "Sort: Name",
				}),
				Spacer("", Hint(-1, 0)),
				Button("proc-kill", "", " ✕ Kill"),
				Button("proc-restart", "", " ↺ Restart"),
				Button("proc-detail", "", " ⬡ Details"),
				Button("proc-refresh", "dialog", " ↻ Refresh"),
			),
			VFlex("proc-table-pane", "", "stretch", 0,
				Border("", "round"), Hint(0, -1),
				Static("proc-table-title", "", " Processes", Font("bold"), Bg("$bg2")),
				Table("proc-table", "", widgets.NewArrayTableProvider(procHeaders, procData), false, Hint(0, -1)),
			),
		),
	)

	container := w.(core.Container)

	core.Find(container, "proc-kill").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		procTable := core.Find(container, "proc-table").(*widgets.Table)
		row, _ := procTable.Selected()
		if row >= 0 && row < len(procData) {
			pid := procData[row][0]
			name := strings.TrimSpace(procData[row][1])
			if lbl, ok := core.Find(container, "proc-summary").(*widgets.Static); ok {
				lbl.Set(fmt.Sprintf("Sent SIGTERM to %s (PID %s)", name, pid))
			}
		}
		return true
	})
	core.Find(container, "proc-refresh").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		if lbl, ok := core.Find(container, "proc-summary").(*widgets.Static); ok {
			lbl.Set(fmt.Sprintf("14 processes  ·  refreshed %s", time.Now().Format("15:04:05")))
		}
		return true
	})

	return w
}

// ── Screen 5: Data Entry ───────────────────────────────────────────────────────

func dataEntryScreen(theme *core.Theme) core.Widget {
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
		Company: "Acme Corporation", Name: "Jane Smith",
		Email: "jane.smith@acme.example", Phone: "+1-555-010-2048", Country: "us",
	}
	shipping := ShippingInfo{
		Address: "1600 Amphitheatre Parkway", City: "Mountain View",
		PostalCode: "94043", Priority: "exp", Insurance: true,
	}
	payment := PaymentInfo{
		Method: "invoice", PO: "PO-2026-0042", VATNumber: "US-123456789",
	}

	orderHeaders := []string{"#", "SKU", "Product", "Qty", "Unit Price", "Total"}
	orderData := [][]string{
		{"1", "ZW-2001", "Zeichenwerk Pro License", "5", "  $249.00", " $1,245.00"},
		{"2", "SW-4412", "Support & Maintenance 1yr", "5", "   $49.00", "   $245.00"},
		{"3", "TR-0801", "Training Package (remote)", "2", "  $599.00", " $1,198.00"},
	}

	w := Build(theme,
		VFlex("data-entry", "", "stretch", 0,
			Padding(1, 2),
			HFlex("de-hdr", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("de-title", "", "New Order Entry", Font("bold"), Fg("$cyan")),
				Spacer("", Hint(-1, 0)),
				Static("de-ref", "", "Draft  ·  REF #2026-0099", Fg("$gray")),
			),
			HRule("", "thin", Padding(0, 0, 1, 0)),
			Grid("de-body", "", []int{0}, []int{-1, -1}, false,
				Hint(0, -1),
				Cell(0, 0, 1, 1,
					VFlex("de-form-col", "", "stretch", 0,
						Collapsible("de-cust-section", "", "  ① Customer Information", true,
							HFlex("de-cust-content", "", "stretch", 0,
								Padding(0, 2),
								Form("de-cust-form", "", "", &customer,
									FormGroup("de-cust-grp", "", "", false, 1),
								),
							),
						),
						Spacer("", Hint(0, 1)),
						Collapsible("de-ship-section", "", "  ② Shipping Details", true,
							VFlex("de-ship-content", "", "stretch", 0,
								Padding(0, 2),
								Form("de-ship-form", "", "", &shipping,
									FormGroup("de-ship-grp", "", "", false, 1),
								),
							),
						),
						Spacer("", Hint(0, 1)),
						Collapsible("de-pay-section", "", "  ③ Payment & Terms", false,
							VFlex("de-pay-content", "", "stretch", 0,
								Padding(0, 2),
								Form("de-pay-form", "", "", &payment,
									FormGroup("de-pay-grp", "", "", false, 1),
								),
							),
						),
					),
				),
				Cell(1, 0, 1, 1,
					VFlex("de-items-col", "", "stretch", 0,
						Padding(0, 0, 0, 2),
						Static("de-items-title", "", "Order Items", Font("bold"), Fg("$fg1"), Padding(0, 0, 1, 0)),
						Table("de-items-table", "", widgets.NewArrayTableProvider(orderHeaders, orderData), true, Hint(0, 8)),
						Spacer("", Hint(0, 1)),
						VFlex("de-summary", "", "stretch", 0,
							Border("", "round"), Padding(1, 2),
							Static("de-sum-title", "", "Order Summary", Font("bold"), Fg("$fg1")),
							HRule("", "thin"),
							HFlex("de-sum-row1", "", "stretch", 0,
								Static("de-sum-subtotal-lbl", "", "Subtotal", Fg("$gray")),
								Spacer("", Hint(-1, 0)),
								Static("de-sum-subtotal-val", "", "$2,688.00"),
							),
							HFlex("de-sum-row2", "", "stretch", 0,
								Static("de-sum-tax-lbl", "", "Tax (0%)", Fg("$gray")),
								Spacer("", Hint(-1, 0)),
								Static("de-sum-tax-val", "", "    $0.00"),
							),
							HFlex("de-sum-row3", "", "stretch", 0,
								Static("de-sum-ship-lbl", "", "Shipping (Express)", Fg("$gray")),
								Spacer("", Hint(-1, 0)),
								Static("de-sum-ship-val", "", "   $35.00"),
							),
							HRule("", "thin"),
							HFlex("de-sum-total-row", "", "stretch", 0,
								Static("de-sum-total-lbl", "", "Total", Font("bold"), Fg("$fg0")),
								Spacer("", Hint(-1, 0)),
								Static("de-sum-total-val", "", "$2,723.00", Font("bold"), Fg("$cyan")),
							),
						),
						Spacer("", Hint(0, -1)),
						HFlex("de-actions", "", "end", 2,
							Button("de-btn-draft", "", " ↓ Save Draft"),
							Button("de-btn-cancel", "", " ✕ Cancel"),
							Button("de-btn-submit", "dialog", " ✓ Submit Order"),
						),
						Static("de-status", "", "", Fg("$green"), Padding(1, 0, 0, 0)),
					),
				),
			),
		),
	)

	container := w.(core.Container)

	core.Find(container, "de-btn-submit").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		if lbl, ok := core.Find(container, "de-status").(*widgets.Static); ok {
			lbl.Set(fmt.Sprintf("✓  Order REF #2026-0099 submitted at %s", time.Now().Format("15:04:05")))
		}
		if title, ok := core.Find(container, "de-ref").(*widgets.Static); ok {
			title.Set("Submitted  ·  REF #2026-0099")
		}
		return true
	})
	core.Find(container, "de-btn-draft").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		if lbl, ok := core.Find(container, "de-status").(*widgets.Static); ok {
			lbl.Set(fmt.Sprintf("Draft saved at %s", time.Now().Format("15:04:05")))
		}
		return true
	})
	core.Find(container, "de-btn-cancel").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		if lbl, ok := core.Find(container, "de-status").(*widgets.Static); ok {
			lbl.Set("Changes discarded.")
		}
		return true
	})

	return w
}

// ── Screen 6: Code Editor ──────────────────────────────────────────────────────

func codeEditorScreen(theme *core.Theme) core.Widget {
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

	// Build file tree imperatively — must happen before Build so we can inject via Include
	root := widgets.NewTreeNode("zeichenwerk")
	widgetsNode := widgets.NewTreeNode("widgets")
	widgetsNode.Add(widgets.NewTreeNode("main.go", 0))
	widgetsNode.Add(widgets.NewTreeNode("table.go", 1))
	widgetsNode.Add(widgets.NewTreeNode("flex.go"))
	widgetsNode.Add(widgets.NewTreeNode("editor.go"))
	cmd := widgets.NewTreeNode("cmd")
	cmd.Add(widgets.NewTreeNode("showcase"))
	cmd.Add(widgets.NewTreeNode("demo"))
	root.Add(widgetsNode)
	root.Add(cmd)
	root.Add(widgets.NewTreeNode("README.md", 2))
	root.Add(widgets.NewTreeNode("go.mod"))

	w := Build(theme,
		VFlex("code-editor", "", "stretch", 0,
			Padding(1, 2),
			HFlex("ce-hdr", "", "center", 2,
				Padding(0, 0, 1, 0),
				Static("ce-title", "", "Code Editor", Font("bold"), Fg("$cyan")),
				Spacer("", Hint(-1, 0)),
				Static("ce-status", "", "main.go — Ln 1, Col 1", Fg("$gray")),
				Button("ce-btn-new", "", " + New"),
			),
			HRule("", "thin", Padding(0, 0, 1, 0)),
			Grid("ce-body", "", []int{-1}, []int{26, -1}, false,
				Hint(0, -1), Border("none"),
				Cell(0, 0, 1, 1,
					VFlex("ce-tree-pane", "", "stretch", 0,
						Border("", "round"),
						Static("ce-tree-title", "", " Project", Font("bold"), Bg("$bg2")),
						Tree("ce-tree", "", Hint(0, -1)),
					),
				),
				Cell(1, 0, 1, 1,
					VFlex("ce-edit-col", "", "stretch", 0,
						Hint(0, -1),
						Tabs("ce-tabs", ""),
						Switcher("ce-switcher", "",
							Hint(0, -1),
							Include(func(t *core.Theme) core.Widget {
								return Build(t, Editor("ce-editor-main", "", Hint(0, -1), Content(mainGoContent), LineNumbers(true)))
							}),
							Include(func(t *core.Theme) core.Widget {
								return Build(t, Editor("ce-editor-table", "", Hint(0, -1), Content(tableGoContent), LineNumbers(true)))
							}),
							Include(func(t *core.Theme) core.Widget {
								return Build(t, Viewport("ce-viewport", "", "",
									Styled("ce-preview", "", readmeContent),
								))
							}),
						),
					),
				),
			),
		),
	)

	container := w.(core.Container)

	// Add tab names imperatively
	tabs := core.Find(container, "ce-tabs").(*widgets.Tabs)
	tabs.Add("main.go")
	tabs.Add("table.go")
	tabs.Add("README.md")

	// Populate tree and expand top-level directories
	tree := core.Find(container, "ce-tree").(*widgets.Tree)
	tree.Add(root)
	tree.Expand(root)
	tree.Expand(widgetsNode)

	// Wire tree → tabs + switcher
	switcher := core.Find(container, "ce-switcher").(*widgets.Switcher)
	tree.On(widgets.EvtSelect, func(_ core.Widget, _ core.Event, _ ...any) bool {
		if node := tree.Selected(); node != nil {
			if idx, ok := node.Data().(int); ok {
				tabs.Set(idx)
				switcher.Select(idx)
			}
		}
		return false
	})

	// Wire editor → status bar
	mainEditor := core.Find(container, "ce-editor-main").(*widgets.Editor)
	statusLbl := core.Find(container, "ce-status").(*widgets.Static)
	mainEditor.On(widgets.EvtChange, func(_ core.Widget, _ core.Event, _ ...any) bool {
		cx, cy, _ := mainEditor.Cursor()
		statusLbl.Set(fmt.Sprintf("main.go — Ln %d, Col %d", cy+1, cx+1))
		return false
	})

	// Wire "New" button → popup dialog
	core.Find(container, "ce-btn-new").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
		ui := widgets.FindRoot(container).(*z.UI)
		dlg := widgets.NewDialog("ce-dlg", "dialog", " New File")
		body := Build(theme,
			VFlex("ce-dlg-body", "", "stretch", 0,
				Static("ce-dlg-lbl", "", "Filename:", Fg("$fg1")),
				Input("ce-dlg-input", "", []string{"untitled.go"}, Hint(28, 1)),
				HFlex("ce-dlg-btns", "", "end", 2,
					Padding(1, 0, 0, 0),
					Button("ce-dlg-ok", "dialog", " ✓ Create"),
					Button("ce-dlg-cancel", "", " ✕ Cancel"),
				),
			),
		)
		dlg.Add(body)
		core.Find(body.(core.Container), "ce-dlg-cancel").On(widgets.EvtActivate, func(_ core.Widget, _ core.Event, _ ...any) bool {
			ui.Close()
			return true
		})
		ui.Popup(-1, -1, 0, 0, dlg)
		return true
	})

	return w
}
