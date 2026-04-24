package rules

import "testing"

// ---- sudo ----
func TestSudo(t *testing.T) {
	matches := []string{
		"Permission denied",
		"permission denied",
		"npm ERR! Error: EACCES, unlink",
		"requested operation requires superuser privilege",
		"need to be root",
		"need root",
		"shutdown: NOT super-user",
		"Error: This command has to be run with superuser privileges (under the root user on most systems).",
		"updatedb: can not open a temporary file for `/var/lib/mlocate/mlocate.db",
		"must be root",
		"You don't have access to the history DB.",
		"error: [Errno 13] Permission denied: '/usr/local/lib/python2.7/dist-packages/ipaddr.py'",
	}
	for _, out := range matches {
		assertMatch(t, "sudo", cmd("", out), true)
	}
	assertMatch(t, "sudo", cmd("", ""), false)
	assertMatch(t, "sudo", cmd("sudo ls", "Permission denied"), false)

	assertNewCommand(t, "sudo", cmd("ls", ""), "sudo ls")
	assertNewCommand(t, "sudo", cmd("echo a > b", ""), `sudo sh -c "echo a > b"`)
	assertNewCommand(t, "sudo", cmd(`echo "a" >> b`, ""), `sudo sh -c "echo \"a\" >> b"`)
	assertNewCommand(t, "sudo", cmd("mkdir && touch a", ""), `sudo sh -c "mkdir && touch a"`)
}

// ---- pip_install ----
func TestPipInstall(t *testing.T) {
	r1 := "\n    Could not install packages due to an EnvironmentError: [Errno 13] Permission denied: '/Library/Python/2.7/site-packages/entrypoints.pyc'\nConsider using the `--user` option or check the permissions.\n"
	assertMatch(t, "pip_install", cmd("pip install -r requirements.txt", r1), true)
	r2 := "Successfully installed bacon-0.3.1"
	assertMatch(t, "pip_install", cmd("pip install bacon", r2), false)
	assertNewCommand(t, "pip_install", cmd("pip install -r requirements.txt", ""),
		"pip install --user -r requirements.txt")
	assertNewCommand(t, "pip_install", cmd("pip install bacon", ""), "pip install --user bacon")
	assertNewCommand(t, "pip_install", cmd("pip install --user -r requirements.txt", ""),
		"sudo pip install -r requirements.txt")
}

// ---- pip_unknown_command ----
func TestPipUnknownCommand(t *testing.T) {
	out := `ERROR: unknown command "instatl" - maybe you meant "install"`
	assertMatch(t, "pip_unknown_command", cmd("pip instatl", out), true)
	assertMatch(t, "pip_unknown_command", cmd("pip i", `ERROR: unknown command "i"`), false)
	assertNewCommand(t, "pip_unknown_command", cmd("pip instatl", out), "pip install")
	outUn := `ERROR: unknown command "un+install" - maybe you meant "uninstall"`
	assertNewCommand(t, "pip_unknown_command", cmd("pip un+install thefuck", outUn), "pip uninstall thefuck")
}

// ---- terraform_init ----
func TestTerraformInit(t *testing.T) {
	for _, tc := range []struct{ s, o string }{
		{"terraform plan", "Error: Initialization required. Please see the error message above."},
		{"terraform plan", `This module is not yet installed. Run "terraform init" to install all modules required by this configuration.`},
		{"terraform apply", "Error: Initialization required. Please see the error message above."},
		{"terraform apply", `This module is not yet installed. Run "terraform init" to install all modules required by this configuration.`},
	} {
		assertMatch(t, "terraform_init", cmd(tc.s, tc.o), true)
	}
	for _, tc := range []struct{ s, o string }{
		{"terraform --version", "Terraform v0.12.2"},
		{"terraform plan", "No changes. Infrastructure is up-to-date."},
		{"terraform apply", "Apply complete! Resources: 0 added, 0 changed, 0 destroyed."},
	} {
		assertMatch(t, "terraform_init", cmd(tc.s, tc.o), false)
	}
	assertNewCommand(t, "terraform_init", cmd("terraform plan", ""), "terraform init && terraform plan")
	assertNewCommand(t, "terraform_init", cmd("terraform apply", ""), "terraform init && terraform apply")
}

