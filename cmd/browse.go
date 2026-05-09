package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"go-git-get/config"
	"go-git-get/ui"

	"github.com/spf13/cobra"
)

var browseCmd = &cobra.Command{
	Use:     "browse <name>",
	Short:   "Open a repository's remote URL in the browser",
	GroupID: GroupRepo,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			return unsupportedJSON("browse")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		r, err := resolveOneRepo(cfg.Repos, args[0])
		if err != nil {
			return err
		}

		browseURL := httpURL(r.URL)
		fmt.Printf("  %s Opening %s\n", ui.Info.Render("●"), ui.Repo.Render(browseURL))
		return openBrowser(browseURL)
	},
}

func httpURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return strings.TrimSuffix(rawURL, ".git")
	}
	if strings.Contains(rawURL, "@") && strings.Contains(rawURL, ":") {
		parts := strings.SplitN(rawURL, "@", 2)
		hostAndPath := parts[1]
		hostAndPath = strings.Replace(hostAndPath, ":", "/", 1)
		hostAndPath = strings.TrimSuffix(hostAndPath, ".git")
		return "https://" + hostAndPath
	}
	return rawURL
}

func openBrowser(url string) error {
	if override := strings.TrimSpace(os.Getenv("GGG_TEST_BROWSER_CMD")); override != "" {
		return exec.Command(override, url).Run()
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(browseCmd)
}
