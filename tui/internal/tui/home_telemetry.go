package tui

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/anthropics/lingtai-tui/i18n"
	"github.com/anthropics/lingtai-tui/internal/fs"
)

// Home telemetry row — a single muted line shown between the input box and the
// bottom path/shortcut footer. It condenses the current session's token usage
// and the live context-window pressure into one high-density line:
//
//	tok 18.4k / 128k  ctx 14%  ▓▓▓░░░░░░░░░░░
//
// meaning: session tokens so far / model context limit, context-usage percent,
// and a small adaptive bar. It is scalar-only — never the noisy `_meta` block
// hidden by PR #440. Data sources (all already read elsewhere in the TUI):
//   - sessionTokens: fs.SumTokenLedger(...).Input+Output+Thinking (props.go)
//   - contextLimit:  manifest.llm.context_limit (fs.ReadInitManifest)
//   - contextUsage:  latest notification Meta.Context.Usage (0..1, -1 = none)
//
// When no data is available the row is omitted entirely (graceful hide), matching
// the TUI's "show nothing rather than a placeholder zero" footer style.

// homeTelemetry holds the already-resolved scalars for the home row. Keeping the
// data plain (no rendering) makes formatHomeTelemetry trivially testable.
type homeTelemetry struct {
	sessionTokens int64   // total session tokens (input+output+thinking); 0 = unknown
	contextLimit  int64   // model context window; 0 = unknown
	contextUsage  float64 // latest context-usage fraction 0..1; <0 = unknown
}

// gatherHomeTelemetry resolves the three telemetry scalars for the orchestrator
// agent from data the TUI already reads elsewhere:
//   - sessionTokens from logs/token_ledger.jsonl (fs.SumTokenLedger, cached)
//   - contextLimit from manifest.llm.context_limit (fs.ReadInitManifest)
//   - contextUsage from the freshest notification Meta.Context.Usage in the
//     current message list (the same value the notification footer renders)
//
// Every source degrades to its "unknown" sentinel independently, so a missing
// ledger / manifest / notification just drops that fragment rather than the row.
func (m MailModel) gatherHomeTelemetry() homeTelemetry {
	t := homeTelemetry{contextUsage: -1}
	if m.orchestrator != "" {
		ledger := fs.SumTokenLedger(filepath.Join(m.orchestrator, "logs", "token_ledger.jsonl"))
		t.sessionTokens = ledger.Input + ledger.Output + ledger.Thinking
		if manifest, err := fs.ReadInitManifest(m.orchestrator); err == nil {
			if llm, ok := manifest["llm"].(map[string]interface{}); ok {
				if cl, ok := llm["context_limit"].(float64); ok && cl > 0 {
					t.contextLimit = int64(cl)
				}
			}
		}
	}
	// Latest context-usage fraction: scan the built messages backward for the
	// most recent notification that carried a context block.
	for i := len(m.messages) - 1; i >= 0; i-- {
		msg := m.messages[i]
		if msg.Type == "notification" && msg.Meta != nil && msg.Meta.Context != nil && msg.Meta.Context.Usage >= 0 {
			t.contextUsage = msg.Meta.Context.Usage
			break
		}
	}
	return t
}

// hasData reports whether any fragment is renderable. With nothing to show the
// caller omits the whole row.
func (t homeTelemetry) hasData() bool {
	return t.sessionTokens > 0 || t.contextUsage >= 0
}

// formatHomeTelemetry renders the telemetry row for the given terminal width, or
// "" when there is nothing to show. The returned string is already styled and
// left-padded to align with the status-bar path label ("  " indent). width is
// the full terminal width; the bar adapts to it and is hidden entirely below
// homeTelemetryBarMinWidth so narrow terminals keep the numbers.
func formatHomeTelemetry(t homeTelemetry, width int) string {
	if !t.hasData() {
		return ""
	}

	var segs []string

	// tok 18.4k / 128k  (the "/ limit" half is dropped when the limit is unknown)
	if t.sessionTokens > 0 {
		tok := i18n.T("mail.telemetry_tok") + " " + humanizeTokenCount(t.sessionTokens)
		if t.contextLimit > 0 {
			tok += " / " + humanizeTokenCount(t.contextLimit)
		}
		segs = append(segs, tok)
	}

	// ctx 14%  ▓▓▓░░  (the bar is dropped on narrow terminals)
	if t.contextUsage >= 0 {
		pct := t.contextUsage * 100
		ctx := fmt.Sprintf("%s %.0f%%", i18n.T("mail.telemetry_ctx"), pct)
		if barW := homeTelemetryBarWidth(width); barW > 0 {
			ctx += "  " + renderContextBar(pct, barW)
		}
		segs = append(segs, ctx)
	}

	if len(segs) == 0 {
		return ""
	}
	// Two spaces between segments for a calm, low-density-feeling separation; the
	// label words themselves are muted by the caller's style.
	return "  " + StyleFaint.Render(strings.Join(segs, "  "))
}

const (
	// homeTelemetryBarMinWidth is the narrowest terminal that still shows the
	// context bar; below it we keep "tok …" and "ctx N%" but drop the bar so the
	// row never wraps. Jason asked for width<40 → numbers only.
	homeTelemetryBarMinWidth = 40
	// homeTelemetryBarMax caps the bar so it stays a compact gauge, not a ruler.
	homeTelemetryBarMax = 14
)

// homeTelemetryBarWidth picks an adaptive bar cell count for the terminal width,
// or 0 to hide the bar on narrow terminals.
func homeTelemetryBarWidth(width int) int {
	if width < homeTelemetryBarMinWidth {
		return 0
	}
	// Scale gently with width; clamp to a compact range so it reads as a gauge.
	w := width / 8
	if w < 6 {
		w = 6
	}
	if w > homeTelemetryBarMax {
		w = homeTelemetryBarMax
	}
	return w
}

// renderContextBar returns a small filled/empty bar proportional to pct (0..100)
// with width cells, colored by pressure: <70% muted green/teal, 70–89% amber,
// >=90% muted red. Empty cells stay dim gray. Uses ▓ (filled) and ░ (empty) to
// match Jason's mock; both are box-drawing glyphs the TUI already relies on.
func renderContextBar(pct float64, width int) string {
	if width < 1 {
		width = 1
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int((pct / 100.0) * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	full := lipgloss.NewStyle().Foreground(contextBarColor(pct)).Render(strings.Repeat("▓", filled))
	empty := lipgloss.NewStyle().Foreground(ColorTextFaint).Render(strings.Repeat("░", width-filled))
	return full + empty
}

// contextBarColor maps context pressure to a muted theme color. Thresholds match
// Jason's spec: calm below 70%, caution to 89%, alarm at 90%+. All three are the
// theme's existing muted state colors — no bright red, consistent with the
// beige/dim footer palette.
func contextBarColor(pct float64) color.Color {
	switch {
	case pct >= 90:
		return ColorSuspended // 朱砂 — muted red
	case pct >= 70:
		return ColorAccent // 琥珀 — amber
	default:
		return ColorActive // 竹青 — muted green/teal
	}
}
