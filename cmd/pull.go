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

var pullCmd = &cobra.Command{
	Use:     "pull [name]",
	Short:   "Pull latest changes (all or a specific repo, only if clean)",
	GroupID: GroupRepo,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, err := loadRepos(cmd)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		if len(args) > 0 {
			filtered, err := filterRepo(repos, args[0])
			if err != nil {
				return err
			}
			repos = filtered
		} else {
			ok, err := confirmAll(fmt.Sprintf("Pull %d repositories?", len(repos)), "Yes, pull all")
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		type pullJob struct {
			repo     config.Repo
			fullPath string
			strategy string
		}
		type result struct {
			url string
			err error
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
			fmt.Println(ui.Info.Render("No repositories to pull (all clean, dirty, or not cloned)."))
			return nil
		}

		// Pull in parallel with spinner
		results := make([]result, len(jobs))
		var wg sync.WaitGroup

		action := func() {
			for i, j := range jobs {
				wg.Add(1)
				go func(idx int, job pullJob) {
					defer wg.Done()
					err := repo.Pull(job.fullPath, job.strategy)
					results[idx] = result{url: job.repo.URL, err: err}
				}(i, j)
			}
			wg.Wait()
		}

		title := fmt.Sprintf("Pulling %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Pulling %s...", jobs[0].repo.URL)
		}

		if err := spinner.New().Title(title).Action(action).Run(); err != nil {
			return err
		}

		// Print results
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
	pullCmd.Flags().StringP("group", "g", "", "Pull only repos in this group")
	rootCmd.AddCommand(pullCmd)
}
