package cli

import (
	"fmt"

	"go-git-get/internal/config"
	"go-git-get/internal/repo"
	"go-git-get/internal/ui"

	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:     "clone [filter]",
	Short:   "Clone repositories (all or filtered)",
	GroupID: GroupRepo,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, repos, filter, err := resolveBulkRepos(cmd, args)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			if done, err := maybeJSON(map[string]any{"results": []any{}, "skipped": []any{}}); done {
				return err
			}
			return nil
		}

		// Separate already cloned from pending
		type cloneJob struct {
			repo     config.Repo
			fullPath string
		}
		type skipped struct {
			URL    string `json:"url"`
			Path   string `json:"path,omitempty"`
			Reason string `json:"reason"`
		}
		var jobs []cloneJob
		var skips []skipped
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				skips = append(skips, skipped{URL: r.URL, Reason: err.Error()})
				if !jsonOutput {
					fmt.Println(ui.Error.Render(fmt.Sprintf("  ✗ Skipping %s: %v", r.URL, err)))
				}
				continue
			}
			if repo.IsCloned(fullPath) {
				skips = append(skips, skipped{URL: r.URL, Path: fullPath, Reason: "already_cloned"})
				if !jsonOutput {
					fmt.Printf("  %s %s\n", ui.Muted.Render("●"), ui.Muted.Render("Already cloned: "+fullPath))
				}
				continue
			}
			jobs = append(jobs, cloneJob{repo: r, fullPath: fullPath})
		}

		type result struct {
			URL   string `json:"url"`
			Path  string `json:"path"`
			Error string `json:"error,omitempty"`
		}

		if len(jobs) == 0 {
			if done, err := maybeJSON(map[string]any{"results": []result{}, "skipped": skips}); done {
				return err
			}
			return nil
		}

		ok, err := confirmBulkAction(filter, len(jobs), "Clone %d repositories?", "Yes, clone all")
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		title := fmt.Sprintf("Cloning %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Cloning %s...", jobs[0].repo.URL)
		}

		results, err := runParallelWithSpinner(jobs, title, func(job cloneJob) result {
			return result{
				URL:   job.repo.URL,
				Path:  job.fullPath,
				Error: errString(repo.Clone(job.repo.URL, job.fullPath)),
			}
		})
		if err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"results": results, "skipped": skips}); done {
			return err
		}

		// Print results
		fmt.Println()
		for _, r := range results {
			if r.Error != "" {
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(r.URL), ui.Error.Render(r.Error))
			} else {
				fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.URL), ui.Path.Render(r.Path))
			}
		}
		return nil
	},
}

func init() {
	cloneCmd.Flags().StringP("group", "g", "", "Clone only repos in this group")
	cloneCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	rootCmd.AddCommand(cloneCmd)
}
