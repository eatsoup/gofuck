# TODO — gofuck (Go port of `thefuck`)

Goal: pass every test in upstream `thefuck/tests/rules/*.py` against this Go
port, and ship a working CLI so we can actually run corrections end-to-end.

Upstream reference clone: `/tmp/thefuck-ref` (re-clone with
`git clone --depth 1 https://github.com/nvbn/thefuck.git /tmp/thefuck-ref`).

---

## Headline numbers

- Upstream rules: **169** (plus `test.py` as `test.py.py` — odd filename).
- Upstream test files: **167** (3 are renamed; 5 rules have no upstream test).
- Go rules registered: **163** → 6 missing.
- Go tests passing: **129 test functions** covering **128 rules**.
- Rules flagged as divergent from upstream: **20**.
- CLI binary: **not yet built** — there is no `cmd/` directory.

The remainder of the work breaks down as:

| Bucket | Count | Notes |
| --- | --- | --- |
| Rules implemented + tested + agreeing with upstream | ~108 | "OK" rows below. |
| Rules implemented but missing Go tests | 34 | Mostly need fs/PATH scaffolding. |
| Rules implemented but diverging from upstream | 20 | Need subprocess/history/pkgfile infra. |
| Upstream rules not yet implemented in Go | 6 | `apt_get`, `history`, `man`, `pacman`, `pacman_not_found`, `test.py`. |
| Missing top-level CLI app | 1 | No `main.go`; nothing to run by hand. |

---

## Per-rule coverage matrix

Legend:
- **Go impl** — `+` registered in `internal/rules`, `-` not yet, `?` unknown.
- **Go tests** — `+` at least one assertion in `*_test.go`, `-` none, `n/a` rule not implemented.
- **Up tests** — does upstream have a `test_<rule>.py`?
- **Status** — `OK`, `NEEDS-RULE`, `NEEDS-TEST`, `DIVERGENT`, or combinations.
  `DIVERGENT` means upstream and Go disagree about how the rule should
  behave (typically because upstream shells out and we use a static list).

