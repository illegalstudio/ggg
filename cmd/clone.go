package cmd

import (
	"fmt"

	"go-git-get/config"
	"go-git-get/repo"

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
			fmt.Println("No repositories configured.")
			return nil
		}

		repos := cfg.Repos
		if len(args) > 0 {
			filtered, err := filterRepo(cfg.Repos, args[0])
			if err != nil {
				return err
			}
			repos = filtered
		}

		for _, r := range repos {
			fullPath, err := repo.FullPath(cfg.BaseDir, r)
			if err != nil {
				fmt.Printf("Skipping %s: %v\n", r.URL, err)
				continue
			}

			if repo.IsCloned(fullPath) {
				fmt.Printf("Already cloned: %s\n", fullPath)
				continue
			}

			fmt.Printf("Cloning %s → %s\n", r.URL, fullPath)
			if err := repo.Clone(r.URL, fullPath); err != nil {
				fmt.Printf("Error cloning %s: %v\n", r.URL, err)
			}
		}
		return nil
	},
}

func filterRepo(repos []config.Repo, name string) ([]config.Repo, error) {
	for _, r := range repos {
		derived, _ := repo.DerivePathFromURL(r.URL)
		if r.Path == name || r.URL == name || derived == name {
			return []config.Repo{r}, nil
		}
	}
	return nil, fmt.Errorf("repository %q not found in config", name)
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