// ---- terraform_no_command ----
func TestTerraformNoCommand(t *testing.T) {
	assertMatch(t, "terraform_no_command",
		cmd("terraform appyl", `Terraform has no command named "appyl". Did you mean "apply"?`), true)
	assertMatch(t, "terraform_no_command",
		cmd("terraform destory", `Terraform has no command named "destory". Did you mean "destroy"?`), true)
	assertMatch(t, "terraform_no_command", cmd("terraform --version", "Terraform v0.12.2"), false)
	assertNewCommand(t, "terraform_no_command",
		cmd("terraform appyl", `Terraform has no command named "appyl". Did you mean "apply"?`), "terraform apply")
	assertNewCommand(t, "terraform_no_command",
		cmd("terraform destory --some-other-option",
			`Terraform has no command named "destory". Did you mean "destroy"?`),
		"terraform destroy --some-other-option")
}

// ---- vagrant_up ----
func TestVagrantUp(t *testing.T) {
	sshOut := "VM must be running to open SSH connection. Run `vagrant up`\nto start the virtual machine."
	rdpOut := "VM must be created before running this command. Run `vagrant up` first."
	assertMatch(t, "vagrant_up", cmd("vagrant ssh", sshOut), true)
	assertMatch(t, "vagrant_up", cmd("vagrant ssh devbox", sshOut), true)
	assertMatch(t, "vagrant_up", cmd("vagrant rdp", rdpOut), true)
	assertMatch(t, "vagrant_up", cmd("vagrant rdp devbox", rdpOut), true)
	assertMatch(t, "vagrant_up", cmd("vagrant ssh", ""), false)
	assertMatch(t, "vagrant_up", cmd("", ""), false)
	assertNewCommand(t, "vagrant_up", cmd("vagrant ssh", sshOut), "vagrant up && vagrant ssh")
	assertNewCommands(t, "vagrant_up", cmd("vagrant ssh devbox", sshOut),
		[]string{"vagrant up devbox && vagrant ssh devbox", "vagrant up && vagrant ssh devbox"})
}

// ---- nixos_cmd_not_found ----
func TestNixosCmdNotFound(t *testing.T) {
	assertMatch(t, "nixos_cmd_not_found", cmd("vim", "nix-env -iA nixos.vim"), true)
	assertMatch(t, "nixos_cmd_not_found", cmd("vim", ""), false)
	assertMatch(t, "nixos_cmd_not_found", cmd("", ""), false)
	assertNewCommand(t, "nixos_cmd_not_found", cmd("vim", "nix-env -iA nixos.vim"), "nix-env -iA nixos.vim && vim")
}

// ---- hostscli ----
func TestHostscli(t *testing.T) {
	out := "\nhostscli.errors.WebsiteImportError:\n\nNo Domain list found for website: a_website_that_does_not_exist\n\nPlease raise an Issue here: https://github.com/dhilipsiva/hostscli/issues/new\nif you think we should add domains for this website.\n\ntype `hostscli websites` to see a list of websites that you can block/unblock\n"
	assertMatch(t, "hostscli", cmd("hostscli block a_website_that_does_not_exist", out), true)
	assertNewCommands(t, "hostscli", cmd("hostscli block a_website_that_does_not_exist", out), []string{"hostscli websites"})
}

// ---- lein_not_task ----
func TestLeinNotTask(t *testing.T) {
	out := "'rpl' is not a task. See 'lein help'.\n\nDid you mean this?\n         repl\n         jar\n"
	assertMatch(t, "lein_not_task", cmd("lein rpl", out), true)
	assertMatch(t, "lein_not_task", cmd("ls", out), false)
	assertNewCommands(t, "lein_not_task", cmd("lein rpl --help", out), []string{"lein repl --help", "lein jar --help"})
}

// ---- aws_cli ----
func TestAwsCli(t *testing.T) {
	misCmd := "usage: aws [options] <command>\naws: error: argument command: Invalid choice\n\nInvalid choice: 'dynamdb', maybe you meant:\n\n  * dynamodb\n"
	misSub := "usage: aws [options] <command>\naws: error: argument operation: Invalid choice\n\nInvalid choice: 'scn', maybe you meant:\n\n  * scan\n"
	misMult := "usage: aws [options] <command>\naws: error: argument operation: Invalid choice\n\nInvalid choice: 't-item', maybe you meant:\n\n  * put-item\n  * get-item\n"
	assertMatch(t, "aws_cli", cmd("aws dynamdb scan", misCmd), true)
	assertMatch(t, "aws_cli", cmd("aws dynamodb scn", misSub), true)
	assertMatch(t, "aws_cli", cmd("aws dynamodb t-item", misMult), true)
	assertNewCommands(t, "aws_cli", cmd("aws dynamdb scan", misCmd), []string{"aws dynamodb scan"})
	assertNewCommands(t, "aws_cli", cmd("aws dynamodb scn", misSub), []string{"aws dynamodb scan"})
	assertNewCommands(t, "aws_cli", cmd("aws dynamodb t-item", misMult), []string{"aws dynamodb put-item", "aws dynamodb get-item"})
}

