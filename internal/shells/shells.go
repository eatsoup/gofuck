package shells

import (
	"os"
	"strings"
)

// Shell is the interface implemented by every shell integration (bash, zsh, fish, generic).
type Shell interface {
	FriendlyName() string
	GetAliases() map[string]string
	FromShell(script string) string
	ToShell(script string) string
	AppAlias(name string) string
	And(cmds ...string) string
	Or(cmds ...string) string
	SplitCommand(cmd string) []string
	Quote(s string) string
	GetBuiltinCommands() []string
	GetHistory() []string
	PutToHistory(cmd string)
	Info() string
}

// Current shell — set by CLI at startup. Tests typically set this to Generic.
var Current Shell = &Generic{}

// Generic is the fallback shell implementation (POSIX-ish defaults).
type Generic struct{}

func (g *Generic) FriendlyName() string         { return "Generic Shell" }
func (g *Generic) GetAliases() map[string]string { return map[string]string{} }

func (g *Generic) FromShell(script string) string {
	aliases := g.GetAliases()
	parts := strings.SplitN(script, " ", 2)
	binary := parts[0]
	if repl, ok := aliases[binary]; ok {
		return strings.Replace(script, binary, repl, 1)
	}
	return script
}

func (g *Generic) ToShell(script string) string { return script }

func (g *Generic) AppAlias(name string) string {
	// Go port: mirrors thefuck's generic.app_alias.
	return `alias ` + name + `='eval "$(TF_ALIAS=` + name + ` gofuck "$(fc -ln -1)")"'`
}

func (g *Generic) And(cmds ...string) string { return strings.Join(cmds, " && ") }
func (g *Generic) Or(cmds ...string) string  { return strings.Join(cmds, " || ") }

// SplitCommand splits like shlex with the \ + space preservation hack from thefuck.
func (g *Generic) SplitCommand(cmd string) []string {
	replaced := strings.ReplaceAll(cmd, `\ `, "??")
	parts, err := shlexSplit(replaced)
	if err != nil {
		parts = strings.Split(cmd, " ")
	}
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = strings.ReplaceAll(p, "??", `\ `)
	}
	return out
}

func (g *Generic) Quote(s string) string {
	if s == "" {
		return "''"
	}
	safe := true
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			r == '@' || r == '%' || r == '+' || r == '=' || r == ':' ||
			r == ',' || r == '.' || r == '/' || r == '-' || r == '_') {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func (g *Generic) GetBuiltinCommands() []string {
	return []string{
		"alias", "bg", "bind", "break", "builtin", "case", "cd", "command",
		"compgen", "complete", "continue", "declare", "dirs", "disown", "echo",
		"enable", "eval", "exec", "exit", "export", "fc", "fg", "getopts",
		"hash", "help", "history", "if", "jobs", "kill", "let", "local",
		"logout", "popd", "printf", "pushd", "pwd", "read", "readonly",
		"return", "set", "shift", "shopt", "source", "suspend", "test", "times",
		"trap", "type", "typeset", "ulimit", "umask", "unalias", "unset",
		"until", "wait", "while",
	}
}

func (g *Generic) GetHistory() []string  { return nil }
func (g *Generic) PutToHistory(cmd string) {}
func (g *Generic) Info() string           { return "Generic Shell" }

// Use updates the global current shell based on the $SHELL env var.
func Use(shellEnv string) {
	switch {
	case strings.Contains(shellEnv, "bash"):
		Current = &Bash{}
	case strings.Contains(shellEnv, "zsh"):
		Current = &Zsh{}
	case strings.Contains(shellEnv, "fish"):
		Current = &Fish{}
	default:
		Current = &Generic{}
	}
}

// Auto picks a shell based on env.
func Auto() {
	Use(os.Getenv("SHELL"))
}
