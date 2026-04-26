package exec

import (
	"errors"
	"testing"
)

func TestRunDelegatesToRunner(t *testing.T) {
	var gotName string
	var gotArgs []string
	defer WithRunner(func(name string, args ...string) Result {
		gotName, gotArgs = name, args
		return Result{Stdout: []byte("hello")}
	})()

	res := Run("echo", "a", "b")
	if gotName != "echo" || len(gotArgs) != 2 || gotArgs[0] != "a" || gotArgs[1] != "b" {
		t.Fatalf("runner not called with expected args: %q %v", gotName, gotArgs)
	}
	if string(res.Stdout) != "hello" {
		t.Fatalf("stdout = %q, want %q", res.Stdout, "hello")
	}
}

func TestWithRunnerRestores(t *testing.T) {
	prev := Runner
	restore := WithRunner(func(string, ...string) Result {
		return Result{Stderr: []byte("mocked")}
	})
	if &Runner == &prev {
		// not strictly meaningful but documents intent
	}
	restore()
	if got := Runner("noop").Stderr; len(got) != 0 {
		t.Fatalf("Runner not restored: stderr=%q", got)
	}
}

func TestResultOK(t *testing.T) {
	cases := []struct {
		r    Result
		want bool
	}{
		{Result{Stdout: []byte("x")}, true},
		{Result{Stderr: []byte("x")}, true},
		{Result{}, false},
		{Result{Stdout: []byte("x"), Err: errors.New("boom")}, false},
	}
	for _, tc := range cases {
		if got := tc.r.OK(); got != tc.want {
			t.Errorf("OK(%+v) = %v, want %v", tc.r, got, tc.want)
		}
	}
}
