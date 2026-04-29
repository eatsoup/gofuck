package rules

import (
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/specific"
	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

// pacmanGetPkgfile is the seam tests swap to inject canned package lists.
var pacmanGetPkgfile = func(script string) []string { return specific.GetPkgfile(script) }

// pacmanCmd is the wrapper invocation used in suggestions. Tests override.
var pacmanCmd = func() string { return specific.PacmanCmd }

func pacmanIsWrapper(name string) bool {
	switch name {
	case "pacman", "yay", "pikaur", "yaourt":
		return true
	}
	return false
}

func init() {
	Register(&types.Rule{
		Name:             "pacman",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			if !strings.Contains(c.Output, "not found") {
				return false
			}
			return len(pacmanGetPkgfile(c.Script)) > 0
		},
		GetNewCommand: func(c *types.Command) []string {
			packages := pacmanGetPkgfile(c.Script)
			if len(packages) == 0 {
				return nil
			}
			pacman := pacmanCmd()
			if pacman == "" {
				pacman = "pacman"
			}
			out := make([]string, 0, len(packages))
			for _, pkg := range packages {
				out = append(out, shells.Current.And(pacman+" -S "+pkg, c.Script))
			}
			return out
		},
	})

	Register(&types.Rule{
		Name:             "pacman_not_found",
		EnabledByDefault: true,
		RequiresOutput:   true,
		Match: func(c *types.Command) bool {
			if !strings.Contains(c.Output, "error: target not found:") {
				return false
			}
			parts := c.ScriptParts()
			if len(parts) == 0 {
				return false
			}
			if pacmanIsWrapper(parts[0]) {
				return true
			}
			return len(parts) >= 2 && parts[0] == "sudo" && parts[1] == "pacman"
		},
		GetNewCommand: func(c *types.Command) []string {
			parts := c.ScriptParts()
			if len(parts) == 0 {
				return nil
			}
			broken := parts[len(parts)-1]
			return utils.ReplaceCommand(c, broken, pacmanGetPkgfile(broken))
		},
	})
}
