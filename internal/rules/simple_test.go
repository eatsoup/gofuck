package rules

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

var fmtSprintf = fmt.Sprintf

// ---- cd_parent ----
func TestCdParent(t *testing.T) {
	assertMatch(t, "cd_parent", cmd("cd..", "cd..: command not found"), true)
	assertMatch(t, "cd_parent", cmd("", ""), false)
	assertNewCommand(t, "cd_parent", cmd("cd..", ""), "cd ..")
}

// ---- sl_ls ----
func TestSlLs(t *testing.T) {
	assertMatch(t, "sl_ls", cmd("sl", ""), true)
	assertMatch(t, "sl_ls", cmd("ls", ""), false)
	assertNewCommand(t, "sl_ls", cmd("sl", ""), "ls")
}

// ---- dry ----
func TestDry(t *testing.T) {
	assertMatch(t, "dry", cmd("cd cd foo", ""), true)
	assertMatch(t, "dry", cmd("git git push origin/master", ""), true)
	assertNewCommand(t, "dry", cmd("cd cd foo", ""), "cd foo")
	assertNewCommand(t, "dry", cmd("git git push origin/master", ""), "git push origin/master")
}

// ---- quotation_marks ----
func TestQuotationMarks(t *testing.T) {
	cases := []struct{ in, want string }{
		{`git commit -m 'My Message"`, `git commit -m "My Message"`},
		{`git commit -am "Mismatched Quotation Marks'`, `git commit -am "Mismatched Quotation Marks"`},
		{`echo "hello'`, `echo "hello"`},
	}
	for _, tc := range cases {
		assertMatch(t, "quotation_marks", cmd(tc.in, ""), true)
		assertNewCommand(t, "quotation_marks", cmd(tc.in, ""), tc.want)
	}
}

// ---- remove_trailing_cedilla ----
func TestRemoveTrailingCedilla(t *testing.T) {
	const cedilla = "ç"
	assertMatch(t, "remove_trailing_cedilla", cmd("wrong"+cedilla, ""), true)
	assertMatch(t, "remove_trailing_cedilla", cmd("wrong with args"+cedilla, ""), true)
	assertNewCommand(t, "remove_trailing_cedilla", cmd("wrong"+cedilla, ""), "wrong")
	assertNewCommand(t, "remove_trailing_cedilla", cmd("wrong with args"+cedilla, ""), "wrong with args")
}

// ---- ag_literal ----
func TestAgLiteral(t *testing.T) {
	out := "ERR: Bad regex! pcre_compile() failed at position 1: missing )\nIf you meant to search for a literal string, run ag with -Q\n"
	assertMatch(t, "ag_literal", cmd(`ag \(`, out), true)
	assertMatch(t, "ag_literal", cmd("ag foo", ""), false)
	assertNewCommand(t, "ag_literal", cmd(`ag \(`, out), `ag -Q \(`)
}

// ---- cargo ----
func TestCargo(t *testing.T) {
	assertMatch(t, "cargo", cmd("cargo", ""), true)
	assertNewCommand(t, "cargo", cmd("cargo", ""), "cargo build")
}

// ---- cargo_no_command ----
func TestCargoNoCommand(t *testing.T) {
	oldOut := "No such subcommand\n\n        Did you mean `build`?\n"
	newOut := "error: no such subcommand\n\n\tDid you mean `build`?\n"
	assertMatch(t, "cargo_no_command", cmd("cargo buid", oldOut), true)
	assertMatch(t, "cargo_no_command", cmd("cargo buils", newOut), true)
	assertNewCommand(t, "cargo_no_command", cmd("cargo buid", oldOut), "cargo build")
	assertNewCommand(t, "cargo_no_command", cmd("cargo buils", newOut), "cargo build")
}

