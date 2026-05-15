# Rule

Horizontal or vertical line for visual separation.

**Constructors:**

- `NewHRule(class, style string) *Rule` — horizontal rule (id `"hrule"`, fixed height 1)
- `NewVRule(class, style string) *Rule` — vertical rule (id `"vrule"`, fixed width 1)

`style` is a theme-registered border name (`"thin"`, `"thick"`, `"double"`, `"dashed"`, `"lines"`, …). The rule renders the border's inner-horizontal or inner-vertical character along its length. If the theme has no matching border, falls back to the `"default"` border, then skips rendering — a missing theme asset degrades silently.

## Notes

The id is hard-coded by the constructors. Pass through `Builder.HRule(style)` / `VRule(style)` if you don't need to customise the id.

Rules dispatch no events and accept no input.
