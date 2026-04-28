# Specification: triebwerk

**Binary:** `cmd/triebwerk/main.go`  
**Package:** `main`  
**Depends on:** `zeichenwerk`, `github.com/fsnotify/fsnotify`

---

## Purpose

`triebwerk` is a terminal UI for Makefile and Justfile targets. It provides
target discovery, one-key execution, live output streaming, and per-target
watch mode — all without leaving the terminal.

---

## Layout

```
┌─ triebwerk ── ~/Projects/myapp ──────────────────────────────────────────────┐
│ ┌─ Targets ───────────────────┐ ┌─ Output ───────────────────────────────┐ │
│ │                             │ │                                        │ │
│ │  [make] build               │ │  $ make build                          │ │
│ │         Build all binaries  │ │  go build ./...                        │ │
│ │                             │ │  ✓ exit 0  (1.2s)                      │ │
│ │  [make] test                │ │  ────────────────────────────────────  │ │
│ │         Run test suite      │ │  $ make build                          │ │
│ │                             │ │  go build ./...                        │ │
│ │  [just] fmt                 │ │                                        │ │
│ │         Format source       │ │                                        │ │
│ │                             │ │                                        │ │
│ │  [just] release             │ │                                        │ │
│ │         Tag and push        │ │                                        │ │
│ │                             │ │                                        │ │
│ │  filter  [x] make  [x] just │ │                                        │ │
│ └─────────────────────────────┘ └────────────────────────────────────────┘ │
│  ◉  ▶ running  [queue: 2]  │  r run  w watch  d dir  m make  j just  c clear  q quit │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Regions

| Region | Widget | Notes |
|--------|--------|-------|
| Header | `Static` | `triebwerk  —  <dir>` with `~/`-relative path |
| Left panel | `Flex` (vertical) | `Deck` (fills height) + filter section below |
| Filter section | `Checkbox` × 2 | `[x] make` / `[x] just`; toggled with `m` / `j` |
| Right panel | `Terminal` | Shared output; persists across runs |
| Footer left | `Scanner` + `Static` | `Scanner` pulses when watch is active; `Static` shows `▶ running` / `[queue: N]` |
| Footer right | `Shortcuts` | Colour-coded key/label pairs via the `Shortcuts` widget |

---

## Target Cards (Deck)

Each card renders two lines:

```
  [make] build
         Build all binaries
```

- **Badge** `[make]` or `[just]` — styled with an accent colour distinct from
  the target name
- **Target name** — bold when focused
- **Description** — dimmed; sourced from `## comment` (Makefile) or doc comment
  (Justfile); empty string if none found
- Focused card gets a highlighted background (standard Deck focus behaviour)
- Selected/running card shows a `▶` indicator

---

## Footer

### Left — Scanner

The `Scanner` widget scrolls text continuously when active. Content:

```
● src/**/*.go  [queue: 2]
```

- `●` pulse indicator — accent colour when active, dimmed when idle
- Glob pattern of the currently watched target
- `[queue: N]` — number of pending runs; hidden when queue is empty
- Entire left section is hidden when no watch is active and queue is empty

### Right — Keybinding hints

Rendered via the `Shortcuts` widget:

```
r run   w watch   d dir   m make   j just   c clear   q quit
```

- When watch is active, `w watch` changes to `w stop`
- `m make` / `j just` reflect the active filter state visually via the checkboxes

---

## Keyboard Bindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate target list |
| `r` / `Enter` | Run selected target (enqueue if busy) |
| `w` | Toggle watch mode for selected target |
| `d` | Open directory switcher |
| `m` | Toggle make filter checkbox |
| `j` | Toggle just filter checkbox |
| `c` | Clear terminal output |
| `q` / `Ctrl+C` | Quit (waits for active process to finish or kills it) |

---

## Execution Model

### Process queue

- A `chan runRequest` with a small buffer (e.g. 8) acts as the run queue
- A single background goroutine drains the queue one request at a time
- Each request carries: target name, runner (`make`/`just`), working directory
- While a process is running, incoming requests (manual or watch-triggered) are
  appended to the queue; the footer shows `[queue: N]`
