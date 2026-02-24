# SDK Agents

This project uses lightweight SDK agents for documentation quality checks.
Both agents share the same architecture and authentication model.

## Prerequisites

- `go-fred-mcp` MCP server installed (e.g. `~/bin/go-fred-mcp`)
- Agent binaries installed (e.g. `~/bin/godoc-check`, `~/bin/markdown-check`)

## How they work

Each agent connects to `go-fred-mcp` via MCP stdio and runs hunspell spellcheck
against the project wordlist (`.github/wordlist.txt`).

- **godoc-check** scans Go doc comments. Spelling issues are enriched with
  suggestions regarding godoc conventions (e.g. use links for known identifiers).
- **markdown-check** scans markdown files. It also performs markdown linting
  (broken links, formatting issues) and suppresses Go exported symbols from
  being flagged as misspellings.

Both agents then call Claude haiku for each finding to classify it as a real
error or a false positive. Only these small judgment calls use the AI API —
the rest is deterministic Go orchestration.

## Authentication

The agents automatically use the Claude Code OAuth token from
`~/.claude/.credentials.json`. No separate API key needed.

Falls back to `ANTHROPIC_API_KEY` or `ANTHROPIC_AUTH_TOKEN` if set.

## Cost model

Unlike a prompt-based approach that uses the full model for the entire workflow
(reading wordlists, parsing results, formatting reports), these SDK agents only
call haiku for classification decisions per finding. Everything else is plain Go.
