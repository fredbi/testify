# Refactor Assertions Tests - Synthetic Skill

A comprehensive guide for refactoring tests in `internal/assertions/` to follow best practices, achieve high coverage, and maintain consistency.

## When to Use This Skill

Use this skill when refactoring test files in `internal/assertions/` to:
- Improve test organization and structure
- Add parallel execution support
- Consolidate and standardize test patterns
- Increase test coverage to 99%+
- Reduce code duplication
- Make tests more maintainable

## Core Architecture Principles

### Test Organization Structure

Tests in `internal/assertions/` follow a three-layer architecture:

**Layer 1: Source of Truth** (`internal/assertions/*.go`)
- Assertion implementations organized by domain
- Each assertion has comprehensive tests
- 94%+ coverage target

**Layer 2: Generated Tests** (`assert/`, `require/`)
- Smoke tests generated from "Examples:" in doc comments
- 100% mechanical coverage of generated forwarding code
- No error message testing (covered in Layer 1)

**Layer 3: Integration Tests** (via code generation)
- Meta-tests that verify code generation correctness
- Optional golden file testing

**Key Point**: Focus refactoring efforts on Layer 1 tests - they are the source of truth.

---

## Refactoring Patterns

### Pattern 1: Iterator Pattern for All Table-Driven Tests

**Mandate**: Use `iter.Seq[T]` for ALL table-driven tests with 2+ cases.

**Structure**:
```go
import (
	"iter"
	"slices"
	"testing"
)

// 1. Define explicit test case type
type testEqualCase struct {
	name       string
	expected   any
	actual     any
	shouldPass bool
}

// 2. Create iterator function
func testEqualCases() iter.Seq[testEqualCase] {
	return slices.Values([]testEqualCase{
		{
			name:       "integers equal",
			expected:   123,
			actual:     123,
			shouldPass: true,
		},
		{
			name:       "integers not equal",
			expected:   123,
			actual:     456,
			shouldPass: false,
		},
		// More cases...
	})
}

// 3. Test function iterates over cases
func TestEqual(t *testing.T) {
	t.Parallel()

	for c := range testEqualCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			mock := new(mockT)
			result := Equal(mock, c.expected, c.actual)

			if c.shouldPass {
				True(t, result)
				False(t, mock.Failed())
			} else {
				False(t, result)
				True(t, mock.Failed())
			}
		})
	}
}
```

**Benefits**:
- Clean separation of test data from test logic
- Easy to add new test cases
- Type-safe iteration
- Excellent for parallel execution
- Reusable iterator functions

### Pattern 2: Group Related Tests into Subtests

Transform flat test functions into structured subtests.

**Before**:
```go
import (
	"testing"
	"time"
)

func TestConditionEventuallyTrue(t *testing.T) {
	condition := func() bool { return true }
	True(t, Eventually(t, condition, 100*time.Millisecond, 20*time.Millisecond))
}

func TestConditionEventuallyFalse(t *testing.T) {
	mock := new(testing.T)
	condition := func() bool { return false }
	False(t, Eventually(mock, condition, 100*time.Millisecond, 20*time.Millisecond))
}
```

**After**:
```go
import "testing"

func TestConditionEventually(t *testing.T) {
	t.Parallel()

	t.Run("condition becomes true", func(t *testing.T) {
		t.Parallel()

		mock := new(errorsCapturingT)
		condition := func() bool { return true }
		True(t, Eventually(mock, condition, testTimeout, testTick))
	})

	t.Run("condition stays false", func(t *testing.T) {
		t.Parallel()

		mock := new(errorsCapturingT)
		condition := func() bool { return false }
		False(t, Eventually(mock, condition, testTimeout, testTick))
	})
}
```

### Pattern 3: Always Enable Parallel Execution

Add `t.Parallel()` at the start of:
- Each top-level test function
- Each subtest

**Pattern**:
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

**Critical**: Each subtest must declare its own variables (no shared state).

### Pattern 4: Extract Constants for Repeated Values

