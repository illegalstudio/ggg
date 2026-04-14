package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
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

		useHTTP, _ := cmd.Flags().GetBool("http")
		group, _ := cmd.Flags().GetString("group")

		var repos []ghRepo
		var fetchErr error
		action := func() {
			repos, fetchErr = fetchRepos(org)
		}

		if err := spinner.New().Title(fmt.Sprintf("Fetching repositories from %s...", org)).Action(action).Run(); err != nil {
			return err
		}
		if fetchErr != nil {
			return fetchErr
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
			url := r.SSHURL
			if useHTTP {
				url = r.CloneURL
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

const personalAccount = "__personal__"

func currentUser() (string, error) {
	out, err := exec.Command("gh", "api", "/user", "--jq", ".login").Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch user: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func selectOrg() (string, error) {
	username, err := currentUser()
	if err != nil {
		return "", err
	}

	out, err := exec.Command("gh", "api", "/user/orgs", "--jq", ".[].login").Output()
	if err != nil {
		return "", fmt.Errorf("failed to fetch organizations: %w", err)
	}

	options := []huh.Option[string]{
		huh.NewOption(fmt.Sprintf("%s (personal)", username), personalAccount),
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
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

func fetchRepos(account string) ([]ghRepo, error) {
	var out []byte
	var err error

	if account == personalAccount {
		// Authenticated endpoint: returns all repos (public + private)
		out, err = exec.Command("gh", "api", "--paginate",
			"/user/repos?per_page=100&affiliation=owner&sort=full_name",
		).Output()
	} else {
		// Try as org first, fall back to user
		out, err = exec.Command("gh", "api", "--paginate",
			fmt.Sprintf("/orgs/%s/repos?per_page=100&sort=full_name", account),
		).Output()
		if err != nil {
			out, err = exec.Command("gh", "api", "--paginate",
				fmt.Sprintf("/users/%s/repos?per_page=100&sort=full_name", account),
			).Output()
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repos: %w", err)
	}

	var repos []ghRepo
	if err := json.Unmarshal(out, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse repos: %w", err)
	}
	return repos, nil
}

func selectRepos(repos []ghRepo) ([]ghRepo, error) {
	if selected, ok, err := selectReposFromEnv(repos); ok || err != nil {
		return selected, err
	}

	options := make([]huh.Option[int], len(repos))
	for i, r := range repos {
		label := r.FullName
		if r.Private {
			label += " 🔒"
		}
		options[i] = huh.NewOption(label, i)
	}

	var selected []int
	err := huh.NewMultiSelect[int]().
		Title(fmt.Sprintf("Select repositories to import (%d found)", len(repos))).
		Description("ctrl+a select all · / filter · space toggle · enter confirm").
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

func selectReposFromEnv(repos []ghRepo) ([]ghRepo, bool, error) {
	raw := strings.TrimSpace(os.Getenv("GGG_TEST_IMPORT_SELECTION"))
	if raw == "" {
		return nil, false, nil
	}

	if raw == "*" || strings.EqualFold(raw, "all") {
		return repos, true, nil
	}

	parts := strings.Split(raw, ",")
	selected := make([]ghRepo, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if idx, err := strconv.Atoi(part); err == nil {
			if idx < 0 || idx >= len(repos) {
				return nil, true, fmt.Errorf("test import selection index %d out of range", idx)
			}
			selected = append(selected, repos[idx])
			continue
		}

		matched := false
		for _, r := range repos {
			if r.FullName == part {
				selected = append(selected, r)
				matched = true
				break
			}
		}
		if !matched {
			return nil, true, fmt.Errorf("test import selection %q not found", part)
		}
	}

	return selected, true, nil
}

func init() {
	importCmd.Flags().Bool("http", false, "Use HTTPS URLs instead of SSH")
	importCmd.Flags().StringP("group", "g", "", "Assign imported repos to a group")
	rootCmd.AddCommand(importCmd)
}
