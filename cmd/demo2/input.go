package main

import (
	"fmt"

	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	. "github.com/tekugo/zeichenwerk/widgets"
)

var inputEntries = []Entry{
	{
		Category: "Input",
		Name:     "Button",
		Summary:  "Clickable button with text label.",
		DocFile:  "button.md",
		DemoFn:   buttonDemo,
		Builder: `builder.Button("save", "Save")
btn := builder.Find("save").(*Button)
btn.On(EvtActivate, func(_ Widget, _ Event, _ ...any) bool {
    println("clicked")
    return true
})`,
		Compose: `compose.Button("save", "", "Save",
    compose.OnActivate(func(_ core.Widget, _ core.Event, _ ...any) bool {
        println("clicked")
        return true
    }),
)`,
	},
	{
		Category: "Input",
		Name:     "Checkbox",
		Summary:  "Toggleable boolean input with a label.",
		DocFile:  "checkbox.md",
		DemoFn:   checkboxDemo,
		Builder: `builder.Checkbox("notify", "Enable notifications", false)
cb := builder.Find("notify").(*Checkbox)
cb.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
    fmt.Println("notify =", data[0])
    return true
})`,
		Compose: `compose.Checkbox("notify", "", "Enable notifications", false)`,
	},
	{
		Category: "Input",
		Name:     "Combo",
		Summary:  "Free-text input combined with a popup suggestion list.",
		DocFile:  "combo.md",
		DemoFn:   comboDemoFn,
		Builder: `builder.Combo("history", "fix bug", "add feature", "refactor")
combo := builder.Find("history").(*Combo)
combo.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
    fmt.Println("submitted:", data[0])
    return true
})`,
		Compose: `compose.Combo("history", "", []string{"fix bug", "add feature", "refactor"})`,
	},
	{
		Category: "Input",
		Name:     "Editor",
		Summary:  "Multi-line text editor with cursor and line numbers.",
		DocFile:  "editor.md",
		DemoFn:   editorDemo,
		Builder: `builder.Editor("editor").Hint(0, -1)
ed := builder.Find("editor").(*Editor)
ed.ShowLineNumbers(true)
ed.Load("Multi-line\ntext editor.\nTab to indent.")`,
		Compose: `compose.Editor("editor", "", compose.Hint(0, -1))`,
	},
	{
		Category: "Input",
		Name:     "Filter",
		Summary:  "Typeahead input bound to a Filterable widget (e.g. List).",
		DocFile:  "filter.md",
		DemoFn:   filterDemoFn,
		Builder: `builder.Filter("search").
List("results", items...).Hint(0, -1)
filter := builder.Find("search").(*Filter)
list := builder.Find("results").(*List)
filter.Bind(list)`,
		Compose: `compose.Filter("search", ""),
compose.List("results", "", items, compose.Hint(0, -1))
// After build:
filter := core.Find(ui, "search").(*widgets.Filter)
list := core.Find(ui, "results").(*widgets.List)
filter.Bind(list)`,
	},
	{
		Category: "Input",
		Name:     "Input",
		Summary:  "Single-line text field with optional placeholder and mask.",
		DocFile:  "input.md",
		DemoFn:   inputDemo,
		Builder: `builder.Input("name", "", "Your name…")           // placeholder
builder.Input("password", "", "", "•").Flag(FlagMasked, true) // masked`,
		Compose: `compose.Input("name", "", []string{"", "Your name…"}),
compose.Input("password", "", []string{"", "", "•"})`,
	},
	{
		Category: "Input",
		Name:     "List",
		Summary:  "Scrollable, filterable, selectable list of strings.",
		DocFile:  "list.md",
		DemoFn:   listDemo,
		Builder: `builder.List("colors", "Red", "Green", "Blue", "Cyan", "Magenta", "Yellow")
list := builder.Find("colors").(*List)
list.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
    fmt.Println("picked:", list.Items()[data[0].(int)])
    return true
})`,
		Compose: `compose.List("colors", "", []string{"Red", "Green", "Blue", "Cyan", "Magenta", "Yellow"})`,
	},
	{
		Category: "Input",
		Name:     "Radio",
		Summary:  "Inline mutually-exclusive choice; same key/label pairs as Select.",
		DocFile:  "radio.md",
		DemoFn:   radioDemo,
		Builder: `builder.Radio("size", "s", "Small", "m", "Medium", "l", "Large")
r := builder.Find("size").(*Radio)
r.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
    fmt.Println("size =", data[0])
    return true
})`,
		Compose: `compose.Radio("size", "", []string{"s", "Small", "m", "Medium", "l", "Large"})`,
	},
	{
		Category: "Input",
		Name:     "Select",
		Summary:  "Dropdown selection with key/label pairs.",
		DocFile:  "select.md",
		DemoFn:   selectDemo,
		Builder: `builder.Select("sex", "f", "Female", "m", "Male", "d", "Diverse")`,
		Compose: `compose.Select("sex", "", []string{"f", "Female", "m", "Male", "d", "Diverse"})`,
	},
	{
		Category: "Input",
		Name:     "Slider",
		Summary:  "Horizontal int range input — compact one-row bar at h=1, rounded box at h≥2.",
		DocFile:  "slider.md",
		DemoFn:   sliderDemo,
		Builder: `builder.Slider("volume").Hint(0, 1)
s := builder.Find("volume").(*Slider)
s.SetMin(0); s.SetMax(11); s.SetStep(1); s.Set(7)
s.On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
    fmt.Println("volume =", data[0])
    return true
})`,
		Compose: `compose.Slider("volume", "",
    compose.Range(0, 11),
    compose.Step(1),
    compose.Value(7),
    compose.Hint(0, 1),
)`,
	},
	{
		Category: "Input",
		Name:     "Tree",
		Summary:  "Expandable hierarchy of nodes with lazy-loading support.",
		DocFile:  "tree.md",
		DemoFn:   treeDemo,
		Builder: `t := NewTree("tree", "")
root := NewTreeNode("Repo")
src := NewTreeNode("src")
src.Add(NewTreeNode("main.go"))
src.Add(NewTreeNode("ui.go"))
root.Add(src)
root.Add(NewTreeNode("README.md"))
t.Add(root)
builder.Add(t)`,
		Compose: `// Compose API: pass a fully-built Tree via Include.
compose.Include(func(_ *core.Theme) core.Widget {
    t := widgets.NewTree("tree", "")
    // … build TreeNodes …
    return t
})`,
	},
	{
		Category: "Input",
		Name:     "TreeFS",
		Summary:  "Tree pre-wired for filesystem navigation, with lazy directory loading.",
		DocFile:  "tree-fs.md",
		DemoFn:   treeFSDemoFn,
		Builder: `tfs := NewTreeFS("fs", "", ".", false) // false = include files
builder.Add(tfs.Tree)`,
		Compose: `compose.TreeFS("fs", "", ".", false)`,
	},
	{
		Category: "Input",
		Name:     "Typeahead",
		Summary:  "Input with ghost-text suggestion completion.",
		DocFile:  "typeahead.md",
		DemoFn:   typeaheadDemoFn,
		Builder: `langs := []string{"Go", "Rust", "Python", "Zig", "C", "C++"}
builder.Typeahead("lang", "", "Type a language…")
ta := builder.Find("lang").(*Typeahead)
ta.SetSuggest(Suggest(langs))`,
		Compose: `// Build the widget then attach the suggester.
compose.Typeahead("lang", "", []string{"", "Type a language…"})`,
	},
}

