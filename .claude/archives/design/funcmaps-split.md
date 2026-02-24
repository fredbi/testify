# Funcmaps Package Split - Phase 1

## Goal

Extract funcmaps into a separate, testable sub-package **before** splitting the generators.

## Current State

```
generator/
├── funcmap.go           (381 lines) - Template functions
├── funcmap_enhanced.go  (136 lines) - Markdown processing
├── funcmap_enhanced_test.go         - Tests
└── generator.go         (494 lines) - Uses funcMap()
```

**Total funcmaps: ~517 lines**

## Proposed Structure

```
generator/
├── funcmaps/
│   ├── doc.go
│   ├── funcmaps.go          - Core template functions (from funcmap.go)
│   ├── funcmaps_test.go     - Tests for core functions
│   ├── markdown.go          - Markdown processing (from funcmap_enhanced.go)
│   └── markdown_test.go     - Tests for markdown (from funcmap_enhanced_test.go)
│
├── generator.go             - Uses funcmaps.FuncMap()
└── doc_generator.go         - Uses funcmaps.FuncMap()
```

## What Goes in `funcmaps/`

### funcmaps.go (Core Functions)
From current `funcmap.go`:
- ✅ `FuncMap()` - Main function map constructor
- ✅ `printImports()` - Import formatting
- ✅ `comment()` - Comment formatting
- ✅ `params()` - Parameter formatting
- ✅ `forward()` - Forward function formatting
- ✅ `printReturns()` - Return value formatting
- ✅ `docStringFor()` - Doc string generation
- ✅ `docStringPackage()` - Package-specific doc strings
- ✅ `sourceLink()` - GitHub source links
- ✅ `titleize()` - Title casing
- ✅ `quote()` - String quoting
- ✅ `godocbadge()` - Godoc badge URL
- ✅ `relocate()` - Test value relocation
- ✅ Helper functions: `concatStrings()`, `pathParts()`, etc.

### markdown.go (Markdown Processing)
From current `funcmap_enhanced.go`:
- ✅ `MarkdownFormat()` - Enhanced markdown formatting
- ✅ Reference link processing
- ✅ Godoc link conversion
- ✅ Hugo shortcode generation

## Public API

```go
package funcmaps

import "text/template"

// FuncMap returns the complete function map for templates
func FuncMap() template.FuncMap

// Individual functions (if needed for testing)
func FormatMarkdown(in string) string
func FormatComment(str string) string
func FormatParams(args model.Parameters) string

// ... etc
```

## Benefits

✅ **Independently testable** - Can test funcmaps without generator
✅ **No options coupling** - Funcmaps are pure functions, no config needed
✅ **Reusable** - Both code and doc generators use same funcmap
✅ **Clear scope** - Template functions only
✅ **Easy to extend** - Add new template functions in one place

## Migration Steps

### Step 1: Create funcmaps package

```bash
mkdir -p internal/generator/funcmaps
```

### Step 2: Move and adapt funcmap.go

```go
// internal/generator/funcmaps/funcmaps.go
package funcmaps

import (
    "fmt"
    "text/template"

    "github.com/go-openapi/testify/v2/codegen/internal/model"
)

// FuncMap returns the template function map
func FuncMap() template.FuncMap {
    return map[string]any{
        "imports":          printImports,
        "comment":          comment,
        "params":           params,
        "forward":          forward,
        "docStringFor":     docStringFor,
        "docStringPackage": docStringPackage,
        "returns":          printReturns,
        "concat":           concatStrings,
        "pathparts":        pathParts,
        "relocate":         relocate,
        "hasSuffix":        strings.HasSuffix,
        "sourceLink":       sourceLink,
        "titleize":         titleize,
        "quote":            quote,
        "mdformat":         FormatMarkdown,  // From markdown.go
        "godocbadge":       godocbadge,
    }
}

// printImports formats import statements
func printImports(in model.ImportMap) string {
    // ... existing implementation
}

// ... all other functions from funcmap.go
```

