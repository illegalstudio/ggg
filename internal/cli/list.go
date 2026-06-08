package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:               "list [filter]",
	Short:             "List configured repositories and their status",
	GroupID:           GroupInfo,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: repoCompletion,
	RunE: func(cmd *cobra.Command, args []string) error {
		showGroups, _ := cmd.Flags().GetBool("groups")
		if showGroups {
			return listGroups()
		}

		cfg, repos, err := loadRepos(cmd)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			if done, err := maybeJSON(map[string]any{"repos": []any{}}); done {
				return err
			}
			return nil
		}

		repos = filterByName(repos, getFilter(cmd, args))

		type listEntry struct {
			URL    string   `json:"url"`
			Path   string   `json:"path"`
			Groups []string `json:"groups,omitempty"`
			Cloned bool     `json:"cloned"`
			Error  string   `json:"error,omitempty"`
		}

		entries := make([]listEntry, 0, len(repos))
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				entries = append(entries, listEntry{URL: r.URL, Groups: r.Groups, Error: err.Error()})
				continue
			}
			entries = append(entries, listEntry{
				URL:    r.URL,
				Path:   fullPath,
				Groups: r.Groups,
				Cloned: repo.IsCloned(fullPath),
			})
		}

		if done, err := maybeJSON(map[string]any{"repos": entries}); done {
			return err
		}

		fmt.Println(ui.Title.Render("Repositories"))
		fmt.Println()
		for _, e := range entries {
			if e.Error != "" {
				fmt.Printf("  %s %s\n", ui.Error.Render("✗"), ui.Error.Render(e.URL+" (invalid URL)"))
				continue
			}
			if e.Cloned {
				fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(e.URL), ui.Path.Render(e.Path))
			} else {
				fmt.Printf("  %s %s → %s\n", ui.Muted.Render("○"), ui.Repo.Render(e.URL), ui.Path.Render(e.Path))
			}
		}
		fmt.Println()
		return nil
	},
}

// countGroups counts how many repos belong to each group. A group is counted
// at most once per repo (duplicate entries within a repo don't inflate the
// count), and blank entries (empty or whitespace-only) are skipped. Group
// names are kept verbatim so they stay consistent with the filter and
// completion, which match exactly.
func countGroups(repos []config.Repo) map[string]int {
	counts := map[string]int{}
	for _, r := range repos {
		seen := map[string]bool{}
		for _, g := range r.Groups {
			if strings.TrimSpace(g) == "" || seen[g] {
				continue
			}
			seen[g] = true
			counts[g]++
		}
	}
	return counts
}

func listGroups() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	groups := countGroups(cfg.Repos)

	names := make([]string, 0, len(groups))
	for g := range groups {
		names = append(names, g)
	}
	sort.Strings(names)

	type groupEntry struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	entries := make([]groupEntry, 0, len(names))
	for _, g := range names {
		entries = append(entries, groupEntry{Name: g, Count: groups[g]})
	}

	if done, err := maybeJSON(map[string]any{"groups": entries}); done {
		return err
	}

	if len(groups) == 0 {
		fmt.Println(ui.Info.Render("No groups defined."))
		return nil
	}

	fmt.Println(ui.Title.Render("Groups"))
	fmt.Println()
	for _, g := range names {
		fmt.Printf("  %s %s\n", ui.Repo.Render(g), ui.Muted.Render(fmt.Sprintf("(%d repos)", groups[g])))
	}
	fmt.Println()
	return nil
}

func init() {
	listCmd.Flags().StringP("group", "g", "", "List only repos in this group")
	listCmd.Flags().Bool("groups", false, "Show available groups")
	listCmd.Flags().StringP("filter", "f", "", "Filter repos by name")
	registerGroupCompletion(listCmd)
	registerFilterCompletion(listCmd)
	rootCmd.AddCommand(listCmd)
}
