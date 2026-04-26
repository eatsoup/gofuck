// Package exec is a thin shim around os/exec that gives rules a swappable
// subprocess seam. Upstream thefuck calls subprocess.Popen directly inside
// rules; we hide that behind Runner so tests can inject canned outputs
// without touching the filesystem.
package exec

import (
	"bytes"
	"os/exec"
	"time"
)

// Result holds the outcome of a captured subprocess call.
type Result struct {
	Stdout []byte
	Stderr []byte
	Err    error
}

// OK reports whether the call produced any output and didn't error.
// Helpful guard for "fall back to a static list when the tool isn't installed".
func (r Result) OK() bool {
	return r.Err == nil && (len(r.Stdout) > 0 || len(r.Stderr) > 0)
}

// RunFn is the signature tests substitute. The default value calls os/exec.
type RunFn func(name string, args ...string) Result

// Default invokes os/exec with a 2s deadline so a hanging subprocess can't
// freeze rule evaluation.
var Default RunFn = func(name string, args ...string) Result {
	cmd := exec.Command(name, args...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return Result{Err: err}
	}
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return Result{Stdout: out.Bytes(), Stderr: errb.Bytes(), Err: err}
	case <-time.After(2 * time.Second):
		_ = cmd.Process.Kill()
		<-done
		return Result{Stdout: out.Bytes(), Stderr: errb.Bytes(), Err: context_DeadlineExceededSentinel}
	}
}

// Runner is the seam rules call. Tests swap it via WithRunner.
var Runner = Default

// Run is the public entry point rules use.
func Run(name string, args ...string) Result {
	return Runner(name, args...)
}

// WithRunner swaps Runner for the lifetime of the returned restore() func.
// Idiomatic test usage:
//
//	defer exec.WithRunner(func(name string, args ...string) exec.Result {
//	    return exec.Result{Stdout: []byte("...")}
//	})()
func WithRunner(r RunFn) func() {
	prev := Runner
	Runner = r
	return func() { Runner = prev }
}

// context_DeadlineExceededSentinel is a stable sentinel for the 2s timeout
// path; we don't pull in context just for this.
var context_DeadlineExceededSentinel = &timeoutErr{}

type timeoutErr struct{}

func (t *timeoutErr) Error() string { return "exec: subprocess timed out" }
