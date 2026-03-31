package repo

import (
	"os"
	"path/filepath"
	"testing"

	"go-git-get/config"
)

func TestDerivePathFromURL_SSH(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"git@github.com:user/repo.git", "github.com/user/repo"},
		{"git@gitlab.com:org/project.git", "gitlab.com/org/project"},
		{"git@github.com:user/repo", "github.com/user/repo"},
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
		{"https://github.com/user/repo.git", "github.com/user/repo"},
		{"https://gitlab.com/org/project.git", "gitlab.com/org/project"},
		{"https://github.com/user/repo", "github.com/user/repo"},
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
	want := filepath.Join("/base", "github.com/user/repo")
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
}
