# Current Plan Review - 2025-12-31

## Overview

This document consolidates the current established plan for testify v2, integrating:
- Code generation architecture (completed)
- Documentation site improvements (in progress)
- Tool development for Claude (planned)
- Outstanding issues and enhancements

## Status Summary

**Major Milestones:**
- ✅ Code generation refactor COMPLETE (76 functions × 8 variants = 608 generated functions)
- ✅ Documentation site infrastructure COMPLETE (Hugo + generation pipeline)
- 🔄 Documentation content polish IN PROGRESS
- 📋 Go development tools for Claude PLANNED
- 📋 Markdown linting agent PLANNED

---

## Part 1: Documentation Generator Enhancements

### 1.1 Missing Coverage: Types, Variables, Constants, Helpers

**Issue:** The documentation generator currently only covers assertion functions. It doesn't handle:
- Types (e.g., `TestingT`, `Assertions`, `ValueAssertionFunc`)
- Variables (if any exported)
- Constants (none currently, but future-proof)
- Helper functions (non-assertion utilities like `ObjectsAreEqual`, `CallerInfo`)

**Location:** `codegen/internal/generator/domains/` package has "TODO" markers for this.

**Impact:**
- Incomplete API reference
- Users don't know about helper types/functions
- Documentation site looks unfinished

**Current Status:**
- ✅ Most types already have domain annotations
- ⚠️ **Specific issue:** Function types (e.g., `ValueAssertionFunc`, `PanicAssertionFunc`) need proper comment annotations
- ✅ Helper functions not covered yet

**Proposed Solution:**
1. Fix function type annotations in source comments:
   - Add domain tags to func type doc comments
   - Ensure consistent documentation format
2. ✅ Extend scanner to detect helper functions (non-assertion exports)
3. Update templates to render function types properly
4. ✅ Add "Helpers" domain for utility functions if needed

**Priority:** Medium (func types are the main gap)

**Status:** TODO - focus on func type annotations first

---

### 1.2 Testable Examples - ALREADY IMPLEMENTED ✅

**Status:** ✅ **COMPLETE** - Testable `Example*` functions are already generated!

**What Exists:**
- `Example*` test functions generated for assertions
- Appear in godoc automatically
- Tested via `go test`
- Runnable and verified

**Remaining Low-Priority Enhancement:**
- Optionally render testable example code in Hugo documentation site
- Would show the actual Go code from `Example*` functions in HTML docs
- **Issue:** This would duplicate/overlap with godoc
- **Decision:** Low priority - godoc already serves this purpose

**Priority:** Low (godoc already provides this)

**Status:** Core feature complete, HTML rendering optional

---

### 1.3 TODO Marker Issue in Example Values

**Problem:** Using `// TODO` as a marker for ignored/placeholder example values:
```go
// Examples:
//   success: &structValue{Field: "value"}, // TODO
//   failure: complexType{}, // TODO
```

This triggers false positives in:
- Code quality analyzers (linters looking for TODOs)
- Project management tools (TODO trackers)
- IDE warnings about unfinished work

**Impact:**
- Noise in linter reports
- Confusion about what's actually TODO vs what's a placeholder
- Professional polish issue

**Proposed Solution:**

Replace `// TODO` with more neutral marker:

**Option A: `// NOT IMPLEMENTED`**
```go
// Examples:
//   success: &structValue{Field: "value"}, // NOT IMPLEMENTED
//   failure: complexType{}, // NOT IMPLEMENTED
```

**Option B: `// PLACEHOLDER`**
```go
// Examples:
//   success: &structValue{Field: "value"}, // PLACEHOLDER
//   failure: complexType{}, // PLACEHOLDER
```

**Option C: `// SYNTHETIC`**
```go
// Examples:
//   success: &structValue{Field: "value"}, // SYNTHETIC
//   failure: complexType{}, // SYNTHETIC
```

**Recommendation:**
- **Option A: `// NOT IMPLEMENTED`** - clearest meaning
- Update all existing `// TODO` markers in `internal/assertions/*.go`
- Update generator to recognize `NOT IMPLEMENTED` as placeholder marker
- Document in CLAUDE.md and contributor guidelines

