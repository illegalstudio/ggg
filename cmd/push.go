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

var pushCmd = &cobra.Command{
	Use:     "push [filter]",
	Short:   "Push commits to remote for repositories that are ahead",
	GroupID: GroupRepo,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, err := loadRepos(cmd)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		filter := getFilter(cmd, args)
		repos = filterByName(repos, filter)

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
			fmt.Println(ui.Info.Render("All repositories are up to date — nothing to push."))
			return nil
		}

		// Show what will be pushed
		for _, j := range jobs {
			fmt.Printf("  %s %s %s\n", ui.Info.Render("↑"), ui.Repo.Render(j.repo.URL), ui.Muted.Render(fmt.Sprintf("(%d commits ahead)", j.ahead)))
		}
		fmt.Println()

		if filter == "" {
			ok, err := confirmAll(fmt.Sprintf("Push %d repositories?", len(jobs)), "Yes, push all")
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
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
				go func(idx int, job pushJob) {
					defer wg.Done()
					results[idx] = result{
						url: job.repo.URL,
						err: repo.Push(job.fullPath),
					}
				}(i, j)
			}
			wg.Wait()
		}

		title := fmt.Sprintf("Pushing %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Pushing %s...", jobs[0].repo.URL)
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
	pushCmd.Flags().StringP("group", "g", "", "Push only repos in this group")
	pushCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	rootCmd.AddCommand(pushCmd)
}
