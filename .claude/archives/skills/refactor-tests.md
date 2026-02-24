# Skill: Refactor Go Tests

This skill guides the refactoring of Go tests to follow the project's established patterns for clarity, maintainability, and parallel execution.

## When to Use This Skill

Use this skill when:
- Refactoring existing test files to improve structure
- Writing new tests that need to follow the project's patterns
- Consolidating multiple related test functions
- Adding parallel test execution support

## Core Refactoring Patterns

### 1. Group Related Tests into Subtests

Transform flat test functions into structured subtests using `t.Run()`.

**Before:**
```go
import (
	"testing"
	"time"
)

func TestConditionNeverFalse(t *testing.T) {
	condition := func() bool { return false }
	True(t, Never(t, condition, 100*time.Millisecond, 20*time.Millisecond))
}

func TestConditionNeverTrue(t *testing.T) {
	mock := new(testing.T)
	returns := make(chan bool, 2)
	returns <- false
	returns <- true
	defer close(returns)
	condition := func() bool { return <-returns }
	False(t, Never(mock, condition, 100*time.Millisecond, 20*time.Millisecond))
}

func TestConditionNeverFailQuickly(t *testing.T) {
	mock := new(testing.T)
	condition := func() bool { return true }
	False(t, Never(mock, condition, 100*time.Millisecond, time.Second))
}
```

**After:**
```go
import (
	"testing"
	"time"
)

func TestConditionNever(t *testing.T) {
	t.Parallel()

	t.Run("should never be true", func(t *testing.T) {
		t.Parallel()

		mock := new(errorsCapturingT)
		condition := func() bool { return false }
		True(t, Never(mock, condition, testTimeout, testTick))
	})

	t.Run("should never be true fails", func(t *testing.T) {
		t.Parallel()

		mock := new(errorsCapturingT)
		returns := make(chan bool, 2)
		returns <- false
		returns <- true
		defer close(returns)
		condition := func() bool { return <-returns }
		False(t, Never(mock, condition, testTimeout, testTick))
	})

	t.Run("should never be true fails, with ticker never triggered", func(t *testing.T) {
		t.Parallel()

		mock := new(errorsCapturingT)
		condition := func() bool { return true }
		False(t, Never(mock, condition, testTimeout, time.Second))
	})
}
```

### 2. Enable Parallel Execution

Add `t.Parallel()` at the start of:
- Each top-level test function
- Each subtest

```go
import "testing"

func TestSomething(t *testing.T) {
	t.Parallel() // Top-level parallel

	t.Run("case one", func(t *testing.T) {
		t.Parallel() // Subtest parallel

		// test code...
	})

	t.Run("case two", func(t *testing.T) {
		t.Parallel() // Subtest parallel

		// test code...
	})
}
```

### 3. Define Constants for Repeated Values

Replace magic numbers with descriptive constants at package level.

**Before:**
```go
Eventually(mock, condition, 100*time.Millisecond, 20*time.Millisecond)
```

**After:**
```go
const (
    testTimeout = 100 * time.Millisecond
    testTick    = 20 * time.Millisecond
)

// In tests:
Eventually(mock, condition, testTimeout, testTick)
```

### 4. Define Constants for Expected Values in Assertions

Use named constants instead of magic numbers in assertions.

**Before:**
```go
Len(t, mock.errors, 4, "expected errors")
```

**After:**
```go
const expectedErrors = 4
Len(t, mock.errors, expectedErrors, "expected 2 errors from the condition, and 2 additional errors from Eventually")
```

### 5. Consolidate Mock Types in `mock_test.go`

Move all mock types to a dedicated file and add interface compliance verification.

```go
// mock_test.go
package assertions

import (
    "bytes"
    "context"
    "fmt"
)

// Interface compliance verification
var (
    _ T         = &mockT{}
    _ T         = &mockFailNowT{}
    _ failNower = &mockFailNowT{}
    _ T         = &errorsCapturingT{}
    _ T         = &outputT{}
)

// errorsCapturingT is a mock implementation of TestingT that captures errors reported with Errorf.
type errorsCapturingT struct {
    errors []error
    ctx    context.Context //nolint:containedctx // this is ok to support context injection tests
}

func (errorsCapturingT) Helper() {}

func (t errorsCapturingT) Context() context.Context {
    if t.ctx == nil {
        return context.Background()
    }
    return t.ctx
}

func (t *errorsCapturingT) WithContext(ctx context.Context) *errorsCapturingT {
    t.ctx = ctx
    return t
}

func (t *errorsCapturingT) Errorf(format string, args ...any) {
    t.errors = append(t.errors, fmt.Errorf(format, args...))
}
```

