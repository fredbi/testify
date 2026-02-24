# equal_test.go Refactoring Plan

## Executive Summary

**Target**: `internal/assertions/equal_test.go` (1,423 lines, 18 test functions)

**Goal**: Apply unified test matrix pattern (Pattern 11) to reduce duplication and improve maintainability

**Expected Impact**: 185-245 lines saved (13-17% reduction), zero duplication in high-priority areas

## Current State Analysis

### File Statistics
- **Total lines**: 1,423
- **Test functions**: 18
- **Coverage**: High, but with significant duplication
- **Pattern**: Mix of isolated test functions with duplicated logic

### Duplication Hotspots (Priority Order)

1. **EqualT/NotEqualT** - 193 lines total
   - Identical type switch logic in both functions
   - Savings potential: ~50% (~96 lines)
   - Priority: **HIGHEST**

2. **Empty/NotEmpty** - 104 lines total
   - Symmetric logic for opposite assertions
   - Savings potential: ~40% (~42 lines)
   - Priority: **HIGH**

3. **SameT/NotSameT** - 72 lines total
   - Symmetric logic for pointer comparison
   - Savings potential: ~45% (~32 lines)
   - Priority: **HIGH**

4. **Same/NotSame** - 57 lines total
   - Reflection-based pointer comparison
   - Savings potential: ~45% (~25 lines)
   - Priority: **MEDIUM**

5. **Nil/NotNil** - 33 lines total
   - Simple but duplicated
   - Savings potential: ~70% (~23 lines)
   - Priority: **MEDIUM**

**Total potential savings**: 185-245 lines (13-17% reduction)

## Semantic Foundation

### Equality Domain Assertions (16 total)

#### Unary Assertions (4)
Tests properties of single objects:
- `Nil(t, obj)` - Is obj nil?
- `NotNil(t, obj)` - Is obj not nil?
- `Empty(t, obj)` - Is obj empty? (len=0, nil, false, 0, "")
- `NotEmpty(t, obj)` - Is obj not empty?

#### Binary Assertions (12)
Tests relationships between two objects:

**Equality variants:**
- `Equal(t, expected, actual)` - Deep equality (reflection-based)
- `EqualT[T](t, expected, actual)` - Deep equality (generic, comparable constraint)
- `NotEqual(t, expected, actual)` - Not deeply equal (reflection)
- `NotEqualT[T](t, expected, actual)` - Not deeply equal (generic)

**Value equality:**
- `EqualValues(t, expected, actual)` - Value equality (coercion allowed)
- `NotEqualValues(t, expected, actual)` - Not value equal

**Pointer identity:**
- `Same(t, expected, actual)` - Same memory address (reflection)
- `SameT[T](t, expected, actual)` - Same memory address (generic)
- `NotSame(t, expected, actual)` - Different memory addresses (reflection)
- `NotSameT[T](t, expected, actual)` - Different memory addresses (generic)

**Element-wise:**
- `ElementsMatch(t, expected, actual)` - Same elements, any order
- `NotElementsMatch(t, expected, actual)` - Different element sets

### Object Categories

**For single objects (unary):**
1. `nil` - Nil values
2. `empty-non-nil` - Empty but not nil ([], "", 0, false)
3. `non-empty-comparable` - Non-empty values that can be compared
4. `non-empty-non-comparable` - Non-empty values with unexported fields, etc.

**For object pairs (binary):**
1. `both-nil` - Both objects are nil
2. `one-nil` - One nil, one non-nil
3. `same-identity` - Same memory address (a == b)
4. `equal-value-comparable` - Equal values, different addresses
5. `equal-value-non-comparable` - Equal values but can't use ==
6. `different-value-same-type` - Different values, same type
7. `different-value-coercible` - Different values, but EqualValues could work (int vs float)
8. `different-type-incomparable` - Completely different types

### Semantic Relationships

#### Truth Tables for Unary Assertions

| Category | Nil | NotNil | Empty | NotEmpty |
|----------|-----|--------|-------|----------|
| nil | ✅ | ❌ | ✅ | ❌ |
| empty-non-nil | ❌ | ✅ | ✅ | ❌ |
| non-empty-comparable | ❌ | ✅ | ❌ | ✅ |
| non-empty-non-comparable | ❌ | ✅ | ❌ | ✅ |

#### Implication Chains for Binary Assertions

