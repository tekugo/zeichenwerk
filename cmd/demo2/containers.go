package main

import (
	"time"

	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/values"
	. "github.com/tekugo/zeichenwerk/widgets"
)

var containerEntries = []Entry{
	{
		Category: "Containers",
		Name:     "Box",
		Summary:  "Bordered single-child container with optional title.",
		DocFile:  "box.md",
		DemoFn:   boxDemo,
		Builder: `builder.Box("info", "Information").Border("round").Padding(1).
    Static("info-text", "Boxes hold a single child and add a border with optional title.").
End()`,
		Compose: `compose.Box("info", "", "Information", compose.Border("", "round"), compose.Padding(1),
    compose.Static("info-text", "", "Boxes hold a single child and add a border with optional title."),
)`,
	},
	{
		Category: "Containers",
		Name:     "Card",
		Summary:  "Bordered container with title in border and an optional footer.",
		DocFile:  "card.md",
		DemoFn:   cardDemo,
		Builder: `// First Add() = body, second Add() = footer.
card := NewCard("user-card", "", "Profile")
card.Add(NewStatic("body", "", "Jane Doe — admin@example.com"))
card.Add(NewStatic("foot", "", "Last login: 2026-05-04 09:12"))
builder.Add(card)`,
		Compose: `// Compose API exposes Card via the Card option helper.
compose.Card("user-card", "", "Profile",
    compose.Static("body", "", "Jane Doe — admin@example.com"),
    compose.Static("foot", "", "Last login: 2026-05-04 09:12"),
)`,
	},
	{
		Category: "Containers",
		Name:     "Collapsible",
		Summary:  "Toggleable header that hides/reveals one child.",
		DocFile:  "collapsible.md",
		DemoFn:   collapsibleDemo,
		Builder: `builder.Collapsible("details", "Details (Enter to toggle)", false).
    Static("body", "Hidden body — appears when expanded.").
End()`,
		Compose: `compose.Collapsible("details", "", "Details (Enter to toggle)", false,
    compose.Static("body", "", "Hidden body — appears when expanded."),
)`,
	},
	{
		Category: "Containers",
		Name:     "CRT",
		Summary:  "Matrix-style power-on / power-off animation wrapper.",
		DocFile:  "crt.md",
		DemoFn:   crtDemo,
		Builder: `crt := NewCRT("crt", "")
crt.Add(NewStatic("hello", "", "Hello, terminal."))

builder.VFlex("pane", Stretch, 1).Margin(2).
    Box("crt-box", "CRT").Border("round").Hint(-1, 20).
        Add(crt).
    End().
    HFlex("ctl", Center, 2).Padding(1, 0, 0, 0).
        Button("start", "▶ Power on").
        Button("stop",  "⏻ Power off").
    End().
End()

builder.Find("start").(*Button).On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
    crt.Start(30 * time.Millisecond)              // power-on (no-op while running)
    return true
})
builder.Find("stop").(*Button).On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
    crt.PowerOff(30*time.Millisecond, func() {})  // contracting animation; pass ui.Quit to exit
    return true
})`,
		Compose: `compose.VFlex("pane", "", core.Stretch, 1, compose.Margin(2),
    compose.Box("crt-box", "", "CRT", compose.Border("", "round"), compose.Hint(-1, 20),
        compose.CRT("crt", "",
            compose.Static("hello", "", "Hello, terminal."),
        ),
    ),
    compose.HFlex("ctl", "", core.Center, 2, compose.Padding(1, 0, 0, 0),
        compose.Button("start", "", "▶ Power on"),
        compose.Button("stop",  "", "⏻ Power off"),
    ),
)`,
	},
	{
		Category: "Containers",
		Name:     "Dialog",
		Summary:  "Single-child container shown as a popup layer.",
		DocFile:  "dialog.md",
		DemoFn:   dialogDemo,
		Builder: `dialog := ui.NewBuilder().
    Dialog("confirm", "Confirm").Padding(1).
        VFlex("body", Stretch, 1).
            Static("", "Save changes before closing?").
            HFlex("buttons", End, 2).
                Button("ok", "Save").
                Button("no", "Discard").
            End().
        End().
    End().
    Container()
ui.Popup(-1, -1, 0, 0, dialog)`,
		Compose: `dialog := compose.Build(theme,
    compose.Dialog("confirm", "", "Confirm", compose.Padding(1),
        compose.VFlex("body", "", core.Stretch, 1,
            compose.Static("", "", "Save changes before closing?"),
        ),
    ),
).(core.Container)
ui.Popup(-1, -1, 0, 0, dialog)`,
	},
	{
		Category: "Containers",
		Name:     "Flex",
		Summary:  "Linear layout: stack children horizontally or vertically with alignment + spacing.",
		DocFile:  "flex.md",
		DemoFn:   flexDemo,
		Builder: `builder.HFlex("row", Stretch, 1).
    Static("a", "Left").
    Static("b", "Centre").Hint(-1, 0).
    Static("c", "Right").
End()
// VFlex(...) is the vertical-orientation shortcut.`,
		Compose: `compose.HFlex("row", "", core.Stretch, 1,
    compose.Static("a", "", "Left"),
    compose.Static("b", "", "Centre", compose.Hint(-1, 0)),
    compose.Static("c", "", "Right"),
)`,
	},
	{
		Category: "Containers",
		Name:     "Form / FormGroup",
		Summary:  "Reflection-bound form. Fields are generated from struct tags.",
		DocFile:  "form.md",
		DemoFn:   formDemo,
		Builder: `data := struct {
    Username string ` + "`width:\"30\"`" + `
    Password string ` + "`control:\"password\" width:\"30\"`" + `
}{Username: "admin"}

builder.Form("login", "Sign in", &data).
    Group("login-group", "", "", false, 1).Border("", "round").
    End().
End()`,
		Compose: `data := struct {
    Username string ` + "`width:\"30\"`" + `
    Password string ` + "`control:\"password\" width:\"30\"`" + `
}{Username: "admin"}

compose.Form("login", "", "Sign in", &data,
    compose.FormGroup("login-group", "", "", false, 1, compose.Border("", "round")),
)`,
	},
	{
		Category: "Containers",
		Name:     "Grid",
		Summary:  "Table-style layout. Cells specify (x, y, w, h) span and individual sizes.",
		DocFile:  "grid.md",
		DemoFn:   gridDemo,
		Builder: `builder.Grid("layout", 3, 3, true).Columns(20, -1, 12).Rows(3, -1, 3).
    Cell(0, 0, 3, 1).Static("hdr", "Header — spans 3 columns").
    Cell(0, 1, 1, 1).Static("nav", "Nav").
    Cell(1, 1, 2, 1).Static("body", "Body — fills remaining space").
    Cell(0, 2, 3, 1).Static("ftr", "Footer — spans 3 columns").
End()`,
		Compose: `compose.Grid("layout", "", []int{3, -1, 3}, []int{20, -1, 12}, true,
    compose.Cell(0, 0, 3, 1, compose.Static("hdr", "", "Header — spans 3 columns")),
    compose.Cell(0, 1, 1, 1, compose.Static("nav", "", "Nav")),
    compose.Cell(1, 1, 2, 1, compose.Static("body", "", "Body — fills remaining space")),
    compose.Cell(0, 2, 3, 1, compose.Static("ftr", "", "Footer — spans 3 columns")),
)`,
	},
	{
		Category: "Containers",
		Name:     "Grow",
		Summary:  "Animated reveal wrapper — its child slides into view.",
		DocFile:  "grow.md",
		DemoFn:   growDemo,
		Builder: `g := NewGrow("grow", "", false) // false = vertical reveal
g.Add(NewStatic("body", "", "Revealed!"))
g.Start(20 * time.Millisecond)
builder.Add(g)`,
		Compose: `compose.Grow("grow", "", false,
    compose.Static("body", "", "Revealed!"),
)`,
	},
	{
		Category: "Containers",
		Name:     "Switcher",
		Summary:  "Shows exactly one of its children at a time.",
		DocFile:  "switcher.md",
		DemoFn:   switcherDemo,
		Builder: `builder.Tabs("tabs", "First", "Second", "Third").
Switcher("pages", true). // connect=true wires Tabs activation to Select
    Static("p1", "First page").
    Static("p2", "Second page").
    Static("p3", "Third page").
End()`,
		Compose: `compose.Tabs("tabs", "", []string{"First", "Second", "Third"}),
compose.Switcher("pages", "",
    compose.Static("p1", "", "First page"),
    compose.Static("p2", "", "Second page"),
    compose.Static("p3", "", "Third page"),
)`,
	},
	{
		Category: "Containers",
		Name:     "Viewport",
		Summary:  "Scrollable container for a single, oversized child.",
		DocFile:  "viewport.md",
		DemoFn:   viewportDemo,
		Builder: `builder.Viewport("viewport", "Scroll me").Border("thin").Hint(40, 8).
    Text("body", longLines, false, 0).
End()`,
		Compose: `compose.Viewport("viewport", "", "Scroll me",
    compose.Border("", "thin"), compose.Hint(40, 8),
    compose.Text("body", "", longLines, false, 0),
)`,
	},
}

