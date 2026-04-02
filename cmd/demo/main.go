package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	. "github.com/tekugo/zeichenwerk"
)

func parseFlags() (*Theme, bool, bool, bool) {
	t := flag.String("t", "tokyo", "Theme: midnight, tokyo, nord, gruvbox-dark, gruvbox-light, lipstick")
	dbg := flag.Bool("debug", false, "Start in debug mode")
	dmp := flag.Bool("dump", false, "Dump widget hierarchy to stdout and exit")
	dmpV := flag.Bool("dump-verbose", false, "Dump widget hierarchy with style details to stdout and exit")
	flag.Parse()
	var theme *Theme
	switch *t {
	case "midnight":
		theme = MidnightNeonTheme()
	case "nord":
		theme = NordTheme()
	case "gruvbox-dark":
		theme = GruvboxDarkTheme()
	case "gruvbox-light":
		theme = GruvboxLightTheme()
	case "lipstick":
		theme = LipstickTheme()
	default:
		theme = TokyoNightTheme()
	}
	return theme, *dbg, *dmp, *dmpV
}

// main function
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

// Create the terminal UI.
func createUI(theme *Theme) *UI {
	ui := NewBuilder(theme).
		Flex("main", false, "stretch", 0).
		Flex("header", true, "stretch", 2).
		Static("title", "Zeichenwerk Demo").
		Static("subtitle", "A terminal UI framework").
		End().
		Grid("content", 2, 2, true).Hint(0, -1).Columns(32, -1).Rows(-1, 10).
		Cell(0, 0, 1, 2).
		List("navigation", "Box", "Canvas", "Checkbox", "Collapsible", "Deck", "Digits", "Editor", "Form", "Grid", "Progress", "Scanner", "Select", "Spinner", "Styled", "Table", "Tabs", "Terminal", "Tree FS", "Typeahead", "Viewport", "Dialog", "Confirm", "Prompt").
		Cell(1, 0, 1, 1).
		Switcher("switcher", false).
		With(box).
		With(canvas).
		With(checkbox).
		With(collapsibleDemo).
		With(func(b *Builder) { deckDemo(b, theme) }).
		With(digits).
		With(editor).
		With(form).
		With(grid).
		With(progress).
		With(scanner).
		With(dropdown).
		With(spinner).
		With(styled).
		With(table).
		With(tabs).
		With(terminalDemo).
		With(treeFSDemo).
		With(typeaheadDemo).
		With(viewport).
		End().
		Cell(1, 1, 1, 1).
		Flex("debug-log-pane", false, "stretch", 0).Hint(0, 10).
		Static("debug-log-title", "Debug Log").Background("green").
		Text("debug-log", []string{"Hello, World!"}, true, 100).Hint(0, -1).
		End().
		End().
		Flex("footer", true, "center", 0).
		Shortcuts("footer-shortcuts", "↑↓", "navigate", "Enter", "select", "i", "debug", "t", "theme", "q", "quit").
		Spacer().Hint(-1, 0).
		Static("theme-label", " Theme: ").
		Select("theme-select", "tokyo", "Tokyo Night", "gruvbox-dark", "Gruvbox Dark", "gruvbox-light", "Gruvbox Light", "nrrd", "Nord", "neon", "Midnight Neon").
		End().
		Build()

	themes := map[string]*Theme{
		"tokyo":         TokyoNightTheme(),
		"gruvbox-dark":  GruvboxDarkTheme(),
		"gruvbox-light": GruvboxLightTheme(),
		"nrrd":          NordTheme(),
		"neon":          MidnightNeonTheme(),
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

	switcher := Find(ui, "switcher").(*Switcher)
	Find(ui, "navigation").On(EvtActivate, func(_ Widget, event Event, data ...any) bool {
		if len(data) == 1 {
			if selected, ok := data[0].(int); ok {
				if selected < len(switcher.Children()) {
					switcher.Select(selected)
				} else {
					switch selected {
					case 20:
						dialog := ui.NewBuilder().
							Dialog("dialog", "Test Dialog").
							Class("dialog").
							Flex("dialog-content", false, "stretch", 1).
							Static("", "Do you really want to do this?").
							Flex("dialog-buttons", true, "end", 2).
							Button("btn-yes", "YES").
							Button("btn-no", "NO").
							End().
							End().
							Class("").
							Container()
						Find(dialog, "btn-yes").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
							ui.Close()
							return true
						})
						Find(dialog, "btn-no").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
							ui.Close()
							return true
						})
						ui.Popup(-1, -1, 0, 0, dialog)
					case 21:
						ui.Confirm("Confirm Action", "Do you really want to do this?",
							func() {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Confirm → OK")
								}
							},
							func() {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Confirm → Cancel")
								}
							},
						)
					case 22:
						ui.Prompt("Enter Value", "Please enter a value:",
							func(text string) {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Prompt → " + text)
								}
							},
							func() {
								if log, ok := Find(ui, "debug-log").(*Text); ok {
									log.Add("Prompt → Cancel")
								}
							},
						)
					}
				}
			}
		}
		return true
	})

	return ui
}

