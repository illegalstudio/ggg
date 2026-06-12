# Owner Aliases Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let the config declare owner aliases that rewrite the first (owner) segment of a repo's URL-derived path, applied everywhere a repo path is derived, matched, or completed.

**Architecture:** A new optional `aliases map[string]string` field on `config.Config`. A pure helper `repo.ApplyOwnerAlias` rewrites the owner segment; `repo.DerivedPath` composes derive + alias; `repo.FullPath` gains an `aliases` parameter. CLI call sites pass `cfg.Aliases`. Explicit `path:` and `import` behavior are unchanged (alias applies lazily at derivation time).

**Tech Stack:** Go, cobra, viper, yaml.v3. Tests with the standard `testing` package.

**Key fact:** `DerivePathFromURL` returns `owner/repo` — there is **no host** segment. The owner is segment `[0]`. So `~/Developer/nahime0/repo` → with alias `nahime0: nahime` → `~/Developer/nahime/repo`.

---

## File Structure

- `internal/config/config.go` — add `Aliases` field; add commented example to `WriteDefault`.
- `internal/config/config_test.go` — test that `aliases` unmarshals.
- `internal/repo/repo.go` — add `ApplyOwnerAlias`, `DerivedPath`; change `FullPath` signature.
- `internal/repo/repo_test.go` — tests for the above; update existing `FullPath` tests.
- `internal/cli/*.go` — pass `cfg.Aliases` to `FullPath`; thread `aliases` through matching/completion helpers.
- `tests/e2e_test.go` — end-to-end test of alias-applied clone path + `cd` matching.
- `docs/configuration.md`, `README.md` — document the field.

---

## Task 1: Config `aliases` field

**Files:**
- Modify: `internal/config/config.go` (struct `Config`, func `WriteDefault`)
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

Add to `internal/config/config_test.go`. This mirrors the existing
`TestSaveAndLoad_WithHomeConfigPath` Save→Load round-trip pattern:

```go
func TestSaveAndLoad_Aliases(t *testing.T) {
	testutil.SetupHome(t)
	cfg := &Config{
		BaseDir: "~/Developer",
		Aliases: map[string]string{"nahime0": "nahime"},
		Repos: []Repo{
			{URL: "git@github.com:nahime0/repo.git"},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Aliases["nahime0"] != "nahime" {
		t.Errorf("Aliases[nahime0] = %q, want %q", loaded.Aliases["nahime0"], "nahime")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestLoad_Aliases -v`
Expected: FAIL — `cfg.Aliases` is undefined (compile error) or nil.

- [ ] **Step 3: Add the field**

In `internal/config/config.go`, add to the `Config` struct (after `PullStrategy`):

```go
type Config struct {
	BaseDir      string            `mapstructure:"base_dir" yaml:"base_dir"`
	PullStrategy PullStrategy      `mapstructure:"pull_strategy" yaml:"pull_strategy,omitempty"`
	Aliases      map[string]string `mapstructure:"aliases" yaml:"aliases,omitempty"`
	Repos        []Repo            `mapstructure:"repos" yaml:"repos"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/config/ -run TestLoad_Aliases -v`
Expected: PASS

- [ ] **Step 5: Add commented example to the default config**

In `internal/config/config.go`, `WriteDefault`, update the heredoc `content` to include an `aliases` example between `pull_strategy` and `repos`:

```go
	content := `# GGG Configuration
base_dir: ~/Developer

# Default pull strategy for all repos: merge, rebase, ff-only
# pull_strategy: merge

# Map a repo owner to a folder name. Repos owned by the key are cloned under
# the value instead, on any host. Repos with an explicit path are unaffected.
# aliases:
#   nahime0: nahime

repos:
  - url: git@github.com:user/repo.git
  # path: custom/path           # optional, derived from URL if omitted
  # groups: [work, oss]         # optional, assign the repo to one or more groups
  # pull_strategy: rebase       # optional, overrides global strategy
`
```

- [ ] **Step 6: Run config tests**

Run: `go test ./internal/config/`
Expected: PASS (existing `TestWriteDefault*` only assert `base_dir`/`repos` presence, so the new comment is safe).

- [ ] **Step 7: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add owner aliases field"
```

