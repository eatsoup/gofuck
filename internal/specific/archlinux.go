package specific

import (
	"strings"

	"github.com/eatsoup/gofuck/internal/specific/exec"
	"github.com/eatsoup/gofuck/internal/utils"
)

// ArchlinuxWhich is the seam tests swap to mock PATH lookups when probing
// for pacman wrappers and pkgfile.
var ArchlinuxWhich = utils.Which

// PacmanCmd holds the active pacman wrapper string ("yay", "pikaur", "yaourt"
// or "sudo pacman"). Empty when none of those wrappers is on PATH. Tests can
// reassign it directly.
var PacmanCmd = detectPacmanCmd()

// PkgfileEnabled reports whether `pkgfile` is on PATH at startup. Tests can
// reassign.
var PkgfileEnabled = ArchlinuxWhich("pkgfile") != ""

func detectPacmanCmd() string {
	switch {
	case ArchlinuxWhich("yay") != "":
		return "yay"
	case ArchlinuxWhich("pikaur") != "":
		return "pikaur"
	case ArchlinuxWhich("yaourt") != "":
		return "yaourt"
	case ArchlinuxWhich("pacman") != "":
		return "sudo pacman"
	}
	return ""
}

// GetPkgfile asks `pkgfile -b -v <command>` for packages providing `command`.
// A leading "sudo " is stripped and only the first whitespace-separated token
// is queried, matching upstream's archlinux.get_pkgfile. Returns nil on
// subprocess error or empty output.
var GetPkgfile = func(script string) []string {
	cmd := strings.TrimSpace(script)
	cmd = strings.TrimPrefix(cmd, "sudo ")
	if i := strings.IndexAny(cmd, " \t"); i >= 0 {
		cmd = cmd[:i]
	}
	if cmd == "" {
		return nil
	}
	res := exec.Run("pkgfile", "-b", "-v", cmd)
	if res.Err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(res.Stdout), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, strings.Fields(line)[0])
	}
	return out
}