**Package-level constants**:
```go
import "time"

const (
	testTimeout = 100 * time.Millisecond
	testTick    = 20 * time.Millisecond
)
```

**Test-specific constants inside test functions**:
```go
import "testing"

func TestSomething(t *testing.T) {
	const expectedErrors = 4

	// Later in test:
	Len(t, mock.errors, expectedErrors, "expected 2 from condition, 2 from Eventually")
}
```

### Pattern 5: Consolidate Mock Types in `mock_test.go`

**Move all mock types to a dedicated file**:

```go
// mock_test.go
package assertions

import "context"

// Interface compliance verification
var (
    _ T         = &mockT{}
    _ T         = &errorsCapturingT{}
    _ failNower = &mockFailNowT{}
)

// mockT is a minimal mock implementation of T.
type mockT struct {
    failed bool
}

func (mockT) Helper() {}

func (m *mockT) Errorf(format string, args ...any) {
    m.failed = true
}

func (m *mockT) Failed() bool {
    return m.failed
}

// errorsCapturingT captures all errors for detailed assertions.
type errorsCapturingT struct {
    errors []error
    ctx    context.Context //nolint:containedctx // ok for test
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

**Delete duplicate mock definitions** from individual test files.

### Pattern 6: Testing Generic and Non-Generic Variants Together

When you have both generic (`GreaterT[V Ordered]`) and non-generic (`Greater`) versions:

**Structure**:
```go
import (
	"iter"
	"slices"
	"testing"
)

// 1. Single source of truth for test data
func greaterCases() iter.Seq[genericTestCase] {
	return slices.Values([]genericTestCase{
		{"int", testAllGreater(int(2), int(1), int(1), int(2))},
		{"float64", testAllGreater(2.0, 1.0, 1.0, 2.0)},
		{"string", testAllGreater("b", "a", "a", "b")},
		// Special types need dedicated functions
		{"time.Time", testGreaterTime()},
		{"[]byte", testGreaterBytes()},
	})
}

// 2. testAll* helper tests both variants with same input
func testAllGreater[V Ordered](successE1, successE2, failE1, failE2 V) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		t.Run("generic version", testGreaterT(successE1, successE2, failE1, failE2))
		t.Run("reflect version", testGreater(successE1, successE2, failE1, failE2))
	}
}

// 3. Individual test helpers for each variant
func testGreaterT[V Ordered](successE1, successE2, failE1, failE2 V) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		mock := new(mockT)

		True(t, GreaterT(mock, successE1, successE2))  // 2 > 1
		False(t, GreaterT(mock, failE1, failE2))       // 1 > 2
		False(t, GreaterT(mock, successE1, successE1)) // 2 > 2
	}
}

func testGreater(successE1, successE2, failE1, failE2 any) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		mock := new(mockT)

		True(t, Greater(mock, successE1, successE2))
		False(t, Greater(mock, failE1, failE2))
		False(t, Greater(mock, successE1, successE1))
	}
}

// 4. Test function
func TestCompareGreater(t *testing.T) {
	t.Parallel()

	for tc := range greaterCases() {
		t.Run(tc.name, tc.test)
	}
}
```

**Benefits**:
- Single source of truth for test cases
- Guaranteed consistency between generic and non-generic
- Easy to add new types (automatically tests both)
- Clear evidence both follow same logic

**Key insight**: Type parameters are resolved at closure creation time, not iteration time.

### Pattern 7: Minimize Unnecessary Type Arguments

Go can infer types in many cases. Use explicit type conversions instead of type parameters.

**Before**:
```go
{"int", testAllGreater[int](2, 1, 1, 2)},
{"float64", testAllGreater[float64](2.0, 1.0, 1.0, 2.0)},
{"float32", testAllGreater[float32](2.0, 1.0, 1.0, 2.0)},
```

**After**:
```go
{"int", testAllGreater(int(2), int(1), int(1), int(2))},       // int via conversion
{"float64", testAllGreater(2.0, 1.0, 1.0, 2.0)},               // inferred
{"float32", testAllGreater(float32(2.0), float32(1.0), float32(1.0), float32(2.0))},
```

### Pattern 8: Descriptive Subtest Names

**Format**: `"should <expected behavior>"` or `"<condition> should <behavior>"`

**Good examples**:
- `"should return true for equal values"`
- `"should fail when values differ"`
- `"context cancellation should stop polling"`
- `"should complete before timeout"`

**Avoid**:
- `"test case 1"`, `"case_a"`
- Names that just repeat function name

### Pattern 9: File Organization with Section Markers

Structure test files with clear sections following natural call flow:

```go
// ============================================================================
// Exported test functions
// ============================================================================

