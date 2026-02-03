# Error-Aware Assertions: The (T, error) Pattern

**Status**: Planning
**Priority**: High
**Date**: 2026-01-23

## Context

Discussion emerged from competitive analysis with Ginkgo/Gomega about handling Go's idiomatic `(value, error)` return pattern in assertions. This pattern appears repeatedly across:
- JSON unmarshaling
- Async operations (Eventually)
- HTTP requests
- Database queries
- Any Go API following `(T, error)` convention

## Problem Statement

Current testify requires manual error handling before assertions:

```go
// Current: verbose and awkward
var actual MyStruct
err := json.Unmarshal(jsonBytes, &actual)
require.NoError(t, err)
assert.Equal(t, expected, actual)

// Current: EventuallyT limited to bool
assert.Eventually(t, func() bool {
    user, err := FetchUser(123)
    return err == nil && user.Status == "active"
}, timeout, tick)
```

## Proposed Solution

Leverage generics to create assertion helpers that unwrap `(T, error)` and enable chaining:

### Core Pattern: result[T] Helper

```go
// Internal helper type (unexported)
type result[T any] struct {
    t TestingT
    value T
    succeeded bool
}

// Implement chainable assertions
func (r *result[T]) Equal(expected T, msgAndArgs ...any) bool {
    if !r.succeeded {
        return false
    }
    return EqualT(r.t, expected, r.value, msgAndArgs...)
}

func (r *result[T]) NotNil(msgAndArgs ...any) bool {
    if !r.succeeded {
        return false
    }
    return NotNil(r.t, r.value, msgAndArgs...)
}

func (r *result[T]) GreaterOrEqual(threshold T, msgAndArgs ...any) bool {
    if !r.succeeded {
        return false
    }
    return GreaterOrEqualT(r.t, r.value, threshold, msgAndArgs...)
}

// Add other assertions as applicable based on type constraints
```

## Proposed Features

### Phase 1: Core Implementations (High Priority)

#### 1. UnmarshalJSONAsT - JSON Domain

**Motivation**: Frequent pattern in API testing

```go
// domain: json

// UnmarshalJSONAsT unmarshals JSON data and returns an assertion helper for the result.
//
// Examples:
//   success: []byte(`{"name":"Alice"}`), User{Name: "Alice"}
//   failure: []byte(`invalid json`), User{}
func UnmarshalJSONAsT[T any](t TestingT, data []byte, msgAndArgs ...any) *result[T] {
    var value T
    if err := json.Unmarshal(data, &value); err != nil {
        Fail(t, fmt.Sprintf("JSON unmarshal failed: %v", err), msgAndArgs...)
        return &result[T]{t: t, value: value, succeeded: false}
    }
    return &result[T]{t: t, value: value, succeeded: true}
}
```

**Usage:**
```go
// Simple case
assert.UnmarshalJSONAsT[User](t, jsonBytes).Equal(expectedUser)

// With custom message
assert.UnmarshalJSONAsT[User](t, jsonBytes).Equal(expectedUser, "user from API")

// Chain multiple assertions
user := assert.UnmarshalJSONAsT[User](t, jsonBytes)
assert.Equal(t, "Alice", user.value.Name)
assert.Greater(t, user.value.Age, 0)
```

**Benefits:**
- ✅ Solves immediate need for JSON testing
- ✅ Establishes the pattern
- ✅ Type-safe with generics
- ✅ Clear error messages
- ✅ Reduces boilerplate

#### 2. EventuallyT - Async Domain (Enhanced)

**Motivation**: Handle `(T, error)` in async operations naturally

```go
// domain: testing

// EventuallyT repeatedly calls f until it returns a value without error,
// or waitFor timeout expires. Returns an assertion helper for the eventual result.
//
// Examples:
//   success: func() (int, error) { return 42, nil }, 1*time.Second, 100*time.Millisecond, 42
//   failure: func() (int, error) { return 0, errors.New("failed") }, 100*time.Millisecond, 10*time.Millisecond, 42
func EventuallyT[T any](
    t TestingT,
    f func() (T, error),
    waitFor, tick time.Duration,
    msgAndArgs ...any,
) *result[T] {
    timer := time.NewTimer(waitFor)
    defer timer.Stop()

    ticker := time.NewTicker(tick)
    defer ticker.Stop()

    for {
        select {
        case <-timer.C:
            var zero T
            return &result[T]{
                t: t,
                value: zero,
                succeeded: Fail(t, "Eventually timed out", msgAndArgs...),
            }
        case <-ticker.C:
            value, err := f()
            if err == nil {
                return &result[T]{t: t, value: value, succeeded: true}
            }
        }
    }
}
```

