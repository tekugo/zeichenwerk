package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ── Regexes ──────────────────────────────────────────────────────────────────

// makeTargetRe matches a Makefile target line, e.g. "build:" or "build: deps".
// The [^:=] lookahead prevents matching variable assignments (CC := ...).
var makeTargetRe = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9_\-\./]*)\s*:[^:=]`)

// makeTargetOnlyRe matches a target-only line with no prerequisites, e.g. "build:".
var makeTargetOnlyRe = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9_\-\./]*)\s*:\s*$`)

// makeDescRe matches a Make description comment directly above a target.
var makeDescRe = regexp.MustCompile(`^##\s*(.+)`)

// justRecipeRe matches the start of a just recipe line.
var justRecipeRe = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9_\-]*)(\s.*)?:`)

// justDocRe matches a just doc comment directly above a recipe.
var justDocRe = regexp.MustCompile(`^#\s*(.+)`)

// ── Makefile parser ───────────────────────────────────────────────────────────

// parseMakefile reads a Makefile and extracts targets with optional descriptions.
// A description is taken from a `## comment` line immediately above the target.
// Targets whose names start with `.` (special Make directives) are skipped;
// .PHONY targets are intentionally kept — they are the most common runnable targets.
func parseMakefile(path string) ([]Target, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var targets []Target
	var prevDesc string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip special Make directives (lines starting with `.`).
		if strings.HasPrefix(line, ".") {
			prevDesc = ""
			continue
		}

		// Capture description comment.
		if m := makeDescRe.FindStringSubmatch(line); m != nil {
			prevDesc = strings.TrimSpace(m[1])
			continue
		}

		// Match target (with or without prerequisites).
		name := ""
		if m := makeTargetRe.FindStringSubmatch(line); m != nil {
			name = m[1]
		} else if m := makeTargetOnlyRe.FindStringSubmatch(line); m != nil {
			name = m[1]
		}

		if name != "" {
			targets = append(targets, Target{
				Name:        name,
				Description: prevDesc,
				Runner:      "make",
			})
		}

		// Reset description on any non-comment line.
		prevDesc = ""
	}

	return targets, scanner.Err()
}

// ── Justfile parser ───────────────────────────────────────────────────────────

// justListJSON is the structure returned by `just --list --list-format json`.
type justListJSON struct {
	Recipes []struct {
		Name string `json:"name"`
		Doc  string `json:"doc"`
	} `json:"recipes"`
}

// parseJustfile extracts recipes from a Justfile in dir.
// It first tries `just --list --list-format json`; if just is not on PATH or
// the flag is unsupported, it falls back to regex parsing of the Justfile.
func parseJustfile(dir string) ([]Target, error) {
	if targets, err := parseJustfileJSON(dir); err == nil {
		return targets, nil
	}
	return parseJustfileRegex(filepath.Join(dir, "justfile"))
}

// parseJustfileJSON shells out to just for structured recipe data.
func parseJustfileJSON(dir string) ([]Target, error) {
	cmd := exec.Command("just", "--list", "--list-format", "json", "--unsorted")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var data justListJSON
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, err
	}

	targets := make([]Target, 0, len(data.Recipes))
	for _, r := range data.Recipes {
		if strings.HasPrefix(r.Name, "_") {
			continue // private recipes
		}
		targets = append(targets, Target{
			Name:        r.Name,
			Description: strings.TrimSpace(r.Doc),
			Runner:      "just",
		})
	}
	return targets, nil
}

// parseJustfileRegex parses a Justfile with a simple line-by-line regex scan.
// Doc comments (`# text`) on the line immediately above a recipe are used as
// descriptions. Private recipes (name starting with `_`) are skipped.
func parseJustfileRegex(path string) ([]Target, error) {
	f, err := os.Open(path)
	if err != nil {
		// Try capitalised variant
		f, err = os.Open(filepath.Join(filepath.Dir(path), "Justfile"))
		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	var targets []Target
	var prevDoc string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		if m := justDocRe.FindStringSubmatch(line); m != nil {
			prevDoc = strings.TrimSpace(m[1])
			continue
		}

		if m := justRecipeRe.FindStringSubmatch(line); m != nil {
			name := m[1]
			if !strings.HasPrefix(name, "_") {
				targets = append(targets, Target{
					Name:        name,
					Description: prevDoc,
					Runner:      "just",
				})
			}
		}

		prevDoc = ""
	}

	return targets, scanner.Err()
}

// ── Auto-detect and merge ─────────────────────────────────────────────────────

// makefileNames lists the filenames recognised as Makefiles, in priority order.
var makefileNames = []string{"GNUmakefile", "makefile", "Makefile"}

// justfileNames lists the filenames recognised as Justfiles, in priority order.
var justfileNames = []string{"justfile", "Justfile"}

// loadTargets detects and parses Makefile and/or Justfile targets from dir.
// Makefile targets are listed first, Justfile targets second.
// Returns an empty slice (no error) if neither file is found.
func loadTargets(dir string) ([]Target, error) {
	var all []Target
	var errs []string

	// Makefile
	for _, name := range makefileNames {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			targets, err := parseMakefile(path)
			if err != nil {
				errs = append(errs, "Makefile: "+err.Error())
			} else {
				all = append(all, targets...)
			}
			break
		}
	}

	// Justfile
	for _, name := range justfileNames {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			targets, err := parseJustfile(dir)
			if err != nil {
				errs = append(errs, name+": "+err.Error())
			} else {
				all = append(all, targets...)
			}
			break
		}
	}

	if len(errs) > 0 && len(all) == 0 {
		return nil, fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return all, nil
}
