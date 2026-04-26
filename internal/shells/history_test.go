package shells

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

// stringReadCloser is a tiny ReadCloser backed by a string for the
// historyOpenFile seam.
type stringReadCloser struct{ *strings.Reader }

func (stringReadCloser) Close() error { return nil }

// withFakeHistory swaps historyOpenFile to return `body` when asked for
// `wantPath`. Anything else returns os.ErrNotExist.
func withFakeHistory(t *testing.T, wantPath, body string) {
	t.Helper()
	prev := historyOpenFile
	t.Cleanup(func() { historyOpenFile = prev })
	historyOpenFile = func(name string) (io.ReadCloser, error) {
		if name != wantPath {
			t.Errorf("historyOpenFile called with %q, want %q", name, wantPath)
			return nil, io.EOF
		}
		return stringReadCloser{strings.NewReader(body)}, nil
	}
}

func withHomeDir(t *testing.T, dir string) {
	t.Helper()
	prev := historyHomeDir
	t.Cleanup(func() { historyHomeDir = prev })
	historyHomeDir = func() string { return dir }
}

func TestBashGetHistory(t *testing.T) {
	withHomeDir(t, "/home/u")
	withFakeHistory(t, "/home/u/.bash_history", "ls\ngit status\n\n  echo hi  \n")
	t.Setenv("HISTFILE", "")

	got := (&Bash{}).GetHistory()
	want := []string{"ls", "git status", "echo hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Bash.GetHistory = %v, want %v", got, want)
	}
}

func TestBashGetHistoryRespectsHISTFILE(t *testing.T) {
	withHomeDir(t, "/home/u")
	withFakeHistory(t, "/tmp/custom_history", "ls -la\n")
	t.Setenv("HISTFILE", "/tmp/custom_history")

	got := (&Bash{}).GetHistory()
	want := []string{"ls -la"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Bash.GetHistory($HISTFILE) = %v, want %v", got, want)
	}
}

func TestZshGetHistoryParsesExtended(t *testing.T) {
	withHomeDir(t, "/home/u")
	body := ": 1700000000:0;ls -la\n" +
		": 1700000001:0;git status\n" +
		"plain entry\n" +
		": 1700000002:0;echo hello\n"
	withFakeHistory(t, "/home/u/.zsh_history", body)
	t.Setenv("HISTFILE", "")

	got := (&Zsh{}).GetHistory()
	want := []string{"ls -la", "git status", "plain entry", "echo hello"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Zsh.GetHistory = %v, want %v", got, want)
	}
}

func TestFishGetHistoryParsesYAMLishLines(t *testing.T) {
	withHomeDir(t, "/home/u")
	body := "- cmd: ls -la\n   when: 1700000000\n" +
		"- cmd: git status\n   when: 1700000001\n" +
		"  paths:\n    - foo\n"
	withFakeHistory(t, "/home/u/.config/fish/fish_history", body)

	got := (&Fish{}).GetHistory()
	want := []string{"ls -la", "git status"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Fish.GetHistory = %v, want %v", got, want)
	}
}

func TestGetHistoryMissingFileReturnsNil(t *testing.T) {
	withHomeDir(t, "/home/u")
	t.Setenv("HISTFILE", "")
	prev := historyOpenFile
	t.Cleanup(func() { historyOpenFile = prev })
	historyOpenFile = func(string) (io.ReadCloser, error) { return nil, io.EOF }
	if got := (&Bash{}).GetHistory(); got != nil {
		t.Fatalf("expected nil for missing file, got %v", got)
	}
}

func TestHistoryLimitTrimsTail(t *testing.T) {
	withHomeDir(t, "/home/u")
	t.Setenv("HISTFILE", "")
	body := strings.Repeat("cmd\n", 5)
	withFakeHistory(t, "/home/u/.bash_history", body)

	prev := historyLimit
	t.Cleanup(func() { historyLimit = prev })
	historyLimit = 3

	got := (&Bash{}).GetHistory()
	if len(got) != 3 {
		t.Fatalf("expected 3 entries after limit, got %d (%v)", len(got), got)
	}
}

func TestParseAliasEnvBash(t *testing.T) {
	t.Setenv("TF_SHELL_ALIASES", "alias ll='ls -lah'\nalias gco='git checkout'\nalias bad-line\n")
	got := parseAliasEnv(true)
	want := map[string]string{"ll": "ls -lah", "gco": "git checkout"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseAliasEnv(true) = %v, want %v", got, want)
	}
}

func TestParseAliasEnvZsh(t *testing.T) {
	t.Setenv("TF_SHELL_ALIASES", `ll='ls -lah'
gco="git checkout"
empty=
`)
	got := parseAliasEnv(false)
	want := map[string]string{"ll": "ls -lah", "gco": "git checkout", "empty": ""}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseAliasEnv(false) = %v, want %v", got, want)
	}
}
