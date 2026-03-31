package repo

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go-git-get/config"
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

// Pull runs git pull in the given repo directory (quiet mode).
func Pull(repoPath string) error {
	cmd := exec.Command("git", "pull", "--quiet")
	cmd.Dir = repoPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}
	return nil
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