// ── Demo functions ────────────────────────────────────────────────────────────

func boxDemo(b *Builder) {
	b.VFlex("box-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Containers that hold one child and draw a border around it.").Padding(0, 0, 1, 0).
		Box("simple", "Simple Box").Border("", "thin").Padding(1).
		Static("c1", "Thin border, default padding.").
		End().
		Box("double", "Double Border").Border("", "double").Padding(1).
		Static("c2", "Double-line border style.").
		End().
		Box("round", "Round Border").Border("", "round").Padding(2).
		Static("c3", "Rounded corners with extra padding.").
		End().
		End()
}

func cardDemo(b *Builder) {
	card1 := NewCard("user-card", "", "Profile")
	_ = card1.Add(NewStatic("body", "", "Jane Doe — admin@example.com"))
	_ = card1.Add(NewStatic("foot", "", "Last login: 2026-05-04 09:12"))

	card2 := NewCard("metric-card", "", "Uptime")
	_ = card2.Add(NewStatic("body2", "", "14d 7h 22m"))
	_ = card2.Add(NewStatic("foot2", "", "Restart scheduled Sun 02:00 UTC"))

	b.VFlex("card-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Cards have a title in the border and an optional footer below the body.").
		Padding(0, 0, 1, 0).
		Add(card1).End().
		Add(card2).End().
		End()
}

