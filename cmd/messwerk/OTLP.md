# Configuring Claude Code for OTLP

Messwerk receives Claude Code telemetry over **OTLP/gRPC** and displays token usage, cost, and activity in real time.

## Quick start

Add the following to your Claude Code `settings.json` (user, project, or local scope):

```json
{
  "env": {
    "CLAUDE_CODE_ENABLE_TELEMETRY": "1",
    "OTEL_METRICS_EXPORTER": "otlp",
    "OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
    "OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4317"
  }
}
```

Then start messwerk before launching Claude Code:

```sh
messwerk
```

## Variables

| Variable | Value | Purpose |
|---|---|---|
| `CLAUDE_CODE_ENABLE_TELEMETRY` | `1` | Enables telemetry collection |
| `OTEL_METRICS_EXPORTER` | `otlp` | Routes metrics to the OTLP exporter |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc` | Uses gRPC transport |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4317` | Messwerk's listen address |

## Options

**Custom port** — if 4317 is already in use, start messwerk on a different port and update the endpoint accordingly:

```sh
messwerk -port 4318
```

```json
"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318"
```

**Faster updates** — by default Claude Code exports metrics every 60 seconds. For a more responsive display during development, reduce the interval:

```json
"OTEL_METRIC_EXPORT_INTERVAL": "10000"
```

**Theme** — messwerk ships with several built-in themes:

```sh
messwerk -t tokyo       # default
messwerk -t midnight
messwerk -t nord
messwerk -t gruvbox-dark
messwerk -t gruvbox-light
messwerk -t lipstick
```
