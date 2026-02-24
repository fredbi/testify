# testify v2 project plan update (2026-01-30)

> [!NOTE]
>
> CLAUDE seems to like unicode emojis a lot (sign of the times, perhaps).
>
> Let's make it happy and adopt the following conventions.
>
> Current status symbols:
>
> * ✅, [x]: Stand for DONE, COMPLETE or MOSTLY COMPLETE
> * ⏳, []:  Stand for IN PROGRESS
> * ❌ Stands for ISSUE or CONCERN
>
> Categorization symbols:
>
> * 🛠️ Stands for INTERNAL TOOLING
> * 🏁 Stands for TESTING
> * 📚 Stands for DOCUMENTATION
> * 😇 Stands for COMPLIANCE
>
> Prioritization symbols:
>
> * 🔍 Stands for NEED INVESTIGATION (before acting)
> * ⛔ Stands for WONT'DO (usually with the reason why)
> * 🔥 Stands for URGENT (requires immediate action)
> * ⚠️ Stands for NEED ATTENTION (requires swift action)
> * ♥️ Stands for ENHANCEMENT (would love it)
> * 🎨 Stands for COSMETIC, LAYOUT, VISUAL IMPROVEMENT
> * 🧪 Stands for EXPERIMENTAL
> * ⚡ Stands for PERFORMANCE RELATED
> * 📝 Stands for PLANNED
>
> Qualitative assessment symbols and terms:
>
> * ⭐⭐⭐: first-class achievement (outstanding, brilliant, A+)
> * ⭐⭐: great achievement (great, excellent, A)
> * ⭐: subpar achievement (decent, correct, B, B-)
> * 👎: not good (inadequate, wrong, C)
>
> The archive of all past plannings, notes and separate endeavors has been
> moved to `./ramblings/*.md`, although everything is not rambling.
> If need be we put a direct reference in the following.

---

## Context

### Motivations

The decision to fork github.com/stretchr/testify and add yet another library
to the already code-heavy go-openapi repos was not a light one.

1. Primary motivation

It all stemmed from the desire expressed by many users of go-openapi libs
to reduce the sprawl of dependencies, in particular deprecated dependencies.
Most of these originated in the testing dependency to github.com/stretch/testify.

After a brief exchange with the maintainers of testify, it appeared clearly that they were entangled
in (expected) questionings like breaking or not breaking change, v2 or not v2.

I do understand totally their reaction: go-openapi libs have _exactly_ the same kind of problems.
I thought the best way to help myself and to help them was to fork straight away.

Other repos in go-openapi have benefited from the techniques adopted here (internalization & separate modules):
* [x] go-openapi/swag
* [x] go-openapi/runtime
* [] go-openapi/strfmt (in progress)

2. Ancillary motives

* more factorizable test utilities

go-openapi and go-swagger repos use *a lot* of tests (many of our repos get >80% coverage).
Many have internal helpers to run tests. Eventually, we'd like to share some of those
to a common "testify/v2" repos.

Example: `JSONEq` tests that JSON are _semantically_ equivalent but can't verify if the ordering of keys is the same.
(we have a helper for that in swag).

* more readable tests in go-openapi repos

By improving the signature of functions in testify/v2, we help our other repos maintain readable tests.

* the guys maintaining stretchr are nice people with whom we have a lot in common (maintaining decade old libraries).
  I find it cool too lend them a hand by feeding them with novel ideas.
  Some contributors at stretchr are already cherry-picking some of my approaches as PRs to the original repo.

* 🧪 Innovation lab repo to work with Claude Code for advanced usage

* 🧪 Innovation lab repo to work with Hugo and eventually help revamp our existing main doc site https://goswagger.io

* 🧪 Innovation lab repo to work with go ast parsing (reusable for go-swagger revamp of the code scanner).

* Further reuse in all personal repos at github.com/fredbi

* Attract further community interest to the go-openapi project in general, leveraging the common need
  for testing go programs as a mean to catch interest.

## Trajectory

The project is progressing in several major directions.

With the help of Claude, it has been possible to make significant advances in all directions
in a few days.

1. Features & fixes
  > We want this project to be a forward exploration base for testify concepts.

  1. ✅ Adopt radical zero-dependency approach
  2. ✅ Adopt & adapt merge request proposed at github.com/stretchr/testify (v1)
  3. ✅ Type safety & other critical fixes needed
  4. ✅ Generics
  5. Features leveraging the internalized dependencies pattern
    * ✅ Enhanced diff output [🎨]
  6. Features with the controlled dependencies pattern
    * ✅ Colorized output [🎨]
  7. Technical
    * ✅ Performance re-assurance (benchmarks), possibly optimization where needed [⚡]
  8. ✅ New features
  9. ✅ Removed extraneous features
    * ✅ Remove test suites (want to use the more specilized mockery tool instead)
    * ✅ Remove http tooling (not sure yet if it is useful, will perhaps be reintroduced in the future)
    * ✅ Remove deprecated methods and types

2. Maintainability
  > We want to reduce technical debt and easily expand or reduce our API.

  1. ✅ Internalized external dependencies, with modernized & relinted code
  2. ✅ Refactored  & modernized code base in single internal/assertion repo
  3. ✅ 100% generated API with variants (8 variants)
  4. ✅ CI automation [🛠️]
  5. ✅ Standardization: align with the rest of go-openapi libraries

3. Documentation [📚]
  > We want to address the (deeply rooted) problem of the bloated API by
  > providing an organized, well-indexed and searchable documentation that
  > supplements the standard godoc at pkg.go.dev.

  1. ✅ Generated documentation organized in domains
  2. ✅ Static documentation site generated with hugo, with the hugo-relearn theme
  3. ✅ Complete project documentation for contributors and maintainers
  4. ✅ Examples, tutorials to demonstrate testing best practices with go

4. Examplarity [😇]
  > We want this project to stand out as an open source golang library. 
  > The project should shine on many aspects, and particularly on its
  > own main topic, which is testing.

  1. ✅ Code quality [😇]
  2. ✅ Test coverage [🏁😇]
  3. ✅ Documentation [📚]
  4. ✅ Respect of community standards (licensing, security, etc.)
  5. ✅ Showcase best practices
  6. ✅ Innovation (AI-aided development, advanced go patterns)

## Active Todo List

**Last Updated:** 2026-01-30

This section tracks current focus items, prioritized according to the [roadmap](../../docs/doc-site/project/maintainers/ROADMAP.md).

### Recently Completed (v2.2.0 - Released)

**Major Features:**
1. ✅ **Generic assertions implementation** (38 functions across 10 domains)
   - IsOfTypeT/IsNotOfTypeT, SeqContainsT/SeqNotContainsT
   - SortedT/NotSortedT, EqualT/NotEqualT, Contains variants, ordering assertions
   - Breakdown: boolean(2), collection(12), compare(6), equal(4), json(1), number(2), order(6), string(2), type(2), yaml(1)
2. ✅ **Colorized output feature**
   - StringColorizer pattern for zero-allocation string coloring
   - Dark/light themes, CLI flags, env vars
   - Documentation and comprehensive tests