// Box demo
func box(builder *Builder) {
	builder.Flex("box-demo", false, "stretch", 1).Padding(1).
		Static("box-title", "Box Widget Demo").Padding(0, 0, 1, 0).
		HRule("thin").
		Flex("box-examples", false, "stretch", 1).
		Box("simple-box", "Simple Box").Padding(1).
		Static("box-content1", "This is content inside a simple box widget.").
		End().
		Box("styled-box", "Styled Box").Padding(1).Border("", "double").
		Static("box-content2", "This box has a double border style.").
		End().
		Box("padded-box", "Padded Box").Padding(2).Border("", "round").
		Static("box-content3", "This box has extra padding and rounded borders.").
		End().
		End().
		HRule("double").
		Static("box-info", "Boxes are containers that can hold a single child widget with optional borders and titles.").Padding(1, 0, 0, 0).
		End()
}

// Canvas demo
func canvas(builder *Builder) {
	// Create a 40x20 canvas with a simple pattern
	c := NewCanvas("demo-canvas", "", 1, 40, 20)

	// Set styles for normal and insert modes
	normalStyle := NewStyle("").WithColors("white", "black").WithCursor("block")
	insertStyle := NewStyle("").WithColors("cyan", "black").WithCursor("bar")

	c.SetStyle("", normalStyle)
	c.SetStyle(":insert", insertStyle)

	// Fill with empty cells
	c.Fill("", normalStyle)

	// Draw a border on the edges
	borderStyle := NewStyle("").WithColors("yellow", "black")
	c.SetCell(0, 0, "+", borderStyle)
	c.SetCell(39, 0, "+", borderStyle)
	c.SetCell(0, 19, "+", borderStyle)
	c.SetCell(39, 19, "+", borderStyle)
	// Horizontal lines
	for x := 1; x < 39; x++ {
		c.SetCell(x, 0, "-", borderStyle)
		c.SetCell(x, 19, "-", borderStyle)
	}
	// Vertical lines
	for y := 1; y < 19; y++ {
		c.SetCell(0, y, "|", borderStyle)
		c.SetCell(39, y, "|", borderStyle)
	}

	// Add some instruction text
	titleStyle := NewStyle("").WithColors("green", "black")
	c.SetCell(2, 2, "Canvas Widget Demo", titleStyle)

	infoStyle := NewStyle("").WithColors("gray", "black")
	info := "NORMAL mode: hjkl/arrows move, 'i' or 'a' enters INSERT mode, ESC returns"
	for i, ch := range info {
		if i < 38 {
			c.SetCell(2+i, 4, string(ch), infoStyle)
		}
	}

	info2 := "INSERT mode: type to insert chars, arrows still move, ESC returns"
	for i, ch := range info2 {
		if i < 38 {
			c.SetCell(2+i, 5, string(ch), infoStyle)
		}
	}

	// Add it to the builder
	builder.Flex("canvas-demo", false, "stretch", 1).Padding(1).
		Static("canvas-title", "Canvas Widget (press 'i' to start editing)").Padding(0, 0, 1, 0).
		Add(c).
		End()
}

