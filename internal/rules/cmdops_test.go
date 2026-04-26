package rules

import (
	"reflect"
	"testing"

	specexec "github.com/eatsoup/gofuck/internal/specific/exec"
)

// withMockRunner swaps the exec runner for a single block. Returns a function
// to call (deferred or invoked manually) that restores the previous runner.
func mockRunner(t *testing.T, byCmd map[string]specexec.Result) func() {
	t.Helper()
	return specexec.WithRunner(func(name string, args ...string) specexec.Result {
		key := name
		for _, a := range args {
			key += " " + a
		}
		if r, ok := byCmd[key]; ok {
			return r
		}
		// Default: tool not mocked → simulate "not installed".
		return specexec.Result{Err: errNotMocked}
	})
}

var errNotMocked = &mockMissErr{}

type mockMissErr struct{}

func (*mockMissErr) Error() string { return "exec: not mocked" }

func TestParseAptHelp(t *testing.T) {
	help := []string{
		"apt 2.4.5 (amd64)",
		"Usage: apt [options] command",
		"",
		"Most used commands:",
		"  list - list packages based on package names",
		"  search - search in package descriptions",
		"  install - install packages",
		"  remove - remove packages",
		"",
		"See apt(8) for more information.",
	}
	got := parseAptHelp(help)
	want := []string{"list", "search", "install", "remove", "See"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseAptHelp = %v, want %v", got, want)
	}
}

func TestParseAptGetHelp(t *testing.T) {
	help := []string{
		"apt-get 2.4.5",
		"Usage: apt-get …",
		"Commands:",
		"  update - retrieve new lists of packages",
		"  upgrade - perform an upgrade",
		"  install - install new packages",
		"",
		"trailing junk",
	}
	got := parseAptGetHelp(help)
	want := []string{"update", "upgrade", "install"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseAptGetHelp = %v, want %v", got, want)
	}
}

func TestGetAptOpsFallback(t *testing.T) {
	defer mockRunner(t, nil)() // empty map → all calls error → fallback
	got := getAptOps("apt")
	if !reflect.DeepEqual(got, aptOps) {
		t.Fatalf("expected fallback aptOps, got %v", got)
	}
}

func TestGetAptOpsDynamic(t *testing.T) {
	// `apt --help` parsing keeps emitting tokens after blank lines (upstream's
	// _parse_apt_operations doesn't terminate on a blank). For the apt-get/apt-cache
	// variant the parser DOES terminate on blank — exercise both shapes.
	defer mockRunner(t, map[string]specexec.Result{
		"apt --help":     {Stdout: []byte("Most used commands:\n  list - x\n  install - y\n")},
		"apt-get --help": {Stdout: []byte("Commands:\n  update - x\n  upgrade - y\n\nfoo trailing\n")},
	})()
	if got, want := getAptOps("apt"), []string{"list", "install"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("getAptOps('apt') = %v, want %v", got, want)
	}
	if got, want := getAptOps("apt-get"), []string{"update", "upgrade"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("getAptOps('apt-get') = %v, want %v", got, want)
	}
}

func TestGetGolangCmdsDynamic(t *testing.T) {
	stderr := "Go is a tool for managing Go source code.\n\nUsage:\n\n\tgo command [arguments]\n\nThe commands are:\n\n\tbuild       compile packages and dependencies\n\trun         compile and run Go program\n\ttest        test packages\n\n"
	defer mockRunner(t, map[string]specexec.Result{
		"go": {Stderr: []byte(stderr)},
	})()
	got := getGolangCmds()
	want := []string{"build", "run", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getGolangCmds dynamic = %v, want %v", got, want)
	}
}

func TestGetGemCmdsDynamic(t *testing.T) {
	stdout := "GEM commands are:\n    install   Install a gem\n    uninstall Uninstall gems\n    list      Display gems\nNot a command line\n"
	defer mockRunner(t, map[string]specexec.Result{
		"gem help commands": {Stdout: []byte(stdout)},
	})()
	got := getGemCmds()
	want := []string{"install", "uninstall", "list"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getGemCmds dynamic = %v, want %v", got, want)
	}
}

func TestGetGulpTasksDynamic(t *testing.T) {
	defer mockRunner(t, map[string]specexec.Result{
		"gulp --tasks-simple": {Stdout: []byte("default\nbuild\ntest\n")},
	})()
	got := getGulpTasks()
	want := []string{"default", "build", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getGulpTasks dynamic = %v, want %v", got, want)
	}
}

func TestGetGulpTasksFallback(t *testing.T) {
	defer mockRunner(t, nil)()
	got := getGulpTasks()
	if !reflect.DeepEqual(got, gulpTasksFallback) {
		t.Fatalf("getGulpTasks fallback = %v, want %v", got, gulpTasksFallback)
	}
}

// Integration: gradle_no_task should pick up the dynamic task list end-to-end.
func TestGradleNoTaskUsesDynamicTasks(t *testing.T) {
	gradleHelp := "" +
		"Tasks runnable from root project\n\n" +
		"------------------------------------------------------------\n" +
		"customTask - my custom thing\n" +
		"build - assembles and tests this project\n" +
		"\n"
	defer mockRunner(t, map[string]specexec.Result{
		"gradle tasks": {Stdout: []byte(gradleHelp)},
	})()

	got := mustRule(t, "gradle_no_task").GetNewCommand(
		cmd("gradle custmTsk", "Task 'custmTsk' not found"),
	)
	if len(got) == 0 {
		t.Fatalf("expected at least one suggestion, got none")
	}
	// 'customTask' is the closest dynamic match for 'custmTsk'; the static
	// fallback list does NOT contain it, so this confirms the seam fired.
	if got[0] != "gradle customTask" {
		t.Fatalf("first suggestion = %q, want %q (dynamic tasks not used?)", got[0], "gradle customTask")
	}
}
