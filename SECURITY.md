# Security Policy

## Supported Versions

Zeichenwerk is in active pre-release development. Security fixes are applied
to the latest commit on `main` only. No backport releases are made for older
versions.

| Version | Supported |
| ------- | --------- |
| latest (`main`) | ✓ |
| older tags | ✗ |

## Scope

Zeichenwerk is a terminal UI library. Its attack surface is limited to:

- **Terminal input handling** — key and mouse events from tcell
- **ANSI/VT sequence parsing** — inside the `Terminal` widget
- **File-system access** — `TreeFS` reads directory listings; no writes

The library does **not** open network connections, spawn processes, or handle
user-supplied data on behalf of applications beyond what the embedding
application explicitly passes in.

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Report vulnerabilities by email to:

**thomas.rustemeyer@me.com**

Include:

1. A brief description of the issue and its potential impact
2. Steps to reproduce or a minimal proof-of-concept
3. The version or commit hash you tested against

You will receive an acknowledgement within 72 hours. If the issue is confirmed,
a fix will be prepared and released as soon as practical, and you will be
credited in the release notes unless you prefer to remain anonymous.

## Out of Scope

- Issues in dependencies (tcell, etc.) — report those upstream
- Denial-of-service via extremely large terminal output — not considered a
  vulnerability in a local terminal library
- Behaviour that requires the attacker to already control the process
