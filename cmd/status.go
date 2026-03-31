package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all configured repositories",
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

		fmt.Println(ui.Title.Render("Repository Status"))
		fmt.Println()
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				fmt.Printf("  %s %s\n", ui.Error.Render("✗"), ui.Error.Render(r.URL+" (invalid URL)"))
				continue
			}

			if !repo.IsCloned(fullPath) {
				fmt.Printf("  %s %s %s\n", ui.Muted.Render("○"), ui.Repo.Render(r.URL), ui.Muted.Render("(not cloned)"))
				continue
			}

			branch, _ := repo.CurrentBranch(fullPath)
			dirty, _ := repo.IsDirty(fullPath)
			ahead, behind, _ := repo.AheadBehind(fullPath)

			status := ui.Success.Render("clean")
			if dirty {
				status = ui.Error.Render("dirty")
			}

			branchStr := ui.Info.Render(branch)

			var syncStr string
			if ahead > 0 && behind > 0 {
				syncStr = ui.Error.Render(fmt.Sprintf(" ↑%d ↓%d", ahead, behind))
			} else if ahead > 0 {
				syncStr = ui.Info.Render(fmt.Sprintf(" ↑%d", ahead))
			} else if behind > 0 {
				syncStr = ui.Error.Render(fmt.Sprintf(" ↓%d", behind))
			}

			fmt.Printf("  %s %s [%s] %s%s\n", ui.Success.Render("✓"), ui.Repo.Render(r.URL), branchStr, status, syncStr)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	statusCmd.Flags().StringP("group", "g", "", "Show only repos in this group")
	rootCmd.AddCommand(statusCmd)
}
