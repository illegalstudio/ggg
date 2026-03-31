package cmd

import (
	"fmt"
	"strings"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a repository from the configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.LoadRaw()
		if err != nil {
			return err
		}

		idx, err := findRepoIndex(cfg.Repos, name)
		if err != nil {
			return err
		}

		removed := cfg.Repos[idx]
		cfg.Repos = append(cfg.Repos[:idx], cfg.Repos[idx+1:]...)

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("  %s Removed %s\n", ui.Success.Render("✓"), ui.Repo.Render(removed.URL))
		return nil
	},
}

func findRepoIndex(repos []config.Repo, name string) (int, error) {
	// Exact match
	for i, r := range repos {
		derived, _ := repo.DerivePathFromURL(r.URL)
		if r.Path == name || r.URL == name || derived == name {
			return i, nil
		}
	}

	// Partial match
	nameLower := strings.ToLower(name)
	var matches []int
	for i, r := range repos {
		derived, _ := repo.DerivePathFromURL(r.URL)
		if strings.Contains(strings.ToLower(r.URL), nameLower) ||
			strings.Contains(strings.ToLower(r.Path), nameLower) ||
			strings.Contains(strings.ToLower(derived), nameLower) {
			matches = append(matches, i)
		}
	}

	if len(matches) == 0 {
		return -1, fmt.Errorf("repository %q not found in config", name)
	}
	if len(matches) == 1 {
		return matches[0], nil
	}

	options := make([]huh.Option[int], len(matches))
	for i, idx := range matches {
		options[i] = huh.NewOption(repos[idx].URL, idx)
	}

	var choice int
	err := huh.NewSelect[int]().
		Title(fmt.Sprintf("Multiple repositories match %q", name)).
		Options(options...).
		Value(&choice).
		Run()
	if err != nil {
		return -1, err
	}

	return choice, nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
