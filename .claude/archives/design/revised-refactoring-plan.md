# Revised Refactoring Plan - Simplified

## Goal

Split well-isolated, testable chunks without separating the generators.

## Rationale

**Keep generators together:**
- ✅ Sharing options is obvious (both embed same struct)
- ✅ Sharing render utilities is natural (same package)
- ✅ Simpler structure
- ✅ No circular dependency concerns

**Split what's truly independent:**
- ✅ Funcmaps - Pure template functions, no dependencies
- ✅ Domains - Well-isolated domain organization logic

## Proposed Structure

```
generator/
├── funcmaps/              - NEW: Template functions
│   ├── doc.go
│   ├── funcmaps.go        (~300 lines) - Core template functions
│   ├── funcmaps_test.go
│   ├── markdown.go        (~150 lines) - Markdown processing
│   └── markdown_test.go
│
├── domains/               - NEW: Domain organization
│   ├── doc.go
│   ├── domains.go         (~366 lines) - Domain metadata & organization
│   └── domains_test.go    - NEW: Tests for domain logic
│
├── doc.go                 - Package documentation
├── options.go             (123 lines) - Shared options
├── render.go              (36 lines)  - Shared rendering
├── generator.go           (494 lines) - Code generator
├── doc_generator.go       (238 lines) - Doc generator
└── templates/             - Template files
    ├── assertion_*.gotmpl
    ├── requirement_*.gotmpl
    └── doc_*.gotmpl
```

## What Changes

### Before
```
generator/
├── funcmap.go           (381 lines)
├── funcmap_enhanced.go  (136 lines)
├── doc_domains.go       (366 lines)
├── generator.go         (494 lines)
├── doc_generator.go     (238 lines)
├── options.go           (123 lines)
├── render.go            (36 lines)
└── templates/
```

**Total in main package: ~1,774 lines**

### After
```
generator/
├── funcmaps/            (~450 lines total)
├── domains/             (~366 lines total)
├── generator.go         (494 lines)
├── doc_generator.go     (238 lines)
├── options.go           (123 lines)
├── render.go            (36 lines)
└── templates/
```

**Total in main package: ~891 lines** (50% reduction)
**Well-organized sub-packages: ~816 lines**

## Phase 1: Split Funcmaps

See **FUNCMAPS_SPLIT.md** for details.

