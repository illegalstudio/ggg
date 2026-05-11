package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/illegalstudio/ggg/internal/config"
	"github.com/illegalstudio/ggg/internal/ui"

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
	Use:     "import [org] [repo]",
	Short:   "Import repositories from GitHub via gh CLI",
	GroupID: GroupConfig,
	Args:    cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		org := ""
		repoQuery := ""
		if len(args) > 0 {
			org = args[0]
			if len(args) > 1 {
				repoQuery = args[1]
			}
		} else if jsonOutput {
			return fmt.Errorf("--json mode requires explicit org/user and repository arguments")
		}
		if jsonOutput && repoQuery == "" {
			return fmt.Errorf("--json mode requires a repository argument")
		}

		if err := requireBinary("gh"); err != nil {
			return err
		}

		if len(args) == 0 {
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

		if jsonOutput {
			action()
		} else {
			if err := spinner.New().Title(fmt.Sprintf("Fetching repositories from %s...", org)).Action(action).Run(); err != nil {
				return err
			}
		}
		if fetchErr != nil {
			return fetchErr
		}

		if len(repos) == 0 {
			if done, err := maybeJSON(map[string]any{"added": []string{}, "skipped": []string{}}); done {
				return err
			}
			fmt.Println(ui.Info.Render("No repositories found."))
			return nil
		}

		sort.Slice(repos, func(i, j int) bool {
			return repos[i].FullName < repos[j].FullName
		})

		var selected []ghRepo
		if repoQuery != "" {
			selectedRepo, err := resolveImportedRepo(repos, repoQuery)
			if err != nil {
				return err
			}
			selected = []ghRepo{selectedRepo}
		} else {
			var err error
			selected, err = selectRepos(repos)
			if err != nil {
				return err
			}
		}

		if len(selected) == 0 {
			if done, err := maybeJSON(map[string]any{"added": []string{}, "skipped": []string{}}); done {
				return err
			}
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

		var addedURLs, skippedURLs []string
		for _, r := range selected {
			url := r.SSHURL
			if useHTTP {
				url = r.CloneURL
			}

			if existing[url] {
				skippedURLs = append(skippedURLs, url)
				if !jsonOutput {
					fmt.Printf("  %s %s %s\n", ui.Muted.Render("●"), ui.Muted.Render(r.FullName), ui.Muted.Render("(already configured)"))
				}
				continue
			}

			cfg.Repos = append(cfg.Repos, config.Repo{URL: url, Group: group})
			addedURLs = append(addedURLs, url)
			if !jsonOutput {
				fmt.Printf("  %s %s\n", ui.Success.Render("✓"), ui.Repo.Render(r.FullName))
			}
		}

		if len(addedURLs) == 0 {
			if done, err := maybeJSON(map[string]any{"added": addedURLs, "skipped": skippedURLs}); done {
				return err
			}
			fmt.Println(ui.Info.Render("All selected repositories are already configured."))
			return nil
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		if done, err := maybeJSON(map[string]any{"added": addedURLs, "skipped": skippedURLs}); done {
			return err
		}

		fmt.Printf("\n  %s\n", ui.Success.Render(fmt.Sprintf("Added %d repositories.", len(addedURLs))))
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

	if jsonOutput {
		return nil, fmt.Errorf("--json mode requires a repository argument")
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
		Description("/ filter · ctrl+a select all · space toggle · enter confirm").
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

func resolveImportedRepo(repos []ghRepo, query string) (ghRepo, error) {
	query = strings.TrimSpace(query)
	queryLower := strings.ToLower(query)
	matches := make([]ghRepo, 0, 1)
	seen := map[string]bool{}

	for _, r := range repos {
		name := r.FullName
		if slash := strings.LastIndex(name, "/"); slash >= 0 {
			name = name[slash+1:]
		}

		if strings.EqualFold(r.FullName, query) ||
			strings.EqualFold(name, query) ||
			strings.ToLower(r.SSHURL) == queryLower ||
			strings.ToLower(r.CloneURL) == queryLower {
			if !seen[r.FullName] {
				matches = append(matches, r)
				seen[r.FullName] = true
			}
		}
	}

	if len(matches) == 0 {
		return ghRepo{}, fmt.Errorf("repository %q not found in imported repositories", query)
	}
	if len(matches) > 1 {
		return ghRepo{}, fmt.Errorf("repository %q matches multiple imported repositories", query)
	}
	return matches[0], nil
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
