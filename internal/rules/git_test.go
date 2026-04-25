package rules

import "testing"

// ---- git_add ----
func TestGitAdd(t *testing.T) {
	withTmpDir(t)
	touchFile(t, "unknown")
	
	out := "error: pathspec 'unknown' did not match any file(s) known to git."
	
	assertMatch(t, "git_add", cmd("git submodule update unknown", out), true)
	assertMatch(t, "git_add", cmd("git commit unknown", out), true)
	
	assertMatch(t, "git_add", cmd("git submodule update known", ""), false)
	assertMatch(t, "git_add", cmd("git commit known", ""), false)
	outMissing := "error: pathspec 'missing' did not match any file(s) known to git."
	assertMatch(t, "git_add", cmd("git submodule update missing", outMissing), false) 
	
	assertNewCommand(t, "git_add", cmd("git submodule update unknown", out), "git add -- unknown && git submodule update unknown")
	assertNewCommand(t, "git_add", cmd("git commit unknown", out), "git add -- unknown && git commit unknown")
}

// ---- git_pull ----
func TestGitPull(t *testing.T) {
	out := `There is no tracking information for the current branch.
Please specify which branch you want to merge with.
See git-pull(1) for details

    git pull <remote> <branch>

If you wish to set tracking information for this branch you can do so with:

    git branch --set-upstream-to=<remote>/<branch> master

`
	assertMatch(t, "git_pull", cmd("git pull", out), true)
	assertMatch(t, "git_pull", cmd("git pull", ""), false)
	assertMatch(t, "git_pull", cmd("ls", out), false)
	
	assertNewCommand(t, "git_pull", cmd("git pull", out), "git branch --set-upstream-to=origin/master master && git pull")
}

// ---- git_add_force ----
func TestGitAddForce(t *testing.T) {
	out := "The following paths are ignored by one of your .gitignore files:\n" +
		"dist/app.js\ndist/background.js\ndist/options.js\n" +
		"Use -f if you really want to add them.\n"
	assertMatch(t, "git_add_force", cmd("git add dist/*.js", out), true)
	assertMatch(t, "git_add_force", cmd("git add dist/*.js", ""), false)
	assertNewCommand(t, "git_add_force", cmd("git add dist/*.js", out), "git add --force dist/*.js")
}

// ---- git_commit_add ----
func TestGitCommitAdd(t *testing.T) {
	assertMatch(t, "git_commit_add", cmd(`git commit -m "test"`, "no changes added to commit"), true)
	assertMatch(t, "git_commit_add", cmd("git commit", "no changes added to commit"), true)
	assertMatch(t, "git_commit_add", cmd(`git commit -m "test"`, " 1 file changed"), false)
	assertMatch(t, "git_commit_add", cmd("git branch foo", ""), false)
	assertMatch(t, "git_commit_add", cmd("git checkout feature/test_commit", ""), false)
	assertMatch(t, "git_commit_add", cmd("git push", ""), false)
	assertNewCommands(t, "git_commit_add", cmd("git commit", ""), []string{"git commit -a", "git commit -p"})
	assertNewCommands(t, "git_commit_add", cmd(`git commit -m "foo"`, ""),
		[]string{`git commit -a -m "foo"`, `git commit -p -m "foo"`})
}

// ---- git_commit_amend ----
func TestGitCommitAmend(t *testing.T) {
	assertMatch(t, "git_commit_amend", cmd(`git commit -m "test"`, "test output"), true)
	assertMatch(t, "git_commit_amend", cmd("git commit", ""), true)
	assertMatch(t, "git_commit_amend", cmd("git branch foo", ""), false)
	assertMatch(t, "git_commit_amend", cmd("git checkout feature/test_commit", ""), false)
	assertMatch(t, "git_commit_amend", cmd("git push", ""), false)
	assertNewCommand(t, "git_commit_amend", cmd(`git commit -m "test commit"`, ""), "git commit --amend")
	assertNewCommand(t, "git_commit_amend", cmd("git commit", ""), "git commit --amend")
}

