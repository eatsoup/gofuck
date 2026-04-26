package shells

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// historyOpenFile is the seam tests swap to inject a fixture history file
// (or simulate "file missing"). Default is os.Open.
var historyOpenFile = func(name string) (io.ReadCloser, error) { return os.Open(name) }

// historyHomeDir returns the user's home directory. Swappable for tests.
var historyHomeDir = func() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return os.Getenv("HOME")
}

// historyLimit caps how many lines are read from the tail of the history
// file. 0 means "no limit". Mirrors thefuck.conf.history_limit.
var historyLimit = 2000

// readHistoryFile reads the named history file, splits it into lines,
// applies parseLine to each (which extracts the bare command from any
// shell-specific framing), trims the result, and returns non-empty entries
// in chronological order. Returns nil when the file is missing or unreadable.
func readHistoryFile(name string, parseLine func(string) string) []string {
	f, err := historyOpenFile(name)
	if err != nil {
		return nil
	}
	defer f.Close()
	var raw []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024) // some history lines are long (multi-line zsh entries)
	for scanner.Scan() {
		raw = append(raw, scanner.Text())
	}
	if historyLimit > 0 && len(raw) > historyLimit {
		raw = raw[len(raw)-historyLimit:]
	}
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		s := strings.TrimSpace(parseLine(line))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// expandHome expands a leading "~/" using historyHomeDir.
func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		return filepath.Join(historyHomeDir(), p[2:])
	}
	if p == "~" {
		return historyHomeDir()
	}
	return p
}

// envOrDefault returns os.Getenv(env) if non-empty, else def (run through
// expandHome so callers can pass "~/.bash_history").
func envOrDefault(env, def string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return expandHome(def)
}

// parseAliasEnv parses TF_SHELL_ALIASES, the env var that the AppAlias bash
// function exports (`export TF_SHELL_ALIASES=$(alias)`). If stripPrefix is
// true, the leading `alias ` is dropped from each line — that's bash's
// `alias` output format. zsh's `alias` already omits it.
func parseAliasEnv(stripPrefix bool) map[string]string {
	out := map[string]string{}
	raw := os.Getenv("TF_SHELL_ALIASES")
	for _, line := range strings.Split(raw, "\n") {
		if line == "" {
			continue
		}
		if stripPrefix {
			line = strings.TrimPrefix(line, "alias ")
		}
		i := strings.Index(line, "=")
		if i <= 0 {
			continue
		}
		name, value := line[:i], line[i+1:]
		if len(value) >= 2 && value[0] == value[len(value)-1] && (value[0] == '"' || value[0] == '\'') {
			value = value[1 : len(value)-1]
		}
		out[name] = value
	}
	return out
}
