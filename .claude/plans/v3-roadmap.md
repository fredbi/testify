# Testify v3 Roadmap

**Status**: Planning
**Target**: Post v1.2 release
**Date**: 2026-01-23

## Vision Statement

Testify v3 will deliver **three strategic pillars**:

1. **Extensibility**: Make the library freely customizable without framework complexity
2. **Type Safety**: Leverage generics for error-aware assertions (T, error) pattern
3. **Safety Testing**: Catch resource leaks (goroutines, file descriptors)

All while maintaining our core principles:
- ✅ Zero external dependencies
- ✅ Idiomatic Go (no magical DSL)
- ✅ Type-safe with generics
- ✅ Performance-focused
- ✅ Works with standard `go test`

---

## Strategic Context

### Competitive Position

From [COMPETITIVE_ANALYSIS.md](./COMPETITIVE_ANALYSIS.md):
- **Testify's strength**: Pragmatic assertion library, zero-deps, type-safe
- **Ginkgo's strength**: BDD framework, rich tooling, advanced async
- **Our strategy**: Narrow gaps where valuable (async, safety), avoid becoming framework

### Key Insights

1. **Extensibility > Feature Bloat**: Enable users to build custom assertions rather than pre-building everything
2. **(T, error) Pattern**: Fundamental Go idiom poorly served by existing assertions
3. **Safety Assertions**: Real bugs (goroutine/FD leaks) that tests should catch
4. **Domain Focus**: JSON/YAML assertions serve go-openapi needs directly

---

## Roadmap Phases

### Phase 1: Core Infrastructure (Foundation)

**Goal**: Enable extensibility and establish patterns for future features

#### 1.1: Make Assertions Implement TestingT Interface

**Why**: This is the **keystone feature** that enables everything else.

**What it does:**
```go
type Assertions struct {
    t TestingT
}

// Implement TestingT by delegation
func (a *Assertions) Errorf(format string, args ...any) {
    if h, ok := a.t.(H); ok {
        h.Helper()
    }
    a.t.Errorf(format, args...)
}

func (a *Assertions) FailNow() {
    if h, ok := a.t.(H); ok {
        h.Helper()
    }
    a.t.FailNow()
}

// ... other TestingT methods
```

**Benefits:**

1. **Free Customization** - Users can compose their own assertions:
```go
type MyAssertions struct {
    *assert.Assertions
}

func (ma *MyAssertions) ValidUser(user *User) bool {
    return ma.NotNil(user) &&
           ma.NotEmpty(user.Email) &&
           ma.Greater(user.Age, 0)
}

// MyAssertions IS a TestingT - works everywhere
```

2. **Generics Workaround** - Forward-style users can access generics:
```go
// Can't do this (method type parameters):
a := assert.New(t)
a.EqualT(x, y)  // ❌ Not allowed in Go

// But can do this (Assertions implements TestingT):
assert.EqualT(a, x, y)  // ✅ Works! a IS a TestingT
```

3. **Ecosystem Growth** - Users build domain-specific assertion libraries without framework buy-in

**Implementation:**
- Update `assert/assertions.go` and `require/requirements.go`
- Implement all `TestingT` interface methods by delegation
- Document pattern in new "Custom Assertions" guide
- Add examples to `docs/doc-site/usage/EXAMPLES.md`

**Complexity**: Low (50-100 LOC)
**Value**: Very High (enables ecosystem)
**Priority**: P0 (must have)

---

#### 1.2: Implement Result[T] Pattern

**Why**: Foundation for error-aware assertions handling `(T, error)` pattern

**What it is:**
```go
// Exported for user extensibility
type Result[T any] struct {
    t         TestingT
    value     T
    succeeded bool
}

// Chainable assertions based on type constraints
func (r *Result[T]) Equal(expected T, msgAndArgs ...any) bool {
    if !r.succeeded {
        return false
    }
    return EqualT(r.t, expected, r.value, msgAndArgs...)
}

func (r *Result[T]) NotNil(msgAndArgs ...any) bool {
    if !r.succeeded {
        return false
    }
    return NotNil(r.t, r.value, msgAndArgs...)
}

// Type-constrained methods
func (r *Result[T]) GreaterOrEqual(threshold T, msgAndArgs ...any) bool
    where T is Ordered {
    if !r.succeeded {
        return false
    }
    return GreaterOrEqualT(r.t, r.value, threshold, msgAndArgs...)
}
```

