package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"

	"github.com/spf13/cobra"
)

var cdCmd = &cobra.Command{
	Use:   "cd <name>",
	Short: "Print the path of a repository (use ggg shell-init for seamless cd)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		repos, err := filterRepo(cfg.Repos, args[0])
		if err != nil {
			return err
		}

		fullPath, err := repo.FullPath(cfg.BaseDir, repos[0])
		if err != nil {
			return err
		}

		if !repo.IsCloned(fullPath) {
			return fmt.Errorf("repository is not cloned: %s", fullPath)
		}

		fmt.Print(fullPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cdCmd)
}
