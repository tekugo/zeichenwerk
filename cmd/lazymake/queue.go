package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/tekugo/zeichenwerk"
)

// crlfWriter wraps an io.Writer and translates bare LF (\n) to CRLF (\r\n).
// It is stateful so that a \r arriving at the end of one Write chunk and a \n
// at the start of the next are not double-expanded.
type crlfWriter struct {
	w    io.Writer
	prev byte
}

func (c *crlfWriter) Write(p []byte) (int, error) {
	buf := make([]byte, 0, len(p)+8)
	for _, b := range p {
		if b == '\n' && c.prev != '\r' {
			buf = append(buf, '\r')
		}
		buf = append(buf, b)
		c.prev = b
	}
	if _, err := c.w.Write(buf); err != nil {
		return 0, err
	}
	return len(p), nil
}

// runRequest is a single queued execution.
type runRequest struct {
	target Target
}

// Runner owns the run queue and executes targets sequentially.
type Runner struct {
	queue   chan runRequest
	term    *Terminal
	status  *Static
	scanner *Scanner
	dir     string
	busy    atomic.Bool
	written atomic.Bool // true once anything has been written to term

	// deduplication: tracks target names currently sitting in the queue
	dedupMu sync.Mutex
	inQueue map[string]bool
}

// NewRunner creates a Runner and starts the background drain goroutine.
func NewRunner(term *Terminal, status *Static, scanner *Scanner, dir string) *Runner {
	r := &Runner{
		queue:   make(chan runRequest, 8),
		term:    term,
		status:  status,
		scanner: scanner,
		dir:     dir,
		inQueue: make(map[string]bool),
	}
	go r.drain()
	return r
}

// Enqueue adds t to the run queue. Silently drops the request if the queue
// is full, t is a placeholder (empty Runner field), or the same target is
// already waiting in the queue (deduplication for watch mode).
func (r *Runner) Enqueue(t Target) {
	if t.Runner == "" {
		return
	}
	r.dedupMu.Lock()
	already := r.inQueue[t.Name]
	if !already {
		r.inQueue[t.Name] = true
	}
	r.dedupMu.Unlock()

	if already {
		return
	}

	select {
	case r.queue <- runRequest{target: t}:
	default:
		// Queue full — remove the dedup entry so a future attempt can retry.
		r.dedupMu.Lock()
		delete(r.inQueue, t.Name)
		r.dedupMu.Unlock()
	}
	r.updateStatus()
}

// ClearTerminal clears the terminal output and resets the separator state.
func (r *Runner) ClearTerminal() {
	r.term.Clear()
	r.written.Store(false)
}

// SetWatchActive controls the scanner animation in the footer.
func (r *Runner) SetWatchActive(active bool) {
	if r.scanner == nil {
		return
	}
	if active {
		r.scanner.Start(120 * time.Millisecond)
	} else {
		r.scanner.Stop()
	}
}

// drain runs in a goroutine and processes requests one at a time.
func (r *Runner) drain() {
	for req := range r.queue {
		r.dedupMu.Lock()
		delete(r.inQueue, req.target.Name)
		r.dedupMu.Unlock()

		r.busy.Store(true)
		r.updateStatus()
		r.run(req)
		r.busy.Store(false)
		r.updateStatus()
	}
}

// run executes a single request, streaming output to the terminal.
func (r *Runner) run(req runRequest) {
	t := req.target

	// Separator between runs.
	if r.written.Load() {
		fmt.Fprintf(r.term, "\033[2m%s\033[0m\r\n", r.separator(t))
	}
	r.written.Store(true)

	// Command prompt line.
	fmt.Fprintf(r.term, "\033[1;32m$\033[0m %s %s\r\n", t.Runner, t.Name)

	// Build and run the subprocess.
	start := time.Now()
	var cmd *exec.Cmd
	switch t.Runner {
	case "make":
		cmd = exec.Command("make", t.Name)
	case "just":
		cmd = exec.Command("just", t.Name)
	default:
		return
	}
	cmd.Dir = r.dir
	out := &crlfWriter{w: r.term}
	cmd.Stdout = out
	cmd.Stderr = out

	err := cmd.Run()
	elapsed := time.Since(start)

	// Exit summary line.
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(r.term, "\033[31m✗ exit %d  (%.1fs)\033[0m\r\n", exitErr.ExitCode(), elapsed.Seconds())
		} else {
			fmt.Fprintf(r.term, "\033[31m✗ %s\033[0m\r\n", err)
		}
	} else {
		fmt.Fprintf(r.term, "\033[32m✓ exit 0  (%.1fs)\033[0m\r\n", elapsed.Seconds())
	}
}

// separator builds the ── runner target ───...─ line sized to the terminal.
func (r *Runner) separator(t Target) string {
	_, _, w, _ := r.term.Bounds()
	if w <= 0 {
		w = 80
	}
	label := fmt.Sprintf("── %s %s ", t.Runner, t.Name)
	dashes := w - len(label)
	dashes = max(dashes, 2)
	return label + strings.Repeat("─", dashes)
}

// updateStatus refreshes the footer status widget.
func (r *Runner) updateStatus() {
	n := len(r.queue)
	busy := r.busy.Load()
	switch {
	case busy && n > 0:
		r.status.SetText(fmt.Sprintf("▶ running  [queue: %d]", n))
	case busy:
		r.status.SetText("▶ running")
	case n > 0:
		r.status.SetText(fmt.Sprintf("[queue: %d]", n))
	default:
		r.status.SetText("")
	}
}
