# TODO — thefuck test suite compliance

Goal: pass every test in upstream `thefuck/tests/rules/*.py` against this Go port.

## Current state

- 129 tests passing in `internal/rules/` (porting upstream rule tests one-by-one).
- Test harness lives in `internal/rules/rules_test.go` with `assertMatch`,
  `assertNewCommand`, `assertNewCommands`, `assertNewCommandIn` helpers.
- Ported tests are split across `simple_test.go`, `git_test.go`, `tools_test.go`,
  `batch3_test.go`.

## Rules still to port (tests not yet written)

These mostly require filesystem or PATH manipulation in tests (tmpdir + chdir,
or `t.Setenv("PATH", ...)`). The rules themselves already exist in Go.

- `cat_dir` — needs isdir check
- `cd_correction` — needs real subdir listing
- `chmod_x` — needs file exists + non-exec mode
- `dirty_untar`, `dirty_unzip` — need real tar/zip archives
- `fix_file` — needs file exists + `$EDITOR`
- `git_add` — needs `Path.exists` mock
- `git_clone_missing` — needs controlled `$PATH`
- `git_rebase_merge_dir` — known divergence (ordering; see below)
- `gradle_wrapper` — needs `./gradlew` file + `$PATH`
- `grep_arguments_order` — needs real file
- `has_exists_script` — needs real file
- `ln_s_order` — needs real file
- `missing_space_before_subcommand` — needs controlled `$PATH`
- `no_command` — needs controlled `$PATH`
- `prove_recursively` — needs isdir
- `scm_correction` — needs `.git`/`.hg` dir
- `sudo_command_from_user_path` — needs controlled `$PATH`
- `wrong_hyphen_before_subcommand` — needs controlled `$PATH`

## Known divergences from upstream (tests will fail as-is)

These rules diverge because the Go port lacks the subprocess/shell-history
infrastructure that thefuck uses. Either add the infrastructure or accept the
gap (with a comment in the rule).

- `apt_invalid_operation` — thefuck calls `apt --help`; port uses a static op list
- `dnf_no_such_command` — thefuck calls `dnf --help`; port uses static
- `docker_not_command` — thefuck calls `docker --help`; port uses static
- `fab_command_not_found` — thefuck parses fab's `-l` output; port parses stderr
- `gem_unknown_command` — thefuck calls `gem help commands`; port uses static
- `go_unknown_command` — thefuck calls `go help`; port uses static
- `gradle_no_task` — thefuck calls `gradle tasks`; port uses static
- `grunt_task_not_found` — thefuck calls `grunt --help`; port uses static
- `gulp_not_task` — thefuck calls gulp; port uses static
- `history` — not implemented (no shell-history integration)
- `ifconfig_device_not_found` — thefuck enumerates interfaces; port uses static
- `npm_missing_script` — thefuck calls `npm run-script`; not implemented
- `npm_run_script` — thefuck calls `npm run-script`; not implemented
- `pacman`, `pacman_not_found` — thefuck calls `pkgfile`; not implemented
- `path_from_history` — thefuck uses shell history; simplified fallback
- `react_native_command_unrecognized` — thefuck calls `react-native --help`
- `workon_doesnt_exists` — port logic diverges; needs rewrite against
  `~/.virtualenvs/` enumeration to match upstream
- `yarn_command_not_found` — thefuck calls `yarn help`; port uses static list
- `yum_invalid_operation` — thefuck calls `yum --help`; port uses static

## Rule fixes already made while porting tests

- `GetCloseMatches` in `internal/utils/utils.go`: arg order now matches
  Python's `difflib.get_close_matches` (seq1=candidate, seq2=word — the
  algorithm is asymmetric), plus tie-break on string-desc to match
  `heapq.nlargest` tuple comparison.
- `brew_reinstall`: removed Perl-only regex (Go's RE2 rejects `(?!…)`); now
  uses plain `strings.Contains`.
- `choco_install`: removed dead-code `if` blocks from the Python-to-Go port.
- `grunt_task_not_found`: replaced no-op `strings.Replace(x, x, x)` with a
  static task list.
- `npm_wrong_command`: now parses commands from npm's own help output instead
  of a static list, matching upstream.
- `open`: returns either the URL-prefixed form or the touch/mkdir pair,
  not both at once.
- `tmux`: regex captures the full candidate list (`(.*)` instead of `([^ ]*)`).
- `touch`: regex now matches BSD output (no surrounding quotes).
- `whois`: complete rewrite to match upstream (strip URL scheme or
  recursive subdomain stripping; was stubbed to an unrelated `-h whois.*` form).

## Infrastructure still needed

- Subprocess invocation helpers for rules that call external tools
  (`apt --help`, `gem help commands`, etc.) with test-time mocking seams.
- Shell-history reader for `no_command`, `history`, `path_from_history`.
- `ifconfig`/interface-enumeration helper.
- `pkgfile` integration for the pacman rules.

## Reference

Upstream tests live at `/tmp/thefuck-ref/tests/rules/test_*.py` (cloned during
this session). Re-clone with:

```sh
git clone --depth 1 https://github.com/nvbn/thefuck.git /tmp/thefuck-ref
```
