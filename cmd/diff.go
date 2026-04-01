package cmd

import (
	"fmt"

	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:     "diff [name]",
	Short:   "Show a summary of changed files in dirty repositories",
	GroupID: GroupRepo,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, err := loadRepos(cmd)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		if len(args) > 0 {
			filtered, err := filterRepo(repos, args[0])
			if err != nil {
				return err
			}
			repos = filtered
		} else {
			ok, err := confirmAll(fmt.Sprintf("Show diff for all %d repositories?", len(repos)), "Yes, show all")
			if err != nil {
				return err
			}
			if !ok {
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
