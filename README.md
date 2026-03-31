# GGG (Go Git Get)

Clone and manage git repositories from a YAML configuration file.

## Installation

```bash
go install github.com/illegalstudio/go-git-get@latest
```

Or build from source:

```bash
git clone https://github.com/illegalstudio/go-git-get.git
cd go-git-get
go build -o ggg .
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
| `ggg list` | List configured repositories and their clone status |
| `ggg clone [name]` | Clone repositories (all or a specific one) |
| `ggg pull [name]` | Pull latest changes (only if repo is clean) |
| `ggg status` | Show branch, dirty/clean, ahead/behind for all repos |
| `ggg add <url>` | Add a repository to the configuration |
| `ggg remove <name>` | Remove a repository from the configuration |
| `ggg cd <name>` | Shell integration — `eval $(ggg cd <name>)` |
| `ggg doctor` | Run health checks on config and repositories |
| `ggg outdated` | Show repositories that are behind their remote |

Most commands support `--group/-g` to filter by group.

## Documentation

Full documentation is available in the [`doc/`](doc/) directory:

- [Configuration Reference](doc/configuration.md)
- [Commands](doc/commands.md)
- [Groups](doc/groups.md)
- [Shell Integration](doc/shell-integration.md)

## License

MIT
