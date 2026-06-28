package preset

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateInitJSONWithOptsPreservesStorageBlockWithoutSecretExpansion(t *testing.T) {
	t.Setenv("AWS_SECRET_ACCESS_KEY", "super-secret-value")
	t.Setenv("NOKV_BUCKET", "secret-bucket-value")

	tmp := t.TempDir()
	globalDir := filepath.Join(tmp, "global")
	lingtaiDir := filepath.Join(tmp, "project", ".lingtai")
	agentDir := filepath.Join(lingtaiDir, "alice")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}

	storage := map[string]interface{}{
		"enabled": true,
		"backend": "nokv",
		"nokv": map[string]interface{}{
			"namespace_root":        "/lingtai/projects/test-project/agents/alice",
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
		"manifest": map[string]interface{}{
			"agent_name": "alice",
			"preset": map[string]interface{}{
				"active":  "~/.lingtai-tui/presets/templates/codex.json",
				"default": "~/.lingtai-tui/presets/templates/codex.json",
				"allowed": []interface{}{"~/.lingtai-tui/presets/templates/codex.json"},
			},
		},
		"storage": storage,
	}
	seedData, err := json.MarshalIndent(seed, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "init.json"), seedData, 0o644); err != nil {
		t.Fatal(err)
	}

	p := Preset{
		Name: "codex",
		Manifest: map[string]interface{}{
			"llm": map[string]interface{}{"provider": "codex", "model": "gpt-5.5"},
		},
		Source: SourceTemplate,
	}
	opts := DefaultAgentOpts()
	opts.PreserveActivePreset = true

	if err := GenerateInitJSONWithOpts(p, "alice", "alice", lingtaiDir, globalDir, opts); err != nil {
		t.Fatalf("GenerateInitJSONWithOpts: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(agentDir, "init.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("generated init.json is invalid json: %v\n%s", err, data)
	}
	gotStorage, ok := got["storage"].(map[string]interface{})
	if !ok {
		t.Fatalf("generated init.json dropped top-level storage block: %s", data)
	}
	if gotStorage["backend"] != "nokv" || gotStorage["enabled"] != true {
		t.Fatalf("storage block changed unexpectedly: %#v", gotStorage)
	}
	nokv, _ := gotStorage["nokv"].(map[string]interface{})
	if nokv["secret_access_key_env"] != "AWS_SECRET_ACCESS_KEY" {
		t.Fatalf("secret env var name was not preserved: %#v", nokv)
	}
	if raw := string(data); strings.Contains(raw, "super-secret-value") || strings.Contains(raw, "secret-bucket-value") {
		t.Fatalf("generated init.json expanded secret env values: %s", raw)
	}
}
