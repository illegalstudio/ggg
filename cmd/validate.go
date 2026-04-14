package cmd

import (
	"fmt"
	"strings"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

// conflict groups indices of repos that clash on a duplicate URL or resolved path.
type conflict struct {
	kind    string // "Duplicate URL" or "Path conflict"
	key     string // the duplicated URL or the conflicting path
	indices []int  // positions in cfg.Repos
}

var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Validate configuration for errors and conflicts",
	GroupID: GroupInfo,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadRaw()
		if err != nil {
			return err
		}

		cfgExpanded, err := config.Load()
		if err != nil {
			return err
		}

		var warnings []string
		var conflicts []conflict

		// Check 1: duplicate URLs
		urlIndices := map[string][]int{}
		for i, r := range cfg.Repos {
			urlIndices[r.URL] = append(urlIndices[r.URL], i)
		}
		for u, idxs := range urlIndices {
			if len(idxs) > 1 {
				conflicts = append(conflicts, conflict{
					kind:    "Duplicate URL",
					key:     u,
					indices: idxs,
				})
			}
		}

		// Check 2: conflicting resolved paths
		pathIndices := map[string][]int{}
		for i, r := range cfgExpanded.Repos {
			fullPath, err := repo.FullPath(cfgExpanded.BaseDir, r)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Invalid URL: %s", r.URL))
				continue
			}
			pathIndices[fullPath] = append(pathIndices[fullPath], i)
		}
		for p, idxs := range pathIndices {
			if len(idxs) > 1 {
				// Skip if this is already covered by a duplicate URL conflict with the same indices
				if isDuplicateOf(conflicts, idxs) {
					continue
				}
				conflicts = append(conflicts, conflict{
					kind:    "Path conflict",
					key:     p,
					indices: idxs,
				})
			}
		}

		// Check 3: invalid pull strategies
		validStrategies := map[config.PullStrategy]bool{
			config.PullMerge: true, config.PullRebase: true, config.PullFFOnly: true, "": true,
		}
		if !validStrategies[cfgExpanded.PullStrategy] {
			warnings = append(warnings, fmt.Sprintf("Invalid global pull_strategy: %q", cfgExpanded.PullStrategy))
		}
		for _, r := range cfg.Repos {
			if !validStrategies[r.PullStrategy] {
				warnings = append(warnings, fmt.Sprintf("Invalid pull_strategy %q for %s", r.PullStrategy, r.URL))
			}
		}

		// Check 4: empty groups (group field set to whitespace)
		for _, r := range cfg.Repos {
			if r.Group != "" && strings.TrimSpace(r.Group) == "" {
				warnings = append(warnings, fmt.Sprintf("Blank group for %s", r.URL))
			}
		}

		// Print results
		fmt.Println(ui.Title.Render("Validate"))
		fmt.Println()

		if len(warnings) == 0 && len(conflicts) == 0 {
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), ui.Muted.Render("Configuration is valid — no issues found."))
			fmt.Println()
			return nil
		}

		for _, c := range conflicts {
			urls := make([]string, len(c.indices))
			for i, idx := range c.indices {
				urls[i] = cfg.Repos[idx].URL
			}
			fmt.Printf("  %s %s: %s → %s\n", ui.Error.Render("✗"), ui.Muted.Render(c.kind), ui.Path.Render(c.key), ui.Repo.Render(strings.Join(urls, ", ")))
		}
		for _, w := range warnings {
			fmt.Printf("  %s %s\n", ui.Error.Render("✗"), ui.Muted.Render(w))
		}
		fmt.Println()

		// Offer cleanup wizard if there are fixable conflicts
		if len(conflicts) == 0 {
			return nil
		}

		var runWizard bool
		err = huh.NewConfirm().
			Title("Fix conflicts interactively?").
			Affirmative("Yes, clean up").
			Negative("No, skip").
			Value(&runWizard).
			Run()
		if err != nil {
			return err
		}
		if !runWizard {
			return nil
		}

		// Collect indices to remove
		removeSet := map[int]bool{}

		for _, c := range conflicts {
			options := make([]huh.Option[int], len(c.indices))
			for i, idx := range c.indices {
				r := cfg.Repos[idx]
				label := r.URL
				if r.Path != "" {
					label += " (path: " + r.Path + ")"
				}
				if r.Group != "" {
					label += " [" + r.Group + "]"
				}
				options[i] = huh.NewOption(label, idx)
			}

			var keep int
			err := huh.NewSelect[int]().
				Title(fmt.Sprintf("%s: %s — which entry to keep?", c.kind, c.key)).
				Options(options...).
				Value(&keep).
				Run()
			if err != nil {
				return err
			}

			for _, idx := range c.indices {
				if idx != keep {
					removeSet[idx] = true
				}
			}
		}

		// Build new repos list preserving order
		var cleaned []config.Repo
		for i, r := range cfg.Repos {
			if !removeSet[i] {
				cleaned = append(cleaned, r)
			}
		}
		cfg.Repos = cleaned

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("  %s Removed %d duplicate entries. Config saved.\n", ui.Success.Render("✓"), len(removeSet))
		fmt.Println()
		return nil
	},
}

// isDuplicateOf checks whether indices match exactly with an existing conflict.
func isDuplicateOf(conflicts []conflict, indices []int) bool {
	for _, c := range conflicts {
		if len(c.indices) != len(indices) {
			continue
		}
		match := true
		for i := range c.indices {
			if c.indices[i] != indices[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
