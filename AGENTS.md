# GGG (Go Git Get) — Agent Guide

## Project Overview

GGG is a CLI tool written in Go that manages git repositories from a YAML configuration file. It clones, pulls, and monitors repositories defined in `~/.config/ggg/repositories.yaml`.

## Architecture

```
main.go          → Entry point, calls cmd.Execute()
cmd/             → Cobra CLI commands (root, init, list, clone, pull, status, add, remove, cd, doctor, outdated)
config/          → Configuration loading, saving, validation (Viper + YAML)
repo/            → Git operations (clone, pull, fetch, status, branch detection)
ui/              → Shared lipgloss styles for terminal output
```

## Tech Stack

- **CLI framework**: [cobra](https://github.com/spf13/cobra)
- **Config management**: [viper](https://github.com/spf13/viper) (loading) + [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) (saving)
- **Terminal UI**: [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) (styling) + [charmbracelet/huh](https://github.com/charmbracelet/huh) (prompts, spinners)
- **Git operations**: `os/exec` calling the `git` binary directly

## Coding Style

- **Language**: Go 1.24+
- **Formatting**: Standard `gofmt`
- **Naming**: Follow Go conventions — exported names are PascalCase, unexported are camelCase
- **Error handling**: Return `fmt.Errorf("context: %w", err)` for wrapping; print user-facing errors with `ui.Error` styling
- **Struct tags**: Use both `mapstructure` (for Viper) and `yaml` (for yaml.v3) tags on config structs
- **Commands**: Each command lives in its own file under `cmd/` (e.g., `cmd/clone.go`). Register via `func init() { rootCmd.AddCommand(...) }`
- **Shared logic**: `filterRepo` and `filterByGroup` in `cmd/` are shared across commands; `findRepoIndex` is used by `remove`
- **Parallelism**: I/O-heavy operations (clone, pull, fetch, status checks) run in parallel goroutines with `sync.WaitGroup`, wrapped in a `huh/spinner`
- **Git calls**: Always use `--quiet` flag for git commands to suppress output; capture stdout only when parsing output (e.g., `git status --porcelain`)
- **UI output**: Always use styles from `ui/styles.go` — never raw `fmt.Println` for user-facing messages

## Testing

- **Run all tests**: `go test ./...`
- **Run with verbose**: `go test ./... -v`
- **Test files**: Named `*_test.go` in the same package (not `_test` suffix packages)
- **Test helpers**: `initTestRepo(t)` in `repo/repo_test.go` creates a temporary git repo with one commit
- **Coverage areas**:
  - `repo/`: URL parsing, path derivation, git operations (clone, branch, dirty, ahead/behind)
  - `config/`: Write, save/load roundtrip, path validation
  - `cmd/`: filterRepo (exact, partial, case-insensitive), filterByGroup, findRepoIndex
- **Convention**: Use table-driven tests for functions with multiple input/output cases; use `t.TempDir()` for filesystem tests

## Configuration

Config file location: `~/.config/ggg/repositories.yaml`

```yaml
base_dir: ~/Developer
repos:
  - url: git@github.com:user/repo.git
  - url: https://github.com/org/project.git
    path: custom/path     # optional
    group: work           # optional
```

## Build

```bash
go build -o ggg .
```

The `ggg` binary is in `.gitignore`.
