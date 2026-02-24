package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func connectMCP(ctx context.Context, workspace string) (*mcp.ClientSession, error) {
	client := mcp.NewClient(
		&mcp.Implementation{Name: "godoc-check", Version: "v1.0.0"},
		nil,
	)

	cmd := exec.Command("go-fred-mcp", "--workspace", workspace)
	cmd.Env = append(os.Environ(), "MCP_GOPLS_LOG_LEVEL=info")

	transport := &mcp.CommandTransport{Command: cmd}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to go-fred-mcp: %w", err)
	}

	return session, nil
}

type godocOpts struct {
	packagePattern string
	analyzers      string
	filters        string
	issuesOnly     bool
	customWords    string
	symbolFilter   string
}

func callGodoc(ctx context.Context, s *mcp.ClientSession, opts godocOpts) (map[string]json.RawMessage, error) {
	args := map[string]any{
		"package_pattern": opts.packagePattern,
		"analyzers":       opts.analyzers,
		"filters":         opts.filters,
		"issues_only":     opts.issuesOnly,
		"custom_words":    opts.customWords,
	}
	if opts.symbolFilter != "" {
		args["symbol_filter"] = opts.symbolFilter
	}

	result, err := s.CallTool(ctx, &mcp.CallToolParams{
		Name:      "godoc",
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("calling godoc tool: %w", err)
	}

	text := extractTextContent(result)
	if text == "" {
		return nil, nil
	}

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, fmt.Errorf("parsing godoc result: %w", err)
	}

	return parsed, nil
}

func callUpdateGodoc(ctx context.Context, s *mcp.ClientSession, updates map[string]any) (string, error) {
	updatesJSON, err := json.Marshal(updates)
	if err != nil {
		return "", fmt.Errorf("marshaling updates: %w", err)
	}

	result, err := s.CallTool(ctx, &mcp.CallToolParams{
		Name:      "update_godoc",
		Arguments: map[string]any{"updates": string(updatesJSON)},
	})
	if err != nil {
		return "", fmt.Errorf("calling update_godoc tool: %w", err)
	}

	return extractTextContent(result), nil
}

func extractTextContent(result *mcp.CallToolResult) string {
	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			return tc.Text
		}
	}
	return ""
}
