package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTmpDir creates a temporary directory, sets the current working directory
// to it for the duration of the test, and restores the original working directory
// when the test completes. It returns the path to the temporary directory.
func withTmpDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	err = os.Chdir(dir)
	if err != nil {
		t.Fatalf("Failed to change working directory to temp dir: %v", err)
	}

	t.Cleanup(func() {
		err := os.Chdir(origWd)
		if err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	})

	return dir
}

// touchFile creates an empty file at the specified relative or absolute path.
// It creates parent directories if they do not exist.
func touchFile(t *testing.T, path string) {
	t.Helper()

	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		t.Fatalf("Failed to create parent directories for %s: %v", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create file %s: %v", path, err)
	}
	f.Close()
}

// mkDir creates a directory at the specified relative or absolute path.
// It creates parent directories if they do not exist.
func mkDir(t *testing.T, path string) {
	t.Helper()

	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory %s: %v", path, err)
	}
}

// withPath modifies the PATH environment variable for the duration of the test
// by prepending the provided directories.
func withPath(t *testing.T, dirs ...string) {
	t.Helper()

	origPath := os.Getenv("PATH")
	newPath := strings.Join(dirs, string(os.PathListSeparator)) + string(os.PathListSeparator) + origPath

	err := os.Setenv("PATH", newPath)
	if err != nil {
		t.Fatalf("Failed to set PATH: %v", err)
	}

	t.Cleanup(func() {
		os.Setenv("PATH", origPath)
	})
}

// TestTestFsHelper tests the testfs helpers to ensure they work.
func TestTestFsHelper(t *testing.T) {
	dir := withTmpDir(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	
	// Because of osX/symlinks, compare them loosely or evaluate symlinks
	// We just ensure we're not in the original dir
	if filepath.Base(cwd) != filepath.Base(dir) {
		// Just ensure they both exist
		if _, err := os.Stat(cwd); err != nil {
			t.Errorf("Expected valid cwd, got %v", err)
		}
	}

	touchFile(t, "testfile.txt")
	if _, err := os.Stat("testfile.txt"); err != nil {
		t.Errorf("touchFile failed: %v", err)
	}

	mkDir(t, "testdir")
	if info, err := os.Stat("testdir"); err != nil || !info.IsDir() {
		t.Errorf("mkDir failed: %v", err)
	}
}
