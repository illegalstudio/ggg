package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured repositories and their status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Repos) == 0 {
			fmt.Println(ui.Info.Render("No repositories configured."))
			return nil
		}

		fmt.Println(ui.Title.Render("Repositories"))
		fmt.Println()
		for _, r := range cfg.Repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				fmt.Printf("  %s %s\n", ui.Error.Render("✗"), ui.Error.Render(r.URL+" (invalid URL)"))
				continue
			}

			if repo.IsCloned(fullPath) {
				fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.URL), ui.Path.Render(fullPath))
			} else {
				fmt.Printf("  %s %s → %s\n", ui.Muted.Render("○"), ui.Repo.Render(r.URL), ui.Path.Render(fullPath))
			}
		}
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
