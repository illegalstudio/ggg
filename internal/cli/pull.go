package cli

import (
	"fmt"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:               "pull [filter]",
	Short:             "Pull latest changes (all or filtered repos, only if clean)",
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

		type pullJob struct {
			repo     config.Repo
			fullPath string
			strategy string
		}

		var jobs []pullJob

		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				continue
			}
			if !repo.IsCloned(fullPath) {
				continue
			}
			dirty, err := repo.IsDirty(fullPath)
			if err != nil {
				continue
			}
			if dirty {
				continue
			}
			strategy := string(cfg.ResolvePullStrategy(r))
			jobs = append(jobs, pullJob{repo: r, fullPath: fullPath, strategy: strategy})
		}

		if len(jobs) == 0 {
			if done, err := maybeJSON(map[string]any{"results": []result{}}); done {
				return err
			}
			fmt.Println(ui.Info.Render("No repositories to pull (all clean, dirty, or not cloned)."))
			return nil
		}

		ok, err := confirmBulkAction(filter, len(jobs), "Pull %d repositories?", "Yes, pull all")
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		title := fmt.Sprintf("Pulling %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Pulling %s...", jobs[0].repo.URL)
		}

		results, err := runParallelWithSpinner(jobs, title, func(job pullJob) result {
			return result{URL: job.repo.URL, Error: errString(repo.Pull(job.fullPath, job.strategy))}
		})
		if err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"results": results}); done {
			return err
		}

		// Print results
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
	pullCmd.Flags().StringArrayP("group", "g", nil, "Pull only repos in this group")
	pullCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	registerGroupCompletion(pullCmd)
	registerFilterCompletion(pullCmd)
	rootCmd.AddCommand(pullCmd)
}
