// Command gofuck is the CLI for the rule pipeline. Two main entry points:
//
//   - `gofuck <cmd>` — run the rules against a (script, output) pair and
//     print the top correction (or all of them with --all). This is what
//     the AppAlias shell function calls behind the scenes.
//   - `gofuck --alias [name]` — print the shell function to source from
//     your rc file (e.g. `eval "$(gofuck --alias)"` in ~/.bashrc).
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eatsoup/gofuck/internal/corrector"
	"github.com/eatsoup/gofuck/internal/shells"
	"github.com/eatsoup/gofuck/internal/types"
)

func main() {
	all := flag.Bool("all", false, "print every candidate, one per line")
	outputFlag := flag.String("output", "", "captured stdout/stderr of the previous command")
	stdinOutput := flag.Bool("stdin", false, "read the previous command's output from stdin")
	aliasMode := flag.Bool("alias", false, "print the shell function to source from your rc file")
	shellName := flag.String("shell", "", "override shell detection (bash|zsh|fish|generic)")
	flag.Usage = usage
	flag.Parse()

	if *shellName != "" {
		shells.Use(*shellName)
	} else {
		shells.Auto()
	}

	if *aliasMode {
		name := "fuck"
		if rest := flag.Args(); len(rest) > 0 {
			name = rest[0]
		}
		fmt.Println(shells.Current.AppAlias(name))
		return
	}

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
	fmt.Fprintln(os.Stderr, "usage: gofuck [--output OUTPUT | --stdin] [--all] [--shell NAME] -- <command...>")
	fmt.Fprintln(os.Stderr, "       gofuck --alias [function-name]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  Runs the rule pipeline against the given command and prints the")
	fmt.Fprintln(os.Stderr, "  top correction (or all of them with --all). Exits 1 when no rule")
	fmt.Fprintln(os.Stderr, "  matches.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  --alias prints a shell function to install in your rc file:")
	fmt.Fprintln(os.Stderr, "    bash/zsh:  eval \"$(gofuck --alias)\"")
	fmt.Fprintln(os.Stderr, "    fish:      gofuck --alias | source")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  examples:")
	fmt.Fprintln(os.Stderr, "    gofuck git brnch")
	fmt.Fprintln(os.Stderr, "    gofuck --output \"$(git brnch 2>&1)\" -- git brnch")
	fmt.Fprintln(os.Stderr, "    git brnch 2>&1 | gofuck --stdin -- git brnch")
}