### 6. Remove Duplicate Mock Definitions

Delete mock type definitions from individual test files after consolidating them in `mock_test.go`.

### 7. Use Descriptive Subtest Names

Format: `"should <expected behavior>"` or `"<condition> should <behavior>"`

**Good examples:**
- `"should never be true"`
- `"should fail on timeout"`
- `"should complete with false"`
- `"should succeed before the first tick"`
- `"condition should be true"`
- `"reported errors should include the context cancellation"`

**Avoid:**
- Names that just repeat the function being tested
- Vague names like `"test case 1"`

### 8. Declare Variables Inside Subtests

Each subtest should declare its own mock/state variables to avoid shared state between parallel tests.

**Before:**
```go
import "testing"

func TestSomething(t *testing.T) {
	mock := new(testing.T) // Shared - problematic with parallel

	t.Run("case one", func(t *testing.T) {
		// uses shared mock
	})

	t.Run("case two", func(t *testing.T) {
		// uses shared mock - race condition!
	})
}
```

**After:**
```go
import "testing"

func TestSomething(t *testing.T) {
	t.Parallel()

	t.Run("case one", func(t *testing.T) {
		t.Parallel()

		mock := new(errorsCapturingT) // Local to this subtest
		// test code...
	})

	t.Run("case two", func(t *testing.T) {
		t.Parallel()

		mock := new(errorsCapturingT) // Local to this subtest
		// test code...
	})
}
```

### 9. Use Nested Subtests for Related Assertions

When you need to make follow-up assertions after the main test logic:

```go
t.Run("should fail on parent test failed", func(t *testing.T) {
    t.Parallel()

    parentCtx, failParent := context.WithCancel(context.WithoutCancel(t.Context()))
    mock := new(errorsCapturingT).WithContext(parentCtx)
    condition := func() bool {
        time.Sleep(testTick)
        failParent()
        time.Sleep(2 * testTick)
        return true
    }

    False(t, Eventually(mock, condition, testTimeout, testTick))

    t.Run("reported errors should include the context cancellation", func(t *testing.T) {
        // Nested assertions about the error state
        Len(t, mock.errors, 2, "expected 2 error messages")

        var hasContextCancelled, hasFailedCondition bool
        for _, err := range mock.errors {
            msg := err.Error()
            switch {
            case strings.Contains(msg, "context canceled"):
                hasContextCancelled = true
            case strings.Contains(msg, "never satisfied"):
                hasFailedCondition = true
            }
        }
        True(t, hasContextCancelled, "expected a context cancelled error")
        True(t, hasFailedCondition, "expected a condition never satisfied error")
    })
})
```

### 10. Prefer Package Assertions Over Raw `t.Error()`

**Before:**
```go
if !Condition(mock, func() bool { return true }, "Truth") {
    t.Error("Condition should return true")
}
```

**After:**
```go
t.Run("condition should be true", func(t *testing.T) {
    t.Parallel()

    mock := new(testing.T)
    if !Condition(mock, func() bool { return true }, "Truth") {
        t.Error("condition should return true")  // lowercase for consistency
    }
})
```

Or when appropriate, use the assertion directly:
```go
True(t, Condition(mock, func() bool { return true }, "Truth"), "condition should return true")
```

### 11. Use Lowercase Error Messages

Error messages should be lowercase and descriptive.

**Before:**
```go
t.Error("Condition should return true")
```

**After:**
```go
t.Error("condition should return true")
```

### 12. Table-Driven Tests with Iterators for Generic/Non-Generic Variants

When testing both generic and non-generic versions of the same function, use a single test data source with helper functions to ensure both implementations follow identical logic.

