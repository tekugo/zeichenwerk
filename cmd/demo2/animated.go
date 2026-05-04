package main

import (
	"time"

	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/widgets"
)

var animatedEntries = []Entry{
	{
		Category: "Animated",
		Name:     "Clock",
		Summary:  "Live wall-clock display built on Digits.",
		DocFile:  "clock.md",
		DemoFn:   clockDemo,
		Builder: `builder.Clock("clock", time.Second, "15:04:05") // tick interval, format
// Start the ticker after the UI exists (e.g. in EvtShow):
clock := builder.Find("clock").(*Clock)
clock.Start()    // uses the interval passed to NewClock`,
		Compose: `compose.Clock("clock", "", time.Second, "15:04:05", "")
// After build:
clock := core.Find(ui, "clock").(*widgets.Clock)
clock.Start()`,
	},
	{
		Category: "Animated",
		Name:     "Marquee",
		Summary:  "Continuously-scrolling text ticker; pauses while hovered.",
		DocFile:  "marquee.md",
		DemoFn:   marqueeDemoFn,
		Builder: `builder.Marquee("ticker").Hint(-1, 1)
m := builder.Find("ticker").(*Marquee)
m.SetText("Status — all systems operational. CPU 4% MEM 1.2 GB NET ↑ 0.8 MB/s")
m.Start(80 * time.Millisecond)`,
		Compose: `compose.Marquee("ticker", "", compose.Hint(-1, 1))
// then: m := core.Find(ui, "ticker").(*widgets.Marquee); m.SetText(…); m.Start(…)`,
	},
	{
		Category: "Animated",
		Name:     "Progress",
		Summary:  "Determinate or indeterminate progress bar; horizontal or vertical.",
		DocFile:  "progress.md",
		DemoFn:   progressDemo,
		Builder: `p := NewProgress("progress", "", true) // horizontal
p.SetTotal(100)
p.Set(42)
builder.Add(p)`,
		Compose: `compose.Progress("progress", "", true)`,
	},
	{
		Category: "Animated",
		Name:     "Scanner",
		Summary:  "Back-and-forth scanning animation with a fading trail.",
		DocFile:  "scanner.md",
		DemoFn:   scannerDemo,
		Builder: `builder.Scanner("scan", 12, "blocks") // width, char style
s := builder.Find("scan").(*Scanner)
s.Start(50 * time.Millisecond)`,
		Compose: `compose.Scanner("scan", "", 12, "blocks")`,
	},
	{
		Category: "Animated",
		Name:     "Shimmer",
		Summary:  "Static text with a sweeping highlight band — stepped or gradient.",
		DocFile:  "shimmer.md",
		DemoFn:   shimmerDemoFn,
		Builder: `builder.Shimmer("title").Hint(-1, 1)
sh := builder.Find("title").(*Shimmer)
sh.SetText("Loading…")
sh.SetBandWidth(10).SetEdgeWidth(4).SetGradient(true)
sh.Start(40 * time.Millisecond)`,
		Compose: `compose.Shimmer("title", "", compose.Hint(-1, 1))`,
	},
	{
		Category: "Animated",
		Name:     "Spinner",
		Summary:  "Animated loading indicator with built-in character sequences.",
		DocFile:  "spinner.md",
		DemoFn:   spinnerDemo,
		Builder: `builder.Spinner("loader", Spinners["dots"])
sp := builder.Find("loader").(*Spinner)
sp.Start(100 * time.Millisecond)`,
		Compose: `compose.Spinner("loader", "", widgets.Spinners["dots"])`,
	},
	{
		Category: "Animated",
		Name:     "Typewriter",
		Summary:  "Character-by-character text reveal with optional cursor and repeat.",
		DocFile:  "typewriter.md",
		DemoFn:   typewriterDemoFn,
		Builder: `builder.Typewriter("tw")
tw := builder.Find("tw").(*Typewriter)
tw.SetText("Initialising subsystems…")
tw.Start(30 * time.Millisecond)`,
		Compose: `compose.Typewriter("tw", "")`,
	},
}

// ── Demo functions ────────────────────────────────────────────────────────────

func clockDemo(b *Builder) {
	b.VFlex("clock-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Clock builds on Digits and ticks once per interval. Call Start() — the Builder method does not auto-start it.").
		Padding(0, 0, 1, 0).
		HFlex("row", Center, 0).
		Clock("clock-large", time.Second, "15:04:05").
		End().
		Spacer().Hint(0, 1).
		HFlex("row2", Center, 0).
		Clock("clock-small", time.Second, "Mon, 02 Jan 2006  15:04 MST").
		End().
		End()

	pane := b.Find("clock-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		for _, c := range FindAll[*Clock](pane) {
			c.Start()
		}
		return true
	})
	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		for _, c := range FindAll[*Clock](pane) {
			c.Stop()
		}
		return true
	})
}

func marqueeDemoFn(b *Builder) {
	b.VFlex("marq-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Marquee scrolls long text in a single row. Hover with the mouse to pause.").
		Padding(0, 0, 1, 0).
		Marquee("ticker").Hint(-1, 1).
		End()
	m := b.Find("ticker").(*Marquee)
	m.SetText("Status — all systems operational.   CPU 4%   MEM 1.2 GB   NET ↑ 0.8 MB/s ↓ 2.1 MB/s   DISK 42%   TEMP 38°C   UPTIME 14d 7h")
	pane := b.Find("marq-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		m.Start(80 * time.Millisecond)
		return true
	})
	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		m.Stop()
		return true
	})
}

