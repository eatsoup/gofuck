package rules

import "testing"

// ---- tsuru_not_command ----
func TestTsuruNotCommand(t *testing.T) {
	mkOut := func(body string) string {
		return `tsuru: "tchururu" is not a tsuru command. See "tsuru help".` + "\n\nDid you mean?\n" + body
	}
	assertMatch(t, "tsuru_not_command", cmd("tsuru log", mkOut("\tapp-log\n\tlogin\n\tlogout\n")), true)
	assertMatch(t, "tsuru_not_command", cmd("tsuru app-l", mkOut("\tapp-list\n\tapp-log\n")), true)
	assertMatch(t, "tsuru_not_command", cmd("tsuru version", "tsuru version 0.16.0."), false)
}

// ---- cp_create_destination ----
func TestCpCreateDestination(t *testing.T) {
	assertMatch(t, "cp_create_destination", cmd("cp", "cp: directory foo does not exist\n"), true)
	assertMatch(t, "cp_create_destination", cmd("mv", "No such file or directory"), true)
	assertMatch(t, "cp_create_destination", cmd("cp", ""), false)
	assertMatch(t, "cp_create_destination", cmd("mv", ""), false)
	assertMatch(t, "cp_create_destination", cmd("ls", "No such file or directory"), false)
	assertNewCommand(t, "cp_create_destination", cmd("cp foo bar/", "cp: directory foo does not exist\n"),
		"mkdir -p bar/ && cp foo bar/")
	assertNewCommand(t, "cp_create_destination", cmd("mv foo bar/", "No such file or directory"),
		"mkdir -p bar/ && mv foo bar/")
	assertNewCommand(t, "cp_create_destination", cmd("cp foo bar/baz/", "cp: directory foo does not exist\n"),
		"mkdir -p bar/baz/ && cp foo bar/baz/")
}

// ---- cp_omitting_directory ----
func TestCpOmittingDirectory(t *testing.T) {
	assertMatch(t, "cp_omitting_directory", cmd("cp dir", "cp: dor: is a directory"), true)
	assertMatch(t, "cp_omitting_directory", cmd("cp dir", "cp: omitting directory 'dir'"), true)
	assertMatch(t, "cp_omitting_directory", cmd("some dir", "cp: dor: is a directory"), false)
	assertMatch(t, "cp_omitting_directory", cmd("cp dir", ""), false)
	assertNewCommand(t, "cp_omitting_directory", cmd("cp dir", ""), "cp -a dir")
}

// ---- mvn_unknown_lifecycle_phase ----
func TestMvnUnknownLifecyclePhase(t *testing.T) {
	out := `[ERROR] Unknown lifecycle phase "cle". You must specify a valid lifecycle phase or a goal in the format <plugin-prefix>:<goal> or <plugin-group-id>:<plugin-artifact-id>[:<plugin-version>]:<goal>. Available lifecycle phases are: validate, initialize, generate-sources, process-sources, generate-resources, process-resources, compile, process-classes, generate-test-sources, process-test-sources, generate-test-resources, process-test-resources, test-compile, process-test-classes, test, prepare-package, package, pre-integration-test, integration-test, post-integration-test, verify, install, deploy, pre-clean, clean, post-clean, pre-site, site, post-site, site-deploy. -> [Help 1]`
	assertMatch(t, "mvn_unknown_lifecycle_phase", cmd("mvn cle", out), true)
	assertMatch(t, "mvn_unknown_lifecycle_phase", cmd("mvn --help", ""), false)
	assertMatch(t, "mvn_unknown_lifecycle_phase", cmd("mvn -v", ""), false)
	assertNewCommands(t, "mvn_unknown_lifecycle_phase", cmd("mvn cle", out),
		[]string{"mvn clean", "mvn compile"})
	out2 := `[ERROR] Unknown lifecycle phase "claen". Available lifecycle phases are: validate, initialize, compile, test, prepare-package, package, clean, pre-clean, post-clean. -> [Help 1]`
	assertNewCommands(t, "mvn_unknown_lifecycle_phase", cmd("mvn claen package", out2),
		[]string{"mvn clean package"})
}