**Pattern Structure:**
```go
import (
	"iter"
	"slices"
	"testing"
)

// 1. Test data iterator returns test cases
func deltaCases() iter.Seq[genericTestCase] {
	return slices.Values([]genericTestCase{
		{"simple/within-delta", testAllDelta(1.001, 1.0, 0.01, true)},
		{"simple/exceeds-delta", testAllDelta(1.0, 2.0, 0.5, false)},
		{"int/success", testAllDelta(int(2), int(1), int(1), true)},
		// ... more cases
	})
}

// 2. testAll* helper tests both variants with same input
func testAllDelta[Number Measurable](expected, actual, delta Number, shouldPass bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		if shouldPass {
			t.Run("should pass", func(t *testing.T) {
				t.Run("with InDelta", testDelta(expected, actual, delta, true))
				t.Run("with InDeltaT", testDeltaT(expected, actual, delta, true))
			})
		} else {
			t.Run("should fail", func(t *testing.T) {
				t.Run("with InDelta", testDelta(expected, actual, delta, false))
				t.Run("with InDeltaT", testDeltaT(expected, actual, delta, false))
			})
		}
	}
}

// 3. Individual test helpers for each variant
func testDelta[Number Measurable](expected, actual, delta Number, shouldPass bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		mock := new(mockT)
		// InDelta requires delta as float64, so convert it
		result := InDelta(mock, expected, actual, float64(delta))

		if shouldPass {
			True(t, result)
			False(t, mock.Failed())
		} else {
			False(t, result)
			True(t, mock.Failed())
		}
	}
}

func testDeltaT[Number Measurable](expected, actual, delta Number, shouldPass bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		mock := new(mockT)
		result := InDeltaT(mock, expected, actual, delta)

		if shouldPass {
			True(t, result)
			False(t, mock.Failed())
		} else {
			False(t, result)
			True(t, mock.Failed())
		}
	}
}

// 4. Test function loops over cases
func TestNumberInDelta(t *testing.T) {
	t.Parallel()

	// Optional: Add variant-specific tests here

	// Run all test cases with both InDelta and InDeltaT
	for tc := range deltaCases() {
		t.Run(tc.name, tc.test)
	}
}
```

**Benefits:**
- Single source of truth for test cases
- Guaranteed consistency between generic and non-generic variants
- Easy to add new test cases (automatically tests both variants)
- Reduced code duplication
- Clear evidence that both implementations follow the same logic

**Example from `number_test.go`:**
- `deltaCases()` → `testAllDelta()` → `testDelta()` + `testDeltaT()`
- `epsilonCases()` → `testAllEpsilon()` → `testEpsilon()` + `testEpsilonT()`
- `deltaSliceCases()` → `testDeltaSlice()` (single variant)
- `epsilonSliceCases()` → `testEpsilonSlice()` (single variant)

### 13. Minimize Unnecessary Type Arguments

Go can infer types in many cases. Remove explicit type arguments when the compiler can infer them.

**Before:**
```go
{"simple/within-delta", testAllDelta[float64](1.001, 1.0, 0.01, true)},
{"int/success", testAllDelta[int](2, 1, 1, true)},
{"float32/success", testAllDelta[float32](2.0, 1.0, 1.0, true)},
```

**After:**
```go
{"simple/within-delta", testAllDelta(1.001, 1.0, 0.01, true)},  // float64 inferred
{"int/success", testAllDelta(int(2), int(1), int(1), true)},    // int via conversion
{"float32/success", testAllDelta(float32(2.0), float32(1.0), float32(1.0), true)}, // float32 via conversion
```

**Guidelines:**
- Use explicit type conversions like `int(100)`, `float32(1.0)` instead of type parameters
- Only use type parameters `[T]` when the type cannot be inferred
- For float64 literals, no conversion needed (e.g., `1.0` is already float64)
- For integer literals with specific types, use conversions: `int8(100)`, `uint64(1000000000)`

**Benefits:**
- Cleaner, more concise code
- Follows Go idioms
- Avoids linter warnings about unnecessary type arguments
- Compiler handles type inference automatically

### 14. Organize Test Files with Clear Section Markers

Structure test files with clear sections and natural call flow.

