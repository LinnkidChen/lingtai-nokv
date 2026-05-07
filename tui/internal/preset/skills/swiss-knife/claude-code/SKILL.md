---
name: claude-code
description: >
  Delegate code implementation, patch writing, documentation, and refactoring to
  Claude Code CLI (Anthropic's coding agent). Runs non-interactively from bash,
  uses the human's Claude Max subscription (no additional API costs), and supports
  quality/effort/budget controls. Use this when you need to write code, generate
  patches, refactor files, create documentation, or do any multi-file code work
  that would be faster delegated than done manually.
version: 1.0.0
tags: [cli, code, delegation, claude, implementation]
---

# Claude Code CLI — Code Delegation

Delegate code work to [Claude Code](https://docs.anthropic.com/en/docs/claude-code) — Anthropic's coding agent — running non-interactively from bash.

## Prerequisites

- Claude Code installed: `which claude` → `/Users/huangzesen/.local/bin/claude`
- Uses the human's **Claude Max subscription** — no additional API costs
- Rate limit tier: `default_claude_max_20x` (effectively unlimited for typical use)

## Quick Usage

```bash
claude -p "your prompt here" --dangerously-skip-permissions
```

This runs Claude Code in non-interactive mode (`-p` = print and exit), skipping permission checks for automation.

## Key Flags

| Flag | Purpose |
|------|---------|
| `-p` / `--print` | Non-interactive mode — run, print result, exit |
| `--dangerously-skip-permissions` | Skip permission prompts (required for automation) |
| `--effort max` | Maximum reasoning effort for complex tasks |
| `--model opus` | Use Opus model for highest quality |
| `--model sonnet` | Use Sonnet model for speed (default) |
| `--max-budget-usd N` | Spending limit per call |
| `--allowedTools "Bash Edit Read Write"` | Restrict which tools Claude can use |
| `--system-prompt "..."` | Custom system prompt |
| `--add-dir /path/to/dir` | Grant access to additional directories |
| `-d /path/to/repo` | Set working directory |

## Recommended Patterns

### Simple task (default quality)
```bash
claude -p "fix the typo in README.md" --dangerously-skip-permissions
```

### Complex implementation (max quality)
```bash
claude -p "implement the caching layer as described in DESIGN.md" \
  --dangerously-skip-permissions \
  --effort max \
  --model opus
```

### With budget control
```bash
claude -p "refactor the auth module" \
  --dangerously-skip-permissions \
  --effort max \
  --model opus \
  --max-budget-usd 5.0
```

### Working in a specific repo
```bash
claude -p "add unit tests for the parser module" \
  --dangerously-skip-permissions \
  -d /path/to/repo
```

### Restricted tools (safer)
```bash
claude -p "generate a patch for issue #42" \
  --dangerously-skip-permissions \
  --allowedTools "Bash Edit Read Write"
```

## Best Practices

1. **Increase bash timeout**: Set `timeout=900` (15 minutes) on the bash tool call for complex tasks. Claude Code has no built-in timeout — the bash tool's timeout controls it.

2. **Use `--effort max` for complex work**: This tells Claude to think harder. Worth it for architecture, refactoring, and multi-file changes.

3. **Use `--model opus` for quality**: Opus produces better code for complex logic. Use Sonnet (default) for simple tasks.

4. **Split large tasks**: Multiple smaller `claude -p` calls chained together beat one monolithic prompt. Each call has its own context window.

5. **Write clear prompts**: Claude Code reads the repo context itself. Give it the goal, constraints, and acceptance criteria — don't dump the entire codebase into the prompt.

6. **Set budget for unknown tasks**: Use `--max-budget-usd` to prevent runaway spending on ambiguous tasks.

## Workflow for Patch/PR Creation

1. **Design**: Write a clear spec (what to change, why, constraints)
2. **Delegate**: Run `claude -p "implement: <spec>" --dangerously-skip-permissions --effort max`
3. **Review**: Check the output, run tests
4. **Push**: Create branch, commit, push as PR

## What to Delegate

- **Code implementation**: New features, bug fixes, refactoring
- **Patch generation**: Multi-file changes, API migrations
- **Documentation**: READMEs, docstrings, API docs
- **Test writing**: Unit tests, integration tests
- **Code review**: Ask Claude to review a PR or diff

## What NOT to Delegate

- **Simple one-line edits**: Use the `edit` tool directly
- **File reading/searching**: Use `read`/`grep`/`glob` directly
- **Shell commands**: Use `bash` directly for non-code tasks
- **Tasks requiring your full context**: Claude Code doesn't share your conversation history

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Timeout after 30s | Increase `timeout=900` on bash call |
| Claude Code not found | Check `which claude` → `/Users/huangzesen/.local/bin/claude` |
| Permission errors | Always include `--dangerously-skip-permissions` |
| Output truncated | Check if Claude hit the budget limit |
| Rate limited | Wait and retry; Max tier has generous limits |
