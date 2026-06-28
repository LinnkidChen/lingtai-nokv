package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupPortalTestNetwork(t *testing.T) string {
	t.Helper()
	base := t.TempDir()

	// alice: active agent, has a ledger entry for bob (relative path)
	aliceDir := filepath.Join(base, "alice")
	os.MkdirAll(filepath.Join(aliceDir, "delegates"), 0o755)
	os.MkdirAll(filepath.Join(aliceDir, "mailbox", "inbox"), 0o755)
	writeAgentManifest(t, aliceDir, "alice", false)

	// ledger entry uses relative path — ReadLedger will resolve to absolute
	ledger := `{"event":"avatar","name":"bob","working_dir":"bob","ts":1000}`
	os.WriteFile(filepath.Join(aliceDir, "delegates", "ledger.jsonl"), []byte(ledger+"\n"), 0o644)

	// bob: discovered by DiscoverAgents (relative address from .agent.json)
	bobDir := filepath.Join(base, "bob")
	os.MkdirAll(filepath.Join(bobDir, "mailbox", "inbox"), 0o755)
	writeAgentManifest(t, bobDir, "bob", false)
	writeHeartbeat(t, bobDir)

	// human: discovered by DiscoverAgents (relative address)
	humanDir := filepath.Join(base, "human")
	os.MkdirAll(filepath.Join(humanDir, "mailbox", "inbox"), 0o755)
	writeAgentManifest(t, humanDir, "human", true)

	return base
}

func writeHeartbeat(t *testing.T, dir string) {
	t.Helper()
	content := time.Now().Format(time.RFC3339)
	os.WriteFile(filepath.Join(dir, ".agent.heartbeat"), []byte(content), 0o644)
}

func TestBuildNetwork_Portal(t *testing.T) {
	base := setupPortalTestNetwork(t)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("build network: %v", err)
	}

	if len(net.Nodes) != 3 {
		t.Errorf("nodes = %d, want 3", len(net.Nodes))
	}
}

func TestBuildNetwork_AllAddressesRelative(t *testing.T) {
	base := setupPortalTestNetwork(t)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("build network: %v", err)
	}

	for _, n := range net.Nodes {
		if len(n.Address) > 0 && n.Address[0] == '/' {
			t.Errorf("node %s has absolute address: %s", n.AgentName, n.Address)
		}
	}
}

// Regression test: ledger entries using relative paths must be relativized
// so they don't create duplicate nodes alongside DiscoverAgents entries.
func TestBuildNetwork_NoDuplicateNodesFromLedger(t *testing.T) {
	base := setupPortalTestNetwork(t)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("build network: %v", err)
	}

	// Count nodes by address — no duplicates allowed
	seen := make(map[string]bool)
	for _, n := range net.Nodes {
		if seen[n.Address] {
			t.Errorf("duplicate node address: %s", n.Address)
		}
		seen[n.Address] = true
	}
}

func TestBuildNetwork_WorkingDirAlwaysAbsolute(t *testing.T) {
	base := setupPortalTestNetwork(t)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("build network: %v", err)
	}

	for _, n := range net.Nodes {
		if !filepath.IsAbs(n.WorkingDir) {
			t.Errorf("node %s has relative WorkingDir: %s", n.AgentName, n.WorkingDir)
		}
	}
}

func TestBuildNetwork_OmitsGhostAvatarNodeFromLedger(t *testing.T) {
	base := t.TempDir()

	aliceDir := filepath.Join(base, "alice")
	os.MkdirAll(filepath.Join(aliceDir, "delegates"), 0o755)
	writeAgentManifest(t, aliceDir, "alice", false)

	// carol is referenced from ledger but has no discoverable .agent.json
	ledger := `{"event":"avatar","name":"carol","working_dir":"carol","ts":1000}` + "\n" +
		`{"event":"avatar","name":"bob","working_dir":"bob","ts":1001}` + "\n"
	os.WriteFile(filepath.Join(aliceDir, "delegates", "ledger.jsonl"), []byte(ledger), 0o644)

	bobDir := filepath.Join(base, "bob")
	writeAgentManifest(t, bobDir, "bob", false)
	writeHeartbeat(t, bobDir)

	humanDir := filepath.Join(base, "human")
	writeAgentManifest(t, humanDir, "human", true)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("build network: %v", err)
	}

	if len(net.Nodes) != 3 {
		t.Fatalf("nodes = %d, want 3", len(net.Nodes))
	}

	for _, n := range net.Nodes {
		if n.Address == "carol" {
			t.Fatalf("unexpected ghost avatar node present: %+v", n)
		}
	}

	// Edge for carol must be dropped; only alice -> bob must remain.
	if len(net.AvatarEdges) != 1 {
		t.Fatalf("avatar edges = %d, want 1", len(net.AvatarEdges))
	}
	if net.AvatarEdges[0].Parent != "alice" || net.AvatarEdges[0].Child != "bob" {
		t.Errorf("unexpected avatar edge: %+v", net.AvatarEdges[0])
	}
}