// ---- git_commit_reset ----
func TestGitCommitReset(t *testing.T) {
	assertMatch(t, "git_commit_reset", cmd(`git commit -m "test"`, "test output"), true)
	assertMatch(t, "git_commit_reset", cmd("git commit", ""), true)
	assertMatch(t, "git_commit_reset", cmd("git branch foo", ""), false)
	assertMatch(t, "git_commit_reset", cmd("git checkout feature/test_commit", ""), false)
	assertMatch(t, "git_commit_reset", cmd("git push", ""), false)
	assertNewCommand(t, "git_commit_reset", cmd(`git commit -m "test commit"`, ""), "git reset HEAD~")
	assertNewCommand(t, "git_commit_reset", cmd("git commit", ""), "git reset HEAD~")
}

// ---- git_diff_no_index ----
func TestGitDiffNoIndex(t *testing.T) {
	assertMatch(t, "git_diff_no_index", cmd("git diff foo bar", ""), true)
	assertMatch(t, "git_diff_no_index", cmd("git diff --no-index foo bar", ""), false)
	assertMatch(t, "git_diff_no_index", cmd("git diff foo", ""), false)
	assertMatch(t, "git_diff_no_index", cmd("git diff foo bar baz", ""), false)
	assertNewCommand(t, "git_diff_no_index", cmd("git diff foo bar", ""), "git diff --no-index foo bar")
}

// ---- git_diff_staged ----
func TestGitDiffStaged(t *testing.T) {
	assertMatch(t, "git_diff_staged", cmd("git diff foo", ""), true)
	assertMatch(t, "git_diff_staged", cmd("git diff", ""), true)
	assertMatch(t, "git_diff_staged", cmd("git diff --staged", ""), false)
	assertMatch(t, "git_diff_staged", cmd("git tag", ""), false)
	assertMatch(t, "git_diff_staged", cmd("git branch", ""), false)
	assertMatch(t, "git_diff_staged", cmd("git log", ""), false)
	assertNewCommand(t, "git_diff_staged", cmd("git diff", ""), "git diff --staged")
	assertNewCommand(t, "git_diff_staged", cmd("git diff foo", ""), "git diff --staged foo")
}

// ---- git_merge ----
func TestGitMerge(t *testing.T) {
	out := "merge: local - not something we can merge\n\nDid you mean this?\n\tremote/local"
	assertMatch(t, "git_merge", cmd("git merge test", out), true)
	assertMatch(t, "git_merge", cmd("git merge master", ""), false)
	assertMatch(t, "git_merge", cmd("ls", out), false)
	assertNewCommand(t, "git_merge", cmd("git merge local", out), "git merge remote/local")
	assertNewCommand(t, "git_merge", cmd(`git merge -m "test" local`, out), `git merge -m "test" remote/local`)
	assertNewCommand(t, "git_merge", cmd(`git merge -m "test local" local`, out), `git merge -m "test local" remote/local`)
}

// ---- git_merge_unrelated ----
func TestGitMergeUnrelated(t *testing.T) {
	out := "fatal: refusing to merge unrelated histories"
	assertMatch(t, "git_merge_unrelated", cmd("git merge test", out), true)
	assertMatch(t, "git_merge_unrelated", cmd("git merge master", ""), false)
	assertMatch(t, "git_merge_unrelated", cmd("ls", out), false)
	assertNewCommand(t, "git_merge_unrelated", cmd("git merge local", out), "git merge local --allow-unrelated-histories")
	assertNewCommand(t, "git_merge_unrelated", cmd(`git merge -m "test" local`, out), `git merge -m "test" local --allow-unrelated-histories`)
}

// ---- git_pull_clone ----
func TestGitPullClone(t *testing.T) {
	err := "\nfatal: Not a git repository (or any parent up to mount point /home)\n" +
		"Stopping at filesystem boundary (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).\n"
	assertMatch(t, "git_pull_clone", cmd("git pull git@github.com:mcarton/thefuck.git", err), true)
	assertNewCommand(t, "git_pull_clone", cmd("git pull git@github.com:mcarton/thefuck.git", err),
		"git clone git@github.com:mcarton/thefuck.git")
}