// Checkbox demo
func checkbox(builder *Builder) {
	builder.Flex("checkbox-demo", false, "stretch", 1).Padding(1, 2).
		Static("checkbox-title", "Checkbox Widget Demo").Padding(0, 0, 1, 0).
		Static("checkbox-info", "Checkboxes toggle between checked and unchecked states.").Padding(0, 0, 1, 0).
		Checkbox("cb1", "Enable notifications", false).
		Checkbox("cb2", "Remember login", true).
		Checkbox("cb3", "Auto-save documents", false).
		Checkbox("cb4", "Show hidden files", true).
		Checkbox("cb5", "I agree to the terms and conditions", false).
		Static("checkbox-status", "Toggle checkboxes with Space or Enter key!").Padding(1, 0, 0, 0).
		End()

	container := builder.Container()
	for i := 1; i <= 5; i++ {
		cbID := fmt.Sprintf("cb%d", i)
		if cb := Find(container, cbID); cb != nil {
			cb.On(EvtChange, func(_ Widget, event Event, data ...any) bool {
				checked := data[0].(bool)
				if statusLabel := Find(container, "checkbox-status"); statusLabel != nil {
					if label, ok := statusLabel.(*Static); ok {
						var name string
						switch cbID {
						case "cb1":
							name = "Notifications"
						case "cb2":
							name = "Remember login"
						case "cb3":
							name = "Auto-save"
						case "cb4":
							name = "Show hidden"
						case "cb5":
							name = "Terms agreed"
						}
						label.SetText(fmt.Sprintf("%s: %v", name, checked))
					}
				}
				return true
			})
		}
	}
}

// Collapsible demo
func collapsibleDemo(builder *Builder) {
	builder.Flex("collapsible-demo", false, "stretch", 1).Padding(1, 2).
		Static("collapsible-title", "Collapsible Widget Demo").Padding(0, 0, 1, 0).
		Static("collapsible-info", "Click the header or press Enter/Space to toggle. → expands, ← collapses.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Collapsible("col-basic", "Basic section (starts expanded)", true).
		Flex("col-basic-content", false, "stretch", 1).Padding(0, 1).
		Static("", "This is the body of the first collapsible.").
		Static("", "It can contain any widget — here a few statics.").
		Static("", "Collapse me with ← or by clicking the header.").
		End().
		End().
		Collapsible("col-list", "List section (starts collapsed)", false).
		List("col-list-items", "Alpha", "Beta", "Gamma", "Delta", "Epsilon").
		End().
		Collapsible("col-inputs", "Input section (starts collapsed)", false).
		Flex("col-inputs-content", false, "stretch", 1).Padding(0, 1).
		Static("", "Name:").
		Input("col-name", "").
		Static("", "Email:").
		Input("col-email", "").
		End().
		End().
		Static("col-status", "").Padding(1, 0, 0, 0).
		End()

	container := builder.Container()
	for _, id := range []string{"col-basic", "col-list", "col-inputs"} {
		id := id
		if w := Find(container, id); w != nil {
			w.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
				if v, ok := data[0].(bool); ok {
					state := "collapsed"
					if v {
						state = "expanded"
					}
					if label, ok := Find(container, "col-status").(*Static); ok {
						label.SetText(fmt.Sprintf("%s: %s", id, state))
					}
				}
				return true
			})
		}
	}
}

