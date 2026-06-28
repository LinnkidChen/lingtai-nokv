package fs

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestReadStorageStatusReturnsLocalWhenArtifactMissing(t *testing.T) {
	dir := t.TempDir()

	status, err := ReadStorageStatus(dir)
	if err != nil {
		t.Fatalf("ReadStorageStatus: %v", err)
	}
	if status.Backend != "local" || status.Enabled {
		t.Fatalf("status = %#v, want disabled local fallback", status)
	}
}

func TestReadStorageStatusReadsLocalResolvedArtifact(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {
	      "mount": "knowledge",
	      "local_root": "/tmp/project/.lingtai/alice/knowledge",
	      "backend": "nokv",
	      "remote_root": "/lingtai/projects/test-project/agents/alice/knowledge"
	    }
	  ],
	  "nokv": {
	    "metadata_addr": "127.0.0.1:7777",
	    "bucket": "nokv",
	    "endpoint": "http://127.0.0.1:9000"
	  }
	}`)

	status, err := ReadStorageStatus(dir)
	if err != nil {
		t.Fatalf("ReadStorageStatus: %v", err)
	}
	if !status.Enabled || status.Backend != "routed" {
		t.Fatalf("status = %#v, want enabled routed storage", status)
	}
	if len(status.Routes) != 1 {
		t.Fatalf("routes = %#v, want one route", status.Routes)
	}
	if status.Routes[0].Mount != "knowledge" || status.Routes[0].Backend != "nokv" {
		t.Fatalf("route = %#v, want knowledge nokv route", status.Routes[0])
	}
	if status.Nokv.MetadataAddr != "127.0.0.1:7777" || status.Nokv.Endpoint != "http://127.0.0.1:9000" {
		t.Fatalf("nokv display metadata = %#v", status.Nokv)
	}
	if raw := status.String(); strings.Contains(raw, "AWS_SECRET_ACCESS_KEY") || strings.Contains(raw, "secret") {
		t.Fatalf("storage status string leaked secret material: %s", raw)
	}
}

func TestReadStorageStatusRejectsUnsupportedSchema(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v0",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {"mount": "knowledge", "backend": "nokv", "remote_root": "/secret/namespace"}
	  ]
	}`)

	_, err := ReadStorageStatus(dir)
	if err == nil {
		t.Fatal("ReadStorageStatus succeeded, want unsupported schema error")
	}
	if !strings.Contains(err.Error(), "unsupported storage.resolved.json schema") {
		t.Fatalf("error = %v, want unsupported schema", err)
	}
}

func TestReadStorageStatusRejectsStaleArtifact(t *testing.T) {
	dir := t.TempDir()
	writeAgentFile(t, dir, "init.json", `{}`)
	writeAgentFile(t, dir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed"
	}`)
	now := time.Now()
	touchAgentFile(t, dir, "system/storage.resolved.json", now.Add(-2*time.Hour))
	touchAgentFile(t, dir, "init.json", now)

	_, err := ReadStorageStatus(dir)
	if err == nil {
		t.Fatal("ReadStorageStatus succeeded, want stale artifact error")
	}
	if !strings.Contains(err.Error(), "stale storage.resolved.json") {
		t.Fatalf("error = %v, want stale artifact", err)
	}
	if os.IsNotExist(err) {
		t.Fatalf("error = %v, want stale artifact rather than missing artifact", err)
	}
}
