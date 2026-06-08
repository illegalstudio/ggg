# Repo Groups as a List Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a repository belong to multiple groups by changing the config field `group` (string) to `groups` (list), while keeping the `-g/--group` filter flag single-valued (matches repos whose list contains the group).

**Architecture:** Clean break — the `Group string` field is replaced by `Groups []string` with no backward compatibility. The change is cross-cutting within the `cli` package, so all consumers are updated atomically in one task; the `config` package and the e2e suite are updated in their own tasks.

**Tech Stack:** Go 1.24, cobra (CLI flags + completion), viper + yaml.v3 (config), standard `slices`/`strings`.

---

## File map

- `internal/config/config.go` — data model (`Repo.Groups`), `WriteDefault` template.
- `internal/config/config_test.go` — config-package tests.
- `internal/cli/root.go` — `filterByGroup` membership match (add `slices` import).
- `internal/cli/cmd_test.go` — `filterByGroup` tests.
- `internal/cli/add.go` — repeatable `-g` flag, assign `Groups`.
- `internal/cli/import.go` — repeatable `-g` flag, assign `Groups`.
- `internal/cli/list.go` — `listEntry.Groups`, `listGroups` counting.
- `internal/cli/helpers.go` — `repoChoiceLabel` rendering.
- `internal/cli/completion.go` — `groupCompletion` iterates the list.
- `internal/cli/validate.go` — blank-group check over the list.
- `tests/e2e_test.go` — config fixtures + assertions.

---

## Task 1: Config data model + config tests

**Files:**
- Modify: `internal/config/config.go:24-29` (struct), `internal/config/config.go:114-131` (`WriteDefault`)
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Update the failing config tests to the new field**

In `internal/config/config_test.go`, change every `Group:` literal and assertion to `Groups`.

`TestSaveAndLoadRaw` — line ~63:
```go
		{URL: "git@github.com:user/repo3.git", Groups: []string{"work"}},
```
and the assertion at ~94:
```go
	if len(loaded.Repos[2].Groups) != 1 || loaded.Repos[2].Groups[0] != "work" {
		t.Errorf("Repos[2].Groups = %v, want [work]", loaded.Repos[2].Groups)
	}
```

`TestSaveAndLoad_WithHomeConfigPath` — line ~158:
```go
			{URL: "git@github.com:a/b.git", Groups: []string{"work"}},
```

`TestSave_MarshalRoundtrip` — line ~196:
```go
			{URL: "git@github.com:c/d.git", Path: "custom", Groups: []string{"personal"}},
```
and the assertion at ~217:
```go
	if len(loaded.Repos[1].Groups) != 1 || loaded.Repos[1].Groups[0] != "personal" {
		t.Errorf("Repos[1].Groups = %v, want [personal]", loaded.Repos[1].Groups)
	}
```

- [ ] **Step 2: Add a multi-group roundtrip test**

Append to `internal/config/config_test.go`:
```go
func TestSave_MultipleGroupsRoundtrip(t *testing.T) {
	cfg := &Config{
		BaseDir: "~/Dev",
		Repos: []Repo{
			{URL: "git@github.com:a/b.git", Groups: []string{"work", "oss"}},
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}

	if len(loaded.Repos[0].Groups) != 2 ||
		loaded.Repos[0].Groups[0] != "work" ||
		loaded.Repos[0].Groups[1] != "oss" {
		t.Errorf("Groups = %v, want [work oss]", loaded.Repos[0].Groups)
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./internal/config/`
Expected: FAIL — compile error `unknown field 'Groups' in struct literal of type config.Repo` (field still named `Group`).

- [ ] **Step 4: Change the data model**

In `internal/config/config.go`, replace the `Group` field (line ~27):
```go
type Repo struct {
	URL          string       `mapstructure:"url" yaml:"url"`
	Path         string       `mapstructure:"path" yaml:"path,omitempty"`
	Groups       []string     `mapstructure:"groups" yaml:"groups,omitempty"`
	PullStrategy PullStrategy `mapstructure:"pull_strategy" yaml:"pull_strategy,omitempty"`
}
```

- [ ] **Step 5: Update the WriteDefault template**

In `internal/config/config.go`, inside `WriteDefault`, update the `content` template's `repos:` block:
```go
	content := `# GGG Configuration