// Deck demo — displays all theme colors as rich multi-line cards.
func deckDemo(builder *Builder, theme *Theme) {
	type colorItem struct{ name, hex string }

	colors := theme.Colors()

	names := make([]string, 0, len(colors))
	for k := range colors {
		names = append(names, k)
	}
	sort.Strings(names)

	items := make([]any, len(names))
	for i, name := range names {
		items[i] = colorItem{name: name, hex: colors[name]}
	}

	const previewW = 10 // columns for color swatch
	const borderW = 1   // column for left-side selection indicator
	const padW = 1      // padding column between indicator and text

	// deck is declared first so the render closure can reference it.
	var deck *Deck

	renderFn := func(r *Renderer, x, y, w, h, _ int, data any, selected, focused bool) {
		item := data.(colorItem)
		textW := w - borderW - padW - previewW

		bg := theme.Color("$bg1")

		var fg, font, indicator, indicatorFg string
		if selected {
			font = "bold"
			if focused {
				fg = theme.Color("$cyan")
				indicatorFg = theme.Color("$cyan")
			} else {
				fg = theme.Color("$fg1")
				indicatorFg = theme.Color("$gray")
			}
			indicator = "▍"
		} else {
			fg = theme.Color("$fg0")
			font = ""
			indicator = " "
			indicatorFg = theme.Color("$gray")
		}

		// Left-side border indicator — spans all rows.
		r.Set(indicatorFg, bg, "")
		for row := 0; row < h; row++ {
			r.Put(x, y+row, indicator)
		}

		// Padding column.
		r.Set("", bg, "")
		r.Fill(x+borderW, y, padW, h, " ")

		// Text area: name row (bold/colored), hex row, empty padding row.
		textX := x + borderW + padW
		r.Set(fg, bg, font)
		r.Text(textX, y, item.name, textW)
		r.Set(fg, bg, "")
		r.Text(textX, y+1, item.hex, textW)
		r.Fill(textX, y+2, textW, 1, " ")

		// Color swatch — 2 rows tall, right-aligned.
		swatchX := x + borderW + padW + textW
		r.Set("", item.hex, "")
		r.Fill(swatchX, y, previewW, 2, " ")
		// Clear the third row of the swatch area.
		r.Set("", bg, "")
		r.Fill(swatchX, y+2, previewW, 1, " ")
	}

	deck = NewDeck("deck-demo", "", renderFn, 3)
	deck.SetItems(items)
	// Wrap in a non-focusable Flex so the left/right padding is stable and
	// unaffected by the deck's own focus state changing its style.
	builder.Flex("deck-wrapper", false, "stretch", 0).Padding(0, 1).
		Static("deck-title", "Color constants from the Tokyo Night theme").Padding(0, 0, 1, 0).
		Add(deck).Hint(0, -1).
		End()
}

func custom() Widget {
	result := NewCustom("custom", "", func(widget Widget, r *Renderer) {
		_, _, width, height := widget.Content()
		for x := 10; x < width; x += 10 {
			for y := 10; y < height; y += 10 {
				r.Put(x, y, "*")
			}
		}
	})
	result.SetStyle("", NewStyle().WithColors("green", "black").WithMargin(0).WithPadding(0))
	result.SetHint(200, 100)
	return result
}

// Digits demo
func digits(builder *Builder) {
	builder.Flex("digits-demo", false, "stretch", 1).Padding(1).
		Static("digits-title", "Digits Widget Demo").Padding(0, 0, 1, 0).
		Flex("digits-content", true, "center", 1).
		Digits("digits", "12:34").
		End().
		Static("digits-info", "Large ASCII art-style digits using Unicode box-drawing characters.").Padding(1, 0, 0, 0).
		End()
}

func dropdown(builder *Builder) {
	builder.Flex("select-demo", false, "start", 1).Padding(1, 2).
		Static("select-title", "Select Widget Demo").Padding(0, 0, 1, 0).
		Static("select-info", "Select is a dropdown selection widget.").Padding(0, 0, 1, 0).
		Select("s1", "f", "Female", "m", "Male", "d", "Diverse").
		Static("select-status", "You selected: ").Padding(1, 0, 0, 0).
		End()
}

// Editor demo
func editor(builder *Builder) {
	builder.Editor("editor-demo").Hint(0, -1).Padding(1)
	if ed := Find(builder.Container(), "editor-demo"); ed != nil {
		if editor, ok := ed.(*Editor); ok {
			editor.ShowLineNumbers(true)
			editor.Load("This is a sample text.\nYou can edit me!\n\nPress Tab to insert tabs,\nBackspace to delete,\nand arrow keys to navigate.")
		}
	}
}