// ---- az_cli ----
func TestAzCli(t *testing.T) {
	misCmd := "az: 'providers' is not in the 'az' command group. See 'az --help'.\n\nThe most similar choice to 'providers' is:\n    provider\n"
	misSub := "az provider: 'lis' is not in the 'az provider' command group. See 'az provider --help'.\n\nThe most similar choice to 'lis' is:\n    list\n"
	assertMatch(t, "az_cli", cmd("az providers", misCmd), true)
	assertMatch(t, "az_cli", cmd("az provider lis", misSub), true)
	assertNewCommands(t, "az_cli", cmd("az providers list", misCmd), []string{"az provider list"})
	assertNewCommands(t, "az_cli", cmd("az provider lis", misSub), []string{"az provider list"})
}

// ---- rails_migrations_pending ----
func TestRailsMigrationsPending(t *testing.T) {
	devOut := "\nMigrations are pending. To resolve this issue, run:\n\n        rails db:migrate RAILS_ENV=development\n"
	testOut := "\nMigrations are pending. To resolve this issue, run:\n\n        bin/rails db:migrate RAILS_ENV=test\n"
	assertMatch(t, "rails_migrations_pending", cmd("", devOut), true)
	assertMatch(t, "rails_migrations_pending", cmd("", testOut), true)
	assertMatch(t, "rails_migrations_pending",
		cmd("Environment data not found in the schema. To resolve this issue, run: \n\n", ""), false)
	assertNewCommand(t, "rails_migrations_pending", cmd("bin/rspec", devOut),
		"rails db:migrate RAILS_ENV=development && bin/rspec")
	assertNewCommand(t, "rails_migrations_pending", cmd("bin/rspec", testOut),
		"bin/rails db:migrate RAILS_ENV=test && bin/rspec")
}

// ---- gulp_not_task ----
func TestGulpNotTask(t *testing.T) {
	out := "[00:41:11] Using gulpfile gulpfile.js\n[00:41:11] Task 'srve' is not in your gulpfile\n[00:41:11] Please check the documentation for proper gulpfile formatting\n"
	assertMatch(t, "gulp_not_task", cmd("gulp srve", out), true)
	assertMatch(t, "gulp_not_task", cmd("gulp serve", ""), false)
	assertMatch(t, "gulp_not_task", cmd("cat srve", out), false)
	// get_new_command uses a static task list (divergence from thefuck)
}

// ---- grep_recursive ----
func TestGrepRecursive(t *testing.T) {
	assertMatch(t, "grep_recursive", cmd("grep blah .", "grep: .: Is a directory"), true)
	assertMatch(t, "grep_recursive", cmd("grep café .", "grep: .: Is a directory"), true)
	assertMatch(t, "grep_recursive", cmd("", ""), false)
	assertNewCommand(t, "grep_recursive", cmd("grep blah .", ""), "grep -r blah .")
	assertNewCommand(t, "grep_recursive", cmd("grep café .", ""), "grep -r café .")
}

// ---- systemctl ----
func TestSystemctl(t *testing.T) {
	assertMatch(t, "systemctl", cmd("systemctl nginx start", "Unknown operation 'nginx'."), true)
	assertMatch(t, "systemctl", cmd("sudo systemctl nginx start", "Unknown operation 'nginx'."), true)
	assertMatch(t, "systemctl", cmd("systemctl start nginx", ""), false)
	assertMatch(t, "systemctl", cmd("sudo systemctl nginx", "Unknown operation 'nginx'."), false)
	assertMatch(t, "systemctl", cmd("systemctl nginx", "Unknown operation 'nginx'."), false)
	assertMatch(t, "systemctl", cmd("systemctl start wtf",
		"Failed to start wtf.service: Unit wtf.service failed to load: No such file or directory."), false)
	assertNewCommand(t, "systemctl", cmd("systemctl nginx start", ""), "systemctl start nginx")
	assertNewCommand(t, "systemctl", cmd("sudo systemctl nginx start", ""), "sudo systemctl start nginx")
}

