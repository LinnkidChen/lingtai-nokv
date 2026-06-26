package tui

import (
	"strings"
	"testing"
)

// The home telemetry row condenses current-session token usage and live context
// pressure into one muted line: "tok 18.4k / 128k  ctx 14%  ▓▓▓░░". It is
// scalar-only (never the _meta block hidden by PR #440) and hides gracefully
// when no data is available.
func TestFormatHomeTelemetry(t *testing.T) {
	tests := []struct {
		name    string
		tel     homeTelemetry
		width   int
		exact   string   // when set, output must equal this exactly ("" = hidden)
		want    []string // substrings that must appear
		notWant []string // substrings that must NOT appear
	}{
		{
			name:  "no data — row hidden",
			tel:   homeTelemetry{contextUsage: -1},
			width: 120,
			exact: "",
		},
		{
			name:  "tokens + limit + ctx + bar (wide)",
			tel:   homeTelemetry{sessionTokens: 18432, contextLimit: 128000, contextUsage: 0.14},
			width: 120,
			want:  []string{"tok", "18.4k", "/", "128.0k", "ctx", "14%", "▓", "░"},
		},
		{
			name:  "tokens without a known context limit — drop the / half",
			tel:   homeTelemetry{sessionTokens: 5000, contextLimit: 0, contextUsage: -1},
			width: 120,
			want:  []string{"tok", "5.0k"},
			// no limit, no ctx, no bar
			notWant: []string{"/", "ctx", "▓"},
		},
		{
			name:  "context only — no session tokens yet",
			tel:   homeTelemetry{sessionTokens: 0, contextLimit: 128000, contextUsage: 0.5},
			width: 120,
			want:  []string{"ctx", "50%", "▓"},
			// nothing to show for tokens, so no "tok"
			notWant: []string{"tok"},
		},
		{
			name:    "narrow terminal (<40) keeps numbers, hides bar",
			tel:     homeTelemetry{sessionTokens: 18432, contextLimit: 128000, contextUsage: 0.14},
			width:   30,
			want:    []string{"tok", "18.4k", "ctx", "14%"},
			notWant: []string{"▓", "░"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatHomeTelemetry(tt.tel, tt.width)
			if tt.exact != "" || !tt.tel.hasData() {
				if got != tt.exact {
					t.Fatalf("got %q, want exact %q", got, tt.exact)
				}
				return
			}
			for _, w := range tt.want {
				if !strings.Contains(got, w) {
					t.Errorf("telemetry %q missing %q", got, w)
				}
			}
			for _, nw := range tt.notWant {
				if strings.Contains(got, nw) {
					t.Errorf("telemetry %q should not contain %q", got, nw)
				}
			}
			// The label must resolve through i18n, not leak the raw key — guards
			// against the localized `mail.telemetry_*` keys going missing.
			if strings.Contains(got, "mail.telemetry") {
				t.Errorf("telemetry %q leaked a raw i18n key (missing translation)", got)
			}
		})
	}
}

func TestRenderContextBar(t *testing.T) {
	// 30% of 10 cells = 3 filled, 7 empty.
	bar := renderContextBar(30, 10)
	if got := strings.Count(bar, "▓"); got != 3 {
		t.Errorf("filled cells = %d, want 3 (bar=%q)", got, bar)
	}
	if got := strings.Count(bar, "░"); got != 7 {
		t.Errorf("empty cells = %d, want 7 (bar=%q)", got, bar)
	}

	// Clamping: over 100 and under 0 must not panic or overflow the width.
	if got := strings.Count(renderContextBar(250, 8), "▓"); got != 8 {
		t.Errorf("pct>100 should fill all 8 cells, got %d", got)
	}
	if got := strings.Count(renderContextBar(-5, 8), "▓"); got != 0 {
		t.Errorf("pct<0 should fill 0 cells, got %d", got)
	}
}

func TestHomeTelemetryBarWidth(t *testing.T) {
	if w := homeTelemetryBarWidth(30); w != 0 {
		t.Errorf("width 30 (<40) should hide bar (0), got %d", w)
	}
	if w := homeTelemetryBarWidth(40); w <= 0 {
		t.Errorf("width 40 should show a bar, got %d", w)
	}
	// Adaptive but capped: a very wide terminal must not produce a ruler.
	if w := homeTelemetryBarWidth(4000); w != homeTelemetryBarMax {
		t.Errorf("very wide terminal should cap bar at %d, got %d", homeTelemetryBarMax, w)
	}
}

func TestContextBarColorThresholds(t *testing.T) {
	// The three muted theme colors must be distinct across the thresholds so the
	// bar visibly escalates calm → caution → alarm. We can't assert exact RGB
	// without the theme initialized, but we can assert the mapping picks the
	// expected named color variables.
	if contextBarColor(50) != ColorActive {
		t.Error("50% should map to ColorActive (calm)")
	}
	if contextBarColor(75) != ColorAccent {
		t.Error("75% should map to ColorAccent (caution)")
	}
	if contextBarColor(95) != ColorSuspended {
		t.Error("95% should map to ColorSuspended (alarm)")
	}
}