**Implementation:**
- Create `internal/assertions/result.go`
- Implement methods for all applicable assertions
- Use type constraints to limit methods by type (comparable, Ordered, etc.)
- Export as `Result[T]` for user extensibility
- Document in [error-aware-assertions.md](./error-aware-assertions.md)

**Complexity**: Medium (200-300 LOC)
**Value**: High (enables Phase 2)
**Priority**: P0 (foundation for Phase 2)

---

### Phase 2: Error-Aware Assertions (Immediate Value)

**Goal**: Handle Go's idiomatic `(T, error)` pattern naturally

See [error-aware-assertions.md](./error-aware-assertions.md) for full details.

#### 2.1: UnmarshalJSONAsT

**Why**: Frequent pattern in API testing (especially go-openapi)

```go
// domain: json

// UnmarshalJSONAsT unmarshals JSON data and returns an assertion helper.
func UnmarshalJSONAsT[T any](t TestingT, data []byte, msgAndArgs ...any) *Result[T] {
    var value T
    if err := json.Unmarshal(data, &value); err != nil {
        Fail(t, fmt.Sprintf("JSON unmarshal failed: %v", err), msgAndArgs...)
        return &Result[T]{t: t, value: value, succeeded: false}
    }
    return &Result[T]{t: t, value: value, succeeded: true}
}
```

**Usage:**
```go
// Clean, type-safe JSON assertions
assert.UnmarshalJSONAsT[User](t, jsonBytes).Equal(expectedUser)
assert.UnmarshalJSONAsT[Config](t, configJSON).NotNil()

// Chain multiple assertions
user := assert.UnmarshalJSONAsT[User](t, jsonBytes)
assert.Equal(t, "Alice", user.value.Name)
assert.Greater(t, user.value.Age, 0)
```

**Implementation:**
- Add to `internal/assertions/json.go`
- Comprehensive tests in `internal/assertions/json_test.go`
- Generate all variants (assert/require, format, forward)
- Update codegen to handle `Result[T]` return types
- Document in API reference and examples

**Complexity**: Low (uses Result[T] pattern)
**Value**: High (direct go-openapi need)
**Priority**: P1 (high value, clear use case)

---

#### 2.2: UnmarshalYAMLAsT

**Why**: Same pattern as JSON for YAML (if yaml enabled)

```go
// domain: yaml

func UnmarshalYAMLAsT[T any](t TestingT, data []byte, msgAndArgs ...any) *Result[T]
```

**Usage:**
```go
assert.UnmarshalYAMLAsT[Config](t, yamlBytes).Equal(expected)
```

**Implementation:**
- Similar to UnmarshalJSONAsT
- Lives in `enable/yaml` module (optional dependency)
- Only available when YAML is enabled

**Complexity**: Low (copy JSON pattern)
**Value**: Medium (less common than JSON)
**Priority**: P2 (after JSON working)

---

#### 2.3: EventuallyT

**Why**: Handle `(T, error)` in async operations naturally

```go
// domain: testing

// EventuallyT polls f() until it returns (value, nil), then returns Result[T]
func EventuallyT[T any](
    t TestingT,
    f func() (T, error),
    waitFor, tick time.Duration,
    msgAndArgs ...any,
) *Result[T] {
    timer := time.NewTimer(waitFor)
    defer timer.Stop()

    ticker := time.NewTicker(tick)
    defer ticker.Stop()

    for {
        select {
        case <-timer.C:
            var zero T
            return &Result[T]{
                t:         t,
                value:     zero,
                succeeded: false,
            }
        case <-ticker.C:
            value, err := f()
            if err == nil {
                return &Result[T]{t: t, value: value, succeeded: true}
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
```

