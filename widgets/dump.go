package widgets

import (
	"fmt"
	"io"
	"os"
	"strings"

	. "github.com/tekugo/zeichenwerk/v2/core"
)

// ==== AI ===================================================================

// DumpOptions configures the output produced by Dump.
type DumpOptions struct {
	// Style appends a second line per widget showing the effective border,
	// padding, margin, foreground, and background for the widget's current state.
	Style bool
}

// Dump writes a human- and LLM-readable text tree of the widget hierarchy
// rooted at root to w. Each widget occupies one line; children are indented
// two spaces per depth level. Hidden widgets are shown with a [HIDDEN] flag
// and their children are always included so the full tree is visible.
//
// Suitable for feeding into an AI agent as a snapshot of the current UI state.
func Dump(w io.Writer, root Widget, opts ...DumpOptions) {
	var opt DumpOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	dumpNode(w, root, 0, opt)
}

// DumpToStdout is a convenience wrapper that calls Dump(os.Stdout, root, opts...).
func DumpToStdout(root Widget, opts ...DumpOptions) {
	Dump(os.Stdout, root, opts...)
}

// dumpNode writes one line for widget, then recurses into its children.
// Hidden containers are printed but their children are still walked.
func dumpNode(w io.Writer, widget Widget, depth int, opt DumpOptions) {
	indent := strings.Repeat("  ", depth)

	// ── type tag ──────────────────────────────────────────────────────────────
	typ := WidgetType(widget) // from inspector.go
	id := widget.ID()
	class := ""
	if c, ok := widget.(interface{ Class() string }); ok {
		class = c.Class()
	}
	tag := typ
	if id != "" {
		tag += "#" + id
	}
	if class != "" {
		tag += "." + class
	}

	// ── content summary ───────────────────────────────────────────────────────
	summary := ""
	if s, ok := widget.(Summarizer); ok {
		summary = s.Summary()
	}

	// ── bounds ────────────────────────────────────────────────────────────────
	x, y, bw, bh := widget.Bounds()
	bounds := fmt.Sprintf("@%d,%d %dx%d", x, y, bw, bh)

	// ── flags ─────────────────────────────────────────────────────────────────
	var flags []string
	if widget.Flag(FlagHidden) {
		flags = append(flags, "HIDDEN")
	}
	if widget.Flag(FlagFocused) {
		flags = append(flags, "FOCUSED")
	}
	if widget.Flag(FlagDisabled) {
		flags = append(flags, "DISABLED")
	}

	// ── assemble widget line ──────────────────────────────────────────────────
	line := indent + "[" + tag + "]"
	if summary != "" {
		line += " " + summary
	}
	line += "  " + bounds
	if len(flags) > 0 {
		line += "  [" + strings.Join(flags, ", ") + "]"
	}
	fmt.Fprintln(w, line)

	// ── optional style line ───────────────────────────────────────────────────
	if opt.Style {
		state := widget.State()
		sel := ""
		if state != "" {
			sel = ":" + state
		}
		style := widget.Style(sel)
		fmt.Fprintln(w, indent+"  style: "+formatStyle(style))
	}

	// ── recurse into children (always, including hidden) ──────────────────────
	if container, ok := widget.(Container); ok {
		for _, child := range container.Children() {
			dumpNode(w, child, depth+1, opt)
		}
	}
}

// formatStyle returns a compact single-line description of a style's
// border, padding, margin, foreground, and background.
func formatStyle(s *Style) string {
	var parts []string

	if b := s.Border(); b != "" {
		// Border spec may include color after a space ("thin $fg2"); keep only the type.
		parts = append(parts, "border="+strings.Fields(b)[0])
	}

	if p := s.Padding(); p != nil && p.Horizontal()+p.Vertical() != 0 {
		parts = append(parts, "pad="+formatInsets(p))
	}

	if m := s.Margin(); m != nil && m.Horizontal()+m.Vertical() != 0 {
		parts = append(parts, "margin="+formatInsets(m))
	}

	if fg := s.Foreground(); fg != "" {
		parts = append(parts, "fg="+fg)
	}

	if bg := s.Background(); bg != "" {
		parts = append(parts, "bg="+bg)
	}

	if len(parts) == 0 {
		return "(none)"
	}
	return strings.Join(parts, "  ")
}

// formatInsets formats an Insets value compactly:
// all-zero → "0", uniform → "N", symmetric → "TB,LR", full → "T,R,B,L".
func formatInsets(i *Insets) string {
	if i == nil || (i.Top == 0 && i.Right == 0 && i.Bottom == 0 && i.Left == 0) {
		return "0"
	}
	if i.Top == i.Right && i.Right == i.Bottom && i.Bottom == i.Left {
		return fmt.Sprintf("%d", i.Top)
	}
	if i.Top == i.Bottom && i.Left == i.Right {
		return fmt.Sprintf("%d,%d", i.Top, i.Left)
	}
	return fmt.Sprintf("%d,%d,%d,%d", i.Top, i.Right, i.Bottom, i.Left)
}
