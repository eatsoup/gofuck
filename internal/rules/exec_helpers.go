package rules

import (
	"strings"

	"github.com/eatsoup/gofuck/internal/specific"
)

func getDockerCommands() []string {
	out, errOut, err := specific.Run("docker")
	if err != nil && out == "" && errOut == "" {
		return dockerCmds
	}
	var res []string
	listing := false
	for _, l := range strings.Split(out+"\n"+errOut, "\n") {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "Commands:") || strings.HasPrefix(l, "Management Commands:") {
			listing = true
			continue
		}
		if listing && l == "" {
			listing = false
			continue
		}
		if listing && l != "" {
			parts := strings.Fields(l)
			if len(parts) > 0 {
				res = append(res, parts[0])
			}
		}
	}
	if len(res) == 0 {
		return dockerCmds
	}
	return res
}

func getGolangCommands() []string {
	_, errOut, err := specific.Run("go")
	if err != nil && errOut == "" {
		return golangCmds
	}
	var res []string
	listing := false
	for _, l := range strings.Split(errOut, "\n") {
		l = strings.TrimSpace(l)
		if l == "The commands are:" {
			listing = true
			continue
		}
		if listing && l == "" {
			break
		}
		if listing {
			parts := strings.Fields(l)
			if len(parts) > 0 {
				res = append(res, parts[0])
			}
		}
	}
	if len(res) == 0 {
		return golangCmds
	}
	return res
}

func getGemCommands() []string {
	out, _, err := specific.Run("gem", "help", "commands")
	if err != nil && out == "" {
		return gemCmds
	}
	var res []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "    ") {
			parts := strings.Fields(l)
			if len(parts) > 0 {
				res = append(res, parts[0])
			}
		}
	}
	if len(res) == 0 {
		return gemCmds
	}
	return res
}

func getGruntTasks() []string {
	out, _, err := specific.Run("grunt", "--help")
	if err != nil && out == "" {
		return []string{"default", "build", "test", "watch", "concat", "clean", "copy", "uglify", "jshint", "lint"}
	}
	var res []string
	listing := false
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "Available tasks") {
			listing = true
			continue
		}
		if listing && strings.TrimSpace(l) == "" {
			break
		}
		if listing && strings.HasPrefix(l, "  ") {
			parts := strings.Fields(l)
			if len(parts) > 0 {
				res = append(res, parts[0])
			}
		}
	}
	if len(res) == 0 {
		return []string{"default", "build", "test", "watch", "concat", "clean", "copy", "uglify", "jshint", "lint"}
	}
	return res
}

func getGulpTasks() []string {
	out, _, err := specific.Run("gulp", "--tasks-simple")
	if err != nil && out == "" {
		return []string{"default", "build", "test", "watch", "lint"} // sensible fallback since it was hardcoded in rule
	}
	var res []string
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			res = append(res, l)
		}
	}
	if len(res) == 0 {
		return []string{"default", "build", "test", "watch", "lint"}
	}
	return res
}

func getYarnTasks() []string {
	out, _, err := specific.Run("yarn", "--help")
	if err != nil && out == "" {
		return yarnTasks
	}
	var res []string
	listing := false
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if strings.Contains(l, "Commands:") {
			listing = true
			continue
		}
		if listing && strings.Contains(l, "- ") {
			parts := strings.Split(l, " - ")
			if len(parts) > 0 {
				res = append(res, strings.TrimSpace(parts[0]))
			}
		}
	}
	if len(res) == 0 {
		return yarnTasks
	}
	return res
}

func getReactNativeCmds() []string {
	out, _, err := specific.Run("react-native", "--help")
	if err != nil && out == "" {
		return []string{"start", "run-android", "run-ios", "link", "unlink", "upgrade", "init", "log-android", "log-ios", "info", "bundle"}
	}
	var res []string
	listing := false
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if strings.Contains(l, "Commands:") {
			listing = true
			continue
		}
		if listing {
			parts := strings.Fields(l)
			if len(parts) > 0 {
				res = append(res, parts[0])
			}
		}
	}
	return res
}

func getAptOperations(app string) []string {
	out, _, err := specific.Run(app, "--help")
	if err != nil && out == "" {
		if app == "apt" {
			return aptOps
		}
		return []string{"update", "upgrade", "install", "remove", "purge", "autoremove", "dist-upgrade", "clean", "autoclean", "check", "source", "build-dep"} // sensible fallback
	}
	var res []string
	listing := false
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if app == "apt" {
			if listing && l != "" {
				res = append(res, strings.Fields(l)[0])
			} else if strings.HasPrefix(l, "Basic commands:") || strings.HasPrefix(l, "Most used commands:") {
				listing = true
			}
		} else {
			if listing {
				if l == "" {
					break
				}
				res = append(res, strings.Fields(l)[0])
			} else if strings.HasPrefix(l, "Commands:") || strings.HasPrefix(l, "Most used commands:") {
				listing = true
			}
		}
	}
	if len(res) == 0 && app == "apt" {
		return aptOps
	}
	return res
}

func getDnfOperations() []string {
	out, _, err := specific.Run("dnf", "--help")
	if err != nil && out == "" {
		return dnfOps
	}
	var res []string
	listing := false
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if listing && l != "" {
			res = append(res, strings.Fields(l)[0])
		} else if strings.HasPrefix(l, "Main Commands") || strings.HasPrefix(l, "List of Main Commands") {
			listing = true
		} else if listing && strings.HasPrefix(l, "Plugin Commands") {
			break
		}
	}
	if len(res) == 0 {
		return dnfOps
	}
	return res
}

func getYumOperations() []string {
	out, _, err := specific.Run("yum", "--help")
	if err != nil && out == "" {
		return yumOps
	}
	var res []string
	listing := false
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if listing && l != "" {
			res = append(res, strings.Fields(l)[0])
		} else if strings.HasPrefix(l, "Main Commands") || strings.HasPrefix(l, "List of Main Commands") {
			listing = true
		} else if listing && strings.HasPrefix(l, "Plugin Commands") {
			break
		}
	}
	if len(res) == 0 {
		return yumOps
	}
	return res
}

func getGradleTasks() []string {
	out, _, err := specific.Run("gradle", "tasks")
	if err != nil && out == "" {
		return []string{"assemble", "build", "check", "clean", "test", "install", "publish", "bootRun", "run", "jar", "war", "compileJava", "dependencies", "tasks", "wrapper"}
	}
	var res []string
	listing := false
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if strings.Contains(l, "----------") { // separator under "Build tasks" etc
			listing = true
			continue
		}
		if listing && l == "" {
			listing = false
			continue
		}
		if listing && strings.Contains(l, " - ") {
			parts := strings.Split(l, " - ")
			if len(parts) > 0 {
				res = append(res, strings.TrimSpace(parts[0]))
			}
		}
	}
	if len(res) == 0 {
		return []string{"assemble", "build", "check", "clean", "test", "install", "publish", "bootRun", "run", "jar", "war", "compileJava", "dependencies", "tasks", "wrapper"}
	}
	return res
}
