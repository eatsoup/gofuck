// Package utils ports the helpers from thefuck/utils.py that rules depend on.
package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eatsoup/gofuck/internal/conf"
	"github.com/eatsoup/gofuck/internal/types"
)

// ReplaceArgument replaces a single occurrence of `from` with `to` in the
// script, preferring an end-of-string match, and falling back to a space-bounded
// mid-string match. Mirrors thefuck.utils.replace_argument.
func ReplaceArgument(script, from, to string) string {
	endPat := regexp.MustCompile(` ` + regexp.QuoteMeta(from) + `$`)
	if endPat.MatchString(script) {
		return endPat.ReplaceAllString(script, " "+to)
	}
	return strings.Replace(script, " "+from+" ", " "+to+" ", 1)
}

// IsApp reports whether the command invokes any of the named apps. `atLeast`
// is the minimum number of trailing args.
func IsApp(cmd *types.Command, atLeast int, apps ...string) bool {
	parts := cmd.ScriptParts()
	if len(parts) > atLeast {
		bin := filepath.Base(parts[0])
		for _, a := range apps {
			if bin == a {
				return true
			}
		}
	}
	return false
}

// GetAllMatchedCommands parses a block of lines following a separator and
// yields each line until the block ends. Mirrors
// thefuck.utils.get_all_matched_commands.
func GetAllMatchedCommands(stderr string, separators ...string) []string {
	if len(separators) == 0 {
		separators = []string{"Did you mean"}
	}
	var out []string
	shouldYield := false
	for _, line := range strings.Split(stderr, "\n") {
		matched := false
		for _, sep := range separators {
			if strings.Contains(line, sep) {
				matched = true
				break
			}
		}
		if matched {
			shouldYield = true
			continue
		}
		if shouldYield && line != "" {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

// GetCloseMatches is an SequenceMatcher-based ratio matching close strings
// — thefuck uses difflib.get_close_matches. This is a Go port of the same
// algorithm (Ratcliff/Obershelp).
func GetCloseMatches(word string, possibilities []string, n int, cutoff float64) []string {
	type scored struct {
		s     string
		ratio float64
	}
	var ranked []scored
	for _, p := range possibilities {
		// Matches Python difflib.get_close_matches: seq1=candidate, seq2=word,
		// which is asymmetric in the algorithm (b2j built from seq2).
		r := ratio(p, word)
		if r >= cutoff {
			ranked = append(ranked, scored{p, r})
		}
	}
	// Insertion sort by (ratio desc, string desc) — matches Python's
	// heapq.nlargest tuple comparison used by difflib.get_close_matches.
	for i := 1; i < len(ranked); i++ {
		for j := i; j > 0; j-- {
			if ranked[j-1].ratio < ranked[j].ratio ||
				(ranked[j-1].ratio == ranked[j].ratio && ranked[j-1].s < ranked[j].s) {
				ranked[j-1], ranked[j] = ranked[j], ranked[j-1]
			} else {
				break
			}
		}
	}
	if n > 0 && len(ranked) > n {
		ranked = ranked[:n]
	}
	out := make([]string, len(ranked))
	for i, r := range ranked {
		out[i] = r.s
	}
	return out
}

// GetClosest returns the closest match or the first element if fallback is true.
func GetClosest(word string, possibilities []string, cutoff float64, fallbackToFirst bool) string {
	matches := GetCloseMatches(word, possibilities, 1, cutoff)
	if len(matches) > 0 {
		return matches[0]
	}
	if fallbackToFirst && len(possibilities) > 0 {
		return possibilities[0]
	}
	return ""
}

// ReplaceCommand is the helper used by *_no_command rules: it finds close
// matches and produces a list of rewritten scripts.
func ReplaceCommand(cmd *types.Command, broken string, matched []string) []string {
	newCmds := GetCloseMatches(broken, matched, conf.Current.NumCloseMatches, 0.1)
	out := make([]string, 0, len(newCmds))
	for _, n := range newCmds {
		out = append(out, ReplaceArgument(cmd.Script, broken, strings.TrimSpace(n)))
	}
	return out
}

// Which returns the path to `program` using PATH, or "".
func Which(program string) string {
	path, err := exec_LookPath(program)
	if err != nil {
		return ""
	}
	return path
}

// exec.LookPath wrapped for easy testing.
func exec_LookPath(program string) (string, error) {
	return lookPath(program)
}

func lookPath(program string) (string, error) {
	if strings.ContainsRune(program, os.PathSeparator) {
		if isExec(program) {
			return program, nil
		}
		return "", os.ErrNotExist
	}
	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, dir := range paths {
		if dir == "" {
			continue
		}
		cand := filepath.Join(dir, program)
		if isExec(cand) {
			return cand, nil
		}
	}
	return "", os.ErrNotExist
}

func isExec(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	return info.Mode()&0o111 != 0
}

// --- internals: Ratcliff/Obershelp ratio ---

// ratio implements difflib.SequenceMatcher.ratio().
func ratio(a, b string) float64 {
	ar, br := []rune(a), []rune(b)
	la, lb := len(ar), len(br)
	if la+lb == 0 {
		return 1.0
	}
	matches := matchingBlocks(ar, br)
	var total int
	for _, m := range matches {
		total += m.size
	}
	return 2.0 * float64(total) / float64(la+lb)
}

type match struct{ a, b, size int }

func matchingBlocks(a, b []rune) []match {
	// Recursive longest-match approach used by difflib.
	var out []match
	var rec func(alo, ahi, blo, bhi int)
	rec = func(alo, ahi, blo, bhi int) {
		m := longestMatch(a, b, alo, ahi, blo, bhi)
		if m.size == 0 {
			return
		}
		if m.a > alo && m.b > blo {
			rec(alo, m.a, blo, m.b)
		}
		out = append(out, m)
		if m.a+m.size < ahi && m.b+m.size < bhi {
			rec(m.a+m.size, ahi, m.b+m.size, bhi)
		}
	}
	rec(0, len(a), 0, len(b))
	return out
}

func longestMatch(a, b []rune, alo, ahi, blo, bhi int) match {
	// b2j: map each rune in b to positions.
	b2j := make(map[rune][]int)
	for j := blo; j < bhi; j++ {
		b2j[b[j]] = append(b2j[b[j]], j)
	}
	bestI, bestJ, bestSize := alo, blo, 0
	j2len := make(map[int]int)
	for i := alo; i < ahi; i++ {
		next := make(map[int]int)
		for _, j := range b2j[a[i]] {
			k := j2len[j-1] + 1
			next[j] = k
			if k > bestSize {
				bestI = i - k + 1
				bestJ = j - k + 1
				bestSize = k
			}
		}
		j2len = next
	}
	return match{bestI, bestJ, bestSize}
}