// ---- no_such_file ----
func TestNoSuchFile(t *testing.T) {
	assertMatch(t, "no_such_file", cmd("mv foo bar/foo", "mv: cannot move 'foo' to 'bar/foo': No such file or directory"), true)
	assertMatch(t, "no_such_file", cmd("mv foo bar/", "mv: cannot move 'foo' to 'bar/': No such file or directory"), true)
	assertMatch(t, "no_such_file", cmd("mv foo bar/", ""), false)
	assertMatch(t, "no_such_file", cmd("mv foo bar/foo", "mv: permission denied"), false)
	assertNewCommand(t, "no_such_file",
		cmd("mv foo bar/foo", "mv: cannot move 'foo' to 'bar/foo': No such file or directory"),
		"mkdir -p bar && mv foo bar/foo")
	assertNewCommand(t, "no_such_file",
		cmd("mv foo bar/", "mv: cannot move 'foo' to 'bar/': No such file or directory"),
		"mkdir -p bar && mv foo bar/")
}

// ---- choco_install ----
func TestChocoInstall(t *testing.T) {
	cases := []struct{ before, after string }{
		{"choco install logstitcher", "choco install logstitcher.install"},
		{"cinst logstitcher", "cinst logstitcher.install"},
		{"choco install logstitcher -y", "choco install logstitcher.install -y"},
		{"cinst logstitcher -y", "cinst logstitcher.install -y"},
		{"choco install logstitcher -y -n=test", "choco install logstitcher.install -y -n=test"},
		{"cinst logstitcher -y -n=test", "cinst logstitcher.install -y -n=test"},
		{"choco install chocolatey -y", "choco install chocolatey.install -y"},
		{"cinst chocolatey -y", "cinst chocolatey.install -y"},
	}
	for _, tc := range cases {
		assertMatch(t, "choco_install", cmd(tc.before, ""), true)
		assertNewCommand(t, "choco_install", cmd(tc.before, ""), tc.after)
	}
	// not match
	for _, s := range []string{"choco /?", "choco upgrade logstitcher", "cup logstitcher"} {
		assertMatch(t, "choco_install", cmd(s, ""), false)
	}
}

// ---- sed_unterminated_s ----
func TestSedUnterminatedS(t *testing.T) {
	out := "sed: -e expression #1, char 9: unterminated `s' command"
	assertMatch(t, "sed_unterminated_s", cmd("sed -e s/foo/bar", out), true)
	assertMatch(t, "sed_unterminated_s", cmd("sed -es/foo/bar", out), true)
	assertMatch(t, "sed_unterminated_s", cmd("sed -e s/foo/bar -e s/baz/quz", out), true)
	assertMatch(t, "sed_unterminated_s", cmd("sed -e s/foo/bar", ""), false)
	assertNewCommand(t, "sed_unterminated_s", cmd("sed -e s/foo/bar", out), "sed -e s/foo/bar/")
	assertNewCommand(t, "sed_unterminated_s", cmd("sed -es/foo/bar", out), "sed -es/foo/bar/")
	assertNewCommand(t, "sed_unterminated_s", cmd("sed -e s/foo/bar -es/baz/quz", out),
		"sed -e s/foo/bar/ -es/baz/quz/")
}

// ---- ln_no_hard_link ----
func TestLnNoHardLink(t *testing.T) {
	err := "hard link not allowed for directory"
	for _, tc := range []struct{ script, out string }{
		{"ln barDir barLink", "ln: 'barDir': " + err},
		{"sudo ln a b", "ln: 'a': " + err},
		{"sudo ln -nbi a b", "ln: 'a': " + err},
	} {
		assertMatch(t, "ln_no_hard_link", cmd(tc.script, tc.out), true)
	}
	for _, tc := range []struct{ script, out string }{
		{"", ""},
		{"ln a b", "... hard link"},
		{"sudo ln a b", "... hard link"},
		{"a b", err},
	} {
		assertMatch(t, "ln_no_hard_link", cmd(tc.script, tc.out), false)
	}
	assertNewCommand(t, "ln_no_hard_link", cmd("ln barDir barLink", ""), "ln -s barDir barLink")
	assertNewCommand(t, "ln_no_hard_link", cmd("sudo ln barDir barLink", ""), "sudo ln -s barDir barLink")
	assertNewCommand(t, "ln_no_hard_link", cmd("sudo ln -nbi a b", ""), "sudo ln -s -nbi a b")
	assertNewCommand(t, "ln_no_hard_link", cmd("ln a ln", ""), "ln -s a ln")
	assertNewCommand(t, "ln_no_hard_link", cmd("sudo ln a ln", ""), "sudo ln -s a ln")
}

