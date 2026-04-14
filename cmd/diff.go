package cmd

import (
	"fmt"

	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:     "diff [filter]",
	Short:   "Show a summary of changed files in dirty repositories",
	GroupID: GroupRepo,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, filter, err := resolveBulkRepos(cmd, args)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		type diffEntry struct {
			url     string
			summary string
			err     error
		}

		var entries []diffEntry
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
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
			entries = append(entries, diffEntry{url: r.URL, summary: summary, err: err})
		}

		if len(entries) == 0 {
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

		for _, entry := range entries {
			if entry.err != nil {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(entry.url), ui.Error.Render(entry.err.Error()))
				continue
			}

			fmt.Printf("\n  %s %s\n", ui.Repo.Render("●"), ui.Repo.Render(entry.url))
			fmt.Printf("%s\n", entry.summary)
		}

		return nil
	},
}

func init() {
	diffCmd.Flags().StringP("group", "g", "", "Show diff only for repos in this group")
	diffCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	rootCmd.AddCommand(diffCmd)
}
