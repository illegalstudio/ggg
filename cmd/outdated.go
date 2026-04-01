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

type outdatedResult struct {
	url    string
	branch string
	behind int
	err    error
	skip   string // "not_cloned", "fetch_err", ""
}

var outdatedCmd = &cobra.Command{
	Use:     "outdated",
	Short:   "Show repositories that are behind their remote",
	GroupID: GroupInfo,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, err := loadRepos(cmd)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		results := make([]outdatedResult, len(repos))
		var wg sync.WaitGroup

		action := func() {
			for i, r := range repos {
				wg.Add(1)
				go func(idx int, r config.Repo) {
					defer wg.Done()
					res := outdatedResult{url: r.URL}

					fullPath, err := repo.FullPath(cfg.BaseDir, r)
					if err != nil {
						res.err = err
						results[idx] = res
						return
					}

					if !repo.IsCloned(fullPath) {
						res.skip = "not_cloned"
						results[idx] = res
						return
					}

					if err := repo.Fetch(fullPath); err != nil {
						res.skip = "fetch_err"
						res.err = err
						results[idx] = res
						return
					}

					res.branch, _ = repo.CurrentBranch(fullPath)
					_, res.behind, _ = repo.AheadBehind(fullPath)
					results[idx] = res
				}(i, r)
			}
			wg.Wait()
		}

		if err := spinner.New().Title("Fetching from remotes...").Action(action).Run(); err != nil {
			return err
		}

		outdated := 0
		for _, r := range results {
			if r.behind > 0 {
				outdated++
			}
		}

		if outdated == 0 {
			fmt.Println(ui.Success.Render("All repositories are up to date."))
			return nil
		}

		fmt.Println(ui.Title.Render("Outdated Repositories"))
		fmt.Println()
		for _, r := range results {
			if r.skip == "not_cloned" {
				continue
			}
			if r.skip == "fetch_err" {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.url), ui.Muted.Render("(fetch failed)"))
				continue
			}
			if r.behind > 0 {
				fmt.Printf("  %s %s [%s] %s\n",
					ui.Error.Render("↓"),
					ui.Repo.Render(r.url),
					ui.Info.Render(r.branch),
					ui.Error.Render(fmt.Sprintf("%d commits behind", r.behind)),
				)
			}
		}
		fmt.Println()
		return nil
	},
}

func init() {
	outdatedCmd.Flags().StringP("group", "g", "", "Check only repos in this group")
	rootCmd.AddCommand(outdatedCmd)
}
