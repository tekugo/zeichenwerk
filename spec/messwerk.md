# Messwerk

A persistent terminal dashboard for Claude Code that combines a
**Config Inspector** and an **Observability Dashboard**. It runs as two
cooperating processes: a background daemon that receives OTLP telemetry and
writes it to SQLite, and a TUI viewer that reads from that database and from
Claude Code's configuration files.

---

## Architecture

### Two modes

```
messwerk serve    # Daemon: OTLP receiver + SQLite writer
messwerk          # TUI: viewer (reads SQLite + config files)
```

> **Note:** the binary is named `messwerk`, not `claudmin`. The `serve`
> sub-command and SQLite persistence are not yet implemented (see
> Implementation status below).

The daemon runs continuously in the background. The TUI is a pure read-only
viewer — closing it does not interrupt collection. Starting or stopping the TUI
has no effect on the data stream.

### Data flow

```
Claude Code
  │  OTLP (gRPC, localhost:4317)
  ▼
messwerk serve
  │  SQLite  (~/.messwerk/telemetry.db)
  ▼
messwerk (TUI)
  ├  reads telemetry from SQLite
  └  reads config files directly from the filesystem
```

All processing is local. No data leaves the developer's machine.

---

## Daemon (`claudmin serve`)

Starts a gRPC OTLP receiver on `localhost:4317`. For each incoming batch:

- Metrics: write one row per data point to `metrics` table with timestamp,
  metric name, value, and all attributes.
- Logs (events): write one row per log record to `events` table with
  timestamp, event name, body, and attributes.

Sessions are identified by the `session.id` attribute present on every record.
A session is considered **active** while data with its ID arrives within a
configurable timeout window (default: 2 minutes). After the timeout elapses
with no new data, the session is marked **ended** in the `sessions` table.

The daemon accepts these flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `4317` | gRPC listen port |
| `--db` | `~/.messwerk/telemetry.db` | SQLite database path |
| `--session-timeout` | `2m` | Idle duration before a session is marked ended |

### OTLP activation (Claude Code settings)

```json
{
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "OTEL_METRICS_EXPORTER": "otlp",
    "OTEL_LOGS_EXPORTER": "otlp",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317",
    "OTEL_METRIC_EXPORT_INTERVAL": "10000",
    "OTEL_LOGS_EXPORT_INTERVAL": "5000"
  }
}
```

### Available telemetry

**Metrics** (exported every 60 s):

| Metric | Attributes |
|--------|-----------|
| `claude_code.token.usage` | `type` (input/output/cache_read/cache_write), model, `session.id` |
| `claude_code.cost.usage` | USD, model, `session.id` |
| `claude_code.session.count` | — |
| `claude_code.active_time.duration` | active usage duration |
| `claude_code.lines_of_code.count` | `type` (added/removed) |
| `claude_code.commits.count` | — |
| `claude_code.pull_requests.count` | — |
| `claude_code.code_edit_tool.decision` | `decision` (accept/reject/ask) |

**Events / Logs** (exported every 5 s):

| Event | Fields |
|-------|--------|
| `user_prompt` | length, `prompt.id` |
| `api_request` | model, tokens, cost, latency, cache-hit flag |
| `api_error` | error type |
| `tool_result` | tool name, success/error, duration, decision |
| `tool_decision` | permission (allow/deny/ask) |

---

## Layout

### Top-level structure

```
┌─────────────────────┬──────────────────────────────────────┐
│  Deck (navigation)  │  Detail view                         │
│                     │                                      │
│  ● Gesamt           │  (depends on selected deck item)     │
│  12.4k tok  $0.08   │                                      │
│  ▁▂▃▄▅▆▃▂▁▂▃▄▅▇█▆  │                                      │
│                     │                                      │
│  ○ ~/projekt-a      │                                      │
│   8.1k tok  $0.05   │                                      │
│  ▁▁▂▂▃▃▄▄▅▅▆▆▇▇██  │                                      │
│                     │                                      │
└─────────────────────┴──────────────────────────────────────┘
        [Observability]  [Config]  [About]
```

The window is split into:
- **Left column** — a `Deck` widget used as navigation.
- **Right column** — context-sensitive detail view.
- **Bottom bar** — tab strip (`Observability`, `Config`, `About`).

The left/right split uses a `Splitter` with a fixed left width (default: 22
characters, user-adjustable).

### Navigation deck (left column)

Each item in the navigation deck renders three rows:

```
Row 0: status-dot + session name / path
Row 1: token count + cost (right-aligned)
Row 2: Sparkline (last 20 minutes, Relative scale)
```

**Status dot**: `●` (green) = active, `●` (yellow) = idle, `○` (dim) = ended.

The first item is always **"Gesamt"** (aggregate across all sessions). Below it
follow individual sessions in reverse-start-time order.

Selecting an item updates the detail view immediately.

