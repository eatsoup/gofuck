package rules

import (
	"strings"

	"github.com/eatsoup/gofuck/internal/types"
	"github.com/eatsoup/gofuck/internal/utils"
)

func init() {
	Register(&types.Rule{
		Name:             "man",
		EnabledByDefault: true,
		RequiresOutput:   false,
		Match: func(c *types.Command) bool {
			return utils.IsApp(c, 1, "man")
		},
		GetNewCommand: func(c *types.Command) []string {
			if strings.Contains(c.Script, "3") {
				return []string{strings.Replace(c.Script, "3", "2", 1)}
			}
			if strings.Contains(c.Script, "2") {
				return []string{strings.Replace(c.Script, "2", "3", 1)}
			}

			parts := c.ScriptParts()
			lastArg := parts[len(parts)-1]
			helpCommand := lastArg + " --help"

			if strings.TrimSpace(c.Output) == "No manual entry for "+lastArg {
				return []string{helpCommand}
			}

			cmd2 := make([]string, 0, len(parts)+1)
			cmd2 = append(cmd2, parts[0], "2")
			cmd2 = append(cmd2, parts[1:]...)

			cmd3 := make([]string, 0, len(parts)+1)
			cmd3 = append(cmd3, parts[0], "3")
			cmd3 = append(cmd3, parts[1:]...)

			return []string{
				strings.Join(cmd3, " "),
				strings.Join(cmd2, " "),
				helpCommand,
			}
		},
	})
}
