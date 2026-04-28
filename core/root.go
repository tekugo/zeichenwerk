package core

// Root is the top-level container of a running UI. It extends Container
// with the capabilities that only make sense at the application root:
// managing focus, owning the active theme, tearing down the screen on
// shutdown, and opening modal popups layered over the main tree.
//
// Concrete implementations live outside the core package (typically in a
// root or ui package). Widgets reach the root by walking Parent() up to a
// widget that also satisfies Root, which is how individual widgets request
// focus changes, repaints, or theme lookups without holding a direct
// reference to the application shell.
type Root interface {
	Container

	// Close releases the terminal screen and any resources held by the
	// root, typically stopping the event loop. After Close returns, the
	// root is no longer usable.
	Close()

	// Focus moves keyboard focus to the given widget. Implementations are
	// expected to emit blur/focus events on the previously and newly
	// focused widgets respectively.
	Focus(widget Widget)

	// Popup overlays the given container on top of the main widget tree at
	// the requested position and size. The popup is responsible for its
	// own layout within those bounds and remains visible until dismissed
	// by the application.
	Popup(x, y, w, h int, container Container)

	// Redraw marks the given widget dirty so that the next rendering pass
	// repaints only that widget's bounds. Passing the root itself requests
	// a full redraw. This is the optimized counterpart to Widget.Refresh,
	// which defaults to a full-screen refresh: use Redraw when the caller
	// knows that only the widget's own bounds need to change on screen.
	Redraw(widget Widget)

	// Theme returns the currently active theme, used by widgets during
	// Apply to resolve styles and by the renderer to resolve colour
	// variables and named borders.
	Theme() *Theme
}
