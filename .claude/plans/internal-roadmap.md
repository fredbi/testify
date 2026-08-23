> [!NOTE]
> Last revision: 2026-08-23 — opened just after v2.7.0

# Internal roadmap

## Summary

The library is essentially done. The API is stable since v2.4, the assertion surface covers the ground we set out to
cover (140 assertions across 19 domains), the fork's premises — zero dependencies, generics, codegen, no mock —
all hold. From here the project moves slowly and on purpose.

What remains is maintenance (patch releases, quarterly upstream sweeps, dependency and CI hygiene), one mechanical
release, and exactly **two topics worth a design round**: reviving test suites, and an option channel for the
forward form (of which the configurable diff hunk size would be the first user).

This is the internal companion to `docs/doc-site/project/maintainers/ROADMAP.md`. That one announces; this one records
what we actually think, including the parts we would not put in an announcement.

## Context

v2.7.0 shipped in August 2026 (generic forward methods on go1.27, `ErrorNotContains`, six failure-message fixes).
Publicly announced next steps: **v2.8** end of September, go1.26+ required; **v2.9** by end of 2026, tentative, on
test suites. Nothing else is planned this year beyond patches.

"Mostly complete" is not a reason to stop paying attention. The August upstream sweep found four inherited bugs and
two of our own, all in failure messages nobody had looked at closely. The quarterly sweep earns its keep.

## Trajectory

1. Maintenance
   > The default state. Nothing here needs a decision.

   1. 📝 Patch releases as bugs surface
   2. 📝 Quarterly upstream sweep — next **November 2026**
   3. 📝 Dependency bumps, CI hygiene, doc-site upkeep

2. 📝 v2.8 — go1.26+ floor (September 2026)
   > Mechanical. Drops go1.25, which lets the `//go:build go1.26` guard on `ErrorAsType`/`NotErrorAsType` disappear
   > and their assertions fold back into the default files.

3. 🔍 v2.9 — revive test suites (tentative, December 2026)
   > The one real feature left. Undecided in shape, not in appetite.

4. 🧪 An option channel for the forward form, with diff hunk size as its first user
   > Design settled (options on the `Assertions` receiver, discovered by type assertion). **No release target:**
   > spare-time work, picked up when there is room. Value stays low; the design is the interesting part.

5. ⚠️ Carried-over debt: the codegen toolchain safety rail

## Actions

### 🔍 Reviving test suites (v2.9, tentative)

We dropped `suite` from the fork on purpose. It was plagued with issues upstream and the design did not offer
anything worth salvaging: reflection-based test discovery, a `suite.Suite` embedded struct doing pseudo-inheritance,
and a lifecycle whose failure modes keep generating upstream bug reports — #1781 (skipping in `BeforeTest` still runs
`AfterTest`), #1723 / #1802 / #1877 (panics when a `WithStats` suite skips early), #1503 (subtest panic), #1836
(synctest), #1109 (parallel sub-tests).