// ---- tmux ----
func TestTmux(t *testing.T) {
	out := "ambiguous command: list, could be: list-buffers, list-clients, list-commands, list-keys, list-panes, list-sessions, list-windows"
	assertMatch(t, "tmux", cmd("tmux list", out), true)
	assertNewCommands(t, "tmux", cmd("tmux list", out),
		[]string{"tmux list-keys", "tmux list-panes", "tmux list-windows"})
}

// ---- long_form_help ----
func TestLongFormHelp(t *testing.T) {
	assertMatch(t, "long_form_help", cmd("grep -h", "Try 'grep --help' for more information."), true)
	assertMatch(t, "long_form_help", cmd("", ""), false)
	assertNewCommand(t, "long_form_help", cmd("grep -h", ""), "grep --help")
	assertNewCommand(t, "long_form_help", cmd("tar -h", ""), "tar --help")
	assertNewCommand(t, "long_form_help", cmd("docker run -h", ""), "docker run --help")
	assertNewCommand(t, "long_form_help", cmd("cut -h", ""), "cut --help")
}

// ---- adb_unknown_command ----
func TestAdbUnknownCommand(t *testing.T) {
	out := "Android Debug Bridge version 1.0.31\n\n -d                            - directs command\n -e                            - directs command\n -s <specific device>          - directs command"
	assertMatch(t, "adb_unknown_command", cmd("adb lgcat", out), true)
	assertMatch(t, "adb_unknown_command", cmd("adb puhs", out), true)
	assertMatch(t, "adb_unknown_command", cmd("git branch foo", ""), false)
	assertMatch(t, "adb_unknown_command", cmd("abd push", ""), false)
	assertNewCommand(t, "adb_unknown_command",
		cmd("adb puhs test.bin /sdcard/test.bin", out), "adb push test.bin /sdcard/test.bin")
	assertNewCommand(t, "adb_unknown_command",
		cmd("adb -d logcatt", out), "adb -d logcat")
	assertNewCommand(t, "adb_unknown_command",
		cmd("adb -e reboott", out), "adb -e reboot")
}

// ---- cd_cs ----
func TestCdCs(t *testing.T) {
	assertMatch(t, "cd_cs", cmd("cs", "cs: command not found"), true)
	assertMatch(t, "cd_cs", cmd("cs /etc/", "cs: command not found"), true)
	assertNewCommand(t, "cd_cs", cmd("cs /etc/", "cs: command not found"), "cd /etc/")
}

// ---- unknown_command ----
func TestUnknownCommand(t *testing.T) {
	out := "ls: Unknown command\nDid you mean -ls?  This command begins with a dash."
	assertMatch(t, "unknown_command", cmd("./bin/hdfs dfs ls", out), true)
	assertMatch(t, "unknown_command", cmd("hdfs dfs ls", out), true)
}

// ---- switch_lang (partial — Cyrillic only since Go port is minimal) ----
func TestSwitchLangCyrillic(t *testing.T) {
	assertMatch(t, "switch_lang", cmd("фзе-пуе", "command not found: фзе-пуе"), true)
	assertNewCommand(t, "switch_lang", cmd("фзе-пуе штыефдд мшь", ""), "apt-get install vim")
}

// ---- heroku_multiple_apps ----
func TestHerokuMultipleApps(t *testing.T) {
	out := `
 ▸    Multiple apps in git remotes
 ▸    Heroku remotes in repo:
 ▸    myapp (heroku)
 ▸    myapp-dev (heroku-dev)
 ▸
 ▸    https://devcenter.heroku.com/articles/multiple-environments
`
	assertMatch(t, "heroku_multiple_apps", cmd("heroku pg", out), true)
	assertMatch(t, "heroku_multiple_apps", cmd("heroku pg", "Continuous Protection: Off"), false)
	assertNewCommands(t, "heroku_multiple_apps", cmd("heroku pg", out),
		[]string{"heroku pg --app myapp", "heroku pg --app myapp-dev"})
}

// ---- heroku_not_command ----
func TestHerokuNotCommand(t *testing.T) {
	out := "\n ▸    log is not a heroku command.\n ▸    Perhaps you meant logs?\n ▸    Run heroku _ to run heroku logs.\n ▸    Run heroku help for a list of available commands."
	assertMatch(t, "heroku_not_command", cmd("heroku log", out), true)
	assertMatch(t, "heroku_not_command", cmd("cat log", out), false)
	assertNewCommand(t, "heroku_not_command", cmd("heroku log", out), "heroku logs")
}

