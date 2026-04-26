package rules

import (
	"testing"

	"github.com/eatsoup/gofuck/internal/types"
)

// withHistory swaps historyValidEntries (mirrors upstream's mocker.patch on
// get_valid_history_without_current). Restores on cleanup.
func withHistory(t *testing.T, entries []string) {
	t.Helper()
	prev := historyValidEntries
	t.Cleanup(func() { historyValidEntries = prev })
	historyValidEntries = func(*types.Command) []string { return entries }
}

func TestHistoryMatch(t *testing.T) {
	withHistory(t, []string{"ls cat", "diff x"})
	for _, script := range []string{"ls cet", "daff x"} {
		assertMatch(t, "history", cmd(script, ""), true)
	}
}

func TestHistoryNotMatch(t *testing.T) {
	withHistory(t, []string{"ls cat", "diff x"})
	for _, script := range []string{"apt-get", "nocommand y"} {
		assertMatch(t, "history", cmd(script, ""), false)
	}
}

func TestHistoryNewCommand(t *testing.T) {
	withHistory(t, []string{"ls cat", "diff x"})
	cases := []struct {
		script, want string
	}{
		{"ls cet", "ls cat"},
		{"daff x", "diff x"},
	}
	for _, tc := range cases {
		assertNewCommand(t, "history", cmd(tc.script, ""), tc.want)
	}
}
