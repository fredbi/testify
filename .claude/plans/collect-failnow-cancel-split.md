# Plan: split `CollectT.FailNow` semantics into `FailNow` (tick-only) + `Cancel` (whole assertion)

Status: 📝 Draft — pending review
Owner: @fredbi
Target: v2.4 (within the "experimental phase" window where breaking changes are allowed)
Related: `internal/assertions/condition.go`, `EventuallyWith`, `CollectT`

## Problem statement

In the current `EventuallyWith` design, `CollectT.FailNow()` cancels the
polling context and aborts the whole assertion immediately. This makes the
typical pattern of writing `require.X(collect, …)` inside an
`EventuallyWith` block effectively useless: the very first failing tick
terminates the assertion instead of letting the poller retry on the next
tick.

This contradicts the spirit of the "eventually" pattern — assertions are
expected to fail initially and converge over time — and it also diverges
from `stretchr/testify`, where `require.*` inside `EventuallyWithT` only
fails the current evaluation and lets the poller try again.

## Design

Split the two intents into two distinct methods on `CollectT`:

| Method                  | Semantics                                                                                                       |
|-------------------------|-----------------------------------------------------------------------------------------------------------------|
| `FailNow()` (changed)   | Marks the **current tick** as failed and exits the condition goroutine via `runtime.Goexit()`. Polling retries. |
| `Cancel()` (new)        | Marks the assertion as cancelled, **cancels the polling context**, and exits via `runtime.Goexit()`.            |

### Behavioral contract

- `FailNow`:
  - Append a `errFailNow` sentinel to `c.errors`.
  - `runtime.Goexit()` (no `cancelContext()` call).
  - Inner goroutine unwinds, `conditionWg.Wait()` returns, poller loop continues, next tick fires.
  - If a later tick succeeds, the assertion succeeds; the failed-tick errors are discarded (current behavior — `lastCollectedErrors` is overwritten on each tick).
  - On timeout, the **last** failing tick's collected errors are reported via `onFailure → copyCollected`.

- `Cancel`:
  - Append a distinct `errCancelled` sentinel to `c.errors`.
  - Call `c.cancelContext()`.
  - `runtime.Goexit()`.
  - Inner goroutine exits, outer poller loop hits `ctx.Done()` → `failFunc` fires → assertion fails immediately.
  - The cancelled tick's errors are reported via `onFailure`.

### Sentinel errors

Define package-private sentinels in `internal/assertions/condition.go`:

```go
var (
    errFailNow   = errors.New("collect: failed now (tick aborted)")
    errCancelled = errors.New("collect: cancelled (assertion aborted)")
)
```

Rationale: keeping them distinguishable lets future tooling (and tests) tell
apart "this tick was aborted by require" from "the user explicitly cancelled
the whole assertion". They are not exported because users should not depend
on the marker shape; they should only observe behavior.

## Scope: 3 PRs

### PR 1 — Feature (godoc + tests)

Goal: ship the API change and prove it works. Self-contained, no doc-site
churn, no migration tool churn.

Files:

- `internal/assertions/condition.go`:
  - Define `errFailNow`, `errCancelled` sentinels.
  - Rewrite `CollectT.FailNow()` — drop `cancelContext()`, append `errFailNow`, `Goexit`.
  - Add `CollectT.Cancel()` — append `errCancelled`, call `cancelContext()`, `Goexit`.
  - Update `EventuallyWith` godoc (line 213+): rewrite the "Calling `CollectT.FailNow`…" sentence and mention `Cancel`.
  - Update the maintainer comment block on `CollectT` (lines 691–705): document the new split, remove the now-stale "FailNow no longer just exits the goroutine" note.
  - Update the `Examples:` doc-comment block on `EventuallyWith` to include at least one new failure example demonstrating `Cancel()` short-circuit semantics, and clarify the existing one is `FailNow`-style retry.

- `internal/assertions/condition_test.go` (new tests):
  - `TestEventuallyWith_FailNowRetries` — first N ticks call `require.X(collect, …)` failing; flip a flag so tick N+1 passes; assert `EventuallyWith` returns `true`.
  - `TestEventuallyWith_CancelShortCircuits` — first tick calls `collect.Cancel()`; assert assertion fails immediately (well under the timeout) and the collected errors include the cancellation marker (or at least a single recorded error from the cancelled tick).
  - `TestEventuallyWith_FailNowTimesOut` — every tick fails via `require.*`; assert assertion fails on timeout and the last tick's errors are reported on `t`.
  - Goroutine leak check around all three (use `internal/leak` if available).

### Actions run manually by Fred

