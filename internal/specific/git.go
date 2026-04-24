package specific

import (
	"regexp"
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

// GitSupportMatch wraps a match fn so it only runs for `git`/`hub` commands,
// also expanding any git alias-expansion trace found in output.
func GitSupportMatch(fn types.MatchFn) types.MatchFn {
	return func(c *types.Command) bool {
		if !utils.IsApp(c, 0, "git", "hub") {
			return false
		}
		if c.Output != "" && strings.Contains(c.Output, "trace: alias expansion:") {
			c = expandAlias(c)
		}
		return fn(c)
	}
}

// GitSupportRewrite wraps a get_new_command fn the same way.
func GitSupportRewrite(fn types.GetNewCommandFn) types.GetNewCommandFn {
	return func(c *types.Command) []string {
		if !utils.IsApp(c, 0, "git", "hub") {
			return nil
		}
		if c.Output != "" && strings.Contains(c.Output, "trace: alias expansion:") {
			c = expandAlias(c)
		}
		return fn(c)
	}
}

var aliasExpansionRe = regexp.MustCompile(`trace: alias expansion: ([^ ]*) => ([^\n]*)`)

func expandAlias(c *types.Command) *types.Command {
	m := aliasExpansionRe.FindStringSubmatch(c.Output)
	if m == nil {
		return c
	}
	alias := m[1]
	parts := shells.Current.SplitCommand(m[2])
	quoted := make([]string, len(parts))
	for i, p := range parts {
		quoted[i] = shells.Current.Quote(p)
	}
	expansion := strings.Join(quoted, " ")
	pat := regexp.MustCompile(`\b` + regexp.QuoteMeta(alias) + `\b`)
	newScript := pat.ReplaceAllString(c.Script, expansion)
	return c.Update(&newScript, nil)
}