base_dir: ~/Developer

# Default pull strategy for all repos: merge, rebase, ff-only
# pull_strategy: merge

repos:
  - url: git@github.com:user/repo.git
  # path: custom/path           # optional, derived from URL if omitted
  # groups: [work, oss]         # optional, assign the repo to one or more groups
  # pull_strategy: rebase       # optional, overrides global strategy
`
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./internal/config/`
Expected: PASS (all config tests, including `TestSave_MultipleGroupsRoundtrip`).

- [ ] **Step 7: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): make repo groups a list (groups)"
```

---

## Task 2: CLI consumers + single-group filter (atomic)

The `Group` rename breaks every `cli` file that reads it, so the package only
compiles once all of them are updated. Edit all production files, then the
tests, then build and run the package suite once.

**Files:**
- Modify: `internal/cli/root.go` (imports + `filterByGroup`), `internal/cli/add.go:21,36,110`, `internal/cli/import.go:60,139,356`, `internal/cli/list.go:42,51,57,90-95`, `internal/cli/helpers.go:183-185`, `internal/cli/completion.go:78-80`, `internal/cli/validate.go:92-97`
- Test: `internal/cli/cmd_test.go`

- [ ] **Step 1: filterByGroup — match by membership**

In `internal/cli/root.go`, add `"slices"` to the import block (alongside `errors`, `fmt`, `os`, `strings`), then replace the function body (line ~85):
```go
func filterByGroup(repos []config.Repo, group string) []config.Repo {
	if group == "" {
		return repos
	}
	var filtered []config.Repo
	for _, r := range repos {
		if slices.Contains(r.Groups, group) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
```

- [ ] **Step 2: add.go — repeatable flag, assign Groups**

In `internal/cli/add.go`, read the flag as a slice (line ~21):
```go
		groups, _ := cmd.Flags().GetStringArray("group")
```
build the repo (line ~36):
```go
		newRepo := config.Repo{URL: url, Groups: groups, Path: path}
```
and change the flag registration (line ~110):
```go
	addCmd.Flags().StringArrayP("group", "g", nil, "Assign the repo to one or more groups (repeatable)")
```

- [ ] **Step 3: import.go — repeatable flag, assign Groups**

In `internal/cli/import.go`, read the flag as a slice (line ~60):
```go
		groups, _ := cmd.Flags().GetStringArray("group")
```
build each repo (line ~139):
```go
			cfg.Repos = append(cfg.Repos, config.Repo{URL: url, Groups: groups})
```
and change the flag registration (line ~356):
```go
	importCmd.Flags().StringArrayP("group", "g", nil, "Assign imported repos to one or more groups (repeatable)")
```

- [ ] **Step 4: list.go — entry field + group counting**

In `internal/cli/list.go`, change the `listEntry` struct field (line ~42):
```go
			Groups []string `json:"groups,omitempty"`
```
the error-path append (line ~51):
```go
				entries = append(entries, listEntry{URL: r.URL, Groups: r.Groups, Error: err.Error()})
```
the success append (line ~57):
```go
				Groups: r.Groups,
```
and the counting loop inside `listGroups` (line ~90):
```go
	groups := map[string]int{}
	for _, r := range cfg.Repos {
		for _, g := range r.Groups {
			if g != "" {
				groups[g]++
			}
		}
	}
```

- [ ] **Step 5: helpers.go — label rendering**

In `internal/cli/helpers.go`, `repoChoiceLabel` (line ~183) — `strings` is already imported:
```go
	if len(r.Groups) > 0 {
		label += " [" + strings.Join(r.Groups, ", ") + "]"
	}
```

- [ ] **Step 6: completion.go — iterate the list**

In `internal/cli/completion.go`, `groupCompletion` (line ~78):
```go
	for _, r := range cfg.Repos {
		for _, g := range r.Groups {
			addCompletion(&comps, seen, g, toComplete)
		}
	}
```

- [ ] **Step 7: validate.go — blank check over the list**

In `internal/cli/validate.go` (line ~92) — `strings` is already imported:
```go
		// Check 4: blank group entries (empty or whitespace-only)
		for _, r := range cfg.Repos {
			for _, g := range r.Groups {
				if strings.TrimSpace(g) == "" {
					warnings = append(warnings, fmt.Sprintf("Blank group for %s", r.URL))
				}
			}
		}
```

- [ ] **Step 8: Update filterByGroup tests + add multi-group case**

In `internal/cli/cmd_test.go`, rewrite the three group tests:
```go
func TestFilterByGroup_Empty(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/a.git", Groups: []string{"work"}},
		{URL: "git@github.com:user/b.git", Groups: []string{"personal"}},
		{URL: "git@github.com:user/c.git"},
	}

	result := filterByGroup(repos, "")
	if len(result) != 3 {
		t.Errorf("filterByGroup with empty group should return all repos, got %d", len(result))
	}
}

func TestFilterByGroup_Match(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/a.git", Groups: []string{"work"}},
		{URL: "git@github.com:user/b.git", Groups: []string{"personal"}},
		{URL: "git@github.com:user/c.git", Groups: []string{"work"}},
	}

	result := filterByGroup(repos, "work")
	if len(result) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(result))
	}
	if result[0].URL != "git@github.com:user/a.git" {
		t.Errorf("result[0].URL = %q, want a.git", result[0].URL)
	}
	if result[1].URL != "git@github.com:user/c.git" {
		t.Errorf("result[1].URL = %q, want c.git", result[1].URL)
	}
}