// ---- apt_get_search ----
func TestAptGetSearch(t *testing.T) {
	assertMatch(t, "apt_get_search", cmd("apt-get search foo", ""), true)
	for _, script := range []string{
		"apt-cache search foo", "aptitude search foo", "apt search foo",
		"apt-get install foo", "apt-get source foo", "apt-get clean",
		"apt-get remove", "apt-get update",
	} {
		assertMatch(t, "apt_get_search", cmd(script, ""), false)
	}
	assertNewCommand(t, "apt_get_search", cmd("apt-get search foo", ""), "apt-cache search foo")
}

// ---- apt_list_upgradable ----
func TestAptListUpgradable(t *testing.T) {
	fullEnglish := "\nHit:1 http://us.archive.ubuntu.com/ubuntu zesty InRelease\n" +
		"8 packages can be upgraded. Run 'apt list --upgradable' to see them.\n"
	german := "Führen Sie »apt list --upgradable« aus, um sie anzuzeigen."
	for _, out := range []string{fullEnglish, german} {
		assertMatch(t, "apt_list_upgradable", cmd("sudo apt update", out), true)
	}
	for _, tc := range []struct{ script, out string }{
		{"apt-cache search foo", ""},
		{"aptitude search foo", ""},
		{"apt search foo", ""},
		{"apt-get install foo", ""},
		{"apt-get source foo", ""},
		{"apt-get clean", ""},
		{"apt-get remove", ""},
		{"apt-get update", ""},
		{"sudo apt update", "All packages are up to date."},
	} {
		assertMatch(t, "apt_list_upgradable", cmd(tc.script, tc.out), false)
	}
	assertNewCommand(t, "apt_list_upgradable", cmd("sudo apt update", fullEnglish), "sudo apt list --upgradable")
	assertNewCommand(t, "apt_list_upgradable", cmd("apt update", fullEnglish), "apt list --upgradable")
}

// ---- apt_upgrade ----
func TestAptUpgrade(t *testing.T) {
	matchOut := "\nListing... Done\nheroku/stable 6.15.2-1 amd64 [upgradable from: 6.14.43-1]\n"
	noMatch := "\nListing... Done\n"
	assertMatch(t, "apt_upgrade", cmd("apt list --upgradable", matchOut), true)
	assertMatch(t, "apt_upgrade", cmd("sudo apt list --upgradable", matchOut), true)
	assertMatch(t, "apt_upgrade", cmd("apt list --upgradable", noMatch), false)
	assertMatch(t, "apt_upgrade", cmd("sudo apt list --upgradable", noMatch), false)
	assertNewCommand(t, "apt_upgrade", cmd("apt list --upgradable", matchOut), "apt upgrade")
	assertNewCommand(t, "apt_upgrade", cmd("sudo apt list --upgradable", matchOut), "sudo apt upgrade")
}

// ---- fix_alt_space ----
func TestFixAltSpace(t *testing.T) {
	assertMatch(t, "fix_alt_space", cmd("ps -ef | grep foo", "-bash:  grep: command not found"), true)
	assertMatch(t, "fix_alt_space", cmd("ps -ef | grep foo", ""), false)
	assertMatch(t, "fix_alt_space", cmd("", ""), false)
	assertNewCommand(t, "fix_alt_space", cmd("ps -ef | grep foo", "-bash:  grep: command not found"), "ps -ef | grep foo")
}

// ---- java ----
func TestJava(t *testing.T) {
	assertMatch(t, "java", cmd("java foo.java", ""), true)
	assertNewCommand(t, "java", cmd("java foo.java", ""), "java foo")
	assertNewCommand(t, "java", cmd("java bar.java", ""), "java bar")
}

// ---- javac ----
func TestJavac(t *testing.T) {
	assertMatch(t, "javac", cmd("javac foo", ""), true)
	assertNewCommand(t, "javac", cmd("javac foo", ""), "javac foo.java")
	assertNewCommand(t, "javac", cmd("javac bar", ""), "javac bar.java")
}

// ---- go_run ----
func TestGoRun(t *testing.T) {
	assertMatch(t, "go_run", cmd("go run foo", ""), true)
	assertNewCommand(t, "go_run", cmd("go run foo", ""), "go run foo.go")
	assertNewCommand(t, "go_run", cmd("go run bar", ""), "go run bar.go")
}

