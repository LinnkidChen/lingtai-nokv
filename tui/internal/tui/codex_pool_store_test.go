package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCodexPoolPath_DefaultUnderGlobalDir verifies the pool file lands directly
// under globalDir when LINGTAI_TUI_DIR is unset.
func TestCodexPoolPath_DefaultUnderGlobalDir(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	dir := t.TempDir()
	got := codexPoolPath(dir)
	want := filepath.Join(dir, codexPoolFileName)
	if got != want {
		t.Fatalf("pool path = %q, want %q", got, want)
	}
}

// TestCodexPoolPath_HonorsEnvOverride verifies LINGTAI_TUI_DIR wins over
// globalDir, matching the kernel reader's precedence.
func TestCodexPoolPath_HonorsEnvOverride(t *testing.T) {
	dir := t.TempDir()
	override := t.TempDir()
	t.Setenv("LINGTAI_TUI_DIR", override)
	got := codexPoolPath(dir)
	want := filepath.Join(override, codexPoolFileName)
	if got != want {
		t.Fatalf("pool path with env override = %q, want %q", got, want)
	}
}

// TestCodexPoolRefForPath_RelativeRefs verifies the legacy file serializes as
// "codex-auth.json" and per-account files as "codex-auth/<slug>.json" — stable
// TUI-dir-relative refs, never absolute and never token contents.
func TestCodexPoolRefForPath_RelativeRefs(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	dir := t.TempDir()

	legacy := legacyCodexAuthPath(dir)
	if ref := codexPoolRefForPath(dir, legacy); ref != "codex-auth.json" {
		t.Errorf("legacy pool ref = %q, want %q", ref, "codex-auth.json")
	}

	perAccount := filepath.Join(dir, codexAuthSubdir, "work.json")
	if ref := codexPoolRefForPath(dir, perAccount); ref != "codex-auth/work.json" {
		t.Errorf("per-account pool ref = %q, want %q", ref, "codex-auth/work.json")
	}
}

// TestCodexPoolRefRoundTrip verifies resolveCodexPoolRef inverts
// codexPoolRefForPath for both the legacy and per-account cases.
func TestCodexPoolRefRoundTrip(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	dir := t.TempDir()
	for _, abs := range []string{
		legacyCodexAuthPath(dir),
		filepath.Join(dir, codexAuthSubdir, "work.json"),
	} {
		ref := codexPoolRefForPath(dir, abs)
		if got := resolveCodexPoolRef(dir, ref); got != abs {
			t.Errorf("round-trip for %q: ref=%q resolved back to %q", abs, ref, got)
		}
	}
}

// TestSaveCodexPool_WritesOnlyRefsAndWeights is the core secrecy guarantee: the
// pool file holds version + relative refs + integer weights and NOTHING that
// looks like token material, even when the referenced token files contain
// secrets.
func TestSaveCodexPool_WritesOnlyRefsAndWeights(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	dir := t.TempDir()

	// Seed real-shaped (fake) token files so their contents exist on disk; the
	// pool file must not pick any of it up.
	writeStubCodexToken(t, legacyCodexAuthPath(dir), "primary@example.com")
	writeStubCodexToken(t, filepath.Join(dir, codexAuthSubdir, "work.json"), "work@example.com")

	pool := codexPool{
		Accounts: []codexPoolAccount{
			{Path: codexPoolRefForPath(dir, legacyCodexAuthPath(dir)), Weight: 1},
			{Path: codexPoolRefForPath(dir, filepath.Join(dir, codexAuthSubdir, "work.json")), Weight: 2},
		},
	}
	if err := saveCodexPool(dir, pool); err != nil {
		t.Fatalf("saveCodexPool: %v", err)
	}

	raw, err := os.ReadFile(codexPoolPath(dir))
	if err != nil {
		t.Fatalf("read pool file: %v", err)
	}
	body := string(raw)

	// No token material may appear.
	for _, secret := range []string{"stub-access", "stub-refresh", "access_token", "refresh_token"} {
		if strings.Contains(body, secret) {
			t.Errorf("pool file leaked secret-shaped content %q; body=%s", secret, body)
		}
	}
	// Refs must be the stable relative form, not absolute paths.
	if strings.Contains(body, dir) {
		t.Errorf("pool file contains an absolute path; body=%s", body)
	}
	if !strings.Contains(body, "codex-auth.json") || !strings.Contains(body, "codex-auth/work.json") {
		t.Errorf("pool file missing expected relative refs; body=%s", body)
	}

	// Version is stamped.
	var reloaded codexPool
	if err := json.Unmarshal(raw, &reloaded); err != nil {
		t.Fatalf("reparse pool: %v", err)
	}
	if reloaded.Version != codexPoolVersion {
		t.Errorf("version = %d, want %d", reloaded.Version, codexPoolVersion)
	}
	if len(reloaded.Accounts) != 2 {
		t.Fatalf("accounts = %d, want 2", len(reloaded.Accounts))
	}
}

