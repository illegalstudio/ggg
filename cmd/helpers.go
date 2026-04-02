package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/charmbracelet/huh"
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
		fmt.Println(ui.Info.Render("No repositories configured."))
		return cfg, nil, nil
	}

	return cfg, repos, nil
}

// confirmAll shows a select prompt asking the user to confirm an action on all repos.
// Returns true if confirmed, false if aborted.
func confirmAll(title, yesLabel string) (bool, error) {
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
