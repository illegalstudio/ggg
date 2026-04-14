package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

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
	Use:     "outdated [filter]",
	Short:   "Show repositories that are behind their remote",
	GroupID: GroupInfo,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, _, err := resolveBulkRepos(cmd, args)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		results, err := runParallelWithSpinner(repos, "Fetching from remotes...", func(r config.Repo) outdatedResult {
			res := outdatedResult{url: r.URL}

			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				res.err = err
				return res
			}

			if !repo.IsCloned(fullPath) {
				res.skip = "not_cloned"
				return res
			}

			if err := repo.Fetch(fullPath); err != nil {
				res.skip = "fetch_err"
				res.err = err
				return res
			}

			res.branch, _ = repo.CurrentBranch(fullPath)
			_, res.behind, _ = repo.AheadBehind(fullPath)
			return res
		})
		if err != nil {
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
	outdatedCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	rootCmd.AddCommand(outdatedCmd)
}
