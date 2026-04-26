package shells

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Fish is the fish-specific shell.
type Fish struct{ Generic }

func (f *Fish) FriendlyName() string { return "Fish Shell" }
func (f *Fish) Info() string         { return "Fish Shell" }

func (f *Fish) AppAlias(name string) string {
	return `function ` + name + ` -d "Correct your previous console command"
  set -l fucked_up_command $history[1]
  env TF_SHELL=fish TF_ALIAS=` + name + ` gofuck $fucked_up_command | read -l unfucked_command
  if [ "$unfucked_command" != "" ]
    eval $unfucked_command
    builtin history delete --exact --case-sensitive -- $fucked_up_command
    builtin history merge ^ /dev/null
  end
end`
}

func (f *Fish) And(cmds ...string) string { return join(cmds, "; and ") }
func (f *Fish) Or(cmds ...string) string  { return join(cmds, "; or ") }

func (f *Fish) GetHistory() []string {
	return readHistoryFile(expandHome("~/.config/fish/fish_history"), fishScriptFromHistory)
}

// PutToHistory appends a command to fish's history file. Bash/zsh do this
// at shell-level via the AppAlias function, but fish doesn't support that
// from a function, so the rule pipeline writes it directly. Errors are
// silently swallowed (mirrors upstream's "log and continue" behaviour).
func (f *Fish) PutToHistory(cmd string) {
	path := expandHome("~/.config/fish/fish_history")
	if _, err := os.Stat(path); err != nil {
		return
	}
	fp, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer fp.Close()
	_, _ = fmt.Fprintf(fp, "- cmd: %s\n   when: %d\n", cmd, time.Now().Unix())
}

// fishScriptFromHistory: fish history is yaml-ish. Each entry starts with
// `- cmd: <script>` followed by `  when: <unix>`. Lines that don't match are
// skipped.
func fishScriptFromHistory(line string) string {
	const prefix = "- cmd: "
	if i := strings.Index(line, prefix); i >= 0 {
		return line[i+len(prefix):]
	}
	return ""
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += sep + p
	}
	return out
}
