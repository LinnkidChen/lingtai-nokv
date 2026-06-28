package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/lingtai-tui/i18n"
)

func TestBuildAgentCodexEntriesShowsBoundedNoticeForNoKVBackedKnowledge(t *testing.T) {
	i18n.SetLang("en")
	t.Cleanup(func() { i18n.SetLang("en") })

	agentDir := t.TempDir()
	writeAgentFile(t, agentDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {
	      "mount": "knowledge",
	      "local_root": "`+filepath.ToSlash(filepath.Join(agentDir, "knowledge"))+`",
	      "backend": "nokv",
	      "remote_root": "/lingtai/projects/test-project/agents/alice/knowledge"
	    }
	  ]
	}`)

	entries := buildAgentCodexEntries(agentDir)
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want one bounded NoKV notice: %+v", len(entries), entries)
	}
	got := entries[0]
	if got.Path != "" {
		t.Fatalf("NoKV-backed knowledge notice must not be path-backed: %#v", got)
	}
	if got.Content == "" {
		t.Fatalf("NoKV-backed knowledge notice must provide bounded content: %#v", got)
	}
	text := strings.ToLower(got.Label + "\n" + got.Description + "\n" + got.Content)
	for _, want := range []string{"nokv", "knowledge", "not browsed"} {
		if !strings.Contains(text, want) {
			t.Fatalf("NoKV notice missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "aws_secret_access_key") || strings.Contains(text, "secret") {
		t.Fatalf("NoKV notice leaked secret wording:\n%s", text)
	}
}

func TestBuildKnowledgeFolderEntriesShowsBoundedNoticeForNoKVBackedKnowledgeRoot(t *testing.T) {
	i18n.SetLang("en")
	t.Cleanup(func() { i18n.SetLang("en") })

	agentDir := t.TempDir()
	knowledgeDir := filepath.Join(agentDir, "knowledge")
	writeAgentFile(t, agentDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {
	      "mount": "knowledge",
	      "local_root": "`+filepath.ToSlash(knowledgeDir)+`",
	      "backend": "nokv",
	      "remote_root": "/lingtai/projects/test-project/agents/alice/knowledge"
	    }
	  ]
	}`)

	entries := buildKnowledgeFolderEntries(knowledgeDir)
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want one bounded NoKV notice: %+v", len(entries), entries)
	}
	if entries[0].Path != "" {
		t.Fatalf("NoKV-backed knowledge root notice must not walk local or remote files: %#v", entries[0])
	}
	text := strings.ToLower(entries[0].Label + "\n" + entries[0].Description + "\n" + entries[0].Content)
	if !strings.Contains(text, "nokv") || !strings.Contains(text, "not browsed") {
		t.Fatalf("NoKV-backed knowledge root notice is not explicit enough:\n%s", text)
	}
}

func TestNoKVKnowledgeNoticeUsesLocaleStrings(t *testing.T) {
	agentDir := t.TempDir()
	writeAgentFile(t, agentDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {"mount": "knowledge", "backend": "nokv", "remote_root": "/lingtai/projects/test/agents/alice/knowledge"}
	  ]
	}`)

	i18n.SetLang("zh")
	t.Cleanup(func() { i18n.SetLang("en") })

	entries := buildAgentCodexEntries(agentDir)
	if len(entries) != 1 {
		t.Fatalf("entries = %+v, want one NoKV notice", entries)
	}
	text := entries[0].Label + "\n" + entries[0].Description + "\n" + entries[0].Content
	if !strings.Contains(text, "NoKV") || !strings.Contains(text, "knowledge") || !strings.Contains(text, "浏览") {
		t.Fatalf("NoKV notice did not use zh locale strings:\n%s", text)
	}
	if strings.Contains(text, "NoKV-backed knowledge is not browsed by TUI") {
		t.Fatalf("NoKV notice used hardcoded English:\n%s", text)
	}
}

func TestBuildAgentCodexEntriesUsesLocalKnowledgeWhenResolvedArtifactIsStale(t *testing.T) {
	agentDir := t.TempDir()
	knowledgeDir := filepath.Join(agentDir, "knowledge")
	writeAgentFile(t, agentDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {"mount": "knowledge", "backend": "nokv", "remote_root": "/lingtai/projects/test/agents/alice/knowledge"}
	  ]
	}`)
	time.Sleep(10 * time.Millisecond)
	writeAgentFile(t, agentDir, "init.json", `{"storage":{"enabled":false}}`)
	writeAgentFile(t, agentDir, "knowledge/local/KNOWLEDGE.md", "---\nname: local-note\ndescription: Local knowledge remains visible.\n---\n# Local\n")

	entries := buildAgentCodexEntries(agentDir)
	if len(entries) != 1 {
		t.Fatalf("entries = %+v, want local knowledge entry", entries)
	}
	if entries[0].Label != "local-note" || entries[0].Path != filepath.Join(knowledgeDir, "local", "KNOWLEDGE.md") {
		t.Fatalf("stale artifact suppressed local knowledge: %+v", entries[0])
	}
	if strings.Contains(strings.ToLower(entries[0].Label+"\n"+entries[0].Description+"\n"+entries[0].Content), "not browsed") {
		t.Fatalf("stale artifact produced NoKV notice instead of local entry: %+v", entries[0])
	}

	folderEntries := buildKnowledgeFolderEntries(knowledgeDir)
	if len(folderEntries) == 1 && strings.Contains(strings.ToLower(folderEntries[0].Content), "not browsed") {
		t.Fatalf("stale artifact suppressed local knowledge root: %+v", folderEntries[0])
	}
}

func TestBuildAgentCodexEntriesUsesLocalKnowledgeWhenResolvedArtifactSchemaInvalid(t *testing.T) {
	agentDir := t.TempDir()
	writeAgentFile(t, agentDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v0",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {"mount": "knowledge", "backend": "nokv"}
	  ]
	}`)
	writeAgentFile(t, agentDir, "knowledge/local/KNOWLEDGE.md", "---\nname: local-note\ndescription: Local knowledge remains visible.\n---\n# Local\n")

	entries := buildAgentCodexEntries(agentDir)
	if len(entries) != 1 || entries[0].Label != "local-note" {
		t.Fatalf("invalid schema should not suppress local knowledge: %+v", entries)
	}
}
