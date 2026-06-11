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
