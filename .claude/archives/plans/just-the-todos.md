# DOC

Simplified project plan with only the todo list with status.

## POLISH

   ✅ Fixed repeated sections in doc site [📚]
   ✅ **Address duplicate markdown documents** [📚] Now minimal: security & license stuff only
   ✅ run mdsf to format code snippets in docs [📚]
   ✅ Roadmap to announce forthcoming features [📚]
   ✅ Polish existing godoc[📚]
   ✅ Complete doc review by github agent [📚]

## DOC GENERATION

   ✅ Generated doc shows tab for any godoc section, not just examples and usage (ex: concurrency) [📚]
   ✅ Render godoc links to current package [📚]
   ✅ (templates /ExtraComments) Multiline private notes won't render great in the template. Need some markdown reformating for correct rendering as a markdown blockquote.
   ✅ (comments parser) Maintainer, note etc private comment annotation don't support pluralization (e.g. "maintainers:" is not
      detected. [📚]
   ✅ In generated doc, internal tab for generics, function signature doesn't show the type parameters
   ✅ Render testable examples in generated doc[📚] This one is hard: see github.com/golang/pkgsite/godoc/dochtml/dochtml.go

   ✅ Unstable package name rendered in testable examples (sometimes from package assert, sometimes from package require)
   ✅ Render multiple examples: assert, require
   ✅ Testable example for NoGoRoutineLeak

   ♥️ Render multiple examples: multiple success or failure values
   ⛔ ~Replace or extend "Usage:" sections in godoc with references to generated examples (not sure yet this will actually improve anything)~
  

## RICHER CONTENT

   ✅ **Advanced examples: async test** [📚]
   ✅ Philosophy / compare to gingko [📚]
   ✅ examples(async): mention that this is a test code small anyway (flaky tests etc)
   ✅ **Document custom YAML serialization injection** [📚]
   ✅ **Repo architecture re opt-in dependencies pattern** [📚]

   ♥️ **Educational content expansion** [📚]
   ♥️ difflib doc could borrow more from the updated source https://docs.python.org/3/library/difflib.html#module-difflib [📚]
   ♥️ Showcase best practices
   

## HUGO-related

   ✅ doc: hugo theme style: style cards to be wider [📚]
   📝 **Configure doc site versioning** [📚🛠️]
   🧪 **News section for doc site** [📚]
   ♥️ **Improve navigation for large documentation pages** [📚🎨]
   ♥️ **Assertion index side-bar**/**All core assertions index**/**All generics index** [📚🎨]
   ♥️ doc: use hugo file resource to import go code in markdown (easier to format source/alternative to mdsf) [📚]
   🧪 **Custom card shortcode for doc site** [📚]


# CI

   ✅ Fix mono-repo release notes
   ♥️ CI: Run broken link check [📚]
   ♥️ CI: Run go code snippets formating (mdsf) [📚]
   ♥️ run mdsf in CI / githook [📚] (attention! may break imports) (ci-worflows project)
   ♥️ CI: automate doc contributors update for doc site [📚] (ci-worflows project)
   ♥️ CI: re-instate past project with doc spellcheck & markdownlint [📚] (ci-workflows project)

# CODE QUALITY

   ✅ **Improve private comments in internal/assertions**
   ✅ **Fix remaining linter issues** [😇]: spew

   ✅ **Fix remaining linter issues** [😇] : now it's more about reducing nolint directives (we have 66 - most a justified - a few remain temporary (complexity))
   ✅ Address issues reported by codeFactor.io (-> complex functions, overlaps linting) (3/6 fixes - remaining "complex" functions are reasonable)
   ✅ **Remove unnecessary type arguments in generated code**


