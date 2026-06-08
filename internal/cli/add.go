package cli

import (
	"fmt"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/repo"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <url>",
	Short:   "Add a repository to the configuration",
	GroupID: GroupConfig,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		groups, _ := cmd.Flags().GetStringArray("group")
		path, _ := cmd.Flags().GetString("path")
		clone, _ := cmd.Flags().GetBool("clone")

		cfg, err := config.LoadRaw()
		if err != nil {
			return err
		}

		for _, r := range cfg.Repos {
			if r.URL == url {
				return fmt.Errorf("repository %s is already configured", url)
			}
		}

		newRepo := config.Repo{URL: url, Groups: groups, Path: path}
		cfg.Repos = append(cfg.Repos, newRepo)

		if err := config.Save(cfg); err != nil {
			return err
		}

		out := map[string]any{"added": true, "url": url}

		if !jsonOutput {
			fmt.Printf("  %s Added %s\n", ui.Success.Render("✓"), ui.Repo.Render(url))
		}

		if clone {
			cfgFull, err := config.Load()
			if err != nil {
				return err
			}

			fullPath, err := repo.FullPath(cfgFull.BaseDir, newRepo)
			if err != nil {
				return err
			}

			out["path"] = fullPath

			if repo.IsCloned(fullPath) {
				out["cloned"] = true
				out["already_cloned"] = true
				if done, err := maybeJSON(out); done {
					return err
				}
				fmt.Printf("  %s %s\n", ui.Muted.Render("●"), ui.Muted.Render("Already cloned: "+fullPath))
				return nil
			}

			var cloneErr error
			action := func() {
				cloneErr = repo.Clone(url, fullPath)
			}

			if jsonOutput {
				action()
			} else {
				if err := spinner.New().Title(fmt.Sprintf("Cloning %s...", url)).Action(action).Run(); err != nil {
					return err
				}
			}

			if cloneErr != nil {
				out["cloned"] = false
				out["error"] = cloneErr.Error()
				if done, err := maybeJSON(out); done {
					return err
				}
				fmt.Printf("  %s %s %s\n", ui.Error.Render("✗"), ui.Repo.Render(url), ui.Error.Render(cloneErr.Error()))
				return cloneErr
			}
			out["cloned"] = true
			if done, err := maybeJSON(out); done {
				return err
			}
			fmt.Printf("  %s %s → %s\n", ui.Success.Render("✓"), ui.Repo.Render(url), ui.Path.Render(fullPath))
			return nil
		}

		if done, err := maybeJSON(out); done {
			return err
		}
		return nil
	},
}

func init() {
	addCmd.Flags().StringArrayP("group", "g", nil, "Assign the repo to one or more groups (repeatable)")
	addCmd.Flags().StringP("path", "p", "", "Custom clone path (relative to base_dir)")
	addCmd.Flags().BoolP("clone", "c", false, "Clone the repo immediately after adding")
	registerGroupCompletion(addCmd)
	rootCmd.AddCommand(addCmd)
}
