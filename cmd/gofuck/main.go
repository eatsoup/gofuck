// Command gofuck is a minimal CLI that runs the rule pipeline against a
// (script, output) pair and prints the candidate fix(es). It is intentionally
// small: enough to drive end-to-end tests by hand while we keep porting the
// upstream test suite. Shell integration (alias, history capture, the
// auto-rerun ceremony) is out of scope for now.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eatsoup/gofuck/internal/corrector"
	"github.com/eatsoup/gofuck/internal/types"
)

func main() {
	all := flag.Bool("all", false, "print every candidate, one per line")
	outputFlag := flag.String("output", "", "captured stdout/stderr of the previous command")
	stdinOutput := flag.Bool("stdin", false, "read the previous command's output from stdin")
	flag.Usage = usage
	flag.Parse()

	scriptParts := flag.Args()
	if len(scriptParts) == 0 {
		fmt.Fprintln(os.Stderr, "gofuck: no command given")
		usage()
		os.Exit(2)
	}
	script := strings.Join(scriptParts, " ")

	output := *outputFlag
	if *stdinOutput {
		buf, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gofuck: read stdin: %v\n", err)
			os.Exit(2)
		}
		output = string(buf)
	}

	cmd := types.NewCommand(script, output)
	candidates := corrector.GetCorrectedCommands(cmd)
	if len(candidates) == 0 {
		fmt.Fprintln(os.Stderr, "gofuck: no correction found")
		os.Exit(1)
	}

	if *all {
		for _, c := range candidates {
			fmt.Println(c.Script)
		}
		return
	}
	fmt.Println(candidates[0].Script)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gofuck [--output OUTPUT | --stdin] [--all] -- <command...>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  Runs the rule pipeline against the given command and prints the")
	fmt.Fprintln(os.Stderr, "  top correction (or all of them with --all). Exits 1 when no rule")
	fmt.Fprintln(os.Stderr, "  matches.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  examples:")
	fmt.Fprintln(os.Stderr, "    gofuck git brnch")
	fmt.Fprintln(os.Stderr, "    gofuck --output \"$(git brnch 2>&1)\" -- git brnch")
	fmt.Fprintln(os.Stderr, "    git brnch 2>&1 | gofuck --stdin -- git brnch")
}
