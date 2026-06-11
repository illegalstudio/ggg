# Owner Aliases — Design

## Problem

Repositories are cloned to a path derived from their git URL. The derived
path is `owner/repo` (the host is **not** part of it):
`git@github.com:nahime0/repo.git` → `nahime0/repo` → `~/Developer/nahime0/repo`.

A user whose GitHub username is `nahime0` wants those repos kept under a
`nahime` folder instead, without writing a custom `path:` for every single
repo.

## Goal

Let the config declare aliases that rewrite the **owner** segment of the
derived path, applied transparently wherever a repo path is derived.

## Decisions

- **Scope of replacement:** owner-only — the first segment (`[0]`) of the
  derived path. The rest of the path is preserved.
  `nahime0/repo` → `nahime/repo`. The derived path has no host segment, so
  there is nothing host-related to preserve.
- **Matching:** owner name, host-independent. The URL's host does not appear in
  the derived path, so `nahime0` is matched regardless of which host the repo
  came from.
- **Explicit `path:` wins:** a repo with an explicit `path:` bypasses alias
  resolution entirely (the existing `FullPath` short-circuit is unchanged).
- **Import stays implicit:** `import` keeps saving only the `url` (no `path:`).
  The alias is applied at derivation time, so imported repos resolve to the
  aliased folder automatically. Changing an alias later re-aligns every repo.
- **Consistency:** alias-applied paths are the canonical derived path used by
  listing, matching (`cd`, bulk commands), and shell completion — not just the
  clone destination. Substring matching on the raw URL still works, so the
  original owner (`nahime0`) remains a valid query.

## Config

New optional top-level field `aliases` (`map[string]string`), owner → folder.
Absent or empty map = current behavior, unchanged.

```yaml
base_dir: ~/Developer
aliases:
  nahime0: nahime
repos:
  - url: git@github.com:nahime0/ggg.git   # -> ~/Developer/nahime/ggg
```

Added to `config.Config` as:

```go
Aliases map[string]string `mapstructure:"aliases" yaml:"aliases,omitempty"`
```

## Core logic (package `repo`)

A pure, independently testable helper:

```go
// ApplyOwnerAlias replaces the first segment (the owner) of a derived relative
// path when it matches an alias key. Returns relPath unchanged when there is no
// match, the map is empty/nil, or the path has no owner segment.
func ApplyOwnerAlias(relPath string, aliases map[string]string) string
```

Examples:
- `ApplyOwnerAlias("nahime0/repo", {"nahime0":"nahime"})` → `nahime/repo`
- `ApplyOwnerAlias("grp/sub/repo", {"grp":"work"})` → `work/sub/repo`
- `ApplyOwnerAlias("other/repo", {"nahime0":"nahime"})` → unchanged
- `ApplyOwnerAlias("nahime0/repo", nil)` → unchanged

## Wiring (approach A — thread the alias map through `repo`)

All path semantics stay in the `repo` package; the alias map becomes an
explicit dependency.

- `FullPath(baseDir string, r config.Repo)` →
  `FullPath(baseDir string, aliases map[string]string, r config.Repo)`.
  When `r.Path == ""`, derive from URL then `ApplyOwnerAlias`. Explicit
  `r.Path` still returns early, before any alias logic.
- New `DerivedPath(r config.Repo, aliases map[string]string) (string, error)`
  = derive + `ApplyOwnerAlias`, for code that needs the relative derived path
  (matching, completion).
- Update the ~14 `repo.FullPath(cfg.BaseDir, r)` call sites in `internal/cli/`
  to pass `cfg.Aliases`. Each already holds the relevant `cfg`.
- Update `internal/cli/helpers.go` (`filterByName`, `matchRepoIndices`) and
  `internal/cli/completion.go` to use `DerivedPath(r, cfg.Aliases)` instead of
  `repo.DerivePathFromURL(r.URL)`.

## Free effects

- **`import`:** no code change. It saves only the URL; aliases apply at
  derivation time.
- **`validate`:** already detects duplicate `FullPath`s. Because it goes
  through the alias-aware `FullPath`, two owners mapping to the same folder and
  colliding are reported automatically.

## Testing

- **Unit (`repo`):** `ApplyOwnerAlias` — match, no-match, nil/empty map,
  different hosts, path lacking an owner segment.
- **Unit (`repo`):** `FullPath` and `DerivedPath` with and without aliases;
  explicit `path:` bypasses aliasing.
- **E2E:** config with `aliases`; assert clone destination uses the aliased
  folder and that `cd`/matching resolves the imported repo.

## Docs

- `docs/configuration.md`: new `aliases` field row + worked example.
- `README.md`: brief mention.

## Out of scope

- Host-scoped aliases (`github.com/nahime0`) — explicitly rejected; matching is
  owner-name only.
- Materializing an explicit `path:` into config on `import`.
- Aliasing anything other than the owner segment.
