package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchitectureEntryLinks(t *testing.T) {
	root := filepath.Clean("..")

	for _, pair := range []struct {
		path  string
		links []string
	}{
		{"ANATOMY.md", []string{"CONTRACT.md", "dev-guide-skill/SKILL.md"}},
		{"CONTRACT.md", []string{"ANATOMY.md", "dev-guide-skill/SKILL.md"}},
	} {
		text := readArchitectureFile(t, root, pair.path)
		for _, target := range pair.links {
			if !strings.Contains(text, "\n  - "+target+"\n") {
				t.Errorf("%s related_files must include %s", pair.path, target)
			}
		}
	}

	for _, path := range []string{"README.md", "README.zh.md", "README.wen.md", "CLAUDE.md"} {
		text := readArchitectureFile(t, root, path)
		for _, target := range []string{"ANATOMY.md", "CONTRACT.md", "dev-guide-skill/SKILL.md"} {
			if !hasMarkdownLink(text, target) {
				t.Errorf("%s must link to %s", path, target)
			}
		}
	}
}

func readArchitectureFile(t *testing.T, root, path string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func hasMarkdownLink(text, target string) bool {
	return strings.Contains(text, "]("+target+")") ||
		strings.Contains(text, "](./"+target+")")
}
