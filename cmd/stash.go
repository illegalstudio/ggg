package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var stashCmd = &cobra.Command{
	Use:     "stash [filter]",
	Short:   "Stash changes in dirty repositories",
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

		type result struct {
			url string
			err error
		}

		title := fmt.Sprintf("Stashing %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Stashing %s...", jobs[0].repo.URL)
		}

		results, err := runParallelWithSpinner(jobs, title, func(job stashJob) result {
			return result{
				url: job.repo.URL,
				err: repo.Stash(job.fullPath),
			}
		})
		if err != nil {
			return err
		}

		fmt.Println()
		for _, r := range results {
			if r.err != nil {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.url), ui.Error.Render(r.err.Error()))
			} else {
				fmt.Printf("  %s %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.url))
			}
		}
		return nil
	},
}

func init() {
	stashCmd.Flags().StringP("group", "g", "", "Stash only repos in this group")
	stashCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	rootCmd.AddCommand(stashCmd)
}