// ---- open ----
func TestOpen(t *testing.T) {
	mkOut := func(arg string) string { return "The file " + arg + " does not exist.\n" }
	urlScripts := []string{
		"open foo.com", "xdg-open foo.com", "gnome-open foo.com", "kde-open foo.com",
	}
	for _, s := range urlScripts {
		// extract the arg
		arg := s[len(s)-len("foo.com"):]
		assertMatch(t, "open", cmd(s, mkOut(arg)), true)
	}
	assertMatch(t, "open", cmd("open nonest", mkOut("nonest")), true)
	assertNewCommands(t, "open", cmd("open foo.io", mkOut("foo.io")), []string{"open http://foo.io"})
	assertNewCommands(t, "open", cmd("xdg-open foo.io", mkOut("foo.io")), []string{"xdg-open http://foo.io"})
	assertNewCommands(t, "open", cmd("open nonest", mkOut("nonest")),
		[]string{"touch nonest && open nonest", "mkdir nonest && open nonest"})
}

// ---- omnienv_no_such_command ----
func TestOmnienvNoSuchCommand(t *testing.T) {
	mkOut := func(c string) string { return "pyenv: no such command `" + c + "'" }
	for _, tc := range []struct{ script, c string }{
		{"pyenv globe", "globe"},
		{"pyenv intall 3.8.0", "intall"},
		{"pyenv list", "list"},
	} {
		assertMatch(t, "omnienv_no_such_command", cmd(tc.script, mkOut(tc.c)), true)
	}
	// goenv's quoted output
	assertMatch(t, "omnienv_no_such_command", cmd("goenv list", "goenv: no such command 'list'"), true)
	// not match
	assertMatch(t, "omnienv_no_such_command", cmd("pyenv global", "system"), false)
	// typo replacement — Python test uses `result in get_new_command(...)`
	assertNewCommandIn(t, "omnienv_no_such_command", cmd("pyenv list", mkOut("list")),
		"pyenv install --list")
	assertNewCommandIn(t, "omnienv_no_such_command", cmd("pyenv remove 3.8.0", mkOut("remove")),
		"pyenv uninstall 3.8.0")
}

// ---- cd_mkdir ----
func TestCdMkdir(t *testing.T) {
	assertMatch(t, "cd_mkdir", cmd("cd foo", "cd: foo: No such file or directory"), true)
	assertMatch(t, "cd_mkdir", cmd("cd foo/bar/baz", "cd: foo: No such file or directory"), true)
	assertMatch(t, "cd_mkdir", cmd("cd foo/bar/baz", "cd: can't cd to foo/bar/baz"), true)
	assertMatch(t, "cd_mkdir", cmd("cd /foo/bar/", `cd: The directory "/foo/bar/" does not exist`), true)
	assertMatch(t, "cd_mkdir", cmd("cd foo", ""), false)
	assertMatch(t, "cd_mkdir", cmd("", ""), false)
	assertNewCommand(t, "cd_mkdir", cmd("cd foo", ""), "mkdir -p foo && cd foo")
	assertNewCommand(t, "cd_mkdir", cmd("cd foo/bar/baz", ""), "mkdir -p foo/bar/baz && cd foo/bar/baz")
}

// ---- git_branch_0flag ----
func TestGitBranch0flag(t *testing.T) {
	branchExists := "fatal: A branch named 'bar' already exists."
	for _, s := range []string{
		"git branch 0a", "git branch 0d", "git branch 0f", "git branch 0r", "git branch 0v",
		"git branch 0d foo", "git branch 0D foo",
	} {
		assertMatch(t, "git_branch_0flag", cmd(s, branchExists), true)
	}
	for _, s := range []string{"git branch -a", "git branch -r", "git branch -v", "git branch -d foo", "git branch -D foo"} {
		assertMatch(t, "git_branch_0flag", cmd(s, ""), false)
	}
	for _, tc := range []struct{ script, want string }{
		{"git branch 0a", "git branch -D 0a && git branch -a"},
		{"git branch 0v", "git branch -D 0v && git branch -v"},
		{"git branch 0d foo", "git branch -D 0d && git branch -d foo"},
		{"git branch 0D foo", "git branch -D 0D && git branch -D foo"},
	} {
		assertNewCommand(t, "git_branch_0flag", cmd(tc.script, branchExists), tc.want)
	}
	// Not-valid-object output: no D-fixup
	notValid := "fatal: Not a valid object name: 'bar'."
	for _, tc := range []struct{ script, want string }{
		{"git branch 0l 'maint-*'", "git branch -l 'maint-*'"},
		{"git branch 0u upstream", "git branch -u upstream"},
	} {
		assertNewCommand(t, "git_branch_0flag", cmd(tc.script, notValid), tc.want)
	}
}

