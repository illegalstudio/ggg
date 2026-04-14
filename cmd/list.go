package cmd

import (
	"fmt"
	"sort"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list [filter]",
	Short:   "List configured repositories and their status",
	GroupID: GroupInfo,
	Args:    cobra.MaximumNArgs(1),
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
			return nil
		}

		repos = filterByName(repos, getFilter(cmd, args))

		fmt.Println(ui.Title.Render("Repositories"))
		fmt.Println()
		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				fmt.Printf("  %s %s\n", ui.Error.Render("✗"), ui.Error.Render(r.URL+" (invalid URL)"))
				continue
			}

			if repo.IsCloned(fullPath) {
				fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.URL), ui.Path.Render(fullPath))
			} else {
				fmt.Printf("  %s %s → %s\n", ui.Muted.Render("○"), ui.Repo.Render(r.URL), ui.Path.Render(fullPath))
			}
		}
		fmt.Println()
		return nil
	},
}

func listGroups() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	groups := map[string]int{}
	for _, r := range cfg.Repos {
		if r.Group != "" {
			groups[r.Group]++
		}
	}

	if len(groups) == 0 {
		fmt.Println(ui.Info.Render("No groups defined."))
		return nil
	}

	names := make([]string, 0, len(groups))
	for g := range groups {
		names = append(names, g)
	}
	sort.Strings(names)

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
	rootCmd.AddCommand(listCmd)
}
