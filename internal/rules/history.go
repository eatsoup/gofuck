package rules

import (
	"os"
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

// historyValidEntries is the seam tests swap to inject a deterministic
// history slice (matches upstream's mocker.patch on
// `get_valid_history_without_current`). Production reads from the active
// shell's history file via shells.Current.GetHistory().
var historyValidEntries = defaultValidHistory

func defaultValidHistory(c *types.Command) []string {
	hist := shells.Current.GetHistory()
	if len(hist) == 0 {
		return nil
	}
	tfAlias := os.Getenv("TF_ALIAS")
	if tfAlias == "" {
		tfAlias = "fuck"
	}

	// Build the executable + builtin set used to filter "real" lines.
	exes := pathExecutables()
	for _, b := range shells.Current.GetBuiltinCommands() {
		exes = append(exes, b)
	}
	known := make(map[string]struct{}, len(exes))
	for _, e := range exes {
		known[e] = struct{}{}
	}

	// notCorrected: drop the line that came right before a `fuck`
	// invocation (that's the broken script the user is now fixing).
	notCorrected := make([]string, 0, len(hist))
	var prev string
	have := false
	for _, line := range hist {
		if have && line != tfAlias {
			notCorrected = append(notCorrected, prev)
		}
		prev = line
		have = true
	}
	if have {
		notCorrected = append(notCorrected, prev)
	}

	out := make([]string, 0, len(notCorrected))
	for _, line := range notCorrected {
		if strings.HasPrefix(line, tfAlias) || line == c.Script {
			continue
		}
		first := line
		if i := strings.Index(line, " "); i >= 0 {
			first = line[:i]
		}
		if _, ok := known[first]; !ok {
			continue
		}
		out = append(out, line)
	}
	return out
}

func init() {
	Register(&types.Rule{
		Name:             "history",
		EnabledByDefault: true,
		RequiresOutput:   false,
		Priority:         9999,
		Match: func(c *types.Command) bool {
			return len(utils.GetCloseMatches(c.Script, historyValidEntries(c), 3, 0.6)) > 0
		},
		GetNewCommand: func(c *types.Command) []string {
			closest := utils.GetClosest(c.Script, historyValidEntries(c), 0.6, false)
			if closest == "" {
				return nil
			}
			return []string{closest}
		},
	})
}