| Rule | Go impl | Go tests | Up tests | Status |
| --- | --- | --- | --- | --- |
| adb_unknown_command | + | + | + | OK |
| ag_literal | + | + | + | OK |
| apt_get | - | n/a | + | NEEDS-RULE |
| apt_get_search | + | + | + | OK |
| apt_invalid_operation | + | - | + | NEEDS-TEST, DIVERGENT |
| apt_list_upgradable | + | + | + | OK |
| apt_upgrade | + | + | + | OK |
| aws_cli | + | + | + | OK |
| az_cli | + | + | + | OK |
| brew_cask_dependency | + | + | + | OK |
| brew_install | + | + | + | OK |
| brew_link | + | + | + | OK |
| brew_reinstall | + | + | + | OK |
| brew_uninstall | + | + | + | OK |
| brew_unknown_command | + | - | + | NEEDS-TEST |
| brew_update_formula | + | + | + | OK |
| cargo | + | + | - | OK |
| cargo_no_command | + | + | + | OK |
| cat_dir | + | - | + | NEEDS-TEST |
| cd_correction | + | - | + | NEEDS-TEST |
| cd_cs | + | + | + | OK |
| cd_mkdir | + | + | + | OK |
| cd_parent | + | + | + | OK |
| chmod_x | + | - | + | NEEDS-TEST |
| choco_install | + | + | + | OK |
| composer_not_command | + | + | + | OK |
| conda_mistype | + | + | + | OK |
| cp_create_destination | + | + | + | OK |
| cp_omitting_directory | + | + | + | OK |
| cpp11 | + | - | - | NEEDS-TEST |
| dirty_untar | + | - | + | NEEDS-TEST |
| dirty_unzip | + | - | + | NEEDS-TEST |
| django_south_ghost | + | + | + | OK |
| django_south_merge | + | + | + | OK |
| dnf_no_such_command | + | + | + | DIVERGENT |
| docker_image_being_used_by_container | + | + | + | OK |
| docker_login | + | + | + | OK |
| docker_not_command | + | - | + | NEEDS-TEST, DIVERGENT |
| dry | + | + | + | OK |
| fab_command_not_found | + | + | + | DIVERGENT |
| fix_alt_space | + | + | + | OK |
| fix_file | + | - | + | NEEDS-TEST |
| gem_unknown_command | + | + | + | DIVERGENT |
| git_add | + | - | + | NEEDS-TEST |
| git_add_force | + | + | + | OK |
| git_bisect_usage | + | + | + | OK |
| git_branch_0flag | + | + | + | OK |
| git_branch_delete | + | + | + | OK |
| git_branch_delete_checked_out | + | + | + | OK |
| git_branch_exists | + | + | + | OK |
| git_branch_list | + | + | + | OK |
| git_checkout | + | + | + | OK |
| git_clone_git_clone | + | + | + | OK |
| git_clone_missing | + | - | + | NEEDS-TEST |
| git_commit_add | + | + | + | OK |
| git_commit_amend | + | + | + | OK |
| git_commit_reset | + | + | + | OK |
| git_diff_no_index | + | + | + | OK |
| git_diff_staged | + | + | + | OK |
| git_fix_stash | + | + | + | OK |
| git_flag_after_filename | + | + | + | OK |
| git_help_aliased | + | + | + | OK |
| git_hook_bypass | + | + | + | OK |
| git_lfs_mistype | + | + | + | OK |
| git_main_master | + | + | + | OK |
| git_merge | + | + | + | OK |
| git_merge_unrelated | + | + | + | OK |
| git_not_command | + | + | + | OK |
| git_pull | + | - | + | NEEDS-TEST |
| git_pull_clone | + | + | + | OK |
| git_pull_uncommitted_changes | + | + | + | OK |
| git_push | + | + | + | OK |
| git_push_different_branch_names | + | + | + | OK |
| git_push_force | + | + | + | OK |
| git_push_pull | + | + | + | OK |
| git_push_without_commits | + | + | + | OK |
| git_rebase_merge_dir | + | - | + | NEEDS-TEST |
| git_rebase_no_changes | + | + | + | OK |
| git_remote_delete | + | + | + | OK |
| git_remote_seturl_add | + | + | + | OK |
| git_rm_local_modifications | + | + | + | OK |
| git_rm_recursive | + | + | + | OK |
| git_rm_staged | + | + | + | OK |
| git_stash | + | + | + | OK |
| git_stash_pop | + | + | + | OK |
| git_tag_force | + | + | + | OK |
| git_two_dashes | + | + | + | OK |
| go_run | + | + | + | OK |
| go_unknown_command | + | - | + | NEEDS-TEST, DIVERGENT |
| gradle_no_task | + | - | + | NEEDS-TEST, DIVERGENT |
| gradle_wrapper | + | - | + | NEEDS-TEST |
| grep_arguments_order | + | - | + | NEEDS-TEST |
| grep_recursive | + | + | + | OK |
| grunt_task_not_found | + | - | + | NEEDS-TEST, DIVERGENT |
| gulp_not_task | + | + | + | DIVERGENT |
| has_exists_script | + | - | + | NEEDS-TEST |
| heroku_multiple_apps | + | + | + | OK |
| heroku_not_command | + | + | + | OK |
| history | - | n/a | + | NEEDS-RULE, DIVERGENT |
| hostscli | + | + | + | OK |
| ifconfig_device_not_found | + | - | + | NEEDS-TEST, DIVERGENT |
| java | + | + | + | OK |
| javac | + | + | + | OK |
| lein_not_task | + | + | + | OK |
| ln_no_hard_link | + | + | + | OK |
| ln_s_order | + | - | + | NEEDS-TEST |
| long_form_help | + | + | + | OK |
| ls_all | + | + | + | OK |
| ls_lah | + | + | + | OK |
| man | + | + | + | OK |
| man_no_space | + | + | + | OK |
| mercurial | + | + | + | OK |
| missing_space_before_subcommand | + | - | + | NEEDS-TEST |
| mkdir_p | + | + | + | OK |
| mvn_no_command | + | + | + | OK |
| mvn_unknown_lifecycle_phase | + | + | + | OK |
| nixos_cmd_not_found | + | + | + | OK |
| no_command | + | - | + | NEEDS-TEST |
| no_such_file | + | + | + | OK |
| npm_missing_script | + | - | + | NEEDS-TEST, DIVERGENT |
| npm_run_script | + | - | + | NEEDS-TEST, DIVERGENT |
| npm_wrong_command | + | + | + | OK |
| omnienv_no_such_command | + | + | + | OK |
| open | + | + | + | OK |
| pacman | - | n/a | + | NEEDS-RULE, DIVERGENT |
| pacman_invalid_option | + | + | + | OK |
| pacman_not_found | - | n/a | + | NEEDS-RULE, DIVERGENT |
| path_from_history | + | - | + | NEEDS-TEST, DIVERGENT |
| php_s | + | + | + | OK |
| pip_install | + | + | + | OK |
| pip_unknown_command | + | + | + | OK |
| port_already_in_use | + | + | + | OK |
| prove_recursively | + | - | + | NEEDS-TEST |
| python_command | + | + | + | OK |
| python_execute | + | + | + | OK |
| python_module_error | + | + | + | OK |
| quotation_marks | + | + | + | OK |
| rails_migrations_pending | + | + | + | OK |
| react_native_command_unrecognized | + | - | + | NEEDS-TEST, DIVERGENT |
| remove_shell_prompt_literal | + | + | + | OK |
| remove_trailing_cedilla | + | + | + | OK |
| rm_dir | + | + | + | OK |
| rm_root | + | + | + | OK |
| scm_correction | + | - | + | NEEDS-TEST |
| sed_unterminated_s | + | + | + | OK |
| sl_ls | + | + | + | OK |
| ssh_known_hosts | + | + | + | OK |
| sudo | + | + | + | OK |
| sudo_command_from_user_path | + | - | + | NEEDS-TEST |
| switch_lang | + | + | + | OK |
| systemctl | + | + | + | OK |
| terraform_init | + | + | + | OK |
| terraform_no_command | + | + | + | OK |
| test.py | - | n/a | - | NEEDS-RULE |
| tmux | + | + | + | OK |
| touch | + | + | + | OK |
| tsuru_login | + | + | + | OK |
| tsuru_not_command | + | + | + | OK |
| unknown_command | + | + | + | OK |
| unsudo | + | + | + | OK |
| vagrant_up | + | + | + | OK |
| whois | + | + | + | OK |
| workon_doesnt_exists | + | - | + | NEEDS-TEST, DIVERGENT |
| wrong_hyphen_before_subcommand | + | - | + | NEEDS-TEST |
| yarn_alias | + | + | + | OK |
| yarn_command_not_found | + | - | + | NEEDS-TEST, DIVERGENT |
| yarn_command_replaced | + | + | + | OK |
| yarn_help | + | + | + | OK |
| yum_invalid_operation | + | + | + | DIVERGENT |

