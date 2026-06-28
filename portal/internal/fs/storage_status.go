package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const storageResolvedSchema = "lingtai.storage.resolved/v1"

type StorageStatus struct {
	Schema  string         `json:"schema"`
	Enabled bool           `json:"enabled"`
	Backend string         `json:"backend"`
	Routes  []StorageRoute `json:"routes"`
	Nokv    StorageNokv    `json:"nokv"`
}

type StorageRoute struct {
	Mount      string `json:"mount"`
	LocalRoot  string `json:"local_root"`
	Backend    string `json:"backend"`
	RemoteRoot string `json:"remote_root"`
}

type StorageNokv struct {
	MetadataAddr string `json:"metadata_addr"`
	Bucket       string `json:"bucket"`
	Endpoint     string `json:"endpoint"`
}

func ReadStorageStatus(agentDir string) (StorageStatus, error) {
	artifactPath := filepath.Join(agentDir, "system", "storage.resolved.json")
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		if os.IsNotExist(err) {
			return StorageStatus{Backend: "local"}, nil
		}
		return StorageStatus{}, err
	}
	if storageResolvedArtifactStale(filepath.Join(agentDir, "init.json"), artifactPath) {
		return StorageStatus{}, fmt.Errorf("stale storage.resolved.json")
	}
	var status StorageStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return StorageStatus{}, err
	}
	if status.Schema != storageResolvedSchema {
		return StorageStatus{}, fmt.Errorf("unsupported storage.resolved.json schema %q", status.Schema)
	}
	if status.Backend == "" {
		status.Backend = "local"
	}
	return status, nil
}

func storageResolvedArtifactStale(initPath, artifactPath string) bool {
	initInfo, err := os.Stat(initPath)
	if err != nil {
		return false
	}
	artifactInfo, err := os.Stat(artifactPath)
	if err != nil {
		return false
	}
	return artifactInfo.ModTime().Before(initInfo.ModTime())
}

func (s StorageStatus) Summary() *StorageSummary {
	backend := strings.TrimSpace(s.Backend)
	if backend == "" {
		backend = "local"
	}
	summary := &StorageSummary{
		Enabled: s.Enabled,
		Backend: backend,
	}
	for _, route := range s.Routes {
		mount := strings.TrimSpace(route.Mount)
		if mount == "" {
			continue
		}
		routeBackend := strings.TrimSpace(route.Backend)
		if routeBackend == "" {
			routeBackend = "unknown"
		}
		summary.Routes = append(summary.Routes, StorageRouteSummary{
			Mount:      mount,
			Backend:    routeBackend,
			RemoteRoot: strings.TrimSpace(route.RemoteRoot),
		})
	}
	return summary
}

func (s StorageStatus) String() string {
	backend := strings.TrimSpace(s.Backend)
	if backend == "" {
		backend = "local"
	}
	var parts []string
	parts = append(parts, "storage: "+backend)
	for _, route := range s.Routes {
		if strings.TrimSpace(route.Mount) == "" {
			continue
		}
		routeBackend := strings.TrimSpace(route.Backend)
		if routeBackend == "" {
			routeBackend = "unknown"
		}
		text := fmt.Sprintf("%s -> %s", route.Mount, routeBackend)
		if route.RemoteRoot != "" {
			text += " (" + route.RemoteRoot + ")"
		}
		parts = append(parts, text)
	}
	if s.Nokv.MetadataAddr != "" {
		parts = append(parts, "NoKV metadata: "+s.Nokv.MetadataAddr)
	}
	if s.Nokv.Bucket != "" {
		parts = append(parts, "NoKV bucket: "+s.Nokv.Bucket)
	}
	if s.Nokv.Endpoint != "" {
		parts = append(parts, "NoKV endpoint: "+s.Nokv.Endpoint)
	}
	return strings.Join(parts, "\n")
}