// ---- ls_lah ----
func TestLsLah(t *testing.T) {
	assertMatch(t, "ls_lah", cmd("ls", ""), true)
	assertMatch(t, "ls_lah", cmd("ls file.py", ""), true)
	assertMatch(t, "ls_lah", cmd("ls /opt", ""), true)
	assertMatch(t, "ls_lah", cmd("ls -lah /opt", ""), false)
	assertMatch(t, "ls_lah", cmd("pacman -S binutils", ""), false)
	assertMatch(t, "ls_lah", cmd("lsof", ""), false)
	assertNewCommand(t, "ls_lah", cmd("ls file.py", ""), "ls -lah file.py")
	assertNewCommand(t, "ls_lah", cmd("ls", ""), "ls -lah")
}

// ---- ls_all ----
func TestLsAll(t *testing.T) {
	assertMatch(t, "ls_all", cmd("ls", ""), true)
	assertMatch(t, "ls_all", cmd("ls", "file.py\n"), false)
	assertNewCommand(t, "ls_all", cmd("ls empty_dir", ""), "ls -A empty_dir")
	assertNewCommand(t, "ls_all", cmd("ls", ""), "ls -A")
}

// ---- man_no_space ----
func TestManNoSpace(t *testing.T) {
	assertMatch(t, "man_no_space", cmd("mandiff", "mandiff: command not found"), true)
	assertMatch(t, "man_no_space", cmd("", ""), false)
	assertNewCommand(t, "man_no_space", cmd("mandiff", ""), "man diff")
}

// ---- php_s ----
func TestPhpS(t *testing.T) {
	assertMatch(t, "php_s", cmd("php -s localhost:8000", ""), true)
	assertMatch(t, "php_s", cmd("php -t pub -s 0.0.0.0:8080", ""), true)
	assertMatch(t, "php_s", cmd("php -S localhost:8000", ""), false)
	assertMatch(t, "php_s", cmd("vim php -s", ""), false)
	assertNewCommand(t, "php_s", cmd("php -s localhost:8000", ""), "php -S localhost:8000")
	assertNewCommand(t, "php_s", cmd("php -t pub -s 0.0.0.0:8080", ""), "php -t pub -S 0.0.0.0:8080")
}

// ---- python_execute ----
func TestPythonExecute(t *testing.T) {
	assertMatch(t, "python_execute", cmd("python foo", ""), true)
	assertNewCommand(t, "python_execute", cmd("python foo", ""), "python foo.py")
	assertNewCommand(t, "python_execute", cmd("python bar", ""), "python bar.py")
}

// ---- python_command ----
func TestPythonCommand(t *testing.T) {
	assertMatch(t, "python_command", cmd("temp.py", "Permission denied"), true)
	assertMatch(t, "python_command", cmd("", ""), false)
	assertNewCommand(t, "python_command", cmd("./test_sudo.py", ""), "python ./test_sudo.py")
}

// ---- python_module_error ----
func TestPythonModuleError(t *testing.T) {
	tmpl := "Traceback (most recent call last):\n  File \"%s\", line 1, in <module>\n    import %s\nModuleNotFoundError: No module named '%s'"
	cases := []struct{ script, filename, mod, want string }{
		{"python some_script.py", "some_script.py", "more_itertools", "pip install more_itertools && python some_script.py"},
		{"./some_other_script.py", "some_other_script.py", "a_module", "pip install a_module && ./some_other_script.py"},
	}
	for _, tc := range cases {
		out := fmtSprintf(tmpl, tc.filename, tc.mod, tc.mod)
		assertMatch(t, "python_module_error", cmd(tc.script, out), true)
		assertNewCommand(t, "python_module_error", cmd(tc.script, out), tc.want)
	}
	assertMatch(t, "python_module_error", cmd("python hello_world.py", "Hello World"), false)
}