3. ✅ **Enhanced diff output**
   - Fixed time.Time rendering (#1829)
   - Deterministic map ordering (#1822)
   - Benchmarked against 4 libraries (fastest: ~15.8 µs)

**Critical Fixes:**
4. ✅ **Goroutine leak fix** (#1611)
   - Consolidated eventually/never/eventuallyWithT into pollCondition
5. ✅ **Spew panic/hang fixes**
   - Property-based testing with random type generator
   - Fixed circular reference edge cases
6. ✅ **Reflect safety improvements** (#1826)
   - Kind/NotKind assertions added (#1803)

**Code Quality & Testing:**
7. ✅ **pollCondition refactoring** (~190 lines → state machine pattern)
   - Reduced complexity from 40+ to distributed (18/11/8/4/3)
8. ✅ **Test refactoring** (collection_test.go, equal_impl_test.go)
   - Unified test matrix pattern
   - 50 unchecked type assertions fixed
   - Early return pattern applied to 7 helper functions
9. ✅ **Function types API cleanup**
   - PanicTestFunc and related types cleaned up
10. ✅ **Package refactoring**
    - Moved assert/enable to enable/stubs for symmetry
11. ✅ **Documentation improvements**
    - Fixed index page count mismatch
    - Added "When to use generics" guidance section
    - Reported Go stdlib fmt.Printf bug (#77313)

### Recent Additions (2026-01-27)

**Documentation Enhancements:**
1. ✅ **Advanced async testing examples** ⭐⭐
   - Added comprehensive "Asynchronous Testing" section to EXAMPLES.md
   - Eventually: Background processor, cache warming patterns
   - Never: Data corruption checks, rate limiter verification
   - EventuallyWithT: Complex multi-assertion scenarios with CollectT
   - Real-world examples with goroutines, mutexes, atomic operations
   - Best practices guidance for async testing

2. ✅ **YAML customization documentation** ⭐⭐
   - Added "Customization" section to USAGE.md
   - Complete example using goccy/go-yaml as alternative unmarshaler
   - Comparison table of YAML libraries (performance, features)
   - How the registration mechanism works (4-step explanation)
   - Advanced pattern: wrapping unmarshalers for custom behavior
   - Important notes about concurrent safety and signature compatibility

**Architectural Decisions:**
3. ✅ **Extension philosophy documentation** ⭐⭐
   - Created `ramblings/yaml-extension-analysis.md`
     - Analysis: Should we add built-in `enable/goccy-yaml` support?
     - **Decision:** No - conflicts with zero-dependency philosophy
     - Rationale: 10 lines of user code preserves flexibility
     - Better alternatives: Documentation, community integrations page
   - Created `ramblings/json-extension-analysis.md`
     - Analysis: Should we make JSON assertions pluggable like YAML?
     - **Decision:** No - JSON in stdlib (fundamental difference from YAML)
     - Rationale: YAGNI, stdlib-first philosophy, performance negligible in tests
     - Key comparison: YAML (no stdlib) vs JSON (in stdlib)
   - Documents preserve reasoning for future architectural decisions

### Recent Additions (2026-01-30)

**New Features (v2.3):**
1. ✅ **UnmarshalJSONAsT[T]** [🧪] (Phase 2.1)
   - Handle JSON unmarshal + assertions in one step
   - Direct go-openapi/go-swagger need
2. ✅ **UnmarshalYAMLAsT[T]** [🧪] (Phase 2.2)
   - Same pattern as UnmarshalJSONAsT for YAML
3. ✅ **Demonstrate extensibility** [🧪] (Phase 1.2)
   - Showcased user extensibility patterns
4. ✅ **Make Assertions implement TestingT interface** [🧪♥️]
   - Errorf conflicts with a different signature → used `Assertions.T()` workaround
   - Added adhoc examples
28. ✅ **NoGoroutineLeak** [🧪⚠️] (Phase 3.1) — research complete, implementation started

**Maintainability:**
5. ✅ **Rename EventuallyWithT** — conflicts with the convention adopted for generics
6. ✅ **Testing messages maintainability** — internal/assertions: testing messages refactored for easier maintenance
7. ✅ **Codegen noise reduction** — code generator no longer systematically changes files (sha/timestamp noise)
8. ✅ **Domain descriptions in godoc** — comments parser now adds domain descriptions to godoc docstrings
9. ✅ **Improve private comments** in internal/assertions
10. ✅ **Fix remaining linter issues (spew)** [😇]

**Documentation:**
11. ✅ **Fixed repeated sections in doc site** [📚]
12. ✅ **Address duplicate markdown documents** [📚] — now minimal: security & license stuff only
13. ✅ **Run mdsf to format code snippets in docs** [📚]
14. ✅ **Roadmap to announce forthcoming features** [📚]
15. ✅ **Polish existing godoc** [📚]
16. ✅ **Complete doc review by github agent** [📚]
17. ✅ **Render testable examples in generated doc** [📚]

**Doc Generation:**
18. ✅ **Generated doc shows tab for any godoc section** (not just examples and usage)
19. ✅ **Render godoc links to current package** [📚]
20. ✅ **Multiline private notes rendering** — fixed markdown blockquote formatting in templates
21. ✅ **Comment annotation pluralization** — maintainer, note etc. annotations now support pluralization
22. ✅ **Generic function signature** — type parameters now shown in generated doc

**Tests:**
23. ✅ **Coverage from main module** [🏁] — no longer missing codegen, enable/yaml, enable/colors
24. ✅ **Missing coverage concentrated in difflib** — addressed
25. ✅ **Integration tests coverage** — testintegration/spew (unavoidable unreachable code), testintegration/yaml, testintegration/colors

**Rejected:**
26. ⛔ **Implement Result[T] Pattern** [🧪] (Phase 1.2)
27. ⛔ **Update codegen for Result[T] returns** [🛠️]

### High Priority (v2.3 - Quality & Maintainability)

**Code Generator Improvements:**
1. ✅ **Simplify code generator templates** [🛠️]
   - Move template logic to Go code
   - Enrich model with computed properties
   - Add helper methods to reduce `{{ if }}` filter logic
   - Impact: Reduces maintenance burden, improves template readability
   - See section 2.2 "Remaining concerns"

2. 🎨 **Remove unnecessary type arguments in generated code**
   - 36 instances in `assert/assert_assertions.go` flagged by `infertypeargs` analyzer
   - Fix: Update code generator templates
   - Impact: Code style improvement (severity: hint)
   - Examples: `assertions.Any[bool](...)` → `assertions.Any(...)`

**Testing & Linting:**
3. ✅ **Complete golangci-lint compliance for test files** [😇]
   - ✅ themes_test.go completed (2026-01-26)
   - Apply linter-compliant patterns across remaining test files
   - Eliminate code duplication in validation logic
   - See `.claude/skills/learned/golangci-lint-compliant-tests.md`

4. ✅ **Fix remaining linter issues** [😇]
   - ✅ Complex functions in internalized libs (mostly code duplication) (spew)
   - Complex table-driven tests in generator
   - Now mostly about reducing nolint directives (66 total — most justified, a few temporary)
   - Target: Clean CodeFactor.io report

**Foundation for Extensibility (from v3 roadmap - Phase 1.1):**
5. ✅ **Make Assertions implement TestingT interface** [🧪♥️]
   - Enable users to compose custom assertions freely
   - Workaround for generic method limitations
   - Foundation for ecosystem growth
   - Complexity: Low (50-100 LOC)
   - Value: Very High (enables user extensibility)
   - See `.claude/plans/v3-roadmap.md` Phase 1.1
   - **Completed:** Errorf conflicts with a different signature, so direct TestingT implementation is not possible.
     Workaround: added `Assertions.T()` method to expose the underlying `TestingT`.
     Added adhoc examples demonstrating the pattern.

   **What it enables:**
   ```go
   // Custom assertions via composition
      type MyAssertions struct {
          *assert.Assertions
      }

      func (ma *MyAssertions) ValidUser(user *User) bool {
          return ma.NotNil(user) &&
                 ma.NotEmpty(user.Email) &&
                 ma.Greater(user.Age, 0)
      }

      // Works with generic assertions
      assert.EqualT(a.T(), x, y)  // a.T() returns the TestingT
   ```

### Medium Priority (v2.3 - Documentation & Usability)

**Documentation:**
1. ✅ **Address duplicate markdown documents** [📚]
   - Now minimal: security & license stuff only
   - Remaining (acceptable duplication):
     - CONTRIBUTORS.md
     - SECURITY.md
     - DCO, LICENSE, NOTICE

2. ✅ **Document custom YAML serialization injection** [📚] (completed 2026-01-27)
   - Show how to inject alternative YAML library
   - Document the enable pattern for optional dependencies
   - Added comprehensive "Customization" section to USAGE.md
   - See section 1.1 "Zero-dependency approach"

3. ♥️ **Improve navigation for large documentation pages** [📚🎨]
   - Add "Back to top" links for pages >500 lines
   - Add section separators between function groups
   - Consider sticky TOC navigation
   - Affects: Collection (993 lines), Equality (871 lines), Comparison (693 lines)

4. 📝 **Configure doc site versioning** [📚🛠️]
   - Adapt release workflow to support doc versioning
   - Multi-version Hugo configuration
   - See section 3.2 "Static documentation site"

**CI/CD Improvements:**
5. ⏳ **CI improvements** [🛠️]
   - Done:
     - ✅ Allow tests with configurable go build tags (want cgo tests in spew)
     - ✅ Fuzz: upload testdata as artifacts on error
     - ✅ Coverage: -coverpkg list
     - ✅ Fix mono-repo release notes

### Lower Priority (v2.3 - Performance & Polish)

**Performance:**
1. ♥️ **Difflib memory allocation optimization** [⚡]
   - Adopt gotextdiff's allocation pattern (fewer, larger allocations)
   - Potential ~20% memory reduction (20.4 KB → 16.5 KB per diff)
   - Reduce allocations from 183 to ~102
   - Low priority: our difflib is already the fastest, marginal benefit

**Code Polish:**
2. ♥️ **Remove extraneous helper functions**
   - Helper methods are mostly unhelpful (not assertions)
   - Identify candidates for removal
   - Examples: InDeltaMap, InDeltaSlice (consider deprecation)
   - See section 1.9 "Removed extraneous features"

3. ✅ **Improve private comments in internal/assertions**
   - Private comments within functions improved
   - Better documentation of complex logic

**Testing:**
4. ✅ **Improve test coverage of generators** [🏁]
   - Integration tests for code generation
   - Golden file testing for template output
   - Currently <0.01% gap overall
   - See section 4.2 "Test coverage"

5. ✅ **Full coverage integration with codecov**
   - ✅ Some tested areas remain unreported
   - ✅ Need to add -coverpkg option
   - See section 4.2 "Test coverage"

### High Priority (v2.4 - Advanced Features & Safety)

**Note:** These items may start in v2.3 if time permits, otherwise they are the primary focus for v2.4.

**Foundation & Error-Aware Assertions (from v3 roadmap):**
1. ⛔ **Implement Result[T] Pattern** [🧪] (Phase 1.2)
   - Foundation for error-aware assertions handling `(T, error)` pattern
   - Chainable assertions based on type constraints
   - Export for user extensibility
   - See `.claude/plans/v3-roadmap.md` Phase 1.2
   - See `.claude/plans/error-aware-assertions.md` for detailed design

2. ✅ **UnmarshalJSONAsT[T]** [🧪] (Phase 2.1)
   - Handle JSON unmarshal + assertions in one step
   - Direct go-openapi/go-swagger need
   - Value: High (frequent pattern in API testing)
   - See `.claude/plans/v3-roadmap.md` Phase 2.1

   **Usage example:**
   ```go
   assert.UnmarshalJSONAsT[User](t, jsonBytes).Equal(expectedUser)
      assert.UnmarshalJSONAsT[Config](t, configJSON).NotNil()
   ```

**Safety Assertions (from v3 roadmap - proven demand):**
3. ✅ **NoGoroutineLeak** [🧪⚠️] (Phase 3.1)
   - Detect goroutine leaks in tests
   - High value - catches real bugs
   - Proven demand (uber-go/goleak widely used)
   - Implement ourselves (maintain zero dependencies)
   - Complexity: Medium (stack parsing, filtering)
   - Value: Very High
   - Priority: P1 (Fred's preference - catches real bugs)
   - See `.claude/plans/v3-roadmap.md` Phase 3.1
   - See `.claude/plans/new-safety-features.md` for discussion

   **Usage example:**
   ```go
   assert.NoGoroutineLeak(t, func() {
          startServer()
          makeRequest()
          stopServer()
      }, IgnoreGoroutine("database/sql.(*DB).connectionOpener"))
   ```

4. 📝 **EventuallyT[T]** [🧪] (Phase 2.3)
   - Handle `(T, error)` pattern in async operations
   - Returns Result[T] for chaining
   - Natural Go idioms for async testing
   - Complexity: Medium (async + generics)
   - Value: High (improves async testing)
   - Priority: P1 (proven need)
   - See `.claude/plans/v3-roadmap.md` Phase 2.3

   **Usage example:**
   ```go
   assert.EventuallyT(t,
          func() (int, error) { return client.FetchCount() },
          5*time.Second, 100*time.Millisecond,
      ).GreaterOrEqual(17)
   ```

5. 📝 **EventuallyWithContextT[T]** [🧪] (Phase 2.4)
   - Context-aware version of EventuallyT
   - Respects context cancellation
   - Idiomatic Go for modern async code
   - Complexity: Medium (context handling)
   - Value: High
   - Priority: P1 (pairs with EventuallyT)
   - See `.claude/plans/v3-roadmap.md` Phase 2.4

6. 📝 **NoFileDescriptorLeak (Unix)** [🧪⚠️] (Phase 3.2)
   - Detect file descriptor leaks on Linux/macOS/BSD
   - Windows support deferred (too complex for v2.4)
   - Complexity: Medium (Unix variants, filtering)
   - Value: High (catches real bugs)
   - Priority: P2 (after goroutine leak detection)
   - See `.claude/plans/v3-roadmap.md` Phase 3.2
   - See `.claude/plans/new-safety-features.md` for discussion

   **Usage example:**
   ```go
   assert.NoFileDescriptorLeak(t, func() {
          f, _ := os.Open("test.txt")
          // If f not closed, test fails
      }, IgnoreNetworkFDs())
   ```

**Code Generator Updates:**
7. ⛔ **Update codegen for Result[T] returns** [🛠️]
   - Rejected: Result[T] pattern not adopted (see item 1 above)

### Features for Future Consideration (v2.5+ / v3.0)

These items are not scheduled for v2.3 or v2.4 but may be considered for future releases.

**From v3 Roadmap (deferred):**
1. ✅ **UnmarshalYAMLAsT[T]** [🧪] (Phase 2.2)
   - Same pattern as UnmarshalJSONAsT for YAML
   - Lives in `enable/yaml` module (optional dependency)
   - Only available when YAML is enabled

2. 📝 **JSONPointerT[T]** [🧪] (Phase 4.1)
   - Type-safe deep JSON assertions via RFC 6901 JSON Pointer
   - Clean syntax for nested structures
   - Zero dependencies (just string parsing)
   - Complexity: Low-Medium
   - Value: Medium (niche but useful for API testing)
   - See `.claude/plans/v3-roadmap.md` Phase 4.1

   **Usage example:**
   ```go
   assert.JSONPointerT[string](t, response, "/user/profile/name").Equal("Alice")
      assert.JSONPointerT[int](t, response, "/user/profile/age").GreaterThan(18)
   ```

3. 📝 **NoFileDescriptorLeak (Windows)** [🧪⚠️]
   - Windows handle leak detection
   - Deferred due to complexity
   - Phase 1: Coarse-grained handle count
   - Phase 2: Full handle enumeration via Native API
   - See `.claude/plans/v3-roadmap.md` Phase 3.2
   - See `.claude/plans/new-safety-features.md` for discussion

**Other New Assertions:**
4. ♥️ **JSON/YAML equivalence vs equality**
   - Distinguish semantic equivalence (current) from key ordering equality
   - Important for testing "verbatim JSON/YAML" features
   - Use cases: go-openapi/swag, fredbi/core/json
   - See section 1.8 "New features"

5. ♥️ **Better time comparison assertions**
   - Enhanced time assertion beyond WithinDuration
   - Time zone handling, DST awareness
   - See section 1.8 "New features"

6. 📝 **Upstream PR candidates** (monitoring)
   - #1087 - Consistently assertion
   - #1601 - NoFieldIsZero
   - Quarterly review cycle continues

**Documentation & Site:**
7. 🧪 **Custom card shortcode for doc site** [📚]
   - Checkbox to display all variants vs only main variant
   - Improves documentation usability for large API
   - See section 3.2 "Static documentation site"

8. 🧪 **News section for doc site** [📚]
   - Hugo blog for release announcements
   - Feature highlights, migration guides
   - See section 3.2 "Static documentation site"

9. ♥️ **Educational content expansion** [📚]
   - Go testing best practices paper
   - Educational code examples at root package level
   - Enrich tutorial with more insights
   - See section 3.4 "Examples, tutorials"

**Explicitly Rejected (from v3 roadmap):**
- ⛔ **General memory leak detection** - Too many false positives, better tools exist (pprof)
- ⛔ **Context caching** - Complexity doesn't justify gains, race condition risks
- ⛔ **BDD framework features** - Keep testify focused on assertions
- ⛔ **Matcher DSL** - Maintain function-based assertion style

**See Also:**
- [v3 Roadmap](./.claude/plans/v3-roadmap.md) - Full v3 vision and design
- [Error-Aware Assertions](./.claude/plans/error-aware-assertions.md) - Result[T] pattern detailed design
- [New Safety Features](./.claude/plans/new-safety-features.md) - Goroutine/FD leak detection discussion
- [Competitive Analysis](./.claude/plans/COMPETITIVE_ANALYSIS.md) - Testify vs Ginkgo/Gomega
- [YAML Extension Analysis](./ramblings/yaml-extension-analysis.md) - Why not add built-in goccy/yaml support
- [JSON Extension Analysis](./ramblings/json-extension-analysis.md) - Why JSON assertions are not pluggable

### Ongoing Maintenance

1. 📝 **Periodic review of upstream PRs** (quarterly)
    - Latest review: 2026-01-19
    - Next review: April 2026
    - See [upstream-prs-catalog-2026-01-19.md](./ramblings/upstream-prs-catalog-2026-01-19.md)
    - **Implemented from upstream**: #1805 (IsOfTypeT/IsNotOfTypeT), #1685 partial (SeqContainsT/SeqNotContainsT)
    - Monitoring: #1087 (Consistently assertion), #1601 (NoFieldIsZero)
    - Superseded by our context-based pollCondition: #1830, #1819

2. 📝 **Standard documentation updates**
   - README, SECURITY, LICENSE, CONTRIBUTING alignment
   - Update in light of doc site review/rewrite iterations
   - See section 4.4 "Respect of community standards"

3. 📝 **Dependency updates**
   - Dependabot automatic updates with auto-merge
   - Quarterly security scan reviews
   - See section 4.4 "Respect of community standards"

## Project Metrics Summary

**Last Updated:** 2026-01-30
**Latest Release:** v2.2.0

| Metric | Value | Notes |
|--------|-------|-------|
| **Test Coverage** | 96% | 100% on all public APIs |
| **Generic Functions** | 38 | Across 10 domains with full type safety |
| **Total Generated Functions** | 608+ | 76 assertions × 8 variants |
| **Performance Improvement** | 1.2-81x | Generic vs reflection (benchmarked) |
| **Code Quality** | golangci-lint clean | Zero linting issues in main packages |
| **Benchmark Refactoring** | 50% reduction | 902 → 450 lines |
| **Equality Test Organization** | 4 files, 1,681 lines | Zero duplication, 96% coverage |
| **Bugs Found via Refactoring** | 4+ | Behavior inconsistencies caught |
| **Documentation Domains** | 19 | Organized by concern |
| **Development Velocity Multiplier** | 3-5x | vs manual approach |
| **Documentation Site** | Production-ready | <https://go-openapi.github.io/testify/> |

**v2.2.0 Highlights:**
- 38 generic assertions with full type safety
- Colorized output with dark/light themes
- Enhanced diff rendering (fastest implementation benchmarked)
- Goroutine leak fixes and pollCondition state machine refactoring
- Comprehensive test quality improvements

**Architecture Summary:**
- Single source of truth: `internal/assertions/` (~5,000 LOC)
- Zero external dependencies (optional YAML via enable pattern)
- Code generation: 100% of assert/require packages
- Test strategy: Layered (exhaustive internal tests + generated smoke tests)
- Pattern library: Unified test matrix, iterator-based testing, type-safe dispatch
- Skill library: 3 Claude Code skills for Go testing best practices

## Detailed plan

### Features & fixes

  1. ✅ Adopt radical zero-dependency approach

  > Achievements
  >
  > * Zero external dependencies ⭐⭐⭐
  > * Optional YAML serialization dependency is optional ⭐⭐⭐
  > * YAML serialization library is injectable ⭐⭐⭐
  > * Pattern in place for further optional dependencies ⭐⭐⭐
  > * ✅ Comprehensive customization documentation (2026-01-27) ⭐⭐
  >   - Added "Customization" section to USAGE.md
  >   - Example using goccy/go-yaml as alternative unmarshaler
  >   - Comparison table of YAML libraries
  >   - Advanced wrapping patterns documented
  >
  > Outstanding todo items: None

  2. ✅ Adopt & adapt merge request proposed at github.com/stretchr/testify (v1)

  > Achievements
  >
  > Adapted & merged upstream fixes in PRs
  >
  > * ✅ **#1513** - Added JSONEqByte
  > * ✅ **#1828** - Fixed panic in spew with unexported fields (critical - we have internalized spew)
  > * ✅ **#1803** - Add Kind/NotKind assertions (candidate for future adoption - aligns with type safety goals)
  > * ✅ Strategy to test more comprehensively type safety issues (e.g. with spew, difflib)
  >   - ✅ Fuzz testing of the spew package, using a property-based values generator
  >   - ✅ Fixed spew hang on edge case (circular pointer wrapped in interface)
  >   - ✅ Fixed spew hang on edge case (circular map reference)
  >
  > Upstream PR catalog established (2026-01-02, updated 2026-01-12)
  >
  > * ✅ Rescanned 129 open upstream PRs and cataloged 12 relevant candidates
  > * See: [upstream-prs-catalog-2026-01-12.md](./ramblings/upstream-prs-catalog-2026-01-12.md)
  > * ⛔ **#1780** - Invalid require examples (rejected as irrelevant - our codegen is completely rewritten)
  > * ⛔ **#1830** - CollectT.Halt() (superseded by our context-based pollCondition)
  > * ⛔ **#1819** - Handle unexpected exits (superseded by our context-based pollCondition)
  > * ✅ **#1817** - Clarify Regexp/NotRegexp documentation (adapted)
  > * ✅ **#1821** - Fix CollectT documentation example (reviewed, our docs correct)

  > Outstanding todo items:
  >
  > * 📝 Periodic review of upstream PRs (quarterly recommended) - **Latest: 2026-01-19**
  > * ✅ Analyzed the following proposals from upstream:
  >   - ✅ [x] **github.com/stretchr/testify#1685** - Iterator support (`iter.Seq`) for Contains/ElementsMatch assertions (Go 1.23+)
  >     - **Status**: ✅ Implemented (partial - Contains only, 2026-01-19)
  >     - **Implementation**: SeqContainsT[E] and SeqNotContainsT[E] using explicit generics (not reflection)
  >     - **Decision**: Skipped SeqElementsMatch - complex edge cases, limited use case
  >     - **Location**: `internal/assertions/collection.go`
  >   - ✅ [x] **github.com/stretchr/testify#1805** - Proposal for generic `IsOfType[T]()` to avoid dummy value instantiation in type checks
  >     - **Status**: ✅ Implemented (2026-01-19)
  >     - **Implementation**: IsOfTypeT[EType] and IsNotOfTypeT[EType] in type.go domain
  >     - **Benefits**: Eliminates dummy instances, solves linter conflicts, cleaner API
  >     - **Location**: `internal/assertions/type.go`

  3. ✅ Type safety & other critical fixes needed [⚠️]

  > Achievements
  >
  > Adapted & merged upstream fixes in PRs
  >
  > * ✅ **#1825** - Fix panic when using EqualValues with uncomparable types
  > * ✅ **#1818** - Fix panic on invalid regex in Regexp/NotRegexp assertions
  > * ✅ **#1223** - Display uint values in decimal instead of hex in diffs
  > * ✅ **#1813** - Panic with unexported fields (fixed via PR #1828)
  > * ✅ **Issue #1611** (go routine leak) - Fixed by consolidating eventually/never/eventuallyWithT into pollCondition
  > * ✅ Other safety fixes (reference to pending issues in the original repo):
  >   Most of these issues stem from an unwary use of the `reflect` package.
  >   - ✅ **#1826** - Reported issue (investigate)
  > * ✅ Fixed edge case bug in Subset/NotSubset (nil/nil interface list input)
  > * ✅ Function types
  >   Problem statement: `PanicTestFunc` and similar types reference internal package types
  >   currently work via re-export aliases but this is untidy.
  >   Exacerbated by move to internal/assertions (was already poor API design in original).
  >   Need to (slightly) rework type mapping to generate clean signatures.
  >
  > Outstanding todo items:
  >
  > * ⛔ **#1824** - Follow/adapt (investigate) (irrelevant: won't do)
  > * ⛔ fuzz test in CI reported an error: fix it (issue: CI doesn't upload the failing case, only the corpus. Need to fix CI first)
  >      (can't reproduce, seems to have been a CI-related issue rather than an issue with our code)

  4. ✅ Generics

  > Achievements
  >
  > * ✅ Establish prioritized list of candidates for adopting generics ⭐⭐⭐
  > * ✅ Target API design: extra method with "T" suffix (e.g., GreaterT[T cmp.Ordered]) ⭐⭐
  >   - Package-level produces extra confusion and now we're well equipped to handle a large API
  >   - Consistent naming convention across all generic variants
  >
  > * ✅ **38 generic assertions fully implemented and tested** (completed 2026-01-19) ⭐⭐⭐
  >
  >   Complete implementation by domain:
  >   - ✅ **Boolean (2)**: TrueT, FalseT
  >   - ✅ **Collection (12)**: StringContainsT, SliceContainsT, MapContainsT, StringNotContainsT,
  >     SliceNotContainsT, MapNotContainsT, SliceSubsetT, SliceNotSubsetT, ElementsMatchT, NotElementsMatchT,
  >     SeqContainsT, SeqNotContainsT
  >   - ✅ **Comparison (6)**: GreaterT, GreaterOrEqualT, LessT, LessOrEqualT, PositiveT, NegativeT
  >   - ✅ **Equality (4)**: EqualT, NotEqualT, SameT, NotSameT
  >   - ✅ **JSON (1)**: JSONEqT
  >   - ✅ **Number (2)**: InDeltaT, InEpsilonT
  >   - ✅ **Ordering (6)**: IsIncreasingT, SortedT, NotSortedT, IsNonIncreasingT, IsDecreasingT, IsNonDecreasingT
  >   - ✅ **String (2)**: RegexpT, NotRegexpT
  >   - ✅ **Type (2)**: IsOfTypeT, IsNotOfTypeT
  >   - ✅ **YAML (1)**: YAMLEqT
  >
  >   Type constraints used:
  >   - `Text` interface: string | []byte (⛔ won't do []rune - can't convert easily)
  >   - `Ordered` constraint: all numeric types and strings
  >   - `comparable` constraint: types that support == and !=
  >   - ⛔ `Len[L Countable]` - won't do: complex type constraint not supported
  >   Won't do as generics:
  >     - ⛔ EqualValues, EqualExportedValues, Exactly, Empty, Nil: better as reflection-based
  >     - ⛔ Condition, Eventually, Never, EventuallyWithT: no point with going generics
  >     - ⛔ Error[E error], ErrorIs, ErrorContains: no need - already use the error interface
  >     - ⛔ File, HTTP, Panic: no opportunity to go generics
  >     - ⛔ InDeltaSlice, InEpsilonSlice, InDeltaMapValues: already much redundant, no need to extend the bloat here
  >     - ⛔ Type (Implements, IsType, Kind, Zero): better as reflection-based
  >
  > * ✅ Comprehensive test suite with iterator pattern ⭐⭐⭐
  >   - Extensive test coverage across all 38 generic functions
  >   - Type dispatch for all constraint types (Ordered, comparable, Text, iter.Seq[E])
  >   - Edge case coverage (equal values, edge boundaries, type safety)
  >   - All Ordered types tested (13 types: int, int8-64, uint, uint8-64, float32-64, string)
  >
  > * ✅ Code generation support ⭐⭐⭐
  >   - Scanner handles generic type parameters
  >   - Generator creates all 8 variants for generic functions
  >   - Documentation generation includes generic type signatures
  >
  > * ✅ Analyze and compare with `github/alecthomas/assert` (see [ramblings](./ramblings/alecthomas-assert.md)) ⭐
  > * ✅ Adapt code generation to support generics ⭐⭐⭐
  > * ✅ Refactor number assertions to level the behavior on edge cases ⭐⭐
  > * ✅ Adapt doc generation to support generics ⭐⭐
  > * ✅ Assertions: refactor tests (factorize generic vs reflection-based) ⭐⭐
  >      See [the detailed note](ramblings/project-status-2026-01-20.md)
  >
  > Other fixes & additions:
  >
  > * ✅ Added YAMLEqBytes to be consistent with JSON
  > * ✅ Added IEEE 754 edge cases handling to InDelta and InEpsilon
  > * ✅ Added support for 0 expected in InEpsilon (falls back to absolute error)
  > * ✅ Fixed invalid type conversion for uintptr (reflect-based compare)
  > * ✅ Fixed quirks with unwary usage of Regexp (unexpected behavior on some input types)
  > * ✅ Refactored Regexp
  > * ✅ Refactored tests for all assertions with new generics
  > * ✅ Added benchmarks to showcase perf improvement with generics
  > - ✅ Further test refactoring opportunities (7700 LOC to test 5000 LOC)
  >
  > Outstanding todo items: N/A

  5. Features leveraging the internalized dependencies pattern

  > Leveraging internalized dependencies (go-spew, difflib)

    * ✅ Enhanced diff output [🎨]

  > Achievements
  >
  > * ✅ **#1828** - Fixed panic in spew with unexported fields (adapted to `internal/spew/`)
  > * ✅ **#1816** - Fix panic on unexported struct key in map (internalized go-spew - may need deeper fix)
  > * ✅ verify that error reporting (with source and line number of the failing test)
  >      is not adversely affected by our refactoring (helpers calling helpers calling helpers).
  > * ✅ Leverage internalized libs to address the following (original) pending issues:
  >   - **#1829** - Fix time.Time rendering in diffs (internalized go-spew)
  >   - **#1822** - Deterministic map ordering in diffs (internalized go-spew)
  > * ✅ Comprehensive difflib comparison (2026-01-12) ⭐⭐⭐
  >   - Compared 4 libraries: our difflib, gotextdiff, go-udiff, sergi/go-diff
  >   - **Our difflib is the fastest** (~15.8 µs vs gotextdiff ~18.1 µs)
  >   - Identical output quality to gotextdiff (industry standard)
  >   - See [internal/testintegration/difflib/README.md](../../internal/testintegration/difflib/README.md)

  > Outstanding todo items:
  >
  > * ♥️ Memory allocation optimization: adopt gotextdiff's allocation pattern [⚡]
  >   - Potential ~20% memory reduction (183 → 102 allocations)
  >   - Low priority: already fastest, marginal benefit

  6. Features with the controlled dependencies pattern
    * ✅  Colorized output

  > Achievements
  >
  > * ✅ Determined which colorization lib we want to enable ⭐⭐
  >    -> none, just ANSI colors. We indulge with a limited dependency to golang.org/x/term.
  > * ✅ Implemented opt-in colorization feature to render diff and unequal values ⭐⭐⭐
  > * ✅ StringColorizer pattern: zero-allocation string coloring (cleaner than bufio.Writer approach) ⭐⭐
  > * ✅ Theme support: dark (bright colors) and light (normal colors) themes ⭐⭐
  > * ✅ Configuration: CLI flag `-testify.colorized`, env vars TESTIFY_COLORIZED, TESTIFY_THEME ⭐⭐
  > * ✅ CI support: TESTIFY_COLORIZED_NOTTY for forcing colors in non-TTY environments ⭐
  > * ✅ Documentation added to EXAMPLES.md "Colorized Output (Optional)" section ⭐
  > * ✅ Unit tests for StringColorizer using iter.Seq pattern ⭐
  > * ✅ Performance re-assurance (benchmarks), possibly optimization where needed. [⚡]
  > * ✅ **Behavior re-alignment** (discoveries through unified testing): ⭐⭐⭐
  >      - **IsNonDecreasing/IsNonIncreasing**: Logic was inverted or unclear - fixed to match documentation
  >      - **Equal vs EqualValues**: Aligned edge case behavior (both should fail with functions)
  >      - **Pointer identity**: Two nil pointers of same type now correctly considered "same"
  >      - **Empty assertions**: Comprehensive testing of nil, zero values, empty collections
  >      - **Impact**: Unified test approach acts as semantic validator, catching logical inconsistencies missed by traditional tests
  > * ✅ More defensive guards re-reflect panic risk: EqualExportedValues
  > * ✅ Verified that overall performance
  >      is not adversely affected by our refactoring (helpers calling helpers calling helpers).
  > * ✅ Add benchmark suite for hot paths (Equal, Contains, Empty).[⚡]

  > Original PRs that have been considered for our implementation (optional in `enable/color` module):
  >
  > * **#1467** - Colorized output with terminal detection (most mature)
  > * **#1480** - Colorized diffs via TESTIFY_COLORED_DIFF env var
  > * **#1232** - Colorized output for expected/actual/errors
  > * **#994**  - Colorize expected vs actual values
  >
  > * ⛔ also consider modern alternatives such as charmbracelet (won't do: that's an entire ecosystem to load).

  > Outstanding todo items: N/A (feature complete)

  7. Technical
  > Outstanding todo items: N/A (feature complete)

  8. ⏳ New features

  > Achievements (2026-01-30):
  >
  > * ✅ **UnmarshalJSONAsT[T]** [🧪] (Phase 2.1) — JSON unmarshal + assertions in one step
  > * ✅ **UnmarshalYAMLAsT[T]** [🧪] (Phase 2.2) — YAML unmarshal + assertions
  > * ✅ **Demonstrate extensibility** [🧪] (Phase 1.2)
  > * ✅ **Make Assertions implement TestingT interface** [🧪♥️]
  >   - Errorf conflicts with a different signature → used Assertions.T() workaround
  >   - Added adhoc examples
  > * ✅ **NoGoroutineLeak** [🧪⚠️] (Phase 3.1) — done
  > * ⛔ **Implement Result[T] Pattern** [🧪] (Phase 1.2) — rejected
  > * ⛔ **Update codegen for Result[T] returns** [🛠️] — rejected (depends on Result[T])
  >
  > Outstanding todo items (minor):
  >
  > The time is ripe to leverage our forking approach to improve tests in the go-openapi repos.
  >
  > Desirable features:
  >
  > * Distinguish json / yaml equivalence (semantical) and equality (with ordered keys).
  >   It matters to test some "verbatim JSON/YAML" features (e.g. go-openapi/swag, fredbi/core/json).
  > * Better time comparison assertions
  >
  > Explicitly rejected features (with rationale documented):
  >
  > * ⛔ **Custom JSON serializer injection** (2026-01-27)
  >   - Considered making JSON pluggable like YAML
  >   - **Rejected:** JSON in stdlib (fundamental difference from YAML situation)
  >   - Rationale: YAGNI, stdlib-first philosophy, performance negligible in tests, compatibility risks
  >   - See [ramblings/json-extension-analysis.md](./ramblings/json-extension-analysis.md)
  > * ⛔ **Built-in goccy/yaml support** (2026-01-27)
  >   - Considered adding `enable/goccy-yaml` as built-in option
  >   - **Rejected:** Conflicts with zero-dependency philosophy, slippery slope
  >   - Rationale: Problem already solved with public API (10 lines of user code)
  >   - See [ramblings/yaml-extension-analysis.md](./ramblings/yaml-extension-analysis.md)

  9. ♥️ Removed extraneous features

    * ✅ Removed test suites and mocks (want to use the more specilized mockery tool instead)
    * ✅ Removed http tooling (not sure yet if it is useful, will perhaps be reintroduced in the future)
    * ✅ Removed deprecated methods and types
    * ✅ Removed the few deprecated functions remaining in `internal/assertions`
  > * ✅ remove useles `New`  from internal/assertions

  > Outstanding todo items:
  >
  > * ♥️ helpers methods are mostly unhelpful (not assertions). See if there are useful candidates to removal.
  > * ♥️ remove extraneous "helper" type definitions actually unhelpful
  > * ♥️ the time is ripe for us to start reflecting on removing extraneous stuff like `InDeltaMap` or `InDeltaSlice`

### Maintainability

  1. ✅ Internalized external dependencies, with modernized & relinted code

  > Completion notes
  >
  > * ✅ Made sure code attribution and licensing is respected
  > * ✅ Modernized code for internalized dependencies ⭐⭐⭐

  2. ✅ Refactored  & modernized code base in single internal/assertion repo
  3. ✅ Refactored tests with unified test cases (single matrix of expectations for reflect-based/generic variants)

  > * ✅ Test refactoring now complete with:
  >      - ✅ collection_test.go
  >      - ✅ equal_impl_test.go (redundant)
  > * ✅ Move `assert/enable` to `enable/stubs` (not a separate module) to reexport internals.
  > * ✅ Simplify `pollCondition` function in `internal/assertions/condition.go`. [🛠️]
  >      The consolidated poller (issue #1611 fix) merges 3 functions into one but is complex (~190 lines).
  >      Consider splitting into smaller helpers or using state machine pattern for clarity.
  > * ✅ Rename EventuallyWithT (conflicts with the convention adopted for generics) (2026-01-30)
  > * ✅ Testing messages maintainability: refactored error message checks across test files (2026-01-30)
  > * ✅ Eliminated assertion self-use in test code: replaced ~273 instances across 15 test files (2026-01-30)

  > Achievements
  >
  > * Centralized maintenance with reduced code base in `internal/assertions` ⭐⭐⭐
  > * Reorganized assertions code into manageable domain-organized files (e.g. `boolean.go`) ⭐⭐⭐
  > * Overhauled the existing code generation with a more capable code parser and code and doc generator ⭐⭐
  > * Iterator-based test pattern (iter.Seq) applied project-wide ⭐⭐⭐
  >   - Clean separation between test data and test logic
  >   - Type-safe test case definitions
  >   - Parallel execution support
  >   - Reusable test case iterators
  > * Early return pattern for test helpers (2026-01-19) ⭐⭐
  >   - Applied to 7 helper functions
  >   - Reduced nesting, improved readability
  >   - Consistent pattern across test files
  > * Type-safe assertion pattern (2026-01-19) ⭐⭐⭐
  >   - All 50 unchecked type assertions fixed
  >   - Safe pattern: `value, ok := assertion` with descriptive errors
  >   - Zero linting issues (forcetypeassert clean)
  > * **Test refactoring with unified test matrix pattern** (2026-01-20) ⭐⭐⭐
  >   - **Equality domain**: Refactored into 4 focused test files (1,681 lines total, 96% coverage, zero duplication)
  >     - `equal_test.go` (928 lines) - Deep equality tests
  >     - `equal_pointer_test.go` (180 lines) - Pointer identity tests
  >     - `equal_unary_test.go` (281 lines) - Unary assertion tests
  >     - `equal_impl_test.go` (292 lines) - Internal helper tests
  >   - **Benchmarking refactoring**: 37 benchmark functions, reduced duplication 50% (902→450 lines)
  >   - **Documented performance gains**: 1.2x to 81x faster with generics vs reflection
  >   - **Pattern proven**: Unified test matrix successfully applied to equality and ordering domains
  >   - **Bugs discovered**: 4+ behavioral inconsistencies caught and fixed through comprehensive testing
  > * **Development velocity improvement** (2026-01-20) ⭐⭐⭐
  >   - Code generation provides **3-5x faster** development for: adding assertions, fixing bugs, ensuring consistency, maintaining docs
  >   - Single source of truth prevents drift and enables rapid feature additions (hours vs days)
  >   - Mechanical consistency eliminates entire classes of bugs
  >
  > The problem that we solved
  >
  > Adding a single new assertion (FileEmpty/FileNotEmpty) exposed significant maintainability issues:
  > - Touching files containing thousands of lines of code
  > - Manual duplication across assert/require packages
  > - Hand-writing format variants, forward methods, and all their tests
  > - Difficulty maintaining consistency across 76 assertion functions
  > - Fear of being unable to maintain the fork long-term
  >
  > * ✅ Our code generator has rapidly grown into a complex beast.[🛠️]
  >      Need to closely monitor code complexity and readability, to keep our design trade-off viable (generating
  >      is supposed to _reduce_ maintenance, not the other way around).
  > * ✅ One or two test functions are detected as complex by linters. Need to refactor. [🏁]
  > * ✅ Need to simplify templates (e.g. deport some of the template functionality to go or use define blocks in templates).[🛠️]
  >      Enrich model with computed properties to avoid declaring variables in templates.
  >      Add helper methods to reduce `{{ if }}` filter logic.
  > * ✅ Addressed the problem of duplicate markdown documents maintained for the github repo
  >      *AND* the doc site. Now minimal: security & license stuff only.[📚]
  > * ✅ (comments parser) Domain descriptions are only parsed from private comments. We might want to add them
  >      to the godoc docstring.
  > * ✅ (comments parser) Maintainer, note etc private comment annotation don't support pluralization (e.g. "maintainers:" is not
  >      detected.
  > * ✅ (templates) Multiline private notes won't render great in the template. Need some markdown reformating for correct
  >      blockquote. [📚]
  > * ✅ Code generator systematically changes files, if only to modify the sha and timestamp. This produce a lot of noise.
  > * ✅ Running codegen with "go generate" does not produce exactly the same result as when running the binary from ./codegen:
  >      tool and headers are empty.
  >
  > Remaining concerns
  >
  > For detailed strategic reflections, lessons learned, and philosophical notes on code quality,
  > see [project-status-2026-01-20.md](./ramblings/project-status-2026-01-20.md).
  >
  > Potential enhancements & known minor glitches
  >
  > * ♥️ Function types (e.g., `ValueAssertionFunc`, `PanicAssertionFunc`) would need some proper comment for categorization. [📚]
  > * ♥️ Code generator testing using golden file. Currently the generator integration test is merely a smoke test. 
  > * ⛔ Add option to skip deprecated functions in codegen. Wont't do: remove deprecated stuff from internal if we don't want it to be generated.

  3. ✅ 100% generated API with variants (8 variants)

  > Achievements
  >
  > * ✅ Full generated code for `assert` and `require` from a single source of truth ⭐⭐⭐
  >      76 functions × 8 variants = 608 generated functions [TODO: these figures must be revisited]
  > * ✅ Centralized maintainance with reduced code base in `internal/assertions` ⭐⭐⭐
  >
  > Completion notes
  > * ✅ fully relinted code
  > * ✅ refactored tests
  >
  > Outstanding todo items (minor):
  >
  > * ✅ Outstanding "proposal for enhancement" comments: should be addressed with "diff rendering" work item

  4. ✅ CI automation [🛠️]

  > Done:
  > * ✅ Allow to run tests with a configurable go build tag (want to enable cgo tests in spew and achieve 100% coverage there)
  > * ✅ Fuzz: in case of error, upload testdata as artifacts
  > * ✅ Coverage: -coverpkg list
  > * ✅ Fix mono-repo release notes


  > Achievements
  >
  > * ✅ align with go-openapi repos: adopt shared workflows go-openapi/ci-workflows
  > * ✅ adopt shared mono-repo release workflows (release automation for go mono-repo)
  > * ✅ create doc update workflow (possible later reused it into shared ci-workflows)
  >
  > Outstanding todo items:
  > * ♥️ further CI improvements (~better release notes~, documentation quality checks, etc)
  >   are delegated to the ci-workflows repo. We won't follow up these in this project.
  
  5. ✅ Standardization: align with the rest of go-openapi libraries

  > * Standard documentation: README, SECURITY, LICENSE, NOTICE, CONTRIBUTING, STYLE, DCO
  >   Should be more or less aligned but now this repo is having more advanced versions,
  >   to be back-ported to other repos.
  >
  > * Standard config: .golangci.yml, .cliff.toml, dependabot.yaml should be already aligned (need a sanity check).
  > * CI config: part of (4) above.
  > * Standard github settings: to be checked, but should come very close to the others (need a sanity check)

### Documentation

  1. ✅ Generated documentation organized in domains

  > Achievements
  >
  > * Auto-generated documentation with clear and searchable domains to organize the big API ⭐⭐⭐
  >
  > * Documentation improvements:
  >
  > * ✅ Document the concept of domains and how we've established this split [📚]
  > * ✅ More documentation for generics usage patterns [📚]
  > * ✅ Roadmap contains an invalid reference to twitter from the original example [📚]
  > * ✅ Update published roadmap
  > * ✅ Better document breaking changes from v1 (w/ apidiff?) [📚] (go-apidiff not usable on fork)
  > * ✅ Many remaining typos, redundant text [📚]
  >
  > Completion notes
  >
  > * ⛔ Won't generate doc for types, variables and const: doc site for assertions index is already very loaded,
  >      and I don't want too much overlap with pkg.go.dev. Writing a fully featured alternative to pkg.go.dev
  >      is an interesting project, but if we want to do that, we'll do it in a separate project.[📚]
  >
  > Outstanding todo items (minor):
  >
  > * ✅ Polish existing godoc[📚]
  > * ✅ Render testable examples in generated doc[📚] This one is hard: see github.com/golang/pkgsite/godoc/dochtml/dochtml.go
  > * ❌ Render multiple examples, fix stability issue with examples (sometimes from package assert, sometimes from package require)
  > * ⛔ ~Replace or extend "Usage:" sections in godoc with references to generated examples (not sure yet this will actually improve anything)~
  > * ✅ Improve private (non-godoc) comments in internal/assertions
  >        Private comments within functions improved.
  > * ✅ Generated doc shows tab for any godoc section, not just examples and usage (ex: concurrency) [📚]
  > * ✅ Run broken link check [📚]
  >
  > **From documentation assessment (2026-01-20):** [📚🎨]
  >
  > * ✅ **Index page count mismatch**: Says "18 domains" but lists 17 (missing "Common"?)
  >   - Location: docs/doc-site/api/_index.md:28
  > * ♥️ **Large page navigation improvements**: Collection (993 lines), Equality (871 lines), Comparison (693 lines)
  >   - Add "Back to top" links for pages >500 lines
  >   - Add section separators between major function groups
  >   - Consider sticky TOC navigation for large pages
  > * ✅ **Missing guidance section**: Add quick reference explaining:
  >   - When to use generic (*T) vs reflection variants
  >   - Type safety benefits
  >   - Performance implications (reference benchmarking results)
  >   - Migration guide from reflection to generic variants
  > * ✅ **Generic function grouping**: Within domains, consider grouping related functions together
  >   - Example: Equal/EqualT/NotEqual/NotEqualT as a group
  >   - Or create "Generic variants" subsections (done)
  >
  > Overall assessment: **Production-ready** ⭐⭐⭐
  > Domain-based structure successfully handles API expansion (38 generics across 10 domains).
  >
  > For complete documentation review, see [doc-assessement-2026-01-20.md](./ramblings/doc-assessement-2026-01-20.md).

  2. ✅ Static documentation site generated with hugo, with the hugo-relearn theme

  > Achievements
  >
  > * ✅ hugo configuration with modern theme Relearn ⭐⭐⭐
  > * ✅ Production-ready documentation site ⭐⭐⭐
  >
  > Completion notes (2026-01-01)
  >
  > Site is now READY FOR PUBLIC LAUNCH (final grade: A-) <https://go-openapi.github.io/testify/>
  >
  > * ✅ Comprehensive documentation review completed by CLAUDE[📚]
  > * ✅ Hugo versioning configuration fixed/removed[🎨📚]
  > * ✅ All content sections complete (API Reference: A, Examples: A, Project Docs: A-, Maintainers: A)
  > * ✅ 13/13 critical issues resolved
  >
  > Fixed issues (2026-01-01):
  >
  > * ✅ variabilized max card width in css (customized the theme's css using the "legit" documented way)
  > * ✅ wrong rendering of markdown badges (hugo version-dependant, needed config for goldmark renderer)
  > * ✅ godoc reference badge left-aligned
  > * ✅ Untidy capitalization for "OS" (was rendered as Os)
  > * ✅ Hugo versioning configuration fixed/removed (was causing browser console errors)
  > * ✅ Fixed all typos across documentation
  > * ✅ Documentation content polish completed [🎨]
  > * ✅ EXAMPLES.md written (482 lines comprehensive)
  > * ✅ TUTORIAL.md written (640 lines with iterator pattern)
  > * ✅ CODEGEN.md written (286 lines + 3 mermaid diagrams)
  > * ✅ ARCHITECTURE.md enhanced with improved mermaid diagram
  > * ✅ MAINTAINERS.md all broken links fixed
  > * ✅ All project README text issues resolved
  > * ✅ Publish to github pages (next critical step)[📚]
  > * ✅ Automate github pages update in CI [🛠️]
  > * ✅ Refer to the doc site in the project's README
  >
  > See <./ramblings/docs-assessment-2026-01-01.md> for a complete review.

  > Outstanding todo items:
  >
  > * 📝 Configure doc site versioning, adapt release workflow to support doc versioning [📚🛠️]
  > * ♥️ Custom card shortcode to add a checkbox to display all variants vs only the main variant[📚]
  > * ♥️ Add News section to the doc site, with release and feature announcements (hugo blog)

  3. ✅ Complete project documentation for contributors and maintainers

  > Achievements
  >
  > * Project documentation completed ⭐⭐⭐
  >
  > Completion notes
  >
  > * ✅ Documentation content polish COMPLETED (2026-01-01)[📚]
  >   * ✅ Maintainers documentation with ARCHITECTURE.md (enhanced mermaid diagram)
  >   * ✅ CODEGEN.md with comprehensive workflow documentation + 3 mermaid diagrams
  >   * ✅ MAINTAINERS.md with all broken links fixed
  >   * ✅ Roadmap
  >   * ✅ All text issues in README.md resolved
  >   * ✅ Hugo versioning configuration fixed
  >   See docs-assessment-2026-01-01.md (final grade: A-)

  4. ✅ Examples, tutorials to demonstrate testing best practices with go[📚]

  > Achievements
  >
  > * ✅ Getting started documentation completed ⭐⭐⭐
  >
  > Completion notes (2026-01-01)
  >   * ✅ EXAMPLES.md - 482 lines comprehensive (Quick Start, assert vs require, common assertions,
  >         variants, table-driven tests, real-world examples, advanced patterns, YAML support,
  >         best practices, migration guide)
  >   * ✅ Advanced async testing examples added (2026-01-27) ⭐⭐
  >     - Eventually: Background processor, cache warming examples
  >     - Never: Data corruption check, rate limiter examples
  >     - EventuallyWithT: API consistency, distributed cache, complex assertions
  >     - Best practices section for async testing
  >   * ✅ TUTORIAL.md - 640 lines with Go 1.23+ iterator pattern focus (What makes a good test,
  >         patterns, simple test logic, iterator pattern examples, parallel execution, setup/teardown,
  >         edge cases, error testing, complete examples)
  >   * ✅ `examples/_index.md` - Uses Hugo children shortcode for dynamic card layout
  >   * ✅ TL;DR notice added for experienced testify users
  >   See <./ramblings/docs-assessment-2026-01-01.md> (final grade: A)
  >
  > Outstanding todo items:
  >
  > * ♥️ Enrich the tutorial document with more insightful tips about go testing (future enhancement)
  > * ✅ Write educational code examples ~at root package level~ in doc-site
  > * ♥️ Write go testing best practice paper based on techniques developed here

### Examplarity

  1. ✅ Code quality

  > Achievements
  >
  > * ✅ Code fully relinted according to our (documented) linting standard ⭐⭐⭐
  > * ✅ Improved readability of table-driven tests, using our iterator pattern ⭐⭐⭐
  > * ✅ Code quality assessment after merge (e.g. CodeFactor.io, possibly an additional review from CLAUDE) ⭐⭐
  > * ✅ Addressed or annotated all TODOs in code ⭐
  > * ✅ Early return pattern refactoring (2026-01-19) ⭐⭐
  >   - Applied to 7 helper functions across 3 test files
  >   - Reduced nesting levels, improved code readability
  >   - Consistent pattern: return early from success case, eliminate else blocks
  > * ✅ Type assertion safety fixes (2026-01-19) ⭐⭐⭐
  >   - Fixed 50 unchecked type assertions across collection_test.go and equal_test.go
  >   - All type assertions now use safe pattern: `value, ok := assertion` with descriptive errors
  >   - Zero linting issues: `golangci-lint run --enable-only forcetypeassert` returns clean
  >
  > Outstanding todo items (minor):
  >
  > * ✅ Remaining linting issues in internalized libs: spew fully relinted (2026-01-30)
  > * ♥️ Remaining linting issues in generator (complex table-driven tests with inlined assertion functions).
  > * ♥️ above mentioned complex functions are reported by CodeFactor.io - reducing nolint directives (we have 66 - most are justified - a few remain temporary)

  2. ✅ Test coverage

  > Achievements
  >
  > * ✅ Auto-generated tests for generated packages with nearly 100% coverage ⭐⭐⭐
  >   Test generation is driven by example values provided as comments. ⭐⭐⭐
  >   Test values are actually parsed and checked to be legit go expressions. ⭐⭐⭐
  >
  > * ✅ Great test coverage across the board (>90%) ⭐⭐⭐
  > * ✅ Outstanding test coverage for the assertions package (>94%) ⭐⭐⭐
  > * ✅ Outstanding test coverage for unitary packages forming the code generator (>99%) ⭐⭐⭐
  > * ✅ Fixed issue with codecov not reporting coverage for all modules: only the root module is reported
  > * ✅ Generic function tests expanded (2026-01-19) ⭐⭐⭐
  >   - **38 generic functions** now fully tested across 10 domains
  >   - Latest additions: SortedT/NotSortedT, IsOfTypeT/IsNotOfTypeT, SeqContainsT/SeqNotContainsT
  >   - Comprehensive coverage includes:
  >     - All constraint types: Ordered (13 types), comparable, Text (string | []byte), iter.Seq[E]
  >     - Edge cases: equal values, boundaries, type safety, nil handling, sequence materialization
  >     - Both positive and negative assertions for all applicable functions
  >     - Iterator-based test pattern (iter.Seq) for all test suites
  >
  > Outstanding todo items (minor):
  >
  > * ✅ Add integration test for YAMLEq.[🏁]
  > * ✅ Add unit tests for colorization StringColorizer.[🏁] (colors_test.go)
  > * ✅ Subtest naming - Some generic names could be more descriptive
  > * ✅ Full coverage is now properly integrated by codecov
  >   - Added -coverpkg option to shared test workflow in CI
  > * ♥️ A few recently added features in the generators are not fully tested (amounts for <.01% overall).[🏁]
  > * ♥️ Improve test coverage of the generators (integration tests).[🏁]

  3. ✅ Documentation

  > Achievements
  >
  > * ✅ Auto-generated documentation with clear and searchable domains to organize the big API ⭐⭐⭐
  > * ✅ Auto-generated testable examples that show up in godoc on pkg.go.dev ⭐⭐⭐
  > * ✅ Additional manually added ad'hoc testable examples to examplify more complex cases ⭐⭐⭐
  > * ✅ Document the code generation maintenance flow, with examples showing the full development workflow
  > * ✅ Production-ready documentation site (2026-01-01) ⭐⭐⭐
  >   - API Reference (A): Comprehensive, professional, well-organized
  >   - Examples (A): Complete transformation with 482-line EXAMPLES.md + comprehensive TUTORIAL.md
  >   - Project Documentation (A-): All text issues fixed, professional presentation
  >   - Maintainer Docs (A): 100% complete with excellent mermaid diagrams
  > * ✅ Publish documentation site to github pages (next step)
  > * ✅ Verify all README.md items are documented in plans
  > * ✅ Cross-check README roadmap with this plan file. Ensure nothing is missing.
  >
  > Outstanding todo items: N/A

  4. ✅ Respect of community standards (licensing, security, etc.).[😇]

  > Achievements
  >
  > * ✅ Standard documentation (README, LICENSE, SECURITY, CONTRIBUTING) ⭐
  > * ✅ Community-oriented: CONTRIBUTORS.md, discord server ⭐⭐
  > * ✅ Nice looking README with status badges ⭐
  > * ✅ Dependabot automatic updates, with auto-merge ⭐⭐
  > * ✅ Github automatic security scan ⭐
  > * ✅ Enable multi-tools for vulnerability scans (with re-aligned CI) ⭐⭐
  >
  > Outstanding todo items (important):
  > * ♥️ Need to update the standard docs in light of the improvements made during the doc site review/rewrite
  >   iterations.

  5. ✅ Showcase best practices

  > Achievements
  >
  > * ✅ Improved our own tests with a novel practice (iter.Seq for table-driven) ⭐⭐⭐
  > * ✅ Clearly separated out package API tests from testing internals (e.g. `equal_test.go` vs `equal_impl_test.go`) ⭐⭐
  >   ✅ Quality-focused: test error output and internal helpers ⭐⭐
  > * ✅ Clean architecture for the generator (minimalistic `main.go`, internal packages, specialized testable sub-packages) ⭐⭐
  > * ✅ Improved own tests with a clear namespace organization (in internal/assertions) ⭐
  >      Tests mirror implementation structure: e.g. `boolean_test.go` → `boolean.go`
  > * ✅ Parallel test execution at all levels ⭐
  > * ✅ Early return pattern in test helpers (2026-01-19) ⭐⭐
  >   - Demonstrates clean control flow patterns
  >   - Reduces cognitive complexity
  >   - Example in `.claude/skills/refactor-tests.md`
  > * ✅ Safe type assertion pattern (2026-01-19) ⭐⭐⭐
  >   - All type assertions checked with descriptive error messages
  >   - Zero tolerance for unchecked type assertions
  >   - Demonstrates Go safety best practices
  > * ✅ Comprehensive test coverage strategy ⭐⭐⭐
  >   - Layer 1: Exhaustive tests in internal/assertions (94% coverage)
  >   - Layer 2: Generated smoke tests in assert/require (~100% coverage)
  >   - Example-driven test generation from doc comments
  >
  > Outstanding todo items:
  > * ♥️ openCSF badge

  6. ✅ Innovation (AI-aided development, advanced go patterns) [🛠️]

  > Achievements:
  >
  > * ✅ Successfully leveraged Claude Code's assistance in the following areas ⭐⭐⭐
  >   - scaffolding great functionality in record time (comment parsing, markdown reformating)
  >   - writing exhaustive tests with golangci-lint compliance
  >   - project planning and tracking
  >   - documentation review and improvement
  >   - brainstorming & challenging ideas
  >   - test refactoring with pattern extraction
  >
  > * ✅ CLAUDE.md documentation ⭐⭐
  > * ✅ Claude Code skills library (4 specialized skills) ⭐⭐⭐
  >   - **golangci-lint-compliant-tests.md** (2026-01-26) ⭐⭐⭐
  >     - Patterns for writing linter-compliant tests
  >     - Eliminates code duplication in validation logic
  >     - Fixes gochecknoglobals, unparam, dupl issues
  >     - 80% code reduction example (347→267 lines)
  >   - **complex-conditionals-to-state-machine.md** ⭐⭐
  >     - State machine pattern for complexity reduction
  >     - Reduces complexity 40+ to distributed (18/11/8/4/3)
  >   - **refactor-tests.md** ⭐⭐
  >     - 11 patterns identified from actual refactoring work
  >     - Refactoring checklist and code examples
  >   - **3 specialized go AST handling skills** ⭐
  >     - go-types-ast-bridge.md, hugo-docs.md
  > * ✅ Project planning excellence ⭐⭐
  >
  > Outstanding todo items:
  >
  >  * 📝 Go development tools for Claude (MCP+specialized agents for recurring tasks)
  >  * 📝 Markdown linting agent (tool+specialized agent)
