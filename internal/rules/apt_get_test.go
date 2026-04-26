package rules

import (
	"testing"
)

// withAptGetMocks swaps aptGetLookup and aptGetWhich for one block, mirroring
// upstream's mocker.patch on `_get_packages` and `which`.
func withAptGetMocks(t *testing.T, packages []string, whichRet string, fn func()) {
	t.Helper()
	prevL, prevW := aptGetLookup, aptGetWhich
	t.Cleanup(func() {
		aptGetLookup = prevL
		aptGetWhich = prevW
	})
	aptGetLookup = func(string) string {
		if len(packages) == 0 {
			return ""
		}
		return packages[0]
	}
	aptGetWhich = func(string) string { return whichRet }
	fn()
}

func TestAptGetMatch(t *testing.T) {
	cases := []struct {
		script   string
		output   string
		packages []string
	}{
		{"vim", "vim: command not found", []string{"vim", "vim-tiny"}},
		{"sudo vim", "vim: command not found", []string{"vim", "vim-tiny"}},
		{"vim", "The program 'vim' is currently not installed. You can install it by typing: sudo apt install vim", []string{"vim", "vim-tiny"}},
	}
	for _, tc := range cases {
		withAptGetMocks(t, tc.packages, "", func() {
			assertMatch(t, "apt_get", cmd(tc.script, tc.output), true)
		})
	}
}

func TestAptGetNotMatch(t *testing.T) {
	// Mirrors upstream's parametrize:
	//   (a_bad_cmd, []packages, no which) → not match (no package known)
	//   (vim, '', empty output)            → not match (output empty)
	//   ('', '')                           → not match (no exe)
	//   (vim, …, which='/usr/bin/vim')     → not match (already on PATH)
	//   (sudo vim, …, which='/usr/bin/vim')→ not match (already on PATH)
	cases := []struct {
		script, output, whichRet string
		packages                 []string
	}{
		{"a_bad_cmd", "a_bad_cmd: command not found", "", nil},
		{"vim", "", "", nil},
		{"", "", "", nil},
		{"vim", "vim: command not found", "/usr/bin/vim", []string{"vim"}},
		{"sudo vim", "vim: command not found", "/usr/bin/vim", []string{"vim"}},
	}
	for _, tc := range cases {
		withAptGetMocks(t, tc.packages, tc.whichRet, func() {
			assertMatch(t, "apt_get", cmd(tc.script, tc.output), false)
		})
	}
}

func TestAptGetNewCommand(t *testing.T) {
	cases := []struct {
		script   string
		want     string
		packages []string
	}{
		{"vim", "sudo apt-get install vim && vim", []string{"vim", "vim-tiny"}},
		{"convert", "sudo apt-get install imagemagick && convert", []string{"imagemagick", "graphicsmagick-imagemagick-compat"}},
		{"sudo vim", "sudo apt-get install vim && sudo vim", []string{"vim", "vim-tiny"}},
		{"sudo convert", "sudo apt-get install imagemagick && sudo convert", []string{"imagemagick", "graphicsmagick-imagemagick-compat"}},
	}
	for _, tc := range cases {
		withAptGetMocks(t, tc.packages, "", func() {
			assertNewCommand(t, "apt_get", cmd(tc.script, ""), tc.want)
		})
	}
}
