## Why Choose Testify?

For Go developers who embrace Go's philosophy of simplicity, explicitness, and building on standards, testify is the
natural choice. It brings powerful testing capabilities while staying true to what makes Go great.

### 1. Aligned with Go Values

**Testify embodies Go's design philosophy** applied to testing:

- **Simplicity**: Just functions and assertions, no framework complexity
- **Explicitness**: Clear, direct assertion calls; no composition magic
- **Standard library first**: Builds on `testing.T`, works with `go test`
- **Minimal abstraction**: Solves the assertion problem, nothing more

If you chose Go for these values, testify extends them to your test suite.

### 2. Immediate Adoption

**No learning curve.** If you know Go testing, you already know how to use testify:

```go
import (
	"testing"

	"gotest.tools/assert"
)

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	assert.Equal(t, 5, result) // That's it. You're done.
}
```

No new DSL. No framework concepts. No special runners. Just better assertions.

### 3. Type Safety and Performance

**Catch errors at compile time**, not runtime:

```go
// Generic assertions provide type safety
assert.EqualT(t, 42, Calculate())           // OK: both int
assert.EqualT(t, "42", Calculate())         // Compile error: type mismatch

// Performance benefits
assert.ElementsMatchT(t, expected, actual)  // 21-81x faster than reflection
```

With **82 non-generic assertions** and **38 generic variants**, you have powerful tools that don't compromise on safety
or speed.

### 4. Zero External Dependencies

**Your tests are isolated** from the dependency ecosystem:

- No `go.mod` bloat from test dependencies
- No security vulnerabilities in transient dependencies
- No version conflicts with your application code
- Stable, predictable behavior across Go versions

Everything is internalized and self-contained. When you import testify, that's all you get: no surprises.

### 5. Comprehensive Assertion Library

**120 assertion functions** organized into 18 logical domains:

- **Equality**: `Equal`, `NotEqual`, `EqualValues`, `Same`, `NotSame`
- **Collections**: `Contains`, `ElementsMatch`, `Subset`, `Len`
- **Errors**: `Error`, `NoError`, `ErrorIs`, `ErrorAs`, `ErrorContains`
- **Comparisons**: `Greater`, `Less`, `GreaterOrEqual`, `LessOrEqual`
- **Types**: `IsType`, `Implements`, `Zero`, `NotZero`
- **HTTP**: `HTTPSuccess`, `HTTPError`, `HTTPStatusCode`, `HTTPBodyContains`
- **Files**: `FileExists`, `DirExists`, `FileEmpty`
- **Panic**: `Panics`, `NotPanics`, `PanicsWithValue`
- **Async**: `Eventually`, `Never`, `EventuallyWithT`
- And more: boolean, number, string, time, JSON, YAML...

Every assertion is available in **8 variants**:

```go
// Package-level functions
assert.Equal(t, expected, actual)
assert.Equalf(t, expected, actual, "user %s", name)

// Forward methods (for chaining)
a := assert.New(t)
a.Equal(expected, actual)
a.Equalf(expected, actual, "user %s", name)

// Fatal variants (stop test immediately)
require.Equal(t, expected, actual)
require.Equalf(t, expected, actual, "user %s", name)

// Fatal forward methods
r := require.New(t)
r.Equal(expected, actual)
r.Equalf(expected, actual, "user %s", name)
```

**Total: 818 generated functions** from a single source of truth, ensuring mechanical consistency across the entire API.

### 6. Standard Go Patterns

**Use the patterns you already know**:

```go
import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

// Table-driven tests with Go 1.23 iterators
func TestMathOperations(t *testing.T) {
	t.Parallel()

	for c := range testCases() {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			result := Calculate(c.input)
			assert.EqualT(t, c.expected, result)
		})
	}
}

// Subtests for organization
func TestUserAPI(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		user := CreateUser("alice@example.com")
		assert.NotNil(t, user)
	})

	t.Run("validation", func(t *testing.T) {
		err := ValidateUser(invalidUser)
		assert.Error(t, err)
	})
}

// Standard parallelization
func TestParallel(t *testing.T) {
	t.Parallel() // Just works

	// Test runs in parallel with other parallel tests
}
```

No framework-specific constructs. No special setup nodes. Just Go.

### 7. IDE-Friendly

**Your IDE already understands testify**:

- Full autocomplete for all assertions
- Jump to definition works perfectly
- Inline documentation in tooltips
- Refactoring support (rename, extract, etc.)
- No framework-specific plugins required

Because testify uses standard Go packages and functions, every IDE feature just works.

### 8. Clear, Actionable Error Messages

**Failures tell you exactly what went wrong**:

```
Error:      Not equal:
            expected: []int{1, 2, 3}
            actual  : []int{1, 2, 4}

            Diff:
            --- Expected
            +++ Actual
            @@ -1,3 +1,3 @@
             ([]int) (len=3) {
              (int) 1,
              (int) 2,
            - (int) 3
            + (int) 4
             }
Test:       TestSliceEquality
```

No cryptic matcher errors. No confusing matcher composition failures. Just clear diffs and context.

---

## When to Choose What

### Choose Testify (Assertion-Style) When:

✅ **You embrace Go's philosophy**: Simplicity, explicitness, standard library first

✅ **You value zero dependencies**: No external packages, no supply chain concerns

✅ **You want immediate productivity**: No learning curve, works with `go test`

✅ **Type safety matters**: 38 generic assertions with compile-time checking

✅ **You're building libraries**: Minimize transitive dependencies for users

✅ **You want IDE integration**: Standard Go code, all tools work automatically

**If you chose Go because you value its design philosophy, assertion-style testing is the natural extension of those
values to your test suite.**

### Consider BDD Frameworks When:

- **Your team prioritizes narrative specifications**: Tests as documentation for stakeholders
- **You want framework-managed organization**: Hierarchical structure (Describe/Context/It)
- **You're comfortable with framework-specific tooling**: `ginkgo` CLI, specialized runners
- **You need advanced async patterns**: Sophisticated Eventually/Consistently support
- **You value BDD methodology**: Behavior-driven development is a team practice

**Both styles are valid in the broader software industry**; they optimize for different priorities. For Go developers
specifically, the question is whether your testing style aligns with the values that drew you to Go in the first place.

---

## Get Started

Ready to enhance your Go tests?

```bash
go get github.com/go-openapi/testify/v2
```

Then in your tests:

```go
import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
)

func TestYourCode(t *testing.T) {
	result := YourFunction()
	require.NotNil(t, result)
	assert.Equal(t, expected, result)
}
```

That's it. No configuration. No setup. No framework initialization. Just better assertions.

---

## Philosophy in Action

The go-openapi project chose testify (and created this fork) because we believe testing should reflect Go's core
values:

- **Simplicity over sophistication**: Tests are code, not specifications
- **Enhancement over replacement**: Build on `testing.T`, don't replace it
- **Standards over frameworks**: Work with `go test`, not against it
- **Explicitness over magic**: Direct assertions, no hidden behavior
- **Dependencies that don't add dependencies**: Zero external packages

This fork takes the assertion-style approach even further than the original stretchr/testify,
internalizing all dependencies to provide a completely self-contained testing toolkit,
embracing modern go and generics.

**One import. Zero dependencies. Pure Go values.**

For Go developers who believe in Go's philosophy, assertion-style testing isn't just a preference: it's the idiomatic
approach. Testify makes that approach powerful.

---
