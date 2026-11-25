package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

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

// Valid reports whether s is a known pull strategy. The empty value is valid
// and means "not set".
func (s PullStrategy) Valid() bool {
	switch s {
	case "", PullMerge, PullRebase, PullFFOnly:
		return true
	}
	return false
}

// BaseDirEnv names the environment variable that overrides base_dir on the
// current machine. An empty value counts as unset.
const BaseDirEnv = "GGG_BASE_DIR"

// DefaultBaseDir is the base_dir used when no source sets one.
const DefaultBaseDir = "~/Developer"

const (
	configFileName      = "repositories.yaml"
	localConfigFileName = "repositories.local.yaml"
)

// BaseDirSource records where the effective base_dir came from.
type BaseDirSource string

const (
	BaseDirFromDefault BaseDirSource = "default"
	BaseDirFromConfig  BaseDirSource = "config"
	BaseDirFromLocal   BaseDirSource = "local"
	BaseDirFromEnv     BaseDirSource = "env"
)

// Describe returns a short human-readable origin, e.g. "from GGG_BASE_DIR".
func (s BaseDirSource) Describe() string {
	switch s {
	case BaseDirFromConfig:
		return "from " + configFileName
	case BaseDirFromLocal:
		return "from " + localConfigFileName
	case BaseDirFromEnv:
		return "from " + BaseDirEnv
	default:
		return "default"
	}
}

type Repo struct {
	URL          string       `mapstructure:"url" yaml:"url"`
	Path         string       `mapstructure:"path" yaml:"path,omitempty"`
	Groups       []string     `mapstructure:"groups" yaml:"groups,omitempty"`
	PullStrategy PullStrategy `mapstructure:"pull_strategy" yaml:"pull_strategy,omitempty"`
}

type Config struct {
	BaseDir              string            `mapstructure:"base_dir" yaml:"base_dir"`
	PullStrategy         PullStrategy      `mapstructure:"pull_strategy" yaml:"pull_strategy,omitempty"`
	Aliases              map[string]string `mapstructure:"aliases" yaml:"aliases,omitempty"`
	SuppressSkillsNotice bool              `mapstructure:"suppress_skills_notice" yaml:"suppress_skills_notice,omitempty"`
	Repos                []Repo            `mapstructure:"repos" yaml:"repos"`

	// BaseDirSource is set by Load and never read from or written to a file.
	BaseDirSource BaseDirSource `mapstructure:"-" yaml:"-"`
}

// LocalConfig holds the machine-specific overrides read from
// LocalConfigPath(). It never contains repositories, so the shared
// repositories.yaml stays the single list of repos.
type LocalConfig struct {
	BaseDir              string            `yaml:"base_dir,omitempty"`
	PullStrategy         PullStrategy      `yaml:"pull_strategy,omitempty"`
	Aliases              map[string]string `yaml:"aliases,omitempty"`
	SuppressSkillsNotice *bool             `yaml:"suppress_skills_notice,omitempty"`
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

// applyLocal overlays machine-local values on top of the shared config.
// Aliases are merged by key, with the local value winning.
func (c *Config) applyLocal(l *LocalConfig) {
	if l.BaseDir != "" {
		c.BaseDir = l.BaseDir
		c.BaseDirSource = BaseDirFromLocal
	}
	if l.PullStrategy != "" {
		c.PullStrategy = l.PullStrategy
	}
	if len(l.Aliases) > 0 {
		if c.Aliases == nil {
			c.Aliases = make(map[string]string, len(l.Aliases))
		}
		for owner, folder := range l.Aliases {
			c.Aliases[owner] = folder
		}
	}
	if l.SuppressSkillsNotice != nil {
		c.SuppressSkillsNotice = *l.SuppressSkillsNotice
	}
}

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ggg", configFileName)
}

// LocalConfigPath returns the path of the machine-local override file, which
// lives next to the shared config and is meant to stay out of sync tools.
func LocalConfigPath() string {
	return filepath.Join(filepath.Dir(ConfigPath()), localConfigFileName)
}

// SuppressSkillsNotice reports whether the stale-skill notice is disabled. A
// missing config file or an unset key yields (false, nil). A value in the
// local config file overrides the shared one.
func SuppressSkillsNotice() (bool, error) {
	suppress := false
	cfg, err := LoadRaw()
	switch {
	case err == nil:
		suppress = cfg.SuppressSkillsNotice
	case !errors.Is(err, os.ErrNotExist):
		return false, err
	}

	local, err := LoadLocal()
	if err != nil {
		return false, err
	}
	if local != nil && local.SuppressSkillsNotice != nil {
		suppress = *local.SuppressSkillsNotice
	}
	return suppress, nil
}

