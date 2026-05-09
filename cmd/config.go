package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"go-git-get/config"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Open the configuration file in your default editor",
	GroupID: GroupConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			return unsupportedJSON("config")
		}

		path := config.ConfigPath()

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("config file not found at %s (run ggg init first)", path)
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
	rootCmd.AddCommand(configCmd)
}
