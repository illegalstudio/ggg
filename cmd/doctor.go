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
	label   string
	ok      bool
	message string
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run health checks on configuration and repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		var results []checkResult

		// Check 1: config file exists
		configPath := config.ConfigPath()
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			results = append(results, checkResult{"Config file", false, "not found at " + configPath})
			printResults(results)
			return nil
		}
		results = append(results, checkResult{"Config file", true, configPath})

		// Check 2: config is parseable
		cfg, err := config.Load()
		if err != nil {
			results = append(results, checkResult{"Config syntax", false, err.Error()})
			printResults(results)
			return nil
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
		type remoteCheck struct {
			url       string
			reachable bool
		}
		remoteResults := make([]remoteCheck, len(cfg.Repos))
		var wg sync.WaitGroup

		action := func() {
			for i, r := range cfg.Repos {
				wg.Add(1)
				go func(idx int, url string) {
					defer wg.Done()
					remoteResults[idx] = remoteCheck{url: url, reachable: repo.RemoteReachable(url)}
				}(i, r.URL)
			}
			wg.Wait()
		}

		if err := spinner.New().Title("Checking remotes...").Action(action).Run(); err != nil {
			return err
		}

		unreachable := 0
		for _, rc := range remoteResults {
			if !rc.reachable {
				unreachable++
			}
		}
		if unreachable > 0 {
			results = append(results, checkResult{"Remotes", false, fmt.Sprintf("%d unreachable", unreachable)})
		} else {
			results = append(results, checkResult{"Remotes", true, "all reachable"})
		}

		// Print all results
		printResults(results)

		// Print unreachable details
		if unreachable > 0 {
			fmt.Println()
			fmt.Println(ui.Muted.Render("  Unreachable remotes:"))
			for _, rc := range remoteResults {
				if !rc.reachable {
					fmt.Printf("    %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(rc.url))
				}
			}
		}

		fmt.Println()
		return nil
	},
}

func printResults(results []checkResult) {
	fmt.Println(ui.Title.Render("Doctor"))
	fmt.Println()
	for _, r := range results {
		icon := ui.Success.Render("✓")
		if !r.ok {
			icon = ui.Error.Render("✗")
		}
		fmt.Printf("  %s %s %s\n", icon, ui.Repo.Render(r.label+":"), ui.Muted.Render(r.message))
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