// ---- docker_image_being_used_by_container ----
func TestDockerImageBeingUsedByContainer(t *testing.T) {
	out := "Error response from daemon: conflict: unable to delete cd809b04b6ff (cannot be forced) - image is being used by running container e5e2591040d1"
	assertMatch(t, "docker_image_being_used_by_container", cmd("docker image rm -f cd809b04b6ff", out), true)
	assertMatch(t, "docker_image_being_used_by_container", cmd("docker image rm -f cd809b04b6ff", "bash: docker: command not found"), false)
	assertMatch(t, "docker_image_being_used_by_container", cmd("git image rm -f cd809b04b6ff", out), false)
	assertNewCommand(t, "docker_image_being_used_by_container", cmd("docker image rm -f cd809b04b6ff", out),
		"docker container rm -f e5e2591040d1 && docker image rm -f cd809b04b6ff")
}

// ---- docker_login ----
func TestDockerLogin(t *testing.T) {
	err1 := "\n    Sending build context to Docker daemon  118.8kB\nStep 1/6 : FROM foo/bar:fdb7c6d\npull access denied for foo/bar, repository does not exist or may require 'docker login'\n"
	err2 := "\n    The push refers to repository [artifactory:9090/foo/bar]\npush access denied for foo/bar, repository does not exist or may require 'docker login'\n"
	err3 := "\n    docker push artifactory:9090/foo/bar:fdb7c6d\nThe push refers to repository\n9c29c7ad209d: Preparing\n"
	assertMatch(t, "docker_login", cmd("docker build -t artifactory:9090/foo/bar:fdb7c6d .", err1), true)
	assertMatch(t, "docker_login", cmd("docker push artifactory:9090/foo/bar:fdb7c6d", err2), true)
	assertMatch(t, "docker_login", cmd("docker push artifactory:9090/foo/bar:fdb7c6d", err3), false)
	assertNewCommand(t, "docker_login", cmd("docker build -t artifactory:9090/foo/bar:fdb7c6d .", ""),
		"docker login && docker build -t artifactory:9090/foo/bar:fdb7c6d .")
	assertNewCommand(t, "docker_login", cmd("docker push artifactory:9090/foo/bar:fdb7c6d", ""),
		"docker login && docker push artifactory:9090/foo/bar:fdb7c6d")
}

// ---- mvn_no_command ----
func TestMvnNoCommand(t *testing.T) {
	out := "No goals have been specified for this build"
	assertMatch(t, "mvn_no_command", cmd("mvn", out), true)
	assertNewCommands(t, "mvn_no_command", cmd("mvn", out), []string{"mvn clean package", "mvn clean install"})
}

// ---- pacman_invalid_option ----
func TestPacmanInvalidOption(t *testing.T) {
	for _, o := range "surqfdvt" {
		assertMatch(t, "pacman_invalid_option",
			cmd("pacman -"+string(o)+"v meat", "error: invalid option '-"), true)
	}
	for _, o := range "azxcbnm" {
		assertMatch(t, "pacman_invalid_option",
			cmd("pacman -"+string(o)+"v meat", "error: invalid option '-"), false)
	}
	for _, o := range "surqfdvt" {
		assertNewCommand(t, "pacman_invalid_option",
			cmd("pacman -"+string(o)+"v meat", ""),
			"pacman -"+string(o-32)+"v meat") // uppercase
	}
}

// ---- port_already_in_use ----
func TestPortAlreadyInUse(t *testing.T) {
	for _, out := range []string{
		"bind on address ('127.0.0.1', 8080)",
		"Unable to bind 0.0.0.0:8080",
		"can't listen on port 8080",
		"listen EADDRINUSE 0.0.0.0:8080",
	} {
		assertMatch(t, "port_already_in_use", cmd("python server.py", out), true)
	}
}

// ---- conda_mistype already tested ----

// ---- gem_unknown_command ----
func TestGemUnknownCommand(t *testing.T) {
	out := "\nERROR:  While executing gem ... (Gem::CommandLineError)\n    Unknown command isntall\n"
	assertMatch(t, "gem_unknown_command", cmd("gem isntall jekyll", out), true)
	assertMatch(t, "gem_unknown_command", cmd("gem install jekyll", ""), false)
}

