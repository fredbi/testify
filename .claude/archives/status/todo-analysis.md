# TODO Analysis: internal/assertions

**Date:** 2025-12-31
**Total TODOs:** 11
**Categories:** Error message quality (8), Code structure (1), Test enhancement (1), Test cleanup (1)

---

## Category 1: Error Message Quality Issues (8 TODOs)

**Files:** `equal_test.go` (lines 595, 602, 608, 617, 624, 636, 663)

### Problem Summary

The `Empty()` and `NotEmpty()` assertions produce confusing error messages when dealing with whitespace, control characters, and special Unicode characters. The error messages don't make it clear what the actual value was.

### Specific Issues

**1. Line 595: Spaces only**
```go
value:          "   "
expectedErrMsg: "Should be empty, but was    \n"
// TODO FIX THIS strange error message
```
**Issue:** Three spaces followed by newline looks like one trailing space in output.

**2. Lines 601-602: Single newline**
```go
value:          "\n"
expectedErrMsg: "Should be empty, but was \n"
// TODO This is the exact same error message as for an empty string
// TODO FIX THIS strange error message
```
**Issue:** Message shows nothing after "was" - looks identical to empty string message.

**3. Line 608: Tabs and newlines**
```go
value:          "\n\t\n"
expectedErrMsg: "Should be empty, but was \n\t\n"
// TODO The line feeds and tab are not helping to spot what is expected
```
**Issue:** Control characters are invisible in error output.

**4. Line 617: Multiple trailing newlines**
```go
value:          "foo\n\n"
expectedErrMsg: "Should be empty, but was foo\n\n"
// TODO it's not clear if one or two lines feed are expected
```
**Issue:** Can't tell if 1 or 2 newlines from the message alone.

**5. Line 624: Complex whitespace**
```go
value:          "\n\nfoo\t\n\t\n"
expectedErrMsg: "Should be empty, but was \n\nfoo\t\n\t\n"
// TODO The line feeds and tab are not helping to figure what is expected
```
**Issue:** Multiple control characters are invisible.

**6. Line 636: Non-breaking space**
```go
value:          "\u00a0" // NO-BREAK SPACE UNICODE CHARACTER
expectedErrMsg: "Should be empty, but was \u00a0\n"
// TODO here you cannot figure out what is expected
```
**Issue:** Unicode non-printable character is invisible in output.

**7. Line 663: Empty string NotEmpty**
```go
value:          ""
expectedErrMsg: `Should NOT be empty, but was ` + "\n"
// TODO FIX THIS strange error message
```
**Issue:** Message shows "was" followed by nothing - unclear.

### Root Cause

The error formatting in `isEmpty()` helper doesn't escape or visualize:
- Whitespace characters (spaces, tabs, newlines)
- Control characters
- Non-printable Unicode characters

This is inherited from the original testify - the issue exists upstream too.

### Proposed Solutions

**Option A: Escape special characters (Recommended)**

Modify error message formatting to use Go's `%q` format (quoted string with escapes):

```go
// Current:
"Should be empty, but was    \n"

// Proposed:
"Should be empty, but was \"   \"\n"
// or even better:
"Should be empty, but was: %q", "   " → "Should be empty, but was: \"   \""
```

**Benefits:**
- ✅ All whitespace visible as `\n`, `\t`, `\s`
- ✅ Unicode escapes visible as `\u00a0`
- ✅ Clear distinction between empty and whitespace
- ✅ Matches Go's standard string representation

**Option B: Add hex dump for non-printable**

Show hex representation for strings with control characters:

```go
"Should be empty, but was \"   \" (hex: 20 20 20)"
"Should be empty, but was \"\\n\" (hex: 0a)"
```

**Benefits:**
- ✅ Very clear what the actual bytes are
- ⚠️ More verbose

**Option C: Use repr/spew formatting**

Use the internalized spew package for better visualization:

```go
"Should be empty, but was: (string) (len=3) \"   \""
```

**Benefits:**
- ✅ Shows length
- ✅ Shows type
- ✅ Escapes special characters
- ⚠️ More verbose for simple cases

### Recommendation

**Use Option A (quoted strings) as the primary fix:**

1. Update `isEmpty()` in `equal.go` or wherever the error message is generated
2. Change format from:
   ```go
   fmt.Sprintf("Should be empty, but was %v", actual)
   ```
   To:
   ```go
   fmt.Sprintf("Should be empty, but was %q", actual)  // if string
      // or for any type:
      fmt.Sprintf("Should be empty, but was: %#v", actual)
   ```

