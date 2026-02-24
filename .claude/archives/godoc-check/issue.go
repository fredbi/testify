package main

import (
	"encoding/json"
	"regexp"
)

// Issue represents a single spell-check finding for a symbol.
type Issue struct {
	Symbol     string
	Word       string
	Suggestion string
	RawMessage string

	Classification string // "REAL" or "FALSE_POSITIVE"
	Reason         string
}

var reQuotedWord = regexp.MustCompile(`"([^"]+)"`)

// extractIssues parses godoc triage output into Issue objects.
//
// The triage format (issues_only=true) is:
//
//	{symbol: {analyzer: [{message, suggestion, ...}, ...]}}
func extractIssues(triageResult map[string]json.RawMessage) []Issue {
	var issues []Issue

	for symbol, rawData := range triageResult {
		// Each symbol maps to {analyzer_name: [entries...]}
		var analyzerMap map[string]json.RawMessage
		if err := json.Unmarshal(rawData, &analyzerMap); err != nil {
			continue
		}

		for _, rawEntries := range analyzerMap {
			var entries []struct {
				Message    string `json:"message"`
				Suggestion string `json:"suggestion"`
			}
			if err := json.Unmarshal(rawEntries, &entries); err != nil {
				continue
			}

			for _, entry := range entries {
				word := ""
				if m := reQuotedWord.FindStringSubmatch(entry.Message); len(m) > 1 {
					word = m[1]
				}
				if word != "" {
					issues = append(issues, Issue{
						Symbol:     symbol,
						Word:       word,
						Suggestion: entry.Suggestion,
						RawMessage: entry.Message,
					})
				}
			}
		}
	}

	return issues
}