**File organization pattern:**
```go
// ============================================================================
// Exported test functions
// ============================================================================

func TestNumberInDelta(t *testing.T) { /* ... */ }
func TestNumberInDeltaSlice(t *testing.T) { /* ... */ }
func TestNumberInEpsilon(t *testing.T) { /* ... */ }

// ============================================================================
// Helper functions and test data for InDelta/InDeltaT
// ============================================================================

func deltaCases() iter.Seq[genericTestCase] { /* ... */ }
func testAllDelta[Number Measurable](...) func(*testing.T) { /* ... */ }
func testDelta[Number Measurable](...) func(*testing.T) { /* ... */ }
func testDeltaT[Number Measurable](...) func(*testing.T) { /* ... */ }

// ============================================================================
// Helper functions and test data for InEpsilon/InEpsilonT
// ============================================================================

func epsilonCases() iter.Seq[genericTestCase] { /* ... */ }
func testAllEpsilon[Number Measurable](...) func(*testing.T) { /* ... */ }
func testEpsilon[Number Measurable](...) func(*testing.T) { /* ... */ }
func testEpsilonT[Number Measurable](...) func(*testing.T) { /* ... */ }
```

**Ordering principles:**
1. **Exported functions first**: All `Test*` functions at the top
2. **Natural call flow**: Functions appear closest below their callers
3. **Logical grouping**: Group related helpers and test data together
4. **Clear markers**: Use comment separators to delineate sections

**Benefits:**
- Easy to navigate large test files
- Clear separation of concerns
- Follows Go conventions (exported first, then helpers)
- Makes dependencies explicit through ordering

### 15. Extract Inline Check Closures into Named Helper Functions

When test case structs contain inline `func(*testing.T, ...)` closures for complex verification, extract them into named functions to reduce cognitive complexity and enable the removal of `//nolint` directives.

**Before:**
```go
//nolint:gocognit,gocyclo,cyclop
func testCases() iter.Seq[testCase] {
	return slices.Values([]testCase{
		{
			name: "complete index",
			setup: buildTestData(),
			check: func(t *testing.T, index Index) {
				t.Helper()
				count := 0
				for key, entry := range index.Entries() {
					count++
					switch key {
					case "alpha":
						if entry.Description() != "Alpha things" { t.Error("wrong description") }
						if len(entry.Items()) != 2 { t.Error("wrong count") }
					case "beta":
						if entry.Description() != "Beta things" { t.Error("wrong description") }
						// ... more checks
					default:
						t.Errorf("unexpected: %s", key)
					}
				}
				if count != 2 { t.Errorf("expected 2, got %d", count) }
			},
		},
	})
}
```

**After:**
```go
func testCases() iter.Seq[testCase] {
	return slices.Values([]testCase{
		{
			name:  "complete index",
			setup: buildTestData(),
			check: checkCompleteIndex, // named function reference
		},
	})
}

// Dispatcher uses a map to route per-key checks
func checkCompleteIndex(t *testing.T, index Index) {
	t.Helper()

	checkers := map[string]func(*testing.T, Entry){
		"alpha": checkAlphaEntry,
		"beta":  checkBetaEntry,
	}

	count := 0
	for key, entry := range index.Entries() {
		count++
		if checker, ok := checkers[key]; ok {
			checker(t, entry)
		} else {
			t.Errorf("unexpected: %s", key)
		}
	}

	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

// Per-key helpers are trivially simple
func checkAlphaEntry(t *testing.T, entry Entry) {
	t.Helper()
	if entry.Description() != "Alpha things" { t.Error("wrong description") }
	if len(entry.Items()) != 2 { t.Error("wrong count") }
}

func checkBetaEntry(t *testing.T, entry Entry) {
	t.Helper()
	if entry.Description() != "Beta things" { t.Error("wrong description") }
}
```

**When to apply:**
- An inline closure contains a `switch` or multi-branch `if/else` over iterated entries
- A `//nolint:gocognit,gocyclo,cyclop` directive is needed to silence linters
- The closure exceeds ~20 lines of assertions

**Technique: map-based dispatcher**
Replace a `switch` over dynamic keys with `map[string]func(*testing.T, ValueType)`. This:
- Eliminates deeply nested `switch` cases that spike cognitive complexity
- Makes each per-key check independently testable
- Keeps the dispatcher function flat (loop + map lookup + fallback error)

