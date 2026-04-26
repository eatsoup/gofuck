# gofuck

A Go port of [**thefuck**](https://github.com/nvbn/thefuck), the magnificent app that corrects errors in previous console commands.

## Credits

**All credit for the design, rules, behaviour, and test suite goes to [Vladimir Iakovlev (@nvbn)](https://github.com/nvbn) and the thefuck contributors.**

- Original project: https://github.com/nvbn/thefuck
- Original author: Vladimir Iakovlev ([@nvbn](https://github.com/nvbn))
- Inspired by [@liamosaur's tweet](https://twitter.com/liamosaur/status/506975850596536320)
- Original license: MIT (Copyright (c) 2015-2022 Vladimir Iakovlev)

This repository is an unofficial, derivative Go port. It faithfully mirrors thefuck's rules
and is tested against ports of thefuck's own test cases. Every rule here is a direct
translation of the corresponding rule in `thefuck/rules/*.py`.

If you like what this does, star [nvbn/thefuck](https://github.com/nvbn/thefuck) —
that's the project that made it all possible.

## License

Distributed under the MIT License, matching the upstream project. See
[thefuck's LICENSE.md](https://github.com/nvbn/thefuck/blob/master/LICENSE.md) for the
full original license text.

## Status

This is a work-in-progress port. See [TODO.md](TODO.md) for the current state of test
coverage and documented divergences from upstream behaviour.

## Usage

```sh
go build ./cmd/gofuck

# pass the previous command and its output:
./gofuck --output "mkdir: cannot create directory 'a/b/c': No such file or directory" -- mkdir a/b/c
# → mkdir -p a/b/c

# or pipe the output via stdin:
mkdir a/b/c 2>&1 | ./gofuck --stdin -- mkdir a/b/c

# print every candidate, not just the top one:
./gofuck --all --output "..." -- <command>
```

### Shell integration

Install the alias function in your rc file. The function exports `TF_SHELL`,
`TF_ALIAS`, `TF_SHELL_ALIASES` and `TF_HISTORY` before invoking `gofuck`,
then `eval`s the result and pushes the corrected command back to history.

```sh
# bash (~/.bashrc) and zsh (~/.zshrc):
eval "$(gofuck --alias)"

# fish (~/.config/fish/config.fish):
gofuck --alias | source
```

`--alias` takes an optional name (defaults to `fuck`), so e.g.
`gofuck --alias fix` defines a `fix` function instead. Use `--shell NAME`
to override auto-detection.
