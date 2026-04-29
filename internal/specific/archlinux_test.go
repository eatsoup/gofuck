package specific

import (
	"reflect"
	"testing"

	"github.com/eatsoup/gofuck/internal/specific/exec"
)

func TestGetPkgfileParsesOutput(t *testing.T) {
	defer exec.WithRunner(func(name string, args ...string) exec.Result {
		if name != "pkgfile" || len(args) != 3 || args[0] != "-b" || args[1] != "-v" || args[2] != "vim" {
			t.Fatalf("unexpected exec call: %q %v", name, args)
		}
		return exec.Result{Stdout: []byte(
			"extra/gvim 7.4.712-1        \t/usr/bin/vim\n" +
				"extra/gvim-python3 7.4.712-1\t/usr/bin/vim\n" +
				"extra/vim 7.4.712-1         \t/usr/bin/vim\n" +
				"extra/vim-minimal 7.4.712-1 \t/usr/bin/vim\n" +
				"extra/vim-python3 7.4.712-1 \t/usr/bin/vim\n",
		)}
	})()
	got := GetPkgfile("vim")
	want := []string{"extra/gvim", "extra/gvim-python3", "extra/vim", "extra/vim-minimal", "extra/vim-python3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("GetPkgfile(\"vim\") = %v, want %v", got, want)
	}
}

func TestGetPkgfileStripsSudoAndArgs(t *testing.T) {
	var gotArg string
	defer exec.WithRunner(func(name string, args ...string) exec.Result {
		gotArg = args[len(args)-1]
		return exec.Result{Stdout: []byte("core/sudo 1.8.13-13/usr/bin/sudo\n")}
	})()
	if got := GetPkgfile("sudo vim --foo"); !reflect.DeepEqual(got, []string{"core/sudo"}) {
		t.Fatalf("GetPkgfile = %v", got)
	}
	if gotArg != "vim" {
		t.Fatalf("pkgfile called with %q, want %q", gotArg, "vim")
	}
}

func TestGetPkgfileReturnsNilOnError(t *testing.T) {
	defer exec.WithRunner(func(string, ...string) exec.Result {
		return exec.Result{Err: errBoom}
	})()
	if got := GetPkgfile("vim"); got != nil {
		t.Fatalf("GetPkgfile error path returned %v, want nil", got)
	}
}

var errBoom = &boomErr{}

type boomErr struct{}

func (*boomErr) Error() string { return "boom" }
