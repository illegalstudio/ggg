package cmd

import (
	"testing"

	"go-git-get/config"
)

func TestFilterByGroup_Empty(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/a.git", Group: "work"},
		{URL: "git@github.com:user/b.git", Group: "personal"},
		{URL: "git@github.com:user/c.git"},
	}

	result := filterByGroup(repos, "")
	if len(result) != 3 {
		t.Errorf("filterByGroup with empty group should return all repos, got %d", len(result))
	}
}

func TestFilterByGroup_Match(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/a.git", Group: "work"},
		{URL: "git@github.com:user/b.git", Group: "personal"},
		{URL: "git@github.com:user/c.git", Group: "work"},
	}

	result := filterByGroup(repos, "work")
	if len(result) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(result))
	}
	if result[0].URL != "git@github.com:user/a.git" {
		t.Errorf("result[0].URL = %q, want a.git", result[0].URL)
	}
	if result[1].URL != "git@github.com:user/c.git" {
		t.Errorf("result[1].URL = %q, want c.git", result[1].URL)
	}
}

func TestFilterByGroup_NoMatch(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/a.git", Group: "work"},
	}

	result := filterByGroup(repos, "oss")
	if len(result) != 0 {
		t.Errorf("expected 0 repos, got %d", len(result))
	}
}

func TestFilterRepo_ExactMatch_URL(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/alpha.git"},
		{URL: "git@github.com:user/beta.git"},
	}

	result, err := filterRepo(repos, "git@github.com:user/beta.git")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].URL != "git@github.com:user/beta.git" {
		t.Errorf("expected beta.git, got %v", result)
	}
}

func TestFilterRepo_ExactMatch_Path(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/alpha.git", Path: "my/alpha"},
		{URL: "git@github.com:user/beta.git"},
	}

	result, err := filterRepo(repos, "my/alpha")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].URL != "git@github.com:user/alpha.git" {
		t.Errorf("expected alpha.git, got %v", result)
	}
}

func TestFilterRepo_ExactMatch_Derived(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/myrepo.git"},
	}

	result, err := filterRepo(repos, "user/myrepo")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
}

func TestFilterRepo_PartialMatch(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/elephc.git"},
		{URL: "git@github.com:user/other.git"},
	}

	result, err := filterRepo(repos, "eleph")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].URL != "git@github.com:user/elephc.git" {
		t.Errorf("expected elephc.git, got %v", result)
	}
}

func TestFilterRepo_PartialMatch_CaseInsensitive(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/MyProject.git"},
		{URL: "git@github.com:user/other.git"},
	}

	result, err := filterRepo(repos, "myproject")
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].URL != "git@github.com:user/MyProject.git" {
		t.Errorf("expected MyProject.git, got %v", result)
	}
}

func TestFilterRepo_NotFound(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/alpha.git"},
	}

	_, err := filterRepo(repos, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent repo")
	}
}

func TestFindRepoIndex_ExactMatch(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/alpha.git"},
		{URL: "git@github.com:user/beta.git"},
	}

	idx, err := findRepoIndex(repos, "git@github.com:user/beta.git")
	if err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
}

func TestFindRepoIndex_PartialMatch(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/alpha.git"},
		{URL: "git@github.com:user/beta.git"},
	}

	idx, err := findRepoIndex(repos, "beta")
	if err != nil {
		t.Fatal(err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
}

func TestFindRepoIndex_NotFound(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:user/alpha.git"},
	}

	_, err := findRepoIndex(repos, "zzz")
	if err == nil {
		t.Error("expected error for nonexistent repo")
	}
}

func TestHTTPURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "ssh",
			in:   "git@github.com:user/repo.git",
			want: "https://github.com/user/repo",
		},
		{
			name: "https",
			in:   "https://github.com/user/repo.git",
			want: "https://github.com/user/repo",
		},
		{
			name: "plain passthrough",
			in:   "file:///tmp/repo",
			want: "file:///tmp/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := httpURL(tt.in); got != tt.want {
				t.Errorf("httpURL(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