### Tab strip

`Observability` is the default tab. Switching tabs changes the detail view
for the currently selected deck item; the deck selection is preserved across
tab switches.

---

## Observability tab

### Aggregate view ("Gesamt" selected)

Answers: *"What is happening right now, and how has today gone?"*

```
┌──────────────┐ ┌──────────────┐ ┌─────────────────────┐
│  Tokens today│ │  Cost today  │ │  Cache-hit rate      │
│   █ 48.2k   │ │   █ $0.31   │ │    [Gauge  72%]      │
│  ▁▂▃▄▅▆▃▂▁▂ │ │  ▁▂▃▂▁▂▃▄▅▇ │ └─────────────────────┘
└──────────────┘ └──────────────┘
┌──────────────────────────────────────────────────────┐
│  Active sessions                                      │
│  ● ~/projekt-a   claude-sonnet-4-6   active  12s ago │
│  ● ~/projekt-b   claude-opus-4-6     active  45s ago │
│  ○ ~/projekt-c   claude-haiku-4-5    idle    8m ago  │
└──────────────────────────────────────────────────────┘
```

Top row: three panels side by side using `Tiles` (equal width).
- **Tokens today**: big number + `Sparkline` (today, Absolute scale).
- **Cost today**: big number + `Sparkline` (today, Absolute scale).
- **Cache-hit rate**: `Gauge` showing `cache_read / (input + cache_read)`.

Below: session list showing each known session with its name, current model,
status label, and time-since-last-data.

### Session detail view (individual session selected)

Answers: *"What is this session doing?"*

Layout (top to bottom):

1. **Big numbers row** — input tokens, output tokens, cache tokens, cost.
   Each is a `digits.go` big-number display with an animated roll-up on update.
2. **Model badge** — current model name, right-aligned.
3. **Sparkline** — selectable time window (last hour / today / 7 days),
   Absolute scale, switchable with `←`/`→`.
4. **Live event stream** — scrollable list of recent `tool_result` events:
   tool name, duration, success/error indicator.
5. **Permission decisions** — counts of allow/deny/ask for `tool_decision`
   events in this session.

### Time window and scale rules

| Context | Window | Scale |
|---------|--------|-------|
| Deck sparkline | Last 20 minutes | Relative |
| Detail sparkline — short | Last 1 hour | Absolute |
| Detail sparkline — medium | Today | Absolute |
| Detail sparkline — long | Last 7 days | Absolute |

---

## Config tab

### Structure

The Config tab reuses the same left-deck / right-detail layout as
Observability. The deck lists configuration categories:

- Settings
- Permissions
- Hooks
- MCP Servers
- Agents
- Skills
- Slash Commands
- CLAUDE.md

### Config source hierarchy (highest priority first)

1. **Managed** — enterprise policies (MDM / Registry / `managed-settings.json`)
2. **Local** — `.claude/settings.local.json` (not committed)
3. **Project** — `.claude/settings.json` (committed)
4. **User** — `~/.claude/settings.json`
5. **Plugin defaults**

Additional files always read:
- `~/.claude.json` — OAuth tokens, theme, MCP user/local scope
- `.mcp.json` — project MCP servers
- `CLAUDE.md`, `.claude/CLAUDE.md`, `~/.claude/CLAUDE.md`, `CLAUDE.local.md`

Agent/skill/command directories:
- User: `~/.claude/agents/`, `~/.claude/skills/`, `~/.claude/commands/`
- Project: `.claude/agents/`, `.claude/skills/`, `.claude/commands/`

### Settings category

**Master view** — horizontally scrollable table:

```
Key                  │ Effective    │ User       │ Project    │ Local   │ Managed
─────────────────────┼──────────────┼────────────┼────────────┼─────────┼────────
model                │ sonnet-4-6   │ opus-4-6   │ —          │ sonnet… │ —
cleanupPeriodDays    │ 20           │ 30         │ 20         │ —       │ —
permissions.allow    │ [4 rules]    │ [2]        │ [2]        │ —       │ —
```

- Long values are truncated with `…`.
- Array settings are summarised as `[N rules]`.
- Values that are overridden by a higher-priority scope are rendered dim.
- Managed-scope values are rendered with a distinct accent colour.

**Detail view** — scalar setting:

```
model
──────────────────────────────────────────────────────
  User  (~/.claude/settings.json)
    opus-4-6  ← overridden

  Local  (.claude/settings.local.json)
    sonnet-4-6  ✓ effective
```

**Detail view** — array setting (merged, not overridden):

```
permissions.allow
──────────────────────────────────────────────────────
  Merged from 2 scopes:

  User  (~/.claude/settings.json)
    ├─ Bash(npm run *)
    └─ Bash(git log *)

  Project  (.claude/settings.json)
    ├─ Bash(npm run lint)
    └─ Read(./.env)  ✗  (denied by higher scope)
```

