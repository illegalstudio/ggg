package cli

import (
	"fmt"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:               "remove [name]",
	Short:             "Remove a repository from the configuration",
	GroupID:           GroupConfig,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: repoCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadRaw()
		if err != nil {
			return err
		}
		if len(cfg.Repos) == 0 {
			if done, err := maybeJSON(map[string]any{"removed": false, "reason": "no repositories configured"}); done {
				return err
			}
			fmt.Println(ui.Info.Render("No repositories configured."))
			return nil
		}

		var idx int
		if len(args) == 0 {
			indices := make([]int, len(cfg.Repos))
			for i := range cfg.Repos {
				indices[i] = i
			}
			idx, err = selectRepoIndex(cfg.Repos, indices, "Select a repository to remove")
			if err != nil {
				return err
			}
		} else {
			idx, err = resolveOneRepoIndex(cfg.Repos, args[0], cfg.Aliases)
			if err != nil {
				return err
			}
		}

		removed := cfg.Repos[idx]
		cfg.Repos = append(cfg.Repos[:idx], cfg.Repos[idx+1:]...)

		if err := config.Save(cfg); err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"removed": true, "url": removed.URL}); done {
			return err
		}

		fmt.Printf("  %s Removed %s\n", ui.Success.Render("✓"), ui.Repo.Render(removed.URL))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