func form(builder *Builder) {
	data := struct {
		Database string `width:"40"`
		Username string `width:"20"`
		Password string `control:"password" width:"20" line:"1"`
	}{
		Database: "sqlite:mem",
		Username: "admin",
		Password: "secret",
	}

	user := struct {
		ID         string `readonly:"true"`
		Login      string `label:"Login Name:" width:"40"`
		Name       string `width:"40"`
		Sex        string `control:"select" options:",n/a,m,Male,f,Female,d,Diverse"`
		Department string `width:"40"`
		Email      string `label:"E-Mail-Address" width:"40"`
		Phone      string `label:"Phone Number" width:"40"`
		Mobile     string `label:"Mobile Phone" width:"40"`
		Password   string `control:"password" width:"40"`
		Temporary  bool   `label:"Temporary"`
		Pending    bool   `label:"Pending"`
		Active     bool   `label:"Active"`
		Fixed      bool   `label:"Fixed" readOnly:"true"`
	}{ID: "JD", Name: "John Doe", Sex: "m"}

	builder.Flex("form-demo", false, "start", 1).Margin(2).Border("", "round").Padding(2).
		Form("form", "Connect", &data).
		Group("form-group", "", "", false, 1).Border("", "round").
		End().
		End().
		Form("form2", "User", &user).
		Group("form-group-2", "user", "", true, 1).Border("", "round").
		End().
		End().
		Flex("form-buttons", true, "start", 1).Margin(1).
		Button("save-button", "Save").
		Static("info-label", "Info").
		End().
		End()

	builder.Find("save-button").On(EvtActivate, func(widget Widget, _ Event, _ ...any) bool {
		Update(FindUI(widget), "info-label", "Activate "+time.Now().String())
		text, _ := json.Marshal(user)
		widget.Log(widget, Debug, string(text))
		return true
	})
}

// Grid demo
func grid(builder *Builder) {
	builder.Grid("grid-demo", 4, 4, true).Margin(1).Border("", "round").
		Cell(0, 0, 4, 1).Static("", "First row, spans 4 columns").
		Cell(0, 1, 1, 3).Static("", "Spans 3 rows").
		Cell(2, 2, 2, 2).Static("", "2 x 2").
		End()
}

// Progress demo
func progress(builder *Builder) {
	builder.Flex("progress-demo", false, "stretch", 1).Padding(1).
		Static("progress-title", "Progress Widget Demo").Padding(0, 0, 1, 0).
		Flex("progress-content", false, "stretch", 1)
	// Indeterminate progress
	pIndet := NewProgress("progress-indet", "", true)
	builder.Add(pIndet)
	builder.Spacer().Size(0, 1)
	// Determinate: 25%
	p25 := NewProgress("progress-25", "", true)
	p25.SetTotal(100)
	p25.SetValue(25)
	builder.Add(p25)
	builder.Spacer().Size(0, 1)
	// 50%
	p50 := NewProgress("progress-50", "", true)
	p50.SetTotal(100)
	p50.SetValue(50)
	builder.Add(p50)
	builder.Spacer().Size(0, 1)
	// 75%
	p75 := NewProgress("progress-75", "", true)
	p75.SetTotal(100)
	p75.SetValue(75)
	builder.Add(p75)
	builder.Spacer().Size(0, 1)
	// 100%
	p100 := NewProgress("progress-full", "", true)
	p100.SetTotal(100)
	p100.SetValue(100)
	builder.Add(p100)
	builder.End().
		Static("progress-info", "Progress bars support determinate (with total>0) and indeterminate (total=0) modes. Use SetTotal/SetValue to control.").Padding(1, 0, 0, 0).
		End()
}

// Scanner demo
func scanner(builder *Builder) {
	builder.Flex("scanner-container", false, "stretch", 1).Padding(1).
		Static("scanner-title", "Scanner Widget Demo").Padding(0, 0, 1, 0).
		Static("scanner-info", "Back-and-forth scanning animation with fading trail.").Padding(0, 0, 1, 0).
		Flex("scanner-flex", false, "center", 1).
		Scanner("scanner-blocks", 12, "blocks").
		Scanner("scanner-circles", 12, "circles").
		Scanner("scanner-diamonds", 12, "diamonds").
		End().
		Static("scanner-note", "Scanner uses a dimmed trail and cycles: forward → hold → backward → hold").Padding(1, 0, 0, 0).
		End()

	container := builder.Find("scanner-container").(Container)
	container.On(EvtShow, func(_ Widget, event Event, data ...any) bool {
		container.Log(container, Debug, "Scanner panel shown")
		for _, scanner := range FindAll[*Scanner](container) {
			scanner.Start(50 * time.Millisecond)
		}
		return true
	})

	container.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		container.Log(container, Debug, "Scanner panel hidden")
		for _, scanner := range FindAll[*Scanner](container) {
			scanner.Stop()
		}
		return true
	})
}

