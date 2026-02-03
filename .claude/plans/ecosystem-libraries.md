# Ecosystem Libraries: Exporting Spew and Difflib

**Status**: Proposal
**Priority**: Low-Medium (Opportunistic)
**Category**: Ecosystem Contribution
**Date**: 2026-01-23

## Executive Summary

Export testify's cleaned-up, optimized versions of `spew` (pretty-printing) and `difflib` (unified diff) as standalone modules for the broader Go community. This aligns with the "examplarity" mission and fills a maintenance gap in the Go ecosystem.

**Key Decision**: Separate modules in same repo, drop-in compatible with upstream, quality over adoption.

---

## Strategic Rationale

### The Opportunity

**Current State of Ecosystem:**

| Library | Status | Last Update | Issues |
|---------|--------|-------------|--------|
| **go-spew/spew** | Unmaintained | 2016 | Performance, bugs, stale |
| **pmezard/go-difflib** | Unmaintained | 2016 | Edge cases, formatting |
| **testify/internal/spew** | Active | 2026 | Optimized, clean |
| **testify/internal/difflib** | Active | 2026 | Enhanced, reliable |

**The Gap**: Widely-used libraries are unmaintained. Community needs quality alternatives.

### Why This Matters

**1. Ecosystem Contribution**
- Fill maintenance void in Go testing ecosystem
- Provide quality, maintained alternatives
- Demonstrate go-openapi's commitment to excellence

**2. Strategic Positioning**
- Establishes go-openapi as "quality testing tools" provider
- Attracts contributors who care about code quality
- Builds reputation beyond API tooling

**3. Validation of Work**
- Your cleanup/optimization already done (sunk cost)
- External adoption validates improvements
- Battle-tested in production (testify)

**4. Alignment with Mission**
- "Examplarity": Doing it right, showing how
- Zero dependencies maintained
- Performance-focused
- Clean, maintainable code

---

## Technical Approach

### Packaging Strategy: Separate Modules (Recommended)

```
github.com/go-openapi/testify/
├── v2/                          # Main testify module
├── spew/                        # Standalone spew module
│   ├── v2/
│   │   ├── go.mod               # Separate go.mod
│   │   ├── spew.go
│   │   ├── config.go
│   │   ├── dump.go
│   │   └── format.go
│   └── README.md
└── difflib/                     # Standalone difflib module
    ├── v2/
    │   ├── go.mod               # Separate go.mod
    │   ├── difflib.go
    │   └── unified.go
    └── README.md
```

**Module Paths:**
- `github.com/go-openapi/testify/spew/v2`
- `github.com/go-openapi/testify/difflib/v2`

**Benefits:**
- ✅ Independent versioning
- ✅ Can be used without testify dependency
- ✅ Clear separation of concerns
- ✅ Single repo (easier maintenance)
- ✅ Minimal coupling

**Testify Integration:**
```go
// testify/v2 depends on exported versions
import (
    "github.com/go-openapi/testify/spew/v2"
    "github.com/go-openapi/testify/difflib/v2"
)

// Remove internal/spew and internal/difflib
// Use exported versions instead
```

---

## Spew: Pretty-Printing Library

### Value Proposition

**What it does**: Deep pretty-printing of Go data structures for debugging and testing

**Why users need it**:
- Debug complex structures
- Generate test fixtures
- Logging and diagnostics
- Test failure messages

**Current pain points with go-spew**:
- Unmaintained since 2016
- Performance issues with large structures
- Limited configuration
- Bugs in edge cases

### Your Improvements (Document These!)

Based on testify's internal/spew:

**Performance Optimizations:**
- Reduced allocations in hot paths
- Better buffer management
- Optimized reflection usage
- Benchmark showing X% improvement

**Code Quality:**
- Modern Go practices (1.23+)
- Better error handling
- Cleaner API surface
- Comprehensive tests

**Features:**
- Enhanced configuration options
- Better formatting control
- Improved edge case handling
- Zero dependencies maintained

### API Strategy: Drop-In Compatible + Enhanced

