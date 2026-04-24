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

var (
	gitPathspecRe      = regexp.MustCompile(`error: pathspec '([^']*)' did not match any file\(s\) known to git\.?`)
	gitBranchExistsRe  = regexp.MustCompile(`fatal: A branch named '(.+)' already exists\.`)
	gitMergeUnknownRe  = regexp.MustCompile(`merge: (.+) - not something we can merge`)
	gitMergeDidMeanRe  = regexp.MustCompile(`Did you mean this\?\n\t([^\n]+)`)
	gitNotCmdRe        = regexp.MustCompile(`git: '([^']*)' is not a git command`)
	gitLfsUnknownRe    = regexp.MustCompile(`Error: unknown command "([^"]*)" for "git-lfs"`)
	gitBisectBrokenRe  = regexp.MustCompile(`git bisect ([^ $]*).*`)
	gitBisectUsageRe   = regexp.MustCompile(`usage: git bisect \[([^\]]+)\]`)
	gitPushUpstreamRe  = regexp.MustCompile(`(?m)^ +(git push [^\s]+ [^\s]+)`)
	gitFlagAfterFnRe   = regexp.MustCompile(`fatal: bad flag '(.*?)' used after filename`)
	gitFlagAfterFnRe2  = regexp.MustCompile(`fatal: option '(.*?)' must come before non-option arguments`)
	gitPushArgsRe      = regexp.MustCompile(`git push (.*)`)
	gitNoRefspecRe     = regexp.MustCompile(`src refspec \w+ does not match any`)
)

func contains(parts []string, s string) bool {
	for _, p := range parts {
		if p == s {
			return true
		}
	}
	return false
}

func indexOf(parts []string, s string) int {
	for i, p := range parts {
		if p == s {
			return i
		}
	}
	return -1
}

