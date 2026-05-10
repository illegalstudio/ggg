package cli

import (
	"fmt"

	"go-git-get/internal/config"
	"go-git-get/internal/repo"
	"go-git-get/internal/ui"

	"github.com/spf13/cobra"
)

type repoStatus struct {
	URL     string `json:"url"`
	Cloned  bool   `json:"cloned"`
	Branch  string `json:"branch,omitempty"`
	Dirty   bool   `json:"dirty"`
	Ahead   int    `json:"ahead"`
	Behind  int    `json:"behind"`
	Err     error  `json:"-"`
	PathErr bool   `json:"-"`
	Error   string `json:"error,omitempty"`
}

var statusCmd = &cobra.Command{
	Use:     "status [filter]",
	Short:   "Show status of all configured repositories",
	GroupID: GroupInfo,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, _, err := resolveBulkRepos(cmd, args)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			if done, err := maybeJSON(map[string]any{"repos": []repoStatus{}}); done {
				return err
			}
			return nil
		}

		statuses, err := runParallelWithSpinner(repos, "Fetching status...", func(r config.Repo) repoStatus {
			s := repoStatus{URL: r.URL}

			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				s.PathErr = true
				s.Error = err.Error()
				return s
			}

			if !repo.IsCloned(fullPath) {
				return s
			}
			s.Cloned = true
			s.Branch, _ = repo.CurrentBranch(fullPath)
			s.Dirty, _ = repo.IsDirty(fullPath)
			s.Ahead, s.Behind, _ = repo.AheadBehind(fullPath)
			return s
		})
		if err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"repos": statuses}); done {
			return err
		}

		fmt.Println(ui.Title.Render("Repository Status"))
		fmt.Println()
		for _, s := range statuses {
			if s.PathErr {
				fmt.Printf("  %s %s\n", ui.Error.Render("✗"), ui.Error.Render(s.URL+" (invalid URL)"))
				continue
			}
			if !s.Cloned {
				fmt.Printf("  %s %s %s\n", ui.Muted.Render("○"), ui.Repo.Render(s.URL), ui.Muted.Render("(not cloned)"))
				continue
			}

			status := ui.Success.Render("clean")
			if s.Dirty {
				status = ui.Error.Render("dirty")
			}

			var syncStr string
			if s.Ahead > 0 && s.Behind > 0 {
				syncStr = ui.Error.Render(fmt.Sprintf(" ↑%d ↓%d", s.Ahead, s.Behind))
			} else if s.Ahead > 0 {
				syncStr = ui.Info.Render(fmt.Sprintf(" ↑%d", s.Ahead))
			} else if s.Behind > 0 {
				syncStr = ui.Error.Render(fmt.Sprintf(" ↓%d", s.Behind))
			}

			fmt.Printf("  %s %s [%s] %s%s\n", ui.Success.Render("✓"), ui.Repo.Render(s.URL), ui.Info.Render(s.Branch), status, syncStr)
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
