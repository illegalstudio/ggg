package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/ui"
	gggskills "github.com/illegalstudio/ggg/skills"

	"github.com/spf13/cobra"
)

// maybeNoticeStaleSkill prints a quiet stderr reminder when an installed agent
// skill disagrees with the one bundled in this binary. It runs after a command
// succeeded, never in --json mode, never on commands whose output is
// machine-consumed (shell-init, cd, completion) or that already report
// skill state themselves (skills), and never breaks the command it follows.
func maybeNoticeStaleSkill(cmd *cobra.Command) {
	if jsonOutput || !skillNoticeApplies(cmd) {
		return
	}
	if suppress, err := config.SuppressSkillsNotice(); err != nil || suppress {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	var stale []skillTarget
	forceNeeded := false
	for _, target := range skillTargets(home) {
		status, err := gggskills.Verify(target.Path)
		if err != nil || !status.Stale() {
			continue
		}
		stale = append(stale, target)
		if status == gggskills.VerifyModified {
			forceNeeded = true
		}
	}
	if len(stale) == 0 {
		return
	}

	var b strings.Builder
	if len(stale) == 1 {
		fmt.Fprintf(&b, "note: the ggg agent skill at %s is not in sync with this ggg version.\n", displayPath(stale[0].Path))
	} else {
		b.WriteString("note: these ggg agent skills are not in sync with this ggg version:\n")
		for _, target := range stale {
			fmt.Fprintf(&b, "        %s\n", displayPath(target.Path))
		}
	}
	update := "ggg skills install"
	if forceNeeded {
		update = "ggg skills install --force (the installed copy was modified)"
	}
	fmt.Fprintf(&b, "      update:  %s\n", update)
	b.WriteString("      verify:  ggg skills verify\n")
	b.WriteString("      silence: set suppress_skills_notice: true in ~/.config/ggg/repositories.yaml")

	// Style line by line: styling the whole block would pad every line to the
	// widest one, leaving stray trailing spaces in captured output.
	lines := strings.Split(b.String(), "\n")
	for i, line := range lines {
		lines[i] = ui.Muted.Render(line)
	}
	fmt.Fprintln(os.Stderr, strings.Join(lines, "\n"))
}

// skillNoticeApplies reports whether a finished command should be followed by
// the stale-skill notice.
func skillNoticeApplies(cmd *cobra.Command) bool {
	if cmd == nil || cmd == rootCmd {
		return false // bare `ggg` (help) and --version
	}
	if cmd.Hidden {
		return false // cobra's __complete / __completeNoDesc shell helpers
	}
	if help, _ := cmd.Flags().GetBool("help"); help {
		return false // `ggg <cmd> --help` returns the command itself, not "help"
	}
	if strings.HasPrefix(cmd.CommandPath(), rootCmd.Name()+" skills") ||
		strings.HasPrefix(cmd.CommandPath(), rootCmd.Name()+" completion") {
		return false
	}
	switch cmd.Name() {
	case "help", "shell-init", "cd":
		return false
	}
	return true
}
