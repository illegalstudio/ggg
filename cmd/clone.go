package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

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
		if len(args) == 0 {
			fmt.Printf("Clone all %d repositories? [y/N] ", len(cfg.Repos))
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(input)) != "y" {
				fmt.Println("Aborted.")
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

	// Multiple matches: prompt user to choose
	fmt.Printf("Multiple repositories match %q:\n", name)
	for i, r := range matches {
		fmt.Printf("  %d) %s\n", i+1, r.URL)
	}
	fmt.Print("Choose a number: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(matches) {
		return nil, fmt.Errorf("invalid choice")
	}

	return []config.Repo{matches[choice-1]}, nil
}

func init() {
	rootCmd.AddCommand(cloneCmd)
}
