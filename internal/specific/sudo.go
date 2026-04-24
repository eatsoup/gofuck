// Package specific mirrors thefuck/specific — small decorators used by rules.
package specific

import (
	"strings"

	"github.com/eatsoup/gofuck/internal/types"
)

// SudoMatch wraps a match function to transparently strip a leading "sudo "
// from the command before testing.
func SudoMatch(fn types.MatchFn) types.MatchFn {
	return func(c *types.Command) bool {
		if !strings.HasPrefix(c.Script, "sudo ") {
			return fn(c)
		}
		stripped := c.Script[5:]
		return fn(c.Update(&stripped, nil))
	}
}

// SudoRewrite wraps a get_new_command function to transparently strip a
// leading "sudo " before calling fn, and re-prefix the returned scripts.
func SudoRewrite(fn types.GetNewCommandFn) types.GetNewCommandFn {
	return func(c *types.Command) []string {
		if !strings.HasPrefix(c.Script, "sudo ") {
			return fn(c)
		}
		stripped := c.Script[5:]
		out := fn(c.Update(&stripped, nil))
		for i, s := range out {
			out[i] = "sudo " + s
		}
		return out
	}
}