3. Update all test expectations in `equal_test.go` accordingly:
   ```go
   expectedErrMsg: `Should be empty, but was "   "`
      expectedErrMsg: `Should be empty, but was "\n"`
      expectedErrMsg: `Should be empty, but was "\n\t\n"`
      expectedErrMsg: `Should be empty, but was "\u00a0"`
   ```

4. Remove all 8 TODO comments once fixed

**Action Items:**
- [ ] Locate error message generation code for `Empty()`/`NotEmpty()`
- [ ] Update to use `%q` or `%#v` formatting for string types
- [ ] Update test expectations in `equal_test.go`
- [ ] Remove 8 TODO comments
- [ ] Consider applying same fix to other assertions with similar issues

**Priority:** Medium - Quality of life improvement for users, not a bug

---

## Category 2: Code Structure (1 TODO)

**File:** `error.go:18`

```go
import "errors"

var ErrTest = errors.New("assert.ErrTest general error for testing")

// TODO: make a type and a const.
```

### Problem

`ErrTest` is currently a package-level variable created with `errors.New()`. The TODO suggests converting it to a custom error type with a constant.

### Analysis

**Current approach:**
```go
import "errors"

var ErrTest = errors.New("assert.ErrTest general error for testing")
```

**Suggested approach:**
```go
type testError string

func (e testError) Error() string {
	return string(e)
}

const ErrTest = testError("assert.ErrTest general error for testing")
```

### Why Change It?

**Benefits of custom type + const:**
- ✅ More efficient - no heap allocation
- ✅ Compile-time constant - cannot be modified
- ✅ Type-safe - distinct type from other errors
- ✅ Can use in const expressions
- ⚠️ More code

**Downsides:**
- ⚠️ More complex for a simple test error
- ⚠️ Breaks existing code if users are doing type assertions on ErrTest
- ⚠️ The current approach is idiomatic Go

### Recommendation

**Don't change it - remove the TODO.**

**Rationale:**
1. The current approach is idiomatic Go for sentinel errors
2. `errors.New()` is perfectly fine for test-only values
3. No performance benefit in test code
4. No functional benefit
5. Risk of breaking existing test code that checks for this error
6. The Go standard library uses this pattern extensively (e.g., `io.EOF`)

**If you really want to change it:**

The modern Go approach (Go 1.13+) would be:
```go
import "errors"

var ErrTest = errors.New("assert.ErrTest general error for testing")

// Keep it as-is, it's already correct!
```

