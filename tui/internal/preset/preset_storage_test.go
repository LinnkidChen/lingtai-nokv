package preset

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateInitJSONWithOptsPreservesStorageBlockWithoutSecretExpansion(t *testing.T) {
	lingtaiDir := t.TempDir()
	globalDir := t.TempDir()
	agentDir := filepath.Join(lingtaiDir, "alice")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}

	storage := map[string]interface{}{
		"enabled": true,
		"backend": "nokv",
		"nokv": map[string]interface{}{
			"namespace_root":        "/lingtai/projects/${project_hash}/agents/${agent_name}",
			"metadata_addr_env":     "NOKV_METADATA_ADDR",
			"bucket_env":            "NOKV_BUCKET",
			"endpoint_env":          "NOKV_ENDPOINT",
			"access_key_id_env":     "AWS_ACCESS_KEY_ID",
			"secret_access_key_env": "AWS_SECRET_ACCESS_KEY",
			"region_env":            "AWS_REGION",
		},
		"mounts": []interface{}{"artifacts", "reports", "checkpoints", "knowledge"},
	}
	seed := map[string]interface{}{
		"manifest": map[string]interface{}{},
		"storage":  storage,
	}
	data, _ := json.Marshal(seed)
	if err := os.WriteFile(filepath.Join(agentDir, "init.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := GenerateInitJSONWithOpts(DefaultPreset(), "alice", "alice", lingtaiDir, globalDir, DefaultAgentOpts()); err != nil {
		t.Fatalf("GenerateInitJSONWithOpts: %v", err)
	}

	updated, err := os.ReadFile(filepath.Join(agentDir, "init.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(updated, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got["storage"], storage) {
		t.Fatalf("storage block was not preserved:\n got=%#v\nwant=%#v", got["storage"], storage)
	}
	if raw := string(updated); raw == "" || strings.Contains(raw, "rustfsadmin") {
		t.Fatalf("init.json should store env var names only, not secret values")
	}
}
