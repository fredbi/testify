# Upstream PR Catalog - stretchr/testify

**Generated:** 2026-01-12 (updated from 2026-01-02)
**Source:** https://github.com/stretchr/testify
**Purpose:** Reference catalog of upstream PRs and their status in our fork

---

## Executive Summary

This document catalogs open pull requests from the upstream stretchr/testify repository and tracks their status in our fork.

**Status Legend:**
- ✅ **Adapted/Merged** - Applied to our fork
- ⛔ **Won't Do** - Not applicable or rejected for our fork
- 🔍 **Monitor** - Watching upstream for developments
- ⏳ **In Progress** - Currently being worked on

**Summary (2026-01-12):**
- 4 PRs adapted and merged
- 6 PRs marked won't do (removed features, different CI, etc.)
- 2 PRs still monitoring

---

## ✅ Adapted/Merged PRs

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

## ⛔ Won't Do PRs

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

## 🔍 Monitoring

### PR #1087 (Issue) - Feature: assert.Consistently

**Status:** 🔍 **MONITORING**
**Description:** Counterpart to Eventually for validating persistent conditions

This could be a useful addition. Watching for upstream design decisions before considering implementation.

---

### PR #1601 (Issue) - Proposal: assert.NoFieldIsZero

**Status:** 🔍 **MONITORING**
**Description:** Assert struct fields aren't zero-valued

Interesting but complex to implement correctly with reflection. Low priority.

---

## Notable Issues - Resolved in Our Fork

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

**Last Updated:** 2026-01-12
**Next Review:** 2026-04-01 (quarterly)
