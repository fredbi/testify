// SPDX-FileCopyrightText: Copyright 2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package difflib

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aymanbagabas/go-udiff"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/go-openapi/testify/v2/internal/difflib"
	"github.com/go-openapi/testify/v2/internal/spew"
)

// Test data structures for comparison

type Person struct {
	Name    string
	Age     int
	Email   string
	Address Address
	Tags    []string
}

type Address struct {
	Street  string
	City    string
	Country string
	Zip     string
}

// diffCase represents a test case for diff comparison.
type diffCase struct {
	name     string
	original any
	modified any
}

func TestDiffLibComparison(t *testing.T) {
	cases := []diffCase{
		structDiffCase(),
		mapDiffCase(),
		sliceDiffCase(),
		nestedStructDiffCase(),
		mixedChangesDiffCase(),
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Serialize both values using spew
			originalStr := spewDump(tc.original)
			modifiedStr := spewDump(tc.modified)

			t.Logf("\n%s", strings.Repeat("=", 80))
			t.Logf("TEST CASE: %s", tc.name)
			t.Logf("%s\n", strings.Repeat("=", 80))

			t.Logf("\n--- ORIGINAL ---\n%s", originalStr)
			t.Logf("\n--- MODIFIED ---\n%s", modifiedStr)

			// Compare outputs from all three libraries
			t.Logf("\n%s", strings.Repeat("-", 40))
			t.Logf("1. OUR DIFFLIB (yardstick)")
			t.Logf("%s\n", strings.Repeat("-", 40))
			ourDiff := ourDifflib(originalStr, modifiedStr)
			t.Logf("%s", ourDiff)

			t.Logf("\n%s", strings.Repeat("-", 40))
			t.Logf("2. GOTEXTDIFF (hexops)")
			t.Logf("%s\n", strings.Repeat("-", 40))
			gotextDiff := gotextdiffLib(originalStr, modifiedStr)
			t.Logf("%s", gotextDiff)

			t.Logf("\n%s", strings.Repeat("-", 40))
			t.Logf("3. GO-UDIFF (aymanbagabas)")
			t.Logf("%s\n", strings.Repeat("-", 40))
			goudiffResult := goudiffLib(originalStr, modifiedStr)
			t.Logf("%s", goudiffResult)

			t.Logf("\n%s", strings.Repeat("-", 40))
			t.Logf("4. GO-DIFF (sergi/diffmatchpatch)")
			t.Logf("%s\n", strings.Repeat("-", 40))
			sergidiff := sergiDiffLib(originalStr, modifiedStr)
			t.Logf("%s", sergidiff)
		})
	}
}

// Test cases

func structDiffCase() diffCase {
	return diffCase{
		name: "simple_struct",
		original: Person{
			Name:  "Alice",
			Age:   30,
			Email: "alice@example.com",
		},
		modified: Person{
			Name:  "Alice",
			Age:   31,                        // changed
			Email: "alice.smith@example.com", // changed
		},
	}
}

func mapDiffCase() diffCase {
	return diffCase{
		name: "simple_map",
		original: map[string]int{
			"apple":  1,
			"banana": 2,
			"cherry": 3,
		},
		modified: map[string]int{
			"apple":  1,
			"banana": 5, // changed
			"date":   4, // added (cherry removed)
		},
	}
}

func sliceDiffCase() diffCase {
	return diffCase{
		name: "simple_slice",
		original: []string{
			"first",
			"second",
			"third",
			"fourth",
		},
		modified: []string{
			"first",
			"SECOND", // changed
			"third",
			"fourth",
			"fifth", // added
		},
	}
}

func nestedStructDiffCase() diffCase {
	return diffCase{
		name: "nested_struct",
		original: Person{
			Name:  "Bob",
			Age:   25,
			Email: "bob@example.com",
			Address: Address{
				Street:  "123 Main St",
				City:    "Boston",
				Country: "USA",
				Zip:     "02101",
			},
			Tags: []string{"developer", "golang"},
		},
		modified: Person{
			Name:  "Bob",
			Age:   26, // changed
			Email: "bob@example.com",
			Address: Address{
				Street:  "456 Oak Ave", // changed
				City:    "Boston",
				Country: "USA",
				Zip:     "02102", // changed
			},
			Tags: []string{"developer", "golang", "senior"}, // added
		},
	}
}