**Phase 1: Compatible API**
```go
// Drop-in replacement for go-spew
package spew

// Compatible functions
func Dump(a ...any) string
func Fdump(w io.Writer, a ...any)
func Printf(format string, a ...any)
func Fprintf(w io.Writer, format string, a ...any)

// Config compatible
type Config struct {
    Indent         string
    MaxDepth       int
    DisableMethods bool
    // ... same as go-spew
}
```

**Phase 2: Enhanced API**
```go
// New, improved APIs (additive, non-breaking)
package spew

// Better control over output
type Formatter struct {
    config *Config
}

func NewFormatter(opts ...Option) *Formatter
func (f *Formatter) Format(a any) (string, error)
func (f *Formatter) FormatValue(v reflect.Value) (string, error)

// Functional options pattern
type Option func(*Config)

func WithMaxDepth(d int) Option
func WithIndent(s string) Option
func WithCompact(enabled bool) Option
```

**Usage Examples:**
```go
// Drop-in replacement
import "github.com/go-openapi/testify/spew/v2"

spew.Dump(complexStruct)  // Works exactly like go-spew

// Enhanced API
formatter := spew.NewFormatter(
    spew.WithMaxDepth(10),
    spew.WithCompact(true),
)
output, _ := formatter.Format(complexStruct)
```

### Migration Guide

**From go-spew:**
```go
// Before
import "github.com/davecgh/go-spew/spew"

// After (drop-in)
import "github.com/go-openapi/testify/spew/v2"

// No code changes needed!
```

**Benefits over go-spew:**
- ✅ Actively maintained
- ✅ X% faster (show benchmarks)
- ✅ Better formatting options
- ✅ Modern Go (1.23+)
- ✅ Comprehensive tests
- ✅ Zero dependencies

---

## Difflib: Unified Diff Library

### Value Proposition

**What it does**: Generate unified diffs between text sequences

**Why users need it**:
- Test failure messages (show expected vs actual)
- Code review tools
- File comparison utilities
- Version control integrations

**Current pain points with go-difflib**:
- Unmaintained since 2016
- Edge case bugs
- Limited formatting options
- Performance issues on large files

### Your Improvements

**Enhancements in testify's internal/difflib:**

**Correctness:**
- Fixed edge cases (empty files, trailing newlines)
- Better handling of large files
- Correct context handling
- Comprehensive test coverage

**Performance:**
- Optimized diff algorithm
- Better memory efficiency
- Reduced allocations
- Benchmark showing X% improvement

**Features:**
- Enhanced unified diff formatting
- Better control over context lines
- Cleaner API
- Modern Go practices

### API Strategy: Compatible + Improved

**Phase 1: Compatible API**
```go
// Drop-in replacement for go-difflib
package difflib

// Compatible types
type UnifiedDiff struct {
    A        []string
    FromFile string
    FromDate string
    B        []string
    ToFile   string
    ToDate   string
    Eol      string
    Context  int
}

// Compatible functions
func WriteUnifiedDiff(w io.Writer, diff UnifiedDiff) error
func GetUnifiedDiffString(diff UnifiedDiff) (string, error)
```

**Phase 2: Enhanced API**
```go
// New APIs for better control
package difflib

type Differ struct {
    config *Config
}

func NewDiffer(opts ...Option) *Differ
func (d *Differ) Diff(a, b []string) (string, error)
func (d *Differ) DiffStrings(a, b string) (string, error)

// Functional options
type Option func(*Config)

func WithContext(lines int) Option
func WithColor(enabled bool) Option
func WithCompact(enabled bool) Option
```

### Migration Guide

**From go-difflib:**
```go
// Before
import "github.com/pmezard/go-difflib/difflib"

// After (drop-in)
import "github.com/go-openapi/testify/difflib/v2"

// No code changes needed!
```

---

## Implementation Plan

### Phase 1: Setup and Extraction (Week 1-2)

**1. Create Module Structure**
```bash
# Create separate modules
cd github.com/go-openapi/testify
mkdir -p spew/v2
mkdir -p difflib/v2

# Initialize modules
cd spew/v2 && go mod init github.com/go-openapi/testify/spew/v2
cd difflib/v2 && go mod init github.com/go-openapi/testify/difflib/v2
```

