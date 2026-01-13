// SPDX-FileCopyrightText: Copyright 2025 go-swagger maintainers
// SPDX-License-Identifier: Apache-2.0

package difflib

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-openapi/testify/v2/internal/spew"
)

func BenchmarkSplitLines100(b *testing.B) {
	b.Run("splitLines", benchmarkSplitLines(100))
}

func BenchmarkSplitLines10000(b *testing.B) {
	b.Run("splitLines", benchmarkSplitLines(10000))
}

func benchmarkSplitLines(count int) func(*testing.B) {
	return func(b *testing.B) {
		str := strings.Repeat("foo\n", count)

		b.ResetTimer()

		n := 0
		for b.Loop() {
			n += len(SplitLines(str))
		}
	}
}

func BenchmarkDiffLib(b *testing.B) {
	tc := nestedStructDiffCase()
	originalStr := spewDump(tc.original)
	modifiedStr := spewDump(tc.modified)

	b.Run("our_difflib", func(b *testing.B) {
		for range b.N {
			ourDifflib(originalStr, modifiedStr)
		}
	})
}

// spewDump serializes a value using our internalized spew.
func spewDump(v any) string {
	return spew.Sdump(v)
}

// ourDifflib produces a unified diff using our internalized difflib.
func ourDifflib(a, b string) string {
	diff := UnifiedDiff{
		A:        SplitLines(a),
		B:        SplitLines(b),
		FromFile: "original",
		ToFile:   "modified",
		Context:  3,
	}

	result, err := GetUnifiedDiffString(diff)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	if result == "" {
		return "(no differences)"
	}

	return result
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
