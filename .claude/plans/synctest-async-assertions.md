# Plan: Migrate async assertions to testing/synctest (Approach A)

**Status:** DESIGN APPROVED — ready for implementation

**Decision log:**
- 2026-04-16: Identified synctest as a way to eliminate flakiness in async tests
- 2026-04-17: Compared two opt-in approaches (typed wrapper vs. variadic option)
- 2026-04-17: **Picked A (typed condition wrapper)** — narrow, reversible, aligned with
  library's primitive-not-configuration ethos; B was rejected due to unbounded drift
  risk from opening a universal options slot

## Goal

Leverage Go 1.25's `testing/synctest` to eliminate timing-dependent flakiness
in async assertions (`Eventually`, `Never`, `Consistently`, `EventuallyWith`).
Activation is **opt-in per call** via typed condition wrappers.

Inside a synctest bubble, `time.Ticker`, `time.After`, `context.WithTimeout`
use a fake clock that advances only when all goroutines are durably blocked.
This makes polling loops deterministic.

## User-facing UX

```go
// Real-time polling (existing behavior, unchanged)
assert.Eventually(t, func() bool { return isReady() }, 5*time.Second, 100*time.Millisecond)

// Fake-time polling (new)
assert.Eventually(t, assert.WithSynctest(func() bool { return isReady() }),
    5*time.Second, 100*time.Millisecond)
```

## Design reference: trial1

See `internal/assertions/trial1/` for a working minimal prototype covering `Eventually`
with `WithBubble` / `WithBubbleContext`. The final implementation will:
- Rename `WithBubble` → `WithSynctest` (clearer intent)
- Cover all four async assertions
- Add comprehensive fake-time validation tests

## synctest API recap (Go 1.25)

```go
func synctest.Test(t *testing.T, f func(*testing.T))  // runs f in a bubble
func synctest.Wait()                                   // blocks until bubble idle
```

Key facts:
- `time.Sleep`, `time.Ticker`, `time.After`, `context.WithTimeout` use fake time
- Goroutines in the bubble **cannot escape**: `Test` blocks until all exit
- Inside a bubble: `t.Run()`, `t.Parallel()`, `t.Deadline()` panic
- Durably blocking (advances fake clock): `time.Sleep`, channels, `sync.Cond.Wait`
- Non-durably blocking (does NOT advance): `sync.Mutex`, network I/O, syscalls
- `internal/synctest.IsInBubble()` exists but is NOT public

## Type design

Introduce four named types (aligned with existing `Conditioner` / `CollectibleConditioner`):

** Document the type definitions.

```go
// For Eventually / Never / Consistently
type WithSynctest        func() bool
type WithSynctestContext func(context.Context) error

// For EventuallyWith
type WithSynctestCollect        func(*CollectT)
type WithSynctestCollectContext func(context.Context, *CollectT)
```

Extend the type unions:

```go
type Conditioner interface {
    func() bool | func(context.Context) error | WithSynctest | WithSynctestContext
}

type CollectibleConditioner interface {
    func(*CollectT) | func(context.Context, *CollectT) |
    WithSynctestCollect | WithSynctestCollectContext
}
```

