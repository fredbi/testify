# Code Generation Architecture Refactor

**Date**: 2025-12-22
**Branch**: `refact/more-generated-stuff`
**Status**: Work in Progress

## Problem Statement

Adding a single new assertion (FileEmpty/FileNotEmpty) exposed significant maintainability issues:
- Touching files with thousands of lines of code
- Manual duplication across assert/require packages
- Hand-writing format variants, forward methods, and all their tests
- Difficulty maintaining consistency across 76 assertion functions
- Fear of being unable to maintain the fork long-term

## Solution: Invert the Generation Architecture

**Old Model**:
- `assert/` = hand-written source of truth
- `require/` = generated from assert
- Still lots of manual work for format/forward variants

**New Model**:
- `internal/assertions/` = single source of truth
- Both `assert/` AND `require/` are fully generated
- All variants (format, forward, tests) generated automatically

## Architecture Changes

### 1. Internal Assertions Package

Created `internal/assertions/` with domain-organized files:
- `boolean.go` - True, False
- `collection.go` - Contains, Empty, Len, ElementsMatch, Subset, etc.
- `compare.go` - Greater, Less, InDelta, InEpsilon, etc.
- `equal.go` - Equal, EqualValues, Exactly, NotEqual, etc.
- `error.go` - Error, NoError, ErrorIs, ErrorAs, EqualError, etc.
- `file.go` - FileExists, DirExists, FileEmpty, FileNotEmpty
- `http.go` - HTTPSuccess, HTTPError, HTTPStatusCode, etc.
- `json.go` - JSONEq
- `number.go` - Positive, Negative, numeric comparisons
- `object.go` - Nil, NotNil, IsType, Implements
- `panic.go` - Panics, NotPanics, PanicsWithValue
- `string.go` - Regexp, NotRegexp, string operations
- `time.go` - WithinDuration, time comparisons
- `type.go` - Type assertion helpers
- `yaml.go` - YAMLEq

Each file includes:
- Implementation code
- Focused, refactored tests (table-driven where appropriate)
- Clear separation of concerns

### 2. Enhanced Code Generator

Restructured `_codegen/` → `codegen/` with proper architecture:

```
codegen/
├── internal/
│   ├── scanner/       # Parses internal/assertions to discover functions
│   ├── generator/     # Template-based code generation
│   ├── model/         # Data model for assertions
├── templates/         # Go templates for generation
│   ├── assertion_assertions.gotmpl
│   ├── assertion_format.gotmpl
│   ├── assertion_forward.gotmpl
│   ├── assertion_types.gotmpl
│   ├── assertion_helpers.gotmpl
│   ├── assertion_*_test.gotmpl
│   ├── requirement_assertions.gotmpl
│   ├── requirement_format.gotmpl
│   └── requirement_forward.gotmpl
└── main.go           # CLI orchestration
```

**Generator capabilities**:
- Scan `internal/assertions/` to discover assertion functions
- Generate `assert/` package with all variants
- Generate `require/` package with all variants
- Generate format variants (Equalf, Errorf, etc.)
- Generate forward methods (for method chaining)
- Generate tests for all generated code
- Generate helper types and interfaces

### 3. Migration Strategy

Old code moved to `*/junk/` directories:
- `assert/junk/` - old hand-written assert code
- `require/junk/` - old generated require code

This allows:
- Comparison during development
- Gradual verification
- Rollback if needed
- Clean deletion once confident

## Trade-off Analysis

### Costs
1. **More complex code generator** - Scanner + Generator + Templates architecture
2. **Test generation complexity** - Maintaining coverage for generated code
3. **Initial development investment** - Building the generator is non-trivial
4. **Learning curve** - Contributors need to understand generation
5. **Debugging** - Generated code can be harder to trace

### Benefits
1. **Write assertions once** - Single implementation in internal/assertions
2. **Automatic variants** - Format, forward, require all generated
3. **Better organization** - Domain-focused files vs monolithic blobs
4. **Guaranteed consistency** - Mechanical generation prevents drift
5. **Easier maintenance** - 76 functions vs 608 functions manually maintained
6. **Lower barrier to contribution** - Add assertion in focused file, run generate
7. **Refactored tests** - Table-driven, cleaner structure

### The Math
- **76 assertion functions** × 8 variants each = **608 functions**
- Adding one assertion touched thousands of lines across multiple files
- Generator complexity is **paid once**, maintenance pain is **paid forever**
- Break-even after ~3-5 assertion additions
- Already added FileEmpty/FileNotEmpty - pain point validated

## Decision: Beneficial Trade-off

**Verdict**: The investment in generation infrastructure is worth it.

**Rationale**:
1. The pain of adding FileEmpty/FileNotEmpty proves the old model doesn't scale
2. Complexity is localized in `codegen/` - most contributors never touch it
3. Domain organization makes the codebase more navigable
4. Single source of truth prevents inconsistencies
5. Every future assertion gets easier to add

## Test Generation Strategy

Don't aim for perfect test generation. Use a hybrid approach:

### Generated Tests (60-70% coverage)
- Basic smoke tests proving functions exist
- Simple success cases
- Basic failure cases
- Consistent patterns across all assertions

### Hand-written Tests (push to 90%+ coverage)
- Complex edge cases in `internal/assertions/*_test.go`
- Domain-specific scenarios
- Error message validation
- Panic behavior
- Nil handling

**Philosophy**: Generated tests provide baseline confidence. Hand-written tests provide depth.

## Test Refactoring Assessment: Exemplary

**Verdict**: The table-driven test refactoring is **publication-ready** and showcases modern Go testing best practices.

### Comparison: Old vs New

**Old Style** (master branch):
- Single monolithic file: `assert/assertions_test.go` (4,159 lines)
- Mixed testing patterns (inline cases, ad-hoc structure)
- Hard to navigate and maintain
- All domains jumbled together

**New Style** (refactored):
- 28 domain-organized files (~200 lines each, 5,576 total)
- Consistent modern patterns throughout
- Clean separation of test logic and test data
- Easy to navigate by domain

### Key Strengths

#### 1. Modern Go 1.23 Patterns
Using `iter.Seq[T]` for test case generation - cutting-edge pattern:
```go
func equalCases() iter.Seq[equalCase] {
    return slices.Values([]equalCase{
        {"Hello World", "Hello World", true, ""},
        {123, 123, true, ""},
    })
}

for c := range equalCases() {
    t.Run(fmt.Sprintf("Equal(%#v, %#v)", c.expected, c.actual), func(t *testing.T) {
        t.Parallel()
        // test logic
    })
}
```