// ---- remove_shell_prompt_literal ----
func TestRemoveShellPromptLiteral(t *testing.T) {
	out := "$: command not found"
	for _, s := range []string{"$ cd newdir", " $ cd newdir", "$ $ cd newdir", " $ $ cd newdir"} {
		assertMatch(t, "remove_shell_prompt_literal", cmd(s, out), true)
	}
	for _, tc := range []struct{ script, out string }{
		{"$", "$: command not found"},
		{" $", "$: command not found"},
		{"$?", "127: command not found"},
		{" $?", "127: command not found"},
		{"", ""},
	} {
		assertMatch(t, "remove_shell_prompt_literal", cmd(tc.script, tc.out), false)
	}
	assertNewCommand(t, "remove_shell_prompt_literal", cmd("$ cd newdir", out), "cd newdir")
	assertNewCommand(t, "remove_shell_prompt_literal", cmd("$ $ cd newdir", out), "cd newdir")
	assertNewCommand(t, "remove_shell_prompt_literal", cmd("$ python3 -m virtualenv env", out), "python3 -m virtualenv env")
	assertNewCommand(t, "remove_shell_prompt_literal", cmd(" $ $ $ python3 -m virtualenv env", out), "python3 -m virtualenv env")
}

// ---- django_south_ghost ----
func TestDjangoSouthGhost(t *testing.T) {
	out := "south.exceptions.GhostMigrations:\n ! with the south_migrationhistory table, or pass --delete-ghost-migrations\n"
	assertMatch(t, "django_south_ghost", cmd("./manage.py migrate", out), true)
	assertMatch(t, "django_south_ghost", cmd("python manage.py migrate", out), true)
	assertMatch(t, "django_south_ghost", cmd("./manage.py migrate", ""), false)
	assertMatch(t, "django_south_ghost", cmd("app migrate", out), false)
	assertMatch(t, "django_south_ghost", cmd("./manage.py test", out), false)
	assertNewCommand(t, "django_south_ghost", cmd("./manage.py migrate auth", ""),
		"./manage.py migrate auth --delete-ghost-migrations")
}

// ---- django_south_merge ----
func TestDjangoSouthMerge(t *testing.T) {
	out := "    --merge: will just attempt the migration ignoring conflicts.\n"
	assertMatch(t, "django_south_merge", cmd("./manage.py migrate", out), true)
	assertMatch(t, "django_south_merge", cmd("python manage.py migrate", out), true)
	assertMatch(t, "django_south_merge", cmd("./manage.py migrate", ""), false)
	assertMatch(t, "django_south_merge", cmd("app migrate", out), false)
	assertMatch(t, "django_south_merge", cmd("./manage.py test", out), false)
	assertNewCommand(t, "django_south_merge", cmd("./manage.py migrate auth", ""), "./manage.py migrate auth --merge")
}

// ---- unsudo ----
func TestUnsudo(t *testing.T) {
	assertMatch(t, "unsudo", cmd("sudo ls", "you cannot perform this operation as root"), true)
	assertMatch(t, "unsudo", cmd("", ""), false)
	assertMatch(t, "unsudo", cmd("sudo ls", "Permission denied"), false)
	assertMatch(t, "unsudo", cmd("ls", "you cannot perform this operation as root"), false)
	assertNewCommand(t, "unsudo", cmd("sudo ls", ""), "ls")
	assertNewCommand(t, "unsudo", cmd("sudo pacaur -S helloworld", ""), "pacaur -S helloworld")
}

// ---- conda_mistype ----
func TestCondaMistype(t *testing.T) {
	out := "\n\nCommandNotFoundError: No command 'conda lst'.\nDid you mean 'conda list'?\n\n    "
	assertMatch(t, "conda_mistype", cmd("conda lst", out), true)
	assertMatch(t, "conda_mistype", cmd("codna list", "bash: codna: command not found"), false)
	assertNewCommands(t, "conda_mistype", cmd("conda lst", out), []string{"conda list"})
}