**Implementation:**
- Update `internal/assertions/testing.go`
- Add alongside existing `Eventually` (maintain backward compat)
- Ensure goroutine leak fixes from v1.x are preserved
- Comprehensive async tests

**Complexity**: Medium (async + generics)
**Value**: High (improves async testing significantly)
**Priority**: P1 (proven need from error-aware discussion)

---

#### 2.4: EventuallyWithContextT

**Why**: Modern Go emphasizes context.Context for cancellation

```go
// domain: testing

// EventuallyWithContextT respects context cancellation
func EventuallyWithContextT[T any](
    ctx context.Context,
    t TestingT,
    f func(context.Context) (T, error),
    tick time.Duration,
    msgAndArgs ...any,
) *Result[T] {
    ticker := time.NewTicker(tick)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            var zero T
            return &Result[T]{
                t:         t,
                value:     zero,
                succeeded: false,
            }
        case <-ticker.C:
            value, err := f(ctx)
            if err == nil {
                return &Result[T]{t: t, value: value, succeeded: true}
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

**Implementation:**
- Add to `internal/assertions/testing.go`
- Ensure proper context handling
- Test cancellation behavior
- Document context patterns

**Complexity**: Medium (context handling)
**Value**: High (idiomatic Go)
**Priority**: P1 (pairs with EventuallyT)

---

### Phase 3: Safety Assertions (Proven Demand)

**Goal**: Catch resource leaks that indicate real bugs

See [new-safety-features.md](./new-safety-features.md) for background discussion.

#### 3.1: NoGoroutineLeak

**Why**: Goroutine leaks are real bugs. Fred has used uber-go/goleak extensively.

```go
// domain: testing

type GoroutineOption func(*goroutineConfig)

// Filter options
func IgnoreGoroutine(functionName string) GoroutineOption
func IgnoreGoroutinePattern(pattern *regexp.Regexp) GoroutineOption
func IgnoreCurrent() GoroutineOption

// NoGoroutineLeak asserts f() doesn't leak goroutines
func NoGoroutineLeak(t TestingT, f func(), opts ...GoroutineOption) bool {
    before := captureGoroutineStacks()

    f()

    // Wait for goroutines to settle
    time.Sleep(10 * time.Millisecond)
    runtime.Gosched()

    after := captureGoroutineStacks()

    leaked := findLeakedGoroutines(before, after, opts)
    if len(leaked) > 0 {
        return Fail(t, formatLeakedGoroutines(leaked))
    }
    return true
}

func captureGoroutineStacks() []goroutineStack {
    buf := make([]byte, 1<<20) // 1MB buffer
    n := runtime.Stack(buf, true)
    return parseGoroutineStacks(buf[:n])
}

func parseGoroutineStacks(buf []byte) []goroutineStack {
    // Parse runtime.Stack() output
    // Format: "goroutine NNN [state]:\nfunc()\n\tfile.go:line\n..."
}

func findLeakedGoroutines(before, after []goroutineStack, opts ...GoroutineOption) []goroutineStack {
    // Compare before/after, apply filters
}
```

**Usage:**
```go
// Basic usage
assert.NoGoroutineLeak(t, func() {
    startServer()
    makeRequest()
    stopServer()
    // If goroutines leaked, test fails
})

// With filtering
assert.NoGoroutineLeak(t, testFunc,
    IgnoreGoroutine("database/sql.(*DB).connectionOpener"),
    IgnoreGoroutine("internal/poll.runtime_pollWait"),
    IgnoreCurrent(),  // Ignore goroutines already running
)
```

**Implementation Strategy:**

**Option A**: Wrap uber-go/goleak (❌ breaks zero-dependency principle)

**Option B**: Implement ourselves (✅ recommended)
- Study goleak implementation for patterns
- Parse `runtime.Stack(buf, true)` output
- ~200-300 LOC total
- Maintain zero dependencies

**Implementation:**
- Create `internal/assertions/goroutines.go`
- Implement stack capture and parsing
- Implement filtering logic
- Comprehensive tests with known leak patterns
- Document common filters for standard library goroutines

**Complexity**: Medium (stack parsing, filtering)
**Value**: Very High (catches real bugs, proven demand)
**Priority**: P1 (Fred prefers this over Eventually improvements)

---

#### 3.2: NoFileDescriptorLeak (Unix)

**Why**: File descriptor leaks cause resource exhaustion

**Phase 1: Unix-only** (Linux, macOS, BSD)

```go
// domain: testing

