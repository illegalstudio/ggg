# Shell Integration

## The Problem

`ggg cd` prints a `cd` command to stdout, but a subprocess cannot change the parent shell's directory. You need `eval` to execute the output in the current shell:

```bash
eval $(ggg cd myrepo)
```

## Setting Up a Shell Function

For a seamless experience, define a wrapper function so you can simply type `gcd myrepo`.

### Bash / Zsh

Add to `~/.bashrc` or `~/.zshrc`:

```bash
gcd() {
  local output
  output=$(ggg cd "$@" 2>&1)
  if [ $? -eq 0 ]; then
    eval "$output"
  else
    echo "$output" >&2
    return 1
  fi
}
```

Reload your shell:

```bash
source ~/.bashrc   # or source ~/.zshrc
```

### Fish

Add to `~/.config/fish/functions/gcd.fish`:

```fish
function gcd
    set -l output (ggg cd $argv 2>&1)
    if test $status -eq 0
        eval $output
    else
        echo $output >&2
        return 1
    end
end
```

## Usage

After setting up the function:

```bash
# Navigate to a repo by name (partial match works)
gcd myrepo

# Full URL also works
gcd git@github.com:user/repo.git
```

If multiple repos match the name, GGG presents an interactive selector to choose one.