---

## Task 2: `ApplyOwnerAlias` and `DerivedPath` helpers

**Files:**
- Modify: `internal/repo/repo.go`
- Test: `internal/repo/repo_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `internal/repo/repo_test.go`:

```go
func TestApplyOwnerAlias(t *testing.T) {
	aliases := map[string]string{"nahime0": "nahime", "grp": "work"}
	tests := []struct {
		name    string
		relPath string
		aliases map[string]string
		want    string
	}{
		{"match owner", "nahime0/repo", aliases, "nahime/repo"},
		{"match nested owner", "grp/sub/repo", aliases, "work/sub/repo"},
		{"no match", "other/repo", aliases, "other/repo"},
		{"nil map", "nahime0/repo", nil, "nahime0/repo"},
		{"empty map", "nahime0/repo", map[string]string{}, "nahime0/repo"},
		{"single segment", "nahime0", aliases, "nahime0"},
		{"empty path", "", aliases, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ApplyOwnerAlias(tt.relPath, tt.aliases); got != tt.want {
				t.Errorf("ApplyOwnerAlias(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestDerivedPath(t *testing.T) {
	aliases := map[string]string{"nahime0": "nahime"}
	r := config.Repo{URL: "git@github.com:nahime0/repo.git"}
	got, err := DerivedPath(r, aliases)
	if err != nil {
		t.Fatal(err)
	}
	if got != "nahime/repo" {
		t.Errorf("DerivedPath = %q, want %q", got, "nahime/repo")
	}
}

func TestDerivedPath_NoAlias(t *testing.T) {
	r := config.Repo{URL: "https://github.com/org/project.git"}
	got, err := DerivedPath(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "org/project" {
		t.Errorf("DerivedPath = %q, want %q", got, "org/project")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/repo/ -run 'TestApplyOwnerAlias|TestDerivedPath' -v`
Expected: FAIL — `ApplyOwnerAlias`/`DerivedPath` undefined (compile error).

- [ ] **Step 3: Implement the helpers**

In `internal/repo/repo.go`, add after `DerivePathFromURL` (the `strings` package is already imported):

```go
// ApplyOwnerAlias replaces the first segment (the owner) of a derived relative
// path when it matches an alias key. Returns relPath unchanged when there is no
// match, the map is empty/nil, or the path has no owner segment.
func ApplyOwnerAlias(relPath string, aliases map[string]string) string {
	if len(aliases) == 0 || relPath == "" {
		return relPath
	}
	owner, rest, hasRest := strings.Cut(relPath, "/")
	alias, ok := aliases[owner]
	if !ok {
		return relPath
	}
	if !hasRest {
		return alias
	}
	return alias + "/" + rest
}

// DerivedPath returns the repo's path derived from its URL, with owner aliases
// applied. It ignores any explicit Path set on the repo.
func DerivedPath(r config.Repo, aliases map[string]string) (string, error) {
	derived, err := DerivePathFromURL(r.URL)
	if err != nil {
		return "", err
	}
	return ApplyOwnerAlias(derived, aliases), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/repo/ -run 'TestApplyOwnerAlias|TestDerivedPath' -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/repo/repo.go internal/repo/repo_test.go
git commit -m "feat(repo): add owner alias path helpers"
```

---

## Task 3: Thread aliases through `FullPath` and all call sites

Changing the `FullPath` signature breaks every caller, so this whole task is one compiling commit.

**Files:**
- Modify: `internal/repo/repo.go` (`FullPath`)
- Modify: `internal/repo/repo_test.go` (3 existing `FullPath` tests)
- Modify (call sites, all `repo.FullPath(<base>, r)` → `repo.FullPath(<base>, <aliases>, r)`):
  `internal/cli/add.go:55`, `list.go:50`, `completion.go:99`, `stash.go:44`,
  `pull.go:45`, `open.go:45`, `checkout.go:49`, `cd.go:29`, `diff.go:38`,
  `push.go:46`, `clone.go:44`, `validate.go:58`, `doctor.go:102`,
  `outdated.go:42`, `status.go:56`

- [ ] **Step 1: Update the existing `FullPath` tests (they fail to compile first)**

In `internal/repo/repo_test.go`, update the three calls:

```go
// TestFullPath_WithCustomPath
got, err := FullPath("/base", nil, r)

// TestFullPath_DerivedFromURL
got, err := FullPath("/base", nil, r)

// TestFullPath_HTTPSUrl
got, err := FullPath("/home/dev", nil, r)
```

Then add a new test for the alias path:

```go
func TestFullPath_WithAlias(t *testing.T) {
	r := config.Repo{URL: "git@github.com:nahime0/repo.git"}
	aliases := map[string]string{"nahime0": "nahime"}
	got, err := FullPath("/base", aliases, r)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/base", "nahime/repo")
	if got != want {
		t.Errorf("FullPath = %q, want %q", got, want)
	}
}

func TestFullPath_ExplicitPathIgnoresAlias(t *testing.T) {
	r := config.Repo{URL: "git@github.com:nahime0/repo.git", Path: "custom/path"}
	aliases := map[string]string{"nahime0": "nahime"}
	got, err := FullPath("/base", aliases, r)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/base", "custom/path")
	if got != want {
		t.Errorf("FullPath = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run repo tests to verify they fail**

Run: `go test ./internal/repo/ -run TestFullPath -v`
Expected: FAIL — compile error (`FullPath` still takes 2 args).

- [ ] **Step 3: Change `FullPath` to accept aliases**

In `internal/repo/repo.go`, replace `FullPath`:

```go
// FullPath returns the absolute path where a repo should be cloned.
func FullPath(baseDir string, aliases map[string]string, r config.Repo) (string, error) {
	if r.Path != "" {
		return filepath.Join(baseDir, r.Path), nil
	}
	rel, err := DerivedPath(r, aliases)
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, rel), nil
}
```

- [ ] **Step 4: Run repo tests to verify they pass**

Run: `go test ./internal/repo/`
Expected: PASS

- [ ] **Step 5: Update all CLI `FullPath` call sites**

In each file below, insert the config's alias map as the second argument. The variable is `cfg` everywhere except `add.go` (`cfgFull`) and `validate.go` (`cfgExpanded`):

```go
// internal/cli/add.go:55
fullPath, err := repo.FullPath(cfgFull.BaseDir, cfgFull.Aliases, newRepo)

// internal/cli/validate.go:58
fullPath, err := repo.FullPath(cfgExpanded.BaseDir, cfgExpanded.Aliases, r)

// each of: list.go:50, completion.go:99, stash.go:44, pull.go:45, open.go:45,
// checkout.go:49, cd.go:29, diff.go:38, push.go:46, clone.go:44, doctor.go:102,
// outdated.go:42, status.go:56
fullPath, err := repo.FullPath(cfg.BaseDir, cfg.Aliases, r)
```

- [ ] **Step 6: Build to confirm everything compiles**

Run: `go build ./...`
Expected: no output (success).

- [ ] **Step 7: Run the full test suite**

Run: `go test ./...`
Expected: PASS (matching/completion still use `DerivePathFromURL` and are untouched here).

- [ ] **Step 8: Commit**

```bash
git add internal/repo/repo.go internal/repo/repo_test.go internal/cli/
git commit -m "feat(repo): apply owner aliases when resolving clone path"
```

---

## Task 4: Alias-aware repo matching

Make `cd`, `open`, `browse`, `remove`, and the bulk commands resolve repos by their alias-applied derived path.

**Files:**
- Modify: `internal/cli/helpers.go` (`filterByName`, `resolveOneRepo`, `resolveOneRepoIndex`, `matchRepoIndices`, `resolveBulkRepos`)
- Modify callers: `internal/cli/checkout.go:34`, `list.go:38`, `browse.go:32`, `cd.go:24`, `open.go:40`, `remove.go:42`

- [ ] **Step 1: Update the matching helpers in `helpers.go`**

Change the four helper signatures to take `aliases map[string]string` and use `repo.DerivedPath` instead of `repo.DerivePathFromURL`:

```go
// filterByName filters repos by substring match on URL, path, or derived path.
func filterByName(repos []config.Repo, filter string, aliases map[string]string) []config.Repo {
	if filter == "" {
		return repos
	}
	filterLower := strings.ToLower(filter)
	var filtered []config.Repo
	for _, r := range repos {
		derived, _ := repo.DerivedPath(r, aliases)
		if strings.Contains(strings.ToLower(r.URL), filterLower) ||
			strings.Contains(strings.ToLower(r.Path), filterLower) ||
			strings.Contains(strings.ToLower(derived), filterLower) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func resolveOneRepo(repos []config.Repo, query string, aliases map[string]string) (config.Repo, error) {
	idx, err := resolveOneRepoIndex(repos, query, aliases)
	if err != nil {
		return config.Repo{}, err
	}
	return repos[idx], nil
}

func resolveOneRepoIndex(repos []config.Repo, query string, aliases map[string]string) (int, error) {
	matches := matchRepoIndices(repos, query, aliases)
	if len(matches) == 0 {
		return -1, fmt.Errorf("repository %q not found in config", query)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return selectRepoIndex(repos, matches, fmt.Sprintf("Multiple repositories match %q", query))
}

func matchRepoIndices(repos []config.Repo, query string, aliases map[string]string) []int {
	// First pass: exact match.
	for i, r := range repos {
		derived, _ := repo.DerivedPath(r, aliases)
		if r.Path == query || r.URL == query || derived == query {
			return []int{i}
		}
	}

	// Second pass: partial match (substring, case-insensitive).
	queryLower := strings.ToLower(query)
	var matches []int
	for i, r := range repos {
		derived, _ := repo.DerivedPath(r, aliases)
		if strings.Contains(strings.ToLower(r.URL), queryLower) ||
			strings.Contains(strings.ToLower(r.Path), queryLower) ||
			strings.Contains(strings.ToLower(derived), queryLower) {
			matches = append(matches, i)
		}
	}

	return matches
}
```

And update `resolveBulkRepos` (same file) to pass the alias map it already has via `cfg`:

```go
	filter := getFilter(cmd, args)
	repos = filterByName(repos, filter, cfg.Aliases)
	return cfg, repos, filter, nil
```

- [ ] **Step 2: Update the callers**

```go
// internal/cli/list.go:38
repos = filterByName(repos, getFilter(cmd, args), cfg.Aliases)

// internal/cli/checkout.go:34
repos = filterByName(repos, filter, cfg.Aliases)

// internal/cli/cd.go:24
r, err := resolveOneRepo(cfg.Repos, args[0], cfg.Aliases)

// internal/cli/browse.go:32
r, err := resolveOneRepo(cfg.Repos, args[0], cfg.Aliases)

// internal/cli/open.go:40
r, err := resolveOneRepo(cfg.Repos, args[0], cfg.Aliases)

// internal/cli/remove.go:42
idx, err = resolveOneRepoIndex(cfg.Repos, args[0], cfg.Aliases)
```

For `checkout.go`, confirm `cfg` is in scope at line 34 (it is — `cfg` is loaded earlier in the command). If the local variable has a different name, use that name.

- [ ] **Step 3: Build to confirm everything compiles**

Run: `go build ./...`
Expected: no output (success).

- [ ] **Step 4: Run the full test suite**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): resolve repos by alias-applied path"
```

---

## Task 5: Alias-aware shell completion

**Files:**
- Modify: `internal/cli/completion.go` (`repoCompletionItems` and its caller `configuredRepoCompletions`)

- [ ] **Step 1: Thread aliases into `repoCompletionItems`**

In `internal/cli/completion.go`, change the signature and the derive call:

```go
func repoCompletionItems(repos []config.Repo, toComplete string, aliases map[string]string) []string {
	seen := map[string]bool{}
	var comps []string

	for _, r := range repos {
		repoPath := r.Path
		if repoPath == "" {
			derived, err := repo.DerivedPath(r, aliases)
			if err != nil {
				continue
			}
			repoPath = derived
		}
		// ... rest unchanged ...
```

- [ ] **Step 2: Update the caller**

In `configuredRepoCompletions` (same file, ~line 39):

```go
	return repoCompletionItems(repos, toComplete, cfg.Aliases), cobra.ShellCompDirectiveNoFileComp
```

- [ ] **Step 3: Build and test**

Run: `go build ./... && go test ./...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/cli/completion.go
git commit -m "feat(cli): apply owner aliases in shell completion"
```

---

## Task 6: End-to-end test

**Files:**
- Modify: `tests/e2e_test.go`

- [ ] **Step 1: Write the e2e test**

Add to `tests/e2e_test.go` (uses the existing `runGGG`, `writeConfig`, `testutil.SetupHome` helpers — confirm their exact names/signatures against existing tests in the file and match them):

```go
func TestCLIOwnerAlias(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	writeConfig(t, home, fmt.Sprintf(`
base_dir: %s
aliases:
  nahime0: nahime
repos:
  - url: git@github.com:nahime0/repo.git
`, filepath.ToSlash(baseDir)))

	// list should show the alias-applied path, not the raw owner.
	out, err := runGGG(t, home, "list")
	if err != nil {
		t.Fatalf("ggg list failed: %v\n%s", err, out)
	}
	aliased := filepath.Join(baseDir, "nahime", "repo")
	if !strings.Contains(out, aliased) {
		t.Fatalf("list output missing aliased path %q:\n%s", aliased, out)
	}
	if strings.Contains(out, filepath.Join(baseDir, "nahime0", "repo")) {
		t.Fatalf("list output still shows un-aliased path:\n%s", out)
	}

	// cd should resolve the repo and print the alias-applied path.
	out, err = runGGG(t, home, "cd", "repo")
	if err != nil {
		t.Fatalf("ggg cd failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, aliased) {
		t.Fatalf("cd output missing aliased path %q:\n%s", aliased, out)
	}
}
```

If `list` prints paths in a form other than the raw `filepath.Join` string (e.g. with `~` collapsing), adjust the assertion to match how the existing `TestCLIAddListValidateRemove` asserts the custom path — it uses `filepath.Join(baseDir, "custom", "two")`, so the same form is expected to work here.

- [ ] **Step 2: Run the e2e test**

Run: `go test ./tests/ -run TestCLIOwnerAlias -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add tests/e2e_test.go
git commit -m "test(e2e): cover owner aliases in list and cd"
```

---

## Task 7: Documentation

**Files:**
- Modify: `docs/configuration.md`
- Modify: `README.md`

- [ ] **Step 1: Document the field in `docs/configuration.md`**

Add a row to the Top-Level fields table (after `pull_strategy`):

```
| `aliases` | `map` | No | — | Map of repo owner → folder name. Repos owned by the key are cloned under the value (on any host) unless they set an explicit `path`. |
```

Add a section after "Pull Strategy":

````markdown
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
````

- [ ] **Step 2: Mention it in `README.md`**

In the Configuration/Documentation area of `README.md`, add a short line near the config reference link, e.g.:

```
- Owner aliases let you map a repo owner (e.g. `nahime0`) to a folder name (`nahime`); see the [Configuration Reference](docs/configuration.md#owner-aliases).
```

Place it wherever it reads naturally alongside the existing doc links.

- [ ] **Step 3: Commit**

```bash
git add docs/configuration.md README.md
git commit -m "docs: document owner aliases"
```

---

## Final Verification

- [ ] Run `go build ./...` — success.
- [ ] Run `go test ./...` — all pass.
- [ ] Run `gofmt -l internal/ tests/` — no files listed (all formatted).
- [ ] Manually skim `git diff main --stat` to confirm only intended files changed.
