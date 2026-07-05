package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/anthropics/lingtai-tui/i18n"
)

// codexEntryIndex finds the entry index for a Codex account by its token-file
// path. Fails the test when absent.
func codexEntryIndex(t *testing.T, m LoginModel, absPath string) int {
	t.Helper()
	for i := range m.entries {
		if m.entries[i].Provider == "codex" && m.entries[i].CodexPath == absPath {
			return i
		}
	}
	t.Fatalf("could not find codex entry for %q; entries=%#v", absPath, m.entries)
	return -1
}

// TestLoginModel_DefaultPoolWeightForValidAccounts verifies that with NO pool
// file, a valid Codex account renders the default weight 1 and its row shows the
// "pool weight" label (not "disabled").
func TestLoginModel_DefaultPoolWeightForValidAccounts(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	_, globalDir := withTempCodexHome(t)

	acctPath := filepath.Join(globalDir, codexAuthSubdir, "work.json")
	writeCodexTokenAt(t, acctPath, "work@example.com")

	m := NewLoginModel("", globalDir)
	idx := codexEntryIndex(t, m, acctPath)
	m.entries[idx].Status = loginValid

	if w := m.codexEntryWeight(m.entries[idx]); w != 1 {
		t.Errorf("valid account with no pool file should default to weight 1; got %d", w)
	}

	m.width = 100
	view := m.View()
	wantLabel := strings.TrimSpace(i18n.T("login.codex_pool_weight"))
	// The format verb region ("pool weight: %d") — check the literal prefix.
	prefix := wantLabel
	if i := strings.Index(prefix, "%"); i >= 0 {
		prefix = strings.TrimSpace(prefix[:i])
	}
	if !strings.Contains(view, prefix) {
		t.Errorf("view should show the pool weight label %q; view=%q", prefix, view)
	}
}

// TestLoginModel_PlusIncrementsPoolWeight verifies the "+" key increases the
// selected Codex account's pool weight and persists it (lazy-writing the pool
// file), without touching any preset binding.
func TestLoginModel_PlusIncrementsPoolWeight(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	_, globalDir := withTempCodexHome(t)

	acctPath := filepath.Join(globalDir, codexAuthSubdir, "work.json")
	writeCodexTokenAt(t, acctPath, "work@example.com")

	m := NewLoginModel("", globalDir)
	idx := codexEntryIndex(t, m, acctPath)
	m.entries[idx].Status = loginValid
	m.cursor = idx

	// Default weight is 1; "+" should make it 2.
	m, cmd := m.Update(tea.KeyPressMsg{Text: "+", Code: '+'})
	if cmd != nil {
		t.Fatal("editing pool weight must not start a command")
	}
	if got := m.poolWeights[acctPath]; got != 2 {
		t.Fatalf("in-memory weight after + = %d, want 2", got)
	}
	// Persisted to disk with the relative ref.
	if got := codexPoolWeights(globalDir)[acctPath]; got != 2 {
		t.Fatalf("persisted weight after + = %d, want 2", got)
	}
}

// TestLoginModel_MinusClampsAtZero verifies "-" decrements and never goes below
// 0 (disabled).
func TestLoginModel_MinusClampsAtZero(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	_, globalDir := withTempCodexHome(t)

	acctPath := filepath.Join(globalDir, codexAuthSubdir, "work.json")
	writeCodexTokenAt(t, acctPath, "work@example.com")

	m := NewLoginModel("", globalDir)
	idx := codexEntryIndex(t, m, acctPath)
	m.entries[idx].Status = loginValid
	m.cursor = idx

	// From default 1: one "-" → 0, another "-" stays 0.
	m, _ = m.Update(tea.KeyPressMsg{Text: "-", Code: '-'})
	if got := m.poolWeights[acctPath]; got != 0 {
		t.Fatalf("weight after first - = %d, want 0", got)
	}
	m, _ = m.Update(tea.KeyPressMsg{Text: "-", Code: '-'})
	if got := m.poolWeights[acctPath]; got != 0 {
		t.Fatalf("weight after second - should clamp at 0; got %d", got)
	}
}

