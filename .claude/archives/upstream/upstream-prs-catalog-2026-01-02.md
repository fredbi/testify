# Upstream PR Catalog - stretchr/testify

**Generated:** 2026-01-02
**Source:** https://github.com/stretchr/testify
**Purpose:** Reference catalog of upstream PRs not yet mentioned in our project plan

---

## Executive Summary

This document catalogs **12 open pull requests** from the upstream stretchr/testify repository that may be relevant to our fork but are not yet referenced in `.claude/plans/testify-v2-project.md`.

**Priority Categories:**
- 🔥 **High Priority** (3 PRs): Critical fixes or features aligned with our goals
- ⚠️ **Medium Priority** (5 PRs): Useful improvements worth considering
- 💭 **Low Priority** (4 PRs): Interesting but less relevant to our fork

---

## 🔥 High Priority PRs

### PR #1828 - Fix Panic in Spew with Unexported Fields

**Status:** Open, ready for review
**Author:** ccoVeille
**Relevance:** ⭐⭐⭐ **CRITICAL - We have internalized spew!**

**Problem:**
Panic occurs when spew attempts to sort unexported struct fields during assertion failures.

**Solution:**
Implements safety check using `CanInterface()` before calling `Interface()` on reflected values.

**Why it matters for us:**
- We have internalized `go-spew` in `internal/spew/`
- This fix prevents panics in our assertion error reporting
- Closes issues #480 and #1813

**Recommendation:**
**MUST ADAPT** - Apply this fix to our internalized spew implementation immediately. This is a safety fix that prevents crashes.

**Related Plan Items:**
- Section 2.5: "Enhanced diff output" - leveraging internalized dependencies
- Listed issue #1813 already noted in plan as needing investigation

---

### PR #1780 - Fix Invalid Examples in Doc Comments (Codegen)

**Status:** Open, requested changes (but improvements made)
**Author:** PCloud63514
**Relevance:** ⭐⭐⭐ **CRITICAL - Directly related to our codegen!**

**Problem:**
The `require` package had invalid documentation examples showing patterns like:
```go
if require.NoError(t, err) { ... }  // WRONG - require doesn't return bool
```

**Solution:**
- Corrects invalid examples throughout codebase
- Introduces `requireCommentParseIf` helper function for bulk transformation
- Includes unit tests for the helper

**Why it matters for us:**
- We generate extensive documentation from examples in doc comments
- Our codegen relies on correct example patterns
- We should verify our generated require examples don't have this issue

**Recommendation:**
**REVIEW AND ADAPT** - Check our generated require examples for this pattern. The helper function approach might be useful for our templates.

**Related Plan Items:**
- Section 3.1: "Generated documentation organized in domains"
- Section 3.4: "Examples, tutorials" - our example-driven test generation
- Our `codegen/internal/scanner/comments-parser/examples.go` parser

**Action Items:**
1. Review our generated `require/` examples for conditional patterns
2. Add validation to our example parser to detect this anti-pattern
3. Consider adapting the `requireCommentParseIf` approach for our templates

---

### PR #1803 - Add Kind and NotKind Assertions

**Status:** Open, awaiting final review
**Author:** segogoreng
**Relevance:** ⭐⭐⭐ **New feature - aligns with our goals**