### Permissions category

Three sections: **Allow**, **Ask**, **Deny**. Each rule shows its scope label.
Conflicts (same pattern allowed in one scope and denied in another) are
highlighted in a warning colour.

### Hooks category

Table: `Event | Matcher | Type | Scope | Command / URL`

Detail: full configuration of the selected hook rendered as formatted JSON or
YAML.

### MCP Servers category

Table: `Name | Scope | Transport | Command / URL | Status`

Connection status is checked once when the category is opened. A manual
refresh is available.

### Agents, Skills, Slash Commands categories

Table per category: `Name | Scope | Path | Description`

Detail: Markdown file content with syntax highlighting (uses the existing ANSI
parser / renderer infrastructure).

### Config ↔ Observability cross-link

`tool_decision` events and `code_edit_tool.decision` metrics together reveal
how often Claude requests permission and how often it is denied. The Config
tab surfaces a contextual hint when the deny rate for a tool exceeds a
threshold, e.g.:

> *"Bash commands are frequently denied — consider tightening the permission
> rule in `.claude/settings.json`."*

---

## About tab

Displays:
- Claudmin version and build info.
- Daemon status: running / not running, uptime, database path and size.
- OTLP receiver endpoint.
- SQLite retention policy (configurable, default: 30 days).

---

## New zeichenwerk widgets required

Claudmin requires the following new zeichenwerk widgets. Each has a separate
spec:

| Widget | Spec | Primary use |
|--------|------|-------------|
| `Sparkline` | `spec/sparkline.md` | Token-rate history in deck items and detail panels |
| `Gauge` | `spec/gauge.md` | Cache-hit rate display in the aggregate view |
| `Heatmap` | `spec/heatmap.md` | Hour-of-day × weekday activity grid |

Additionally, `digits.go` is extended with:
- Animated roll-up counter on value update.
- Optional unit label rendered as small text below the number.

---

## Decisions

- No collapsible session preview panel.
- Idle timeout is user-configurable via `--timeout` flag (default: `2m`).
  Sessions transition: active → idle at `timeout/2`, idle → ended at `timeout`.
- First milestone: telemetry deck only — no config inspector, no SQLite.
  OTLP metrics are received and held in-memory; data is lost when messwerk exits.
- SQLite retention TBD for a later milestone.

---

## Implementation status

### Done

- **OTLP gRPC receiver** (`cmd/messwerk/receiver.go`) — handles
  `claude_code.token.usage` (all four token types) and `claude_code.cost.usage`.
  Session identity is resolved from `session.id`; display name is derived from
  `claude_code.session.path` / `process.working_directory` resource attributes,
  shortened to the last two path components.
- **In-memory store** (`cmd/messwerk/model.go`) — per-session rolling 20-bucket
  (1-bucket = 1 minute) history; Gesamt aggregate bucket; `onChange` callback for
  live refresh; `SessionStatus` (active / idle / ended) derived from
  `idleTimeout`.
- **TUI — Milestone 1** (`cmd/messwerk/ui.go`, `cmd/messwerk/main.go`):
  - Left deck: status dot + name (row 0), token count + cost right-aligned
    (row 1), per-session Sparkline in Absolute mode (row 2).
  - Right detail: token stats bar (Input / Output / Cache / Cost), horizontal
    rule, detail Sparkline, horizontal rule, session list (for Gesamt) or
    placeholder (for individual sessions).
  - 15-second background ticker keeps status dots fresh between data arrivals.
  - Theme flag (`-t`) supports midnight, tokyo (default), nord, gruvbox-dark,
    gruvbox-light, lipstick.
- **Sparkline widget** (`sparkline.go`) — implemented and used in deck items and
  detail panel.

### Not yet started

- Config tab (Settings, Permissions, Hooks, MCP Servers, Agents, Skills, Slash
  Commands, CLAUDE.md categories).
- Observability tab — aggregate "Gesamt" panel with Tiles layout (Tokens today /
  Cost today / Cache-hit Gauge + active session list).
- Session detail view — big-number row with animated roll-up, model badge, time-
  window switcher (`←`/`→`), live event stream, permission-decision counters.
- Tab strip (`Observability` / `Config` / `About`).
- About tab (daemon status, OTLP endpoint, DB path/size, retention policy).
- SQLite persistence (daemon `serve` sub-command, `sessions`/`metrics`/`events`
  tables, configurable retention).
- Log/event ingestion — `user_prompt`, `api_request`, `api_error`, `tool_result`,
  `tool_decision` event types.
- Gauge widget (`spec/gauge.md` — needed for cache-hit rate panel).
- Heatmap widget (`spec/heatmap.md` — `heatmap.go` + `heatmap_test.go` present
  but not yet integrated into messwerk).
- `digits.go` animated roll-up counter and unit-label extension.
- Config ↔ Observability cross-link (deny-rate hint).
