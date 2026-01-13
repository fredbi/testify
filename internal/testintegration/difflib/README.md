# difflib comparison

Compare side by side:

1. `github.com/go-openapi/testify/v2/internal/difflib` (our internalized library - yardstick)
2. `github.com/hexops/gotextdiff` (Myers algorithm, used by gopls)
3. `github.com/aymanbagabas/go-udiff` (modern unified diff library)
4. `github.com/sergi/go-diff/diffmatchpatch` (Google's diff-match-patch port)

Colorization is no longer a comparison criterion (handled separately in `enable/colors`).

## Summary of Findings (2026-01-12)

### Output Quality

| Library | Output Format | Line-based | Unified Diff | Readability |
|---------|--------------|------------|--------------|-------------|
| Our difflib | Unified diff | Yes | Native | Excellent |
| gotextdiff | Unified diff | Yes | Native | Excellent |
| go-udiff | Unified diff | Yes | Native | Excellent |
| sergi/go-diff | Character-level | No (requires conversion) | Via patches | Poor for structured data |

**Key observations:**

1. **Our difflib and gotextdiff produce identical output** for the unified diff format.
   Both use the Myers diff algorithm and produce clean, readable line-based diffs.

2. **go-udiff produces similar output but with different grouping** - it interleaves
   deletions and insertions (each `-` followed by its `+`), while our difflib and
   gotextdiff group all deletions followed by all insertions. Both are valid unified
   diff formats.

3. **sergi/go-diff is fundamentally different** - it's designed for character-level
   diffing (text editing, spell checking) rather than line-based diffs. The patch
   conversion produces URL-encoded output that's hard to read.

4. **sergi/go-diff has stability issues** - it can panic on certain inputs
   (observed with larger nested structures).

### Performance (nested struct diff)

```
BenchmarkDiffLibs/our_difflib-16      85182    15821 ns/op   20399 B/op    183 allocs/op
BenchmarkDiffLibs/gotextdiff-16       66986    18124 ns/op   16513 B/op    102 allocs/op
BenchmarkDiffLibs/go_udiff-16         24613    48964 ns/op   18430 B/op    156 allocs/op
BenchmarkDiffLibs/sergi_godiff-16      8048   165871 ns/op   75535 B/op   1247 allocs/op
```

| Library | Speed | Memory | Allocations | vs Our difflib |
|---------|-------|--------|-------------|----------------|
| Our difflib | ~15.8 µs | 20.4 KB | 183 | baseline |
| gotextdiff | ~18.1 µs | 16.5 KB | 102 | 1.15x slower |
| go-udiff | ~49.0 µs | 18.4 KB | 156 | 3.1x slower |
| sergi/go-diff | ~165.9 µs | 75.5 KB | 1247 | 10.5x slower |

**Performance notes:**

1. **Our difflib is the fastest** - slightly faster than gotextdiff while producing
   identical output.

2. **gotextdiff uses fewer allocations** (102 vs 183) but is ~15% slower overall.
   The allocation reduction doesn't translate to speed improvement.

3. **go-udiff is ~3x slower** than our difflib/gotextdiff despite similar allocation
   count and memory usage. The interleaved output style may contribute to overhead.

4. **sergi/go-diff is ~10x slower** with 6x more memory and 12x more allocations.
   This is expected - character-level diffing is more expensive than line-level.

### Recommendation

**No action needed.** Our internalized difflib is:

- **Fastest** among all tested libraries
- Producing excellent output quality (identical to gotextdiff)
- Already supports our colorization hooks via printer builders
- Stable (no panics observed)

The only potential improvement would be to adopt gotextdiff's allocation pattern
(fewer, larger allocations), but the benefit would be marginal and might not
translate to actual speed improvement.

## Output Style Comparison

### Our difflib / gotextdiff (grouped style)

Deletions are grouped, then insertions:

```diff
- Age: (int) 30,
- Email: (string) (len=17) "alice@example.com",
+ Age: (int) 31,
+ Email: (string) (len=23) "alice.smith@example.com",
```

### go-udiff (interleaved style)

Each deletion is immediately followed by its corresponding insertion:

```diff
- Age: (int) 30,
+ Age: (int) 31,
- Email: (string) (len=17) "alice@example.com",
+ Email: (string) (len=23) "alice.smith@example.com",
```

Both are valid unified diff formats. The grouped style (our current output) is
more traditional and matches what `diff -u` produces.

## Test Cases

The comparison tests cover:

1. **simple_struct** - Basic struct with field changes
2. **simple_map** - Map with key/value changes
3. **simple_slice** - Slice with element changes and additions
4. **nested_struct** - Deeply nested struct with multiple changes
5. **mixed_changes** - Complex map with mixed value types

Run tests with:

```bash
go test -v -run TestDiffLibComparison ./...
```

Run benchmarks with:

```bash
go test -bench=BenchmarkDiffLibs -benchmem ./...
```

## Example Output

### Our difflib (yardstick)

```diff
--- original
+++ modified
@@ -1,7 +1,7 @@
 (difflib.Person) {
  Name: (string) (len=5) "Alice",
- Age: (int) 30,
- Email: (string) (len=17) "alice@example.com",
+ Age: (int) 31,
+ Email: (string) (len=23) "alice.smith@example.com",
  Address: (difflib.Address) {
   Street: (string) "",
   City: (string) "",
```

### gotextdiff

Produces identical output to our difflib for unified diff format.

### go-udiff

Similar output but with interleaved deletions/insertions (see style comparison above).

### sergi/go-diff

```
=== Character-level diff (native) ===
[ ] ...(65 chars)...
[-] 0
[+] 1
[ ] ,\n Email: (string) (len=
[-] 17
[+] 23
[ ] ) "alice
[+] .smith
[ ] ...(162 chars)...
```

Character-level output is fine for text editing but poor for code/data comparison.
