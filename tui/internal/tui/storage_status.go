package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/lingtai-tui/i18n"
)

type storageResolvedStatus struct {
	Schema  string          `json:"schema"`
	Enabled bool            `json:"enabled"`
	Backend string          `json:"backend"`
	Routes  []storageRoute  `json:"routes"`
	Streams []storageStream `json:"streams"`
	Health  *storageHealth  `json:"health"`
	NoKV    struct {
		MetadataAddr string `json:"metadata_addr"`
		Bucket       string `json:"bucket"`
		Endpoint     string `json:"endpoint"`
	} `json:"nokv"`
}

type storageRoute struct {
	Mount      string `json:"mount"`
	LocalRoot  string `json:"local_root"`
	Backend    string `json:"backend"`
	RemoteRoot string `json:"remote_root"`
}

type storageStream struct {
	Stream     string `json:"stream"`
	LocalPath  string `json:"local_path"`
	Backend    string `json:"backend"`
	RemoteRoot string `json:"remote_root"`
	Mode       string `json:"mode"`
}

type storageHealth struct {
	Status          string   `json:"status"`
	Backend         string   `json:"backend"`
	Streams         []string `json:"streams"`
	LastError       string   `json:"last_error"`
	LastErrorStream string   `json:"last_error_stream"`
	UpdatedAt       string   `json:"updated_at"`
}

func readStorageResolvedStatus(agentDir string) (storageResolvedStatus, error) {
	var status storageResolvedStatus
	data, err := os.ReadFile(filepath.Join(agentDir, "system", "storage.resolved.json"))
	if err != nil {
		return status, err
	}
	if err := json.Unmarshal(data, &status); err != nil {
		return status, err
	}
	if status.Schema != "lingtai.storage.resolved/v1" {
		return status, fmt.Errorf("unsupported storage.resolved.json schema")
	}
	return status, nil
}

func storageStatusDoctorLines(agentDir string) []doctorLine {
	status, err := readStorageResolvedStatus(agentDir)
	if err != nil {
		return []doctorLine{{Text: i18n.T("doctor.storage_unknown"), Warn: true}}
	}
	if !status.Enabled {
		return []doctorLine{{Text: i18n.T("doctor.storage_local"), OK: true}}
	}
	var mounts []string
	for _, route := range status.Routes {
		if route.Backend == "nokv" && route.Mount != "" {
			mounts = append(mounts, route.Mount)
		}
	}
	for _, stream := range status.Streams {
		if stream.Backend != "nokv" || stream.Stream == "" {
			continue
		}
		label := "stream:" + stream.Stream
		if stream.Mode != "" {
			label = fmt.Sprintf("%s[%s]", label, stream.Mode)
		}
		mounts = append(mounts, label)
	}
	if len(mounts) == 0 {
		return []doctorLine{{Text: i18n.T("doctor.storage_enabled_no_routes"), Warn: true}}
	}
	lines := []doctorLine{{
		Text: i18n.TF("doctor.storage_enabled", strings.Join(mounts, ", ")),
		OK:   true,
	}}
	if status.Health != nil && strings.EqualFold(status.Health.Status, "degraded") {
		detail := status.Health.LastErrorStream
		if detail == "" {
			detail = strings.Join(status.Health.Streams, ", ")
		}
		if detail == "" {
			detail = status.Health.Backend
		}
		if detail == "" {
			detail = "mirror"
		}
		lines = append(lines, doctorLine{
			Text: i18n.TF("doctor.storage_degraded", detail),
			Warn: true,
		})
	}
	return lines
}

func knowledgeBackedByNoKV(agentDir string) bool {
	status, err := readStorageResolvedStatus(agentDir)
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

func knowledgeDirBackedByNoKV(knowledgeDir string) bool {
	agentDir := filepath.Dir(knowledgeDir)
	status, err := readStorageResolvedStatus(agentDir)
	if err != nil || !status.Enabled {
		return false
	}
	cleanKnowledge, _ := filepath.Abs(knowledgeDir)
	for _, route := range status.Routes {
		if route.Mount != "knowledge" || route.Backend != "nokv" || route.LocalRoot == "" {
			continue
		}
		cleanRoute, _ := filepath.Abs(route.LocalRoot)
		if cleanKnowledge == cleanRoute {
			return true
		}
	}
	return false
}

func nokvKnowledgeNotice() MarkdownEntry {
	return MarkdownEntry{
		Label:   i18n.T("knowledge.nokv_notice_label"),
		Group:   "Knowledge",
		Content: i18n.T("knowledge.nokv_notice_body"),
	}
}
