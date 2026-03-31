package cmd

import (
	"fmt"
	"os"

	"go-git-get/config"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ggg",
	Short: "Go Git Get — clone and manage git repositories from a config file",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// filterByGroup filters repos by group if group is non-empty.
func filterByGroup(repos []config.Repo, group string) []config.Repo {
	if group == "" {
		return repos
	}
	var filtered []config.Repo
	for _, r := range repos {
		if r.Group == group {
			filtered = append(filtered, r)
		}
	}
	return filtered
}
