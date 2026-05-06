package core

import "io"

// LayoutForm captures a single child's Add-parameters on a particular
// container (e.g. a Grid cell's coordinates). One layout form per
// (parent, child) pair.
//
// LayoutForm lives in core rather than inspector because its method
// signatures appear in the public API of widget *Form structs (the
// return type of GridForm.LayoutForm and friends), and the widgets
// package must not depend on inspector. Container forms in widgets/
// return values of this type so the inspector can drive them
// structurally without an import-cycle workaround.
type LayoutForm interface {
	Load(parent Container, child Widget)
	Store(parent Container, child Widget)
	Validate(field string) error

	// Emit writes the chain prefix placed in front of the child's
	// own call. In Builder mode this is typically a single chained
	// method like ".Cell(0, 1, 1, 1)" — leading ".\n" since the
	// child's constructor follows on the same chain. Indentation is
	// gofmt's responsibility.
	Emit(w io.Writer, mode string) error
}
