package cli

import (
	"reflect"
	"testing"

	"github.com/illegalstudio/ggg/internal/config"
)

func TestRepoCompletionItemsIncludesPathsAndBasenames(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:acme/project.git"},
		{URL: "git@github.com:acme/service.git", Path: "team/service-api"},
	}

	got := repoCompletionItems(repos, "", nil)
	want := []string{"acme/project", "project", "service-api", "team/service-api"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestRepoCompletionItemsFiltersByPrefix(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:acme/project.git"},
		{URL: "git@github.com:acme/service.git", Path: "team/service-api"},
	}

	got := repoCompletionItems(repos, "ser", nil)
	want := []string{"service-api"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestRepoCompletionItems_Alias(t *testing.T) {
	repos := []config.Repo{
		{URL: "git@github.com:nahime0/repo.git"},
	}
	aliases := map[string]string{"nahime0": "nahime"}

	got := repoCompletionItems(repos, "", aliases)

	contains := func(items []string, target string) bool {
		for _, s := range items {
			if s == target {
				return true
			}
		}
		return false
	}

	if !contains(got, "nahime/repo") {
		t.Errorf("expected aliased path %q in completions, got %v", "nahime/repo", got)
	}
	if contains(got, "nahime0/repo") {
		t.Errorf("un-aliased path %q should not appear in completions, got %v", "nahime0/repo", got)
	}
}
