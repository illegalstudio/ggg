package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

// loadRepos loads the config, applies --group filter, and returns the config
// with the filtered repos. Returns an error message if no repos match.
func loadRepos(cmd *cobra.Command) (*config.Config, []config.Repo, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	group, _ := cmd.Flags().GetString("group")
	repos := filterByGroup(cfg.Repos, group)

	if len(repos) == 0 {
		if !jsonOutput {
			fmt.Println(ui.Info.Render("No repositories configured."))
		}
		return cfg, nil, nil
	}

	return cfg, repos, nil
}

// confirmAll shows a select prompt asking the user to confirm an action on all repos.
// Returns true if confirmed, false if aborted. In JSON mode, auto-confirms.
func confirmAll(title, yesLabel string) (bool, error) {
	if jsonOutput {
		return true, nil
	}
	var choice string
	err := huh.NewSelect[string]().
		Title(title).
		Options(
			huh.NewOption(yesLabel, "yes"),
			huh.NewOption("No, abort", "no"),
		).
		Value(&choice).
		Run()
	if err != nil {
		return false, err
	}
	if choice == "no" {
		fmt.Println(ui.Muted.Render("Aborted."))
		return false, nil
	}
	return true, nil
}

// getFilter returns the filter string from --filter flag or the first positional argument.
func getFilter(cmd *cobra.Command, args []string) string {
	filter, _ := cmd.Flags().GetString("filter")
	if filter == "" && len(args) > 0 {
		filter = args[len(args)-1]
	}
	return filter
}

// resolveBulkRepos loads config, applies --group and filter resolution, and
// returns the final repo set plus the effective filter string.
func resolveBulkRepos(cmd *cobra.Command, args []string) (*config.Config, []config.Repo, string, error) {
	cfg, repos, err := loadRepos(cmd)
	if err != nil || len(repos) == 0 {
		return cfg, repos, "", err
	}

	filter := getFilter(cmd, args)
	repos = filterByName(repos, filter)
	return cfg, repos, filter, nil
}

// filterByName filters repos by substring match on URL, path, or derived path.
func filterByName(repos []config.Repo, filter string) []config.Repo {
	if filter == "" {
		return repos
	}
	filterLower := strings.ToLower(filter)
	var filtered []config.Repo
	for _, r := range repos {
		derived, _ := repo.DerivePathFromURL(r.URL)
		if strings.Contains(strings.ToLower(r.URL), filterLower) ||
			strings.Contains(strings.ToLower(r.Path), filterLower) ||
			strings.Contains(strings.ToLower(derived), filterLower) {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// resolveOneRepo returns a single repo using exact-match, partial-match, then
// interactive disambiguation.
func resolveOneRepo(repos []config.Repo, query string) (config.Repo, error) {
	idx, err := resolveOneRepoIndex(repos, query)
	if err != nil {
		return config.Repo{}, err
	}
	return repos[idx], nil
}

// resolveOneRepoIndex returns the index of a single repo using exact-match,
// partial-match, then interactive disambiguation.
func resolveOneRepoIndex(repos []config.Repo, query string) (int, error) {
	matches := matchRepoIndices(repos, query)
	if len(matches) == 0 {
		return -1, fmt.Errorf("repository %q not found in config", query)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return selectRepoIndex(repos, matches, fmt.Sprintf("Multiple repositories match %q", query))
}

func matchRepoIndices(repos []config.Repo, query string) []int {
	// First pass: exact match.
	for i, r := range repos {
		derived, _ := repo.DerivePathFromURL(r.URL)
		if r.Path == query || r.URL == query || derived == query {
			return []int{i}
		}
	}

	// Second pass: partial match (substring, case-insensitive).
	queryLower := strings.ToLower(query)
	var matches []int
	for i, r := range repos {
		derived, _ := repo.DerivePathFromURL(r.URL)
		if strings.Contains(strings.ToLower(r.URL), queryLower) ||
			strings.Contains(strings.ToLower(r.Path), queryLower) ||
			strings.Contains(strings.ToLower(derived), queryLower) {
			matches = append(matches, i)
		}
	}

	return matches
}

func selectRepoIndex(repos []config.Repo, indices []int, title string) (int, error) {
	if jsonOutput {
		return -1, fmt.Errorf("%s: refusing interactive prompt in --json mode (be more specific)", title)
	}

	options := make([]huh.Option[int], len(indices))
	for i, idx := range indices {
		options[i] = huh.NewOption(repoChoiceLabel(repos[idx]), idx)
	}

	var choice int
	err := huh.NewSelect[int]().
		Title(title).
		Description("Start typing to filter").
		Options(options...).
		Filtering(true).
		Height(20).
		Value(&choice).
		Run()
	if err != nil {
		return -1, err
	}

	return choice, nil
}

func repoChoiceLabel(r config.Repo) string {
	label := r.URL
	if r.Path != "" {
		label += " (path: " + r.Path + ")"
	}
	if len(r.Groups) > 0 {
		label += " [" + strings.Join(r.Groups, ", ") + "]"
	}
	return label
}

func confirmBulkAction(filter string, count int, title, yesLabel string) (bool, error) {
	if filter != "" {
		return true, nil
	}
	return confirmAll(fmt.Sprintf(title, count), yesLabel)
}

func runParallelWithSpinner[T any, R any](jobs []T, title string, runner func(T) R) ([]R, error) {
	results := make([]R, len(jobs))
	var wg sync.WaitGroup

	action := func() {
		for i, job := range jobs {
			wg.Add(1)
			go func(idx int, job T) {
				defer wg.Done()
				results[idx] = runner(job)
			}(i, job)
		}
		wg.Wait()
	}

	if jsonOutput {
		action()
		return results, nil
	}

	if err := spinner.New().Title(title).Action(action).Run(); err != nil {
		return nil, err
	}

	return results, nil
}

// defaultEditor returns the user's preferred editor from $EDITOR, $VISUAL, or "vi".
func defaultEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	return "vi"
}

// requireBinary checks if a binary is available in $PATH.
// Returns nil if found, or a user-friendly error if not.
func requireBinary(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		return fmt.Errorf("%s is not installed. Install it from https://cli.github.com", name)
	}
	return nil
}
