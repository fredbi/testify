# Generator Package Refactoring Proposal

## Current Structure (Monolithic)

```
internal/generator/
├── generator.go           (494 lines) - Code generation for assert/require
├── funcmap.go             (381 lines) - Template functions
├── doc_domains.go         (366 lines) - Domain organization
├── doc_generator.go       (238 lines) - Documentation generation
├── funcmap_enhanced.go    (136 lines) - Enhanced markdown processing
├── options.go             (123 lines) - Generator options
├── render.go              (36 lines)  - Rendering utilities
├── doc.go                 (8 lines)   - Package documentation
└── templates/             (15 files)  - Template files
    ├── assertion_*.gotmpl
    ├── requirement_*.gotmpl
    └── doc_*.gotmpl
```

**Total: ~2,179 lines in main package**

## Problems with Current Structure

1. **Large monolithic package** - All generation logic in one namespace
2. **Mixed responsibilities** - Code gen, doc gen, template management, rendering all together
3. **Hard to test** - Functions tightly coupled, difficult to unit test in isolation
4. **Template loading duplicated** - Both Generator and DocGenerator load templates similarly
5. **Unclear separation** - funcmap.go used by both code and doc generation

## Proposed Structure (Modular)

Following the `scanner` package pattern with focused sub-packages:

```
internal/generator/
├── doc.go                 - Package documentation
├── options.go             - Public API options (shared)
├── generator.go           - Main Generator orchestrator (simplified)
│
├── templates/             - Template management and rendering
│   ├── doc.go
│   ├── loader.go          - Template loading logic
│   ├── registry.go        - Template registry/cache
│   ├── render.go          - Rendering functions (Go + Markdown)
│   ├── funcmap.go         - Core template functions
│   ├── funcmap_markdown.go - Markdown-specific functions (from funcmap_enhanced)
│   ├── funcmap_test.go    - Tests for template functions
│   └── assets/            - Embedded template files
│       ├── assertion_*.gotmpl
│       ├── requirement_*.gotmpl
│       └── doc_*.gotmpl
│
├── codegen/               - Code generation (assert/require packages)
│   ├── doc.go
│   ├── generator.go       - Code generator implementation
│   ├── generator_test.go  - Code generation tests
│   ├── transformer.go     - Model transformation logic
│   └── targets.go         - Target package definitions (assert, require)
│
└── docs/                  - Documentation generation
    ├── doc.go
    ├── generator.go       - Documentation generator implementation
    ├── domains.go         - Domain organization and metadata
    ├── domains_test.go    - Domain organization tests
    └── indexer.go         - Index page generation
```

## Responsibilities by Sub-package

### 1. `generator/templates/` - Template Management & Rendering

**Purpose:** Centralized template loading, caching, and rendering

**Files:**
- `loader.go` - Load templates from embedded FS
- `registry.go` - Template registry with lazy loading
- `render.go` - render() and renderMD() functions
- `funcmap.go` - Core template functions (from current funcmap.go)
- `funcmap_markdown.go` - Markdown processing (from funcmap_enhanced.go)
- `assets/*.gotmpl` - Embedded template files

**Public API:**
```go
type Registry struct { ... }
func NewRegistry() *Registry
func (r *Registry) Load(name string) (*template.Template, error)
func (r *Registry) Render(tpl *template.Template, data any) ([]byte, error)
func (r *Registry) RenderGo(tpl *template.Template, data any) ([]byte, error)
func (r *Registry) RenderMarkdown(tpl *template.Template, data any) ([]byte, error)
```

**Benefits:**
- Single source of truth for template management
- Shared by both code and doc generation
- Easy to test template functions in isolation
- Clear separation of rendering concerns

### 2. `generator/codegen/` - Code Generation

**Purpose:** Generate assert and require packages from model

**Files:**
- `generator.go` - Main CodeGenerator type and logic
- `transformer.go` - Transform model for different targets (assert vs require)
- `targets.go` - Target-specific configuration (assert, require)
- `generator_test.go` - Unit tests

**Extracted from:**
- Current `generator.go` (main generation logic)
- Parts of `generator.go` dealing with model transformation

**Public API:**
```go
type CodeGenerator struct { ... }
func NewCodeGenerator(source *model.AssertionPackage, opts ...Option) *CodeGenerator
func (g *CodeGenerator) Generate(target string, opts ...GenerateOption) error
func (g *CodeGenerator) GenerateAssert(opts ...GenerateOption) error
func (g *CodeGenerator) GenerateRequire(opts ...GenerateOption) error
func (g *CodeGenerator) Documentation() model.Documentation
```

**Benefits:**
- Focused on code generation only
- Clear separation from documentation
- Easier to test transformation logic
- Can add new targets (e.g., generics variants) easily

### 3. `generator/docs/` - Documentation Generation

**Purpose:** Generate markdown documentation for Hugo site

