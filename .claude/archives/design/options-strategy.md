# Options Strategy for Package Split

## Current Pattern

```go
// options.go
type (
	Option         func(*options) // Currently unused
	GenerateOption func(*generateOptions)
)

type options struct{} // Empty placeholder

type generateOptions struct { // Real options
	targetPkg     string
	targetRoot    string
	targetDoc     string
	enableForward bool
	enableFormat  bool
	// ... etc
}

// generator.go
type Generator struct {
	options // Embedded (empty)
	source  *model.AssertionPackage
	ctx     *genCtx
}

// doc_generator.go
type DocGenerator struct {
	options // Embedded (empty)
	ctx     *genCtx
	doc     model.Documentation
}
```

## The Challenge

When we split into sub-packages:
- `codegen/generator.go` needs options
- `docs/generator.go` needs options
- Both currently embed private `options` struct

**Question:** How do we share the private `options` struct across sub-packages?

## Options (Pun Intended)

### Option 1: Keep Options in Parent Package ✅ RECOMMENDED

**Structure:**
```go
// generator/options.go (stays in parent)
type Option func(*options)
type GenerateOption func(*generateOptions)
type options struct { ... }
type generateOptions struct { ... }

// generator/codegen/generator.go
import "github.com/.../generator"

type Generator struct {
    opts generator.generateOptions  // Not embedded, just stored
    // ...
}

func New(source *model.AssertionPackage, opts ...generator.GenerateOption) *Generator {
    return &Generator{
        opts: generator.GenerateOptionsWithDefaults(opts),
    }
}

// generator/docs/generator.go
import "github.com/.../generator"

type Generator struct {
    opts generator.generateOptions  // Same pattern
    // ...
}
```

**Pros:**
- ✅ Single source of truth for options
- ✅ No duplication
- ✅ No circular dependencies (parent doesn't import subs)
- ✅ Options stay private to generator package family

**Cons:**
- ⚠️ Sub-packages depend on parent for options (acceptable)

### Option 2: Separate Options Package

**Structure:**
```go
// generator/config/config.go
package config

type Option func(*Options)
type GenerateOption func(*GenerateOptions)
type Options struct { ... }
type GenerateOptions struct { ... }

// generator/codegen/generator.go
import "github.com/.../generator/config"

type Generator struct {
    opts config.GenerateOptions
}

// generator/docs/generator.go
import "github.com/.../generator/config"

type Generator struct {
    opts config.GenerateOptions
}
```

**Pros:**
- ✅ Clean separation
- ✅ No parent dependency

**Cons:**
- ❌ Extra package for what's currently an empty struct
- ❌ More complex than needed

### Option 3: Duplicate Options in Each Sub-package

**Structure:**
```go
// generator/codegen/options.go
type Options struct { ... }

// generator/docs/options.go
type Options struct { ... }
```

**Pros:**
- ✅ Complete independence

**Cons:**
- ❌ Duplication
- ❌ Divergence risk
- ❌ Violates DRY

### Option 4: Make Options Public

**Structure:**
```go
// generator/options.go
type Options struct { // Public
	// exposed fields
}

type GenerateOptions struct { // Public
	// exposed fields
}
```

**Pros:**
- ✅ Simple
- ✅ Easy to use

**Cons:**
- ❌ Breaks encapsulation
- ❌ Exposes internal details
- ❌ Can't evolve privately

## Recommended Approach: Option 1

**Keep options in parent package, sub-packages consume them.**

### Implementation

```go
// generator/options.go (unchanged location)
package generator

type Option func(*options)
type GenerateOption func(*generateOptions)

type options struct {
    // Future shared config
}

type generateOptions struct {
    targetPkg        string
    targetRoot       string
    targetDoc        string
    enableForward    bool
    enableFormat     bool
    enableGenerics   bool
    generateHelpers  bool
    generateTests    bool
    generateExamples bool
    runnableExamples bool
    generateDoc      bool
    formatOptions    *imports.Options
}

// Export the constructor for sub-packages
func GenerateOptionsWithDefaults(opts []GenerateOption) generateOptions {
    // ... existing implementation
}

// generator/codegen/generator.go
package codegen

import (
    "github.com/go-openapi/testify/v2/codegen/internal/generator"
    "github.com/go-openapi/testify/v2/codegen/internal/model"
)

type Generator struct {
    source *model.AssertionPackage
    opts   generator.generateOptions  // Stored, not embedded
}

func New(source *model.AssertionPackage, opts ...generator.GenerateOption) *Generator {
    return &Generator{
        source: source,
        opts:   generator.GenerateOptionsWithDefaults(opts),
    }
}

func (g *Generator) Generate(opts ...generator.GenerateOption) error {
    // Use g.opts or merge with new opts
    finalOpts := generator.GenerateOptionsWithDefaults(opts)
    // ... generation logic
}

// generator/docs/generator.go
package docs

import (
    "github.com/go-openapi/testify/v2/codegen/internal/generator"
    "github.com/go-openapi/testify/v2/codegen/internal/model"
)

type Generator struct {
    doc  model.Documentation
    opts generator.generateOptions  // Same pattern
}

func New(doc model.Documentation, opts ...generator.GenerateOption) *Generator {
    return &Generator{
        doc:  doc,
        opts: generator.GenerateOptionsWithDefaults(opts),
    }
}
```

### Why This Works

1. **No circular dependencies:**
   - Parent package defines options
   - Sub-packages import parent for options
   - Parent orchestrator can import subs for execution

2. **Clear ownership:**
   - Options belong to generator package family
   - Sub-packages are just specialized executors

3. **Future-proof:**
   - Can add common options to `options` struct
   - Sub-packages automatically get them

4. **Similar to scanner pattern:**
   - scanner/comments imports scanner for types
   - scanner/signature imports scanner for types

## Migration Path

### Phase 1: Extract Funcmaps (No Options Issues)
```go
// generator/funcmaps/funcmaps.go
package funcmaps

// All funcmap functions, no options needed
func FuncMap() template.FuncMap { ... }
```

No options coupling - clean split.

### Phase 2: Split Generators (Options Stay in Parent)

**Before:**
```go
// generator/generator.go
type Generator struct {
	options // embedded
	// ...
}
```

**After:**
```go
// generator/codegen/generator.go
type Generator struct {
	opts generator.generateOptions // composition, not embedding
	// ...
}

// generator/options.go
// Export constructor
func GenerateOptionsWithDefaults(opts []GenerateOption) generateOptions {
	// ... existing
}
```

### Phase 3: Parent Orchestrator

```go
// generator/generator.go (simplified orchestrator)
package generator

import (
    "github.com/go-openapi/testify/v2/codegen/internal/generator/codegen"
    "github.com/go-openapi/testify/v2/codegen/internal/generator/docs"
)

// High-level API stays the same
type Generator struct {
    codeGen *codegen.Generator
    docGen  *docs.Generator
}

func New(source *model.AssertionPackage, opts ...Option) *Generator {
    return &Generator{
        codeGen: codegen.New(source, /* generate opts */),
        docGen:  docs.New(/* doc */, /* generate opts */),
    }
}
```

## Summary

✅ **Recommended:** Keep options in parent `generator/` package
✅ Sub-packages import parent for option types
✅ No circular dependencies (parent can import subs)
✅ Single source of truth
✅ Matches scanner package pattern
