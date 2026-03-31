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
		hostAndPath = strings.Replace(hostAndPath, ":", "/", 1)
		hostAndPath = strings.TrimSuffix(hostAndPath, ".git")
		return hostAndPath, nil
	}

	// Handle HTTPS-style URLs
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	path := strings.TrimSuffix(parsed.Host+parsed.Path, ".git")
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

// Clone clones a git repository to the given path.
func Clone(repoURL, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("cannot create parent directory: %w", err)
	}

	cmd := exec.Command("git", "clone", repoURL, destPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}
	return nil
}