// ---- git_pull_uncommitted_changes ----
func TestGitPullUncommittedChanges(t *testing.T) {
	out1 := "error: Cannot pull with rebase: You have unstaged changes."
	out2 := "error: Cannot pull with rebase: Your index contains uncommitted changes."
	for _, out := range []string{out1, out2} {
		assertMatch(t, "git_pull_uncommitted_changes", cmd("git pull", out), true)
		assertMatch(t, "git_pull_uncommitted_changes", cmd("git pull", ""), false)
		assertMatch(t, "git_pull_uncommitted_changes", cmd("ls", out), false)
		assertNewCommand(t, "git_pull_uncommitted_changes", cmd("git pull", out), "git stash && git pull && git stash pop")
	}
}

// ---- git_rebase_no_changes ----
func TestGitRebaseNoChanges(t *testing.T) {
	out := "Applying: Test commit\nNo changes - did you forget to use 'git add'?\nRun git rebase --continue.\n"
	assertMatch(t, "git_rebase_no_changes", cmd("git rebase --continue", out), true)
	assertMatch(t, "git_rebase_no_changes", cmd("git rebase --continue", ""), false)
	assertMatch(t, "git_rebase_no_changes", cmd("git rebase --skip", ""), false)
	assertNewCommand(t, "git_rebase_no_changes", cmd("git rebase --continue", out), "git rebase --skip")
}

// ---- git_remote_delete ----
func TestGitRemoteDelete(t *testing.T) {
	assertMatch(t, "git_remote_delete", cmd("git remote delete foo", ""), true)
	assertMatch(t, "git_remote_delete", cmd("git remote remove foo", ""), false)
	assertMatch(t, "git_remote_delete", cmd("git remote add foo", ""), false)
	assertMatch(t, "git_remote_delete", cmd("git commit", ""), false)
	assertNewCommand(t, "git_remote_delete", cmd("git remote delete foo", ""), "git remote remove foo")
	assertNewCommand(t, "git_remote_delete", cmd("git remote delete delete", ""), "git remote remove delete")
}

// ---- git_remote_seturl_add ----
func TestGitRemoteSeturlAdd(t *testing.T) {
	assertMatch(t, "git_remote_seturl_add", cmd("git remote set-url origin url", "fatal: No such remote"), true)
	assertMatch(t, "git_remote_seturl_add", cmd("git remote set-url origin url", ""), false)
	assertMatch(t, "git_remote_seturl_add", cmd("git remote add origin url", ""), false)
	assertMatch(t, "git_remote_seturl_add", cmd("git remote remove origin", ""), false)
	assertMatch(t, "git_remote_seturl_add", cmd("git remote prune origin", ""), false)
	assertMatch(t, "git_remote_seturl_add", cmd("git remote set-branches origin branch", ""), false)
	assertNewCommand(t, "git_remote_seturl_add", cmd("git remote set-url origin git@github.com:nvbn/thefuck.git", ""),
		"git remote add origin git@github.com:nvbn/thefuck.git")
}

// ---- git_stash ----
func TestGitStash(t *testing.T) {
	cp := "error: Your local changes would be overwritten by cherry-pick.\n" +
		"hint: Commit your changes or stash them to proceed.\nfatal: cherry-pick failed"
	rb := "Cannot rebase: Your index contains uncommitted changes.\nPlease commit or stash them."
	assertMatch(t, "git_stash", cmd("git cherry-pick a1b2c3d", cp), true)
	assertMatch(t, "git_stash", cmd("git rebase -i HEAD~7", rb), true)
	assertMatch(t, "git_stash", cmd("git cherry-pick a1b2c3d", ""), false)
	assertMatch(t, "git_stash", cmd("git rebase -i HEAD~7", ""), false)
	assertNewCommand(t, "git_stash", cmd("git cherry-pick a1b2c3d", cp), "git stash && git cherry-pick a1b2c3d")
	assertNewCommand(t, "git_stash", cmd("git rebase -i HEAD~7", rb), "git stash && git rebase -i HEAD~7")
}

