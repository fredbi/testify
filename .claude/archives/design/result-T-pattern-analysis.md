# Result[T] Pattern Analysis

**Date:** 2026-01-30
**Context:** Design conversation about whether a general-purpose `Result[T]` pattern is worth implementing

## Background

The v3 roadmap proposed a `Result[T]` generic type to wrap `(T, error)` returns and enable chainable, type-safe assertions:

```go
type Result[T any] struct {
    t         TestingT
    value     T
    succeeded bool
}

// Usage vision:
assert.UnmarshalJSONAsT[User](t, data).Equal(expectedUser)
```

This came from two separate threads that got conflated:
1. **Assertions extensibility + generics**: Making the `Assertions` type usable with generic functions. Solved pragmatically with `a.T` escape hatch (`EqualT(a.T, expected, actual)`).
2. **Error-aware assertions**: Wrapping `(T, error)` returns so error checks and value assertions compose fluently. This is where `Result[T]` was proposed.

## The Go Constraint

Go does not allow injecting extra arguments before multi-return values:

```go
// This works:
doIt(json.Marshal(data))  // doIt(b []byte, err error)

// This does NOT compile:
doIt(t, json.Marshal(data))  // doIt(t T, b []byte, err error)
```

This means you can't write `JSONMarshalAsT(t, json.Marshal(data))` directly, which was the initial intuition.

## Approaches Evaluated

### 1. Result as `func() (T, error)` with compound assertions

```go
type Result[Object any] func() (Object, error)

func NoErrorAndEqualT[Object comparable](t T, expected Object, result Result[Object], ...) bool

// Usage:
assert.NoErrorAndEqual(t, expected, func() ([]byte, error) { return json.Marshal(actual) })
```

**Verdict: Worse than the problem it solves.**
- Closure syntax is noisy
- Longer than the 2-line alternative
- Combinatorial explosion: need `NoErrorAndX` for every assertion X
- No ergonomic gain

### 2. Traditional 2-line Go

```go
jazon, err := json.Marshal(actual)
require.NoError(t, err)
assert.EqualT(t, expected, jazon)
```

**Verdict: This is the Go way.** Idiomatic, reads top to bottom, no new concepts.

### 3. Domain-specific assertions

```go
assert.JSONMarshalAs(t, expected, actual)
```

**Verdict: Best UX for known operations.** Encodes *intent* ("assert this marshals to that JSON") rather than *mechanism* ("marshal, check error, compare bytes"). Not general, but doesn't need to be.

### 4. Original Result[T] with methods (from v3 roadmap)

```go
assert.UnmarshalJSONAsT[User](t, data).Equal(expectedUser)
```

**Verdict: Marginal gain over 2 lines of plain Go.** Adds a new type and mental model. Codegen complexity cost is real. The savings over the traditional approach are one line at best.

## Conclusion

**Do not implement a general-purpose `Result[T]` pattern.** The rationale:

1. **Go works against it.** The explicit `(T, error)` pattern is a language-level design choice. Abstracting over it adds ceremony that competes with the straightforward 2-line idiomatic alternative. This project's spirit is not to work against Go or add abstractions for abstraction's sake.

2. **Domain-specific assertions are the better UX.** `JSONMarshalAs`, `JSONUnmarshalAs`, `YAMLMarshalAs`, the existing HTTP assertions -- these express intent directly and are more readable than any general wrapping mechanism. The set of operations worth wrapping is small and known.

3. **The combinatorial problem kills generality.** Any general `Result[T]` either needs `NoErrorAndX` for every assertion (explosion), or methods on the Result type (new mental model + codegen complexity). Neither is worth it.

4. **One exception: `Eventually`.** The `Eventually` use case is fundamentally different because the error is part of a retry loop, not a precondition to check once. The framework needs to *own* the error handling (retry vs give up). If `EventuallyT[T]` is ever implemented, a result-like return may make sense there specifically -- but that's a localized design decision, not a general pattern.

## Recommendation

- Keep domain-specific assertions (`JSONMarshalAs`, `JSONUnmarshalAs`, `YAMLMarshalAs`, HTTP assertions) as the primary approach for error-producing operations
- Check that existing HTTP assertions follow a consistent UX pattern
- Revisit result wrapping only in the context of `EventuallyT` if/when that gets implemented
- Remove `Result[T]` from the v3 roadmap as a general foundation; it's not needed
