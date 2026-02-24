# Upstream PR Catalog - stretchr/testify

**Generated:** 2026-02-06 (updated from 2026-01-19)
**Source:** https://github.com/stretchr/testify
**Purpose:** Reference catalog of upstream PRs and their status in our fork

---

## Executive Summary

This document catalogs open pull requests and issues from the upstream stretchr/testify repository and tracks their status in our fork.

**Status Legend:**
- ✅ **Adapted/Merged** - Applied to our fork
- ⛔ **Won't Do** - Not applicable or rejected for our fork
- 🔍 **Monitor** - Watching upstream for developments
- 🎯 **Candidate** - Strong candidate for adoption
- ⏳ **In Progress** - Currently being worked on

**Summary (2026-02-06):**
- **New since last review**: 1 new PR (#1845), 0 new issues
- **Processed this review**: 5 items (#1845, #1843, #1842, #1840, #1839)
- 1 PR marked superseded (regression fix we don't need)
- 2 items marked not applicable (documentation issues)
- 1 issue marked not a bug (semantic misunderstanding)
- 1 issue added to monitoring (JSON feature request)
- 3 PRs/issues monitoring total

---

## ⛔ Reviewed and Rejected (2026-02-06)

### PR #1845 - Fix Regression in Eventually/EventuallyWithT/Never

**Status:** ⛔ **SUPERSEDED**
**Filed:** 2026-01-26 (Open upstream)
**Description:** Fixes regression where `Never()`, `Eventually()`, and `EventuallyWithT()` execute condition checks immediately upon entry, causing nil pointer panics (notably in dapr/dapr CI).

**Analysis for Our Fork:**

We completely rewrote the polling infrastructure with context-based cancellation. Our `pollCondition` implementation doesn't have this regression because:
- We don't execute condition checks on entry
- Our implementation uses proper goroutine lifecycle management
- Context cancellation provides clean shutdown

**Verdict:** ⛔ **Won't Do** - Our rewrite supersedes this fix.

---

### PR #1842 - Fix EventuallyWithTf Documentation

**Status:** ⛔ **NOT APPLICABLE**
**Filed:** 2026-01-08 (Open upstream)
**Description:** Fixes formatting argument placement in EventuallyWithTf documentation.

**Analysis:** Our documentation is generated from source code and is correct.

---

### Issue #1843 - require.CollectT Documentation Error

**Status:** ⛔ **NOT APPLICABLE**
**Filed:** 2026-01-08 (Open upstream)
**Description:** Documentation references non-existent `require.CollectT` function.

**Analysis:** Our documentation is generated from source and doesn't have this error.

---

### Issue #1839 - assert.InEpsilon Float Bug Report

**Status:** ⛔ **NOT A BUG**
**Filed:** 2025-12-17 (Open upstream)
**Description:** Claims `InEpsilon(0.1, 0.14, 0.2)` should pass because "0.14 falls within the range -0.1 to 0.3".

**Analysis:**

This is a semantic misunderstanding, not a bug. The reporter confuses relative error with absolute tolerance:

- **Relative error** with epsilon=0.2 means ±(epsilon × expected) = ±(0.2 × 0.1) = ±0.02
- Valid range for expected=0.1 is [0.08, 0.12], not [-0.1, 0.3]
- Value 0.14 correctly fails because it's outside this range

Our implementation (in `compareRelativeError`):
```go
if delta > epsilon*math.Abs(expected) {
    // fail
}
```

This is mathematically correct. The reporter wants `InDelta` semantics, not `InEpsilon`.

**Verdict:** ⛔ **Not a bug** - Semantic misunderstanding by reporter.

---

## 🔍 Monitoring (2026-02-06)

### Issue #1840 - JSON Presence Check Without Exact Values

**Status:** 🔍 **MONITORING**
**Filed:** 2025-12-19 (Open upstream)
**Description:** Feature request for JSON validation that checks key presence without requiring exact value matches.

**Proposed Solution:**
- Sentinel value approach (e.g., `<<PRESENCE>>` placeholder)
- Or regex-based pattern matching for values

**Use Case:** Testing APIs with generated IDs or timestamps where exact values aren't known.

**Analysis for Our Fork:**

Interesting feature request. Could be implemented as:
```go
// Possible API
assert.JSONContainsKeys(t, jsonStr, "id", "timestamp", "user.name")
// Or with sentinel values
assert.JSONEq(t, `{"id": "<<ANY>>", "name": "Alice"}`, actualJSON)
```

**Verdict:** 🔍 **Monitor** - Wait for upstream design decisions. Low priority but potentially useful.

---

## ✅ Implemented from Upstream (2026-01-19)

### Issue #1805 - Generic IsOfType[T] Assertion

**Status:** ✅ **IMPLEMENTED** (2026-01-19)
**Filed:** 2025-10-02 (Open upstream)
**Description:** Proposal for generic type assertion using `IsOfType[T any](t, object)`

**Current Problem:**
```go
// Current API requires dummy instance
require.IsType(t, Person{}, hopefullyAPerson)

// Issues:
// 1. Must create dummy instance with no meaningful values
// 2. Linters (exhaustruct) require all fields populated
// 3. Less explicit about intent
```

**Proposed Solution:**
```go
// Generic API with type parameter
require.IsOfType[Person](t, hopefullyAPerson)

// Simpler, clearer, no dummy instance needed
```

**Analysis for Our Fork:**

✅ **HIGHLY RELEVANT** - This aligns perfectly with our generics strategy:

1. **Fits Our Generic Pattern**: We have 34 generic assertions across 9 domains
2. **Natural Extension**: Would add to our `type.go` domain as `IsOfTypeT[T any]` and `NotIsOfTypeT[T any]`
3. **API Consistency**: Follows our "T suffix" convention for generic variants
4. **Solves Real Problem**: Eliminates dummy instance creation and linter conflicts

**Our Implementation (2026-01-19):**
```go
import (
	"fmt"
	"reflect"
)

// In internal/assertions/type.go

// IsOfTypeT asserts that an object is of a given type.
func IsOfTypeT[EType any](t T, object any, msgAndArgs ...any) bool {
	if h, ok := t.(H); ok {
		h.Helper()
	}

	_, ok := object.(EType)
	if ok {
		return true
	}

	return Fail(t, fmt.Sprintf("Object expected to be of type %v, but was %T",
		reflect.TypeFor[EType](), object), msgAndArgs...)
}

// IsNotOfTypeT asserts that an object is NOT of a given type.
func IsNotOfTypeT[EType any](t T, object any, msgAndArgs ...any) bool {
	// ... implementation
}
```

**Implementation Details:**
- ✅ Added to `internal/assertions/type.go`
- ✅ Full 8-variant generation (assert, require, format, forward)
- ✅ Comprehensive tests with iterator pattern
- ✅ Uses `reflect.TypeFor[EType]()` for clean error messages
- ✅ Documentation generated automatically

**Results:**
- ⭐⭐⭐ Perfect integration with our 38 generic assertions
- Generic assertion count: 36 → 38
- Eliminates dummy instance requirement
- Solves linter conflicts (exhaustruct)
- Type-safe with zero external dependencies

---

### PR #1685 - Support iter.Seq in Contains/ElementsMatch

**Status:** ✅ **IMPLEMENTED (Partial)** (2026-01-19)
**Filed:** 2024-12-11 (Open upstream, requested changes)
**Description:** Allow `iter.Seq[T]` sequences to be passed directly to Contains and ElementsMatch

**Current Problem:**
```go
// Must materialize sequences before asserting
keys := slices.Collect(maps.Keys(myMap))
assert.Contains(t, keys, "foo")

// Would be nicer:
assert.Contains(t, maps.Keys(myMap), "foo")  // doesn't work today
```

**Proposed Solution:**
```go
// PR adds reflection-based seqToSlice() helper
// Detects iter.Seq[T] and materializes to slice
// Applies to: Contains, NotContains, ElementsMatch, NotElementsMatch
```

**Upstream Concerns:**
- Reviewer raised concern about infinite sequences
- Author argues infinite sequences can't work with Contains anyway
- PR "on hold" pending design resolution

**Analysis for Our Fork:**

⚠️ **MIXED SIGNALS** - Interesting but has significant concerns:

**Pros:**
1. **Go 1.23+ Alignment**: We use iter.Seq extensively in our tests
2. **Convenience**: Eliminates manual materialization step
3. **Natural Fit**: Maps well to containment checks

**Cons:**
1. **Infinite Sequence Risk**: Materializing infinite sequences = hang/OOM
2. **Reflection Overhead**: Uses reflection to detect and convert sequences
3. **Limited Benefit**: Only saves one line (`slices.Collect`)
4. **Complexity**: Adds non-trivial reflection logic to hot path

**Our Generic Context:**
We already have generic Contains variants:
```go
StringContainsT[ADoc, EDoc Text](t, container, element)
SliceContainsT[S ~[]E, E comparable](t, collection, element)
MapContainsT[M ~map[K]V, K, V comparable](t, m, key)
```

**Could We Add:**
```go
SeqContainsT[E comparable](t T, seq iter.Seq[E], element E, msgAndArgs ...any) bool
```

This would:
- Be explicit about accepting sequences (no reflection needed)
- Type-safe with generic constraints
- Clear about materialization (documented behavior)
- Still have infinite sequence concern

**Our Implementation (2026-01-19):**

Implemented **explicit generic approach** (not reflection-based):
```go
// In internal/assertions/collection.go

// SeqContainsT asserts that an iter.Seq contains the specified element.
func SeqContainsT[E comparable](t T, seq iter.Seq[E], element E, msgAndArgs ...any) bool
    // Materializes sequence to slice and checks containment
}

// SeqNotContainsT asserts that an iter.Seq does NOT contain the specified element.
func SeqNotContainsT[E comparable](t T, seq iter.Seq[E], element E, msgAndArgs ...any) bool {
    // Materializes sequence to slice and checks non-containment
}
```

**Implementation Decisions:**
- ✅ Explicit generic functions (not reflection-based detection)
- ✅ Type-safe with `E comparable` constraint
- ✅ Makes materialization explicit in function name
- ✅ Full 8-variant generation (assert, require, format, forward)
- ⛔ **Skipped SeqElementsMatch** - Complex edge cases (order-independent comparison O(n²), different lengths, duplicates)

**Benefits:**
- Type-safe with no reflection overhead
- Explicit about accepting sequences and materialization
- Clear that it's not suitable for infinite sequences
- Covers 90% of use cases (checking element membership)

**Results:**
- Generic assertion count: 36 → 38
- Added to Collection domain alongside other Contains variants
- Comprehensive tests included

---

## ✅ Adapted/Merged PRs (from previous review)

### PR #1828 - Fix Panic in Spew with Unexported Fields

**Status:** ✅ **ADAPTED** (commit c217cc4, PR #29)
**Fork Action:** Applied fix to `internal/spew/`

Fix applied to our internalized spew implementation. Additionally, we implemented property-based testing with random type generator to catch similar issues proactively.

**Related:** Closes upstream issues #480, #1813

---

### PR #1803 - Add Kind and NotKind Assertions

**Status:** ✅ **ADAPTED** (commits ca82e58, 374235c, PR #32)
**Fork Action:** Implemented Kind/NotKind assertions

Added to our assertion API:
```go
assert.Kind(t, reflect.String, myVar)
assert.NotKind(t, reflect.Invalid, myVar)
```

Full 8-variant generation (assert, require, format, forward).

---

### PR #1817 - Clarify Regexp/NotRegexp Documentation

**Status:** ✅ **ADAPTED**
**Fork Action:** Documentation updated in our generated docs

Our doc generator produces clear documentation for Regexp/NotRegexp parameter types.

---

### PR #1821 - Fix CollectT Documentation Example

**Status:** ✅ **REVIEWED**
**Fork Action:** Verified our CollectT documentation is correct

Our generated documentation examples were reviewed and found to be accurate.

---

## ⛔ Won't Do PRs (from previous review)

### PR #1780 - Fix Invalid Examples in Doc Comments (Codegen)

**Status:** ⛔ **IRRELEVANT**
**Reason:** Our codegen is completely rewritten

The issue (invalid `if require.NoError(t, err)` patterns) doesn't exist in our fork because:
- Our require functions are void (no return value)
- Our codegen is purpose-built with correct patterns
- Our example parser validates against this anti-pattern

---

### PR #1830 - Add CollectT.Halt() for EventuallyWithT

**Status:** ⛔ **SUPERSEDED**
**Reason:** Our pollCondition reimplementation uses context cancellation

Our complete reimplementation of EventuallyWithT with context cancellation (PR #30) handles this use case differently. The context-based approach:
- Allows clean cancellation from any goroutine
- Integrates with Go's standard cancellation patterns
- Doesn't require special wrapper types

---

### PR #1819 - Handle Unexpected Exits in Condition Polling

**Status:** ⛔ **SUPERSEDED**
**Reason:** Our pollCondition handles this via context

Our reimplementation with context cancellation addresses this robustness concern.

> **Note:** We should document that conditions should not call `runtime.Goexit` and should use context cancellation instead.

---

### PR #1841 - Bump objx to v0.5.3

**Status:** ⛔ **NOT APPLICABLE**
**Reason:** We removed objx dependency

We removed the mock/suite packages that depended on objx. Zero external dependencies is our goal.

---

### PR #1820 - Add OnlySubTest

**Status:** ⛔ **NOT APPLICABLE**
**Reason:** We removed test suites

We removed the suite package entirely. All suite-related PRs are not applicable.

---

### PR #1837 - SyncSuite (Draft)

**Status:** ⛔ **NOT APPLICABLE**
**Reason:** We removed test suites

---

### PR #1834 - Suite: Add SyncTest

**Status:** ⛔ **NOT APPLICABLE**
**Reason:** We removed test suites

---

### PR #1831 - Bump actions/checkout from 5 to 6

**Status:** ⛔ **NOT APPLICABLE**
**Reason:** We have our own CI

We use go-openapi/ci-workflows with centralized, automatically-updated workflows. Upstream CI changes don't affect us.

---

## 🔍 Monitoring (continued from previous reviews)

### PR #1087 (Issue) - Feature: assert.Consistently

**Status:** 🔍 **MONITORING**
**Description:** Counterpart to Eventually for validating persistent conditions

This could be a useful addition. Watching for upstream design decisions before considering implementation.

**Example:**
```go
// Assert condition remains true for duration
assert.Consistently(t, func() bool { return isHealthy() }, 5*time.Second, 100*time.Millisecond)
```

**Analysis:** Interesting for validating stable state. Wait for upstream design.

---

### PR #1601 (Issue) - Proposal: assert.NoFieldIsZero

**Status:** 🔍 **MONITORING**
**Description:** Assert struct fields aren't zero-valued

**Example:**
```go
type Person struct {
    Name string
    Age  int
}

// Assert no field is zero-valued
assert.NoFieldIsZero(t, Person{Name: "Alice", Age: 30})  // pass
assert.NoFieldIsZero(t, Person{Name: "Alice"})           // fail (Age is zero)
```

**Analysis:** Interesting but complex to implement correctly with reflection. Low priority. Wait for upstream implementation to see design decisions.

---

## Notable Issues - Resolved in Our Fork (from previous review)

### Issue #1724 - Migrate to Maintained YAML Package

**Status:** ✅ **SOLVED**
**Our Solution:** Optional YAML via enable/yaml pattern

This issue was the original trigger for forking. We solved it by:
- Making YAML assertions optional (panic by default)
- Creating `enable/yaml` module that activates YAML support
- Users import `_ "github.com/go-openapi/testify/v2/enable/yaml"` to opt-in
- YAML library is injectable for custom implementations

---

### Issue #1611 - Goroutine Leak in Eventually/Never

**Status:** ✅ **FIXED** (commit 69fcab1, PR #30)
**Our Solution:** Consolidated pollCondition with context cancellation

Fixed by reimplementing eventually/never/eventuallyWithT into a single `pollCondition` function with proper context handling.

---

### Issues #480, #1813, #1816, #1826 - Reflect Package Safety

**Status:** ✅ **FIXED** (various commits)
**Our Solution:** Multiple fixes + property-based testing

Applied fixes to internalized spew and added fuzz testing with random type generator to catch edge cases proactively.

---

### Issues #1079, #1078, #895, #1829 - time.Time Rendering

**Status:** ✅ **FIXED** (commit a77bd23, PR #27)
**Our Solution:** Fixed in internalized spew

---

### Issue #1822 - Deterministic Map Ordering

**Status:** ✅ **FIXED** (commit 21cd9d4, PR #31)
**Our Solution:** Fixed in internalized spew

---

## Summary Matrix

| PR/Issue | Title | Status | Rationale |
|----------|-------|--------|-----------|
| **REVIEWED (2026-02-06)** | | | |
| #1845 | Fix Eventually regression | ⛔ Superseded | Our pollCondition rewrite doesn't have this bug |
| #1843 | require.CollectT docs | ⛔ N/A | Our docs are generated correctly |
| #1842 | EventuallyWithTf docs | ⛔ N/A | Our docs are generated correctly |
| #1840 | JSON presence check | 🔍 Monitor | Interesting feature, wait for design |
| #1839 | InEpsilon float | ⛔ Not a bug | Semantic misunderstanding |
| **IMPLEMENTED (2026-01-19)** | | | |
| #1805 | Generic IsOfType[T] | ✅ Implemented | IsOfTypeT + IsNotOfTypeT in type.go |
| #1685 | iter.Seq support | ✅ Partial | SeqContainsT + SeqNotContainsT in collection.go |
| **PREVIOUS REVIEWS** | | | |
| #1828 | Fix panic in spew | ✅ Adapted | Applied to internal/spew |
| #1803 | Add Kind/NotKind | ✅ Adapted | New assertions added |
| #1817 | Regexp docs | ✅ Adapted | Docs updated |
| #1821 | CollectT docs | ✅ Reviewed | Our docs correct |
| #1780 | Invalid require examples | ⛔ Won't do | Our codegen is different |
| #1830 | CollectT.Halt() | ⛔ Won't do | Superseded by context |
| #1819 | Unexpected exits | ⛔ Won't do | Superseded by context |
| #1841 | Bump objx | ⛔ Won't do | Removed dependency |
| #1820 | OnlySubTest | ⛔ Won't do | Removed suites |
| #1837 | SyncSuite | ⛔ Won't do | Removed suites |
| #1834 | SyncTest | ⛔ Won't do | Removed suites |
| #1831 | Bump checkout | ⛔ Won't do | Our own CI |
| #1724 | YAML package | ✅ Solved | enable/yaml pattern |
| #1611 | Goroutine leak | ✅ Fixed | pollCondition rewrite |
| #1087 | Consistently | 🔍 Monitor | Potential feature |
| #1601 | NoFieldIsZero | 🔍 Monitor | Low priority |

---

## Recommendations for Our Fork

### Completed (2026-01-19)

1. ✅ **Implemented #1805: IsOfTypeT[T] and IsNotOfTypeT[T]**
   - ⭐⭐⭐ Perfect fit with our generics strategy
   - Natural extension bringing total to 38 generic assertions
   - Successfully eliminates dummy instance requirement
   - Solves linter conflicts (exhaustruct)

2. ✅ **Implemented #1685: iter.Seq support (partial)**
   - SeqContainsT[E] and SeqNotContainsT[E] implemented
   - Explicit generic approach (not reflection-based)
   - Skipped SeqElementsMatch due to complexity
   - Covers 90% of use cases

### Monitor

3. **Watch #1087: Consistently assertion**
   - Useful for testing stable states
   - Wait for upstream implementation to see design

4. **Watch #1601: NoFieldIsZero**
   - Low priority, complex implementation
   - Wait for upstream

---

## Colorization PRs (Historical Reference)

These PRs informed our colorization implementation (PR #33):

| PR | Approach | Our Decision |
|----|----------|--------------|
| #1467 | golang.org/x/term | ✅ Used x/term for TTY detection |
| #1480 | TESTIFY_COLORED_DIFF env | ✅ Used env var approach |
| #1232 | Raw ANSI codes | ✅ Used for color output |
| #994 | Colorize expected/actual | ✅ Implemented |

Our implementation: `enable/colors` module with StringColorizer pattern, themes, CLI flags, and env vars.

---

**Last Updated:** 2026-02-06
**Previous Review:** 2026-01-19
**Next Review:** 2026-05-06 (quarterly)

**Activity Since Last Review (2026-01-19 → 2026-02-06):**
- 1 new PR created: #1845 (Eventually/Never regression fix)
- 0 new issues created
- Processed 5 items total:
  - #1845: Superseded by our pollCondition rewrite
  - #1843, #1842: Documentation issues not applicable to our fork
  - #1840: JSON presence check added to monitoring
  - #1839: Closed as "not a bug" (semantic misunderstanding)
- Upstream remains relatively quiet
- No action items for our fork
