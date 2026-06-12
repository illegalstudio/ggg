package cli

import (
	"fmt"

	"github.com/illegalstudio/ggg/internal/repo"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:               "diff [filter]",
	Short:             "Show a summary of changed files in dirty repositories",
	GroupID:           GroupRepo,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: repoCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, filter, err := resolveBulkRepos(cmd, args)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			if done, err := maybeJSON(map[string]any{"diffs": []any{}}); done {
				return err
			}
			return nil
		}

		type diffEntry struct {
			URL     string `json:"url"`
			Summary string `json:"summary"`
			Error   string `json:"error,omitempty"`
		}

		var entries []diffEntry
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, cfg.Aliases, r)
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
			entries = append(entries, diffEntry{URL: r.URL, Summary: summary, Error: errString(err)})
		}

		if len(entries) == 0 {
			if done, err := maybeJSON(map[string]any{"diffs": []diffEntry{}}); done {
				return err
			}
			fmt.Println(ui.Info.Render("All repositories are clean."))
			return nil
		}

		ok, err := confirmBulkAction(filter, len(entries), "Show diff for %d repositories?", "Yes, show all")
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		if done, err := maybeJSON(map[string]any{"diffs": entries}); done {
			return err
		}

		for _, entry := range entries {
			if entry.Error != "" {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(entry.URL), ui.Error.Render(entry.Error))
				continue
			}

			fmt.Printf("\n  %s %s\n", ui.Repo.Render("●"), ui.Repo.Render(entry.URL))
			fmt.Printf("%s\n", entry.Summary)
		}

		return nil
	},
}

func init() {
	diffCmd.Flags().StringArrayP("group", "g", nil, "Show diff only for repos in this group")
	diffCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	registerGroupCompletion(diffCmd)
	registerFilterCompletion(diffCmd)
	rootCmd.AddCommand(diffCmd)
}