```
same-identity implies:
  ✅ Same/SameT
  ✅ Equal/EqualT (for comparable types)
  ✅ EqualValues
  ❌ NotSame/NotSameT
  ❌ NotEqual/NotEqualT
  ❌ NotEqualValues

both-nil implies:
  ✅ Equal/EqualT
  ✅ EqualValues
  Note: Same/SameT behavior varies (pointers vs values)

equal-value-comparable implies:
  ✅ Equal/EqualT
  ✅ EqualValues
  ❌ NotEqual/NotEqualT
  Note: Same/SameT = false (different addresses)

different-value implies:
  ❌ Equal/EqualT
  ❌ Same/SameT
  ✅ NotEqual/NotEqualT
  ✅ NotSame/NotSameT
```

## Refactoring Strategy: Unified Test Matrix Pattern

### Pattern Overview

Apply Pattern 11 from refactor-assertions-tests.md:

1. **Define test case properties** - Categorize test data by semantic properties
2. **Enumerate assertion kinds** - Create enum for assertion types being tested
3. **Encode semantics algorithmically** - Map (property, assertion) → expected result
4. **Single source of test data** - One iterator for all test cases
5. **Test all assertions per case** - Nested subtests for each assertion
6. **Type dispatch for generics** - Handle generic vs reflection variants

### Success Model: order_test.go

Before: 660+ lines, massive duplication
After: 503 lines, zero duplication, algorithmic expected results

Key insight: Organize by test case properties, not by assertion functions

## Phased Implementation Plan

### Phase 1: Foundation (Unary Assertions)

**Target**: Nil, NotNil, Empty, NotEmpty (4 assertions)

**Step 1.1: Define test case structure**
```go
type objectCategory int

const (
	nilCategory objectCategory = iota
	emptyNonNil
	nonEmptyComparable
	nonEmptyNonComparable
	errorCase
)

type unaryTestCase struct {
	name     string
	object   any
	category objectCategory
}
```

**Step 1.2: Create unified test data**
```go
import (
	"iter"
	"slices"
)

func unifiedUnaryCases() iter.Seq[unaryTestCase] {
	return slices.Values([]unaryTestCase{
		// Nil cases
		{"nil/nil-ptr", (*int)(nil), nilCategory},
		{"nil/nil-slice", []int(nil), nilCategory},
		{"nil/nil-map", map[string]int(nil), nilCategory},
		{"nil/nil-chan", (chan int)(nil), nilCategory},
		{"nil/nil-func", (func())(nil), nilCategory},
		{"nil/nil-interface", (any)(nil), nilCategory},

		// Empty but non-nil
		{"empty-non-nil/slice", []int{}, emptyNonNil},
		{"empty-non-nil/string", "", emptyNonNil},
		{"empty-non-nil/map", map[string]int{}, emptyNonNil},
		{"empty-non-nil/zero-int", 0, emptyNonNil},
		{"empty-non-nil/false", false, emptyNonNil},

		// Non-empty comparable
		{"non-empty/int", 42, nonEmptyComparable},
		{"non-empty/string", "hello", nonEmptyComparable},
		{"non-empty/slice", []int{1, 2}, nonEmptyComparable},
		{"non-empty/true", true, nonEmptyComparable},

		// Non-empty non-comparable (if needed)
		// ...
	})
}
```

**Step 1.3: Define assertion enum**
```go
type unaryAssertionKind int

const (
	nilKind unaryAssertionKind = iota
	notNilKind
	emptyKind
	notEmptyKind
)
```

**Step 1.4: Encode semantics**
```go
import "fmt"

func expectedStatusForUnaryAssertion(assertionKind unaryAssertionKind, category objectCategory) bool {
	if category == errorCase {
		return false
	}

	switch assertionKind {
	case nilKind:
		return category == nilCategory
	case notNilKind:
		return category != nilCategory
	case emptyKind:
		return category == nilCategory || category == emptyNonNil
	case notEmptyKind:
		return category == nonEmptyComparable || category == nonEmptyNonComparable
	default:
		panic(fmt.Sprintf("unknown assertion kind: %v", assertionKind))
	}
}
```

