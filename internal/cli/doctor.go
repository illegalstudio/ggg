package cli

import (
	"fmt"
	"os"
	"sync"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"
	"github.com/illegalstudio/ggg/internal/ui"
	gggskills "github.com/illegalstudio/ggg/skills"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

type checkResult struct {
	Label   string `json:"label"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
	Warn    bool   `json:"warn,omitempty"`
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
			checks = append(checks, skillChecks()...)
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
			results = append(results, checkResult{Label: "Config file", OK: false, Message: "not found at " + configPath})
			return emit(results, nil)
		}
		results = append(results, checkResult{Label: "Config file", OK: true, Message: configPath})

		// Check 2: config is parseable
		cfg, err := config.Load()
		if err != nil {
			results = append(results, checkResult{Label: "Config syntax", OK: false, Message: err.Error()})
			return emit(results, nil)
		}
		results = append(results, checkResult{Label: "Config syntax", OK: true, Message: "valid"})

		// Check 3: base_dir exists
		if _, err := os.Stat(cfg.BaseDir); os.IsNotExist(err) {
			results = append(results, checkResult{Label: "Base directory", OK: false, Message: cfg.BaseDir + " does not exist"})
		} else {
			results = append(results, checkResult{Label: "Base directory", OK: true, Message: cfg.BaseDir})
		}

		// Check 4: repos count
		results = append(results, checkResult{Label: "Repositories", OK: true, Message: fmt.Sprintf("%d configured", len(cfg.Repos))})

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
			results = append(results, checkResult{Label: "Duplicates", OK: false, Message: fmt.Sprintf("%d duplicate URLs found", dupes)})
		} else {
			results = append(results, checkResult{Label: "Duplicates", OK: true, Message: "none"})
		}

		// Check 6: clone status
		cloned, notCloned := 0, 0
		for _, r := range cfg.Repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, cfg.Aliases, r)
			if err != nil {
				continue
			}
			if repo.IsCloned(fullPath) {
				cloned++
			} else {
				notCloned++
			}
		}
		results = append(results, checkResult{Label: "Cloned", OK: true, Message: fmt.Sprintf("%d cloned, %d missing", cloned, notCloned)})

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
			results = append(results, checkResult{Label: "Remotes", OK: false, Message: fmt.Sprintf("%d unreachable", unreachable)})
		} else {
			results = append(results, checkResult{Label: "Remotes", OK: true, Message: "all reachable"})
		}

		return emit(results, remoteResults)
	},
}

func printResults(results []checkResult) {
	fmt.Println(ui.Title.Render("Doctor"))
	fmt.Println()
	for _, r := range results {
		icon := ui.Success.Render("✓")
		switch {
		case r.Warn:
			icon = ui.Warning.Render("⚠")
		case !r.OK:
			icon = ui.Error.Render("✗")
		}
		fmt.Printf("  %s %s %s\n", icon, ui.Repo.Render(r.Label+":"), ui.Muted.Render(r.Message))
	}
}

// skillChecks reports one row per *existing* skill installation. Destinations
// that were never installed produce no row, so doctor stays quiet for users who
// do not use AI agents.
func skillChecks() []checkResult {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var results []checkResult
	for _, target := range skillTargets(home) {
		state, err := gggskills.Inspect(target.Path)
		if err != nil {
			results = append(results, checkResult{
				Label:   "AI agent skill",
				Message: fmt.Sprintf("%s: %v", target.Label, err),
			})
			continue
		}

		switch state {
		case gggskills.StateMissing:
			continue
		case gggskills.StateCurrent:
			results = append(results, checkResult{
				Label:   "AI agent skill",
				OK:      true,
				Message: target.Label + ": up to date",
			})
		case gggskills.StateOutdated:
			results = append(results, checkResult{
				Label:   "AI agent skill",
				Warn:    true,
				Message: target.Label + ": outdated — run: ggg skills install",
			})
		case gggskills.StateModified:
			results = append(results, checkResult{
				Label:   "AI agent skill",
				Warn:    true,
				Message: target.Label + ": locally modified — run: ggg skills install --force",
			})
		default:
			results = append(results, checkResult{
				Label:   "AI agent skill",
				Warn:    true,
				Message: target.Label + ": unmanaged directory — run: ggg skills install --force",
			})
		}
	}
	return results
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
