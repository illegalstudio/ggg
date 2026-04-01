package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"go-git-get/config"
	"go-git-get/repo"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <name> [editor]",
	Short: "Open a repository in an editor",
	Long: `Open a repository directory in an editor.

Without an editor argument, uses $EDITOR, $VISUAL, or vi as fallback.

Examples:
  ggg open myrepo              # open in default editor
  ggg open myrepo cursor       # open in Cursor
  ggg open myrepo code         # open in VS Code
  ggg open myrepo idea         # open in IntelliJ IDEA`,
	GroupID: GroupRepo,
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		repos, err := filterRepo(cfg.Repos, args[0])
		if err != nil {
			return err
		}
		r := repos[0]

		fullPath, err := repo.FullPath(cfg.BaseDir, r)
		if err != nil {
			return err
		}
		if !repo.IsCloned(fullPath) {
			return fmt.Errorf("repository is not cloned: %s", fullPath)
		}

		editor := defaultEditor()
		if len(args) == 2 {
			editor = args[1]
		}

		fmt.Printf("  %s Opening %s in %s\n", ui.Info.Render("●"), ui.Repo.Render(r.URL), ui.Repo.Render(editor))

		c := exec.Command(editor, fullPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