Of those seven, three are still open (#1781, #1836, #1109), three were closed without a merge (#1723, #1802, #1503),
and one was actually fixed (#1877, merged May 2026). That ratio is the argument: the design keeps producing reports
faster than they get resolved.

**Upstream will not guide this design.** Not as a starting point, not as a checklist of features to match. The
material we have is better and closer to hand: the go-openapi repos are full of ad-hoc test harnesses, written over
years by people solving real problems in a hurry. That bunch is diverse and largely undocumented, which is exactly
what makes it worth reading.

The study, when we get to it:

1. Survey the harnesses across the go-openapi repos — table drivers, fixture loaders, golden-file comparators,
   per-suite temp dirs and servers, the setup/teardown shapes people reinvented.
2. Sort them: what recurs, what is repo-specific, what is a workaround for something the assertion library should
   have offered.
3. Isolate the patterns worth generalising. The suite API, if there is one, falls out of that — a shape we can show
   three existing harnesses collapsing into, not a shape borrowed from elsewhere.

What we already believe, to be confirmed or dropped by the survey:

- **Valued:** lifecycle hooks (setup/teardown at suite and test level) and per-test isolation.
- **Disliked:** reflection-based discovery — the part that makes `go test -run` behave oddly and defeats IDE
  navigation.
- **Constraint:** generics, zero dependencies, no reflection magic. Explicit registration over discovery.
- **Weak signal so far:** the [GitHub discussion](https://github.com/go-openapi/testify/discussions/75) ranged from
  "yes, I want playwright" (a misunderstanding) to "no, suites are a pain". Our own repos are the better witness.

If the survey turns up no pattern worth generalising, that is a result too, and v2.9 can carry something else.

### 🧪 Configurable diff hunk size (upstream #1878) — unscheduled

Not attached to v2.8 or v2.9. Spare-time work, no deadline; the design below is settled enough to pick up cold.

Upstream proposes a global setting for the context size of `GetUnifiedDiffString`. We hardcode it:
`internal/assertions/diff.go` builds the `difflib.UnifiedDiff` literal with `Context: 1`.

The interest is not the feature, it is where it lands. **An assertion has no option channel, by design.** Every
assertion is `func Name(t T, args…, msgAndArgs ...any) bool`, and codegen derives all 8 variants from exactly that
shape. A global setter (the `colors.Enable` precedent) is documented "not intended for concurrent use" and would race
under `t.Parallel()`; hijacking `msgAndArgs` is already in the public ROADMAP's dropped endeavors; `…WithOptions`
variants would multiply the matrix across 140 assertions.

#### Fred's design: options on the forward form only

Carry the options on the `Assertions` receiver, and let the assertion discover them by type-asserting on `t`:

- `New()` hydrates an options value on the `Assertions` it returns;
- `Assertions` exposes `ApplyOptions(o *Options)`, which copies its settings onto the options the assertion is about
  to use;
- an assertion that has an option to honour does `if applier, ok := t.(interface{ ApplyOptions(*Options) }); ok`.
  A successful assertion means `t` is an `Assertions`, not a bare `*testing.T`, and the settings apply.

This accepts the asymmetry deliberately: `assert.New(t).Equal(…)` can be configured, `assert.Equal(t, …)` keeps the
defaults forever. That is the right trade — the package-level form stays exactly what it is today, and nothing in the
variant matrix grows.

**The funnel is narrow enough to make this cheap.** `diff()` has three call sites, all in `equal.go`: `EqualValues`
(l.160), `EqualExportedValues` (l.250), and `failWithDiff` (l.291), which is itself reached from `Equal` (l.46) and
`EqualT` (l.75). So four assertions in total, and one function to thread the option through.

#### Two things to get right before writing it

1. **Forward methods pass `a.T`, not `a`.** Generated today:

       return assertions.Equal(a.T, expected, actual, msgAndArgs...)

   so the `t` an assertion receives is the inner `*testing.T` and the type assertion can never succeed. The forward
   templates have to pass the receiver itself. `*Assertions` embeds `T`, so it satisfies `T` — the change is one line
   per forward template.

2. **Passing `a` silently breaks the `H` check.** `T` declares only `Errorf`; `Helper()` lives in the separate `H`
   interface. `Assertions` embeds the *interface* `T`, so `*Assertions` has no `Helper()` in its method set and
   `if h, ok := t.(H); ok` stops firing inside every assertion. Nothing fails loudly — the reported caller line just
   moves. `Assertions` needs an explicit `Helper()` forwarding to the inner `T` at the same time.

Also to settle: `assertion_types.gotmpl` re-exports what `internal/assertions` exports, so an exported `Options` type
would surface as `assert.Options` unless the template excludes it. Keeping the type inside `internal/` is what stops
anyone outside the module implementing the interface by accident — a property worth preserving on purpose.

**Value, honestly:** low. One upstream issue, nobody in our ecosystem has asked, and `Context: 1` is a deliberate
terse default. The design is worth writing down because it is the first option channel the library would have, not
because hunk size deserves one. If it gets built, build it for the channel and let hunk size be its first user.

### ⚠️ Codegen toolchain safety rail (carried over from v2.6)

Hard-fail when the codegen toolchain sits below a `//go:build go1.N` guard it would emit. Detection must be textual,
since a too-old toolchain drops the file before the scanner sees it.

Two reasons now, not one:

- codegen must observe guarded *source* files (the original v2.6 reason);
- since v2.7, codegen must also **parse what it emits**. `x/tools/imports` formats every generated file with the
  `go/parser` of the running toolchain, and before go1.27 that parser rejects a generic method. A maintainer on
  go1.26 running `go generate` gets a bare "method must have no type parameters" with nothing pointing at the cause.

Small, self-contained, and it stops the next person losing an afternoon. Good candidate to fold into any release.

### 📝 Standing chores

- **Regenerate the GitHub PAT.** The `ghp_` token in `GH_TOKEN` returns 401; the August sweep ran anonymously at
  60 requests/hour, enough for lists but not for comment threads or review state.
- Open buckets from the August sweep (`.claude/plans/upstream-prs-catalog-2026-08-23.md`): five design candidates
  (#1563 `*AssertionFunc` aliases, #1930 IDE-linkable `file:line`, #1929 `Comparer` interface, #1928 unordered
  nested slices, #1910 OSS-Fuzz), four cheap doc wins (#1903, #1919/#1893, #1832), three open questions
  (#1652/#1654 `Eventually`/`Never` semantics, #1934 YAML numerics, #1892 module retraction).

## Achievements

Recorded per release in the plans they came from; this document tracks what is still ahead.

- ✅ v2.7.0 (August 2026) — `.claude/plans/generic-forward-methods.md`, `.claude/plans/upstream-prs-catalog-2026-08-23.md`

## Appendix: what "mostly complete" rules out

Worth writing down so it is not re-litigated every sweep:

- **`mock` is not coming back.** Stated in the README.
- **No option-taking assertion signatures.** See the hunk-size analysis above.
- **No new external dependency in the root module.** Optional features go through the `enable/` pattern.
- **`NoFieldIsZero`** (#1591/#1601) — prototyped, rejected: the semantics are ambiguous (map keys, `[]byte`, pointer
  targets, unexported fields, cycles, `time.Time`-style smart zeros) and every pitfall fix adds a knob until it is a
  struct validator rather than an assertion.
- **`CollectT` redesign / `testing.TB` interop** (#1862) — studied, parked: `testing.TB.private()` blocks any clean
  proxy, and the workaround for affected users is a 3-line per-helper adapter.