func init() {
	// git_add
	Register(&types.Rule{
		Name: "git_add", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			m := gitPathspecRe.FindStringSubmatch(c.Output)
			if m == nil {
				return false
			}
			if !strings.Contains(c.Output, "did not match any file(s) known to git.") {
				return false
			}
			_, err := os.Stat(m[1])
			return err == nil
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			m := gitPathspecRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{shells.Current.And("git add -- "+m[1], c.Script)}
		}),
	})

	// git_add_force
	Register(&types.Rule{
		Name: "git_add_force", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return contains(c.ScriptParts(), "add") &&
				strings.Contains(c.Output, "Use -f if you really want to add them.")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "add", "add --force")}
		}),
	})

	// git_bisect_usage
	Register(&types.Rule{
		Name: "git_bisect_usage", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return contains(c.ScriptParts(), "bisect") && strings.Contains(c.Output, "usage: git bisect")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			m := gitBisectBrokenRe.FindStringSubmatch(c.Script)
			u := gitBisectUsageRe.FindStringSubmatch(c.Output)
			if m == nil || u == nil {
				return nil
			}
			matched := strings.Split(u[1], "|")
			return utils.ReplaceCommand(c, m[1], matched)
		}),
	})

	// git_branch_0flag (0 typed instead of -)
	Register(&types.Rule{
		Name: "git_branch_0flag", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) < 2 || parts[1] != "branch" {
				return false
			}
			for _, p := range parts {
				if len(p) == 2 && strings.HasPrefix(p, "0") {
					return true
				}
			}
			return false
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			var flag string
			for _, p := range c.ScriptParts() {
				if len(p) == 2 && strings.HasPrefix(p, "0") {
					flag = p
					break
				}
			}
			fixed := strings.Replace(flag, "0", "-", 1)
			fixedScript := strings.Replace(c.Script, flag, fixed, 1)
			if strings.Contains(c.Output, "A branch named '") && strings.Contains(c.Output, "' already exists.") {
				return []string{shells.Current.And("git branch -D "+flag, fixedScript)}
			}
			return []string{fixedScript}
		}),
	})

	// git_branch_delete
	Register(&types.Rule{
		Name: "git_branch_delete", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "branch -d") &&
				strings.Contains(c.Output, "If you are sure you want to delete it")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "-d", "-D")}
		}),
	})

	// git_branch_delete_checked_out
	Register(&types.Rule{
		Name: "git_branch_delete_checked_out", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return (strings.Contains(c.Script, "branch -d") || strings.Contains(c.Script, "branch -D")) &&
				strings.Contains(c.Output, "error: Cannot delete branch '") &&
				strings.Contains(c.Output, "' checked out at '")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{shells.Current.And("git checkout master", utils.ReplaceArgument(c.Script, "-d", "-D"))}
		}),
	})

	// git_branch_exists
	Register(&types.Rule{
		Name: "git_branch_exists", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Output, "fatal: A branch named '") &&
				strings.Contains(c.Output, "' already exists.")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			m := gitBranchExistsRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			name := strings.ReplaceAll(m[1], "'", `\'`)
			templates := [][]string{
				{"git branch -d NAME", "git branch NAME"},
				{"git branch -d NAME", "git checkout -b NAME"},
				{"git branch -D NAME", "git branch NAME"},
				{"git branch -D NAME", "git checkout -b NAME"},
				{"git checkout NAME"},
			}
			var out []string
			for _, t := range templates {
				replaced := make([]string, len(t))
				for i, s := range t {
					replaced[i] = strings.ReplaceAll(s, "NAME", name)
				}
				out = append(out, shells.Current.And(replaced...))
			}
			return out
		}),
	})

	// git_branch_list
	Register(&types.Rule{
		Name: "git_branch_list", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) < 3 {
				return false
			}
			return strings.Join(parts[1:], " ") == "branch list"
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{shells.Current.And("git branch --delete list", "git branch")}
		}),
	})

	// git_checkout
	Register(&types.Rule{
		Name: "git_checkout", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Output, "did not match any file(s) known to git") &&
				!strings.Contains(c.Output, "Did you forget to 'git add'?")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			m := gitPathspecRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			missing := m[1]
			var cmds []string
			parts := c.ScriptParts()
			if len(parts) > 1 && parts[1] == "checkout" {
				cmds = append(cmds, utils.ReplaceArgument(c.Script, "checkout", "checkout -b"))
			}
			if len(cmds) == 0 {
				cmds = append(cmds, shells.Current.And("git branch "+missing, c.Script))
			}
			return cmds
		}),
	})

	// git_clone_git_clone (pasted `git clone git clone ...`)
	Register(&types.Rule{
		Name: "git_clone_git_clone", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, " git clone ") &&
				strings.Contains(c.Output, "fatal: Too many arguments.")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{strings.Replace(c.Script, " git clone ", " ", 1)}
		}),
	})

	// git_clone_missing
	Register(&types.Rule{
		Name: "git_clone_missing", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) != 1 {
				return false
			}
			if utils.Which(parts[0]) != "" {
				return false
			}
			if !(strings.Contains(c.Output, "No such file or directory") ||
				strings.Contains(c.Output, "not found") ||
				strings.Contains(c.Output, "is not recognised as")) {
				return false
			}
			s := c.Script
			// Accept http(s)://host/path and user@host:path (ssh)
			if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
				rest := s
				if strings.HasPrefix(s, "http://") {
					rest = s[7:]
				} else {
					rest = s[8:]
				}
				slash := strings.IndexByte(rest, '/')
				host := rest
				if slash >= 0 {
					host = rest[:slash]
				}
				return host != ""
			}
			if strings.Contains(s, "@") && strings.Contains(s, ":") {
				return true
			}
			return false
		},
		GetNewCommand: func(c *types.Command) []string {
			return []string{"git clone " + c.Script}
		},
	})

	// git_commit_add (no changes added to commit)
	Register(&types.Rule{
		Name: "git_commit_add", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return contains(c.ScriptParts(), "commit") &&
				strings.Contains(c.Output, "no changes added to commit")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{
				utils.ReplaceArgument(c.Script, "commit", "commit -a"),
				utils.ReplaceArgument(c.Script, "commit", "commit -p"),
			}
		}),
	})

	// git_commit_amend
	Register(&types.Rule{
		Name: "git_commit_amend", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return contains(c.ScriptParts(), "commit")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{"git commit --amend"}
		}),
	})

	// git_commit_reset
	Register(&types.Rule{
		Name: "git_commit_reset", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return contains(c.ScriptParts(), "commit")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{"git reset HEAD~"}
		}),
	})

	// git_diff_no_index
	Register(&types.Rule{
		Name: "git_diff_no_index", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) < 3 {
				return false
			}
			files := 0
			for _, a := range parts[2:] {
				if !strings.HasPrefix(a, "-") {
					files++
				}
			}
			return strings.Contains(c.Script, "diff") &&
				!strings.Contains(c.Script, "--no-index") &&
				files == 2
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "diff", "diff --no-index")}
		}),
	})

	// git_diff_staged
	Register(&types.Rule{
		Name: "git_diff_staged", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "diff") && !strings.Contains(c.Script, "--staged")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "diff", "diff --staged")}
		}),
	})

	// git_fix_stash
	stashCmds := []string{"apply", "branch", "clear", "drop", "list", "pop", "save", "show"}
	Register(&types.Rule{
		Name: "git_fix_stash", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			if len(parts) <= 1 {
				return false
			}
			return parts[1] == "stash" && strings.Contains(c.Output, "usage:")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			parts := c.ScriptParts()
			if len(parts) < 3 {
				return nil
			}
			stashCmd := parts[2]
			fixed := utils.GetClosest(stashCmd, stashCmds, 0.6, false)
			if fixed != "" {
				return []string{utils.ReplaceArgument(c.Script, stashCmd, fixed)}
			}
			cp := append([]string{}, parts...)
			ins := append([]string{}, cp[:2]...)
			ins = append(ins, "save")
			ins = append(ins, cp[2:]...)
			return []string{strings.Join(ins, " ")}
		}),
	})

	// git_flag_after_filename
	Register(&types.Rule{
		Name: "git_flag_after_filename", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return gitFlagAfterFnRe.MatchString(c.Output) || gitFlagAfterFnRe2.MatchString(c.Output)
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			var badFlag string
			if m := gitFlagAfterFnRe.FindStringSubmatch(c.Output); m != nil {
				badFlag = m[1]
			} else if m := gitFlagAfterFnRe2.FindStringSubmatch(c.Output); m != nil {
				badFlag = m[1]
			} else {
				return nil
			}
			parts := append([]string{}, c.ScriptParts()...)
			flagIdx := indexOf(parts, badFlag)
			if flagIdx < 0 {
				return nil
			}
			fnIdx := -1
			for i := flagIdx - 1; i >= 0; i-- {
				if !strings.HasPrefix(parts[i], "-") {
					fnIdx = i
					break
				}
			}
			if fnIdx < 0 {
				return nil
			}
			parts[flagIdx], parts[fnIdx] = parts[fnIdx], parts[flagIdx]
			return []string{strings.Join(parts, " ")}
		}),
	})

	// git_help_aliased
	Register(&types.Rule{
		Name: "git_help_aliased", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "help") && strings.Contains(c.Output, " is aliased to ")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			// Mirrors: output.split('`', 2)[2].split("'", 1)[0].split(' ', 1)[0]
			a := strings.SplitN(c.Output, "`", 3)
			if len(a) < 3 {
				return nil
			}
			b := strings.SplitN(a[2], "'", 2)
			first := strings.SplitN(b[0], " ", 2)[0]
			return []string{"git help " + first}
		}),
	})

	// git_hook_bypass
	hooked := []string{"am", "commit", "push"}
	Register(&types.Rule{
		Name: "git_hook_bypass", EnabledByDefault: true, RequiresOutput: false,
		Priority: 1100,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			for _, h := range hooked {
				if contains(parts, h) {
					return true
				}
			}
			return false
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			parts := c.ScriptParts()
			for _, h := range hooked {
				if contains(parts, h) {
					return []string{utils.ReplaceArgument(c.Script, h, h+" --no-verify")}
				}
			}
			return nil
		}),
	})

	// git_lfs_mistype
	Register(&types.Rule{
		Name: "git_lfs_mistype", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "lfs") && strings.Contains(c.Output, "Did you mean this?")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			m := gitLfsUnknownRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			matched := utils.GetAllMatchedCommands(c.Output, "Did you mean", " for usage.")
			return utils.ReplaceCommand(c, m[1], matched)
		}),
	})

	// git_main_master
	Register(&types.Rule{
		Name: "git_main_master", EnabledByDefault: true, RequiresOutput: true,
		Priority: 1200,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Output, "'master'") || strings.Contains(c.Output, "'main'")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			if strings.Contains(c.Output, "'master'") {
				return []string{strings.ReplaceAll(c.Script, "master", "main")}
			}
			return []string{strings.ReplaceAll(c.Script, "main", "master")}
		}),
	})

	// git_merge
	Register(&types.Rule{
		Name: "git_merge", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "merge") &&
				strings.Contains(c.Output, " - not something we can merge") &&
				strings.Contains(c.Output, "Did you mean this?")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			unk := gitMergeUnknownRe.FindStringSubmatch(c.Output)
			rem := gitMergeDidMeanRe.FindStringSubmatch(c.Output)
			if unk == nil || rem == nil {
				return nil
			}
			return []string{utils.ReplaceArgument(c.Script, unk[1], rem[1])}
		}),
	})

	// git_merge_unrelated
	Register(&types.Rule{
		Name: "git_merge_unrelated", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "merge") &&
				strings.Contains(c.Output, "fatal: refusing to merge unrelated histories")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{c.Script + " --allow-unrelated-histories"}
		}),
	})

	// git_not_command
	Register(&types.Rule{
		Name: "git_not_command", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Output, " is not a git command. See 'git --help'.") &&
				(strings.Contains(c.Output, "The most similar command") || strings.Contains(c.Output, "Did you mean"))
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			m := gitNotCmdRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			matched := utils.GetAllMatchedCommands(c.Output, "The most similar command", "Did you mean")
			return utils.ReplaceCommand(c, m[1], matched)
		}),
	})

	// git_pull
	Register(&types.Rule{
		Name: "git_pull", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "pull") && strings.Contains(c.Output, "set-upstream")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			lines := strings.Split(c.Output, "\n")
			if len(lines) < 3 {
				return nil
			}
			line := strings.TrimSpace(lines[len(lines)-3])
			parts := strings.Split(line, " ")
			branch := parts[len(parts)-1]
			setUpstream := strings.ReplaceAll(strings.ReplaceAll(line, "<remote>", "origin"), "<branch>", branch)
			return []string{shells.Current.And(setUpstream, c.Script)}
		}),
	})

	// git_pull_clone
	Register(&types.Rule{
		Name: "git_pull_clone", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Output, "fatal: Not a git repository") &&
				strings.Contains(c.Output, "Stopping at filesystem boundary (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "pull", "clone")}
		}),
	})

	// git_pull_uncommitted_changes
	Register(&types.Rule{
		Name: "git_pull_uncommitted_changes", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "pull") &&
				(strings.Contains(c.Output, "You have unstaged changes") ||
					strings.Contains(c.Output, "contains uncommitted changes"))
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{shells.Current.And("git stash", "git pull", "git stash pop")}
		}),
	})

	// git_push
	Register(&types.Rule{
		Name: "git_push", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return contains(c.ScriptParts(), "push") && strings.Contains(c.Output, "git push --set-upstream")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			idx := indexOf(parts, "--set-upstream")
			if idx < 0 {
				idx = indexOf(parts, "-u")
			}
			if idx >= 0 {
				parts = append(parts[:idx], parts[idx+1:]...)
				if len(parts) > idx {
					parts = append(parts[:idx], parts[idx+1:]...)
				}
			} else {
				pushIdx := indexOf(parts, "push") + 1
				for len(parts) > pushIdx && !strings.HasPrefix(parts[len(parts)-1], "-") {
					parts = parts[:len(parts)-1]
				}
			}
			all := gitPushArgsRe.FindAllStringSubmatch(c.Output, -1)
			if len(all) == 0 {
				return nil
			}
			args := strings.ReplaceAll(strings.TrimSpace(all[len(all)-1][1]), "'", `\'`)
			return []string{utils.ReplaceArgument(strings.Join(parts, " "), "push", "push "+args)}
		}),
	})

	// git_push_different_branch_names
	Register(&types.Rule{
		Name: "git_push_different_branch_names", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "push") &&
				strings.Contains(c.Output, "The upstream branch of your current branch does not match")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			m := gitPushUpstreamRe.FindStringSubmatch(c.Output)
			if m == nil {
				return nil
			}
			return []string{m[1]}
		}),
	})

	// git_push_force
	Register(&types.Rule{
		Name: "git_push_force", EnabledByDefault: false, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "push") &&
				strings.Contains(c.Output, "! [rejected]") &&
				strings.Contains(c.Output, "failed to push some refs to") &&
				strings.Contains(c.Output, "Updates were rejected because the tip of your current branch is behind")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "push", "push --force-with-lease")}
		}),
	})

	// git_push_pull
	Register(&types.Rule{
		Name: "git_push_pull", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "push") &&
				strings.Contains(c.Output, "! [rejected]") &&
				strings.Contains(c.Output, "failed to push some refs to") &&
				(strings.Contains(c.Output, "Updates were rejected because the tip of your current branch is behind") ||
					strings.Contains(c.Output, "Updates were rejected because the remote contains work that you do"))
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{shells.Current.And(utils.ReplaceArgument(c.Script, "push", "pull"), c.Script)}
		}),
	})

	// git_push_without_commits
	Register(&types.Rule{
		Name: "git_push_without_commits", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return gitNoRefspecRe.MatchString(c.Output)
		}),
		GetNewCommand: func(c *types.Command) []string {
			return []string{shells.Current.And(`git commit -m "Initial commit"`, c.Script)}
		},
	})

	// git_rebase_merge_dir
	Register(&types.Rule{
		Name: "git_rebase_merge_dir", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, " rebase") &&
				strings.Contains(c.Output, "It seems that there is already a rebase-merge directory") &&
				strings.Contains(c.Output, "I wonder if you are in the middle of another rebase")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			commands := []string{"git rebase --continue", "git rebase --abort", "git rebase --skip"}
			lines := strings.Split(c.Output, "\n")
			if len(lines) >= 4 {
				commands = append(commands, strings.TrimSpace(lines[len(lines)-4]))
			}
			return utils.GetCloseMatches(c.Script, commands, 4, 0)
		}),
	})

	// git_rebase_no_changes
	Register(&types.Rule{
		Name: "git_rebase_no_changes", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			parts := c.ScriptParts()
			return contains(parts, "rebase") && contains(parts, "--continue") &&
				strings.Contains(c.Output, "No changes - did you forget to use 'git add'?")
		}),
		GetNewCommand: func(c *types.Command) []string {
			return []string{"git rebase --skip"}
		},
	})

	// git_remote_delete
	Register(&types.Rule{
		Name: "git_remote_delete", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "remote delete")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{strings.Replace(c.Script, "delete", "remove", 1)}
		}),
	})

	// git_remote_seturl_add
	Register(&types.Rule{
		Name: "git_remote_seturl_add", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "set-url") && strings.Contains(c.Output, "fatal: No such remote")
		}),
		GetNewCommand: func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "set-url", "add")}
		},
	})

	// git_rm_local_modifications
	Register(&types.Rule{
		Name: "git_rm_local_modifications", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, " rm ") &&
				strings.Contains(c.Output, "error: the following file has local modifications") &&
				strings.Contains(c.Output, "use --cached to keep the file, or -f to force removal")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			i := indexOf(parts, "rm") + 1
			a := append(append([]string{}, parts[:i]...), "--cached")
			a = append(a, parts[i:]...)
			b := append(append([]string{}, parts[:i]...), "-f")
			b = append(b, parts[i:]...)
			return []string{strings.Join(a, " "), strings.Join(b, " ")}
		}),
	})

	// git_rm_recursive
	Register(&types.Rule{
		Name: "git_rm_recursive", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, " rm ") &&
				strings.Contains(c.Output, "fatal: not removing '") &&
				strings.Contains(c.Output, "' recursively without -r")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			i := indexOf(parts, "rm") + 1
			out := append(append([]string{}, parts[:i]...), "-r")
			out = append(out, parts[i:]...)
			return []string{strings.Join(out, " ")}
		}),
	})

	// git_rm_staged
	Register(&types.Rule{
		Name: "git_rm_staged", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, " rm ") &&
				strings.Contains(c.Output, "error: the following file has changes staged in the index") &&
				strings.Contains(c.Output, "use --cached to keep the file, or -f to force removal")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			parts := append([]string{}, c.ScriptParts()...)
			i := indexOf(parts, "rm") + 1
			a := append(append([]string{}, parts[:i]...), "--cached")
			a = append(a, parts[i:]...)
			b := append(append([]string{}, parts[:i]...), "-f")
			b = append(b, parts[i:]...)
			return []string{strings.Join(a, " "), strings.Join(b, " ")}
		}),
	})

	// git_stash
	Register(&types.Rule{
		Name: "git_stash", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Output, "or stash them")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{shells.Current.And("git stash", c.Script)}
		}),
	})

	// git_stash_pop
	Register(&types.Rule{
		Name: "git_stash_pop", EnabledByDefault: true, RequiresOutput: true,
		Priority: 900,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Script, "stash") && strings.Contains(c.Script, "pop") &&
				strings.Contains(c.Output, "Your local changes to the following files would be overwritten by merge")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{shells.Current.And("git add --update", "git stash pop", "git reset .")}
		}),
	})

	// git_tag_force
	Register(&types.Rule{
		Name: "git_tag_force", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return contains(c.ScriptParts(), "tag") && strings.Contains(c.Output, "already exists")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			return []string{utils.ReplaceArgument(c.Script, "tag", "tag --force")}
		}),
	})

	// git_two_dashes
	Register(&types.Rule{
		Name: "git_two_dashes", EnabledByDefault: true, RequiresOutput: true,
		Match: specific.GitSupportMatch(func(c *types.Command) bool {
			return strings.Contains(c.Output, "error: did you mean `") && strings.Contains(c.Output, "` (with two dashes ?)")
		}),
		GetNewCommand: specific.GitSupportRewrite(func(c *types.Command) []string {
			bits := strings.Split(c.Output, "`")
			if len(bits) < 2 {
				return nil
			}
			to := bits[1]
			if len(to) < 2 {
				return nil
			}
			return []string{utils.ReplaceArgument(c.Script, to[1:], to)}
		}),
	})
}
