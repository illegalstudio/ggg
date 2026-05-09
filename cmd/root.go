package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go-git-get/config"

	"github.com/spf13/cobra"
)

const (
	GroupConfig = "config"
	GroupRepo   = "repo"
	GroupInfo   = "info"
)

var rootCmd = &cobra.Command{
	Use:           "ggg",
	Short:         "Go Git Get — clone and manage git repositories from a config file",
	SilenceErrors: true,
	Long: `Go Git Get — clone and manage git repositories from a config file.

Shell Integration:
  Run "ggg shell-init" to generate a "gcd" shell function for quick navigation.

    eval "$(ggg shell-init zsh)"    # add to ~/.zshrc
    eval "$(ggg shell-init bash)"   # add to ~/.bashrc
    ggg shell-init fish | source    # add to ~/.config/fish/config.fish

  Then use "gcd <name>" to cd into any repository.

Use "ggg <command> --help" for details on any command.`,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format (suppresses spinners and prompts)")
}

func Execute() {
	rootCmd.SilenceUsage = jsonFlagRequested(os.Args[1:])

	rootCmd.AddGroup(
		&cobra.Group{ID: GroupConfig, Title: "Configuration:"},
		&cobra.Group{ID: GroupRepo, Title: "Repository Operations:"},
		&cobra.Group{ID: GroupInfo, Title: "Info & Diagnostics:"},
	)
	rootCmd.SetHelpCommandGroupID(GroupInfo)
	rootCmd.SetCompletionCommandGroupID(GroupConfig)

	if err := rootCmd.Execute(); err != nil {
		if jsonOutput {
			var payloadErr interface{ JSONPayload() any }
			if errors.As(err, &payloadErr) {
				_ = emitJSON(payloadErr.JSONPayload())
			} else {
				_ = emitJSON(map[string]any{"error": err.Error()})
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func jsonFlagRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--json" {
			return true
		}
		if strings.HasPrefix(arg, "--json=") {
			value := strings.TrimPrefix(arg, "--json=")
			return value != "false" && value != "0"
		}
	}
	return false
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
