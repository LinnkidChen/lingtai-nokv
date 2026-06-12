# Text output spacing implementation summary

## Problem

In the TUI ctrl+o transcript, consecutive `tool_call` / `tool_result` groups already receive a blank separator when their `api_call_id` changes, so different LLM API responses read as separate groups. Consecutive `text_output` entries did not use the same grouping, so assistant text output from distinct API responses could run together visually.

## Root cause

`MailModel.buildMessages` derived a current API-call id from hidden `llm_response` markers, but only copied that id onto `tool_call` and `tool_result` entries. `text_output` entries therefore reached `renderMessages` without API grouping metadata, and the render loop only tracked visible tool groups.

## Fix

- `ChatMessage.ApiCallID` now documents that it applies to `text_output` as well as tool events.
- `MailModel.buildMessages` assigns derived/explicit `api_call_id` to `text_output` entries.
- `renderMessages` tracks consecutive visible `text_output` entries and inserts a blank line when the API-call group changes.
- `textOutputGroupSeparatorBefore` mirrors the tool separator rule but stays conservative for legacy entries with no grouping metadata.
- `tui/internal/tui/ANATOMY.md` documents the behavior.

## Validation

- `cd tui && go test ./internal/tui -run 'TestBuildMessagesAssignsApiCallIDToTextOutput|TestTextOutputGroupSeparatorBefore|TestRenderMessages_.*TextOutput|TestRenderMessages_.*ApiCallGroup|TestToolGroupSeparatorBefore' -count=1`
- `cd tui && go test ./internal/tui -count=1`
- `cd tui && go test ./... -count=1`
- `git diff --check`