func TestEqual(t *testing.T) { /* ... */ }
func TestNotEqual(t *testing.T) { /* ... */ }

// ============================================================================
// Helper functions and test data for Equal
// ============================================================================

func equalCases() iter.Seq[testEqualCase] { /* ... */ }
func testEqual(...) func(*testing.T) { /* ... */ }

// ============================================================================
// Helper functions and test data for NotEqual
// ============================================================================

func notEqualCases() iter.Seq[testNotEqualCase] { /* ... */ }
func testNotEqual(...) func(*testing.T) { /* ... */ }
```

**Ordering principles**:
1. Exported `Test*` functions first
2. Natural call flow: functions appear closest below callers
3. Logical grouping of related helpers and test data
4. Clear comment separators

### Pattern 10: Extract Inline Functions for Testability

**Problem**: Inline anonymous functions cannot be tested independently.

**Before**:
```go
slices.SortFunc(result, func(a, b model.Ident) int {
    return strings.Compare(a.Name, b.Name)
})
```

**After**:
```go
slices.SortFunc(result, compareIdents)

// compareIdents compares two Idents by their Name field.
func compareIdents(a, b model.Ident) int {
    return strings.Compare(a.Name, b.Name)
}

// Then test it:
func TestCompareIdents(t *testing.T) {
    t.Parallel()
    for c := range compareIdentsCases() {
        t.Run(c.name, func(t *testing.T) {
            t.Parallel()
            result := compareIdents(c.a, c.b)
            // Assertions...
        })
    }
}
```

### Pattern 11: Unified Test Matrix for Multi-Assertion Testing

**When to use**: Testing multiple related assertions (6+) that operate on the same data with different semantics.

**Problem**: Testing 6 ordering assertions × 80+ collections × 2 implementations = massive duplication.

**Solution**: Organize by test case properties, not by assertion. Test all assertions for each collection.

#### The Breakthrough Insight

**❌ Wrong organization** (by assertion):
```go
func TestIsIncreasing(t *testing.T) {
    // Test all collections for IsIncreasing
    for allCollections { test IsIncreasing }
}
func TestIsDecreasing(t *testing.T) {
    // Test all collections for IsDecreasing (duplicate data!)
    for allCollections { test IsDecreasing }
}
// ... 4 more test functions with duplicate collections
```

**✅ Right organization** (by test case):
```go
import "testing"

func TestOrder(t *testing.T) {
	for collection := range unifiedOrderCases() {
		// Test all assertions for this collection
		testIsIncreasing(collection)
		testIsDecreasing(collection)
		testIsNonIncreasing(collection)
		testIsNonDecreasing(collection)
		testSorted(collection)
		testNotSorted(collection)
	}
}
```

**Why this is better**:
- Single source of truth for test data
- When a test fails, you see: collection → which assertions passed/failed
- Adding a new collection automatically tests all 6 assertions
- Zero duplication

#### Complete Pattern Structure

**Step 1: Define data properties (not assertion names)**

```go
// Describes the PROPERTY of the data, not what we're testing
type collectionKind int