**Guidelines:**
- Name helpers after what they check: `checkEqualDomain`, `checkCompleteIndexMetadata`
- Always include `t.Helper()` in every extracted helper
- The case iterator function (`testCases()`) should only contain data setup and function references — no inline assertion logic
- Group extracted helpers under a `/* section comment */` near the cases they serve

**Example in codebase:**
See `codegen/internal/generator/domains/domains_test.go`:
- `checkCompleteIndexMetadata` — metadata assertions
- `checkCompleteIndexDomains` — map-based dispatcher over 4 domain checkers
- `checkEqualDomain`, `checkBooleanDomain`, `checkTestingDomain`, `checkCommonDomain` — per-domain helpers
- `checkDanglingDomainExclusion` — second test case's check function

## Refactoring Checklist

When refactoring a test file:

### Basic Refactoring (Patterns 1-11)
1. [ ] Identify related test functions that can be grouped
2. [ ] Create a parent test function with `t.Parallel()`
3. [ ] Convert each original test to a subtest with `t.Run()` and `t.Parallel()`
4. [ ] Extract repeated constants (timeouts, expected values)
5. [ ] Move mock types to `mock_test.go` if not already there
6. [ ] Remove duplicate mock definitions from the file
7. [ ] Ensure each subtest declares its own variables
8. [ ] Update subtest names to be descriptive
9. [ ] Add interface compliance checks for mocks
10. [ ] Run tests with `-race` flag to catch data races
11. [ ] Verify tests still pass: `go test -v -race ./...`

### Advanced Refactoring (Patterns 12-15)
12. [ ] For generic/non-generic pairs, consolidate into table-driven tests with iterators
13. [ ] Remove unnecessary type arguments (use type conversions instead)
14. [ ] Organize file with clear section markers and natural call flow
15. [ ] Extract inline check closures into named helpers; use map-based dispatchers for switch-over-keys
16. [ ] Verify linter warnings are resolved (no `//nolint` directives needed)
17. [ ] Check that all test variants use identical test data

## Example Refactorings

### Basic Refactoring (Patterns 1-11)
See the diff between commits:
- `c217cc4a12fd6c9bca890a15563fcbd16c96c7fe` (before)
- `e6b0793ba519fb22dc1887392e1465649a5a95ff` (after)

Key files:
- `internal/assertions/condition_test.go` - Main test refactoring example
- `internal/assertions/mock_test.go` - Mock consolidation example

### Advanced Refactoring (Patterns 12-14)
See `internal/assertions/number_test.go` for examples of:
- **Table-driven tests with iterators** for generic/non-generic variants
  - `deltaCases()` → `testAllDelta()` → `testDelta()` + `testDeltaT()`
  - `epsilonCases()` → `testAllEpsilon()` → `testEpsilon()` + `testEpsilonT()`
- **Single source of truth** ensuring both implementations follow identical logic
- **Slice tests** using the same iterator pattern
  - `deltaSliceCases()` → `testDeltaSlice()`
  - `epsilonSliceCases()` → `testEpsilonSlice()`
- **Type inference** with explicit conversions instead of type parameters
- **Clear organization** with section markers and natural call flow

**Results from number_test.go refactoring:**
- Reduced from 791 to 667 lines (-15.7%)
- Eliminated redundancy between `InDelta`/`InDeltaT` and `InEpsilon`/`InEpsilonT`
- Single test data source guarantees consistency
- All linter warnings resolved

### Complexity Reduction (Pattern 15)
See `codegen/internal/generator/domains/domains_test.go` for examples of:
- **Inline closure extraction**: `checkMetadata` and `checkDomains` closures replaced by named function references
- **Map-based dispatcher**: `checkCompleteIndexDomains` uses `map[string]func(*testing.T, Entry)` instead of a `switch`
- **Per-key helpers**: `checkEqualDomain`, `checkBooleanDomain`, `checkTestingDomain`, `checkCommonDomain`

**Results from domains_test.go refactoring:**
- Removed `//nolint:gocognit,gocyclo,cyclop` directive
- Case iterator function reduced to pure data setup + function references
- Each extracted helper is trivially simple (3-6 assertions)
