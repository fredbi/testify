> [!NOTE]
> Last revision: 2026-08-23 — **shipped in v2.7.0**. Closed; nothing outstanding.

# Generic assertions as forward methods (go1.27)

## Summary

go1.27 allows type parameters on methods. That lifts the one restriction that kept the 55 generic assertions out of
the `Assertions` receiver: `a.EqualT(...)`, `a.SliceContainsT(...)`, `a.Eventually(...)` become possible.

Because the repo supports go1.26 for six more months, the new methods must live in their own files guarded by
`//go:build go1.27`, generated as a separate build variant, exactly like the go1.26 `ErrorAsType` files from v2.6.

Shipped in **v2.7.0** (tagged August 2026, ahead of the September slot), as PR #156.

## Context

The v2.6 build-guard work already gives us most of the machinery: `model.Function.GoBuild`, `Functions.BuildVariants`,
`model.GoBuildTag`, the `//go:build` line in `header.gotmpl`, the `_go1NN` filename suffix, the orphan sweep, and the
`goversion` Hugo shortcode. See the `go-version-guarded-assertions` memory.

What that machinery does *not* cover: the go1.27 guard here is not inherited from a source file. Every generic assertion
lives in an unguarded source file (`equal.go`, `collection.go`, ...); it is the *forward* variant alone that needs the
guard. So the forward category needs a partition of its own, orthogonal to the source-guard partition.

Two verified facts that make this work:

1. `//go:build go1.27` raises the language version of that file above the module's `go 1.25.0`. Checked with a scratch
   module: a generic method in a guarded file compiles under `go 1.25.0`.
2. Every shape we need compiles as a method: inference from arguments (`a.SliceContainsT([]string{"a"}, "a")`),
   explicit instantiation (`a.IsOfTypeT[int](x)`), and method values (`f := a.EqualT[int]`).

Two generic assertions — `ErrorAsType`, `NotErrorAsType` — are themselves guarded `go1.26`. Their methods need
`max(go1.26, go1.27)` = `go1.27`, so they land in the same file. The rule is written as a max so a future `go1.28`
generic assertion does not silently end up in a `go1.27` file.

## Trajectory

1. ✅ 🛠️ Codegen: forward methods for generic assertions
   1. ✅ Model: resolve the forward guard per function (`max(fn.GoBuild, go1.27)`), partition on it
   2. ✅ Generator: forward files get their own partition loop, keyed on `ForwardGoBuild`
   3. ✅ Templates: the existing forward templates render both, no new template file
   4. ✅ Fix the `Assertions` godoc note, which said generics are not supported as methods
2. ✅ 🏁 Generated output and its tests
   1. ✅ Regenerate `assert/` and `require/`
   2. ✅ Forward tests for generics in `assert_forward_go127_test.go` / `require_forward_go127_test.go`
   3. ✅ `go test ./...` on go1.27, plus go1.26 and go1.25 runs proving the guarded files are excluded
3. ✅ 📚 Documentation
   1. ✅ Doc generator: method rows for generic assertions, with the go1.27 badge
   2. ✅ Metrics: generic assertions now count 4 package variants, not 2
   3. ✅ Hand-written prose that stated the restriction: `docs/doc-site/_index.md`, `ARCHITECTURE.md` ("the maths")
4. ✅ 🛠️ Workspace: bump the `go.work` toolchain floor to go1.27.0, and rewrite the comment that ties the floor to
   `internal/assertions` guards only

## Actions

### Outstanding

Nothing. Both items raised during the work were closed before the release:

- ✅ **`require` method rows in the API docs claimed a `bool` return** — fixed in `bb4e0df`, which also removed the
  `args ..any` typo and the twelve dead links to helper variants that codegen never generates.
- ✅ **`/usr/bin/gofmt` pointed at the go-1.26 distribution** — Fred installed go1.27 by hand; the distro package was
  not out yet.

### Done for release v2.7

1. ✅ **Forward guard resolution** (Phase 1.1)
   - `model.Function.ForwardGoBuild()` returns `fn.GoBuild` for non-generics, `max(fn.GoBuild, "go1.27")` for generics
   - `Functions.ForwardBuildVariants()` lists the distinct forward guards, default partition first
   - New file `codegen/internal/model/buildtags.go` holds all of it, `GoBuildTag`/`BuildVariants` moved there