**2. Extract Code**
- Copy from `internal/spew` → `spew/v2/`
- Copy from `internal/difflib` → `difflib/v2/`
- Make APIs public (unexport internals)
- Ensure backward compatibility

**3. Update Testify**
```go
// testify/v2/go.mod
require (
    github.com/go-openapi/testify/spew/v2 v2.0.0
    github.com/go-openapi/testify/difflib/v2 v2.0.0
)

// Replace internal imports
// internal/spew → github.com/go-openapi/testify/spew/v2
// internal/difflib → github.com/go-openapi/testify/difflib/v2
```

**4. Testing**
- Ensure all testify tests still pass
- No regression in functionality
- Validate module independence

### Phase 2: Documentation (Week 3)

**1. README Files**

**spew/README.md:**
```markdown
# Spew: Deep Pretty-Printing for Go

A maintained, optimized fork of go-spew with better performance and modern Go practices.

## Why Testify Spew?

- ✅ **Actively Maintained**: Regular updates, bug fixes
- ✅ **Better Performance**: X% faster than go-spew (see benchmarks)
- ✅ **Modern Go**: Uses Go 1.23+ features
- ✅ **Zero Dependencies**: No external deps
- ✅ **Drop-In Compatible**: Works with existing go-spew code

## Installation

go get github.com/go-openapi/testify/spew/v2

## Quick Start

[Examples...]

## Migration from go-spew

[Migration guide...]

## Benchmarks

[Comparative benchmarks...]
```

**difflib/README.md:** Similar structure

**2. Godoc Examples**
```go
package spew_test

import (
    "fmt"
    "github.com/go-openapi/testify/spew/v2"
)

func ExampleDump() {
    type Person struct {
        Name string
        Age  int
    }

    p := Person{Name: "Alice", Age: 30}
    spew.Dump(p)
    // Output:
    // (spew_test.Person) {
    //  Name: (string) (len=5) "Alice",
    //  Age: (int) 30
    // }
}

func ExampleNewFormatter() {
    formatter := spew.NewFormatter(
        spew.WithMaxDepth(5),
        spew.WithCompact(true),
    )

    data := map[string]any{"key": "value"}
    output, _ := formatter.Format(data)
    fmt.Println(output)
}
```

**3. Migration Guides**
- Create `spew/MIGRATION.md`
- Create `difflib/MIGRATION.md`
- Document improvements and changes
- Provide side-by-side examples

### Phase 3: Benchmarking (Week 4)

**1. Create Comparative Benchmarks**

**spew/v2/bench_test.go:**
```go
package spew_test

import (
    "testing"

    oldspew "github.com/davecgh/go-spew/spew"
    newspew "github.com/go-openapi/testify/spew/v2"
)

func BenchmarkDump_Old(b *testing.B) {
    data := createComplexStruct()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        oldspew.Sdump(data)
    }
}

func BenchmarkDump_New(b *testing.B) {
    data := createComplexStruct()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        newspew.Sdump(data)
    }
}
```

**2. Document Results**
```markdown
## Performance Comparison

| Operation | go-spew | testify/spew | Improvement |
|-----------|---------|--------------|-------------|
| Dump (small) | 1234 ns/op | 890 ns/op | 28% faster |
| Dump (large) | 45.6 µs/op | 32.1 µs/op | 30% faster |
| Format (nested) | 2.3 µs/op | 1.6 µs/op | 30% faster |
```

### Phase 4: Soft Launch (Week 5-6)

**1. Internal Adoption**
- Use in go-openapi projects (go-swagger, spec, etc.)
- Gather feedback from team
- Identify any issues
- Refine documentation

**2. Limited Announcement**
- Mention in testify v2 release notes
- Post on go-openapi GitHub discussions
- Get feedback from existing users
- Don't promote widely yet

**3. Iterate Based on Feedback**
- Fix any reported issues
- Improve documentation
- Add requested features
- Stabilize API

### Phase 5: Public Launch (Week 7+)

**Only if soft launch successful:**

