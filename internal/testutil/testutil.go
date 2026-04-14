package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testAuthorName  = "GGG Test"
	testAuthorEmail = "ggg@example.com"
)

// SetupHome points HOME-related environment variables at a temporary directory
// and disables global/system git config for the current test process.
func SetupHome(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	ensureHomeLayout(t, home)

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, ".gitconfig"))
	t.Setenv("GIT_TERMINAL_PROMPT", "0")

	return home
}

// ChildEnv returns a hermetic environment suitable for child processes that use
// git, config files, or os.UserHomeDir.
func ChildEnv(t *testing.T, home string) []string {
	t.Helper()
	ensureHomeLayout(t, home)

	return append(os.Environ(),
		"HOME="+home,
		"USERPROFILE="+home,
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL="+filepath.Join(home, ".gitconfig"),
		"GIT_TERMINAL_PROMPT=0",
	)
}

// ConfigPath returns the expected GGG config path for a specific test home.
func ConfigPath(home string) string {
	return filepath.Join(home, ".config", "ggg", "repositories.yaml")
}

// WriteConfig writes config content to the standard GGG config location inside
// the provided test home.
func WriteConfig(t *testing.T, home, content string) string {
	t.Helper()

	path := ConfigPath(home)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimLeft(content, "\n")), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

// RunGit executes git with a hermetic environment and conservative config that
// disables commit signing and hooks.
func RunGit(t *testing.T, dir, home string, args ...string) string {
	t.Helper()

	gitArgs := []string{
		"-c", "commit.gpgsign=false",
		"-c", "tag.gpgsign=false",
		"-c", "core.hooksPath=" + filepath.Join(home, ".githooks"),
	}
	gitArgs = append(gitArgs, args...)

	cmd := exec.Command("git", gitArgs...)
	cmd.Dir = dir
	cmd.Env = append(ChildEnv(t, home),
		"GIT_AUTHOR_NAME="+testAuthorName,
		"GIT_AUTHOR_EMAIL="+testAuthorEmail,
		"GIT_COMMITTER_NAME="+testAuthorName,
		"GIT_COMMITTER_EMAIL="+testAuthorEmail,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}

	return strings.TrimSpace(string(out))
}

// InitGitRepo creates a temporary git repository with one commit.
func InitGitRepo(t *testing.T) string {
	t.Helper()

	home := SetupHome(t)
	dir := t.TempDir()

	RunGit(t, dir, home, "init", "-b", "main")
	RunGit(t, dir, home, "config", "user.email", testAuthorEmail)
	RunGit(t, dir, home, "config", "user.name", testAuthorName)

	readmePath := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# test\n"), 0644); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	RunGit(t, dir, home, "add", ".")
	RunGit(t, dir, home, "commit", "-m", "initial commit")

	return dir
}

func ensureHomeLayout(t *testing.T, home string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(home, ".config", "ggg"), 0755); err != nil {
		t.Fatalf("create xdg config dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(home, ".githooks"), 0755); err != nil {
		t.Fatalf("create hooks dir: %v", err)
	}

	gitConfigPath := filepath.Join(home, ".gitconfig")
	if _, err := os.Stat(gitConfigPath); os.IsNotExist(err) {
		if err := os.WriteFile(gitConfigPath, []byte{}, 0644); err != nil {
			t.Fatalf("create git config: %v", err)
		}
	}
}
