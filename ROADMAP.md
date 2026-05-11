# Roadmap

## v1.0.0

- [x] Define YAML configuration format (`~/.config/ggg.yaml`)
- [x] Integrate Cobra as CLI framework
- [x] `ggg init` command — generate sample configuration file
- [x] `ggg list` command — show configured repositories and their clone status
- [x] `ggg clone` command — clone all configured repositories
- [x] `ggg clone <name>` command — clone a single repository
- [x] Auto-derive destination path from repo URL (e.g. `github.com/user/repo`)
- [x] Support `base_dir` in configuration for clone root directory
- [x] Error handling (already cloned, invalid URL, non-writable directory)
- [x] Add unit tests
- [x] Complete README with installation, usage, and examples
- [x] `ggg pull` command — pull all repos (only if clean)
- [x] `ggg status` command — show branch, dirty/clean, ahead/behind
- [x] `ggg add <url>` command — add a repo to config from CLI
- [x] `ggg remove <name>` command — remove a repo from config
- [x] Parallel clone and pull with spinner
- [x] Group/tag support with `--group/-g` flag
- [x] `ggg cd <name>` command — shell integration (`eval $(ggg cd <name>)`)
- [x] Styled output with Charm stack (lipgloss + huh)
- [ ] Cross-platform build (Linux, macOS, Windows)
- [ ] First GitHub release with prebuilt binaries

## v1.1.0

- [x] `ggg doctor` command — health check: valid config, orphaned repos, reachable remotes
- [x] `ggg outdated` command — show repos behind their remote
- [x] `ggg open <name> [editor]` command — open repo in editor (default: $EDITOR)
- [x] `ggg browse <name>` command — open repo in browser
- [ ] `ggg sync` command — clone missing + pull clean in a single command
- [x] `ggg import [org]` command — import repos from GitHub via `gh` CLI
- [x] `ggg export` command — export config in a shareable format
- [x] Shell alias — `gcd` shell function for navigation without `eval`
- [ ] Dynamic completions — repo name autocompletion for bash/zsh/fish
- [ ] Config watch — `--watch` flag on status with periodic refresh
- [ ] Dirty notifications — shell prompt integration (starship/p10k)

## v1.2.0

- [x] `ggg add <url> --clone` command — add and clone in one shot
- [x] `ggg add <url> --group <name>` command — specify group and path from CLI
- [x] `ggg list --groups` command — show available groups
- [ ] `ggg rename <old> <new>` command — rename a repo path/alias
- [x] `ggg stash [name]` command — run `git stash` on all dirty repos
- [ ] `ggg branch [name]` command — show or filter repos by current branch
- [x] `ggg checkout <branch> [name]` command — checkout a branch across repos
- [x] `ggg validate` command — deep config validation (duplicate URLs, conflicting paths)
- [x] `ggg diff [name]` command — summary of changed files across dirty repos
- [ ] Multi-config support — merge multiple YAML files (e.g. `work.yaml` + `personal.yaml`)
