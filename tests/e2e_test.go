package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"go-git-get/internal/testutil"
)

var (
	buildBaseEnv = os.Environ()
	buildOnce    sync.Once
	binaryPath   string
	buildErr     error
)

func TestCLIInitAndShellInit(t *testing.T) {
	home := testutil.SetupHome(t)

	out, err := runGGG(t, home, "init")
	if err != nil {
		t.Fatalf("ggg init failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Config file created") {
		t.Fatalf("unexpected init output:\n%s", out)
	}

	if _, err := os.Stat(testutil.ConfigPath(home)); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	out, err = runGGG(t, home, "shell-init", "zsh")
	if err != nil {
		t.Fatalf("ggg shell-init failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "gcd()") {
		t.Fatalf("unexpected shell-init output:\n%s", out)
	}
}

func TestCLIAddListValidateRemove(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	testutil.WriteConfig(t, home, `
base_dir: ~/Developer
repos: []
`)

	out, err := runGGG(t, home, "add", "git@github.com:acme/one.git", "--group", "work")
	if err != nil {
		t.Fatalf("ggg add one failed: %v\n%s", err, out)
	}

	out, err = runGGG(t, home, "add", "git@github.com:acme/two.git", "--path", "custom/two")
	if err != nil {
		t.Fatalf("ggg add two failed: %v\n%s", err, out)
	}

	out, err = runGGG(t, home, "list")
	if err != nil {
		t.Fatalf("ggg list failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "git@github.com:acme/one.git") || !strings.Contains(out, "git@github.com:acme/two.git") {
		t.Fatalf("list output missing repos:\n%s", out)
	}
	if !strings.Contains(out, filepath.Join(baseDir, "custom", "two")) {
		t.Fatalf("list output missing custom path:\n%s", out)
	}

	out, err = runGGG(t, home, "validate")
	if err != nil {
		t.Fatalf("ggg validate failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Configuration is valid") {
		t.Fatalf("unexpected validate output:\n%s", out)
	}

	out, err = runGGG(t, home, "remove", "two")
	if err != nil {
		t.Fatalf("ggg remove failed: %v\n%s", err, out)
	}

	out, err = runGGG(t, home, "list")
	if err != nil {
		t.Fatalf("ggg list after remove failed: %v\n%s", err, out)
	}
	if strings.Contains(out, "git@github.com:acme/two.git") {
		t.Fatalf("repo still present after remove:\n%s", out)
	}
}

func TestCLICDPrintsResolvedPath(t *testing.T) {
	home := testutil.SetupHome(t)
	repoDir := filepath.Join(home, "Developer", "acme", "project")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	testutil.WriteConfig(t, home, `
base_dir: ~/Developer
repos:
  - url: git@github.com:acme/project.git
`)

	out, err := runGGG(t, home, "cd", "project")
	if err != nil {
		t.Fatalf("ggg cd failed: %v\n%s", err, out)
	}
	if strings.TrimSpace(out) != repoDir {
		t.Fatalf("cd output = %q, want %q", strings.TrimSpace(out), repoDir)
	}
}

func runGGG(t *testing.T, home string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(buildBinary(t), args...)
	cmd.Dir = moduleRoot(t)
	cmd.Env = testutil.ChildEnv(t, home)

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func buildBinary(t *testing.T) string {
	t.Helper()

	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "ggg-e2e-*")
		if err != nil {
			buildErr = err
			return
		}

		binaryPath = filepath.Join(dir, "ggg")
		if runtime.GOOS == "windows" {
			binaryPath += ".exe"
		}

		cmd := exec.Command("go", "build", "-o", binaryPath, ".")
		cmd.Dir = moduleRoot(t)
		cmd.Env = append([]string{}, buildBaseEnv...)
		cmd.Env = append(cmd.Env, "GOCACHE="+filepath.Join(dir, "gocache"))
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = &buildFailure{err: err, output: string(out)}
		}
	})

	if buildErr != nil {
		t.Fatal(buildErr)
	}

	return binaryPath
}

func moduleRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine current file path")
	}

	return filepath.Dir(filepath.Dir(file))
}

type buildFailure struct {
	err    error
	output string
}

func (b *buildFailure) Error() string {
	return b.err.Error() + "\n" + b.output
}
