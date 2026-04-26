package shells

import "strings"

// Zsh is the zsh-specific shell.
type Zsh struct{ Generic }

func (z *Zsh) FriendlyName() string { return "ZSH" }
func (z *Zsh) Info() string         { return "ZSH" }

// AppAlias mirrors thefuck/shells/zsh.py.
func (z *Zsh) AppAlias(name string) string {
	return name + ` () {
    export TF_SHELL=zsh;
    export TF_ALIAS=` + name + `;
    TF_SHELL_ALIASES=$(alias);
    export TF_SHELL_ALIASES;
    TF_HISTORY="$(fc -ln -10)";
    export TF_HISTORY;
    TF_CMD=$(gofuck "$@") && eval $TF_CMD;
    unset TF_HISTORY;
    test -n "$TF_CMD" && print -s $TF_CMD;
}`
}

func (z *Zsh) GetHistory() []string {
	return readHistoryFile(envOrDefault("HISTFILE", "~/.zsh_history"), zshScriptFromHistory)
}

// GetAliases — zsh's `alias` output omits the leading `alias ` keyword.
func (z *Zsh) GetAliases() map[string]string {
	return parseAliasEnv(false)
}

// zshScriptFromHistory: zsh extended history lines look like
// `: 1700000000:0;ls -la`. The actual script is everything after the first ';'.
// Plain (non-extended) lines pass through untouched.
func zshScriptFromHistory(line string) string {
	if strings.HasPrefix(line, ": ") {
		if i := strings.Index(line, ";"); i >= 0 {
			return line[i+1:]
		}
		return ""
	}
	return line
}