// TestLoginModel_ZeroDisablesAccount verifies "0" sets the weight straight to 0
// (disabled) and the row renders the disabled label.
func TestLoginModel_ZeroDisablesAccount(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	_, globalDir := withTempCodexHome(t)

	acctPath := filepath.Join(globalDir, codexAuthSubdir, "work.json")
	writeCodexTokenAt(t, acctPath, "work@example.com")

	m := NewLoginModel("", globalDir)
	idx := codexEntryIndex(t, m, acctPath)
	m.entries[idx].Status = loginValid
	m.cursor = idx

	// Bump to 3 first so we can prove "0" is absolute, not a decrement.
	m, _ = m.Update(tea.KeyPressMsg{Text: "+", Code: '+'})
	m, _ = m.Update(tea.KeyPressMsg{Text: "+", Code: '+'})
	if got := m.poolWeights[acctPath]; got != 3 {
		t.Fatalf("precondition: weight should be 3; got %d", got)
	}
	m, _ = m.Update(tea.KeyPressMsg{Text: "0", Code: '0'})
	if got := m.poolWeights[acctPath]; got != 0 {
		t.Fatalf("weight after 0 = %d, want 0", got)
	}

	m.width = 100
	view := m.View()
	if !strings.Contains(view, i18n.T("login.codex_pool_disabled")) {
		t.Errorf("disabled account row should show the disabled label; view=%q", view)
	}
}

// TestLoginModel_PoolEditDoesNotRewritePresets guards the core separation-of-
// concerns rule: editing a pool weight touches ONLY the pool file, never the
// active-account preset binding.
func TestLoginModel_PoolEditDoesNotRewritePresets(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	_, globalDir := withTempCodexHome(t)

	// A legacy-bound saved codex preset (no codex_auth_path).
	saveCodexPresetForTest(t, "codex-a", "")

	acctPath := filepath.Join(globalDir, codexAuthSubdir, "work.json")
	writeCodexTokenAt(t, acctPath, "work@example.com")

	m := NewLoginModel("", globalDir)
	idx := codexEntryIndex(t, m, acctPath)
	m.entries[idx].Status = loginValid
	m.cursor = idx

	m, _ = m.Update(tea.KeyPressMsg{Text: "+", Code: '+'})

	// The preset must remain legacy-bound — pool editing never rewrites it.
	if ref, ok := reloadSavedRef(t, "codex-a"); ok {
		t.Errorf("pool weight edit must not bind the preset; got codex_auth_path=%q", ref)
	}
}

// TestLoginModel_PoolWeightIgnoredOnVirtualRow verifies the pool keys are inert
// on the virtual "add account" row (no crash, no pool write).
func TestLoginModel_PoolWeightIgnoredOnVirtualRow(t *testing.T) {
	t.Setenv("LINGTAI_TUI_DIR", "")
	_, globalDir := withTempCodexHome(t)

	acctPath := filepath.Join(globalDir, codexAuthSubdir, "work.json")
	writeCodexTokenAt(t, acctPath, "work@example.com")

	m := NewLoginModel("", globalDir)
	// Put the cursor on the virtual add row (index == len(entries)).
	m.cursor = len(m.entries)
	if !m.cursorOnVirtualRow() {
		t.Fatalf("precondition: cursor should be on the virtual add row")
	}
	m, cmd := m.Update(tea.KeyPressMsg{Text: "+", Code: '+'})
	if cmd != nil {
		t.Fatal("pool key on virtual row must not start a command")
	}
	// No pool file should have been created.
	if w := codexPoolWeights(globalDir); len(w) != 0 {
		t.Errorf("virtual-row pool edit must not write the pool file; got %v", w)
	}
}
