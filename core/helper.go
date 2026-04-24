package core

import (
	"errors"
	"fmt"
)

// ---- Container Helper Functions -------------------------------------------
// This file groups free functions that operate on the widget tree. They are
// deliberately kept as standalone helpers rather than methods on Container so
// they can be composed and extended without changing the interface, and so
// that they continue to work uniformly across all container implementations.

// Find locates the first widget whose ID matches id, searching the subtree
// rooted at container in depth-first order. The container itself is
// considered as the first candidate, so calling Find with the root of the
// widget tree will find any widget in the UI. The search short-circuits on
// the first match.
//
// Parameters:
//   - container: Root of the subtree to search.
//   - id:        The identifier to match; comparison is exact.
//
// Returns:
//   - Widget: The matching widget, or nil when no widget in the subtree
//     carries the requested ID.
func Find(container Container, id string) Widget {
	if container.ID() == id {
		return container
	}
	for _, child := range container.Children() {
		if child.ID() == id {
			return child
		}
		inner, ok := child.(Container)
		if ok {
			widget := Find(inner, id)
			if widget != nil {
				return widget
			}
		}
	}
	return nil
}

// MustFind locates the widget with the given id and returns it as T. It is
// a strict counterpart to Find + type assertion intended for call sites
// where the caller built the widget tree and can guarantee the lookup
// succeeds: a mismatch indicates an invariant violation (renamed ID,
// refactored widget type, typo) rather than a runtime condition.
//
// MustFind panics with a descriptive message when:
//   - no widget in the subtree has the requested id, or
//   - the matched widget cannot be asserted to T.
//
// The diagnostic includes the id and both the actual and expected types,
// which makes failures far easier to debug than the raw `interface
// conversion: <nil>` panic produced by a plain `Find(c, id).(T)` cast.
//
// Type parameters:
//   - T: The expected widget type, usually a pointer type such as *Input.
//
// Parameters:
//   - c:  Root of the subtree to search.
//   - id: Identifier of the target widget.
//
// Returns:
//   - T: The matching widget, asserted to T.
func MustFind[T Widget](c Container, id string) T {
	w := Find(c, id)
	if w == nil {
		panic(fmt.Sprintf("MustFind: no widget with id=%q", id))
	}
	t, ok := w.(T)
	if !ok {
		var zero T
		panic(fmt.Sprintf("MustFind: widget with id=%q has type %T, expected %T", id, w, zero))
	}
	return t
}

// FindAll returns every descendant of container that satisfies the type
// assertion to T, collected in depth-first visit order. T is typically a
// concrete widget pointer type. The container itself is not included in the
// result, even if it matches.
//
// The returned slice is non-nil; an empty subtree — or one with no matches —
// yields a slice with length zero. Allocation is lazy: nothing is allocated
// until the first match is found.
//
// Type parameters:
//   - T: The target type to filter by.
//
// Parameters:
//   - container: Root of the subtree to search.
//
// Returns:
//   - []T: All descendants that satisfy the type assertion to T.
func FindAll[T any](container Container) []T {
	var result []T
	Traverse(container, func(widget Widget) bool {
		if val, ok := widget.(T); ok {
			result = append(result, val)
		}
		return true
	})
	return result
}

// FindAt returns the deepest widget whose bounds contain the screen
// coordinate (x, y). It walks the widget tree top-down, descending into
// every child whose bounds enclose the point, so the result is the most
// specific visible widget under the point. If (x, y) lies within container
// but no child matches, container itself is returned as the fallback hit.
//
// Widgets marked with FlagHidden are skipped; containers whose own bounds
// do not contain the point cause the recursion to stop, so hidden subtrees
// never participate in hit testing. This makes FindAt suitable for
// translating mouse events into focus or click targets.
//
// Parameters:
//   - container: Root of the subtree to hit-test.
//   - x, y:      Screen coordinates, in the same coordinate system as
//     Widget.Bounds().
//
// Returns:
//   - Widget: The most specific widget whose bounds contain the point, or
//     nil when the point lies outside container.
func FindAt(container Container, x, y int) Widget {
	cx, cy, cw, ch := container.Bounds()

	// Check if it is inside the bounds
	if x < cx || y < cy || x >= cx+cw || y >= cy+ch {
		return nil
	}

	for _, child := range container.Children() {
		visible := !child.Flag(FlagHidden)
		if !visible {
			continue
		}
		cx, cy, cw, ch = child.Bounds()
		if x >= cx && y >= cy && x < cx+cw && y < cy+ch {

			inner, ok := child.(Container)
			if ok {
				widget := FindAt(inner, x, y)
				if widget != nil {
					return widget
				}
			}
			return child
		}
	}

	return container
}

// Layout invokes Container.Layout on each direct child of container that is
// itself a Container. Children that are leaf widgets are left untouched;
// callers are expected to have already laid out container itself before
// invoking this helper.
//
// Errors from individual Layout calls are logged through the parent
// container's logger but do not abort the iteration — every child is
// visited regardless of earlier failures. All collected errors are
// joined with errors.Join and returned as a single value; callers that
// need to inspect individual failures can use errors.Is or iterate via
// the Unwrap() []error method on the joined error.
//
// Parameters:
//   - container: The container whose direct child containers should be
//     laid out.
//
// Returns:
//   - error: A joined error containing every child's failure, or nil
//     when every child container succeeded.
func Layout(container Container) error {
	var errs []error
	for _, child := range container.Children() {
		if inner, ok := child.(Container); ok {
			if err := inner.Layout(); err != nil {
				container.Log(inner, Error, "Layout failed", "error", err)
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

// Traverse walks the subtree rooted at container, invoking fn on every
// direct child in depth-first order. The container itself is not passed to
// fn — the function is called for descendants only, so a call always
// visits at least one level below the root.
//
// fn's return value gates descent, not continuation: returning true means
// "visit this widget's children too", while returning false prunes the
// subtree below the current widget but lets iteration continue with its
// siblings. In other words, returning false cannot abort the whole walk;
// to stop early, capture a flag in the closure and have fn return false
// once the flag is set.
//
// Parameters:
//   - container: Root of the subtree to traverse.
//   - fn:        Visitor invoked for each descendant. Return true to
//     descend into the widget's children, false to skip them.
func Traverse(container Container, fn func(Widget) bool) {
	for _, child := range container.Children() {
		if !fn(child) {
			continue
		}
		if inner, ok := child.(Container); ok {
			Traverse(inner, fn)
		}
	}
}