type FileDescriptorOption func(*fdConfig)

// Filter options
func IgnoreNetworkFDs() FileDescriptorOption
func IgnorePipeFDs() FileDescriptorOption
func IgnoreSocketFDs() FileDescriptorOption

// NoFileDescriptorLeak asserts f() doesn't leak file descriptors
func NoFileDescriptorLeak(t TestingT, f func(), opts ...FileDescriptorOption) bool {
    if runtime.GOOS == "windows" {
        Skip(t, "NoFileDescriptorLeak not supported on Windows")
        return true
    }

    before := captureOpenFDs()

    f()

    after := captureOpenFDs()

    leaked := findLeakedFDs(before, after, opts)
    if len(leaked) > 0 {
        return Fail(t, formatLeakedFDs(leaked))
    }
    return true
}

// Linux: read /proc/self/fd
// macOS: use syscall or lsof
func captureOpenFDs() []fileDescriptor {
    switch runtime.GOOS {
    case "linux":
        return captureLinuxFDs()    // Read /proc/self/fd
    case "darwin":
        return captureDarwinFDs()   // Use lsof or libproc
    case "freebsd", "openbsd", "netbsd":
        return captureBSDFDs()      // Similar to macOS
    default:
        return nil
    }
}

func captureLinuxFDs() []fileDescriptor {
    entries, _ := os.ReadDir("/proc/self/fd")
    fds := make([]fileDescriptor, 0, len(entries))

    for _, entry := range entries {
        fdNum := entry.Name()
        target, _ := os.Readlink("/proc/self/fd/" + fdNum)

        fds = append(fds, fileDescriptor{
            fd:     fdNum,
            target: target,
            typ:    classifyFD(target),
        })
    }
    return fds
}

type fileDescriptor struct {
    fd     string
    target string  // What the FD points to
    typ    fdType  // file, socket, pipe, etc.
}

type fdType int

const (
    fdTypeFile fdType = iota
    fdTypeSocket
    fdTypePipe
    fdTypeUnknown
)

func classifyFD(target string) fdType {
    switch {
    case strings.HasPrefix(target, "socket:["):
        return fdTypeSocket
    case strings.HasPrefix(target, "pipe:["):
        return fdTypePipe
    case strings.HasPrefix(target, "/"):
        return fdTypeFile
    default:
        return fdTypeUnknown
    }
}
```

**Usage:**
```go
// Basic usage
assert.NoFileDescriptorLeak(t, func() {
    f, _ := os.Open("test.txt")
    // If f not closed, test fails
})

// With filtering
assert.NoFileDescriptorLeak(t, testFunc,
    IgnoreNetworkFDs(),  // Ignore sockets (e.g., HTTP client)
    IgnorePipeFDs(),     // Ignore pipes
)
```

**Phase 2: Windows (Future)** - Coarse-grained handle count only

```go
// Windows: track total handle count
func getWindowsHandleCount() (int, error) {
    // Use GetProcessHandleCount() WinAPI
    // Requires syscall.Syscall
}
```

**Implementation:**
- Create `internal/assertions/filedescriptors.go`
- Implement Unix FD capture (Linux, macOS, BSD)
- Implement filtering logic
- Skip on Windows with clear message
- Document limitations
- Consider Phase 2 (Windows) based on demand

**Complexity**: Medium (Unix variants, filtering)
**Value**: High (catches real bugs)
**Priority**: P2 (after goroutine leak detection)

**Note**: fdooze (https://github.com/thediveo/fdooze) exists but is Linux-only and not well-implemented. We can do better.

---

### Phase 4: Nice-to-Have (Evaluate Demand)

**Goal**: Features that are interesting but lower priority

#### 4.1: JSONPointerT (JSON Pointer Assertions)

**Why**: Clean syntax for asserting on deeply nested JSON structures

```go
// domain: json

