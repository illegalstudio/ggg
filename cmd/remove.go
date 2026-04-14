package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Remove a repository from the configuration",
	GroupID: GroupConfig,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.LoadRaw()
		if err != nil {
			return err
		}

		idx, err := resolveOneRepoIndex(cfg.Repos, name)
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

func init() {
	rootCmd.AddCommand(removeCmd)
}
