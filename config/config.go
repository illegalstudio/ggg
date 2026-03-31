package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Repo struct {
	URL  string `mapstructure:"url" yaml:"url"`
	Path string `mapstructure:"path" yaml:"path,omitempty"`
}

type Config struct {
	BaseDir string `mapstructure:"base_dir" yaml:"base_dir"`
	Repos   []Repo `mapstructure:"repos" yaml:"repos"`
}

func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ggg", "repositories.yaml")
}

func Load() (*Config, error) {
	viper.SetConfigName("repositories")
	viper.SetConfigType("yaml")

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	viper.AddConfigPath(filepath.Join(home, ".config", "ggg"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
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

repos:
  - url: git@github.com:user/repo.git
  # path: custom/path  # optional, derived from URL if omitted
`
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}
