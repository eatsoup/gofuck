package rules

import (
	"reflect"
	"testing"
)

// ---- cpp11 (S3.4) ----
//
// Upstream has no `test_cpp11.py`. We assert the Go behavior directly: match
// only on g++/clang++ when the output mentions C++11 support, and append
// `-std=c++11` to the script.

func TestCpp11Match(t *testing.T) {
	const wantErr = "error: This file requires compiler and library support for the ISO C++ 2011 standard."
	cases := []struct {
		script, output string
		want           bool
	}{
		{"g++ a.cpp", wantErr, true},
		{"clang++ a.cpp", "warning: -Wc++11-extensions", true},
		{"gcc a.c", wantErr, false},
		{"g++ a.cpp", "unrelated", false},
	}
	for _, tc := range cases {
		assertMatch(t, "cpp11", cmd(tc.script, tc.output), tc.want)
	}
}

func TestCpp11NewCommand(t *testing.T) {
	const out = "error: This file requires compiler and library support for the ISO C++ 2011 standard."
	assertNewCommand(t, "cpp11", cmd("g++ a.cpp", out), "g++ a.cpp -std=c++11")
	assertNewCommand(t, "cpp11", cmd("clang++ a.cpp", out), "clang++ a.cpp -std=c++11")
}

// ---- git_pull (S3.10) ----

const gitPullSetUpstreamOutput = "There is no tracking information for the current branch.\n" +
	"Please specify which branch you want to merge with.\n" +
	"See git-pull(1) for details\n" +
	"\n" +
	"    git pull <remote> <branch>\n" +
	"\n" +
	"If you wish to set tracking information for this branch you can do so with:\n" +
	"\n" +
	"    git branch --set-upstream-to=<remote>/<branch> master\n" +
	"\n"

func TestGitPullMatch(t *testing.T) {
	assertMatch(t, "git_pull", cmd("git pull", gitPullSetUpstreamOutput), true)
	assertMatch(t, "git_pull", cmd("git pull", ""), false)
	assertMatch(t, "git_pull", cmd("ls", gitPullSetUpstreamOutput), false)
}

func TestGitPullNewCommand(t *testing.T) {
	assertNewCommand(t, "git_pull",
		cmd("git pull", gitPullSetUpstreamOutput),
		"git branch --set-upstream-to=origin/master master && git pull",
	)
}

// ---- brew_unknown_command (S3.22) ----

func TestBrewUnknownCommandMatch(t *testing.T) {
	const inst = "Error: Unknown command: inst"
	assertMatch(t, "brew_unknown_command", cmd("brew inst", inst), true)
	for _, command := range brewCmds {
		assertMatch(t, "brew_unknown_command", cmd("brew "+command, ""), false)
	}
}

func TestBrewUnknownCommandNewCommand(t *testing.T) {
	got := mustRule(t, "brew_unknown_command").GetNewCommand(cmd("brew inst", "Error: Unknown command: inst"))
	want := []string{"brew list", "brew install", "brew uninstall"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("brew_unknown_command(brew inst) = %v, want %v", got, want)
	}

	cmds := mustRule(t, "brew_unknown_command").GetNewCommand(cmd("brew instaa", "Error: Unknown command: instaa"))
	if !contains(cmds, "brew install") || !contains(cmds, "brew uninstall") {
		t.Errorf("brew_unknown_command(brew instaa) = %v, want it to contain install+uninstall", cmds)
	}
}
