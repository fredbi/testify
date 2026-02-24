# Full Stock Analysis: stretchr/testify Open Issues and PRs

**Date:** 2026-02-06
**Source:** https://github.com/stretchr/testify
**Purpose:** Validate decision to exclude mock/suite from fork; discover overlooked topics

---

## Executive Summary

This analysis examines **all 373 open items** (243 issues + 130 PRs) in the upstream stretchr/testify repository to:
1. Quantify the maintenance burden from mock and suite packages
2. Identify patterns in the remaining issues
3. Discover any overlooked topics that might be relevant to our fork

### Key Findings

| Category | Issues | PRs | Total | % of All |
|----------|--------|-----|-------|----------|
| **Mock-related** | 52 | 27 | 79 | **21.2%** |
| **Suite-related** | 23 | 20 | 43 | **11.5%** |
| **Mock + Suite combined** | 75 | 47 | 122 | **32.7%** |
| **Other (assert/require)** | 168 | 83 | 251 | **67.3%** |

**Conclusion:** Mock and suite together account for ~1/3 of the open issue/PR backlog. This validates the decision to exclude them from our fork, but also shows that 2/3 of the work is in assert/require—which we do support.

---

## Detailed Breakdown

### Mock Package (79 items = 21.2%)

**52 Open Issues:**
- Data races (#1597)
- API complexity (#1578 cleanup argument matching)
- Panic conditions (#1599, #1179, #1209)
- Missing features: spy (#1811), AnyTimes (#1372), variadic support (#1269, #1006, #1005)
- Concurrency issues (#1128, #964)
- Expectation ordering (#718, #1542)
- Documentation and error messages (#1383, #1144)

**27 Open PRs:**
- Race condition fixes (#1693, #1598)
- New features: OnF (#1814), mock builder (#1806), AtLeast (#1648)
- Panic fixes (#1799)
- API improvements (#1577, #1572, #1336)

**Assessment:** High complexity, significant maintenance burden. Mock frameworks are notoriously difficult to get right in Go due to the language's type system. External alternatives like gomock, mockery, or moq are better suited.

---

### Suite Package (43 items = 11.5%)

**23 Open Issues:**
- Parallel test issues (#187, #1139)
- Setup/Teardown lifecycle bugs (#1781, #1722, #1123)
- Missing hooks: BeforeSubTest/AfterSubTest (#1363)
- Stats and reporting (#761, #244)
- Panic handling (#771, #849)

**20 Open PRs:**
- SyncSuite/SyncTest for synctest compatibility (#1837, #1834)
- Filtering and control (#1749, #1750, #1051)
- Bug fixes (#1802, #1723, #1244)
- Feature extensions (#1142 fuzz support, #495 benchmark suite)

**Assessment:** Suite is a wrapper around `testing.T` that adds setup/teardown. Modern Go testing patterns (table-driven tests with `t.Run`, `t.Cleanup`) have reduced the need for test suites. The complexity of lifecycle management and parallel test support creates significant edge cases.

---

### Assert/Require Package (251 items = 67.3%)

This is the core functionality we support. Breaking down by subcategory:

#### Eventually/Never/Polling (15 items)
| # | Title | Status |
|---|-------|--------|
| 1843 | require.CollectT docs missing | ⛔ N/A (our docs correct) |
| 1833 | EventuallyWithTf docs wrong | ⛔ N/A (our docs correct) |
| 1810 | Eventually hangs on failed condition | 🔍 Check our impl |
| 1774 | NeverWithT for assertions in Never | 🔍 Consider |
| 1654 | Never passes if condition doesn't return | ✅ Fixed in our context rewrite |
| 1652 | Eventually can fail without running | ✅ Fixed in our context rewrite |
| 1611 | Eventually leaks goroutine | ✅ Fixed (pollCondition) |
| 1396 | require with EventuallyWithT | ✅ Handled differently |

**Our status:** Most fixed by our context-based pollCondition rewrite.

#### Equal/Comparison (50+ issues, 9 PRs)

Common themes:
- **Custom comparers** (#1616, #1204, #1184) - Request for user-defined equality
- **Time.Time handling** (#1388, #1078, #984) - Already fixed in our spew
- **Panic safety** (#1699, #1813) - Already addressed
- **Diff improvements** (#1628, #1479, #1325) - Partially addressed with colors
- **Pointer comparison** (#1403 atomic.Pointer, #1076)
- **Float comparison** (#1576)

**Potential additions:**
- #1184: Generic equality assertion - We have this via generics
- #1616: Custom equality interface - Interesting but complex

#### Collection/Contains (21 issues, 10 PRs)

Common themes:
- **Map support** (#1173 MapSubset, #704) - Worth considering
- **Slice ordering** (#806, #759) - Already have ElementsMatch
- **Error messages** (#1376, #1263, #561)
- **Large object handling** (#1801)

**Potential additions:**
- #1173: assert.MapSubset - Natural extension
- #1027: sync.Map support in Contains

#### Error Assertions (12 issues, 8 PRs)

- #1350: PanicsWithErrorIs - Worth considering
- #1304: PanicsWithErrorRegex - Worth considering
- #1115: Improve ErrorAs messages

#### Feature Proposals (interesting)

| # | Title | Relevance |
|---|-------|-----------|
| 1087 | Consistently (opposite of Eventually) | 🔍 Monitor |
| 1601 | NoFieldIsZero | 🔍 Monitor |
| 1332 | Cap/Capf (slice capacity) | Low priority |
| 1350 | PanicsWithErrorIs | Worth considering |
| 1446 | Assertions.Run wrapper | Low priority |
| 1111 | NonPositive/NonNegative | Trivial to add |
| 673 | NotJSONEq | Already have via JSONEq negation |
| 617 | NotFileExists/NotDirExists | Already have |

#### Color/Output/Formatting (44 issues, 19 PRs)

Major theme! Many requests for:
- Colorized diffs (#1479, etc.) - ✅ We have enable/colors
- Better diff output (#1628, #1325)
- Message formatting (#1355, #1034)

**Our status:** Largely addressed with enable/colors module.

#### Dependencies (7 issues, 5 PRs)

- #1826: go-spew vendoring - ✅ We internalized it
- #1752: Do we need objx? - ✅ We removed it
- #1724: YAML package migration - ✅ enable/yaml pattern

**Our status:** Fully addressed via internalization strategy.

---

## Undiscovered Topics Worth Considering

### 1. MapSubset Assertion (#1173, #704)
```go
// Proposed
assert.MapSubset(t, fullMap, expectedSubset)
```
We have Subset for slices but not maps. Natural extension.

### 2. PanicsWithErrorIs (#1350)
```go
// Check panic wraps specific error
assert.PanicsWithErrorIs(t, targetErr, func() { ... })
```
Combines PanicsWithError and ErrorIs semantics.

### 3. Consistently Assertion (#1087)
```go
// Assert condition remains true for duration
assert.Consistently(t, condition, 5*time.Second, 100*time.Millisecond)
```
Opposite of Eventually - useful for testing stable states.

### 4. Custom Equality Interface (#1616, #1204)
Allow users to define custom comparison logic for specific types. Complex but powerful.

### 5. Cap Assertion (#1332)
```go
assert.Cap(t, slice, expectedCap)
```
Trivial to implement alongside Len.

### 6. NonPositive/NonNegative (#1111)
```go
assert.NonPositive(t, -5)  // passes for 0 and negative
assert.NonNegative(t, 0)   // passes for 0 and positive
```
Trivial to add.

---

## Recommendations

### Validated Decisions

1. **Excluding mock package:** ✅ VALIDATED
   - 21% of issues, high complexity
   - Data races, API design issues, panic conditions
   - Better alternatives exist (gomock, mockery)

2. **Excluding suite package:** ✅ VALIDATED
   - 12% of issues, lifecycle complexity
   - Modern Go patterns reduce need
   - Parallel test issues are hard to solve

3. **Internalizing spew:** ✅ VALIDATED
   - Multiple issues about vendoring, panics
   - We control fixes directly

4. **Optional YAML:** ✅ VALIDATED
   - Single issue about package migration
   - enable/yaml pattern works well

### Potential Additions

| Feature | Recommendation |
|---------|----------------|
| Consistently (#1087) | 🔍 Monitor - Only feature worth considering |

### Not Adding (Chrome)

The following trivial proposals do not add substantial value. Upstream maintainers have expressed regret about accumulated "chrome" - small features that bloat the API without solving real problems:

| Feature | Reason Not Adding |
|---------|-------------------|
| MapSubset | Users can iterate; doesn't solve real problem |
| PanicsWithErrorIs | Niche; users can compose existing assertions |
| Cap assertion | Trivial; users can write `assert.Equal(t, cap(s), n)` |
| NonPositive/NonNegative | Trivial; covered by `assert.LessOrEqual(t, x, 0)` |
| Custom equality interface | Too complex, reflection-heavy |
| sync.Map support | Edge case, users can convert |
| Generic matchers (#729) | Over-engineering for Go |

**Philosophy:** Keep the API lean. Every addition has ongoing maintenance cost. If a feature can be trivially expressed with existing primitives, it doesn't belong in the library.

---

## Statistics Summary

```
Total Open Items: 373 (243 issues + 130 PRs)

By Package:
  Mock:     79 items (21.2%)
  Suite:    43 items (11.5%)
  Other:   251 items (67.3%)

Already Tracked by Us: ~30 items
New Items Discovered: 341 items
  - Relevant to our fork: ~200 (assert/require core)
  - Mock/Suite (excluded): 122
  - Misc/CI/Deps: ~19
```

---

**Conclusion:** The decision to exclude mock and suite is strongly validated by the data. These packages account for 1/3 of the maintenance burden and have complex, hard-to-solve issues. Our fork's focus on assert/require with zero dependencies is the right strategy.

The remaining assert/require issues are mostly about:
1. Output formatting (addressed with colors)
2. Edge cases we've already fixed (spew, time.Time, panics)
3. Feature requests that are nice-to-have but not critical

---

**Generated:** 2026-02-06
**Analysis by:** Claude Code