**1. Create Announcement Content**
- Blog post: "Introducing testify/spew and testify/difflib"
- Highlight improvements, benchmarks
- Show migration examples
- Emphasize maintenance commitment

**2. Promote Strategically**
- Reddit /r/golang
- Twitter/X Go community
- golang-nuts mailing list
- Gophers Slack

**3. Engage Community**
- Respond to feedback
- Accept PRs
- Fix issues promptly
- Build trust

---

## Success Criteria

### Minimum Viable Success (Soft Launch)
- ✅ Modules compile and pass all tests
- ✅ Testify v2 successfully uses exported versions
- ✅ go-openapi projects adopt without issues
- ✅ Documentation clear and complete
- ✅ No regressions vs internal versions

### Public Launch Success
- ✅ 10+ external projects adopt
- ✅ Positive community feedback
- ✅ Zero critical bugs reported
- ✅ Performance benchmarks validated
- ✅ Documentation praised

### Long-Term Success
- ✅ Becomes de facto maintained alternative
- ✅ Regular external contributions
- ✅ Used in high-profile projects
- ✅ Cited as example of quality Go library

---

## Risk Management

### Risks and Mitigations

**Risk**: Fragmenting ecosystem with yet another fork
- **Mitigation**: Emphasize maintenance + quality, not just "different"
- **Validation**: Show benchmarks, document improvements clearly

**Risk**: Low adoption (people stick with go-spew)
- **Mitigation**: Drop-in compatibility makes switching painless
- **Acceptance**: Even if low adoption, helps go-openapi projects
- **Fallback**: Keep as internal modules if no external interest

**Risk**: Maintenance burden increases
- **Mitigation**: Start small, expand based on demand
- **Limit**: Don't accept every feature request
- **Strategy**: Quality over features

**Risk**: Upstream maintainers return
- **Mitigation**: Welcome! Offer to contribute improvements back
- **Strategy**: Position as collaboration, not competition

**Risk**: Breaking changes needed
- **Mitigation**: Separate v2 module, maintain v1 compatibility
- **Strategy**: Additive changes only in minor versions

### Exit Strategy

**If adoption is low:**
- Keep as internal modules
- Maintain for go-openapi projects only
- No public marketing
- Archive public modules gracefully

**If maintenance becomes burden:**
- Clearly communicate status
- Accept maintainer help
- Consider donating to foundation
- Don't just abandon

---

## Effort Estimation

### Initial Setup (Spew Only)
- Module creation: 4 hours
- Code extraction: 8 hours
- Testing: 8 hours
- Documentation: 16 hours
- Benchmarking: 8 hours
- **Total: ~44 hours (~1 week)**

### Difflib Addition
- Similar effort: ~40 hours
- **Total: ~1 week**

### Ongoing Maintenance
- Issues/PRs: ~2-4 hours/week
- Updates: ~4 hours/month
- Documentation: ~2 hours/month
- **Total: ~10-15 hours/month if active adoption**

---

## Prioritization

