package cli

import (
	"fmt"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"

	"github.com/spf13/cobra"
)

var cdCmd = &cobra.Command{
	Use:               "cd <name>",
	Short:             "Print a repository path (chdir requires shell integration — see `ggg shell-init`)",
	GroupID:           GroupRepo,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: repoCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		r, err := resolveOneRepo(cfg.Repos, args[0])
		if err != nil {
			return err
		}

		fullPath, err := repo.FullPath(cfg.BaseDir, r)
		if err != nil {
			return err
		}

		if !repo.IsCloned(fullPath) {
			return fmt.Errorf("repository is not cloned: %s", fullPath)
		}

		if done, err := maybeJSON(map[string]any{"path": fullPath, "url": r.URL}); done {
			return err
		}

		fmt.Print(fullPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cdCmd)
}