// Spinner demo
func spinner(builder *Builder) {
	builder.Box("spinner-demo", "Spinner").Border("", "round").Margin(1).Padding(1, 5).
		Flex("spinner-flex", true, "start", 2).
		Spinner("spinner", Spinners["bar"]).
		Spinner("spinner", Spinners["dot"]).
		Spinner("spinner", Spinners["dots"]).
		Spinner("spinner", Spinners["arrow"]).
		Spinner("spinner", Spinners["circle"]).
		Spinner("spinner", Spinners["bounce"]).
		Spinner("spinner", Spinners["braille"]).
		End().
		End()

	container := builder.Find("spinner-demo").(Container)
	container.On(EvtShow, func(_ Widget, event Event, data ...any) bool {
		for _, spinner := range FindAll[*Spinner](container) {
			spinner.Start(100 * time.Millisecond)
		}
		return true
	})

	container.On(EvtHide, func(_ Widget, _ Event, _ ...any) bool {
		for _, spinner := range FindAll[*Spinner](container) {
			spinner.Stop()
		}
		return true
	})
}

// Styled text demo
func styled(builder *Builder) {
	builder.Styled("styled-demo", "Styled *Text* __Widget__ **Demo**! ~~Not found~~ We are producing a very long text to test word wrapping functionality for the styled text widget and verify, that long lines are wrapped, if they are wider than the widget content area.").Padding(1)
}

// Table demo
func table(builder *Builder) {
	headers := []string{
		"First name", "Last name", "Street address", "ZIP", "City", "State", "Country", "Phone", "E-Mail", "Date of Birth", "Age", "Place of Birth", "Income", "SSN", "Sex",
	}
	data := people(100)
	builder.Table("table-demo", NewArrayTableProvider(headers, data)).Hint(0, -1)
}

// Tabs demo
func tabs(builder *Builder) {
	builder.Flex("tabs-demo", false, "stretch", 1).Padding(1, 2).
		Tabs("tabs", "First", "Second", "Third", "Fourth").
		End()
}

// Terminal demo — feeds representative ANSI/VT sequences into a Terminal widget.
func terminalDemo(builder *Builder) {
	term := NewTerminal("terminal-demo", "")
	term.SetHint(0, -1)

	builder.Flex("terminal-pane", false, "stretch", 0).Hint(0, -1).Padding(0, 1).
		Static("terminal-title", "Terminal Widget Demo").Padding(0, 0, 1, 0).
		Add(term).Hint(0, -1).
		End()

	// Re-render on each visit so the content fits the current dimensions.
	pane := builder.Find("terminal-pane").(Container)
	pane.On(EvtShow, func(_ Widget, _ Event, _ ...any) bool {
		term.Clear()
		writeTerminalDemo(term)
		return true
	})
}

