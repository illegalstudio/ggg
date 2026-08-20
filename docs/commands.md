# Commands Reference

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Emit machine-readable JSON instead of styled text for supported commands. Suppresses spinners and bulk-confirmation prompts (auto-confirms), and refuses interactive disambiguation when a query matches multiple repositories. |

The exact JSON shape depends on the command but follows consistent conventions:

- Bulk operations (`clone`, `pull`, `push`, `stash`, `checkout`, `diff`, `status`, `outdated`, `list`) emit a `{"results": [...]}` or `{"repos": [...]}` array.
- Single-target operations (`add`, `remove`, `cd`, `init`, `export`, `shell-init`) emit a small object describing the action and its target.
- `ggg skills install` emits `{"name": ..., "installations": [...]}` and never prompts in JSON mode; use `--target` to narrow the destinations.
- Per-item errors are reported as a non-empty `"error"` string field on the item; the command's overall exit code is unaffected.
- Commands that launch external applications (`config`, `open`, `browse`) do not support `--json`; they return a JSON error object with a non-zero exit code.
- `ggg import --json` requires both an account and one repository argument, for example `ggg --json import myorg myrepo`.

## `ggg skills install`

Install the AI agent skill bundled with the `ggg` binary.

```bash
# Interactive: multi-select menu, both destinations preselected
ggg skills install

# Non-interactive: one destination
ggg skills install --target claude

# Non-interactive: every destination, machine-readable result
ggg --json skills install
```

| Flag | Description |
|------|-------------|
| `--target` | Install only to this destination (`agents`, `claude`). Repeatable. Skips the menu. |
| `--force` | Replace an existing skill that differs from the bundled version. |

Destinations:

- `~/.agents/skills/ggg` (`agents`) — Codex and other Agent Skills hosts
- `~/.claude/skills/ggg` (`claude`) — Claude Code

The two destinations are installed independently: a conflict in one is reported on that destination and does not stop the other.

### Behavior

Reinstalling is safe. GGG records a SHA-256 digest of what it installed in a `.ggg-managed.json` marker inside the destination, and compares it against both the bundled skill and the files on disk:

| Situation | Status | Needs `--force` |
|---|---|---|
| Destination does not exist | `installed` | no |
| Files match the bundled skill | `up-to-date` | no |
| GGG installed it and you have not edited it | `updated` | no |
| You edited the files, or the directory was not created by GGG | `replaced` | **yes** |

Installation is atomic: contents are staged in a sibling temporary directory and moved into place, and the previous copy is kept until the move succeeds.

### JSON output

`ggg --json skills install` never prompts. Without `--target` it installs every destination. It emits:

```json
{
  "name": "ggg",
  "installations": [
    { "target": "agents", "path": "/Users/me/.agents/skills/ggg", "status": "installed" },
    { "target": "claude", "path": "/Users/me/.claude/skills/ggg", "error": "skill already exists at ... rerun with --force to replace it" }
  ]
}
```

Per-destination failures appear in `error` and do not change the exit code, matching the other bulk commands. An unknown `--target` is a command-level error and does exit non-zero.

---

## `ggg init`

Generate a default configuration file.

```bash
ggg init
```

Creates `~/.config/ggg/repositories.yaml` with a starter template. Fails if the file already exists.

---

## `ggg list`

List configured repositories and their clone status.

```bash
ggg list
ggg list --group work
ggg list --groups
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | List only repos in this group. |
| `--groups` | | Show available groups with repo count. |
| `--filter` | `-f` | Filter repos by name. |

Output shows each repo with:
- `✓` — cloned, with the full path
- `○` — not yet cloned, with the expected path

---

## `ggg clone`

Clone repositories — all at once or a specific one.

```bash
# Clone all configured repos (prompts for confirmation)
ggg clone

# Clone a specific repo by name, URL, or path
ggg clone myrepo

# Clone only repos in a group
ggg clone --group work
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Clone only repos in this group. |

### Matching Behavior

When a `[name]` argument is provided:

1. **Exact match** — checks against `url`, `path`, and the derived path.
2. **Partial match** — case-insensitive substring match against the same fields.
3. **Multiple matches** — presents an interactive selector to choose one.

