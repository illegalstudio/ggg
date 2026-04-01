package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const shellFunction = `# GGG shell integration — gcd alias
# Add this to your .bashrc or .zshrc
gcd() {
  if [ -z "$1" ]; then
    echo "Usage: gcd <name>"
    return 1
  fi
  local dir
  dir=$(ggg cd "$1" 2>&1)
  if [ $? -eq 0 ]; then
    cd "$dir"
  else
    echo "$dir"
    return 1
  fi
}
`

const fishFunction = `# GGG shell integration for fish — gcd alias
# Add this to your ~/.config/fish/config.fish
function gcd
  if test -z "$argv[1]"
    echo "Usage: gcd <name>"
    return 1
  end
  set -l dir (command ggg cd $argv[1] 2>&1)
  if test $status -eq 0
    cd $dir
  else
    echo $dir
    return 1
  end
end
`

var shellInitCmd = &cobra.Command{
	Use:     "shell-init [bash|zsh|fish]",
	Short:   "Print shell integration script (gcd alias)",
	GroupID: GroupConfig,
	Long: `Print a shell function that defines the "gcd" alias for quick navigation.

Add to your shell configuration:
  bash:  eval "$(ggg shell-init bash)"   (in ~/.bashrc)
  zsh:   eval "$(ggg shell-init zsh)"    (in ~/.zshrc)
  fish:  ggg shell-init fish | source    (in ~/.config/fish/config.fish)

Then use "gcd <name>" to navigate to a repository.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := "zsh"
		if len(args) > 0 {
			shell = args[0]
		}

		switch shell {
		case "bash", "zsh":
			fmt.Print(shellFunction)
		case "fish":
			fmt.Print(fishFunction)
		default:
			return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellInitCmd)
}
