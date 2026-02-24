# JSON Extension Analysis

**Date:** 2026-01-27
**Topic:** Should we make JSON assertions pluggable like YAML?

## Current State

JSON assertions in testify use `encoding/json` from the Go standard library:
- `JSONEq(t, expected, actual)` compares JSON strings semantically
- `JSONEqBytes(t, expected, actual)` compares JSON byte slices
- `JSONEqT[E, A Text](t, expected, actual)` generic variant
- No abstraction layer or pluggability mechanism
- Works everywhere, zero dependencies

## The Question

Should we generalize the pluggable serializer concept used for YAML to JSON as well?

This would allow users to swap in alternative JSON libraries like:
- `jsoniter` (faster, more compatible)
- `sonic` (extremely fast on x86-64)
- `goccy/go-json` (faster, better errors)
- Custom JSON implementations

## Arguments For Making JSON Pluggable

### 1. Consistency
If YAML is pluggable, JSON being pluggable creates symmetry:
- Predictable API patterns
- Users expect similar extensibility
- Architecture feels complete

### 2. Future-Proofing
Potential use cases:
- Testing systems that use `jsoniter` for compatibility (some projects rely on jsoniter-specific behaviors)
- Need for specific number handling (big decimals, precision)
- Want better error messages in failing tests
- Large test suites where every microsecond counts

### 3. Low Implementation Cost
Pattern already exists from YAML:

```go
import "encoding/json"

// internal/assertions/enable/json/json.go
var (
	enableJSONMarshal   = json.Marshal   // default to stdlib
	enableJSONUnmarshal = json.Unmarshal // default to stdlib
)

func EnableJSON(marshal func(any) ([]byte, error), unmarshal func([]byte, any) error) {
	enableJSONMarshal = marshal
	enableJSONUnmarshal = unmarshal
}
```

Implementation is ~30 lines of code. Maintenance is minimal.

### 4. Demonstrates Flexibility
Shows that testify's architecture is well-designed and consistently extensible across all serialization formats.

## Arguments Against Making JSON Pluggable

### 1. JSON is in the Standard Library ⚠️

This is the **crucial difference** from YAML:

| Aspect | YAML | JSON |
|--------|------|------|
| Standard library? | ❌ No | ✅ Yes (`encoding/json`) |
| External dependency required? | ✅ Yes | ❌ No |
| Multiple competing versions? | ✅ Yes (v2, v3) | ❌ No |
| Canonical implementation? | ❌ No | ✅ Yes |
| Zero dependencies possible? | ❌ No | ✅ Yes |

Making JSON pluggable contradicts testify's **"standard library first"** philosophy (from APPROACH.md):
- Go values built-in solutions over external dependencies
- `encoding/json` is well-tested, stable, and available everywhere
- Adding abstraction fights the design rather than enhancing it

### 2. Performance in Tests is Rarely Critical ⚠️

Test execution time is dominated by:
- Setup/teardown (DB connections, API calls, file I/O): **milliseconds to seconds**
- The code under test itself: **microseconds to milliseconds**
- Test framework overhead: **microseconds**
- JSON serialization in assertions: **microseconds**

Alternative libraries like `jsoniter` or `sonic` might be 2-3x faster:
- `encoding/json`: 3μs per assertion
- `sonic`: 1μs per assertion
- **Savings: 2 microseconds per assertion**

In a test suite where:
- Each test runs for 10ms minimum
- Setup/teardown dominates
- Most tests have 1-5 JSON assertions

Total savings: **Negligible** (< 0.1% of test runtime)

Only matters if:
- Massive test suites (10,000+ tests)
- Each test has dozens of JSON assertions
- JSON payloads are huge (MBs)

This is an **extreme edge case**, not a common scenario.

### 3. Compatibility Risks ⚠️

Different JSON libraries have **subtle behavioral differences**:

| Behavior | encoding/json | jsoniter | sonic | Impact |
|----------|---------------|----------|-------|--------|
| Number precision | Standard float64 | Configurable | Configurable | Tests pass/fail differently |
| Struct tags | Full support | Partial | Partial | Fields unexpectedly included/excluded |
| Map key ordering | Consistent | Library-specific | Library-specific | Flaky test failures |
| Unicode escaping | RFC 8259 | Variations | Variations | String comparisons fail |
| Null vs missing | Distinct | Configurable | Configurable | Assertion behavior changes |
| Invalid JSON handling | Strict | Tolerant modes | Tolerant modes | Bad data not caught |

**Real-world problem:**
```go
// Test passes with encoding/json
assert.JSONEq(t, `{"value": 1.23456789012345}`, actual)

// Test FAILS with sonic (different precision handling)
// Expected: 1.23456789012345
// Actual:   1.2345678901234500

// Now test results depend on which JSON library is configured!
```