**Step 1.5: Implement unified test**
```go
import "testing"

func TestUnaryAssertions(t *testing.T) {
	t.Parallel()

	for tc := range unifiedUnaryCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("with Nil", testUnaryAssertion(tc, nilKind, Nil))
			t.Run("with NotNil", testUnaryAssertion(tc, notNilKind, NotNil))
			t.Run("with Empty", testUnaryAssertion(tc, emptyKind, Empty))
			t.Run("with NotEmpty", testUnaryAssertion(tc, notEmptyKind, NotEmpty))
		})
	}
}

func testUnaryAssertion(tc unaryTestCase, kind unaryAssertionKind, fn func(T, any, ...any) bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		mock := new(mockT)
		result := fn(mock, tc.object)
		expected := expectedStatusForUnaryAssertion(kind, tc.category)

		if result != expected {
			t.Errorf("expected %v, got %v for %s with category %v",
				expected, result, tc.name, tc.category)
		}
	}
}
```

**Expected savings**: ~50 lines (from 66 to ~16 + test cases)

### Phase 2: Pointer Identity (Generic + Reflection)

**Target**: Same, SameT, NotSame, NotSameT (4 assertions)

**Step 2.1: Define test case structure**
```go
type pairRelationship int

const (
	bothNil pairRelationship = iota
	oneNil
	sameIdentity
	equalValueDifferentAddress
	differentValue
	differentType
)

type pointerPairTestCase struct {
	name         string
	expected     any
	actual       any
	relationship pairRelationship
	genericOnly  bool // Can only test with generics
	reflectOnly  bool // Can only test with reflection
}
```

**Step 2.2: Create unified test data**
```go
import (
	"iter"
	"slices"
)

func unifiedPointerPairCases() iter.Seq[pointerPairTestCase] {
	v1 := 42
	v2 := 42

	return slices.Values([]pointerPairTestCase{
		// Both nil
		{"both-nil/ptr", (*int)(nil), (*int)(nil), bothNil, false, false},
		{"both-nil/slice", []int(nil), []int(nil), bothNil, false, false},

		// One nil
		{"one-nil/first", (*int)(nil), &v1, oneNil, false, false},
		{"one-nil/second", &v1, (*int)(nil), oneNil, false, false},

		// Same identity
		{"same-identity/ptr", &v1, &v1, sameIdentity, false, false},
		{"same-identity/slice", []int{1, 2}, nil, sameIdentity, false, false}, // Will set actual = expected

		// Equal value, different address
		{"equal-diff-addr/ptr", &v1, &v2, equalValueDifferentAddress, false, false},

		// Different value
		{"diff-value/int", 42, 43, differentValue, false, false},

		// Different type
		{"diff-type/int-string", 42, "42", differentType, false, true}, // Reflection only
	})
}
```

**Step 2.3: Define assertion enum**
```go
type pointerAssertionKind int

const (
	sameKind pointerAssertionKind = iota
	sameTKind
	notSameKind
	notSameTKind
)
```

**Step 2.4: Encode semantics**
```go
import "fmt"

func expectedStatusForPointerAssertion(assertionKind pointerAssertionKind, relationship pairRelationship) bool {
	positive := assertionKind == sameKind || assertionKind == sameTKind

	switch relationship {
	case sameIdentity:
		return positive
	case bothNil, equalValueDifferentAddress, differentValue, differentType:
		return !positive
	case oneNil:
		return !positive
	default:
		panic(fmt.Sprintf("unknown relationship: %v", relationship))
	}
}
```

**Step 2.5: Implement unified test with generic dispatch**
```go
import (
	"fmt"
	"testing"
)

func TestPointerIdentity(t *testing.T) {
	t.Parallel()

	for tc := range unifiedPointerPairCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !tc.genericOnly {
				t.Run("with Same", testPointerAssertion(tc, sameKind, Same))
				t.Run("with NotSame", testPointerAssertion(tc, notSameKind, NotSame))
			}

			if !tc.reflectOnly {
				t.Run("with SameT", testPointerAssertionT(tc, sameTKind))
				t.Run("with NotSameT", testPointerAssertionT(tc, notSameTKind))
			}
		})
	}
}

func testPointerAssertionT(tc pointerPairTestCase, kind pointerAssertionKind) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		mock := new(mockT)
		result := callPointerGenericAssertion(mock, kind, tc.expected, tc.actual)
		expected := expectedStatusForPointerAssertion(kind, tc.relationship)

		if result != expected {
			t.Errorf("expected %v, got %v", expected, result)
		}
	}
}

func callPointerGenericAssertion[T comparable](mock T, kind pointerAssertionKind, expected, actual T) bool {
	switch kind {
	case sameTKind:
		return SameT(mock, expected, actual)
	case notSameTKind:
		return NotSameT(mock, expected, actual)
	default:
		panic(fmt.Sprintf("unknown kind: %v", kind))
	}
}
```