**Files to Update:**
```bash
# Find all occurrences
grep -r "// TODO" internal/assertions/*.go | grep "Examples:"
```

**Priority:** Low (cosmetic but professional)

**Status:** New issue raised today

---

## Part 2: Documentation Site Polish

### 2.1 Critical Fixes (From Assessment)

**Priority: HIGH - Must fix before public launch**

1. ✅ **Incomplete Attribution**
   - File: `docs/doc-site/api/_index.md`
   - Issue: "Generated with @" with nothing after it
   - Fix: Complete the attribution line or remove it

2. ✅ **Inconsistent Capitalization**
   - Files: All domain description files
   - Issue: Mixed title case and sentence case
   - Fix: Standardize to title case for all domain descriptions

3. ✅ **Empty Table Headers**
   - Files: All domain markdown files
   - Issue: `<th></th><th></th>` in variant tables
   - Fix: Add meaningful headers or remove `<thead>` entirely

4. ✅ **Bare Project Documentation**
   - File: `docs/doc-site/project/_index.md`
   - Issue: Just cards, no introduction
   - Fix: Add welcoming introduction and context

5. **Typo: "Os Files"**
   - File: Domain description
   - Issue: Should be "OS Files" (uppercase)
   - Fix: Capitalize properly

**Tracking:** See `docs-assessment.md` for full details

---

### 2.2 Missing Context (Medium Priority)

Add explanatory content:
1. What are "domains" in this context?
2. When to use `assert` vs `require`?
3. What are "formatted variants"?
4. What are "method variants" (forward pattern)?
5. Quick start guide
6. Installation instructions

**Proposed Location:** New pages in `docs/doc-site/guides/`

---

### 2.3 Hugo Site TODOs

**From `hack/doc-site/hugo/TODO.md`:**
1. ✅ Variabilize width in CSS (done: changed directly in theme CSS for now)
2. ✅ GoDoc reference badge left-aligned (done)

**From `notes/TODO.md`:**
1. ✅ {{ .Tool }}: without sha (done)
2. ✅ {{ .ToolHeader }}: sha, timestamp (done)
3. ✅ RefCount in domain description (for cards)
4. ✅ Add weight in frontmatter to put Common first or last
5. [ ] ~Cover types (see 1.1 above)~ won't do: the generated is complex (and complete) enough. No need to enrich it further with types.
6. ✅ Cover helpers (see 1.1 above)
7. ⏳ godoc link in notes is not processed

---

## Part 3: Tool Development for Claude

### 3.1 Go Development Tools

**Purpose:** Help Claude work more efficiently with Go code

**Planned Features (from discussion):**

**Tier 0: Navigation & Understanding (High Priority)**
1. Find References - All usages of function/type/variable
2. Go to Definition - Jump to symbol definition
3. List Symbols - Functions/types in file or package
4. Type Information - Get type of expression
5. Find Implementations - Types implementing an interface

**Tier 1: Quality & Refactoring (Medium Priority)**
6. Diagnostics - Real compile errors with line numbers
7. Organize Imports - Add missing, remove unused
8. Rename Symbol - Safe rename across codebase

**Tier 2: Nice to Have (Low Priority)**
9. Package Structure - List packages, files, dependencies
10. Call Hierarchy - See what calls/is called by a function

**Design Considerations:**
- Fast (sub-second responses)
- Reliable (works consistently)
- Simple API (clear parameters, predictable output)
- JSON output (structured, easy to parse)
- Works with absolute paths (no workspace issues)

**Status:** Planned, not started

---

### 3.2 Documentation Tools

**Purpose:** Help Claude work efficiently with documentation

**3.2.1 Go Documentation Tools**

**Planned Features (from discussion):**

**Tier 0: Documentation Operations (Critical)**
1. List/Extract Doc Comments
   - Get all godoc comments from file/package as structured data
   - Include function signature + doc comment
   - Extract metadata (domain tags, Examples sections)
   - Filter by exported/unexported, missing comments
   - Fast - no AST parsing overhead

