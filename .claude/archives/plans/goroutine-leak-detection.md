# Plan: Goroutine Leak Detection via pprof Labels

**Date:** 2026-02-01
**Branch:** TBD (new feature branch off master)
**Status:** Design finalized, ready for implementation

## Summary

Replace the goleak-derived stack-parsing approach in `internal/leak/` with a
pprof-label-based mechanism. Instead of capturing all goroutine stacks and
filtering out known system goroutines with heuristics, we label the tested
function's goroutine using `pprof.Do`. Child goroutines inherit the label
automatically. After the function returns, any goroutines still carrying the
label are leaks.

This eliminates false positives, removes brittle filter lists, and naturally
supports `t.Parallel()`.

## API

```go
// In internal/assertions/safety.go

// NoGoRoutineLeak asserts that the tested function does not leak any goroutines.
//
// The function is executed in an instrumented context. After it returns, any
// goroutines it spawned (directly or transitively) that are still running
// are reported as leaks.
//
// Cleanup of resources (servers, channels, etc.) should happen inside the
// tested function itself, not via t.Cleanup().
//
// # Usage
//
//     NoGoRoutineLeak(t, func() {
//         server := startServer()
//         defer server.Shutdown()
//         // ... test code ...
//     })
//
// Domain: safety
func NoGoRoutineLeak(t T, tested func(), msgAndArgs ...any) bool
```

No options. No filter configuration. Labels handle isolation automatically.

## Implementation Steps

### ✅ Step 1: Rewrite `internal/leak/` core

Replace the current goleak-derived code with the label-based approach.

**Files to rewrite:**
- `internal/leak/leaks.go` — new `Find(tested func()) error` entry point

**New implementation:**

```
Find(tested func()) error:
  1. Generate a unique ID (crypto/rand, hex-encoded)
  2. Create pprof labels: pprof.Labels("testify-leak-check", id)
  3. Build the needle: fmt.Appendf(nil, "%q:%q", key, id)
  4. Run tested() inside pprof.Do, in a separate goroutine:
     - Start goroutine
     - defer recover() to capture panics
     - pprof.Do(ctx, labels, func(_ context.Context) { tested() })
     - Wait for goroutine to complete
  5. Retry loop with exponential backoff:
     - Capture goroutine profile (debug=1)
     - Search for needle in profile
     - If not found → return nil (clean)
     - If found and retries exhausted → parse profile, return error
     - Otherwise → backoff, retry
  6. If tested() panicked, re-panic after leak check
```

**Context handling:**
- ✅ Use `context.Background()` for now (for when T does not implement Context() context.Context)
- ✅ Future enhancement: accept `t.Context()` when the assertion (done by default)
  interface supports it (Go 1.21+ `testing.TB.Context()`)

### ✅ Step 2: Clean up `internal/leak/` — remove unused code

**Files to delete or gut:**
- ✅ `internal/leak/options.go` — remove `Option` interface, all filter
  functions (`IgnoreTopFunction`, `IgnoreAnyFunction`, `IgnoreCreatedBy`,
  `IgnoreCurrent`, `Cleanup`, `RunOnFailure`), `buildOpts`, built-in
  filters (`isTestStack`, `isSyscallStack`, `isStdLibStack`, `isTraceStack`),
  retry logic (moved into `leaks.go`)
- ✅ `internal/leak/stacks.go` — remove `Stack` struct, full stack parser
  (`stackParser`, `parseStack`, `parseFuncName`, `parseGoStackHeader`),
  `All()`, `Current()`, `getStacks()`, `getStackBuffer()`
- ✅ `internal/leak/scan.go` — remove custom scanner (no longer needed)

**Files to keep (rewritten):**
- ✅ `internal/leak/leaks.go` — the new label-based `Find`

