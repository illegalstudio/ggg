package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:     "clone [filter]",
	Short:   "Clone repositories (all or filtered)",
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

		// Separate already cloned from pending
		type cloneJob struct {
			repo     config.Repo
			fullPath string
		}
		var jobs []cloneJob
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				fmt.Println(ui.Error.Render(fmt.Sprintf("  ✗ Skipping %s: %v", r.URL, err)))
				continue
			}
			if repo.IsCloned(fullPath) {
				fmt.Printf("  %s %s\n", ui.Muted.Render("●"), ui.Muted.Render("Already cloned: "+fullPath))
				continue
			}
			jobs = append(jobs, cloneJob{repo: r, fullPath: fullPath})
		}

		if len(jobs) == 0 {
			return nil
		}

		ok, err := confirmBulkAction(filter, len(jobs), "Clone %d repositories?", "Yes, clone all")
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		type result struct {
			url  string
			path string
			err  error
		}

		title := fmt.Sprintf("Cloning %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Cloning %s...", jobs[0].repo.URL)
		}

		results, err := runParallelWithSpinner(jobs, title, func(job cloneJob) result {
			return result{
				url:  job.repo.URL,
				path: job.fullPath,
				err:  repo.Clone(job.repo.URL, job.fullPath),
			}
		})
		if err != nil {
			return err
		}

		// Print results
		fmt.Println()
		for _, r := range results {
			if r.err != nil {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.url), ui.Error.Render(r.err.Error()))
			} else {
				fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.url), ui.Path.Render(r.path))
			}
		}
		return nil
	},
}

func init() {
	cloneCmd.Flags().StringP("group", "g", "", "Clone only repos in this group")
	cloneCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	rootCmd.AddCommand(cloneCmd)
}
