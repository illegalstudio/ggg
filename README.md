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

## Configuration

The configuration file is located at `~/.config/ggg/repositories.yaml`.

```yaml
# Base directory for all cloned repositories
base_dir: ~/Developer

repos:
  - url: git@github.com:user/repo.git
  - url: https://github.com/org/project.git
    path: custom/path  # optional, derived from URL if omitted
```

### Fields

| Field | Description | Required |
|-------|-------------|----------|
| `base_dir` | Root directory for clones. Defaults to `~/Developer` | No |
| `repos[].url` | Git repository URL (SSH or HTTPS) | Yes |
| `repos[].path` | Custom clone path relative to `base_dir`. If omitted, derived from the URL (e.g. `github.com/user/repo`) | No |

## Commands

| Command | Description |
|---------|-------------|
| `ggg init` | Generate a default configuration file |
| `ggg list` | List configured repositories and their clone status |
| `ggg clone` | Clone all configured repositories |
| `ggg clone <name>` | Clone a specific repository (by URL, path, or derived name) |

## License

MIT
