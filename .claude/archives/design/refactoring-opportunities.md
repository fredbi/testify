# Generator Refactoring Opportunities

Analysis of redundancies and code sharing opportunities between `Generator` and `DocGenerator`.

## Current State

### Generator (generator.go)
- Generates assert/require packages from internal/assertions
- Uses `genCtx` with templates, index, target, formatOptions
- Has `loadTemplates()`, `render()`, `transformModel()`, etc.
- Calls `render()` for Go code (with goimports formatting)

### DocGenerator (doc_generator.go)
- Generates documentation markdown from accumulated docs
- Uses `genCtx` with templates, index
- Has `loadTemplates()`, `render()`, `reorganizeByDomain()`, etc.
- Calls `renderMD()` for markdown (no formatting)

## Identified Redundancies

### 1. **Nearly Identical `render()` Methods** ⭐⭐⭐

Both generators have the same render method structure:

```go
import "fmt"

// Generator.render()
func (g *Generator) render(name string, target string, data any) error {
	tplName, ok := g.ctx.index[name]
	if !ok {
		panic(fmt.Errorf("internal error: expect template name %q", name))
	}

	tpl, ok := g.ctx.templates[tplName]
	if !ok {
		panic(fmt.Errorf("internal error: expect template %q", name))
	}

	return render(tpl, target, data, g.ctx.formatOptions)
}

// DocGenerator.render()
func (d *DocGenerator) render(name string, target string, data any) error {
	tplName, ok := d.ctx.index[name]
	if !ok {
		panic(fmt.Errorf("internal error: expect template name %q", name))
	}

	tpl, ok := d.ctx.templates[tplName]
	if !ok {
		panic(fmt.Errorf("internal error: expect template %q", name))
	}

	return renderMD(tpl, target, data)
}
```

**Difference:** Only the final render function call (`render()` vs `renderMD()`).

### 2. **Duplicate `loadTemplates()` Logic** ⭐⭐⭐

Both have very similar template loading:

```go
import (
	"fmt"
	"path"
	"sort"
	"text/template"
)

// Both generators
func loadTemplates() error {
	const (
		tplExt            = ".gotmpl" // or ".md.gotmpl"
		expectedTemplates = 10
	)

	index := make(map[string]string, expectedTemplates)
	// ... populate index ...

	d.ctx.index = index
	needed := make([]string, 0, len(index))
	for _, v := range index {
		needed = append(needed, v)
	}
	sort.Strings(needed)

	d.ctx.templates = make(map[string]*template.Template, len(needed))
	for _, name := range needed {
		file := name + tplExt
		tpl, err := template.New(file).Funcs(funcmaps.FuncMap()).ParseFS(templatesFS, path.Join("templates", file))
		if err != nil {
			return fmt.Errorf("failed to load template %q from %q: %w", name, file, err)
		}

		d.ctx.templates[name] = tpl
	}

	return nil
}
```

**Differences:**
- Template extension (`.gotmpl` vs `.md.gotmpl`)
- How the index is populated (different logic)

### 3. **Shared Context Fields** ⭐⭐

Both use `genCtx` with overlapping fields:

```go
import "text/template"

type genCtx struct {
	generateOptions

	index      map[string]string             // SHARED
	templates  map[string]*template.Template // SHARED
	target     *model.AssertionPackage       // Generator only
	docs       *model.Documentation          // Generator only
	targetBase string                        // Generator only
}
```

DocGenerator has its own inline anonymous struct with similar fields.

### 4. **Shared Constants** ⭐

```go
import "embed"

const (
	dirPermissions  = 0o750 // SHARED
	filePermissions = 0o600 // SHARED
)

//go:embed templates/*.gotmpl
var templatesFS embed.FS // SHARED
```

### 5. **Template Index Building Pattern** ⭐

Both follow the same pattern:
1. Create `map[string]string` index
2. Populate with logical name → template file name mappings
3. Extract values to slice
4. Sort alphabetically
5. Load each template

## Refactoring Proposals

### Proposal 1: Extract Common Template Loading ⭐⭐⭐

**Impact:** High - eliminates major duplication

Create shared template loading infrastructure:

```go
import (
	"embed"
	"fmt"
	"path"
	"sort"
	"text/template"
)

// Common template context
type templateCtx struct {
	index     map[string]string
	templates map[string]*template.Template
}

// Common template loader
func loadTemplatesFromIndex(
	index map[string]string,
	tplExt string,
	fs embed.FS,
) (map[string]*template.Template, error) {
	needed := make([]string, 0, len(index))
	for _, v := range index {
		needed = append(needed, v)
	}
	sort.Strings(needed)

	templates := make(map[string]*template.Template, len(needed))
	for _, name := range needed {
		file := name + tplExt
		tpl, err := template.New(file).Funcs(funcmaps.FuncMap()).ParseFS(fs, path.Join("templates", file))
		if err != nil {
			return nil, fmt.Errorf("failed to load template %q from %q: %w", name, file, err)
		}
		templates[name] = tpl
	}

	return templates, nil
}
```

