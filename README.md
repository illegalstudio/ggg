# GGG (Go Git Get)

Clone and manage git repositories from a YAML configuration file.

## Installation

```bash
go install github.com/illegalstudio/ggg/cmd/ggg@latest
```

Or build from source:

```bash
git clone https://github.com/illegalstudio/ggg.git
cd ggg
go build -o ggg ./cmd/ggg
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
| `ggg cd <name>` | Shell integration — `eval $(ggg cd <name>)` |
| `ggg import [org] [repo]` | Import repositories from GitHub via `gh` CLI |
| `ggg export [path]` | Export the configuration file to a given path |
| `ggg stash [name]` | Stash changes in dirty repositories |
| `ggg checkout <branch>` | Checkout a branch across repositories |
| `ggg diff [name]` | Show changed files in dirty repositories |
| `ggg doctor` | Run health checks on config and repositories |
| `ggg outdated` | Show repositories that are behind their remote |
| `ggg validate` | Validate config for duplicates and conflicts |
| `ggg shell-init` | Print shell integration script (`gcd` alias) |

Most commands support `--group/-g` to filter by group. Data-producing commands support `--json` for machine-readable output; commands that launch an editor or browser report that JSON is unsupported.

## Documentation

Full documentation is available in the [`docs/`](docs/) directory:

- [Configuration Reference](docs/configuration.md)
- [Commands](docs/commands.md)
- [Groups](docs/groups.md)
- [Shell Integration](docs/shell-integration.md)

## License

MIT