// ---- yum_invalid_operation ----
func TestYumInvalidOperation(t *testing.T) {
	mkOut := func(cmd string) string {
		return "Loaded plugins: extras_suggestions, langpacks, priorities, update-motd\nNo such command: " + cmd + ". Please use /usr/bin/yum --help\n"
	}
	for _, c := range []string{"saerch", "uninstall"} {
		assertMatch(t, "yum_invalid_operation", cmd("yum "+c, mkOut(c)), true)
	}
	for _, tc := range []struct{ s, o string }{
		{"vim", ""},
		{"yum", "Usage: yum [options] COMMAND"},
	} {
		assertMatch(t, "yum_invalid_operation", cmd(tc.s, tc.o), false)
	}
	assertNewCommand(t, "yum_invalid_operation", cmd("yum uninstall vim", mkOut("uninstall")), "yum remove vim")
}

// ---- dnf_no_such_command ----
func TestDnfNoSuchCommand(t *testing.T) {
	invalidCmd := func(c string) string {
		return "No such command: " + c + ". Please use /usr/bin/dnf --help\nIt could be a DNF plugin command, try: \"dnf install 'dnf-command(" + c + ")'\"\n"
	}
	for _, o := range []string{invalidCmd("saerch"), invalidCmd("isntall")} {
		assertMatch(t, "dnf_no_such_command", cmd("dnf", o), true)
	}
	assertMatch(t, "dnf_no_such_command", cmd("pip", invalidCmd("isntall")), false)
	assertMatch(t, "dnf_no_such_command", cmd("vim", ""), false)
}

// ---- composer_not_command ----
func TestComposerNotCommand(t *testing.T) {
	notCmd := "\n\n                                    \n  [InvalidArgumentException]        \n  Command \"udpate\" is not defined.  \n  Did you mean this?                \n      update                        \n                                    \n\n\n"
	oneOf := "\n\n                                   \n  [InvalidArgumentException]       \n  Command \"pdate\" is not defined.  \n  Did you mean one of these?       \n      selfupdate                   \n      self-update                  \n      update                       \n                                   \n\n\n"
	reqInstead := `Invalid argument package. Use "composer require package" instead to add packages to your composer.json.`
	assertMatch(t, "composer_not_command", cmd("composer udpate", notCmd), true)
	assertMatch(t, "composer_not_command", cmd("composer pdate", oneOf), true)
	assertMatch(t, "composer_not_command", cmd("composer install package", reqInstead), true)
	assertMatch(t, "composer_not_command", cmd("ls update", notCmd), false)
	assertNewCommand(t, "composer_not_command", cmd("composer udpate", notCmd), "composer update")
	assertNewCommand(t, "composer_not_command", cmd("composer pdate", oneOf), "composer selfupdate")
	assertNewCommand(t, "composer_not_command", cmd("composer install package", reqInstead), "composer require package")
}

// ---- mercurial ----
func TestMercurial(t *testing.T) {
	cases := []struct{ script, out, want string }{
		{"hg base", "hg: unknown command 'base'\n(did you mean one of blame, phase, rebase?)", "hg rebase"},
		{"hg branchch", "hg: unknown command 'branchch'\n(did you mean one of branch, branches?)", "hg branch"},
		{"hg vert", "hg: unknown command 'vert'\n(did you mean one of revert?)", "hg revert"},
		{"hg lgo -r tip", "hg: command 're' is ambiguous:\n(did you mean one of log?)", "hg log -r tip"},
		{"hg rerere", "hg: unknown command 'rerere'\n(did you mean one of revert?)", "hg revert"},
	}
	for _, tc := range cases {
		assertMatch(t, "mercurial", cmd(tc.script, tc.out), true)
		assertNewCommand(t, "mercurial", cmd(tc.script, tc.out), tc.want)
	}
	// Not-match cases
	for _, tc := range []struct{ s, o string }{
		{"hg", "\nMercurial Distributed SCM\n\nbasic commands:"},
		{"hg asdf", "hg: unknown command 'asdf'\nMercurial Distributed SCM\n\nbasic commands:"},
		{"hg me", "\nabort: no repository found in './thefuck' (.hg not found)!"},
	} {
		assertMatch(t, "mercurial", cmd(tc.s, tc.o), false)
	}
}