**Expected savings**: ~60 lines (from 129 to ~69)

### Phase 3: Deep Equality (Highest Priority)

**Target**: Equal, EqualT, NotEqual, NotEqualT (4 assertions, 193 lines total)

**Step 3.1: Define comprehensive pair relationships**
```go
type equalityRelationship int

const (
	eqBothNil equalityRelationship = iota
	eqOneNil
	eqSameIdentity
	eqEqualValueComparable
	eqEqualValueNonComparable
	eqDifferentValueSameType
	eqDifferentType
	eqErrorCase
)
```

**Step 3.2: Create extensive test data**
```go
import (
	"iter"
	"slices"
)

func unifiedEqualityCases() iter.Seq[equalityTestCase] {
	return slices.Values([]equalityTestCase{
		// Both nil
		{"both-nil/ptr", (*int)(nil), (*int)(nil), eqBothNil, false, false},
		{"both-nil/slice", []int(nil), []int(nil), eqBothNil, false, false},
		{"both-nil/interface", (any)(nil), (any)(nil), eqBothNil, false, false},

		// One nil
		{"one-nil/first", (*int)(nil), ptr(42), eqOneNil, false, false},
		{"one-nil/second", ptr(42), (*int)(nil), eqOneNil, false, false},

		// Same identity
		{"same-identity/ptr", ptr(42), nil, eqSameIdentity, false, false}, // actual = expected in test
		{"same-identity/slice", []int{1, 2}, nil, eqSameIdentity, false, false},

		// Equal value, comparable
		{"equal-comparable/int", 42, 42, eqEqualValueComparable, false, false},
		{"equal-comparable/string", "hello", "hello", eqEqualValueComparable, false, false},
		{"equal-comparable/slice", []int{1, 2}, []int{1, 2}, eqEqualValueComparable, false, false},
		{"equal-comparable/struct", simpleStruct{X: 1}, simpleStruct{X: 1}, eqEqualValueComparable, false, false},

		// Equal value, non-comparable
		{"equal-non-comparable/slice-in-struct", structWithSlice{S: []int{1}}, structWithSlice{S: []int{1}}, eqEqualValueNonComparable, false, true},

		// Different value, same type
		{"diff-value/int", 42, 43, eqDifferentValueSameType, false, false},
		{"diff-value/string", "hello", "world", eqDifferentValueSameType, false, false},
		{"diff-value/slice-len", []int{1, 2}, []int{1, 2, 3}, eqDifferentValueSameType, false, false},
		{"diff-value/slice-elem", []int{1, 2}, []int{1, 3}, eqDifferentValueSameType, false, false},

		// Different type
		{"diff-type/int-int64", 42, int64(42), eqDifferentType, false, true},
		{"diff-type/int-string", 42, "42", eqDifferentType, false, true},

		// Type coverage: all numeric types
		{"equal-comparable/int8", int8(42), int8(42), eqEqualValueComparable, false, false},
		{"equal-comparable/int16", int16(42), int16(42), eqEqualValueComparable, false, false},
		{"equal-comparable/int32", int32(42), int32(42), eqEqualValueComparable, false, false},
		{"equal-comparable/int64", int64(42), int64(42), eqEqualValueComparable, false, false},
		{"equal-comparable/uint", uint(42), uint(42), eqEqualValueComparable, false, false},
		{"equal-comparable/uint8", uint8(42), uint8(42), eqEqualValueComparable, false, false},
		{"equal-comparable/uint16", uint16(42), uint16(42), eqEqualValueComparable, false, false},
		{"equal-comparable/uint32", uint32(42), uint32(42), eqEqualValueComparable, false, false},
		{"equal-comparable/uint64", uint64(42), uint64(42), eqEqualValueComparable, false, false},
		{"equal-comparable/float32", float32(42.0), float32(42.0), eqEqualValueComparable, false, false},
		{"equal-comparable/float64", float64(42.0), float64(42.0), eqEqualValueComparable, false, false},
		{"equal-comparable/complex64", complex64(42 + 0i), complex64(42 + 0i), eqEqualValueComparable, false, false},
		{"equal-comparable/complex128", complex128(42 + 0i), complex128(42 + 0i), eqEqualValueComparable, false, false},

		// Custom types
		{"equal-comparable/~int", myInt(42), myInt(42), eqEqualValueComparable, false, false},
		{"equal-comparable/~string", myString("hello"), myString("hello"), eqEqualValueComparable, false, false},
	})
}
```

