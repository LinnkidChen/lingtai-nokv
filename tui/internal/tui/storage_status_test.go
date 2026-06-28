package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/lingtai-tui/i18n"
)

func TestStorageStatusDoctorLinesReadLocalResolvedArtifact(t *testing.T) {
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
	    },
	    {
	      "mount": "reports",
	      "local_root": "`+filepath.ToSlash(filepath.Join(agentDir, "reports"))+`",
	      "backend": "local"
	    }
	  ],
	  "nokv": {
	    "metadata_addr": "127.0.0.1:7777",
	    "bucket": "nokv",
	    "endpoint": "http://127.0.0.1:9000"
	  }
	}`)

	lines := storageStatusDoctorLines(agentDir)
	text := joinDoctorLineText(lines)
	for _, want := range []string{
		"storage: routed",
		"knowledge -> nokv",
		"/lingtai/projects/test-project/agents/alice/knowledge",
		"reports -> local",
		"127.0.0.1:7777",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("storage status output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "AWS_SECRET_ACCESS_KEY") || strings.Contains(text, "super-secret") {
		t.Fatalf("storage status output leaked secret material:\n%s", text)
	}
}

func TestStorageStatusDoctorLinesUsesLocaleStrings(t *testing.T) {
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

	text := joinDoctorLineText(storageStatusDoctorLines(agentDir))
	for _, want := range []string{"存储", "knowledge -> nokv"} {
		if !strings.Contains(text, want) {
			t.Fatalf("localized storage status missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "storage: routed") {
		t.Fatalf("doctor storage status used hardcoded English:\n%s", text)
	}
}

func TestStorageStatusDoctorLinesWarnsOnStaleResolvedArtifact(t *testing.T) {
	i18n.SetLang("en")
	t.Cleanup(func() { i18n.SetLang("en") })

	agentDir := t.TempDir()
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

	lines := storageStatusDoctorLines(agentDir)
	text := joinDoctorLineText(lines)
	if !strings.Contains(text, "unknown") || !strings.Contains(text, "stale") {
		t.Fatalf("stale storage artifact should warn unknown/stale, got:\n%s", text)
	}
	if strings.Contains(text, "storage: routed") {
		t.Fatalf("stale storage artifact must not report routed:\n%s", text)
	}
	if len(lines) == 0 || !lines[0].Warn {
		t.Fatalf("stale storage artifact should be a warning: %+v", lines)
	}
}

func TestStorageStatusDoctorLinesWarnsOnInvalidSchema(t *testing.T) {
	i18n.SetLang("en")
	t.Cleanup(func() { i18n.SetLang("en") })

	agentDir := t.TempDir()
	writeAgentFile(t, agentDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v0",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {"mount": "knowledge", "backend": "nokv"}
	  ]
	}`)

	text := joinDoctorLineText(storageStatusDoctorLines(agentDir))
	if !strings.Contains(text, "unknown") || !strings.Contains(text, "schema") {
		t.Fatalf("invalid schema should warn unknown/schema, got:\n%s", text)
	}
	if strings.Contains(text, "storage: routed") {
		t.Fatalf("invalid schema must not report routed:\n%s", text)
	}
}

func joinDoctorLineText(lines []doctorLine) string {
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(line.Text)
		b.WriteByte('\n')
	}
	return b.String()
}
