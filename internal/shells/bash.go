package shells

// Bash is the bash-specific shell.
type Bash struct{ Generic }

func (b *Bash) FriendlyName() string { return "Bash" }
func (b *Bash) Info() string         { return "Bash" }

// AppAlias mirrors thefuck/shells/bash.py — a multi-line shell function that
// exports the env vars gofuck consults (TF_SHELL, TF_ALIAS, TF_SHELL_ALIASES,
// TF_HISTORY) before invoking gofuck and eval'ing the result. The result is
// also pushed to bash's interactive history so the corrected command is
// recallable.
func (b *Bash) AppAlias(name string) string {
	return `function ` + name + ` () {
    export TF_SHELL=bash;
    export TF_ALIAS=` + name + `;
    export TF_SHELL_ALIASES=$(alias);
    export TF_HISTORY=$(fc -ln -10);
    TF_CMD=$(gofuck "$@") && eval "$TF_CMD";
    unset TF_HISTORY;
    history -s "$TF_CMD";
}`
}

func (b *Bash) GetHistory() []string {
	return readHistoryFile(envOrDefault("HISTFILE", "~/.bash_history"), bashScriptFromHistory)
}

// GetAliases parses TF_SHELL_ALIASES (the output of `alias` that the bash
// AppAlias function exports). Each line looks like `alias name='value'`.
func (b *Bash) GetAliases() map[string]string {
	return parseAliasEnv(true)
}

// bashScriptFromHistory: bash history is plain command lines, one per entry.
func bashScriptFromHistory(line string) string { return line }
