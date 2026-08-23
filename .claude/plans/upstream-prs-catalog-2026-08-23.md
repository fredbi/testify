---
name: Upstream scan — 2026-08-23
description: Delta since the 2026-04-17 catalog — new issues/PRs on stretchr/testify, verified against our code
type: plan
---

# Upstream PR Catalog — stretchr/testify (delta since 2026-04-17)

**Generated:** 2026-08-23 — buckets 🔴 and 🔎 **shipped in v2.7.0** (PR #159)
**Status:** the six bugs are fixed; the 🎯, 📚 and 🔍 buckets are still open
**Prior snapshot:** `.claude/plans/upstream-prs-catalog-2026-04-17.md`
**Tracking ledger:** `docs/doc-site/usage/TRACKING.md`

> Collected anonymously (60 req/h): the `ghp_` PAT in `GH_TOKEN` returns 401, so it needs regenerating before any
> sweep that wants comment threads or review state.

## TL;DR

**96 items** have been active upstream since the last sweep and are not yet in our ledger. Most are noise: a bulk
closure of ancient PRs on 2026-05-12, dependabot bumps, upstream CI, and mock/suite work that does not apply to us.

The real signal was **4 bugs we share**, **2 bugs of our own found while checking**, and **5 design candidates**.
Every bug claim was reproduced against our code, not read off a title. The six bugs shipped in **v2.7.0**; the
candidates, the documentation wins and the three open questions are still on the table.

| Verdict bucket | Count | Items |
|---|---|---|
| ✅ **Actionable — fixed in v2.7.0** | 4 | #1908, #1874/#1875, #1931/#1899, #1898 |
| ✅ **Our own finds — fixed in v2.7.0** | 2 | `InDeltaSlice` argument swap, `EqualValues` signed/unsigned |
| 🎯 **Candidate — worth a design round** | 5 | #1563, #1930, #1929, #1928, #1910 |
| ✅ **Checked, does not affect us** | 7 | #1925, #1531, #1916/#1939/#1922, #1920, #1918, #1896, #1932 |
| 📚 **Doc-only, cheap to adopt** | 4 | #1903, #1919/#1893, #1832, #1776 |
| 🔍 **Needs investigation** | 3 | #1652/#1654, #1934, #1892 |
| ⛔ **Not applicable** | rest | mock/suite, dependabot, upstream CI, 2026-05-12 cleanup |

## Context

Method: list open PRs (134), closed PRs by `updated` desc (100), open issues; filter to activity since 2026-04-17;
subtract every number already named in `TRACKING.md` or the two prior catalog plans (66 known). Each candidate that
claims a bug was then run against our code in a scratch module rather than judged from its title — which changed the
verdict in both directions, so it is worth keeping that step.

Two notes on the raw data. The 2026-05-12 spike (about 25 items) is a maintainer sweep closing PRs from 2015–2020;
it carries no new information. And `#1942` is the upstream PR behind `#1940`, the `ErrorNotContains` we shipped
today — it belongs in the ledger next to `#1940`.

## Trajectory

1. ✅ Fix the four bugs we share with upstream — v2.7.0
2. ✅ Fix the two we found on our own — v2.7.0
3. 📝 🎯 Decide on the five design candidates
4. 📝 📚 Fold in the cheap documentation wins
5. 🔍 Investigate the three open questions
6. ✅ Update `TRACKING.md` once the verdicts are in — v2.7.0

## Actions

### ✅ Actionable — fixed in v2.7.0

All four landed on `reporting-bugs` (PR #159), one commit each, with the fix verified against a probe before and
after. Two things worth carrying forward:

- The **`EqualValues` hole was wider than this catalog first said**: four shapes passed wrongly, not one —
  `int64/uint64`, `int8/uint64`, `int/uint`, in either argument order. Converting smaller-to-larger cannot help,
  since `int8(-1)` wraps to `uint64(math.MaxUint64)` exactly as `int64(-1)` does.
- On **#1908 we adopted half the report**: the rune-element bug is fixed, the invalid-UTF-8 half is deliberately left
  as byte semantics. Upstream's `[]rune` conversion costs an allocation on every string `Contains` to reject a
  malformed needle. Recorded in `TRACKING.md` so it is easy to revisit.

#### Original findings

1. **#1908 — `Contains` is not rune-safe** (open PR)
   - Upstream's stated rationale is half wrong and half right. A *valid* UTF-8 needle can never match mid-rune, but an
     **invalid** one can: `strings.Contains("é", string([]byte{0xa9}))` is `true`, because `0xa9` is the second byte
     of `é`. That part stands.
   - **Our real bug is different and worse.** `containsElement` in `internal/assertions/collection.go` calls
     `elementValue.String()` on the element. For a rune, `reflect.Value.String()` returns the literal
     `"<int32 Value>"`, so `assert.Contains(t, "héllo", 'é')` **fails** with a message naming a type, not a character.
   - Verified: `Contains("héllo", 'é')` → `false`; `Contains("héllo", "é")` → `true`.
   - Suggested action: handle `reflect.Int32` elements explicitly (and reject non-string, non-rune elements with a
     clear message) rather than importing upstream's `[]rune` sliding window, which is a costlier fix for the
     invalid-UTF-8 half only.

2. **#1874 / #1875 — string values unquoted in `Empty` failures** (both closed upstream, idea still good)
   - `assert.Empty(t, "  ")` reports ``Should be empty, but was   `` — the value is invisible.
   - Suggested action: quote the value when it is a string, in `Empty`/`NotEmpty`. Small and self-contained.

3. **#1931 / #1899 — `InEpsilonSlice` drops `msgAndArgs`** (open / closed)
   - `internal/assertions/number.go`: the per-element call is `InEpsilon(t, …, "at index %d", i)`. The caller's
     `msgAndArgs` is not just unforwarded, it is *replaced* by the index context.
   - Suggested action: keep the index context and append the caller's message, rather than choosing one.

4. **#1898 — `InDeltaMapValues` does not name the key** (open PR)
   - We do report a *missing* key (`missing key %q in actual map`), but when the delta comparison fails for a key
     that exists, the message comes from `InDelta` and never says which key it was.
   - Suggested action: same treatment as #1931 — pass the key as context into the per-key call.

### ✅ Our own finds — fixed in v2.7.0

5. **`InDeltaSlice` passes expected and actual the wrong way round**
   - `internal/assertions/number.go`: `InDelta(t, actualSlice.Index(i)…, expectedSlice.Index(i)…, delta, …)`.
   - The verdict is unaffected (the delta is symmetric) but every failure message names them swapped.
   - Inherited from upstream, which still has it; no upstream issue covers it.

6. **`EqualValues` is a false positive for same-size signed/unsigned**
   - `assert.EqualValues(t, int64(-1), uint64(math.MaxUint64))` returns **true**.
   - We already carry the #1531 fix (`ObjectsAreEqualValues` converts the smaller type to the larger one, so
     `EqualValues(int(270), int8(14))` is correctly `false`). The size rule simply cannot separate `int64` from
     `uint64` — same width, and the conversion wraps. Upstream's fix takes the same approach and shares the hole.
   - Suggested action: when both types are numeric and of equal size but differ in signedness, compare through a
     path that cannot wrap.

### 🎯 Candidates — worth a design round

7. **#1563 — make the `*AssertionFunc` types aliases** (MERGED upstream)
   - Ours are defined types in `internal/assertions/ifaces.go`; upstream switched to aliases so a plain func literal
     can be passed without conversion.
   - Interaction to think through first: `assertion_types.gotmpl` already overrides these for `require`, where the
     assertion returns nothing. An alias in `assert` and an override in `require` may read oddly side by side.

8. **#1930 — put a space before `file:line` so IDEs linkify it** (open PR)
   - We format identically to upstream (`fmt.Sprintf("%s:%d", file, line)`, joined with `\n\t\t\t`). Cosmetic, but
     it is the kind of thing people notice every day.

9. **#1929 — `Comparer` interface for custom `Equal`** (closed upstream)
   - We have no custom-comparer hook at all. Closed upstream does not mean uninteresting: it fits our
     options-and-generics direction better than it fits theirs.

10. **#1928 — `ObjectsMatch` / `JsonContentsMatch` for unordered nested slices** (closed upstream)
    - Overlaps our JSON domain and the `Redactor` pattern from #1840. Worth a look before anyone asks for it.

11. **#1910 / #1909 — OSS-Fuzz integration** (open / closed)
    - We already fuzz spew property-style. Upstream wiring up OSS-Fuzz is a prompt to ask whether we want the same
      continuous coverage, not something to adopt as-is.

### 📚 Doc-only, cheap to adopt

12. **#1903** — clarify that `Zero` is unrelated to `encoding/json` `omitempty`/`omitzero`. Applies to our `Zero` doc.
13. **#1919 / #1893** — clarify `EventuallyWithT` goroutine and `FailNow` behaviour. Applies to `EventuallyWith`.
14. **#1832** — `New(t)` is undocumented upstream. Check ours reads well now that it also carries generic methods.
15. **#1776** — *"require: invalid examples from doc comments as functions don't return bool"*. This is exactly the
    defect fixed in `bb4e0df` two commits ago, in their generated docs rather than ours. Nothing to do; worth
    recording in the ledger as independently found and fixed.

### 🔍 Needs investigation

16. **#1652 / #1654 — `Eventually` never runs the condition; `Never` succeeds if the timeout beats the condition**
    - Our `pollCondition` was rewritten around `context.Context` (#1611), so the shape of the bug is likely gone, but
      neither case was reproduced in this sweep. Closed PR #1901 targets the `Never` half.

17. **#1934 — equivalent YAML numerics** (closed upstream)
    - Ours is inconsistent: `YAMLEq("a: 0x10", "a: 16")` is `true`, but `"a: 1"` vs `"a: 1.0"` and `"a: 1e3"` vs
      `"a: 1000"` are both `false`. Hex normalises, float and exponent do not. Defensible, but decide deliberately.

18. **#1892 — retract `github.com/stretchrcom/testify`** — informational; check we have no such legacy path.

### ⛔ Not applicable

Mock and suite work (we ship neither): #1873, #1876, #1877, #1895, #1904, #1912, #1938, #1814, #1802, #1723, #1503,
#1597, #1695, #1719, #1781, #1836, #1905, #375, #663, #965, #969.
Dependabot and upstream CI: #1906, #1913, #1914, #1917, #1923, #1924, #1926, #1883, #1885, #1889, #1941, #1460, #1858.
Bulk stale-PR closure of 2026-05-12: #134, #620, #622, #627, #629, #661, #674, #732, #753, #1531 (metadata touch).
Already ours or already ledgered: #1888 (= our #1848 fix), #1935/#1879/#1772 (yaml migration), #1872/#1897/#1817/#1793
(Regexp, = #1818), #1937, #1861, #1591, #1891, #1921, #1927, #1936, #1918, #1922.

## Achievements

### Sweep ⭐⭐⭐

- 96 uncatalogued items triaged; 13 claims reproduced or refuted against our own code rather than judged from titles,
  which changed the verdict in both directions. Seven candidates were dismissed on evidence (#1925, #1531,
  #1916/#1939/#1922, #1920, #1918, #1896, #1932) — worth keeping so nobody re-opens them.
- Two defects found that no upstream report covers.

### Fixes shipped in v2.7.0 ⭐⭐

- Six commits, one per issue, plus the ledger update referencing all nine upstream numbers including #1776.
- The `InDeltaSlice` swap needed a test that could actually fail: the delta is symmetric, so only the message betrays
  it. Verified by reverting the fix and watching the test catch `"Max difference between 9 and 2"`.

### Ledger ⭐⭐

- Two pre-existing defects fixed alongside: `[#1937]` was referenced with no link definition and rendered as literal
  text on the doc site, and #1576/#1829/#1859 each had duplicate definitions. Counts recounted per row: 34/6/2/5 = 47.

## Appendix: ledger chores

All three done in `fcd4e8a`:

- ✅ the review-frequency line now reads "last review: August 2026, next review: November 2026";
- ✅ `#1942` joins `#1940` in the implemented table;
- ✅ `#1776` has its informational row, recording that both projects found the defect independently.

**Next sweep: November 2026.** Regenerate the PAT first — the `ghp_` token returned 401 for this one, so it ran
anonymously at 60 requests/hour, which was enough for lists but not for comment threads or review state.
