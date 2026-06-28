package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStorageStatusDoctorLinesReadLocalResolvedArtifact(t *testing.T) {
	agentDir := t.TempDir()
	statusPath := filepath.Join(agentDir, "system", "storage.resolved.json")
	if err := os.MkdirAll(filepath.Dir(statusPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statusPath, []byte(`{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {"mount": "knowledge", "local_root": "/tmp/a/knowledge", "backend": "nokv", "remote_root": "/remote/a/knowledge"}
	  ],
	  "nokv": {"metadata_addr": "127.0.0.1:7777", "bucket": "nokv", "endpoint": "http://127.0.0.1:9000"}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	lines := storageStatusDoctorLines(agentDir)
	joined := strings.Join(doctorLineTexts(lines), "\n")
	if !strings.Contains(joined, "NoKV storage: enabled") {
		t.Fatalf("doctor lines did not show enabled storage: %s", joined)
	}
	if !strings.Contains(joined, "knowledge") {
		t.Fatalf("doctor lines did not list route mounts: %s", joined)
	}
	if strings.Contains(joined, "secret") || strings.Contains(joined, "AWS_SECRET") {
		t.Fatalf("doctor lines leaked secret-like text: %s", joined)
	}
}

func TestStorageStatusDoctorLinesMissingArtifactIsUnknown(t *testing.T) {
	lines := storageStatusDoctorLines(t.TempDir())
	joined := strings.Join(doctorLineTexts(lines), "\n")
	if !strings.Contains(joined, "NoKV storage: unknown") {
		t.Fatalf("doctor lines should report unknown when artifact is missing: %s", joined)
	}
}

func TestStorageStatusDoctorLinesListJsonlStreamMirrors(t *testing.T) {
	agentDir := t.TempDir()
	statusPath := filepath.Join(agentDir, "system", "storage.resolved.json")
	if err := os.MkdirAll(filepath.Dir(statusPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statusPath, []byte(`{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [],
	  "streams": [
	    {"stream": "logs/events", "local_path": "/tmp/a/logs/events.jsonl", "backend": "nokv", "remote_root": "/remote/a/logs/events", "mode": "mirror"},
	    {"stream": "history/chat_history", "local_path": "/tmp/a/history/chat_history.jsonl", "backend": "nokv", "remote_root": "/remote/a/history/chat_history", "mode": "mirror"}
	  ],
	  "nokv": {"metadata_addr": "127.0.0.1:7777", "bucket": "nokv", "endpoint": "http://127.0.0.1:9000"}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	lines := storageStatusDoctorLines(agentDir)
	joined := strings.Join(doctorLineTexts(lines), "\n")
	if !strings.Contains(joined, "NoKV storage: enabled") {
		t.Fatalf("doctor lines did not show enabled storage: %s", joined)
	}
	if !strings.Contains(joined, "logs/events") || !strings.Contains(joined, "history/chat_history") {
		t.Fatalf("doctor lines did not list stream mirrors: %s", joined)
	}
	if strings.Contains(joined, "no active routes") {
		t.Fatalf("doctor lines treated configured stream mirrors as no routes: %s", joined)
	}
}

func TestStorageStatusDoctorLinesWarnForDegradedMirrorHealth(t *testing.T) {
	agentDir := t.TempDir()
	statusPath := filepath.Join(agentDir, "system", "storage.resolved.json")
	if err := os.MkdirAll(filepath.Dir(statusPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statusPath, []byte(`{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [],
	  "streams": [
	    {"stream": "logs/events", "local_path": "/tmp/a/logs/events.jsonl", "backend": "nokv", "remote_root": "/remote/a/logs/events", "mode": "mirror"}
	  ],
	  "health": {
	    "status": "degraded",
	    "backend": "mirror",
	    "streams": ["logs/events"],
	    "last_error": "RuntimeError: mirror write failed SECRET_TOKEN=abc123",
	    "last_error_stream": "logs/events"
	  },
	  "nokv": {"metadata_addr": "127.0.0.1:7777", "bucket": "nokv", "endpoint": "http://127.0.0.1:9000"}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	lines := storageStatusDoctorLines(agentDir)
	joined := strings.Join(doctorLineTexts(lines), "\n")
	if !strings.Contains(joined, "degraded") || !strings.Contains(joined, "logs/events") {
		t.Fatalf("doctor lines did not show degraded stream health: %s", joined)
	}
	if strings.Contains(joined, "SECRET_TOKEN") || strings.Contains(joined, "abc123") {
		t.Fatalf("doctor lines leaked secret-like text: %s", joined)
	}
}

func doctorLineTexts(lines []doctorLine) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, line.Text)
	}
	return out
}