// ---- ssh_known_hosts (match only) ----
func TestSshKnownHosts(t *testing.T) {
	out := "@@@@@@@@@@@@@@@@@@@@\n@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @\n@@@@@@@@@@@@@@@@@@@@\nOffending RSA key in /tmp/known:2\n"
	assertMatch(t, "ssh_known_hosts", cmd("ssh user@host", out), true)
	assertMatch(t, "ssh_known_hosts", cmd("scp something something", out), true)
	assertMatch(t, "ssh_known_hosts", cmd(out, ""), false)
	assertMatch(t, "ssh_known_hosts", cmd("notssh", out), false)
	assertMatch(t, "ssh_known_hosts", cmd("ssh", ""), false)
	assertNewCommand(t, "ssh_known_hosts", cmd("ssh user@host", out), "ssh user@host")
}

// ---- touch ----
func TestTouch(t *testing.T) {
	linuxOut := "touch: cannot touch '/a/b/c': No such file or directory"
	bsdOut := "touch: /a/b/c: No such file or directory"
	for _, out := range []string{linuxOut, bsdOut} {
		assertMatch(t, "touch", cmd("touch /a/b/c", out), true)
		assertNewCommand(t, "touch", cmd("touch /a/b/c", out), "mkdir -p /a/b && touch /a/b/c")
	}
	assertMatch(t, "touch", cmd("touch /a/b/c", ""), false)
	assertMatch(t, "touch", cmd("ls /a/b/c", linuxOut), false)
}

// ---- brew_install ----
func TestBrewInstall(t *testing.T) {
	one := `Warning: No available formula with the name "giss". Did you mean gist?`
	two := `Warning: No available formula with the name "elasticserar". Did you mean elasticsearch or elasticsearch@6?`
	three := `Warning: No available formula with the name "gitt". Did you mean git, gitg or gist?`
	already := "Warning: git-2.3.5 already installed"
	noArg := "Install a formula or cask. Additional options specific to a formula may be"
	assertMatch(t, "brew_install", cmd("brew install giss", one), true)
	assertMatch(t, "brew_install", cmd("brew install elasticserar", two), true)
	assertMatch(t, "brew_install", cmd("brew install gitt", three), true)
	assertMatch(t, "brew_install", cmd("brew install git", already), false)
	assertMatch(t, "brew_install", cmd("brew install", noArg), false)
	assertNewCommands(t, "brew_install", cmd("brew install giss", one), []string{"brew install gist"})
	assertNewCommands(t, "brew_install", cmd("brew install elasticsear", two),
		[]string{"brew install elasticsearch", "brew install elasticsearch@6"})
	assertNewCommands(t, "brew_install", cmd("brew install gitt", three),
		[]string{"brew install git", "brew install gitg", "brew install gist"})
}

// ---- brew_link ----
func TestBrewLink(t *testing.T) {
	out := "Error: Could not symlink bin/gcp\nTarget /usr/local/bin/gcp\nalready exists. You may want to remove it:\n  rm '/usr/local/bin/gcp'\n\nTo force the link and overwrite all conflicting files:\n  brew link --overwrite coreutils\n\nTo list all files that would be deleted:\n  brew link --overwrite --dry-run coreutils\n"
	for _, s := range []string{"brew link coreutils", "brew ln coreutils"} {
		assertMatch(t, "brew_link", cmd(s, out), true)
	}
	assertMatch(t, "brew_link", cmd("brew link coreutils", ""), false)
	assertNewCommand(t, "brew_link", cmd("brew link coreutils", out), "brew link --overwrite --dry-run coreutils")
}

// ---- brew_reinstall ----
func TestBrewReinstall(t *testing.T) {
	out := "Warning: thefuck 9.9 is already installed and up-to-date\nTo reinstall 9.9, run `brew reinstall thefuck`"
	assertMatch(t, "brew_reinstall", cmd("brew install thefuck", out), true)
	assertMatch(t, "brew_reinstall", cmd("brew reinstall thefuck", ""), false)
	assertMatch(t, "brew_reinstall", cmd("brew install foo", ""), false)
	assertNewCommand(t, "brew_reinstall", cmd("brew install foo", out), "brew reinstall foo")
	assertNewCommand(t, "brew_reinstall", cmd("brew install bar zap", out), "brew reinstall bar zap")
}

