package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ReadStorageStatus(agentDir string) (StorageStatus, error) {
	var status StorageStatus
	data, err := os.ReadFile(filepath.Join(agentDir, "system", "storage.resolved.json"))
	if err != nil {
		return status, fmt.Errorf("read storage.resolved.json: %w", err)
	}
	if err := json.Unmarshal(data, &status); err != nil {
		return status, fmt.Errorf("parse storage.resolved.json: %w", err)
	}
	if status.Schema != "lingtai.storage.resolved/v1" {
		return status, fmt.Errorf("unsupported storage.resolved.json schema")
	}
	return status, nil
}
