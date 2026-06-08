# Design: repo groups become a list (`group` → `groups`)

**Date:** 2026-06-08
**Status:** Approved

## Goal

Let a repository belong to **multiple groups** instead of exactly one. The
config key changes from the singular `group` (string) to the plural `groups`
(list of strings). The `-g/--group` *filter* flag stays singular: it selects all
repos whose group list **contains** the requested group.

## Decisions

- **Clean break, no backward compat.** The app has no users yet, so `group` is
  removed entirely — no legacy field, no migration command, no deprecation
  warning. A config still using `group:` simply has its value ignored by YAML
  (unknown key), and the repo ends up with no groups.
- **Filter stays single-group.** All `-g` flags on read/bulk commands
  (`list`, `pull`, `push`, `status`, `diff`, `clone`, `checkout`, `stash`,
  `outdated`) keep taking one group name and match by membership.
- **Assignment becomes multi-group.** `-g` on `add` and `import` becomes
  **repeatable** (`-g work -g oss`), assigning the full list.

## Data model — `internal/config/config.go`

```go
type Repo struct {
    URL          string       `mapstructure:"url" yaml:"url"`
    Path         string       `mapstructure:"path" yaml:"path,omitempty"`
    Groups       []string     `mapstructure:"groups" yaml:"groups,omitempty"`
    PullStrategy PullStrategy `mapstructure:"pull_strategy" yaml:"pull_strategy,omitempty"`
}
```

The old `Group string` field is deleted. `yaml:"groups,omitempty"` keeps repos
without groups serialized cleanly. `WriteDefault`'s template comment is updated
to show `# groups: [oss]` instead of the old single-group hint.

Config file example after the change:

```yaml
repos:
  - url: git@github.com:user/repo.git
    groups:
      - work
      - oss
```

## Filtering — `internal/cli/root.go`

`filterByGroup(repos []config.Repo, group string)` keeps its signature; only the
match changes from equality to membership:

```go
if slices.Contains(r.Groups, group) {   // was: r.Group == group
```

Empty `group` still returns all repos. No changes to any command that calls it —
they all pass a single group string as before.

## Assignment — `add.go` / `import.go`

- Flag definition: `StringP("group", "g", ...)` → `StringArrayP("group", "g", nil, ...)`.
- Read with `cmd.Flags().GetStringArray("group")` → `[]string`.
- Build repo as `config.Repo{URL: url, Groups: groups, Path: path}`.
- `import` applies the same `groups` slice to every imported repo.
- Help text updated (e.g. "Assign the repo to one or more groups (repeatable)").

## Display & listing — `list.go`, `helpers.go`, `completion.go`

- `repoChoiceLabel` (helpers.go): render `[work, oss]` via
  `strings.Join(r.Groups, ", ")`, only when the list is non-empty.
- `listGroups` (list.go): count by iterating each `r.Groups` entry, so a repo in
  two groups increments both counts.
- `list` JSON entry: field `Group string \`json:"group,omitempty"\`` becomes
  `Groups []string \`json:"groups,omitempty"\``. **Breaking change** to JSON
  consumers — acceptable given no users yet.
- `groupCompletion` (completion.go): iterate `r.Groups` and add each name to the
  completion set.

## Validation — `validate.go` / doctor

- "Blank group" check: iterate `r.Groups`; warn for any entry that is empty or
  whitespace-only.
- (Optional, nice-to-have) warn on duplicate group names within a single repo's
  list.
- No legacy-`group` detection (clean break).

## Tests to update

- `internal/cli/cmd_test.go` — `TestFilterByGroup_*`: change `Group: "work"` to
  `Groups: []string{"work"}`; add a case where a repo has multiple groups and
  the filter matches one of them.
- `internal/config/config_test.go` — replace `Group:` literals (lines ~63, 158,
  196) and the `loaded.Repos[*].Group` assertions (lines ~94, 217) with the
  `Groups` slice equivalents.
- `tests/e2e_test.go` — the inline config at line ~142 (`group: work`) becomes
  `groups:` list form; the `strings.Count(configText, "group: work")` assertion
  (line ~566) is updated to the new sequence serialization (e.g. count `- work`);
  the `add`/`import --group work` invocations still work unchanged (single
  repeated value).

## Out of scope

- No comma-separated parsing for `-g` (repeatable flag only).
- No migration tooling.
- No change to filter semantics (stays single-group membership).