**New files:**
- ✅ `internal/leak/profile.go` — profile capture and label extraction (renamed leak.go)
  (from prototype's `captureProfile`, `extractLabeledBlocks`, `buildNeedle`)
- `internal/leak/format.go` — error message formatting (parse matched
  profile blocks into readable diagnostics)

### ✅ Step 3: Unique ID generation

**File:** `internal/leak/id.go`

- Use `crypto/rand` to generate 16 bytes
- Hex-encode to 32-character string
- No atomic counter, no timestamps — pure randomness

### ✅ Step 4: Goroutine wrapper with panic guard

**In `internal/leak/leaks.go`:**

```go
func runGuarded(ctx context.Context, labels pprof.LabelSet, tested func()) (panicked bool, panicVal any) {
    done := make(chan struct{})
    go func() {
        defer close(done)
        defer func() {
            if r := recover(); r != nil {
                panicked = true
                panicVal = r
            }
        }()
        pprof.Do(ctx, labels, func(_ context.Context) {
            tested()
        })
    }()
    <-done
    return
}
```

Note: `runtime.Goexit()` (from `t.FailNow()`) terminates the goroutine
without triggering `recover()`. The `<-done` unblocks via `defer close(done)`,
and `panicked` remains false. This is acceptable — if the test already
failed fatally, the leak check result is moot.

### ✅ Step 5: Exponential backoff retry

**In `internal/leak/leaks.go`:**

Same strategy as goleak: start at 1µs, double each time, cap at 100ms,
max 20 retries. Total worst-case wait ~2s.

```go
const (
    maxRetries = 20
    maxSleep   = 100 * time.Millisecond
)

func retry(attempt int) bool {
    if attempt >= maxRetries {
        return false
    }
    d := min(time.Duration(int(time.Microsecond)<<uint(attempt)), maxSleep)
    time.Sleep(d)
    return true
}
```

### ✅ Step 6: Error message formatting

**File:** `internal/leak/format.go`

Parse matched profile blocks from debug=1 output to produce readable
error messages. Extract:
- Number of leaked goroutines (the count prefix in each block)
- Blocking state (from the stack frames) (❌ not implemented)
- Top 2-3 stack frames with file:line (❌ not implemented)

Target output:
```
found 2 leaked goroutines:
  goroutine [chan receive]: mypackage.worker
      /path/to/worker.go:42
  goroutine [chan receive]: mypackage.listener
      /path/to/listener.go:88
```

### ✅ Step 7: Profile format guard

**File:** `internal/leak/profile.go`

Add a package-level `init()` or lazy-initialized check that validates
the label round-trip works:
1. Set a known label via `pprof.Do`
2. Capture profile
3. Verify the label appears in expected format
4. If not, set a package-level flag that causes `Find` to return
   a clear error: "goroutine leak detection unavailable: pprof label
   format not recognized (Go version X.Y)"

This guards against silent failures on future Go versions that change
the profile text format.

### ✅ Step 8: Update `internal/assertions/safety.go`

Simplify to:

```go
func NoGoRoutineLeak(t T, tested func(), msgAndArgs ...any) bool {
    // Domain: safety
    if h, ok := t.(H); ok {
        h.Helper()
    }

    err := leak.Find(tested) // replaced Find by Leaked. Added context.
    if err != nil { // no longer return an error but just a string
        return Fail(t, err.Error(), msgAndArgs...)
    }

    return true
}
```

Remove:
- `LeadOption` / `LeakOption` / `leakOption` type aliases
- `[]LeakOption` parameter

### ✅ Step 9: Tests for `internal/leak/`

**File:** `internal/leak/leaks_test.go`

Rewrite from scratch (current tests are commented-out goleak copies).

Test cases:
- **No leak**: clean function → `Find` returns nil
- **Direct leak**: goroutine blocked on channel → `Find` returns error
- **Transitive leak**: grandchild goroutine inherits label → detected
- **Pre-existing goroutine**: started before `Find` → not attributed
- **Parallel isolation**: concurrent `Find` calls don't interfere
- **Panic propagation**: `tested()` panics → leak check runs, then re-panics
- **Goexit handling**: `tested()` calls `runtime.Goexit()` → doesn't hang
- **Fast cleanup**: goroutine exits during retry window → no false positive
- **Error message content**: verify leaked goroutine stack appears in error

**File:** `internal/leak/profile_test.go`

- Needle construction format
- Profile capture returns valid data
- Label extraction from real profile blocks
- Format guard validation

### ✅ Step 10: Tests for `internal/assertions/safety.go`

**File:** `internal/assertions/safety_test.go`

- `NoGoRoutineLeak` with clean function → returns true
- `NoGoRoutineLeak` with leaking function → returns false, mock reports failure
- Error message contains useful goroutine information

### Step 11: Update doc comment for codegen

Ensure `NoGoRoutineLeak` has proper doc comment with:
- `// Domain: safety`
- `// Examples:` section (may need `// NOT IMPLEMENTED` if the test
  args can't be expressed as simple literals)
- Usage example showing the pattern

### ✅ Step 12: Run codegen and verify

```bash
go generate ./...
```

Verify:
- `assert.NoGoRoutineLeak` generated
- `require.NoGoRoutineLeak` generated
- Format variants generated
- Forward methods generated
- Documentation generated in `docs/doc-site/api/safety.md`

### ✅ Step 13: Remove prototype

Delete `internal/assertions/exp-leak/` after the production
implementation is complete and passing.

## File Summary

| File | Action |
|------|--------|
| `internal/leak/leaks.go` | Rewrite (label-based Find) |
| `internal/leak/profile.go` | New (capture, extract, format guard) |
| `internal/leak/format.go` | New (error message formatting) |
| `internal/leak/id.go` | New (crypto/rand unique ID) |
| `internal/leak/options.go` | Delete |
| `internal/leak/stacks.go` | Delete |
| `internal/leak/scan.go` | Delete |
| `internal/leak/leaks_test.go` | Rewrite |
| `internal/leak/options_test.go` | Delete |
| `internal/leak/stacks_test.go` | Delete |
| `internal/leak/profile_test.go` | New |
| `internal/assertions/safety.go` | Simplify (remove options) |
| `internal/assertions/safety_test.go` | New |
| `internal/assertions/exp-leak/` | Delete (after completion) |

## Design Decisions

1. **No options/filters** — labels provide natural isolation; filter
   lists are unnecessary
2. **`func()` not `func(context.Context)`** — labels propagate regardless
   of context; simpler API wins. Context propagation is a future enhancement.
3. **Separate goroutine for `tested()`** — protects against `Goexit`/panic
   killing the leak check
4. **crypto/rand for IDs** — eliminates collision risk in parallel tests
5. **debug=1 profile format** — includes labels; debug>=2 does not
6. **Format guard** — defensive check against Go version changes in
   pprof output format
7. **Cleanup inside `tested()`** — document that `t.Cleanup()` runs
   after the leak check; resources should be released inside the function

## v2.4 Roadmap: Generalization to Resource Leak Detection

### NoFileDescriptorLeak (v2.4)

The pprof label approach does **not** generalize to file descriptors:
- FDs are kernel resources (ints in the process fd table), not Go runtime
  objects — there is no pprof profile for open fds
- The stack that called `os.Open` is gone by the time we check for leaks
- Intercepting `os.Open` would require the user to use a wrapped API
  or build-time instrumentation — neither is acceptable

**The snapshot-and-diff approach remains correct for fds:**

```
before: readdir /proc/self/fd → set of ints
run tested()
after:  readdir /proc/self/fd → set of ints
diff:   after - before = leaked fds
```

**Platform support:**
- Linux: `/proc/self/fd` (straightforward)
- macOS: `/dev/fd` or `lsof -p $PID` or `proc_pidinfo` via cgo
- Windows: `GetProcessHandleCount()` for coarse count (phase 1),
  `NtQuerySystemInformation(SystemHandleInformation)` for full
  enumeration (phase 2). Ship Unix-only first.

**API:**

```go
// Domain: safety
func NoFileDescriptorLeak(t T, tested func(), msgAndArgs ...any) bool
```

Same wrapping pattern as `NoGoRoutineLeak` — the `tested()` function
is already executed in a controlled way, so layering fd snapshots
before/after is cheap.

**Implementation sketch:**
- `internal/fdleak/` package (separate from `internal/leak/`)
- `Snapshot() → map[int]fdInfo` where `fdInfo` includes the symlink
  target from `/proc/self/fd/N` (socket, pipe, file path, etc.)
- `Diff(before, after) → []LeakedFD` with fd number + target
- Filter out fds opened by the runtime/test harness between snapshots
  (e.g., epoll fds, pipe fds for signal delivery) — this is the Unix
  equivalent of goleak's filter lists, but the set is small and stable
- Error message shows leaked fd number + what it points to:
  ```
  found 2 leaked file descriptors:
    fd 7: /tmp/testdata/unclosed.txt
    fd 9: socket:[12345]
  ```

### NoResourceLeak (v2.4, stretch goal)

A combined assertion that runs both checks in a single `tested()` call:

```go
// Domain: safety
func NoResourceLeak(t T, tested func(), msgAndArgs ...any) bool
```

Internally:
1. Snapshot open fds
2. Run `tested()` inside pprof-labeled goroutine (same as `NoGoRoutineLeak`)
3. After return: check for leaked goroutines (labels) AND leaked fds (diff)
4. Report both in a single failure message if either (or both) leak

This avoids running `tested()` twice and gives a comprehensive
resource leak report in one assertion.

### Not planned

- Memory leak detection — not feasible without instrumenting allocation
  sites; Go's GC handles reachability correctly, "leaks" are design
  issues best caught by static analysis and heap profiling (pprof)
- `VerifyTestMain` variant — not needed; labels support `t.Parallel()`
- Context propagation to `tested()` — future enhancement if demand arises
