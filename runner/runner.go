// Package runner defines per-language runners and their orchestration.
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tamnd/github-compat/tap"
)

// Runner describes a single language/tool test subprocess.
type Runner struct {
	Name    string
	Command []string
	Gates   []string
	Timeout time.Duration
}

// scriptDir returns the path to the runners/ directory relative to this file.
func scriptDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "runners"
	}
	return filepath.Join(filepath.Dir(filepath.Dir(file)), "runners")
}

// All returns all registered runners.
func All() []Runner {
	sd := scriptDir()
	return []Runner{
		{
			Name:    "git",
			Command: []string{"bash", filepath.Join(sd, "shell", "git_test.sh")},
			Gates:   []string{"G1", "G2", "G3", "G4", "G5", "G6", "G7", "G10", "G11"},
			Timeout: 3 * time.Minute,
		},
		{
			Name:    "gh-cli",
			Command: []string{"bash", filepath.Join(sd, "shell", "gh_cli_test.sh")},
			Gates:   ghGates(),
			Timeout: 3 * time.Minute,
		},
		{
			Name:    "oct-js",
			Command: []string{"node", filepath.Join(sd, "js", "octokit_test.mjs")},
			Gates:   octGates(),
			Timeout: 2 * time.Minute,
		},
		{
			Name:    "python",
			Command: []string{"python3", filepath.Join(sd, "python", "pygithub_test.py")},
			Gates:   pyGates(),
			Timeout: 2 * time.Minute,
		},
		{
			Name:    "go-clients",
			Command: []string{"go", "run", filepath.Join(sd, "go", "main.go")},
			Gates:   goGates(),
			Timeout: 2 * time.Minute,
		},
		{
			Name:    "ruby",
			Command: []string{"ruby", filepath.Join(sd, "ruby", "octokit_rb_test.rb")},
			Gates:   rbGates(),
			Timeout: 2 * time.Minute,
		},
		{
			Name:    "dotnet",
			Command: []string{"dotnet", "test", filepath.Join(sd, "dotnet"), "--logger", "tap"},
			Gates:   csGates(),
			Timeout: 3 * time.Minute,
		},
		{
			Name:    "java",
			Command: []string{"mvn", "-pl", filepath.Join(sd, "java"), "-q", "test"},
			Gates:   javGates(),
			Timeout: 5 * time.Minute,
		},
	}
}

// Result is the outcome of running a single runner.
type Result struct {
	Runner     string
	TAPResults []tap.Result
	Duration   time.Duration
	Err        error
}

// Run executes a runner subprocess and parses its TAP output.
func (r *Runner) Run(ctx context.Context, env []string) Result {
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.Command[0], r.Command[1:]...)
	cmd.Env = env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()

	results, parseErr := tap.Parse(&stdout)
	if parseErr != nil && runErr == nil {
		runErr = parseErr
	}

	var err error
	if runErr != nil {
		err = fmt.Errorf("%w\nstderr: %s", runErr, strings.TrimSpace(stderr.String()))
	}

	return Result{
		Runner:     r.Name,
		TAPResults: results,
		Duration:   time.Since(start),
		Err:        err,
	}
}

func ghGates() []string {
	var gs []string
	for i := 1; i <= 22; i++ {
		gs = append(gs, fmt.Sprintf("GH-A%d", i))
	}
	return gs
}

func octGates() []string {
	var gs []string
	for i := 1; i <= 11; i++ {
		gs = append(gs, fmt.Sprintf("OCT-%d", i))
	}
	return gs
}

func pyGates() []string {
	var gs []string
	for i := 1; i <= 15; i++ {
		gs = append(gs, fmt.Sprintf("PY-%d", i))
	}
	return gs
}

func goGates() []string {
	var gs []string
	for i := 1; i <= 12; i++ {
		gs = append(gs, fmt.Sprintf("GOG-%d", i))
	}
	return gs
}

func rbGates() []string {
	var gs []string
	for i := 1; i <= 8; i++ {
		gs = append(gs, fmt.Sprintf("RB-%d", i))
	}
	return gs
}

func csGates() []string {
	var gs []string
	for i := 1; i <= 8; i++ {
		gs = append(gs, fmt.Sprintf("CS-%d", i))
	}
	return gs
}

func javGates() []string {
	var gs []string
	for i := 1; i <= 7; i++ {
		gs = append(gs, fmt.Sprintf("JAV-%d", i))
	}
	return gs
}
