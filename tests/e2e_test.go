package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/illegalstudio/ggg/internal/testutil"
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
	if !strings.Contains(out, "#compdef ggg") {
		t.Fatalf("shell-init output missing zsh completion:\n%s", out)
	}
}

func TestCLIAddListValidateRemove(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	writeConfig(t, home, fmt.Sprintf(`
base_dir: %s
repos: []
`, filepath.ToSlash(baseDir)))

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

	writeConfig(t, home, `
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

	out, err = runGGG(t, home, "__complete", "cd", "pro")
	if err != nil {
		t.Fatalf("ggg cd completion failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "project") {
		t.Fatalf("completion output missing repo name:\n%s", out)
	}
}

func TestCLIExportCopiesConfig(t *testing.T) {
	home := testutil.SetupHome(t)
	writeConfig(t, home, `
base_dir: ~/Developer
repos:
  - url: git@github.com:acme/exported.git
    group: work
`)

	destDir := t.TempDir()
	out, err := runGGG(t, home, "export", destDir)
	if err != nil {
		t.Fatalf("ggg export failed: %v\n%s", err, out)
	}

	exportedPath := filepath.Join(destDir, "repositories.yaml")
	data, err := os.ReadFile(exportedPath)
	if err != nil {
		t.Fatalf("read exported config: %v", err)
	}
	if !strings.Contains(string(data), "git@github.com:acme/exported.git") {
		t.Fatalf("exported config missing repo:\n%s", data)
	}
}

func TestCLIGitFlowCloneStatusDiffStashCheckout(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, remoteOne := createRemoteRepo(t, home)
	_, remoteTwo := createRemoteRepo(t, home)

	writeConfig(t, home, fmt.Sprintf(`
base_dir: %s
repos:
  - url: %s
    path: team/demo
  - url: %s
    path: team/uncloned
`, filepath.ToSlash(baseDir), filepath.ToSlash(remoteOne), filepath.ToSlash(remoteTwo)))

	out, err := runGGG(t, home, "clone", "demo")
	if err != nil {
		t.Fatalf("ggg clone failed: %v\n%s", err, out)
	}

	clonedPath := filepath.Join(baseDir, "team", "demo")
	if _, err := os.Stat(clonedPath); err != nil {
		t.Fatalf("cloned repo missing: %v", err)
	}

	out, err = runGGG(t, home, "status")
	if err != nil {
		t.Fatalf("ggg status failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "(not cloned)") {
		t.Fatalf("status output missing not cloned repo:\n%s", out)
	}
	if !strings.Contains(out, "[main]") || !strings.Contains(out, "clean") {
		t.Fatalf("status output missing cloned repo state:\n%s", out)
	}

	testutil.WriteFile(t, clonedPath, "README.md", "# dirty\n")

	out, err = runGGG(t, home, "status", "demo")
	if err != nil {
		t.Fatalf("ggg status dirty failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "dirty") {
		t.Fatalf("status output missing dirty state:\n%s", out)
	}

	out, err = runGGG(t, home, "diff", "demo")
	if err != nil {
		t.Fatalf("ggg diff failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "README.md") {
		t.Fatalf("diff output missing changed file:\n%s", out)
	}

	out, err = runGGG(t, home, "stash", "demo")
	if err != nil {
		t.Fatalf("ggg stash failed: %v\n%s", err, out)
	}

	out, err = runGGG(t, home, "status", "demo")
	if err != nil {
		t.Fatalf("ggg status after stash failed: %v\n%s", err, out)
	}
	if strings.Contains(out, "dirty") || !strings.Contains(out, "clean") {
		t.Fatalf("status should be clean after stash:\n%s", out)
	}

	testutil.RunGit(t, clonedPath, home, "branch", "feature")
	out, err = runGGG(t, home, "__complete", "checkout", "fea")
	if err != nil {
		t.Fatalf("ggg checkout completion failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "feature") {
		t.Fatalf("checkout completion output missing branch:\n%s", out)
	}

	out, err = runGGG(t, home, "checkout", "feature", "demo")
	if err != nil {
		t.Fatalf("ggg checkout failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "feature") {
		t.Fatalf("checkout output missing branch:\n%s", out)
	}

	branch := testutil.RunGit(t, clonedPath, home, "rev-parse", "--abbrev-ref", "HEAD")
	if branch != "feature" {
		t.Fatalf("current branch = %q, want feature", branch)
	}
}

func TestCLIStatusWorktrees(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, remote := createRemoteRepo(t, home)

	writeConfig(t, home, fmt.Sprintf(`
base_dir: %s
repos:
  - url: %s
    path: team/wt-demo
`, filepath.ToSlash(baseDir), filepath.ToSlash(remote)))

	if out, err := runGGG(t, home, "clone", "wt-demo"); err != nil {
		t.Fatalf("ggg clone failed: %v\n%s", err, out)
	}

	clonedPath := filepath.Join(baseDir, "team", "wt-demo")
	wt1 := filepath.Join(t.TempDir(), "feature-wt")
	wt2 := filepath.Join(t.TempDir(), "bugfix-wt")
	testutil.RunGit(t, clonedPath, home, "worktree", "add", "-b", "feature", wt1)
	testutil.RunGit(t, clonedPath, home, "worktree", "add", "-b", "bugfix", wt2)

	// Make one of the worktrees dirty.
	testutil.WriteFile(t, wt1, "scratch.txt", "wip\n")

	// Default view: count badge appears, no per-worktree tree.
	out, err := runGGG(t, home, "status")
	if err != nil {
		t.Fatalf("ggg status failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "⎇2") {
		t.Fatalf("status output missing worktree badge ⎇2:\n%s", out)
	}
	if strings.Contains(out, "├─") || strings.Contains(out, "└─") {
		t.Fatalf("status (no --detailed) should not show worktree tree:\n%s", out)
	}

	// Detailed view: tree includes both worktrees with branch + status.
	out, err = runGGG(t, home, "status", "--detailed")
	if err != nil {
		t.Fatalf("ggg status --detailed failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "├─") || !strings.Contains(out, "└─") {
		t.Fatalf("status --detailed missing tree connectors:\n%s", out)
	}
	if !strings.Contains(out, "[feature]") || !strings.Contains(out, "[bugfix]") {
		t.Fatalf("status --detailed missing worktree branches:\n%s", out)
	}
	if !strings.Contains(out, "dirty") {
		t.Fatalf("status --detailed missing dirty worktree status:\n%s", out)
	}

	// JSON output exposes the worktrees array.
	out, err = runGGG(t, home, "--json", "status")
	if err != nil {
		t.Fatalf("ggg --json status failed: %v\n%s", err, out)
	}
	var payload struct {
		Repos []struct {
			Worktrees []struct {
				Branch string `json:"branch"`
				Dirty  bool   `json:"dirty"`
			} `json:"worktrees"`
		} `json:"repos"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("status --json invalid JSON: %v\n%s", err, out)
	}
	if len(payload.Repos) != 1 || len(payload.Repos[0].Worktrees) != 2 {
		t.Fatalf("expected 1 repo with 2 worktrees in JSON, got: %+v", payload)
	}
}