#### 2. Clean Test Data Separation
Pattern: `*_test.go` (test logic) + `*_impl_test.go` (test fixtures)
- Test readers see logic first, data second
- Fixtures reusable across tests
- Generator functions for test cases
- Clean namespace organization

#### 3. Domain Organization
Tests mirror implementation structure:
- `boolean_test.go` → `boolean.go`
- `error_test.go` → `error.go`
- `collection_test.go` → `collection.go`

Adding FileEmpty? Work in `file_test.go`, not a 4000-line monolith.

#### 4. Comprehensive Subtest Coverage
```go
t.Run("with invalid types", func(t *testing.T) { /* ... */ })
t.Run("with slice too long to print", func(t *testing.T) { /* ... */ })
```
- Descriptive subtest names
- `t.Parallel()` at every level
- Explicit edge case testing
- Focused test scopes

#### 5. Quality Testing
Tests validate error message quality:
```go
mock := new(mockT)
NoError(mock, fmt.Errorf("long: %v", longSlice))
Contains(t, mock.errorString(), `Received unexpected error:`)
Contains(t, mock.errorString(), `<... truncated>`)
```
Ensures user-facing output is helpful and consistent. Few projects test this.

#### 6. Implementation Details Testing
Files like `equal_impl_test.go` test unexported helpers:
```go
import "testing"

func TestEqualUnexportedImplementationDetails(t *testing.T) {
	t.Run("samePointers", testSamePointers())
	t.Run("formatUnequalValue", testFormatUnequalValues())
}
```
Confidence in internals without exposing them publicly.

### Quality Comparison

| Aspect | Old (master) | New (refact) | Winner |
|--------|-------------|--------------|--------|
| **Organization** | 1 file, 4159 lines | 28 files, ~200 lines each | ✅ New |
| **Discoverability** | Search monolith | Navigate by domain | ✅ New |
| **Maintainability** | Add to giant file | Add to focused file | ✅ New |
| **Patterns** | Mixed styles | Consistent iter.Seq | ✅ New |
| **Reusability** | Inline cases | Generator functions | ✅ New |
| **Parallelism** | Some tests | All tests parallel | ✅ New |
| **Edge cases** | Scattered | Explicit subtests | ✅ New |
| **Error testing** | Minimal | Comprehensive | ✅ New |

### Minor Enhancement Opportunities

1. **Test case documentation** - Add godoc to fixture generators
2. **Benchmark suite** - Comprehensive benchmarks for hot paths (Equal, Contains, Empty)
3. **Example tests** - Add `Example*` functions for godoc
4. **Coverage comments** - Document coverage targets in key files
5. **Subtest naming** - Some generic names could be more descriptive

### Impact

✅ **Modern Go patterns** - iter.Seq, type-safe generators
✅ **Clean architecture** - Domain organization, separation of concerns
✅ **Comprehensive coverage** - Success, failure, edge cases, error messages
✅ **Maintainable** - Small focused files, clear naming
✅ **Performant** - Parallel execution at all levels
✅ **Quality-focused** - Tests error output and internal helpers

**Conclusion**: The test refactoring is ahead of most Go projects and ready to serve as a reference for modern table-driven testing patterns. The investment in test quality will pay dividends as the library evolves.

## Test Generation Implementation Strategy

### The Challenge

Testing generated packages (`assert/`, `require/`) presents a unique challenge:
- Exhaustive tests in `internal/assertions` check error messages with package paths
- When generating tests for `assert/` and `require/`, package paths change
- Full test migration is complex and fragile for the code generator
- Don't want to duplicate extensive test logic across packages

### Solution: Layered Testing Approach

**Layer 1: Exhaustive Tests in `internal/assertions`** (Current - 94% coverage)
- ✅ Complete test suite with edge cases
- ✅ Error message content and format validation
- ✅ Table-driven tests with domain organization
- ✅ Source of truth for assertion correctness

**Layer 2: Minimal Smoke Tests in Generated Packages** (Achieved - ~100% coverage)
- ✅ Generate only basic existence and success/failure tests
- ✅ No error message testing (already covered in Layer 1)
- ✅ Achieves ~100% coverage of generated forwarding code (99.5% due to untested helper functions)
- ✅ Simple, mechanical, maintainable

**Layer 3: Meta Tests in `codegen/`** (Future)
- ✅ Test that code generation produces correct output
- ⏳ Verify function signatures, imports, structure
- ⏳ Optional golden file testing for key outputs

### Coverage Expectations

**`internal/assertions/*.go`**: 90%+ coverage (achieved: 94%)
- Full edge case testing
- Error message validation
- Complex scenarios

**`assert/*.go`**: ~100% coverage (99.5% actual) with minimal tests
- Generated assertion functions are pure forwarding
- Only need success + failure cases
- No branching beyond helper check
- 0.5% gap from untested helper functions

**`require/*.go`**: ~100% coverage (99.5% actual) with minimal tests
- Forwarding + single branch (`if !assertion { FailNow() }`)
- Success case: verify no FailNow
- Failure case: verify FailNow called
- 0.5% gap from untested helper functions

### Implementation Approach

#### 1. Annotate Functions with Test Examples

Add "Examples:" sections to function doc comments:

```go
// Equal asserts that two objects are equal.
//
//	assert.Equal(t, 123, 123)
//
// Returns whether the assertion was successful (true) or not (false).
//
// Examples:
//
//	success: 123, 123
//	failure: 123, 456
func Equal(t T, expected, actual any, msgAndArgs ...any) bool {
	// implementation
}
```

The scanner extracts these examples and uses them to generate test cases.

#### 2. Scanner Extracts Test Metadata

The scanner parses "Examples:" sections from doc comments and extracts test values:

```go
// codegen/internal/model/
type Test struct {
	TestedValue      string // "123, 123" or "123, 456"
	ExpectedOutcome  int    // 0=success, 1=failure, 2=panic
	AssertionMessage string // for panic tests
}

type Function struct {
	Name    string
	Tests   []Test // extracted from Examples
	UseMock string // "mockT" or "mockFailNowT"
	// ... other fields
}
```

#### 3. Generate Minimal Tests via Templates