const (
	allEqual        collectionKind = iota // all values equal
	strictlyAsc                           // strictly ascending (each < next)
	strictlyDesc                          // strictly descending (each > next)
	nonStrictlyAsc                        // non-strictly ascending (each <= next)
	nonStrictlyDesc                       // non-strictly descending (each >= next)
	unsorted                              // no ordering
	passAll                               // empty or single element
	errorCase                             // should fail with error
)
```

**Step 2: Define assertion semantics (separate from data)**

```go
// Describes WHAT we're testing
type orderAssertionKind int

const (
	increasingKind orderAssertionKind = iota
	notIncreasingKind
	decreasingKind
	notDecreasingKind
	sortedKind
	notSortedKind
)
```

**Step 3: Unified test data with properties**

```go
import (
	"iter"
	"slices"
	"time"
)

type orderTestCase struct {
	name           string
	collection     any
	kind           collectionKind // Data property
	reflectionOnly bool
}

func unifiedOrderCases() iter.Seq[orderTestCase] {
	return slices.Values([]orderTestCase{
		// Single source of truth - define each collection once
		{"all-equal/int", []int{2, 2, 2}, allEqual, false},
		{"strictly-asc/int", []int{1, 2, 3}, strictlyAsc, false},
		{"strictly-desc/int", []int{3, 2, 1}, strictlyDesc, false},
		{"non-strictly-asc/int", []int{1, 1, 2, 3}, nonStrictlyAsc, false},
		{"unsorted/int", []int{1, 4, 2}, unsorted, false},

		// Add all types for each property
		{"strictly-asc/float64", []float64{1.1, 2.2, 3.3}, strictlyAsc, false},
		{"strictly-asc/string", []string{"a", "b", "c"}, strictlyAsc, false},
		{"strictly-asc/time.Time", []time.Time{t0, t1, t2}, strictlyAsc, false},
		// ... more types and properties
	})
}
```

**Step 4: Algorithmic expected results**

```go
import "fmt"

// Determine pass/fail from data property + assertion semantics
func expectedStatusForAssertion(assertionKind orderAssertionKind, kind collectionKind) bool {
	// Error cases always fail
	if kind == errorCase {
		return false
	}

	switch assertionKind {
	case increasingKind:
		// IsIncreasing: only strictly ascending passes
		return kind == strictlyAsc || kind == passAll
	case notIncreasingKind:
		// IsNonIncreasing: NOT strictly ascending
		return kind != strictlyAsc && kind != passAll
	case decreasingKind:
		// IsDecreasing: only strictly descending passes
		return kind == strictlyDesc || kind == passAll
	case notDecreasingKind:
		// IsNonDecreasing: NOT strictly descending
		return kind != strictlyDesc && kind != passAll
	case sortedKind:
		// SortedT: non-strictly ascending (allows equal)
		return kind == allEqual || kind == strictlyAsc || kind == nonStrictlyAsc || kind == passAll
	case notSortedKind:
		// NotSortedT: inverse of SortedT
		return kind != allEqual && kind != strictlyAsc && kind != nonStrictlyAsc && kind != passAll
	default:
		panic(fmt.Errorf("invalid orderAssertionKind: %d", assertionKind))
	}
}
```

**Step 5: Test all assertions for each collection**

```go
import "testing"

func TestOrder(t *testing.T) {
	t.Parallel()

	for tc := range unifiedOrderCases() {
		t.Run(tc.name, func(t *testing.T) {
			// Test all assertions for this collection
			t.Run("with IsIncreasing", func(t *testing.T) {
				t.Parallel()
				shouldPass := expectedStatusForAssertion(increasingKind, tc.kind)
				t.Run("with reflection", testOrderReflect(IsIncreasing, tc.collection, shouldPass))
				if !tc.reflectionOnly {
					t.Run("with generic", testOrderGeneric(increasingKind, tc.collection, shouldPass))
				}
			})

			t.Run("with IsDecreasing", func(t *testing.T) {
				t.Parallel()
				shouldPass := expectedStatusForAssertion(decreasingKind, tc.kind)
				t.Run("with reflection", testOrderReflect(IsDecreasing, tc.collection, shouldPass))
				if !tc.reflectionOnly {
					t.Run("with generic", testOrderGeneric(decreasingKind, tc.collection, shouldPass))
				}
			})

			// ... test all other assertions
		})
	}
}
```

**Step 6: Enum-based dispatch (avoid reflection hacks)**

```go
import (
	"fmt"
	"testing"
)

