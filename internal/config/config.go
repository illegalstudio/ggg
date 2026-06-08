package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// PullStrategy defines how `git pull` behaves.
// Valid values: "merge" (default), "rebase", "ff-only".
type PullStrategy string

const (
	PullMerge  PullStrategy = "merge"
	PullRebase PullStrategy = "rebase"
	PullFFOnly PullStrategy = "ff-only"
)

const DefaultPullStrategy = PullMerge

type Repo struct {
	URL          string       `mapstructure:"url" yaml:"url"`
	Path         string       `mapstructure:"path" yaml:"path,omitempty"`
	Groups       []string     `mapstructure:"groups" yaml:"groups,omitempty"`
	PullStrategy PullStrategy `mapstructure:"pull_strategy" yaml:"pull_strategy,omitempty"`
}

type Config struct {
	BaseDir      string       `mapstructure:"base_dir" yaml:"base_dir"`
	PullStrategy PullStrategy `mapstructure:"pull_strategy" yaml:"pull_strategy,omitempty"`
	Repos        []Repo       `mapstructure:"repos" yaml:"repos"`
}

// ResolvePullStrategy returns the effective pull strategy for a repo:
// repo-level > global > default (merge).
func (c *Config) ResolvePullStrategy(r Repo) PullStrategy {
	if r.PullStrategy != "" {
		return r.PullStrategy
	}
	if c.PullStrategy != "" {
		return c.PullStrategy
	}
	return DefaultPullStrategy
}

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ggg", "repositories.yaml")
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	v := viper.New()
	v.SetConfigName("repositories")
	v.SetConfigType("yaml")
	v.AddConfigPath(filepath.Join(home, ".config", "ggg"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	// Expand ~ in base_dir
	if cfg.BaseDir == "" {
		cfg.BaseDir = filepath.Join(home, "Developer")
	} else if len(cfg.BaseDir) >= 2 && cfg.BaseDir[:2] == "~/" {
		cfg.BaseDir = filepath.Join(home, cfg.BaseDir[2:])
	}

	// Validate repos
	for i, r := range cfg.Repos {
		if r.URL == "" {
			return nil, fmt.Errorf("repo #%d has no URL", i+1)
		}
	}

	return &cfg, nil
}

// LoadRaw reads the config file without expanding paths.
func LoadRaw() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}
	return &cfg, nil
}

// Save writes the config to the config file.
func Save(cfg *Config) error {
	path := ConfigPath()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot serialize config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

func WriteDefault(path string) error {
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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}