// ---- git_branch_delete ----
func TestGitBranchDelete(t *testing.T) {
	out := "error: The branch 'branch' is not fully merged.\nIf you are sure you want to delete it, run 'git branch -D branch'.\n\n"
	assertMatch(t, "git_branch_delete", cmd("git branch -d branch", out), true)
	assertMatch(t, "git_branch_delete", cmd("git branch -d branch", ""), false)
	assertMatch(t, "git_branch_delete", cmd("ls", out), false)
	assertNewCommand(t, "git_branch_delete", cmd("git branch -d branch", out), "git branch -D branch")
}

// ---- git_branch_delete_checked_out ----
func TestGitBranchDeleteCheckedOut(t *testing.T) {
	out := "error: Cannot delete branch 'foo' checked out at '/bar/foo'"
	assertMatch(t, "git_branch_delete_checked_out", cmd("git branch -d foo", out), true)
	assertMatch(t, "git_branch_delete_checked_out", cmd("git branch -D foo", out), true)
	assertMatch(t, "git_branch_delete_checked_out", cmd("git branch -d foo", "Deleted branch foo (was a1b2c3d)."), false)
	assertNewCommand(t, "git_branch_delete_checked_out", cmd("git branch -d foo", out), "git checkout master && git branch -D foo")
	assertNewCommand(t, "git_branch_delete_checked_out", cmd("git branch -D foo", out), "git checkout master && git branch -D foo")
}

// ---- git_branch_exists ----
func TestGitBranchExists(t *testing.T) {
	mkOut := func(n string) string { return "fatal: A branch named '" + n + "' already exists." }
	for _, s := range []string{"git branch foo", "git checkout bar"} {
		assertMatch(t, "git_branch_exists", cmd(s, mkOut("foo")), true)
		assertMatch(t, "git_branch_exists", cmd(s, ""), false)
	}
	assertNewCommands(t, "git_branch_exists", cmd("git branch foo", mkOut("foo")), []string{
		"git branch -d foo && git branch foo",
		"git branch -d foo && git checkout -b foo",
		"git branch -D foo && git branch foo",
		"git branch -D foo && git checkout -b foo",
		"git checkout foo",
	})
}

// ---- git_branch_list ----
func TestGitBranchList(t *testing.T) {
	assertMatch(t, "git_branch_list", cmd("git branch list", ""), true)
	assertMatch(t, "git_branch_list", cmd("", ""), false)
	assertMatch(t, "git_branch_list", cmd("git commit", ""), false)
	assertMatch(t, "git_branch_list", cmd("git branch", ""), false)
	assertMatch(t, "git_branch_list", cmd("git stash list", ""), false)
	assertNewCommand(t, "git_branch_list", cmd("git branch list", ""),
		"git branch --delete list && git branch")
}

// ---- git_fix_stash ----
func TestGitFixStash(t *testing.T) {
	err := "\nusage: git stash list [<options>]\n   or: git stash show [<stash>]\n   or: git stash ( pop | apply ) [--index] [-q|--quiet] [<stash>]\n"
	for _, s := range []string{"git stash opp", "git stash Some message", "git stash saev Some message"} {
		assertMatch(t, "git_fix_stash", cmd(s, err), true)
	}
	assertMatch(t, "git_fix_stash", cmd("git", err), false)
	assertNewCommand(t, "git_fix_stash", cmd("git stash opp", err), "git stash pop")
	assertNewCommand(t, "git_fix_stash", cmd("git stash Some message", err), "git stash save Some message")
	assertNewCommand(t, "git_fix_stash", cmd("git stash saev Some message", err), "git stash save Some message")
}

