# Messwerk

- Receives OTLP telemetry data from Claude Code
- Tracks the currently running sessions
- For every session, keeps track of:
  - ID, start
  - Token usage input/output/cache per model
  - Token cost per model
- Display all running sessions in a Deck
- For every session show item (41 chars wide):
  - Line 1: Status indicator, Session ID
  - Line 2: Model, last token input, last token output (numbers), total cost
  - Line 3: "Input" label, Sparkline 30 chars wide, input tokens every 2 minutes
  - Line 4: "Output" label, Sparkline 30 chars wide, output tokens every 2
    minutes
