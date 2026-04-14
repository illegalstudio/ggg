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

var stashCmd = &cobra.Command{
	Use:     "stash [name]",
	Short:   "Stash changes in dirty repositories",
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
			ok, err := confirmAll(fmt.Sprintf("Stash changes in all %d repositories?", len(repos)), "Yes, stash all")
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
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

		type result struct {
			url string
			err error
		}
		results := make([]result, len(jobs))
		var wg sync.WaitGroup

		action := func() {
			for i, j := range jobs {
				wg.Add(1)
				go func(idx int, job stashJob) {
					defer wg.Done()
					results[idx] = result{
						url: job.repo.URL,
						err: repo.Stash(job.fullPath),
					}
				}(i, j)
			}
			wg.Wait()
		}

		title := fmt.Sprintf("Stashing %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Stashing %s...", jobs[0].repo.URL)
		}

		if err := spinner.New().Title(title).Action(action).Run(); err != nil {
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
	rootCmd.AddCommand(stashCmd)
}
