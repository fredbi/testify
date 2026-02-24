# Goroutine Leak Detection: Implementation Analysis

**Date:** 2026-01-30

## Context

Designing a goroutine leak detection assertion for go-openapi/testify. Studied five implementations plus the Go 1.26 experimental GC-based approach.

## The Family Tree

```
Go stdlib: net/http/main_test.go (the ancestor, ~2013)
  |
  +-- CockroachDB pkg/util/leaktest/  (fork, adds snapshot-diff + more filters)
  |     |
  |     +-- fortytw2/leaktest  (fork of CRDB, standalone package)
  |
  +-- Uber goleak  (ground-up rewrite, same fundamental technique)
  |
  +-- Lantern grtrack  (minimal ID-only variant)
  |
  +-- Go 1.26: runtime/goroutineleakprofile  (GC-based, fundamentally different)
```

All stack-based variants descend from the same ~30-line `interestingGoroutines()` pattern.

## The Fundamental Technique

Every stack-based implementation does the same thing:

1. Call `runtime.Stack(buf, true)` to get all goroutine stacks as text
2. Parse the text to identify goroutines
3. Filter out known system/testing goroutines
4. Report whatever remains as leaked

The differences are in parsing depth, filtering strategy, retry mechanism, and API surface.

## Implementation Comparison

### net/http/main_test.go (the original)

- **Size**: ~80 lines (not a library, test-internal)
- **Two mechanisms**:
  - `goroutineLeaked()`: runs in TestMain after all tests. Checks for *any* interesting goroutine. 5 retries x 100ms = 0.5s max.
  - `afterTest(t)`: runs per-test via defer, checks for **specific known-bad patterns** (readLoop, writeLoop, httptest.Server, etc.). 2500 retries x 1ms = 2.5s max.
- **Parsing**: `strings.Cut(g, "\n")` discards header, matches substrings against stack body only. No ID extraction, no state parsing.
- **Filtering**: Hard-coded substrings for testing infrastructure + runtime goroutines.
- **t.Parallel()**: Explicitly incompatible. Uses a deliberately non-atomic `leakReported` flag so the race detector catches concurrent leak checks.
- **Key insight**: Domain-specific. It knows exactly what net/http creates and checks for those specifically. Not a general-purpose tool.

### CockroachDB leaktest

- **Size**: ~200 lines
- **Approach**: Snapshot-and-diff. Captures goroutine IDs at test start via `AfterTest(t)`, returns closure that diffs at end.
- **Parsing**: Uses `allstacks.Get()` (their own wrapper around runtime/pprof). Extracts goroutine IDs + full stack strings.
- **Filtering**: Significantly more hard-coded exclusions than stdlib: Sentry, OpenCensus, pgconn, TLS handshake (Go <1.23 bug), `log.flushDaemon`.
- **Retry**: 5s timeout, 50ms polling (fixed ticker).
- **API**: `defer leaktest.AfterTest(t)()`
- **Key insight**: Shows the maintenance burden of hard-coded filters. Every new dependency that spawns background goroutines needs a new exclusion string.

### fortytw2/leaktest (from CockroachDB)

