# Inspector

A debugging overlay that lets you explore the live widget hierarchy, inspect
bounds and styles, and browse the debug log.

**Constructor:** `NewInspector(root Container) *Inspector`

`root` is the container whose subtree will be inspected. Call `UI()` on the
returned inspector and pass it to `ui.Popup` (or add it to a layer) to show it.

## Methods

- `UI() Container` — returns the inspector's own UI container, suitable for passing to `ui.Popup`
- `Hint() (int, int)` — returns the preferred size of the inspector overlay
- `Refresh()` — re-reads the current container's children and updates the display
- `Activate(index int) bool` — navigates into the selected child container
- `SelectWidget(index int) bool` — shows details for the highlighted widget
- `SelectStyle(index int) bool` — shows details for the highlighted style selector

## Notes

The inspector has two built-in tabs:

- **Widgets** — tree navigation: select a child to see its bounds, hint, state,
  and style selectors; press Enter to descend into a container; press Backspace
  to ascend to the parent.
- **Debug Log** — scrollable table of all structured log entries emitted by the
  running UI.

No events are emitted. The inspector is a self-contained diagnostic tool and is
not intended for production UI trees.
