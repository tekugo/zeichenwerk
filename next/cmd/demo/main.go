package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	. "github.com/tekugo/zeichenwerk/next"
)

func main() {
	ui := createUI()
	ui.Run()
}

func createUI() *UI {
	ui := NewBuilder(TokyoNightTheme()).
		Flex("main", false, "stretch", 0).
		Flex("header", true, "stretch", 2).
		Static("title", "Zeichenwerk Demo").
		Static("subtitle", "A terminal UI framework").
		End().
		Grid("content", 2, 2, true).Hint(0, -1).Columns(32, -1).Rows(-1, 10).
		Cell(0, 0, 1, 2).
		List("navigation", "Box", "Checkbox", "Digits", "Editor", "Grid", "Scanner", "Spinner", "Styled", "Table", "Tabs").
		Cell(1, 0, 1, 1).
		Switcher("switcher", false).
		With(box).
		With(checkbox).
		With(digits).
		With(editor).
		With(grid).
		With(scanner).
		With(spinner).
		With(styled).
		With(table).
		With(tabs).
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
				switcher.Select(selected)
				if selected == 5 {
					for _, spinner := range FindAll[*Spinner](ui) {
						spinner.Start(100 * time.Millisecond)
					}
				} else if selected == 4 {
					for _, scanner := range FindAll[*Scanner](ui) {
						scanner.Start(50 * time.Millisecond)
					}
				}
				if selected != 5 {
					for _, spinner := range FindAll[*Spinner](ui) {
						spinner.Stop()
					}
				}
				if selected != 4 {
					for _, scanner := range FindAll[*Scanner](ui) {
						scanner.Stop()
					}
				}
				return true
			}
		}
		return false
	})

	return ui
}

func box(builder *Builder) {
	builder.Flex("box-demo", false, "stretch", 1).Padding(1).
		Static("box-title", "Box Widget Demo").Padding(0, 0, 1, 0).
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
		Static("box-info", "Boxes are containers that can hold a single child widget with optional borders and titles.").Padding(1, 0, 0, 0).
		End()
}

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

func digits(builder *Builder) {
	builder.Flex("digits-demo", false, "stretch", 1).Padding(1).
		Static("digits-title", "Digits Widget Demo").Padding(0, 0, 1, 0).
		Flex("digits-content", true, "center", 1).
		Digits("digits", "12:34").
		End().
		Static("digits-info", "Large ASCII art-style digits using Unicode box-drawing characters.").Padding(1, 0, 0, 0).
		End()
}

func editor(builder *Builder) {
	builder.Editor("editor-demo").Hint(0, -1).Padding(1)
	if ed := Find(builder.Container(), "editor-demo"); ed != nil {
		if editor, ok := ed.(*Editor); ok {
			editor.Load("This is a sample text.\nYou can edit me!\n\nPress Tab to insert tabs,\nBackspace to delete,\nand arrow keys to navigate.\n\nLine numbers are disabled by default.\nEnable them with ShowLineNumbers(true).")
		}
	}
}

func grid(builder *Builder) {
	builder.Grid("grid-demo", 4, 4, true).Margin(1).Border("", "round").
		Cell(0, 0, 4, 1).Static("", "First row, spans 4 columns").
		Cell(0, 1, 1, 3).Static("", "Spans 3 rows").
		Cell(2, 2, 2, 2).Static("", "2 x 2").
		End()
}

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
}

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
}

func styled(builder *Builder) {
	builder.Styled("styled-demo", "Styled *Text* __Widget__ **Demo**! ~~Not found~~ We are producing a very long text to test word wrapping functionality for the styled text widget and verify, that long lines are wrapped, if they are wider than the widget content area.").Padding(1)
}

func table(builder *Builder) {
	headers := []string{
		"First name", "Last name", "Street address", "ZIP", "City", "State", "Country", "Phone", "E-Mail", "Date of Birth", "Age", "Place of Birth", "Income", "SSN", "Sex",
	}
	data := people(100)
	builder.Table("table-demo", NewArrayTableProvider(headers, data)).Hint(0, -1)
}

func tabs(builder *Builder) {
	builder.Flex("tabs-demo", false, "stretch", 1).Padding(1, 2).
		Tabs("tabs", "First", "Second", "Third", "Fourth").
		End().
		Flex("tab-content", true, "stretch", 0).
		Static("tab-1", "Content for the first tab.").
		Static("tab-2", "Content for the second tab.").
		Static("tab-3", "Content for the third tab.").
		Static("tab-4", "Content for the fourth tab.").
		End()
}

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
