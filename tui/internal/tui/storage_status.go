package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/lingtai-tui/i18n"
)

const storageResolvedSchema = "lingtai.storage.resolved/v1"

type storageStatus struct {
	Schema  string         `json:"schema"`
	Enabled bool           `json:"enabled"`
	Backend string         `json:"backend"`
	Routes  []storageRoute `json:"routes"`
	Nokv    storageNokv    `json:"nokv"`
}

type storageRoute struct {
	Mount      string `json:"mount"`
	LocalRoot  string `json:"local_root"`
	Backend    string `json:"backend"`
	RemoteRoot string `json:"remote_root"`
}

type storageNokv struct {
	MetadataAddr string `json:"metadata_addr"`
	Bucket       string `json:"bucket"`
	Endpoint     string `json:"endpoint"`
}

func readStorageStatus(agentDir string) (storageStatus, error) {
	artifactPath := filepath.Join(agentDir, "system", "storage.resolved.json")
	if storageResolvedArtifactStale(filepath.Join(agentDir, "init.json"), artifactPath) {
		return storageStatus{}, fmt.Errorf("stale storage.resolved.json")
	}
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return storageStatus{}, err
	}
	var status storageStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return storageStatus{}, err
	}
	if status.Schema != storageResolvedSchema {
		return storageStatus{}, fmt.Errorf("unsupported storage.resolved.json schema %q", status.Schema)
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

func storageStatusDoctorLines(agentDir string) []doctorLine {
	status, err := readStorageStatus(agentDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []doctorLine{{Text: "✓ " + i18n.T("storage.local_no_artifact"), OK: true}}
		}
		return []doctorLine{{Text: "! " + i18n.TF("storage.unknown_reason", err), Warn: true}}
	}
	if !status.Enabled {
		return []doctorLine{{Text: "✓ " + i18n.T("storage.local"), OK: true}}
	}

	backend := strings.TrimSpace(status.Backend)
	if backend == "" {
		backend = "unknown"
	}
	lines := []doctorLine{{Text: "✓ " + i18n.TF("storage.backend", backend), OK: true}}
	for _, route := range status.Routes {
		mount := strings.TrimSpace(route.Mount)
		if mount == "" {
			continue
		}
		routeBackend := strings.TrimSpace(route.Backend)
		if routeBackend == "" {
			routeBackend = "unknown"
		}
		text := i18n.TF("storage.route", mount, routeBackend)
		if route.RemoteRoot != "" {
			text = i18n.TF("storage.route_remote", mount, routeBackend, route.RemoteRoot)
		}
		lines = append(lines, doctorLine{Text: "• " + text, Warn: routeBackend == "nokv", OK: routeBackend != "nokv"})
	}
	if status.Nokv.MetadataAddr != "" {
		lines = append(lines, doctorLine{Text: "• " + i18n.TF("storage.nokv_metadata", status.Nokv.MetadataAddr), Warn: true})
	}
	if status.Nokv.Bucket != "" {
		lines = append(lines, doctorLine{Text: "• " + i18n.TF("storage.nokv_bucket", status.Nokv.Bucket), Warn: true})
	}
	if status.Nokv.Endpoint != "" {
		lines = append(lines, doctorLine{Text: "• " + i18n.TF("storage.nokv_endpoint", status.Nokv.Endpoint), Warn: true})
	}
	return lines
}

func knowledgeMountBackedByNoKV(agentDir string) bool {
	status, err := readStorageStatus(agentDir)
	if err != nil || !status.Enabled {
		return false
	}
	for _, route := range status.Routes {
		if route.Mount == "knowledge" && route.Backend == "nokv" {
			return true
		}
	}
	return false
}

func nokvKnowledgeNoticeEntry() MarkdownEntry {
	return MarkdownEntry{
		Label:       i18n.T("knowledge.nokv_notice.label"),
		Description: i18n.T("knowledge.nokv_notice.description"),
		Content:     i18n.T("knowledge.nokv_notice.content"),
	}
}