### Step 3: Move and adapt funcmap_enhanced.go

```go
// internal/generator/funcmaps/markdown.go
package funcmaps

import (
    "fmt"
    "regexp"
    "strings"
)

// FormatMarkdown processes markdown in godoc comments
// Handles reference links, godoc links, and Hugo shortcodes
func FormatMarkdown(in string) string {
    // ... existing implementation from markdownFormatEnhanced
}

// Helper functions for markdown processing
func extractReferenceLinks(in string) map[string]string {
    // ... extracted logic
}

func convertGodocLinks(in string) string {
    // ... extracted logic
}
```

### Step 4: Move tests

```go
// internal/generator/funcmaps/markdown_test.go
package funcmaps

import "testing"

func TestFormatMarkdown(t *testing.T) {
    // ... existing tests from funcmap_enhanced_test.go
}

func TestReferenceLinks(t *testing.T) { ... }
func TestGodocLinks(t *testing.T) { ... }
```

### Step 5: Update generator.go

```go
// internal/generator/generator.go
package generator

import (
    "github.com/go-openapi/testify/v2/codegen/internal/generator/funcmaps"
)

func (g *Generator) loadTemplates() error {
    // ... existing code

    // OLD: funcMap()
    // NEW: funcmaps.FuncMap()
    tpl, err := template.New(name).Funcs(funcmaps.FuncMap()).ParseFS(templatesFS, pattern)

    // ... rest unchanged
}
```

### Step 6: Update doc_generator.go

```go
// internal/generator/doc_generator.go
package generator

import (
    "github.com/go-openapi/testify/v2/codegen/internal/generator/funcmaps"
)

func (d *DocGenerator) loadTemplates() error {
    // ... existing code

    // OLD: funcMap()
    // NEW: funcmaps.FuncMap()
    tpl, err := template.New(name).Funcs(funcmaps.FuncMap()).ParseFS(templatesFS, pattern)

    // ... rest unchanged
}
```

### Step 7: Remove old files

```bash
git rm internal/generator/funcmap.go
git rm internal/generator/funcmap_enhanced.go
git mv internal/generator/funcmap_enhanced_test.go \
	internal/generator/funcmaps/markdown_test.go
```

## Testing Strategy

### Before Split (Current)
```bash
go test ./internal/generator -run TestMarkdownFormatEnhanced
```

### After Split
```bash
# Test funcmaps independently
go test ./internal/generator/funcmaps -v

# Test generators still work
go test ./internal/generator -v
```

## No Breaking Changes

✅ **Generator API unchanged** - Still `New()`, `Generate()`
✅ **DocGenerator API unchanged** - Still works the same
✅ **Templates unchanged** - Same function names available
✅ **Main.go unchanged** - No import changes needed

## File Size After Split

**funcmaps/funcmaps.go:** ~300 lines (core functions)
**funcmaps/markdown.go:** ~150 lines (markdown processing)
**funcmaps/markdown_test.go:** ~140 lines (tests)

All under 300 lines, easy to understand and maintain.

## Next Steps After Funcmaps Split

Once funcmaps is extracted and tested:

1. ✅ **Phase 1 Complete:** Funcmaps independent and tested
2. **Phase 2:** Split codegen (assert/require generation)
3. **Phase 3:** Split docs (documentation generation)
4. **Phase 4:** Simplify parent orchestrator

## Decision Point: Markdown Location

**Question:** Should markdown.go stay in funcmaps/ or move to a docs-specific location?

**Option A:** Keep in funcmaps/ (Recommended)
- ✅ It's template functionality
- ✅ Used by doc generation templates
- ✅ May be useful for other template uses later

**Option B:** Move to docs/
- ❌ Couples it to docs generation
- ❌ Less reusable

**Recommendation:** Keep `markdown.go` in `funcmaps/` - it's template functionality.