func TestCLIOutdatedAndPull(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	seedRepo, remote := createRemoteRepo(t, home)

	writeConfig(t, home, fmt.Sprintf(`
base_dir: %s
repos:
  - url: %s
    path: team/pull-demo
`, filepath.ToSlash(baseDir), filepath.ToSlash(remote)))

	out, err := runGGG(t, home, "clone", "pull-demo")
	if err != nil {
		t.Fatalf("ggg clone failed: %v\n%s", err, out)
	}

	testutil.CommitFile(t, home, seedRepo, "README.md", "# updated\n", "update readme")
	testutil.RunGit(t, seedRepo, home, "push", "origin", "main")

	out, err = runGGG(t, home, "outdated", "pull-demo")
	if err != nil {
		t.Fatalf("ggg outdated failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "1 commits behind") {
		t.Fatalf("outdated output missing behind count:\n%s", out)
	}

	out, err = runGGG(t, home, "pull", "pull-demo")
	if err != nil {
		t.Fatalf("ggg pull failed: %v\n%s", err, out)
	}

	clonedPath := filepath.Join(baseDir, "team", "pull-demo")
	if got := testutil.ReadFile(t, clonedPath, "README.md"); got != "# updated\n" {
		t.Fatalf("cloned README not updated after pull: %q", got)
	}

	out, err = runGGG(t, home, "outdated", "pull-demo")
	if err != nil {
		t.Fatalf("ggg outdated after pull failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "All repositories are up to date.") {
		t.Fatalf("unexpected outdated output after pull:\n%s", out)
	}
}

func TestCLIPush(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	seedRepo, remote := createRemoteRepo(t, home)

	writeConfig(t, home, fmt.Sprintf(`
base_dir: %s
repos:
  - url: %s
    path: team/push-demo
`, filepath.ToSlash(baseDir), filepath.ToSlash(remote)))

	out, err := runGGG(t, home, "clone", "push-demo")
	if err != nil {
		t.Fatalf("ggg clone failed: %v\n%s", err, out)
	}

	clonedPath := filepath.Join(baseDir, "team", "push-demo")
	testutil.CommitFile(t, home, clonedPath, "README.md", "# pushed\n", "push change")
	localHead := testutil.RunGit(t, clonedPath, home, "rev-parse", "HEAD")

	out, err = runGGG(t, home, "push", "push-demo")
	if err != nil {
		t.Fatalf("ggg push failed: %v\n%s", err, out)
	}

	testutil.RunGit(t, seedRepo, home, "fetch", "origin")
	remoteHead := testutil.RunGit(t, seedRepo, home, "rev-parse", "origin/main")
	if remoteHead != localHead {
		t.Fatalf("remote head = %q, want %q", remoteHead, localHead)
	}
}

func TestCLIConfigAndOpenUseEditor(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, remote := createRemoteRepo(t, home)
	writeConfig(t, home, fmt.Sprintf(`
base_dir: %s
repos:
  - url: %s
    path: team/editor-target
`, filepath.ToSlash(baseDir), filepath.ToSlash(remote)))

	out, err := runGGG(t, home, "clone", "editor-target")
	if err != nil {
		t.Fatalf("ggg clone failed: %v\n%s", err, out)
	}

	stubDir := t.TempDir()
	logPath := filepath.Join(stubDir, "editor.log")
	editorPath := testutil.WriteStubCommand(t, stubDir, "fake-editor", logPath, "")

	t.Setenv("EDITOR", editorPath)

	out, err = runGGG(t, home, "config")
	if err != nil {
		t.Fatalf("ggg config failed: %v\n%s", err, out)
	}

	out, err = runGGG(t, home, "open", "editor-target", editorPath)
	if err != nil {
		t.Fatalf("ggg open failed: %v\n%s", err, out)
	}

	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read editor log: %v", err)
	}

	log := string(logData)
	if !strings.Contains(log, testutil.ConfigPath(home)) {
		t.Fatalf("config command did not invoke editor with config path:\n%s", log)
	}
	if !strings.Contains(log, filepath.Join(baseDir, "team", "editor-target")) {
		t.Fatalf("open command did not invoke editor with repo path:\n%s", log)
	}
}

func TestCLIBrowseUsesBrowserOverride(t *testing.T) {
	home := testutil.SetupHome(t)
	writeConfig(t, home, `
base_dir: ~/Developer
repos:
  - url: git@github.com:acme/browse-demo.git
`)

	stubDir := t.TempDir()
	logPath := filepath.Join(stubDir, "browser.log")
	browserPath := testutil.WriteStubCommand(t, stubDir, "fake-browser", logPath, "")

	t.Setenv("GGG_TEST_BROWSER_CMD", browserPath)

	out, err := runGGG(t, home, "browse", "browse-demo")
	if err != nil {
		t.Fatalf("ggg browse failed: %v\n%s", err, out)
	}

	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read browser log: %v", err)
	}

	if got := strings.TrimSpace(string(logData)); got != "https://github.com/acme/browse-demo" {
		t.Fatalf("browser called with %q", got)
	}
}