**Step 3.3: Define assertion enum**
```go
type equalityAssertionKind int

const (
	equalKind equalityAssertionKind = iota
	equalTKind
	notEqualKind
	notEqualTKind
)
```

**Step 3.4: Encode semantics**
```go
import "fmt"

func expectedStatusForEqualityAssertion(assertionKind equalityAssertionKind, relationship equalityRelationship) bool {
	if relationship == eqErrorCase {
		return false
	}

	positive := assertionKind == equalKind || assertionKind == equalTKind

	switch relationship {
	case eqBothNil, eqSameIdentity, eqEqualValueComparable, eqEqualValueNonComparable:
		return positive
	case eqOneNil, eqDifferentValueSameType, eqDifferentType:
		return !positive
	default:
		panic(fmt.Sprintf("unknown relationship: %v", relationship))
	}
}
```

**Step 3.5: Implement unified test with type dispatch**
```go
import (
	"fmt"
	"testing"
)

func TestEquality(t *testing.T) {
	t.Parallel()

	for tc := range unifiedEqualityCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Handle same-identity case
			actual := tc.actual
			if actual == nil && tc.relationship == eqSameIdentity {
				actual = tc.expected
			}

			if !tc.genericOnly {
				t.Run("with Equal", testEqualityAssertion(tc, equalKind, Equal, tc.expected, actual))
				t.Run("with NotEqual", testEqualityAssertion(tc, notEqualKind, NotEqual, tc.expected, actual))
			}

			if !tc.reflectOnly {
				t.Run("with EqualT", testEqualityAssertionT(tc, equalTKind, tc.expected, actual))
				t.Run("with NotEqualT", testEqualityAssertionT(tc, notEqualTKind, tc.expected, actual))
			}
		})
	}
}

func testEqualityAssertionT(tc equalityTestCase, kind equalityAssertionKind, expected, actual any) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		mock := new(mockT)

		// Type switch to call appropriate generic variant
		result := false
		switch expected := expected.(type) {
		case int:
			result = callEqualityGenericAssertion(mock, kind, expected, actual.(int))
		case string:
			result = callEqualityGenericAssertion(mock, kind, expected, actual.(string))
		case []int:
			result = callEqualityGenericAssertion(mock, kind, expected, actual.([]int))
		case simpleStruct:
			result = callEqualityGenericAssertion(mock, kind, expected, actual.(simpleStruct))
		// ... more types
		default:
			t.Fatalf("unsupported type for generic test: %T", expected)
		}

		expectedResult := expectedStatusForEqualityAssertion(kind, tc.relationship)
		if result != expectedResult {
			t.Errorf("expected %v, got %v", expectedResult, result)
		}
	}
}

func callEqualityGenericAssertion[T comparable](mock T, kind equalityAssertionKind, expected, actual T) bool {
	switch kind {
	case equalTKind:
		return EqualT(mock, expected, actual)
	case notEqualTKind:
		return NotEqualT(mock, expected, actual)
	default:
		panic(fmt.Sprintf("unknown kind: %v", kind))
	}
}
```

**Expected savings**: ~96 lines (from 193 to ~97)

### Phase 4: Value Equality and ElementsMatch

**Target**: EqualValues, NotEqualValues, ElementsMatch, NotElementsMatch

This phase is lower priority and can use similar patterns to Phase 3.

### Phase 5: Error Message Tests

**Target**: Consolidate error message tests into single table-driven test

Similar to what we did in order_test.go:

