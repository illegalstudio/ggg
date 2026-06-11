package cli

import (
	"sort"
	"strings"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"

	"github.com/spf13/cobra"
)

func repoCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return configuredRepoCompletions(cmd, toComplete)
}

func checkoutCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return branchCompletion(cmd, toComplete)
	case 1:
		return configuredRepoCompletions(cmd, toComplete)
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func configuredRepoCompletions(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	group, _ := groupFilter(cmd)
	repos := filterByGroup(cfg.Repos, group)
	return repoCompletionItems(repos, toComplete), cobra.ShellCompDirectiveNoFileComp
}

func repoCompletionItems(repos []config.Repo, toComplete string) []string {
	seen := map[string]bool{}
	var comps []string

	for _, r := range repos {
		repoPath := r.Path
		if repoPath == "" {
			derived, err := repo.DerivePathFromURL(r.URL)
			if err != nil {
				continue
			}
			repoPath = derived
		}

		addCompletion(&comps, seen, repoPath, toComplete)
		base := completionPathBase(repoPath)
		if base != repoPath {
			addCompletion(&comps, seen, base, toComplete)
		}
		if toComplete != "" {
			addCompletion(&comps, seen, r.URL, toComplete)
		}
	}

	sort.Strings(comps)
	return comps
}

func groupCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	seen := map[string]bool{}
	var comps []string
	for _, r := range cfg.Repos {
		for _, g := range r.Groups {
			addCompletion(&comps, seen, g, toComplete)
		}
	}
	sort.Strings(comps)
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func branchCompletion(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.Load()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	group, _ := groupFilter(cmd)
	repos := filterByGroup(cfg.Repos, group)
	seen := map[string]bool{}
	var comps []string

	for _, r := range repos {
		fullPath, err := repo.FullPath(cfg.BaseDir, cfg.Aliases, r)
		if err != nil || !repo.IsCloned(fullPath) {
			continue
		}

		branches, err := repo.LocalBranches(fullPath)
		if err != nil {
			continue
		}
		for _, branch := range branches {
			addCompletion(&comps, seen, branch, toComplete)
		}
	}

	sort.Strings(comps)
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func shellCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	seen := map[string]bool{}
	var comps []string
	for _, shell := range []string{"bash", "zsh", "fish"} {
		addCompletion(&comps, seen, shell, toComplete)
	}
	return comps, cobra.ShellCompDirectiveNoFileComp
}

func addCompletion(comps *[]string, seen map[string]bool, candidate, toComplete string) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" || seen[candidate] || !completionMatches(candidate, toComplete) {
		return
	}
	seen[candidate] = true
	*comps = append(*comps, candidate)
}

func completionMatches(candidate, toComplete string) bool {
	return toComplete == "" || strings.HasPrefix(strings.ToLower(candidate), strings.ToLower(toComplete))
}

func completionPathBase(path string) string {
	path = strings.TrimRight(path, "/\\")
	idx := strings.LastIndexAny(path, "/\\")
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}

func registerGroupCompletion(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("group", groupCompletion)
}

func registerFilterCompletion(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("filter", repoFlagCompletion)
}

func repoFlagCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return configuredRepoCompletions(cmd, toComplete)
}
