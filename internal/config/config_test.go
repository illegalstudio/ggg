package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/illegalstudio/ggg/internal/testutil"

	"gopkg.in/yaml.v3"
)

func TestWriteDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ggg", "repositories.yaml")

	if err := WriteDefault(path, ""); err != nil {
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
	if !strings.Contains(content, "suppress_skills_notice") {
		t.Error("config should document suppress_skills_notice")
	}
}

func TestSuppressSkillsNotice(t *testing.T) {
	home := testutil.SetupHome(t)

	if got, err := SuppressSkillsNotice(); err != nil || got {
		t.Fatalf("no config file: got (%v, %v), want (false, nil)", got, err)
	}

	testutil.WriteConfig(t, home, `
repos: []
`)
	if got, err := SuppressSkillsNotice(); err != nil || got {
		t.Fatalf("key unset: got (%v, %v), want (false, nil)", got, err)
	}

	testutil.WriteConfig(t, home, `
suppress_skills_notice: true
repos: []
`)
	if got, err := SuppressSkillsNotice(); err != nil || !got {
		t.Fatalf("key true: got (%v, %v), want (true, nil)", got, err)
	}
}

func TestWriteDefault_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "deep", "repositories.yaml")

	if err := WriteDefault(path, ""); err != nil {
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
			{URL: "git@github.com:user/repo3.git", Groups: []string{"work"}},
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
	if len(loaded.Repos[2].Groups) != 1 || loaded.Repos[2].Groups[0] != "work" {
		t.Errorf("Repos[2].Groups = %v, want [work]", loaded.Repos[2].Groups)
	}
}

func TestConfigPath(t *testing.T) {
	home := testutil.SetupHome(t)

	path := ConfigPath()
	want := filepath.Join(home, ".config", "ggg", "repositories.yaml")
	if path != want {
		t.Errorf("ConfigPath = %q, want %q", path, want)
	}
}

