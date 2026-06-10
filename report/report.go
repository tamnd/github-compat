// Package report handles conformance report serialization.
package report

import (
	"encoding/json"
	"io"
	"sort"
	"time"

	"github.com/tamnd/github-compat/tap"
)

// Status is the outcome of a single gate.
type Status string

const (
	Pass Status = "pass"
	Fail Status = "fail"
	Skip Status = "skip"
)

// GateResult is the result of a single gate test.
type GateResult struct {
	Status     Status `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// Report is the top-level JSON structure written to --report FILE.
type Report struct {
	Timestamp      time.Time              `json:"timestamp"`
	GithomeVersion string                 `json:"githome_version"`
	Results        map[string]*GateResult `json:"results"`
	Summary        Summary                `json:"summary"`
}

// Summary aggregates pass/fail/skip counts.
type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// New returns an empty report stamped with now.
func New(version string) *Report {
	return &Report{
		Timestamp:      time.Now().UTC(),
		GithomeVersion: version,
		Results:        make(map[string]*GateResult),
	}
}

// Add records a gate result from a parsed TAP result.
func (r *Report) Add(gate string, status Status, durationMs int64, errMsg string) {
	r.Results[gate] = &GateResult{
		Status:     status,
		DurationMs: durationMs,
		Error:      errMsg,
	}
}

// AddTAP records all results from a TAP parse.
func (r *Report) AddTAP(results []tap.Result, durationMs int64) {
	for _, tr := range results {
		if tr.Gate == "" {
			continue
		}
		st := Pass
		if !tr.OK {
			st = Fail
		}
		if tr.OK && tr.Error != "" && len(tr.Error) > 5 && tr.Error[:5] == "SKIP " {
			st = Skip
		}
		r.Results[tr.Gate] = &GateResult{
			Status:     st,
			DurationMs: durationMs,
			Error:      tr.Error,
		}
	}
}

// Finalize computes the Summary counts.
func (r *Report) Finalize() {
	r.Summary.Total = len(r.Results)
	for _, gr := range r.Results {
		switch gr.Status {
		case Pass:
			r.Summary.Passed++
		case Fail:
			r.Summary.Failed++
		case Skip:
			r.Summary.Skipped++
		}
	}
}

// Write serializes the report as pretty-printed JSON.
func (r *Report) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// Matrix renders a text gate matrix to w. gateIDs is the full expected set.
func (r *Report) Matrix(w io.Writer, gateIDs []string) {
	ids := append([]string(nil), gateIDs...)
	sort.Strings(ids)
	for _, id := range ids {
		gr, ok := r.Results[id]
		if !ok {
			io.WriteString(w, "  MISSING "+id+"\n")
			continue
		}
		sym := "  pass"
		if gr.Status == Fail {
			sym = "  FAIL"
		} else if gr.Status == Skip {
			sym = "  skip"
		}
		io.WriteString(w, sym+" "+id+"\n")
	}
}