// JSONPointerT extracts value at JSON Pointer path with type safety
func JSONPointerT[T any](t TestingT, data any, pointer string) *Result[T] {
    value, ok := navigateJSONPointer(data, pointer)
    if !ok {
        var zero T
        return &Result[T]{t: t, value: zero, succeeded: false}
    }

    // Type assertion
    typed, ok := value.(T)
    if !ok {
        return &Result[T]{
            t: t,
            succeeded: Fail(t, "JSON pointer value has wrong type: expected %T, got %T",
                typed, value),
        }
    }

    return &Result[T]{t: t, value: typed, succeeded: true}
}

// navigateJSONPointer implements RFC 6901 JSON Pointer
func navigateJSONPointer(data any, pointer string) (any, bool) {
    if pointer == "" {
        return data, true
    }
    if !strings.HasPrefix(pointer, "/") {
        return nil, false
    }

    parts := parseJSONPointer(pointer)  // Split by /, handle ~0 and ~1 escaping

    current := data
    for _, part := range parts {
        switch v := current.(type) {
        case map[string]any:
            current = v[part]
        case []any:
            index, _ := strconv.Atoi(part)
            if index >= 0 && index < len(v) {
                current = v[index]
            } else {
                return nil, false
            }
        default:
            return nil, false
        }

        if current == nil {
            return nil, false
        }
    }

    return current, true
}
```

**Usage:**
```go
response := map[string]any{
    "user": map[string]any{
        "profile": map[string]any{
            "name": "Alice",
            "age":  30,
        },
    },
    "items": []any{"a", "b", "c"},
}

// Type-safe deep assertions
assert.JSONPointerT[string](t, response, "/user/profile/name").Equal("Alice")
assert.JSONPointerT[int](t, response, "/user/profile/age").GreaterThan(18)

// Array indexing
assert.JSONPointerT[string](t, response, "/items/0").Equal("a")
assert.JSONPointerT[string](t, response, "/items/2").Equal("c")
```

**Benefits:**
- Clean syntax for deep JSON assertions
- Zero dependencies (just string parsing)
- Leverages Result[T] pattern
- Useful for API testing

**Limitations:**
- Only works with `map[string]any` / `[]any` structures
- Doesn't work with typed structs
- JSON Pointer spec (RFC 6901) is niche

**Implementation:**
- Create `internal/assertions/jsonpointer.go`
- Implement RFC 6901 parsing (~100 LOC)
- Implement navigation logic
- Comprehensive tests with edge cases

**Complexity**: Low-Medium (string parsing, navigation)
**Value**: Medium (niche but useful for API testing)
**Priority**: P3 (after core features proven)

**Decision**: Implement **after** Phase 3 complete and if demand emerges from go-openapi usage.

---

#### 4.2: Context Caching (Deferred/Rejected)

**Why considered**: Cache compiled regexps, reduce redundant work

**Why rejected**:
- Contexts are read-only (can't inject cache)
- Workaround (pointer in context) causes race conditions
- Regexp compilation is fast enough
- Better pattern: document "compile once" approach

```go
// Document this pattern instead:
var phonePattern = regexp.MustCompile(`^\d{3}-\d{4}$`)