// writeTerminalDemo pipes a sequence of ANSI escape sequences into t to
// showcase the terminal widget's rendering capabilities.
func writeTerminalDemo(t *Terminal) {
	w := func(s string) { t.Write([]byte(s)) }

	// ---- Title ----
	w("\033[1;37mANSI / VT Terminal Demo\033[0m\r\n")
	w("\033[2m─────────────────────────────────────────────────────\033[0m\r\n\r\n")

	// ---- Text attributes ----
	w("\033[1mBold\033[0m  ")
	w("\033[2mDim\033[0m  ")
	w("\033[3mItalic\033[0m  ")
	w("\033[9mStrikethrough\033[0m  ")
	w("\033[5mBlink\033[0m\r\n\r\n")

	// ---- Underline styles ----
	w("Underline styles: ")
	w("\033[4mSingle\033[0m  ")
	w("\033[21mDouble\033[0m  ")
	w("\033[4:3mCurly\033[0m  ")
	w("\033[4:4mDotted\033[0m  ")
	w("\033[4:5mDashed\033[0m\r\n\r\n")

	// ---- Standard 16 ANSI colours ----
	w("Standard colours:  ")
	for i := 0; i < 8; i++ {
		w(fmt.Sprintf("\033[%dm  \033[0m", 40+i))
	}
	w("\r\n                   ")
	for i := 0; i < 8; i++ {
		w(fmt.Sprintf("\033[%dm  \033[0m", 100+i))
	}
	w("\r\n\r\n")

	// ---- 256-colour palette strip ----
	w("256-colour palette:\r\n")
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
	// Greyscale ramp
	w("  ")
	for i := 232; i <= 255; i++ {
		w(fmt.Sprintf("\033[48;5;%dm  \033[0m", i))
	}
	w("\r\n\r\n")

	// ---- True colour gradient ----
	w("True colour: ")
	steps := 48
	for i := 0; i < steps; i++ {
		r := 255 * i / (steps - 1)
		g := 128
		b := 255 - r
		w(fmt.Sprintf("\033[48;2;%d;%d;%dm \033[0m", r, g, b))
	}
	w("\r\n\r\n")

	// ---- Underline colour ----
	w("\033[4:3m\033[58;2;255;100;0mCurly underline in orange\033[0m\r\n\r\n")

	// ---- Box-drawing ----
	w("Box-drawing: ┌──────────┐\r\n")
	w("             │  \033[1;32mHello!\033[0m  │\r\n")
	w("             └──────────┘\r\n\r\n")

	// ---- Reverse video ----
	w("\033[7m Reverse video \033[0m\r\n\r\n")

	// ---- OSC title (not visible but exercises the parser) ----
	w("\033]0;Terminal Widget Demo\007")
}

func typeaheadDemo(builder *Builder) {
	languages := []string{
		"Ada", "Clojure", "C", "C++", "C#", "Crystal", "D", "Dart", "Elixir",
		"Elm", "Erlang", "F#", "Go", "Groovy", "Haskell", "Java", "JavaScript",
		"Julia", "Kotlin", "Lisp", "Lua", "Nim", "OCaml", "Pascal", "Perl",
		"PHP", "Python", "R", "Ruby", "Rust", "Scala", "Scheme", "Swift",
		"TypeScript", "Zig",
	}
	countries := []string{
		"Afghanistan", "Albania", "Algeria", "Argentina", "Australia", "Austria",
		"Belgium", "Bolivia", "Brazil", "Bulgaria", "Canada", "Chile", "China",
		"Colombia", "Croatia", "Czech Republic", "Denmark", "Egypt", "Finland",
		"France", "Germany", "Greece", "Hungary", "India", "Indonesia", "Iran",
		"Iraq", "Ireland", "Israel", "Italy", "Japan", "Jordan", "Kenya",
		"Malaysia", "Mexico", "Morocco", "Netherlands", "New Zealand", "Nigeria",
		"Norway", "Pakistan", "Peru", "Philippines", "Poland", "Portugal",
		"Romania", "Russia", "Saudi Arabia", "Serbia", "Singapore", "South Africa",
		"South Korea", "Spain", "Sweden", "Switzerland", "Thailand", "Turkey",
		"Ukraine", "United Kingdom", "United States", "Vietnam",
	}

	builder.Flex("typeahead-demo", false, "stretch", 1).Padding(1, 2).
		Static("typeahead-title", "Typeahead Widget Demo").Padding(0, 0, 1, 0).
		Static("typeahead-desc", "Type to see inline ghost-text completions. Tab or → accepts.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Static("", "Programming language:").
		Typeahead("ta-lang", "", "e.g. Go, Rust, Python…").
		Static("", "Country:").
		Typeahead("ta-country", "", "e.g. Germany, Japan…").
		Static("ta-accepted", "").Padding(1, 0, 0, 0).
		End()

	container := builder.Container()

	langTA := Find(container, "ta-lang").(*Typeahead)
	langTA.SetSuggest(Suggest(languages))
	langTA.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if s, ok := data[0].(string); ok {
			if label, ok := Find(container, "ta-accepted").(*Static); ok {
				label.SetText("Accepted: " + s)
			}
		}
		return true
	})

	countryTA := Find(container, "ta-country").(*Typeahead)
	countryTA.SetSuggest(Suggest(countries))
	countryTA.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if s, ok := data[0].(string); ok {
			if label, ok := Find(container, "ta-accepted").(*Static); ok {
				label.SetText("Accepted: " + s)
			}
		}
		return true
	})
}

