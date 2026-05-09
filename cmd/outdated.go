package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

type outdatedResult struct {
	URL    string `json:"url"`
	Branch string `json:"branch,omitempty"`
	Behind int    `json:"behind"`
	Skip   string `json:"skip,omitempty"` // "not_cloned", "fetch_err", ""
	Error  string `json:"error,omitempty"`
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
			if done, err := maybeJSON(map[string]any{"repos": []outdatedResult{}}); done {
				return err
			}
			return nil
		}

		results, err := runParallelWithSpinner(repos, "Fetching from remotes...", func(r config.Repo) outdatedResult {
			res := outdatedResult{URL: r.URL}

			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				res.Error = err.Error()
				return res
			}

			if !repo.IsCloned(fullPath) {
				res.Skip = "not_cloned"
				return res
			}

			if err := repo.Fetch(fullPath); err != nil {
				res.Skip = "fetch_err"
				res.Error = err.Error()
				return res
			}

			res.Branch, _ = repo.CurrentBranch(fullPath)
			_, res.Behind, _ = repo.AheadBehind(fullPath)
			return res
		})
		if err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"repos": results}); done {
			return err
		}

		outdated := 0
		for _, r := range results {
			if r.Behind > 0 {
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
			if r.Skip == "not_cloned" {
				continue
			}
			if r.Skip == "fetch_err" {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.URL), ui.Muted.Render("(fetch failed)"))
				continue
			}
			if r.Behind > 0 {
				fmt.Printf("  %s %s [%s] %s\n",
					ui.Error.Render("↓"),
					ui.Repo.Render(r.URL),
					ui.Info.Render(r.Branch),
					ui.Error.Render(fmt.Sprintf("%d commits behind", r.Behind)),
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