func TestPhones(t *testing.T) {
    assert.Regexp(t, phonePattern, phone1)
    assert.Regexp(t, phonePattern, phone2)
}
```

**Decision**: ❌ **Rejected**. Not worth the complexity.

---

## Implementation Strategy

### Sequencing

**Phase 1 (Foundation)**
1. Assertions implements T interface
2. Result[T] pattern implementation

**Then parallel tracks:**

**Track A (Error-Aware):**
3. UnmarshalJSONAsT
4. EventuallyT + EventuallyWithContextT
5. UnmarshalYAMLAsT

**Track B (Safety):**
3. NoGoroutineLeak
4. NoFileDescriptorLeak (Unix)

**Track C (Nice-to-Have):**
5. Evaluate demand, implement JSONPointerT if needed

### Code Generation

All new functions must integrate with codegen:
- Update `codegen/internal/scanner/` to recognize new patterns
- Update `codegen/internal/generator/` to handle Result[T] returns
- Generate all variants (assert/require, format, forward)
- Generate comprehensive tests from Examples
- Generate documentation

### Testing Strategy

**Layer 1**: Internal assertions tests (exhaustive)
- All edge cases
- Error conditions
- Type constraints
- Resource cleanup

**Layer 2**: Generated tests (smoke tests)
- Basic success/failure cases
- All variants work (assert/require/format/forward)

**Layer 3**: Integration tests
- Real-world scenarios from go-openapi
- Performance benchmarks
- Resource leak detection under load

### Documentation

Update all documentation:
1. **Custom Assertions Guide** (new) - How to extend using Assertions as T
2. **Error-Aware Assertions Guide** (new) - Result[T] pattern and usage
3. **Safety Testing Guide** (new) - Goroutine/FD leak detection
4. **Generics Guide** - Add Result[T] section
5. **Examples** - Add real-world examples for each feature
6. **API Reference** - Auto-generated from code
7. **Migration Guide** - How to adopt new features

---

## Success Criteria

### Phase 1 Success
- ✅ Assertions implements TestingT
- ✅ Users can compose custom assertions
- ✅ Generic assertions work with Assertions object: `assert.EqualT(a, x, y)`
- ✅ Result[T] pattern established and documented
- ✅ Code generator handles new patterns

### Phase 2 Success
- ✅ UnmarshalJSONAsT used in go-openapi tests
- ✅ EventuallyT handles (T, error) naturally
- ✅ EventuallyWithContextT respects cancellation
- ✅ No regression in existing tests
- ✅ Performance maintained or improved

### Phase 3 Success
- ✅ NoGoroutineLeak catches real leaks in go-openapi
- ✅ NoFileDescriptorLeak works on Linux/macOS
- ✅ False positive rate acceptable (<5%)
- ✅ Clear error messages showing leaked resources
- ✅ Documentation with common filter patterns

### Overall Success
- ✅ Zero external dependencies maintained
- ✅ Type safety extended to new features
- ✅ Extensibility proven with real-world custom assertions
- ✅ go-openapi team adopts new features
- ✅ Community feedback positive

---

## Non-Goals (What We Won't Do)

1. **❌ BDD Framework Features**
   - No Describe/Context/When hierarchy
   - Keep testify focused on assertions
   - Users wanting BDD should use Ginkgo

2. **❌ Matcher DSL**
   - No fluent matcher composition like Gomega
   - Keep function-based assertion style
   - Maintain simplicity

3. **❌ Required Tooling**
   - Must work with `go test` (no custom CLI)
   - No required code generation at user's project
   - Stay compatible with standard Go toolchain

4. **❌ External Dependencies**
   - Zero-dependency principle is sacred
   - Implement ourselves rather than wrap existing libs
   - Exception: optional modules (yaml, colors)

5. **❌ General Memory Leak Detection**
   - Too many false positives
   - Can't distinguish leak from caching
   - Better tools exist (pprof)

6. **❌ Windows FD Leak Detection (Phase 1)**
   - Too complex for initial release
   - Unix-only acceptable for v1
   - Revisit if proven demand

7. **❌ Context Caching**
   - Complexity doesn't justify gains
   - Race condition risks
   - Document better patterns instead

---

## Risk Management

### Technical Risks

**Risk**: Result[T] pattern too complex for users
- **Mitigation**: Excellent documentation with examples
- **Fallback**: Keep simple function variants alongside

**Risk**: Goroutine leak detection has false positives
- **Mitigation**: Study goleak's filtering, provide good defaults
- **Fallback**: Provide flexible filtering options

**Risk**: FD leak detection breaks on some Unix variants
- **Mitigation**: Test on Linux, macOS, FreeBSD
- **Fallback**: Skip gracefully on unsupported platforms

**Risk**: Codegen can't handle Result[T] returns
- **Mitigation**: Prototype codegen changes early
- **Fallback**: Manual generation for v1, fix codegen later

### Product Risks

**Risk**: Features don't get adopted
- **Mitigation**: Work closely with go-openapi team
- **Validation**: Use in real go-swagger tests

**Risk**: Breaks backward compatibility
- **Mitigation**: All new features are additive
- **Validation**: Existing tests must pass

**Risk**: Documentation insufficient
- **Mitigation**: Write docs alongside code
- **Validation**: External reviewer feedback

---

## Open Questions

### For Fred to Decide

1. **Should Result[T] be exported?**
   - Pro: Enables user extensibility (can return from custom assertions)
   - Con: Locks into public API
   - **Recommendation**: Yes, export as `assert.Result[T]`

2. **Deprecate old Eventually?**
   - Pro: Simplify API surface
   - Con: Breaking change
   - **Recommendation**: Keep both, document preference for EventuallyT

3. **MarshalJSONAsT too?**
   - Question: Do you need marshal direction or just unmarshal?
   - **Recommendation**: Start with unmarshal only, add marshal if needed

4. **Custom Assertions in separate package?**
   - Question: Should examples live in `testify/examples` or in docs only?
   - **Recommendation**: Examples in docs, maybe `testify-contrib` repo later

### Design Decisions Needed

1. **Which assertions should Result[T] support?**
   - All comparable: Equal, NotEqual, Same, NotSame
   - All Ordered: Greater, Less, GreaterOrEqual, LessOrEqual
   - Others: NotNil, Len, Contains, etc.
   - **Decision**: Implement based on type constraints, document clearly

2. **NoGoroutineLeak default filters?**
   - What should be ignored by default?
   - Standard library goroutines?
   - Testing framework goroutines?
   - **Decision**: Start conservative (no defaults), add based on feedback

3. **FD leak detection on macOS?**
   - Use lsof (requires exec)?
   - Use native syscalls (complex)?
   - Use libproc (requires cgo)?
   - **Decision**: Research further, pick most reliable

---

## Related Documents

- [COMPETITIVE_ANALYSIS.md](./COMPETITIVE_ANALYSIS.md) - Testify vs Ginkgo/Gomega
- [error-aware-assertions.md](./error-aware-assertions.md) - Result[T] pattern detailed design
- [new-safety-features.md](./new-safety-features.md) - Safety assertions discussion
- [Generics Guide](../docs/doc-site/usage/GENERICS.md) - Existing generic assertions
- [Benchmarks](../docs/doc-site/project/maintainers/BENCHMARKS.md) - Performance data

---

## Timeline (Tentative)

**Post v1.2 Release:**

**Month 1-2**: Phase 1 (Foundation)
- Assertions implements T
- Result[T] pattern
- Code generator updates
- Documentation framework

**Month 3-4**: Phase 2 (Error-Aware Assertions)
- UnmarshalJSONAsT
- EventuallyT variants
- Testing with go-openapi

**Month 5-6**: Phase 3 (Safety Assertions)
- NoGoroutineLeak
- NoFileDescriptorLeak (Unix)
- Real-world validation

**Month 7**: Phase 4 (Evaluate)
- Gather feedback
- Decide on JSONPointerT
- Plan v4 features

**Target**: v3.0.0 release ~6 months post v1.2

---

## Next Steps

1. ✅ Document the roadmap (this file)
2. ⏳ Fix bugs for v1.2 release (in progress)
3. ⏳ Cut v1.2 release
4. ⏳ Begin Phase 1: Design Assertions as TestingT
5. ⏳ Prototype Result[T] pattern
6. ⏳ Update code generator
7. ⏳ Implement first features (UnmarshalJSONAsT, NoGoroutineLeak)

---

**Note**: This roadmap is a living document. Priorities may shift based on:
- User feedback from v1.2
- go-openapi/go-swagger needs
- Community requests
- Technical discoveries during implementation
