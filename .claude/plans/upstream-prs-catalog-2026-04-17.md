---
name: Upstream scan — 2026-04-17
description: Delta since the 2026-02-06 catalog — new issues/PRs on stretchr/testify and our verdicts
type: plan
---

# Upstream PR Catalog — stretchr/testify (delta since 2026-02-06)

**Generated:** 2026-04-17
**Prior snapshot:** `.claude/plans/upstream-prs-catalog-2026-02-06.md`
**Tracking ledger:** `docs/doc-site/usage/TRACKING.md`

## TL;DR

As expected, not much has moved upstream. **9 new items** since the previous review, roughly half of which are mock-specific (not applicable to the fork). On the assert/require side, there are two genuinely interesting signals and one small actionable bug.

| Verdict bucket | Count | Items |
|---|---|---|
| 🔴 **Actionable — we have the same bug** | 1 | #1848 (Subset/NotSubset `%q` format) |
| 🎯 **Candidate — worth a design round** | 2 | #1859 (channel assertions), #1860/#1861 (ErrorAsType) |
| 📝 **Already covered (design-explored)** | 1 | #1862 (CollectT/TB interop — see `collectT-interface-tb-interop.md`) |
| ⛔ **Superseded / not a bug in our fork** | 3 | #1857, #1847, #1863 |
| ⛔ **Not applicable (mock-only)** | 3 | #1852, #1853, #1866, #1870 (merged list) |

## Summary by status

### 🔴 Actionable — we have the same bug

#### PR #1848 — `fix: use %#v instead of %q in Subset/NotSubset error messages`

- **Upstream filed:** 2026-02-10 (still open)
- **What it fixes:** `Subset`/`NotSubset` uses `%q` to render the subset/list, which produces broken output for non-string types (e.g., `%!q(bool=true)` for booleans, `'\x01'` for ints).
- **Our status:** We inherited the same bug. `internal/assertions/collection.go` lines **531, 808, 832** all use `truncatingFormat("%q", …)`.
- **Suggested action:** Small, self-contained fix — mirror the upstream change (switch to `%#v`). Update the corresponding test fixtures.
- **Fred (2026-04-17):** fix asap.

### 🎯 Candidates — worth a design round

#### Issue #1860 + PR #1861 — `ErrorAsType[E]` for Go 1.26+

- **Upstream filed:** 2026-03-13 (issue + PR, both open)
- **Proposal:** Mirror Go 1.26's `errors.AsType[E]` as `require.ErrorAsType[E]`, eliminating the `var x *T; require.ErrorAs(t, err, &x)` dance.
- **Fred's comment on PR:** "Yes interesting. The assertion expression is definitely lighter."
- **Constraints for our fork:**
  - Blocked on our Go version policy until we move to go1.26+ (currently go1.25 minimum).
  - Fits naturally with our generics-first direction — would live alongside `IsOfTypeT`, `ErrorAs`, etc.
- **Suggested action:** Add to the v2.6 wishlist. Implement once we bump the minimum to go1.26.
- **Fred (2026-04-17):** interesting, but definitely not great — it's just one more assertion. To be considered later. Would need a **go1.26 build tag** and **specific codegen provisions** to support conditional assertion generation.

#### Issue #1859 — Channel assertions (synctest-friendly)

- **Upstream filed:** 2026-03-09 (open, under discussion)
- **Proposal:** Helpers like `requireBlocked[T]` / `requireUnblocked[T]` for validating channel state. Explicitly motivated by `testing/synctest`.
- **Dolmen's reservation:** Signatures don't match the `assert`/`require` family (they return `T` rather than `bool`). No concrete fit-for-family design yet.
- **Why relevant to us:** We just shipped the synctest opt-in. Channel-state assertions pair naturally with fake-clock polling.
- **Suggested action:** Let upstream iterate on the signature shape. If a clean design emerges, adopt. Otherwise, consider a minimal `ChanBlocks` / `ChanUnblocks` pair scoped to our idioms. Park for now.

### 📝 Already covered (design-exploration in place)

#### Issue #1862 — CollectT/TB interop

- **Upstream filed:** 2026-03-13
- **Already catalogued:** `.claude/plans/collectT-interface-tb-interop.md` (status: EXPLORING, no commitment).
- Fred's upstream comment (interface idea) and Dolmen's goroutine-safety concern are both captured in our plan. No new information needed.
- **Fred (2026-04-17):** still exploring — not really sure how to make `CollectT` a great feature from where it started. No rush.

### ⛔ Superseded or not a bug in our fork

#### PR #1857 — Regexp/NotRegexp panic on invalid regex

- **Already handled:** Implemented via upstream #1818 in v2.1 (see TRACKING.md). Our fork doesn't panic on invalid patterns.

#### PR #1847 — `fix: stabilize relative error calculation in InEpsilon`

- **Upstream filed:** 2026-02-06
- Fred already left a comment on the parent issue (#1839) pointing to our more thorough InEpsilon spec (zero handling, NaN/±Inf semantics). The upstream fix is narrower (normalizes by max-absolute value) and doesn't address the edge cases we cover.
- **Verdict:** Superseded by our implementation. No change needed.

#### PR #1863 — `Fix: float32 and float64 comparisons with EqualValues`

- **Upstream filed:** 2026-03-17
- **Proposal:** Round float32 values to 6 decimal places before comparing to float64 in `EqualValues`.
- **Concerns:** Rounding silently changes comparison semantics. 6-digit heuristic is arbitrary (float32 precision is ~7.2 decimal digits; 6 is a rough lower bound). Callers who genuinely want "close enough" should use `InDelta`/`InEpsilon`.
- **Verdict:** Monitor upstream. Not adopting. The root issue (#1576) is better solved by pointing users at `InDelta`/`InEpsilon`.
- **Fred (2026-04-17):** if such a *symmetric* comparison is desirable at all, it should land as a **new assertion** (e.g. `NumberClose(t, a, b)` or similar) — **not as a semantics change to `InEpsilon`**. Changing `InEpsilon` silently would break existing callers relying on the asymmetric relative-error formula.

### ⛔ Not applicable (mock-related)

- PR #1870 — drop objx dependency in mock
- PR #1866 — mock data race fix
- Issue #1852 — mock: replace objx.Map (parent of #1870)
- PR #1853 — mock: `AnythingImplementing`

Mock is excluded from the fork; these do not affect us.

## Fred's involvement snapshot

| Item | Type | Fred's position |
|---|---|---|
| #1862 | Issue | Proposed CollectT-as-interface; led to our design-exploration doc |
| #1861 | PR | Supportive ("lighter syntax") |
| #1839 | Issue | Pushed back on misunderstanding; linked to our InEpsilon spec |
| #1829 | PR (his own) | Still open upstream — we've already merged equivalent into internal spew |

## Proposed follow-ups

1. **Pick up #1848 (Subset `%q` → `%#v`) — asap.** Three call sites in `collection.go` (531, 808, 832) + fixture updates. Fold into v2.5.
2. **Park #1859 (channel assertions)** — wait for upstream design to converge. Revisit in next quarterly review.
3. **Queue #1860/#1861 (ErrorAsType) for later.** Needs go1.26 build tag + codegen provisions for conditional assertion generation. Attach to the go1.26 bump line in the roadmap.
4. **#1862 (CollectT/TB)** — keep in design-exploration state; no commitment.
5. **#1863 (float EqualValues)** — if we ever want a symmetric float comparison, introduce it as a **new** assertion (`NumberClose` or similar), not a change to `InEpsilon`.
6. **Update `TRACKING.md`** to add the four new items above with their verdicts.
