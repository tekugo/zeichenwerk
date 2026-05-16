package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	. "github.com/tekugo/zeichenwerk"
	. "github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/themes"
	. "github.com/tekugo/zeichenwerk/widgets"
)

// App owns the editor's runtime state and direct references to the widgets we
// mutate frequently (status line, editor widget).
type App struct {
	ui     *UI
	theme  *Theme
	editor *Editor
	status *Static

	path     string
	dirty    bool
	boxDraw  bool
	lineDraw bool
	border   string

	// paintMask tracks the *true* stroke mask for every cell we've painted
	// in the current session, keyed by [row, col]. Glyph reverse-lookup is
	// lossy for straight runes (`─` could be a passing-through line OR the
	// end of one; both render the same), so the sidecar is the only way to
	// produce correct corners when the user turns at a line endpoint.
	paintMask map[[2]int]int
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [file]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Shortcuts: Ctrl+P palette, Ctrl+O open, Ctrl+S save,")
		fmt.Fprintln(os.Stderr, "           Ctrl+Shift+S save-as, Ctrl+K delete line, Ctrl+Q quit.")
		fmt.Fprintln(os.Stderr, "Open the palette (Ctrl+P) for markdown preview, box-drawing,")
		fmt.Fprintln(os.Stderr, "line-drawing, and border style selection.")
	}
	flag.Parse()

	app := &App{
		theme:     themes.TokyoNight(),
		border:    "thin",
		paintMask: map[[2]int]int{},
	}
	app.buildUI()
	app.wireEvents()
	app.registerCommands()

	if args := flag.Args(); len(args) > 0 {
		if err := app.openPath(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "open: %v\n", err)
			os.Exit(1)
		}
	} else {
		app.refreshChrome()
	}

	app.ui.SetFocus("editor")
	if err := app.ui.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (a *App) buildUI() {
	a.ui = NewBuilder(a.theme).
		VFlex("root", Stretch, 0).
		Editor("editor").Hint(-1, -1).
		Static("status", "").
		Hint(-1, 1).
		Background("$fg0").
		Foreground("$bg0").
		End().
		Build()

	a.editor = MustFind[*Editor](a.ui, "editor")
	a.status = MustFind[*Static](a.ui, "status")

	a.editor.ShowLineNumbers(true)
	a.editor.SetTabWidth(4)
	a.editor.UseSpaces(true)
	a.editor.SetAutoIndent(true)
}

func (a *App) wireEvents() {
	a.editor.On(EvtChange, func(_ Widget, _ Event, _ ...any) bool {
		if !a.dirty {
			a.dirty = true
		}
		a.refreshChrome()
		return false
	})
	a.editor.On(EvtMove, func(_ Widget, _ Event, _ ...any) bool {
		a.refreshChrome()
		return false
	})
	OnKey(a.editor, a.handleKey)
}

// shortcutsHint is the right-aligned reminder of the most-used shortcuts. The
// caret style mirrors the convention used by Emacs / nano / VS Code keymaps.
// Mode toggles and the markdown preview are reachable only through the
// command palette (Ctrl+P) to avoid clashing with terminal bindings like
// Ctrl+M (Enter) and tmux's Ctrl+B prefix.
const shortcutsHint = "^P pal  ^O open  ^S save  ^K del-line  ^Q quit "

// refreshChrome rebuilds the status bar from the current app state. The line
// is split into a left section (filename, dirty marker, cursor position, and
// box-draw state) and a right section (shortcut reminders), separated by
// padding so the right section hugs the right edge of the screen.
func (a *App) refreshChrome() {
	if a.status == nil {
		return
	}

	name := a.path
	if name == "" {
		name = "Untitled"
	}
	if a.dirty {
		name += " *"
	}

	row, col := a.editor.Line(), a.editor.Column()
	left := fmt.Sprintf(" %s  |  Ln %d, Col %d", name, row+1, col+1)
	if mode := a.drawingModeLabel(); mode != "" {
		left += "  |  " + mode
	}

	right := shortcutsHint

	width := 0
	if _, _, w, _ := a.ui.Bounds(); w > 0 {
		width = w
	}
	gap := max(width-len(left)-len(right), 1)
	a.status.Set(left + strings.Repeat(" ", gap) + right)
}

// drawingModeLabel returns the status-bar fragment describing which drawing
// mode(s) are active and which border style they use. Returns "" when both
// modes are off.
func (a *App) drawingModeLabel() string {
	switch {
	case a.boxDraw && a.lineDraw:
		return fmt.Sprintf("draw: box+line (%s)", a.border)
	case a.boxDraw:
		return fmt.Sprintf("draw: box (%s)", a.border)
	case a.lineDraw:
		return fmt.Sprintf("draw: line (%s)", a.border)
	}
	return ""
}
