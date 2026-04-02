# Specification: lazymake

**Binary:** `cmd/lazymake/main.go`  
**Package:** `main`  
**Depends on:** `zeichenwerk`, `github.com/fsnotify/fsnotify`

---

## Purpose

`lazymake` is a terminal UI for Makefile and Justfile targets. It provides
target discovery, one-key execution, live output streaming, and per-target
watch mode — all without leaving the terminal.

---

## Layout

```
┌─ lazymake ── ~/Projects/myapp ──────────────────────────────────────────────┐
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
│ └─────────────────────────────┘ └────────────────────────────────────────┘ │
│  ● src/**/*.go  [queue: 2]  │  r run  w watch  d dir  c clear  q quit      │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Regions

| Region | Widget | Notes |
|--------|--------|-------|
| Header | `Static` | Project name + resolved absolute path; `[d]` hint |
| Left panel | `Deck` | Target cards; fills available height |
| Right panel | `Terminal` | Shared output; persists across runs |
| Footer left | `Scanner` | Watch glob + queue depth; hidden when idle |
| Footer right | `Static` / `Styled` | Colour-coded keybinding hints |

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

Each hint is rendered as two adjacent `Styled` spans:

```
r run   w watch   d dir   c clear   q quit
```

- **Key** (`r`, `w`, `d`, `c`, `q`) — accent colour, bold
- **Description** (`run`, `watch`, `dir`, `clear`, `quit`) — muted/dimmed foreground
- Hints are separated by three spaces
- When watch is active, `w watch` hint changes to `w stop` (same colours)

Style keys used:

```
"hint:key"    — accent colour, bold  (e.g. $accent or $green)
"hint:label"  — $fg2 (dimmed)
```

---

## Keyboard Bindings

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate target list |
| `r` / `Enter` | Run selected target (enqueue if busy) |
| `w` | Toggle watch mode for selected target |
| `d` | Open directory switcher |
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
2. If no glob is stored for this target in `.lazymake.json`, a `UI.Prompt`
   dialog opens asking for a glob pattern (e.g. `**/*.go`)
3. Pattern is saved to `.lazymake.json` immediately
4. `fsnotify.Watcher` resolves matching files via `filepath.Glob` / `doublestar`
   and registers them
5. On any `Write` or `Create` event matching the glob: enqueue a run
6. Pressing `w` again stops the watcher for that target
7. Only one watcher is active per target at a time; switching targets does not
   stop other watchers

---

## Configuration — `.lazymake.json`

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
| `cmd/lazymake/main.go` | CLI flag parsing, `main()` entry point |
| `cmd/lazymake/ui.go` | `buildUI()`, widget helpers (card render, hints render) |
| `cmd/lazymake/model.go` | `Target` struct; placeholder data for Step 2 |
| `cmd/lazymake/targets.go` | Makefile/Justfile parsers — added in Step 3 |
| `cmd/lazymake/queue.go` | Run queue, subprocess execution — added in Step 4 |
| `cmd/lazymake/watcher.go` | `fsnotify` wrapper, glob watching — added in Step 5 |
| `cmd/lazymake/config.go` | `.lazymake.json` load/save — added in Step 5 |

### Step 1 — Project scaffold ✓

- Created `cmd/lazymake/main.go`, `ui.go`, `model.go`
- Added `github.com/fsnotify/fsnotify` to `go.mod`
- Added `github.com/bmatcuk/doublestar/v4` for glob matching
- `go build ./cmd/lazymake` passes

### Step 2 — Static UI (no data, no logic) ✓

- Root: vertical `Flex` (header / body / footer)
- Header: `Static` showing `lazymake — <cwd>`
- Body: horizontal `Flex`
  - `Box("Targets")` hint(34) wrapping `Deck` with 8 placeholder cards (3 rows each)
  - `Box("Output")` hint(-1) wrapping `Terminal` pre-seeded with fake output + separator
- Footer: horizontal `Flex`
  - `Static("footer-status")` — placeholder `● src/**/*.go`
  - `Spacer` (flexible fill)
  - `Custom("hints")` — colour-coded key/label pairs via `hint:key` / `hint:label`
- `hint:key` and `hint:label` added to all six themes (Tokyo, Midnight Neon, Nord,
  Gruvbox Dark, Gruvbox Light, Lipstick)

Acceptance: `go run ./cmd/lazymake` renders the full layout, cards look right,
footer hints show correct colours, terminal shows fake output.

### Step 3 — Target parsing

- Implement `parseMakefile(path string) ([]Target, error)` — regex-based
- Implement `parseJustfile(dir string) ([]Target, error)` — `just --list --format json` with regex fallback
- Implement `loadTargets(dir string) ([]Target, error)` — auto-detects and merges
- Wire into UI: replace placeholder cards with live parsed targets
- Handle empty state (no Makefile/Justfile found)

Acceptance: targets from a real project appear in the Deck with correct badges
and descriptions.

### Step 4 — Execution

- Implement run queue (`chan runRequest`, background drain goroutine)
- Implement `runTarget`: `exec.Cmd` with `Stdout`/`Stderr` wired to the
  `Terminal` widget via its `io.Writer` interface
- Print separator line + summary lines around each run
- Wire `r` / `Enter` to enqueue the selected target
- Update footer queue counter while queue is non-empty
- Wire `c` to `Terminal.Clear()`

Acceptance: pressing `r` runs the selected target, output streams into the
terminal, separator and exit summary appear, queue counter updates correctly.

### Step 5 — Watch mode

- Implement `Watcher` wrapper around `fsnotify.Watcher` with doublestar glob
  resolution
- Implement `.lazymake.json` load/save
- Wire `w` key: prompt for glob if none saved, start/stop watcher
- Watcher events enqueue runs (with deduplication)
- Scanner shows active glob and queue depth
- `●` indicator pulses on trigger

Acceptance: saving a watched file triggers an automatic re-run, footer updates,
duplicate triggers are coalesced.

### Step 6 — Directory switcher

- Wire `d` key to `UI.Prompt` for directory input
- On accept: stop watchers, flush queue, re-parse targets, update header
- Preserve terminal output history across directory switches

Acceptance: switching to a different project directory updates the target list
without clearing the terminal.

### Step 7 — Polish

- Quit confirmation if a process is running (`UI.Confirm`)
- Keyboard focus management (Deck vs Terminal)
- Resize handling (Terminal.Resize on window resize)
- Error states: unparseable Makefile, `just` not found, invalid glob
- README for `cmd/lazymake`

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
