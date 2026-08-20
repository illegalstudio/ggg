package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillTargets(t *testing.T) {
	home := t.TempDir()
	targets := skillTargets(home)

	if len(targets) != 2 {
		t.Fatalf("target count = %d, want 2", len(targets))
	}
	if targets[0].Key != "agents" || targets[1].Key != "claude" {
		t.Fatalf("target keys = %q, %q", targets[0].Key, targets[1].Key)
	}
	if want := filepath.Join(home, ".agents", "skills", "ggg"); targets[0].Path != want {
		t.Fatalf("agents path = %q, want %q", targets[0].Path, want)
	}
	if want := filepath.Join(home, ".claude", "skills", "ggg"); targets[1].Path != want {
		t.Fatalf("claude path = %q, want %q", targets[1].Path, want)
	}
}

func TestSkillTargetKeys(t *testing.T) {
	keys := skillTargetKeys()
	if len(keys) != 2 || keys[0] != "agents" || keys[1] != "claude" {
		t.Fatalf("keys = %v, want [agents claude]", keys)
	}
}

func TestFilterSkillTargetsPreservesDeclarationOrder(t *testing.T) {
	all := skillTargets(t.TempDir())

	selected, err := filterSkillTargets(all, []string{"claude", "agents"})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 {
		t.Fatalf("selected count = %d, want 2", len(selected))
	}
	if selected[0].Key != "agents" || selected[1].Key != "claude" {
		t.Fatalf("selected = %q, %q; want agents, claude", selected[0].Key, selected[1].Key)
	}
}

func TestFilterSkillTargetsDeduplicatesAndNormalizes(t *testing.T) {
	all := skillTargets(t.TempDir())

	selected, err := filterSkillTargets(all, []string{"Claude", " claude ", "claude"})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].Key != "claude" {
		t.Fatalf("selected = %+v, want one claude target", selected)
	}
}

func TestFilterSkillTargetsRejectsUnknownKey(t *testing.T) {
	all := skillTargets(t.TempDir())

	_, err := filterSkillTargets(all, []string{"cursor"})
	if err == nil {
		t.Fatal("expected an error for an unknown target")
	}
	if !strings.Contains(err.Error(), "unknown skill target") {
		t.Fatalf("error = %v, want it to mention an unknown skill target", err)
	}
	if !strings.Contains(err.Error(), "agents, claude") {
		t.Fatalf("error = %v, want it to list the valid targets", err)
	}
}

func TestFilterSkillTargetsRejectsEmptySelection(t *testing.T) {
	all := skillTargets(t.TempDir())

	if _, err := filterSkillTargets(all, []string{"  "}); err == nil {
		t.Fatal("expected an error when no target remains")
	}
}

func TestSelectSkillTargetsUsesAllTargetsInJSONMode(t *testing.T) {
	previous := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = previous })

	all := skillTargets(t.TempDir())
	selected, err := selectSkillTargets(all, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 2 {
		t.Fatalf("selected count = %d, want 2", len(selected))
	}
}

func TestSelectSkillTargetsPrefersExplicitTargetsOverJSONMode(t *testing.T) {
	previous := jsonOutput
	jsonOutput = true
	t.Cleanup(func() { jsonOutput = previous })

	all := skillTargets(t.TempDir())
	selected, err := selectSkillTargets(all, []string{"claude"})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].Key != "claude" {
		t.Fatalf("selected = %+v, want only claude", selected)
	}
}

func TestInstallSkillTargetsKeepsDestinationsIndependent(t *testing.T) {
	all := skillTargets(t.TempDir())

	if result := installSkillTargets(all[:1], false); result.Installations[0].Error != "" {
		t.Fatalf("first install failed: %s", result.Installations[0].Error)
	}
	if err := os.WriteFile(filepath.Join(all[0].Path, "SKILL.md"), []byte("hand edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result := installSkillTargets(all, false)
	if result.Name != "ggg" {
		t.Fatalf("name = %q, want %q", result.Name, "ggg")
	}
	if len(result.Installations) != 2 {
		t.Fatalf("installation count = %d, want 2", len(result.Installations))
	}
	if result.Installations[0].Error == "" {
		t.Fatal("modified destination did not report a conflict")
	}
	if result.Installations[1].Error != "" {
		t.Fatalf("second destination failed: %s", result.Installations[1].Error)
	}
	if _, err := os.Stat(filepath.Join(all[1].Path, "SKILL.md")); err != nil {
		t.Fatalf("second destination was not installed: %v", err)
	}
}

func TestSkillChecksAreSilentWhenNothingIsInstalled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if checks := skillChecks(); len(checks) != 0 {
		t.Fatalf("checks = %+v, want none", checks)
	}
}

func TestSkillChecksReportInstalledDestination(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	target := skillTargets(home)[1] // claude
	if result := installSkillTargets([]skillTarget{target}, false); result.Installations[0].Error != "" {
		t.Fatalf("install failed: %s", result.Installations[0].Error)
	}

	checks := skillChecks()
	if len(checks) != 1 {
		t.Fatalf("checks = %+v, want exactly one", checks)
	}
	if checks[0].Label != "AI agent skill" {
		t.Fatalf("label = %q, want %q", checks[0].Label, "AI agent skill")
	}
	if !checks[0].OK || checks[0].Warn {
		t.Fatalf("check = %+v, want OK without a warning", checks[0])
	}
	if !strings.Contains(checks[0].Message, target.Label) {
		t.Fatalf("message = %q, want it to mention %q", checks[0].Message, target.Label)
	}
}

func TestSkillChecksWarnAboutLocallyModifiedSkill(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	target := skillTargets(home)[0] // agents
	if result := installSkillTargets([]skillTarget{target}, false); result.Installations[0].Error != "" {
		t.Fatalf("install failed: %s", result.Installations[0].Error)
	}
	if err := os.WriteFile(filepath.Join(target.Path, "SKILL.md"), []byte("hand edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	checks := skillChecks()
	if len(checks) != 1 {
		t.Fatalf("checks = %+v, want exactly one", checks)
	}
	if !checks[0].Warn || checks[0].OK {
		t.Fatalf("check = %+v, want a warning", checks[0])
	}
	if !strings.Contains(checks[0].Message, "--force") {
		t.Fatalf("message = %q, want it to suggest --force", checks[0].Message)
	}
}