func TestBuildNetwork_IncludesAvatarEdgeForLiveChild(t *testing.T) {
	base := t.TempDir()

	aliceDir := filepath.Join(base, "alice")
	os.MkdirAll(filepath.Join(aliceDir, "delegates"), 0o755)
	writeAgentManifest(t, aliceDir, "alice", false)

	ledger := `{"event":"avatar","name":"bob","working_dir":"bob","ts":1000}` + "\n"
	os.WriteFile(filepath.Join(aliceDir, "delegates", "ledger.jsonl"), []byte(ledger), 0o644)

	bobDir := filepath.Join(base, "bob")
	writeAgentManifest(t, bobDir, "bob", false)
	writeHeartbeat(t, bobDir)

	humanDir := filepath.Join(base, "human")
	writeAgentManifest(t, humanDir, "human", true)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("build network: %v", err)
	}

	if len(net.AvatarEdges) != 1 {
		t.Fatalf("avatar edges = %d, want 1", len(net.AvatarEdges))
	}

	e := net.AvatarEdges[0]
	if e.Parent != "alice" || e.Child != "bob" || e.ChildName != "bob" {
		t.Errorf("unexpected avatar edge: %+v", e)
	}
}

func TestBuildNetwork_ContactEdgesPreserved(t *testing.T) {
	base := t.TempDir()

	aliceDir := filepath.Join(base, "alice")
	os.MkdirAll(filepath.Join(aliceDir, "delegates"), 0o755)
	writeAgentManifest(t, aliceDir, "alice", false)

	// carol is dead reference in ledger, alive in contacts
	ledger := `{"event":"avatar","name":"carol","working_dir":"carol","ts":1000}` + "\n"
	os.WriteFile(filepath.Join(aliceDir, "delegates", "ledger.jsonl"), []byte(ledger), 0o644)

	contacts := []contactRecord{
		{Address: "bob", Name: "Bob"},
		{Address: "carol", Name: "Carol"},
	}
	data, _ := json.Marshal(contacts)
	os.MkdirAll(filepath.Join(aliceDir, "mailbox"), 0o755)
	os.WriteFile(filepath.Join(aliceDir, "mailbox", "contacts.json"), data, 0o644)

	bobDir := filepath.Join(base, "bob")
	writeAgentManifest(t, bobDir, "bob", false)
	writeHeartbeat(t, bobDir)

	humanDir := filepath.Join(base, "human")
	writeAgentManifest(t, humanDir, "human", true)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("build network: %v", err)
	}

	// Nodes: alice + bob + human. No ghost node for carol.
	if len(net.Nodes) != 3 {
		t.Fatalf("nodes = %d, want 3", len(net.Nodes))
	}
	if len(net.AvatarEdges) != 0 {
		t.Fatalf("avatar edges = %d, want 0", len(net.AvatarEdges))
	}

	if len(net.ContactEdges) != 2 {
		t.Fatalf("contact edges = %d, want 2", len(net.ContactEdges))
	}

	got := map[string]bool{}
	for _, e := range net.ContactEdges {
		if e.Owner != "alice" {
			t.Errorf("unexpected contact owner: %s", e.Owner)
		}
		got[e.Target] = true
	}
	if !got["bob"] || !got["carol"] {
		t.Errorf("expected alice -> bob and alice -> carol contact edges, got %+v", net.ContactEdges)
	}
}