// ── Demo functions ────────────────────────────────────────────────────────────

func buttonDemo(b *Builder) {
	b.VFlex("button-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Buttons fire EvtActivate on Enter, Space, or click.").
		Padding(0, 0, 1, 0).
		HFlex("row", Center, 2).
		Button("save", "Save").
		Button("cancel", "Cancel").
		Button("apply", "Apply").
		End().
		Static("status", "").Padding(1, 0, 0, 0).
		End()

	for _, id := range []string{"save", "cancel", "apply"} {
		id := id
		Find(b.Container(), id).On(EvtActivate, func(w Widget, _ Event, _ ...any) bool {
			if s, ok := Find(b.Container(), "status").(*Static); ok {
				s.Set("Activated: " + id)
			}
			return true
		})
	}
}

func checkboxDemo(b *Builder) {
	b.VFlex("checkbox-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Checkboxes toggle a boolean state.").
		Padding(0, 0, 1, 0).
		Checkbox("c1", "Enable notifications", true).
		Checkbox("c2", "Auto-save documents", false).
		Checkbox("c3", "Sync with cloud", false).
		Checkbox("c4", "I agree to the terms", false).
		Static("status", "Toggle with Space or Enter.").Padding(1, 0, 0, 0).
		End()
	for _, id := range []string{"c1", "c2", "c3", "c4"} {
		id := id
		Find(b.Container(), id).On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
			if s, ok := Find(b.Container(), "status").(*Static); ok {
				s.Set(fmt.Sprintf("%s = %v", id, data[0]))
			}
			return true
		})
	}
}

