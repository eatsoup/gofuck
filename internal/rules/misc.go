package rules

import (
	"os"
	"regexp"
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/specific"
	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

func init() {
	// cd_parent: `cd..` -> `cd ..`
	Register(&types.Rule{
		Name:             "cd_parent",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return c.Script == "cd.."
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{"cd .."}
		},
	})

	// dry: `git git log` -> `git log`
	Register(&types.Rule{
		Name:             "dry",
		Priority:         900,
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			sp := c.ScriptParts()
			return len(sp) >= 2 && sp[0] == sp[1]
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.Join(c.ScriptParts()[1:], " ")}
		},
	})

	// sl_ls: `sl` -> `ls`
	Register(&types.Rule{
		Name:             "sl_ls",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match:            func(c *types.Command) bool { return c.Script == "sl" },
		GetNewCommand:    func(c *types.Command) []string { return []string{"ls"} },
	})

	// quotation_marks: ' and " mixed
	Register(&types.Rule{
		Name:             "quotation_marks",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return strings.Contains(c.Script, "'") && strings.Contains(c.Script, `"`)
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.ReplaceAll(c.Script, "'", `"`)}
		},
	})

	// remove_trailing_cedilla
	Register(&types.Rule{
		Name:             "remove_trailing_cedilla",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return strings.HasSuffix(c.Script, "ç")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.TrimSuffix(c.Script, "ç")}
		},
	})

	// remove_shell_prompt_literal
	rspPat := regexp.MustCompile(`^\s*\$ \S+`)
	Register(&types.Rule{
		Name:             "remove_shell_prompt_literal",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return strings.Contains(c.Output, "$: command not found") && rspPat.MatchString(c.Script)
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.TrimLeft(c.Script, "$ ")}
		},
	})

	// fix_alt_space (NBSP -> regular space)
	Register(&types.Rule{
		Name:             "fix_alt_space",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return strings.Contains(strings.ToLower(c.Output), "command not found") &&
				strings.Contains(c.Script, " ")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{strings.ReplaceAll(c.Script, " ", " ")}
		}),
	})

	// sudo: prepends sudo when output indicates permission denied etc.
	sudoPatterns := []string{
		"permission denied",
		"eacces",
		"pkg: insufficient privileges",
		"you cannot perform this operation unless you are root",
		"non-root users cannot",
		"operation not permitted",
		"not super-user",
		"superuser privilege",
		"root privilege",
		"this command has to be run under the root user.",
		"this operation requires root.",
		"requested operation requires superuser privilege",
		"must be run as root",
		"must run as root",
		"must be superuser",
		"must be root",
		"need to be root",
		"need root",
		"needs to be run as root",
		"only root can ",
		"you don't have access to the history db.",
		"authentication is required",
		"edspermissionerror",
		"you don't have write permissions",
		"use `sudo`",
		"sudorequirederror",
		"error: insufficient privileges",
		"updatedb: can not open a temporary file",
	}
	Register(&types.Rule{
		Name:             "sudo",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			parts := c.ScriptParts()
			hasAmp := false
			for _, p := range parts {
				if p == "&&" {
					hasAmp = true
					break
				}
			}
			if len(parts) > 0 && !hasAmp && parts[0] == "sudo" {
				return false
			}
			lo := strings.ToLower(c.Output)
			for _, p := range sudoPatterns {
				if strings.Contains(lo, p) {
					return true
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			if strings.Contains(c.Script, "&&") {
				parts := c.ScriptParts()
				filtered := make([]string, 0, len(parts))
				for _, p := range parts {
					if p != "sudo" {
						filtered = append(filtered, p)
					}
				}
				return []string{`sudo sh -c "` + strings.Join(filtered, " ") + `"`}
			}
			if strings.Contains(c.Script, ">") {
				return []string{`sudo sh -c "` + strings.ReplaceAll(c.Script, `"`, `\"`) + `"`}
			}
			return []string{"sudo " + c.Script}
		},
	})

	// unsudo: runs non-sudo command when root isn't allowed
	Register(&types.Rule{
		Name:             "unsudo",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			sp := c.ScriptParts()
			if len(sp) == 0 || sp[0] != "sudo" {
				return false
			}
			return strings.Contains(strings.ToLower(c.Output), "you cannot perform this operation as root")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.TrimSpace(c.Script[len("sudo"):])}
		},
	})

	// ls_lah: adds -lah when no flag given
	Register(&types.Rule{
		Name:             "ls_lah",
		EnabledByDefault: true,
		RequiresOutput:   false,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "ls") {
				return false
			}
			return !strings.Contains(c.Script, "ls -")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			parts[0] = "ls -lah"
			return []string{strings.Join(parts, " ")}
		},
	})

	// ls_all: shows hidden files when output empty
	Register(&types.Rule{
		Name:             "ls_all",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "ls") && strings.TrimSpace(c.Output) == ""
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := append([]string{"ls", "-A"}, c.ScriptParts()[1:]...)
			return []string{strings.Join(parts, " ")}
		},
	})

	// mkdir_p: `mkdir a/b` -> `mkdir -p a/b`
	mkdirPat := regexp.MustCompile(`\bmkdir (.*)`)
	Register(&types.Rule{
		Name:             "mkdir_p",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "mkdir") &&
				strings.Contains(c.Output, "No such file or directory")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{mkdirPat.ReplaceAllString(c.Script, "mkdir -p $1")}
		}),
	})

	// rm_dir
	rmPat := regexp.MustCompile(`\brm (.*)`)
	Register(&types.Rule{
		Name:             "rm_dir",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "rm") &&
				strings.Contains(strings.ToLower(c.Output), "is a directory")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			args := "-rf"
			if strings.Contains(c.Script, "hdfs") {
				args = "-r"
			}
			return []string{rmPat.ReplaceAllString(c.Script, "rm "+args+" $1")}
		}),
	})

	// rm_root: must be explicitly enabled
	Register(&types.Rule{
		Name:             "rm_root",
		EnabledByDefault: false,
		RequiresOutput:   true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			sp := c.ScriptParts()
			hasRm, hasSlash := false, false
			for _, p := range sp {
				if p == "rm" {
					hasRm = true
				}
				if p == "/" {
					hasSlash = true
				}
			}
			return hasRm && hasSlash &&
				!strings.Contains(c.Script, "--no-preserve-root") &&
				strings.Contains(c.Output, "--no-preserve-root")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{c.Script + " --no-preserve-root"}
		}),
	})

	// chmod_x
	Register(&types.Rule{
		Name:             "chmod_x",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			if !strings.HasPrefix(c.Script, "./") {
				return false
			}
			if !strings.Contains(strings.ToLower(c.Output), "permission denied") {
				return false
			}
			parts := c.ScriptParts()
			if len(parts) == 0 {
				return false
			}
			info, err := os.Stat(parts[0])
			if err != nil || info.IsDir() {
				return false
			}
			return info.Mode()&0o111 == 0
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			return []string{shells.Current.And(
				"chmod +x "+parts[0][2:],
				c.Script,
			)}
		},
	})

	// has_exists_script (forgot to prefix ./)
	Register(&types.Rule{
		Name:             "has_exists_script",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) == 0 {
				return false
			}
			if _, err := os.Stat(parts[0]); err != nil {
				return false
			}
			return strings.Contains(c.Output, "command not found")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{"./" + c.Script}
		}),
	})

	// touch: `touch a/b/c` when parent missing -> mkdir -p parent && touch ...
	touchPat := regexp.MustCompile(`touch: (?:cannot touch ')?(.+)/.+?'?:`)
	Register(&types.Rule{
		Name:             "touch",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "touch") && strings.Contains(c.Output, "No such file or directory")
		},
		GetNewCommand: func(c *types.Command) []string {
			m := touchPat.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{shells.Current.And("mkdir -p "+m[1], c.Script)}
		},
	})

	// whois: strip URL scheme/subdomains
	Register(&types.Rule{
		Name:             "whois",
		EnabledByDefault: true,
		RequiresOutput:   false,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 1, "whois")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			url := parts[1]
			if strings.Contains(c.Script, "/") {
				rest := url
				if i := strings.Index(rest, "://"); i >= 0 {
					rest = rest[i+3:]
				}
				if i := strings.IndexByte(rest, '/'); i >= 0 {
					rest = rest[:i]
				}
				return []string{"whois " + rest}
			}
			if strings.Contains(c.Script, ".") {
				path := strings.Split(url, ".")
				var out []string
				for n := 1; n < len(path); n++ {
					out = append(out, "whois "+strings.Join(path[n:], "."))
				}
				return out
			}
			return nil
		},
	})

	// man_no_space
	Register(&types.Rule{
		Name:             "man_no_space",
		Priority:         2000,
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return strings.HasPrefix(c.Script, "man") &&
				strings.Contains(strings.ToLower(c.Output), "command not found")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{"man " + c.Script[3:]}
		},
	})

	// cat_dir: cat on a directory
	Register(&types.Rule{
		Name:             "cat_dir",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 1, "cat") {
				return false
			}
			if !strings.HasPrefix(c.Output, "cat: ") {
				return false
			}
			parts := c.ScriptParts()
			info, err := os.Stat(parts[1])
			return err == nil && info.IsDir()
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.Replace(c.Script, "cat", "ls", 1)}
		},
	})

	// cpp11: add -std=c++11 for g++/clang++
	Register(&types.Rule{
		Name:             "cpp11",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 0, "g++", "clang++") {
				return false
			}
			return strings.Contains(c.Output,
				"This file requires compiler and library support for the ISO C++ 2011 standard.") ||
				strings.Contains(c.Output, "-Wc++11-extensions")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script + " -std=c++11"}
		},
	})

	// java: `java foo.java` -> strip .java
	Register(&types.Rule{
		Name:             "java",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "java") && strings.HasSuffix(c.Script, ".java")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script[:len(c.Script)-5]}
		},
	})

	// javac: `javac foo` -> add .java
	Register(&types.Rule{
		Name:             "javac",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "javac") && !strings.HasSuffix(c.Script, ".java")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script + ".java"}
		},
	})

	// php_s: `-s` -> `-S`
	Register(&types.Rule{
		Name:             "php_s",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			if !utils.IsApp(c, 1, "php") {
				return false
			}
			parts := c.ScriptParts()
			hasS := false
			for _, p := range parts {
				if p == "-s" {
					hasS = true
					break
				}
			}
			return hasS && parts[len(parts)-1] != "-s"
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "-s", "-S")}
		},
	})

	// python_command: suggest prefixing with python
	Register(&types.Rule{
		Name:             "python_command",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: specific.SudoMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) == 0 || !strings.HasSuffix(parts[0], ".py") {
				return false
			}
			return strings.Contains(c.Output, "Permission denied") ||
				strings.Contains(c.Output, "command not found")
		}),
		GetNewCommand: specific.SudoRewrite(func(c *types.Command) []string {
			return []string{"python " + c.Script}
		}),
	})

	// python_execute: `python foo` -> `python foo.py`
	Register(&types.Rule{
		Name:             "python_execute",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "python") && !strings.HasSuffix(c.Script, ".py")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{c.Script + ".py"}
		},
	})

	// ag_literal
	Register(&types.Rule{
		Name:             "ag_literal",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "ag") && strings.HasSuffix(c.Output, "run ag with -Q\n")
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{strings.Replace(c.Script, "ag", "ag -Q", 1)}
		},
	})

	// no_such_file (cp/mv into missing dir)
	nsfPats := []*regexp.Regexp{
		regexp.MustCompile(`mv: cannot move '[^']*' to '([^']*)': No such file or directory`),
		regexp.MustCompile(`mv: cannot move '[^']*' to '([^']*)': Not a directory`),
		regexp.MustCompile(`cp: cannot create regular file '([^']*)': No such file or directory`),
		regexp.MustCompile(`cp: cannot create regular file '([^']*)': Not a directory`),
	}
	Register(&types.Rule{
		Name:             "no_such_file",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			for _, p := range nsfPats {
				if p.MatchString(c.Output) {
					return true
				}
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			for _, p := range nsfPats {
				m := p.FindStringSubmatch(c.Output)
				if m != nil {
					file := m[1]
					idx := strings.LastIndex(file, "/")
					dir := ""
					if idx > 0 {
						dir = file[:idx]
					}
					return []string{shells.Current.And("mkdir -p "+dir, c.Script)}
				}
			}
			return nil
		},
	})

	// sed_unterminated_s
	Register(&types.Rule{
		Name:             "sed_unterminated_s",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 0, "sed") && strings.Contains(c.Output, "unterminated `s' command")
		},
		GetNewCommand: func(c *types.Command) []string {
			parts, err := shlexSplitLocal(c.Script)
			if err != nil {
				parts = strings.Split(c.Script, " ")
			}
			for i, e := range parts {
				if (strings.HasPrefix(e, "s/") || strings.HasPrefix(e, "-es/")) && !strings.HasSuffix(e, "/") {
					parts[i] = e + "/"
				}
			}
			out := make([]string, len(parts))
			for i, p := range parts {
				out[i] = shells.Current.Quote(p)
			}
			return []string{strings.Join(out, " ")}
		},
	})
}
