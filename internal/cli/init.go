package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Generate a default configuration file",
	GroupID: GroupConfig,
	Long: `Generate a default configuration file.

Without flags, init creates the shared ~/.config/ggg/repositories.yaml.

With --local, init creates ~/.config/ggg/repositories.local.yaml instead: a
machine-local file that overrides base_dir (and optionally pull_strategy,
aliases, suppress_skills_notice) on this machine only. Use it when the shared
config is synced between machines whose clone root differs. Without --base-dir
the base directory is asked interactively; --json requires --base-dir.`,
	Example: `  ggg init
  ggg init --base-dir ~/code
  ggg init --local --base-dir /data/dev`,
	RunE: func(cmd *cobra.Command, args []string) error {
		baseDir, _ := cmd.Flags().GetString("base-dir")
		if local, _ := cmd.Flags().GetBool("local"); local {
			return runInitLocal(cmd, baseDir)
		}

		path := config.ConfigPath()

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists at %s", path)
		}

		if err := config.WriteDefault(path, strings.TrimSpace(baseDir)); err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"created": true, "path": path}); done {
			return err
		}

		fmt.Println(ui.Success.Render("✓") + " Config file created at " + ui.Path.Render(path))
		return nil
	},
}

func runInitLocal(cmd *cobra.Command, baseDir string) error {
	path := config.LocalConfigPath()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("local config file already exists at %s", path)
	}

	if !cmd.Flags().Changed("base-dir") {
		if jsonOutput {
			return errors.New("--base-dir is required with --local and --json")
		}
		baseDir = suggestedLocalBaseDir()
		err := huh.NewInput().
			Title("Base directory on this machine").
			Description("Overrides base_dir from repositories.yaml on this machine only.").
			Value(&baseDir).
			Validate(validateBaseDirInput).
			Run()
		if err != nil {
			return err
		}
	}
	baseDir = strings.TrimSpace(baseDir)
	if err := validateBaseDirInput(baseDir); err != nil {
		return err
	}

	if err := config.WriteLocal(path, baseDir); err != nil {
		return err
	}

	if done, err := maybeJSON(map[string]any{"created": true, "path": path, "base_dir": baseDir}); done {
		return err
	}

	fmt.Println(ui.Success.Render("✓") + " Local config file created at " + ui.Path.Render(path))
	fmt.Println(ui.Muted.Render("  Keep it out of whatever syncs repositories.yaml between machines."))
	return nil
}

// suggestedLocalBaseDir pre-fills the interactive prompt with the shared
// base_dir, falling back to the default.
func suggestedLocalBaseDir() string {
	if cfg, err := config.LoadRaw(); err == nil && cfg.BaseDir != "" {
		return cfg.BaseDir
	}
	return config.DefaultBaseDir
}

func validateBaseDirInput(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("base directory cannot be empty")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	if _, err := config.ExpandPath(s, home); err != nil {
		return err
	}
	return nil
}

func init() {
	initCmd.Flags().Bool("local", false, "Create the machine-local override file instead of the shared config")
	initCmd.Flags().String("base-dir", "", "Base directory to write into the generated file")
	rootCmd.AddCommand(initCmd)
}
