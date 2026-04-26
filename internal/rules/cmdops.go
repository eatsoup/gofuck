package rules

import (
	"regexp"
	"strings"

	"github.com/eatsoup/gofuck/internal/specific/exec"
)

// This file holds the per-tool helpers used by rules that upstream
// implements by shelling out (apt --help, gem help commands, …). Each
// helper invokes exec.Run via the swappable seam, parses the output
// using the same logic as upstream, and falls back to a static list
// when the tool isn't installed (or its output is unparseable).
//
// Tests mock exec.Runner to assert both the dynamic and the fallback paths.

// runHelp is a tiny convenience: combine stdout+stderr (some tools print
// help to stderr) and split into lines. Returns nil on subprocess error
// to signal "fall back".
func runHelp(name string, args ...string) []string {
	res := exec.Run(name, args...)
	if res.Err != nil {
		return nil
	}
	combined := string(res.Stdout)
	if combined == "" {
		combined = string(res.Stderr)
	}
	if combined == "" {
		return nil
	}
	return strings.Split(combined, "\n")
}

// firstWord returns the first whitespace-separated token of s, or "".
func firstWord(s string) string {
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}

// ---- apt / apt-get / apt-cache ----

func parseAptHelp(lines []string) []string {
	var out []string
	inList := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if inList {
			if line == "" {
				continue
			}
			out = append(out, firstWord(line))
		} else if strings.HasPrefix(line, "Basic commands:") || strings.HasPrefix(line, "Most used commands:") {
			inList = true
		}
	}
	return out
}

func parseAptGetHelp(lines []string) []string {
	var out []string
	inList := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if inList {
			if line == "" {
				return out
			}
			out = append(out, firstWord(line))
		} else if strings.HasPrefix(line, "Commands:") || strings.HasPrefix(line, "Most used commands:") {
			inList = true
		}
	}
	return out
}

func getAptOps(app string) []string {
	lines := runHelp(app, "--help")
	if lines == nil {
		return aptOps
	}
	var parsed []string
	if app == "apt" {
		parsed = parseAptHelp(lines)
	} else {
		parsed = parseAptGetHelp(lines)
	}
	if len(parsed) == 0 {
		return aptOps
	}
	return parsed
}

// ---- dnf ----

var dnfOpRe = regexp.MustCompile(`(?m)^([a-z-]+) +`)

func getDnfOps() []string {
	res := exec.Run("dnf", "--help")
	if res.Err != nil || len(res.Stdout) == 0 {
		return dnfOps
	}
	matches := dnfOpRe.FindAllStringSubmatch(string(res.Stdout), -1)
	if len(matches) == 0 {
		return dnfOps
	}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

// ---- yum ----

func getYumOps() []string {
	lines := runHelp("yum")
	if lines == nil {
		return yumOps
	}
	var parsed []string
	inList := false
	headerSeen := false
	for _, raw := range lines {
		if !inList {
			if strings.HasPrefix(strings.TrimSpace(raw), "List of Commands:") {
				inList = true
			}
			continue
		}
		// Skip the two header lines after "List of Commands:" (matches
		// upstream islice(lines, 2, None)).
		if !headerSeen {
			headerSeen = true
			continue
		}
		line := strings.TrimSpace(raw)
		if line == "" {
			break
		}
		parsed = append(parsed, firstWord(line))
	}
	if len(parsed) == 0 {
		return yumOps
	}
	return parsed
}

// ---- docker ----

// parseDockerCmds collects the section beginning with `startsWith`, skipping
// the header line, until a blank line. Mirrors upstream's _parse_commands.
func parseDockerCmds(lines []string, startsWith string) []string {
	var out []string
	inSection := false
	headerSeen := false
	for _, raw := range lines {
		if !inSection {
			if strings.HasPrefix(raw, startsWith) {
				inSection = true
			}
			continue
		}
		if !headerSeen {
			headerSeen = true
			continue
		}
		if strings.TrimSpace(raw) == "" {
			break
		}
		out = append(out, firstWord(strings.TrimSpace(raw)))
	}
	return out
}

func getDockerCmds() []string {
	lines := runHelp("docker")
	if lines == nil {
		return dockerCmds
	}
	hasMgmt := false
	for _, l := range lines {
		if l == "Management Commands:" {
			hasMgmt = true
			break
		}
	}
	var parsed []string
	if hasMgmt {
		parsed = parseDockerCmds(lines, "Management Commands:")
	}
	parsed = append(parsed, parseDockerCmds(lines, "Commands:")...)
	if len(parsed) == 0 {
		return dockerCmds
	}
	return parsed
}

// ---- gem ----

func getGemCmds() []string {
	lines := runHelp("gem", "help", "commands")
	if lines == nil {
		return gemCmds
	}
	var out []string
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") {
			out = append(out, firstWord(strings.TrimSpace(line)))
		}
	}
	if len(out) == 0 {
		return gemCmds
	}
	return out
}