func TestBuildNetwork_MailEdgesUnaffected(t *testing.T) {
	base := t.TempDir()

	aliceDir := filepath.Join(base, "alice")
	os.MkdirAll(filepath.Join(aliceDir, "delegates"), 0o755)
	writeAgentManifest(t, aliceDir, "alice", false)

	// dead child referenced in ledger but missing from discovery
	ledger := `{"event":"avatar","name":"ghost","working_dir":"ghost","ts":1000}` + "\n"
	os.WriteFile(filepath.Join(aliceDir, "delegates", "ledger.jsonl"), []byte(ledger), 0o644)

	ghostDir := filepath.Join(base, "ghost")
	os.MkdirAll(filepath.Join(ghostDir, "mailbox", "inbox"), 0o755)
	writeMailMessage(t, ghostDir, "inbox", "msg-1", MailMessage{
		ID:         "msg-1",
		From:       "ghost",
		To:         "alice",
		ReceivedAt: time.Now().Format(time.RFC3339),
	})

	humanDir := filepath.Join(base, "human")
	writeAgentManifest(t, humanDir, "human", true)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("build network: %v", err)
	}

	// ghost node must not be added
	for _, n := range net.Nodes {
		if n.Address == "ghost" {
			t.Fatalf("unexpected ghost node present: %+v", n)
		}
	}

	if len(net.MailEdges) != 0 {
		t.Fatalf("mail edges = %d, want 0", len(net.MailEdges))
	}
}

func TestBuildNetwork_IncludesSecretSafeStorageSummary(t *testing.T) {
	base := t.TempDir()
	aliceDir := filepath.Join(base, "alice")
	writeAgentManifest(t, aliceDir, "alice", false)
	writeHeartbeat(t, aliceDir)
	writeAgentFile(t, aliceDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {
	      "mount": "knowledge",
	      "local_root": "`+filepath.ToSlash(filepath.Join(aliceDir, "knowledge"))+`",
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

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("BuildNetwork: %v", err)
	}
	if len(net.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(net.Nodes))
	}
	node := net.Nodes[0]
	if node.Storage == nil {
		t.Fatalf("node storage missing: %+v", node)
	}
	if node.Storage.Backend != "routed" {
		t.Fatalf("storage backend = %q, want routed", node.Storage.Backend)
	}
	if len(node.Storage.Routes) != 1 {
		t.Fatalf("storage routes = %#v, want one route", node.Storage.Routes)
	}
	route := node.Storage.Routes[0]
	if route.Mount != "knowledge" || route.Backend != "nokv" || route.RemoteRoot != "/lingtai/projects/test-project/agents/alice/knowledge" {
		t.Fatalf("storage route = %#v, want NoKV-backed knowledge route", route)
	}

	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	payload := string(raw)
	for _, forbidden := range []string{aliceDir, "local_root", "WorkingDir", "AWS_SECRET_ACCESS_KEY", "secret"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("storage network payload leaked %q: %s", forbidden, payload)
		}
	}
	if !strings.Contains(payload, `"storage"`) || !strings.Contains(payload, `"knowledge"`) || !strings.Contains(payload, `"nokv"`) {
		t.Fatalf("storage network payload missing NoKV knowledge summary: %s", payload)
	}
}