// TestLoadCodexPool_MissingFileIsEmpty verifies a missing pool file is not an
// error — it yields an empty (versioned) pool so "no pool yet" reads cleanly.
func TestLoadCodexPool_MissingFileIsEmpty(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	dir := t.TempDir()
	pool, err := loadCodexPool(dir)
	if err != nil {
		t.Fatalf("missing pool file must not error; got %v", err)
	}
	if len(pool.Accounts) != 0 {
		t.Errorf("missing pool should have no accounts; got %d", len(pool.Accounts))
	}
	if pool.Version != codexPoolVersion {
		t.Errorf("missing pool version = %d, want %d", pool.Version, codexPoolVersion)
	}
}

// TestLoadCodexPool_MalformedFileErrors verifies a corrupt pool file surfaces an
// error rather than silently resetting the user's weights.
func TestLoadCodexPool_MalformedFileErrors(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	dir := t.TempDir()
	if err := os.WriteFile(codexPoolPath(dir), []byte("{not json"), 0o644); err != nil {
		t.Fatalf("seed malformed pool: %v", err)
	}
	if _, err := loadCodexPool(dir); err == nil {
		t.Fatal("malformed pool file should return a parse error")
	}
}

// TestCodexPoolWeightFor_DefaultsForMissingEntries verifies the default-weight
// policy: a valid account absent from the pool defaults to 1, an invalid one to
// 0, and a recorded weight (including 0) always wins.
func TestCodexPoolWeightFor_DefaultsForMissingEntries(t *testing.T) {
	weights := map[string]int{
		"/x/recorded.json":  3,
		"/x/explicit0.json": 0,
	}
	if w := codexPoolWeightFor(weights, "/x/valid-missing.json", true); w != 1 {
		t.Errorf("valid missing account default = %d, want 1", w)
	}
	if w := codexPoolWeightFor(weights, "/x/invalid-missing.json", false); w != 0 {
		t.Errorf("invalid missing account default = %d, want 0", w)
	}
	if w := codexPoolWeightFor(weights, "/x/recorded.json", true); w != 3 {
		t.Errorf("recorded weight = %d, want 3", w)
	}
	if w := codexPoolWeightFor(weights, "/x/explicit0.json", true); w != 0 {
		t.Errorf("explicit 0 (disabled) should win over the valid default; got %d", w)
	}
}

// TestSetCodexPoolWeight_LazyCreateAndUpdate verifies the lazy-write policy: the
// pool file is created on first edit, updated in place on subsequent edits, and
// unrelated accounts survive.
func TestSetCodexPoolWeight_LazyCreateAndUpdate(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	dir := t.TempDir()
	poolFile := codexPoolPath(dir)

	if _, err := os.Stat(poolFile); !os.IsNotExist(err) {
		t.Fatalf("precondition: pool file should not exist yet; stat err=%v", err)
	}

	work := filepath.Join(dir, codexAuthSubdir, "work.json")
	home := filepath.Join(dir, codexAuthSubdir, "home.json")

	// First edit creates the file.
	if err := setCodexPoolWeight(dir, work, 2); err != nil {
		t.Fatalf("set work weight: %v", err)
	}
	if _, err := os.Stat(poolFile); err != nil {
		t.Fatalf("pool file should be created on first edit; stat err=%v", err)
	}

	// Second account is added, not replacing the first.
	if err := setCodexPoolWeight(dir, home, 5); err != nil {
		t.Fatalf("set home weight: %v", err)
	}
	// Update the first in place.
	if err := setCodexPoolWeight(dir, work, 4); err != nil {
		t.Fatalf("update work weight: %v", err)
	}

	weights := codexPoolWeights(dir)
	if weights[work] != 4 {
		t.Errorf("work weight = %d, want 4", weights[work])
	}
	if weights[home] != 5 {
		t.Errorf("home weight = %d, want 5", weights[home])
	}
	if len(weights) != 2 {
		t.Errorf("expected exactly 2 pool accounts; got %d (%v)", len(weights), weights)
	}

	// The stored refs must be the relative form.
	pool, err := loadCodexPool(dir)
	if err != nil {
		t.Fatalf("reload pool: %v", err)
	}
	for _, acct := range pool.Accounts {
		if filepath.IsAbs(acct.Path) || strings.HasPrefix(acct.Path, "~") {
			t.Errorf("pool ref should be TUI-dir-relative; got %q", acct.Path)
		}
	}
}

// TestSetCodexPoolWeight_ClampsNegative verifies weights never go below 0.
func TestSetCodexPoolWeight_ClampsNegative(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	dir := t.TempDir()
	work := filepath.Join(dir, codexAuthSubdir, "work.json")
	if err := setCodexPoolWeight(dir, work, -3); err != nil {
		t.Fatalf("set negative weight: %v", err)
	}
	if w := codexPoolWeights(dir)[work]; w != 0 {
		t.Errorf("negative weight should clamp to 0; got %d", w)
	}
}