// ---- git_flag_after_filename ----
func TestGitFlagAfterFilename(t *testing.T) {
	cases := []struct{ script, out, want string }{
		{"git log README.md -p", "fatal: bad flag '-p' used after filename", "git log -p README.md"},
		{"git log README.md -p CONTRIBUTING.md", "fatal: bad flag '-p' used after filename", "git log -p README.md CONTRIBUTING.md"},
		{"git log -p README.md --name-only", "fatal: bad flag '--name-only' used after filename", "git log -p --name-only README.md"},
		{"git log README.md -p", "fatal: option '-p' must come before non-option arguments", "git log -p README.md"},
		{"git log README.md -p CONTRIBUTING.md", "fatal: option '-p' must come before non-option arguments", "git log -p README.md CONTRIBUTING.md"},
		{"git log -p README.md --name-only", "fatal: option '--name-only' must come before non-option arguments", "git log -p --name-only README.md"},
	}
	for _, tc := range cases {
		assertMatch(t, "git_flag_after_filename", cmd(tc.script, tc.out), true)
		assertNewCommand(t, "git_flag_after_filename", cmd(tc.script, tc.out), tc.want)
	}
	assertMatch(t, "git_flag_after_filename", cmd("git log README.md", ""), false)
	assertMatch(t, "git_flag_after_filename", cmd("git log -p README.md", ""), false)
}

// ---- git_bisect_usage ----
func TestGitBisectUsage(t *testing.T) {
	out := "usage: git bisect [help|start|bad|good|new|old|terms|skip|next|reset|visualize|replay|log|run]"
	for _, s := range []string{"git bisect strt", "git bisect rset", "git bisect goood"} {
		assertMatch(t, "git_bisect_usage", cmd(s, out), true)
	}
	for _, s := range []string{"git bisect", "git bisect start", "git bisect good"} {
		assertMatch(t, "git_bisect_usage", cmd(s, ""), false)
	}
	assertNewCommands(t, "git_bisect_usage", cmd("git bisect goood", out),
		[]string{"git bisect good", "git bisect old", "git bisect log"})
	assertNewCommands(t, "git_bisect_usage", cmd("git bisect strt", out),
		[]string{"git bisect start", "git bisect terms", "git bisect reset"})
	assertNewCommands(t, "git_bisect_usage", cmd("git bisect rset", out),
		[]string{"git bisect reset", "git bisect next", "git bisect start"})
}

// ---- git_push ----
func TestGitPush(t *testing.T) {
	mkOut := func(branch string) string {
		return "fatal: The current branch " + branch + " has no upstream branch.\nTo push the current branch and set the remote as upstream, use\n\n    git push --set-upstream origin " + branch + "\n\n"
	}
	master := mkOut("master")
	for _, s := range []string{"git push", "git push origin"} {
		assertMatch(t, "git_push", cmd(s, master), true)
	}
	assertMatch(t, "git_push", cmd("git push master", ""), false)
	assertMatch(t, "git_push", cmd("ls", master), false)
	cases := []struct{ script, want string }{
		{"git push", "git push --set-upstream origin master"},
		{"git push master", "git push --set-upstream origin master"},
		{"git push -u", "git push --set-upstream origin master"},
		{"git push -u origin", "git push --set-upstream origin master"},
		{"git push origin", "git push --set-upstream origin master"},
		{"git push --set-upstream origin", "git push --set-upstream origin master"},
		{"git push --quiet", "git push --set-upstream origin master --quiet"},
		{"git push --quiet origin", "git push --set-upstream origin master --quiet"},
		{"git -c test=test push --quiet origin", "git -c test=test push --set-upstream origin master --quiet"},
		{"git push --force", "git push --set-upstream origin master --force"},
		{"git push --force-with-lease", "git push --set-upstream origin master --force-with-lease"},
	}
	for _, tc := range cases {
		assertNewCommand(t, "git_push", cmd(tc.script, master), tc.want)
	}
}

// ---- git_push_force ----
func TestGitPushForce(t *testing.T) {
	err := "\nTo /tmp/foo\n ! [rejected]        master -> master (non-fast-forward)\n error: failed to push some refs to '/tmp/bar'\n hint: Updates were rejected because the tip of your current branch is behind\n"
	for _, s := range []string{"git push", "git push nvbn", "git push nvbn master"} {
		assertMatch(t, "git_push_force", cmd(s, err), true)
	}
	// Not matched outputs
	assertMatch(t, "git_push_force", cmd("git push", "Everything up-to-date"), false)
	assertNewCommand(t, "git_push_force", cmd("git push", err), "git push --force-with-lease")
	assertNewCommand(t, "git_push_force", cmd("git push nvbn", err), "git push --force-with-lease nvbn")
	assertNewCommand(t, "git_push_force", cmd("git push nvbn master", err), "git push --force-with-lease nvbn master")
}

