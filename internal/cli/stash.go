package cli

import (
	"fmt"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/spf13/cobra"
)

var stashCmd = &cobra.Command{
	Use:               "stash [filter]",
	Short:             "Stash changes in dirty repositories",
	GroupID:           GroupRepo,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: repoCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		type result struct {
			URL   string `json:"url"`
			Error string `json:"error,omitempty"`
		}

		cfg, repos, filter, err := resolveBulkRepos(cmd, args)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			if done, err := maybeJSON(map[string]any{"results": []result{}}); done {
				return err
			}
			return nil
		}

		type stashJob struct {
			repo     config.Repo
			fullPath string
		}

		var jobs []stashJob

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
			jobs = append(jobs, stashJob{repo: r, fullPath: fullPath})
		}

		if len(jobs) == 0 {
			if done, err := maybeJSON(map[string]any{"results": []result{}}); done {
				return err
			}
			fmt.Println(ui.Info.Render("No dirty repositories to stash."))
			return nil
		}

		ok, err := confirmBulkAction(filter, len(jobs), "Stash changes in %d repositories?", "Yes, stash all")
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		title := fmt.Sprintf("Stashing %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Stashing %s...", jobs[0].repo.URL)
		}

		results, err := runParallelWithSpinner(jobs, title, func(job stashJob) result {
			return result{
				URL:   job.repo.URL,
				Error: errString(repo.Stash(job.fullPath)),
			}
		})
		if err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"results": results}); done {
			return err
		}

		fmt.Println()
		for _, r := range results {
			if r.Error != "" {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.URL), ui.Error.Render(r.Error))
			} else {
				fmt.Printf("  %s %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.URL))
			}
		}
		return nil
	},
}

func init() {
	stashCmd.Flags().StringArrayP("group", "g", nil, "Stash only repos in this group")
	stashCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	registerGroupCompletion(stashCmd)
	registerFilterCompletion(stashCmd)
	rootCmd.AddCommand(stashCmd)
}