func TestLoad_ExpandsBaseDir(t *testing.T) {
	home := testutil.SetupHome(t)
	testutil.WriteConfig(t, home, `
base_dir: ~/Projects
pull_strategy: rebase
repos:
  - url: git@github.com:user/repo.git
`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.BaseDir != filepath.Join(home, "Projects") {
		t.Errorf("BaseDir = %q, want %q", cfg.BaseDir, filepath.Join(home, "Projects"))
	}
	if cfg.PullStrategy != PullRebase {
		t.Errorf("PullStrategy = %q, want %q", cfg.PullStrategy, PullRebase)
	}
	if len(cfg.Repos) != 1 || cfg.Repos[0].URL != "git@github.com:user/repo.git" {
		t.Fatalf("unexpected repos: %+v", cfg.Repos)
	}
}

func TestLoad_DefaultBaseDir_WhenMissing(t *testing.T) {
	home := testutil.SetupHome(t)
	testutil.WriteConfig(t, home, `
repos:
  - url: git@github.com:user/repo.git
`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(home, "Developer")
	if cfg.BaseDir != want {
		t.Errorf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}
}

func TestSaveAndLoad_WithHomeConfigPath(t *testing.T) {
	home := testutil.SetupHome(t)
	cfg := &Config{
		BaseDir:      "~/Dev",
		PullStrategy: PullFFOnly,
		Repos: []Repo{
			{URL: "git@github.com:a/b.git", Groups: []string{"work"}},
			{URL: "https://github.com/c/d.git", Path: "custom/path", PullStrategy: PullRebase},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	raw, err := LoadRaw()
	if err != nil {
		t.Fatal(err)
	}
	if raw.BaseDir != "~/Dev" {
		t.Errorf("raw BaseDir = %q, want %q", raw.BaseDir, "~/Dev")
	}
	if raw.PullStrategy != PullFFOnly {
		t.Errorf("raw PullStrategy = %q, want %q", raw.PullStrategy, PullFFOnly)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.BaseDir != filepath.Join(home, "Dev") {
		t.Errorf("loaded BaseDir = %q, want %q", loaded.BaseDir, filepath.Join(home, "Dev"))
	}
	if len(loaded.Repos) != 2 {
		t.Fatalf("got %d repos, want 2", len(loaded.Repos))
	}
}

func TestSave_MarshalRoundtrip(t *testing.T) {
	cfg := &Config{
		BaseDir: "~/Dev",
		Repos: []Repo{
			{URL: "git@github.com:a/b.git"},
			{URL: "git@github.com:c/d.git", Path: "custom", Groups: []string{"personal"}},
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
	if len(loaded.Repos[1].Groups) != 1 || loaded.Repos[1].Groups[0] != "personal" {
		t.Errorf("Repos[1].Groups = %v, want [personal]", loaded.Repos[1].Groups)
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

func TestSave_MultipleGroupsRoundtrip(t *testing.T) {
	cfg := &Config{
		BaseDir: "~/Dev",
		Repos: []Repo{
			{URL: "git@github.com:a/b.git", Groups: []string{"work", "oss"}},
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

	if len(loaded.Repos[0].Groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(loaded.Repos[0].Groups))
	}
	if loaded.Repos[0].Groups[0] != "work" || loaded.Repos[0].Groups[1] != "oss" {
		t.Errorf("Groups = %v, want [work oss]", loaded.Repos[0].Groups)
	}
}

func TestSaveAndLoad_Aliases(t *testing.T) {
	testutil.SetupHome(t)
	cfg := &Config{
		BaseDir: "~/Developer",
		Aliases: map[string]string{"nahime0": "nahime"},
		Repos: []Repo{
			{URL: "git@github.com:nahime0/repo.git"},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.Aliases["nahime0"] != "nahime" {
		t.Errorf("Aliases[nahime0] = %q, want %q", loaded.Aliases["nahime0"], "nahime")
	}
}

func TestWriteDefault_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ggg", "repositories.yaml")

	if err := WriteDefault(path, ""); err != nil {
		t.Fatal(err)
	}
	data1, _ := os.ReadFile(path)

	if err := WriteDefault(path, ""); err != nil {
		t.Fatal(err)
	}
	data2, _ := os.ReadFile(path)

	if string(data1) != string(data2) {
		t.Error("WriteDefault should be idempotent")
	}
}

func TestWriteDefault_CustomBaseDir(t *testing.T) {
	home := testutil.SetupHome(t)
	if err := WriteDefault(ConfigPath(), "~/code"); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "code"); cfg.BaseDir != want {
		t.Errorf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}
}

func TestExpandPath(t *testing.T) {
	home := filepath.FromSlash("/home/tester")
	t.Setenv("GGG_TEST_ROOT", filepath.FromSlash("/data"))
	t.Setenv("GGG_TEST_EMPTY", "")

	tests := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{name: "bare tilde", in: "~", want: home},
		{name: "tilde prefix", in: "~/Developer", want: filepath.Join(home, "Developer")},
		{name: "absolute", in: "/srv/dev", want: "/srv/dev"},
		{name: "dollar var", in: "$GGG_TEST_ROOT/dev", want: filepath.FromSlash("/data") + "/dev"},
		{name: "braced var", in: "${GGG_TEST_ROOT}/dev", want: filepath.FromSlash("/data") + "/dev"},
		{name: "unset var", in: "$GGG_TEST_UNSET/dev", wantErr: "GGG_TEST_UNSET"},
		{name: "empty var", in: "${GGG_TEST_EMPTY}/dev", wantErr: "GGG_TEST_EMPTY"},
		{name: "tilde not first", in: "/x/~/y", want: "/x/~/y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.in, home)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ExpandPath(%q) error = %v, want mention of %q", tt.in, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ExpandPath(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestLoad_BaseDirPrecedence(t *testing.T) {
	tests := []struct {
		name       string
		shared     string
		local      string
		env        string
		wantRel    string
		wantSource BaseDirSource
	}{
		{name: "default", shared: "repos: []\n", wantRel: "Developer", wantSource: BaseDirFromDefault},
		{name: "shared", shared: "base_dir: ~/Shared\nrepos: []\n", wantRel: "Shared", wantSource: BaseDirFromConfig},
		{name: "local over shared", shared: "base_dir: ~/Shared\nrepos: []\n", local: "base_dir: ~/Local\n", wantRel: "Local", wantSource: BaseDirFromLocal},
		{name: "local without base_dir keeps shared", shared: "base_dir: ~/Shared\nrepos: []\n", local: "pull_strategy: rebase\n", wantRel: "Shared", wantSource: BaseDirFromConfig},
		{name: "env over local", shared: "base_dir: ~/Shared\nrepos: []\n", local: "base_dir: ~/Local\n", env: "~/Env", wantRel: "Env", wantSource: BaseDirFromEnv},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := testutil.SetupHome(t)
			testutil.WriteConfig(t, home, tt.shared)
			if tt.local != "" {
				testutil.WriteLocalConfig(t, home, tt.local)
			}
			t.Setenv(BaseDirEnv, tt.env)

			cfg, err := Load()
			if err != nil {
				t.Fatal(err)
			}
			if want := filepath.Join(home, tt.wantRel); cfg.BaseDir != want {
				t.Errorf("BaseDir = %q, want %q", cfg.BaseDir, want)
			}
			if cfg.BaseDirSource != tt.wantSource {
				t.Errorf("BaseDirSource = %q, want %q", cfg.BaseDirSource, tt.wantSource)
			}
		})
	}
}

func TestLoad_BaseDirWithUnsetVariableFails(t *testing.T) {
	home := testutil.SetupHome(t)
	testutil.WriteConfig(t, home, `
base_dir: $GGG_TEST_UNSET_ROOT/dev
repos: []
`)

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "GGG_TEST_UNSET_ROOT") {
		t.Fatalf("Load() error = %v, want mention of the unset variable", err)
	}
}

func TestLoad_BaseDirExpandsVariable(t *testing.T) {
	home := testutil.SetupHome(t)
	root := t.TempDir()
	t.Setenv("GGG_TEST_DEV_ROOT", root)
	testutil.WriteConfig(t, home, `
base_dir: ${GGG_TEST_DEV_ROOT}/src
repos: []
`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if want := root + "/src"; cfg.BaseDir != want {
		t.Errorf("BaseDir = %q, want %q", cfg.BaseDir, want)
	}
}

func TestLoad_LocalOverridesOtherSettings(t *testing.T) {
	home := testutil.SetupHome(t)
	testutil.WriteConfig(t, home, `
pull_strategy: merge
aliases:
  shared: one
  both: shared-value
repos:
  - url: git@github.com:user/repo.git
`)
	testutil.WriteLocalConfig(t, home, `
pull_strategy: ff-only
aliases:
  both: local-value
  local: two
suppress_skills_notice: true
`)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PullStrategy != PullFFOnly {
		t.Errorf("PullStrategy = %q, want %q", cfg.PullStrategy, PullFFOnly)
	}
	wantAliases := map[string]string{"shared": "one", "both": "local-value", "local": "two"}
	for k, v := range wantAliases {
		if cfg.Aliases[k] != v {
			t.Errorf("Aliases[%q] = %q, want %q", k, cfg.Aliases[k], v)
		}
	}
	if !cfg.SuppressSkillsNotice {
		t.Error("SuppressSkillsNotice should come from the local file")
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(cfg.Repos))
	}
}

func TestLoadLocal(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{name: "comments only", content: "# nothing here\n"},
		{name: "empty", content: ""},
		{name: "repos rejected", content: "repos:\n  - url: git@github.com:a/b.git\n", wantErr: "cannot define repos"},
		{name: "unknown key rejected", content: "basedir: ~/typo\n", wantErr: `unknown key "basedir"`},
		{name: "invalid pull strategy", content: "pull_strategy: squash\n", wantErr: "invalid pull_strategy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			home := testutil.SetupHome(t)
			testutil.WriteLocalConfig(t, home, tt.content)

			local, err := LoadLocal()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("LoadLocal() error = %v, want mention of %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadLocal() unexpected error: %v", err)
			}
			if local == nil {
				t.Fatal("LoadLocal() = nil, want an empty LocalConfig for an existing file")
			}
		})
	}
}

func TestLoadLocal_MissingFile(t *testing.T) {
	testutil.SetupHome(t)

	local, err := LoadLocal()
	if err != nil || local != nil {
		t.Fatalf("LoadLocal() = (%v, %v), want (nil, nil)", local, err)
	}
}

func TestLoad_InvalidLocalFails(t *testing.T) {
	home := testutil.SetupHome(t)
	testutil.WriteConfig(t, home, "repos: []\n")
	testutil.WriteLocalConfig(t, home, "repos: []\n")

	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "cannot define repos") {
		t.Fatalf("Load() error = %v, want local repos rejection", err)
	}
}

func TestWriteLocal_Roundtrip(t *testing.T) {
	home := testutil.SetupHome(t)
	testutil.WriteConfig(t, home, "base_dir: /shared\nrepos: []\n")

	// A bare ~ must be quoted, or YAML would read it as null.
	if err := WriteLocal(LocalConfigPath(), "~"); err != nil {
		t.Fatal(err)
	}

	local, err := LoadLocal()
	if err != nil {
		t.Fatal(err)
	}
	if local.BaseDir != "~" {
		t.Errorf("local BaseDir = %q, want %q", local.BaseDir, "~")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BaseDir != home || cfg.BaseDirSource != BaseDirFromLocal {
		t.Errorf("Load() BaseDir = %q (%s), want %q (local)", cfg.BaseDir, cfg.BaseDirSource, home)
	}
}

func TestSave_DoesNotLeakLocalValues(t *testing.T) {
	home := testutil.SetupHome(t)
	testutil.WriteConfig(t, home, `
base_dir: ~/Shared
repos:
  - url: git@github.com:user/repo.git
`)
	testutil.WriteLocalConfig(t, home, `
base_dir: /machine/only
pull_strategy: rebase
`)

	raw, err := LoadRaw()
	if err != nil {
		t.Fatal(err)
	}
	raw.Repos = append(raw.Repos, Repo{URL: "git@github.com:user/other.git"})
	if err := Save(raw); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "/machine/only") || strings.Contains(content, "rebase") {
		t.Fatalf("shared config picked up local values:\n%s", content)
	}
	if !strings.Contains(content, "base_dir: ~/Shared") {
		t.Fatalf("shared base_dir changed:\n%s", content)
	}
}

func TestSuppressSkillsNotice_LocalOverride(t *testing.T) {
	home := testutil.SetupHome(t)

	// Local file alone, without a shared config.
	testutil.WriteLocalConfig(t, home, "suppress_skills_notice: true\n")
	if got, err := SuppressSkillsNotice(); err != nil || !got {
		t.Fatalf("local only: got (%v, %v), want (true, nil)", got, err)
	}

	// Local false beats shared true.
	testutil.WriteConfig(t, home, "suppress_skills_notice: true\nrepos: []\n")
	testutil.WriteLocalConfig(t, home, "suppress_skills_notice: false\n")
	if got, err := SuppressSkillsNotice(); err != nil || got {
		t.Fatalf("local false: got (%v, %v), want (false, nil)", got, err)
	}

	// Local without the key keeps the shared value.
	testutil.WriteLocalConfig(t, home, "base_dir: /x\n")
	if got, err := SuppressSkillsNotice(); err != nil || !got {
		t.Fatalf("local unset: got (%v, %v), want (true, nil)", got, err)
	}
}
