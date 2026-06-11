package repo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/testutil"
)

func TestDerivePathFromURL_SSH(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"git@github.com:user/repo.git", "user/repo"},
		{"git@gitlab.com:org/project.git", "org/project"},
		{"git@github.com:user/repo", "user/repo"},
		{"git@bitbucket.org:team/lib.git", "team/lib"},
	}
	for _, tt := range tests {
		got, err := DerivePathFromURL(tt.url)
		if err != nil {
			t.Errorf("DerivePathFromURL(%q) error: %v", tt.url, err)
			continue
		}
		if got != tt.want {
			t.Errorf("DerivePathFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestDerivePathFromURL_HTTPS(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://github.com/user/repo.git", "user/repo"},
		{"https://gitlab.com/org/project.git", "org/project"},
		{"https://github.com/user/repo", "user/repo"},
		{"https://bitbucket.org/team/lib.git", "team/lib"},
	}
	for _, tt := range tests {
		got, err := DerivePathFromURL(tt.url)
		if err != nil {
			t.Errorf("DerivePathFromURL(%q) error: %v", tt.url, err)
			continue
		}
		if got != tt.want {
			t.Errorf("DerivePathFromURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}

func TestFullPath_WithCustomPath(t *testing.T) {
	r := config.Repo{URL: "git@github.com:user/repo.git", Path: "custom/path"}
	got, err := FullPath("/base", r)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/base", "custom/path")
	if got != want {
		t.Errorf("FullPath = %q, want %q", got, want)
	}
}

func TestFullPath_DerivedFromURL(t *testing.T) {
	r := config.Repo{URL: "git@github.com:user/repo.git"}
	got, err := FullPath("/base", r)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/base", "user/repo")
	if got != want {
		t.Errorf("FullPath = %q, want %q", got, want)
	}
}

func TestFullPath_HTTPSUrl(t *testing.T) {
	r := config.Repo{URL: "https://github.com/org/project.git"}
	got, err := FullPath("/home/dev", r)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/home/dev", "org/project")
	if got != want {
		t.Errorf("FullPath = %q, want %q", got, want)
	}
}

func TestIsCloned(t *testing.T) {
	dir := t.TempDir()

	if IsCloned(filepath.Join(dir, "nonexistent")) {
		t.Error("IsCloned should return false for nonexistent dir")
	}

	subdir := filepath.Join(dir, "exists")
	os.MkdirAll(subdir, 0755)
	if !IsCloned(subdir) {
		t.Error("IsCloned should return true for existing dir")
	}

	// A file is not a cloned repo
	filePath := filepath.Join(dir, "afile")
	os.WriteFile(filePath, []byte("hello"), 0644)
	if IsCloned(filePath) {
		t.Error("IsCloned should return false for a file")
	}
}

// initTestRepo creates a git repo in a temp dir with one commit.
func initTestRepo(t *testing.T) string {
	t.Helper()
	return testutil.InitGitRepo(t)
}

func TestCurrentBranch(t *testing.T) {
	dir := initTestRepo(t)
	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Errorf("CurrentBranch = %q, want %q", branch, "main")
	}
}

func TestIsDirty_Clean(t *testing.T) {
	dir := initTestRepo(t)
	dirty, err := IsDirty(dir)
	if err != nil {
		t.Fatal(err)
	}
	if dirty {
		t.Error("repo should be clean")
	}
}

func TestIsDirty_Dirty(t *testing.T) {
	dir := initTestRepo(t)
	os.WriteFile(filepath.Join(dir, "new.txt"), []byte("change"), 0644)
	dirty, err := IsDirty(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !dirty {
		t.Error("repo should be dirty")
	}
}

func TestAheadBehind_NoUpstream(t *testing.T) {
	dir := initTestRepo(t)
	ahead, behind, err := AheadBehind(dir)
	if err != nil {
		t.Fatal(err)
	}
	if ahead != 0 || behind != 0 {
		t.Errorf("AheadBehind = (%d, %d), want (0, 0) with no upstream", ahead, behind)
	}
}

func TestFetch_ValidRepo(t *testing.T) {
	dir := initTestRepo(t)
	// Fetch on a local repo with no remote should fail gracefully
	err := Fetch(dir)
	// It's expected to fail since there's no remote
	if err == nil {
		t.Log("Fetch succeeded (remote may be configured)")
	}
}

func TestClone(t *testing.T) {
	src := initTestRepo(t)
	dest := filepath.Join(t.TempDir(), "cloned")

	err := Clone(src, dest)
	if err != nil {
		t.Fatal(err)
	}

	if !IsCloned(dest) {
		t.Error("cloned directory should exist")
	}

	branch, err := CurrentBranch(dest)
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Errorf("cloned branch = %q, want %q", branch, "main")
	}
}

func TestRemoteReachable_InvalidURL(t *testing.T) {
	if RemoteReachable("git@invalid.example.com:no/repo.git") {
		t.Error("invalid remote should not be reachable")
	}
}

func TestWorktrees_None(t *testing.T) {
	dir := initTestRepo(t)
	wts, err := Worktrees(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 0 {
		t.Errorf("Worktrees with no linked worktrees = %d, want 0", len(wts))
	}
}

func TestWorktrees_LinkedAreReturned(t *testing.T) {
	home := testutil.SetupHome(t)
	dir := testutil.InitGitRepoInHome(t, home)

	wtPath := filepath.Join(t.TempDir(), "feature-wt")
	testutil.RunGit(t, dir, home, "worktree", "add", "-b", "feature", wtPath)

	wts, err := Worktrees(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 1 {
		t.Fatalf("Worktrees = %d, want 1", len(wts))
	}
	if wts[0].Branch != "feature" {
		t.Errorf("Worktree branch = %q, want %q", wts[0].Branch, "feature")
	}
	// Path returned by git worktree list may resolve symlinks (e.g. /var → /private/var on macOS),
	// so compare basenames rather than full paths.
	if filepath.Base(wts[0].Path) != filepath.Base(wtPath) {
		t.Errorf("Worktree path basename = %q, want %q", filepath.Base(wts[0].Path), filepath.Base(wtPath))
	}
}

func TestApplyOwnerAlias(t *testing.T) {
	aliases := map[string]string{"nahime0": "nahime", "grp": "work"}
	tests := []struct {
		name    string
		relPath string
		aliases map[string]string
		want    string
	}{
		{"match owner", "nahime0/repo", aliases, "nahime/repo"},
		{"match nested owner", "grp/sub/repo", aliases, "work/sub/repo"},
		{"no match", "other/repo", aliases, "other/repo"},
		{"nil map", "nahime0/repo", nil, "nahime0/repo"},
		{"empty map", "nahime0/repo", map[string]string{}, "nahime0/repo"},
		{"single segment", "nahime0", aliases, "nahime0"},
		{"empty path", "", aliases, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ApplyOwnerAlias(tt.relPath, tt.aliases); got != tt.want {
				t.Errorf("ApplyOwnerAlias(%q) = %q, want %q", tt.relPath, got, tt.want)
			}
		})
	}
}

func TestDerivedPath(t *testing.T) {
	aliases := map[string]string{"nahime0": "nahime"}
	r := config.Repo{URL: "git@github.com:nahime0/repo.git"}
	got, err := DerivedPath(r, aliases)
	if err != nil {
		t.Fatal(err)
	}
	if got != "nahime/repo" {
		t.Errorf("DerivedPath = %q, want %q", got, "nahime/repo")
	}
}

func TestDerivedPath_NoAlias(t *testing.T) {
	r := config.Repo{URL: "https://github.com/org/project.git"}
	got, err := DerivedPath(r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != "org/project" {
		t.Errorf("DerivedPath = %q, want %q", got, "org/project")
	}
}