**Relative to v3 Roadmap:**
- Lower priority than core features
- Opportunistic (do if time permits)
- Parallel track (doesn't block v3)

**When to do:**
- After Phase 1 (Foundation) complete
- During Phase 2/3 (error-aware + safety)
- As "break" from main feature work
- February timeframe alongside v3 work

**Sequencing:**
1. Start with **spew only** (higher demand)
2. Validate approach, gather feedback
3. Add **difflib** if spew successful
4. Expand based on actual adoption

---

## Decision Points

### Go/No-Go Decisions

**Before Starting:**
- ✅ Is code cleanup complete?
- ✅ Are improvements documented?
- ✅ Is team bandwidth available?

**Before Public Launch:**
- ✅ Soft launch successful?
- ✅ No critical issues found?
- ✅ Documentation complete?
- ✅ Performance validated?

**After 6 Months:**
- ❓ Has adoption grown?
- ❓ Is maintenance sustainable?
- ❓ Is community engaged?

**Decision**: Continue, scale back, or archive accordingly

---

## Documentation Deliverables

### For Each Module

**1. README.md**
- Why use this library
- Installation
- Quick start
- Migration from upstream
- Benchmarks
- Contributing

**2. MIGRATION.md**
- Detailed migration guide
- API differences
- Improvements list
- Side-by-side examples

**3. CHANGELOG.md**
- Version history
- Improvements over upstream
- Breaking changes (if any)

**4. Godoc Examples**
- All public functions
- Common use cases
- Advanced patterns

**5. Benchmarks**
- Comparative benchmarks
- Performance analysis
- When improvements matter

---

## Marketing Strategy (If Pursuing Public Adoption)

### Positioning

**Headline**: "Maintained, optimized alternatives to go-spew and go-difflib"

**Key Messages:**
1. **Quality**: Battle-tested in testify, cleaned up, optimized
2. **Maintenance**: Actively maintained, regular updates
3. **Performance**: Measurably faster (show benchmarks)
4. **Compatibility**: Drop-in replacement, zero risk
5. **Trust**: From go-openapi, proven track record

### Content Plan

**Launch Content:**
- Blog post: "Why we forked spew and difflib"
- Technical deep-dive: "Optimizing Go pretty-printing"
- Migration guide: "Switching from go-spew"

**Ongoing Content:**
- Performance tips
- Advanced usage examples
- Community showcase

### Channels

**Primary:**
- GitHub (repo README, releases)
- go-openapi blog
- Twitter/X

**Secondary:**
- Reddit /r/golang
- Gophers Slack
- golang-nuts

**Tertiary:**
- Conference talks (if successful)
- Podcast appearances
- Guest blog posts

---

## Relationship to Testify

### Integration Strategy

**Testify depends on exported versions:**
```go
// testify/v2/go.mod
module github.com/go-openapi/testify/v2

require (
    github.com/go-openapi/testify/spew/v2 v2.0.0
    github.com/go-openapi/testify/difflib/v2 v2.0.0
)
```

**Benefits:**
- ✅ Dogfooding (we use what we recommend)
- ✅ Quality signal (testify depends on it)
- ✅ Shared maintenance (improvements benefit both)

**Considerations:**
- Circular dependency avoided (separate modules)
- Version synchronization needed
- Breaking changes must be coordinated

### Versioning Strategy

**Independent Versioning:**
- spew/v2: Own semver
- difflib/v2: Own semver
- testify/v2: Own semver

**Compatibility Commitment:**
- spew/v2 and difflib/v2 must remain stable
- Breaking changes require major version bump
- Testify can depend on specific versions

---

## Next Steps

### Immediate (If Approved)
1. ⏳ Finalize code cleanup for spew
2. ⏳ Create spew/v2 module structure
3. ⏳ Extract code to spew/v2
4. ⏳ Write initial documentation
5. ⏳ Create benchmarks vs go-spew

### Short-Term (Soft Launch)
6. ⏳ Test in go-openapi projects
7. ⏳ Gather internal feedback
8. ⏳ Refine based on usage
9. ⏳ Prepare for potential public release

### Long-Term (If Successful)
10. ⏳ Public announcement
11. ⏳ Community engagement
12. ⏳ Add difflib/v2 similarly
13. ⏳ Ongoing maintenance

---

## Conclusion

**This is an opportunistic ecosystem contribution** that:
- ✅ Aligns with examplarity mission
- ✅ Leverages work already done
- ✅ Fills real gap in Go ecosystem
- ✅ Low risk (soft launch first)
- ✅ Moderate effort (mostly docs)
- ✅ High potential impact (if adopted)

**Recommendation**: Proceed with **spew first**, soft launch internally, evaluate adoption, then decide on public promotion and difflib.

**Timeline**: February alongside v3 work, as time permits.

**Success Definition**: Even if public adoption is low, validates our work and helps go-openapi projects. If adoption is high, becomes strategic asset for go-openapi brand.

---

## Related Documents

- [v3-roadmap.md](./v3-roadmap.md) - Main feature roadmap
- [COMPETITIVE_ANALYSIS.md](./COMPETITIVE_ANALYSIS.md) - Ecosystem positioning
- Testify internal/spew and internal/difflib - Source code