// ---- mkdir_p ----
func TestMkdirP(t *testing.T) {
	assertMatch(t, "mkdir_p", cmd("mkdir foo/bar/baz", "mkdir: foo/bar: No such file or directory"), true)
	assertMatch(t, "mkdir_p", cmd("./bin/hdfs dfs -mkdir foo/bar/baz", "mkdir: `foo/bar/baz': No such file or directory"), true)
	assertMatch(t, "mkdir_p", cmd("hdfs dfs -mkdir foo/bar/baz", "mkdir: `foo/bar/baz': No such file or directory"), true)
	assertMatch(t, "mkdir_p", cmd("mkdir foo/bar/baz", ""), false)
	assertMatch(t, "mkdir_p", cmd("mkdir foo/bar/baz", "foo bar baz"), false)
	assertMatch(t, "mkdir_p", cmd("hdfs dfs -mkdir foo/bar/baz", ""), false)
	assertMatch(t, "mkdir_p", cmd("./bin/hdfs dfs -mkdir foo/bar/baz", ""), false)
	assertMatch(t, "mkdir_p", cmd("", ""), false)
	assertNewCommand(t, "mkdir_p", cmd("mkdir foo/bar/baz", ""), "mkdir -p foo/bar/baz")
	assertNewCommand(t, "mkdir_p", cmd("hdfs dfs -mkdir foo/bar/baz", ""), "hdfs dfs -mkdir -p foo/bar/baz")
	assertNewCommand(t, "mkdir_p", cmd("./bin/hdfs dfs -mkdir foo/bar/baz", ""), "./bin/hdfs dfs -mkdir -p foo/bar/baz")
}

// ---- rm_dir ----
func TestRmDir(t *testing.T) {
	assertMatch(t, "rm_dir", cmd("rm foo", "rm: foo: is a directory"), true)
	assertMatch(t, "rm_dir", cmd("rm foo", "rm: foo: Is a directory"), true)
	assertMatch(t, "rm_dir", cmd("hdfs dfs -rm foo", "rm: `foo`: Is a directory"), true)
	assertMatch(t, "rm_dir", cmd("./bin/hdfs dfs -rm foo", "rm: `foo`: Is a directory"), true)
	assertMatch(t, "rm_dir", cmd("rm foo", ""), false)
	assertMatch(t, "rm_dir", cmd("hdfs dfs -rm foo", ""), false)
	assertMatch(t, "rm_dir", cmd("./bin/hdfs dfs -rm foo", ""), false)
	assertMatch(t, "rm_dir", cmd("", ""), false)
	assertNewCommand(t, "rm_dir", cmd("rm foo", ""), "rm -rf foo")
	assertNewCommand(t, "rm_dir", cmd("hdfs dfs -rm foo", ""), "hdfs dfs -rm -r foo")
}

// ---- rm_root ----
func TestRmRoot(t *testing.T) {
	assertMatch(t, "rm_root", cmd("rm -rf /", "add --no-preserve-root"), true)
	assertMatch(t, "rm_root", cmd("ls", "add --no-preserve-root"), false)
	assertMatch(t, "rm_root", cmd("rm --no-preserve-root /", "add --no-preserve-root"), false)
	assertMatch(t, "rm_root", cmd("rm -rf /", ""), false)
	assertNewCommand(t, "rm_root", cmd("rm -rf /", ""), "rm -rf / --no-preserve-root")
}

// ---- whois ----
func TestWhois(t *testing.T) {
	assertMatch(t, "whois", cmd("whois https://en.wikipedia.org/wiki/Main_Page", ""), true)
	assertMatch(t, "whois", cmd("whois https://en.wikipedia.org/", ""), true)
	assertMatch(t, "whois", cmd("whois meta.unix.stackexchange.com", ""), true)
	assertMatch(t, "whois", cmd("whois", ""), false)
	assertNewCommand(t, "whois", cmd("whois https://en.wikipedia.org/wiki/Main_Page", ""), "whois en.wikipedia.org")
	assertNewCommand(t, "whois", cmd("whois https://en.wikipedia.org/", ""), "whois en.wikipedia.org")
	assertNewCommands(t, "whois", cmd("whois meta.unix.stackexchange.com", ""),
		[]string{"whois unix.stackexchange.com", "whois stackexchange.com", "whois com"})
}

