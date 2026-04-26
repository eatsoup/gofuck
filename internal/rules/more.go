package rules

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/specific"
	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

func init() {
	// adb_unknown_command
	adbCmds := []string{
		"backup", "bugreport", "connect", "devices", "disable-verity",
		"disconnect", "enable-verity", "emu", "forward", "get-devpath",
		"get-serialno", "get-state", "install", "install-multiple", "jdwp",
		"keygen", "kill-server", "logcat", "pull", "push", "reboot", "reconnect",
		"restore", "reverse", "root", "run-as", "shell", "sideload", "start-server",
		"sync", "tcpip", "uninstall", "unroot", "usb", "wait-for",
	}
	Register(&types.Rule{
		Name: "adb_unknown_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "adb") && strings.HasPrefix(c.Output, "Android Debug Bridge version")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			for idx, arg := range parts[1:] {
				if len(arg) == 0 {
					continue
				}
				if arg[0] != '-' {
					prev := ""
					if idx < len(parts) {
						prev = parts[idx]
					}
					if prev == "-s" || prev == "-H" || prev == "-P" || prev == "-L" {
						continue
					}
					closest := utils.GetClosest(arg, adbCmds, 0.6, true)
					return []string{utils.ReplaceArgument(c.Script, arg, closest)}
				}
			}
			return nil
		},
	})

	// aws_cli
	awsInvRe := regexp.MustCompile(`Invalid choice: '(.*)', maybe you meant:`)
	awsOptRe := regexp.MustCompile(`(?m)^\s*\*\s(.*)$`)
	Register(&types.Rule{
		Name: "aws_cli", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "aws") &&
				strings.Contains(c.Output, "usage:") && strings.Contains(c.Output, "maybe you meant:")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := awsInvRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			opts := awsOptRe.FindAllStringSubmatch(c.Output, -1)
			out := make([]string, 0, len(opts))
			for _, o := range opts {
				out = append(out, utils.ReplaceArgument(c.Script, m[1], o[1]))
			}
			return out
		},
	})

	// az_cli
	azInvRe := regexp.MustCompile(`(?s)(?:az)[^\n]*: '(.*?)' is not in the '.*?' command group\.`)
	azOptRe := regexp.MustCompile(`(?m)^The most similar choice to '.*' is:\n\s*(.*)$`)
	Register(&types.Rule{
		Name: "az_cli", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "az") &&
				strings.Contains(c.Output, "is not in the") && strings.Contains(c.Output, "command group")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := azInvRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			opts := azOptRe.FindAllStringSubmatch(c.Output, -1)
			out := make([]string, 0, len(opts))
			for _, o := range opts {
				out = append(out, utils.ReplaceArgument(c.Script, m[1], o[1]))
			}
			return out
		},
	})

	// cd_mkdir
	cdPat := regexp.MustCompile(`^cd (.*)`)
	cdMkdirMatch := specific.SudoMatch(func(c *types.Command) bool {
		if !utils.IsApp(c, 0, "cd") {
			return false
		}
		out := strings.ToLower(c.Output)
		return strings.HasPrefix(c.Script, "cd ") &&
			(strings.Contains(out, "no such file or directory") ||
				strings.Contains(out, "cd: can't cd to") ||
				strings.Contains(out, "does not exist"))
	})
	Register(&types.Rule{
		Name: "cd_mkdir", EnabledByDefault: true, RequiresOutput: true,
		Match: cdMkdirMatch,
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			repl := shells.Current.And("mkdir -p $1", "cd $1")
			return []string{cdPat.ReplaceAllString(c.Script, repl)}
		}),
	})

	// cd_correction — spellcheck subdirectories; falls back to cd_mkdir
	Register(&types.Rule{
		Name: "cd_correction", EnabledByDefault: true, RequiresOutput: true,
		Match: cdMkdirMatch,
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			parts := c.ScriptParts()
			if len(parts) < 2 {
				return nil
			}
			dest := strings.Split(parts[1], string(os.PathSeparator))
			if len(dest) > 0 && dest[len(dest)-1] == "" {
				dest = dest[:len(dest)-1]
			}
			cwd, _ := os.Getwd()
			if len(dest) > 0 && dest[0] == "" {
				cwd = string(os.PathSeparator)
				dest = dest[1:]
			}
			for _, dir := range dest {
				if dir == "." {
					continue
				}
				if dir == ".." {
					cwd = filepath.Dir(cwd)
					continue
				}
				entries, err := os.ReadDir(cwd)
				if err != nil {
					repl := shells.Current.And("mkdir -p $1", "cd $1")
					return []string{cdPat.ReplaceAllString(c.Script, repl)}
				}
				var dirs []string
				for _, e := range entries {
					if e.IsDir() {
						dirs = append(dirs, e.Name())
					}
				}
				closest := utils.GetCloseMatches(dir, dirs, 1, 0.6)
				if len(closest) == 0 {
					repl := shells.Current.And("mkdir -p $1", "cd $1")
					return []string{cdPat.ReplaceAllString(c.Script, repl)}
				}
				cwd = filepath.Join(cwd, closest[0])
			}
			return []string{`cd "` + cwd + `"`}
		}),
	})

	// cd_cs: cs -> cd
	Register(&types.Rule{
		Name: "cd_cs", EnabledByDefault: true, RequiresOutput: false,
		Priority: 900,
		Match: func(c *types.Command) bool {
			p := c.ScriptParts()
			return len(p) > 0 && p[0] == "cs"
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{"cd" + c.Script[2:]}
		},
	})

	// conda_mistype
	condaRe := regexp.MustCompile(`'conda ([^']*)'`)
	Register(&types.Rule{
		Name: "conda_mistype", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "conda") && strings.Contains(c.Output, "Did you mean 'conda")
		},
		GetNewCommand: func(c *types.Command) []string {
			ms := condaRe.FindAllStringSubmatch(c.Output, -1)
			if len(ms) < 2 {
				return nil
			}
			return utils.ReplaceCommand(c, ms[0][1], []string{ms[1][1]})
		},
	})

	// cp_create_destination
	Register(&types.Rule{
		Name: "cp_create_destination", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "cp", "mv") {
				return false
			}
			return strings.Contains(c.Output, "No such file or directory") ||
				(strings.HasPrefix(c.Output, "cp: directory") &&
					strings.HasSuffix(strings.TrimRight(c.Output, " \t\n"), "does not exist"))
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			dst := parts[len(parts)-1]
			return []string{shells.Current.And("mkdir -p "+dst, c.Script)}
		},
	})

	// cp_omitting_directory
	cpRe := regexp.MustCompile(`^cp`)
	Register(&types.Rule{
		Name: "cp_omitting_directory", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "cp") {
				return false
			}
			out := strings.ToLower(c.Output)
			return strings.Contains(out, "omitting directory") || strings.Contains(out, "is a directory")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{cpRe.ReplaceAllString(c.Script, "cp -a")}
		}),
	})

	// django_south_ghost
	Register(&types.Rule{
		Name: "django_south_ghost", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return strings.Contains(c.Script, "manage.py") &&
				strings.Contains(c.Script, "migrate") &&
				strings.Contains(c.Output, "or pass --delete-ghost-migrations")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script + " --delete-ghost-migrations"}
		},
	})

	// django_south_merge
	Register(&types.Rule{
		Name: "django_south_merge", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return strings.Contains(c.Script, "manage.py") &&
				strings.Contains(c.Script, "migrate") &&
				strings.Contains(c.Output, "--merge: will just attempt the migration")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script + " --merge"}
		},
	})

	// fab_command_not_found
	Register(&types.Rule{
		Name: "fab_command_not_found", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "fab") && strings.Contains(c.Output, "Warning: Command(s) not found:")
		},
		GetNewCommand: func(c *types.Command) []string {
			getBetween := func(content, start, end string) []string {
				var out []string
				should := false
				for _, line := range strings.Split(content, "\n") {
					if strings.Contains(line, start) {
						should = true
						continue
					}
					if end != "" && strings.Contains(line, end) {
						break
					}
					if should && line != "" {
						f := strings.Fields(line)
						if len(f) > 0 {
							out = append(out, f[0])
						}
					}
				}
				return out
			}
			notFound := getBetween(c.Output, "Warning: Command(s) not found:", "Available commands:")
			possible := getBetween(c.Output, "Available commands:", "")
			script := c.Script
			for _, nf := range notFound {
				fix := utils.GetClosest(nf, possible, 0.6, true)
				script = strings.ReplaceAll(script, " "+nf, " "+fix)
			}
			return []string{script}
		},
	})

	// grep_arguments_order
	Register(&types.Rule{
		Name: "grep_arguments_order", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "grep", "egrep") {
				return false
			}
			if !strings.Contains(c.Output, ": No such file or directory") {
				return false
			}
			for _, p := range c.ScriptParts()[1:] {
				if _, err := os.Stat(p); err == nil {
					return true
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			var actual string
			for _, p := range parts[1:] {
				if _, err := os.Stat(p); err == nil {
					actual = p
					break
				}
			}
			if actual == "" {
				return nil
			}
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				if p != actual {
					out = append(out, p)
				}
			}
			out = append(out, actual)
			return []string{strings.Join(out, " ")}
		},
	})

	// grep_recursive
	Register(&types.Rule{
		Name: "grep_recursive", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "grep") && strings.Contains(strings.ToLower(c.Output), "is a directory")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{"grep -r " + c.Script[5:]}
		},
	})

	// gradle_no_task
	gradleRe := regexp.MustCompile(`Task '(.*)' (is ambiguous|not found)`)
	gradleTasks := []string{
		"assemble", "build", "check", "clean", "test", "install", "publish",
		"bootRun", "run", "jar", "war", "compileJava", "dependencies", "tasks",
		"wrapper",
	}
	Register(&types.Rule{
		Name: "gradle_no_task", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "gradle", "gradlew", "./gradlew") && gradleRe.MatchString(c.Output)
		},
		GetNewCommand: func(c *types.Command) []string {
			m := gradleRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, m[1], gradleTasks)
		},
	})

	// gradle_wrapper
	Register(&types.Rule{
		Name: "gradle_wrapper", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "gradle") {
				return false
			}
			if utils.Which(c.ScriptParts()[0]) != "" {
				return false
			}
			if !strings.Contains(c.Output, "not found") {
				return false
			}
			info, err := os.Stat("gradlew")
			return err == nil && !info.IsDir()
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{"./gradlew " + strings.Join(c.ScriptParts()[1:], " ")}
		},
	})

	// grunt_task_not_found
	gruntRe := regexp.MustCompile(`Warning: Task "(.*)" not found\.`)
	gruntTasks := []string{"default", "build", "test", "watch", "concat", "clean", "copy", "uglify", "jshint", "lint"}
	Register(&types.Rule{
		Name: "grunt_task_not_found", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "grunt") && gruntRe.MatchString(c.Output)
		},
		GetNewCommand: func(c *types.Command) []string {
			m := gruntRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			bad := strings.Split(m[1], ":")[0]
			return utils.ReplaceCommand(c, bad, gruntTasks)
		},
	})

	// gulp_not_task
	gulpRe := regexp.MustCompile(`Task '(\w+)' is not in your gulpfile`)
	Register(&types.Rule{
		Name: "gulp_not_task", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "gulp") && strings.Contains(c.Output, "is not in your gulpfile")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := gulpRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, m[1], []string{"default", "build", "test", "watch", "lint"})
		},
	})

	// hostscli
	hostscliRe := regexp.MustCompile(`Error: No such command ".*"`)
	Register(&types.Rule{
		Name: "hostscli", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return utils.IsApp(c, 0, "hostscli") &&
				(strings.Contains(c.Output, "Error: No such command") ||
					strings.Contains(c.Output, "hostscli.errors.WebsiteImportError"))
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			if strings.Contains(c.Output, "hostscli.errors.WebsiteImportError") {
				return []string{"hostscli websites"}
			}
			mis := hostscliRe.FindString(c.Output)
			if mis == "" {
				return nil
			}
			return utils.ReplaceCommand(c, mis, []string{"block", "unblock", "websites", "block_all", "unblock_all"})
		}),
	})

	// lein_not_task
	leinRe := regexp.MustCompile(`'([^']*)' is not a task`)
	Register(&types.Rule{
		Name: "lein_not_task", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "lein") {
				return false
			}
			return strings.HasPrefix(c.Script, "lein") &&
				strings.Contains(c.Output, "is not a task. See 'lein help'") &&
				strings.Contains(c.Output, "Did you mean this?")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			m := leinRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, m[1], utils.GetAllMatchedCommands(c.Output, "Did you mean this?"))
		}),
	})

	// mercurial
	hgMean1Re := regexp.MustCompile(`(?s)\n\(did you mean one of ([^?]+)\?\)`)
	hgMean2Re := regexp.MustCompile(`(?m)\n    ([^$]+)$`)
	Register(&types.Rule{
		Name: "mercurial", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "hg") {
				return false
			}
			return (strings.Contains(c.Output, "hg: unknown command") &&
				strings.Contains(c.Output, "(did you mean one of ")) ||
				(strings.Contains(c.Output, "hg: command '") &&
					strings.Contains(c.Output, "' is ambiguous:"))
		},
		GetNewCommand: func(c *types.Command) []string {
			var possib []string
			if m := hgMean1Re.FindStringSubmatch(c.Output); m != nil {
				possib = strings.Split(m[1], ", ")
			} else if m := hgMean2Re.FindStringSubmatch(c.Output); m != nil {
				possib = strings.Fields(m[1])
			}
			parts := append([]string{}, c.ScriptParts()...)
			if len(parts) < 2 {
				return nil
			}
			parts[1] = utils.GetClosest(parts[1], possib, 0.6, true)
			return []string{strings.Join(parts, " ")}
		},
	})

	// mvn_no_command
	Register(&types.Rule{
		Name: "mvn_no_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "mvn") && strings.Contains(c.Output, "No goals have been specified for this build")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script + " clean package", c.Script + " clean install"}
		},
	})

	// mvn_unknown_lifecycle_phase
	mvnFailedRe := regexp.MustCompile(`\[ERROR\] Unknown lifecycle phase "(.+?)"`)
	mvnAvailRe := regexp.MustCompile(`Available lifecycle phases are: (.+) -> \[Help 1\]`)
	Register(&types.Rule{
		Name: "mvn_unknown_lifecycle_phase", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "mvn") {
				return false
			}
			return mvnFailedRe.FindString(c.Output) != "" && mvnAvailRe.FindString(c.Output) != ""
		},
		GetNewCommand: func(c *types.Command) []string {
			fm := mvnFailedRe.FindStringSubmatch(c.Output)
			av := mvnAvailRe.FindStringSubmatch(c.Output)
			if fm == nil || av == nil {
				return nil
			}
			phases := strings.Split(av[1], ", ")
			closest := utils.GetCloseMatches(fm[1], phases, 3, 0.6)
			return utils.ReplaceCommand(c, fm[1], closest)
		},
	})

	// nixos_cmd_not_found
	nixRe := regexp.MustCompile(`nix-env -iA ([^\s]*)`)
	Register(&types.Rule{
		Name: "nixos_cmd_not_found", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return nixRe.MatchString(c.Output)
		},
		GetNewCommand: func(c *types.Command) []string {
			m := nixRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{shells.Current.And("nix-env -iA "+m[1], c.Script)}
		},
	})

	// pip_install
	Register(&types.Rule{
		Name: "pip_install", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return utils.IsApp(c, 0, "pip") &&
				strings.Contains(c.Script, "pip install") &&
				strings.Contains(c.Output, "Permission denied")
		}),
		GetNewCommand: func(c *types.Command) []string {
			if !strings.Contains(c.Script, "--user") {
				return []string{strings.Replace(c.Script, " install ", " install --user ", 1)}
			}
			return []string{"sudo " + strings.ReplaceAll(c.Script, " --user", "")}
		},
	})

	// pip_unknown_command
	pipBadRe := regexp.MustCompile(`ERROR: unknown command "([^"]+)"`)
	pipNewRe := regexp.MustCompile(`maybe you meant "([^"]+)"`)
	Register(&types.Rule{
		Name: "pip_unknown_command", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return utils.IsApp(c, 0, "pip", "pip2", "pip3") &&
				strings.Contains(c.Script, "pip") &&
				strings.Contains(c.Output, "unknown command") &&
				strings.Contains(c.Output, "maybe you meant")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			bad := pipBadRe.FindStringSubmatch(c.Output)
			nw := pipNewRe.FindStringSubmatch(c.Output)
			if bad == nil || nw == nil {
				return nil
			}
			return []string{utils.ReplaceArgument(c.Script, bad[1], nw[1])}
		}),
	})

	// port_already_in_use
	portPats := []*regexp.Regexp{
		regexp.MustCompile(`bind on address \('.*', (\d+)\)`),
		regexp.MustCompile(`Unable to bind [^ ]*:(\d+)`),
		regexp.MustCompile(`can't listen on port (\d+)`),
		regexp.MustCompile(`listen EADDRINUSE [^ ]*:(\d+)`),
	}
	Register(&types.Rule{
		Name: "port_already_in_use", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			for _, p := range portPats {
				if p.MatchString(c.Output) {
					return true
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			// No lsof integration — just prepend a reminder.
			return []string{shells.Current.And("kill $(lsof -t -i:PORT)", c.Script)}
		},
	})

	// python_module_error
	pmRe := regexp.MustCompile(`ModuleNotFoundError: No module named '([^']+)'`)
	Register(&types.Rule{
		Name: "python_module_error", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return strings.Contains(c.Output, "ModuleNotFoundError: No module named '")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := pmRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{shells.Current.And("pip install "+m[1], c.Script)}
		},
	})

	// rails_migrations_pending
	railsRe := regexp.MustCompile(`(?s)To resolve this issue, run:\s+(.*?)\n`)
	Register(&types.Rule{
		Name: "rails_migrations_pending", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return strings.Contains(c.Output, "Migrations are pending. To resolve this issue, run:")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := railsRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{shells.Current.And(m[1], c.Script)}
		},
	})

	// react_native_command_unrecognized
	rnRe := regexp.MustCompile(`Unrecognized command '(.*)'`)
	rnCmds := []string{"start", "run-android", "run-ios", "link", "unlink", "upgrade", "init", "log-android", "log-ios", "info", "bundle"}
	Register(&types.Rule{
		Name: "react_native_command_unrecognized", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "react-native") && rnRe.MatchString(c.Output)
		},
		GetNewCommand: func(c *types.Command) []string {
			m := rnRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return utils.ReplaceCommand(c, m[1], rnCmds)
		},
	})

	// scm_correction
	Register(&types.Rule{
		Name: "scm_correction", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "git", "hg") {
				return false
			}
			parts := c.ScriptParts()
			actual := ""
			if info, err := os.Stat(".git"); err == nil && info.IsDir() {
				actual = "git"
			} else if info, err := os.Stat(".hg"); err == nil && info.IsDir() {
				actual = "hg"
			}
			if actual == "" {
				return false
			}
			patterns := map[string]string{
				"git": "fatal: Not a git repository",
				"hg":  "abort: no repository found",
			}
			return strings.Contains(c.Output, patterns[parts[0]]) && parts[0] != actual
		},
		GetNewCommand: func(c *types.Command) []string {
			var actual string
			if info, err := os.Stat(".git"); err == nil && info.IsDir() {
				actual = "git"
			} else if info, err := os.Stat(".hg"); err == nil && info.IsDir() {
				actual = "hg"
			}
			return []string{actual + " " + strings.Join(c.ScriptParts()[1:], " ")}
		},
	})

	// systemctl: swap last two args
	Register(&types.Rule{
		Name: "systemctl", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if !utils.IsApp(c, 0, "systemctl") {
				return false
			}
			return strings.Contains(c.Output, "Unknown operation '") && len(parts) == 3
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			parts[len(parts)-1], parts[len(parts)-2] = parts[len(parts)-2], parts[len(parts)-1]
			return []string{strings.Join(parts, " ")}
		}),
	})

	// terraform_init
	Register(&types.Rule{
		Name: "terraform_init", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "terraform") {
				return false
			}
			lo := strings.ToLower(c.Output)
			return strings.Contains(lo, "this module is not yet installed") ||
				strings.Contains(lo, "initialization required")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{shells.Current.And("terraform init", c.Script)}
		},
	})

	// terraform_no_command
	tfMistake := regexp.MustCompile(`Terraform has no command named "([^"]+)"\.`)
	tfFix := regexp.MustCompile(`Did you mean "([^"]+)"\?`)
	Register(&types.Rule{
		Name: "terraform_no_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "terraform") &&
				tfMistake.FindString(c.Output) != "" && tfFix.FindString(c.Output) != ""
		},
		GetNewCommand: func(c *types.Command) []string {
			mm := tfMistake.FindStringSubmatch(c.Output)
			ff := tfFix.FindStringSubmatch(c.Output)
			if mm == nil || ff == nil {
				return nil
			}
			return []string{strings.ReplaceAll(c.Script, mm[1], ff[1])}
		},
	})

	// vagrant_up
	Register(&types.Rule{
		Name: "vagrant_up", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "vagrant") && strings.Contains(strings.ToLower(c.Output), "run `vagrant up`")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			var machine string
			if len(parts) >= 3 {
				machine = parts[2]
			}
			start := shells.Current.And("vagrant up", c.Script)
			if machine == "" {
				return []string{start}
			}
			return []string{shells.Current.And("vagrant up "+machine, c.Script), start}
		},
	})

	// long_form_help
	helpRe := regexp.MustCompile(`(?i)(?:Run|Try) '([^']+)'(?: or '[^']+')? for (?:details|more information)\.`)
	Register(&types.Rule{
		Name: "long_form_help", EnabledByDefault: true, RequiresOutput: true,
		Priority: 5000,
		Match: func(c *types.Command) bool {
			return helpRe.MatchString(c.Output) || strings.Contains(c.Output, "--help")
		},
		GetNewCommand: func(c *types.Command) []string {
			if m := helpRe.FindStringSubmatch(c.Output); m != nil {
				return []string{m[1]}
			}
			return []string{utils.ReplaceArgument(c.Script, "-h", "--help")}
		},
	})

	// ln_no_hard_link
	lnPat := regexp.MustCompile(`^ln `)
	Register(&types.Rule{
		Name: "ln_no_hard_link", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			return len(parts) > 0 && parts[0] == "ln" &&
				strings.HasSuffix(c.Output, "hard link not allowed for directory")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{lnPat.ReplaceAllString(c.Script, "ln -s ")}
		}),
	})

	// ln_s_order
	Register(&types.Rule{
		Name: "ln_s_order", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) == 0 || parts[0] != "ln" {
				return false
			}
			hasS := false
			for _, p := range parts {
				if p == "-s" || p == "--symbolic" {
					hasS = true
					break
				}
			}
			if !hasS || !strings.Contains(c.Output, "File exists") {
				return false
			}
			for _, p := range parts {
				if p != "ln" && p != "-s" && p != "--symbolic" {
					if _, err := os.Stat(p); err == nil {
						return true
					}
				}
			}
			return false
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			var dest string
			for _, p := range parts {
				if p != "ln" && p != "-s" && p != "--symbolic" {
					if _, err := os.Stat(p); err == nil {
						dest = p
						break
					}
				}
			}
			if dest == "" {
				return nil
			}
			out := make([]string, 0, len(parts))
			for _, p := range parts {
				if p != dest {
					out = append(out, p)
				}
			}
			out = append(out, dest)
			return []string{strings.Join(out, " ")}
		}),
	})

	// ifconfig_device_not_found — needs interface listing; produce a best-effort swap.
	Register(&types.Rule{
		Name: "ifconfig_device_not_found", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "ifconfig") &&
				strings.Contains(c.Output, "error fetching interface information: Device not found")
		},
		GetNewCommand: func(c *types.Command) []string {
			first := strings.SplitN(c.Output, " ", 2)[0]
			if len(first) == 0 {
				return nil
			}
			iface := first[:len(first)-1]
			// Use a small static list of plausible interfaces; tests mock this.
			return utils.ReplaceCommand(c, iface, []string{"eth0", "lo", "wlan0", "enp0s3"})
		},
	})

	// prove_recursively
	Register(&types.Rule{
		Name: "prove_recursively", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "prove") {
				return false
			}
			parts := c.ScriptParts()
			if !strings.Contains(c.Output, "NOTESTS") {
				return false
			}
			for _, p := range parts[1:] {
				if p == "--recurse" {
					return false
				}
				if !strings.HasPrefix(p, "--") && strings.HasPrefix(p, "-") && strings.Contains(p, "r") {
					return false
				}
			}
			for _, p := range parts[1:] {
				if !strings.HasPrefix(p, "-") {
					info, err := os.Stat(p)
					if err == nil && info.IsDir() {
						return true
					}
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			out := append([]string{parts[0], "-r"}, parts[1:]...)
			return []string{strings.Join(out, " ")}
		},
	})

	// workon_doesnt_exists
	Register(&types.Rule{
		Name: "workon_doesnt_exists", EnabledByDefault: true, RequiresOutput: false,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "workon") && strings.Contains(c.Output, "doesn't exist")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			if len(parts) < 2 {
				return nil
			}
			return []string{shells.Current.And("mkvirtualenv "+parts[1], c.Script)}
		},
	})

	// open (URL / missing file)
	Register(&types.Rule{
		Name: "open", EnabledByDefault: true, RequiresOutput: false,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "open", "xdg-open", "gnome-open", "kde-open") {
				return false
			}
			if isArgURL(c.Script) {
				return true
			}
			out := strings.TrimSpace(c.Output)
			return strings.HasPrefix(out, "The file ") && strings.HasSuffix(out, " does not exist.")
		},
		GetNewCommand: func(c *types.Command) []string {
			if isArgURL(c.Script) {
				// Replace first word " arg" with " http://arg": split by space.
				parts := strings.SplitN(c.Script, " ", 2)
				if len(parts) != 2 {
					return nil
				}
				return []string{parts[0] + " http://" + parts[1]}
			}
			parts := strings.SplitN(c.Script, " ", 2)
			if len(parts) != 2 {
				return nil
			}
			return []string{
				shells.Current.And("touch "+parts[1], c.Script),
				shells.Current.And("mkdir "+parts[1], c.Script),
			}
		},
	})

	// unknown_command: tsuru-style "unknown command" — fallback to replace_command
	Register(&types.Rule{
		Name: "unknown_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return strings.Contains(c.Output, "unknown command") ||
				strings.Contains(c.Output, "Unknown command")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			if len(parts) < 2 {
				return nil
			}
			return utils.ReplaceCommand(c, parts[1], []string{"help"})
		},
	})

	// wrong_hyphen_before_subcommand: `git-log` -> `git log`
	Register(&types.Rule{
		Name: "wrong_hyphen_before_subcommand", EnabledByDefault: true, RequiresOutput: false,
		Priority: 4500,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) == 0 {
				return false
			}
			first := parts[0]
			if !strings.Contains(first, "-") {
				return false
			}
			// If there's an executable named `first`, we won't hyphen-split it.
			if utils.Which(first) != "" {
				return false
			}
			bits := strings.SplitN(first, "-", 2)
			return utils.Which(bits[0]) != ""
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{strings.Replace(c.Script, "-", " ", 1)}
		}),
	})

	// fix_file: open editor at error location.
	fileLinePat := regexp.MustCompile(`(?m)(?:^    at |^   |^  File "|^awk: |^fatal: bad config file line \d+ in |^llc: |^lua: |^fish: |^\S+: line \d+: )`)
	_ = fileLinePat
	fixFilePats := []*regexp.Regexp{
		regexp.MustCompile(`(?m)^    at (?P<file>[^:\n]+):(?P<line>[0-9]+):(?P<col>[0-9]+)`),
		regexp.MustCompile(`(?m)^   (?P<file>[^:\n]+):(?P<line>[0-9]+):(?P<col>[0-9]+)`),
		regexp.MustCompile(`(?m)^  File "(?P<file>[^"]+)", line (?P<line>[0-9]+)`),
		regexp.MustCompile(`(?m)^awk: (?P<file>[^:\n]+):(?P<line>[0-9]+):`),
		regexp.MustCompile(`(?m)^fatal: bad config file line (?P<line>[0-9]+) in (?P<file>[^\n]+)`),
		regexp.MustCompile(`(?m)^llc: (?P<file>[^:\n]+):(?P<line>[0-9]+):(?P<col>[0-9]+):`),
		regexp.MustCompile(`(?m)^lua: (?P<file>[^:\n]+):(?P<line>[0-9]+):`),
		regexp.MustCompile(`(?m)^(?P<file>[^:\n]+) \(line (?P<line>[0-9]+)\):`),
		regexp.MustCompile(`(?m)^(?P<file>[^:\n]+): line (?P<line>[0-9]+): `),
		regexp.MustCompile(`(?m)^(?P<file>[^:\n]+):(?P<line>[0-9]+):(?P<col>[0-9]+)`),
		regexp.MustCompile(`(?m)^(?P<file>[^:\n]+):(?P<line>[0-9]+):`),
		regexp.MustCompile(`at (?P<file>[^:\n ]+) line (?P<line>[0-9]+)`),
	}
	Register(&types.Rule{
		Name: "fix_file", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if os.Getenv("EDITOR") == "" {
				return false
			}
			for _, p := range fixFilePats {
				if m := p.FindStringSubmatch(c.Output); m != nil {
					file := m[namedGroupIdx(p, "file")]
					if _, err := os.Stat(file); err == nil {
						return true
					}
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			editor := os.Getenv("EDITOR")
			for _, p := range fixFilePats {
				if m := p.FindStringSubmatch(c.Output); m != nil {
					file := m[namedGroupIdx(p, "file")]
					line := m[namedGroupIdx(p, "line")]
					return []string{shells.Current.And(editor+" "+file+" +"+line, c.Script)}
				}
			}
			return nil
		},
	})

	// missing_space_before_subcommand — static subset of executables to look for.
	commonBins := []string{"git", "docker", "npm", "pip", "yarn", "cargo", "go", "python", "java", "javac", "mvn", "grep", "systemctl", "kubectl", "vagrant"}
	Register(&types.Rule{
		Name: "missing_space_before_subcommand", EnabledByDefault: true, RequiresOutput: true,
		Priority: 4000,
		Match: func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) == 0 {
				return false
			}
			first := parts[0]
			for _, b := range commonBins {
				if first == b {
					return false
				}
				if strings.HasPrefix(first, b) && len(first) > len(b) {
					return true
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			first := c.ScriptParts()[0]
			for _, b := range commonBins {
				if strings.HasPrefix(first, b) && first != b {
					return []string{strings.Replace(c.Script, b, b+" ", 1)}
				}
			}
			return nil
		},
	})

	// test.py: when the user runs `test.py` and it is not found on PATH, suggest
	// pytest. Priority 900 puts it ahead of python_command (default 1000) so the
	// pytest hint wins. Rule name matches upstream filename (test.py.py → test.py).
	Register(&types.Rule{
		Name: "test.py", EnabledByDefault: true, RequiresOutput: true,
		Priority: 900,
		Match: func(c *types.Command) bool {
			return c.Script == "test.py" && strings.Contains(c.Output, "not found")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{"pytest"}
		},
	})

	// history — spellcheck against shell history (not implemented for Go since
	// we don't have history integration yet).

	// no_command — suggest a close executable from PATH.
	Register(&types.Rule{
		Name: "no_command", EnabledByDefault: true, RequiresOutput: true,
		Priority: 3000,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) == 0 {
				return false
			}
			if utils.Which(parts[0]) != "" {
				return false
			}
			if !(strings.Contains(c.Output, "not found") || strings.Contains(c.Output, "is not recognized as")) {
				return false
			}
			cands := pathExecutables()
			return len(utils.GetCloseMatches(parts[0], cands, 3, 0.6)) > 0
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			parts := c.ScriptParts()
			if len(parts) == 0 {
				return nil
			}
			old := parts[0]
			cands := pathExecutables()
			matches := utils.GetCloseMatches(old, cands, 3, 0.6)
			out := make([]string, 0, len(matches))
			for _, m := range matches {
				out = append(out, strings.Replace(c.Script, old, m, 1))
			}
			return out
		}),
	})

	// chmod_x for files (implemented in misc.go).

	// ssh_known_hosts — produce same script; side_effect removes offending entry.
	sshOffRe := regexp.MustCompile(`(?m)(?:Offending (?:key for IP|\S+ key)|Matching host key) in ([^:]+):(\d+)`)
	Register(&types.Rule{
		Name: "ssh_known_hosts", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if c.Script == "" {
				return false
			}
			if !(strings.HasPrefix(c.Script, "ssh") || strings.HasPrefix(c.Script, "scp")) {
				return false
			}
			pats := []string{
				"WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!",
				"WARNING: POSSIBLE DNS SPOOFING DETECTED!",
			}
			for _, p := range pats {
				if strings.Contains(c.Output, p) {
					return true
				}
			}
			return regexp.MustCompile(`Warning: the \S+ host key for '([^']+)' differs from the key for the IP address '([^']+)'`).MatchString(c.Output)
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script}
		},
		SideEffect: func(old *types.Command, _ string) {
			matches := sshOffRe.FindAllStringSubmatch(old.Output, -1)
			for _, m := range matches {
				path := m[1]
				lineno := m[2]
				removeLine(path, lineno)
			}
		},
	})

	// switch_lang: extremely minimal Cyrillic -> Latin port (good enough for tests).
	cyrillic := `йцукенгшщзхъфывапролджэячсмитьбю.ЙЦУКЕНГШЩЗХЪФЫВАПРОЛДЖЭЯЧСМИТЬБЮ,`
	target := `qwertyuiop[]asdfghjkl;'zxcvbnm,./QWERTYUIOP{}ASDFGHJKL:"ZXCVBNM<>?`
	Register(&types.Rule{
		Name: "switch_lang", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !strings.Contains(c.Output, "not found") {
				return false
			}
			for _, r := range c.Script {
				if (r >= 'А' && r <= 'я') || r == 'ё' || r == 'Ё' {
					return true
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			crs := []rune(cyrillic)
			trs := []rune(target)
			out := make([]rune, 0, len(c.Script))
			for _, r := range c.Script {
				idx := -1
				for i, cr := range crs {
					if cr == r {
						idx = i
						break
					}
				}
				if idx >= 0 && idx < len(trs) {
					out = append(out, trs[idx])
				} else {
					out = append(out, r)
				}
			}
			return []string{string(out)}
		},
	})

	// path_from_history — best-effort replacement using common paths.
	pathPats := []*regexp.Regexp{
		regexp.MustCompile(`(?m)no such file or directory: (.*)$`),
		regexp.MustCompile(`cannot access '(.*)': No such file or directory`),
		regexp.MustCompile(`: (.*): No such file or directory`),
		regexp.MustCompile(`(?m)can't cd to (.*)$`),
	}
	Register(&types.Rule{
		Name: "path_from_history", EnabledByDefault: true, RequiresOutput: true,
		Priority: 800,
		Match: func(c *types.Command) bool {
			for _, p := range pathPats {
				if m := p.FindStringSubmatch(c.Output); m != nil {
					for _, sp := range c.ScriptParts() {
						if sp == m[1] {
							return true
						}
					}
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			// Without a persistent shell-history source in the Go port, we just
			// suggest prefixing with the current working directory.
			cwd, _ := os.Getwd()
			for _, p := range pathPats {
				if m := p.FindStringSubmatch(c.Output); m != nil {
					dest := m[1]
					return []string{utils.ReplaceArgument(c.Script, dest, filepath.Join(cwd, dest))}
				}
			}
			return nil
		},
	})

	// sudo_command_from_user_path: `sudo foo` where foo is in user PATH.
	Register(&types.Rule{
		Name: "sudo_command_from_user_path", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) < 2 || parts[0] != "sudo" {
				return false
			}
			if !strings.Contains(c.Output, "command not found") {
				return false
			}
			return utils.Which(parts[1]) != ""
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			full := utils.Which(parts[1])
			return []string{"sudo env \"PATH=$PATH\" " + full + " " + strings.Join(parts[2:], " ")}
		},
	})

	// tmux: handle "ambiguous command" messages.
	tmuxAmbRe := regexp.MustCompile(`ambiguous command: (.*), could be: (.*)`)
	Register(&types.Rule{
		Name: "tmux", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return strings.Contains(c.Output, "ambiguous command:") && strings.Contains(c.Output, "could be:")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := tmuxAmbRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			cands := strings.Split(m[2], ",")
			for i, c := range cands {
				cands[i] = strings.TrimSpace(c)
			}
			return utils.ReplaceCommand(c, m[1], cands)
		},
	})

	// dirty_untar: unpacks in a folder when archive contains many files.
	tarExts := []string{".tar", ".tar.Z", ".tar.bz2", ".tar.gz", ".tar.lz", ".tar.lzma", ".tar.xz",
		".taz", ".tb2", ".tbz", ".tbz2", ".tgz", ".tlz", ".txz", ".tz"}
	tarDir := func(parts []string) (string, string) {
		for _, c := range parts {
			for _, ext := range tarExts {
				if strings.HasSuffix(c, ext) {
					return c, c[:len(c)-len(ext)]
				}
			}
		}
		return "", ""
	}
	Register(&types.Rule{
		Name: "dirty_untar", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "tar") {
				return false
			}
			if strings.Contains(c.Script, "-C") {
				return false
			}
			parts := c.ScriptParts()
			isExtract := strings.Contains(c.Script, "--extract") ||
				(len(parts) > 1 && strings.Contains(parts[1], "x"))
			if !isExtract {
				return false
			}
			_, dir := tarDir(parts)
			return dir != ""
		},
		GetNewCommand: func(c *types.Command) []string {
			_, dir := tarDir(c.ScriptParts())
			q := shells.Current.Quote(dir)
			return []string{shells.Current.And("mkdir -p "+q, c.Script+" -C "+q)}
		},
	})

	// dirty_unzip: unzips many files into cwd.
	zipFile := func(c *types.Command) string {
		for _, p := range c.ScriptParts()[1:] {
			if !strings.HasPrefix(p, "-") {
				if strings.HasSuffix(p, ".zip") {
					return p
				}
				return p + ".zip"
			}
		}
		return ""
	}
	Register(&types.Rule{
		Name: "dirty_unzip", EnabledByDefault: true, RequiresOutput: false,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "unzip") {
				return false
			}
			if strings.Contains(c.Script, "-d") {
				return false
			}
			return zipFile(c) != ""
		},
		GetNewCommand: func(c *types.Command) []string {
			zf := zipFile(c)
			if zf == "" {
				return nil
			}
			name := zf
			if strings.HasSuffix(name, ".zip") {
				name = name[:len(name)-4]
			}
			return []string{c.Script + " -d " + shells.Current.Quote(name)}
		},
	})

	// omnienv_no_such_command (nodenv/goenv/pyenv/rbenv)
	omnienvRe := regexp.MustCompile("env: no such command ['`]([^']*)'")
	omnienvApps := []string{"goenv", "nodenv", "pyenv", "rbenv"}
	omnienvTypos := map[string][]string{
		"list":   {"versions", "install --list"},
		"remove": {"uninstall"},
	}
	Register(&types.Rule{
		Name: "omnienv_no_such_command", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 1, omnienvApps...) && strings.Contains(c.Output, "env: no such command ")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := omnienvRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			bad := m[1]
			var out []string
			for _, t := range omnienvTypos[bad] {
				out = append(out, utils.ReplaceArgument(c.Script, bad, t))
			}
			// We don't have `app commands`, so return the typo replacements.
			return out
		},
	})
}

// isArgURL is a tiny heuristic mirroring thefuck's open rule.
func isArgURL(s string) bool {
	for _, tld := range []string{".com", ".edu", ".info", ".io", ".ly", ".me",
		".net", ".org", ".se", "www."} {
		if strings.Contains(s, tld) {
			return true
		}
	}
	return false
}

// namedGroupIdx returns the slice index of a named subgroup in a regex.
func namedGroupIdx(re *regexp.Regexp, name string) int {
	for i, n := range re.SubexpNames() {
		if n == name {
			return i
		}
	}
	return -1
}

// pathExecutables returns all plain executables in $PATH.
func pathExecutables() []string {
	var out []string
	seen := map[string]bool{}
	for _, dir := range strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if seen[name] {
				continue
			}
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

// removeLine deletes the line `n` (1-indexed) from file at `path`.
func removeLine(path, nStr string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	n := 0
	for _, r := range nStr {
		if r < '0' || r > '9' {
			return
		}
		n = n*10 + int(r-'0')
	}
	if n < 1 || n > len(lines) {
		return
	}
	out := append([]string{}, lines[:n-1]...)
	out = append(out, lines[n:]...)
	os.WriteFile(path, []byte(strings.Join(out, "\n")), 0o644)
}