**Problem:**
No built-in way to assert the reflection kind of an object (addresses issue #633).

**Solution:**
Adds two new assertion functions:
```go
assert.Kind(t, reflect.String, myVar)      // Assert kind matches
assert.NotKind(t, reflect.Invalid, myVar)  // Assert kind doesn't match
```

**Features:**
- Validates against `reflect.Invalid` kinds
- Handles `nil` objects properly
- Comprehensive test coverage
- Both assert and require variants

**Why it matters for us:**
- Expands our assertion API with useful reflection helpers
- Type safety focus aligns with our plan section 1.3
- Clean implementation we could adapt

**Recommendation:**
**CONSIDER ADOPTION** - This is a well-designed feature that fits our API. Worth reviewing for inclusion.

**Related Plan Items:**
- Section 1.3: "Type safety & other critical fixes needed"
- Section 1.8: "New features"

**Decision Needed:**
Should we add Kind/NotKind assertions to our fork? They seem useful and well-implemented.

---

## ⚠️ Medium Priority PRs

### PR #1830 - Add CollectT.Halt() for EventuallyWithT

**Status:** Open
**Author:** twz123
**Relevance:** ⭐⭐ **Useful improvement to Eventually pattern**

**Problem:**
Fatal assertions within `EventuallyWithT` block the retry loop until timeout instead of failing fast.

**Solution:**
Introduces `HaltT` wrapper for `CollectT` that signals early exit:
```go
condition := func(c *CollectT) {
    require.True(c.Halt(), socketsOpen(), "must be open")
    assert.True(c, eventuallyTrue(), "non-fatal checks")
}
assert.EventuallyWithT(t, condition, time.Second, 10*time.Millisecond)
```

**Why it matters for us:**
- Improves usability of Eventually assertions
- Fail-fast behavior is more developer-friendly
- Small, focused change

**Recommendation:**
**MONITOR** - Wait to see if upstream merges. If we get user requests for this, consider adopting.

**Related Plan Items:**
- Section 1.8: "New features" - evaluating useful additions

---

### PR #1819 - Handle Unexpected Exits in Condition Polling

**Status:** Open, ready for review (7 commits)
**Author:** ubunatic
**Relevance:** ⭐⭐ **Robustness improvement**

**Problem:**
`Eventually`, `Never`, and `EventuallyWithT` could hang or give false positives when condition goroutines exit via `runtime.Goexit`.

**Solution:**
Detects unexpected goroutine exits and fails immediately with "Condition exited unexpectedly" error.

**Changes:**
- `Eventually`: Fails early instead of hanging until timeout
- `EventuallyWithT`: Returns false and propagates failures correctly
- `Never`: Fails early to prevent incorrect results

**Why it matters for us:**
- Makes Eventually/Never more robust
- Prevents confusing test hangs
- Addresses issues #1810 and related

**Recommendation:**
**REVIEW** - This is a good robustness fix. Consider adopting if we use Eventually/Never assertions.

**Related Plan Items:**
- Section 1.3: "Type safety & other critical fixes"
- Section 4.1: "Code quality" - robustness improvements

---

### PR #1817 - Clarify Regexp/NotRegexp Documentation

**Status:** Open
**Author:** kdt523
**Relevance:** ⭐⭐ **Documentation improvement**

**Problem:**
Users confused about which types the `rx` parameter accepts in Regexp/NotRegexp assertions.

**Solution:**
Updates documentation to clarify:
- Preferred type: `*regexp.Regexp`
- Backward compatibility: Other values are stringified and compiled via `regexp.MustCompile`
- Adds proper GoDoc links

**Why it matters for us:**
- Improves API documentation clarity
- No code changes, just better docs
- We have these same assertions

**Recommendation:**
**ADOPT** - Low-risk documentation improvement. Update our generated docs similarly.

**Related Plan Items:**
- Section 3.1: "Generated documentation organized in domains"
- Section 4.3: "Documentation" - polish existing godoc

**Action Items:**
1. Review our Regexp/NotRegexp documentation
2. Add similar clarification to our generated docs
3. Consider if our doc generator templates need updates

---

### PR #1821 - Fix CollectT Documentation Example

**Status:** Open
**Author:** a2not
**Relevance:** ⭐⭐ **Documentation fix**

**Problem:**
Incorrect example usage of `assert.CollectT` in require package documentation.

**Solution:**
Corrects the doc comment example to show proper CollectT usage.

**Why it matters for us:**
- We generate documentation examples extensively
- Accuracy in examples is critical
- Small, targeted fix

**Recommendation:**
**REVIEW** - Check if our CollectT documentation has similar issues.

**Related Plan Items:**
- Section 3.4: "Examples, tutorials" - accuracy in examples

---

### PR #1841 - Bump objx to v0.5.3

**Status:** Open
**Author:** WhyNotHugo
**Relevance:** ⭐ **Not applicable - we don't use objx**

**Problem:**
Updates objx dependency to latest version.

**Why it doesn't matter for us:**
- We removed mocking functionality (which used objx)
- Zero external dependencies is our goal
- Not relevant to our fork

**Recommendation:**
**IGNORE** - We don't use objx.

---

## 💭 Low Priority PRs

### PR #1820 - Add OnlySubTest

**Status:** Open
**Author:** kelvinfloresta
**Relevance:** ⭐ **Suite-related - we removed suites**

**Problem:**
Running full test suite is slow when debugging specific subtests.

**Solution:**
Adds `OnlySubTest` method to run isolated subtests efficiently.

**Why it doesn't matter for us:**
- We removed the suite package
- Feature is suite-specific
- Not applicable to our minimal approach

**Recommendation:**
**IGNORE** - We don't support suites.

---

### PR #1837 - Add SyncSuite (Draft)

**Status:** Draft
**Author:** Nocccer
**Relevance:** ⭐ **Suite-related - we removed suites**

**Why it doesn't matter for us:**
- Suite functionality removed from our fork
- Still in draft status
- Not aligned with our minimal dependencies approach

**Recommendation:**
**IGNORE** - We don't support suites.

---

### PR #1834 - Suite: Add SyncTest

**Status:** Open (10 comments)
**Author:** ikonst
**Relevance:** ⭐ **Suite-related - we removed suites**

**Why it doesn't matter for us:**
- Suite-specific feature
- We removed suite support

**Recommendation:**
**IGNORE** - We don't support suites.

---

### PR #1831 - Bump actions/checkout from 5 to 6

**Status:** Open
**Author:** dependabot
**Relevance:** ⭐ **CI infrastructure - not code**

**Why it doesn't matter for us:**
- GitHub Actions dependency bump
- We maintain our own CI configuration
- Can update independently

**Recommendation:**
**IGNORE** - Manage our own CI dependencies.

---

## Notable Open Issues

These issues were mentioned as highly discussed but don't have PRs yet:

### Issue #1724 - Migrate to Maintained YAML Package
**Relevance:** ⚠️ **We already handle this!**

**Issue:** gopkg.in/yaml.v3 is no longer actively maintained.

**Our Status:** We have made YAML optional via the enable/yaml pattern! This is a solved problem for us.

### Issue #1601 - Proposal: assert.NoFieldIsZero
**Relevance:** 💭 **Potential feature**

**Issue:** Add assertion to verify struct fields aren't zero-valued.

**Consideration:** Interesting feature, but complex to implement correctly with reflection.

### Issue #1087 - Feature: assert.Consistently
**Relevance:** 💭 **Potential feature**

**Issue:** Counterpart to Eventually for validating persistent conditions.

**Consideration:** Useful pattern. Could be valuable addition.

---

## Summary Matrix

| PR # | Title | Priority | Action | Rationale |
|------|-------|----------|--------|-----------|
| 1828 | Fix panic in spew | 🔥 High | MUST ADAPT | We have internalized spew |
| 1780 | Fix invalid require examples | 🔥 High | REVIEW & ADAPT | Codegen doc examples |
| 1803 | Add Kind/NotKind | 🔥 High | CONSIDER | New feature - type safety |
| 1830 | CollectT.Halt() | ⚠️ Medium | MONITOR | Useful Eventually improvement |
| 1819 | Handle unexpected exits | ⚠️ Medium | REVIEW | Robustness fix |
| 1817 | Regexp docs | ⚠️ Medium | ADOPT | Documentation improvement |
| 1821 | CollectT docs | ⚠️ Medium | REVIEW | Documentation fix |
| 1820 | OnlySubTest | 💭 Low | IGNORE | Suite-related |
| 1837 | SyncSuite | 💭 Low | IGNORE | Suite-related |
| 1834 | SyncTest | 💭 Low | IGNORE | Suite-related |
| 1831 | Bump checkout action | 💭 Low | IGNORE | CI dependency |
| 1841 | Bump objx | 💭 Low | IGNORE | We don't use objx |

---

## PRs Already in Project Plan

These are NOT included above since they're already tracked:

**Completed/Merged:**
- #1825 - Fix panic with EqualValues and uncomparable types ✅
- #1818 - Fix panic on invalid regex ✅
- #1223 - Display uint in decimal ✅
- #1513 - Add JSONEqByte ✅

**Under Investigation:**
- #1829 - Fix time.Time rendering in diffs 🔍
- #1822 - Deterministic map ordering in diffs 🔍
- #1816 - Don't panic on unexported struct key 🔍
- #1824, #1826, #1611, #1813 - Various issues to investigate 🔍

**Colorization PRs:**
- #1467, #1480, #1232, #994 - Colorized output variations 🔍

---

## Recommended Next Steps

1. **Immediate Action (This Week):**
   - ✅ Adapt PR #1828 fix to our `internal/spew/` implementation
   - ✅ Review PR #1780 and audit our generated require examples

2. **Short Term (This Month):**
   - Review PR #1803 (Kind/NotKind) for potential adoption
   - Check PR #1817 documentation improvements for our docs
   - Review PR #1819 Eventually/Never robustness fixes

3. **Ongoing Monitoring:**
   - Watch for PR #1830 merge status
   - Track upstream issue #1087 (Consistently assertion)
   - Monitor other documentation improvements

4. **Update Project Plan:**
   - Add PR #1828 to section 2.5 "Enhanced diff output"
   - Add PR #1780 to section 3.1 "Generated documentation"
   - Add PR #1803 to section 1.8 "New features" for evaluation
   - Note that issue #1724 (YAML package) is solved in our fork

---

## Future Reference Location

**Suggested permanent location:** `.claude/plans/ramblings/upstream-prs-catalog-2026-01-02.md`

This document should be updated periodically (e.g., quarterly) to track new upstream PRs.

---

**Last Updated:** 2026-01-02
**Next Review:** 2026-04-01 (quarterly)
