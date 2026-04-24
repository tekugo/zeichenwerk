// Package zeichenwerk provides the Zeichenwerk terminal UI toolkit.
//
// Zeichenwerk (German for "character works") is a complete Terminal UI toolkit
// for building interactive terminal applications in Go. It offers two APIs for
// constructing UIs: a fluent builder and a functional composition API.
//
// # Builder API
//
// The [Builder] type provides a chainable, method-based API. Calls return the
// builder itself so the entire layout can be expressed as a single expression.
// Child scopes are opened implicitly and closed with End().
//
//	func main() {
//	    NewBuilder(TokyoNightTheme()).
//	        Flex("root", false, Stretch, 0).
//	            Flex("header", true, Center, 1).
//	                Static("title", "My App").Font("bold").Foreground("$cyan").
//	            End().
//	            Grid("body", 1, 2, false).Columns(20, -1).
//	                Cell(0, 0, 1, 1).List("nav", "Home", "Settings", "About").
//	                Cell(1, 0, 1, 1).With(content).
//	            End().
//	        End().
//	        Run()
//	}
//
//	func content(b *Builder) {
//	    b.Static("body-text", "Select an item from the menu.")
//	}
//
// # Composition API
//
// The compose sub-package (github.com/tekugo/zeichenwerk/compose) provides a
// functional alternative. Every widget is an Option — a plain function value —
// that can be nested directly or passed around as data. The theme flows through
// the tree automatically so no global state is required.
//
//	import (
//	    z  "github.com/tekugo/zeichenwerk"
//	    .  "github.com/tekugo/zeichenwerk/compose"
//	)
//
//	func main() {
//	    UI(z.TokyoNightTheme(),
//	        Flex("root", "", false, Stretch, 0,
//	            Flex("header", "", true, Center, 1,
//	                Static("title", "", "My App", Font("bold"), Fg("$cyan")),
//	            ),
//	            Grid("body", "", []int{0}, []int{20, -1}, false,
//	                Cell(0, 0, 1, 1, List("nav", "", []string{"Home", "Settings", "About"})),
//	                Cell(1, 0, 1, 1, Include(content)),
//	            ),
//	        ),
//	    ).Run()
//	}
//
//	func content(theme *z.Theme) z.Widget {
//	    return Build(theme, Static("body-text", "", "Select an item from the menu."))
//	}
//
// Where direct widget access is needed after construction — to wire events,
// populate a Tree, or start animations — retrieve the widget with [Find] and
// call methods on it directly.
//
// # Widgets
//
// Nearly everything, including the root [UI], implements the [Widget] interface,
// which is fully provided by the embedded [Component] type. Creating new widgets
// is straightforward: embed Component and implement Render.
//
// For simple examples see [Static] and [Button]; for containers see [Flex],
// [Grid], and [Switcher].
package zeichenwerk