- Parallel watch triggers for the same target are deduplicated: if the target
  is already in the queue, the duplicate is dropped

### Terminal output

- Each run prints a separator line before its output if the terminal is non-empty:

  ```
  ── make build ────────────────────────────────────────
  ```

  Separator uses a `Rule`-style line with the target name inset; styled with
  `$bg2` / `$fg2`.
- After the process exits, a one-line summary is appended:

  ```
  ✓ exit 0  (1.4s)        (green on success)
  ✗ exit 2  (0.3s)        (red on failure)
  ```

- The terminal is **never** auto-cleared between runs; `c` is the only way to
  clear it.

### Watch mode

1. User selects a target and presses `w`
2. If no glob is stored for this target in `.triebwerk.json`, a `UI.Prompt`
   dialog opens asking for a glob pattern (e.g. `**/*.go`)
3. Pattern is saved to `.triebwerk.json` immediately
4. `fsnotify.Watcher` resolves matching files via `filepath.Glob` / `doublestar`
   and registers them
5. On any `Write` or `Create` event matching the glob: enqueue a run
6. Pressing `w` again stops the watcher for that target
7. Only one watcher is active per target at a time; switching targets does not
   stop other watchers

---

## Configuration — `.triebwerk.json`

Stored in the project root (working directory). Pretty-printed. Created on
first watch pattern save; safe to commit.

```json
{
  "targets": {
    "build": {
      "watch": "src/**/*.go"
    },
    "test": {
      "watch": "**/*.go"
    }
  }
}
```

No other configuration is stored here. Directory history and window state are
not persisted in the MVP.

---

## Parsing

### Makefile

Regex-based; no shell-out to `make -p` in the MVP.

Rules:
1. A target line matches `^([a-zA-Z0-9_\-\.]+)\s*:` that is **not** `.PHONY`
   and does not start with `.`
2. A description comment directly above the target line matches `^##\s*(.+)`
3. Variables and includes are ignored
4. If the same target name appears in multiple included files, the first
   occurrence wins

### Justfile

Shell out to `just --list --list-format json` to get a structured list of
recipes, their parameters, and doc comments. Fall back to regex parsing if
`just` is not on `$PATH`.

### Auto-detect

On startup (and on directory change):
1. Look for `Makefile` (also `makefile`, `GNUmakefile`) → parse, badge `[make]`
2. Look for `Justfile` (also `justfile`) → parse, badge `[just]`
3. Merge into a single list: Makefile targets first, then Justfile targets
4. If neither file is found, show an empty state message in the Deck

---

## Directory Switcher

Activated by pressing `d`. Opens a `UI.Prompt` with a pre-filled path equal to
the current working directory. On accept:
1. Resolve and validate the path
2. Re-parse targets from the new directory
3. Stop all active watchers
4. Reset the queue
5. Update the header
6. Do **not** clear the terminal (output history is preserved)

---

## Implementation Steps

### Source file layout

| File | Responsibility |
|------|---------------|
| `cmd/triebwerk/main.go` | CLI flag parsing, `main()` entry point |
| `cmd/triebwerk/ui.go` | `buildUI()`, widget helpers (card render, hints render) |
| `cmd/triebwerk/model.go` | `Target` struct; placeholder data for Step 2 |
| `cmd/triebwerk/targets.go` | Makefile/Justfile parsers — added in Step 3 |
| `cmd/triebwerk/queue.go` | Run queue, subprocess execution — added in Step 4 |
| `cmd/triebwerk/watcher.go` | `fsnotify` wrapper, glob watching — added in Step 5 |
| `cmd/triebwerk/config.go` | `.triebwerk.json` load/save — added in Step 5 |

### Step 1 — Project scaffold ✓

- Created `cmd/triebwerk/main.go`, `ui.go`, `model.go`
- Added `github.com/fsnotify/fsnotify` to `go.mod`
- Added `github.com/bmatcuk/doublestar/v4` for glob matching
- `go build ./cmd/triebwerk` passes

### Step 2 — Static UI (no data, no logic) ✓