**Files:**
- `generator.go` - Main DocGenerator type and logic
- `domains.go` - Domain organization (from doc_domains.go)
- `indexer.go` - Index page generation logic
- `domains_test.go` - Tests for domain organization

**Extracted from:**
- Current `doc_generator.go`
- Current `doc_domains.go`

**Public API:**
```go
type DocGenerator struct { ... }
func NewDocGenerator(doc model.Documentation, opts ...Option) *DocGenerator
func (g *DocGenerator) Generate(targetDir string, opts ...GenerateOption) error
func (g *DocGenerator) OrganizeByDomain() (iter.Seq2[string, model.Document], error)
```

**Benefits:**
- Self-contained documentation generation
- Domain logic separated and testable
- Can evolve independently from code generation
- Easy to add new documentation formats

### 4. `generator/` - Main Package (Orchestrator)

**Purpose:** Public API and orchestration

**Files:**
- `doc.go` - Package documentation
- `options.go` - Shared options for all generators
- `generator.go` - Main orchestrator (simplified)

**Public API:**
```go
// High-level convenience API
func GenerateAll(source *model.AssertionPackage, opts ...Option) error
func GenerateCode(source *model.AssertionPackage, opts ...Option) (*codegen.CodeGenerator, error)
func GenerateDocs(doc model.Documentation, opts ...Option) (*docs.DocGenerator, error)
```

## Migration Strategy

### Phase 1: Create Sub-packages (No Breaking Changes)
1. Create `templates/` sub-package
   - Move template loading logic
   - Move funcmap.go functions
   - Move render.go utilities
   - Keep public exports

2. Create `codegen/` sub-package
   - Move Generator type and code gen logic
   - Keep public API compatible

3. Create `docs/` sub-package
   - Move DocGenerator type and doc gen logic
   - Move domain organization

### Phase 2: Update Imports
1. Update `generator/generator.go` to use sub-packages
2. Update `main.go` to use new structure
3. Update tests

### Phase 3: Cleanup
1. Remove old files from main package
2. Consolidate duplicate logic
3. Add comprehensive tests for sub-packages

## Testing Benefits

### Current Testing Issues
- Functions in generator.go hard to test in isolation
- Template functions mixed with generation logic
- Domain organization not independently tested

### After Refactoring
```go
// templates/funcmap_test.go
func TestMarkdownFormat(t *testing.T) { ... }
func TestGodocLinks(t *testing.T) { ... }

// codegen/transformer_test.go
func TestTransformModel(t *testing.T) { ... }
func TestTransformForAssert(t *testing.T) { ... }

// docs/domains_test.go
func TestOrganizeByDomain(t *testing.T) { ... }
func TestDomainMetadata(t *testing.T) { ... }
```

## File Size Comparison

**Before:**
- generator.go: 494 lines
- funcmap.go: 381 lines
- doc_domains.go: 366 lines
- doc_generator.go: 238 lines

**After:**
- templates/loader.go: ~100 lines
- templates/registry.go: ~80 lines
- templates/render.go: ~60 lines
- templates/funcmap.go: ~250 lines
- templates/funcmap_markdown.go: ~150 lines
- codegen/generator.go: ~200 lines
- codegen/transformer.go: ~150 lines
- codegen/targets.go: ~80 lines
- docs/generator.go: ~150 lines
- docs/domains.go: ~200 lines
- docs/indexer.go: ~100 lines

**All files under 250 lines**, easy to understand and test.

## Parallels with Scanner Package

| Scanner | Generator (Proposed) |
|---------|---------------------|
| `scanner/` | `generator/` (orchestrator) |
| `scanner/comments/` | `generator/templates/` |
| `scanner/comments-parser/` | `generator/codegen/` |
| `scanner/signature/` | `generator/docs/` |

Both follow the same pattern:
- Main package as orchestrator
- Focused sub-packages with clear responsibilities
- Each sub-package independently testable
- Separation of concerns

## Open Questions

1. Should `funcmap_markdown.go` stay in templates/ or move to docs/?
   - **Recommendation:** Keep in templates/ - it's template functionality, used by docs

2. Should template registry be a singleton or instance-based?
   - **Recommendation:** Instance-based for testability

3. Should we keep backwards compatibility in the main API?
   - **Recommendation:** Yes, main Generator API stays the same, internal refactored

4. Where should the embedded templates live?
   - **Recommendation:** templates/assets/ with templates package

## Benefits Summary

✅ **Smaller, focused files** - All files under 250 lines
✅ **Clear separation of concerns** - Each sub-package has one responsibility
✅ **Better testability** - Can test components in isolation
✅ **Easier to maintain** - Changes localized to relevant sub-package
✅ **Follows established pattern** - Consistent with scanner package structure
✅ **No breaking changes** - Main API remains compatible
✅ **Future extensibility** - Easy to add new generators or targets