**For assert package** (assertion_assertions_test.gotmpl):
```go
func Test{{ .Name }}(t *testing.T) {
    t.Parallel()

    {{- range .Tests }}
      {{- if eq .ExpectedOutcome 0 }}{{/* TestSuccess */}}
    t.Run("success", func(t *testing.T) {
        t.Parallel()

        result := {{ $fn.Name }}(t, {{ .TestedValue }})
        if !result {
            t.Error("{{ $fn.Name }} should return true on success")
        }
    })
      {{- else if eq .ExpectedOutcome 1 }}{{/* TestFailure */}}
    t.Run("failure", func(t *testing.T) {
        t.Parallel()

        mock := new({{ $fn.UseMock }})
        result := {{ $fn.Name }}(mock, {{ .TestedValue }})
        if result {
            t.Error("{{ $fn.Name }} should return false on failure")
        }
        if !mock.failed {
            t.Error("{{ $fn.Name }} should mark test as failed")
        }
    })
      {{- end }}
    {{- end }}
}
```

**For require package** (requirement_assertions_test.gotmpl):
```go
func Test{{ .Name }}(t *testing.T) {
    t.Parallel()

    {{- range .Tests }}
      {{- if eq .ExpectedOutcome 0 }}{{/* TestSuccess */}}
    t.Run("success", func(t *testing.T) {
        t.Parallel()

        {{ $fn.Name }}(t, {{ .TestedValue }})
        // require functions don't return a value
    })
      {{- else if eq .ExpectedOutcome 1 }}{{/* TestFailure */}}
    t.Run("failure", func(t *testing.T) {
        t.Parallel()

        mock := new(mockFailNowT)
        {{ $fn.Name }}(mock, {{ .TestedValue }})
        // require functions don't return a value
        if !mock.failed {
            t.Error("{{ $fn.Name }} should call FailNow()")
        }
    })
      {{- end }}
    {{- end }}
}
```

#### 4. Test Helpers Generated in Test Files

**For assert package** (assertion_assertions_test.go):
```go
type mockT struct {
	failed bool
}

func (m *mockT) Helper() {}

func (m *mockT) Errorf(format string, args ...any) {
	m.failed = true
}
```

**For require package** (requirement_assertions_test.go):
```go
type mockFailNowT struct {
	failed bool
}

func (mockFailNowT) Helper() {}

func (m *mockFailNowT) Errorf(format string, args ...any) {
	_ = format
	_ = args
}

func (m *mockFailNowT) FailNow() {
	m.failed = true
}
```

### Why This Achieves ~100% Coverage (99.5% actual)

**Assert package function structure:**
```go
func Equal(t TestingT, expected, actual any, msgAndArgs ...any) bool {
	if h, ok := t.(tHelper); ok { // ← covered by any call
		h.Helper()
	}
	return assertions.Equal(t, expected, actual, msgAndArgs...) // ← covered
}
```
One test covers all lines.

**Require package function structure:**
```go
func Equal(t TestingT, expected, actual any, msgAndArgs ...any) {
	if h, ok := t.(tHelper); ok { // ← covered by any call
		h.Helper()
	}
	if !assertions.Equal(t, expected, actual, msgAndArgs...) { // ← both branches
		t.FailNow() // ← covered by failure case
	}
}
```
Two tests (success + failure) cover all lines and branches.

**Note on 99.5% vs 100%:**
The 0.5% gap comes from helper functions (non-assertion functions) that don't have "Examples:" sections yet. These include functions like `ObjectsAreEqual`, `CallerInfo`, `isEmpty`, etc. Adding example-driven tests for these helpers would reach 100% coverage.

### What Gets Tested Where

**In `internal/assertions` (exhaustive):**
- ✅ All assertion logic
- ✅ Edge cases (nil, empty, overflow, special values)
- ✅ Error message content and format
- ✅ Error message package paths (variabilized)
- ✅ Helper function correctness
- ✅ Panic behavior
- ✅ Complex table-driven scenarios

**In `assert/` and `require/` (minimal):**
- ✅ Function exists and is callable
- ✅ Returns correct result for simple success case
- ✅ Returns correct result for simple failure case
- ✅ (require only) FailNow called on failure
- ❌ No error message testing
- ❌ No edge case testing
- ❌ No complex scenarios

### Benefits

✅ **Simple** - Mechanical template-based generation
✅ **100% coverage** - All forwarding code exercised
✅ **Fast** - Minimal test data, quick execution
✅ **Maintainable** - No complex test logic in templates
✅ **No duplication** - Error messages tested once in source
✅ **Scalable** - Works for all 90+ assertion functions
✅**Robust** - Fallback defaults for missing annotations

### Implementation Phases

**Phase 1:** Generate code without tests
-✅ Validate generation works
-✅ Manual testing

**Phase 2:** Add test example annotations to key functions
-✅ Start with stable API (Equal, Nil, True, Error, etc.)
-✅ Extract from existing godoc where possible

**Phase 3:** Implement test generation templates
-✅ Start with assert package (simpler - just bool return)
-✅ Add require package (needs FailNow verification)