**Usage:**

```go
// Generator
func (g *Generator) loadTemplates() error {
	index := make(map[string]string, 10)
	// ... populate index based on targetBase ...

	templates, err := loadTemplatesFromIndex(index, ".gotmpl", templatesFS)
	if err != nil {
		return err
	}

	g.ctx.index = index
	g.ctx.templates = templates
	return nil
}

// DocGenerator
func (d *DocGenerator) loadTemplates() error {
	index := map[string]string{
		"doc_index": "doc_index",
		"doc_page":  "doc_page",
	}

	templates, err := loadTemplatesFromIndex(index, ".md.gotmpl", templatesFS)
	if err != nil {
		return err
	}

	d.ctx.index = index
	d.ctx.templates = templates
	return nil
}
```

### Proposal 2: Unified Render Method ⭐⭐⭐

**Impact:** High - eliminates duplicate render logic

Create a render function type and use it as a parameter:

```go
import (
	"fmt"
	"text/template"
)

type renderFunc func(*template.Template, string, any) error

// Common render implementation
func renderTemplate(
	index map[string]string,
	templates map[string]*template.Template,
	name string,
	target string,
	data any,
	renderFn renderFunc,
) error {
	tplName, ok := index[name]
	if !ok {
		return fmt.Errorf("internal error: expect template name %q", name)
	}

	tpl, ok := templates[tplName]
	if !ok {
		return fmt.Errorf("internal error: expect template %q", name)
	}

	return renderFn(tpl, target, data)
}
```

**Usage:**

```go
import "text/template"

// Generator
func (g *Generator) render(name string, target string, data any) error {
	return renderTemplate(
		g.ctx.index,
		g.ctx.templates,
		name,
		target,
		data,
		func(tpl *template.Template, target string, data any) error {
			return render(tpl, target, data, g.ctx.formatOptions)
		},
	)
}

// DocGenerator
func (d *DocGenerator) render(name string, target string, data any) error {
	return renderTemplate(
		d.ctx.index,
		d.ctx.templates,
		name,
		target,
		data,
		renderMD,
	)
}
```

### Proposal 3: Embed Common Template Context ⭐⭐

**Impact:** Medium - reduces struct duplication

```go
import "html/template"

// Shared template infrastructure
type templateContext struct {
	index     map[string]string
	templates map[string]*template.Template
}

type genCtx struct {
	templateContext // EMBEDDED
	generateOptions

	target     *model.AssertionPackage
	docs       *model.Documentation
	targetBase string
}

// DocGenerator can also embed
type docGenCtx struct {
	templateContext // EMBEDDED
	generateOptions
}
```

### Proposal 4: Extract Template Index Builder ⭐

**Impact:** Low-Medium - improves code clarity

```go
import "sort"

// Helper to build and sort template index
func buildTemplateIndex(index map[string]string) []string {
	needed := make([]string, 0, len(index))
	for _, v := range index {
		needed = append(needed, v)
	}
	sort.Strings(needed)
	return needed
}
```

## Recommended Implementation Order

1. **Phase 1: Low-risk extraction** (Proposal 4)
   - Extract `buildTemplateIndex()` helper
   - Test: Verify both generators still work

2. **Phase 2: Template loading** (Proposal 1)
   - Extract `loadTemplatesFromIndex()` function
   - Update both generators to use it
   - Test: Run full test suite

3. **Phase 3: Render unification** (Proposal 2)
   - Create `renderTemplate()` function
   - Update both generators to use it
   - Test: Verify all code generation and docs work

4. **Phase 4: Context consolidation** (Proposal 3, optional)
   - Extract `templateContext` struct
   - Embed in both `genCtx` types
   - Test: Full integration tests

## Benefits

**Code Reduction:**
- ~30-40 lines of duplicate code eliminated
- Reduced maintenance burden
- Single source of truth for template operations

**Improved Maintainability:**
- Changes to template loading logic only need to be made once
- Clearer separation of concerns
- Easier to test template infrastructure independently

**Better Consistency:**
- Both generators use identical template patterns
- Reduces chance of divergence
- Makes codebase easier to understand

## Risks & Considerations

**Low Risk:**
- Proposals 1, 2, 4 are straightforward extractions
- No changes to public APIs
- Easy to test incrementally

**Medium Risk:**
- Proposal 3 (context consolidation) touches more code
- Consider as optional enhancement

**Testing Strategy:**
- Each phase should be followed by full test suite
- Verify code generation output matches before/after
- Check documentation generation still works

## Notes

- There's already a TODO comment in `doc_generator.go:197` noting the duplication
- Both generators deliberately kept in same package for obvious sharing
- This refactoring maintains that design while reducing repetition
- All shared code would remain in the `generator` package