Note on test naming: upstream has three test files whose name does not match
their rule:
- `test_git_pull_unstaged_changes.py` exercises `git_pull_uncommitted_changes`
  (it's a duplicate fixture file kept under a stale name).
- `test_gradle_not_task.py` exercises `gradle_no_task`.
- `test_ssh_known_host.py` exercises `ssh_known_hosts`.

---

## Rules with no upstream test

These are upstream rules without a `test_*.py` in the upstream repo. We
should still keep parity for the rule itself, but there is no upstream
test to satisfy:

- `cargo` (no upstream test, Go has one)
- `cpp11` (no upstream test, Go has none — `NEEDS-TEST` against our own behaviour)
- `gradle_no_task` (no upstream test, but `test_gradle_not_task` covers it)
- `ssh_known_hosts` (no upstream test, but `test_ssh_known_host` covers it)
- `test.py` (no upstream test)

---

## Known divergences from upstream

These rules diverge because the Go port lacks the subprocess/shell-history
infrastructure that thefuck uses. Either close the gap by building the
infrastructure (preferred) or accept the divergence with a comment in the
rule. Each entry says what upstream does vs. what the port currently does.

| Rule | Upstream | Port |
| --- | --- | --- |
| `apt_invalid_operation` | calls `apt --help` | static op list |
| `dnf_no_such_command` | calls `dnf --help` | static |
| `docker_not_command` | calls `docker --help` | static |
| `fab_command_not_found` | parses fab's `-l` output | parses stderr |
| `gem_unknown_command` | calls `gem help commands` | static |
| `go_unknown_command` | calls `go help` | static |
| `gradle_no_task` | calls `gradle tasks` | static |
| `grunt_task_not_found` | calls `grunt --help` | static |
| `gulp_not_task` | calls gulp | static |
| `history` | shell-history integration | not implemented |
| `ifconfig_device_not_found` | enumerates interfaces | static |
| `npm_missing_script` | calls `npm run-script` | not implemented |
| `npm_run_script` | calls `npm run-script` | not implemented |
| `pacman` | calls `pkgfile` | not implemented |
| `pacman_not_found` | calls `pkgfile` | not implemented |
| `path_from_history` | shell history | simplified fallback |
| `react_native_command_unrecognized` | calls `react-native --help` | static |
| `workon_doesnt_exists` | enumerates `~/.virtualenvs/` | logic diverges |
| `yarn_command_not_found` | calls `yarn help` | static list |
| `yum_invalid_operation` | calls `yum --help` | static |

---

## Infrastructure gaps blocking divergence closure

- Subprocess invocation helpers (with test-time mocking seam) for rules
  that call external tools (`apt --help`, `gem help commands`, etc.).
- Shell-history reader for `no_command`, `history`, `path_from_history`.
- `ifconfig` / interface enumeration helper.
- `pkgfile` integration for the pacman rules.
- Test-helper for tmpdir + chdir + `t.Setenv("PATH", ...)` so the
  filesystem-touching rules can be tested cleanly.

---

## Rule fixes already made while porting tests

(Carried over from the previous TODO — keep for context.)

- `GetCloseMatches` in `internal/utils/utils.go`: arg order now matches
  Python's `difflib.get_close_matches` (seq1=candidate, seq2=word — the
  algorithm is asymmetric), plus tie-break on string-desc to match
  `heapq.nlargest` tuple comparison.
