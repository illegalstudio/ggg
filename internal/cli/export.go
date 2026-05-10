package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"go-git-get/internal/config"
	"go-git-get/internal/ui"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:     "export [path]",
	Short:   "Export the configuration file to a given path",
	Args:    cobra.MaximumNArgs(1),
	GroupID: GroupConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		src := config.ConfigPath()
		if _, err := os.Stat(src); err != nil {
			return fmt.Errorf("config file not found at %s", src)
		}

		var dest string
		if len(args) == 1 {
			dest = args[0]
		} else if !jsonOutput {
			err := huh.NewInput().
				Title("Export path").
				Placeholder("./repositories.yaml").
				Value(&dest).
				Run()
			if err != nil {
				return err
			}
		}

		if dest == "" {
			dest = "repositories.yaml"
		}

		// Expand ~ prefix
		if len(dest) >= 2 && dest[:2] == "~/" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cannot determine home directory: %w", err)
			}
			dest = filepath.Join(home, dest[2:])
		}

		// If dest is a directory, append the filename
		if info, err := os.Stat(dest); err == nil && info.IsDir() {
			dest = filepath.Join(dest, "repositories.yaml")
		}

		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("cannot read config: %w", err)
		}

		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("cannot create directory: %w", err)
		}

		if err := os.WriteFile(dest, data, 0644); err != nil {
			return fmt.Errorf("cannot write file: %w", err)
		}

		abs, _ := filepath.Abs(dest)
		if done, err := maybeJSON(map[string]any{"exported": true, "path": abs}); done {
			return err
		}
		fmt.Println(ui.Success.Render("✓") + " Config exported to " + ui.Path.Render(abs))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
