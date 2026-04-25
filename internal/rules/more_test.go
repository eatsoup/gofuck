package rules

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

// ---- gradle_wrapper ----
func TestGradleWrapper(t *testing.T) {
	withTmpDir(t)
	withPath(t, "") // clear PATH so 'gradle' is not found
	touchFile(t, "gradlew")
	
	out := "gradle: not found"
	
	assertMatch(t, "gradle_wrapper", cmd("gradle tasks", out), true)
	
	// negative: no gradlew file
	_ = os.Remove("gradlew")
	assertMatch(t, "gradle_wrapper", cmd("gradle tasks", out), false)
	
	// negative: gradlew exists, but command not gradle
	touchFile(t, "gradlew")
	assertMatch(t, "gradle_wrapper", cmd("npm tasks", out), false)
	
	// GetNewCommand
	assertNewCommand(t, "gradle_wrapper", cmd("gradle tasks", out), "./gradlew tasks")
	assertNewCommand(t, "gradle_wrapper", cmd("gradle build -c", out), "./gradlew build -c")
}

// ---- grep_arguments_order ----
func TestGrepArgumentsOrder(t *testing.T) {
	withTmpDir(t)
	touchFile(t, "testfile")
	
	out := func(f string) string {
		return fmt.Sprintf("grep: %s: No such file or directory", f)
	}
	
	assertMatch(t, "grep_arguments_order", cmd("grep testfile test", out("test")), true)
	assertMatch(t, "grep_arguments_order", cmd("grep -lir . test", out("test")), true)
	assertMatch(t, "grep_arguments_order", cmd("egrep testfile test", out("test")), true)
	
	assertMatch(t, "grep_arguments_order", cmd("cat testfile", out("test")), false)
	assertMatch(t, "grep_arguments_order", cmd("grep test testfile", ""), false)
	
	assertNewCommand(t, "grep_arguments_order", cmd("grep testfile test", out("test")), "grep test testfile")
	assertNewCommand(t, "grep_arguments_order", cmd("grep -lir . test", out("test")), "grep -lir test .")
	assertNewCommand(t, "grep_arguments_order", cmd("egrep testfile test", out("test")), "egrep test testfile")
	assertNewCommand(t, "grep_arguments_order", cmd("grep . test -lir", out("test")), "grep test -lir .")
}

// ---- fix_file ----
func TestFixFile(t *testing.T) {
	withTmpDir(t)
	touchFile(t, "a.c")
	
	outGCC := "a.c: In function 'main':\na.c:3:1: error: expected expression before '}' token"
	
	t.Setenv("EDITOR", "dummy_editor")
	assertMatch(t, "fix_file", cmd("gcc a.c", outGCC), true)
	
	t.Setenv("EDITOR", "")
	assertMatch(t, "fix_file", cmd("gcc a.c", outGCC), false)
	
	t.Setenv("EDITOR", "dummy_editor")
	assertMatch(t, "fix_file", cmd("gcc missing.c", "missing.c:3: error..."), false)
	
	assertNewCommand(t, "fix_file", cmd("gcc a.c", outGCC), "dummy_editor a.c +3 && gcc a.c")
}

// ---- ln_s_order ----
func TestLnSOrder(t *testing.T) {
	withTmpDir(t)
	touchFile(t, "source")
	touchFile(t, "dest") // Need this to mirror upstream's blanket os.path.exists return_value=True!
	
	out := func(f string) string {
		return fmt.Sprintf("ln: failed to create symbolic link '%s': File exists", f)
	}
	
	assertMatch(t, "ln_s_order", cmd("ln -s dest source", out("source")), true)
	assertMatch(t, "ln_s_order", cmd("ln dest -s source", out("source")), true)
	assertMatch(t, "ln_s_order", cmd("ln dest source -s", out("source")), true)
	
	assertMatch(t, "ln_s_order", cmd("ln dest source", out("source")), false)
	assertMatch(t, "ln_s_order", cmd("ls -s dest source", out("source")), false)
	assertMatch(t, "ln_s_order", cmd("ln -s dest source", ""), false)
	
	// if file doesn't exist
	_ = os.Remove("source")
	_ = os.Remove("dest")
	assertMatch(t, "ln_s_order", cmd("ln -s dest source", out("source")), false)
	
	touchFile(t, "source")
	touchFile(t, "dest")
	
	assertNewCommand(t, "ln_s_order", cmd("ln -s dest source", out("source")), "ln -s source dest")
	assertNewCommand(t, "ln_s_order", cmd("ln dest -s source", out("source")), "ln -s source dest")
	assertNewCommand(t, "ln_s_order", cmd("ln dest source -s", out("source")), "ln source -s dest")
}

