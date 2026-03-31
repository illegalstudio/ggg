package cmd

import (
	"fmt"
	"strings"
	"sync"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var cloneCmd = &cobra.Command{
	Use:   "clone [name]",
	Short: "Clone repositories (all or a specific one)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if len(cfg.Repos) == 0 {
			fmt.Println(ui.Info.Render("No repositories configured."))
			return nil
		}

		group, _ := cmd.Flags().GetString("group")
		repos := filterByGroup(cfg.Repos, group)

		if len(repos) == 0 {
			fmt.Println(ui.Info.Render("No repositories match the criteria."))
			return nil
		}

		if len(args) == 0 {
			var confirm bool
			err := huh.NewConfirm().
				Title(fmt.Sprintf("Clone all %d repositories?", len(repos))).
				Value(&confirm).
				Run()
			if err != nil {
				return err
			}
			if !confirm {
				fmt.Println(ui.Muted.Render("Aborted."))
				return nil
			}
		}
		if len(args) > 0 {
			filtered, err := filterRepo(cfg.Repos, args[0])
			if err != nil {
				return err
			}
			repos = filtered
		}

		// Separate already cloned from pending
		type cloneJob struct {
			repo     config.Repo
			fullPath string
		}
		var jobs []cloneJob
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				fmt.Println(ui.Error.Render(fmt.Sprintf("  ✗ Skipping %s: %v", r.URL, err)))
				continue
			}
			if repo.IsCloned(fullPath) {
				fmt.Printf("  %s %s\n", ui.Muted.Render("●"), ui.Muted.Render("Already cloned: "+fullPath))
				continue
			}
			jobs = append(jobs, cloneJob{repo: r, fullPath: fullPath})
		}

		if len(jobs) == 0 {
			return nil
		}

		// Clone in parallel with spinner
		type result struct {
			url  string
			path string
			err  error
		}
		results := make([]result, len(jobs))
		var wg sync.WaitGroup

		action := func() {
			for i, j := range jobs {
				wg.Add(1)
				go func(idx int, job cloneJob) {
					defer wg.Done()
					results[idx] = result{
						url:  job.repo.URL,
						path: job.fullPath,
						err:  repo.Clone(job.repo.URL, job.fullPath),
					}
				}(i, j)
			}
			wg.Wait()
		}

		title := fmt.Sprintf("Cloning %d repositories...", len(jobs))
		if len(jobs) == 1 {
			title = fmt.Sprintf("Cloning %s...", jobs[0].repo.URL)
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
				fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.url), ui.Path.Render(r.path))
			}
		}
		return nil
	},
}

func filterRepo(repos []config.Repo, name string) ([]config.Repo, error) {
	// First pass: exact match
	for _, r := range repos {
		derived, _ := repo.DerivePathFromURL(r.URL)
		if r.Path == name || r.URL == name || derived == name {
			return []config.Repo{r}, nil
		}
	}

	// Second pass: partial match (substring, case-insensitive)
	nameLower := strings.ToLower(name)
	var matches []config.Repo
	for _, r := range repos {
		derived, _ := repo.DerivePathFromURL(r.URL)
		if strings.Contains(strings.ToLower(r.URL), nameLower) ||
			strings.Contains(strings.ToLower(r.Path), nameLower) ||
			strings.Contains(strings.ToLower(derived), nameLower) {
			matches = append(matches, r)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("repository %q not found in config", name)
	}
	if len(matches) == 1 {
		return matches, nil
	}

	// Multiple matches: prompt user to choose via huh select
	options := make([]huh.Option[int], len(matches))
	for i, r := range matches {
		options[i] = huh.NewOption(r.URL, i)
	}

	var choice int
	err := huh.NewSelect[int]().
		Title(fmt.Sprintf("Multiple repositories match %q", name)).
		Options(options...).
		Value(&choice).
		Run()
	if err != nil {
		return nil, err
	}

	return []config.Repo{matches[choice]}, nil
}

func init() {
	cloneCmd.Flags().StringP("group", "g", "", "Clone only repos in this group")
	rootCmd.AddCommand(cloneCmd)
}
