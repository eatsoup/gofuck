package rules

import (
	"reflect"
	"testing"

	specexec "github.com/eatsoup/gofuck/internal/specific/exec"
)

// Tests for rules that upstream implements by shelling out to the underlying
// tool (apt --help, docker, go, grunt, react-native, yarn). The Go port
// reaches the tool via internal/specific/exec; here we mock that seam with
// the same fixture text upstream's tests inject via subprocess.Popen.
//
// `workon_doesnt_exists` is also covered: the port DIVERGES from upstream
// (port matches on the "doesn't exist" marker; upstream enumerates
// ~/.virtualenvs/). Test asserts the port's actual behavior, with the
// divergence documented at the top of that subtest.

// ---- apt_invalid_operation (S3.23) ----

const aptHelpFixture = `apt 1.0.10.2ubuntu1 for amd64 compiled on Oct  5 2015 15:55:05
Usage: apt [options] command

CLI for apt.
Basic commands:
 list - list packages based on package names
 search - search in package descriptions
 show - show package details

 update - update list of available packages

 install - install packages
 remove  - remove packages

 upgrade - upgrade the system by installing/upgrading packages
 full-upgrade - upgrade the system by removing/installing/upgrading packages

 edit-sources - edit the source information file
`

const aptGetHelpFixture = `apt 1.0.10.2ubuntu1 for amd64 compiled on Oct  5 2015 15:55:05
Usage: apt-get [options] command

apt-get is a simple command line interface for downloading and
installing packages.

Commands:
   update - Retrieve new lists of packages
   upgrade - Perform an upgrade
   install - Install new packages (pkg is libc6 not libc6.deb)
   remove - Remove packages
   autoremove - Remove automatically all unused packages
   purge - Remove packages and config files
   source - Download source archives
   build-dep - Configure build-dependencies for source packages
   dist-upgrade - Distribution upgrade, see apt-get(8)
   dselect-upgrade - Follow dselect selections
   clean - Erase downloaded archive files
   autoclean - Erase old downloaded archive files
   check - Verify that there are no broken dependencies
   changelog - Download and display the changelog for the given package
   download - Download the binary package into the current directory

Options:
`

func TestAptInvalidOperationMatch(t *testing.T) {
	cases := []struct {
		script, output string
		want           bool
	}{
		{"apt", "E: Invalid operation saerch", true},
		{"apt-get", "E: Invalid operation isntall", true},
		{"apt-cache", "E: Invalid operation rumove", true},
		{"vim", "E: Invalid operation vim", false},
		{"apt-get", "", false},
	}
	for _, tc := range cases {
		assertMatch(t, "apt_invalid_operation", cmd(tc.script, tc.output), tc.want)
	}
}

func TestAptInvalidOperationNewCommand(t *testing.T) {
	defer mockRunner(t, map[string]specexec.Result{
		"apt --help":     {Stdout: []byte(aptHelpFixture)},
		"apt-get --help": {Stdout: []byte(aptGetHelpFixture)},
	})()

	cases := []struct {
		script, output, want string
	}{
		{"apt-get isntall vim", "E: Invalid operation isntall", "apt-get install vim"},
		{"apt saerch vim", "E: Invalid operation saerch", "apt search vim"},
		{"apt uninstall vim", "E: Invalid operation uninstall", "apt remove vim"},
	}
	for _, tc := range cases {
		assertNewCommand(t, "apt_invalid_operation", cmd(tc.script, tc.output), tc.want)
	}
}

// ---- docker_not_command (S3.23) ----

const dockerHelpFixture = `Usage: docker [OPTIONS] COMMAND [arg...]

Commands:
    attach    Attach to a running container
    build     Build an image from a Dockerfile
    commit    Create a new image from a container's changes
    cp        Copy files/folders from a container's filesystem to the host path
    create    Create a new container
    diff      Inspect changes on a container's filesystem
    events    Get real time events from the server
    exec      Run a command in a running container
    export    Stream the contents of a container as a tar archive
    history   Show the history of an image
    images    List images
    import    Create a new filesystem image from the contents of a tarball
    info      Display system-wide information
    inspect   Return low-level information on a container or image
    kill      Kill a running container
    load      Load an image from a tar archive
    login     Register or log in to a Docker registry server
    logout    Log out from a Docker registry server
    logs      Fetch the logs of a container
    pause     Pause all processes within a container
    port      Lookup the public-facing port that is NAT-ed to PRIVATE_PORT
    ps        List containers
    pull      Pull an image or a repository from a Docker registry server
    push      Push an image or a repository to a Docker registry server
    rename    Rename an existing container
    restart   Restart a running container
    rm        Remove one or more containers
    rmi       Remove one or more images
    run       Run a command in a new container
    save      Save an image to a tar archive
    search    Search for an image on the Docker Hub
    start     Start a stopped container
    stats     Display a stream of a containers' resource usage statistics
    stop      Stop a running container
    tag       Tag an image into a repository
    top       Lookup the running processes of a container
    unpause   Unpause a paused container
    version   Show the Docker version information
    wait      Block until a container stops, then print their exit codes

Run 'docker COMMAND --help' for more information on a command.
`