func TestFilterByGroup_NoMatch(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/a.git", Groups: []string{"work"}},
	}

	result := filterByGroup(repos, "oss")
	if len(result) != 0 {
		t.Errorf("expected 0 repos, got %d", len(result))
	}
}

func TestFilterByGroup_MultiGroupMembership(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/a.git", Groups: []string{"work", "oss"}},
		{URL: "git@github.com:user/b.git", Groups: []string{"personal"}},
	}

	if got := filterByGroup(repos, "oss"); len(got) != 1 || got[0].URL != "git@github.com:user/a.git" {
		t.Errorf("filterByGroup(\"oss\") = %v, want [a.git]", got)
	}
	if got := filterByGroup(repos, "work"); len(got) != 1 || got[0].URL != "git@github.com:user/a.git" {
		t.Errorf("filterByGroup(\"work\") = %v, want [a.git]", got)
	}
}
```

- [ ] **Step 9: Build the package to confirm it compiles**

Run: `go build ./...`
Expected: no output (success). If anything still references `r.Group`, the
compiler names the file and line — fix it.

- [ ] **Step 10: Run the cli package tests**

Run: `go test ./internal/cli/`
Expected: PASS, including `TestFilterByGroup_MultiGroupMembership`.

- [ ] **Step 11: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): support multiple groups per repo"
```

---

## Task 3: End-to-end tests

**Files:**
- Modify: `tests/e2e_test.go:138-143` (export fixture), `tests/e2e_test.go:566` (import assertion)

- [ ] **Step 1: Update the export-test config fixture**

In `tests/e2e_test.go`, `TestCLIExportCopiesConfig` (line ~138), change the inline config to the list form:
```go
	writeConfig(t, home, `
base_dir: ~/Developer
repos:
  - url: git@github.com:acme/exported.git
    groups:
      - work
`)
```

- [ ] **Step 2: Update the import group assertion**

In `tests/e2e_test.go`, `TestCLIImportFromGitHub` (line ~566) — after import, each
of the two repos serializes its group list as a `- work` sequence item, so count
that instead of the old `group: work`:
```go
	if strings.Count(configText, "- work") != 2 {
		t.Fatalf("imported group missing from config:\n%s", configText)
	}
```

(The `add`/`import ... --group work` invocations at lines ~63, ~550, ~652 need no
change: a single `--group work` is a one-element repeatable value.)

- [ ] **Step 3: Run the full suite**

Run: `go test ./...`
Expected: PASS across all packages.

- [ ] **Step 4: Commit**

```bash
git add tests/e2e_test.go
git commit -m "test(e2e): adapt fixtures to groups list"
```

---

## Notes

- `slices.Contains` is in the Go 1.24 standard library — no new dependency.
- `--group` filter semantics are unchanged for users: pass one group, get repos
  that contain it.
- The `list --json` field rename (`group` → `groups`) is a deliberate breaking
  change to JSON output; acceptable as the tool has no users yet.