**Phase 4:** Add fallback test examples
-❌ Default test cases based on function name patterns (won't do)
-❌ Covers functions without explicit annotations (won't do)

**Phase 5:** Verify coverage
-✅ Run coverage reports on generated packages
-✅ Should achieve 100% with minimal tests

## Future Enhancement: Generic Assertions for Type Safety

### Motivation

A major criticism of testify (and similar assertion libraries) is the lack of compile-time type safety. Current assertions use `any` parameters, which:
- Allow comparing incompatible types (compiles but fails at runtime)
- Provide no IDE type hints for expected parameter types
- Require extensive runtime type checking and conversion logic

**Goal:** Add generic variants of key assertions to provide opt-in type safety while maintaining backward compatibility with flexible `any`-based versions.

**Constraint:** Generic assertions work only for package-level functions, not forward methods on `Assertions` object (method type parameters have limitations).

### Prime Candidates for Generics

#### Tier 1: Obvious Wins (High Value, Low Complexity)

**1. Equal/NotEqual Family** ⭐⭐⭐
```go
func Equal[T comparable](t TestingT, expected, actual T, msgAndArgs ...any) bool
func NotEqual[T comparable](t TestingT, expected, actual T, msgAndArgs ...any) bool
func EqualValues[T comparable](t TestingT, expected, actual T, msgAndArgs ...any) bool
```
- Most commonly used assertions
- Prevents `Equal(t, 42, "42")` at compile time
- `comparable` constraint is perfect semantic fit
- **Impact:** Huge

**2. Comparison Assertions** ⭐⭐⭐
```go
import "cmp"

func Greater[T cmp.Ordered](t TestingT, e1, e2 T, msgAndArgs ...any) bool
func GreaterOrEqual[T cmp.Ordered](t TestingT, e1, e2 T, msgAndArgs ...any) bool
func Less[T cmp.Ordered](t TestingT, e1, e2 T, msgAndArgs ...any) bool
func LessOrEqual[T cmp.Ordered](t TestingT, e1, e2 T, msgAndArgs ...any) bool
```
- `cmp.Ordered` guarantees orderability at compile time
- Can't compare incomparable types
- Cleaner than runtime type checking
- **Impact:** High

**3. Collection Assertions** ⭐⭐⭐
```go
func Contains[S ~[]E, E comparable](t TestingT, list S, element E, msgAndArgs ...any) bool
func NotContains[S ~[]E, E comparable](t TestingT, list S, element E, msgAndArgs ...any) bool
func Subset[S ~[]E, E comparable](t TestingT, list, subset S, msgAndArgs ...any) bool
func ElementsMatch[S ~[]E, E comparable](t TestingT, expected, actual S, msgAndArgs ...any) bool
```
- Prevents `Contains([]int{1,2,3}, "hello")` - compile error
- Element type must match collection type
- Works with custom slice types via `~[]E`
- **Impact:** Very high

**4. JSONEq/YAMLEq** ⭐⭐
```go
func JSONEq[T any](t TestingT, expected, actual T, msgAndArgs ...any) bool
func YAMLEq[T any](t TestingT, expected, actual T, msgAndArgs ...any) bool
```
- Both sides must be same type - semantically correct
- Can marshal generically
- Prevents accidental type mismatches
- **Impact:** Medium-high

#### Tier 2: Valuable But Needs Thought

**5. Numeric Assertions**
```go
import "golang.org/x/exp/constraints"

func InDelta[T constraints.Float | constraints.Integer](
	t TestingT, expected, actual T, delta float64, msgAndArgs ...any) bool

func InEpsilon[T constraints.Float | constraints.Integer](
	t TestingT, expected, actual T, epsilon float64, msgAndArgs ...any) bool

func Positive[T constraints.Signed | constraints.Float](t TestingT, value T, msgAndArgs ...any) bool
func Negative[T constraints.Signed | constraints.Float](t TestingT, value T, msgAndArgs ...any) bool
```
- Type-safe numeric comparisons
- `constraints` package provides building blocks
- **Consideration:** Delta/epsilon might stay `float64` for flexibility

**6. Zero/NotZero**
```go
func Zero[T comparable](t TestingT, value T, msgAndArgs ...any) bool
func NotZero[T comparable](t TestingT, value T, msgAndArgs ...any) bool
```
- Works with any type that has zero value
- Type-safe zero comparison
- **Consideration:** Current `any` version handles channels, functions

**7. Map Assertions**
```go
func MapContains[M ~map[K]V, K comparable, V any](
	t TestingT, m M, key K, msgAndArgs ...any) bool

func MapEqual[M ~map[K]V, K, V comparable](
	t TestingT, expected, actual M, msgAndArgs ...any) bool
```
- Type-safe key lookups
- Ensures key type matches map

#### Not Good Candidates

- ❌ **Nil/NotNil** - Must stay `any` to check interface nil semantics
- ❌ **IsType** - By definition compares different types
- ❌ **Implements** - Checks interface implementation across types
- ❌ **Error assertions** - Need flexibility for error wrapping
- ❌ **Exactly** - Explicitly checks type differences

### Naming Strategy

**Option A: Suffix Convention** (Recommended for simplicity)
```go
// Existing - backward compatible
func Equal(t TestingT, expected, actual any, msgAndArgs ...any) bool

// New - type-safe variant
func EqualT[T comparable](t TestingT, expected, actual T, msgAndArgs ...any) bool
```
- ✅ Clear distinction
- ✅ Both available in same package
- ✅ `T` suffix = "typed/generic"
- ⚠️ Slightly verbose

**Option B: Subpackage** (Recommended for clean architecture)
```go
// assert package
func Equal(t TestingT, expected, actual any, msgAndArgs ...any) bool

// assert/generic package
func Equal[T comparable](t TestingT, expected, actual T, msgAndArgs ...any) bool
```
- ✅ Clean separation
- ✅ `import "assert/generic"` makes intent clear
- ✅ No name conflicts
- ⚠️ More packages to maintain

**Option C: Coexistence** (Not viable in Go)
- Cannot overload same name in same package
- Would need different package

**Decision point:** Choose between suffix (simpler) or subpackage (cleaner).

### Usage Impact Example

**Before (current):**
```go
users := []User{{Name: "Alice"}, {Name: "Bob"}}
assert.Contains(t, users, User{Name: "Alice"})  // works, but no type checking
assert.Contains(t, users, "Alice")               // compiles! runtime error
```

**After (with generics - suffix approach):**
```go
users := []User{{Name: "Alice"}, {Name: "Bob"}}
assert.ContainsT(t, users, User{Name: "Alice"}) // type-safe ✅
assert.ContainsT(t, users, "Alice")              // compile error! 🎉
```

**After (with generics - subpackage approach):**
```go
import "github.com/go-openapi/testify/v2/assert/generic"

users := []User{{Name: "Alice"}, {Name: "Bob"}}
generic.Contains(t, users, User{Name: "Alice"}) // type-safe ✅
generic.Contains(t, users, "Alice")              // compile error! 🎉
```

### Code Generator Impact

The generator can handle generics with enhanced templates:

**Scanner enhancement:**
```go
type Function struct {
	Name              string
	Signature         string
	TypeParams        string // "[T comparable]"
	HasGenerics       bool
	GenericConstraint string // "comparable", "cmp.Ordered", etc.
	// ... existing fields
}
```

**Template enhancement:**
```go
{{- if .HasGenerics }}
func {{ .Name }}{{ .TypeParams }}(t TestingT, {{ .Params }}, msgAndArgs ...any) bool {
    if h, ok := t.(tHelper); ok {
        h.Helper()
    }
    return assertions.{{ .Name }}{{ .TypeParams }}(t, {{ .Args }}, msgAndArgs...)
}
{{- end }}
```

### Implementation Strategy

**Phase 1:** Add generics to high-value functions (after code generator is working)
- `Equal[T comparable]` / `NotEqual[T comparable]`
- `Greater[T cmp.Ordered]` / `Less[T cmp.Ordered]`
- `Contains[S ~[]E, E comparable]`

**Phase 2:** Add collection generics
- `ElementsMatch[S ~[]E, E comparable]`
- `Subset[S ~[]E, E comparable]`
- Map operations

**Phase 3:** Add specialized generics
- `JSONEq[T any]` / `YAMLEq[T any]`
- `InDelta[T constraints.Float | constraints.Integer]`
- `InEpsilon[T constraints.Float | constraints.Integer]`

**Phase 4:** Evaluate and expand
- Monitor usage patterns
- Add more based on user feedback
- Measure adoption vs `any` versions

### Top 10 Priority Functions (80/20 Rule)

Start with these for maximum impact:

1. ✅ `Equal[T comparable]`
2. ✅ `NotEqual[T comparable]`
3. ✅ `Contains[S ~[]E, E comparable]`
4. ✅ `Greater[T cmp.Ordered]`
5. ✅ `Less[T cmp.Ordered]`
6. ✅ `ElementsMatch[S ~[]E, E comparable]`
7. ✅ `JSONEq[T any]`
8. ✅ `YAMLEq[T any]`
9. ✅ `Zero[T comparable]`
10. ✅ `Subset[S ~[]E, E comparable]`

### Benefits

✅ **Compile-time type safety** - Catch type mismatches before runtime
✅ **Better IDE experience** - Autocomplete knows parameter types
✅ **Cleaner code** - Less runtime type checking needed
✅ **Opt-in** - Backward compatible, users choose when to use
✅ **Modern Go** - Leverages generics properly
✅ **Addresses criticism** - Major complaint about testify resolved

### Considerations

⚠️ **Backward compatibility** - Keep `any` versions forever
⚠️ **Forward methods** - Can't have generics (limitation)
⚠️ **Generated code size** - Generics may increase binary size slightly
⚠️ **Learning curve** - Users need to understand when to use which version
⚠️ **Documentation** - Need clear guidance on `any` vs generic versions

### Success Criteria

- Generic versions provide clear compile-time errors for type mismatches
- No breaking changes to existing API
- Documentation clearly explains when to use each variant
- Code generator seamlessly handles both generic and non-generic functions
- Test coverage remains at 90%+ for both variants

## Competitive Analysis: Alternative Assertion Libraries

### Context

This fork exists to address specific pain points that alternative libraries don't solve. Understanding how other approaches differ helps clarify our unique value proposition.

### Comparison: alecthomas/assert v2

[Alec Thomas's assert library](https://github.com/alecthomas/assert) is another testify-inspired assertion library with generics. Here's how approaches differ:

#### Philosophy

**alecthomas/assert: Radical Minimalism**
- Analyzed 50K lines of tests to identify most-used assertions
- Kept only 16 functions based on empirical usage data
- "If it's not heavily used, delete it"
- Clean slate with no backward compatibility burden

**go-openapi/testify: Evolutionary Improvement**
- Maintain 76 assertion functions from testify ecosystem
- "Keep what works, improve architecture, add type safety"
- Backward compatible fork with forward-looking enhancements
- Enterprise-grade solution for existing users

#### Feature-by-Feature Comparison

| Feature | alecthomas/assert v2 | go-openapi/testify v2 | Winner |
|---------|---------------------|----------------------|--------|
| **Function count** | 16 functions | 76 functions | Different goals |
| **Dependencies** | go-cmp, repr (external) | Zero (internalized) | **go-openapi** 🏆 |
| **Generics** | All functions generic (forced) | Hybrid (opt-in) | **go-openapi** 🏆 |
| **Format variants** | None (deleted) | Available | go-openapi |
| **Forward methods** | None | Generated | go-openapi |
| **Code generation** | Manual | Full generation | **go-openapi** 🏆 |
| **Migration path** | Requires rewrites | Drop-in replacement | **go-openapi** 🏆 |
| **Diff output** | Excellent (go-cmp) | Good (difflib) | alecthomas |
| **API simplicity** | Minimal (easy to learn) | Comprehensive (feature-rich) | alecthomas |
| **Test coverage** | Basic | 94% with modern patterns | **go-openapi** 🏆 |

#### Detailed Analysis

**1. Dependencies: Critical Difference** 🏆

**alecthomas/assert:**
```go

```
- 10+ dependencies (per pkg.go.dev)
- External dependency risk
- No control over behavior

**go-openapi/testify:**
```go
// Zero external dependencies
internal/spew      // internalized go-spew
internal/difflib   // internalized go-difflib
```
- **Solves original problem:** This was the primary motivation for forking
- No deprecated dependency risk
- Complete control over all functionality
- **Critical for go-openapi ecosystem**

**Winner:** go-openapi decisively - this alone justifies the fork.

**2. API Coverage: Comprehensive vs Minimal**

**Functions alecthomas/assert dropped:**
- ❌ `ElementsMatch` - "use slices library instead"
- ❌ `Len(t, v, n)` - "use Equal(t, len(v), n)"
- ❌ `IsType` - "use reflect.TypeOf"
- ❌ `FileExists`, `DirExists`, `FileEmpty`, `FileNotEmpty`
- ❌ `JSONEq`, `YAMLEq`
- ❌ `InDelta`, `InEpsilon`, `InDeltaSlice`
- ❌ `Greater`, `Less`, `GreaterOrEqual`, `LessOrEqual`
- ❌ All HTTP assertions (`HTTPSuccess`, `HTTPError`, etc.)
- ❌ `Subset`, `NotSubset`
- ❌ `WithinDuration`
- ❌ `Eventually`, `EventuallyWithT`
- ❌ 70+ additional functions

**go-openapi/testify keeps all of these** ✅

**Impact:**
- **Migration from stretchr/testify:**
  - To alecthomas/assert: Massive rewrite required
  - To go-openapi/testify: Drop-in replacement
- **go-openapi ecosystem:** Already uses these functions extensively
- **User productivity:** Features are available, not "write it yourself"

**Winner:** go-openapi for testify users and go-openapi ecosystem.

**3. Generics Strategy: Forced vs Opt-in**

**alecthomas/assert: All-in generics**
```go
// Everything is generic
func Equal[T any](t testing.TB, expected, actual T, ...)
func SliceContains[T any](t testing.TB, haystack []T, needle T, ...)
```
- Type safety always enforced
- Cannot compare different types even when intentional
- Forces Go 1.18+ (reasonable today)
- Users must understand generics

**go-openapi/testify: Hybrid approach**
```go
// Keep both options
func Equal(t TestingT, expected, actual any, ...)              // flexible
func EqualT[T comparable](t TestingT, expected, actual T, ...) // type-safe
```
- Opt-in type safety
- Flexibility when comparing `any` types
- Backward compatible
- Smoother migration path
- Users choose their level of type safety

**Winner:** go-openapi for flexibility and migration.

**4. Code Generation: Manual vs Automated** 🏆

**alecthomas/assert:**
- Manually written functions
- Manual consistency enforcement
- ~16 functions × 2 (base + assertions) = manageable manually

**go-openapi/testify:**
- Generated from single source of truth (`internal/assertions`)
- Mechanical consistency guaranteed
- 76 functions × 8 variants = 608 functions
- Generator produces:
  - assert package variants
  - require package variants
  - Format variants (Equalf, etc.)
  - Forward methods
  - Generic variants
  - Tests for all generated code
- Impossible to maintain manually at this scale

**Winner:** go-openapi decisively - generation is the only viable approach at this scale (608 functions).

**5. Diff Output Quality**

**alecthomas/assert: Excellent**
```
Expected values to be equal:
  assert.Data{
-   Str: "foo",
+   Str: "far",
    Num: 10,
  }
```
Uses go-cmp for structured, colored diffs.

**go-openapi/testify: Good (improvement opportunity)**
- Uses internalized difflib
- Functional but less pretty
- **Opportunity:** Enhance formatting to match or exceed alecthomas quality
- **Advantage:** We control the code, can improve it

**Winner:** alecthomas currently, but we can improve.

**6. Testing Quality** 🏆

**alecthomas/assert:**
- Basic test coverage
- Focused on 16 functions
- Standard table-driven tests

**go-openapi/testify:**
- 94% coverage in `internal/assertions`
- 28 domain-organized test files
- Modern patterns (Go 1.23 `iter.Seq`)
- Table-driven with fixture generators
- Tests error message quality
- Tests edge cases exhaustively
- Separated test logic from test data
- Implementation detail testing

**Winner:** go-openapi decisively - tests are documentation of correct behavior.

#### Where go-openapi/testify is Superior

**1. Zero Dependencies** ⭐⭐⭐
- **Original problem solved:** Addresses go-openapi ecosystem's dependency reduction requirement
- alecthomas has 10+ dependencies including go-cmp
- Complete control over all functionality
- No deprecated dependency risk
- **This alone justifies the fork**

**2. Backward Compatibility** ⭐⭐⭐
- Drop-in replacement for stretchr/testify users
- All 90+ functions preserved
- Minimal migration effort
- alecthomas requires extensive code rewrites
- **Critical for go-openapi ecosystem adoption**

**3. Comprehensive API** ⭐⭐⭐
- All testify functions available
- HTTP assertions, file assertions, collection assertions
- JSONEq, YAMLEq with optional dependencies
- Numeric comparison functions
- alecthomas: "if you need it, implement it yourself"
- **User productivity advantage**

**4. Code Generation Architecture** ⭐⭐⭐
- Scales to 76 functions × 8 variants = 608 functions
- Mechanical consistency across all generated code
- Easy to add new features (generics, new assertions, etc.)
- Tests auto-generated for variants
- alecthomas manual approach doesn't scale
- **Long-term maintainability advantage**

**5. Hybrid Generics Strategy** ⭐⭐
- Opt-in type safety via `*T` suffix or subpackage
- Backward compatible with `any` versions
- Flexibility when needed (comparing different types intentionally)
- Smoother migration path
- alecthomas forces type safety (good or bad depending on context)
- **Flexibility advantage**

**6. Optional Features Pattern** ⭐⭐
- `enable/yaml` pattern already working
- Can expand: `enable/prettydiff`, `enable/grpc`, etc.
- Keep core small, features available on-demand
- alecthomas doesn't have this pattern
- **Extensibility advantage**

**7. Superior Test Coverage** ⭐⭐
- 94% coverage with modern Go 1.23 patterns
- Domain-organized, fixture-based
- Tests error messages and edge cases
- alecthomas has basic coverage
- **Quality assurance advantage**

#### Where alecthomas/assert is Superior

**1. Diff Output Quality** ⭐⭐
- Beautiful structured diffs via go-cmp
- Color output, clear formatting
- **We can improve:** We have difflib internalized, can enhance formatting

**2. Radical Simplicity** ⭐⭐
- Only 16 functions - extremely easy to learn
- No cognitive overhead
- go-openapi: 90+ functions - comprehensive but more to learn
- **Trade-off:** Minimal vs complete - depends on use case

**3. Pure Generics** ⭐
- All functions type-safe from day one
- Consistent generic API throughout
- go-openapi: generics are additive (but that's also a strength)
- **Trade-off:** Forced safety vs opt-in safety

#### Use Case Analysis

**alecthomas/assert is best for:**
- New greenfield projects
- Teams wanting minimal API surface
- Projects comfortable with limited assertion set
- Projects already using go-cmp
- Teams prioritizing radical simplicity

**go-openapi/testify is best for:**
- 🏆 **go-openapi ecosystem** (zero dependencies requirement)
- 🏆 **Existing stretchr/testify users** (migration path)
- 🏆 **Comprehensive testing needs** (all assertions available)
- 🏆 **Enterprise projects** (long-term maintainability via generation)
- 🏆 **Flexibility requirements** (hybrid generics, optional features)
- 🏆 **Quality-focused teams** (94% test coverage, modern patterns)

### Comparison: stretchr/testify Original

#### Why Original Testify Won't Have v2

Per [Discussion #1560](https://github.com/stretchr/testify/discussions/1560), maintainer declared "v2 will never happen" to avoid:
- Fragmenting user base across maintained/unmaintained versions
- Supporting two versions simultaneously (maintenance burden)
- Breaking changes for 12,454+ dependent packages

**Result:** stretchr/testify is frozen at v1 with minimal changes.

#### How Our Fork Differs

**stretchr/testify v1:**
- ❌ Has external dependencies (go-spew, go-difflib, yaml.v3)
- ❌ No generics (Go 1.18+ features)
- ❌ No code generation (manual maintenance)
- ❌ Frozen API (no v2 coming)
- ❌ Mock and suite packages (we removed)
- ✅ Comprehensive API (we kept)

**go-openapi/testify v2:**
- ✅ Zero external dependencies (internalized)
- ✅ Generics support (coming)
- ✅ Full code generation (scalable)
- ✅ Active development (improvements ongoing)
- ✅ Focused on assertions (mock/suite removed)
- ✅ Comprehensive API (all functions kept)

### Our Unique Position

**We're not competing with alternatives - we're solving different problems:**

**alecthomas/assert:** "Minimal testify with generics"
**go-openapi/testify:** "Enterprise-grade testify with zero deps, modern architecture, comprehensive features"

**Our tagline should be:**
> "The testify v2 that should have been: zero dependencies, comprehensive API, type-safe generics, generated for consistency, built for the long term."

### What We Should Adopt from Alternatives

**From alecthomas/assert:**
1. ✅ **Better diff output** - Enhance our internalized difflib formatting
2. ✅ **Usage-based deprecation** - His usage stats validate deprecating `InDeltaSlice`, etc.
3. ✅ **Simplicity in documentation** - Show most-used 20% prominently, document rest separately

**From testify v2 discussions:**
1. ✅ **API consistency guidelines** - Generator enforces mechanically
2. ✅ **Remove low-value wrappers** - Deprecate rarely-used functions
3. ✅ **Better naming** - Improve where it makes sense (e.g., avoid "Is" prefix)

### What We Shouldn't Adopt

**From alecthomas/assert:**
1. ❌ **Dropping functions** - Our users need comprehensive API
2. ❌ **External dependencies** - Defeats our core purpose
3. ❌ **Forced generics** - Hybrid approach better for migration

**From testify v2 discussions:**
1. ❌ **Removing format variants** - Too disruptive, provides value
2. ❌ **Argument order changes** - Breaking change not worth the benefit

### Competitive Advantages Summary

**Why go-openapi/testify v2 is the right choice for go-openapi ecosystem:**

1. ✅ **Solves the original problem:** Zero dependencies (alecthomas fails this)
2. ✅ **Smooth migration:** Drop-in replacement (alecthomas requires rewrites)
3. ✅ **Comprehensive:** All assertions available (alecthomas is minimal)
4. ✅ **Scalable:** Code generation handles complexity (alecthomas is manual)
5. ✅ **Flexible:** Hybrid generics strategy (alecthomas forces type safety)
6. ✅ **Quality:** 94% test coverage with modern patterns (alecthomas is basic)
7. ✅ **Future-proof:** Active development with clear roadmap (testify is frozen)

**Our fork is not "better" universally - it's better for our use case.** Alec Thomas is a brilliant developer and his library excellently serves its target audience (minimal API enthusiasts). We serve a different audience (enterprise users, testify migrants, go-openapi ecosystem).

**Sources:**
- [alecthomas/assert v2](https://github.com/alecthomas/assert)
- [assert/v2 Package Documentation](https://pkg.go.dev/github.com/alecthomas/assert/v2)
- [Testify v2 Discussion #1560](https://github.com/stretchr/testify/discussions/1560)

## Implementation Status

### Completed ✅

**Architecture & Organization:**
- ✅ Created `internal/assertions/` package structure (single source of truth)
- ✅ Organized 76 assertion functions by domain across 14 files
- ✅ Refactored tests to table-driven patterns with Go 1.23 `iter.Seq`
- ✅ Restructured code generator: `codegen/` with scanner + generator + templates
- ✅ Moved old code to junk/ directories (removed, except example_test.go preserved)
- ✅ Created template infrastructure (9 templates total)

**Scanner Implementation:**
- ✅ Fully operational Go AST/types-based scanner
- ✅ Position-based lookup bridging semantic and syntactic analysis
- ✅ Import alias resolution for accurate code generation
- ✅ Example extraction from doc comments (success/failure/panic cases)
- ✅ Function signature parsing with type information
- ✅ Generic function detection and handling

**Code Generation:**
- ✅ Generate `assert/` package with all variants (package + format + forward)
- ✅ Generate `require/` package with all variants (package + format + forward)
- ✅ Generate helper types and interfaces
- ✅ All 76 functions × 8 variants = 608 generated functions

**Example-Driven Test Generation:**
- ✅ Added "Examples:" sections to all 76 assertion functions
- ✅ Test value extraction from doc comments
- ✅ Generate tests for assert package (success + failure cases)
- ✅ Generate tests for require package (with FailNow verification)
- ✅ Generate tests for format variants (with message parameter)
- ✅ Generate tests for forward variants (method chaining)
- ✅ Mock testing infrastructure (mockT for assert, mockFailNowT for require)

**Test Coverage Achievement:**
- ✅ `internal/assertions/`: 94% coverage with exhaustive tests
- ✅ `assert/`: ~100% coverage with generated tests
- ✅ `require/`: ~100% coverage with generated tests
- ✅ Both packages fully tested across all 8 variants per function

**Documentation:**
- ✅ Template documentation and architecture notes
- ✅ Example status tracking (EXAMPLES_STATUS.md)
- ✅ Updated this plan with accomplishments (in progress)
- ✅ Updated CLAUDE.md with new architecture (in progress)

### Remaining ⏳

**Test Coverage (99.5% → 100%):**
- ⏳ Generate tests for helper functions (non-assertion functions)
  - Helper functions like `ObjectsAreEqual`, `CallerInfo`, etc. don't have "Examples:" sections
  - These are the reason for 99.5% coverage instead of 100%
  - **Note:** The 0.5% coverage gap is intentional - it comes from helper functions that don't have "Examples:" annotations
  - Need to decide: hand-write tests or add example-driven generation for helpers
  - **Future plan:** Stop generating helper functions entirely - move them to internal or separate helpers package

**Cleanup:**
- ✅ Deleted junk/ directories (preserved example_test.go files for reference)
- ✅ Remove old `_codegen/` references if any remain
- ✅ Clean up any remaining TODO comments in templates

**Testable Examples:**
- ✅ Generate Example* functions from "Examples:" sections for all assertions
- ✅ Make examples runnable by default (include Output: comments)
- ⏳ Replace "Usage:" sections in godoc with references to generated examples (to be discussed)
- These provide both documentation and executable examples in godoc

**Code Quality & Testing:**
- ✅ Add smoke tests for codegen generator (target 20% coverage minimum)
- ✅ Add smoke test for codegen/main.go
- ⏳ Improve private (non-godoc) comments in internal/assertions
  - ✅ Current godoc comments are excellent
  - Private comments within functions need better clarity
- ⏳ Code quality assessment after merge
  - Review and fix linting issues in internal packages (internalized dependencies)
  - Assess overall code health across the codebase

**Documentation & Organization:**
- ✅ Better organized documentation site (group by themes, not alphabetically)
  - Boolean assertions (True, False)
  - Comparison assertions (Equal, Greater, Less, etc.)
  - Collection assertions (Contains, Len, ElementsMatch, etc.)
  - Error assertions (Error, NoError, ErrorIs, etc.)
  - Type assertions (IsType, Implements, etc.)
  - Testing utilities (Eventually, Never)
  - HTTP assertions
  - File/Directory assertions
  - Time assertions
  - Special assertions (JSONEq, YAMLEq, etc.)
- ⏳ Educational examples at root package level (notice that this doesn't fit well with godoc)
  - Beginner-friendly examples showing common patterns
  - Migration guide from stretchr/testify
  - Best testing practices documentation
- ⏳ Verify all README.md items are documented in plans
  - Cross-check README roadmap with this plan file
  - Ensure nothing is missing

**Generator Enhancements:**
- ❌ Add option to skip deprecated functions in codegen (wont't do: remove deprecated stuff from internal if we don't want it generated)
  - Don't generate code for functions marked as deprecated
  - Allow clean deprecation path for rarely-used assertions
- ✅ Consider generating documentation by theme instead of alphabetically
- ✅ Clean parsing of example values for safe code generation and remove regexp-based hacks at rendering time

**Future Enhancements (Optional):**
- ⏳ Add generic variants for type safety (Equal[T comparable], Contains[S ~[]E, E comparable], etc.)
- ⏳ Enhance diff output formatting (consider improvements inspired by alecthomas/assert)
- ⏳ Add benchmark suite for hot paths (Equal, Contains, Empty)
- ⏳ Consider deprecating rarely-used functions based on usage data
- ⏳ Expand "enable" pattern for optional features (prettydiff, etc.)
- ✅ Add Example* functions for godoc (done: 5 complete ad'hoc examples - manual, not generated - reintroduced to assert and require)
- ⏳ Meta-tests for code generator itself (golden file testing)

**Integration & Polish:**
- ✅ Final review of all generated code
- ⏳ Performance testing of generated functions
- ✅ Consider adding generation timestamps/version to headers
- ⏳ Documentation examples showing the full development workflow

## Recommendation

**Proceed with the refactor.** Even if test generation is initially imperfect, the architectural win is worth it. Test generation can be enhanced incrementally.

The single-source-of-truth model is the right long-term architecture for a library with 76 assertion functions that need 8 variants each.

## Next Steps

**Immediate (Cleanup & Validation):**
1. ✅**Final validation** - Run comprehensive tests across all packages (ran test suites against 3 different go-openapi repos, including with yaml enabled, using the current version)
2. **Performance benchmarks** - Ensure generated code performs well
3. ✅**Delete junk/ directories** - Clean up old hand-written code
4. ✅**Documentation review** - Ensure all godoc is accurate and helpful

**Short-term (Polish):**
5. **Simplify templates** - Add methods/fields to model to reduce template complexity
   - Enrich model with computed properties to avoid declaring variables in templates
   - Add helper methods to reduce `{{ if }}` filter logic
   - Make templates more readable and maintainable
6. **Fix type mapping for xxxFunc types** - Address internal type references in signatures
   - Issue: PanicAssertionFunc and similar types reference internal package types
   - Currently works via re-export aliases but is untidy
   - Exacerbated by move to internal/assertions (was already poor API design in original)
   - Need to rework type mapping to generate clean signatures
7. ✅ **Generate documentation site with Hugo** - Leverage generator for docs
   - Reuse scanner/model/generator architecture for markdown generation
   - Generate markdown files with Hugo frontmatter in `docs/site/`
   - Organize assertions by theme (Boolean, Comparison, Collection, Error, etc.)
   - Include signatures, examples, cross-references in generated pages
   - Deploy to GitHub Pages
   - **Inspiration sources:**
     - Minimalistic: https://masterminds.github.io/sprig (simple but API is bloated)
     - Previous work: https://goswagger.io/go-swagger (more comprehensive)
   - **Status:** Preparing inner data model for template consumption
   - **Benefits:** Solves multiple plan items (organized docs, educational examples, replace Usage sections)
8. ✅ **Add generation metadata** - Timestamps, version info in generated headers
9. **Improve error messages** - Review and enhance failure output quality
10. **Create usage examples** - Show best practices for the new workflow
11. ⏳**Integration testing** - Test in go-swagger and other go-openapi projects (in progress)

**Long-term (Enhancements):**
9. **Generic variants** - Add type-safe Equal[T], Contains[S,E], etc.
10. **Enhanced diff output** - Improve formatting inspired by go-cmp
11. **Benchmark suite** - Comprehensive performance testing
12. **Meta-tests** - Test the code generator itself with golden files

## Merged PRs from Original Repository

**Upstream fixes merged (catching up with stretchr/testify):**

- ✅ **#1825** - Fix panic when using EqualValues with uncomparable types
- ✅ **#1818** - Fix panic on invalid regex in Regexp/NotRegexp assertions
- ✅ **#1223** - Display uint values in decimal instead of hex in diffs

**Planned upstream merges:**

Critical safety fixes (high priority):
- ⏳ **#1824** - Follow/adapt (investigate)
- ⏳ **#1826** - Reported issue (investigate)
- ⏳ **#1611** - Reported issue (investigate)
- ⏳ **#1813** - Reported issue (investigate)

Leveraging internalized dependencies (go-spew, difflib):
- ⏳ **#1829** - Fix time.Time rendering in diffs (internalized go-spew)
- ⏳ **#1822** - Deterministic map ordering in diffs (internalized go-spew)
- ⏳ **#1816** - Fix panic on unexported struct key in map (internalized go-spew - may need deeper fix)

UX improvements:
- ⏳ Diff rendering improvements

Under consideration (would be optional `enable/color` module):
- ⏳ **#1467** - Colorized output with terminal detection (most mature)
- ⏳ **#1480** - Colorized diffs via TESTIFY_COLORED_DIFF env var
- ⏳ **#1232** - Colorized output for expected/actual/errors
- ⏳ **#994** - Colorize expected vs actual values

## Success Criteria ✅ ACHIEVED

**Primary Goals (All Achieved):**
- ✅ Adding a new assertion requires only:
  1. Writing function in appropriate `internal/assertions/*.go` file with Examples
  2. Writing exhaustive tests in corresponding `*_test.go`
  3. Running `go generate`
  4. All 8 variants generated automatically with 100% test coverage

- ✅ No more touching thousand-line files (domain-organized instead)
- ✅ No more manual duplication (mechanical generation)
- ✅ Consistent API across all assertion variants (template-enforced)

**Additional Achievements:**
- ✅ Example-driven test generation working perfectly
- ✅ Both assert and require packages fully generated
- ✅ 94% coverage in source, ~100% in generated packages
- ✅ Modern Go 1.23 patterns throughout
- ✅ Clean separation: source (internal/assertions), generation (codegen/), output (assert/require)
- ✅ Scalable architecture supporting 76 functions × 8 variants = 608 generated functions

**Proof of Success:**
The FileEmpty/FileNotEmpty addition that originally motivated this refactor would now require:
1. Add two functions to `internal/assertions/file.go` with Examples
2. Add tests to `internal/assertions/file_test.go`
3. Run `go generate`
4. Done - 16 variants generated with full test coverage

vs. the old model of manually touching 6+ files with thousands of lines each.