**Usage:**
```go
// Natural (T, error) handling
assert.EventuallyT(t,
    func() (int, error) { return client.FetchCount() },
    5*time.Second, 100*time.Millisecond,
).GreaterOrEqual(17)

// Works with any type
assert.EventuallyT(t,
    func() (*User, error) { return db.FindUser(id) },
    timeout, tick,
).Equal(expectedUser)

// Complex assertions
assert.EventuallyT(t, fetchStatus, timeout, tick).Equal("ready")
```

**Benefits:**
- ✅ Handles idiomatic Go `(T, error)` pattern
- ✅ Type-safe result assertions
- ✅ Cleaner than bool closures
- ✅ Composable with all assertions

#### 3. EventuallyWithContextT - Context Support

**Motivation**: Modern Go emphasizes context.Context everywhere

```go
// domain: testing

// EventuallyWithContextT like EventuallyT but respects context cancellation.
//
// Examples:
//   success: context.Background(), func(ctx) (int, error) { return 42, nil }, 100*time.Millisecond, 42
func EventuallyWithContextT[T any](
    ctx context.Context,
    t TestingT,
    f func(context.Context) (T, error),
    tick time.Duration,
    msgAndArgs ...any,
) *result[T] {
    ticker := time.NewTicker(tick)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            var zero T
            return &result[T]{
                t: t,
                value: zero,
                succeeded: Fail(t, "Eventually cancelled: "+ctx.Err().Error(), msgAndArgs...),
            }
        case <-ticker.C:
            value, err := f(ctx)
            if err == nil {
                return &result[T]{t: t, value: value, succeeded: true}
            }
        }
    }
}
```

**Usage:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

assert.EventuallyWithContextT(ctx, t,
    func(ctx context.Context) (*User, error) {
        return client.FetchUserWithContext(ctx, id)
    },
    100*time.Millisecond,
).Equal(expectedUser)
```

**Benefits:**
- ✅ Idiomatic context cancellation
- ✅ Better resource cleanup
- ✅ Integrates with context deadlines
- ✅ Modern Go best practices

### Phase 2: Evaluate Expansion (Medium Priority)

Based on usage patterns and demand, consider:

#### YAML Support (if YAML enabled)

```go
// domain: yaml

func UnmarshalYAMLAsT[T any](t TestingT, data []byte, msgAndArgs ...any) *result[T]
```

#### HTTP Support

```go
// domain: http

func GetJSONT[T any](t TestingT, url string, msgAndArgs ...any) *result[T]
```

**Decision criteria:**
- Wait for user requests
- Ensure clear use cases
- Maintain zero-dependency principle

### Phase 3: Custom Assertion Documentation

Provide patterns for users to create domain-specific helpers:

```go
// Example: Database assertions
func (db *TestDB) FindUserT(t *testing.T, id int) *testify.Result[User] {
    user, err := db.FindUser(id)
    if err != nil {
        testify.Fail(t, fmt.Sprintf("FindUser failed: %v", err))
        return &testify.Result[User]{succeeded: false}
    }
    return &testify.Result[User]{value: user, succeeded: true}
}

// Usage
db.FindUserT(t, 123).Equal(expectedUser)
```

## Implementation Strategy

### Step 1: Design Internal result[T] Type

```go
// internal/assertions/result.go

type result[T any] struct {
    t        TestingT
    value    T
    succeeded bool
}

