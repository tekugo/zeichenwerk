package inspector

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"reflect"
	"sort"

	"github.com/tekugo/zeichenwerk/core"
)

// Designer owns the registry mapping widget concrete types to the
// WidgetForm factories that edit them, plus the codegen entry point.
//
// The Designer keeps the inspector's internals out of the widgets
// package: widgets contribute *Form structs that satisfy WidgetForm
// structurally, and the driver wires them up by calling Register on
// each kind it cares about.
//
// Designer is not safe for concurrent use; callers serialise access.
// In a TUI host this is naturally satisfied by the main event loop —
// every Register, FormFor, Add, and Generate call runs on the same
// goroutine.
type Designer struct {
	target core.Container

	// Per-Type registry. byType is the lookup index used by
	// FormFor / Add / the codegen walker; kinds preserves
	// registration order, which is also the picker ordering after
	// the lexicographic Kinds() / KindNames() sort.
	byType map[reflect.Type]Kind
	kinds  []Kind
}

// NewDesigner returns an empty Designer pointed at target. The driver
// is responsible for calling Register on each widget kind to populate
// the registry.
func NewDesigner(target core.Container) *Designer {
	return &Designer{
		target: target,
		byType: map[reflect.Type]Kind{},
	}
}

// Register validates and stores a Kind. The Kind is rejected with a
// non-nil error when:
//
//   - Type or Make is nil;
//   - Make() returns nil;
//   - Make().New() returns nil or a value whose concrete type does
//     not match Type. This catches the common driver bug of pairing
//     a Type with the wrong form (e.g. *Static registered against
//     &GridForm{}), which would otherwise panic during the first
//     Load.
//
// On success the Kind is filled in with metadata derived from the
// freshly-built form (Name / Group / Help) and is available via
// Kinds, KindNames, FormFor, and Add.
func (d *Designer) Register(k Kind) error {
	if k.Type == nil {
		return fmt.Errorf("inspector: Register: Type is nil")
	}
	if k.Make == nil {
		return fmt.Errorf("inspector: Register: Make is nil")
	}
	form := k.Make()
	if form == nil {
		return fmt.Errorf("inspector: Register: Make returned nil for %s", k.Type)
	}
	w := form.New()
	if w == nil {
		return fmt.Errorf("inspector: Register: form.New returned nil for %s", k.Type)
	}
	if got := reflect.TypeOf(w); got != k.Type {
		return fmt.Errorf("inspector: Register: form.New returned %s, expected %s", got, k.Type)
	}
	if k.Name == "" {
		k.Name = form.Name()
	}
	if k.Group == "" {
		k.Group = form.Group()
	}
	if k.Help == "" {
		k.Help = form.Help()
	}
	d.byType[k.Type] = k
	d.kinds = append(d.kinds, k)
	return nil
}

// FormFor returns a fresh form pre-loaded from w, or nil if w's kind
// has no registered form.
func (d *Designer) FormFor(w core.Widget) WidgetForm {
	k, ok := d.byType[reflect.TypeOf(w)]
	if !ok {
		return nil
	}
	f := k.Make()
	f.Load(w)
	return f
}

// Kind returns the Kind metadata registered for w's type, or the
// zero Kind value if none is registered. Useful for capability
// checks (does this widget have a form?) without paying the Load
// cost FormFor incurs.
func (d *Designer) Kind(w core.Widget) Kind {
	return d.byType[reflect.TypeOf(w)]
}

