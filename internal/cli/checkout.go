package cli

import (
	"fmt"

	"go-git-get/internal/config"
	"go-git-get/internal/repo"
	"go-git-get/internal/ui"

	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:     "checkout <branch> [filter]",
	Short:   "Checkout a branch in repositories that have it",
	GroupID: GroupRepo,
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]

		cfg, repos, err := loadRepos(cmd)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			if done, err := maybeJSON(map[string]any{"branch": branch, "results": []any{}}); done {
				return err
			}
			return nil
		}

		filter := getFilter(cmd, args[1:])
		repos = filterByName(repos, filter)

		type checkoutJob struct {
			repo     config.Repo
			fullPath string
		}
		type result struct {
			URL    string `json:"url"`
			Branch string `json:"branch"`
			Error  string `json:"error,omitempty"`
		}

		var jobs []checkoutJob

		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				continue
			}
			if !repo.IsCloned(fullPath) {
				continue
			}
			if !repo.HasBranch(fullPath, branch) {
				continue
			}
			jobs = append(jobs, checkoutJob{repo: r, fullPath: fullPath})
		}

		if len(jobs) == 0 {
			if done, err := maybeJSON(map[string]any{"branch": branch, "results": []result{}}); done {
				return err
			}
			fmt.Println(ui.Info.Render(fmt.Sprintf("No repositories have branch %q.", branch)))
			return nil
		}

		ok, err := confirmBulkAction(filter, len(jobs), fmt.Sprintf("Checkout branch %q in %%d repositories?", branch), "Yes, checkout all")
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		title := fmt.Sprintf("Checking out %q in %d repositories...", branch, len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Checking out %q in %s...", branch, jobs[0].repo.URL)
		}

		results, err := runParallelWithSpinner(jobs, title, func(job checkoutJob) result {
			return result{
				URL:    job.repo.URL,
				Branch: branch,
				Error:  errString(repo.Checkout(job.fullPath, branch)),
			}
		})
		if err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"branch": branch, "results": results}); done {
			return err
		}

		fmt.Println()
		for _, r := range results {
			if r.Error != "" {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.URL), ui.Error.Render(r.Error))
			} else {
				fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.URL), ui.Info.Render(branch))
			}
		}
		return nil
	},
}

func init() {
	checkoutCmd.Flags().StringP("group", "g", "", "Checkout only in repos in this group")
	checkoutCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	rootCmd.AddCommand(checkoutCmd)
}