func collapsibleDemo(b *Builder) {
	b.VFlex("col-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Click the header or press Enter to toggle. → expands, ← collapses.").
		Padding(0, 0, 1, 0).
		Collapsible("first", "Section A (open)", true).
		Static("a-body", "Body A — shown initially.").Padding(0, 1).
		End().
		Collapsible("second", "Section B (closed)", false).
		List("b-items", "Apple", "Banana", "Cherry", "Date").
		End().
		Collapsible("third", "Section C (closed)", false).
		Static("c-body", "Body C — open me to read.").Padding(0, 1).
		End().
		End()
}

func crtDemo(b *Builder) {
	crt := NewCRT("crt-demo-inner", "")
	_ = crt.Add(NewStyled("crt-content", "",
		"```\n"+
			"╔══════════════════════════════════════╗\n"+
			"║                                      ║\n"+
			"║         CRT — power-on               ║\n"+
			"║         animation wrapper            ║\n"+
			"║                                      ║\n"+
			"║   Matrix-style fade-in / fade-out    ║\n"+
			"║   effect over any child subtree.     ║\n"+
			"║                                      ║\n"+
			"║   Press [Start] to replay.           ║\n"+
			"║                                      ║\n"+
			"╚══════════════════════════════════════╝\n"+
			"```\n"))

	b.VFlex("crt-demo", Stretch, 1).Margin(2).
		Static("desc", "CRT wraps any subtree with a Matrix-style power-on / power-off effect. Use the buttons to replay either animation.").
		Padding(0, 0, 1, 0).
		Box("crt-box", "CRT").Border("round").Hint(-1, 20).
		Add(crt).
		End().
		End().
		HFlex("crt-ctl", Center, 2).Padding(1, 0, 0, 0).
		Button("crt-start", "▶ Power on").
		Button("crt-stop", "⏻ Power off").
		End().
		End()

	b.Find("crt-start").(*Button).On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		crt.Start(30 * time.Millisecond)
		return true
	})
	b.Find("crt-stop").(*Button).On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		crt.PowerOff(30*time.Millisecond, func() {})
		return true
	})
	pane := b.Find("crt-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		crt.Start(30 * time.Millisecond)
		return true
	})
}

func dialogDemo(b *Builder) {
	b.VFlex("dialog-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Dialog is shown as a popup layer via ui.Popup. Press the button below to open one.").
		Padding(0, 0, 1, 0).
		Button("open-dialog", "Open dialog").
		Static("dialog-status", "").Padding(1, 0, 0, 0).
		End()

	btn := b.Find("open-dialog").(*Button)
	btn.On(EvtActivate, func(w Widget, _ Event, _ ...any) bool {
		ui := FindRoot(w).(*UI)
		dlg := ui.NewBuilder().
			Dialog("confirm", "Confirm").Padding(1).
			VFlex("body", Stretch, 1).
			Static("", "Save changes before closing?").
			HFlex("buttons", End, 2).
			Button("ok", "Save").
			Button("no", "Discard").
			End().
			End().
			End().
			Container()
		Find(dlg, "ok").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
			ui.Close()
			Update(ui, "dialog-status", "Saved.")
			return true
		})
		Find(dlg, "no").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
			ui.Close()
			Update(ui, "dialog-status", "Discarded.")
			return true
		})
		ui.Popup(-1, -1, 0, 0, dlg)
		return true
	})
}