// ---- go ----

func getGolangCmds() []string {
	res := exec.Run("go")
	if res.Err != nil && len(res.Stdout) == 0 && len(res.Stderr) == 0 {
		return golangCmds
	}
	body := string(res.Stderr)
	if body == "" {
		body = string(res.Stdout)
	}
	var out []string
	inSection := false
	skipped := 0
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if !inSection {
			if line == "The commands are:" {
				inSection = true
			}
			continue
		}
		// Upstream does islice(lines, 2, None) AFTER dropwhile, where dropwhile
		// keeps the marker as the first element. We've already consumed the
		// marker via `continue`, so skip one more line (the blank after it).
		if skipped < 1 {
			skipped++
			continue
		}
		if line == "" {
			break
		}
		out = append(out, firstWord(line))
	}
	if len(out) == 0 {
		return golangCmds
	}
	return out
}

// ---- gradle ----

func getGradleTasks(gradleBin string) []string {
	lines := runHelp(gradleBin, "tasks")
	if lines == nil {
		return gradleTasksFallback
	}
	var out []string
	yield := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.HasPrefix(line, "----") {
			yield = true
			continue
		}
		if line == "" {
			yield = false
			continue
		}
		if yield && !strings.HasPrefix(line, "All tasks runnable from root project") {
			out = append(out, firstWord(line))
		}
	}
	if len(out) == 0 {
		return gradleTasksFallback
	}
	return out
}

// ---- grunt ----

func getGruntTasks() []string {
	lines := runHelp("grunt", "--help")
	if lines == nil {
		return gruntTasksFallback
	}
	var out []string
	yield := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.Contains(line, "Available tasks") {
			yield = true
			continue
		}
		if yield && line == "" {
			break
		}
		if strings.Contains(raw, "  ") {
			out = append(out, firstWord(line))
		}
	}
	if len(out) == 0 {
		return gruntTasksFallback
	}
	return out
}

// ---- gulp ----

func getGulpTasks() []string {
	res := exec.Run("gulp", "--tasks-simple")
	if res.Err != nil || len(res.Stdout) == 0 {
		return gulpTasksFallback
	}
	var out []string
	for _, line := range strings.Split(string(res.Stdout), "\n") {
		line = strings.TrimRight(line, "\r")
		if line != "" {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		return gulpTasksFallback
	}
	return out
}

// ---- react-native ----

func getReactNativeCmds() []string {
	lines := runHelp("react-native", "--help")
	if lines == nil {
		return rnCmdsFallback
	}
	var out []string
	yield := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if strings.Contains(line, "Commands:") {
			yield = true
			continue
		}
		if yield {
			out = append(out, firstWord(line))
		}
	}
	if len(out) == 0 {
		return rnCmdsFallback
	}
	return out
}

// ---- yarn ----

func getYarnTasks() []string {
	lines := runHelp("yarn", "--help")
	if lines == nil {
		return yarnTasks
	}
	var out []string
	yield := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if strings.Contains(line, "Commands:") {
			yield = true
			continue
		}
		if yield && strings.Contains(line, "- ") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				out = append(out, fields[len(fields)-1])
			}
		}
	}
	if len(out) == 0 {
		return yarnTasks
	}
	return out
}

// Fallback static lists used when the subprocess seam can't reach the tool.
// Pulled out of the inline definitions in more.go so cmdops can reuse them.
var (
	gradleTasksFallback = []string{
		"assemble", "build", "check", "clean", "test", "install", "publish",
		"bootRun", "run", "jar", "war", "compileJava", "dependencies", "tasks",
		"wrapper",
	}
	gruntTasksFallback = []string{
		"default", "build", "test", "watch", "concat", "clean", "copy",
		"uglify", "jshint", "lint",
	}
	gulpTasksFallback = []string{"default", "build", "test", "watch", "lint"}
	rnCmdsFallback    = []string{
		"start", "run-android", "run-ios", "link", "unlink", "upgrade",
		"init", "log-android", "log-ios", "info", "bundle",
	}
)
