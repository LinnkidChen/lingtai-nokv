package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAgentCodexEntriesShowsBoundedNoticeForNoKVBackedKnowledge(t *testing.T) {
	agentDir := t.TempDir()
	statusPath := filepath.Join(agentDir, "system", "storage.resolved.json")
	if err := os.MkdirAll(filepath.Dir(statusPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statusPath, []byte(`{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [{"mount": "knowledge", "backend": "nokv", "local_root": "`+filepath.ToSlash(filepath.Join(agentDir, "knowledge"))+`", "remote_root": "/remote/knowledge"}]
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := buildAgentCodexEntries(agentDir)
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want bounded notice: %+v", len(entries), entries)
	}
	if entries[0].Path != "" {
		t.Fatalf("NoKV-backed knowledge notice must not walk local files, got path %q", entries[0].Path)
	}
	if !strings.Contains(entries[0].Content, "NoKV-backed") {
		t.Fatalf("notice content = %q", entries[0].Content)
	}
}

func TestBuildKnowledgeFolderEntriesShowsBoundedNoticeForNoKVBackedKnowledgeRoot(t *testing.T) {
	agentDir := t.TempDir()
	knowledgeDir := filepath.Join(agentDir, "knowledge")
	statusPath := filepath.Join(agentDir, "system", "storage.resolved.json")
	if err := os.MkdirAll(filepath.Dir(statusPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statusPath, []byte(`{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [{"mount": "knowledge", "backend": "nokv", "local_root": "`+filepath.ToSlash(knowledgeDir)+`", "remote_root": "/remote/knowledge"}]
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := buildKnowledgeFolderEntries(knowledgeDir)
	if len(entries) != 1 || !strings.Contains(entries[0].Content, "NoKV-backed") {
		t.Fatalf("got entries=%+v, want bounded NoKV notice", entries)
	}
}
