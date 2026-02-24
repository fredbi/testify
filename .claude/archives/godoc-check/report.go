package main

import (
	"fmt"
	"sort"
)

func printRawReport(pkg string, issues []Issue) int {
	fmt.Println()
	fmt.Println("=== Godoc Spell-Check Report (raw) ===")
	fmt.Printf("Package: %s\n", pkg)
	fmt.Printf("Issues found: %d\n", len(issues))

	// Group by symbol
	bySymbol := make(map[string][]Issue)
	for _, iss := range issues {
		bySymbol[iss.Symbol] = append(bySymbol[iss.Symbol], iss)
	}

	symbols := make([]string, 0, len(bySymbol))
	for s := range bySymbol {
		symbols = append(symbols, s)
	}
	sort.Strings(symbols)

	fmt.Println()
	for _, symbol := range symbols {
		fmt.Printf("  %s:\n", symbol)
		for _, iss := range bySymbol[symbol] {
			sug := ""
			if iss.Suggestion != "" {
				sug = fmt.Sprintf(" (suggestion: %s)", iss.Suggestion)
			}
			fmt.Printf("    \"%s\"%s\n", iss.Word, sug)
		}
	}

	fmt.Println()
	fmt.Println("Use without --report-only to classify with AI,")
	fmt.Println("or use --fix to apply corrections.")
	return 1
}

func printClassifiedReport(pkg string, all, real, fp []Issue, verbose bool) {
	fmt.Println()
	fmt.Println("=== Godoc Spell-Check Report ===")
	fmt.Printf("Package: %s\n", pkg)
	fmt.Printf("Issues found: %d\n", len(all))
	fmt.Printf("  Real errors: %d\n", len(real))
	fmt.Printf("  False positives: %d\n", len(fp))

	if len(real) > 0 {
		fmt.Println()
		fmt.Println("Real errors:")
		for _, iss := range real {
			sug := ""
			if iss.Suggestion != "" {
				sug = fmt.Sprintf(" -> suggest \"%s\"", iss.Suggestion)
			}
			fmt.Printf("  %s: \"%s\"%s\n", iss.Symbol, iss.Word, sug)
			if verbose {
				fmt.Printf("    reason: %s\n", iss.Reason)
			}
		}
	}

	if len(fp) > 0 {
		fpWords := uniqueSorted(fp)
		fmt.Println()
		fmt.Printf("False positives (%d unique words):\n", len(fpWords))
		for _, w := range fpWords {
			fmt.Printf("  %s\n", w)
		}
	}
}

func uniqueSorted(issues []Issue) []string {
	seen := make(map[string]struct{})
	for _, iss := range issues {
		seen[iss.Word] = struct{}{}
	}
	words := make([]string, 0, len(seen))
	for w := range seen {
		words = append(words, w)
	}
	sort.Strings(words)
	return words
}
