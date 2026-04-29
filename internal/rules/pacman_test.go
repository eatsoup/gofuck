package rules

import (
	"reflect"
	"strings"
	"testing"
)

// pacmanFixtureCmd is the wrapper string the upstream test fixtures use.
const pacmanFixtureCmd = "pacman"

var pkgfileVim = []string{
	"extra/gvim", "extra/gvim-python3", "extra/vim",
	"extra/vim-minimal", "extra/vim-python3",
}
var pkgfileConvert = []string{"extra/imagemagick"}
var pkgfileSudo = []string{"core/sudo"}
var pkgfileLLC = []string{"extra/llvm", "extra/llvm35"}

// withPacmanMocks installs canned pkgfile responses (keyed by the lookup
// argument we expect, after the same trim/sudo-strip the seam does for the
// real binary) and pins the wrapper to `wrapper`.
func withPacmanMocks(t *testing.T, byArg map[string][]string, wrapper string, fn func()) {
	t.Helper()
	prevPkg, prevCmd := pacmanGetPkgfile, pacmanCmd
	t.Cleanup(func() {
		pacmanGetPkgfile = prevPkg
		pacmanCmd = prevCmd
	})
	pacmanGetPkgfile = func(script string) []string {
		s := strings.TrimSpace(script)
		s = strings.TrimPrefix(s, "sudo ")
		if i := strings.IndexAny(s, " \t"); i >= 0 {
			s = s[:i]
		}
		return byArg[s]
	}
	pacmanCmd = func() string { return wrapper }
	fn()
}

func TestPacmanMatch(t *testing.T) {
	cases := []struct {
		script, output string
	}{
		{"vim", "vim: command not found"},
		{"sudo vim", "sudo: vim: command not found"},
	}
	for _, tc := range cases {
		withPacmanMocks(t, map[string][]string{"vim": pkgfileVim}, pacmanFixtureCmd, func() {
			assertMatch(t, "pacman", cmd(tc.script, tc.output), true)
		})
	}
}

func TestPacmanNotMatch(t *testing.T) {
	cases := []struct {
		script, output string
	}{
		{"vim", ""},
		{"", ""},
		{"sudo vim", ""},
	}
	for _, tc := range cases {
		withPacmanMocks(t, map[string][]string{"vim": pkgfileVim}, pacmanFixtureCmd, func() {
			assertMatch(t, "pacman", cmd(tc.script, tc.output), false)
		})
	}
}

func TestPacmanNewCommand(t *testing.T) {
	cases := []struct {
		script string
		pkgs   []string
		want   []string
	}{
		{
			"vim", pkgfileVim,
			[]string{
				"pacman -S extra/gvim && vim",
				"pacman -S extra/gvim-python3 && vim",
				"pacman -S extra/vim && vim",
				"pacman -S extra/vim-minimal && vim",
				"pacman -S extra/vim-python3 && vim",
			},
		},
		{
			"sudo vim", pkgfileVim,
			[]string{
				"pacman -S extra/gvim && sudo vim",
				"pacman -S extra/gvim-python3 && sudo vim",
				"pacman -S extra/vim && sudo vim",
				"pacman -S extra/vim-minimal && sudo vim",
				"pacman -S extra/vim-python3 && sudo vim",
			},
		},
		{
			"convert", pkgfileConvert,
			[]string{"pacman -S extra/imagemagick && convert"},
		},
		{
			"sudo convert", pkgfileConvert,
			[]string{"pacman -S extra/imagemagick && sudo convert"},
		},
		{
			"sudo", pkgfileSudo,
			[]string{"pacman -S core/sudo && sudo"},
		},
	}
	for _, tc := range cases {
		// pacmanGetPkgfile is keyed off the script's first non-sudo word, so we
		// register the same packages under whatever that word is.
		s := strings.TrimSpace(tc.script)
		s = strings.TrimPrefix(s, "sudo ")
		if i := strings.IndexAny(s, " \t"); i >= 0 {
			s = s[:i]
		}
		withPacmanMocks(t, map[string][]string{s: tc.pkgs}, pacmanFixtureCmd, func() {
			assertNewCommands(t, "pacman", cmd(tc.script, ""), tc.want)
		})
	}
}

func TestPacmanNotFoundMatch(t *testing.T) {
	cases := []struct {
		script, output string
	}{
		{"yay -S llc", "error: target not found: llc"},
		{"pikaur -S llc", "error: target not found: llc"},
		{"yaourt -S llc", "error: target not found: llc"},
		{"pacman llc", "error: target not found: llc"},
		{"sudo pacman llc", "error: target not found: llc"},
	}
	for _, tc := range cases {
		withPacmanMocks(t, map[string][]string{"llc": pkgfileLLC}, pacmanFixtureCmd, func() {
			assertMatch(t, "pacman_not_found", cmd(tc.script, tc.output), true)
		})
	}
}

func TestPacmanNotFoundNewCommand(t *testing.T) {
	cases := []struct {
		script string
		want   []string
	}{
		{"yay -S llc", []string{"yay -S extra/llvm", "yay -S extra/llvm35"}},
		{"pikaur -S llc", []string{"pikaur -S extra/llvm", "pikaur -S extra/llvm35"}},
		{"yaourt -S llc", []string{"yaourt -S extra/llvm", "yaourt -S extra/llvm35"}},
		{"pacman -S llc", []string{"pacman -S extra/llvm", "pacman -S extra/llvm35"}},
		{"sudo pacman -S llc", []string{"sudo pacman -S extra/llvm", "sudo pacman -S extra/llvm35"}},
	}
	for _, tc := range cases {
		withPacmanMocks(t, map[string][]string{"llc": pkgfileLLC}, pacmanFixtureCmd, func() {
			got := mustRule(t, "pacman_not_found").GetNewCommand(cmd(tc.script, "error: target not found: llc"))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("pacman_not_found: GetNewCommand(%q) = %v, want %v", tc.script, got, tc.want)
			}
		})
	}
}
