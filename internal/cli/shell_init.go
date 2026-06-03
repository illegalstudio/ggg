package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// shellFunction is the shell wrapper for bash/zsh. It intercepts `ggg cd`
// and turns it into a real `cd`; everything else falls through to the binary.
// Only stdout (the resolved path) is captured, so the interactive selector —
// which huh renders to stderr/the TTY — works even for ambiguous names.
const shellFunction = `# ggg shell integration: turn 'ggg cd' into a real chdir.
ggg() {
  if [ "$1" = "cd" ]; then
    shift
    local _ggg_dir
    if ! _ggg_dir=$(command ggg cd "$@"); then
      return 1
    fi
    builtin cd "$_ggg_dir" || return $?
  else
    command ggg "$@"
  fi
}
`

// fishFunction is the equivalent for fish.
const fishFunction = `# ggg shell integration: turn 'ggg cd' into a real chdir.
function ggg
  if test "$argv[1]" = "cd"
    set -e argv[1]
    set -l _ggg_dir (command ggg cd $argv)
    if test $status -ne 0
      return 1
    end
    builtin cd $_ggg_dir
  else
    command ggg $argv
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
	Short:             "Print shell integration script (eval to enable `ggg cd` and completions)",
	GroupID:           GroupConfig,
	ValidArgsFunction: shellCompletion,
	Long: `Print a shell function that makes "ggg cd" actually change directory.

Without it, "ggg cd" can only print the path — a child process cannot change
the parent shell's directory. The generated script also installs Cobra-powered
tab completion for ggg commands, repository names, and group names.

Add to your shell configuration:
  bash:  eval "$(ggg shell-init bash)"   (in ~/.bashrc)
  zsh:   eval "$(ggg shell-init zsh)"    (in ~/.zshrc)
  fish:  ggg shell-init fish | source    (in ~/.config/fish/config.fish)

Then use "ggg cd <name>" to navigate to a repository.`,
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