```go
import (
	"iter"
	"slices"
	"strings"
	"testing"
)

type equalityErrorMessageCase struct {
	name          string
	fn            func(T, any, any, ...any) bool
	expected      any
	actual        any
	msgAndArgs    []any
	expectedInMsg string
}

func equalityErrorMessageCases() iter.Seq[equalityErrorMessageCase] {
	return slices.Values([]equalityErrorMessageCase{
		{
			name:          "Equal fails with diff",
			fn:            Equal,
			expected:      42,
			actual:        43,
			expectedInMsg: "Not equal",
		},
		// More cases...
	})
}

func TestEqualityErrorMessages(t *testing.T) {
	t.Parallel()

	for tc := range equalityErrorMessageCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := newOutputMock()
			result := tc.fn(mock, tc.expected, tc.actual, tc.msgAndArgs...)

			if result {
				t.Errorf("expected assertion to fail")
			}

			if !strings.Contains(mock.buf.String(), tc.expectedInMsg) {
				t.Errorf("expected error message to contain: %s but got %q",
					tc.expectedInMsg, mock.buf.String())
			}
		})
	}
}
```

## Implementation Order and Verification

### Order of Execution
1. **Phase 1**: Unary assertions (easiest, builds confidence)
2. **Phase 3**: Deep equality (highest impact, 96 lines saved)
3. **Phase 2**: Pointer identity (medium impact)
4. **Phase 4**: Value equality and ElementsMatch (lower impact)
5. **Phase 5**: Error message consolidation (cleanup)

### Verification After Each Phase
```bash
# Run tests
go test ./internal/assertions -run TestUnary    # After Phase 1
go test ./internal/assertions -run TestEquality # After Phase 3
go test ./internal/assertions -run TestPointer  # After Phase 2

# Check coverage
go test -cover ./internal/assertions
# Should maintain or improve 94%+ coverage

# Verify line count reduction
wc -l internal/assertions/equal_test.go

# Run full test suite
go test ./...
```

### Success Criteria
- ✅ All tests pass
- ✅ Coverage maintains 94%+
- ✅ Zero duplication in refactored sections
- ✅ Line count reduced by 185-245 lines
- ✅ Test cases organized by semantic properties
- ✅ Expected results computed algorithmically
- ✅ Both generic and reflection variants tested consistently

## Risk Mitigation

### Risk 1: Type dispatch complexity
**Mitigation**: Keep type switch simple, add types incrementally, use exhaustive testing

### Risk 2: Semantic encoding errors
**Mitigation**:
- Document truth tables clearly
- Start with simple cases
- Add test cases incrementally
- Verify against original tests

### Risk 3: Generic type inference issues
**Mitigation**: Use enum-based dispatch like in order_test.go, avoid reflection pointer comparison

### Risk 4: Breaking existing coverage
**Mitigation**:
- Run tests after each phase
- Keep original test functions until refactoring complete
- Use git to track changes incrementally

## Expected Outcome

### Before
```
equal_test.go: 1,423 lines
- 18 test functions
- Significant duplication across pairs (EqualT/NotEqualT, Same/NotSame, etc.)
- Test data scattered across multiple functions
```

### After
```
equal_test.go: ~1,178-1,238 lines (13-17% reduction)
- 5-6 main test functions (TestUnary, TestEquality, TestPointerIdentity, TestValueEquality, TestElementsMatch, TestErrorMessages)
- Zero duplication in high-priority areas
- Single source of truth for test cases
- Algorithmic expected results based on semantic properties
- Consistent testing of generic and reflection variants
```

### Qualitative Improvements
- Easier to add new test cases (append to iterator)
- Easier to add new assertions (add to enum + semantics function)
- Self-documenting through semantic categories
- Parallel test execution
- Consistent test structure across all equality assertions

## Next Steps After Plan Approval

1. Begin Phase 1 implementation
2. Verify Phase 1 passes all tests
3. Proceed to Phase 3 (highest impact)
4. Continue through remaining phases
5. Final verification and cleanup
6. Update documentation if needed

## Lessons from order_test.go

### What Worked Well
✅ Unified test matrix pattern
✅ Enum-based dispatch for generics
✅ Algorithmic expected results
✅ Single source of test data
✅ Comprehensive type coverage

### What to Replicate
✅ Start with clear semantic foundation
✅ Define categories before implementation
✅ Build incrementally, verify frequently
✅ Use const enums, not reflection hacks
✅ Document truth tables explicitly

### What to Improve
✅ Plan type dispatch upfront
✅ Consider reflectionOnly cases early
✅ Keep error message tests separate from functional tests
✅ Add comprehensive type coverage from the start