Already-cloned repos are skipped. Cloning runs in parallel with a spinner.

---

## `ggg pull`

Pull latest changes for repos that have a clean working tree.

```bash
# Pull all repos
ggg pull

# Pull a specific repo
ggg pull myrepo

# Pull only repos in a group
ggg pull --group work
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Pull only repos in this group. |

### Behavior

- **Dirty repos are skipped** — repos with uncommitted changes are not pulled.
- **Not-cloned repos are skipped** with a notice.
- **Pull strategy** is respected per repo (see [Configuration](configuration.md#pull-strategy)).
- Pulls run **in parallel** with a spinner.

---

## `ggg status`

Show the status of all configured repositories.

```bash
ggg status
ggg status --group work
ggg status --detailed
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Show only repos in this group. |
| `--filter` | `-f` | Filter repos by name. |
| `--detailed` | `-d` | Show linked worktrees as a tree under each repo. |

For each cloned repo, displays:
- **Branch** name
- **Clean/dirty** working tree status
- **Ahead/behind** remote (`↑N` / `↓N`)
- **Worktree count** (`⎇N`) when one or more linked worktrees exist

With `--detailed`, every linked worktree is rendered as a tree branch under its
repo, showing the worktree directory name, current branch, dirty/clean state,
and ahead/behind counters.

Status checks run **in parallel** with a spinner.

---

## `ggg add`

Add a repository to the configuration.

```bash
ggg add git@github.com:user/new-repo.git
ggg add git@github.com:user/new-repo.git --clone
ggg add git@github.com:user/new-repo.git --group work --path custom/path
```

| Flag | Short | Description |
|------|-------|-------------|
| `--clone` | `-c` | Clone the repo immediately after adding. |
| `--group` | `-g` | Assign the repo to a group. |
| `--path` | `-p` | Custom clone path (relative to `base_dir`). |

Appends the repo to `~/.config/ggg/repositories.yaml`. Fails if the URL is already configured.

---

## `ggg remove`

Remove a repository from the configuration.

```bash
ggg remove
ggg remove myrepo
ggg remove git@github.com:user/old-repo.git
```

Supports the same matching behavior as `clone`:
1. Exact match against URL, path, or derived path.
2. Partial, case-insensitive substring match.
3. Interactive selector on multiple matches, with live filtering as you type.

If no argument is provided, GGG shows a searchable selector with all configured repositories and removes the one you choose.

This only removes the entry from the config file — it does **not** delete the cloned directory.

---

## `ggg cd`

Print the path for a repository. Designed to be used by shell integration.

```bash
cd "$(ggg cd myrepo)"
```

Fails if the repo is not cloned. Supports the same matching behavior as `clone` and `remove`.

See [Shell Integration](shell-integration.md) for how to set up a helper function.

---

## `ggg doctor`

Run health checks on your configuration and repositories.

```bash
ggg doctor
```

Checks performed:

| Check | Description |
|-------|-------------|
| Config file | Exists at expected path |
| Config syntax | File is valid YAML and parseable |
| Base directory | `base_dir` exists on disk |
| Repositories | Count of configured repos |
| Duplicates | Detects duplicate URLs |
| Cloned | Count of cloned vs. missing repos |
| Remotes | Reachability of each remote URL (parallel) |

Unreachable remotes are listed individually at the end.

---

## `ggg outdated`

Fetch from remotes and show repos that are behind.

```bash
ggg outdated
ggg outdated --group work
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Check only repos in this group. |

Fetches from all remotes **in parallel** with a spinner, then reports repos that have commits behind their remote tracking branch (e.g., `5 commits behind`).

If all repos are up to date, prints a success message.

---

## `ggg push`

Push commits to remote for repositories that are ahead.

```bash
# Push all repos with unpushed commits (prompts for confirmation)
ggg push

# Push a specific repo
ggg push myrepo

# Push only repos in a group
ggg push --group work
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Push only repos in this group. |

Only repos with commits ahead of their remote are pushed. Not-cloned repos and repos with no upstream are skipped. Pushes run **in parallel** with a spinner.

---

## `ggg open`

Open a repository in an editor.

```bash
ggg open myrepo              # open in default editor ($EDITOR)
ggg open myrepo code         # open in VS Code
ggg open myrepo cursor       # open in Cursor
```