This undermines **test reliability** and creates confusion:
- "Tests pass on my machine but fail in CI" (different JSON library)
- "Why does this assertion fail with jsoniter but not stdlib?"
- Hard to debug, hard to document

### 4. YAGNI (You Ain't Gonna Need It) ⚠️

YAML was made pluggable because of **real, demonstrated problems**:
- Projects stuck on gopkg.in/yaml.v2 for compatibility
- v3 introduced breaking changes
- No canonical implementation in stdlib
- Multiple competing libraries with different features
- **Actual user pain points**

For JSON:
- `encoding/json` is canonical and universal
- No version fragmentation (no json/v2, json/v3 split)
- No demonstrated user demand for alternatives **in testing**
- Works perfectly for 99.9% of use cases
- ❌ **No evidence of a problem to solve**

Adding abstraction without demonstrated need violates YAGNI and increases complexity for theoretical benefit.

### 5. Simplicity and Cognitive Load ⚠️

Every abstraction layer adds:
- **Mental overhead:** Users need to understand it exists
- **Documentation burden:** Need to explain when/why to use it
- **Configuration surface:** One more thing that can be wrong
- **Debug complexity:** "Is my test failing because of my JSON library choice?"
- **Maintenance:** Code to maintain, test, document

For YAML, this cost is **justified** (solves real problems).
For JSON, this cost is **unjustified** (solves theoretical problems).

**Testify's strength is simplicity.** Don't add complexity without clear benefit.

### 6. Standard Library Philosophy ⚠️

From `docs/doc-site/project/APPROACH.md`:

> **Go values built-in solutions over external dependencies**
>
> Testify aligns with this by:
> - Working with standard `testing.T`
> - No custom test runners
> - **Maximizing use of standard library**
> - Zero external dependencies (except opt-in features)

Making JSON pluggable **fights this philosophy** rather than embracing it.

## Recommendation

**Do NOT make JSON pluggable.** Keep `encoding/json` as the only implementation.

### Why This is the Right Call

The situations are **fundamentally different**:

```
YAML:  No stdlib → External dependency required → Multiple options → Pluggable makes sense
JSON:  In stdlib → Zero dependencies possible → One canonical impl → Pluggable adds complexity
```

| Factor | YAML | JSON | Pluggable Justified? |
|--------|------|------|---------------------|
| Stdlib implementation | ❌ | ✅ | YAML: Yes |
| Zero dependencies | ❌ | ✅ | JSON: No |
| Version fragmentation | ✅ | ❌ | YAML: Yes |
| User pain points | ✅ | ❌ | JSON: No |
| Performance critical | ❌ | ❌ | Neither: No |
| Compatibility issues | ✅ | ❌ | YAML: Yes |

### What to Do Instead

#### 1. Document the Decision

Add to documentation (e.g., in customization section or FAQ):

```markdown
## Why JSON is Not Pluggable

**Q: YAML assertions are pluggable. Can I use a custom JSON library?**

A: No. JSON assertions use `encoding/json` from the Go standard library, and this is not
configurable.

**Why the difference?**

YAML and JSON have fundamentally different situations:

| Aspect | YAML | JSON |
|--------|------|------|
| Standard library? | No | **Yes** (`encoding/json`) |
| External dependency? | Yes | **No** |
| Version fragmentation? | Yes (v2, v3) | **No** |
| User pain points? | Yes | **No** |

Making YAML pluggable solves real problems:
- Projects stuck on yaml.v2 for compatibility
- Performance needs (goccy/go-yaml is 2-3x faster)
- Better error messages for debugging

Making JSON pluggable would add complexity without solving real problems:
- `encoding/json` works excellently for testing
- Performance differences are negligible in test suites
- Alternative libraries have subtle compatibility differences
- No demonstrated user demand

**If you have a compelling use case, please open an issue to discuss it.**
```

#### 2. Community Integrations Page

Create `docs/doc-site/usage/COMMUNITY.md`:

```markdown
## Community Integrations

### YAML Libraries

Testify supports pluggable YAML unmarshalers. Users have successfully integrated:

**[goccy/go-yaml](https://github.com/goccy/go-yaml)**
- 2-3x faster unmarshaling
- Colored error messages for debugging
- See [customization guide](./USAGE.md#customization) for integration

**[gopkg.in/yaml.v2](https://gopkg.in/yaml.v2)**
- Legacy projects requiring v2 compatibility

### JSON Libraries

Testify uses `encoding/json` from the Go standard library. We intentionally
do **not** support pluggable JSON implementations because:

1. **Standard library coverage is excellent** - `encoding/json` handles all valid JSON
2. **Alternative libraries have compatibility differences** - subtle behavior changes cause test flakiness
3. **Test performance is rarely JSON-bound** - setup/teardown dominates test runtime
4. **No demonstrated user need** - no real-world use cases reported

**Standard library first** is a core Go value that testify embraces.

If you have a use case requiring a different JSON library in assertions,
please [open an issue](https://github.com/go-openapi/testify/issues) to discuss it.
```