func dockerNotCommandOutput(cmd string) string {
	return "docker: '" + cmd + "' is not a docker command.\nSee 'docker --help'."
}

func TestDockerNotCommandMatch(t *testing.T) {
	assertMatch(t, "docker_not_command",
		cmd("docker pes", dockerNotCommandOutput("pes")), true)
	assertMatch(t, "docker_not_command",
		cmd("docker ps", ""), false)
	assertMatch(t, "docker_not_command",
		cmd("cat pes", dockerNotCommandOutput("pes")), false)
}

func TestDockerNotCommandNewCommand(t *testing.T) {
	defer mockRunner(t, map[string]specexec.Result{
		"docker": {Stdout: []byte(dockerHelpFixture)},
	})()
	got := mustRule(t, "docker_not_command").GetNewCommand(
		cmd("docker pes", dockerNotCommandOutput("pes")),
	)
	want := []string{"docker ps", "docker push", "docker pause"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("docker_not_command(pes) = %v, want %v", got, want)
	}
}

// ---- go_unknown_command (S3.23) ----

const goHelpFixture = "Go is a tool for managing Go source code.\n\n" +
	"Usage:\n\n\tgo <command> [arguments]\n\n" +
	"The commands are:\n\n" +
	"\tbug         start a bug report\n" +
	"\tbuild       compile packages and dependencies\n" +
	"\tclean       remove object files and cached files\n" +
	"\trun         compile and run Go program\n" +
	"\ttest        test packages\n" +
	"\tvet         report likely mistakes in packages\n\n"

const goBuildMisspelledOutput = "go bulid: unknown command\nRun 'go help' for usage."

func TestGoUnknownCommandMatch(t *testing.T) {
	assertMatch(t, "go_unknown_command",
		cmd("go bulid", goBuildMisspelledOutput), true)
	assertMatch(t, "go_unknown_command",
		cmd("go run", "go run: no go files listed"), false)
}

func TestGoUnknownCommandNewCommand(t *testing.T) {
	defer mockRunner(t, map[string]specexec.Result{
		"go": {Stderr: []byte(goHelpFixture)},
	})()
	assertNewCommand(t, "go_unknown_command",
		cmd("go bulid", goBuildMisspelledOutput), "go build")
}

// ---- grunt_task_not_found (S3.23) ----

const gruntHelpFixture = `Grunt: The JavaScript Task Runner (v0.4.5)

Usage
 grunt [options] [task [task ...]]

Available tasks
  default  Alias for "build" task.
  build    Build the project.
  test     Run tests.
  watch    Watch files.
  compass  Compile Sass to CSS using Compass.
  concat   Concatenate files.
  clean    Clean files and folders.
  copy     Copy files.

For more information, see http://gruntjs.com/
`

func gruntNotFoundOutput(task string) string {
	return "Warning: Task \"" + task + "\" not found. Use --force to continue.\n"
}

func TestGruntTaskNotFoundMatch(t *testing.T) {
	defer mockRunner(t, map[string]specexec.Result{
		"grunt --help": {Stdout: []byte(gruntHelpFixture)},
	})()
	assertMatch(t, "grunt_task_not_found",
		cmd("grunt defualt", gruntNotFoundOutput("defualt")), true)
	assertMatch(t, "grunt_task_not_found",
		cmd("grunt buld:css", gruntNotFoundOutput("buld:css")), true)
	assertMatch(t, "grunt_task_not_found",
		cmd("npm nuild", gruntNotFoundOutput("nuild")), false)
	assertMatch(t, "grunt_task_not_found",
		cmd("grunt rm", ""), false)
}

func TestGruntTaskNotFoundNewCommand(t *testing.T) {
	defer mockRunner(t, map[string]specexec.Result{
		"grunt --help": {Stdout: []byte(gruntHelpFixture)},
	})()
	assertNewCommand(t, "grunt_task_not_found",
		cmd("grunt defualt", gruntNotFoundOutput("defualt")), "grunt default")
	assertNewCommand(t, "grunt_task_not_found",
		cmd("grunt cmpass:all", gruntNotFoundOutput("cmpass:all")), "grunt compass:all")
	assertNewCommand(t, "grunt_task_not_found",
		cmd("grunt cmpass:all --color", gruntNotFoundOutput("cmpass:all")),
		"grunt compass:all --color")
}

// ---- react_native_command_unrecognized (S3.23) ----

const reactNativeHelpFixture = `Scanning 615 folders for symlinks…

  Usage: react-native [options] [command]

  Commands:

    start [options]                    starts the webserver
    run-ios [options]                  builds your app and starts it on iOS simulator
    run-android [options]              builds your app and starts it on a connected Android emulator or device
    bundle [options]                   builds the javascript bundle for offline use
    unbundle [options]                 builds javascript as "unbundle" for offline use
    link [options] [packageName]       links all native dependencies
    unlink [options] <packageName>     unlink native dependency
    install [options] <packageName>    install and link native dependencies
    log-android [options]              starts adb logcat
    log-ios [options]                  starts iOS device syslog tail
`

