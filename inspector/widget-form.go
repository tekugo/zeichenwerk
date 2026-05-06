// Package inspector hosts the runtime developer console.
//
// The form-driven editing and codegen surface is rooted in the
// WidgetForm and ContainerForm interfaces and the ModeBuilder /
// ModeCompose codegen-mode constants defined in this file. Concrete
// *Form structs that know each widget's internals live in the widgets
// package alongside the widgets they edit so they can read and write
// unexported fields directly; they satisfy these interfaces
// structurally and never reference the inspector package by name.
//
// LayoutForm — the per-child Add-params form returned by container
// forms — lives in core, not inspector, so widget container forms
// can declare it as a return type without importing inspector.
package inspector

import (
	"io"

	"github.com/tekugo/zeichenwerk/core"
)

// Codegen mode strings; passed to Emit on the various form interfaces.
const (
	ModeBuilder = "builder"
	ModeCompose = "compose"
)

// WidgetForm is the editing + codegen surface for one widget kind.
//
// One implementation per registered kind; the implementations live in
// the widgets package alongside the widget they edit so they can read
// and write unexported fields directly.
type WidgetForm interface {
	Name() string
	Group() string
	Help() string

	New() core.Widget
	Load(core.Widget)
	Store(core.Widget)

	Validate(field string) error

	// Style returns a pointer to the form's style snapshot — the
	// editable surface the inspector renders separately from the
	// kind-specific fields. The snapshot is loaded by Load and
	// flushed back to the widget's *core.Style by Store, so edits
	// made through the returned pointer survive across calls
	// without any explicit synchronisation. The implementation
	// lives on the embedded ComponentForm; concrete kind forms
	// inherit it for free.
	Style() *core.StyleForm

	// Emit writes the widget's call shape onto an in-progress
	// chain. The caller has already written whatever precedes this
	// widget; an Emit implementation continues with leading ".\n"
	// so the chain stays continuous. Indentation is the
	// responsibility of go/format.Source, which Designer runs over
	// the whole emitted string before returning. Containers emit
	// only the constructor + chain + style; the codegen walker
	// writes children and the closing ".End()" with a trailing
	// "// Kind#id" marker.
	Emit(w io.Writer, mode string) error
}

// ContainerForm is the optional extension implemented by forms whose
// widget consumes per-child Add parameters (e.g. Grid's cell
// coordinates, FormGroup's line/label). Containers that ignore Add
// params (Box, Flex, Card, …) implement only WidgetForm.
type ContainerForm interface {
	WidgetForm

	// LayoutForm returns a fresh per-child layout form already loaded
	// with child's current Add-params on parent. Returning nil is
	// allowed and signals "no per-child params for this child".
	//
	// The return type is core.LayoutForm so widget container forms can
	// declare this method without importing inspector.
	LayoutForm(parent core.Container, child core.Widget) core.LayoutForm
}
