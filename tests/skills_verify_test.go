package tests

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/illegalstudio/ggg/internal/testutil"
)

// runGGGSplit is runGGG with stdout and stderr captured separately, which the
// stale-skill notice contract requires: the notice lives on stderr and must
// never leak into stdout.
func runGGGSplit(t *testing.T, home string, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cmd := exec.Command(buildBinary(t), args...)
	cmd.Dir = moduleRoot(t)
	cmd.Env = testutil.ChildEnv(t, home)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.String(), errBuf.String(), err
}

func installSkill(t *testing.T, home, target string) {
	t.Helper()

	out, err := runGGG(t, home, "--json", "skills", "install", "--target", target)
	if err != nil {
		t.Fatalf("ggg --json skills install --target %s failed: %v\n%s", target, err, out)
	}
}

func writeMinimalConfig(t *testing.T, home string) {
	t.Helper()
	writeConfig(t, home, `
base_dir: ~/Developer
repos: []
`)
}

// TestCLISkillsVerifyEmitsPayloadOnStale covers the visible JSON contract:
// `ggg --json skills verify` exits 1 when an installed skill is stale, but
// still emits the full verifications payload instead of {"error": ...}.
func TestCLISkillsVerifyEmitsPayloadOnStale(t *testing.T) {
	home := testutil.SetupHome(t)

	installSkill(t, home, "agents")
	skillPath := filepath.Join(home, ".agents", "skills", "ggg", "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("hand edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, err := runGGGSplit(t, home, "--json", "skills", "verify")
	if err == nil {
		t.Fatalf("ggg --json skills verify succeeded on a stale skill:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stale --json verify wrote to stderr: %q", stderr)
	}

	var payload struct {
		Name          string `json:"name"`
		Verifications []struct {
			Target string `json:"target"`
			Status string `json:"status"`
			Error  string `json:"error"`
		} `json:"verifications"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("stale verify did not emit the verifications payload: %v\n%s", err, stdout)
	}
	if payload.Name != "ggg" || len(payload.Verifications) != 2 {
		t.Fatalf("unexpected payload shape: %+v\n%s", payload, stdout)
	}
	if got := payload.Verifications[0]; got.Target != "agents" || got.Status != "modified" || got.Error != "" {
		t.Fatalf("agents verification = %+v, want modified without error\n%s", got, stdout)
	}
	if got := payload.Verifications[1]; got.Target != "claude" || got.Status != "not-installed" {
		t.Fatalf("claude verification = %+v, want not-installed\n%s", got, stdout)
	}

	// The human run fails too, and its closing message must recommend --force
	// for a modified copy, not the plain install that would conflict.
	out, err := runGGG(t, home, "skills", "verify")
	if err == nil {
		t.Fatalf("ggg skills verify succeeded on a stale skill:\n%s", out)
	}
	if !strings.Contains(out, "modified") || !strings.Contains(out, "--force") {
		t.Fatalf("human verify output missing status or --force advice:\n%s", out)
	}
	if !strings.Contains(out, "suppress_skills_notice") {
		t.Fatalf("human verify output missing the suppression hint:\n%s", out)
	}

	// A command-level failure (unknown target) still uses the {"error": ...}
	// shape, with no verifications payload.
	stdout, _, err = runGGGSplit(t, home, "--json", "skills", "verify", "--target", "cursor")
	if err == nil {
		t.Fatalf("ggg skills verify with an unknown target succeeded:\n%s", stdout)
	}
	var errPayload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &errPayload); err != nil || !strings.Contains(errPayload.Error, "unknown skill target") {
		t.Fatalf("unknown target did not emit a JSON error: %v\n%s", err, stdout)
	}
}

// TestCLISkillNoticeFollowsInteractiveCommands covers the post-success notice
// contract end to end: when it appears, where it appears, and where it must
// never appear.
func TestCLISkillNoticeFollowsInteractiveCommands(t *testing.T) {
	home := testutil.SetupHome(t)
	writeMinimalConfig(t, home)
	const notice = "not in sync"

	// Nothing installed: no notice.
	_, stderr, err := runGGGSplit(t, home, "list")
	if err != nil {
		t.Fatalf("ggg list failed: %v", err)
	}
	if strings.Contains(stderr, notice) {
		t.Fatalf("notice appeared with no skill installed:\n%s", stderr)
	}

	// Install, then make the copy stale by editing it.
	installSkill(t, home, "agents")
	skillPath := filepath.Join(home, ".agents", "skills", "ggg", "SKILL.md")
	if err := os.WriteFile(skillPath, []byte("hand edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Interactive command: notice on stderr, stdout untouched, and the
	// message explains update, verify, and suppression.
	stdout, stderr, err := runGGGSplit(t, home, "list")
	if err != nil {
		t.Fatalf("ggg list failed: %v", err)
	}
	if !strings.Contains(stderr, notice) ||
		!strings.Contains(stderr, "ggg skills install --force") ||
		!strings.Contains(stderr, "ggg skills verify") ||
		!strings.Contains(stderr, "suppress_skills_notice") {
		t.Fatalf("stderr notice missing parts:\n%s", stderr)
	}
	if strings.Contains(stdout, notice) {
		t.Fatalf("notice leaked into stdout:\n%s", stdout)
	}

	// --json: machine output stays clean.
	stdout, stderr, err = runGGGSplit(t, home, "--json", "list")
	if err != nil {
		t.Fatalf("ggg --json list failed: %v", err)
	}
	if strings.Contains(stderr, notice) {
		t.Fatalf("notice appeared under --json:\n%s", stderr)
	}
	var listPayload struct {
		Repos []struct {
			URL string `json:"url"`
		} `json:"repos"`
	}
	if err := json.Unmarshal([]byte(stdout), &listPayload); err != nil {
		t.Fatalf("--json list output is not clean JSON: %v\n%s", err, stdout)
	}

	// Per-command help and shell completion internals: no notice.
	if _, stderr, err = runGGGSplit(t, home, "list", "--help"); err != nil {
		t.Fatalf("ggg list --help failed: %v", err)
	}
	if strings.Contains(stderr, notice) {
		t.Fatalf("notice appeared on --help:\n%s", stderr)
	}
	if _, stderr, err = runGGGSplit(t, home, "__complete", "de"); err != nil {
		t.Fatalf("ggg __complete failed: %v", err)
	}
	if strings.Contains(stderr, notice) {
		t.Fatalf("notice appeared on __complete:\n%s", stderr)
	}

	// suppress_skills_notice: true silences the reminder.
	writeConfig(t, home, `
base_dir: ~/Developer
suppress_skills_notice: true
repos: []
`)
	if _, stderr, err = runGGGSplit(t, home, "list"); err != nil {
		t.Fatalf("ggg list failed: %v", err)
	}
	if strings.Contains(stderr, notice) {
		t.Fatalf("notice appeared despite suppress_skills_notice:\n%s", stderr)
	}
}