func reactNativeOutput(cmd string) string { return "Unrecognized command '" + cmd + "'" }

func TestReactNativeCommandUnrecognizedMatch(t *testing.T) {
	assertMatch(t, "react_native_command_unrecognized",
		cmd("react-native star", reactNativeOutput("star")), true)
	assertMatch(t, "react_native_command_unrecognized",
		cmd("react-native android-logs", reactNativeOutput("android-logs")), true)
	assertMatch(t, "react_native_command_unrecognized",
		cmd("gradle star", reactNativeOutput("star")), false)
	assertMatch(t, "react_native_command_unrecognized",
		cmd("react-native start", ""), false)
}

func TestReactNativeCommandUnrecognizedNewCommand(t *testing.T) {
	defer mockRunner(t, map[string]specexec.Result{
		"react-native --help": {Stdout: []byte(reactNativeHelpFixture)},
	})()
	got := mustRule(t, "react_native_command_unrecognized").GetNewCommand(
		cmd("react-native star", reactNativeOutput("star")),
	)
	if len(got) == 0 || got[0] != "react-native start" {
		t.Errorf("react_native(star)[0] = %q, want %q", first(got), "react-native start")
	}

	got = mustRule(t, "react_native_command_unrecognized").GetNewCommand(
		cmd("react-native logsandroid -f", reactNativeOutput("logsandroid")),
	)
	if len(got) == 0 || got[0] != "react-native log-android -f" {
		t.Errorf("react_native(logsandroid -f)[0] = %q, want %q", first(got), "react-native log-android -f")
	}
}

// ---- yarn_command_not_found (S3.23) ----

const yarnHelpFixture = `

  Usage: yarn [command] [flags]

  Commands:

    - access
    - add
    - bin
    - cache
    - check
    - clean
    - config
    - generate-lock-entry
    - global
    - import
    - info
    - init
    - install
    - licenses
    - link
    - list
    - login
    - logout
    - outdated
    - owner
    - pack
    - publish
    - remove
    - run
    - tag
    - team
    - unlink
    - upgrade
    - upgrade-interactive
    - version
    - versions
    - why

  Run yarn help COMMAND for more information on specific commands.
`

func yarnNotFoundOutput(cmd string) string { return "error Command \"" + cmd + "\" not found.\n" }

func TestYarnCommandNotFoundMatch(t *testing.T) {
	assertMatch(t, "yarn_command_not_found",
		cmd("yarn whyy webpack", yarnNotFoundOutput("whyy")), true)
	assertMatch(t, "yarn_command_not_found",
		cmd("npm nuild", yarnNotFoundOutput("nuild")), false)
	assertMatch(t, "yarn_command_not_found",
		cmd("yarn install", ""), false)
}

func TestYarnCommandNotFoundNewCommand(t *testing.T) {
	defer mockRunner(t, map[string]specexec.Result{
		"yarn --help": {Stdout: []byte(yarnHelpFixture)},
	})()
	got := mustRule(t, "yarn_command_not_found").GetNewCommand(
		cmd("yarn whyy webpack", yarnNotFoundOutput("whyy")),
	)
	if first(got) != "yarn why webpack" {
		t.Errorf("yarn whyy → %q, want %q", first(got), "yarn why webpack")
	}

	got = mustRule(t, "yarn_command_not_found").GetNewCommand(
		cmd("yarn require lodash", yarnNotFoundOutput("require")),
	)
	if first(got) != "yarn add lodash" {
		t.Errorf("yarn require → %q, want %q", first(got), "yarn add lodash")
	}
}

// ---- workon_doesnt_exists (S3.23, DIVERGENT) ----
//
// Upstream matches every `workon X` invocation and consults
// ~/.virtualenvs/ for close-matching environment names; misses fall through
// to `mkvirtualenv X`. The Go port instead keys off the literal
// "doesn't exist" marker in stderr and always emits the `mkvirtualenv`
// fallback. Tests below assert the port's behavior; the upstream's
// close-match path is not reproduced here pending a Phase 4-style
// virtualenv enumerator (out of scope for this PR).

func TestWorkonDoesntExistsMatch(t *testing.T) {
	assertMatch(t, "workon_doesnt_exists",
		cmd("workon zzzz", "ERROR: Environment 'zzzz' doesn't exist."), true)
	assertMatch(t, "workon_doesnt_exists",
		cmd("workon zzzz", ""), false)
	assertMatch(t, "workon_doesnt_exists",
		cmd("work on zzzz", "doesn't exist"), false)
}

func TestWorkonDoesntExistsNewCommand(t *testing.T) {
	got := mustRule(t, "workon_doesnt_exists").GetNewCommand(
		cmd("workon zzzz", "ERROR: Environment 'zzzz' doesn't exist."),
	)
	want := []string{"mkvirtualenv zzzz && workon zzzz"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("workon_doesnt_exists = %v, want %v", got, want)
	}
}

// ---- helpers ----

func first(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	return ss[0]
}
