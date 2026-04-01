package cmd

import (
	"fmt"
	"sync"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var checkoutCmd = &cobra.Command{
	Use:     "checkout <branch> [name]",
	Short:   "Checkout a branch in repositories that have it",
	GroupID: GroupRepo,
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]

		cfg, repos, err := loadRepos(cmd)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		if len(args) > 1 {
			filtered, err := filterRepo(repos, args[1])
			if err != nil {
				return err
			}
			repos = filtered
		} else {
			ok, err := confirmAll(fmt.Sprintf("Checkout branch %q in all %d repositories?", branch, len(repos)), "Yes, checkout all")
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		type checkoutJob struct {
			repo     config.Repo
			fullPath string
		}
		var jobs []checkoutJob

		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				continue
			}
			if !repo.IsCloned(fullPath) {
				fmt.Printf("  %s %s %s\n", ui.Muted.Render("○"), ui.Muted.Render(r.URL), ui.Muted.Render("(not cloned)"))
				continue
			}
			if !repo.HasBranch(fullPath, branch) {
				fmt.Printf("  %s %s %s\n", ui.Muted.Render("○"), ui.Muted.Render(r.URL), ui.Muted.Render(fmt.Sprintf("(no branch %q)", branch)))
				continue
			}
			jobs = append(jobs, checkoutJob{repo: r, fullPath: fullPath})
		}

		if len(jobs) == 0 {
			fmt.Println(ui.Info.Render(fmt.Sprintf("No repositories have branch %q.", branch)))
			return nil
		}

		type result struct {
			url string
			err error
		}
		results := make([]result, len(jobs))
		var wg sync.WaitGroup

		action := func() {
			for i, j := range jobs {
				wg.Add(1)
				go func(idx int, job checkoutJob) {
					defer wg.Done()
					results[idx] = result{
						url: job.repo.URL,
						err: repo.Checkout(job.fullPath, branch),
					}
				}(i, j)
			}
			wg.Wait()
		}

		title := fmt.Sprintf("Checking out %q in %d repositories...", branch, len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Checking out %q in %s...", branch, jobs[0].repo.URL)
		}

		if err := spinner.New().Title(title).Action(action).Run(); err != nil {
			return err
		}

		fmt.Println()
		for _, r := range results {
			if r.err != nil {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.url), ui.Error.Render(r.err.Error()))
			} else {
				fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.url), ui.Info.Render(branch))
			}
		}
		return nil
	},
}

func init() {
	checkoutCmd.Flags().StringP("group", "g", "", "Checkout only in repos in this group")
	rootCmd.AddCommand(checkoutCmd)
}