- Root: vertical `Flex` (header / body / footer)
- Header: `Static` showing `triebwerk — <cwd>`
- Body: horizontal `Flex`
  - `Box("Targets")` hint(34) wrapping `Deck` with 8 placeholder cards (3 rows each)
  - `Box("Output")` hint(-1) wrapping `Terminal` pre-seeded with fake output + separator
- Footer: horizontal `Flex`
  - `Static("footer-status")` — placeholder `● src/**/*.go`
  - `Spacer` (flexible fill)
  - `Custom("hints")` — colour-coded key/label pairs via `hint:key` / `hint:label`
- `hint:key` and `hint:label` added to all six themes (Tokyo, Midnight Neon, Nord,
  Gruvbox Dark, Gruvbox Light, Lipstick)

Acceptance: `go run ./cmd/triebwerk` renders the full layout, cards look right,
footer hints show correct colours, terminal shows fake output.

### Step 3 — Target parsing ✓

- `parseMakefile(path string) ([]Target, error)` — regex-based (skips `.`-prefixed
  directives; keeps `.PHONY` targets)
- `parseJustfile(dir string) ([]Target, error)` — `just --list --list-format json
  --unsorted` with regex fallback; private recipes (prefix `_`) skipped
- `loadTargets(dir string) ([]Target, error)` — auto-detects, merges; empty slice
  on no files found
- Wired into UI via `deck.SetItems(toItems(allTargets, dir))`; empty state shows
  placeholder card
- **Extra:** filter section (make/just checkboxes + `m`/`j` keys) added to left
  panel; `refilter` callback re-applies filters on toggle

### Step 4 — Execution ✓

- `Runner` struct owns `chan runRequest` (buffer 8) + background `drain` goroutine
- `crlfWriter` wraps `Terminal` to translate bare LF → CRLF for correct rendering
- Separator line (`── runner target ───…`) printed before each run if terminal
  non-empty; sized to terminal width
- Exit summary: `✓ exit 0  (N.Ns)` (green) / `✗ exit N  (N.Ns)` (red)
- Deduplication: `inQueue map[string]bool` prevents same target appearing in queue
  twice; used by both manual runs and watch triggers
- `r` / `Enter` enqueue; `c` calls `runner.ClearTerminal()` (also resets separator
  state via `written` atomic)
- Footer `Static` shows `▶ running`, `[queue: N]`, or both via `updateStatus()`

### Step 5 — Watch mode ✓

- `Watcher` wraps one `fsnotify.Watcher` per target (independent start/stop)
- Glob resolution via `doublestar.FilepathGlob`; all matching parent dirs registered
  plus root dir for new-file detection
- `.triebwerk.json` load/save implemented (`loadConfig` / `saveConfig` / `pattern`
  / `setPattern`)
- `w` key: stops watcher if active; otherwise prompts for glob (via `UI.Prompt`)
  if no saved pattern, then calls `watcher.Start`
- Watch trigger → `runner.Enqueue` (deduplication inherited from Runner)
- `runner.SetWatchActive(true/false)` starts/stops the `Scanner` animation
- `w watch` shortcut label toggles to `w stop` when watch is active

### Step 6 — Directory switcher

- Wire `d` key to `UI.Prompt` for directory input
- On accept: stop all watchers, flush queue, re-parse targets, update header
- Preserve terminal output history across directory switches

Acceptance: switching to a different project directory updates the target list
without clearing the terminal.

### Step 7 — Polish

- Quit confirmation if a process is running (`UI.Confirm`)
- Keyboard focus management (Deck vs Terminal)
- Resize handling (Terminal.Resize on window resize)
- Error states: unparseable Makefile, `just` not found, invalid glob
- README for `cmd/triebwerk`

---

## Dependencies

| Module | Purpose |
|--------|---------|
| `github.com/fsnotify/fsnotify` | Filesystem event watching |
| `github.com/bmatcuk/doublestar/v4` | `**` glob matching for watch patterns |

---

## Non-goals (MVP)

- No argument/variable injection before running targets
- No per-target environment variable overrides
- No parallel target execution
- No scroll-back in the Terminal beyond what the widget provides
- No remote/SSH project support
- No `make -j` job control
