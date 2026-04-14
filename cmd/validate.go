package cmd

import (
	"fmt"
	"strings"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Validate configuration for errors and conflicts",
	GroupID: GroupInfo,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		var issues []string

		// Check 1: duplicate URLs
		urlCount := map[string]int{}
		for _, r := range cfg.Repos {
			urlCount[r.URL]++
		}
		for u, n := range urlCount {
			if n > 1 {
				issues = append(issues, fmt.Sprintf("Duplicate URL: %s (%d times)", u, n))
			}
		}

		// Check 2: conflicting resolved paths
		pathToURLs := map[string][]string{}
		for _, r := range cfg.Repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				issues = append(issues, fmt.Sprintf("Invalid URL: %s", r.URL))
				continue
			}
			pathToURLs[fullPath] = append(pathToURLs[fullPath], r.URL)
		}
		for p, urls := range pathToURLs {
			if len(urls) > 1 {
				issues = append(issues, fmt.Sprintf("Path conflict: %s → %s", p, strings.Join(urls, ", ")))
			}
		}

		// Check 3: invalid pull strategies
		validStrategies := map[config.PullStrategy]bool{
			config.PullMerge: true, config.PullRebase: true, config.PullFFOnly: true, "": true,
		}
		if !validStrategies[cfg.PullStrategy] {
			issues = append(issues, fmt.Sprintf("Invalid global pull_strategy: %q", cfg.PullStrategy))
		}
		for _, r := range cfg.Repos {
			if !validStrategies[r.PullStrategy] {
				issues = append(issues, fmt.Sprintf("Invalid pull_strategy %q for %s", r.PullStrategy, r.URL))
			}
		}

		// Check 4: empty groups (group field set to whitespace)
		for _, r := range cfg.Repos {
			if r.Group != "" && strings.TrimSpace(r.Group) == "" {
				issues = append(issues, fmt.Sprintf("Blank group for %s", r.URL))
			}
		}

		// Print results
		fmt.Println(ui.Title.Render("Validate"))
		fmt.Println()

		if len(issues) == 0 {
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), ui.Muted.Render("Configuration is valid — no issues found."))
		} else {
			for _, issue := range issues {
				fmt.Printf("  %s %s\n", ui.Error.Render("✗"), ui.Muted.Render(issue))
			}
		}

		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