# MAINTAINABILITY

   ✅ Code generator systematically changes files, if only to modify the sha and timestamp. This produces a lot of noise.
   ✅ Rename EventuallyWithT (conflicts with the convention adopted for generics)
   ✅ (comments parser) Domain descriptions are only parsed from private comments. We might want to add them
      to the godoc docstring.
   ✅ internal/assertions: Testing messages is difficult to maintain

   ✅ **Simplify code generator templates** [🛠️]
      * ✅ use funcmaps to replace {{ printf "{{" }}: tabs-etabs/tab-etab/expand-expand
      * ✅simplify the huge FormatMarkdown function
   ✅ codegen/internal/generator/funcmaps: refact code and tests (markdown.go)

   ♥️ **Remove extraneous helper functions**
   ♥️ Function types (e.g., `ValueAssertionFunc`, `PanicAssertionFunc`) would need some proper comment for categorization. [📚]

# PERF
   ♥️ **Difflib memory allocation optimization** [⚡]
   ♥️ graphic visualization for benchmarks [📚]

# TESTS

   ✅ Only coverage from the main module is recovered [🏁] - don't have codegen, enable/yaml and enable/colors any longer
   ✅ Missing coverage largely concentrated in difflib
   ✅ Missing coverage now moved to testintegration and codegen mostly [🏁]
     ✅* testintegration/spew : unavoidable (unreachable code)
     ✅* testintegration/yaml : ok
     ✅* testintegration/colors : ok
   ✅ **Improve test coverage for error messages** [🏁]
   ✅ **Improve test coverage of generators** [🏁]

   ♥️ **Improve test coverage for generated helpers** [🏁]
   ♥️ Code generator testing using golden file. Currently the generator integration test is merely a smoke test. 
   ⛔ fuzz test in CI reported an error: fix it (issue: CI doesn't upload the failing case, only the corpus. Need to fix CI first)
      - can't reproduce with a 10x longer, 10x larger test sample
      - Most likely, CI did hit a timeout ("hang" failure) when overloaded with tests.

# FEATURES

  ## v2.3
   ✅ **UnmarshalJSONAsT[T]** [🧪] (Phase 2.1)
   ✅ **UnmarshalYAMLAsT[T]** [🧪] (Phase 2.2)
   ✅ Demonstrate extensibility [🧪] (Phase 1.2)
   ✅ **Make Assertions implement TestingT interface** [🧪♥️]
      - actually that is not possible: Errorf conflicts with a different signature
      - used Assertions.T to work around
      - added adhoc examples

   ✅ **NoGoroutineLeak** [🧪⚠️] (Phase 3.1)

   ⛔ **Implement Result[T] Pattern** [🧪] (Phase 1.2)
   ⛔ **Update codegen for Result[T] returns** [🛠️]

  ## v2.4
   📝 **NoFileDescriptorLeak (Unix)** [🧪⚠️] (Phase 3.2)
   📝 **EventuallyT[T]** [🧪] (Phase 2.3) (with support for returned error)
   📝 **EventuallyWithContextT[T]** [🧪] (Phase 2.4)
   📝 **JSONPointerT[T]** [🧪] (Phase 4.1)
   ♥️ **Better time comparison assertions**
   ⛔ **custom JSON serializer injection**

  ## v2.5
   📝 **Eventually and all -> synctest** [🧪⚠️]
   📝 **NoFileDescriptorLeak (Windows)** [🧪⚠️]
   📝 **Upstream PR candidates** (monitoring)
   ♥️ **JSON/YAML equivalence vs equality**

# TOOLING

   ✅ Innovation (AI-aided development, skills with advanced go patterns)
   ✅ Running codegen with "go generate" does not produce exactly the same result as when running the binary from ./codegen:
      tool and headers are empty.

   ♥️ Benchmark viz
   ⏳ go-fred MCP
      * ✅ godoc-check with agent:
           a MCP tool that runs a pipeline with hunspell, godoc-specific filters and a bulk
           in-place comment update tool
      * ⏳test callgraph
   ♥️ git-janitor
   ♥️ go4us: TBD
      * auto-fix feature for funcorder, acting as a custom formatter for golangci-lint, with Fred's ordering rules

