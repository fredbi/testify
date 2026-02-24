# Markdown Link Handling in Documentation Generation

## Problem

The documentation generator needed to handle two types of links found in godoc comments:

1. **Reference-style markdown links**: `[text]: url`
2. **Godoc-style links**: `[errors.Is]`, `[testing.T]`, etc.

These were appearing inside code blocks where they couldn't be rendered properly.

## Solution

The `markdownFormatEnhanced()` function in `funcmap_enhanced.go` processes links in 4 steps:

### Step 1: Extract Reference Link Definitions

```regex
^\[([^\]]+)\]:\s+(.+)$
```

Finds all reference-style link definitions like `[Zero values]: https://...` and stores them in a map.

### Step 2: Convert Used References to Inline Links

For each reference definition, check if it's actually used in the text (e.g., `[Zero values]` without colon).

If used → Convert to inline: `[Zero values](https://go.dev/ref/spec#The_zero_value)`
If unused → Keep as "dangling reference" to append later

### Step 3: Convert Godoc-Style Links

Pattern: `[package.Symbol]` where Symbol contains a dot

Examples:
- `[errors.Is]` → `[errors.Is](https://pkg.go.dev/errors#Is)`
- `[testing.T]` → `[testing.T](https://pkg.go.dev/testing#T)`
- `[fmt.Printf]` → `[fmt.Printf](https://pkg.go.dev/fmt#Printf)`

### Step 4: Append Dangling References

After the Hugo shortcodes (`{{% /expand %}}`), append any unused reference definitions as proper markdown links outside the code blocks.

## Examples

### Dangling Reference (No Usage)

**Input:**
```markdown
Empty asserts value is empty.

# Examples

    success: ""
[Zero values]: https://go.dev/ref/spec#The_zero_value
```

**Output:**
```markdown
Empty asserts value is empty.

{{% expand title="Examples" %}}
{{< tabs >}}
{{% tab title="Examples" %}}
```go
success: ""
```
{{< /tab >}}
{{< /tabs >}}
{{% /expand %}}

[Zero values]: https://go.dev/ref/spec#The_zero_value
```

The reference appears **after** the expand block, where Hugo will render it as a clickable link.

### Used Reference

**Input:**
```markdown
See [Zero values] for definition.

# Examples

    success: ""
[Zero values]: https://go.dev/ref/spec#The_zero_value
```

**Output:**
```markdown
See [Zero values](https://go.dev/ref/spec#The_zero_value) for definition.

{{% expand title="Examples" %}}
...
{{% /expand %}}
```

The reference is converted to an inline link and the definition is removed.

### Godoc Links

**Input:**
```markdown
This wraps [errors.Is] for testing.

# Usage

    assertions.ErrorIs(t, err, target)
```

**Output:**
```markdown
This wraps [errors.Is](https://pkg.go.dev/errors#Is) for testing.

{{% expand title="Examples" %}}
...
{{% /expand %}}
```

## Usage

To switch from the current `markdownFormat` to the enhanced version:

1. In `funcmap.go`, update the funcmap:
   ```go
   "mdformat": markdownFormatEnhanced,  // was: markdownFormat
   ```

2. Regenerate documentation:
   ```bash
   go generate ./...
   ```

## Testing

Run tests:
```bash
go test ./internal/generator -run TestMarkdownFormatEnhanced
```

Tests cover:
- Dangling references (unused definitions)
- Used references (converted to inline)
- Godoc-style links
- Mixed scenarios
