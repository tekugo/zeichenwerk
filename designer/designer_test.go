package designer_test

import (
	"bytes"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	. "github.com/tekugo/zeichenwerk/core"
	"github.com/tekugo/zeichenwerk/designer"
	. "github.com/tekugo/zeichenwerk/widgets"
)

// registerKinds wires the four widget kinds the POC drivers care
// about, so the test exercises the same registry the inspector-poc
// uses. Failures here would catch a mis-registration before it
// propagates into the driver.
func registerKinds(t *testing.T, d *designer.Designer) {
	t.Helper()
	register := func(typ reflect.Type, mk func() designer.WidgetForm) {
		t.Helper()
		if err := d.Register(designer.Kind{Type: typ, Make: mk}); err != nil {
			t.Fatalf("register %s: %v", typ, err)
		}
	}
	register(reflect.TypeOf((*Static)(nil)),
		func() designer.WidgetForm { return &StaticForm{} })
	register(reflect.TypeOf((*Grid)(nil)),
		func() designer.WidgetForm { return &GridForm{} })
	register(reflect.TypeOf((*Flex)(nil)),
		func() designer.WidgetForm { return &FlexForm{} })
	register(reflect.TypeOf((*Input)(nil)),
		func() designer.WidgetForm { return &InputForm{} })
}

// TestGenerateFragment_RoundTrip generates Builder source for a few
// canned trees and verifies:
//  1. The output is syntactically valid Go (parses with go/parser as
//     a single expression);
//  2. The expected widget call names appear in the right order;
//  3. Container closes carry the "// Kind#id" marker comment.
func TestGenerateFragment_RoundTrip(t *testing.T) {
	type expectation struct {
		name string  // human-readable test case label
		root func() Container // tree builder
		// callsInOrder lists method/identifier names that must
		// appear in this exact order somewhere in the output.
		// Each is matched against the formatted text via simple
		// substring search; the index of each match must
		// monotonically increase.
		callsInOrder []string
		// markers lists "// Kind#id" comment fragments that
		// must appear in the output.
		markers []string
	}

	cases := []expectation{
		{
			name: "single static",
			root: func() Container {
				root := NewFlex("root", "", Stretch, 0)
				_ = root.Add(NewStatic("hello", "", "Hello"))
				return root
			},
			callsInOrder: []string{
				"NewBuilder(theme)",
				`HFlex("root"`,
				`Static("hello", "Hello")`,
				"End()",
			},
			// Markers are emitted on container End() only;
			// Static is a leaf, so no "// Static#hello".
			markers: []string{
				"// Flex#root",
			},
		},
		{
			name: "grid with two cells",
			root: func() Container {
				root := NewFlex("outer", "", Stretch, 0)
				g := NewGrid("g", "", 2, 2, false)
				g.Add(NewStatic("a", "", "A"), 0, 0, 1, 1)
				g.Add(NewStatic("b", "highlight", "B"), 1, 0, 1, 1)
				_ = root.Add(g)
				return root
			},
			callsInOrder: []string{
				`Grid("g"`,
				`Cell(0, 0, 1, 1)`,
				`Static("a", "A")`,
				`Cell(1, 0, 1, 1)`,
				`Class("highlight")`,
				`Static("b", "B")`,
				"End()",
			},
			markers: []string{
				"// Grid#g",
				"// Flex#outer",
			},
		},
		{
			name: "vflex picks VFlex constructor",
			root: func() Container {
				root := NewFlex("col", "", Center, 1)
				root.SetFlag(FlagVertical, true)
				_ = root.Add(NewInput("name", "", "", "your name…"))
				return root
			},
			callsInOrder: []string{
				`VFlex("col"`,
				`Input("name", "", "your name…")`,
				"End()",
			},
			markers: []string{
				"// Flex#col",
			},
		},
		{
			name: "non-fixed style emits chain",
			root: func() Container {
				root := NewFlex("frame", "", Stretch, 0)
				s := NewStatic("hi", "", "Hi")
				// Install a non-fixed style override so
				// StyleForm.fixed is false and
				// EmitBuilderChain produces output.
				style := NewStyle("").
					WithPadding(0, 1).
					WithBackground("$blue")
				s.SetStyle("", style)
				_ = root.Add(s)
				return root
			},
			callsInOrder: []string{
				`Static("hi", "Hi")`,
				`Padding(0, 1)`,
				`Background("$blue")`,
			},
			markers: []string{
				"// Flex#frame",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := designer.NewDesigner(tc.root())
			registerKinds(t, d)

			var buf bytes.Buffer
			if err := d.GenerateFragment(designer.ModeBuilder, &buf); err != nil {
				t.Fatalf("GenerateFragment: %v", err)
			}
			out := buf.String()

			// (1) Output must parse as a Go expression. Wrap
			// in a synthetic file so go/parser has something
			// to chew.
			src := "package _expr_test\n\nvar _ = " + out + "\n"
			if _, err := parser.ParseFile(token.NewFileSet(), "", src, 0); err != nil {
				t.Fatalf("output does not parse:\n%s\n--- error ---\n%v", out, err)
			}

			// (2) Listed calls must appear in order.
			pos := 0
			for _, want := range tc.callsInOrder {
				idx := strings.Index(out[pos:], want)
				if idx < 0 {
					t.Errorf("expected call %q not found at or after pos %d in:\n%s",
						want, pos, out)
					continue
				}
				pos += idx + len(want)
			}

			// (3) Each "// Kind#id" marker comment must
			// appear somewhere in the output.
			for _, marker := range tc.markers {
				if !strings.Contains(out, marker) {
					t.Errorf("expected marker %q not found in:\n%s", marker, out)
				}
			}
		})
	}
}

// TestGenerateFile_BuildsValidSource checks that GenerateFile produces
// a complete, parseable Go source file (package decl + imports +
// func body) for a small tree.
func TestGenerateFile_BuildsValidSource(t *testing.T) {
	root := NewFlex("root", "", Stretch, 0)
	_ = root.Add(NewStatic("hi", "", "hi"))

	d := designer.NewDesigner(root)
	registerKinds(t, d)

	var buf bytes.Buffer
	if err := d.GenerateFile(designer.ModeBuilder, &buf, "demo", "BuildUI"); err != nil {
		t.Fatalf("GenerateFile: %v", err)
	}
	out := buf.String()

	if _, err := parser.ParseFile(token.NewFileSet(), "", out, 0); err != nil {
		t.Fatalf("file does not parse:\n%s\n--- error ---\n%v", out, err)
	}

	for _, want := range []string{
		"package demo",
		"func BuildUI(",
		"NewBuilder(theme)",
		`Static("hi", "hi")`,
		"Build()",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in:\n%s", want, out)
		}
	}
}

// TestRegister_RejectsMismatch ensures a Kind whose Make produces a
// form whose New returns the wrong widget type is rejected at
// registration time, not when Load is later called.
func TestRegister_RejectsMismatch(t *testing.T) {
	d := designer.NewDesigner(NewFlex("r", "", Stretch, 0))
	err := d.Register(designer.Kind{
		Type: reflect.TypeOf((*Static)(nil)),
		Make: func() designer.WidgetForm { return &GridForm{} }, // wrong form
	})
	if err == nil {
		t.Fatalf("expected error registering *Static against &GridForm{}, got nil")
	}
	if !strings.Contains(err.Error(), "Register") {
		t.Errorf("error message should mention Register: %v", err)
	}
}