func flexDemo(b *Builder) {
	b.VFlex("flex-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "HFlex stacks children left-to-right. VFlex stacks them top-to-bottom.").
		Padding(0, 0, 1, 0).
		Static("h-label", "HFlex with Stretch alignment, spacing 1:").
		HFlex("h", Stretch, 1).Border("", "thin").Padding(0, 1).
		Static("a", "[ A ]").
		Static("b", "[ B ]").
		Static("c", "[ C — fills ]").Hint(-1, 0).
		Static("d", "[ D ]").
		End().
		Spacer().Hint(0, 1).
		Static("v-label", "VFlex with Center alignment, spacing 0:").
		VFlex("v", Center, 0).Border("", "thin").Padding(1).
		Static("v1", "Top").
		Static("v2", "Middle").
		Static("v3", "Bottom").
		End().
		End()
}

func formDemo(b *Builder) {
	user := struct {
		ID       string `readonly:"true"`
		Name     string `width:"30"`
		Email    string `label:"E-Mail" width:"30"`
		Role     string `control:"select" options:"admin,Admin,editor,Editor,viewer,Viewer"`
		Active   bool
		Password string `control:"password" width:"30"`
	}{
		ID:     "JD",
		Name:   "Jane Doe",
		Email:  "jane@example.com",
		Role:   "admin",
		Active: true,
	}

	b.VFlex("form-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Form binds a Go struct via reflection. Tags drive labels, widths, and control types.").
		Padding(0, 0, 1, 0).
		Form("user-form", "User", &user).
		Group("user-group", "", "", false, 1).Border("", "round").Padding(1).
		End().
		End().
		End()
}

func gridDemo(b *Builder) {
	b.VFlex("grid-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Grid arranges children by (x, y, w, h) cell coordinates with explicit row/column sizes.").
		Padding(0, 0, 1, 0).
		Grid("layout", 3, 3, true).Border("", "round").Columns(20, -1, 14).Rows(3, -1, 3).Hint(0, -1).
		Cell(0, 0, 3, 1).Static("hdr", "Header — spans 3 columns").
		Cell(0, 1, 1, 1).Static("nav", "Nav").
		Cell(1, 1, 2, 1).Static("body", "Body — fills remaining space").
		Cell(0, 2, 3, 1).Static("ftr", "Footer — spans 3 columns").
		End().
		End()
}

func growDemo(b *Builder) {
	g := NewGrow("grow-anim", "", false)
	g.SetHint(-1, 6)
	_ = g.Add(NewStyled("grow-body", "", "**Hello — I appeared.**\n\nGrow reveals its child cell-by-cell."))
	b.VFlex("grow-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Grow animates its single child into view. Open the page to replay.").
		Padding(0, 0, 1, 0).
		Box("grow-box", "Grow").Border("round").
		Add(g).
		End().
		End().
		End()
	pane := b.Find("grow-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		g.Start(20 * time.Millisecond)
		return true
	})
}

func switcherDemo(b *Builder) {
	b.VFlex("switcher-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Switcher shows one child at a time. Combine with Tabs to navigate between panes.").
		Padding(0, 0, 1, 0).
		Tabs("sw-tabs", "Alpha", "Beta", "Gamma").
		Switcher("sw-pages", true).Border("", "thin").Hint(0, 6).
		VFlex("p1", Center, 0).Static("p1-text", "Pane Alpha — selected by default.").End().
		VFlex("p2", Center, 0).Static("p2-text", "Pane Beta — second tab.").End().
		VFlex("p3", Center, 0).Static("p3-text", "Pane Gamma — third tab.").End().
		End().
		End()
}

func viewportDemo(b *Builder) {
	lines := make([]string, 60)
	for i := range lines {
		lines[i] = "Line " + itoa(i+1) + " — scroll the viewport with arrow keys."
	}
	b.VFlex("viewport-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Viewport lets a smaller window scroll over a larger child. Useful for overflowing text or canvases.").
		Padding(0, 0, 1, 0).
		Viewport("viewport", "Scrollable").Border("", "thin").Hint(0, -1).
		Text("body", lines, false, 0).
		End().
		End()
}
