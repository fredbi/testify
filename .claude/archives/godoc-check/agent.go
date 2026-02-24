package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

type config struct {
	pkg            string
	fix            bool
	updateWordlist bool
	reportOnly     bool
	model          string
	workspace      string
	wordlist       string
	verbose        bool
}

func run(ctx context.Context, cfg config) int {
	workspace, err := filepath.Abs(cfg.workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving workspace: %v\n", err)
		return 2
	}
	wordlistPath := filepath.Join(workspace, cfg.wordlist)

	log := func(msg string) {
		if cfg.verbose {
			fmt.Fprintf(os.Stderr, "  [verbose] %s\n", msg)
		}
	}

	// --- Setup AI client ---
	var ai *anthropic.Client
	if !cfg.reportOnly {
		ai, err = newAIClient(cfg.verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}
	}

	// --- Connect MCP ---
	fmt.Printf("Connecting to go-fred-mcp (workspace: %s)...\n", workspace)
	session, err := connectMCP(ctx, workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}
	defer session.Close()
	log("MCP session initialised")

	// --- Step 1: Read wordlist ---
	customWords, err := readWordlist(wordlistPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading wordlist: %v\n", err)
		return 2
	}
	wordCount := 0
	for _, line := range strings.Split(customWords, "\n") {
		if strings.TrimSpace(line) != "" {
			wordCount++
		}
	}
	log(fmt.Sprintf("Loaded %d custom words from %s", wordCount, wordlistPath))

	// --- Step 2: Triage ---
	fmt.Printf("Running spell-check triage on %s...\n", cfg.pkg)
	triage, err := callGodoc(ctx, session, godocOpts{
		packagePattern: cfg.pkg,
		analyzers:      "hunspell",
		filters:        "godoc-filter",
		issuesOnly:     true,
		customWords:    customWords,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 2
	}

	if len(triage) == 0 {
		fmt.Println()
		fmt.Println("=== Godoc Spell-Check Report ===")
		fmt.Printf("Package: %s\n", cfg.pkg)
		fmt.Println("No issues found. Documentation is clean!")
		return 0
	}

	issues := extractIssues(triage)
	if len(issues) == 0 {
		fmt.Println()
		fmt.Println("=== Godoc Spell-Check Report ===")
		fmt.Printf("Package: %s\n", cfg.pkg)
		fmt.Println("No issues found. Documentation is clean!")
		return 0
	}

	fmt.Printf("Found %d issue(s) across %d symbol(s)\n", len(issues), len(triage))

	// --- Report-only mode ---
	if cfg.reportOnly {
		return printRawReport(cfg.pkg, issues)
	}

	// --- Step 3: Classify each issue ---
	fmt.Println("Classifying issues...")
	for i := range issues {
		log(fmt.Sprintf("Classifying %d/%d: %s -> %q", i+1, len(issues), issues[i].Symbol, issues[i].Word))
		if err := classifyIssue(ctx, ai, &issues[i], cfg.model); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 2
		}
		log(fmt.Sprintf("  -> %s: %s", issues[i].Classification, issues[i].Reason))
	}

	var realIssues, falsePositives []Issue
	for _, iss := range issues {
		if iss.Classification == "REAL" {
			realIssues = append(realIssues, iss)
		} else {
			falsePositives = append(falsePositives, iss)
		}
	}

	// --- Step 4: Report ---
	printClassifiedReport(cfg.pkg, issues, realIssues, falsePositives, cfg.verbose)

	// --- Step 5: Update wordlist only (no code fixes) ---
	if cfg.updateWordlist && !cfg.fix {
		if len(falsePositives) > 0 {
			fpWords := uniqueSorted(falsePositives)
			fmt.Printf("\nAdding %d word(s) to wordlist: %s\n", len(fpWords), wordlistPath)
			if err := updateWordlist(wordlistPath, fpWords); err != nil {
				fmt.Fprintf(os.Stderr, "Error updating wordlist: %v\n", err)
				return 2
			}
		}
		if len(realIssues) > 0 {
			fmt.Println()
			fmt.Println("Run with --fix to also apply code corrections.")
		}
		if len(realIssues) > 0 {
			return 1
		}
		return 0
	}

	// --- Dry-run: report only ---
	if !cfg.fix {
		if len(realIssues) > 0 {
			fmt.Println()
			fmt.Println("Run with --fix to apply corrections.")
		}
		if len(realIssues) > 0 {
			return 1
		}
		return 0
	}

	// --- Apply fixes for real issues ---
	if len(realIssues) > 0 {
		fmt.Println()
		fmt.Println("Applying fixes...")

		// Group issues by symbol
		symbolsToFix := make(map[string][]Issue)
		for _, iss := range realIssues {
			symbolsToFix[iss.Symbol] = append(symbolsToFix[iss.Symbol], iss)
		}

		fixedCount := 0
		for symbol, symIssues := range symbolsToFix {
			log(fmt.Sprintf("Fetching full comment for %s", symbol))

			// Extract just the symbol name (after last dot) for the filter
			symName := symbol
			if idx := strings.LastIndex(symbol, "."); idx >= 0 {
				symName = symbol[idx+1:]
			}

			fullData, err := callGodoc(ctx, session, godocOpts{
				packagePattern: cfg.pkg,
				analyzers:      "hunspell",
				filters:        "godoc-filter",
				issuesOnly:     false,
				customWords:    customWords,
				symbolFilter:   "^" + regexp.QuoteMeta(symName) + "$",
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: error fetching comment for %s: %v\n", symbol, err)
				continue
			}

			rawSymbol, ok := fullData[symbol]
			if !ok {
				fmt.Printf("  Warning: could not fetch comment for %s, skipping\n", symbol)
				continue
			}

			var symbolData struct {
				Comment string `json:"comment"`
			}
			if err := json.Unmarshal(rawSymbol, &symbolData); err != nil {
				fmt.Printf("  Warning: could not parse comment for %s, skipping\n", symbol)
				continue
			}
			if symbolData.Comment == "" {
				fmt.Printf("  Warning: empty comment for %s, skipping\n", symbol)
				continue
			}

			// Apply fixes for each issue in this symbol
			currentComment := symbolData.Comment
			for _, iss := range symIssues {
				log(fmt.Sprintf("  Fixing %q in %s", iss.Word, symbol))
				currentComment, err = getFix(ctx, ai, iss, currentComment, cfg.model)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
				}
			}

			// Apply the fix via update_godoc
			result, err := callUpdateGodoc(ctx, session, map[string]any{
				symbol: map[string]any{"comment": currentComment},
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: error updating %s: %v\n", symbol, err)
				continue
			}
			log(fmt.Sprintf("  update_godoc result: %s", result))
			fixedCount++
			fmt.Printf("  Fixed: %s\n", symbol)
		}

		fmt.Printf("\n%d symbol(s) fixed.\n", fixedCount)
	}

	// Update wordlist with false positives
	if len(falsePositives) > 0 {
		fpWords := uniqueSorted(falsePositives)
		fmt.Printf("\nAdding %d word(s) to wordlist: %s\n", len(fpWords), wordlistPath)
		if err := updateWordlist(wordlistPath, fpWords); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating wordlist: %v\n", err)
			return 2
		}
		// Re-read for verification pass
		customWords, _ = readWordlist(wordlistPath)
	}

	// --- Step 6: Verification pass ---
	if len(realIssues) > 0 {
		fmt.Println("\nVerifying fixes...")
		verify, err := callGodoc(ctx, session, godocOpts{
			packagePattern: cfg.pkg,
			analyzers:      "hunspell",
			filters:        "godoc-filter",
			issuesOnly:     true,
			customWords:    customWords,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during verification: %v\n", err)
			return 2
		}

		var remaining []Issue
		if len(verify) > 0 {
			remaining = extractIssues(verify)
		}
		fmt.Printf("Before: %d issues\n", len(issues))
		fmt.Printf("After:  %d issues\n", len(remaining))
		if len(remaining) > 0 {
			fmt.Println("\nRemaining issues:")
			for _, iss := range remaining {
				fmt.Printf("  %s: \"%s\"\n", iss.Symbol, iss.Word)
			}
			return 1
		}
	}

	fmt.Println("\nAll clean!")
	return 0
}