// ---- cat_dir ----
func TestCatDir(t *testing.T) {
	withTmpDir(t)
	mkDir(t, "foo")
	mkDir(t, "cat")
	assertMatch(t, "cat_dir", cmd("cat foo", "cat: foo: Is a directory\n"), true)
	assertMatch(t, "cat_dir", cmd("cat cat", "cat: cat: Is a directory\n"), true)
	
	assertMatch(t, "cat_dir", cmd("cat foo", "foo bar baz"), false)
	assertMatch(t, "cat_dir", cmd("cat foo bar", "foo bar baz"), false)
	assertMatch(t, "cat_dir", cmd("notcat foo bar", "some output"), false)
	
	assertNewCommand(t, "cat_dir", cmd("cat foo", "cat: foo: Is a directory\n"), "ls foo")
	assertNewCommand(t, "cat_dir", cmd("cat cat", "cat: cat: Is a directory\n"), "ls cat")
}

// ---- cpp11 ----
func TestCpp11(t *testing.T) {
	out1 := "This file requires compiler and library support for the ISO C++ 2011 standard."
	out2 := "warning: 'auto' type specifier is a C++11 extension [-Wc++11-extensions]"
	
	assertMatch(t, "cpp11", cmd("g++ main.cpp", out1), true)
	assertMatch(t, "cpp11", cmd("clang++ main.cpp", out1), true)
	assertMatch(t, "cpp11", cmd("g++ main.cpp", out2), true)
	assertMatch(t, "cpp11", cmd("clang++ main.cpp", out2), true)
	
	assertMatch(t, "cpp11", cmd("g++ main.cpp", ""), false)
	assertMatch(t, "cpp11", cmd("clang++ main.cpp", ""), false)
	assertMatch(t, "cpp11", cmd("gcc main.cpp", out1), false)
	
	assertNewCommand(t, "cpp11", cmd("g++ main.cpp", out1), "g++ main.cpp -std=c++11")
	assertNewCommand(t, "cpp11", cmd("clang++ main.cpp", out2), "clang++ main.cpp -std=c++11")
}

// ---- chmod_x ----
func TestChmodX(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("execute bit operations are unreliable on windows")
	}
	withTmpDir(t)
	touchFile(t, "script")
	touchFile(t, "script2")
	_ = os.Chmod("script2", 0755) // executable
	
	assertMatch(t, "chmod_x", cmd("./script", "permission denied"), true)
	assertMatch(t, "chmod_x", cmd("./script", "Permission denied"), true)
	
	assertMatch(t, "chmod_x", cmd("script", "permission denied"), false)
	assertMatch(t, "chmod_x", cmd("./script", ""), false)
	assertMatch(t, "chmod_x", cmd("./script2", "permission denied"), false) // already executable
	assertMatch(t, "chmod_x", cmd("./missing", "permission denied"), false)
	
	assertNewCommand(t, "chmod_x", cmd("./script", "permission denied"), "chmod +x script && ./script")
}

// ---- has_exists_script ----
func TestHasExistsScript(t *testing.T) {
	withTmpDir(t)
	touchFile(t, "main")
	
	assertMatch(t, "has_exists_script", cmd("main", "main: command not found"), true)
	assertMatch(t, "has_exists_script", cmd("main --help", "main: command not found"), true)
	
	assertMatch(t, "has_exists_script", cmd("main", ""), false)
	assertMatch(t, "has_exists_script", cmd("missing", "missing: command not found"), false)
	
	assertNewCommand(t, "has_exists_script", cmd("main", "main: command not found"), "./main")
	assertNewCommand(t, "has_exists_script", cmd("main --help", "main: command not found"), "./main --help")
}
