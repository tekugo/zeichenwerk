# Zeichenwerk Tutorial

A guided tour of building terminal UIs with **zeichenwerk**, from a 20-line
"hello world" to a working SQLite query tool.

## Who this is for

Go developers who want to build interactive terminal applications without
hand-rolling a tcell event loop, focus engine, layout system, and rendering
pipeline. No prior TUI experience required; basic Go familiarity is assumed.

## What you'll build

By the end of this tutorial you'll have written:

- A handful of small example programs (one per concept).
- A real SQLite query tool with a schema list, SQL editor, and result table —
  the kind of app you'd actually keep in your toolbelt.

## How to read this

Each chapter is a single Markdown file. Every code snippet you see is also
shipped as a runnable Go program under [`examples/`](examples/) — clone the
repo and run them as you go.

```bash
go run ./doc/tutorial/examples/01-teaser
go run ./doc/tutorial/examples/03-layout-flex
go run ./doc/tutorial/examples/09-sqlite/step-5-wired
```

Press `q` or `Ctrl-Q` to quit any example.

## API choice

The tutorial uses the **Builder** API throughout. zeichenwerk also ships a
functional [`compose`](../../compose) API that produces equivalent UIs; the
patterns transfer one-to-one. Pick whichever style fits your taste once you
know the framework.

## Table of contents

### Part I — Teaser

1. [A TUI in 20 lines](01-teaser.md) — what a minimal app looks like and what
   the framework does for you.

### Part II — Concepts

2. [Architecture & lifecycle](02-architecture.md) — the four layers, frame
   pipeline, and the build-then-wire pattern.
3. [Layout](03-layout.md) — Flex, Grid, and `Hint` semantics.
4. [Styling & themes](04-styling-themes.md) — theme variables, selectors,
   classes.
5. [Events & focus](05-events-focus.md) — handlers, focus traversal, redraw
   vs. relayout.

### Part III — Widgets in depth

6. [Widget tour](06-widget-tour.md) — quick catalog of every widget.
7. [Containers](07-containers.md) — Switcher, Tabs, Forms, Dialogs.
8. [Custom widgets](08-custom-widgets.md) — extending `Component` with your
   own render code.

### Part IV — Real-world example

9. [Building a SQLite query tool](09-sqlite-query-tool.md) — a full
   application, six incremental steps.

### Appendices

- [Debugging](A-debugging.md) — `.Debug()`, the inspector, the log panel.
- [Cheatsheet](B-cheatsheet.md) — one-page reference.

## Prerequisites

```bash
go get github.com/tekugo/zeichenwerk/v2
```

Go 1.26+ is required. The SQLite tutorial additionally needs `mattn/go-sqlite3`
(pulled in automatically when you run that example).

## Where to go from here

- **Widget reference** — [`doc/reference/overview.md`](../reference/overview.md)
- **Design principles** — [`doc/principles.md`](../principles.md)
- **Showcase app** — [`cmd/showcase/main.go`](../../cmd/showcase/main.go)