func comboDemoFn(b *Builder) {
	history := []string{
		"fix memory leak in renderer",
		"add dark mode support",
		"refactor event handling",
		"update dependencies",
		"implement file chooser",
	}
	b.VFlex("combo-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Type freely or pick from the list. ↓↑ navigate, Tab/→ accept, Enter confirms.").
		Padding(0, 0, 1, 0).
		Combo("history", history...).
		Static("status", "").Padding(1, 0, 0, 0).
		End()
	c := b.Find("history").(*Combo)
	c.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if s, ok := data[0].(string); ok {
			Find(b.Container(), "status").(*Static).Set("Submitted: " + s)
		}
		return true
	})
}

func editorDemo(b *Builder) {
	b.VFlex("editor-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Multi-line editor with line numbers, cursor, Tab indentation.").
		Padding(0, 0, 1, 0).
		Editor("editor").Hint(0, -1).
		End()
	if ed, ok := b.Find("editor").(*Editor); ok {
		ed.ShowLineNumbers(true)
		ed.Load("// Press Tab to indent\n// Backspace to delete\n// Arrow keys to navigate\n\nfunc main() {\n\tfmt.Println(\"Hello, world!\")\n}")
	}
}

func filterDemoFn(b *Builder) {
	items := []string{
		"Go", "Rust", "TypeScript", "Python", "Kotlin", "Swift", "Zig",
		"C", "C++", "C#", "Java", "Scala", "Haskell", "Erlang", "Elixir",
		"Ruby", "PHP", "Dart", "Lua", "Julia", "R", "Clojure",
	}
	b.VFlex("filter-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Type to filter the list. Ghost text suggests the first prefix match.").
		Padding(0, 0, 1, 0).
		Filter("filter").
		List("results", items...).Hint(0, -1).
		End()
	f := b.Find("filter").(*Filter)
	l := b.Find("results").(*List)
	f.Bind(l)
}

func inputDemo(b *Builder) {
	b.VFlex("input-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Single-line text field. Use placeholders, masks for passwords, and read-only mode.").
		Padding(0, 0, 1, 0).
		Static("name-label", "Name:").
		Input("name", "", "Your name…").
		Spacer().Hint(0, 1).
		Static("pw-label", "Password:").
		Input("pw", "", "", "•").Flag(FlagMasked, true).
		Spacer().Hint(0, 1).
		Static("ro-label", "Read-only:").
		Input("ro", "Cannot edit me").Flag(FlagReadonly, true).
		End()
}

func listDemo(b *Builder) {
	colors := []string{"Red", "Green", "Blue", "Cyan", "Magenta", "Yellow", "Orange", "Purple", "Brown", "Pink", "Teal", "Navy"}
	b.VFlex("list-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Scrollable list. ↑↓ to navigate, Home/End to jump, Enter to activate.").
		Padding(0, 0, 1, 0).
		List("colors", colors...).Hint(0, -1).
		Static("status", "").Padding(1, 0, 0, 0).
		End()
	l := b.Find("colors").(*List)
	s := b.Find("status").(*Static)
	l.On(EvtSelect, func(_ Widget, _ Event, data ...any) bool {
		if i, ok := data[0].(int); ok && i >= 0 && i < len(l.Items()) {
			s.Set("Highlighted: " + l.Items()[i])
		}
		return true
	})
	l.On(EvtActivate, func(_ Widget, _ Event, data ...any) bool {
		if i, ok := data[0].(int); ok && i >= 0 && i < len(l.Items()) {
			s.Set("Activated: " + l.Items()[i])
		}
		return true
	})
}

func radioDemo(b *Builder) {
	b.VFlex("radio-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Radio groups show every option inline. ↑↓ (or j/k) change the selection immediately — no separate cursor.").
		Padding(0, 0, 1, 0).
		Static("size-label", "T-shirt size:").
		Radio("size", "s", "Small", "m", "Medium", "l", "Large", "xl", "Extra Large").
		Spacer().Hint(0, 1).
		Static("env-label", "Environment:").
		Radio("env", "dev", "Development", "stg", "Staging", "prd", "Production").
		Spacer().Hint(0, 1).
		Static("status", "Pick a value above.").
		End()
	for _, id := range []string{"size", "env"} {
		id := id
		Find(b.Container(), id).On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
			if v, ok := data[0].(string); ok {
				Find(b.Container(), "status").(*Static).Set(id + " = " + v)
			}
			return true
		})
	}
}

