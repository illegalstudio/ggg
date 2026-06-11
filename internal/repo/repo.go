package repo

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/illegalstudio/ggg/internal/config"
)

// DerivePathFromURL extracts a filesystem path from a git URL.
// e.g. "git@github.com:user/repo.git" -> "github.com/user/repo"
// e.g. "https://github.com/user/repo.git" -> "github.com/user/repo"
func DerivePathFromURL(rawURL string) (string, error) {
	// Handle SSH-style URLs (git@host:user/repo.git)
	if strings.Contains(rawURL, "@") && strings.Contains(rawURL, ":") && !strings.Contains(rawURL, "://") {
		parts := strings.SplitN(rawURL, "@", 2)
		hostAndPath := parts[1]
		// Extract only the path after the host (user/repo)
		colonIdx := strings.Index(hostAndPath, ":")
		path := hostAndPath[colonIdx+1:]
		path = strings.TrimSuffix(path, ".git")
		return path, nil
	}

	// Handle HTTPS-style URLs
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	path := strings.TrimPrefix(parsed.Path, "/")
	path = strings.TrimSuffix(path, ".git")
	return path, nil
}

// ApplyOwnerAlias replaces the first segment (the owner) of a derived relative
// path when it matches an alias key. Returns relPath unchanged when there is no
// match, the map is empty/nil, or the path has no owner segment.
func ApplyOwnerAlias(relPath string, aliases map[string]string) string {
	if len(aliases) == 0 || relPath == "" {
		return relPath
	}
	owner, rest, hasRest := strings.Cut(relPath, "/")
	if !hasRest {
		return relPath
	}
	alias, ok := aliases[owner]
	if !ok {
		return relPath
	}
	return alias + "/" + rest
}

// DerivedPath returns the repo's path derived from its URL, with owner aliases
// applied. It ignores any explicit Path set on the repo.
func DerivedPath(r config.Repo, aliases map[string]string) (string, error) {
	derived, err := DerivePathFromURL(r.URL)
	if err != nil {
		return "", err
	}
	return ApplyOwnerAlias(derived, aliases), nil
}

// FullPath returns the absolute path where a repo should be cloned.
func FullPath(baseDir string, r config.Repo) (string, error) {
	if r.Path != "" {
		return filepath.Join(baseDir, r.Path), nil
	}
	derived, err := DerivePathFromURL(r.URL)
	if err != nil {
		return "", err
	}
	return filepath.Join(baseDir, derived), nil
}

// IsCloned checks if the repo directory already exists.
func IsCloned(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// CurrentBranch returns the current branch name.
func CurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// AheadBehind returns how many commits ahead/behind the tracking branch.
func AheadBehind(repoPath string) (int, int, error) {
	cmd := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return 0, 0, nil // no upstream configured
	}
	var ahead, behind int
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d\t%d", &ahead, &behind)
	return ahead, behind, nil
}

// IsDirty checks if a repo has uncommitted changes.
func IsDirty(repoPath string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// Pull runs git pull in the given repo directory with the specified strategy.
// Valid strategies: "merge", "rebase", "ff-only".
func Pull(repoPath string, strategy string) error {
	args := []string{"pull", "--quiet"}
	switch strategy {
	case "rebase":
		args = append(args, "--rebase")
	case "ff-only":
		args = append(args, "--ff-only")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	return nil
}

// Fetch runs git fetch in the given repo directory.
func Fetch(repoPath string) error {
	cmd := exec.Command("git", "fetch", "--quiet")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}
	return nil
}

// RemoteReachable checks if the git remote is reachable.
func RemoteReachable(repoURL string) bool {
	cmd := exec.Command("git", "ls-remote", "--exit-code", "--quiet", repoURL)
	return cmd.Run() == nil
}

// Stash runs git stash in the given repo directory.
func Stash(repoPath string) error {
	cmd := exec.Command("git", "stash", "--quiet")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git stash failed: %w", err)
	}
	return nil
}

// Checkout switches to the given branch in the repo directory.
func Checkout(repoPath, branch string) error {
	cmd := exec.Command("git", "checkout", "--quiet", branch)
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}
	return nil
}

// HasBranch checks if a branch exists locally in the repo.
func HasBranch(repoPath, branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "--quiet", branch)
	cmd.Dir = repoPath
	return cmd.Run() == nil
}

// LocalBranches returns local branch names for shell completion.
func LocalBranches(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref failed: %w", err)
	}

	var branches []string
	for _, line := range strings.Split(string(out), "\n") {
		branch := strings.TrimSpace(line)
		if branch != "" {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

// DiffSummary returns the short diff stat for a dirty repo.
func DiffSummary(repoPath string) (string, error) {
	cmd := exec.Command("git", "diff", "--stat")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Push runs git push in the given repo directory.
func Push(repoPath string) error {
	cmd := exec.Command("git", "push", "--quiet")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}
	return nil
}

// Worktree describes a linked git worktree.
type Worktree struct {
	Path   string
	Branch string
}

// Worktrees returns the linked worktrees for the repo at repoPath, excluding
// the main worktree. The main worktree is always reported first by
// `git worktree list`, so we drop it and return the rest.
func Worktrees(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	var all []Worktree
	var current Worktree
	flush := func() {
		if current.Path != "" {
			all = append(all, current)
		}
		current = Worktree{}
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			flush()
			continue
		}
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		case line == "detached":
			current.Branch = "(detached)"
		}
	}
	flush()

	if len(all) <= 1 {
		return nil, nil
	}
	return all[1:], nil
}

// Clone clones a git repository to the given path (quiet mode, no stdout).
func Clone(repoURL, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("cannot create parent directory: %w", err)
	}

	cmd := exec.Command("git", "clone", "--quiet", repoURL, destPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}
