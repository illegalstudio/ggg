# Shell Integration

## Quick Setup

Use `ggg shell-init` to generate a `gcd` shell function for quick navigation and shell completion for `ggg`:

### Bash

Add to `~/.bashrc`:

```bash
eval "$(ggg shell-init bash)"
```

### Zsh

Add to `~/.zshrc`:

```bash
eval "$(ggg shell-init zsh)"
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
ggg shell-init fish | source
```

Then reload your shell and use `gcd`:

```bash
gcd myrepo
```

## How It Works

A subprocess cannot change the parent shell's directory. The `ggg shell-init` command outputs a `gcd` shell function that calls `ggg cd <name>` under the hood, captures the path, and runs `cd` in the current shell.

`ggg shell-init` also emits Cobra-powered completion code for the selected shell.

## Completion

Completion includes:

- command and flag names, for example `ggg sta<TAB>` completing to `status`
- configured repository paths and basenames, for example `project` or `team/project`
- configured group names for `--group/-g`
- local branch names for `ggg checkout`

The generated completion supports bash, zsh, and fish through the same `shell-init` command used for `gcd`.

The original `ggg cd <name>` command still works as before — it prints the repository path to stdout, so you can also use it directly with `cd`:

```bash
cd "$(ggg cd myrepo)"
```

## Usage

```bash
# Navigate to a repo by name (partial match works)
gcd myrepo

# Full URL also works
gcd git@github.com:user/repo.git

# Or use ggg cd directly
cd "$(ggg cd myrepo)"
```

If multiple repos match the name, GGG presents an interactive selector to choose one.