func TestBuildNetwork_IncludesUnknownStorageSummaryForMalformedArtifact(t *testing.T) {
	base := t.TempDir()
	aliceDir := filepath.Join(base, "alice")
	writeAgentManifest(t, aliceDir, "alice", false)
	writeHeartbeat(t, aliceDir)
	writeAgentFile(t, aliceDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "routes": [
	    {"mount": "knowledge", "local_root": "`+filepath.ToSlash(filepath.Join(aliceDir, "knowledge"))+`", "backend": "nokv"}
	  ],
	  "secret": "super-secret"
	`)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("BuildNetwork: %v", err)
	}
	if len(net.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(net.Nodes))
	}
	node := net.Nodes[0]
	if node.Storage == nil {
		t.Fatalf("node storage missing for malformed artifact: %+v", node)
	}
	if node.Storage.Backend != "unknown" || node.Storage.Error == "" {
		t.Fatalf("storage summary = %#v, want unknown backend with error", node.Storage)
	}
	if len(node.Storage.Routes) != 0 {
		t.Fatalf("malformed artifact must not expose routes: %#v", node.Storage.Routes)
	}
	raw, err := json.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}
	payload := string(raw)
	for _, forbidden := range []string{aliceDir, "local_root", "super-secret", "secret"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("malformed storage payload leaked %q: %s", forbidden, payload)
		}
	}
	if !strings.Contains(payload, `"storage"`) || !strings.Contains(payload, `"unknown"`) || !strings.Contains(payload, `"error"`) {
		t.Fatalf("malformed storage payload missing unknown/error summary: %s", payload)
	}
}

func TestBuildNetwork_IncludesUnknownStorageSummaryForUnsupportedSchema(t *testing.T) {
	base := t.TempDir()
	aliceDir := filepath.Join(base, "alice")
	writeAgentManifest(t, aliceDir, "alice", false)
	writeHeartbeat(t, aliceDir)
	writeAgentFile(t, aliceDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v0",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {
	      "mount": "knowledge",
	      "local_root": "`+filepath.ToSlash(filepath.Join(aliceDir, "knowledge"))+`",
	      "backend": "nokv",
	      "remote_root": "/lingtai/projects/test-project/agents/alice/knowledge"
	    }
	  ]
	}`)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("BuildNetwork: %v", err)
	}
	if len(net.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(net.Nodes))
	}
	storage := net.Nodes[0].Storage
	if storage == nil {
		t.Fatalf("node storage missing: %+v", net.Nodes[0])
	}
	if storage.Backend != "unknown" || storage.Error == "" {
		t.Fatalf("storage summary = %#v, want unknown backend with error", storage)
	}
	if len(storage.Routes) != 0 {
		t.Fatalf("unsupported schema must not expose routes: %#v", storage.Routes)
	}

	raw, err := json.Marshal(net.Nodes[0])
	if err != nil {
		t.Fatal(err)
	}
	payload := string(raw)
	for _, forbidden := range []string{aliceDir, "local_root"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("unsupported schema payload leaked %q: %s", forbidden, payload)
		}
	}
}

func TestBuildNetwork_IncludesUnknownStorageSummaryForStaleArtifact(t *testing.T) {
	base := t.TempDir()
	aliceDir := filepath.Join(base, "alice")
	writeAgentManifest(t, aliceDir, "alice", false)
	writeHeartbeat(t, aliceDir)
	writeAgentFile(t, aliceDir, "init.json", `{}`)
	writeAgentFile(t, aliceDir, "system/storage.resolved.json", `{
	  "schema": "lingtai.storage.resolved/v1",
	  "enabled": true,
	  "backend": "routed",
	  "routes": [
	    {
	      "mount": "knowledge",
	      "local_root": "`+filepath.ToSlash(filepath.Join(aliceDir, "knowledge"))+`",
	      "backend": "nokv",
	      "remote_root": "/lingtai/projects/test-project/agents/alice/knowledge"
	    }
	  ]
	}`)
	now := time.Now()
	touchAgentFile(t, aliceDir, "system/storage.resolved.json", now.Add(-2*time.Hour))
	touchAgentFile(t, aliceDir, "init.json", now)

	net, err := BuildNetwork(base)
	if err != nil {
		t.Fatalf("BuildNetwork: %v", err)
	}
	if len(net.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(net.Nodes))
	}
	storage := net.Nodes[0].Storage
	if storage == nil {
		t.Fatalf("node storage missing: %+v", net.Nodes[0])
	}
	if storage.Backend != "unknown" || storage.Error == "" {
		t.Fatalf("storage summary = %#v, want unknown backend with error", storage)
	}
	if len(storage.Routes) != 0 {
		t.Fatalf("stale artifact must not expose routes: %#v", storage.Routes)
	}

	raw, err := json.Marshal(net.Nodes[0])
	if err != nil {
		t.Fatal(err)
	}
	payload := string(raw)
	for _, forbidden := range []string{aliceDir, "local_root"} {
		if strings.Contains(payload, forbidden) {
			t.Fatalf("stale artifact payload leaked %q: %s", forbidden, payload)
		}
	}
}
