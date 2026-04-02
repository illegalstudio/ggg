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

type repoStatus struct {
	url       string
	cloned    bool
	branch    string
	dirty     bool
	ahead     int
	behind    int
	err       error
	pathErr   bool
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show status of all configured repositories",
	GroupID: GroupInfo,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, err := loadRepos(cmd)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			return nil
		}

		filter, _ := cmd.Flags().GetString("filter")
		repos = filterByName(repos, filter)

		statuses := make([]repoStatus, len(repos))
		var wg sync.WaitGroup

		action := func() {
			for i, r := range repos {
				wg.Add(1)
				go func(idx int, r config.Repo) {
					defer wg.Done()
					s := repoStatus{url: r.URL}

					fullPath, err := repo.FullPath(cfg.BaseDir, r)
					if err != nil {
						s.pathErr = true
						statuses[idx] = s
						return
					}

					if !repo.IsCloned(fullPath) {
						statuses[idx] = s
						return
					}
					s.cloned = true
					s.branch, _ = repo.CurrentBranch(fullPath)
					s.dirty, _ = repo.IsDirty(fullPath)
					s.ahead, s.behind, _ = repo.AheadBehind(fullPath)
					statuses[idx] = s
				}(i, r)
			}
			wg.Wait()
		}

		if err := spinner.New().Title("Fetching status...").Action(action).Run(); err != nil {
			return err
		}

		fmt.Println(ui.Title.Render("Repository Status"))
		fmt.Println()
		for _, s := range statuses {
			if s.pathErr {
				fmt.Printf("  %s %s\n", ui.Error.Render("✗"), ui.Error.Render(s.url+" (invalid URL)"))
				continue
			}
			if !s.cloned {
				fmt.Printf("  %s %s %s\n", ui.Muted.Render("○"), ui.Repo.Render(s.url), ui.Muted.Render("(not cloned)"))
				continue
			}

			status := ui.Success.Render("clean")
			if s.dirty {
				status = ui.Error.Render("dirty")
			}

			var syncStr string
			if s.ahead > 0 && s.behind > 0 {
				syncStr = ui.Error.Render(fmt.Sprintf(" ↑%d ↓%d", s.ahead, s.behind))
			} else if s.ahead > 0 {
				syncStr = ui.Info.Render(fmt.Sprintf(" ↑%d", s.ahead))
			} else if s.behind > 0 {
				syncStr = ui.Error.Render(fmt.Sprintf(" ↓%d", s.behind))
			}

			fmt.Printf("  %s %s [%s] %s%s\n", ui.Success.Render("✓"), ui.Repo.Render(s.url), ui.Info.Render(s.branch), status, syncStr)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	statusCmd.Flags().StringP("group", "g", "", "Show only repos in this group")
	statusCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	rootCmd.AddCommand(statusCmd)
}
