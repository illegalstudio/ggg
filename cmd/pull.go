package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull [name]",
	Short: "Pull latest changes (all or a specific repo, only if clean)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Repos) == 0 {
			fmt.Println(ui.Info.Render("No repositories configured."))
			return nil
		}

		repos := cfg.Repos
		if len(args) > 0 {
			filtered, err := filterRepo(cfg.Repos, args[0])
			if err != nil {
				return err
			}
			repos = filtered
		}

		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				fmt.Println(ui.Error.Render(fmt.Sprintf("  ✗ Skipping %s: %v", r.URL, err)))
				continue
			}

			if !repo.IsCloned(fullPath) {
				fmt.Printf("  %s %s\n", ui.Muted.Render("○"), ui.Muted.Render("Not cloned: "+r.URL))
				continue
			}

			dirty, err := repo.IsDirty(fullPath)
			if err != nil {
				fmt.Println(ui.Error.Render(fmt.Sprintf("  ✗ %s: %v", r.URL, err)))
				continue
			}
			if dirty {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.URL), ui.Muted.Render("(dirty, skipping)"))
				continue
			}

			fmt.Printf("  %s %s\n", ui.Info.Render("⬇"), ui.Repo.Render(r.URL))
			if err := repo.Pull(fullPath); err != nil {
				fmt.Println(ui.Error.Render(fmt.Sprintf("  ✗ Error: %v", err)))
			} else {
				fmt.Println(ui.Success.Render("  ✓ Up to date"))
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