- **Size**: ~171 lines
- **Approach**: Same snapshot-and-diff as CRDB, cleaned up for standalone use.
- **Parsing**: `strings.Split(buf, "\n\n")` then regex for goroutine header. Extracts ID + full stack string.
- **Filtering**: Hard-coded string matches (subset of CRDB's list).
- **Retry**: Fixed 50ms ticker, default 5s timeout, context-aware (`CheckContext`).
- **Buffer**: Fixed 2MB.
- **API**: `Check(t)`, `CheckTimeout(t, dur)`, `CheckContext(ctx, t)`
- **t.Parallel()**: Not supported (acknowledged).
- **Strengths**: Context-aware timeout. Good simplicity/usability balance.
- **Weaknesses**: Hard-coded filters are brittle across Go versions. No extensibility. Fixed buffer.

### Uber goleak

- **Size**: ~811 lines across 7 files (2 packages)
- **Approach**: Full stack parsing with rich introspection.
- **Parsing**: Custom `stackParser` with `Unscan()` lookahead scanner. Extracts goroutine ID, state, firstFunction, createdBy, allFunctions map.
- **Filtering**: 4 built-in categories (testing, syscall, stdlib/signal, trace) + 5 user option types (`IgnoreTopFunction`, `IgnoreAnyFunction`, `IgnoreCreatedBy`, `IgnoreCurrent`, `Cleanup`).
- **Retry**: Exponential backoff: 1us -> 2us -> 4us -> ... capped at 100ms, 20 attempts.
- **Buffer**: Adaptive: starts 64 KiB, doubles until stack fits.
- **API**: `Find()`, `VerifyNone(t)`, `VerifyTestMain(m)`
- **t.Parallel()**: Incompatible with `VerifyNone`; use `VerifyTestMain` instead.
- **Strengths**: Most complete filtering, rich stack introspection, well-maintained, adaptive buffer.
- **Weaknesses**: Complex parser (~350 lines) for a text format that could change between Go versions. The stack parsing is the most fragile part.

### Lantern grtrack

- **Size**: ~80 lines, single file
- **Approach**: Pure ID-based snapshot comparison.
- **Parsing**: `regexp.MustCompile("goroutine ([0-9]+)")` -- extracts IDs only.
- **Filtering**: None. Pure ID-based snapshot comparison.
- **Retry**: Fixed ticker polling (configurable interval + timeout).
- **Buffer**: Fixed 2MB.
- **API**: `Start()` returns checker closure.
- **Strengths**: Extremely simple. Format-immune (doesn't care about stack format changes).
- **Weaknesses**: No filtering -- false positives on any system goroutine that starts between snapshots. No diagnostics (just IDs).

### Go 1.26 runtime/goroutineleakprofile (experimental)

- **Size**: Runtime-level (not user code).
- **Approach**: GC-based reachability analysis. Fundamentally different from all stack-based tools.
- **How it works**:
  1. Only *runnable* goroutines serve as GC roots (not all goroutines)
  2. Standard GC marking proceeds from those roots
  3. Blocked goroutines on *reachable* primitives are marked "eventually runnable" -> new roots
  4. Fixed-point iteration until stable
  5. Unreached goroutines -> provably leaked
- **Output**: Runtime annotates goroutine header with `(leaked)`: `goroutine 42 [chan receive (leaked)]:`
- **Catches**: Goroutines blocked on channels/mutexes/conds where no running goroutine holds a reference to the other end. Zero false positives (theoretically).
- **Misses**: Spinning goroutines (infinite loop -- they're "runnable", so they're roots). Goroutines blocked on primitives that are still reachable but will never be used. Effectiveness reduced when heap references are heavily interconnected.
- **Status**: Experimental in Go 1.26, gated behind `GOEXPERIMENT=goroutineleakprofile`.
- **Test suite insight**: ~60% of GoKer tests are flaky -- not false positives but false *negatives* due to scheduling non-determinism. A goroutine might not have reached its blocking state when the GC runs. Compensated with GOMAXPROCS=1, asyncpreemptoff=1, and repetitions.
- **Key insight**: Solves a different problem. It answers "which goroutines are provably unreachable?" at a whole-program level. The testing use case is "did this test leave goroutines behind?" -- a snapshot-and-diff problem.

Reference Research Paper: https://dl.acm.org/doi/pdf/10.1145/3676641.3715990

## Go Runtime API Status (as of Go 1.24)

`runtime.Stack()` remains the only way to get all goroutine stacks programmatically. No structured API exists.

Relevant changes since Go 1.20:
- **Go 1.23**: Max stack depth for profiles raised 32 -> 128 frames. Indented error messages in tracebacks (minor format change -- affects text parsers).
- **Go 1.24**: No new goroutine inspection APIs.
- **Go 1.26**: Experimental goroutineleak profile (not usable for Go 1.24 target).

## Parsing Depth Tradeoffs

| Approach | Complexity | Fragility | Diagnostics | Extensibility |
|----------|-----------|-----------|-------------|---------------|
| ID-only snapshot (grtrack) | ~80 lines | None (format-immune) | Poor (just IDs) | None |
| Header-only parse (ID + state) | ~150 lines | Low (header format stable since Go 1.0) | Good (state + raw stack for errors) | Filter by state |
| Full stack parse (goleak) | ~800 lines | Medium (function name extraction breaks across versions) | Excellent | Full (any function, creator, etc.) |
| GC-based (Go 1.26) | Runtime-level | None (built-in) | Profile-level | N/A |

## The Middle Ground: Header-Only Parsing

The goroutine header format (`goroutine N [state]:`) has been stable since Go 1.0. A header-only approach would:

- Parse header to get ID + state
- Capture raw stack block for error reporting (no function name extraction from body)
- Enable state-based filtering (`"chan receive"`, `"syscall"`, `"running"`) -- covers the most useful built-in filters
- Support `IgnoreTopFunction` (first line after header is always the top function)
- Skip the fragile function-name parser (~200 lines in goleak)
- Use CockroachDB/stdlib substring matching on raw stack blocks for user-supplied filters

What you lose: `IgnoreAnyFunction` (match anywhere in stack) and `IgnoreCreatedBy` (match creator). In practice, most users only need `IgnoreTopFunction` or just let the built-in filters handle it.

## Retry Strategy Comparison

| Implementation | Strategy | Total wait |
|---------------|----------|------------|
| net/http goroutineLeaked | 5 x 100ms fixed | 0.5s |
| net/http afterTest | 2500 x 1ms fixed | 2.5s |
| CockroachDB | 50ms ticker, 5s deadline | 5s |
| fortytw2/leaktest | 50ms ticker, 5s deadline | 5s |
| goleak | Exponential 1us..100ms, 20 retries | ~2s |
| grtrack | Configurable ticker + timeout | Configurable |

Goleak's exponential backoff is the most sophisticated: fast for clean tests, patient for slow cleanup.

## t.Parallel() Incompatibility

Every implementation acknowledges this problem. The fundamental issue: with parallel tests, you can't distinguish "goroutine from a still-running test" from "leaked goroutine from a finished test."

Solutions:
- **goleak**: Offers `VerifyTestMain(m)` as alternative -- checks after ALL tests complete.
- **net/http**: Disables parallel tests when leak checking is active (setParallel skips t.Parallel in non-short mode).
- **Others**: Document the limitation and move on.

## Key Design Decisions for Our Implementation

1. **Parsing depth**: Header-only is the sweet spot. We get ID, state, and raw stack for diagnostics without the fragile function-name parser.

2. **Filtering**: Built-in state-based filters (testing, syscall, signal) + substring matching on raw stack body for user extensions. No need for goleak's full function extraction.

3. **Retry**: Exponential backoff (goleak's approach). Fast path for clean tests.

4. **Buffer**: Adaptive (goleak's approach). Start small, double as needed.

5. **API**: Snapshot-based (`defer NoLeakedGoroutines(t)()` or similar). Suite-level `TestMain` variant for parallel test compatibility.

6. **Context support**: Yes (from leaktest). Modern Go code expects context-based cancellation.

7. **Future-proofing**: When Go 1.26's goroutine leak profile stabilizes, the implementation could optionally use it as a backend while keeping the same API. The `(leaked)` annotation makes detection trivial.
