# Plan: Redesign `CollectT` — interface shape + `testing.TB` interop

**Status:** EXPLORING — no option is clearly best; decision deferred.

**Motivation:** upstream issue [stretchr/testify#1862](https://github.com/stretchr/testify/issues/1862)
— user helpers typed as `testing.TB` cannot be invoked inside
`EventuallyWithT` callbacks because `*CollectT` does not (and cannot)
implement `testing.TB`.

**Position in the fork roadmap:** covered by the vague "new candidate
features from upstream" line item on v2.5+. No firm date.

## The problem, restated

Users routinely write custom assertion helpers typed as `testing.TB`:

```go
func AssertProtoEqual(t testing.TB, want, got any, msgAndArgs ...any) {
    t.Helper()
    if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
        assert.Fail(t, fmt.Sprintf("not equal (-want +got):\n%s", diff), msgAndArgs...)
    }
}
```

They want to invoke such helpers from inside an `EventuallyWith` callback:

```go
EventuallyWith(t, func(c *CollectT) {
    AssertProtoEqual(c, want, got)  // ← compile error today
}, ...)
```

Today this fails because `*CollectT` does not satisfy `testing.TB`.

## The hard constraint

`testing.TB` has an unexported `private()` method:

```go
type TB interface {
    // ...
    private()
}
```

**Nothing outside the `testing` package can satisfy `testing.TB` from
scratch.** The only exception is a type that embeds `*testing.T` (or
`*testing.B` / `*testing.F`) — because it inherits `private()` as a
promoted method.

This is a deliberate Go stdlib design choice (for forward compatibility
of the `testing.TB` interface). It means: **we cannot write a pure
routing proxy** that satisfies `testing.TB` without a real `*testing.T`
inside it.

## Goroutine-safety subtext (dolmen's concern upstream)

Per Go's `testing.T` docs:

> A test ends when its Test function returns or calls any of the methods
> `T.FailNow`, `T.Fatal`, `T.Fatalf`, `T.SkipNow`, `T.Skip`, `T.Skipf`.
> Those methods, as well as `T.Parallel`, must be called only from the
> goroutine running the Test function.

`EventuallyWith` polling runs the callback in a spawned goroutine (per-tick
wrap). So any testing.TB method we expose that bottoms out in the real
`*testing.T` from a spawned goroutine is **technically undefined behavior**
— even if it "works" today.

Mitigation: our `CollectT.FailNow` and `Cancel` use `runtime.Goexit` on the
*spawned* goroutine only, never calling the real `t.FailNow`. Any
TB-interop design must preserve this property.

## Design space

### A. Interface widening only — introduce `assert.TB`-like interface

Add a broader interface than our current `T`:

```go
type TB interface {
    Errorf(format string, args ...any)
    Helper()
    FailNow()
    Logf(format string, args ...any)
    Failed() bool
    Name() string
    // ... whatever subset we choose
}
```

Users retype their helpers from `testing.TB` → `assert.TB`. `*CollectT`
already satisfies it (or we extend it to).

- **Pros**: zero runtime complexity, non-breaking, gradual migration.
- **Cons**: does nothing for users unwilling to retype their helpers;
  doesn't close the gap with helpers that genuinely need `testing.TB`
  (e.g. anything taking a `*testing.T` for `TempDir`, `Context`, etc.).

### B. Embed `*testing.T` in `CollectT` (make CollectT concretely testing.TB)

Change `CollectT` struct to embed `*testing.T`. By embedding, it inherits
`private()` and satisfies `testing.TB`. Override the methods we need to
collect (Errorf, Fail, FailNow, Fatal…) to route via the collector.

- **Pros**: drop-in — `testing.TB` helpers "just work".
- **Cons**:
  - Breaking: CollectT construction changes (needs a real `*testing.T`).
  - Mocks can't be used with `EventuallyWith` anymore (must pass
    `*testing.T`), or we add a nil-embedded degraded mode.
  - Lots of testing.TB methods to override carefully. Semantics for
    `Skip*`, `Cleanup`, `Setenv` in a per-tick callback are fuzzy.
  - Goroutine safety for forwarded methods remains a grey area.

### C. Opt-in `CollectTB` variant — same pattern as `WithSynctest`

Keep the current `CollectT` unchanged. Introduce a parallel type
`CollectTB` used only when the caller opts in (e.g. via a
`WithCollectTB` condition-wrapper echoing `WithSynctest`):

```go
EventuallyWith(t, WithCollectTB(func(c *CollectTB) {
    AssertProtoEqual(c, want, got)  // c satisfies testing.TB
}), ...)
```

- **Pros**: narrow, non-breaking, mirrors our existing opt-in idiom.
- **Cons**: doubles the wrapper-type surface; two condition-variant types
  to remember; users still migrate to a new signature to get interop.

### D. Turn `CollectT` into an interface (proposed by @fredbi upstream)

Make `CollectT` an interface; rename the current struct to unexported
`collectT`. Default impl does what today's CollectT does. An alternate
impl (internal, `collectTB`) embeds `*testing.T` and routes all TB methods
through the collector.

Selection is automatic based on `t.(*testing.T)`: if the caller passed a
real `*testing.T`, we construct `collectTB`; otherwise `collectT`.

