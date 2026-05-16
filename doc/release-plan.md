# Publish zeichenwerk: extract apps + launch checklist

## Status — 2026-05-16

**Part 1 extractions: largely done.** Eight apps are now standalone
repos under `tekugo/` at `v0.1.0` (or higher):

| App           | Repo                                                       | Tag    |
|---------------|------------------------------------------------------------|--------|
| dbu           | [tekugo/datenwerk](https://github.com/tekugo/datenwerk)            | v0.1.0 |
| triebwerk     | [tekugo/triebwerk](https://github.com/tekugo/triebwerk) (binary `tw`) | v0.1.0 |
| messwerk      | [tekugo/messwerk](https://github.com/tekugo/messwerk) (binary `mw`)   | v0.1.0 |
| tblr          | [tekugo/tblr](https://github.com/tekugo/tblr)                      | v0.1.0 |
| figlet        | [tekugo/figlet](https://github.com/tekugo/figlet)                  | v0.1.0 |
| malwerk       | [tekugo/malwerk](https://github.com/tekugo/malwerk)                | v0.1.0 |
| hal-explorer  | [tekugo/hal-explorer](https://github.com/tekugo/hal-explorer)      | v0.1.0 |
| (tutorial)    | [tekugo/zeichenwerk-tutorial](https://github.com/tekugo/zeichenwerk-tutorial) | v1.1.0 |

Per-app effect on zeichenwerk's `go.mod`: `fsnotify` + `doublestar` (triebwerk), `otlp` + `grpc` + `grpc-gateway` + `genproto` + `protobuf` (messwerk), `atotto/clipboard` (tblr) are all gone. The library now requires only `tcell`, `uniseg`, `testify`, and `golang.org/x/tools`.

**Not extracted (decision):** `cmd/editor` stays in zeichenwerk as a showcase that will be expanded. `cmd/designer` (the binary driving the root-level `designer/` package) was removed manually; the package itself is library code and stays.

**Remaining Part 1 work:** none — extractions are complete.

Parts 2–4 (CI/CONTRIBUTING, homepage, launch content) are still TODO.

## Context

The library has reached v2 beta and is largely publish-ready: MIT LICENSE, strong README (325 lines, 24 code blocks, comparison table, gif), CHANGELOG (440 lines), substantial `doc/` tree (tutorial / reference / spec / designer subdirs), and a `doc.go` that gives `pkg.go.dev` a real landing page. Release tags exist up to `v2.0.0-beta.6`.

What's missing for a clean public launch is **focus**: the `cmd/` directory currently holds 17 commands, including 9 substantial standalone applications (malwerk 2787 LOC, messwerk 1255, tblr 1336, triebwerk 933, hal-explorer 427, figlet 547, designer 82, editor 80, dbu 156) that dilute the library's story. Once those move out, what remains in `cmd/` is unambiguously "demos + small dev tools that exercise the library."

Beyond extraction, the gaps are: no CI, no `CONTRIBUTING.md` / `CODE_OF_CONDUCT.md`, only one gif (asciinema embeds would carry better), and no marketing presence (homepage, launch posts).

Decisions baked in:
- Extract the 9 substantial apps (see list below); keep demos + dev tools (demo, demo2, compose, showcase, designer-poc, inspector-poc, zw, analyzer, coverage)
- One GitHub repo per extracted app under `tekugo/`
- Build a **dedicated-domain homepage**, not just GitHub Pages
- Launch on **Hacker News**, **awesome-go** / **awesome-tui** PRs, and reddit (`/r/golang` + `/r/commandline` — the original answer wrote `/r/tu`, treated as typo; confirm before posting)

## Part 1 — Extract the 9 apps

For each of `malwerk`, `messwerk`, `tblr`, `triebwerk`, `designer`, `hal-explorer`, `figlet`, `editor`, `dbu`:

1. **Preserve history.** Use `git filter-repo --path cmd/<name>` against a clone of zeichenwerk to get a history-only-of-that-subdir repo. Cleaner than `cp -r`.
2. **Create the new repo** `tekugo/<name>` on GitHub (empty, MIT, no auto README).
3. **Push the filtered repo** to it.
4. **Promote to a module**: `git mv cmd/<name>/* .` so files sit at the repo root; `go mod init github.com/tekugo/<name>`; `go get github.com/tekugo/zeichenwerk@v2.0.0` (or whatever tag is current).
5. **Add per-app files**: `LICENSE` (copy MIT), `README.md` (one-paragraph what-it-is + screenshot/asciicast + install + usage), `go.sum`.
6. **Tag `v0.1.0`** and push. Don't promise stability yet — these are early.
7. **Delete `cmd/<name>/` from zeichenwerk** in the same commit that updates CHANGELOG (entry: "extracted to github.com/tekugo/<name>").

A short shell script in zeichenwerk's `doc/release/` directory automating steps 1–6 would let the user do these in batches without re-deriving the recipe each time. **Recommended**: write `doc/release/extract-app.sh <name>` as part of this plan.

Sequencing: extract one app end-to-end as the pilot (suggest `dbu` — smallest, 156 LOC, lowest risk). Validate the script and the GitHub-side workflow, then extract the rest in a single batch session. The library can release before all 9 extractions are done — apps are independent.

## Part 2 — Library publish-readiness

Add to repo root, none over 100 lines:

- **`CONTRIBUTING.md`** — branch model, how to run tests/lint, commit style, where to file issues. Reference the existing `doc/principles.md` for design philosophy.
- **`CODE_OF_CONDUCT.md`** — adopt Contributor Covenant v2.1 verbatim. Standard, expected.
- **`.github/workflows/ci.yml`** — matrix on Go 1.25 + 1.26 + tip, steps: `go build ./...`, `go test ./...`, `go vet ./...`, `golangci-lint run`. Run on push to main + on PRs.
- **`.github/ISSUE_TEMPLATE/`** — bug-report + feature-request templates.
- **`.github/PULL_REQUEST_TEMPLATE.md`** — checklist (tests, doc, CHANGELOG).
- **README badges**: build status, pkg.go.dev godoc, latest release, license, Go report card. One line each at the top under the title.

Existing **doc/** is good. Don't reorganise — defer to the homepage build (Part 3) to add a navigation layer.

**Release `v2.0.0` stable** when CI is green and the comparison table in README is updated. Don't launch on a beta tag.

## Part 3 — Homepage at a dedicated domain

Domain (in order of preference): `zeichenwerk.dev` (Go-library convention), `zeichenwerk.io`, `zeichenwerk.app`. Cost ~€15–60/year depending on TLD. Recommend `.dev` — owned by Google, requires HTTPS, signals "developer tool."

Tech stack: **Astro** with the [Starlight](https://starlight.astro.build) docs theme. Why Astro/Starlight:
- Out-of-the-box docs nav, search, dark mode, mobile, syntax-highlighted code
- Static output → host anywhere, no runtime
- Markdown + MDX, so the existing `doc/` tree migrates with minimal changes
- Asciinema embeds work cleanly in MDX components

Alternatives considered: Hugo (heavier theme decisions), Docusaurus (React-only, more JS), plain HTML (no nav/search out of box). Astro+Starlight is the lowest-effort path to a polished result.

Hosting: **Cloudflare Pages** (free, fast global CDN, auto-deploy from a `homepage` branch or sibling repo `tekugo/zeichenwerk-site`). GitHub Pages is the fallback if Cloudflare adds friction.

Content scope (minimum viable site):
- **Landing**: hero with a 30-second asciicast (designer popup or showcase demo), one-paragraph pitch, install command (`go get github.com/tekugo/zeichenwerk`), CTA buttons → "Tutorial" / "Widget Gallery" / "GitHub"
- **Tutorial**: port `doc/tutorial/` (already exists). Step-by-step "build a TUI in 10 minutes."
- **Widget Gallery**: one card per widget, asciicast + Builder snippet + Compose snippet + link to pkg.go.dev for the API. Source from `doc/reference/`.
- **Designer / Inspector**: short page each with asciicast + how to enable + screenshot.
- **Comparison**: zeichenwerk vs bubbletea vs tview vs gocui. Pull from the existing README section and expand.
- **API reference**: link out to `pkg.go.dev/github.com/tekugo/zeichenwerk`, don't duplicate.

Source repo: new `tekugo/zeichenwerk-site` (keep separate so CI/CD doesn't entangle with the library). Build time: 1–2 days for the scaffold + content migration, plus the asciicast recording.

**Asciinema recordings to produce** (record locally with `asciinema rec`, host on asciinema.org or self-host as `.cast` files in the site repo):
1. Builder hello-world (30s) — type the code, run, see the TUI
2. Designer popup walkthrough (45s) — Ctrl+Space, select widget, edit, Apply, see change
3. Inspector popup walkthrough (45s) — Ctrl+D, navigate tree, see properties + log
4. One showcase demo in motion (30s) — visual hook for the landing hero

## Part 4 — Launch content

**Hacker News post.** Title under 70 chars, "Show HN" prefix.
- Draft title: `Show HN: Zeichenwerk – A Go TUI library with a built-in visual designer`
- Body: 1 paragraph (what it is + what's different — the visual designer is the differentiator), link to homepage, link to GitHub, link to the designer asciicast inline
- Time it for a weekday Tuesday/Wednesday morning EST
- Be ready to answer "vs bubbletea / tview" within the first hour
- File: `doc/release/hn-post.md`

**awesome-go PR** (`avelino/awesome-go`).
- Section: under "GUI Applications" or "Text Processing" (verify the right section against the repo's TOC).
- Format: `- [zeichenwerk](https://github.com/tekugo/zeichenwerk) - Go TUI library with a built-in visual designer.`
- Must follow contribution guidelines (alphabetical, no trailing period inconsistencies, etc.)
- File: `doc/release/awesome-go-entry.md`

**awesome-tui PR** (`rothgar/awesome-tuis` is the canonical list).
- Section: "Frameworks & Libraries" or similar.
- Similar one-liner.
- File: `doc/release/awesome-tui-entry.md`

**Reddit posts** — `/r/golang` (primary) and `/r/commandline` (secondary).
- More casual tone than HN; lead with the asciicast.
- Title under 90 chars.
- Wait 24–48 hours after HN so timelines don't collide.
- File: `doc/release/reddit-posts.md`

## Sequencing

1. **Week 1 — library publish prep.** Add CI workflow, CONTRIBUTING, CODE_OF_CONDUCT, issue/PR templates, README badges. Tag `v2.0.0` stable once CI is green. Pilot-extract `dbu` (the smallest) to validate the script.
2. **Week 2 — app extractions.** Run `doc/release/extract-app.sh` for the remaining 8 apps. Each gets a per-app README. Tag `v0.1.0` on each.
3. **Week 3 — homepage.** Astro+Starlight scaffold in `tekugo/zeichenwerk-site`, port `doc/tutorial/` + `doc/reference/`. Record the four asciicasts. Buy `zeichenwerk.dev` and wire it to Cloudflare Pages.
4. **Week 4 — launch.** Soft-launch via awesome-go + awesome-tui PRs first (these are evergreen). Then HN post Tuesday/Wednesday morning EST. Reddit posts 24–48 hours after.

This is "calendar order" — not "blocking order." Steps 1, 2, 3 are mostly independent and can be parallelised.

## Critical files / artifacts

**New in zeichenwerk repo:**
- `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`
- `.github/workflows/ci.yml`, `.github/ISSUE_TEMPLATE/`, `.github/PULL_REQUEST_TEMPLATE.md`
- `doc/release/extract-app.sh` — automation for step 1 of each extraction
- `doc/release/hn-post.md`, `doc/release/awesome-go-entry.md`, `doc/release/awesome-tui-entry.md`, `doc/release/reddit-posts.md` — drafts for review
- README updates: badges row + extraction note in the demo section (`cmd/<app>` → `github.com/tekugo/<app>`)
- CHANGELOG entries: stable `v2.0.0` release + each app extraction

**Deleted from zeichenwerk repo (after extraction):**
- `cmd/malwerk/`, `cmd/messwerk/`, `cmd/tblr/`, `cmd/triebwerk/`, `cmd/designer/`, `cmd/hal-explorer/`, `cmd/figlet/`, `cmd/editor/`, `cmd/dbu/`

**New repos under tekugo/:**
- `tekugo/malwerk`, `tekugo/messwerk`, `tekugo/tblr`, `tekugo/triebwerk`, `tekugo/zw-designer` (`tekugo/designer` may collide with generic names — verify availability), `tekugo/hal-explorer`, `tekugo/figlet`, `tekugo/editor`, `tekugo/dbu`
- `tekugo/zeichenwerk-site` — Astro homepage

## Verification

- **Library**: `go build ./...` and `go test ./...` pass on a fresh clone. CI is green on `v2.0.0` tag. `go install github.com/tekugo/zeichenwerk/cmd/demo@v2.0.0` produces a runnable binary.
- **Per extracted app**: clone the new repo, `go build ./...`, `go install .` — runs against the published zeichenwerk dependency. The app's README install instruction (`go install github.com/tekugo/<name>@latest`) works for a fresh user.
- **Homepage**: production build deploys to Cloudflare Pages, the four asciicasts play, the install command on the landing copies correctly, search returns results for "widget", "designer", "theme". `https://zeichenwerk.dev` resolves with HTTPS.
- **Launch posts**: HN post draft proofread for accuracy and "Show HN" guidelines. awesome-go PR passes their CI (alphabetical sort + link health).

## Critical files to reference during execution

- `README.md` — current state of the landing copy (the homepage will port most of this)
- `doc.go` — package doc / pkg.go.dev landing
- `CHANGELOG.md` — model for the new release entry
- `doc/tutorial/`, `doc/reference/` — content to port to the homepage
- `doc/principles.md` — design philosophy (link from CONTRIBUTING.md)
- `cmd/<name>/main.go` — per-app source to extract (start with `cmd/dbu/main.go` as the pilot)