2. Edit Doc Comment
   - Update function's doc comment directly
   - Preserve formatting and indentation
   - Can update metadata tags
   - No grep/sed/awk nonsense

3. Validate Doc Comments
   - Missing comments on exported symbols
   - Style consistency (punctuation, formatting)
   - Required metadata (domain tags, Examples for assertions)
   - Common mistakes (typos in tags, malformed examples)

**Tier 1: Testing Operations (High Value)**
4. List Tests
   - All test functions from file/package
   - Test names, subtests, table-driven test cases
   - Filter by pattern, benchmark vs test
   - Show test file coverage

5. Get Coverage
   - Run tests and return coverage data
   - Per-function coverage percentages
   - Uncovered lines
   - Coverage diff between runs
   - Fast - cached when code unchanged

6. Run Specific Tests
   - Execute tests by pattern/function
   - Return structured results (pass/fail, output, timing)

**Tier 2: Code Organization (Medium Priority)**
7. Reorder Declarations
   - Reorganize functions/types in file
   - Alphabetically, by type, by domain tag
   - Group related functions
   - Preview before applying

8. Auto Format/Organize
   - Run gofmt + goimports on files
   - Batch operation across multiple files
   - Return diff preview
   - Add missing imports, remove unused

**Status:** Planned, partially discussed

---

### 3.2.2 Markdown Linting Agent

**Purpose:** Automated markdown quality checking and fixing

**Type:** Agent (not just tool) - autonomous multi-step workflow

**Planned Features:**

**Auto-Fix (Safe):**
- Trailing whitespace
- Consistent list markers (- vs * vs +)
- Heading level fixes
- Missing language on code blocks (infer from content)
- Multiple consecutive blank lines
- Inconsistent emphasis markers (** vs __)

**Flag for Review (Uncertain):**
- Broken links (might be intentional/coming soon)
- Typos (might be technical terms - see below)
- Heading hierarchy changes that affect structure
- Missing alt text on images

**Typo Handling:**
- Auto-fix obvious typos (spell-checker with tech dictionary)
- Flag uncertain ones (technical terms, project-specific names)
- Learn from user feedback (whitelist)

**Output Format:**
```json
{
  "fixed": [
    {"file": "README.md", "line": 23, "issue": "trailing whitespace"},
    {"file": "docs/api.md", "line": 45, "issue": "heading level skip"}
  ],
  "review_needed": [
    {"file": "README.md", "line": 67, "issue": "Possible typo: 'testify' → 'test'?"},
    {"file": "docs/api.md", "line": 12, "issue": "Broken link: #installation"}
  ],
  "summary": {
    "files_processed": 47,
    "issues_fixed": 156,
    "issues_flagged": 12
  }
}
```

**Agent Workflow:**
1. Run markdown-lint on specified files
2. Categorize issues (safe auto-fix vs needs review)
3. Auto-fix the safe ones
4. Flag uncertain ones
5. Return summary

**Status:** Planned, architecture discussed

---

## Part 4: Outstanding Code Issues

### 4.1 Code Generation

**From main plan:**

1. ⏳ **Helper function tests** (99.5% → 100% coverage)
   - Helper functions (ObjectsAreEqual, CallerInfo, etc.) don't have Examples
   - Current approach: May not need 100% - helpers are tested indirectly
   - Alternative: Move helpers to separate package, don't generate tests

2. ⏳ **Type mapping for xxxFunc types**
   - Issue: PanicAssertionFunc references internal package types
   - Currently: Works via re-export aliases but untidy
   - Inherited: Poor API design from original testify
   - Fix: Rework type mapping for clean signatures

3. ✅ **Testable examples generation** (see 1.2 above)

---

### 4.2 Code Quality

**From main plan:**

1. ⏳ **Improve private comments** in internal/assertions
   - Godoc comments are excellent ✅
   - Private comments within functions need clarity
   - Better explain complex logic

2. ⏳ **Code quality assessment**
   - Review linting issues in internal packages
   - Focus on internalized dependencies (spew, difflib)
   - Overall codebase health check

3. ⏳ **Performance benchmarks**
   - Hot paths: Equal, Contains, Empty
   - Ensure generated code performs well
   - Identify optimization opportunities

