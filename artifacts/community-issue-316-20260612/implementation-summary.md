# Issue #316 Implementation Summary

**Date:** 2026-06-12
**Branch:** fix/setup-preset-policy-316-20260612

## Problem

`/setup` with "Keep current preset" (cursor == -1) synthesises a placeholder
`Preset{Name: "keep_current"}` so downstream wizard steps can read the agent's
existing LLM/capability config. On save, `GenerateInitJSONWithOpts` passed this
sentinel through `RefFor()`, which computed a non-existent path
`~/.lingtai-tui/presets/saved/keep_current.json` and wrote it into
`manifest.preset.default` and `manifest.preset.allowed`.

A second path: `performSetupSaveOnly` would also propagate that synthetic ref to
all peer agents via `propagatePresetPolicyToNetwork`.

## Fix (3 parts)

### 1. `tui/internal/preset/preset.go` — core writer fix

Added `isSyntheticPreset(p Preset) bool` for the exact keep-current sentinel
(`SourceUnknown` + `Name == "keep_current"`). The predicate is intentionally
narrow so hand-built/custom presets that rely on `RefFor`'s saved-preset fallback
are not misclassified.

In `GenerateInitJSONWithOpts`, inside the `PreserveActivePreset` block that already
reads `active` and `allowed` from the existing init.json, also read
`manifest.preset.default`. When `isSyntheticPreset(p)` is true, override `presetRef`
with the existing default string; if an older init lacks `default`, fall back to the
real existing `active` ref. The synthetic name never reaches disk.

### 2. `tui/internal/tui/firstrun.go` — propagation fix

In `performSetupSaveOnly()`, the propagation call used `presetCanonicalRef(p)` which
returns the synthetic path. Now reads the real existing default from
`m.setupKeepInitJSON["manifest"]["preset"]["default"]` when `m.cursor == -1`.

### 3. `tui/internal/tui/firstrun.go` — UX: auto-check just-created preset

In `enterAgentPresets()`, after setup-mode hydration from the existing init.json,
auto-check the row for `m.cursor`'s preset if it is not already in `presetAllowed`.
This ensures a preset created in the editor on the same `/setup` run defaults to
authorized, without turning all saved presets on.

## Files changed

| File | Change |
|------|--------|
| `tui/internal/preset/preset.go` | Added `isSyntheticPreset`; override `presetRef` with existing default when synthetic |
| `tui/internal/tui/firstrun.go` | Propagation uses real existing default for keep-current; auto-check just-created preset |
| `tui/internal/preset/preset_keep_current_test.go` | New regression tests (2 cases) |
| `reports/issue316-setup-preset-policy-20260612.html` | HTML explainer |
| `artifacts/community-issue-316-20260612/implementation-summary.md` | This file |

## Tests

```
go test ./internal/preset/ -run TestSetupModeKeepCurrent -v
# → PASS (2 tests)
go test ./... -timeout 120s
# → all packages pass
```

## Invariants preserved

- `manifest.preset.allowed` remains a static authorization snapshot
- Kernel/TUI allowed-gate unchanged
- No migrations needed (bug only affected newly-written files)
