package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const classifyPrompt = `You are reviewing a hunspell spell-check finding in a Go doc comment.

Symbol: %s
Flagged word: "%s"
Hunspell suggestion: "%s"
Full message: %s

Is this a real spelling error, or a false positive (e.g., a Go identifier,
parameter name, technical term, or valid but uncommon English word)?

Reply with exactly one line:
REAL: <brief reason>
or
FALSE_POSITIVE: <brief reason>`

const fixPrompt = `Fix the spelling error in this Go doc comment. Change "%s" appropriately.
The hunspell suggestion is "%s".
Keep the fix minimal -- only change what's needed. Preserve godoc formatting.
The first line must remain "// SymbolName verb..." per godoc conventions.

Current comment:
%s

Return ONLY the corrected comment text, nothing else.`

func newAIClient(verbose bool) (*anthropic.Client, error) {
	// 1. Standard API key
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		if verbose {
			fmt.Fprintln(os.Stderr, "  [verbose] Using ANTHROPIC_API_KEY")
		}
		c := anthropic.NewClient(option.WithAPIKey(key))
		return &c, nil
	}

	oauthHeader := option.WithHeader("anthropic-beta", "oauth-2025-04-20")

	// 2. Explicit auth token env var
	if token := os.Getenv("ANTHROPIC_AUTH_TOKEN"); token != "" {
		if verbose {
			fmt.Fprintln(os.Stderr, "  [verbose] Using ANTHROPIC_AUTH_TOKEN")
		}
		c := anthropic.NewClient(option.WithAuthToken(token), oauthHeader)
		return &c, nil
	}

	// 3. Claude Code local credentials
	if token := readClaudeCodeToken(); token != "" {
		if verbose {
			fmt.Fprintln(os.Stderr, "  [verbose] Using Claude Code OAuth token from ~/.claude/.credentials.json")
		}
		c := anthropic.NewClient(option.WithAuthToken(token), oauthHeader)
		return &c, nil
	}

	return nil, fmt.Errorf(
		"no Anthropic authentication found\n" +
			"Options:\n" +
			"  - Set ANTHROPIC_API_KEY env var (API key from console.anthropic.com)\n" +
			"  - Set ANTHROPIC_AUTH_TOKEN env var (OAuth bearer token)\n" +
			"  - Log in to Claude Code (claude login) to use its OAuth token\n" +
			"  - Use --report-only to skip AI classification",
	)
}

func readClaudeCodeToken() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(home, ".claude", ".credentials.json"))
	if err != nil {
		return ""
	}

	var creds struct {
		ClaudeAIOAuth struct {
			AccessToken string `json:"accessToken"`
		} `json:"claudeAiOauth"`
	}
	if err := json.Unmarshal(data, &creds); err != nil {
		return ""
	}

	return creds.ClaudeAIOAuth.AccessToken
}

func classifyIssue(ctx context.Context, c *anthropic.Client, issue *Issue, model string) error {
	prompt := fmt.Sprintf(classifyPrompt, issue.Symbol, issue.Word, issue.Suggestion, issue.RawMessage)

	resp, err := c.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 100,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return fmt.Errorf("classifying %q in %s: %w", issue.Word, issue.Symbol, err)
	}

	text := ""
	for _, block := range resp.Content {
		if block.Type == "text" {
			text = block.Text
			break
		}
	}

	issue.Classification, issue.Reason = parseClassification(text)
	return nil
}

func parseClassification(text string) (string, string) {
	text = strings.TrimSpace(text)
	upper := strings.ToUpper(text)

	if strings.HasPrefix(upper, "REAL") {
		reason := ""
		if idx := strings.Index(text, ":"); idx >= 0 {
			reason = strings.TrimSpace(text[idx+1:])
		}
		return "REAL", reason
	}

	if strings.HasPrefix(upper, "FALSE_POSITIVE") {
		reason := ""
		if idx := strings.Index(text, ":"); idx >= 0 {
			reason = strings.TrimSpace(text[idx+1:])
		}
		return "FALSE_POSITIVE", reason
	}

	// Fallback: if the response contains "false positive" anywhere
	if strings.Contains(strings.ToLower(text), "false positive") {
		return "FALSE_POSITIVE", text
	}

	return "REAL", text
}

func getFix(ctx context.Context, c *anthropic.Client, issue Issue, comment, model string) (string, error) {
	prompt := fmt.Sprintf(fixPrompt, issue.Word, issue.Suggestion, comment)

	resp, err := c.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 2000,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return comment, fmt.Errorf("getting fix for %q: %w", issue.Word, err)
	}

	for _, block := range resp.Content {
		if block.Type == "text" {
			return strings.TrimSpace(block.Text), nil
		}
	}

	return comment, nil
}