**Summary:**
- Move funcmap.go → funcmaps/funcmaps.go
- Move funcmap_enhanced.go → funcmaps/markdown.go
- Move tests → funcmaps/*_test.go
- Update generator.go and doc_generator.go to use `funcmaps.FuncMap()`

**Result:**
```go
// generator/generator.go
import "github.com/go-openapi/testify/v2/codegen/internal/generator/funcmaps"

func (g *Generator) loadTemplates() error {
    tpl, err := template.New(name).Funcs(funcmaps.FuncMap()).ParseFS(...)
}

// generator/doc_generator.go
import "github.com/go-openapi/testify/v2/codegen/internal/generator/funcmaps"

func (d *DocGenerator) loadTemplates() error {
    tpl, err := template.New(name).Funcs(funcmaps.FuncMap()).ParseFS(...)
}
```

## Phase 2: Split Domains

### Current State

```go
// generator/doc_domains.go (366 lines)
package generator

type domainInfo struct { ... }
var domainMetadata = map[string]domainInfo{ ... }
var domainRank = map[string]int{ ... }
func rankDomain(domain string) int { ... }
func descriptiveDomain(domain string) string { ... }
func titleDomain(domain string) string { ... }
func keywordsDomain(domain string) string { ... }
```

### Proposed Structure

```go
// generator/domains/doc.go
// Package domains provides domain organization and metadata for assertions.

// generator/domains/domains.go
package domains

// Info contains metadata about an assertion domain
type Info struct {
    Title       string
    Description string
    Keywords    []string
}

// Metadata returns domain information for a given domain
func Metadata(domain string) Info

// Rank returns the sort order for a domain
func Rank(domain string) int

// Title returns the display title for a domain
func Title(domain string) string

// Description returns the description for a domain
func Description(domain string) string

// Keywords returns the keywords for a domain
func Keywords(domain string) []string

// All returns all known domains sorted by rank
func All() []string

// Private data
var metadata = map[string]Info{ ... }
var ranking = map[string]int{ ... }
```

### Usage

```go
// generator/doc_generator.go
import "github.com/go-openapi/testify/v2/codegen/internal/generator/domains"

func (d *DocGenerator) reorganizeByDomain() (iter.Seq2[string, model.Document], uniqueValues) {
    // ...
    for domain := range domainsSeen {
        rank := domains.Rank(domain)
        title := domains.Title(domain)
        // ...
    }
}

func (d *DocGenerator) buildIndexDocument(...) model.Document {
    // ...
    Title:       domains.Title(domain),
    Description: domains.Description(domain),
    Keywords:    domains.Keywords(domain),
}
```

### Testing Benefits

```go
// generator/domains/domains_test.go
package domains

func TestDomainRank(t *testing.T) {
    tests := []struct {
        domain   string
        expected int
    }{
        {"boolean", 1},
        {"equality", 2},
        {"common", 100},
    }

    for _, tt := range tests {
        t.Run(tt.domain, func(t *testing.T) {
            if got := Rank(tt.domain); got != tt.expected {
                t.Errorf("Rank(%q) = %d, want %d", tt.domain, got, tt.expected)
            }
        })
    }
}

func TestAllDomains(t *testing.T) {
    all := All()

    // Test they're sorted by rank
    for i := 1; i < len(all); i++ {
        if Rank(all[i-1]) > Rank(all[i]) {
            t.Errorf("Domains not sorted: %s (rank %d) before %s (rank %d)",
                all[i-1], Rank(all[i-1]), all[i], Rank(all[i]))
        }
    }
}

func TestDomainMetadata(t *testing.T) {
    info := Metadata("boolean")

    if info.Title != "Boolean" {
        t.Errorf("Title = %q, want %q", info.Title, "Boolean")
    }

    if info.Description != "Asserting Boolean Values" {
        t.Errorf("Description = %q, want %q", info.Description, "Asserting Boolean Values")
    }
}
```

## Migration Steps

### Step 1: Create funcmaps package
```bash
mkdir -p internal/generator/funcmaps
# Move files as per FUNCMAPS_SPLIT.md
```

### Step 2: Create domains package
```bash
mkdir -p internal/generator/domains
```

### Step 3: Extract domains.go

```go
// generator/domains/domains.go
package domains

type Info struct {
    Title       string
    Description string
    Keywords    []string
}

var metadata = map[string]Info{
    "boolean": {
        Title:       "Boolean",
        Description: "Asserting Boolean Values",
        Keywords:    []string{"True", "False"},
    },
    // ... rest from doc_domains.go
}

var ranking = map[string]int{
    "boolean":    1,
    "collection": 2,
    // ... rest from doc_domains.go
}

func Metadata(domain string) Info {
    if info, ok := metadata[domain]; ok {
        return info
    }
    return Info{
        Title:       titleCase(domain),
        Description: "Uncategorized",
    }
}

func Rank(domain string) int {
    if rank, ok := ranking[domain]; ok {
        return rank
    }
    return 100 // Default rank for unknown domains
}

func Title(domain string) string {
    return Metadata(domain).Title
}

func Description(domain string) string {
    return Metadata(domain).Description
}

func Keywords(domain string) []string {
    return Metadata(domain).Keywords
}

func All() []string {
    domains := make([]string, 0, len(metadata))
    for domain := range metadata {
        domains = append(domains, domain)
    }

    sort.Slice(domains, func(i, j int) bool {
        return Rank(domains[i]) < Rank(domains[j])
    })

    return domains
}
```

### Step 4: Update doc_generator.go

```go
// generator/doc_generator.go
import (
    "github.com/go-openapi/testify/v2/codegen/internal/generator/domains"
)

func (d *DocGenerator) reorganizeByDomain() (iter.Seq2[string, model.Document], uniqueValues) {
    // Replace: rankDomain() → domains.Rank()
    // Replace: titleDomain() → domains.Title()
}

func (d *DocGenerator) buildIndexDocument(...) model.Document {
    // Replace: titleDomain() → domains.Title()
    // Replace: descriptiveDomain() → domains.Description()
    // Replace: keywordsDomain() → domains.Keywords()
}
```

### Step 5: Remove old file
```bash
git rm internal/generator/doc_domains.go
```

### Step 6: Add tests
```bash
# Create comprehensive tests for domains package
touch internal/generator/domains/domains_test.go
```

## Benefits

### Funcmaps Split
✅ Template functions independently testable
✅ No coupling to generator internals
✅ Reusable across both generators
✅ Clear, focused responsibility

### Domains Split
✅ Domain logic independently testable
✅ Can verify domain metadata correctness
✅ Can test rank ordering
✅ Easy to add new domains
✅ Clear separation from generation logic

### Generators Stay Together
✅ Shared options (obvious embedding)
✅ Shared render utilities (same package)
✅ No circular dependencies
✅ Simpler mental model
✅ Easy to see both generation strategies

## Testing Strategy

### Before Split
```bash
go test ./internal/generator
```

### After Split
```bash
# Test each component independently
go test ./internal/generator/funcmaps -v
go test ./internal/generator/domains -v

# Test generators still work
go test ./internal/generator -v

# Run all
go test ./internal/generator/... -v
```

## File Count Comparison

**Before:**
- 8 files in generator/ (including tests)
- ~1,774 lines in main package

**After:**
- 4 files in generator/
- 6 files in generator/funcmaps/
- 3 files in generator/domains/
- ~891 lines in main package
- Well-organized sub-packages with clear responsibilities

## Summary

This simplified approach:
1. ✅ Splits well-isolated, testable chunks (funcmaps, domains)
2. ✅ Keeps generators together (simpler, obvious sharing)
3. ✅ Avoids options complexity (both in same package)
4. ✅ Improves testability (sub-packages independently tested)
5. ✅ Reduces main package size by 50%
6. ✅ No breaking changes to public API
7. ✅ Clear path for future refactoring if needed
