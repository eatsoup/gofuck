package rules

import (
	"testing"
)

func TestMan(t *testing.T) {
	assertMatch(t, "man", cmd("man read", ""), true)
	assertMatch(t, "man", cmd("man 2 read", ""), true)
	assertMatch(t, "man", cmd("man 3 read", ""), true)
	assertMatch(t, "man", cmd("man -s2 read", ""), true)
	assertMatch(t, "man", cmd("man -s3 read", ""), true)
	assertMatch(t, "man", cmd("man -s 2 read", ""), true)
	assertMatch(t, "man", cmd("man -s 3 read", ""), true)

	assertMatch(t, "man", cmd("man", ""), false)
	assertMatch(t, "man", cmd("man ", ""), false)

	assertNewCommands(t, "man", cmd("man read", ""), []string{"man 3 read", "man 2 read", "read --help"})
	assertNewCommands(t, "man", cmd("man missing", "No manual entry for missing\n"), []string{"missing --help"})
	
	assertNewCommand(t, "man", cmd("man 2 read", ""), "man 3 read")
	assertNewCommand(t, "man", cmd("man 3 read", ""), "man 2 read")
	assertNewCommand(t, "man", cmd("man -s2 read", ""), "man -s3 read")
	assertNewCommand(t, "man", cmd("man -s3 read", ""), "man -s2 read")
	assertNewCommand(t, "man", cmd("man -s 2 read", ""), "man -s 3 read")
	assertNewCommand(t, "man", cmd("man -s 3 read", ""), "man -s 2 read")
}
