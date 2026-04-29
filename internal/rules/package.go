package rules

import (
	"regexp"
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/specific"
	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

// Fallback operation lists for apt / dnf / yum (used when we can't call --help).
var aptOps = []string{
	"autoremove", "build-dep", "changelog", "check", "clean", "deselect-upgrade",
	"dist-upgrade", "download", "dselect-upgrade", "full-upgrade", "install",
	"list", "moo", "purge", "reinstall", "remove", "search", "show", "source",
	"update", "upgrade",
}

var dnfOps = []string{
	"autoremove", "check", "check-update", "clean", "distro-sync", "downgrade",
	"group", "help", "history", "info", "install", "list", "makecache", "mark",
	"provides", "reinstall", "remove", "repolist", "repoquery", "repository-packages",
	"search", "shell", "swap", "updateinfo", "upgrade", "upgrade-minimal",
}

var yumOps = []string{
	"check", "check-update", "clean", "deplist", "distribution-synchronization",
	"downgrade", "erase", "fs", "fssnapshot", "groups", "help", "history", "info",
	"install", "list", "load-transaction", "localinstall", "makecache", "provides",
	"reinstall", "remove", "repo-pkgs", "repolist", "search", "shell", "swap",
	"update", "update-minimal", "updateinfo", "upgrade", "version",
}

// brewCmds is the static fallback command list for brew_unknown_command.
// Lifted to package scope so tests can iterate it the way upstream does
// (`for command in _brew_commands(): assert not match(...)`).
var brewCmds = []string{
	"info", "home", "options", "install", "uninstall",
	"search", "list", "update", "upgrade", "pin", "unpin",
	"doctor", "create", "edit", "cask",
}

func init() {
	// apt_get_search
	aptGetPat := regexp.MustCompile(`^apt-get`)
	Register(&types.Rule{
		Name: "apt_get_search", EnabledByDefault: true, RequiresOutput: false,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "apt-get") && strings.HasPrefix(c.Script, "apt-get search")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{aptGetPat.ReplaceAllString(c.Script, "apt-cache")}
		},
	})

	// apt_list_upgradable
	Register(&types.Rule{
		Name: "apt_list_upgradable", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return utils.IsApp(c, 0, "apt") && strings.Contains(c.Output, "apt list --upgradable")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{"apt list --upgradable"}
		}),
	})

	// apt_upgrade
	Register(&types.Rule{
		Name: "apt_upgrade", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "apt") {
				return false
			}
			return c.Script == "apt list --upgradable" &&
				len(strings.Split(strings.TrimSpace(c.Output), "\n")) > 1
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{"apt upgrade"}
		}),
	})

	// apt_invalid_operation
	Register(&types.Rule{
		Name: "apt_invalid_operation", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return utils.IsApp(c, 0, "apt", "apt-get", "apt-cache") &&
				strings.Contains(c.Output, "E: Invalid operation")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			words := strings.Fields(c.Output)
			if len(words) == 0 {
				return nil
			}
			invalid := words[len(words)-1]
			if invalid == "uninstall" {
				return []string{strings.Replace(c.Script, "uninstall", "remove", 1)}
			}
			return utils.ReplaceCommand(c, invalid, getAptOps(c.ScriptParts()[0]))
		}),
	})

	// dnf_no_such_command
	dnfRe := regexp.MustCompile(`No such command: (.*)\.`)
	Register(&types.Rule{
		Name: "dnf_no_such_command", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return utils.IsApp(c, 0, "dnf") && strings.Contains(strings.ToLower(c.Output), "no such command")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			m := dnfRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, m[1], getDnfOps())
		}),
	})

	// yum_invalid_operation
	Register(&types.Rule{
		Name: "yum_invalid_operation", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return utils.IsApp(c, 0, "yum") && strings.Contains(c.Output, "No such command: ")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			parts := c.ScriptParts()
			if len(parts) < 2 {
				return nil
			}
			inv := parts[1]
			if inv == "uninstall" {
				return []string{strings.Replace(c.Script, "uninstall", "remove", 1)}
			}
			return utils.ReplaceCommand(c, inv, getYumOps())
		}),
	})

	// brew_install
	brewInstallRe := regexp.MustCompile(`Warning: No available formula with the name "(?:[^"]+)"\. Did you mean (.+)\?`)
	Register(&types.Rule{
		Name: "brew_install", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 1, "brew") {
				return false
			}
			return strings.Contains(c.Script, "install") &&
				strings.Contains(c.Output, "No available formula") &&
				strings.Contains(c.Output, "Did you mean")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := brewInstallRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			sug := strings.Split(strings.ReplaceAll(m[1], " or ", ", "), ", ")
			out := make([]string, len(sug))
			for i, s := range sug {
				out[i] = "brew install " + s
			}
			return out
		},
	})

	// brew_link
	Register(&types.Rule{
		Name: "brew_link", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 1, "brew") {
				return false
			}
			parts := c.ScriptParts()
			return (parts[1] == "ln" || parts[1] == "link") &&
				strings.Contains(c.Output, "brew link --overwrite --dry-run")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			parts[1] = "link"
			out := append([]string{parts[0], parts[1], "--overwrite", "--dry-run"}, parts[2:]...)
			return []string{strings.Join(out, " ")}
		},
	})

	// brew_reinstall
	Register(&types.Rule{
		Name: "brew_reinstall", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 1, "brew") {
				return false
			}
			parts := c.ScriptParts()
			if len(parts) < 2 || parts[1] != "install" {
				return false
			}
			return strings.Contains(c.Output, "is already installed and up-to-date") &&
				strings.Contains(c.Output, "To reinstall")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.Replace(c.Script, "install", "reinstall", 1)}
		},
	})

	// brew_uninstall
	Register(&types.Rule{
		Name: "brew_uninstall", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 1, "brew") {
				return false
			}
			p := c.ScriptParts()
			return (p[1] == "uninstall" || p[1] == "rm" || p[1] == "remove") &&
				strings.Contains(c.Output, "brew uninstall --force")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			parts[1] = "uninstall"
			out := append([]string{parts[0], parts[1], "--force"}, parts[2:]...)
			return []string{strings.Join(out, " ")}
		},
	})

	// brew_unknown_command — use a static fallback command list.
	brewUnkRe := regexp.MustCompile(`Error: Unknown command: ([a-z]+)`)
	Register(&types.Rule{
		Name: "brew_unknown_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !(strings.Contains(c.Script, "brew") && strings.Contains(c.Output, "Unknown command")) {
				return false
			}
			m := brewUnkRe.FindStringSubmatch(c.Output)
			if m == nil {
				return false
			}
			return utils.GetClosest(m[1], brewCmds, 0.6, false) != ""
		},
		GetNewCommand: func(c *types.Command) []string {
			m := brewUnkRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, m[1], brewCmds)
		},
	})

	// brew_update_formula
	Register(&types.Rule{
		Name: "brew_update_formula", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 1, "brew") && strings.Contains(c.Script, "update") &&
				strings.Contains(c.Output, "Error: This command updates brew itself") &&
				strings.Contains(c.Output, "Use `brew upgrade")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.Replace(c.Script, "update", "upgrade", 1)}
		},
	})

	// brew_cask_dependency
	Register(&types.Rule{
		Name: "brew_cask_dependency", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			parts := c.ScriptParts()
			hasInstall := false
			for _, p := range parts {
				if p == "install" {
					hasInstall = true
					break
				}
			}
			return utils.IsApp(c, 0, "brew") && hasInstall && strings.Contains(c.Output, "brew cask install")
		},
		GetNewCommand: func(c *types.Command) []string {
			var caskLines []string
			for _, line := range strings.Split(c.Output, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "brew cask install") {
					caskLines = append(caskLines, line)
				}
			}
			if len(caskLines) == 0 {
				return nil
			}
			all := caskLines
			script := shells.Current.And(all...)
			return []string{shells.Current.And(script, c.Script)}
		},
	})

	// pacman_invalid_option
	pacmanInvRe := regexp.MustCompile(` -[dfqrstuv]`)
	Register(&types.Rule{
		Name: "pacman_invalid_option", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "pacman") {
				return false
			}
			if !strings.HasPrefix(c.Output, "error: invalid option '-") {
				return false
			}
			for _, o := range "surqfdvt" {
				if strings.Contains(c.Script, " -"+string(o)) {
					return true
				}
			}
			return false
		}),
		GetNewCommand: func(c *types.Command) []string {
			m := pacmanInvRe.FindStringIndex(c.Script)
			if m == nil {
				return nil
			}
			opt := c.Script[m[0]:m[1]]
			return []string{strings.Replace(c.Script, opt, strings.ToUpper(opt), 1)}
		},
	})

	// choco_install: adds .install suffix to the package name.
	chocoBlacklist := map[string]bool{"list": true, "search": true, "info": true, "-?": true, "-h": true, "--help": true}
	Register(&types.Rule{
		Name: "choco_install", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) < 2 {
				return false
			}
			bin := parts[0]
			if bin != "choco" && bin != "cinst" {
				return false
			}
			sub := ""
			if bin == "choco" {
				if len(parts) < 3 {
					return false
				}
				if parts[1] != "install" {
					return false
				}
				sub = parts[2]
			} else {
				sub = parts[1]
			}
			return !chocoBlacklist[sub] && !strings.HasSuffix(sub, ".install")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			idx := 1
			if parts[0] == "choco" {
				idx = 2
			}
			parts[idx] = parts[idx] + ".install"
			return []string{strings.Join(parts, " ")}
		},
	})
}