func mixedChangesDiffCase() diffCase {
	return diffCase{
		name: "mixed_changes",
		original: map[string]any{
			"name":   "Test",
			"count":  42,
			"active": true,
			"items":  []string{"a", "b", "c"},
			"meta": map[string]string{
				"created": "2025-01-01",
				"version": "1.0",
			},
		},
		modified: map[string]any{
			"name":   "Test Updated", // changed
			"count":  42,
			"active": false,                   // changed
			"items":  []string{"a", "b", "d"}, // changed
			"meta": map[string]string{
				"created": "2025-01-01",
				"version": "2.0", // changed
				"author":  "Bob", // added
			},
		},
	}
}

// Library wrappers

// spewDump serializes a value using our internalized spew.
func spewDump(v any) string {
	return spew.Sdump(v)
}

// ourDifflib produces a unified diff using our internalized difflib.
func ourDifflib(a, b string) string {
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(a),
		B:        difflib.SplitLines(b),
		FromFile: "original",
		ToFile:   "modified",
		Context:  3,
	}

	result, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	if result == "" {
		return "(no differences)"
	}

	return result
}

// gotextdiffLib produces a unified diff using hexops/gotextdiff.
func gotextdiffLib(a, b string) string {
	edits := myers.ComputeEdits(span.URIFromPath("original"), a, b)
	diff := gotextdiff.ToUnified("original", "modified", a, edits)

	result := fmt.Sprint(diff)
	if result == "" {
		return "(no differences)"
	}

	return result
}

// goudiffLib produces a unified diff using aymanbagabas/go-udiff.
func goudiffLib(a, b string) string {
	// Unified computes and formats the diff in one call
	diff := udiff.Unified("original", "modified", a, b)

	if diff == "" {
		return "(no differences)"
	}

	return diff
}

// sergiDiffLib produces a diff using sergi/go-diff.
// Note: This library produces character-level diffs by default,
// so we convert to a line-based format for comparison.
func sergiDiffLib(a, b string) (result string) {
	// Recover from panics - this library has known issues with certain inputs
	defer func() {
		if r := recover(); r != nil {
			result = fmt.Sprintf("PANIC: %v\n(sergi/go-diff has known stability issues with certain inputs)", r)
		}
	}()

	dmp := diffmatchpatch.New()

	// Method 1: Character-level diff (native)
	diffs := dmp.DiffMain(a, b, true)
	dmp.DiffCleanupSemantic(diffs)

	// Convert to a readable format
	var sb strings.Builder
	sb.WriteString("=== Character-level diff (native) ===\n")
	for _, d := range diffs {
		switch d.Type {
		case diffmatchpatch.DiffDelete:
			sb.WriteString(fmt.Sprintf("[-] %s\n", truncateForDisplay(d.Text)))
		case diffmatchpatch.DiffInsert:
			sb.WriteString(fmt.Sprintf("[+] %s\n", truncateForDisplay(d.Text)))
		case diffmatchpatch.DiffEqual:
			// Skip equal parts for brevity, but show context
			if len(d.Text) < 50 {
				sb.WriteString(fmt.Sprintf("[ ] %s\n", truncateForDisplay(d.Text)))
			} else {
				sb.WriteString(fmt.Sprintf("[ ] ...(%d chars)...\n", len(d.Text)))
			}
		}
	}

	// Method 2: Line-level diff (converted to patches)
	sb.WriteString("\n=== Line-level diff (converted to patches) ===\n")
	patches := dmp.PatchMake(a, diffs)
	sb.WriteString(dmp.PatchToText(patches))

	return sb.String()
}

// truncateForDisplay truncates long strings for display.
func truncateForDisplay(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\t", "\\t")
	if len(s) > 60 {
		return s[:57] + "..."
	}
	return s
}

// Benchmark comparison

func BenchmarkDiffLibs(b *testing.B) {
	tc := nestedStructDiffCase()
	originalStr := spewDump(tc.original)
	modifiedStr := spewDump(tc.modified)

	b.Run("our_difflib", func(b *testing.B) {
		for range b.N {
			ourDifflib(originalStr, modifiedStr)
		}
	})

	b.Run("gotextdiff", func(b *testing.B) {
		for range b.N {
			gotextdiffLib(originalStr, modifiedStr)
		}
	})

	b.Run("go_udiff", func(b *testing.B) {
		for range b.N {
			goudiffLib(originalStr, modifiedStr)
		}
	})

	b.Run("sergi_godiff", func(b *testing.B) {
		for range b.N {
			sergiDiffLib(originalStr, modifiedStr)
		}
	})
}
