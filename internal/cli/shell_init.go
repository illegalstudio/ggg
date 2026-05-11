package cli

import (
	"fmt"
	"strings"

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

func generateCompletionScript(shell string) (string, error) {
	var buf strings.Builder
	switch shell {
	case "bash":
		if err := rootCmd.GenBashCompletion(&buf); err != nil {
			return "", err
		}
	case "zsh":
		if err := rootCmd.GenZshCompletion(&buf); err != nil {
			return "", err
		}
	case "fish":
		if err := rootCmd.GenFishCompletion(&buf, true); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
	return buf.String(), nil
}

var shellInitCmd = &cobra.Command{
	Use:               "shell-init [bash|zsh|fish]",
	Short:             "Print shell integration script (gcd alias and completions)",
	GroupID:           GroupConfig,
	ValidArgsFunction: shellCompletion,
	Long: `Print a shell function that defines the "gcd" alias for quick navigation.

The generated script also installs Cobra-powered tab completion for ggg
commands, repository names, group names, and local branches where applicable.

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

		var script string
		switch shell {
		case "bash", "zsh":
			script = shellFunction
		case "fish":
			script = fishFunction
		default:
			return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
		}

		completion, err := generateCompletionScript(shell)
		if err != nil {
			return err
		}
		script += "\n" + completion

		if done, err := maybeJSON(map[string]any{"shell": shell, "script": script}); done {
			return err
		}

		fmt.Print(script)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(shellInitCmd)
}
