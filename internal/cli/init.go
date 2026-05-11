package cli

import (
	"fmt"
	"os"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Generate a default configuration file",
	GroupID: GroupConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath()

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists at %s", path)
		}

		if err := config.WriteDefault(path); err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"created": true, "path": path}); done {
			return err
		}

		fmt.Println(ui.Success.Render("✓") + " Config file created at " + ui.Path.Render(path))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