// Kinds returns the registered kinds, sorted lexicographically by
// Name for stable picker order. The slice is a fresh copy; callers
// can mutate it without affecting the registry.
func (d *Designer) Kinds() []Kind {
	out := make([]Kind, len(d.kinds))
	copy(out, d.kinds)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// KindNames returns the registered kinds' names, sorted
// lexicographically for stable picker order.
func (d *Designer) KindNames() []string {
	out := make([]string, 0, len(d.kinds))
	for _, k := range d.kinds {
		out = append(out, k.Name)
	}
	sort.Strings(out)
	return out
}

// Add appends a fresh widget of the given kind to parent. The
// widget is created via the kind's WidgetForm.New() — i.e. with
// default values — and added through parent.Add without any
// per-child layout parameters. Containers that consume layout
// params (Grid) fall back to their own defaults; for Grid that
// means cell (0, 0, 1, 1). Callers wanting a specific position
// should follow up with the matching LayoutForm.
//
// Theme application is the caller's responsibility: Designer
// has no Theme reference and the new widget is untouched by any
// Apply call. Typical usage is:
//
//	w, err := d.Add(parent, "Static")
//	if err != nil { ... }
//	w.Apply(theme)
//	Relayout(w)
func (d *Designer) Add(parent core.Container, kindName string) (core.Widget, error) {
	if parent == nil {
		return nil, fmt.Errorf("inspector: Add: parent is nil")
	}
	for _, k := range d.kinds {
		if k.Name != kindName {
			continue
		}
		child := k.Make().New()
		if err := parent.Add(child); err != nil {
			return nil, fmt.Errorf("inspector: Add: parent.Add: %w", err)
		}
		return child, nil
	}
	return nil, fmt.Errorf("inspector: Add: no kind named %q", kindName)
}

// Generate writes Builder-mode source for d.target's tree to w.
// Deprecated wrapper around GenerateFragment for the legacy
// signature; new callers should use GenerateFragment or
// GenerateFile directly.
func (d *Designer) Generate(mode string, w io.Writer) error {
	return d.GenerateFragment(mode, w)
}

// GenerateFragment writes a chained Builder expression for
// d.target's subtree to w, starting with "NewBuilder(theme)" and
// ending with the trailing newline of the formatted output. The
// fragment is suitable for splicing into an existing function body
// as a single statement.
//
// Each emitted chain call ends with ".\n" (trailing-dot
// convention) so subsequent chain elements connect without needing
// a leading separator; the very last call (the root container's
// End) drops the trailing dot. Each container's closing ".End()"
// carries a trailing "// Kind#id" comment naming what it closes,
// which survives gofmt and gives the reader an anchor when the
// chain is otherwise visually flat.
//
// The whole emitted string is run through go/format.Source before
// being written, so callers always see canonical Go.
//
// Compose mode is reserved; today it returns an error.
func (d *Designer) GenerateFragment(mode string, w io.Writer) error {
	if mode != ModeBuilder {
		return fmt.Errorf("inspector: unsupported codegen mode: %q", mode)
	}
	var buf bytes.Buffer
	buf.WriteString("NewBuilder(theme).\n")
	if err := d.emit(&buf, d.target, nil); err != nil {
		return err
	}
	formatted, err := formatExpr(buf.Bytes())
	if err != nil {
		return fmt.Errorf("inspector: GenerateFragment: %w", err)
	}
	_, err = w.Write(formatted)
	return err
}

// GenerateFile writes a complete, self-contained Go source file for
// d.target's subtree to w: package declaration, the inferred import
// set, and a func wrapper named funcName whose body is the
// chained-Builder expression GenerateFragment would emit.
//
// pkg is the file's package name; funcName is the name of the
// builder-returning function. The function signature is
// "func <funcName>(theme *core.Theme) *zeichenwerk.UI"; the body
// returns the result of calling .Build() on the chain.
//
// As with GenerateFragment, the result is run through
// go/format.Source before being written.
//
// Compose mode is reserved; today it returns an error.
func (d *Designer) GenerateFile(mode string, w io.Writer, pkg, funcName string) error {
	if mode != ModeBuilder {
		return fmt.Errorf("inspector: unsupported codegen mode: %q", mode)
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", pkg)
	buf.WriteString("import (\n")
	buf.WriteString("\t. \"github.com/tekugo/zeichenwerk\"\n")
	buf.WriteString("\t. \"github.com/tekugo/zeichenwerk/core\"\n")
	buf.WriteString("\t. \"github.com/tekugo/zeichenwerk/widgets\"\n")
	buf.WriteString(")\n\n")
	fmt.Fprintf(&buf, "func %s(theme *Theme) *UI {\n", funcName)
	buf.WriteString("\treturn NewBuilder(theme).\n")
	if err := d.emit(&buf, d.target, nil); err != nil {
		return err
	}
	buf.WriteString("Build()\n}\n")
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("inspector: GenerateFile: %w\n--- raw output ---\n%s", err, buf.String())
	}
	_, err = w.Write(formatted)
	return err
}

// emit writes one widget (and its children if it is a container)
// to w. parent is the widget's parent container, used to look up
// a ContainerForm for layout-prefix emission.
//
// Each chain element ends with ".\n" (trailing-dot convention).
// Container closes are written as "End(). // Kind#id\n" for
// non-root containers (trailing dot continues the chain to the
// next sibling) and "End() // Kind#id\n" for the root container
// close (no trailing dot, the chain terminates). The placement of
// the dot before any trailing comment is what makes the chain
// gofmt-canonical: Go allows "Foo(). // comment" as a continuation
// but not "Foo() // comment\n.Bar()".
func (d *Designer) emit(w io.Writer, widget core.Widget, parent core.Container) error {
	// Layout-prefix from the parent's ContainerForm, if any.
	if parent != nil {
		if pf := d.FormFor(parent); pf != nil {
			if cf, ok := pf.(ContainerForm); ok {
				if lf := cf.LayoutForm(parent, widget); lf != nil {
					if err := lf.Emit(w, ModeBuilder); err != nil {
						return err
					}
				}
			}
		}
	}

	// The widget's own constructor + chain.
	wf := d.FormFor(widget)
	if wf == nil {
		_, err := fmt.Fprintf(w, "/* TODO: no form registered for %T */\n", widget)
		return err
	}
	if err := wf.Emit(w, ModeBuilder); err != nil {
		return err
	}

	// Children, then End() for containers.
	container, isContainer := widget.(core.Container)
	if isContainer {
		for _, child := range container.Children() {
			if err := d.emit(w, child, container); err != nil {
				return err
			}
		}
		marker := closingMarker(wf, widget)
		if parent == nil {
			// Root container: chain terminates here, no
			// trailing dot.
			if _, err := fmt.Fprintf(w, "End() %s\n", marker); err != nil {
				return err
			}
		} else {
			// Non-root container: trailing dot continues the
			// chain to the next sibling. The dot must come
			// before any "// comment" to keep Go syntax valid
			// (a "." after the comment would be part of the
			// comment).
			if _, err := fmt.Fprintf(w, "End(). %s\n", marker); err != nil {
				return err
			}
		}
	}
	return nil
}

// closingMarker formats the trailing comment placed on a
// container's closing End() — typically "// Kind#id" — so a reader
// can match the close back to its opener after gofmt flattens the
// chain. Falls back to "// Kind" when the widget has no id.
func closingMarker(wf WidgetForm, widget core.Widget) string {
	name := wf.Name()
	if id := widget.ID(); id != "" {
		return fmt.Sprintf("// %s#%s", name, id)
	}
	return fmt.Sprintf("// %s", name)
}

// formatExpr wraps a free-standing expression in a synthetic file,
// formats it, then strips the wrapper. Used by GenerateFragment to
// produce gofmt-canonical output for an expression that isn't a
// complete file by itself.
func formatExpr(expr []byte) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("package _expr\n\nvar _ = ")
	buf.Write(expr)
	buf.WriteByte('\n')
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%w\n--- raw output ---\n%s", err, buf.String())
	}
	// Strip "package _expr\n\nvar _ = " and the trailing newline.
	const prefix = "package _expr\n\nvar _ = "
	out := formatted
	if i := bytes.Index(out, []byte(prefix)); i >= 0 {
		out = out[i+len(prefix):]
	}
	out = bytes.TrimRight(out, "\n")
	out = append(out, '\n')
	return out, nil
}

