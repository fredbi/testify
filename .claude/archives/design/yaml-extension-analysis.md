# YAML Extension Analysis

**Date:** 2026-01-27
**Topic:** Should we add built-in support for alternative YAML libraries (e.g., `enable/goccy-yaml`)?

## Current State

YAML assertions in testify are **opt-in and pluggable**:
- By default, YAML assertions panic with a helpful error message
- Users import `_ "github.com/go-openapi/testify/enable/yaml/v2"` to enable YAML support
- This registers `gopkg.in/yaml.v3` as the unmarshaler
- The registration mechanism is public: `enable/stubs/yaml.EnableYAMLWithUnmarshal()`
- Users can register custom unmarshalers (e.g., goccy/go-yaml)

## The Question

Should we add `enable/goccy-yaml` as a built-in option alongside `enable/yaml`?

## Arguments For Built-in goccy/yaml Support

### 1. Discovery
Users might not realize customization is possible. Having `enable/goccy-yaml` in the repo signals it's supported and
encourages adoption.

### 2. Convenience
Saves users from creating their own wrapper package. Instead of ~10 lines of wrapper code, they just import a different
package.

### 3. Documentation by Example
Shows best practices for the enable pattern. Other users can look at the implementation to understand how to integrate
other libraries.

### 4. Performance Benefits
goccy/go-yaml offers:
- 2-3x faster unmarshaling
- Colored error messages for debugging
- Better error reporting with line/column numbers
- Comment preservation (useful for config testing)
- JSON-like syntax support

## Arguments Against Built-in goccy/yaml Support

### 1. Conflicts with Zero Dependencies Philosophy ⚠️
Testify strongly emphasizes zero external dependencies:
- We internalized `go-spew` rather than depend on it
- We internalized `difflib` rather than depend on it
- The project aims to be `github.com/go-openapi/testify/v2` as the **only** import

Adding `enable/goccy-yaml` would mean:
- Taking on maintenance of that integration
- Tracking goccy/go-yaml releases for breaking changes
- Dealing with user issues related to that library
- Potentially dealing with security advisories
- Creating an exception to the zero-dependency rule

### 2. Slippery Slope ⚠️
If we add goccy/go-yaml, where does it stop?
- `enable/yaml-v2` (many projects still on gopkg.in/yaml.v2)
- `enable/toml` (another common config format)
- `enable/json-iter` (faster JSON library)
- `enable/sonic-json` (even faster JSON)
- `enable/custom-diff` (alternative diff libraries)
- `enable/custom-spew` (alternative pretty-printers)

Each one adds:
- Maintenance burden
- Decision-making complexity
- "Which one should I use?" questions
- Compatibility matrix testing

### 3. Problem Already Solved ⚠️
The `EnableYAMLWithUnmarshal` API + documentation means users can add this in ~10 lines of code:

```go
// User writes this once in their codebase:
package testutil

import (
	goccyyaml "github.com/goccy/go-yaml"
	yamlstub "github.com/go-openapi/testify/v2/enable/stubs/yaml"
)

func init() {
	yamlstub.EnableYAMLWithUnmarshal(goccyyaml.Unmarshal)
}
```

The built-in feature would save almost nothing. The complexity is in:
- Deciding which library to use (user still does this)
- Adding the dependency to go.mod (user still does this)
- The registration call is literally one line

### 4. Picking Winners is Risky ⚠️
Including goccy/go-yaml implies endorsement. Risks:
- **Abandonment:** Library gets abandoned (common in Go ecosystem)
- **Better alternative:** Another library emerges that's superior
- **Version conflicts:** Users have conflicts with other dependencies
- **Security:** Library has security issues and we're tied to it
- **Breaking changes:** Library changes API and breaks our integration

We'd be tied to these decisions and responsible for maintenance.

### 5. Better to Teach the Pattern ⚠️
The documentation showing customization is **more valuable** than one specific implementation because:
- Users learn the pattern and can apply it anywhere
- Works for ANY unmarshaler, not just YAML alternatives
- Demonstrates the flexibility of the architecture
- Users maintain control and ownership
- Shows testify as a framework, not just a library

## Recommendation

**Do NOT add `enable/goccy-yaml` as a built-in feature.**

### Why This is the Right Call

1. **Preserves zero-dependency philosophy:** Only `enable/yaml` has a dependency, and it's opt-in
2. **Empowers users:** They control which library and version
3. **Reduces maintenance:** We don't track goccy/go-yaml changes
4. **Avoids commitment:** No implied endorsement of specific libraries
5. **Demonstrates flexibility:** Documentation shows the pattern works for any library

### Better Alternatives

#### 1. Enhanced Documentation ✅
Already added to `docs/doc-site/usage/USAGE.md`:
- Full explanation of the registration mechanism
- Complete example using goccy/go-yaml
- Comparison table of different libraries
- Advanced patterns (wrapping unmarshalers)

#### 2. Community Integrations Section
Add to docs (e.g., `docs/doc-site/usage/COMMUNITY.md`):

```markdown
## Community Integrations

### YAML Libraries

Users have successfully integrated these YAML libraries:

**[goccy/go-yaml](https://github.com/goccy/go-yaml)**
- Better performance (2-3x faster)
- Colored error messages
- Better error reporting with line/column numbers
- Comment preservation for config testing

**[gopkg.in/yaml.v2](https://gopkg.in/yaml.v2)**
- Legacy projects requiring v2 compatibility
- Established, stable API

See [customization guide](./USAGE.md#customization) for the integration pattern.
```

#### 3. Example Repository (Optional)
Create `go-openapi/testify-examples` with community patterns:
- Multiple YAML library integrations
- Custom assertion helpers
- Advanced testing patterns
- Keep separate from main repo to avoid maintenance burden

#### 4. Multiple Examples in Documentation
Show the pattern with 2-3 different libraries to reinforce flexibility:
- Example 1: goccy/go-yaml (performance)
- Example 2: yaml.v2 (compatibility)
- Example 3: Custom wrapper with validation

This teaches the pattern more effectively than a single built-in implementation.

## When to Reconsider

Add built-in support IF:
1. **Overwhelming user demand:** Multiple users request it with real use cases
2. **Critical compatibility issue:** gopkg.in/yaml.v3 has a breaking bug/issue
3. **Standard emerges:** One alternative becomes the de-facto standard
4. **Maintenance is trivial:** The integration becomes completely zero-maintenance

But wait for actual demand rather than anticipating it.

## Design Principle

The current design is elegant:
- `enable/yaml` → Opt-in feature with one blessed implementation
- `enable/stubs/yaml` → Public API for complete customization
- Documentation → Shows users how to take control

This balances:
- **Convenience** (most users): Import enable/yaml, done
- **Flexibility** (power users): Swap in any library with 10 lines
- **Maintenance** (maintainers): Only one YAML dependency to track
- **Philosophy** (project): Minimal dependencies, maximal extensibility

## Conclusion

**The 10 lines of code users need to write is a feature, not a bug.**

It keeps testify:
- Focused on assertions, not library integrations
- Dependency-free (except explicit opt-ins)
- Flexible and extensible
- Maintainable for the long term

Users who want goccy/go-yaml get full control over version, configuration, and behavior. That's more valuable than a
pre-packaged integration.

---

**TL;DR:** Document the pattern extensively, create a community integrations page, but don't add built-in support for
alternative YAML libraries. The customization API is the feature.
