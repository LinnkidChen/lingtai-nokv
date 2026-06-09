package tui

import (
	"fmt"

	"github.com/anthropics/lingtai-tui/internal/fs"
)

// formatToolTimestamp renders a session timestamp as a short local "15:04"
// string for display beside a tool_call / tool_result line. It accepts both
// the whole-second RFC3339 form emitted into events.jsonl and the fractional
// RFC3339Nano form carried by mail-sourced entries (via fs.ParseSessionTs).
// An empty or unparseable timestamp yields an empty string so the caller can
// omit the stamp cleanly.
func formatToolTimestamp(ts string) string {
	if ts == "" {
		return ""
	}
	t := fs.ParseSessionTs(ts)
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("15:04")
}

// truncateToolBody trims a rendered tool_call / tool_result body to at most
// `limit` runes. A non-positive limit means "no truncation" — the body is
// returned verbatim (the default, so full tool call content is shown). When a
// finite limit is set and the body exceeds it, the body is cut deterministically
// at the rune boundary (never mid-codepoint) and a clear indicator reports how
// many characters were hidden.
func truncateToolBody(body string, limit int) string {
	if limit <= 0 {
		return body
	}
	runes := []rune(body)
	if len(runes) <= limit {
		return body
	}
	hidden := len(runes) - limit
	return string(runes[:limit]) + fmt.Sprintf("… (+%d chars)", hidden)
}