// ---- git_stash_pop ----
func TestGitStashPop(t *testing.T) {
	out := "error: Your local changes to the following files would be overwritten by merge:"
	assertMatch(t, "git_stash_pop", cmd("git stash pop", out), true)
	assertMatch(t, "git_stash_pop", cmd("git stash", ""), false)
	assertNewCommand(t, "git_stash_pop", cmd("git stash pop", out),
		"git add --update && git stash pop && git reset .")
}

// ---- git_tag_force ----
func TestGitTagForce(t *testing.T) {
	out := "fatal: tag 'alert' already exists"
	assertMatch(t, "git_tag_force", cmd("git tag alert", out), true)
	assertMatch(t, "git_tag_force", cmd("git tag alert", ""), false)
	assertNewCommand(t, "git_tag_force", cmd("git tag alert", out), "git tag --force alert")
}

// ---- git_two_dashes ----
func TestGitTwoDashes(t *testing.T) {
	mkout := func(dash string) string {
		return "error: did you mean `" + dash + "` (with two dashes ?)"
	}
	cases := []struct{ script, dash, want string }{
		{"git add -patch", "--patch", "git add --patch"},
		{"git checkout -patch", "--patch", "git checkout --patch"},
		{"git init -bare", "--bare", "git init --bare"},
		{"git commit -amend", "--amend", "git commit --amend"},
		{"git push -tags", "--tags", "git push --tags"},
		{"git rebase -continue", "--continue", "git rebase --continue"},
	}
	for _, tc := range cases {
		assertMatch(t, "git_two_dashes", cmd(tc.script, mkout(tc.dash)), true)
		assertNewCommand(t, "git_two_dashes", cmd(tc.script, mkout(tc.dash)), tc.want)
	}
	for _, s := range []string{"git add --patch", "git checkout --patch", "git commit --amend", "git push --tags", "git rebase --continue"} {
		assertMatch(t, "git_two_dashes", cmd(s, ""), false)
	}
}

// ---- git_merge_unrelated already tested ----

// ---- git_hook_bypass ----
func TestGitHookBypass(t *testing.T) {
	assertMatch(t, "git_hook_bypass", cmd("git am file.patch", ""), true)
	assertMatch(t, "git_hook_bypass", cmd("git commit -m test", ""), true)
	assertMatch(t, "git_hook_bypass", cmd("git push origin master", ""), true)
	assertMatch(t, "git_hook_bypass", cmd("git status", ""), false)
	assertNewCommand(t, "git_hook_bypass", cmd("git commit -m test", ""), "git commit --no-verify -m test")
}

// ---- git_main_master ----
func TestGitMainMaster(t *testing.T) {
	assertMatch(t, "git_main_master", cmd("git checkout master", "error: pathspec 'master' did not match any file(s) known to git."), true)
	assertMatch(t, "git_main_master", cmd("git checkout main", "error: pathspec 'main' did not match any file(s) known to git."), true)
	assertMatch(t, "git_main_master", cmd("git checkout foo", "error: pathspec 'foo' did not match any file(s) known to git."), false)
	assertNewCommand(t, "git_main_master", cmd("git checkout master", "error: pathspec 'master' did not match any file(s) known to git."), "git checkout main")
	assertNewCommand(t, "git_main_master", cmd("git checkout main", "error: pathspec 'main' did not match any file(s) known to git."), "git checkout master")
}

// ---- git_not_command ----
func TestGitNotCommand(t *testing.T) {
	out := "git: 'brnch' is not a git command. See 'git --help'.\n\nThe most similar command is\n\tbranch\n"
	assertMatch(t, "git_not_command", cmd("git brnch", out), true)
	assertMatch(t, "git_not_command", cmd("git branch", ""), false)
	assertNewCommand(t, "git_not_command", cmd("git brnch", out), "git branch")
}

// ---- git_rm_recursive ----
func TestGitRmRecursive(t *testing.T) {
	out := "fatal: not removing 'foo' recursively without -r"
	assertMatch(t, "git_rm_recursive", cmd("git rm foo", out), true)
	assertMatch(t, "git_rm_recursive", cmd("git rm foo", ""), false)
	assertNewCommand(t, "git_rm_recursive", cmd("git rm foo", out), "git rm -r foo")
}