func progressDemo(b *Builder) {
	indet := NewProgress("indet", "", true)

	p25 := NewProgress("p25", "", true)
	p25.SetTotal(100)
	p25.Set(25)

	p50 := NewProgress("p50", "", true)
	p50.SetTotal(100)
	p50.Set(50)

	p75 := NewProgress("p75", "", true)
	p75.SetTotal(100)
	p75.Set(75)

	p100 := NewProgress("p100", "", true)
	p100.SetTotal(100)
	p100.Set(100)

	b.VFlex("prog-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Determinate (with total>0) and indeterminate (total=0) progress bars.").
		Padding(0, 0, 1, 0).
		Static("indet-label", "Indeterminate (no total):").
		Add(indet).
		Spacer().Hint(0, 1).
		Static("d-label", "25 / 50 / 75 / 100 %:").
		Add(p25).
		Spacer().Hint(0, 1).
		Add(p50).
		Spacer().Hint(0, 1).
		Add(p75).
		Spacer().Hint(0, 1).
		Add(p100).
		End()

	_ = indet
}

func scannerDemo(b *Builder) {
	b.VFlex("scanner-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Back-and-forth scanning animation with a fading trail. Built-in styles: blocks, circles, diamonds.").
		Padding(0, 0, 1, 0).
		VFlex("rows", Center, 1).
		Scanner("s1", 12, "blocks").
		Scanner("s2", 12, "circles").
		Scanner("s3", 12, "diamonds").
		End().
		End()
	pane := b.Find("scanner-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		for _, s := range FindAll[*Scanner](pane) {
			s.Start(50 * time.Millisecond)
		}
		return true
	})
	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		for _, s := range FindAll[*Scanner](pane) {
			s.Stop()
		}
		return true
	})
}

func shimmerDemoFn(b *Builder) {
	b.VFlex("shim-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Sweeping highlight band over static text. Toggle SetGradient for a smoother transition.").
		Padding(0, 0, 1, 0).
		Static("l1", "Stepped band:").
		Shimmer("s1").Hint(-1, 1).
		Spacer().Hint(0, 1).
		Static("l2", "Gradient band:").
		Shimmer("s2").Hint(-1, 1).
		Spacer().Hint(0, 1).
		Static("l3", "Multi-line gradient:").
		Shimmer("s3").Hint(-1, 3).
		End()

	s1 := b.Find("s1").(*Shimmer)
	s1.SetText("Analysing codebase…  Status: all systems operational.")
	s1.SetBandWidth(10).SetEdgeWidth(5)

	s2 := b.Find("s2").(*Shimmer)
	s2.SetText("Analysing codebase…  Status: all systems operational.")
	s2.SetBandWidth(10).SetEdgeWidth(5).SetGradient(true)

	s3 := b.Find("s3").(*Shimmer)
	s3.SetText("Searching for references…\nProcessing matched files…\nUpdating cross-references…")
	s3.SetBandWidth(10).SetEdgeWidth(5).SetGradient(true)

	pane := b.Find("shim-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		for _, s := range FindAll[*Shimmer](pane) {
			s.Start(40 * time.Millisecond)
		}
		return true
	})
	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		for _, s := range FindAll[*Shimmer](pane) {
			s.Stop()
		}
		return true
	})
}

func spinnerDemo(b *Builder) {
	b.Box("spinner-demo", "Spinners").Border("", "round").Margin(1).Padding(1, 5).
		HFlex("row", Start, 2).
		Spinner("s-bar", Spinners["bar"]).
		Spinner("s-dot", Spinners["dot"]).
		Spinner("s-dots", Spinners["dots"]).
		Spinner("s-arrow", Spinners["arrow"]).
		Spinner("s-circle", Spinners["circle"]).
		Spinner("s-bounce", Spinners["bounce"]).
		Spinner("s-braille", Spinners["braille"]).
		End().
		End()
	pane := b.Find("spinner-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		for _, s := range FindAll[*Spinner](pane) {
			s.Start(100 * time.Millisecond)
		}
		return true
	})
	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		for _, s := range FindAll[*Spinner](pane) {
			s.Stop()
		}
		return true
	})
}

func typewriterDemoFn(b *Builder) {
	phrases := []string{
		"Initialising subsystems…",
		"Loading configuration…",
		"Connecting to services…",
		"All systems operational.",
	}
	idx := 0

	b.VFlex("tw-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Typewriter reveals text character by character. EvtActivate fires when complete.").
		Padding(0, 0, 1, 0).
		Typewriter("tw").
		Spacer().Hint(0, 1).
		HFlex("ctl", Center, 4).
		Checkbox("repeat", "Repeat", false).
		Checkbox("cursor", "Show cursor", true).
		Button("restart", "Restart").
		End().
		End()

	tw := b.Find("tw").(*Typewriter)
	tw.SetText(phrases[idx])
	tw.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		idx = (idx + 1) % len(phrases)
		tw.SetText(phrases[idx])
		tw.Start(30 * time.Millisecond)
		return true
	})

	b.Find("repeat").(*Checkbox).On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if v, ok := data[0].(bool); ok {
			tw.SetRepeat(v)
		}
		return true
	})
	b.Find("cursor").(*Checkbox).On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
		if v, ok := data[0].(bool); ok {
			tw.SetCursor(v)
		}
		return true
	})
	b.Find("restart").(*Button).On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		tw.Stop()
		tw.Reset()
		tw.Start(30 * time.Millisecond)
		return true
	})

	pane := b.Find("tw-demo").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		tw.Reset()
		tw.Start(30 * time.Millisecond)
		return true
	})
	pane.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		tw.Stop()
		return true
	})
}
