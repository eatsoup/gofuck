package rules

import (
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

// aptGetPackage maps a missing executable to the Debian/Ubuntu package that
// provides it. Upstream's apt_get rule shells out to Python's CommandNotFound
// module (backed by /var/lib/command-not-found/commands.db). We don't have
// that integration in Go yet, so we ship a small static map covering the
// most common cases. Tests override aptGetLookup to inject canned data.
var aptGetPackages = map[string]string{
	"vim":     "vim",
	"convert": "imagemagick",
	"ack":     "ack-grep",
	"ag":      "silversearcher-ag",
	"htop":    "htop",
	"git":     "git",
	"node":    "nodejs",
	"npm":     "npm",
	"python":  "python3",
	"pip":     "python3-pip",
	"curl":    "curl",
	"wget":    "wget",
	"tmux":    "tmux",
	"tree":    "tree",
	"jq":      "jq",
	"unzip":   "unzip",
	"zip":     "zip",
	"make":    "build-essential",
	"gcc":     "build-essential",
	"g++":     "build-essential",
	"java":    "default-jre",
	"javac":   "default-jdk",
	"go":      "golang",
	"ruby":    "ruby",
	"perl":    "perl",
	"man":     "man-db",
}

// aptGetLookup is the seam tests swap to inject fake packages. It returns the
// preferred package name for `executable`, or "" when nothing is known.
var aptGetLookup = func(executable string) string {
	return aptGetPackages[executable]
}

// aptGetWhich is the seam tests swap to mock PATH lookups, mirroring upstream's
// mocker.patch on `which`.
var aptGetWhich = utils.Which

func aptGetExecutable(c *types.Command) string {
	parts := c.ScriptParts()
	if len(parts) == 0 {
		return ""
	}
	if parts[0] == "sudo" && len(parts) > 1 {
		return parts[1]
	}
	return parts[0]
}

func init() {
	Register(&types.Rule{
		Name: "apt_get", EnabledByDefault: true, RequiresOutput: true,
		Match: func(c *types.Command) bool {
			if !strings.Contains(c.Output, "not found") && !strings.Contains(c.Output, "not installed") {
				return false
			}
			exe := aptGetExecutable(c)
			if exe == "" {
				return false
			}
			if aptGetWhich(exe) != "" {
				return false
			}
			return aptGetLookup(exe) != ""
		},
		GetNewCommand: func(c *types.Command) []string {
			pkg := aptGetLookup(aptGetExecutable(c))
			return []string{shells.Current.And("sudo apt-get install "+pkg, c.Script)}
		},
	})
}
