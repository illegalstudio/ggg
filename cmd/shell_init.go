package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const shellFunction = `# GGG shell integration
# Add this to your .bashrc, .zshrc, or config.fish
ggg() {
  if [ "$1" = "cd" ]; then
    if [ -z "$2" ]; then
      echo "Usage: ggg cd <name>"
      return 1
    fi
    local dir
    dir=$(command ggg cd "$2" 2>&1)
    if [ $? -eq 0 ]; then
      cd "$dir"
    else
      echo "$dir"
      return 1
    fi
  else
    command ggg "$@"
  fi
}
`

const fishFunction = `# GGG shell integration for fish
# Add this to your ~/.config/fish/config.fish
function ggg
  if test "$argv[1]" = "cd"
    if test -z "$argv[2]"
      echo "Usage: ggg cd <name>"
      return 1
    end
    set -l dir (command ggg cd $argv[2] 2>&1)
    if test $status -eq 0
      cd $dir
    else
      echo $dir
      return 1
    end
  else
    command ggg $argv
  end
end
`

var shellInitCmd = &cobra.Command{
	Use:   "shell-init [bash|zsh|fish]",
	Short: "Print shell integration script (add to your shell rc file)",
	Long: `Print a shell function that wraps ggg to enable seamless "ggg cd".

Add to your shell configuration:
  bash:  eval "$(ggg shell-init bash)"   (in ~/.bashrc)
  zsh:   eval "$(ggg shell-init zsh)"    (in ~/.zshrc)
  fish:  ggg shell-init fish | source    (in ~/.config/fish/config.fish)`,
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
