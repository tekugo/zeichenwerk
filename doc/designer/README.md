# Designer

The designer is zeichenwerk's interactive layout editor. It runs as a popup
on top of a live UI tree and edits that tree directly: every change made
through a form is written back to the running widget, so the underlying
preview reflects edits immediately.

This directory documents both the runtime side (the inspector framework and
the `designer-poc` driver) and the per-widget editing surface
(`*-form.go` files in the `widgets` package).

## Documents

- [overview.md](overview.md) — the moving parts: `inspector.Designer`,
  `WidgetForm`, `ContainerForm`, `LayoutForm`, the registry, and how
  the popup wires them together.
- [forms.md](forms.md) — how to write a `*-form.go` file for a new
  widget kind. Covers the editing surface, struct tags, codegen
  (`Emit`), the standard `EmitFrame` helper, and special cases
  (containers with per-child params, runtime-only dependencies).
- [codegen.md](codegen.md) — what `Designer.GenerateFragment` and
  `Designer.GenerateFile` produce, the chain-element convention, and
  the difference between `ModeBuilder` and `ModeCompose`.
- [designer-poc.md](designer-poc.md) — guided tour of the
  `cmd/designer-poc` driver: how it builds the popup, what each tab
  does, how Apply / Reset / Generate are wired, and how to extend it
  with new widget kinds or new tabs.

## At a glance

```text
┌──────────────────── designer popup ───────────────────────┐
│ Header  file • • TokyoNight    [Save] [Generate] [Run]   │
├──────────────────────┬────────────────────────────────────┤
│ Tree                 │ Tabs: General | Layout | Style | …│
│  ▼ VFlex (#root)     │  ┌──────────────────────────────┐ │
│    ▶ HFlex (#header) │  │ ID            [main       ]  │ │
│    ▶ Grid (#g1)      │  │ Class         [           ]  │ │
│                      │  │ Hint W / H    [0 ] [0 ]      │ │
│                      │  │ Vertical      [x]            │ │
│                      │  │ Alignment     [stretch ▼]    │ │
│                      │  │ Spacing       [0 ]           │ │
│                      │  └──────────────────────────────┘ │
│ [Add] [Del] [Up] …   │  [Apply]  [Reset]                 │
├──────────────────────┴────────────────────────────────────┤
│ status: applied → Flex#root                              │
└────────────────────────────────────────────────────────────┘
```

## Quick start

```bash
go run ./cmd/designer-poc          # open the demo
# Ctrl+Space  open / close the designer popup
# Alt+1..4    jump to General / Layout / Style / Info tab
# Apply       commit form edits to the live widget
# Generate    write Builder-mode source to /tmp/designer-poc-out.go
# Ctrl+Q      quit
```

The designer is a **popup driver around a generic framework**. The
inspector package owns the state (`inspector.Designer`, kind registry,
codegen walker); the driver owns the UX (popup layout, button handlers,
status bar). Anyone can write their own driver — the registry and the
form interfaces are stable.
