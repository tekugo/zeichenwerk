package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	. "github.com/tekugo/zeichenwerk"
)

// main function
func main() {
	ui := createUI()
	ui.Run()
}

// Create the terminal UI.
func createUI() *UI {
	ui := NewBuilder(TokyoNightTheme()).
		Flex("main", false, "stretch", 0).
		Flex("header", true, "stretch", 2).
		Static("title", "Zeichenwerk Demo").
		Static("subtitle", "A terminal UI framework").
		End().
		Grid("content", 2, 2, true).Hint(0, -1).Columns(32, -1).Rows(-1, 10).
		Cell(0, 0, 1, 2).
		List("navigation", "Box", "Canvas", "Checkbox", "Digits", "Editor", "Form", "Grid", "Progress", "Scanner", "Select", "Spinner", "Styled", "Table", "Tabs", "Viewport", "Dialog").
		Cell(1, 0, 1, 1).
		Switcher("switcher", false).
		With(box).
		With(canvas).
		With(checkbox).
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
		With(viewport).
		End().
		Cell(1, 1, 1, 1).
		Flex("debug-log-pane", false, "stretch", 0).Hint(0, 10).
		Static("debug-log-title", "Debug Log").Background("green").
		Text("debug-log", []string{"Hello, World!"}, true, 100).Hint(0, -1).
		End().
		End().
		Flex("footer", true, "stretch", 0).
		Static("footer-text", "Footer").
		End().
		Build()

	switcher := Find(ui, "switcher").(*Switcher)
	Find(ui, "navigation").On("activate", func(_ Widget, event string, data ...any) bool {
		if len(data) == 1 {
			if selected, ok := data[0].(int); ok {
				if selected < len(switcher.Children()) {
					switcher.Select(selected)
				} else {
					switch selected {
					case 14:
						dialog := ui.NewBuilder().
							Dialog("dialog", "Test Dialog").
							Class("dialog").
							Flex("dialog-content", false, "end", 1).
							Static("", "Do you really want to do this?").
							Flex("dialog-buttons", true, "end", 2).
							Button("btn-yes", "YES").
							Button("btn-no", "NO").
							End().
							End().
							Class("").
							Container()
						ui.Popup(-1, -1, 0, 0, dialog)
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
		Flex("box-examples", true, "stretch", 2).
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
	c := NewCanvas("demo-canvas", 40, 20)

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
			cb.On("change", func(_ Widget, event string, data ...any) bool {
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

func custom() Widget {
	result := NewCustom("custom", func(widget Widget, r *Renderer) {
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
			editor.Load("This is a sample text.\nYou can edit me!\n\nPress Tab to insert tabs,\nBackspace to delete,\nand arrow keys to navigate.\n\nLine numbers are disabled by default.\nEnable them with ShowLineNumbers(true).")
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

	builder.Find("save-button").On("click", func(widget Widget, _ string, _ ...any) bool {
		Update(FindUI(widget), "info-label", "Click "+time.Now().String())
		text, _ := json.Marshal(user)
		widget.Log(widget, "debug", string(text))
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
	pIndet := NewProgress("progress-indet", true)
	builder.Add(pIndet)
	builder.Spacer().Size(0, 1)
	// Determinate: 25%
	p25 := NewProgress("progress-25", true)
	p25.SetTotal(100)
	p25.SetValue(25)
	builder.Add(p25)
	builder.Spacer().Size(0, 1)
	// 50%
	p50 := NewProgress("progress-50", true)
	p50.SetTotal(100)
	p50.SetValue(50)
	builder.Add(p50)
	builder.Spacer().Size(0, 1)
	// 75%
	p75 := NewProgress("progress-75", true)
	p75.SetTotal(100)
	p75.SetValue(75)
	builder.Add(p75)
	builder.Spacer().Size(0, 1)
	// 100%
	p100 := NewProgress("progress-full", true)
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
	container.On("show", func(_ Widget, event string, data ...any) bool {
		container.Log(container, "debug", "Scanner panel shown")
		for _, scanner := range FindAll[*Scanner](container) {
			scanner.Start(50 * time.Millisecond)
		}
		return true
	})

	container.On("hide", func(_ Widget, _ string, _ ...any) bool {
		container.Log(container, "debug", "Scanner panel hidden")
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
	container.On("show", func(_ Widget, event string, data ...any) bool {
		for _, spinner := range FindAll[*Spinner](container) {
			spinner.Start(100 * time.Millisecond)
		}
		return true
	})

	container.On("hide", func(_ Widget, _ string, _ ...any) bool {
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