// Load reads the shared config, overlays the machine-local file and the
// GGG_BASE_DIR environment variable, and expands base_dir. The base_dir
// precedence is GGG_BASE_DIR > local file > shared file > ~/Developer.
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

	cfg.BaseDirSource = BaseDirFromDefault
	if cfg.BaseDir != "" {
		cfg.BaseDirSource = BaseDirFromConfig
	}

	local, err := LoadLocal()
	if err != nil {
		return nil, err
	}
	if local != nil {
		cfg.applyLocal(local)
	}

	if env := os.Getenv(BaseDirEnv); env != "" {
		cfg.BaseDir = env
		cfg.BaseDirSource = BaseDirFromEnv
	}

	if cfg.BaseDir == "" {
		cfg.BaseDir = DefaultBaseDir
	}
	expanded, err := ExpandPath(cfg.BaseDir, home)
	if err != nil {
		return nil, fmt.Errorf("invalid base_dir %q (%s): %w", cfg.BaseDir, cfg.BaseDirSource.Describe(), err)
	}
	cfg.BaseDir = expanded

	// Validate repos
	for i, r := range cfg.Repos {
		if r.URL == "" {
			return nil, fmt.Errorf("repo #%d has no URL", i+1)
		}
	}

	return &cfg, nil
}

// ExpandPath expands $VAR and ${VAR} references and a leading ~ in p. It fails
// when p references an environment variable that is unset or empty, so a
// missing variable never silently turns into a different directory.
func ExpandPath(p, home string) (string, error) {
	var missing []string
	expanded := os.Expand(p, func(name string) string {
		value := os.Getenv(name)
		if value == "" && !slices.Contains(missing, name) {
			missing = append(missing, name)
		}
		return value
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("environment variable %s is not set", strings.Join(missing, ", "))
	}

	if expanded == "~" {
		return home, nil
	}
	if strings.HasPrefix(expanded, "~/") || strings.HasPrefix(expanded, "~"+string(filepath.Separator)) {
		return filepath.Join(home, expanded[2:]), nil
	}
	return expanded, nil
}

// LoadRaw reads the shared config file without expanding paths or applying
// machine-local overrides. Commands that rewrite the shared file use it so
// local values never leak into it.
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

// LoadLocal reads the machine-local override file. It returns (nil, nil) when
// the file does not exist. Unknown keys, a repos list, and invalid pull
// strategies are rejected.
func LoadLocal() (*LocalConfig, error) {
	path := LocalConfigPath()
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot read local config file %s: %w", path, err)
	}

	var keys map[string]any
	if err := yaml.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("cannot parse local config file %s: %w", path, err)
	}
	if _, ok := keys["repos"]; ok {
		return nil, fmt.Errorf("local config file %s cannot define repos; keep them in %s", path, configFileName)
	}
	allowed := []string{"base_dir", "pull_strategy", "aliases", "suppress_skills_notice"}
	for _, key := range slices.Sorted(maps.Keys(keys)) {
		if !slices.Contains(allowed, key) {
			return nil, fmt.Errorf("local config file %s has unknown key %q (allowed: %s)", path, key, strings.Join(allowed, ", "))
		}
	}

	var local LocalConfig
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&local); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("cannot parse local config file %s: %w", path, err)
	}
	if !local.PullStrategy.Valid() {
		return nil, fmt.Errorf("local config file %s has invalid pull_strategy %q", path, local.PullStrategy)
	}
	return &local, nil
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

// WriteDefault writes the starter shared config to path. An empty baseDir
// writes the default ~/Developer.
func WriteDefault(path, baseDir string) error {
	if baseDir == "" {
		baseDir = DefaultBaseDir
	}
	content := `# GGG Configuration
# base_dir can be overridden per machine in repositories.local.yaml
# (ggg init --local) or with the GGG_BASE_DIR environment variable.
base_dir: ` + yamlScalar(baseDir) + `

# Default pull strategy for all repos: merge, rebase, ff-only
# pull_strategy: merge

# Map a repo owner to a folder name. Repos owned by the key are cloned under
# the value instead, on any host. Repos with an explicit path are unaffected.
# aliases:
#   nahime0: nahime

# suppress_skills_notice: set to true to stop ggg from reminding you when an
# installed AI agent skill no longer matches the one bundled with the binary.
# suppress_skills_notice: true

repos:
  - url: git@github.com:user/repo.git
  # path: custom/path           # optional, derived from URL if omitted
  # groups: [work, oss]         # optional, assign the repo to one or more groups
  # pull_strategy: rebase       # optional, overrides global strategy
`
	return writeNewFile(path, content)
}

// WriteLocal writes a machine-local override file at path that sets baseDir.
func WriteLocal(path, baseDir string) error {
	content := `# GGG machine-local configuration
# Values here override repositories.yaml on this machine only.
# Keep this file out of whatever syncs repositories.yaml between machines.
base_dir: ` + yamlScalar(baseDir) + `

# Optional machine-specific overrides:
# pull_strategy: rebase
# aliases:              # merged with the shared aliases, local wins
#   nahime0: nahime
# suppress_skills_notice: true
`
	return writeNewFile(path, content)
}

func writeNewFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// yamlScalar encodes s as a YAML scalar, quoting it when needed (a bare ~, for
// example, would otherwise parse as null).
func yamlScalar(s string) string {
	out, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Sprintf("%q", s)
	}
	return strings.TrimSuffix(string(out), "\n")
}
