# Custom

Widget with a user-supplied render function.

**Constructor:** `NewCustom(id, class string, fn func(Widget, *Renderer)) *Custom`

## Notes

Call `c.Component.Render(r)` inside `fn` to draw border and background before custom content.
