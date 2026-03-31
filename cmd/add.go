package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add a repository to the configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		cfg, err := config.LoadRaw()
		if err != nil {
			return err
		}

		for _, r := range cfg.Repos {
			if r.URL == url {
				return fmt.Errorf("repository %s is already configured", url)
			}
		}

		cfg.Repos = append(cfg.Repos, config.Repo{URL: url})

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("  %s Added %s\n", ui.Success.Render("✓"), ui.Repo.Render(url))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