func TestCLIDoctorReportsRemoteHealth(t *testing.T) {
	home := testutil.SetupHome(t)
	baseDir := filepath.Join(home, "Developer")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatal(err)
	}

	_, remote := createRemoteRepo(t, home)
	missing := filepath.Join(t.TempDir(), "missing.git")

	writeConfig(t, home, fmt.Sprintf(`
base_dir: %s
repos:
  - url: %s
    path: team/reachable
  - url: %s
    path: team/missing
`, filepath.ToSlash(baseDir), filepath.ToSlash(remote), filepath.ToSlash(missing)))

	out, err := runGGG(t, home, "doctor")
	if err != nil {
		t.Fatalf("ggg doctor failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "2 configured") {
		t.Fatalf("doctor output missing repo count:\n%s", out)
	}
	if !strings.Contains(out, "1 unreachable") {
		t.Fatalf("doctor output missing unreachable remote count:\n%s", out)
	}
	if !strings.Contains(out, missing) {
		t.Fatalf("doctor output missing unreachable remote details:\n%s", out)
	}
}

func TestCLIImportWithStubGH(t *testing.T) {
	home := testutil.SetupHome(t)
	writeConfig(t, home, `
base_dir: ~/Developer
repos: []
`)

	stubDir := t.TempDir()
	logPath := filepath.Join(stubDir, "gh.log")
	testutil.PrependPath(t, stubDir)
	testutil.WriteStubScript(t, stubDir, "gh", fmt.Sprintf(`
printf '%%s\n' "$*" >> %s
printf '%%s\n' '[{"full_name":"acme/alpha","ssh_url":"git@github.com:acme/alpha.git","clone_url":"https://github.com/acme/alpha.git","private":false},{"full_name":"acme/bravo","ssh_url":"git@github.com:acme/bravo.git","clone_url":"https://github.com/acme/bravo.git","private":true}]'
`, shellQuote(logPath)))

	t.Setenv("GGG_TEST_IMPORT_SELECTION", "all")

	out, err := runGGG(t, home, "import", "acme", "--http", "--group", "work")
	if err != nil {
		t.Fatalf("ggg import failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Added 2 repositories.") {
		t.Fatalf("unexpected import output:\n%s", out)
	}

	configData, err := os.ReadFile(testutil.ConfigPath(home))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	configText := string(configData)
	if !strings.Contains(configText, "https://github.com/acme/alpha.git") || !strings.Contains(configText, "https://github.com/acme/bravo.git") {
		t.Fatalf("imported repos missing from config:\n%s", configText)
	}
	if strings.Count(configText, "group: work") != 2 {
		t.Fatalf("imported group missing from config:\n%s", configText)
	}

	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read gh log: %v", err)
	}
	if !strings.Contains(string(logData), "/orgs/acme/repos") {
		t.Fatalf("gh stub did not receive expected API call:\n%s", logData)
	}
}

