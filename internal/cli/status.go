package cli

import (
	"fmt"
	"path/filepath"

	"go-git-get/internal/config"
	"go-git-get/internal/repo"
	"go-git-get/internal/ui"

	"github.com/spf13/cobra"
)

type worktreeStat struct {
	Path   string `json:"path"`
	Branch string `json:"branch,omitempty"`
	Dirty  bool   `json:"dirty"`
	Ahead  int    `json:"ahead"`
	Behind int    `json:"behind"`
}

type repoStatus struct {
	URL       string         `json:"url"`
	Cloned    bool           `json:"cloned"`
	Branch    string         `json:"branch,omitempty"`
	Dirty     bool           `json:"dirty"`
	Ahead     int            `json:"ahead"`
	Behind    int            `json:"behind"`
	Worktrees []worktreeStat `json:"worktrees,omitempty"`
	Err       error          `json:"-"`
	PathErr   bool           `json:"-"`
	Error     string         `json:"error,omitempty"`
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

			wts, _ := repo.Worktrees(fullPath)
			for _, wt := range wts {
				ws := worktreeStat{Path: wt.Path, Branch: wt.Branch}
				ws.Dirty, _ = repo.IsDirty(wt.Path)
				ws.Ahead, ws.Behind, _ = repo.AheadBehind(wt.Path)
				s.Worktrees = append(s.Worktrees, ws)
			}
			return s
		})
		if err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"repos": statuses}); done {
			return err
		}

		detailed, _ := cmd.Flags().GetBool("detailed")

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

			fmt.Printf("  %s %s [%s] %s%s%s\n",
				ui.Success.Render("✓"),
				ui.Repo.Render(s.URL),
				ui.Info.Render(s.Branch),
				status,
				renderSync(s.Ahead, s.Behind),
				renderWorktreeBadge(len(s.Worktrees)),
			)

			if detailed {
				renderWorktreeTree(s.Worktrees)
			}
		}
		fmt.Println()
		return nil
	},
}

func renderSync(ahead, behind int) string {
	switch {
	case ahead > 0 && behind > 0:
		return ui.Error.Render(fmt.Sprintf(" ↑%d ↓%d", ahead, behind))
	case ahead > 0:
		return ui.Info.Render(fmt.Sprintf(" ↑%d", ahead))
	case behind > 0:
		return ui.Error.Render(fmt.Sprintf(" ↓%d", behind))
	}
	return ""
}

func renderWorktreeBadge(count int) string {
	if count == 0 {
		return ""
	}
	return ui.Info.Render(fmt.Sprintf(" ⎇%d", count))
}

func renderWorktreeTree(wts []worktreeStat) {
	for i, wt := range wts {
		prefix := "├─"
		if i == len(wts)-1 {
			prefix = "└─"
		}
		status := ui.Success.Render("clean")
		if wt.Dirty {
			status = ui.Error.Render("dirty")
		}
		branch := wt.Branch
		if branch == "" {
			branch = "(detached)"
		}
		fmt.Printf("    %s %s [%s] %s%s\n",
			ui.Muted.Render(prefix),
			ui.Path.Render(filepath.Base(wt.Path)),
			ui.Info.Render(branch),
			status,
			renderSync(wt.Ahead, wt.Behind),
		)
	}
}

func init() {
	statusCmd.Flags().StringP("group", "g", "", "Show only repos in this group")
	statusCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	statusCmd.Flags().BoolP("detailed", "d", false, "Show linked worktrees as a tree under each repo")
	rootCmd.AddCommand(statusCmd)
}
