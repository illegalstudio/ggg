# GGG (Go Git Get) — Agent Guide

## Project Overview

GGG is a CLI tool written in Go that manages git repositories from a YAML configuration file. It clones, pulls, pushes, validates, imports, and monitors repositories defined in `~/.config/ggg/repositories.yaml`, with shell integration for quick navigation.

## Architecture

```
main.go          → Entry point, calls cmd.Execute()
cmd/             → Cobra CLI commands grouped as config, repo operations, and diagnostics
                 → Config commands: init, config, add, remove, import, export, shell-init
                 → Repo commands: clone, pull, push, stash, checkout, diff, open, browse, cd
                 → Info/diagnostics: list, status, outdated, doctor, validate
                 → Shared helpers in cmd/helpers.go and matcher helpers in clone.go/remove.go
config/          → Configuration loading, raw parsing, saving, default template, pull-strategy resolution
repo/            → Git operations (clone, pull, push, fetch, stash, checkout, status, branch detection, ahead/behind)
internal/testutil/ → Hermetic test helpers for HOME, config files, and git subprocesses
tests/           → End-to-end CLI tests that build the ggg binary and exercise real commands
ui/              → Shared lipgloss styles for terminal output
```

## Tech Stack

- **CLI framework**: [cobra](https://github.com/spf13/cobra)
- **Config management**: [viper](https://github.com/spf13/viper) (loading) + [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) (saving)
- **Terminal UI**: [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) (styling) + [charmbracelet/huh](https://github.com/charmbracelet/huh) (prompts, spinners)
- **Git operations**: `os/exec` calling the `git` binary directly
- **External integrations**: GitHub import via `gh`, editor launching via `$EDITOR` / `$VISUAL`, browser launching via OS-specific commands

## Coding Style

- **Language**: Go 1.24+
- **Formatting**: Standard `gofmt`
- **Naming**: Follow Go conventions — exported names are PascalCase, unexported are camelCase
- **Error handling**: Return `fmt.Errorf("context: %w", err)` for wrapping; print user-facing errors with `ui.Error` styling
- **Struct tags**: Use both `mapstructure` (for Viper) and `yaml` (for yaml.v3) tags on config structs
- **Commands**: Each command lives in its own file under `cmd/` (e.g. `cmd/clone.go`) and registers itself via `func init() { rootCmd.AddCommand(...) }`
- **Command grouping**: Use the root command group IDs (`GroupConfig`, `GroupRepo`, `GroupInfo`) so help output stays organized
- **Shared logic**:
  - `loadRepos`, `resolveBulkRepos`, `confirmAll`, `confirmBulkAction`, `getFilter`, `filterByName`, `resolveOneRepo`, `resolveOneRepoIndex`, `runParallelWithSpinner`, `defaultEditor`, `requireBinary` live in `cmd/helpers.go`
  - Single-target commands (`cd`, `open`, `browse`, `remove`) share the same exact-match / partial-match / interactive disambiguation flow via `resolveOneRepo*`
- **Filtering model**: Bulk commands should prefer `resolveBulkRepos()` so group filtering, positional `filter`, and `--filter` stay consistent across commands. Single-target commands should prefer `resolveOneRepo()` / `resolveOneRepoIndex()`.
- **Parallelism**: I/O-heavy operations (clone, pull, push, fetch, stash, checkout, status checks, remote reachability checks) run in parallel goroutines with `sync.WaitGroup`, wrapped in a `huh/spinner`
- **Bulk command contract**: Multi-repo commands should be structured as: resolve targets, build actionable jobs, optionally confirm, run with `runParallelWithSpinner()` when appropriate, then render results
- **Git calls**: Use quiet git flags where the command supports them; capture stdout only when output is the actual data being parsed or displayed
- **UI output**: Normal command output is usually `fmt.Print*` combined with styles from `ui/styles.go`; fatal errors are generally returned to Cobra rather than printed inline
- **Interactive flows**: `huh` is used not only for confirmation prompts, but also for repo disambiguation, import multi-select, export path input, and validate cleanup
- **Multi-repo confirmation**: Bulk repo operations that affect multiple repositories or emit repo-by-repo action output should call `confirmAll()` when no explicit filter/name is provided. Current commands following this rule are `clone`, `pull`, `push`, `stash`, `checkout`, and `diff`.

## Testing

- **Run all tests**: `go test ./...`
- **Run with verbose**: `go test ./... -v`
- **Test files**: Named `*_test.go` in the same package (not `_test` suffix packages)
- **Test layout**:
  - Unit/package tests stay next to the code they cover (`cmd/`, `config/`, `repo/`)
  - `tests/` is reserved for end-to-end CLI coverage against the compiled binary
  - `internal/testutil/` contains shared hermetic test helpers and should be preferred over ad-hoc env setup in individual tests
- **Test helpers**: `initTestRepo(t)` in `repo/repo_test.go` creates a temporary git repo with one commit
- **Git test caveat**: `repo/` tests shell out to the real `git` binary. If the developer machine has global commit signing, hooks, or SSH-based signing enabled, `initTestRepo` can fail during `git commit`. Prefer disabling signing for the test process or neutralizing global git config when adjusting those tests.
- **External command tests**: Browser/editor/GitHub CLI flows are covered in `tests/` using local stub executables and hermetic environment variables instead of real external apps or network calls
- **Coverage areas**:
  - `repo/`: URL parsing, path derivation, git operations (clone, branch, dirty, ahead/behind)
  - `config/`: Write, save/load roundtrip, path validation
  - `cmd/`: filterRepo (exact, partial, case-insensitive), filterByGroup, findRepoIndex, URL-to-browser conversion
  - `tests/`: binary-level flows such as `init`, `shell-init`, add/remove/list, export, clone, status, outdated, pull, push, diff, stash, checkout, open, browse, doctor, import, and `cd`
- **Convention**: Use table-driven tests for functions with multiple input/output cases; use `t.TempDir()` for filesystem tests

## Documentation

User-facing documentation lives in `docs/`:

- `docs/configuration.md` — Full configuration reference (fields, pull strategy, examples)
- `docs/commands.md` — All CLI commands with usage, flags, and examples
- `docs/groups.md` — How groups work and filtering with `--group/-g`
- `docs/shell-integration.md` — Shell integration with `ggg cd`

**Documentation must always be kept up to date.** When adding or modifying commands, config fields, or behavior:
1. Update the relevant `docs/` file(s)
2. Update `README.md` if the commands table or quick start section is affected
3. Update this `AGENTS.md` if architecture, coding style, or conventions change

## Configuration

Config file location: `~/.config/ggg/repositories.yaml`

```yaml
base_dir: ~/Developer
pull_strategy: rebase  # optional: merge (default), rebase, ff-only

repos:
  - url: git@github.com:user/repo.git
  - url: https://github.com/org/project.git
    path: custom/path          # optional
    group: work                # optional
    pull_strategy: ff-only     # optional, overrides global
```

## Build

```bash
go build -o ggg .
```

The `ggg` binary is in `.gitignore`.
