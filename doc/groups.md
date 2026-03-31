# Groups

Groups let you organize repositories into logical categories and operate on subsets of your repos.

## Defining Groups

Add a `group` field to any repo in `~/.config/ggg/repositories.yaml`:

```yaml
base_dir: ~/Developer

repos:
  - url: git@github.com:company/backend.git
    group: work

  - url: git@github.com:company/frontend.git
    group: work

  - url: git@github.com:user/dotfiles.git
    group: personal

  - url: git@github.com:user/side-project.git
    group: personal

  # Repos without a group are included in ungrouped operations
  - url: git@github.com:user/notes.git
```

Groups are free-form strings — use whatever names make sense for your workflow.

## Using the `--group` / `-g` Flag

The following commands support `--group`:

| Command | Example |
|---------|---------|
| `list` | `ggg list -g work` |
| `clone` | `ggg clone -g work` |
| `pull` | `ggg pull -g work` |
| `status` | `ggg status -g work` |
| `outdated` | `ggg outdated -g work` |

When `--group` is specified, only repos matching that group are included. Without it, all repos are included regardless of group.

## Examples

```bash
# List only work repos
ggg list --group work

# Clone all personal repos
ggg clone -g personal

# Pull latest for work repos
ggg pull -g work

# Check status of personal repos
ggg status -g personal

# Find outdated work repos
ggg outdated --group work
```

## Tips

- A repo can only belong to **one group** (or no group).
- The group match is **exact** — `ggg list -g Work` will not match `work`.
- Repos without a `group` field are only included when no `--group` filter is applied.
