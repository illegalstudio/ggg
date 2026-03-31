package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"

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
			fmt.Println("No repositories configured.")
			return nil
		}

		for _, r := range cfg.Repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				fmt.Printf("  ✗ %s (invalid URL)\n", r.URL)
				continue
			}

			status := "✗ not cloned"
			if repo.IsCloned(fullPath) {
				status = "✓ cloned"
			}
			fmt.Printf("  %s  %s → %s\n", status, r.URL, fullPath)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