// Implement methods for all applicable assertions
// Based on type constraints (comparable, Ordered, etc.)
```

**Key decisions:**
- Should `result[T]` be exported or internal?
  - **Recommendation**: Export as `Result[T]` for extensibility
- Which assertions to implement?
  - **Recommendation**: All applicable based on type constraints
- How to handle failed state?
  - **Recommendation**: `succeeded` flag, methods check before asserting

### Step 2: Implement UnmarshalJSONAsT

- Create in `internal/assertions/json.go`
- Add comprehensive tests in `internal/assertions/json_test.go`
- Generate variants: `assert.UnmarshalJSONAsT`, `require.UnmarshalJSONAsT`, format variants
- Document with examples
- Test with go-swagger use cases

### Step 3: Enhance Eventually

- Update `internal/assertions/testing.go`
- Add `EventuallyT` alongside existing `Eventually`
- Add `EventuallyWithContextT`
- Maintain backward compatibility (keep `Eventually`)
- Fix existing goroutine leaks (already done)
- Comprehensive async tests

### Step 4: Documentation

Update documentation:
- Add to [Generics Guide](../usage/GENERICS.md) - new section on result pattern
- Add to [Examples](../usage/EXAMPLES.md) - JSON and async examples
- Add to [API Reference](../api/) - json and testing domains
- Create "Custom Assertions" guide showing extensibility

### Step 5: Code Generation

Update `codegen/` to handle:
- `Result[T]` return types
- Method generation for `Result[T]`
- Test generation for new functions
- Documentation generation

## Design Principles

1. **Type Safety First**: Leverage generics for compile-time safety
2. **Zero Dependencies**: No external packages
3. **Idiomatic Go**: Follow `(T, error)` conventions
4. **Composability**: Chain with existing assertions
5. **Simplicity**: No DSL, just function calls
6. **Backward Compatibility**: Keep existing APIs working

## Success Criteria

- ✅ `UnmarshalJSONAsT` successfully replaces verbose unmarshal + assert patterns
- ✅ `EventuallyT` handles `(T, error)` naturally
- ✅ `EventuallyWithContextT` respects context cancellation
- ✅ go-swagger team adopts in real tests
- ✅ Documentation clear and comprehensive
- ✅ Zero performance regression
- ✅ All variants generated correctly

## Alternatives Considered

### Alternative 1: General-Purpose Unwrap

```go
// Too abstract, rejected
func Unwrap[T any](t TestingT, value T, err error) *Result[T]

// Usage
assert.Unwrap(t, json.Unmarshal(data, &actual)).Equal(expected, actual)
```

**Rejected because:**
- Too clever/abstract
- Harder to document
- Less clear intent
- Domain-specific functions are more discoverable

### Alternative 2: Matcher-Style DSL

```go
// Gomega-style, rejected
assert.Eventually(t, fetchUser).Should().Equal(expected)
```

**Rejected because:**
- Conflicts with "no DSL" principle
- Adds complexity
- Not idiomatic testify style

### Alternative 3: Multiple Return Values

```go
// Support (T1, T2, error), rejected for now
func EventuallyT2[T1, T2 any](f func() (T1, T2, error)) (*Result2[T1, T2])
```

**Rejected because:**
- YAGNI - no clear use case yet
- Can revisit if demand emerges
- Adds significant complexity

## Open Questions

1. **Should Result[T] be exported?**
   - Pro: Enables user extensibility
   - Con: Locks us into public API
   - **Decision**: Export as `Result[T]` for extensibility

2. **Should we deprecate old Eventually?**
   - Pro: Simplify API surface
   - Con: Breaking change
   - **Decision**: Keep both, document preference for `EventuallyT`

3. **What assertions should Result[T] support?**
   - All applicable based on type constraints
   - Document which are available for which types
   - Use generics constraints properly

4. **How to handle require vs assert variants?**
   - Same pattern: `require.UnmarshalJSONAsT` returns `Result[T]`
   - Result methods call appropriate require/assert internally
   - **Question**: How does Result[T] know if it's from require or assert?
   - **Answer**: Result methods take `t TestingT` and call underlying functions

## Related Work

- [Generics Guide](../usage/GENERICS.md) - Existing generic assertions
- [COMPETITIVE_ANALYSIS.md](./COMPETITIVE_ANALYSIS.md) - Comparison with Gomega
- [Benchmarks](../project/maintainers/BENCHMARKS.md) - Performance of generics

## Next Steps

1. ✅ Document the plan (this file)
2. ⏳ Design `Result[T]` type and method set
3. ⏳ Implement `UnmarshalJSONAsT` as proof of concept
4. ⏳ Test with real go-swagger use cases
5. ⏳ Implement `EventuallyT` and `EventuallyWithContextT`
6. ⏳ Update code generator
7. ⏳ Write comprehensive documentation
8. ⏳ Gather feedback, iterate

---

## Notes from Discussion

**Fred's feedback:**
- ✅ Likes the (T, error) wrapping pattern
- ✅ Already wanted this for JSON assertions
- ✅ Context support is brilliant
- ✅ Agrees with surgical enhancement approach
- ✅ Focus on 3 functions: UnmarshalJSONAsT, EventuallyT, EventuallyWithContextT

**Strategic alignment:**
- Maintains testify's identity: "simple, type-safe, zero-dependency"
- Leverages generic advantage vs Gomega
- Solves real Go patterns idiomatically
- Minimal API surface (3 functions to start)
- Room for organic growth based on usage

**What to avoid:**
- ❌ Full Gomega-style async DSL
- ❌ `Consistently` (wait for demand)
- ❌ Multiple value returns (YAGNI)
- ❌ General-purpose `Unwrap` (too abstract)
- ❌ Feature creep beyond proven needs