func testOrderGeneric(assertionKind orderAssertionKind, collection any, shouldPass bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		mock := new(mockT)
		var result bool

		// Type switch to call generic variants
		switch coll := collection.(type) {
		case []int:
			result = callGenericAssertion(mock, assertionKind, coll)
		case []float64:
			result = callGenericAssertion(mock, assertionKind, coll)
		case []string:
			result = callGenericAssertion(mock, assertionKind, coll)
		// ... all other types
		default:
			t.Fatalf("unsupported collection type: %T", coll)
		}

		// Verify result matches expected
		if shouldPass {
			t.Run("should pass", func(t *testing.T) {
				if !result || mock.Failed() {
					t.Errorf("expected to pass")
				}
			})
		} else {
			t.Run("should fail", func(t *testing.T) {
				if result || !mock.Failed() {
					t.Errorf("expected to fail")
				}
			})
		}
	}
}

func callGenericAssertion[E Ordered](mock T, assertionKind orderAssertionKind, collection []E) bool {
	// Dispatch based on enum, not reflection pointer comparison
	switch assertionKind {
	case increasingKind:
		return IsIncreasingT(mock, collection)
	case decreasingKind:
		return IsDecreasingT(mock, collection)
	case notIncreasingKind:
		return IsNonIncreasingT(mock, collection)
	case notDecreasingKind:
		return IsNonDecreasingT(mock, collection)
	case sortedKind:
		return SortedT(mock, collection)
	case notSortedKind:
		return NotSortedT(mock, collection)
	default:
		panic(fmt.Errorf("invalid orderAssertionKind: %d", assertionKind))
	}
}
```

#### Benefits of This Pattern

**Before unified approach** (order_test.go old version):
- 660 lines with massive duplication
- Separate test functions for each assertion
- Duplicate test data across 6 functions
- Hard to see relationships between assertions
- Adding a type requires updating 6+ places

**After unified approach** (order_test.go new version):
- 503 lines, zero data duplication
- Single source of truth for test data
- Algorithmic expected results (no manual specification)
- Clear separation: data properties vs assertion semantics
- Adding a type updates automatically for all assertions
- When test fails: see collection → which assertions pass/fail

**Key insight**: The mental model shift from "test each assertion across all collections" to "test all assertions for each collection" is transformative for maintainability and debugging.

#### Error Message Testing

Keep error message testing separate (don't mix with functional testing):

```go
import (
	"iter"
	"slices"
	"strings"
	"testing"
)

func TestOrderErrorMessages(t *testing.T) {
	t.Parallel()

	for tc := range orderErrorMessageCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := newOutputMock()
			result := tc.fn(mock, tc.collection, tc.msgAndArgs...)
			if result {
				t.Errorf("expected to fail")
				return
			}

			if !strings.Contains(mock.buf.String(), tc.expectedInMsg) {
				t.Errorf("expected error message to contain: %s but got %q",
					tc.expectedInMsg, mock.buf.String())
			}
		})
	}
}

type errorMessageTestCase struct {
	name          string
	fn            func(T, any, ...any) bool
	collection    any
	msgAndArgs    []any
	expectedInMsg string
}

