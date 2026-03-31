# Shell Integration

## Quick Setup

Use `ggg shell-init` to generate a shell wrapper that makes `ggg cd` work seamlessly:

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

Then reload your shell and use `ggg cd` directly:

```bash
ggg cd myrepo
```

## How It Works

A subprocess cannot change the parent shell's directory. The `ggg shell-init` command outputs a shell function that wraps the `ggg` binary. When you run `ggg cd <name>`, the wrapper intercepts the `cd` subcommand, captures the path from the binary, and runs `cd` in the current shell. All other commands are passed through to the binary unchanged.

## Usage

```bash
# Navigate to a repo by name (partial match works)
ggg cd myrepo

# Full URL also works
ggg cd git@github.com:user/repo.git
```

If multiple repos match the name, GGG presents an interactive selector to choose one.

## Manual Setup

If you prefer not to use `ggg shell-init`, you can use `eval` directly:

```bash
eval $(ggg cd myrepo)
```