// ---- missing_space_before_subcommand ----
func TestMissingSpaceBeforeSubcommand(t *testing.T) {
	assertMatch(t, "missing_space_before_subcommand", cmd("gitbranch", ""), true)
	assertMatch(t, "missing_space_before_subcommand", cmd("npminstall", ""), true)
	
	assertMatch(t, "missing_space_before_subcommand", cmd("git branch", ""), false)
	assertMatch(t, "missing_space_before_subcommand", cmd("vimfile", ""), false)
	
	assertNewCommand(t, "missing_space_before_subcommand", cmd("gitbranch", ""), "git branch")
	assertNewCommand(t, "missing_space_before_subcommand", cmd("npminstall webpack", ""), "npm install webpack")
}

// ---- no_command ----
func TestNoCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exec.LookPath behaves differently on windows regarding extensions")
	}
	withTmpDir(t)
	withPath(t, "bin")
	mkDir(t, "bin")
	touchFile(t, "bin/git")
	
	out1 := "git: command not found"
	out2 := "gti is not recognized as an internal or external command"
	
	assertMatch(t, "no_command", cmd("gti", out1), true)
	assertMatch(t, "no_command", cmd("gti status", out2), true)
	
	assertMatch(t, "no_command", cmd("gti", ""), false)
	assertMatch(t, "no_command", cmd("git", "git: command not found"), false) 
	
	assertNewCommand(t, "no_command", cmd("gti", out2), "git")
	assertNewCommand(t, "no_command", cmd("gti status", out1), "git status")
}

// ---- prove_recursively ----
func TestProveRecursively(t *testing.T) {
	withTmpDir(t)
	mkDir(t, "t")

	out := "Files=0, Tests=0,  0 wallclock secs ( 0.00 usr +  0.00 sys =  0.00 CPU)\nResult: NOTESTS"

	// match: directory arg, no -r flag
	assertMatch(t, "prove_recursively", cmd("prove -lv t", out), true)
	assertMatch(t, "prove_recursively", cmd("prove t", out), true)

	// not match: directory but already has -r or --recurse
	assertMatch(t, "prove_recursively", cmd("prove -r t", out), false)
	assertMatch(t, "prove_recursively", cmd("prove --recurse t", out), false)

	// not match: arg is not a directory
	assertMatch(t, "prove_recursively", cmd("prove -lv nonexistent", out), false)

	// GetNewCommand
	assertNewCommand(t, "prove_recursively", cmd("prove -lv t", out), "prove -r -lv t")
	assertNewCommand(t, "prove_recursively", cmd("prove t", out), "prove -r t")
}

// ---- scm_correction ----
func TestScmCorrection(t *testing.T) {
	withTmpDir(t)

	gitErr := "fatal: Not a git repository (or any of the parent directories): .git"
	hgErr := "abort: no repository found in '/home/nvbn/exp/thefuck' (.hg not found)!"

	// Simulate a mercurial repo (has .hg), user runs "git log"
	mkDir(t, ".hg")
	assertMatch(t, "scm_correction", cmd("git log", gitErr), true)
	assertNewCommand(t, "scm_correction", cmd("git log", gitErr), "hg log")

	// Clean up and simulate a git repo (has .git), user runs "hg log"
	_ = os.RemoveAll(".hg")
	mkDir(t, ".git")
	assertMatch(t, "scm_correction", cmd("hg log", hgErr), true)
	assertNewCommand(t, "scm_correction", cmd("hg log", hgErr), "git log")

	// Not match: correct scm
	assertMatch(t, "scm_correction", cmd("git log", ""), false)

	// Not match: no scm directory at all
	_ = os.RemoveAll(".git")
	assertMatch(t, "scm_correction", cmd("git log", gitErr), false)
	assertMatch(t, "scm_correction", cmd("hg log", hgErr), false)

	// Not match: not a scm tool
	mkDir(t, ".git")
	assertMatch(t, "scm_correction", cmd("not-scm log", hgErr), false)
}