func orderErrorMessageCases() iter.Seq[errorMessageTestCase] {
	return slices.Values([]errorMessageTestCase{
		// Test msgAndArgs formatting
		{
			"IsIncreasing/with-msgAndArgs", IsIncreasing,
			[]int{2, 1},
			[]any{"format %s %x", "this", 0xc001},
			"format this c001\n",
		},

		// Test specific error messages
		{
			"IsIncreasing/string", IsIncreasing,
			[]string{"b", "a"},
			nil, `"b" is not less than "a"`,
		},
		{
			"IsDecreasing/int", IsDecreasing,
			[]int{1, 2},
			nil, `"1" is not greater than "2"`,
		},
		// ... more error message tests
	})
}
```

#### When to Use This Pattern

**Use this pattern when**:
- Testing 6+ related assertions that operate on similar data
- Assertions have clear semantic relationships (e.g., X vs NotX, strict vs non-strict)
- Test data can be categorized by properties (kinds)
- You want algorithmic correctness verification

**Don't use this pattern when**:
- Testing 1-2 unrelated assertions
- Assertions don't share test data
- Simpler patterns (6-10) are sufficient

#### Real-World Example

See `internal/assertions/order_test.go` for complete implementation:
- 80+ test collections (all numeric types, strings, time.Time, []byte)
- 6 ordering assertions (IsIncreasing, IsNonIncreasing, IsDecreasing, IsNonDecreasing, SortedT, NotSortedT)
- Both reflection and generic implementations tested
- 503 lines, zero duplication
- 100% test coverage

---

## Coverage Strategy

### Target: 99%+ Coverage

**Interpretation**:
- **99-100%**: Excellent, production-ready
- **95-99%**: Good, a few edge cases missed
- **90-95%**: Acceptable, needs more cases
- **<90%**: Needs significant work

### What's OK to Leave Uncovered

- Defensive nil checks that "can't happen" (mark with `// safeguard` comment)
- Panic paths in defensive code (document why they're unreachable)
- Extremely rare error conditions

### Standard Edge Cases to Test

Always include these categories:

1. **Empty/Zero Values**:
   - Empty strings, nil slices, zero numbers
   - Example: `""`; `[]int(nil)`; `0`

2. **Single Element**:
   - Lists with one item
   - Example: `[]int{1}`

3. **Multiple Elements**:
   - Normal case with several items
   - Example: `[]int{1, 2, 3, 4, 5}`

4. **Boundary Conditions**:
   - Maximum/minimum values
   - Example: `math.MaxInt64`; `math.MinInt64`

5. **Special Values**:
   - NaN, Inf for floats
   - Nil interfaces vs typed nil
   - Example: `math.NaN()`; `var x *int = nil`

6. **Type Variations** (for generic functions):
   - All numeric types: `int`, `int8`, `int16`, `int32`, `int64`, `uint`, `uint8`, `uint16`, `uint32`, `uint64`, `uintptr`, `float32`, `float64`
   - String types: `string`
   - Special types: `time.Time`, `[]byte`
   - Custom types: `type myInt int`

7. **Error Conditions**:
   - Expected failures
   - Invalid inputs
   - Context cancellation

### Defensive Code Pattern

When code has defensive checks that "shouldn't happen":

```go
parts := strings.SplitN(identifier, ".", 2)
if len(parts) != 2 {
    // Defensive code: This should never happen with the current regex pattern,
    // which requires at least one dot. If this triggers, it indicates a bug in
    // the regex pattern that needs to be fixed.
    panic(fmt.Errorf("internal error: pattern matched %q but split into %d parts",
        identifier, len(parts)))
}
```

**When to panic vs return error**:
- **Panic**: Code generators, build-time tools, programming errors
- **Return error**: Runtime code, user input validation, expected failures

### Coverage Analysis Workflow

```bash
# Check overall coverage
go test -cover ./internal/assertions

# Generate detailed coverage report
go test -coverprofile=coverage.out ./internal/assertions
go tool cover -func=coverage.out

# View coverage in HTML
go tool cover -html=coverage.out -o coverage.html
# Open coverage.html in browser

# Find specific uncovered lines
go tool cover -func=coverage.out | grep -v "100.0%"
```

---

## Special Patterns for Assertions Tests

### Testing Variadic Parameters

Always test functions with variadic parameters in multiple configurations:

```go
{
    name: "with message and args",
    test: func(t *testing.T) {
        mock := new(errorsCapturingT)
        False(t, Equal(mock, 1, 2, "expected %d", 2))
        if len(mock.errors) != 1 {
            t.Fatal("expected 1 error")
        }
        if !strings.Contains(mock.errors[0].Error(), "expected 2") {
            t.Errorf("expected formatted message, got: %v", mock.errors[0])
        }
    },
},
{
    name: "without message",
    test: func(t *testing.T) {
        mock := new(errorsCapturingT)
        False(t, Equal(mock, 1, 2))
        // Should still work without custom message
    },
},
```

### Testing Context Support

For functions that use `t.Context()`:

```go
t.Run("should respect context cancellation", func(t *testing.T) {
    t.Parallel()

    parentCtx, cancel := context.WithCancel(context.WithoutCancel(t.Context()))
    mock := new(errorsCapturingT).WithContext(parentCtx)

    condition := func() bool {
        time.Sleep(testTick)
        cancel()  // Cancel during execution
        time.Sleep(2 * testTick)
        return true
    }

    False(t, Eventually(mock, condition, testTimeout, testTick))

    // Verify context cancellation was detected
    var foundCancellationError bool
    for _, err := range mock.errors {
        if strings.Contains(err.Error(), "context canceled") {
            foundCancellationError = true
            break
        }
    }
    True(t, foundCancellationError, "expected context cancellation error")
})
```

### Testing Error Messages

When testing error message content:

```go
t.Run("error message should include details", func(t *testing.T) {
    t.Parallel()

    mock := new(errorsCapturingT)
    False(t, Equal(mock, 123, 456))

    if len(mock.errors) != 1 {
        t.Fatalf("expected 1 error, got %d", len(mock.errors))
    }

    msg := mock.errors[0].Error()
    if !strings.Contains(msg, "123") {
        t.Errorf("error should mention expected value 123, got: %s", msg)
    }
    if !strings.Contains(msg, "456") {
        t.Errorf("error should mention actual value 456, got: %s", msg)
    }
})
```

### Testing Type-Specific Behavior

For assertions that behave differently based on type:

```go
{
    name: "compares slices element-wise",
    test: func(t *testing.T) {
        mock := new(mockT)
        a := []int{1, 2, 3}
        b := []int{1, 2, 3}
        // Slices with same elements should be equal
        True(t, Equal(mock, a, b))
    },
},
{
    name: "nil slice equals empty slice",
    test: func(t *testing.T) {
        mock := new(mockT)
        var nilSlice []int
        emptySlice := []int{}
        // This is a design decision - document it
        True(t, Equal(mock, nilSlice, emptySlice))
    },
},
```

---

## Refactoring Workflow

### Step-by-Step Process

1. **Identify the test file to refactor**
   - Start with files that have many similar test functions
   - Look for files without parallel execution
   - Find files with duplicated mock types

2. **Read and understand existing tests**
   - What do they test?
   - What are the edge cases?
   - Are there patterns across multiple test functions?

3. **Group related tests**
   - Identify test functions that test the same assertion
   - Group by: success cases, failure cases, edge cases

4. **Create iterator functions**
   - Extract test data into iterator functions
   - Define explicit test case types
   - Move test logic to dedicated test helpers

5. **Add parallel execution**
   - Add `t.Parallel()` to all tests and subtests
   - Ensure no shared state between subtests
   - Each subtest declares its own variables

6. **Extract constants**
   - Find repeated values (timeouts, expected counts, etc.)
   - Define as package-level or test-local constants

7. **Consolidate mocks**
   - Check if mock type already exists in `mock_test.go`
   - If not, add it there
   - Remove duplicate definitions from test file
   - Add interface compliance verification

8. **Add section markers**
   - Organize file with clear comment separators
   - Group exported tests first
   - Group helpers below the tests that use them

9. **Fill coverage gaps**
   - Run coverage analysis: `go test -cover`
   - Identify uncovered lines
   - Add test cases for missing edge cases
   - Document defensive code that can't be tested

10. **Verify and polish**
    - Run tests: `go test -v -race ./internal/assertions`
    - Check coverage: `go test -coverprofile=coverage.out ./internal/assertions`
    - View coverage: `go tool cover -html=coverage.out`
    - Verify all tests pass and coverage is 99%+

### Checklist for Each Test File

- [ ] All `Test*` functions have `t.Parallel()`
- [ ] All subtests have `t.Parallel()`
- [ ] No shared state between parallel subtests
- [ ] Iterator pattern used for all table-driven tests (2+ cases)
- [ ] Repeated values extracted to constants
- [ ] Mock types consolidated in `mock_test.go`
- [ ] File organized with clear section markers
- [ ] Subtest names are descriptive
- [ ] Edge cases covered (empty, single, multiple, boundary)
- [ ] Generic and non-generic variants use same test data
- [ ] Coverage is 99%+ (run `go test -cover`)
- [ ] All tests pass with race detector (`go test -race`)

---

## Examples from Codebase

**Good examples to reference**:
- `internal/assertions/order_test.go` - **★ Best example** - Unified test matrix pattern (Pattern 11), organizing by test case instead of by assertion, zero duplication, 80+ collections × 6 assertions
- `internal/assertions/string_test.go` - 3-level dispatching (cases × types × assertion variants), comprehensive type coverage
- `internal/assertions/number_test.go` - 2-level dispatching ((cases × types) × assertion variants), generic/non-generic pattern
- `internal/assertions/condition_test.go` - Subtest grouping, parallel execution, constants
- `internal/assertions/mock_test.go` - Mock consolidation, interface compliance
- `internal/assertions/equal_test.go` - Comprehensive edge case coverage

**Patterns demonstrated**:
- Unified test matrix for multi-assertion testing (Pattern 11)
- Iterator pattern with `iter.Seq[T]`
- Generic and non-generic testing with single source of truth
- Enum-based dispatch instead of reflection hacks
- Algorithmic expected result determination
- Mock consolidation and interface compliance verification
- Parallel execution at all levels
- Clear file organization with section markers
- Separate error message testing
- Comprehensive edge case coverage achieving 99%+

---

## Common Mistakes to Avoid

❌ **Forcing 100% coverage**: Don't write artificial tests for unreachable code
❌ **Shared state in parallel tests**: Each subtest must declare its own variables
❌ **Inline test data**: Use iterator pattern instead
❌ **Vague test names**: Use descriptive names like "should fail when..."
❌ **Missing parallel execution**: Always add `t.Parallel()`
❌ **Duplicate mocks**: Consolidate in `mock_test.go`
❌ **Missing edge cases**: Test empty, single, multiple, boundary cases
❌ **Type arguments when unnecessary**: Use type conversions instead
❌ **Inconsistent test data**: Generic and non-generic should use same test data
❌ **Organizing by assertion instead of by test case**: For multi-assertion testing, organize by test case (test all assertions for each collection) rather than by assertion (test all collections for each assertion)

---

## Summary

**Core Patterns**:
1. Iterator pattern (`iter.Seq[T]`) for all table-driven tests
2. Parallel execution everywhere (`t.Parallel()`)
3. Mock consolidation in `mock_test.go`
4. Generic/non-generic variants share single test data source
5. Clear file organization with section markers
6. Descriptive subtest names
7. Extract constants for repeated values
8. Comprehensive edge case coverage
9. 99%+ coverage target
10. Minimize unnecessary type arguments
11. **Unified test matrix for multi-assertion testing** - organize by test case (data property), not by assertion

**Goals**:
- High test coverage (99%+)
- Fast parallel execution
- Maintainable test code
- Clear test organization
- Comprehensive edge case coverage
- Consistent patterns across all test files
- Zero duplication through unified test data

**Result**:
- Tests that are easy to understand and maintain
- High confidence in assertion correctness
- Fast test execution through parallelization
- Single source of truth for test data
- Clear evidence that generic and non-generic variants follow same logic
- Algorithmic correctness verification for related assertions
- Transformative debugging experience: see collection → which assertions pass/fail
