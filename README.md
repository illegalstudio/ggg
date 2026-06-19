<h1 align="center">GGG</h1>

<p align="center">
  <em>All your git repos, declared in one YAML file.</em>
</p>

<p align="center">
  <a href="https://github.com/illegalstudio/ggg/stargazers"><img src="https://img.shields.io/github/stars/illegalstudio/ggg?style=flat-square&logo=github&logoColor=white&label=stars&color=00ADD8" alt="Stars"></a>
  <a href="https://github.com/illegalstudio/ggg/releases"><img src="https://img.shields.io/github/downloads/illegalstudio/ggg/total?style=flat-square&logo=github&logoColor=white&label=downloads&color=00ADD8" alt="Downloads"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/illegalstudio/ggg?style=flat-square&color=00ADD8" alt="License: MIT"></a>
  <a href="https://x.com/nahime0"><img src="https://img.shields.io/badge/Follow-%40nahime0-00ADD8?style=flat-square&logo=x&logoColor=white" alt="Follow @nahime0 on X"></a>
</p>

<p align="center">
  <strong>Declarative YAML config &middot; Clone &amp; sync many repos &middot; Shell <code>cd</code> integration &middot; Single Go binary</strong>
</p>

<p align="center">
  GGG (Go Git Get) clones and manages all your git repositories from a single YAML
  configuration file — clone, pull, push, and check status across every repo at once,
  jump between them with a real shell <code>cd</code>, and import whole orgs from GitHub.
</p>

---

GGG has a brother, [GGW](https://github.com/illegalstudio/ggw), that helps you manage your git worktrees.

## Installation

### Homebrew

```bash
brew install illegalstudio/tap/ggg
```

### Go

```bash
go install github.com/illegalstudio/ggg/cmd/ggg@latest
```

### From source

```bash
git clone https://github.com/illegalstudio/ggg.git
cd ggg
make && make install
```

## Quick Start

```bash
# Generate a default configuration file
ggg init

# Edit the config to add your repositories
# vim ~/.config/ggg/repositories.yaml

# Clone all configured repositories
ggg clone

# Clone a specific repository
ggg clone github.com/user/repo

# List repositories and their status
ggg list

# Optional: enable gcd and shell completions
eval "$(ggg shell-init zsh)"

# Print the installed version
ggg --version
```

## Commands

| Command | Description |
|---------|-------------|
| `ggg init` | Generate a default configuration file |
| `ggg config` | Open the configuration file in your editor |
| `ggg list` | List configured repositories and their clone status |
| `ggg clone [name]` | Clone repositories (all or a specific one) |
| `ggg pull [name]` | Pull latest changes (only if repo is clean) |
| `ggg push [name]` | Push commits to remote for repos that are ahead |
| `ggg status` | Show branch, dirty/clean, ahead/behind for all repos |
| `ggg add <url>` | Add a repository to the configuration |
| `ggg remove [name]` | Remove a repository from the configuration |
| `ggg open <name>` | Open a repository in your editor |
| `ggg browse <name>` | Open a repository's remote URL in the browser |
| `ggg cd <name>` | Print a repository path for shell navigation |
| `ggg import [org] [repo]` | Import repositories from GitHub via `gh` CLI |
| `ggg export [path]` | Export the configuration file to a given path |
| `ggg stash [name]` | Stash changes in dirty repositories |
| `ggg checkout <branch>` | Checkout a branch across repositories |
| `ggg diff [name]` | Show changed files in dirty repositories |
| `ggg doctor` | Run health checks on config and repositories |
| `ggg outdated` | Show repositories that are behind their remote |
| `ggg validate` | Validate config for duplicates and conflicts |
| `ggg shell-init` | Print shell integration script (`gcd` alias and completions) |

Most commands support `--group/-g` to filter by group. Data-producing commands support `--json` for machine-readable output; commands that launch an editor or browser report that JSON is unsupported.

## Documentation

Full documentation is available in the [`docs/`](docs/) directory:

- [Configuration Reference](docs/configuration.md)
- [Commands](docs/commands.md)
- [Groups](docs/groups.md)
- [Shell Integration](docs/shell-integration.md)
- Owner aliases let you map a repo owner (e.g. `nahime0`) to a folder name (`nahime`); see the [Configuration Reference](docs/configuration.md#owner-aliases).

## License

MIT
