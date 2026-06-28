package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadStorageStatusReadsLocalResolvedArtifact(t *testing.T) {
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
	    {"mount": "artifacts", "local_root": "/tmp/a/artifacts", "backend": "nokv", "remote_root": "/remote/a/artifacts"}
	  ],
	  "nokv": {"metadata_addr": "127.0.0.1:7777", "bucket": "nokv", "endpoint": "http://127.0.0.1:9000"}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := ReadStorageStatus(agentDir)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Enabled || status.Backend != "routed" {
		t.Fatalf("status = %+v, want enabled routed", status)
	}
	if len(status.Routes) != 1 || status.Routes[0].Mount != "artifacts" {
		t.Fatalf("routes = %+v", status.Routes)
	}
	if status.NoKV.MetadataAddr != "127.0.0.1:7777" {
		t.Fatalf("nokv metadata addr = %q", status.NoKV.MetadataAddr)
	}
}

func TestReadStorageStatusReadsJsonlStreamMirrors(t *testing.T) {
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
	    {
	      "stream": "logs/events",
	      "local_path": "/tmp/a/logs/events.jsonl",
	      "backend": "nokv",
	      "remote_root": "/remote/a/logs/events",
	      "mode": "mirror"
	    }
	  ],
	  "nokv": {"metadata_addr": "127.0.0.1:7777", "bucket": "nokv", "endpoint": "http://127.0.0.1:9000"}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := ReadStorageStatus(agentDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(status.Streams) != 1 {
		t.Fatalf("streams = %+v, want one stream", status.Streams)
	}
	stream := status.Streams[0]
	if stream.Stream != "logs/events" || stream.Mode != "mirror" || stream.Backend != "nokv" {
		t.Fatalf("stream = %+v", stream)
	}
	if stream.RemoteRoot != "/remote/a/logs/events" {
		t.Fatalf("remote root = %q", stream.RemoteRoot)
	}
}

func TestReadStorageStatusReadsDegradedMirrorHealth(t *testing.T) {
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
	    "last_error": "RuntimeError: mirror write failed",
	    "last_error_stream": "logs/events"
	  },
	  "nokv": {"metadata_addr": "127.0.0.1:7777", "bucket": "nokv", "endpoint": "http://127.0.0.1:9000"}
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := ReadStorageStatus(agentDir)
	if err != nil {
		t.Fatal(err)
	}
	if status.Health == nil {
		t.Fatalf("health missing from storage status")
	}
	if status.Health.Status != "degraded" || status.Health.LastErrorStream != "logs/events" {
		t.Fatalf("health = %+v, want degraded logs/events", status.Health)
	}
	if status.Health.LastError != "RuntimeError: mirror write failed" {
		t.Fatalf("last error = %q", status.Health.LastError)
	}
}