// ---- brew_uninstall ----
func TestBrewUninstall(t *testing.T) {
	out := "Uninstalling /usr/local/Cellar/tbb/4.4-20160916... (118 files, 1.9M)\ntbb 4.4-20160526, 4.4-20160722 are still installed.\nRemove all versions with `brew uninstall --force tbb`.\n"
	for _, s := range []string{"brew uninstall tbb", "brew rm tbb", "brew remove tbb"} {
		assertMatch(t, "brew_uninstall", cmd(s, out), true)
	}
	assertMatch(t, "brew_uninstall", cmd("brew remove gnuplot",
		"Uninstalling /usr/local/Cellar/gnuplot/5.0.4_1... (44 files, 2.3M)\n"), false)
	assertNewCommand(t, "brew_uninstall", cmd("brew uninstall tbb", out), "brew uninstall --force tbb")
}

// ---- brew_update_formula ----
func TestBrewUpdateFormula(t *testing.T) {
	out := "Error: This command updates brew itself, and does not take formula names.\nUse `brew upgrade thefuck`."
	assertMatch(t, "brew_update_formula", cmd("brew update thefuck", out), true)
	assertMatch(t, "brew_update_formula", cmd("brew upgrade foo", ""), false)
	assertMatch(t, "brew_update_formula", cmd("brew update", ""), false)
	assertNewCommand(t, "brew_update_formula", cmd("brew update foo", out), "brew upgrade foo")
	assertNewCommand(t, "brew_update_formula", cmd("brew update bar zap", out), "brew upgrade bar zap")
}

// ---- brew_cask_dependency ----
func TestBrewCaskDependency(t *testing.T) {
	out := "sshfs: OsxfuseRequirement unsatisfied!\n\nYou can install with Homebrew-Cask:\n  brew cask install osxfuse\n\nYou can download from:\n  https://osxfuse.github.io/\nError: An unsatisfied requirement failed this build."
	assertMatch(t, "brew_cask_dependency", cmd("brew install sshfs", out), true)
	assertMatch(t, "brew_cask_dependency", cmd("brew link sshfs", out), false)
	assertMatch(t, "brew_cask_dependency", cmd("cat output", out), false)
	assertMatch(t, "brew_cask_dependency", cmd("brew install sshfs", ""), false)
	assertNewCommand(t, "brew_cask_dependency", cmd("brew install sshfs", out),
		"brew cask install osxfuse && brew install sshfs")
}

// ---- yarn_alias ----
func TestYarnAlias(t *testing.T) {
	outRemove := "error Did you mean `yarn remove`?"
	outEtl := `error Command "etil" not found. Did you mean "etl"?`
	outList := "error Did you mean `yarn list`?"
	assertMatch(t, "yarn_alias", cmd("yarn rm", outRemove), true)
	assertMatch(t, "yarn_alias", cmd("yarn etil", outEtl), true)
	assertMatch(t, "yarn_alias", cmd("yarn ls", outList), true)
	assertNewCommand(t, "yarn_alias", cmd("yarn rm", outRemove), "yarn remove")
	assertNewCommand(t, "yarn_alias", cmd("yarn etil", outEtl), "yarn etl")
	assertNewCommand(t, "yarn_alias", cmd("yarn ls", outList), "yarn list")
}

// ---- yarn_command_replaced ----
func TestYarnCommandReplaced(t *testing.T) {
	mkOut := func(s string) string {
		return "error `install` has been replaced with `add` to add new dependencies. Run \"yarn add " + s + "\" instead."
	}
	for _, s := range []string{"redux", "moment", "lodash"} {
		assertMatch(t, "yarn_command_replaced", cmd("yarn install "+s, mkOut(s)), true)
		assertNewCommand(t, "yarn_command_replaced", cmd("yarn install "+s, mkOut(s)), "yarn add "+s)
	}
	assertMatch(t, "yarn_command_replaced", cmd("yarn install", ""), false)
}

// ---- yarn_help ----
func TestYarnHelp(t *testing.T) {
	out := "\n\n  Usage: yarn [command] [flags]\n\n  Visit https://yarnpkg.com/en/docs/cli/clean for documentation about this command.\n"
	assertMatch(t, "yarn_help", cmd("yarn help clean", out), true)
	assertNewCommand(t, "yarn_help", cmd("yarn help clean", out),
		"xdg-open https://yarnpkg.com/en/docs/cli/clean")
}

// ---- tsuru_login ----
func TestTsuruLogin(t *testing.T) {
	out := "Error: you're not authenticated or session has expired."
	assertMatch(t, "tsuru_login", cmd("tsuru app-list", out), true)
	assertNewCommand(t, "tsuru_login", cmd("tsuru app-list", out), "tsuru login && tsuru app-list")
}
