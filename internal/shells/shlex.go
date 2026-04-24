package shells

import (
	"errors"
	"strings"
	"unicode"
)

// shlexSplit mirrors Python's shlex.split in posix mode for the subset of
// behaviour thefuck relies on. Returns an error on unterminated quotes.
func shlexSplit(s string) ([]string, error) {
	var out []string
	var cur strings.Builder
	inSingle, inDouble, escaped := false, false, false
	started := false
	for _, r := range s {
		if escaped {
			cur.WriteRune(r)
			escaped = false
			started = true
			continue
		}
		if r == '\\' && !inSingle {
			escaped = true
			started = true
			continue
		}
		if inSingle {
			if r == '\'' {
				inSingle = false
			} else {
				cur.WriteRune(r)
			}
			continue
		}
		if inDouble {
			if r == '"' {
				inDouble = false
			} else {
				cur.WriteRune(r)
			}
			continue
		}
		if r == '\'' {
			inSingle = true
			started = true
			continue
		}
		if r == '"' {
			inDouble = true
			started = true
			continue
		}
		if unicode.IsSpace(r) {
			if started {
				out = append(out, cur.String())
				cur.Reset()
				started = false
			}
			continue
		}
		cur.WriteRune(r)
		started = true
	}
	if inSingle || inDouble {
		return nil, errors.New("unterminated quote")
	}
	if started {
		out = append(out, cur.String())
	}
	return out, nil
}
