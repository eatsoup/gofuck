package corrector

import (
	"sort"

	"github.com/eatsoup/gofuck/internal/conf"
	"github.com/eatsoup/gofuck/internal/rules"
	"github.com/eatsoup/gofuck/internal/types"
)

// GetRules returns all enabled rules sorted by priority.
func GetRules() []*types.Rule {
	all := rules.All()
	enabled := make([]*types.Rule, 0, len(all))
	for _, r := range all {
		if isEnabled(r) {
			enabled = append(enabled, r)
		}
	}
	sort.SliceStable(enabled, func(i, j int) bool {
		return enabled[i].Priority < enabled[j].Priority
	})
	return enabled
}

func isEnabled(r *types.Rule) bool {
	if contains(conf.Current.ExcludeRules, r.Name) {
		return false
	}
	if contains(conf.Current.Rules, r.Name) {
		return true
	}
	if r.EnabledByDefault && contains(conf.Current.Rules, conf.ALL_ENABLED) {
		return true
	}
	return false
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// GetCorrectedCommands yields the ordered, de-duplicated candidate fixes for
// the given command.
func GetCorrectedCommands(cmd *types.Command) []*types.CorrectedCommand {
	var all []*types.CorrectedCommand
	for _, r := range GetRules() {
		if !r.IsMatch(cmd) {
			continue
		}
		all = append(all, r.GetCorrectedCommands(cmd)...)
	}
	return organize(all)
}

func organize(cmds []*types.CorrectedCommand) []*types.CorrectedCommand {
	if len(cmds) == 0 {
		return nil
	}
	first := cmds[0]
	// Sort the tail by priority.
	tail := append([]*types.CorrectedCommand(nil), cmds[1:]...)
	sort.SliceStable(tail, func(i, j int) bool {
		return tail[i].Priority < tail[j].Priority
	})
	seen := map[string]bool{first.Script: true}
	var deduped []*types.CorrectedCommand
	for _, c := range tail {
		if seen[c.Script] {
			continue
		}
		seen[c.Script] = true
		deduped = append(deduped, c)
	}
	sort.SliceStable(deduped, func(i, j int) bool {
		return deduped[i].Priority < deduped[j].Priority
	})
	return append([]*types.CorrectedCommand{first}, deduped...)
}
