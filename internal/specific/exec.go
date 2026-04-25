package specific

import (
	"bytes"
	"os/exec"
)

// RunnerFunction defines the signature for running system commands.
type RunnerFunction func(name string, args ...string) (stdout string, stderr string, err error)

// Runner is the swappable seam for executing subprocesses.
// Unit tests should replace this to avoid executing real commands.
var Runner RunnerFunction = DefaultRunner

// DefaultRunner uses os/exec to run commands.
func DefaultRunner(name string, args ...string) (string, string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Run is a convenience function that delegates to Runner.
func Run(name string, args ...string) (string, string, error) {
	return Runner(name, args...)
}