// ---- sudo_command_from_user_path ----
func TestSudoCommandFromUserPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sudo is not available on Windows")
	}
	withTmpDir(t)
	mkDir(t, "bin")
	touchFile(t, "bin/npm")
	// Make the file executable
	_ = os.Chmod("bin/npm", 0755)
	withPath(t, "bin")

	outFmt := "sudo: %s: command not found"

	// match: sudo + command in PATH + "command not found"
	assertMatch(t, "sudo_command_from_user_path", cmd("sudo npm install -g react-native-cli", fmt.Sprintf(outFmt, "npm")), true)

	// not match: not sudo
	assertMatch(t, "sudo_command_from_user_path", cmd("npm --version", fmt.Sprintf(outFmt, "npm")), false)

	// not match: no "command not found" in output
	assertMatch(t, "sudo_command_from_user_path", cmd("sudo npm --version", ""), false)

	// not match: command not in PATH
	assertMatch(t, "sudo_command_from_user_path", cmd("sudo foobar install", fmt.Sprintf(outFmt, "foobar")), false)
}

// ---- wrong_hyphen_before_subcommand ----
func TestWrongHyphenBeforeSubcommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exec.LookPath behaves differently on windows regarding extensions")
	}
	withTmpDir(t)
	mkDir(t, "bin")
	touchFile(t, "bin/git")
	touchFile(t, "bin/apt")
	touchFile(t, "bin/apt-get")
	_ = os.Chmod("bin/git", 0755)
	_ = os.Chmod("bin/apt", 0755)
	_ = os.Chmod("bin/apt-get", 0755)
	withPath(t, "bin")

	// match: git-log → git + log
	assertMatch(t, "wrong_hyphen_before_subcommand", cmd("git-log", ""), true)
	assertMatch(t, "wrong_hyphen_before_subcommand", cmd("apt-install python", ""), true)

	// not match: already a valid executable
	assertMatch(t, "wrong_hyphen_before_subcommand", cmd("apt-get install python", ""), false)

	// not match: no hyphen
	assertMatch(t, "wrong_hyphen_before_subcommand", cmd("git log", ""), false)

	// GetNewCommand
	assertNewCommand(t, "wrong_hyphen_before_subcommand", cmd("git-log", ""), "git log")
	assertNewCommand(t, "wrong_hyphen_before_subcommand", cmd("apt-install python", ""), "apt install python")
}

// ---- brew_unknown_command ----
func TestBrewUnknownCommand(t *testing.T) {
	out1 := "Error: Unknown command: inst"
	out2 := "Error: Unknown command: instaa"

	assertMatch(t, "brew_unknown_command", cmd("brew inst", out1), true)

	// not match: known command
	assertMatch(t, "brew_unknown_command", cmd("brew install", ""), false)
	assertMatch(t, "brew_unknown_command", cmd("brew list", ""), false)

	// GetNewCommand: "inst" is close to "list", "install", "uninstall"
	cmds := getNewCommands(t, "brew_unknown_command", cmd("brew inst", out1))
	assertContains(t, cmds, "brew list")
	assertContains(t, cmds, "brew install")
	assertContains(t, cmds, "brew uninstall")

	cmds2 := getNewCommands(t, "brew_unknown_command", cmd("brew instaa", out2))
	assertContains(t, cmds2, "brew install")
	assertContains(t, cmds2, "brew uninstall")
}

// ---- dirty_untar ----
func TestDirtyUntar(t *testing.T) {
	assertMatch(t, "dirty_untar", cmd("tar xvf foo.tar", ""), true)
	assertMatch(t, "dirty_untar", cmd("tar xvf foo.tar.gz", ""), true)
	assertMatch(t, "dirty_untar", cmd("tar --extract -f foo.tar.bz2", ""), true)

	// not match: has -C flag
	assertMatch(t, "dirty_untar", cmd("tar xvf foo.tar -C /tmp", ""), false)
	// not match: not extract
	assertMatch(t, "dirty_untar", cmd("tar cvf foo.tar file1", ""), false)
	// not match: not tar
	assertMatch(t, "dirty_untar", cmd("unzip foo.zip", ""), false)

	assertNewCommand(t, "dirty_untar", cmd("tar xvf foo.tar.gz", ""), "mkdir -p foo && tar xvf foo.tar.gz -C foo")
}

// ---- dirty_unzip ----
func TestDirtyUnzip(t *testing.T) {
	assertMatch(t, "dirty_unzip", cmd("unzip foo.zip", ""), true)
	assertMatch(t, "dirty_unzip", cmd("unzip foo", ""), true)

	// not match: has -d flag already
	assertMatch(t, "dirty_unzip", cmd("unzip foo.zip -d bar", ""), false)
	// not match: not unzip
	assertMatch(t, "dirty_unzip", cmd("tar xvf foo.tar", ""), false)

	assertNewCommand(t, "dirty_unzip", cmd("unzip foo.zip", ""), "unzip foo.zip -d foo")
}