func TestCLIJSONStatusWithNoReposEmitsValidJSON(t *testing.T) {
	home := testutil.SetupHome(t)
	writeConfig(t, home, `
base_dir: ~/Developer
repos: []
`)

	out, err := runGGG(t, home, "--json", "status")
	if err != nil {
		t.Fatalf("ggg --json status failed: %v\n%s", err, out)
	}

	var payload struct {
		Repos []any `json:"repos"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("status did not emit valid JSON: %v\n%s", err, out)
	}
	if len(payload.Repos) != 0 {
		t.Fatalf("repos length = %d, want 0\n%s", len(payload.Repos), out)
	}
}

func TestCLIJSONUnsupportedCommandEmitsStructuredError(t *testing.T) {
	home := testutil.SetupHome(t)

	out, err := runGGG(t, home, "--json", "browse", "anything")
	if err == nil {
		t.Fatalf("ggg --json browse succeeded unexpectedly:\n%s", out)
	}

	var payload struct {
		Supported bool   `json:"supported"`
		Command   string `json:"command"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("browse did not emit valid JSON error: %v\n%s", err, out)
	}
	if payload.Supported || payload.Command != "browse" || !strings.Contains(payload.Error, "--json is not supported") {
		t.Fatalf("unexpected JSON error payload: %+v\n%s", payload, out)
	}
}