- `brew_reinstall`: removed Perl-only regex (Go's RE2 rejects `(?!…)`); now
  uses plain `strings.Contains`.
- `choco_install`: removed dead-code `if` blocks from the Python-to-Go port.
- `grunt_task_not_found`: replaced no-op `strings.Replace(x, x, x)` with a
  static task list.
- `npm_wrong_command`: parses commands from npm's own help output instead
  of a static list, matching upstream.
- `open`: returns either the URL-prefixed form or the touch/mkdir pair, not
  both at once.
- `tmux`: regex captures the full candidate list (`(.*)` instead of `([^ ]*)`).
- `touch`: regex now matches BSD output (no surrounding quotes).
- `whois`: complete rewrite to match upstream.

---

## Subtask plan (in execution order)

Mark a task `[x]` when complete. Each task should leave the tree
green (`go test ./...`) and committable. Don't bundle tasks together —
they are sized so that progress can be picked up at any time.

### Phase 1 — make it runnable

- [x] **S1.1** Audit upstream vs port and produce the matrix above.
- [x] **S1.2** Add `cmd/gofuck/main.go` — minimal CLI: take previous command
      script + output (flags or stdin), call
      `corrector.GetCorrectedCommands`, print the top candidate. Add a
      `--all` flag to print every candidate. Verify with `go build ./...`
      and a smoke run against a known-good rule.
- [x] **S1.3** Document how to run the binary in `README.md` (one-paragraph
      "usage" section, no marketing).

### Phase 2 — close the implementation gap (6 missing rules)

Each of these is one task. Port the rule from upstream, register it,
write the corresponding Go test against upstream's `test_<rule>.py`.

- [ ] **S2.1** Implement `apt_get` rule (needs `CommandNotFound` lookup —
      pick a static replacement or scaffold a hook for it). Test parity.
- [x] **S2.2** Implement `man` rule. Pure string manipulation; should be
      easy. Test parity with `tests/rules/test_man.py`.
- [ ] **S2.3** Implement `test.py` rule (1-liner; priority 900).
- [ ] **S2.4** Implement `history` rule (depends on shell-history infra —
      see Phase 4). Mark this blocked until S4.2 lands.
- [ ] **S2.5** Implement `pacman` rule (depends on pkgfile infra — Phase 4).
- [ ] **S2.6** Implement `pacman_not_found` rule (depends on pkgfile infra).

### Phase 3 — fill in the missing tests for already-ported rules

These rules already have Go implementations; the goal is just to add
Go tests that mirror upstream `test_<rule>.py`. Many will need a small
`fstest` helper (tmpdir + chdir or `t.Setenv("PATH", ...)`).

- [ ] **S3.0** Add `internal/rules/testfs_test.go` with `withTmpDir`,
      `withPath`, `touchFile` helpers used by the rest of Phase 3.
- [ ] **S3.1** `cat_dir` test — needs isdir.
- [ ] **S3.2** `cd_correction` test — needs real subdir listing.
- [ ] **S3.3** `chmod_x` test — needs file exists + non-exec mode.
- [ ] **S3.4** `cpp11` test — pure string manipulation, no fixture.
- [ ] **S3.5** `dirty_untar` test — needs real tar archive in tmpdir.
- [ ] **S3.6** `dirty_unzip` test — needs real zip archive in tmpdir.
- [ ] **S3.7** `fix_file` test — needs `$EDITOR` and a real file.
- [ ] **S3.8** `git_add` test — needs `Path.exists` mock (use real tmpdir).
- [ ] **S3.9** `git_clone_missing` test — needs controlled `$PATH`.
- [ ] **S3.10** `git_pull` test — output-driven; no fixture.
- [ ] **S3.11** `git_rebase_merge_dir` test — accept ordering divergence
      (or fix ordering).
- [ ] **S3.12** `gradle_wrapper` test — needs `./gradlew` file + `$PATH`.
- [ ] **S3.13** `grep_arguments_order` test — needs real file.
- [ ] **S3.14** `has_exists_script` test — needs real file.
- [ ] **S3.15** `ln_s_order` test — needs real file.
- [ ] **S3.16** `missing_space_before_subcommand` test — controlled `$PATH`.
- [ ] **S3.17** `no_command` test — controlled `$PATH`.
- [ ] **S3.18** `prove_recursively` test — needs isdir.
- [ ] **S3.19** `scm_correction` test — needs `.git`/`.hg` dir.
- [ ] **S3.20** `sudo_command_from_user_path` test — controlled `$PATH`.
- [ ] **S3.21** `wrong_hyphen_before_subcommand` test — controlled `$PATH`.
- [ ] **S3.22** `brew_unknown_command` test — output-driven.
- [ ] **S3.23** Tests for the divergent rules that don't need infra
      (`docker_not_command` / `go_unknown_command` / `gradle_no_task`
      etc. — port the test against the static list and accept skipped
      cases). One subtask per rule, document divergence inline.

### Phase 4 — close divergences by building the missing infra

Each of these unlocks a batch of rules.

- [ ] **S4.1** `internal/specific/exec` — `Run(name, args...) (stdout, stderr,
      err)` with a swappable seam (`var execRunner = exec.Run`) so tests
      can inject canned outputs. Use it from `apt_invalid_operation`,
      `dnf_no_such_command`, `docker_not_command`, `gem_unknown_command`,
      `go_unknown_command`, `gradle_no_task`, `grunt_task_not_found`,
      `gulp_not_task`, `react_native_command_unrecognized`,
      `yarn_command_not_found`, `yum_invalid_operation`.
- [ ] **S4.2** Shell-history reader (`internal/shells/history.go`):
      bash/zsh/fish history file lookup. Unblocks `history`,
      `no_command`, `path_from_history`.
- [ ] **S4.3** Network interface enumerator: net.Interfaces() lookup.
      Unblocks `ifconfig_device_not_found`.
- [ ] **S4.4** `pkgfile` integration: `internal/specific/archlinux.go`.
      Unblocks `pacman`, `pacman_not_found`.
- [ ] **S4.5** Re-test divergent rules now matching upstream behaviour;
      update the table above.

### Phase 5 — exhaustive parity check

- [ ] **S5.1** Write a parity harness (`tools/parity/main.go` — out of
      tree of the binary) that walks `/tmp/thefuck-ref/tests/rules/*.py`,
      extracts cases, runs the Go rule, and reports diffs. Numeric
      target: 167 upstream test files mapped, ≥99% case pass rate.
- [ ] **S5.2** Triage every failing case from S5.1 into a numbered
      follow-up subtask under this section.

---

## Progress log

Append a line each session so anyone picking up the work knows where we
are. Newest at the bottom.

- 2026-04-25 — Re-cloned upstream into `/tmp/thefuck-ref`. Audited
  upstream vs port: 163/169 rules implemented, 129/169 rule-tests
  written, 20 known divergences. New TODO authored with per-rule matrix
  and Phase 1–5 plan.
- 2026-04-25 — S1.2 + S1.3 done. Added `cmd/gofuck/main.go` (flags:
  `--output`, `--stdin`, `--all`, positional script). Smoke-tested
  end-to-end on `mkdir_p`, `git_push_force`, `no_such_file`. README
  gained a "Usage" section. `go test ./...` still green. **Next:
  S2.2 (`man` rule) — pure string manipulation, easy first port.**

---

## Reference

- Upstream tests: `/tmp/thefuck-ref/tests/rules/test_*.py`
- Upstream rules: `/tmp/thefuck-ref/thefuck/rules/*.py`
- Upstream entrypoint (for CLI shape inspiration):
  `/tmp/thefuck-ref/thefuck/entrypoints/fix_command.py`
