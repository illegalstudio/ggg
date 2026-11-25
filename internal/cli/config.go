package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/ui"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Open the configuration file in your default editor",
	Long: `Open the configuration file in your default editor.

With --local, open the machine-local override file
(~/.config/ggg/repositories.local.yaml) instead; create it first with
ggg init --local.`,
	GroupID: GroupConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			return unsupportedJSON("config")
		}

		path := config.ConfigPath()
		hint := "ggg init"
		if local, _ := cmd.Flags().GetBool("local"); local {
			path = config.LocalConfigPath()
			hint = "ggg init --local"
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("config file not found at %s (run %s first)", path, hint)
		}

		editor := defaultEditor()

		fmt.Printf("  %s Opening %s with %s\n", ui.Info.Render("●"), ui.Path.Render(path), ui.Repo.Render(editor))

		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	configCmd.Flags().Bool("local", false, "Open the machine-local override file instead of the shared config")
	rootCmd.AddCommand(configCmd)
}
