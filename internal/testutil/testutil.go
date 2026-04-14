package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	return InitGitRepoInHome(t, home)
}

// InitGitRepoInHome creates a temporary git repository with one commit while
// using the provided hermetic HOME for git subprocesses.
func InitGitRepoInHome(t *testing.T, home string) string {
	t.Helper()

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

// CreateBareRemote clones an existing repo as a bare remote and returns its path.
func CreateBareRemote(t *testing.T, home, source string) string {
	t.Helper()

	remote := filepath.Join(t.TempDir(), "origin.git")
	RunGit(t, ".", home, "clone", "--bare", source, remote)
	return remote
}

// CloneLocalRepo clones a local repository to the given destination path.
func CloneLocalRepo(t *testing.T, home, source, dest string) {
	t.Helper()
	RunGit(t, ".", home, "clone", source, dest)
}

// CommitFile writes or updates a file, stages it, and creates a commit.
func CommitFile(t *testing.T, home, repoPath, fileName, content, message string) {
	t.Helper()

	fullPath := filepath.Join(repoPath, fileName)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("create file dir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	RunGit(t, repoPath, home, "add", ".")
	RunGit(t, repoPath, home, "commit", "-m", message)
}

// WriteFile updates a file in a repo without committing.
func WriteFile(t *testing.T, repoPath, fileName, content string) {
	t.Helper()

	fullPath := filepath.Join(repoPath, fileName)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("create file dir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

// ReadFile returns the contents of a file relative to repoPath.
func ReadFile(t *testing.T, repoPath, fileName string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(repoPath, fileName))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	return string(data)
}

// PrependPath prepends a directory to PATH for the current test process.
func PrependPath(t *testing.T, dir string) {
	t.Helper()

	current := os.Getenv("PATH")
	if current == "" {
		t.Setenv("PATH", dir)
		return
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+current)
}

// WriteStubCommand writes a small executable that records its invocation.
// The template receives:
//
//	{log}    absolute log file path
//	{output} literal stdout payload
func WriteStubCommand(t *testing.T, dir, name, logPath, output string) string {
	t.Helper()

	var path string
	var content string
	if runtime.GOOS == "windows" {
		path = filepath.Join(dir, name+".cmd")
		content = fmt.Sprintf("@echo off\r\nsetlocal enabledelayedexpansion\r\necho %%*>> %s\r\n", windowsBatchPath(logPath))
		if output != "" {
			content += fmt.Sprintf("echo %s\r\n", output)
		}
	} else {
		path = filepath.Join(dir, name)
		content = "#!/bin/sh\n" +
			fmt.Sprintf("printf '%%s\\n' \"$*\" >> %s\n", shellQuote(logPath))
		if output != "" {
			content += fmt.Sprintf("printf '%%s\\n' %s\n", shellQuote(output))
		}
	}

	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("write stub command: %v", err)
	}

	return path
}

// WriteStubScript writes a custom executable script to a directory and returns its path.
func WriteStubScript(t *testing.T, dir, name, body string) string {
	t.Helper()

	var path string
	if runtime.GOOS == "windows" {
		path = filepath.Join(dir, name+".cmd")
	} else {
		path = filepath.Join(dir, name)
		body = "#!/bin/sh\n" + body
	}

	if err := os.WriteFile(path, []byte(body), 0755); err != nil {
		t.Fatalf("write stub script: %v", err)
	}
	return path
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

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func windowsBatchPath(path string) string {
	return strings.ReplaceAll(path, "/", `\`)
}