func selectDemo(b *Builder) {
	b.VFlex("select-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Dropdown selection. Enter or Space opens the popup; arrow keys + Enter pick a value.").
		Padding(0, 0, 1, 0).
		Static("sex-label", "Sex:").
		Select("sex", "f", "Female", "m", "Male", "d", "Diverse").
		Spacer().Hint(0, 1).
		Static("env-label", "Environment:").
		Select("env", "dev", "Development", "stg", "Staging", "prd", "Production").
		Spacer().Hint(0, 1).
		Static("status", "Pick a value above.").
		End()
	for _, id := range []string{"sex", "env"} {
		id := id
		Find(b.Container(), id).On(EvtChange, func(_ Widget, _ Event, data ...any) bool {
			if v, ok := data[0].(string); ok {
				Find(b.Container(), "status").(*Static).Set(id + " = " + v)
			}
			return true
		})
	}
}

func sliderDemo(b *Builder) {
	b.VFlex("slider-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Horizontal int range input. ←/→ (or h/l) step; Home/End jump to bounds. Click on the track to set a value.").
		Padding(0, 0, 1, 0).
		Static("compact-label", "Compact (height 1 — heavy line + bar):").
		Slider("volume").Hint(0, 1).
		Spacer().Hint(0, 1).
		Static("box-label", "Box (height 2 — rounded box + ╥╨ thumb):").
		Slider("brightness").Hint(0, 2).
		Spacer().Hint(0, 1).
		Static("tall-label", "Box, centred in a taller area (height 4):").
		Slider("contrast").Hint(0, 4).
		Static("status", "Volume = 0   Brightness = 50   Contrast = 25").Padding(1, 0, 0, 0).
		End()

	volume := b.Find("volume").(*Slider)
	brightness := b.Find("brightness").(*Slider)
	contrast := b.Find("contrast").(*Slider)
	brightness.Set(50)
	contrast.Set(25)
	status := b.Find("status").(*Static)
	update := func() {
		status.Set(fmt.Sprintf("Volume = %d   Brightness = %d   Contrast = %d",
			volume.Value(), brightness.Value(), contrast.Value()))
	}
	volume.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { update(); return true })
	brightness.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { update(); return true })
	contrast.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool { update(); return true })
}

func treeDemo(b *Builder) {
	t := NewTree("tree", "")
	root := NewTreeNode("zeichenwerk")
	cmd := NewTreeNode("cmd")
	cmd.Add(NewTreeNode("demo"))
	cmd.Add(NewTreeNode("demo2"))
	cmd.Add(NewTreeNode("showcase"))
	root.Add(cmd)
	core := NewTreeNode("core")
	core.Add(NewTreeNode("renderer.go"))
	core.Add(NewTreeNode("style.go"))
	core.Add(NewTreeNode("theme.go"))
	root.Add(core)
	widgets := NewTreeNode("widgets")
	widgets.Add(NewTreeNode("button.go"))
	widgets.Add(NewTreeNode("list.go"))
	widgets.Add(NewTreeNode("tree.go"))
	root.Add(widgets)
	root.Add(NewTreeNode("README.md"))
	t.Add(root)
	root.Expand()

	b.VFlex("tree-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Hierarchical tree. ←→ collapse/expand, ↑↓ navigate, Enter activate.").
		Padding(0, 0, 1, 0).
		Add(t).Hint(0, -1).
		End()
}

func treeFSDemoFn(b *Builder) {
	tfs := NewTreeFS("fs", "", ".", false)
	b.VFlex("tree-fs-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "TreeFS browses the local filesystem with lazy loading. Starting at the current directory.").
		Padding(0, 0, 1, 0).
		Static("path-label", tfs.RootPath()).
		Add(tfs.Tree).Hint(0, -1).
		End()
}

func typeaheadDemoFn(b *Builder) {
	languages := []string{
		"Ada", "Clojure", "C", "C++", "C#", "Crystal", "D", "Dart", "Elixir",
		"Elm", "Erlang", "F#", "Go", "Groovy", "Haskell", "Java", "JavaScript",
		"Julia", "Kotlin", "Lisp", "Lua", "Nim", "OCaml", "Perl", "PHP",
		"Python", "R", "Ruby", "Rust", "Scala", "Swift", "TypeScript", "Zig",
	}
	b.VFlex("ta-demo", Stretch, 1).Padding(1, 2).
		Static("desc", "Type a few characters — the rest appears as ghost text. Tab or → to accept.").
		Padding(0, 0, 1, 0).
		Static("lang-label", "Language:").
		Typeahead("lang", "", "e.g. Go, Rust…").
		Static("status", "").Padding(1, 0, 0, 0).
		End()
	ta := b.Find("lang").(*Typeahead)
	ta.SetSuggest(Suggest(languages))
	ta.On(EvtAccept, func(_ Widget, _ Event, data ...any) bool {
		if s, ok := data[0].(string); ok {
			Find(b.Container(), "status").(*Static).Set("Accepted: " + s)
		}
		return true
	})
}
