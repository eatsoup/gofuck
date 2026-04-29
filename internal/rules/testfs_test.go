package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTmpDir creates a fresh tmpdir, chdirs into it, and restores the previous
// cwd on test cleanup. Returns the absolute path to the tmpdir.
func withTmpDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir %q: %v", dir, err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })
	abs, err := filepath.EvalSymlinks(dir)
	if err != nil {
		abs = dir
	}
	return abs
}

// withPath sets PATH to the given dirs (joined with the OS list separator) for
// the duration of the test, mirroring upstream's mocker.patch on `which`.
func withPath(t *testing.T, dirs ...string) {
	t.Helper()
	t.Setenv("PATH", strings.Join(dirs, string(os.PathListSeparator)))
}

// touchFile creates `name` (relative to cwd) with mode 0o644 and the given
// contents. Use after withTmpDir.
func touchFile(t *testing.T, name, contents string) {
	t.Helper()
	if dir := filepath.Dir(name); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %q: %v", dir, err)
		}
	}
	if err := os.WriteFile(name, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %q: %v", name, err)
	}
}

// touchExec creates `name` (relative to cwd) marked executable.
func touchExec(t *testing.T, name string) {
	t.Helper()
	touchFile(t, name, "#!/bin/sh\nexit 0\n")
	if err := os.Chmod(name, 0o755); err != nil {
		t.Fatalf("chmod %q: %v", name, err)
	}
}
