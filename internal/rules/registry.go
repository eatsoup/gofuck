// Package rules hosts ported thefuck rules. Each rule file calls Register()
// in its init(), which appends to the global registry.
package rules

import (
	"sort"

	"github.com/eatsoup/gofuck/internal/conf"
	"github.com/eatsoup/gofuck/internal/types"
)

var registry []*types.Rule

// Register adds a rule. Priority defaults to DEFAULT_PRIORITY if zero;
// EnabledByDefault defaults to true unless explicitly set.
func Register(r *types.Rule) {
	if r.Priority == 0 {
		r.Priority = conf.DEFAULT_PRIORITY
	}
	registry = append(registry, r)
}

// All returns a copy of all registered rules sorted by name (deterministic).
func All() []*types.Rule {
	out := append([]*types.Rule(nil), registry...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get looks up a registered rule by name.
func Get(name string) *types.Rule {
	for _, r := range registry {
		if r.Name == name {
			return r
		}
	}
	return nil
}
