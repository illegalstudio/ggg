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

A second, optional file holds settings that belong to one machine only:

```
~/.config/ggg/repositories.local.yaml
```

See [Machine-local overrides](#machine-local-overrides).

## Fields

### Top-Level

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `base_dir` | `string` | No | `~/Developer` | Root directory where repositories are cloned. Supports `~`, `$VAR` and `${VAR}` expansion, and can be overridden per machine (see [Machine-local overrides](#machine-local-overrides)). |
| `pull_strategy` | `string` | No | `merge` | Default pull strategy for all repos. |
| `aliases` | `map` | No | - | Map of repo owner → folder name. Repos owned by the key are cloned under the value (on any host) unless they set an explicit `path`. |
| `suppress_skills_notice` | `bool` | No | `false` | Set to `true` to silence the stderr notice ggg prints when an installed AI agent skill no longer matches the bundled one. |
| `repos` | `list` | Yes | - | List of repository entries. |

### Repository Entry (`repos[]`)

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | `string` | Yes | - | Git repository URL (SSH or HTTPS). |
| `path` | `string` | No | Derived from URL | Custom clone path relative to `base_dir`. If omitted, derived from the URL (e.g., `user/repo`). |
| `groups` | `list` | No | - | Group names the repo belongs to, e.g. `[work, oss]`. A repo can be in several groups. Used to filter commands with `--group`. See [Groups](groups.md). |
| `pull_strategy` | `string` | No | - | Pull strategy override for this repo. |

## Machine-local overrides

`repositories.yaml` is meant to be shared: keep it in a dotfiles repository or
sync it between machines. Repo paths are relative to `base_dir`, so the only
thing that usually differs between machines is `base_dir` itself (for example
`~/Developer` on a laptop and `/data/dev` on a workstation). Three mechanisms
handle that without editing the shared file.

### Local file

`~/.config/ggg/repositories.local.yaml` is read on top of the shared file.
Create it with:

```bash
ggg init --local --base-dir /data/dev   # or omit --base-dir to be asked
ggg config --local                      # edit it later
```

```yaml
# repositories.local.yaml
base_dir: /data/dev

# Optional:
# pull_strategy: rebase      # replaces the global pull_strategy
# aliases:                   # merged by key with the shared aliases, local wins
#   nahime0: nahime
# suppress_skills_notice: true
```

Rules:

- Only `base_dir`, `pull_strategy`, `aliases`, and `suppress_skills_notice` are
  accepted. A `repos` list or an unknown key is an error, so the repository list
  stays in one place and typos do not go unnoticed.
- An invalid `pull_strategy` is an error.
- Commands that write the configuration (`add`, `remove`, `import`, `validate`)
  only rewrite `repositories.yaml`; local values never leak into it.
- `ggg export` copies only the shared file.
- Keep the local file out of whatever syncs `repositories.yaml`. If you sync the
  whole `~/.config/ggg/` directory, exclude `repositories.local.yaml`.

### `GGG_BASE_DIR`

The `GGG_BASE_DIR` environment variable overrides `base_dir` from both files.
It is handy for CI, containers, or a one-off run:

```bash
GGG_BASE_DIR=/mnt/scratch ggg clone -g work
```

An empty value counts as unset.

### Variables in `base_dir`

`base_dir` expands `~`, `$VAR`, and `${VAR}`, in any of the sources above, so a
shared file can delegate the location to the environment:

```yaml
base_dir: ${DEV_ROOT}/src
```

A referenced variable that is unset or empty is an error rather than an empty
string, so a missing variable never sends clones to an unexpected directory.
Quote a bare `~` (`base_dir: "~"`): unquoted, YAML reads it as null and the
default applies.

### Precedence

The effective `base_dir` is resolved in this order:

1. `GGG_BASE_DIR` (if set and non-empty)
2. `base_dir` in `repositories.local.yaml`
3. `base_dir` in `repositories.yaml`
4. Default: `~/Developer`

`ggg doctor` shows the effective base directory and where it came from.

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
2. **Global** `pull_strategy` (if set; the local file's value replaces the shared one)
3. **Default**: `merge`

## Stale-skill notice

After any interactive command, ggg compares the AI agent skills installed under
`~/.agents/skills/ggg` and `~/.claude/skills/ggg` with the one bundled in the
running binary. When an installed copy exists and its digest differs, a short
stderr notice explains how to update (`ggg skills install`), how to check on
demand (`ggg skills verify`), and how to silence the reminder:

```yaml
suppress_skills_notice: true
```

The notice never appears in `--json` output, so scripts are unaffected. Set the
flag if you deliberately maintain your own edits of the skill.

## Owner Aliases

`aliases` rewrites the owner segment of a repo's derived path. The derived path
is `owner/repo` (no host), so an alias changes the top-level folder a repo is
cloned into.

```yaml
aliases:
  nahime0: nahime

repos:
  # cloned to <base_dir>/nahime/ggg instead of <base_dir>/nahime0/ggg
  - url: git@github.com:nahime0/ggg.git
```

Notes:

- Matching is by owner name and is host-independent (`github.com`, `gitlab.com`, …).
- A repo with an explicit `path` is never affected by aliases.
- Aliases apply lazily at derivation time, so `import`ed repos (which store only
  the URL) automatically land in the aliased folder, and changing an alias
  re-aligns every repo that uses it.

## Full Example

```yaml
# Base directory for all cloned repositories
base_dir: ~/Developer

# Default pull strategy for all repos: merge, rebase, ff-only
pull_strategy: rebase

# Rewrite owner segments in derived paths (host-independent)
aliases:
  nahime0: nahime

# Silence the stderr notice when an installed AI agent skill no longer
# matches the one bundled with the running binary
# suppress_skills_notice: true

repos:
  # Minimal — URL only, path derived as user/app
  - url: git@github.com:user/app.git

  # Custom path — cloned to ~/Developer/my-tools/app
  - url: https://github.com/org/project.git
    path: my-tools/app

  # With a group, filterable with --group work
  - url: git@github.com:company/backend.git
    groups: [work]

  # In several groups, and a per-repo pull strategy that overrides global rebase
  - url: git@github.com:company/frontend.git
    groups: [work, web]
    pull_strategy: ff-only

  # Personal project in a separate group
  - url: git@github.com:user/dotfiles.git
    groups: [personal]
    pull_strategy: merge
```