Add a `TB() testing.TB` method on the `CollectT` interface:
- Returns `self` on `collectTB` — a routing proxy that satisfies
  `testing.TB` (via the embedded `*testing.T` it carries internally).
- Returns `nil` on `collectT` (no `*testing.T` available).

Users of TB-typed helpers:
```go
EventuallyWith(t, func(c CollectT) {  // interface now, not pointer-to-struct
    if tb := c.TB(); tb != nil {
        chancez.AssertProtoEqual(tb, want, got)
    }
    // direct methods still preferred for common ops:
    assert.Equal(c, want, got)
    c.Logf("attempting...")
}, ...)
```

- **Pros**:
  - Everyone migrates once (`*CollectT` → `CollectT`), mechanical sed.
  - Zero ceremony at call sites when using direct methods.
  - Automatic selection — no user-facing opt-in required.
  - TB() is an explicit escape hatch, documented as such.
  - Leaves room for user-provided CollectT mocks if ever needed.
- **Cons**:
  - Breaking (signature change for every existing callback).
  - Must implement the `collectTB` proxy carefully: many TB methods need
    overrides; Skip/Cleanup/Setenv semantics TBD.
  - Method-set divergence between the two impls (collectT and collectTB)
    is invisible at the type level — could surprise users who rely on
    `TB()` returning non-nil.
  - `TB()` is always on the interface but meaningful only half the time.

## The `TB()` routing subtlety

Crucial point — discovered mid-discussion: if `TB()` returned the raw
`*testing.T`, helpers invoking it would bypass the per-tick collection
and fail the whole test immediately, defeating retry semantics.

**The `TB()` return value must be a routing proxy** that:
- Satisfies `testing.TB` (by embedding `*testing.T`)
- Overrides `Errorf`, `Error`, `Fail`, `FailNow`, `Fatal`, `Fatalf`,
  `Failed` to route through the collector (per-tick collection + Goexit
  for abort methods)
- Forwards read-only methods (`Name`, `TempDir`, `Context`) to the
  embedded `*testing.T`
- Has a decision for `Skip*`, `Cleanup`, `Setenv` (panic? noop? forward?)

## Open design questions (from the conversation)

1. **Naming: keep `CollectT` or rename to `Collector`?**
   - Leaning: keep `CollectT`. The `*` → (none) migration signal is
     enough; rename would create double churn.

2. **Which methods on the `CollectT` interface itself?**
   - Agreed: `Errorf`, `Helper`, `FailNow`, `Cancel`, `Cancelf`, `Logf`,
     `Failed`, `TB()`.
   - Open: should we include `Name`?
   - Open: where does `Log` (without f) go?

3. **How does `collectTB` handle test-lifetime methods?**
   - `Cleanup`: per-tick cleanup piles up — probably **panic / forbid**.
   - `Setenv`: racy across ticks, cross-goroutine — probably **panic /
     forbid**.
   - `Skip*`: skipping the whole test from a tick is surprising — probably
     **panic / forbid** with a clear message, or forward to the embedded T
     only for `Skip`/`Skipf` (non-fatal).
   - Each of these is a judgment call.

4. **`Logf` / `Log` routing:**
   - Via the collector (captured, printed on failure)? Via the real
     `*testing.T` (immediate output)? Leaning toward real `*testing.T`
     for diagnostic visibility.

5. **Default `TB()` — nil or degraded stub?**
   - Leaning: **nil**. Clear failure signal; users nil-check.

## Why this is hard (a.k.a. "all options look terrible")

Every option has a visible cost:

- **A** doesn't fix the root problem for users unwilling/unable to migrate
  helpers.
- **B** forces breaking construction changes and bakes goroutine-safety
  risk into the default type.
- **C** is safe but doubles the wrapper surface and leaves us juggling
  two parallel variants forever.
- **D** is the most elegant conceptually but introduces breaking changes,
  invisible semantic divergence between impls, and non-trivial proxy
  implementation complexity.

Additionally, the Go stdlib's `testing.TB.private()` is a forced
architectural constraint we can't engineer around.

## Possible paths forward (not commitments)

1. **Do nothing** — document the workaround (retype helpers as our minimal
   `T` interface, or write thin adapters per-helper). Chancez's specific
   use case is doable with a 3-line adapter wrapping his helper.

2. **Just A** — widen our `T` into `TB`-like, document the recommended
   pattern, close the issue with "here's how we'd like people to write
   helpers in this fork." Low cost, partial solve.

3. **A + later D** — land A now (non-breaking), revisit D for v3.

4. **D directly** — bold release, one-shot migration, capture the upstream
   design debate's intent.

5. **C** — if D feels too big and A feels too weak.

## Artifacts linked

- Upstream issue: https://github.com/stretchr/testify/issues/1862
- @fredbi's comment (2026-03-20) suggesting CollectT as an interface
- @dolmen's comment on goroutine-safety constraints of `testing.T` methods

## Decision log (so far)

- 2026-04-17: Options A-D identified and catalogued.
- 2026-04-17: Interface-rename from `*CollectT` to `CollectT` is a
  shallow mechanical break, not a deep semantic one.
- 2026-04-17: `TB()` must return a routing proxy (not the raw
  `*testing.T`) to preserve retry semantics.
- 2026-04-17: **No commitment to a direction**. Fred's gut: "all look
  terrible."
