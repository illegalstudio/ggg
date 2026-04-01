package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:     "diff [name]",
	Short:   "Show a summary of changed files in dirty repositories",
	GroupID: GroupRepo,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		group, _ := cmd.Flags().GetString("group")
		repos := filterByGroup(cfg.Repos, group)

		if len(repos) == 0 {
			fmt.Println(ui.Info.Render("No repositories configured."))
			return nil
		}

		if len(args) > 0 {
			filtered, err := filterRepo(cfg.Repos, args[0])
			if err != nil {
				return err
			}
			repos = filtered
		} else {
			var confirm bool
			err := huh.NewConfirm().
				Title(fmt.Sprintf("Show diff for all %d repositories?", len(repos))).
				Value(&confirm).
				Run()
			if err != nil {
				return err
			}
			if !confirm {
				fmt.Println(ui.Muted.Render("Aborted."))
				return nil
			}
		}

		found := false
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				continue
			}
			if !repo.IsCloned(fullPath) {
				continue
			}
			dirty, _ := repo.IsDirty(fullPath)
			if !dirty {
				continue
			}

			summary, err := repo.DiffSummary(fullPath)
			if err != nil {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.URL), ui.Error.Render(err.Error()))
				continue
			}

			found = true
			fmt.Printf("\n  %s %s\n", ui.Repo.Render("●"), ui.Repo.Render(r.URL))
			fmt.Printf("%s\n", summary)
		}

		if !found {
			fmt.Println(ui.Info.Render("All repositories are clean."))
		}

		return nil
	},
}

func init() {
	diffCmd.Flags().StringP("group", "g", "", "Show diff only for repos in this group")
	rootCmd.AddCommand(diffCmd)
}
