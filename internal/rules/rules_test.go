package rules

import (
	"reflect"
	"testing"

	"github.com/eatsoup/gofuck/internal/types"
)

// mustRule returns the registered rule by name or fails the test.
func mustRule(t *testing.T, name string) *types.Rule {
	t.Helper()
	r := Get(name)
	if r == nil {
		t.Fatalf("rule %q not registered", name)
	}
	return r
}

// cmd is a shorthand for types.NewCommand.
func cmd(script, output string) *types.Command {
	return types.NewCommand(script, output)
}

// assertMatch asserts that rule.Match returns want for the command.
func assertMatch(t *testing.T, name string, c *types.Command, want bool) {
	t.Helper()
	got := mustRule(t, name).Match(c)
	if got != want {
		t.Errorf("%s: Match(%q, %q) = %v, want %v", name, c.Script, c.Output, got, want)
	}
}

// assertNewCommand asserts the rule returns exactly want (single-string case).
func assertNewCommand(t *testing.T, name string, c *types.Command, want string) {
	t.Helper()
	got := mustRule(t, name).GetNewCommand(c)
	if len(got) == 0 {
		t.Errorf("%s: GetNewCommand(%q) returned nothing; want %q", name, c.Script, want)
		return
	}
	if got[0] != want {
		t.Errorf("%s: GetNewCommand(%q)[0] = %q, want %q", name, c.Script, got[0], want)
	}
}

// assertNewCommands asserts the full slice matches want.
func assertNewCommands(t *testing.T, name string, c *types.Command, want []string) {
	t.Helper()
	got := mustRule(t, name).GetNewCommand(c)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: GetNewCommand(%q) = %v, want %v", name, c.Script, got, want)
	}
}

// assertNewCommandIn asserts want appears in the returned commands
// (matches Python tests that use `in` rather than `==`).
func assertNewCommandIn(t *testing.T, name string, c *types.Command, want string) {
	t.Helper()
	got := mustRule(t, name).GetNewCommand(c)
	for _, s := range got {
		if s == want {
			return
		}
	}
	t.Errorf("%s: want %q in GetNewCommand(%q) = %v", name, want, c.Script, got)
}
