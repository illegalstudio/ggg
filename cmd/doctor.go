package cmd

import (
	"fmt"
	"os"
	"sync"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

type checkResult struct {
	Label   string `json:"label"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

var doctorCmd = &cobra.Command{
	Use:     "doctor",
	Short:   "Run health checks on configuration and repositories",
	GroupID: GroupInfo,
	RunE: func(cmd *cobra.Command, args []string) error {
		var results []checkResult

		type remoteCheck struct {
			URL       string `json:"url"`
			Reachable bool   `json:"reachable"`
		}

		emit := func(checks []checkResult, remotes []remoteCheck) error {
			if done, err := maybeJSON(map[string]any{"checks": checks, "remotes": remotes}); done {
				return err
			}
			printResults(checks)
			unreachable := 0
			for _, rc := range remotes {
				if !rc.Reachable {
					unreachable++
				}
			}
			if unreachable > 0 {
				fmt.Println()
				fmt.Println(ui.Muted.Render("  Unreachable remotes:"))
				for _, rc := range remotes {
					if !rc.Reachable {
						fmt.Printf("    %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(rc.URL))
					}
				}
			}
			fmt.Println()
			return nil
		}

		// Check 1: config file exists
		configPath := config.ConfigPath()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			results = append(results, checkResult{"Config file", false, "not found at " + configPath})
			return emit(results, nil)
		}
		results = append(results, checkResult{"Config file", true, configPath})

		// Check 2: config is parseable
		cfg, err := config.Load()
		if err != nil {
			results = append(results, checkResult{"Config syntax", false, err.Error()})
			return emit(results, nil)
		}
		results = append(results, checkResult{"Config syntax", true, "valid"})

		// Check 3: base_dir exists
		if _, err := os.Stat(cfg.BaseDir); os.IsNotExist(err) {
			results = append(results, checkResult{"Base directory", false, cfg.BaseDir + " does not exist"})
		} else {
			results = append(results, checkResult{"Base directory", true, cfg.BaseDir})
		}

		// Check 4: repos count
		results = append(results, checkResult{"Repositories", true, fmt.Sprintf("%d configured", len(cfg.Repos))})

		// Check 5: duplicate URLs
		seen := make(map[string]bool)
		dupes := 0
		for _, r := range cfg.Repos {
			if seen[r.URL] {
				dupes++
			}
			seen[r.URL] = true
		}
		if dupes > 0 {
			results = append(results, checkResult{"Duplicates", false, fmt.Sprintf("%d duplicate URLs found", dupes)})
		} else {
			results = append(results, checkResult{"Duplicates", true, "none"})
		}

		// Check 6: clone status
		cloned, notCloned := 0, 0
		for _, r := range cfg.Repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				continue
			}
			if repo.IsCloned(fullPath) {
				cloned++
			} else {
				notCloned++
			}
		}
		results = append(results, checkResult{"Cloned", true, fmt.Sprintf("%d cloned, %d missing", cloned, notCloned)})

		// Check 7: remote reachability (parallel)
		remoteResults := make([]remoteCheck, len(cfg.Repos))
		var wg sync.WaitGroup

		action := func() {
			for i, r := range cfg.Repos {
				wg.Add(1)
				go func(idx int, url string) {
					defer wg.Done()
					remoteResults[idx] = remoteCheck{URL: url, Reachable: repo.RemoteReachable(url)}
				}(i, r.URL)
			}
			wg.Wait()
		}

		if jsonOutput {
			action()
		} else {
			if err := spinner.New().Title("Checking remotes...").Action(action).Run(); err != nil {
				return err
			}
		}

		unreachable := 0
		for _, rc := range remoteResults {
			if !rc.Reachable {
				unreachable++
			}
		}
		if unreachable > 0 {
			results = append(results, checkResult{"Remotes", false, fmt.Sprintf("%d unreachable", unreachable)})
		} else {
			results = append(results, checkResult{"Remotes", true, "all reachable"})
		}

		return emit(results, remoteResults)
	},
}

func printResults(results []checkResult) {
	fmt.Println(ui.Title.Render("Doctor"))
	fmt.Println()
	for _, r := range results {
		icon := ui.Success.Render("✓")
		if !r.OK {
			icon = ui.Error.Render("✗")
		}
		fmt.Printf("  %s %s %s\n", icon, ui.Repo.Render(r.Label+":"), ui.Muted.Render(r.Message))
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
