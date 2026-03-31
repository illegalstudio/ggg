# Configuration Reference

## Config File Location

GGG reads its configuration from:

```
~/.config/ggg/repositories.yaml
```

Generate a default config with:

```bash
ggg init
```

## Fields

### Top-Level

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `base_dir` | `string` | No | `~/Developer` | Root directory where repositories are cloned. Supports `~` expansion. |
| `pull_strategy` | `string` | No | `merge` | Default pull strategy for all repos. |
| `repos` | `list` | Yes | — | List of repository entries. |

### Repository Entry (`repos[]`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | `string` | Yes | — | Git repository URL (SSH or HTTPS). |
| `path` | `string` | No | Derived from URL | Custom clone path relative to `base_dir`. If omitted, derived from the URL (e.g., `github.com/user/repo`). |
| `group` | `string` | No | — | Logical group name. Used to filter commands with `--group`. |
| `pull_strategy` | `string` | No | — | Pull strategy override for this repo. |

## Pull Strategy

The `pull_strategy` field controls how `ggg pull` fetches changes. Valid values:

| Value | Git Equivalent | Description |
|-------|---------------|-------------|
| `merge` | `git pull` | Merge remote changes (default). |
| `rebase` | `git pull --rebase` | Rebase local commits on top of remote. |
| `ff-only` | `git pull --ff-only` | Only fast-forward; fail if not possible. |

### Resolution Order

The effective pull strategy for a repo is resolved in this order:

1. **Repo-level** `pull_strategy` (if set)
2. **Global** `pull_strategy` (if set)
3. **Default**: `merge`

## Full Example

```yaml
# Base directory for all cloned repositories
base_dir: ~/Developer

# Default pull strategy for all repos: merge, rebase, ff-only
pull_strategy: rebase

repos:
  # Minimal — URL only, path derived as github.com/user/app
  - url: git@github.com:user/app.git

  # Custom path — cloned to ~/Developer/my-tools/app
  - url: https://github.com/org/project.git
    path: my-tools/app

  # With group — filterable with --group work
  - url: git@github.com:company/backend.git
    group: work

  # Per-repo pull strategy — overrides global rebase
  - url: git@github.com:company/frontend.git
    group: work
    pull_strategy: ff-only

  # Personal project in a separate group
  - url: git@github.com:user/dotfiles.git
    group: personal
    pull_strategy: merge
```