#### 3. Internal Note for Maintainers

Leave an internal comment in `internal/assertions/json.go`:

```go
// NOTE: Unlike YAML, JSON is intentionally NOT pluggable.
//
// YAML required pluggability due to:
// - No stdlib implementation
// - Version fragmentation (v2 vs v3)
// - Competing implementations with different tradeoffs
//
// JSON has none of these issues:
// - encoding/json is in stdlib, stable, and universal
// - No version fragmentation
// - Alternative libraries have compatibility risks
// - No demonstrated user demand
//
// If this changes in the future (e.g., encoding/json/v2 emerges, or
// multiple users request it), we can reconsider. Until then, YAGNI.
//
// See: .claude/plans/ramblings/json-extension-analysis.md
```

## When to Reconsider

Make JSON pluggable IF:

1. **Multiple users request it** with real, specific use cases (not theoretical)
2. **encoding/json gets a competing stdlib version** (like `encoding/json/v2`)
3. **Performance becomes measurably important** in real test suites:
   - User demonstrates JSON assertions are >10% of test runtime
   - Not synthetic benchmarks, but real-world test suites
4. **Compatibility requirements emerge:**
   - Testing systems that rely on jsoniter-specific behavior
   - Need for specific number/precision handling not in stdlib
5. **Standard library has critical bugs/limitations** that block users

But **wait for actual demand** with evidence, not anticipation.

## Design Principles

### Current JSON Design (Correct)
```
encoding/json → JSONEq → Simple, reliable, zero dependencies
```

### Hypothetical Pluggable JSON (Incorrect)
```
User choice → jsoniter/sonic/stdlib → JSONEq → Complex, compatibility risks, minimal benefit
```

### The Difference
```
YAML: Pluggable solves problems (no stdlib, version fragmentation, user pain)
JSON: Pluggable creates problems (compatibility, complexity, test reliability)
```

## Comparison Table: When to Make Something Pluggable

| Criteria | YAML | JSON | Pluggable? |
|----------|------|------|------------|
| No stdlib implementation | ✅ Yes | ❌ No | YAML: Yes |
| External dependency required | ✅ Yes | ❌ No | YAML: Yes |
| Multiple competing versions | ✅ Yes (v2, v3) | ❌ No | YAML: Yes |
| User pain points demonstrated | ✅ Yes | ❌ No | YAML: Yes |
| Compatibility across libraries | ❌ Poor | ✅ Good | JSON: No |
| Performance critical in tests | ❌ No | ❌ No | Neither: No |
| Stdlib-first philosophy | N/A | ✅ Yes | JSON: No |
| **Recommendation** | **Pluggable** | **Not Pluggable** | - |

## Conclusion

**Don't make JSON pluggable.** The cost-benefit analysis is clear:

**Costs:**
- Increased complexity
- Documentation burden
- Compatibility risks (different libraries have subtle differences)
- Test reliability issues (results change based on library choice)
- Violates "stdlib first" philosophy
- Maintenance burden

**Benefits:**
- Theoretical performance gains (negligible in practice)
- API consistency with YAML (not valuable if it adds complexity)
- Future-proofing (YAGNI - no evidence of need)

The situations are fundamentally different:
- **YAML pluggability:** Solves real problems, justified by necessity
- **JSON pluggability:** Theoretical benefit, unjustified complexity

### Embrace "Stdlib First"

Testify's strength is simplicity and Go idiomatic design. Using `encoding/json` is the right default, and making it the
**only** option reinforces testify's philosophy:
- Standard library first
- Zero dependencies (except explicit opt-ins)
- Simple, predictable, reliable
- Works everywhere

### The Pattern is the Value

The valuable lesson from YAML is not "everything should be pluggable" but rather "abstraction should solve real
problems."

YAML abstraction: ✅ Solves real problems
JSON abstraction: ❌ Theoretical benefit only

---

**TL;DR:** Keep JSON assertions using `encoding/json` only. Don't add pluggability. Document the decision clearly.
The stdlib is excellent for testing, and adding abstraction would increase complexity without solving real problems.

**YAGNI + Stdlib First = Keep JSON Simple.**