// ---- git_push_pull ----
func TestGitPushPull(t *testing.T) {
	err := "\nTo /tmp/foo\n ! [rejected]        master -> master (non-fast-forward)\n error: failed to push some refs to '/tmp/bar'\n hint: Updates were rejected because the tip of your current branch is behind\n"
	err2 := "\nTo /tmp/foo\n ! [rejected]        master -> master (non-fast-forward)\n error: failed to push some refs to '/tmp/bar'\nhint: Updates were rejected because the remote contains work that you do\n"
	for _, out := range []string{err, err2} {
		for _, s := range []string{"git push", "git push nvbn", "git push nvbn master"} {
			assertMatch(t, "git_push_pull", cmd(s, out), true)
		}
	}
	assertMatch(t, "git_push_pull", cmd("git push", "Everything up-to-date"), false)
	assertNewCommand(t, "git_push_pull", cmd("git push", err), "git pull && git push")
	assertNewCommand(t, "git_push_pull", cmd("git push nvbn", err), "git pull nvbn && git push nvbn")
	assertNewCommand(t, "git_push_pull", cmd("git push nvbn master", err), "git pull nvbn master && git push nvbn master")
}

// ---- git_push_different_branch_names ----
func TestGitPushDifferentBranchNames(t *testing.T) {
	out := `fatal: The upstream branch of your current branch does not match
the name of your current branch.  To push to the upstream branch
on the remote, use

    git push origin HEAD:bar

To push to the branch of the same name on the remote, use

    git push origin foo

To choose either option permanently, see push.default in 'git help config'.
`
	assertMatch(t, "git_push_different_branch_names", cmd("git push", out), true)
	assertMatch(t, "git_push_different_branch_names", cmd("vim", ""), false)
	assertMatch(t, "git_push_different_branch_names", cmd("git status", out), false)
	assertMatch(t, "git_push_different_branch_names", cmd("git push", ""), false)
	assertNewCommand(t, "git_push_different_branch_names", cmd("git push", out), "git push origin HEAD:bar")
}

// ---- git_push_without_commits ----
func TestGitPushWithoutCommits(t *testing.T) {
	out := "error: src refspec master does not match any\nerror: failed to..."
	assertMatch(t, "git_push_without_commits", cmd("git push -u origin master", out), true)
	assertMatch(t, "git_push_without_commits", cmd("git push -u origin master", "Everything up-to-date"), false)
	assertNewCommand(t, "git_push_without_commits", cmd("git push -u origin master", out),
		`git commit -m "Initial commit" && git push -u origin master`)
}

// ---- git_clone_git_clone ----
func TestGitCloneGitClone(t *testing.T) {
	out := "\nfatal: Too many arguments.\n\nusage: git clone [<options>] [--] <repo> [<dir>]\n"
	assertMatch(t, "git_clone_git_clone", cmd("git clone git clone foo", out), true)
	assertMatch(t, "git_clone_git_clone", cmd("", ""), false)
	assertMatch(t, "git_clone_git_clone", cmd("git branch", ""), false)
	assertMatch(t, "git_clone_git_clone", cmd("git clone foo", ""), false)
	assertNewCommand(t, "git_clone_git_clone", cmd("git clone git clone foo", out), "git clone foo")
}

// ---- git_checkout ----
func TestGitCheckout(t *testing.T) {
	mkOut := func(t string) string { return "error: pathspec '" + t + "' did not match any file(s) known to git." }
	assertMatch(t, "git_checkout", cmd("git checkout unknown", mkOut("unknown")), true)
	assertMatch(t, "git_checkout", cmd("git commit unknown", mkOut("unknown")), true)
	assertMatch(t, "git_checkout", cmd("git submodule update unknown",
		mkOut("unknown")+"\nDid you forget to 'git add'?"), false)
	assertMatch(t, "git_checkout", cmd("git checkout known", ""), false)
	assertNewCommand(t, "git_checkout", cmd("git checkout unknown", mkOut("unknown")), "git checkout -b unknown")
	assertNewCommand(t, "git_checkout", cmd("git commit unknown", mkOut("unknown")),
		"git branch unknown && git commit unknown")
}

