# Commands Reference

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
```

| Flag | Short | Description |
|------|-------|-------------|
| `--group` | `-g` | Show only repos in this group. |

For each cloned repo, displays:
- **Branch** name
- **Clean/dirty** working tree status
- **Ahead/behind** remote (`↑N` / `↓N`)

Status checks run **in parallel** with a spinner.

---

## `ggg add`

Add a repository to the configuration.

```bash
ggg add git@github.com:user/new-repo.git
```

Appends the repo to `~/.config/ggg/repositories.yaml`. Fails if the URL is already configured. The repo is not cloned automatically — run `ggg clone` afterwards.

---

## `ggg remove`

Remove a repository from the configuration.

```bash
ggg remove myrepo
ggg remove git@github.com:user/old-repo.git
```

Supports the same matching behavior as `clone`:
1. Exact match against URL, path, or derived path.
2. Partial, case-insensitive substring match.
3. Interactive selector on multiple matches.

This only removes the entry from the config file — it does **not** delete the cloned directory.

---

## `ggg cd`

Print a `cd` command for a repository. Designed to be used with `eval` for shell integration.

```bash
eval $(ggg cd myrepo)
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