func treeFSDemo(builder *Builder) {
	var tfs *TreeFS

	tfs = NewTreeFS("tree-fs", "", ".", false)

	builder.Flex("tree-fs-demo", false, "stretch", 0).
		// Toolbar: Up button + current root path
		Flex("tree-fs-toolbar", true, "center", 1).Padding(0, 1).
		Button("tree-fs-up", "↑ Up").
		Static("tree-fs-path", tfs.RootPath()).Padding(0, 1).
		End().
		// The tree itself, takes all remaining height
		Add(tfs.Tree).Hint(0, -1).
		// Status bar showing the highlighted path
		Static("tree-fs-selected", "").Padding(0, 1).
		End()

	container := builder.Container()

	// Up button: navigate to the parent directory
	builder.Find("tree-fs-up").On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
		parent := filepath.Dir(tfs.RootPath())
		if parent == tfs.RootPath() {
			return true // already at filesystem root
		}
		tfs.SetRoot(parent)
		if label, ok := Find(container, "tree-fs-path").(*Static); ok {
			label.SetText(tfs.RootPath())
		}
		if label, ok := Find(container, "tree-fs-selected").(*Static); ok {
			label.SetText("")
		}
		return true
	})

	// Update the status bar whenever the highlighted node changes
	tfs.Tree.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		if len(data) == 0 {
			return true
		}
		node, ok := data[0].(*TreeNode)
		if !ok {
			return true
		}
		if label, ok := Find(container, "tree-fs-selected").(*Static); ok {
			label.SetText(node.Data().(string))
		}
		return true
	})
}

func viewport(builder *Builder) {
	builder.Flex("viewport-demo", false, "stretch", 1).Padding(1, 2).
		Static("viewport-title", "Viewport Demo").Padding(0, 0, 1, 0).
		Static("viewport-info", "A scrollable viewport of the inside widget.").Padding(0, 0, 1, 0).
		HRule("thin").Padding(0, 0, 1, 0).
		Viewport("viewport", "Viewport").Border("thin").Hint(-1, -1).
		Add(custom()).
		End().
		End()
}

// Table demo data generation
func people(n int) [][]string {
	firstNames := []string{"John", "Jane", "Michael", "Emily", "David", "Sophia", "James", "Olivia", "Daniel", "Ava", "Liam", "Emma", "Noah", "Isabella", "Ethan", "Mia", "Lucas", "Charlotte", "Mason", "Amelia"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin"}
	streets := []string{"Maple St", "Oak Ave", "Pine Rd", "Birch Blvd", "Cedar Ln", "Spruce Ct", "Willow Way", "Elm Pl", "Aspen Dr", "Cypress St"}
	cities := []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose"}
	states := []string{"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA", "HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD", "MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ", "NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC", "SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY"}

	result := make([][]string, n)
	for i := range n {
		first := firstNames[rand.Intn(len(firstNames))]
		last := lastNames[rand.Intn(len(lastNames))]
		street := streets[rand.Intn(len(streets))]
		city := cities[rand.Intn(len(cities))]
		state := states[rand.Intn(len(states))]
		zip := fmt.Sprintf("%05d", 10000+rand.Intn(90000))
		phone := fmt.Sprintf("+1-%03d-%03d-%04d", rand.Intn(900)+100, rand.Intn(900)+100, rand.Intn(10000))
		email := strings.ToLower(fmt.Sprintf("%s.%s@example.com", first, last))
		birth := fmt.Sprintf("%04d-%02d-%02d", 1950+rand.Intn(50), rand.Intn(12)+1, rand.Intn(28)+1)
		age := rand.Intn(50) + 20
		pob := cities[rand.Intn(len(cities))]
		income := 20000 + rand.Intn(100000)
		ssn := fmt.Sprintf("%03d-%02d-%04d", rand.Intn(900)+100, rand.Intn(80)+10, rand.Intn(10000))
		sex := []string{"M", "F"}[rand.Intn(2)]

		result[i] = []string{first, last, street, zip, city, state, "USA", phone, email, birth, fmt.Sprintf("%d", age), pob, fmt.Sprintf("$%d", income), ssn, sex}
	}
	return result
}
