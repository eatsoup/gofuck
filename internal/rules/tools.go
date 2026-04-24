package rules

import (
	"regexp"
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/specific"
	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

// Fallback static command lists (since we can't parse `tool --help` at runtime
// for every environment).
var (
	dockerCmds = []string{
		"attach", "build", "commit", "cp", "create", "diff", "events", "exec",
		"export", "history", "images", "import", "info", "inspect", "kill",
		"load", "login", "logout", "logs", "pause", "port", "ps", "pull",
		"push", "rename", "restart", "rm", "rmi", "run", "save", "search",
		"start", "stats", "stop", "tag", "top", "unpause", "update", "version", "wait",
		"container", "image", "network", "node", "plugin", "secret", "service",
		"stack", "swarm", "system", "volume",
	}
	golangCmds = []string{
		"bug", "build", "clean", "doc", "env", "fix", "fmt", "generate", "get",
		"install", "list", "mod", "run", "test", "tool", "version", "vet",
	}
	gemCmds = []string{
		"build", "cert", "check", "cleanup", "contents", "dependency",
		"environment", "fetch", "generate_index", "help", "install", "list",
		"lock", "mirror", "open", "outdated", "owner", "pristine", "push",
		"query", "rdoc", "search", "server", "sources", "specification",
		"stale", "uninstall", "unpack", "update", "which", "yank",
	}
	npmScripts = []string{"start", "test", "build", "lint", "dev", "serve", "watch"}
	yarnTasks  = []string{
		"access", "add", "autoclean", "cache", "check", "config", "create",
		"exec", "generate-lock-entry", "global", "help", "import", "info",
		"init", "install", "licenses", "link", "list", "login", "logout",
		"node", "outdated", "owner", "pack", "policies", "publish", "remove",
		"run", "tag", "team", "test", "unlink", "unplug", "upgrade",
		"upgrade-interactive", "version", "versions", "why", "workspace", "workspaces",
	}
)

func init() {
	// docker_image_being_used_by_container
	Register(&types.Rule{
		Name: "docker_image_being_used_by_container", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "docker") &&
				strings.Contains(c.Output, "image is being used by running container")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := strings.Split(strings.TrimSpace(c.Output), " ")
			id := parts[len(parts)-1]
			return []string{shells.Current.And("docker container rm -f "+id, c.Script)}
		},
	})

	// docker_login
	Register(&types.Rule{
		Name: "docker_login", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "docker") &&
				strings.Contains(c.Output, "access denied") &&
				strings.Contains(c.Output, "may require 'docker login'")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{shells.Current.And("docker login", c.Script)}
		},
	})

	// docker_not_command
	dockerNotRe := regexp.MustCompile(`docker: '(\w+)' is not a docker command.`)
	Register(&types.Rule{
		Name: "docker_not_command", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return utils.IsApp(c, 0, "docker") &&
				(strings.Contains(c.Output, "is not a docker command") ||
					strings.Contains(c.Output, "Usage:\tdocker"))
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			if strings.Contains(c.Output, "Usage:") && len(c.ScriptParts()) > 2 {
				// Management subcommand typo — reuse dockerCmds as approximation.
				return utils.ReplaceCommand(c, c.ScriptParts()[2], dockerCmds)
			}
			m := dockerNotRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, m[1], dockerCmds)
		}),
	})

	// npm_missing_script
	npmMissRe := regexp.MustCompile(`(?m).*missing script: (.*)$`)
	Register(&types.Rule{
		Name: "npm_missing_script", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "npm") {
				return false
			}
			hasRu := false
			for _, p := range c.ScriptParts() {
				if strings.HasPrefix(p, "ru") {
					hasRu = true
					break
				}
			}
			return hasRu && strings.Contains(c.Output, "npm ERR! missing script: ")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := npmMissRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, strings.TrimSpace(m[1]), npmScripts)
		},
	})

	// npm_run_script
	Register(&types.Rule{
		Name: "npm_run_script", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 1, "npm") {
				return false
			}
			parts := c.ScriptParts()
			for _, p := range parts {
				if strings.HasPrefix(p, "ru") {
					return false
				}
			}
			if !strings.Contains(c.Output, "Usage: npm <command>") {
				return false
			}
			for _, s := range npmScripts {
				if parts[1] == s {
					return true
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			out := append([]string{parts[0], "run-script"}, parts[1:]...)
			return []string{strings.Join(out, " ")}
		},
	})

	// npm_wrong_command — parses commands from the npm help block, like thefuck.
	Register(&types.Rule{
		Name: "npm_wrong_command", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) == 0 || parts[0] != "npm" {
				return false
			}
			if !strings.Contains(c.Output, "where <command> is one of:") {
				return false
			}
			for _, p := range parts[1:] {
				if !strings.HasPrefix(p, "-") {
					return true
				}
			}
			return false
		}),
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			var wrong string
			for _, p := range parts[1:] {
				if !strings.HasPrefix(p, "-") {
					wrong = p
					break
				}
			}
			if wrong == "" {
				return nil
			}
			var npmCmds []string
			listing := false
			for _, line := range strings.Split(c.Output, "\n") {
				if strings.HasPrefix(line, "where <command> is one of:") {
					listing = true
					continue
				}
				if listing {
					if strings.TrimSpace(line) == "" {
						break
					}
					for _, cmd := range strings.Split(line, ", ") {
						cmd = strings.TrimSpace(cmd)
						if cmd != "" {
							npmCmds = append(npmCmds, cmd)
						}
					}
				}
			}
			fixed := utils.GetClosest(wrong, npmCmds, 0.6, true)
			return []string{utils.ReplaceArgument(c.Script, wrong, fixed)}
		},
	})

	// yarn_alias
	yarnAliasRe := regexp.MustCompile(`Did you mean [` + "`" + `"](?:yarn )?([^` + "`" + `"]*)[` + "`" + `"]`)
	Register(&types.Rule{
		Name: "yarn_alias", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 1, "yarn") && strings.Contains(c.Output, "Did you mean")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := yarnAliasRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{utils.ReplaceArgument(c.Script, c.ScriptParts()[1], m[1])}
		},
	})

	// yarn_command_not_found
	yarnCnfRe := regexp.MustCompile(`error Command "(.*)" not found\.`)
	Register(&types.Rule{
		Name: "yarn_command_not_found", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "yarn") && yarnCnfRe.MatchString(c.Output)
		},
		GetNewCommand: func(c *types.Command) []string {
			m := yarnCnfRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			task := m[1]
			if task == "require" {
				return []string{utils.ReplaceArgument(c.Script, "require", "add")}
			}
			return utils.ReplaceCommand(c, task, yarnTasks)
		},
	})

	// yarn_command_replaced
	yarnReplRe := regexp.MustCompile(`Run "(.*)" instead`)
	Register(&types.Rule{
		Name: "yarn_command_replaced", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 1, "yarn") && yarnReplRe.MatchString(c.Output)
		},
		GetNewCommand: func(c *types.Command) []string {
			m := yarnReplRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{m[1]}
		},
	})

	// yarn_help
	yarnHelpRe := regexp.MustCompile(`Visit ([^ ]*) for documentation about this command\.`)
	Register(&types.Rule{
		Name: "yarn_help", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 2, "yarn") && c.ScriptParts()[1] == "help" &&
				strings.Contains(c.Output, "for documentation about this command.")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := yarnHelpRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{"xdg-open " + m[1]}
		},
	})

	// cargo
	Register(&types.Rule{
		Name: "cargo", EnabledByDefault: true, RequiresOutput: false,
		Match: func(c *types.Command) bool { return c.Script == "cargo" },
		GetNewCommand: func(c *types.Command) []string {
			return []string{"cargo build"}
		},
	})

	// cargo_no_command
	cargoNoCmdRe := regexp.MustCompile("Did you mean `([^`]*)`")
	Register(&types.Rule{
		Name: "cargo_no_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 1, "cargo") &&
				strings.Contains(strings.ToLower(c.Output), "no such subcommand") &&
				strings.Contains(c.Output, "Did you mean")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			m := cargoNoCmdRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{utils.ReplaceArgument(c.Script, parts[1], m[1])}
		},
	})

	// go_run
	Register(&types.Rule{
		Name: "go_run", EnabledByDefault: true, RequiresOutput: false,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "go") &&
				strings.HasPrefix(c.Script, "go run ") && !strings.HasSuffix(c.Script, ".go")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script + ".go"}
		},
	})

	// go_unknown_command
	Register(&types.Rule{
		Name: "go_unknown_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "go") && strings.Contains(c.Output, "unknown command")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			if len(parts) < 2 {
				return nil
			}
			closest := utils.GetClosest(parts[1], golangCmds, 0.6, true)
			return []string{utils.ReplaceArgument(c.Script, parts[1], closest)}
		},
	})

	// gem_unknown_command
	gemUnkRe := regexp.MustCompile(`Unknown command (.*)$`)
	Register(&types.Rule{
		Name: "gem_unknown_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "gem") &&
				strings.Contains(c.Output, "ERROR:  While executing gem ... (Gem::CommandLineError)") &&
				strings.Contains(c.Output, "Unknown command")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := gemUnkRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, strings.TrimSpace(m[1]), gemCmds)
		},
	})

	// composer_not_command
	composerNotDefRe := regexp.MustCompile(`Command "([^']*)" is not defined`)
	composerDidMeanRe := regexp.MustCompile(`(?s)Did you mean this\?[^\n]*\n\s*([^\n]*)`)
	composerDidMeanManyRe := regexp.MustCompile(`(?s)Did you mean one of these\?[^\n]*\n\s*([^\n]*)`)
	Register(&types.Rule{
		Name: "composer_not_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "composer") {
				return false
			}
			lo := strings.ToLower(c.Output)
			if strings.Contains(lo, "did you mean this?") || strings.Contains(lo, "did you mean one of these?") {
				return true
			}
			parts := c.ScriptParts()
			for _, p := range parts {
				if p == "install" {
					if strings.Contains(lo, "composer require") {
						return true
					}
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			lo := strings.ToLower(c.Output)
			parts := c.ScriptParts()
			hasInstall := false
			for _, p := range parts {
				if p == "install" {
					hasInstall = true
				}
			}
			if hasInstall && strings.Contains(lo, "composer require") {
				return []string{utils.ReplaceArgument(c.Script, "install", "require")}
			}
			brokenM := composerNotDefRe.FindStringSubmatch(c.Output)
			if brokenM == nil {
				return nil
			}
			var newCmd string
			if m := composerDidMeanRe.FindStringSubmatch(c.Output); m != nil {
				newCmd = strings.TrimSpace(m[1])
			} else if m := composerDidMeanManyRe.FindStringSubmatch(c.Output); m != nil {
				newCmd = strings.TrimSpace(m[1])
			}
			if newCmd == "" {
				return nil
			}
			return []string{utils.ReplaceArgument(c.Script, brokenM[1], newCmd)}
		},
	})

	// heroku_multiple_apps
	herokuAppRe := regexp.MustCompile(`([^ ]*) \([^)]*\)`)
	Register(&types.Rule{
		Name: "heroku_multiple_apps", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "heroku") &&
				strings.Contains(c.Output, "https://devcenter.heroku.com/articles/multiple-environments")
		},
		GetNewCommand: func(c *types.Command) []string {
			apps := herokuAppRe.FindAllStringSubmatch(c.Output, -1)
			out := make([]string, 0, len(apps))
			for _, a := range apps {
				out = append(out, c.Script+" --app "+a[1])
			}
			return out
		},
	})

	// heroku_not_command
	herokuNotRe := regexp.MustCompile(`Run heroku _ to run ([^.]*)`)
	Register(&types.Rule{
		Name: "heroku_not_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "heroku") && strings.Contains(c.Output, "Run heroku _ to run")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := herokuNotRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{m[1]}
		},
	})

	// tsuru_login
	Register(&types.Rule{
		Name: "tsuru_login", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "tsuru") &&
				strings.Contains(c.Output, "not authenticated") &&
				strings.Contains(c.Output, "session has expired")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{shells.Current.And("tsuru login", c.Script)}
		},
	})

	// tsuru_not_command
	tsuruNotRe := regexp.MustCompile(`tsuru: "([^"]*)" is not a tsuru command`)
	Register(&types.Rule{
		Name: "tsuru_not_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "tsuru") &&
				strings.Contains(c.Output, ` is not a tsuru command. See "tsuru help".`) &&
				strings.Contains(c.Output, "\nDid you mean?\n\t")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := tsuruNotRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, m[1], utils.GetAllMatchedCommands(c.Output, "Did you mean"))
		},
	})
}