Uses `$EDITOR`, `$VISUAL`, or `vi` as fallback. Fails if the repo is not cloned.

---

## `ggg browse`

Open a repository's remote URL in the browser.

```bash
ggg browse myrepo
```

Converts SSH URLs to HTTPS automatically. Opens the default browser on macOS, Linux, and Windows.

---

## `ggg import`

Import repositories from GitHub via the `gh` CLI.

```bash
# Interactive: select account, then pick repos
ggg import

# Import from a specific org or user
ggg import myorg

# Import a single repo from a specific org or user
ggg import myorg myrepo

# Use HTTPS URLs and assign to a group
ggg import myorg --http --group work
```

| Flag | Short | Description |
|------|-------|-------------|
| `--http` | | Use HTTPS URLs instead of SSH. |
| `--group` | `-g` | Assign imported repos to a group. |

Requires [`gh`](https://cli.github.com) to be installed. Without a repository argument, presents an interactive multi-select to choose which repos to import. Already-configured repos are skipped.
In the repository picker, press `/` to filter, `space` to toggle a repo, and `enter` to confirm.

When a repository argument is provided, it must match a repo name (`myrepo`), full name (`myorg/myrepo`), SSH URL, or HTTPS URL exactly. In `--json` mode this repository argument is required; JSON import never imports all repositories implicitly.

---

## `ggg stash`

Stash changes in dirty repositories.

```bash
# Stash all dirty repos (prompts for confirmation)
ggg stash

# Stash a specific repo
ggg stash myrepo

# Stash only repos in a group
ggg stash --group work
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Stash only repos in this group. |

Clean repos are skipped. Runs in parallel with a spinner.

---

## `ggg checkout`

Checkout a branch in repositories that have it.

```bash
# Checkout in all repos (prompts for confirmation)
ggg checkout main

# Checkout in a specific repo
ggg checkout develop myrepo

# Checkout only in repos of a group
ggg checkout main --group work
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Checkout only in repos in this group. |

Repos that don't have the specified branch are skipped. Runs in parallel with a spinner.

---

## `ggg diff`

Show a summary of changed files in dirty repositories.

```bash
# Show diff for all dirty repos (prompts for confirmation)
ggg diff

# Show diff for a specific repo
ggg diff myrepo

# Show diff only for repos in a group
ggg diff --group work
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Show diff only for repos in this group. |

Clean repos are skipped. Displays `git diff --stat` output for each dirty repo.

---

## `ggg config`

Open the configuration file in your default editor.

```bash
ggg config
```

Uses `$EDITOR`, `$VISUAL`, or `vi` as fallback.

---

## `ggg shell-init`

Print shell integration script for the `gcd` alias and command completion.

```bash
eval "$(ggg shell-init bash)"    # add to ~/.bashrc
eval "$(ggg shell-init zsh)"     # add to ~/.zshrc
ggg shell-init fish | source     # add to ~/.config/fish/config.fish
```

Generates a `gcd` function for quick directory navigation and installs Cobra-powered completion for commands, flags, repository names, group names, and `checkout` branch names. See [Shell Integration](shell-integration.md) for details.

---

## `ggg export`

Export the configuration file to a given path.

```bash
# Export to a specific file
ggg export ./my-config.yaml

# Export to a directory (saves as repositories.yaml inside it)
ggg export ~/Desktop

# Interactive prompt for the destination path
ggg export
```

Copies `~/.config/ggg/repositories.yaml` to the specified destination. If no path is provided, an interactive prompt asks for one (defaults to `repositories.yaml` in the current directory).

---

## `ggg validate`

Validate the configuration for errors and conflicts.

```bash
ggg validate
```

Checks performed:

| Check | Description |
|-------|-------------|
| Duplicate URLs | Detects repos with the same URL |
| Path conflicts | Detects repos that resolve to the same filesystem path |
| Pull strategy | Validates global and per-repo `pull_strategy` values |
| Blank groups | Detects repos with whitespace-only group names |

If no issues are found, prints a success message.

If conflicts are found (duplicate URLs or path conflicts), offers an interactive cleanup wizard that asks which entry to keep for each conflict and removes the duplicates from the config.
