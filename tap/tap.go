// Package tap parses TAP (Test Anything Protocol) output from runner subprocesses.
package tap

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Result is a single TAP test line.
type Result struct {
	Number int
	OK     bool
	Gate   string
	Desc   string
	Error  string
}

// Parse reads TAP lines from r and returns all results.
func Parse(r io.Reader) ([]Result, error) {
	scanner := bufio.NewScanner(r)
	var results []Result
	var last *Result

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "ok "):
			r := parseLine(true, strings.TrimPrefix(line, "ok "))
			results = append(results, r)
			last = &results[len(results)-1]
		case strings.HasPrefix(line, "not ok "):
			r := parseLine(false, strings.TrimPrefix(line, "not ok "))
			results = append(results, r)
			last = &results[len(results)-1]
		case strings.HasPrefix(line, "  # ") && last != nil && !last.OK:
			last.Error = strings.TrimPrefix(line, "  # ")
		}
	}
	return results, scanner.Err()
}

func parseLine(ok bool, rest string) Result {
	r := Result{OK: ok}

	// rest = "1 - GATE-ID: description"
	parts := strings.SplitN(rest, " - ", 2)
	if len(parts) == 2 {
		n, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err == nil {
			r.Number = n
		}
		desc := parts[1]
		if idx := strings.Index(desc, ": "); idx >= 0 {
			r.Gate = desc[:idx]
			r.Desc = desc[idx+2:]
		} else {
			r.Desc = desc
		}
	}
	return r
}

// Writer writes TAP output to w.
type Writer struct {
	w     io.Writer
	total int
	n     int
}

// NewWriter creates a TAP writer. total is the expected test count.
func NewWriter(w io.Writer, total int) *Writer {
	fmt.Fprintf(w, "TAP version 14\n1..%d\n", total)
	return &Writer{w: w, total: total}
}

// Pass emits an ok line.
func (tw *Writer) Pass(gate, desc string) {
	tw.n++
	fmt.Fprintf(tw.w, "ok %d - %s: %s\n", tw.n, gate, desc)
}

// Fail emits a not ok line with an optional error message.
func (tw *Writer) Fail(gate, desc, errMsg string) {
	tw.n++
	fmt.Fprintf(tw.w, "not ok %d - %s: %s\n", tw.n, gate, desc)
	if errMsg != "" {
		fmt.Fprintf(tw.w, "  # %s\n", errMsg)
	}
}

// Skip emits an ok line with a SKIP directive.
func (tw *Writer) Skip(gate, desc, reason string) {
	tw.n++
	fmt.Fprintf(tw.w, "ok %d - %s: %s # SKIP %s\n", tw.n, gate, desc, reason)
}