**Action Items:**
- [x] **Remove TODO comment** - current implementation is fine
- [ ] OR if you want the "type + const" pattern, implement it (but I don't recommend it)

**Priority:** Low - current code is correct, change is cosmetic at best

---

## Category 3: Test Enhancement (1 TODO)

**File:** `json_test.go:8`

```go
// TODO(fred): load fixtures and assertions from embedded testdata
```

### Problem

The JSON tests currently use inline JSON strings. The TODO suggests moving test fixtures to embedded testdata files for better organization.

### Current State

Tests look like:
```go
import "testing"

func TestJSONEq_EqualSONString(t *testing.T) {
	t.Parallel()
	mock := new(testing.T)

	True(t, JSONEq(mock, `{"hello": "world", "foo": "bar"}`,
		`{"hello": "world", "foo": "bar"}`))
}
```

### Proposed Enhancement

**Using embedded testdata:**

1. Create `internal/assertions/testdata/json/` directory:
   ```
   testdata/json/
   ├── equal_simple.json
   ├── equal_complex.json
   ├── not_equal_simple.json
   └── fixtures.yaml  (metadata about test cases)
   ```

2. Embed testdata:
   ```go
   import "embed"
   
   //go:embed testdata/json/*.json
   var jsonTestData embed.FS
   ```

3. Load fixtures in tests:
   ```go
   import "testing"
   
   func TestJSONEq_EqualSONString(t *testing.T) {
   	expected, _ := jsonTestData.ReadFile("testdata/json/equal_simple.json")
   	actual, _ := jsonTestData.ReadFile("testdata/json/equal_simple.json")
   
   	True(t, JSONEq(mock, string(expected), string(actual)))
   }
   ```

### Benefits

**Pros:**
- ✅ Cleaner test code (no long inline JSON)
- ✅ Easier to maintain complex JSON fixtures
- ✅ Can share fixtures across multiple tests
- ✅ Fixtures can be used by other tools (validation, fuzzing)
- ✅ Better for large/realistic JSON examples

**Cons:**
- ⚠️ More files to manage
- ⚠️ Less obvious what's being tested (have to open fixture file)
- ⚠️ Overkill for simple test cases
- ⚠️ Current tests are already clear and working

### Recommendation

**Low priority enhancement - nice to have, not necessary.**

**Hybrid approach:**
- Keep simple inline tests for basic cases (current approach)
- Add testdata fixtures for complex scenarios if/when needed
- Don't migrate existing tests unless they become unwieldy

**Action Items:**
- [ ] **Update TODO to be more specific:**
  ```go
  // TODO(fred): Consider adding testdata fixtures for complex JSON test cases
  //             Current inline approach is fine for simple tests.
  ```
- [ ] Add testdata fixtures only when needed for complex test scenarios
- [ ] Keep current simple inline tests

**Priority:** Low - enhancement, not a problem

---

## Category 4: Test Cleanup (1 TODO)

**File:** `equal_test.go:219`

```go
func TestEqualEmpty(t *testing.T) {
    t.Parallel()

    // TODO(fredbi): redundant test context declaration
    chWithValue := make(chan struct{}, 1)
    chWithValue <- struct{}{}
    var tiP *time.Time
    var tiNP time.Time
    var s *string
    // ... more variable declarations
```

### Problem

The TODO suggests these variable declarations are redundant or should be moved/refactored.

### Analysis

Looking at the test, these variables are likely used in test cases below. The "redundant" comment might mean:

1. **Variables declared but not used** - if some are never referenced
2. **Should use a helper function** - declare them in a fixture generator
3. **Should be inline in test cases** - declare where used instead of top-level
4. **Duplicated across tests** - same variables in multiple test functions

### Recommendation

**Investigate and clean up:**

**Option A: Move to test case generator**
```go
import (
	"iter"
	"slices"
)

func equalEmptyCases() iter.Seq[equalEmptyCase] {
	chWithValue := make(chan struct{}, 1)
	chWithValue <- struct{}{}

	return slices.Values([]equalEmptyCase{
		{value: chWithValue, expected: false},
		// ...
	})
}
```

**Option B: Inline in test cases where used**
```go
// Only declare if actually used
```

**Option C: Remove if truly redundant**
```go
// Delete unused variables
```

**Action Items:**
- [ ] Check if all variables are actually used in the test
- [ ] If unused: delete them and the TODO
- [ ] If used: move to appropriate scope (test case generator or inline)
- [ ] Remove TODO once cleaned up

**Priority:** Low - code cleanup, not affecting functionality

---

## Summary & Recommendations

### Immediate Actions (High Value)

1. **Error Message Quality (8 TODOs):**
   - Fix: Update `Empty()`/`NotEmpty()` error formatting to use `%q` for strings
   - Impact: Better UX for users debugging tests
   - Effort: Medium (update error formatting + ~50 test expectations)
   - **Do this if you want to improve on testify upstream**

2. **ErrTest Type TODO:**
   - Fix: Remove TODO comment (current code is correct)
   - Impact: None (cleanup only)
   - Effort: Trivial (delete 1 comment line)
   - **Do this immediately**

### Later / Optional

3. **Test Cleanup (1 TODO):**
   - Fix: Investigate redundant variables, clean up
   - Impact: Minor (code cleanliness)
   - Effort: Low
   - **Do when refactoring tests**

4. **Testdata Fixtures (1 TODO):**
   - Fix: Clarify TODO or add fixtures when needed
   - Impact: Low (enhancement)
   - Effort: Medium if implementing
   - **Do only if complex JSON tests are added**

---

## Priority Matrix

| TODO | File | Priority | Effort | Impact | Recommendation |
|------|------|----------|--------|--------|----------------|
| Error messages (8) | equal_test.go | Medium | Medium | High | **Fix if improving UX** |
| ErrTest type | error.go | Low | Trivial | None | **Remove TODO** |
| Test cleanup | equal_test.go | Low | Low | Low | Later |
| Testdata fixtures | json_test.go | Low | Medium | Low | When needed |

---

## Next Steps

**Quick Wins (Do Now):**
1. Remove TODO from `error.go:18` - current code is fine
2. Update TODO in `json_test.go:8` to be more specific about "when needed"

**Quality Improvement (Do if Time):**
3. Fix error message formatting for `Empty()`/`NotEmpty()` assertions
4. Update all test expectations
5. Remove 8 error message TODOs

**Later:**
6. Clean up redundant test variables in `equal_test.go:219`

Would you like me to implement any of these fixes?