// ---- git_rm_local_modifications ----
func TestGitRmLocalModifications(t *testing.T) {
	out := "error: the following file has local modifications\n   x\nuse --cached to keep the file, or -f to force removal"
	assertMatch(t, "git_rm_local_modifications", cmd("git rm foo", out), true)
	assertMatch(t, "git_rm_local_modifications", cmd("git rm foo", ""), false)
	assertNewCommands(t, "git_rm_local_modifications", cmd("git rm foo", out),
		[]string{"git rm --cached foo", "git rm -f foo"})
}

// ---- git_rm_staged ----
func TestGitRmStaged(t *testing.T) {
	out := "error: the following file has changes staged in the index\n   x\nuse --cached to keep the file, or -f to force removal"
	assertMatch(t, "git_rm_staged", cmd("git rm foo", out), true)
	assertMatch(t, "git_rm_staged", cmd("git rm foo", ""), false)
	assertNewCommands(t, "git_rm_staged", cmd("git rm foo", out),
		[]string{"git rm --cached foo", "git rm -f foo"})
}

// ---- git_clone_missing ----
func TestGitCloneMissing(t *testing.T) {
	withPath(t, "") // clear PATH so 'https://...' is not found
	
	urls := []string{
		"https://github.com/nvbn/thefuck.git",
		"https://github.com/nvbn/thefuck",
		"http://github.com/nvbn/thefuck.git",
		"git@github.com:nvbn/thefuck.git",
		"git@github.com:nvbn/thefuck",
		"ssh://git@github.com:nvbn/thefuck.git",
	}
	
	invalid := []string{
		"",  // No command
		"notacommand",  // Command not found
		"ssh git@github.com:nvbn/thefrick.git",  // ssh command, not a git clone
		"git clone foo",  // Valid clone
		"git clone https://github.com/nvbn/thefuck.git",  // Full command
		"github.com/nvbn/thefuck.git",  // Missing protocol
		"github.com:nvbn/thefuck.git",  // SSH missing username
		"git clone git clone ssh://git@github.com:nvbn/thefrick.git",  // 2x clone
		"https:/github.com/nvbn/thefuck.git",  // Bad protocol
	}

	outputs := []string{
		"No such file or directory",
		"not found",
		"is not recognised as",
	}

	for _, u := range urls {
		for _, o := range outputs {
			assertMatch(t, "git_clone_missing", cmd(u, o), true)
			assertNewCommand(t, "git_clone_missing", cmd(u, o), "git clone "+u)
		}
	}
	
	for _, u := range invalid {
		for _, o := range outputs {
			assertMatch(t, "git_clone_missing", cmd(u, o), false)
		}
		assertMatch(t, "git_clone_missing", cmd(u, "some other output"), false)
	}
}

// ---- git_rebase_merge_dir ----
func TestGitRebaseMergeDir(t *testing.T) {
	out := `It seems that there is already a rebase-merge directory, and
I wonder if you are in the middle of another rebase.  If that is the
case, please try
	git rebase (--continue | --abort | --skip)
If that is not the case, please
	rm -fr ".git/rebase-merge"
and run me again.  I am stopping in case you still have something
valuable there.
`

	assertMatch(t, "git_rebase_merge_dir", cmd("git rebase master", out), true)
	assertMatch(t, "git_rebase_merge_dir", cmd("git rebase -skip", out), true)
	
	assertMatch(t, "git_rebase_merge_dir", cmd("git rebase master", ""), false)
	assertMatch(t, "git_rebase_merge_dir", cmd("ls", out), false)
	
	// Check new commands generated (the last option parsed from output should be rm -fr ".git/rebase-merge")
	assertNewCommands(t, "git_rebase_merge_dir", cmd("git rebase master", out),
		[]string{"git rebase --abort", "git rebase --skip", "git rebase --continue", `rm -fr ".git/rebase-merge"`})
	
	assertNewCommands(t, "git_rebase_merge_dir", cmd("git rebase -skip", out),
		[]string{"git rebase --skip", "git rebase --abort", "git rebase --continue", `rm -fr ".git/rebase-merge"`})
}
