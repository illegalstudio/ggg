# Groups

Groups let you organize repositories into logical categories and operate on subsets of your repos.

## Defining Groups

Add a `groups` list to any repo in `~/.config/ggg/repositories.yaml`. A repo can belong to several groups:

```yaml
base_dir: ~/Developer

repos:
  - url: git@github.com:company/backend.git
    groups: [work]

  - url: git@github.com:company/frontend.git
    groups: [work, web]

  - url: git@github.com:user/dotfiles.git
    groups: [personal]

  - url: git@github.com:user/side-project.git
    groups: [personal, web]

  # Repos without groups are included in ungrouped operations
  - url: git@github.com:user/notes.git
```

The block form works too:

```yaml
  - url: git@github.com:company/frontend.git
    groups:
      - work
      - web
```

Groups are free-form strings, so use whatever names make sense for your workflow.

The field is plural. A singular `group:` key is not read and the repo ends up in no group; `ggg list --groups` shows which groups ggg actually sees.

## Assigning Groups from the CLI

`ggg add` and `ggg import` accept a repeatable `--group/-g`:

```bash
ggg add git@github.com:company/frontend.git -g work -g web
ggg import myorg -g work
```

## Filtering with `--group` / `-g`

The following commands filter by group:

| Command | Example |
|---------|---------|
| `list` | `ggg list -g work` |
| `clone` | `ggg clone -g work` |
| `pull` | `ggg pull -g work` |
| `push` | `ggg push -g work` |
| `status` | `ggg status -g work` |
| `outdated` | `ggg outdated -g work` |
| `stash` | `ggg stash -g work` |
| `checkout` | `ggg checkout main -g work` |
| `diff` | `ggg diff -g work` |

When `--group` is specified, only repos whose `groups` list contains that group are included. Without it, all repos are included regardless of group. The filter accepts a single group; passing `--group` twice is an error.

## Examples

```bash
# List the available groups with their repo counts
ggg list --groups

# List only work repos
ggg list --group work

# Clone all personal repos
ggg clone -g personal

# Pull latest for web repos (frontend and side-project above)
ggg pull -g web

# Find outdated work repos
ggg outdated --group work
```

## Tips

- A repo can belong to **several groups**, or to none.
- The group match is **exact**: `ggg list -g Work` will not match `work`.
- Repos without groups are only included when no `--group` filter is applied.
- `ggg validate` warns about blank group names.
