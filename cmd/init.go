package cmd

import (
	"fmt"
	"os"

	"go-git-get/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a default configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.ConfigPath()

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("config file already exists at %s", path)
		}

		if err := config.WriteDefault(path); err != nil {
			return err
		}

		fmt.Println("Config file created at", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