2. ✅ **Forward generation pass** (Phase 1.2)
   - A second loop in `Generate`, after the source-guard one, over `ForwardBuildVariants`
   - `selectVariant` takes the partitioning key as a parameter: `SourceGoBuild` or `ForwardGoBuild`
   - Produces `assert_forward.go` (non-generics) and `assert_forward_go127.go` (110 methods)

3. ✅ **Templates** (Phase 1.3)
   - Method name renders as `.GenericName` (`EqualT[V comparable]`), the call as `.GenericCallName` (`EqualT[V]`)
   - Format variants too: `a.EqualTf(...)`
   - `testSetup` gains `fn.GenericSuffix()` on the forward variants, for the two `OfTypeT` assertions that cannot
     infer their type parameter
   - No new template file: for a non-generic function `GenericName` == `Name` and `GenericCallName` == `Name`,
     so one template body renders both partitions

4. ✅ **Doc generator** (Phase 3)
   - Drop `(not .IsGeneric)` from the two method rows in `doc_page.md.gotmpl`, add the go1.27 badge on those rows
   - `genericsVariantsMultiplier` dropped: every assertion now counts 4 variants per package.
     `package_variants` 446 → 556, `total_variants` 892 → 1112, `total_functions` 902 → 1122

## Achievements

### Codegen ⭐⭐

- Forward files partition on `Function.ForwardGoBuild`, a second axis alongside the source guard. `max(source guard,
  go1.27)` keeps `ErrorAsType` and `NotErrorAsType` in the same go1.27 file and would push a future go1.28 generic
  assertion into its own.
- No new template. Because `GenericName`/`GenericCallName` fall back to `Name` for a non-generic function, the two
  forward templates render both partitions from one body; `requirement_forward.gotmpl` also lost its hand-rolled
  filter in favour of `Scope "include-generics"`.
- Tests: `TestForwardGoBuild`, `TestForwardBuildVariants`, `TestSourceGoBuild` in `model`; `TestForwardGenericsAssert`,
  `TestForwardGenericsRequire`, `TestForwardGenericsDisabled` in `generator`.

### Generated output ⭐⭐⭐

- `assert/assert_forward_go127.go` and `require/require_forward_go127.go`: 110 methods each (55 generic assertions ×
  plain + `f`), with the matching `_test.go` files. Compiled and passing first try.
- `Eventually`, `Never`, `Consistently` and `EventuallyWith` gain a method for the first time — they are generic
  without a `T` suffix, so they had none at all before.
- Hand-written `assert/assert_adhoc_generic_methods_test.go` drives the methods against a real `*testing.T`:
  inference, explicit instantiation (`a.IsOfTypeT[int](42)`) and a method value (`a.EqualT[int]`).
- Oldstable guarantee re-proven: under `GOTOOLCHAIN=go1.26.0` and `go1.25.0`, `go list -f '{{ .GoFiles }}' ./assert`
  reports zero `_go12*` files and both packages still build and test.

### Documentation ⭐⭐

- API pages carry the method rows for generic assertions with a `{{% goversion "go1.27" %}}` badge, and the domain
  header explains what the badge means.
- `ARCHITECTURE.md` "the maths" now reads 8 variants for every assertion, with the go1.27 caveat spelled out.
- `docs/doc-site/_index.md` replaces "the `Assertion` type cannot be extended with generic methods" with the
  go1.27 usage.

## Appendix: open points

- CI runs one job with `GOTOOLCHAIN=local` on go1.25 (see f490f00). That job keeps excluding the guarded files, which
  is the oldstable guarantee we want. Re-checked after the `go.work` bump: a `toolchain` directive is a switch *hint*
  that `GOTOOLCHAIN=local` ignores, so the floor neither broke nor protected that job. What protects it is the
  `//go:build go1.27` guard on `forward_generics_test.go` and the probe-skip in `TestExecute` (`7f32ae7`), added after
  the oldstable leg failed on three tests: **codegen cannot run below go1.27 at all**, because `x/tools/imports`
  formats with the `go/parser` of the running toolchain and that parser rejects a generic method.
- **Still open, deferred past v2.7:** the safety rail that hard-fails when the codegen toolchain is below a guard it
  emits. Outstanding since v2.6, and now with a second reason: a maintainer on go1.26 running `go generate` gets the
  raw "method must have no type parameters" with nothing pointing at the cause.
- `Eventually`, `Never`, `Consistently`, `EventuallyWith` are generic without a `T` suffix and therefore have no method
  today at all. They gain one under go1.27. That is a genuine API asymmetry between go1.26 and go1.27 builds, not just
  a type-safe alternative to an existing method.
