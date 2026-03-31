package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWriteDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ggg", "repositories.yaml")

	if err := WriteDefault(path); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if len(content) == 0 {
		t.Error("config file should not be empty")
	}
	if !strings.Contains(content, "base_dir") {
		t.Error("config should contain base_dir")
	}
	if !strings.Contains(content, "repos") {
		t.Error("config should contain repos")
	}
}

func TestWriteDefault_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "repositories.yaml")

	if err := WriteDefault(path); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("config file should exist")
	}
}

func TestSaveAndLoadRaw(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".config", "ggg")
	os.MkdirAll(configDir, 0755)
	path := filepath.Join(configDir, "repositories.yaml")

	cfg := &Config{
		BaseDir: "~/Projects",
		Repos: []Repo{
			{URL: "git@github.com:user/repo1.git"},
			{URL: "git@github.com:user/repo2.git", Path: "custom/path"},
			{URL: "git@github.com:user/repo3.git", Group: "work"},
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Read it back directly (can't use LoadRaw since it uses ConfigPath)
	readData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var loaded Config
	if err := yaml.Unmarshal(readData, &loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.BaseDir != "~/Projects" {
		t.Errorf("BaseDir = %q, want %q", loaded.BaseDir, "~/Projects")
	}
	if len(loaded.Repos) != 3 {
		t.Fatalf("got %d repos, want 3", len(loaded.Repos))
	}
	if loaded.Repos[1].Path != "custom/path" {
		t.Errorf("Repos[1].Path = %q, want %q", loaded.Repos[1].Path, "custom/path")
	}
	if loaded.Repos[2].Group != "work" {
		t.Errorf("Repos[2].Group = %q, want %q", loaded.Repos[2].Group, "work")
	}
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	if !strings.Contains(path, ".config") {
		t.Errorf("ConfigPath should contain .config, got %q", path)
	}
	if !strings.HasSuffix(path, "ggg/repositories.yaml") {
		t.Errorf("ConfigPath should end with ggg/repositories.yaml, got %q", path)
	}
}

func TestSave_MarshalRoundtrip(t *testing.T) {
	cfg := &Config{
		BaseDir: "~/Dev",
		Repos: []Repo{
			{URL: "git@github.com:a/b.git"},
			{URL: "git@github.com:c/d.git", Path: "custom", Group: "personal"},
		},
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var loaded Config
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}

	if len(loaded.Repos) != 2 {
		t.Fatalf("got %d repos, want 2", len(loaded.Repos))
	}
	// Path with omitempty: repo without path should not have path field
	if loaded.Repos[0].Path != "" {
		t.Errorf("Repos[0].Path should be empty, got %q", loaded.Repos[0].Path)
	}
	if loaded.Repos[1].Group != "personal" {
		t.Errorf("Repos[1].Group = %q, want %q", loaded.Repos[1].Group, "personal")
	}
}

func TestResolvePullStrategy_Default(t *testing.T) {
	cfg := &Config{}
	r := Repo{URL: "git@github.com:user/repo.git"}
	if got := cfg.ResolvePullStrategy(r); got != PullMerge {
		t.Errorf("expected %q, got %q", PullMerge, got)
	}
}

func TestResolvePullStrategy_Global(t *testing.T) {
	cfg := &Config{PullStrategy: PullRebase}
	r := Repo{URL: "git@github.com:user/repo.git"}
	if got := cfg.ResolvePullStrategy(r); got != PullRebase {
		t.Errorf("expected %q, got %q", PullRebase, got)
	}
}

func TestResolvePullStrategy_RepoOverridesGlobal(t *testing.T) {
	cfg := &Config{PullStrategy: PullRebase}
	r := Repo{URL: "git@github.com:user/repo.git", PullStrategy: PullFFOnly}
	if got := cfg.ResolvePullStrategy(r); got != PullFFOnly {
		t.Errorf("expected %q, got %q", PullFFOnly, got)
	}
}

func TestResolvePullStrategy_RepoWithoutGlobal(t *testing.T) {
	cfg := &Config{}
	r := Repo{URL: "git@github.com:user/repo.git", PullStrategy: PullRebase}
	if got := cfg.ResolvePullStrategy(r); got != PullRebase {
		t.Errorf("expected %q, got %q", PullRebase, got)
	}
}

func TestWriteDefault_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ggg", "repositories.yaml")

	if err := WriteDefault(path); err != nil {
		t.Fatal(err)
	}
	data1, _ := os.ReadFile(path)

	if err := WriteDefault(path); err != nil {
		t.Fatal(err)
	}
	data2, _ := os.ReadFile(path)

	if string(data1) != string(data2) {
		t.Error("WriteDefault should be idempotent")
	}
}