Helper constructors (so users don't need to write the type literal):

```go
func WithSynctest(f func() bool) WithSynctest { return WithSynctest(f) }
func WithSynctestContext(f func(context.Context) error) WithSynctestContext { return WithSynctestContext(f) }
// ... etc.
```

Wait — the constructor signature collides with the type name. Two options:
- Name the function differently: `Synctest(func() bool) WithSynctest` (type-as-noun, function-as-verb)
- Let users use type conversion: `assert.WithSynctest(cond)` works without a helper if we
  rename the type (e.g. `type synctestCondition func() bool` + export `WithSynctest = synctestCondition`)

Going with **`Synctest(f)` as constructors** — keeps the named types exported and the
call site short. Revisit during implementation if a better naming shakes out.

## Activation logic

In each async assertion's internal function:

```go
func eventually[C Conditioner](t T, condition C, timeout, tick time.Duration, msgAndArgs ...any) bool {
    wantsBubble, cond := makeCondition(condition, false)
    tt, canBubble := t.(*testing.T)

    if canBubble && wantsBubble {
        var result bool
        synctest.Test(tt, func(st *testing.T) {
            result = pollCondition(st, cond, timeout, tick, msgAndArgs...)
        })
        return result
    }
    return pollCondition(t, cond, timeout, tick, msgAndArgs...)
}
```

When `t` is not a `*testing.T` (e.g. a mock, `CollectT`), falls back to real-time
polling silently. Worth documenting: **synctest variants require a real `*testing.T`;
they no-op on mocks.**

## Scope of changes

| Component | Change |
|-----------|--------|
| `internal/assertions/condition.go` | Add 4 wrapper types, extend `Conditioner`/`CollectibleConditioner` unions, update `makeCondition`/`makeCollectibleCondition` to return a `wantsBubble` flag, add bubble activation in `eventually`/`never`/`consistently`/`eventuallyWithT` |
| `internal/assertions/condition_test.go` | Add tests for every wrapper type; validate fake-time behavior for every practical concern (see below) |
| `codegen/internal/scanner/` | Verify scanner doesn't choke on the new types in the type union |
| `codegen/internal/generator/templates/` | Generated signatures don't change (union dispatch is internal) |
| `assert/`, `require/` | Regenerated; re-exports of `WithSynctest` types/constructors |
| Docs | `Eventually`, `Never`, `Consistently`, `EventuallyWith` docstrings get a new "# Synctest" section |

## Critical finding from trial1 validation (2026-04-17)

**Channels consumed inside a bubble MUST be created inside the bubble.**

The initial trial1 prototype created `conditionChan` and `doneChan` in
`newConditionPoller` (outside the bubble). The polling tests hung: all
goroutines were durably blocked in selects, but fake time never advanced.

Root cause: `synctest` only considers a goroutine durably blocked if it is
blocked on a bubble-owned primitive. Channels created outside the bubble
are not bubble-owned — receives on them from inside the bubble do NOT
count as durably blocking. So the bubble waits forever for all goroutines
to reach a durably-blocked state, and fake time stalls.

**Fix:** create channels lazily inside `pollCondition` (runs inside the bubble).

This has a production-code implication: either:
- Move `conditionChan` and `doneChan` creation from `newConditionPoller`
  into `pollCondition` (trial1's fix), OR
- Keep `newConditionPoller` but have it lazy-init the channels on first
  use from inside the bubble.

This finding generalizes: any primitive the polling goroutines block on
(channels, timers) must be bubble-owned. `p.ticker = time.NewTicker(tick)`
is already created inside `pollCondition`, so it's fine.

## Practical concerns requiring thorough testing

These are the things that could silently break when the polling loop runs in a bubble.
Each needs a dedicated test.

### 1. Fake-time ticker

```go
// In a bubble, time.Ticker fires on fake time. The condition should be
// called at fake-tick intervals, not wall-clock intervals.
```
- Test: `assert.Eventually(t, WithSynctest(cond), 1*time.Hour, 1*time.Minute)` should
  complete in milliseconds of real time if `cond` eventually becomes true
- Test: count the number of condition calls matches `timeout/tick` exactly (deterministic)

### 2. `context.WithTimeout` on fake clock

```go
// cancellableContext() uses context.WithTimeout(parentCtx, timeout).
// Inside a bubble, the timeout fires on fake time — good.
```
- Test: timeout fires exactly at the fake `timeout` boundary
- Test: `ctx.Err()` is `context.DeadlineExceeded` at the expected fake-time mark

### 3. `context.WithoutCancel` (used by Never/Consistently)

```go
// For Never/Consistently, we detach from parent cancellation.
// Does WithoutCancel interact correctly with the bubble's fake clock?
```
- Test: `Never(t, WithSynctest(cond), ...)` succeeds if cond stays false for fake `timeout`
- Test: parent cancellation (via `t.Context()`) still fails `Never` appropriately

### 4. Parent context cancellation propagation

```go
// parentContextFromT(t) extracts t.Context() — does this context behave
// like a fake-time context inside the bubble?
```
- Test: parent cancellation from outside the bubble — does it reach the polling loop?
  (Expected: yes, context cancellation crosses bubble boundaries)
- Test: `contextualizer` interface path works identically

### 5. Panic recovery inside the bubble

```go
// recoverCondition() wraps the condition in defer+recover. Does panic
// unwinding work correctly through the bubble's goroutine management?
```
- Test: panicking condition is recovered and becomes an error (same as non-bubble)
- Test: `errConditionPanicked` sentinel is preserved through the bubble
- Test: panic recovery path does not leak goroutines (bubble-exit would fail)

### 6. `runtime.Goexit` from `CollectT.FailNow` / `Cancel`

```go
// CollectT.FailNow() and Cancel() call runtime.Goexit. This aborts the
// current tick goroutine. Does Goexit interact correctly with synctest's
// goroutine tracking?
```
- Test: `FailNow` in a `WithSynctestCollect` condition causes tick retry (not bubble panic)
- Test: `Cancel` / `Cancelf` aborts the whole assertion before the bubble timeout

### 7. Goroutine cleanup (bubble-exit invariant)

```go
// synctest.Test BLOCKS until all goroutines in the bubble exit.
// pollCondition spawns 2 goroutines guarded by wg.Wait().
// This already aligns — but we must verify no goroutine ever leaks.
```
- Test: on every success path, the bubble exits cleanly (no hang)
- Test: on every failure path, the bubble exits cleanly
- Test: on panic recovery, the bubble exits cleanly

### 8. Interactions with `t.Run()` (forbidden inside bubble)

```go
// synctest panics if t.Run() is called inside a bubble.
// TestConditionEventuallyTimeout currently has a nested t.Run().
// If we put that outer test in a bubble, the nested t.Run will panic.
```
- Review: our own tests do not mix `WithSynctest` with inner `t.Run` subtests
- Possibly: add a test asserting that the panic-from-nested-`t.Run` is caught
  gracefully (or document the constraint and trust users)

### 9. Real I/O inside a bubbled condition

```go
// If a user's condition does a real HTTP call, the goroutine blocks
// non-durably, so fake time doesn't advance while it's blocked. The
// timeout could never fire. Is this what we want?
```
- Test: condition that does a real network call inside `WithSynctest` — behavior
  should be documented (best practice: don't; `WithSynctest` is for pure compute /
  `time.Sleep` / channel-based code)
- Document this clearly: **`WithSynctest` is only for conditions that do NOT
  perform real I/O.**

### 10. Subroutine `sync.WaitGroup.Go` + Goexit

```go
// executeCondition uses a child wg.Go() wrapping the condition call, to
// guard against runtime.Goexit. Does wg.Go's goroutine participate in
// the bubble correctly?
```
- Test: bubble exits cleanly when condition calls `runtime.Goexit`
- Test: the inner `wg.Wait()` unblocks correctly when `Goexit` fires

### 11. Deterministic tick counting

```go
// Because fake time is deterministic, we can make strong guarantees
// about the number of ticks that elapse — eliminating the "counter
// could be 4, 5, or 6" fuzzy assertions currently in our test suite.
```
- Refactor `testEventuallyWithShouldCompleteWithFalse` (currently tolerates ±1 on counter)
  to use `WithSynctest` and assert an exact count. This is the motivating demo.

### 12. `synctest.Wait()` — do we need to call it explicitly?

```go
// synctest.Wait() blocks until all goroutines in the bubble are durably
// blocked. Our polling loop self-coordinates via channels/ticker. Do we
// ever need to call Wait() explicitly, or does the natural structure
// suffice?
```
- Investigation task: try without explicit `Wait()` calls first. If deterministic
  counting requires `Wait()` at specific points (e.g. after `conditionChan <-`),
  add it minimally.

### 13. Validated from trial1 (2026-04-17)

- ✅ Fake-time ticker fires deterministically (concerns 1, 11)
- ✅ Panic recovery works through the bubble (concern 5)
- ✅ No goroutine leak; bubble exits cleanly even with slow conditions (concern 7)
- ✅ Fallback path works for non-`*testing.T` (bubble activation is strict)
- ✅ Race-clean under `-race`
- ⚠ Requires the channel-creation-inside-bubble fix (captured above)

## Test strategy

### Two orthogonal test axes

**Axis 1: Behavioral dual-path tests (structural enforcement)**

Every async-assertion behavioral test is written once and automatically
runs under both real time and fake time via a shared dual-runner:

```go
// runDualPath runs fn twice: once with real time, once inside a bubble.
// fn uses plain (non-wrapped) conditions and a mock T so failures can be
// verified without polluting the outer test.
//
// NOTE: fn MUST NOT call t.Parallel(). synctest.Test forbids it inside
// a bubble. This restriction is transparent to real users of the async
// assertions (users do not call t.Parallel() inside condition functions).
func runDualPath(t *testing.T, name string, fn func(t *testing.T)) {
    t.Helper()
    t.Run(name+"/real-time", fn)
    t.Run(name+"/synctest", func(t *testing.T) {
        synctest.Test(t, fn)
    })
}
```

Key insight: **the bubble does not need to be activated via `WithSynctest`
for the internal polling to run under fake time.** If the test harness is
already inside a bubble (via `synctest.Test`), then `Eventually(mock, plainCond, ...)`
naturally uses fake time because `time.NewTicker` and the polling channels
(created inside `pollCondition`) are bubble-owned.

This decouples two concerns:
- **Failure capture** — done by passing a mock T (no bubble activation via API)
- **Fake time** — done by the test harness bubble (independent of API activation)

Benefits:
- Every subtest runs both paths by construction; forgetting is visible at review
- Failure cases can still be verified with mock-captured errors under fake time
- Deterministic counter assertions become possible for every case

**Axis 2: API-level wrapper detection tests (explicit)**

A smaller focused suite per assertion verifies that `WithSynctest(cond)`
correctly activates the internal bubble when `t` is `*testing.T`. Mirrors
the existing trial1 `TestBubble_*` tests. These exercise the surface
users actually write, not just the internal behavior.

### Example dual-runner test case shape

```go
type asyncCase struct {
    name    string
    // build returns a fresh condition + an inspector for scratch state.
    build   func() (cond func() bool, getCounter func() int)
    timeout time.Duration
    tick    time.Duration
    wantOk  bool
    verify  func(t *testing.T, mock *errorsCapturingT, counter int)
}

func (c asyncCase) Run(t *testing.T) {
    runDualPath(t, c.name, func(t *testing.T) {
        // DO NOT call t.Parallel() here — forbidden inside synctest bubble.
        mock := new(errorsCapturingT)
        cond, getCounter := c.build()
        ok := Eventually(mock, cond, c.timeout, c.tick)
        if ok != c.wantOk {
            t.Errorf("want ok=%v, got %v", c.wantOk, ok)
        }
        c.verify(t, mock, getCounter())
    })
}
```

### Why not use `WithSynctest` in the dual-runner?

Using the wrapper in the dual-runner would require `*testing.T` (not a mock),
which means failure-path tests couldn't intercept error messages. The
test-harness bubble wrapping sidesteps that by giving us fake time plus
mock-captured failures in the same test — a strict improvement for coverage.

The wrapper itself is still tested in Axis 2.

## Implementation phases

### Phase 0: validate trial1 end-to-end
- [ ] Write a few synctest-based tests against the trial1 prototype
- [ ] Confirm: (1) fake time works, (2) no goroutines leak, (3) panic recovery works
- [ ] Check gotchas: parent context, `WithoutCancel`, `Goexit`

### Phase 1: promote trial1 to production
- [ ] Move types to `internal/assertions/condition.go` (rename to `WithSynctest*`)
- [ ] Extend both type unions (`Conditioner`, `CollectibleConditioner`)
- [x] Update `makeCondition` to return `(wantsBubble bool, cond func(...))`
- [x] Update `makeCollectibleCondition` analogously
- [x] Add bubble activation in all four async assertion entry points
- [x] Remove `trial1/` directory

### Phase 2: test every practical concern
- [x] Dual-path runner (`runDualPath`) built and applied to Eventually / Never / Consistently / EventuallyWith behavior tests
- [x] API-level tests for all four `WithSynctest*` wrapper types
- [x] Panic recovery + slow-condition no-leak tests under fake time
- [x] Fallback path (mock T with WithSynctest) verified
- [x] All existing tests still pass (no regressions for real-time path)
- [ ] Migrate the fuzzy ±1 tolerances in existing tests to exact counts under fake time (deferred; the dual-path runner already covers behavior parity)

### Phase 3: codegen + docs
- [x] Run `go generate ./...`; `assert/` and `require/` regenerate cleanly (no generator changes needed)
- [x] Codegen scanner handles the new types in the union without adaptation
- [x] Re-exports of `WithSynctest*` and `NeverConditioner` from `assert/` and `require/` generated as aliases automatically
- [x] Docstring on `Eventually`, `Never`, `Consistently`, `EventuallyWith` all reference the Synctest opt-in (propagated via codegen)
- [x] Added "Deterministic polling with synctest (opt-in)" subsection in `docs/doc-site/usage/EXAMPLES.md`
- [x] Added "New Types — synctest opt-in" table and behavior-change entries in `docs/doc-site/usage/CHANGES.md`
- [x] Marked v2.5 synctest item as done in `docs/doc-site/project/maintainers/ROADMAP.md`
- [x] Testable examples `ExampleWithSynctest*` added to `assert/assert_adhoc_example_9_test.go` and `require/require_adhoc_example_9_test.go` (5 each)
- [ ] Reserved for next round: `docs/doc-site/usage/TRACKING.md` (no corresponding upstream issue)

### Phase 4: lint + CI
- [x] `golangci-lint run --new-from-rev master` clean (0 issues)
- [x] `go test work ./...` passes
- [ ] Verify on `stable` and `oldstable` Go versions — **but**: `oldstable` may be Go 1.24,
  which doesn't have `testing/synctest`. Decision point:
  - Drop `oldstable` support temporarily, OR
  - Build-tag the synctest path so it compiles out on Go 1.24
  - The current `go.mod` says `go 1.25` — if that's already the floor, no issue

## Open questions for implementation

1. **Constructor naming**: `Synctest(f func() bool) WithSynctest` vs letting users write
   `WithSynctest(cond)` via type conversion directly. Pick during implementation based on
   which reads cleaner at call sites.

2. **Re-exports**: `assert.WithSynctest` and `require.WithSynctest` — need generator
   support or hand-written aliases.

3. **Go version**: confirm the `oldstable` situation. `go.mod` says 1.25 but CI may still
   exercise 1.24 via `actions/setup-go`.

4. **Documentation placement**: attention point about "not for real I/O" — main docstring
   of `Eventually` or separate synctest section? Probably both.

## Non-goals (explicitly out of scope)

- **Option system for sync assertions** — rejected as unbounded drift risk
- **Auto-detection of bubble context** — `internal/synctest.IsInBubble()` is not public
- **Build tag–based activation** — too inflexible for per-call opt-in

## References

- `internal/assertions/trial1/` — working prototype
- `internal/assertions/condition.go` — production code
- `internal/assertions/condition_test.go` — existing tests
- Go 1.25 release notes: `testing/synctest` proposal