- Regenerate `assert/`, `require/` via `go generate ./...` (so the new `Examples:` propagate through the codegen).
- Run: `go test work ./... -race`, `golangci-lint run --new-from-rev master`.
- commit, push, open PR, merge

PR description: link this plan, summarize the breaking change, give before/after snippets.

### PR 2 — Doc-site update (testable examples)

Goal: bring the user-facing docs into alignment with the new semantics, and
add testable examples that the doc-site renders.

Files to audit / update (already grep'd, list non-exhaustive — re-audit at PR time):

- `docs/doc-site/api/condition.md` — main `EventuallyWith` page. Rewrite the section that describes `FailNow` semantics. Add a `Cancel` section. Add two testable examples (retry-then-succeed, cancel-short-circuit).
- `docs/doc-site/usage/MIGRATION.md` — call out the behavior change explicitly: "If you previously used `FailNow` inside `EventuallyWith` to abort the whole assertion, switch to `Cancel`."
- `docs/doc-site/usage/CHANGES.md` — add a v2.4 entry.
- `docs/doc-site/usage/EXAMPLES.md` — add or update worked examples for both methods.
- `docs/doc-site/usage/TRACKING.md` — refresh if it tracks API surface.
- `docs/doc-site/project/maintainers/ROADMAP.md` — mark the item if listed; otherwise add a note under v2.4.
- `docs/doc-site/api/safety.md` — re-read line 193 area (mentions `runtime.Goexit` semantics in safety guards) and confirm the discussion still holds; cross-reference the new `Cancel` method.

Run the doc-site link-check + spellcheck + markdownlint MCPs before opening
the PR.

### Actions run manually by Fred

- check hugo site locally
- commit, push, open PR, merge

### PR 3 — Migration tool update (`hack/migrate-testify`)

Goal: help users coming from `stretchr/testify` land on the new semantics
without surprises.

Tasks:
- `hack/migrate-testify/rename_map.go` already maps `EventuallyWithT → EventuallyWith` etc. Verify no rename is needed for the methods themselves (`FailNow` keeps its name; `Cancel` is new, so no source-side rename either).
- Add an **advisory pass** in the migration tool (or a doc note in the tool's README) that flags occurrences of `collect.FailNow()` inside `EventuallyWith` blocks where the original `stretchr/testify` semantics matched our **old** behavior (assertion-aborting). Recommendation in the report: review whether `Cancel()` is intended.
  - Implementation can be a simple AST walk: find `*ast.CallExpr` whose `Sel.Name == "FailNow"` on a receiver of type `*assertions.CollectT` (or matching by import-path heuristic during migration when types are not resolvable).
  - Output: a non-fatal warning in the tool's report, not a rewrite — the safe default for stretchr users is the new `FailNow` semantics (which matches stretchr's `CollectT.FailNow`), so silent rewriting would be wrong.
- Add a fixture under the migration tool's testdata for both cases.

### Actions run manually by Fred

- check hugo site locally
- commit, push, open PR, merge

## Risks and rollback

- **Breaking change**. v2.4 is still in the "experimental phase" per the
  roadmap, so this is acceptable. Ensure CHANGES.md and the PR description
  call this out prominently.

OK.

- **Hidden callers in the wild**: anyone who relied on `FailNow` aborting
  the whole `EventuallyWith` will see longer test runs (poller retries
  until timeout) instead of a fast fail. The migration note in PR 2 +
  advisory in PR 3 mitigate this.

Not much of a concern actually: this change came from early adopters of the feature who reported our inappropriate setting.o

- **Rollback**: a single revert of PR 1 restores prior behavior. PR 2 and
  PR 3 are documentation/tooling only, safe to leave or revert
  independently.

OK.

## Out of scope

- Changing `Eventually` / `Never` / `Consistently` semantics — this plan only
  affects `EventuallyWith` + `CollectT`.

-> our first commit, which affects go routines already affects these, but on a slightly different path

- Adding goroutine/panic guards beyond the existing inner-goroutine wrap
  introduced in the `fix/eventuallyWith` branch.

-> we'll do it, but separately

## Open questions

- Should `Cancel()` accept a reason string (`Cancel(format string, args ...any)`)
  to enrich the reported error? Default proposal: yes — symmetric with
  `Errorf`, lets users explain *why* they aborted. Fine to defer to a
  follow-up if it complicates PR 1.

-> keep bare Cancel() for now. But good call: in a follow-up, we'll add Cancelf(string,...any).

- Should we also expose a non-Goexit form (`Failed()` that the user checks
  manually before `return`)? Probably no — the Goexit form matches require's
  ergonomics and is what users expect.

-> not for the moment. There is still not a large consensus on what CollectT should look like in the end
   (perhaps an interface). So definitely later on.