func TestCLIJSONImportRequiresAndImportsSingleRepo(t *testing.T) {
	home := testutil.SetupHome(t)
	writeConfig(t, home, `
base_dir: ~/Developer
repos: []
`)

	stubDir := t.TempDir()
	logPath := filepath.Join(stubDir, "gh.log")
	testutil.PrependPath(t, stubDir)
	testutil.WriteStubScript(t, stubDir, "gh", fmt.Sprintf(`
printf '%%s\n' "$*" >> %s
printf '%%s\n' '[{"full_name":"acme/alpha","ssh_url":"git@github.com:acme/alpha.git","clone_url":"https://github.com/acme/alpha.git","private":false},{"full_name":"acme/bravo","ssh_url":"git@github.com:acme/bravo.git","clone_url":"https://github.com/acme/bravo.git","private":true}]'
`, shellQuote(logPath)))

	out, err := runGGG(t, home, "--json", "import", "acme")
	if err == nil {
		t.Fatalf("ggg --json import without repo succeeded unexpectedly:\n%s", out)
	}
	var errorPayload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(out), &errorPayload); err != nil {
		t.Fatalf("import error was not valid JSON: %v\n%s", err, out)
	}
	if !strings.Contains(errorPayload.Error, "repository argument") {
		t.Fatalf("unexpected import error payload: %+v\n%s", errorPayload, out)
	}

	out, err = runGGG(t, home, "--json", "import", "acme", "alpha", "--http", "--group", "work")
	if err != nil {
		t.Fatalf("ggg --json import single repo failed: %v\n%s", err, out)
	}

	var payload struct {
		Added   []string `json:"added"`
		Skipped []string `json:"skipped"`
	}
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		t.Fatalf("import did not emit valid JSON: %v\n%s", err, out)
	}
	if len(payload.Added) != 1 || payload.Added[0] != "https://github.com/acme/alpha.git" || len(payload.Skipped) != 0 {
		t.Fatalf("unexpected import JSON payload: %+v\n%s", payload, out)
	}

	configData, err := os.ReadFile(testutil.ConfigPath(home))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	configText := string(configData)
	if !strings.Contains(configText, "https://github.com/acme/alpha.git") {
		t.Fatalf("alpha missing from config:\n%s", configText)
	}
	if strings.Contains(configText, "https://github.com/acme/bravo.git") {
		t.Fatalf("json import imported more than the requested repo:\n%s", configText)
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

		cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/ggg")
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

func createRemoteRepo(t *testing.T, home string) (string, string) {
	t.Helper()

	seed := testutil.InitGitRepoInHome(t, home)
	remote := testutil.CreateBareRemote(t, home, seed)
	testutil.RunGit(t, seed, home, "remote", "add", "origin", remote)
	testutil.RunGit(t, seed, home, "push", "-u", "origin", "main")
	return seed, remote
}

func writeConfig(t *testing.T, home, content string) {
	t.Helper()
	testutil.WriteConfig(t, home, content)
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

type buildFailure struct {
	err    error
	output string
}

func (b *buildFailure) Error() string {
	return b.err.Error() + "\n" + b.output
}
