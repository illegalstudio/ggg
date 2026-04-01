package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"go-git-get/config"
	"go-git-get/ui"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/cobra"
)

type ghRepo struct {
	FullName string `json:"full_name"`
	SSHURL   string `json:"ssh_url"`
	CloneURL string `json:"clone_url"`
	Private  bool   `json:"private"`
}

type ghOrg struct {
	Login string `json:"login"`
}

var importCmd = &cobra.Command{
	Use:     "import [org]",
	Short:   "Import repositories from GitHub via gh CLI",
	GroupID: GroupConfig,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireBinary("gh"); err != nil {
			return err
		}

		org := ""
		if len(args) > 0 {
			org = args[0]
		} else {
			selected, err := selectOrg()
			if err != nil {
				return err
			}
			org = selected
		}

		useSSH, _ := cmd.Flags().GetBool("ssh")
		group, _ := cmd.Flags().GetString("group")

		var repos []ghRepo
		action := func() {
			var err error
			repos, err = fetchRepos(org)
			if err != nil {
				repos = nil
			}
		}

		if err := spinner.New().Title(fmt.Sprintf("Fetching repositories from %s...", org)).Action(action).Run(); err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println(ui.Info.Render("No repositories found."))
			return nil
		}

		sort.Slice(repos, func(i, j int) bool {
			return repos[i].FullName < repos[j].FullName
		})

		selected, err := selectRepos(repos)
		if err != nil {
			return err
		}

		if len(selected) == 0 {
			fmt.Println(ui.Muted.Render("No repositories selected."))
			return nil
		}

		cfg, err := config.LoadRaw()
		if err != nil {
			return err
		}

		existing := make(map[string]bool)
		for _, r := range cfg.Repos {
			existing[r.URL] = true
		}

		added := 0
		for _, r := range selected {
			url := r.CloneURL
			if useSSH {
				url = r.SSHURL
			}

			if existing[url] {
				fmt.Printf("  %s %s %s\n", ui.Muted.Render("●"), ui.Muted.Render(r.FullName), ui.Muted.Render("(already configured)"))
				continue
			}

			cfg.Repos = append(cfg.Repos, config.Repo{URL: url, Group: group})
			added++
			fmt.Printf("  %s %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.FullName))
		}

		if added == 0 {
			fmt.Println(ui.Info.Render("All selected repositories are already configured."))
			return nil
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("\n  %s\n", ui.Success.Render(fmt.Sprintf("Added %d repositories.", added)))
		return nil
	},
}

func selectOrg() (string, error) {
	out, err := exec.Command("gh", "api", "/user/orgs", "--jq", ".[].login").Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch organizations: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")

	// Get the authenticated user
	userOut, err := exec.Command("gh", "api", "/user", "--jq", ".login").Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch user: %w", err)
	}
	username := strings.TrimSpace(string(userOut))

	options := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("%s (personal)", username), username),
	}
	for _, line := range lines {
		if line != "" {
			options = append(options, huh.NewOption(line, line))
		}
	}

	var choice string
	err = huh.NewSelect[string]().
		Title("Select an account to import from").
		Options(options...).
		Value(&choice).
		Run()
	if err != nil {
		return "", err
	}

	return choice, nil
}

func fetchRepos(org string) ([]ghRepo, error) {
	out, err := exec.Command("gh", "api",
		fmt.Sprintf("/users/%s/repos?per_page=100&sort=full_name", org),
	).Output()
	if err != nil {
		// Try as org endpoint
		out, err = exec.Command("gh", "api",
			fmt.Sprintf("/orgs/%s/repos?per_page=100&sort=full_name", org),
		).Output()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch repos: %w", err)
		}
	}

	var repos []ghRepo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}
	return repos, nil
}

func selectRepos(repos []ghRepo) ([]ghRepo, error) {
	options := make([]huh.Option[int], len(repos))
	for i, r := range repos {
		label := r.FullName
		if r.Private {
			label += " 🔒"
		}
		options[i] = huh.NewOption(label, i).Selected(true)
	}

	var selected []int
	err := huh.NewMultiSelect[int]().
		Title("Select repositories to import").
		Description("/ filter → esc apply → space toggle · ctrl+a all · enter confirm").
		Options(options...).
		Filterable(true).
		Height(20).
		Value(&selected).
		Run()
	if err != nil {
		return nil, err
	}

	result := make([]ghRepo, len(selected))
	for i, idx := range selected {
		result[i] = repos[idx]
	}
	return result, nil
}

func init() {
	importCmd.Flags().BoolP("ssh", "s", false, "Use SSH URLs instead of HTTPS")
	importCmd.Flags().StringP("group", "g", "", "Assign imported repos to a group")
	rootCmd.AddCommand(importCmd)
}
