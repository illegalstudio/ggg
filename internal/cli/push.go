package cli

import (
	"fmt"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:               "push [filter]",
	Short:             "Push commits to remote for repositories that are ahead",
	GroupID:           GroupRepo,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: repoCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		type result struct {
			URL   string `json:"url"`
			Ahead int    `json:"ahead"`
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

		type pushJob struct {
			repo     config.Repo
			fullPath string
			ahead    int
		}

		var jobs []pushJob

		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				continue
			}
			if !repo.IsCloned(fullPath) {
				continue
			}
			ahead, _, err := repo.AheadBehind(fullPath)
			if err != nil {
				continue
			}
			if ahead == 0 {
				continue
			}
			jobs = append(jobs, pushJob{repo: r, fullPath: fullPath, ahead: ahead})
		}

		if len(jobs) == 0 {
			if done, err := maybeJSON(map[string]any{"results": []result{}}); done {
				return err
			}
			fmt.Println(ui.Info.Render("All repositories are up to date — nothing to push."))
			return nil
		}

		// Show what will be pushed (skipped in JSON mode)
		if !jsonOutput {
			for _, j := range jobs {
				fmt.Printf("  %s %s %s\n", ui.Info.Render("↑"), ui.Repo.Render(j.repo.URL), ui.Muted.Render(fmt.Sprintf("(%d commits ahead)", j.ahead)))
			}
			fmt.Println()
		}

		ok, err := confirmBulkAction(filter, len(jobs), "Push %d repositories?", "Yes, push all")
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		title := fmt.Sprintf("Pushing %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Pushing %s...", jobs[0].repo.URL)
		}

		results, err := runParallelWithSpinner(jobs, title, func(job pushJob) result {
			return result{
				URL:   job.repo.URL,
				Ahead: job.ahead,
				Error: errString(repo.Push(job.fullPath)),
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
	pushCmd.Flags().StringP("group", "g", "", "Push only repos in this group")
	pushCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	registerGroupCompletion(pushCmd)
	registerFilterCompletion(pushCmd)
	rootCmd.AddCommand(pushCmd)
}
