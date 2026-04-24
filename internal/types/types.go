package types

import (
	"strings"

	"github.com/eatsoup/gofuck/internal/shells"
)

// Command represents a shell command the user ran that should be fixed.
type Command struct {
	Script string
	Output string

	partsCache []string
	partsOnce  bool
}

// NewCommand constructs a Command from a script and captured output.
func NewCommand(script, output string) *Command {
	return &Command{Script: script, Output: output}
}

// ScriptParts returns the command split into shell-style tokens.
func (c *Command) ScriptParts() []string {
	if !c.partsOnce {
		c.partsCache = shells.Current.SplitCommand(c.Script)
		c.partsOnce = true
	}
	return c.partsCache
}

// Update returns a new Command with overridden fields (script/output).
func (c *Command) Update(script, output *string) *Command {
	s := c.Script
	o := c.Output
	if script != nil {
		s = *script
	}
	if output != nil {
		o = *output
	}
	return NewCommand(s, o)
}

// MatchFn tests whether a rule applies to a command.
type MatchFn func(*Command) bool

// GetNewCommandFn returns one or more replacement command strings.
type GetNewCommandFn func(*Command) []string

// SideEffectFn runs after a fix is applied.
type SideEffectFn func(old *Command, newScript string)

// Rule is a single correction rule.
type Rule struct {
	Name             string
	Match            MatchFn
	GetNewCommand    GetNewCommandFn
	EnabledByDefault bool
	SideEffect       SideEffectFn
	Priority         int
	RequiresOutput   bool
}

// CorrectedCommand is a candidate fix produced by a rule.
type CorrectedCommand struct {
	Script     string
	SideEffect SideEffectFn
	Priority   int
}

// Key returns an equality key ignoring priority.
func (c *CorrectedCommand) Key() string {
	if c.SideEffect == nil {
		return c.Script + "|nil"
	}
	// function pointers aren't comparable via string; use address-less sentinel.
	return c.Script + "|fn"
}

// Equals compares two corrected commands (ignoring priority, per Python semantics).
func (c *CorrectedCommand) Equals(o *CorrectedCommand) bool {
	if c == nil || o == nil {
		return c == o
	}
	return c.Script == o.Script && sameFn(c.SideEffect, o.SideEffect)
}

func sameFn(a, b SideEffectFn) bool {
	// Best-effort: both nil => equal; both non-nil => treat as equal only if
	// address of reflect value matches. For de-duplication, we collapse all
	// non-nil side effects into a script-based key; callers should normalize.
	if a == nil && b == nil {
		return true
	}
	if (a == nil) != (b == nil) {
		return false
	}
	return true
}

// GetCorrectedCommands expands a rule's get_new_command into CorrectedCommands.
func (r *Rule) GetCorrectedCommands(cmd *Command) []*CorrectedCommand {
	news := r.GetNewCommand(cmd)
	out := make([]*CorrectedCommand, 0, len(news))
	for i, n := range news {
		out = append(out, &CorrectedCommand{
			Script:     n,
			SideEffect: r.SideEffect,
			Priority:   (i + 1) * r.Priority,
		})
	}
	return out
}

// IsMatch runs the rule's match predicate, respecting requires_output.
func (r *Rule) IsMatch(cmd *Command) bool {
	if cmd.Output == "" && r.RequiresOutput {
		return false
	}
	defer func() {
		// Never let a rule panic the pipeline.
		_ = recover()
	}()
	return r.Match(cmd)
}

// TrimSpace is a tiny helper some rules use.
func TrimSpace(s string) string { return strings.TrimSpace(s) }
