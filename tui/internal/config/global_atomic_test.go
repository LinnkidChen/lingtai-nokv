package config

// Regression tests for issue #508: config.json, tui_config.json and .env
// must be written atomically (temp + rename) so an interrupted write can
// never truncate a file that holds API keys.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assertNoTempLeftovers fails if an atomic-write temp file survived in dir.
func assertNoTempLeftovers(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
}

func TestSaveConfigAtomicRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{}
	if err := SaveConfig(dir, cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(dir); err != nil {
		t.Fatalf("LoadConfig after SaveConfig: %v", err)
	}
	info, err := os.Stat(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("config.json perm = %o, want 0600", info.Mode().Perm())
	}
	assertNoTempLeftovers(t, dir)
}

func TestSaveTUIConfigAtomicRoundTrip(t *testing.T) {
	dir := t.TempDir()
	tc := DefaultTUIConfig()
	tc.Language = "zh"
	if err := SaveTUIConfig(dir, tc); err != nil {
		t.Fatal(err)
	}
	if got := LoadTUIConfig(dir); got.Language != "zh" {
		t.Errorf("Language = %q, want zh", got.Language)
	}
	assertNoTempLeftovers(t, dir)
}

func TestWriteEnvLinesPreservesPermsAtomically(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("A=1\n"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := writeEnvLines(path, []string{"A=2", "B=3"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "A=2\nB=3\n" {
		t.Errorf("content = %q", string(data))
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o640 {
		t.Errorf(".env perm = %o, want 0640 (user-tightened perms must survive)", info.Mode().Perm())
	}
	assertNoTempLeftovers(t, dir)
}

// TestAtomicWriteFileKeepsOldContentOnFailure pins the property the issue is
// about: when the new content cannot be written, the old file must survive
// untouched rather than being truncated.
func TestAtomicWriteFileKeepsOldContentOnFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root — read-only dirs are still writable")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"old":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	if err := atomicWriteFile(path, []byte(`{"new":true}`), 0o600); err == nil {
		t.Fatal("expected error writing into read-only dir, got nil")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"old":true}` {
		t.Errorf("old content must survive a failed write, got %q", string(data))
	}
}
