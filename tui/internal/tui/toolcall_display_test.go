package tui

import (
	"strings"
	"testing"
	"time"
)

func TestTruncateToolBody_NoLimitReturnsFull(t *testing.T) {
	body := strings.Repeat("x", 5000)
	got := truncateToolBody(body, 0)
	if got != body {
		t.Errorf("limit 0 must not truncate: got %d chars, want %d", len([]rune(got)), len([]rune(body)))
	}
}

func TestTruncateToolBody_NegativeLimitReturnsFull(t *testing.T) {
	body := strings.Repeat("y", 1000)
	got := truncateToolBody(body, -1)
	if got != body {
		t.Errorf("negative limit must not truncate: got %d chars, want %d", len([]rune(got)), len([]rune(body)))
	}
}

func TestTruncateToolBody_ShorterThanLimitUnchanged(t *testing.T) {
	body := "echo(hello)"
	got := truncateToolBody(body, 200)
	if got != body {
		t.Errorf("body shorter than limit must be unchanged: got %q, want %q", got, body)
	}
}

func TestTruncateToolBody_TruncatesAtLimitWithIndicator(t *testing.T) {
	body := strings.Repeat("z", 300)
	got := truncateToolBody(body, 200)
	// The retained content must be exactly the first 200 runes.
	if !strings.HasPrefix(got, strings.Repeat("z", 200)) {
		t.Errorf("truncated output must begin with first 200 chars of body")
	}
	// The full body must NOT survive intact.
	if got == body {
		t.Errorf("body longer than limit must be truncated")
	}
	// Truncation must be indicated to the user.
	if !strings.Contains(got, "…") && !strings.Contains(got, "...") {
		t.Errorf("truncation must be indicated with an ellipsis, got %q", got)
	}
	// The indicator should communicate how many characters were hidden (100).
	if !strings.Contains(got, "100") {
		t.Errorf("truncation indicator should report hidden char count (100), got %q", got)
	}
}

func TestFormatToolTimestamp_EmptyIsEmpty(t *testing.T) {
	if got := formatToolTimestamp(""); got != "" {
		t.Errorf("empty timestamp must format to empty string, got %q", got)
	}
}

func TestFormatToolTimestamp_RFC3339WholeSeconds(t *testing.T) {
	// events.jsonl uses whole-second RFC3339.
	ts := "2026-06-08T07:08:26Z"
	got := formatToolTimestamp(ts)
	want := time.Date(2026, 6, 8, 7, 8, 26, 0, time.UTC).Local().Format("15:04")
	if got != want {
		t.Errorf("formatToolTimestamp(%q) = %q, want %q", ts, got, want)
	}
}

func TestFormatToolTimestamp_RFC3339Nano(t *testing.T) {
	// mail-sourced entries can carry fractional seconds.
	ts := "2026-06-08T07:08:26.1279Z"
	got := formatToolTimestamp(ts)
	want := time.Date(2026, 6, 8, 7, 8, 26, 0, time.UTC).Local().Format("15:04")
	if got != want {
		t.Errorf("formatToolTimestamp(%q) = %q, want %q", ts, got, want)
	}
}

func TestFormatToolTimestamp_GarbageIsEmpty(t *testing.T) {
	if got := formatToolTimestamp("not-a-timestamp"); got != "" {
		t.Errorf("unparseable timestamp must format to empty string, got %q", got)
	}
}

func TestTruncateToolBody_DeterministicByRune(t *testing.T) {
	// Multi-byte runes must not be split mid-codepoint.
	body := strings.Repeat("世", 300) // each is 3 bytes, 1 rune
	got := truncateToolBody(body, 100)
	// Must be valid UTF-8 — no broken codepoints.
	for _, r := range got {
		if r == '�' {
			t.Fatalf("truncation split a multi-byte rune (got replacement char)")
		}
	}
	// First 100 runes retained.
	if !strings.HasPrefix(got, strings.Repeat("世", 100)) {
		t.Errorf("must retain first 100 runes intact")
	}
}