---

### 4.3 Upstream Merges

**From stretchr/testify (need investigation):**

Critical fixes:
- ⏳ **#1824** - Investigate and adapt if relevant
- ⏳ **#1826** - Reported issue, investigate
- ⏳ **#1611** - Reported issue, investigate
- ⏳ **#1813** - Reported issue, investigate

Internalized dependency improvements:
- ⏳ **#1829** - Fix time.Time rendering in diffs (go-spew)
- ⏳ **#1822** - Deterministic map ordering (go-spew)
- ⏳ **#1816** - Fix panic on unexported struct key (go-spew)

UX improvements:
- ⏳ Diff rendering enhancements

Optional (enable/color module):
- ⏳ **#1467** - Colorized output with terminal detection
- ⏳ **#1480** - Colorized diffs via env var
- ⏳ **#1232** - Colorize expected/actual/errors
- ⏳ **#994** - Colorize values

---

## Part 5: Future Enhancements

### 5.1 Generic Assertions (Type Safety)

**From main plan - detailed design exists**

**Prime candidates:**
- Equal[T comparable]
- NotEqual[T comparable]
- Greater[T cmp.Ordered], Less[T cmp.Ordered]
- Contains[S ~[]E, E comparable]
- ElementsMatch[S ~[]E, E comparable]
- JSONEq[T any], YAMLEq[T any]

**Strategy:** Hybrid approach
- Keep `any` versions (backward compatible)
- Add generic variants (opt-in type safety)
- Both coexist in same package

**Status:** Design complete, implementation pending

---

### 5.2 Enhanced Diff Output

**Inspiration:** alecthomas/assert uses go-cmp for beautiful diffs

**Advantage:** We control difflib (internalized), can improve it

**Potential improvements:**
- Colorized output (via enable/color)
- Better struct formatting
- Improved readability
- Smart truncation

**Status:** Future consideration

---

## Priority Matrix

### Immediate (Before Public Launch)
1. 🔥 Fix documentation site critical issues (2.1)
2. 🔥 Add missing context to docs (2.2)
3. 🔥 Update TODO → NOT IMPLEMENTED markers (1.3)

### Short Term (Next Iteration)
4. 📝 Fix func type annotations for doc generator (1.1)
5. 📝 Add domain descriptions and weights (Hugo TODOs)
6. 📝 Complete Hugo site polish items

### Medium Term (Next Quarter)
7. 🛠️ Build Go development tools for Claude (3.1)
8. 🤖 Build markdown linting agent (3.2.2)
9. 🔍 Investigate upstream merges (4.3)
10. 📝 (Optional/Low) Render testable examples in HTML docs (1.2)

### Long Term (Future)
11. 🎯 Implement generic assertions (5.1)
12. 🎨 Enhanced diff output (5.2)
13. ⚡ Performance optimization
14. 📊 Meta-tests for generator

---

## Next Actions

**This Week:**
1. [ ] Fix documentation site critical issues (see docs-assessment.md)
2. [ ] Update all `// TODO` → `// NOT IMPLEMENTED` in examples
3. [ ] Add missing context pages to documentation site

**This Month:**
4. [ ] Extend doc generator to cover types/helpers
5. [ ] Complete Hugo site polish
6. [ ] Begin Go development tools planning

**This Quarter:**
7. [ ] Build and test Go development tools
8. [ ] Build and test markdown linting agent
9. [ ] Investigate and merge relevant upstream fixes
10. [ ] Plan generic assertions implementation

---

## Decisions Made (2025-12-31)

1. ✅ **TODO marker:** Replace `// TODO` with `// NOT IMPLEMENTED` in example comments
2. ✅ **Testable examples:** Already implemented! Optional HTML rendering is low priority
3. ✅ **Types in docs:** Most have domains. Focus on fixing func type annotations
4. ✅ **Priority confirmed:** Testable examples rendering = low priority
5. ⏳ **Tool vs Agent for markdown:** Agent approach confirmed
6. ⏳ **Go tools priority:** TBD - which features most urgent?

---

## Document Revision History

- 2025-12-31: Initial consolidated plan review
- [Future updates...]