// ---- git_help_aliased ----
func TestGitHelpAliased(t *testing.T) {
	assertMatch(t, "git_help_aliased", cmd("git help st", "`git st' is aliased to `status'"), true)
	assertMatch(t, "git_help_aliased", cmd("git help ds", "`git ds' is aliased to `diff --staged'"), true)
	assertMatch(t, "git_help_aliased", cmd("git help status", "GIT-STATUS(1)...Git Manual...GIT-STATUS(1)"), false)
	assertNewCommand(t, "git_help_aliased", cmd("git help st", "`git st' is aliased to `status'"), "git help status")
	assertNewCommand(t, "git_help_aliased", cmd("git help ds", "`git ds' is aliased to `diff --staged'"), "git help diff")
}

// ---- git_lfs_mistype ----
func TestGitLfsMistype(t *testing.T) {
	out := "\nError: unknown command \"evn\" for \"git-lfs\"\n\nDid you mean this?\n\tenv\n\text\n\nRun 'git-lfs --help' for usage.\n    "
	assertMatch(t, "git_lfs_mistype", cmd("git lfs evn", out), true)
	assertMatch(t, "git_lfs_mistype", cmd("git lfs env", "bash: git: command not found"), false)
	assertMatch(t, "git_lfs_mistype", cmd("docker lfs env", out), false)
	assertNewCommands(t, "git_lfs_mistype", cmd("git lfs evn", out), []string{"git lfs env", "git lfs ext"})
}

// ---- fab_command_not_found ----
func TestFabCommandNotFound(t *testing.T) {
	out := "\nWarning: Command(s) not found:\n    extenson\n    deloyp\n\nAvailable commands:\n\n    update_config\n    prepare_extension\n    Template               A string class for supporting $-substitutions.\n    deploy\n    glob                   Return a list of paths matching a pathname pattern.\n    install_web\n    set_version\n"
	assertMatch(t, "fab_command_not_found", cmd("fab extenson", out), true)
	assertMatch(t, "fab_command_not_found", cmd("fab deloyp", out), true)
	assertMatch(t, "fab_command_not_found", cmd("fab extenson deloyp", out), true)
	assertMatch(t, "fab_command_not_found", cmd("gulp extenson", out), false)
	assertMatch(t, "fab_command_not_found", cmd("fab deloyp", ""), false)
	assertNewCommand(t, "fab_command_not_found", cmd("fab extenson", out), "fab prepare_extension")
	assertNewCommand(t, "fab_command_not_found", cmd("fab extenson:version=2016", out), "fab prepare_extension:version=2016")
	assertNewCommand(t, "fab_command_not_found",
		cmd("fab extenson:version=2016 install_web set_version:val=0.5.0", out),
		"fab prepare_extension:version=2016 install_web set_version:val=0.5.0")
}

// ---- npm_wrong_command ----
func TestNpmWrongCommand(t *testing.T) {
	out := `
Usage: npm <command>

where <command> is one of:
    access, add-user, adduser, apihelp, author, bin, bugs, c,
    cache, completion, config, ddp, dedupe, deprecate, dist-tag,
    dist-tags, docs, edit, explore, faq, find, find-dupes, get,
    help, help-search, home, i, info, init, install, issues, la,
    link, list, ll, ln, login, logout, ls, outdated, owner,
    pack, ping, prefix, prune, publish, r, rb, rebuild, remove,
    repo, restart, rm, root, run-script, s, se, search, set,
    show, shrinkwrap, star, stars, start, stop, t, tag, team,
    test, tst, un, uninstall, unlink, unpublish, unstar, up,
    update, upgrade, v, verison, version, view, whoami

npm <cmd> -h     quick help on <cmd>
`
	for _, s := range []string{"npm urgrdae", "npm urgrade -g", "npm -f urgrade -g", "npm urg"} {
		assertMatch(t, "npm_wrong_command", cmd(s, out), true)
	}
	assertMatch(t, "npm_wrong_command", cmd("npm urgrade", ""), false)
	// "npm" alone (no args) doesn't match
	assertMatch(t, "npm_wrong_command", cmd("npm", out), false)
	assertMatch(t, "npm_wrong_command", cmd("test urgrade", out), false)
	assertMatch(t, "npm_wrong_command", cmd("npm -e", out), false)
	assertNewCommand(t, "npm_wrong_command", cmd("npm urgrade", out), "npm upgrade")
}
